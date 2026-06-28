package model

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a sales transaction.
type Order struct {
	ID            uuid.UUID  `db:"id"             json:"id"`
	TenantID      uuid.UUID  `db:"tenant_id"      json:"tenant_id"`
	DeviceID      *uuid.UUID `db:"device_id"      json:"device_id"`
	OrderNumber   string     `db:"order_number"    json:"order_number"`
	Source        string     `db:"source"          json:"source"`
	TableNumber   string     `db:"table_number"    json:"table_number"`
	Status        string     `db:"status"          json:"status"`
	Total         int64      `db:"total"           json:"total"`
	PaymentMethod string     `db:"payment_method"  json:"payment_method"`
	CreatedAt     time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"      json:"updated_at"`
}

// OrderItem represents a single line in an order.
type OrderItem struct {
	ID             uuid.UUID  `db:"id"               json:"id"`
	TenantID       uuid.UUID  `db:"tenant_id"        json:"tenant_id"`
	OrderID        uuid.UUID  `db:"order_id"         json:"order_id"`
	MenuItemID     *uuid.UUID `db:"menu_item_id"     json:"menu_item_id"`
	Quantity       int        `db:"quantity"          json:"quantity"`
	UnitPrice      int64      `db:"unit_price"       json:"unit_price"`
	LineTotal      int64      `db:"line_total"       json:"line_total"`
	NameArSnapshot string     `db:"name_ar_snapshot" json:"name_ar_snapshot"`
}

// OrderWithItems bundles an order with its line items.
type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

// CreateOrderRequest is the body for POST /api/v1/orders (pushed from POS).
type CreateOrderRequest struct {
	ID            uuid.UUID              `json:"id"`              // Client-generated UUID
	Source        string                 `json:"source"`          // "cashier"
	TableNumber   string                 `json:"table_number"`
	Total         int64                  `json:"total"`
	PaymentMethod string                 `json:"payment_method"`  // "cash"
	Items         []CreateOrderItemInput `json:"items"`
	CreatedAt     time.Time              `json:"created_at"`      // Client timestamp
}

// CreateOrderItemInput is a single line item in a CreateOrderRequest.
type CreateOrderItemInput struct {
	ID             uuid.UUID `json:"id"`
	MenuItemID     uuid.UUID `json:"menu_item_id"`
	Quantity       int       `json:"quantity"`
	UnitPrice      int64     `json:"unit_price"`
	LineTotal      int64     `json:"line_total"`
	NameArSnapshot string    `json:"name_ar_snapshot"`
}

// WebOrderRequest is the body for POST /api/v1/web-orders (from web menu).
// Prices are resolved server-side — customer only sends item IDs and quantities.
type WebOrderRequest struct {
	Items []WebOrderItemInput `json:"items"`
}

// WebOrderItemInput is a single line item in a WebOrderRequest.
type WebOrderItemInput struct {
	MenuItemID string `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
}

// UpdateOrderStatusRequest is the body for PUT /api/v1/orders/{id}/status.
type UpdateOrderStatusRequest struct {
	Status string `json:"status"` // "accepted", "rejected", "completed"
}


