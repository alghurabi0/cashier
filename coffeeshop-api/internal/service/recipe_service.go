package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

// RecipeService handles business logic for recipes.
type RecipeService struct {
	recipeRepo   *repository.RecipeRepository
	menuItemRepo *repository.MenuItemRepository
	inventoryRepo *repository.InventoryRepository
}

// NewRecipeService creates a new RecipeService.
func NewRecipeService(
	recipeRepo *repository.RecipeRepository,
	menuItemRepo *repository.MenuItemRepository,
	inventoryRepo *repository.InventoryRepository,
) *RecipeService {
	return &RecipeService{
		recipeRepo:    recipeRepo,
		menuItemRepo:  menuItemRepo,
		inventoryRepo: inventoryRepo,
	}
}

// GetRecipe returns the recipe for a menu item with ingredient details.
func (s *RecipeService) GetRecipe(menuItemID uuid.UUID) ([]model.RecipeIngredientWithDetails, error) {
	// Verify menu item exists
	_, err := s.menuItemRepo.FindByID(menuItemID)
	if err != nil {
		return nil, fmt.Errorf("menu item not found")
	}

	return s.recipeRepo.FindByMenuItemID(menuItemID)
}

// SetRecipe validates and replaces the recipe for a menu item.
// After saving, recalculates and updates cached_auto_cost.
func (s *RecipeService) SetRecipe(menuItemID uuid.UUID, req model.SetRecipeRequest) ([]model.RecipeIngredientWithDetails, error) {
	// Verify menu item exists
	_, err := s.menuItemRepo.FindByID(menuItemID)
	if err != nil {
		return nil, fmt.Errorf("menu item not found")
	}

	// Validate ingredients
	errors := make(map[string]string)
	for i, ing := range req.Ingredients {
		if ing.InventoryItemID == uuid.Nil {
			errors[fmt.Sprintf("ingredients[%d].inventory_item_id", i)] = "must not be empty"
		}
		if ing.Quantity <= 0 {
			errors[fmt.Sprintf("ingredients[%d].quantity", i)] = "must be greater than 0"
		}
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	// Verify all inventory items exist
	for i, ing := range req.Ingredients {
		_, err := s.inventoryRepo.FindByID(ing.InventoryItemID)
		if err != nil {
			errors[fmt.Sprintf("ingredients[%d].inventory_item_id", i)] = "inventory item not found"
		}
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	// Save the recipe (delete-and-replace)
	if err := s.recipeRepo.SetRecipe(menuItemID, req.Ingredients); err != nil {
		return nil, fmt.Errorf("failed to save recipe: %w", err)
	}

	// Recalculate and update cached_auto_cost
	cost, err := s.recipeRepo.CalculateAutoCost(menuItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate auto cost: %w", err)
	}
	if err := s.menuItemRepo.UpdateCachedAutoCost(menuItemID, cost); err != nil {
		return nil, fmt.Errorf("failed to update cached auto cost: %w", err)
	}

	// Return the saved recipe with details
	return s.recipeRepo.FindByMenuItemID(menuItemID)
}
