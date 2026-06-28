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

// MenuItemHandler handles HTTP requests for menu item operations.
type MenuItemHandler struct {
	menuItemService *service.MenuItemService
	sseHub          *sse.Hub
}

// NewMenuItemHandler creates a new MenuItemHandler.
func NewMenuItemHandler(menuItemService *service.MenuItemService, sseHub *sse.Hub) *MenuItemHandler {
	return &MenuItemHandler{menuItemService: menuItemService, sseHub: sseHub}
}

func (h *MenuItemHandler) broadcastDataChanged(entity, action string) {
	if h.sseHub != nil {
		h.sseHub.Broadcast(sse.Event{
			Type: "data_changed",
			Data: map[string]string{"entity": entity, "action": action},
		})
	}
}

// List handles GET /api/v1/menu-items
func (h *MenuItemHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid 'since' parameter: use ISO 8601 format")
			return
		}
		items, err := h.menuItemService.ListSince(tenantID, since)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to fetch menu items")
			return
		}
		JSON(w, http.StatusOK, items)
		return
	}

	var categoryID *uuid.UUID
	if cidStr := r.URL.Query().Get("category_id"); cidStr != "" {
		parsed, err := uuid.Parse(cidStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid category_id format")
			return
		}
		categoryID = &parsed
	}

	items, err := h.menuItemService.List(tenantID, categoryID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch menu items")
		return
	}

	JSON(w, http.StatusOK, items)
}

// Get handles GET /api/v1/menu-items/{id}
func (h *MenuItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	item, err := h.menuItemService.Get(tenantID, id)
	if err != nil {
		Error(w, http.StatusNotFound, "menu item not found")
		return
	}

	JSON(w, http.StatusOK, item)
}

// Create handles POST /api/v1/menu-items
func (h *MenuItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req model.CreateMenuItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	item, err := h.menuItemService.Create(tenantID, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, item)
	h.broadcastDataChanged("menu_item", "create")
}

// Update handles PUT /api/v1/menu-items/{id}
// Supports optimistic concurrency via X-Expected-Version header (RFC3339).
func (h *MenuItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	var req model.UpdateMenuItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	var item *model.MenuItem
	versionHeader := r.Header.Get("X-Expected-Version")
	if versionHeader != "" {
		expectedVersion, parseErr := time.Parse(time.RFC3339Nano, versionHeader)
		if parseErr != nil {
			Error(w, http.StatusBadRequest, "invalid X-Expected-Version header, use RFC3339")
			return
		}
		item, err = h.menuItemService.UpdateWithVersion(tenantID, id, req, expectedVersion)
	} else {
		item, err = h.menuItemService.Update(tenantID, id, req)
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

	JSON(w, http.StatusOK, item)
	h.broadcastDataChanged("menu_item", "update")
}

// Delete handles DELETE /api/v1/menu-items/{id}
func (h *MenuItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	if err := h.menuItemService.Delete(tenantID, id); err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.broadcastDataChanged("menu_item", "delete")
}
