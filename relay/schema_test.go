package relay

import "testing"

func TestTopicSetBasic(t *testing.T) {
	ts := NewTopicClientSet()

	c := &client{id: "1"}
	ts.Set("hello", c)
	ts.Set("hello", &client{id: "2"})
	ts.Set("hello", &client{id: "3"})

	actualLen := ts.Len("hello")
	expectedLen := 3
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}

	ts.Unset("hello", c)
	expectedLen = 2
	actualLen = ts.Len("hello")
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}
}

func TestTopicSetReference(t *testing.T) {
	ts := NewTopicClientSet()

	c := &client{id: "1"}
	ts.Set("hello", c)
	ts.Set("hello", &client{id: "2"})
	ts.Set("hello", &client{id: "3"})

	actualLen := ts.Len("hello")
	expectedLen := 3
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}

	newts := ts

	newts.Unset("hello", c)
	expectedLen = 2
	actualLen = ts.Len("hello")
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}
}

func TestGetTopicsByClientAndClear(t *testing.T) {
	ts := NewTopicClientSet()

	c := &client{id: "1"}
	ts.Set("hello", c)
	ts.Set("hello1", c)
	ts.Set("hello2", c)
	ts.Set("hello", &client{id: "2"})
	ts.Set("hello", &client{id: "3"})

	actualLen := ts.Len("hello")
	expectedLen := 3
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}

	topics := ts.GetTopicsByClient(c, false)
	actualLen = len(topics)
	expectedLen = 3
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}

	topics = ts.GetTopicsByClient(c, true)
	actualLen = len(topics)
	expectedLen = 3
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}

	topics = ts.GetTopicsByClient(c, true)
	actualLen = len(topics)
	expectedLen = 0
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}
}

func TestTopicGet(t *testing.T) {
	ts := NewTopicSet()

	ts.Set("hello")
	ts.Set("hello1")
	ts.Set("hello2")

	topics := ts.Get()
	actualLen := len(topics)
	expectedLen := 3
	if actualLen != expectedLen {
		t.Errorf("length error, expected: %v, actual: %v", expectedLen, actualLen)
	}
	if _, ok := topics["hello"]; !ok {
		t.Errorf("key does not exists")
	}
	if _, ok := topics["hello1"]; !ok {
		t.Errorf("key does not exists")
	}
	if _, ok := topics["hello2"]; !ok {
		t.Errorf("key does not exists")
	}
}
