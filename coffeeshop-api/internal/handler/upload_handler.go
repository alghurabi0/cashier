package handler

import (
	"coffeeshop-api/internal/storage"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// UploadHandler handles file upload requests.
type UploadHandler struct {
	storage *storage.R2Storage
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(storage *storage.R2Storage) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// Upload handles POST /api/v1/uploads
// Accepts multipart/form-data with a single file field named "file".
// Optional query param "folder" to set the R2 key prefix (default: "menu-items").
// Returns the public URL of the uploaded file.
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if h.storage == nil || !h.storage.IsConfigured() {
		Error(w, http.StatusServiceUnavailable, "file storage is not configured")
		return
	}

	allowedExts := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".mp4":  "video/mp4",
	}

	// 50MB limit for videos, 5MB for images
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		Error(w, http.StatusBadRequest, fmt.Sprintf("failed to read file: %v", err))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType, ok := allowedExts[ext]
	if !ok {
		Error(w, http.StatusBadRequest, "unsupported file type (allowed: jpg, png, webp, mp4)")
		return
	}

	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "menu-items"
	}

	key := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	publicURL, err := h.storage.Upload(r.Context(), key, file, contentType)
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("upload failed: %v", err))
		return
	}

	JSON(w, http.StatusCreated, map[string]string{
		"url": publicURL,
	})
}
