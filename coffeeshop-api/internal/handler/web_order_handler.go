package handler

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"coffeeshop-api/internal/sse"
	"net/http"
)

// WebOrderHandler handles web menu order creation.
type WebOrderHandler struct {
	orderService *service.OrderService
	tableService *service.TableService
	sseHub       *sse.Hub
}

// NewWebOrderHandler creates a new WebOrderHandler.
func NewWebOrderHandler(orderService *service.OrderService, tableService *service.TableService, sseHub *sse.Hub) *WebOrderHandler {
	return &WebOrderHandler{
		orderService: orderService,
		tableService: tableService,
		sseHub:       sseHub,
	}
}

// Create handles POST /api/v1/web-orders?token={table_token}
// The table token resolves the tenant — no JWT auth needed.
func (h *WebOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		Error(w, http.StatusUnauthorized, "table token required")
		return
	}

	table, err := h.tableService.GetByToken(token)
	if err != nil {
		Error(w, http.StatusUnauthorized, "invalid table token")
		return
	}

	var req model.WebOrderRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	// Use the table's tenant_id to scope the order
	order, err := h.orderService.CreateWebOrder(table.TenantID, table.Number, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sseHub.Broadcast(sse.Event{
		Type: "new_order",
		Data: order,
	})

	JSON(w, http.StatusCreated, order)
}
