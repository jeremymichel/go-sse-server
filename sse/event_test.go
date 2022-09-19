package sse

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvent_ToBytes(t *testing.T) {
	event := Event{Topics: []string{"topic"}, Data: map[string]string{"foo": "bar"}}

	bytes, _ := event.ToBytes()

	expectedString := "{\"foo\":\"bar\"}"
	if string(bytes) != expectedString {
		t.Errorf("event data does not match expected, got %s", string(bytes))
	}
}

func TestEvent_HasMatchingTopic(t *testing.T) {
	event := Event{Topics: []string{"foo", "bar"}}

	topics := []string{"foo", "fizz"}

	if !event.HasMatchingTopic(topics) {
		t.Errorf("Failed to match topics")
	}

	topics = []string{"fizz", "buzz"}
	if event.HasMatchingTopic(topics) {
		t.Errorf("Failed to not match topics")
	}
}

func TestEvent_SendMessage(t *testing.T) {
	event := Event{Topics: []string{"foo"}, Data: map[string]interface{}{}}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/?topics=foo&topics=bar", nil)
	rw := httptest.NewRecorder()
	err := event.SendMessage(req, rw, rw)
	if err != nil {
		t.Errorf("Failed to send message, got error %v", err)
	}
}
