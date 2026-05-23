package sync

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// SSEEvent represents a parsed server-sent event.
type SSEEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// SSEClient connects to the API's SSE endpoint for real-time order events.
type SSEClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	onEvent    func(SSEEvent) // callback for received events
}

// NewSSEClient creates a new SSE client.
func NewSSEClient(baseURL string, onEvent func(SSEEvent)) *SSEClient {
	return &SSEClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 0, // no timeout for SSE stream
		},
		onEvent: onEvent,
	}
}

// SetToken sets the auth token.
func (c *SSEClient) SetToken(token string) {
	c.token = token
}

// Connect starts the SSE connection with automatic reconnect.
// Blocks until context is cancelled.
func (c *SSEClient) Connect(ctx context.Context) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			slog.Info("sse-client: stopped")
			return
		default:
		}

		err := c.stream(ctx)
		if err != nil {
			slog.Warn("sse-client: connection error, reconnecting...", "error", err, "backoff", backoff)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		// Exponential backoff, max 30s
		backoff = backoff * 2
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
	}
}

func (c *SSEClient) stream(ctx context.Context) error {
	url := c.baseURL + "/api/v1/orders/stream"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE returned status %d", resp.StatusCode)
	}

	slog.Info("sse-client: connected to SSE stream")

	// Reset backoff on successful connection (caller handles this indirectly)
	scanner := bufio.NewScanner(resp.Body)
	var eventType string
	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event
			if eventType != "" && len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				event := SSEEvent{
					Type: eventType,
					Data: json.RawMessage(data),
				}
				if c.onEvent != nil {
					c.onEvent(event)
				}
			}
			eventType = ""
			dataLines = nil
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		}
		// Ignore comments (lines starting with :)
	}

	return scanner.Err()
}
