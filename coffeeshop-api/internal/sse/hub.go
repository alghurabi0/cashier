package sse

import (
	"encoding/json"
	"sync"
)

// Event represents a server-sent event.
type Event struct {
	Type string      `json:"type"` // "new_order", "order_status"
	Data interface{} `json:"data"`
}

// Hub manages SSE client connections and broadcasts events.
type Hub struct {
	clients map[chan Event]bool
	mu      sync.RWMutex
}

// NewHub creates a new SSE hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan Event]bool),
	}
}

// Subscribe registers a new client and returns its event channel.
func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 16) // buffered to avoid blocking broadcast
	h.mu.Lock()
	h.clients[ch] = true
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a client and closes its channel.
func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

// Broadcast sends an event to all connected clients.
func (h *Hub) Broadcast(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- event:
		default:
			// Client buffer full, skip (non-blocking)
		}
	}
}

// FormatSSE formats an event as an SSE-compliant string.
func FormatSSE(event Event) (string, error) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return "", err
	}
	return "event: " + event.Type + "\ndata: " + string(data) + "\n\n", nil
}
