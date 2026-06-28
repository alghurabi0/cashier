package model

import (
	"time"

	"github.com/google/uuid"
)

// StockAdjustment records inventory changes (deliveries, waste, corrections).
type StockAdjustment struct {
	ID              uuid.UUID `db:"id"                json:"id"`
	TenantID        uuid.UUID `db:"tenant_id"         json:"tenant_id"`
	InventoryItemID uuid.UUID `db:"inventory_item_id" json:"inventory_item_id"`
	Delta           int       `db:"delta"             json:"delta"`
	ReasonAr        string    `db:"reason_ar"         json:"reason_ar"`
	CreatedAt       time.Time `db:"created_at"        json:"created_at"`
}

// CreateStockAdjustmentRequest is the expected JSON body for recording a stock adjustment.
type CreateStockAdjustmentRequest struct {
	ID              *uuid.UUID `json:"id,omitempty"` // Optional: client-generated UUID
	InventoryItemID uuid.UUID  `json:"inventory_item_id"`
	Delta           int        `json:"delta"`     // +50 = received, -10 = waste
	ReasonAr        string     `json:"reason_ar"`
}
