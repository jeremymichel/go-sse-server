package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Event struct {
	Topics []string
	Data   interface{}
}

// ToBytes takes the event's data and turn it into a JSON Byte slice
func (e *Event) ToBytes() ([]byte, error) {
	eventBytes, err := json.Marshal(e.Data)

	if err != nil {
		return nil, err
	}

	return eventBytes, nil
}

// HasMatchingTopic checks given with event's finding a match if available
func (e *Event) HasMatchingTopic(topics []string) bool {
	for _, topic := range topics {
		for _, eventTopic := range e.Topics {
			if eventTopic == topic {
				return true
			}
		}
	}

	return false
}

// SendMessage When an event's topics matches that of the request, send the message as JSON to the client
func (e *Event) SendMessage(req *http.Request, writer http.ResponseWriter, flusher http.Flusher) error {
	if e.HasMatchingTopic(getTopicsFromRequest(req)) {
		dataMessage, err := e.ToBytes()
		if err != nil {
			return fmt.Errorf("failed to convert event data to JSON : %w", err)
		}

		// Write to the ResponseWriter using Server Sent Events Format
		_, err = fmt.Fprintf(writer, "data: %s\n\n", dataMessage)
		if err != nil {
			return fmt.Errorf("failed to write SSE data message : %w", err)
		}

		// Flush the data immediately instead of buffering it for later.
		flusher.Flush()
	}
	return nil
}

// getTopicsFromRequest retrives string slice of topics from the request query string
func getTopicsFromRequest(request *http.Request) []string {
	var topics []string
	topics, ok := request.URL.Query()["topics"]
	if !ok {
		topics = []string{""}
	}

	return topics
}
