package relay

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/RabbyHub/derelay/config"
	"github.com/RabbyHub/derelay/log"
	"github.com/RabbyHub/derelay/metrics"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type WsServer struct {
	config *config.WsConfig

	// connection maintenance
	clients    map[*client]struct{}
	register   chan *client
	unregister chan ClientUnregisterEvent

	redisConn    *redis.Client
	redisSubConn *redis.PubSub

	publishers  *TopicClientSet
	subscribers *TopicClientSet

	localCh chan SocketMessage // for handling message of local clients
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO Only white list allowed origins
	},
}

func NewWSServer(config *config.Config) *WsServer {
	ws := &WsServer{
		config: &config.WsServerConfig, // config

		clients:    make(map[*client]struct{}),
		register:   make(chan *client, 4096),
		unregister: make(chan ClientUnregisterEvent, 4096),

		publishers:  NewTopicClientSet(),
		subscribers: NewTopicClientSet(),

		localCh: make(chan SocketMessage, 2),
	}
	ws.redisConn = redis.NewClient(&redis.Options{
		Addr:     config.RedisServerConfig.ServerAddr,
		Password: config.RedisServerConfig.Password,
		DB:       0,
	})
	ws.redisSubConn = ws.redisConn.Subscribe(context.TODO())

	return ws
}

func (ws *WsServer) NewClientConn(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// ignore the clients who ain't mean to do websocket communication with us
		return
	}

	client := &client{
		conn:      conn,
		id:        generateRandomBytes16(),
		ws:        ws,
		pubTopics: NewTopicSet(),
		subTopics: NewTopicSet(),
		sendbuf:   make(chan SocketMessage, 8),
		quit:      make(chan struct{}),
	}

	ws.register <- client

	go client.read()
	go client.write()
}

func (ws *WsServer) Run() {
	log.Info("Websocket server has been started")

	remoteCh := ws.redisSubConn.Channel()

	for {
		select {
		case message := <-ws.localCh:
			// local message could be "pub", "sub" or "ack" or "ping"
			// pub/sub message handler may contain time-consuming operations(e.g. read/write redis)
			// so put them in separate goroutine to avoid blocking wsserver main loop
			switch message.Type {
			case Pub:
				// do not modify wsserver's local variable in seperate goroutine
				message.client.pubTopics.Set(message.Topic)
				ws.publishers.Set(message.Topic, message.client)
				go ws.pubMessage(message)
				log.Info("local message", zap.Any("client", message.client), zap.Any("message", message))
			case Sub:
				message.client.subTopics.Set(message.Topic)
				ws.subscribers.Set(message.Topic, message.client)
				go ws.subMessage(message)
				log.Info("local message", zap.Any("client", message.client), zap.Any("message", message))
			case Ping:
				ws.handlePingMessage(message)
			}
		case chmessage := <-remoteCh:

			message := SocketMessage{}
			err := json.Unmarshal([]byte(chmessage.Payload), &message)
			if err != nil {
				log.Warn("malformed message from remote", zap.String("payload", chmessage.Payload))
				continue
			}
			log.Info("remote message", zap.Any("message", message))

			// if message is not from `dappNotifyChan`, then must be from `messageChan` and must be a "pub" message
			if !fromDappNotifyChan(chmessage.Channel) {
				for _, subscriber := range ws.GetSubscriber(message.Topic) {
					log.Info("forward to subscriber", zap.Any("client", subscriber), zap.Any("message", message))
					subscriber.send(message)
				}
				continue
			}

			// otherwise the message must be from the `dappNofityChan` channel
			// messages from chanNotifyDapp could be:
			//  * SessionReceived
			//	* SessionSuspended
			//	* SessionResumed
			// 	* relay generated fake "ack" for the wallet
			for _, publisher := range ws.GetDappPublisher(message.Topic) {
				log.Debug("wallet updates, notify dapp", zap.Any("client", publisher), zap.Any("message", message))
				publisher.send(message)
			}

		case client := <-ws.register:
			metrics.IncNewConnection()
			ws.clients[client] = struct{}{}
			metrics.SetCurrentConnections(len(ws.clients))

		case unregisterEvent := <-ws.unregister:
			client, reason := unregisterEvent.client, unregisterEvent.reason

			ws.handleClientDisconnect(client)
			delete(ws.clients, client)

			metrics.IncClosedConnection()
			metrics.SetCurrentConnections(len(ws.clients))
			log.Info("client disconnected", zap.Any("client", client), zap.String("reason", reason.Error()))
		}
	}
}

func (ws *WsServer) GetSubscriber(topic string) []*client {
	clients := []*client{}
	for client := range ws.subscribers.Get(topic) {
		clients = append(clients, client)
	}
	return clients
}

// GetDappPulisher gets the topic publisher and check whether its role is "dapp"
// the return value indicates whether the notification target has been found
func (ws *WsServer) GetDappPublisher(topic string) []*client {
	dapps := []*client{}
	for client := range ws.publishers.Get(topic) {
		if client.role == Dapp {
			dapps = append(dapps, client)
		}
	}
	return dapps
}

// getCachedMessages gets pending notifications from cache by topic
// you can set `clear` to true if you want clear the pending notifications meanwhile
func (ws *WsServer) getCachedMessages(topic string, clear bool) []SocketMessage {
	// Retrieve the notifications from Redis by topic
	notificationBytes, err := ws.redisConn.LRange(context.TODO(), cachedMessageKey(topic), 0, -1).Result()
	if err != nil {
		log.Warn("get cached messages failed", zap.String("topic", topic), zap.Error(err))
		return nil
	}

	// Deserialize the notifications from JSON
	notifications := make([]SocketMessage, 0, len(notificationBytes))
	for _, nb := range notificationBytes {
		var n SocketMessage
		err := json.Unmarshal([]byte(nb), &n)
		if err != nil {
			log.Error("malformed message, unmarshal failed", nil, zap.Any("topic", topic), zap.Any("notification", nb))
			return nil
		}
		notifications = append(notifications, n)
	}

	if clear && len(notifications) > 0 {
		go func() {
			metrics.DecCachedMessages()
			ws.redisConn.Del(context.TODO(), cachedMessageKey(topic))
		}()
	}

	return notifications
}

func (ws *WsServer) Shutdown() {
}
