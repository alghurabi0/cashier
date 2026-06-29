package handler

import (
	"coffeeshop-api/internal/storage"
	"io"
	"net/http"
	"strings"
)

// ImageHandler serves images from R2 storage as a proxy.
type ImageHandler struct {
	storage *storage.R2Storage
}

// NewImageHandler creates a new ImageHandler.
func NewImageHandler(storage *storage.R2Storage) *ImageHandler {
	return &ImageHandler{storage: storage}
}

// Serve handles GET /api/v1/images/{path...}
func (h *ImageHandler) Serve(w http.ResponseWriter, r *http.Request) {
	if h.storage == nil || !h.storage.IsConfigured() {
		Error(w, http.StatusServiceUnavailable, "image storage is not configured")
		return
	}

	key := r.PathValue("path")
	if key == "" || strings.Contains(key, "..") {
		Error(w, http.StatusBadRequest, "invalid image path")
		return
	}

	body, contentType, err := h.storage.Get(r.Context(), key)
	if err != nil {
		Error(w, http.StatusNotFound, "image not found")
		return
	}
	defer body.Close()

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, body)
}
