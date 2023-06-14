package relay

import (
	"testing"
	"time"
)

func TestInsertPendingSession(t *testing.T) {
	sps := NewSortedPendingSessions()

	session1 := &pendingSession{
		expireTime: time.Now().Add(5 * time.Second), // expire
		topic:      "topic1",
	}

	session2 := &pendingSession{
		expireTime: time.Now().Add(4 * time.Second), // expire
		topic:      "topic2",
	}

	session3 := &pendingSession{
		expireTime: time.Now().Add(3 * time.Second), // expire
		topic:      "topic3",
	}

	sps.insert(session1)
	sps.insert(session3)
	sps.insert(session2)

	actual := sps.peak()
	if actual.topic != session3.topic {
		t.Errorf("unexpected top session, expected: %v, actual: %v\n", session3.topic, actual.topic)
	}

	sps.pop()
	actual = sps.peak()
	if actual.topic != session2.topic {
		t.Errorf("unexpected top session, expected: %v, actual: %v\n", session2.topic, actual.topic)
	}

	sps.pop()
	actual = sps.peak()
	if actual.topic != session1.topic {
		t.Errorf("unexpected top session, expected: %v, actual: %v\n", session1.topic, actual.topic)
	}
}

func TestDeletePendingSession(t *testing.T) {
	sps := NewSortedPendingSessions()

	session1 := &pendingSession{
		expireTime: time.Now().Add(5 * time.Second), // expire
		topic:      "topic1",
	}

	session2 := &pendingSession{
		expireTime: time.Now().Add(4 * time.Second), // expire
		topic:      "topic2",
	}

	session3 := &pendingSession{
		expireTime: time.Now().Add(3 * time.Second), // expire
		topic:      "topic3",
	}

	sps.insert(session1)
	sps.insert(session3)
	sps.insert(session2)

	sps.deleteByTopic("topic3")

	actual := sps.peak()
	if actual.topic != session2.topic {
		t.Errorf("unexpected top session, expected: %v, actual: %v\n", session2.topic, actual.topic)
	}
}
