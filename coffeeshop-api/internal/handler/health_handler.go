package handler

import (
	"net/http"
	"time"
)

// HealthHandler handles the health check endpoint.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check handles GET /api/v1/health — returns server status and time.
// This endpoint is unauthenticated so POS can probe connectivity.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
