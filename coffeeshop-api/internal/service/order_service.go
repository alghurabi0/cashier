package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OrderService handles business logic for orders.
type OrderService struct {
	orderRepo *repository.OrderRepository
}

// NewOrderService creates a new OrderService.
func NewOrderService(orderRepo *repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

// Create validates and creates a new order (pushed from POS).
func (s *OrderService) Create(req model.CreateOrderRequest) (*model.OrderWithItems, error) {
	errors := make(map[string]string)

	if req.ID == uuid.Nil {
		errors["id"] = "must not be empty"
	}
	if len(req.Items) == 0 {
		errors["items"] = "must have at least one item"
	}
	if req.Total <= 0 {
		errors["total"] = "must be greater than 0"
	}
	if req.Source == "" {
		req.Source = "cashier"
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = "cash"
	}

	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	return s.orderRepo.Create(req)
}

// UpdateStatus updates an order's status with transition validation.
func (s *OrderService) UpdateStatus(id uuid.UUID, status string) (*model.Order, error) {
	validStatuses := map[string]bool{"accepted": true, "rejected": true, "completed": true}
	if !validStatuses[status] {
		return nil, &ValidationError{Errors: map[string]string{"status": "must be accepted, rejected, or completed"}}
	}
	return s.orderRepo.UpdateStatus(id, status)
}

// CreateWebOrder validates and creates an order from the web menu.
func (s *OrderService) CreateWebOrder(tableNumber string, req model.WebOrderRequest) (*model.OrderWithItems, error) {
	errors := make(map[string]string)
	if tableNumber == "" {
		errors["table"] = "table number is required"
	}
	if len(req.Items) == 0 {
		errors["items"] = "must have at least one item"
	}
	for i, item := range req.Items {
		if item.MenuItemID == "" {
			errors[fmt.Sprintf("items[%d].menu_item_id", i)] = "required"
		}
		if item.Quantity <= 0 {
			errors[fmt.Sprintf("items[%d].quantity", i)] = "must be > 0"
		}
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}
	return s.orderRepo.CreateWebOrder(tableNumber, req.Items)
}

// ListByDateRange returns orders in a date range. from/to should be "YYYY-MM-DD".
func (s *OrderService) ListByDateRange(from, to string) ([]model.OrderWithItems, error) {
	layout := "2006-01-02"
	fromTime, err := time.Parse(layout, from)
	if err != nil {
		return nil, &ValidationError{Errors: map[string]string{"from": "invalid date format, use YYYY-MM-DD"}}
	}
	toTime, err := time.Parse(layout, to)
	if err != nil {
		return nil, &ValidationError{Errors: map[string]string{"to": "invalid date format, use YYYY-MM-DD"}}
	}
	if fromTime.After(toTime) {
		return nil, &ValidationError{Errors: map[string]string{"from": "must be before or equal to 'to'"}}
	}
	return s.orderRepo.ListByDateRange(fromTime, toTime)
}

