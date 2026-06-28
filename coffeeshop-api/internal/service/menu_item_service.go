package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MenuItemService handles business logic for menu items.
type MenuItemService struct {
	menuItemRepo *repository.MenuItemRepository
	categoryRepo *repository.CategoryRepository
}

// NewMenuItemService creates a new MenuItemService.
func NewMenuItemService(menuItemRepo *repository.MenuItemRepository, categoryRepo *repository.CategoryRepository) *MenuItemService {
	return &MenuItemService{
		menuItemRepo: menuItemRepo,
		categoryRepo: categoryRepo,
	}
}

// List returns all active menu items for a tenant, optionally filtered by category.
func (s *MenuItemService) List(tenantID uuid.UUID, categoryID *uuid.UUID) ([]model.MenuItemWithCategory, error) {
	return s.menuItemRepo.FindAll(tenantID, categoryID)
}

// ListSince returns all menu items (including inactive) modified since the given time.
func (s *MenuItemService) ListSince(tenantID uuid.UUID, since time.Time) ([]model.MenuItemWithCategory, error) {
	return s.menuItemRepo.FindAllSince(tenantID, since)
}

// Get returns a single menu item by ID (tenant-scoped).
func (s *MenuItemService) Get(tenantID uuid.UUID, id uuid.UUID) (*model.MenuItemWithCategory, error) {
	return s.menuItemRepo.FindByID(tenantID, id)
}

// Create validates and creates a new menu item under a tenant.
func (s *MenuItemService) Create(tenantID uuid.UUID, req model.CreateMenuItemRequest) (*model.MenuItem, error) {
	errors := make(map[string]string)

	if req.NameAr == "" {
		errors["name_ar"] = "must not be empty"
	}
	if req.Price <= 0 {
		errors["price"] = "must be greater than 0"
	}
	if req.CategoryID == uuid.Nil {
		errors["category_id"] = "must not be empty"
	}

	if req.CostCalcMethod == "" {
		req.CostCalcMethod = "auto"
	}
	if req.CostCalcMethod != "auto" && req.CostCalcMethod != "manual" {
		errors["cost_calc_method"] = "must be 'auto' or 'manual'"
	}

	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	// Verify category exists within tenant
	_, err := s.categoryRepo.FindByID(tenantID, req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	item := &model.MenuItem{
		CategoryID:      req.CategoryID,
		NameAr:          req.NameAr,
		Price:           req.Price,
		CostCalcMethod:  req.CostCalcMethod,
		ManualCostPrice: req.ManualCostPrice,
		ImagePath:       req.ImagePath,
	}
	if req.ID != nil {
		item.ID = *req.ID
	}

	return s.menuItemRepo.Create(tenantID, item)
}

// Update validates and updates an existing menu item (tenant-scoped).
func (s *MenuItemService) Update(tenantID uuid.UUID, id uuid.UUID, req model.UpdateMenuItemRequest) (*model.MenuItem, error) {
	existing, err := s.menuItemRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("menu item not found")
	}

	item := &existing.MenuItem

	if req.CategoryID != nil {
		item.CategoryID = *req.CategoryID
	}
	if req.NameAr != nil {
		item.NameAr = *req.NameAr
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.CostCalcMethod != nil {
		item.CostCalcMethod = *req.CostCalcMethod
	}
	if req.ManualCostPrice != nil {
		item.ManualCostPrice = *req.ManualCostPrice
	}
	if req.ImagePath != nil {
		item.ImagePath = *req.ImagePath
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}

	errors := make(map[string]string)
	if item.NameAr == "" {
		errors["name_ar"] = "must not be empty"
	}
	if item.Price <= 0 {
		errors["price"] = "must be greater than 0"
	}
	if item.CostCalcMethod != "auto" && item.CostCalcMethod != "manual" {
		errors["cost_calc_method"] = "must be 'auto' or 'manual'"
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	if req.CategoryID != nil {
		_, err := s.categoryRepo.FindByID(tenantID, item.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("category not found")
		}
	}

	return s.menuItemRepo.Update(tenantID, item)
}

// Delete soft-deletes a menu item (tenant-scoped).
func (s *MenuItemService) Delete(tenantID uuid.UUID, id uuid.UUID) error {
	return s.menuItemRepo.SoftDelete(tenantID, id)
}

// UpdateWithVersion validates and updates a menu item with optimistic concurrency.
func (s *MenuItemService) UpdateWithVersion(tenantID uuid.UUID, id uuid.UUID, req model.UpdateMenuItemRequest, expectedVersion time.Time) (*model.MenuItem, error) {
	existing, err := s.menuItemRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("menu item not found")
	}

	item := &existing.MenuItem

	if req.CategoryID != nil {
		item.CategoryID = *req.CategoryID
	}
	if req.NameAr != nil {
		item.NameAr = *req.NameAr
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.CostCalcMethod != nil {
		item.CostCalcMethod = *req.CostCalcMethod
	}
	if req.ManualCostPrice != nil {
		item.ManualCostPrice = *req.ManualCostPrice
	}
	if req.ImagePath != nil {
		item.ImagePath = *req.ImagePath
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}

	errors := make(map[string]string)
	if item.NameAr == "" {
		errors["name_ar"] = "must not be empty"
	}
	if item.Price <= 0 {
		errors["price"] = "must be greater than 0"
	}
	if item.CostCalcMethod != "auto" && item.CostCalcMethod != "manual" {
		errors["cost_calc_method"] = "must be 'auto' or 'manual'"
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	if req.CategoryID != nil {
		_, err := s.categoryRepo.FindByID(tenantID, item.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("category not found")
		}
	}

	result, err := s.menuItemRepo.UpdateWithVersion(tenantID, item, expectedVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			current, fetchErr := s.menuItemRepo.FindByID(tenantID, id)
			serverVersion := time.Time{}
			if fetchErr == nil {
				serverVersion = current.UpdatedAt
			}
			return nil, &ConflictError{
				EntityType:    "menu_item",
				EntityID:      id.String(),
				ServerVersion: serverVersion,
				ClientVersion: expectedVersion,
			}
		}
		return nil, fmt.Errorf("failed to update menu item: %w", err)
	}
	return result, nil
}
