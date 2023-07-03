package relay

import (
	"encoding/json"
	"strings"
	"sync"

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
type TopicClientSet struct {
	*sync.RWMutex
	Data map[string]map[*client]struct{}
}

func NewTopicClientSet() *TopicClientSet {
	return &TopicClientSet{
		RWMutex: &sync.RWMutex{},
		Data:    map[string]map[*client]struct{}{},
	}
}

func (ts *TopicClientSet) Get(topic string) map[*client]struct{} {
	ts.RLock()
	defer ts.RUnlock()
	return ts.Data[topic]
}

func (ts *TopicClientSet) Set(topic string, c *client) {
	ts.Lock()
	defer ts.Unlock()
	if _, ok := ts.Data[topic]; !ok {
		ts.Data[topic] = make(map[*client]struct{})
	}

	ts.Data[topic][c] = struct{}{}
}

// GetTopicsByClient returns the topics associated with the specified client,
// meanwhile, remove the client from these topics if `clear` is true
// returns the topics the client has associated with
func (ts *TopicClientSet) GetTopicsByClient(c *client, clear bool) []string {
	// Write lock in a read func because we may remove the client from the topics
	ts.Lock()
	defer ts.Unlock()
	topics := []string{}
	for topic, set := range ts.Data {
		if _, ok := set[c]; ok {
			topics = append(topics, topic)
		}
		if clear {
			delete(set, c)
		}
	}
	return topics
}

func (ts *TopicClientSet) Unset(topic string, c *client) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.Data[topic], c)
}

func (ts *TopicClientSet) Len(topic string) int {
	ts.RLock()
	defer ts.RUnlock()
	return len(ts.Data[topic])
}

func (ts *TopicClientSet) Clear(topic string) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.Data, topic)
}

type TopicSet struct {
	*sync.RWMutex
	Data map[string]struct{}
}

func NewTopicSet() *TopicSet {
	return &TopicSet{
		RWMutex: &sync.RWMutex{},
		Data:    map[string]struct{}{},
	}
}

func (tm TopicSet) Set(topic string) {
	tm.Lock()
	defer tm.Unlock()
	tm.Data[topic] = struct{}{}
}

func (tm TopicSet) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	tm.Lock()
	defer tm.Unlock()
	for topic := range tm.Data {
		encoder.AppendString(topic)
	}
	return nil
}

type ClientUnregisterEvent struct {
	client *client
	reason error
}
