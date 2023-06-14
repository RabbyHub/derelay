package relay

import (
	"encoding/json"
	"strings"

	"go.uber.org/zap/zapcore"
)

type MessageType string

const (
	Pub MessageType = "pub"
	Sub MessageType = "sub"
	Ack MessageType = "ack"

	Ping MessageType = "ping"
	Pong MessageType = "pong"
)

// websocket message
type SocketMessage struct {
	Topic   string      `json:"topic"`
	Type    MessageType `json:"type"` // pub, sub, ack
	Payload string      `json:"payload"`
	Role    string      `json:"role"`
	Phase   string      `json:"phase"`
	Silent  bool        `json:"silent"`

	client *client `json:"-"`
}

func (sm SocketMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(sm)
}

type RoleType string

const (
	Dapp   RoleType = "dapp"
	Wallet RoleType = "wallet"
	Relay  RoleType = "relay"
)

type PhaseType string

const (
	SessionRequest   PhaseType = "sessionRequest"
	SessionReceived  PhaseType = "sessionReceived"
	SessionExpired   PhaseType = "sessionExpired"
	SessionStart     PhaseType = "sessionStart"
	SessionSuspended PhaseType = "sessionSuspended"
	SessionResumed   PhaseType = "sessionResumed"
)

// redis key prefix
const (
	// redis message cache
	cachedMessagePrefix = "wc:relay:cache:pendingMessages:"

	// redis message channels
	messageChan    = "wc:relay:chan:messages:"
	dappNotifyChan = "wc:relay:chan:dappNotify:"
)

func messageChanKey(topic string) string {
	return messageChan + topic
}

func dappNotifyChanKey(topic string) string {
	return dappNotifyChan + topic
}

// fromDappNotifyChan checks whether the redis notify message is from the notfyDapp channel
func fromDappNotifyChan(channel string) bool {
	return strings.HasPrefix(channel, dappNotifyChan)
}

func cachedMessageKey(topic string) string {
	return cachedMessagePrefix + topic
}

// TopicClientSet stores topic -> clients relationship
type TopicClientSet map[string]map[*client]struct{}

func NewTopicClientSet() TopicClientSet {
	return make(map[string]map[*client]struct{})
}

func (ts TopicClientSet) Get(topic string) map[*client]struct{} {
	return ts[topic]
}

func (ts TopicClientSet) Set(topic string, c *client) {
	if _, ok := ts[topic]; !ok {
		ts[topic] = make(map[*client]struct{})
	}

	ts[topic][c] = struct{}{}
}

// GetTopicsByClient returns the topics associated with the specified client,
// meanwhile, remove the client from these topics if `clear` is true
// returns the topics the client has associated with
func (ts TopicClientSet) GetTopicsByClient(c *client, clear bool) []string {
	topics := []string{}
	for topic, set := range ts {
		if _, ok := set[c]; ok {
			topics = append(topics, topic)
		}
		if clear {
			delete(set, c)
		}
	}
	return topics
}

func (ts TopicClientSet) Unset(topic string, c *client) {
	delete(ts[topic], c)
}

func (ts TopicClientSet) Len(topic string) int {
	return len(ts[topic])
}

func (ts TopicClientSet) Clear(topic string) {
	delete(ts, topic)
}

type TopicSet map[string]struct{}

func NewTopicSet() TopicSet {
	return make(map[string]struct{})
}

func (tm TopicSet) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for topic := range tm {
		encoder.AppendString(topic)
	}
	return nil
}

type ClientUnregisterEvent struct {
	client *client
	reason error
}
