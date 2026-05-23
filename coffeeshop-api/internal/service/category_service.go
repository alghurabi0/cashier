package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

// CategoryService handles business logic for categories.
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// List returns all active categories.
func (s *CategoryService) List() ([]model.Category, error) {
	return s.categoryRepo.FindAll()
}

// Get returns a single category by ID.
func (s *CategoryService) Get(id uuid.UUID) (*model.Category, error) {
	return s.categoryRepo.FindByID(id)
}

// Create validates and creates a new category.
func (s *CategoryService) Create(req model.CreateCategoryRequest) (*model.Category, error) {
	errors := make(map[string]string)

	if req.NameAr == "" {
		errors["name_ar"] = "must not be empty"
	}
	if req.SortOrder < 0 {
		errors["sort_order"] = "must be >= 0"
	}

	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	return s.categoryRepo.Create(req.NameAr, req.SortOrder)
}

// Update validates and updates an existing category.
// Uses pointer fields to distinguish between "not provided" and "zero value".
func (s *CategoryService) Update(id uuid.UUID, req model.UpdateCategoryRequest) (*model.Category, error) {
	// Fetch existing
	existing, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	// Merge provided fields
	if req.NameAr != nil {
		existing.NameAr = *req.NameAr
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// Validate merged result
	errors := make(map[string]string)
	if existing.NameAr == "" {
		errors["name_ar"] = "must not be empty"
	}
	if existing.SortOrder < 0 {
		errors["sort_order"] = "must be >= 0"
	}
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	return s.categoryRepo.Update(id, existing.NameAr, existing.SortOrder, existing.IsActive)
}

// Delete soft-deletes a category.
func (s *CategoryService) Delete(id uuid.UUID) error {
	return s.categoryRepo.SoftDelete(id)
}

// ValidationError contains field-level validation errors.
type ValidationError struct {
	Errors map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
