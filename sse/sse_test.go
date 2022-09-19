package sse

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAddNewClient(t *testing.T) {
	broker := NewSSEBroker()
	broker.addClient()
	if len(broker.clients) < 1 {
		t.Errorf("Failed to add new client")
	}
}

func TestRemoveClient(t *testing.T) {
	broker := NewSSEBroker()
	client := broker.addClient()
	broker.removeClient(client)
	//TODO: Figure out a better way to do this, this is garbage
	time.Sleep(1 * time.Millisecond)
	if len(broker.clients) > 0 {
		t.Errorf("Failed to remove client, still has %d client", len(broker.clients))
	}
}

func TestBroker_ServeHTTP(t *testing.T) {
	broker := NewSSEBroker()

	ctx, cancel := context.WithCancel(context.Background())

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost?topics=foo", nil)
	rw := httptest.NewRecorder()

	go func() {
		//TODO: Figure out a better way to do this, this is garbage
		time.Sleep(1 * time.Millisecond)
		broker.PublishEvent(Event{Topics: []string{"foo"}, Data: map[string]string{"hello": "world"}})
		cancel()
	}()
	broker.ServeHTTP(rw, req)

	body, _ := io.ReadAll(rw.Body)
	if string(body) != "data: {\"hello\":\"world\"}\n\n" {
		t.Errorf("Failed to send event data to client, sent: [%v]", string(body))
	}
}
