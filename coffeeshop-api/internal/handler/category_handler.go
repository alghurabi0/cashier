package handler

import (
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"coffeeshop-api/internal/sse"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// CategoryHandler handles HTTP requests for category operations.
type CategoryHandler struct {
	categoryService *service.CategoryService
	sseHub          *sse.Hub
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryService *service.CategoryService, sseHub *sse.Hub) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService, sseHub: sseHub}
}

// broadcastDataChanged sends a data_changed SSE event to all connected clients.
func (h *CategoryHandler) broadcastDataChanged(entity, action string) {
	if h.sseHub != nil {
		h.sseHub.Broadcast(sse.Event{
			Type: "data_changed",
			Data: map[string]string{"entity": entity, "action": action},
		})
	}
}

// List handles GET /api/v1/categories
// Supports optional ?since=<ISO8601> for delta sync.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid 'since' parameter: use ISO 8601 format")
			return
		}
		categories, err := h.categoryService.ListSince(tenantID, since)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to fetch categories")
			return
		}
		JSON(w, http.StatusOK, categories)
		return
	}

	categories, err := h.categoryService.List(tenantID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	JSON(w, http.StatusOK, categories)
}

// Get handles GET /api/v1/categories/{id}
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	category, err := h.categoryService.Get(tenantID, id)
	if err != nil {
		Error(w, http.StatusNotFound, "category not found")
		return
	}

	JSON(w, http.StatusOK, category)
}

// Create handles POST /api/v1/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req model.CreateCategoryRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	category, err := h.categoryService.Create(tenantID, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, category)
	h.broadcastDataChanged("category", "create")
}

// Update handles PUT /api/v1/categories/{id}
// Supports optimistic concurrency via X-Expected-Version header (RFC3339).
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	var req model.UpdateCategoryRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	// Check for optimistic concurrency header
	var category *model.Category
	versionHeader := r.Header.Get("X-Expected-Version")
	if versionHeader != "" {
		expectedVersion, parseErr := time.Parse(time.RFC3339Nano, versionHeader)
		if parseErr != nil {
			Error(w, http.StatusBadRequest, "invalid X-Expected-Version header, use RFC3339")
			return
		}
		category, err = h.categoryService.UpdateWithVersion(tenantID, id, req, expectedVersion)
	} else {
		category, err = h.categoryService.Update(tenantID, id, req)
	}

	if err != nil {
		if ce, ok := err.(*service.ConflictError); ok {
			JSON(w, http.StatusConflict, ce)
			return
		}
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	JSON(w, http.StatusOK, category)
	h.broadcastDataChanged("category", "update")
}

// Delete handles DELETE /api/v1/categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	if err := h.categoryService.Delete(tenantID, id); err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.broadcastDataChanged("category", "delete")
}
