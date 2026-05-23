package handler

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"coffeeshop-api/internal/sse"
	"net/http"

	"github.com/google/uuid"
)

// OrderHandler handles HTTP requests for order operations.
type OrderHandler struct {
	orderService *service.OrderService
	sseHub       *sse.Hub
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(orderService *service.OrderService, sseHub *sse.Hub) *OrderHandler {
	return &OrderHandler{orderService: orderService, sseHub: sseHub}
}

// Create handles POST /api/v1/orders
// Accepts a synced order from the POS (with client-generated UUID).
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	order, err := h.orderService.Create(req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, order)
}

// UpdateStatus handles PUT /api/v1/orders/{id}/status
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	var req model.UpdateOrderStatusRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	order, err := h.orderService.UpdateStatus(id, req.Status)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Broadcast status change to SSE subscribers
	if h.sseHub != nil {
		h.sseHub.Broadcast(sse.Event{
			Type: "order_status",
			Data: order,
		})
	}

	JSON(w, http.StatusOK, order)
}

// List handles GET /api/v1/orders?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		Error(w, http.StatusBadRequest, "from and to query parameters are required (YYYY-MM-DD)")
		return
	}

	orders, err := h.orderService.ListByDateRange(from, to)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if orders == nil {
		orders = []model.OrderWithItems{}
	}

	JSON(w, http.StatusOK, orders)
}

