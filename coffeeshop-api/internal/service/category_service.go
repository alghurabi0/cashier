package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"database/sql"
	"fmt"
	"time"

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

// List returns all active categories for a tenant.
func (s *CategoryService) List(tenantID uuid.UUID) ([]model.Category, error) {
	return s.categoryRepo.FindAll(tenantID)
}

// ListSince returns all categories (including inactive) modified since the given time.
func (s *CategoryService) ListSince(tenantID uuid.UUID, since time.Time) ([]model.Category, error) {
	return s.categoryRepo.FindAllSince(tenantID, since)
}

// Get returns a single category by ID (tenant-scoped).
func (s *CategoryService) Get(tenantID uuid.UUID, id uuid.UUID) (*model.Category, error) {
	return s.categoryRepo.FindByID(tenantID, id)
}

// Create validates and creates a new category under a tenant.
func (s *CategoryService) Create(tenantID uuid.UUID, req model.CreateCategoryRequest) (*model.Category, error) {
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

	return s.categoryRepo.Create(tenantID, req.NameAr, req.SortOrder, req.ID)
}

// Update validates and updates an existing category (tenant-scoped).
func (s *CategoryService) Update(tenantID uuid.UUID, id uuid.UUID, req model.UpdateCategoryRequest) (*model.Category, error) {
	existing, err := s.categoryRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	if req.NameAr != nil {
		existing.NameAr = *req.NameAr
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

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

	return s.categoryRepo.Update(tenantID, id, existing.NameAr, existing.SortOrder, existing.IsActive)
}

// Delete soft-deletes a category (tenant-scoped).
func (s *CategoryService) Delete(tenantID uuid.UUID, id uuid.UUID) error {
	return s.categoryRepo.SoftDelete(tenantID, id)
}

// UpdateWithVersion validates and updates a category with optimistic concurrency.
// If the entity was modified since expectedVersion, returns ConflictError.
func (s *CategoryService) UpdateWithVersion(tenantID uuid.UUID, id uuid.UUID, req model.UpdateCategoryRequest, expectedVersion time.Time) (*model.Category, error) {
	existing, err := s.categoryRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	if req.NameAr != nil {
		existing.NameAr = *req.NameAr
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

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

	result, err := s.categoryRepo.UpdateWithVersion(tenantID, id, existing.NameAr, existing.SortOrder, existing.IsActive, expectedVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			// Conflict: entity was modified by another client
			current, fetchErr := s.categoryRepo.FindByID(tenantID, id)
			serverVersion := time.Time{}
			if fetchErr == nil {
				serverVersion = current.UpdatedAt
			}
			return nil, &ConflictError{
				EntityType:    "category",
				EntityID:      id.String(),
				ServerVersion: serverVersion,
				ClientVersion: expectedVersion,
			}
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return result, nil
}

// ValidationError contains field-level validation errors.
type ValidationError struct {
	Errors map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
