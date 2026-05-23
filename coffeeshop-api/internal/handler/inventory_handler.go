package handler

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"net/http"

	"github.com/google/uuid"
)

// InventoryHandler handles HTTP requests for inventory operations.
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// List handles GET /api/v1/inventory
func (h *InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.inventoryService.List()
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch inventory items")
		return
	}
	JSON(w, http.StatusOK, items)
}

// Get handles GET /api/v1/inventory/{id}
func (h *InventoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid inventory item ID")
		return
	}

	item, err := h.inventoryService.Get(id)
	if err != nil {
		Error(w, http.StatusNotFound, "inventory item not found")
		return
	}
	JSON(w, http.StatusOK, item)
}

// Create handles POST /api/v1/inventory
func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInventoryItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	item, err := h.inventoryService.Create(req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, item)
}

// Update handles PUT /api/v1/inventory/{id}
func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid inventory item ID")
		return
	}

	var req model.UpdateInventoryItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	item, err := h.inventoryService.Update(id, req)
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

// Delete handles DELETE /api/v1/inventory/{id}
func (h *InventoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid inventory item ID")
		return
	}

	if err := h.inventoryService.Delete(id); err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Adjust handles POST /api/v1/inventory/adjust
func (h *InventoryHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	var req model.CreateStockAdjustmentRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	adjustment, err := h.inventoryService.Adjust(req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusCreated, adjustment)
}
