package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"coffeeshop-api/internal/sse"
)

// SSEHandler serves the SSE stream for real-time order events.
type SSEHandler struct {
	hub *sse.Hub
}

// NewSSEHandler creates a new SSEHandler.
func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

// Stream handles GET /api/v1/orders/stream
func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	// Check for streaming support
	flusher, ok := w.(http.Flusher)
	if !ok {
		Error(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Subscribe to events
	ch := h.hub.Subscribe()
	defer h.hub.Unsubscribe(ch)

	slog.Info("sse: client connected")

	// Send initial keepalive
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Keepalive ticker to prevent timeouts
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Stream events until client disconnects
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			slog.Info("sse: client disconnected")
			return
		case <-ticker.C:
			// Send empty comment line as keepalive
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		case event := <-ch:
			formatted, err := sse.FormatSSE(event)
			if err != nil {
				slog.Warn("sse: failed to format event", "error", err)
				continue
			}
			fmt.Fprint(w, formatted)
			flusher.Flush()
		}
	}
}
