package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"database/sql"
	"fmt"
	"time"

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

// List returns all active inventory items for a tenant.
func (s *InventoryService) List(tenantID uuid.UUID) ([]model.InventoryItem, error) {
	return s.inventoryRepo.FindAll(tenantID)
}

// ListSince returns all inventory items (including inactive) modified since the given time.
func (s *InventoryService) ListSince(tenantID uuid.UUID, since time.Time) ([]model.InventoryItem, error) {
	return s.inventoryRepo.FindAllSince(tenantID, since)
}

// Get returns a single inventory item by ID (tenant-scoped).
func (s *InventoryService) Get(tenantID uuid.UUID, id uuid.UUID) (*model.InventoryItem, error) {
	return s.inventoryRepo.FindByID(tenantID, id)
}

// Create validates and creates a new inventory item under a tenant.
func (s *InventoryService) Create(tenantID uuid.UUID, req model.CreateInventoryItemRequest) (*model.InventoryItem, error) {
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

	return s.inventoryRepo.Create(tenantID, req.NameAr, req.BaseUnitAr, req.StockQty, req.LowStockThreshold, req.UnitCost, req.ID)
}

// Update validates and updates an existing inventory item (tenant-scoped).
func (s *InventoryService) Update(tenantID uuid.UUID, id uuid.UUID, req model.UpdateInventoryItemRequest) (*model.InventoryItem, error) {
	existing, err := s.inventoryRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found")
	}

	oldUnitCost := existing.UnitCost

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

	updated, err := s.inventoryRepo.Update(tenantID, id, existing.NameAr, existing.BaseUnitAr, existing.LowStockThreshold, existing.UnitCost, existing.IsActive)
	if err != nil {
		return nil, err
	}

	if oldUnitCost != existing.UnitCost {
		if err := s.recalculateAffectedMenuItems(id); err != nil {
			fmt.Printf("warning: failed to recalculate auto-costs after unit_cost change: %v\n", err)
		}
	}

	return updated, nil
}

// Delete soft-deletes an inventory item (tenant-scoped).
func (s *InventoryService) Delete(tenantID uuid.UUID, id uuid.UUID) error {
	return s.inventoryRepo.SoftDelete(tenantID, id)
}

// UpdateWithVersion validates and updates an inventory item with optimistic concurrency.
func (s *InventoryService) UpdateWithVersion(tenantID uuid.UUID, id uuid.UUID, req model.UpdateInventoryItemRequest, expectedVersion time.Time) (*model.InventoryItem, error) {
	existing, err := s.inventoryRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found")
	}

	oldUnitCost := existing.UnitCost

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

	result, err := s.inventoryRepo.UpdateWithVersion(tenantID, id, existing.NameAr, existing.BaseUnitAr, existing.LowStockThreshold, existing.UnitCost, existing.IsActive, expectedVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			current, fetchErr := s.inventoryRepo.FindByID(tenantID, id)
			serverVersion := time.Time{}
			if fetchErr == nil {
				serverVersion = current.UpdatedAt
			}
			return nil, &ConflictError{
				EntityType:    "inventory_item",
				EntityID:      id.String(),
				ServerVersion: serverVersion,
				ClientVersion: expectedVersion,
			}
		}
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}

	if oldUnitCost != existing.UnitCost {
		if err := s.recalculateAffectedMenuItems(id); err != nil {
			fmt.Printf("warning: failed to recalculate auto-costs after unit_cost change: %v\n", err)
		}
	}

	return result, nil
}

// Adjust records a stock adjustment and atomically updates stock_qty (tenant-scoped).
func (s *InventoryService) Adjust(tenantID uuid.UUID, req model.CreateStockAdjustmentRequest) (*model.StockAdjustment, error) {
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

	_, err := s.inventoryRepo.FindByID(tenantID, req.InventoryItemID)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found")
	}

	return s.inventoryRepo.AdjustStock(tenantID, req.InventoryItemID, req.Delta, req.ReasonAr, req.ID)
}

// recalculateAffectedMenuItems recalculates cached_auto_cost for menu items using this ingredient.
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
