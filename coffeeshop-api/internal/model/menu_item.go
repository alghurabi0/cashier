package model

import "github.com/google/uuid"

// MenuItem represents a front-of-house product (e.g. "لاتيه").
// Prices are stored in fils (IQD × 1000) to avoid floating-point errors.
type MenuItem struct {
	ID              uuid.UUID `db:"id"                json:"id"`
	CategoryID      uuid.UUID `db:"category_id"       json:"category_id"`
	NameAr          string    `db:"name_ar"            json:"name_ar"`
	Price           int64     `db:"price"              json:"price"`
	CostCalcMethod  string    `db:"cost_calc_method"   json:"cost_calc_method"`
	ManualCostPrice int64     `db:"manual_cost_price"  json:"manual_cost_price"`
	CachedAutoCost  int64     `db:"cached_auto_cost"   json:"cached_auto_cost"`
	ImagePath       string    `db:"image_path"         json:"image_path"`
	IsActive        bool      `db:"is_active"          json:"is_active"`
}

// MenuItemWithCategory extends MenuItem with the category name for list/detail responses.
type MenuItemWithCategory struct {
	MenuItem
	CategoryNameAr string `db:"category_name_ar" json:"category_name_ar"`
}

// CreateMenuItemRequest is the expected JSON body for creating a menu item.
type CreateMenuItemRequest struct {
	CategoryID      uuid.UUID `json:"category_id"`
	NameAr          string    `json:"name_ar"`
	Price           int64     `json:"price"`
	CostCalcMethod  string    `json:"cost_calc_method"`
	ManualCostPrice int64     `json:"manual_cost_price"`
	ImagePath       string    `json:"image_path"`
}

// UpdateMenuItemRequest is the expected JSON body for updating a menu item.
type UpdateMenuItemRequest struct {
	CategoryID      *uuid.UUID `json:"category_id,omitempty"`
	NameAr          *string    `json:"name_ar,omitempty"`
	Price           *int64     `json:"price,omitempty"`
	CostCalcMethod  *string    `json:"cost_calc_method,omitempty"`
	ManualCostPrice *int64     `json:"manual_cost_price,omitempty"`
	ImagePath       *string    `json:"image_path,omitempty"`
	IsActive        *bool      `json:"is_active,omitempty"`
}
