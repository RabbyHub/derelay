package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/RabbyHub/derelay/log"
	"github.com/RabbyHub/derelay/metrics"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type client struct {
	conn *websocket.Conn
	ws   *WsServer

	id        string   // randomly generate, just for logging
	active    bool     // heartbeat related
	role      RoleType // dapp or wallet
	session   string   // session id
	pubTopics *TopicSet
	subTopics *TopicSet

	sendbuf chan SocketMessage // send buffer
	ping    chan struct{}
	quit    chan struct{}
}

func (c *client) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	if c != nil {
		encoder.AddString("id", c.id)
		encoder.AddString("role", string(c.role))
		encoder.AddString("session", string(c.session))
		encoder.AddArray("pubTopics", c.pubTopics)
		encoder.AddArray("subTopics", c.subTopics)
	}
	return nil
}

func (c *client) read() {
	for {
		_, m, err := c.conn.ReadMessage()
		if err != nil {
			c.quit <- struct{}{}
			c.conn.Close()
			c.ws.unregister <- ClientUnregisterEvent{client: c, reason: err}
			return
		}

		message := SocketMessage{}
		if err := json.NewDecoder(bytes.NewReader(m)).Decode(&message); err != nil {
			log.Warn("[wsconn] received malformed text message", zap.Error(err), zap.String("raw", string(m)))
			continue
		}

		// Record the client role, this is a customized feature off the offical v1 spec,
		// Rabby dapp always sends `"role": "dapp"` in messages to relay server.
		c.role = RoleType(strings.ToLower(message.Role))

		message.client = c
		c.ws.localCh <- message
	}
}

func (c *client) write() {
	for {
		select {
		case message, more := <-c.sendbuf:
			if !more {
				return
			}
			m := new(bytes.Buffer)
			if err := json.NewEncoder(m).Encode(message); err != nil {
				log.Warn("sending malformed text message", zap.Error(err))
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, m.Bytes())
			if err != nil {
				log.Error("client write error", err, zap.Any("client", c), zap.Any("message", message))
			}
		case _, more := <-c.ping:
			if !more {
				return
			}
			if err := c.conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				log.Error("client ping error", err)
			}
		case <-c.quit:
			return
		}
	}
}

// send implements a non-blocking sending
func (c *client) send(message SocketMessage) {
	select {
	case c.sendbuf <- message:
	default:
		metrics.IncSendBlocking(len(c.sendbuf))
		log.Error("sending to client blocked", fmt.Errorf("sendbuf full"), zap.Any("client", c), zap.Any("len(sendbuf)", len(c.sendbuf)))
	}
}
