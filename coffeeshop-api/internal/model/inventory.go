package model

import (
	"time"

	"github.com/google/uuid"
)

// InventoryItem represents a back-of-house raw material (e.g. "حليب المراعي").
type InventoryItem struct {
	ID                uuid.UUID `db:"id"                  json:"id"`
	TenantID          uuid.UUID `db:"tenant_id"           json:"tenant_id"`
	NameAr            string    `db:"name_ar"              json:"name_ar"`
	BaseUnitAr        string    `db:"base_unit_ar"         json:"base_unit_ar"`
	StockQty          int       `db:"stock_qty"            json:"stock_qty"`
	LowStockThreshold int       `db:"low_stock_threshold"  json:"low_stock_threshold"`
	UnitCost          int64     `db:"unit_cost"            json:"unit_cost"`
	IsActive          bool      `db:"is_active"            json:"is_active"`
	UpdatedAt         time.Time `db:"updated_at"           json:"updated_at"`
}

// CreateInventoryItemRequest is the expected JSON body for creating an inventory item.
type CreateInventoryItemRequest struct {
	ID                *uuid.UUID `json:"id,omitempty"` // Optional: client-generated UUID
	NameAr            string     `json:"name_ar"`
	BaseUnitAr        string     `json:"base_unit_ar"`
	StockQty          int        `json:"stock_qty"`
	LowStockThreshold int        `json:"low_stock_threshold"`
	UnitCost          int64      `json:"unit_cost"`
}

// UpdateInventoryItemRequest is the expected JSON body for updating an inventory item.
// Note: StockQty is NOT updatable here — use POST /api/v1/inventory/adjust instead.
type UpdateInventoryItemRequest struct {
	NameAr            *string `json:"name_ar,omitempty"`
	BaseUnitAr        *string `json:"base_unit_ar,omitempty"`
	LowStockThreshold *int    `json:"low_stock_threshold,omitempty"`
	UnitCost          *int64  `json:"unit_cost,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}
