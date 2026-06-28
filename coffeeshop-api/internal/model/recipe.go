package model

import (
	"time"

	"github.com/google/uuid"
)

// RecipeIngredient links a MenuItem to an InventoryItem with a quantity.
type RecipeIngredient struct {
	ID              uuid.UUID `db:"id"                json:"id"`
	TenantID        uuid.UUID `db:"tenant_id"         json:"tenant_id"`
	MenuItemID      uuid.UUID `db:"menu_item_id"      json:"menu_item_id"`
	InventoryItemID uuid.UUID `db:"inventory_item_id" json:"inventory_item_id"`
	Quantity        int       `db:"quantity"           json:"quantity"`
	UpdatedAt       time.Time `db:"updated_at"         json:"updated_at"`
}

// RecipeIngredientWithDetails extends RecipeIngredient with inventory item info
// for rich GET responses.
type RecipeIngredientWithDetails struct {
	RecipeIngredient
	InventoryNameAr string `db:"inventory_name_ar" json:"inventory_name_ar"`
	BaseUnitAr      string `db:"base_unit_ar"      json:"base_unit_ar"`
	UnitCost        int64  `db:"unit_cost"          json:"unit_cost"`
}

// SetRecipeRequest is the body for PUT /api/v1/menu-items/:id/recipe.
// It replaces ALL existing ingredients with the provided list.
type SetRecipeRequest struct {
	Ingredients []RecipeIngredientInput `json:"ingredients"`
}

// RecipeIngredientInput is a single ingredient in a SetRecipeRequest.
type RecipeIngredientInput struct {
	InventoryItemID uuid.UUID `json:"inventory_item_id"`
	Quantity        int       `json:"quantity"`
}
