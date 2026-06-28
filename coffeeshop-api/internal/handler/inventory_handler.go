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

// InventoryHandler handles HTTP requests for inventory operations.
type InventoryHandler struct {
	inventoryService *service.InventoryService
	sseHub           *sse.Hub
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(inventoryService *service.InventoryService, sseHub *sse.Hub) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService, sseHub: sseHub}
}

func (h *InventoryHandler) broadcastDataChanged(entity, action string) {
	if h.sseHub != nil {
		h.sseHub.Broadcast(sse.Event{
			Type: "data_changed",
			Data: map[string]string{"entity": entity, "action": action},
		})
	}
}

// List handles GET /api/v1/inventory
func (h *InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid 'since' parameter: use ISO 8601 format")
			return
		}
		items, err := h.inventoryService.ListSince(tenantID, since)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to fetch inventory items")
			return
		}
		JSON(w, http.StatusOK, items)
		return
	}

	items, err := h.inventoryService.List(tenantID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch inventory items")
		return
	}
	JSON(w, http.StatusOK, items)
}

// Get handles GET /api/v1/inventory/{id}
func (h *InventoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid inventory item ID")
		return
	}

	item, err := h.inventoryService.Get(tenantID, id)
	if err != nil {
		Error(w, http.StatusNotFound, "inventory item not found")
		return
	}
	JSON(w, http.StatusOK, item)
}

// Create handles POST /api/v1/inventory
func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req model.CreateInventoryItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	item, err := h.inventoryService.Create(tenantID, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, item)
	h.broadcastDataChanged("inventory_item", "create")
}

// Update handles PUT /api/v1/inventory/{id}
// Supports optimistic concurrency via X-Expected-Version header (RFC3339).
func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid inventory item ID")
		return
	}

	var req model.UpdateInventoryItemRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	var item *model.InventoryItem
	versionHeader := r.Header.Get("X-Expected-Version")
	if versionHeader != "" {
		expectedVersion, parseErr := time.Parse(time.RFC3339Nano, versionHeader)
		if parseErr != nil {
			Error(w, http.StatusBadRequest, "invalid X-Expected-Version header, use RFC3339")
			return
		}
		item, err = h.inventoryService.UpdateWithVersion(tenantID, id, req, expectedVersion)
	} else {
		item, err = h.inventoryService.Update(tenantID, id, req)
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
	h.broadcastDataChanged("inventory_item", "update")
}

// Delete handles DELETE /api/v1/inventory/{id}
func (h *InventoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid inventory item ID")
		return
	}

	if err := h.inventoryService.Delete(tenantID, id); err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.broadcastDataChanged("inventory_item", "delete")
}

// Adjust handles POST /api/v1/inventory/adjust
func (h *InventoryHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req model.CreateStockAdjustmentRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	adjustment, err := h.inventoryService.Adjust(tenantID, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusCreated, adjustment)
	h.broadcastDataChanged("inventory_item", "adjust")
}
