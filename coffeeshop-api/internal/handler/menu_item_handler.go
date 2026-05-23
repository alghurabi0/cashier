package handler

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"net/http"

	"github.com/google/uuid"
)

// MenuItemHandler handles HTTP requests for menu item operations.
type MenuItemHandler struct {
	menuItemService *service.MenuItemService
}

// NewMenuItemHandler creates a new MenuItemHandler.
func NewMenuItemHandler(menuItemService *service.MenuItemService) *MenuItemHandler {
	return &MenuItemHandler{menuItemService: menuItemService}
}

// List handles GET /api/v1/menu-items
// Supports optional ?category_id= query parameter.
func (h *MenuItemHandler) List(w http.ResponseWriter, r *http.Request) {
	var categoryID *uuid.UUID

	if cidStr := r.URL.Query().Get("category_id"); cidStr != "" {
		parsed, err := uuid.Parse(cidStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid category_id format")
			return
		}
		categoryID = &parsed
	}

	items, err := h.menuItemService.List(categoryID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch menu items")
		return
	}

	JSON(w, http.StatusOK, items)
}

// Get handles GET /api/v1/menu-items/{id}
func (h *MenuItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	item, err := h.menuItemService.Get(id)
	if err != nil {
		Error(w, http.StatusNotFound, "menu item not found")
		return
	}

	JSON(w, http.StatusOK, item)
}

// Create handles POST /api/v1/menu-items
func (h *MenuItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateMenuItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	item, err := h.menuItemService.Create(req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, item)
}

// Update handles PUT /api/v1/menu-items/{id}
func (h *MenuItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	var req model.UpdateMenuItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	item, err := h.menuItemService.Update(id, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	JSON(w, http.StatusOK, item)
}

// Delete handles DELETE /api/v1/menu-items/{id}
func (h *MenuItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	if err := h.menuItemService.Delete(id); err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
