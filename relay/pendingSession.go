package relay

import (
	"sort"
	"sync"
	"time"
)

type PendingSessions []*pendingSession

// SortedPendingSessions stores pending sessions in sorted manner by expireTime, meanwhile provides an O(1) time if lookup by topic
type SortedPendingSessions struct {
	sessions PendingSessions
	mapping  map[string]*pendingSession
	mutex    sync.Mutex
}

type pendingSession struct {
	expireTime time.Time
	topic      string
	dapp       *client
}

func (pq PendingSessions) Len() int { return len(pq) }

func (pq PendingSessions) Less(i, j int) bool {
	return pq[i].expireTime.Before(pq[j].expireTime)
}

func (pq PendingSessions) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PendingSessions) insert(p *pendingSession) {
	// i points to the smallest index that expires before p
	i := sort.Search(len(*pq), func(i int) bool {
		return (*pq)[i].expireTime.Before(p.expireTime)
	})
	*pq = append(*pq, &pendingSession{}) // allocate a empty value at the end
	copy((*pq)[i+1:], (*pq)[i:])
	(*pq)[i] = p
}

func (pq *PendingSessions) delete(p *pendingSession) {
	// i points to the smallest index that expires earlier or same as p
	i := sort.Search(len(*pq), func(i int) bool {
		return !(*pq)[i].expireTime.After(p.expireTime)
	})

	target := i
	for j, session := range (*pq)[i:] {
		if session.expireTime.Equal((*pq)[i].expireTime) && session.topic == (*pq)[i].topic {
			target = i + j
			break
		}
	}
	if target < len(*pq) {
		copy((*pq)[target:], (*pq)[target+1:])
		*pq = (*pq)[:len(*pq)-1]
	}
}

func (sps *SortedPendingSessions) insert(p *pendingSession) {
	sps.mutex.Lock()
	defer sps.mutex.Unlock()

	if _, ok := sps.mapping[p.topic]; ok {
		return
	}

	sps.sessions.insert(p)
	sps.mapping[p.topic] = p
}

// deleteByTopic deletes the pending session by topic
// the return indicates whether exists the pending session identified by the topic
func (sps *SortedPendingSessions) deleteByTopic(topic string) bool {
	sps.mutex.Lock()
	defer sps.mutex.Unlock()

	p, ok := sps.mapping[topic]
	if ok {
		sps.sessions.delete(p)
	}
	delete(sps.mapping, topic)

	return true
}

// peak peaks the session that will be expired earliest but ain't remove it
func (sps *SortedPendingSessions) peak() *pendingSession {
	if sps.sessions.Len() < 1 {
		return nil
	}
	return sps.sessions[sps.sessions.Len()-1]
}

// pop pops the session that will be expired earliest
func (sps *SortedPendingSessions) pop() {
	sps.mutex.Lock()
	defer sps.mutex.Unlock()

	session := sps.sessions[sps.sessions.Len()-1]
	sps.sessions = sps.sessions[:sps.sessions.Len()-1]
	delete(sps.mapping, session.topic)
}

func NewSortedPendingSessions() *SortedPendingSessions {
	return &SortedPendingSessions{
		sessions: PendingSessions{},
		mapping:  make(map[string]*pendingSession),
	}
}
