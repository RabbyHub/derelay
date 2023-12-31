package relay

import (
	"context"
	"time"

	"github.com/RabbyHub/derelay/log"
	"github.com/RabbyHub/derelay/metrics"
	"go.uber.org/zap"
)

type WsMessageHandler func(*WsServer, SocketMessage)

func (ws *WsServer) pubMessage(message SocketMessage) {
	topic := message.Topic
	publisher := message.client

	if message.Role == string(Dapp) {
		// this `notifyDappChanName(topic)` redis channel is used to notify dapp the wallet's status
		ws.redisSubConn.Subscribe(context.TODO(), dappNotifyChanKey(topic))
		if message.Phase == string(SessionStart) {
			metrics.IncEstablishedSessions()
			return
		}
	}

	log.Debug("publish message", zap.Any("client", publisher), zap.Any("topic", message.Topic))

	metrics.IncTotalMessages()
	key := messageChanKey(topic)
	if count, _ := ws.redisConn.Publish(context.TODO(), key, message).Result(); count >= 1 {
		log.Debug("message published", zap.Any("client", publisher), zap.Any("topic", topic))
		if publisher.role == Dapp {
			publisher.send(SocketMessage{
				Topic: message.Topic,
				Type:  Ack,
				Role:  string(Wallet),
			})
		}
	} else {
		log.Debug("cache message", zap.Any("client", publisher), zap.Any("topic", topic))
		metrics.IncCachedMessages()
		if message.Phase == string(SessionRequest) {
			metrics.IncNewRequestedSessions()
		}
		ws.cacheMessage(message, ws.config.MessageCacheTime)
	}
}

func (ws *WsServer) subMessage(message SocketMessage) {
	topic := message.Topic
	subscriber := message.client

	if err := ws.redisSubConn.Subscribe(context.TODO(), messageChanKey(topic)); err != nil {
		log.Warn("[redisSub] subscribe to topic fail", zap.String("topic", topic), zap.Any("client", subscriber))
	}
	log.Debug("subscribe to topic", zap.String("topic", topic), zap.Any("client", subscriber))

	// forward cached notificatoins if there's any
	notifications := ws.getCachedMessages(topic, true)
	log.Debug("pending notifications", zap.String("topic", topic), zap.Any("num", len(notifications)), zap.Any("client", subscriber))
	for _, notification := range notifications {
		subscriber.send(notification)
	}

	// we need do some more work if it's a wallet that subscribes the topic
	if message.Role != string(Dapp) {
		for _, noti := range notifications {
			// When a wallet subscribe to a topic, either it've just scanned the QRCode to receive the session request
			// or it've just waken up from hibernation and trying to recovering the connection
			//
			// NOTE
			// During its whole lifetime, the wallet subscribes to TWO topics, the first is inferred from the QRCode
			// to received the session request from the dapp, the second is generated by itself locally and used for
			// receiving future messages from dapp.
			//
			// When the wallet scans the QRCode, two subscription messages for the two topics issued consecutively.
			// The first topic is only used once and discarded right after received the session request, while the second
			// topic is subscribed everytime the wallet wakes up from hibernation.
			//
			// Just FYI, for dapp, during its whole lifetime, it only subscribe ONE topic, i.e. the once he uses to receive
			// messages from wallet

			if noti.Phase == string(SessionRequest) { // handle the 1st case stated above
				metrics.IncReceivedSessions()
				log.Debug("session been scanned", zap.Any("topic", topic), zap.Any("client", subscriber))

				// notify the topic publisher, aka the dapp, that the session request has been received by wallet
				key := dappNotifyChanKey(noti.Topic)
				ws.redisConn.Publish(context.TODO(), key, SocketMessage{
					Topic: noti.Topic,
					Phase: string(SessionReceived),
					Type:  Ack,
					Role:  string(Relay),
				})
			}
		}

		// handle the 2nd case
		// NOTE we could check for whether the notifactions of this topic is session request, we don't need reply `sessionResumed`
		// for sessionRequest message, but for simplity we don't do that check here
		key := dappNotifyChanKey(message.Topic)
		ws.redisConn.Publish(context.TODO(), key, SocketMessage{
			Topic: message.Topic,
			Type:  Pub,
			Role:  string(Relay),
			Phase: string(SessionResumed),
		})
	}
}

func (ws *WsServer) handlePingMessage(message SocketMessage) {
	// response to application layer ping message
	client := message.client
	client.send(SocketMessage{
		Type: Pong,
		Role: string(Relay),
	})
}

func (ws *WsServer) cacheMessage(message SocketMessage, cacheTime int) {
	key := cachedMessageKey(message.Topic)
	// Store the notification in Redis with the topic as the key
	if _, err := ws.redisConn.RPush(context.TODO(), key, message).Result(); err != nil {
		log.Warn("cache message to redis fail", zap.Any("message", message), zap.Error(err))
		return
	}
	if _, err := ws.redisConn.Expire(context.TODO(), key, time.Duration(cacheTime)*time.Second).Result(); err != nil {
		log.Warn("set expire for cache message failed", zap.Any("key", key), zap.Any("ttl", cacheTime))
		return
	}
}

func (ws *WsServer) handleClientDisconnect(client *client) {

	// clear the client from the subscribed and published topics
	channelsToClear := []string{}
	//subscribedTopics := ws.subscribers.GetTopicsByClient(client, true)
	subscribedTopics := client.subTopics.Get()

	for topic := range subscribedTopics {
		ws.subscribers.Unset(topic, client)
		if ws.subscribers.Len(topic) == 0 {
			ws.subscribers.Clear(topic)
			channelsToClear = append(channelsToClear, messageChanKey(topic))
		}
	}
	for topic := range client.pubTopics.Get() {
		ws.publishers.Unset(topic, client)
		if ws.publishers.Len(topic) == 0 {
			ws.publishers.Clear(topic)
			channelsToClear = append(channelsToClear, messageChanKey(topic))
			// for dapp, need to further clear notify channels
			if client.role == Dapp {
				channelsToClear = append(channelsToClear, dappNotifyChanKey(topic))
			}
		}
	}

	if len(channelsToClear) > 0 {
		log.Info("clear channels", zap.Any("client", client), zap.Any("channels", channelsToClear))
		// !!! WARNING !!!
		// Only call `Unsubscribe` when the length of `channelsToClear` IS NOT 0.
		// Otherwise redis will unsubscribe all of the previous subscribed channels!!!
		go ws.redisSubConn.Unsubscribe(context.TODO(), channelsToClear...)
	}

	// if the client is wallet, notify the topic publisher that wallet has disconnected
	if client.role == Dapp {
		return
	}
	for topic := range subscribedTopics {
		go func(topic string) {
			key := dappNotifyChanKey(topic)
			ws.redisConn.Publish(context.TODO(), key, SocketMessage{
				Topic: topic,
				Type:  Pub,
				Role:  string(Wallet),
				Phase: string(SessionSuspended),
			})
			log.Debug("notify dapp about the wallet suspension", zap.Any("client", client))
		}(topic)
	}
}
