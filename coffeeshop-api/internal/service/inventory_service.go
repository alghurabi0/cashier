package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

// InventoryService handles business logic for inventory items.
type InventoryService struct {
	inventoryRepo *repository.InventoryRepository
	recipeRepo    *repository.RecipeRepository
	menuItemRepo  *repository.MenuItemRepository
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(
	inventoryRepo *repository.InventoryRepository,
	recipeRepo *repository.RecipeRepository,
	menuItemRepo *repository.MenuItemRepository,
) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		recipeRepo:    recipeRepo,
		menuItemRepo:  menuItemRepo,
	}
}

// List returns all active inventory items.
func (s *InventoryService) List() ([]model.InventoryItem, error) {
	return s.inventoryRepo.FindAll()
}

// Get returns a single inventory item by ID.
func (s *InventoryService) Get(id uuid.UUID) (*model.InventoryItem, error) {
	return s.inventoryRepo.FindByID(id)
}

// Create validates and creates a new inventory item.
func (s *InventoryService) Create(req model.CreateInventoryItemRequest) (*model.InventoryItem, error) {
	errors := make(map[string]string)

	if req.NameAr == "" {
		errors["name_ar"] = "must not be empty"
	}
	if req.BaseUnitAr == "" {
		errors["base_unit_ar"] = "must not be empty"
	}
	if req.UnitCost < 0 {
		errors["unit_cost"] = "must be >= 0"
	}
	if req.StockQty < 0 {
		errors["stock_qty"] = "must be >= 0"
	}

	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	return s.inventoryRepo.Create(req.NameAr, req.BaseUnitAr, req.StockQty, req.LowStockThreshold, req.UnitCost)
}

// Update validates and updates an existing inventory item.
// If unit_cost changes, recalculates cached_auto_cost for all affected menu items.
func (s *InventoryService) Update(id uuid.UUID, req model.UpdateInventoryItemRequest) (*model.InventoryItem, error) {
	existing, err := s.inventoryRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found")
	}

	oldUnitCost := existing.UnitCost

	// Merge provided fields
	if req.NameAr != nil {
		existing.NameAr = *req.NameAr
	}
	if req.BaseUnitAr != nil {
		existing.BaseUnitAr = *req.BaseUnitAr
	}
	if req.LowStockThreshold != nil {
		existing.LowStockThreshold = *req.LowStockThreshold
	}
	if req.UnitCost != nil {
		existing.UnitCost = *req.UnitCost
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// Validate
	errors := make(map[string]string)
	if existing.NameAr == "" {
		errors["name_ar"] = "must not be empty"
	}
	if existing.BaseUnitAr == "" {
		errors["base_unit_ar"] = "must not be empty"
	}
	if existing.UnitCost < 0 {
		errors["unit_cost"] = "must be >= 0"
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	updated, err := s.inventoryRepo.Update(id, existing.NameAr, existing.BaseUnitAr, existing.LowStockThreshold, existing.UnitCost, existing.IsActive)
	if err != nil {
		return nil, err
	}

	// If unit_cost changed, recalculate auto-cost for all menu items using this inventory item
	if oldUnitCost != existing.UnitCost {
		if err := s.recalculateAffectedMenuItems(id); err != nil {
			// Log but don't fail the update
			fmt.Printf("warning: failed to recalculate auto-costs after unit_cost change: %v\n", err)
		}
	}

	return updated, nil
}

// Delete soft-deletes an inventory item.
func (s *InventoryService) Delete(id uuid.UUID) error {
	return s.inventoryRepo.SoftDelete(id)
}

// Adjust records a stock adjustment and atomically updates stock_qty.
func (s *InventoryService) Adjust(req model.CreateStockAdjustmentRequest) (*model.StockAdjustment, error) {
	errors := make(map[string]string)

	if req.InventoryItemID == uuid.Nil {
		errors["inventory_item_id"] = "must not be empty"
	}
	if req.Delta == 0 {
		errors["delta"] = "must not be zero"
	}

	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	// Verify inventory item exists
	_, err := s.inventoryRepo.FindByID(req.InventoryItemID)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found")
	}

	return s.inventoryRepo.AdjustStock(req.InventoryItemID, req.Delta, req.ReasonAr)
}

// recalculateAffectedMenuItems finds all menu items that use the given inventory item
// in their recipe and recalculates their cached_auto_cost.
func (s *InventoryService) recalculateAffectedMenuItems(inventoryItemID uuid.UUID) error {
	menuItemIDs, err := s.recipeRepo.FindMenuItemIDsByInventoryItem(inventoryItemID)
	if err != nil {
		return fmt.Errorf("failed to find affected menu items: %w", err)
	}

	for _, menuItemID := range menuItemIDs {
		cost, err := s.recipeRepo.CalculateAutoCost(menuItemID)
		if err != nil {
			return fmt.Errorf("failed to calculate auto cost for menu item %s: %w", menuItemID, err)
		}
		if err := s.menuItemRepo.UpdateCachedAutoCost(menuItemID, cost); err != nil {
			return fmt.Errorf("failed to update cached auto cost for menu item %s: %w", menuItemID, err)
		}
	}

	return nil
}
