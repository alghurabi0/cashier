package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"

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

// FindByMenuItemID returns all recipe ingredients for a menu item,
// joined with inventory item details.
func (r *RecipeRepository) FindByMenuItemID(menuItemID uuid.UUID) ([]model.RecipeIngredientWithDetails, error) {
	var ingredients []model.RecipeIngredientWithDetails
	err := r.db.Select(&ingredients,
		`SELECT ri.id, ri.menu_item_id, ri.inventory_item_id, ri.quantity,
		        ii.name_ar AS inventory_name_ar, ii.base_unit_ar, ii.unit_cost
		 FROM recipe_ingredients ri
		 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		 WHERE ri.menu_item_id = $1
		 ORDER BY ii.name_ar ASC`,
		menuItemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipe: %w", err)
	}
	return ingredients, nil
}

// SetRecipe replaces all recipe ingredients for a menu item in a transaction.
// Deletes existing ingredients, then inserts the new list.
func (r *RecipeRepository) SetRecipe(menuItemID uuid.UUID, ingredients []model.RecipeIngredientInput) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing recipe
	_, err = tx.Exec(
		`DELETE FROM recipe_ingredients WHERE menu_item_id = $1`,
		menuItemID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear existing recipe: %w", err)
	}

	// Insert new ingredients
	for _, ing := range ingredients {
		_, err = tx.Exec(
			`INSERT INTO recipe_ingredients (menu_item_id, inventory_item_id, quantity)
			 VALUES ($1, $2, $3)`,
			menuItemID, ing.InventoryItemID, ing.Quantity,
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
// Formula: SUM(recipe_ingredient.quantity × inventory_item.unit_cost)
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
	// If no recipe ingredients, cost is NULL → return 0
	if cost == nil {
		return 0, nil
	}
	return *cost, nil
}

// FindMenuItemIDsByInventoryItem returns all menu item IDs that have a recipe
// containing the given inventory item. Used for cascading auto-cost recalculation.
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
