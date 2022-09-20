package sse

import (
	"log"
	"net/http"
	"time"
)

// the amount of time to wait when pushing a message to
// a slow client or a client that closed after `range clients` started.
const patience = time.Second * 1

type Broker struct {
	// Event sent to clients
	notifier chan Event
	// New client connections
	newClients chan chan Event
	// Closed client connections
	closingClients chan chan Event
	// Client connections registry
	clients map[chan Event]bool
	// Shutdown Broker
	shutdown chan struct{}
}

// NewSSEBroker Creates a new broker with channels set-up
func NewSSEBroker() (broker *Broker) {
	// Instantiate a broker
	broker = &Broker{
		notifier:       make(chan Event, 1),
		newClients:     make(chan chan Event),
		closingClients: make(chan chan Event),
		clients:        make(map[chan Event]bool),
		shutdown:       make(chan struct{}),
	}

	// Start listening on channels
	go broker.listen()

	return
}

// ServeHTTP Serve HTTP requests as a keep-alive connection and respond events to connected client
func (b *Broker) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	// Make sure that the writer supports flushing.
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set necessary HTTP headers for SSE
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	// TODO: Setup CORS as options, this is not production ready
	writer.Header().Set("Access-Control-Allow-Origin", "*")

	writer.WriteHeader(http.StatusOK)

	flusher.Flush()

	// Add new client
	messageChan := b.addClient()

	// Remove this client when connection is over
	defer func() {
		b.removeClient(messageChan)
	}()

	// Listen to connection close and un-register messageChan
	requestDone := req.Context().Done()

	for {
		select {
		// Request connection is over, bail now
		case <-requestDone:
			return
		default:
			// When receiving a message, send to this client
			event := <-messageChan
			err := event.SendMessage(req, writer, flusher)
			if err != nil {
				// TODO: Setup logger
				log.Printf("Failed to send message : %v", err)
			}
		}
	}
}

// addClient Adds a new client (chan Event) to the Broker
func (b *Broker) addClient() chan Event {
	// Create the client chan
	messageChan := make(chan Event)
	b.newClients <- messageChan

	return messageChan
}

// removeClient Removes a client (chan Event) from the Broker
func (b *Broker) removeClient(channel chan Event) {
	b.closingClients <- channel
}

// listen on Broker channels for events or client (de)registration and broker shutdown
func (b *Broker) listen() {
	for {
		select {
		case s := <-b.newClients:
			// New client, adds to the client registry
			b.clients[s] = true
			log.Printf("Client added. %d registered clients", len(b.clients))
		case s := <-b.closingClients:
			// Client set to be removed, removes from the client registry
			delete(b.clients, s)
			log.Printf("Removed client. %d registered clients", len(b.clients))
		case event := <-b.notifier:
			// New event comes in, add it to every client's channel in the registry
			for clientMessageChan, _ := range b.clients {
				select {
				case clientMessageChan <- event:
				case <-time.After(patience):
					log.Printf("Skipping client.")
				}
			}
		case <-b.shutdown:
			// Close all Broker channels
			close(b.newClients)
			close(b.closingClients)
			close(b.notifier)
			close(b.shutdown)
		}
	}
}

// PublishEvent adds an event to the broker to be sent
func (b *Broker) PublishEvent(event Event) {
	b.notifier <- event
}
