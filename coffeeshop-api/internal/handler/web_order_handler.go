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
// Validates the table token, creates an order, and pushes it via SSE.
func (h *WebOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Validate table token
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

	// Decode order request
	var req model.WebOrderRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	// Create order with server-side price resolution
	order, err := h.orderService.CreateWebOrder(table.Number, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Push to SSE subscribers (POS clients)
	h.sseHub.Broadcast(sse.Event{
		Type: "new_order",
		Data: order,
	})

	JSON(w, http.StatusCreated, order)
}
