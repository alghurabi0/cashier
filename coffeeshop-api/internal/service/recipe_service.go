package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RecipeService handles business logic for recipes.
type RecipeService struct {
	recipeRepo    *repository.RecipeRepository
	menuItemRepo  *repository.MenuItemRepository
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

// GetRecipe returns the recipe for a menu item (tenant-scoped).
func (s *RecipeService) GetRecipe(tenantID uuid.UUID, menuItemID uuid.UUID) ([]model.RecipeIngredientWithDetails, error) {
	_, err := s.menuItemRepo.FindByID(tenantID, menuItemID)
	if err != nil {
		return nil, fmt.Errorf("menu item not found")
	}

	return s.recipeRepo.FindByMenuItemID(tenantID, menuItemID)
}

// GetAllRecipes returns all recipe ingredients for all active menu items (tenant-scoped).
func (s *RecipeService) GetAllRecipes(tenantID uuid.UUID) ([]model.RecipeIngredientWithDetails, error) {
	return s.recipeRepo.FindAllBulk(tenantID)
}

// GetAllRecipesSince returns recipe ingredients for menu items updated since the given time.
func (s *RecipeService) GetAllRecipesSince(tenantID uuid.UUID, since time.Time) ([]model.RecipeIngredientWithDetails, error) {
	return s.recipeRepo.FindAllBulkSince(tenantID, since)
}

// SetRecipe validates and replaces the recipe for a menu item (tenant-scoped).
func (s *RecipeService) SetRecipe(tenantID uuid.UUID, menuItemID uuid.UUID, req model.SetRecipeRequest) ([]model.RecipeIngredientWithDetails, error) {
	_, err := s.menuItemRepo.FindByID(tenantID, menuItemID)
	if err != nil {
		return nil, fmt.Errorf("menu item not found")
	}

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

	for i, ing := range req.Ingredients {
		_, err := s.inventoryRepo.FindByID(tenantID, ing.InventoryItemID)
		if err != nil {
			errors[fmt.Sprintf("ingredients[%d].inventory_item_id", i)] = "inventory item not found"
		}
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	if err := s.recipeRepo.SetRecipe(tenantID, menuItemID, req.Ingredients); err != nil {
		return nil, fmt.Errorf("failed to save recipe: %w", err)
	}

	cost, err := s.recipeRepo.CalculateAutoCost(menuItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate auto cost: %w", err)
	}
	if err := s.menuItemRepo.UpdateCachedAutoCost(menuItemID, cost); err != nil {
		return nil, fmt.Errorf("failed to update cached auto cost: %w", err)
	}

	return s.recipeRepo.FindByMenuItemID(tenantID, menuItemID)
}
