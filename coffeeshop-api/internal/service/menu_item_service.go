package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"

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

// List returns all active menu items, optionally filtered by category.
func (s *MenuItemService) List(categoryID *uuid.UUID) ([]model.MenuItemWithCategory, error) {
	return s.menuItemRepo.FindAll(categoryID)
}

// Get returns a single menu item by ID.
func (s *MenuItemService) Get(id uuid.UUID) (*model.MenuItemWithCategory, error) {
	return s.menuItemRepo.FindByID(id)
}

// Create validates and creates a new menu item.
func (s *MenuItemService) Create(req model.CreateMenuItemRequest) (*model.MenuItem, error) {
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

	// Validate cost_calc_method
	if req.CostCalcMethod == "" {
		req.CostCalcMethod = "auto"
	}
	if req.CostCalcMethod != "auto" && req.CostCalcMethod != "manual" {
		errors["cost_calc_method"] = "must be 'auto' or 'manual'"
	}

	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	// Verify category exists
	_, err := s.categoryRepo.FindByID(req.CategoryID)
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

	return s.menuItemRepo.Create(item)
}

// Update validates and updates an existing menu item.
func (s *MenuItemService) Update(id uuid.UUID, req model.UpdateMenuItemRequest) (*model.MenuItem, error) {
	// Fetch existing item
	existing, err := s.menuItemRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("menu item not found")
	}

	// Merge provided fields into the existing item
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

	// Validate merged result
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

	// If category changed, verify new category exists
	if req.CategoryID != nil {
		_, err := s.categoryRepo.FindByID(item.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("category not found")
		}
	}

	return s.menuItemRepo.Update(item)
}

// Delete soft-deletes a menu item.
func (s *MenuItemService) Delete(id uuid.UUID) error {
	return s.menuItemRepo.SoftDelete(id)
}
