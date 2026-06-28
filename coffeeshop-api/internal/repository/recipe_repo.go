package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// RecipeRepository handles database operations for recipe ingredients.
type RecipeRepository struct {
	db *sqlx.DB
}

// NewRecipeRepository creates a new RecipeRepository.
func NewRecipeRepository(db *sqlx.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

const recipeDetailCols = `ri.id, ri.tenant_id, ri.menu_item_id, ri.inventory_item_id, ri.quantity, ri.updated_at,
		        ii.name_ar AS inventory_name_ar, ii.base_unit_ar, ii.unit_cost`

// FindByMenuItemID returns all recipe ingredients for a menu item (tenant-scoped).
func (r *RecipeRepository) FindByMenuItemID(tenantID uuid.UUID, menuItemID uuid.UUID) ([]model.RecipeIngredientWithDetails, error) {
	var ingredients []model.RecipeIngredientWithDetails
	err := r.db.Select(&ingredients,
		`SELECT `+recipeDetailCols+`
		 FROM recipe_ingredients ri
		 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		 WHERE ri.tenant_id = $1 AND ri.menu_item_id = $2
		 ORDER BY ii.name_ar ASC`,
		tenantID, menuItemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipe: %w", err)
	}
	return ingredients, nil
}

// FindAllBulk returns all recipe ingredients for all active menu items (tenant-scoped).
func (r *RecipeRepository) FindAllBulk(tenantID uuid.UUID) ([]model.RecipeIngredientWithDetails, error) {
	var ingredients []model.RecipeIngredientWithDetails
	err := r.db.Select(&ingredients,
		`SELECT `+recipeDetailCols+`
		 FROM recipe_ingredients ri
		 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		 JOIN menu_items mi ON mi.id = ri.menu_item_id
		 WHERE ri.tenant_id = $1 AND mi.is_active = true
		 ORDER BY ri.menu_item_id, ii.name_ar ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all recipes: %w", err)
	}
	return ingredients, nil
}

// FindAllBulkSince returns recipe ingredients for menu items updated since the given time (tenant-scoped).
func (r *RecipeRepository) FindAllBulkSince(tenantID uuid.UUID, since time.Time) ([]model.RecipeIngredientWithDetails, error) {
	var ingredients []model.RecipeIngredientWithDetails
	err := r.db.Select(&ingredients,
		`SELECT `+recipeDetailCols+`
		 FROM recipe_ingredients ri
		 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		 JOIN menu_items mi ON mi.id = ri.menu_item_id
		 WHERE ri.tenant_id = $1 AND mi.updated_at > $2
		 ORDER BY ri.menu_item_id, ii.name_ar ASC`,
		tenantID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes since %v: %w", since, err)
	}
	return ingredients, nil
}

// SetRecipe replaces all recipe ingredients for a menu item (tenant-scoped).
func (r *RecipeRepository) SetRecipe(tenantID uuid.UUID, menuItemID uuid.UUID, ingredients []model.RecipeIngredientInput) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing recipe
	_, err = tx.Exec(
		`DELETE FROM recipe_ingredients WHERE tenant_id = $1 AND menu_item_id = $2`,
		tenantID, menuItemID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear existing recipe: %w", err)
	}

	// Insert new ingredients
	for _, ing := range ingredients {
		_, err = tx.Exec(
			`INSERT INTO recipe_ingredients (tenant_id, menu_item_id, inventory_item_id, quantity)
			 VALUES ($1, $2, $3, $4)`,
			tenantID, menuItemID, ing.InventoryItemID, ing.Quantity,
		)
		if err != nil {
			return fmt.Errorf("failed to insert recipe ingredient: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CalculateAutoCost computes the total auto-cost for a menu item's recipe.
func (r *RecipeRepository) CalculateAutoCost(menuItemID uuid.UUID) (int64, error) {
	var cost *int64
	err := r.db.Get(&cost,
		`SELECT SUM(ri.quantity::BIGINT * ii.unit_cost)
		 FROM recipe_ingredients ri
		 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		 WHERE ri.menu_item_id = $1`,
		menuItemID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate auto cost: %w", err)
	}
	if cost == nil {
		return 0, nil
	}
	return *cost, nil
}

// FindMenuItemIDsByInventoryItem returns all menu item IDs that use the given inventory item.
func (r *RecipeRepository) FindMenuItemIDsByInventoryItem(inventoryItemID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Select(&ids,
		`SELECT DISTINCT menu_item_id FROM recipe_ingredients WHERE inventory_item_id = $1`,
		inventoryItemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find menu items by inventory item: %w", err)
	}
	return ids, nil
}
