package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CategoryRepository handles database operations for categories.
type CategoryRepository struct {
	db *sqlx.DB
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FindAll returns all active categories for a tenant, sorted by sort_order.
func (r *CategoryRepository) FindAll(tenantID uuid.UUID) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Select(&categories,
		`SELECT id, tenant_id, name_ar, sort_order, is_active, updated_at
		 FROM categories
		 WHERE tenant_id = $1 AND is_active = true
		 ORDER BY sort_order ASC, name_ar ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	return categories, nil
}

// FindAllSince returns all categories (including inactive) modified since the given time.
func (r *CategoryRepository) FindAllSince(tenantID uuid.UUID, since time.Time) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Select(&categories,
		`SELECT id, tenant_id, name_ar, sort_order, is_active, updated_at
		 FROM categories
		 WHERE tenant_id = $1 AND updated_at > $2
		 ORDER BY sort_order ASC, name_ar ASC`,
		tenantID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories since %v: %w", since, err)
	}
	return categories, nil
}

// FindByID returns a single category by ID (tenant-scoped).
func (r *CategoryRepository) FindByID(tenantID uuid.UUID, id uuid.UUID) (*model.Category, error) {
	var cat model.Category
	err := r.db.Get(&cat,
		`SELECT id, tenant_id, name_ar, sort_order, is_active, updated_at
		 FROM categories
		 WHERE tenant_id = $1 AND id = $2`,
		tenantID, id,
	)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return &cat, nil
}

// Create inserts a new category and returns it with the generated ID.
func (r *CategoryRepository) Create(tenantID uuid.UUID, nameAr string, sortOrder int, clientID *uuid.UUID) (*model.Category, error) {
	var cat model.Category
	if clientID != nil {
		err := r.db.Get(&cat,
			`INSERT INTO categories (id, tenant_id, name_ar, sort_order)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, tenant_id, name_ar, sort_order, is_active, updated_at`,
			*clientID, tenantID, nameAr, sortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create category: %w", err)
		}
	} else {
		err := r.db.Get(&cat,
			`INSERT INTO categories (tenant_id, name_ar, sort_order)
			 VALUES ($1, $2, $3)
			 RETURNING id, tenant_id, name_ar, sort_order, is_active, updated_at`,
			tenantID, nameAr, sortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create category: %w", err)
		}
	}
	return &cat, nil
}

// Update modifies an existing category (tenant-scoped).
func (r *CategoryRepository) Update(tenantID uuid.UUID, id uuid.UUID, nameAr string, sortOrder int, isActive bool) (*model.Category, error) {
	var cat model.Category
	err := r.db.Get(&cat,
		`UPDATE categories
		 SET name_ar = $1, sort_order = $2, is_active = $3
		 WHERE tenant_id = $4 AND id = $5
		 RETURNING id, tenant_id, name_ar, sort_order, is_active, updated_at`,
		nameAr, sortOrder, isActive, tenantID, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return &cat, nil
}

// UpdateWithVersion modifies a category only if its updated_at matches expectedVersion.
// Returns sql.ErrNoRows if the version doesn't match (conflict detected).
func (r *CategoryRepository) UpdateWithVersion(tenantID uuid.UUID, id uuid.UUID, nameAr string, sortOrder int, isActive bool, expectedVersion time.Time) (*model.Category, error) {
	var cat model.Category
	err := r.db.Get(&cat,
		`UPDATE categories
		 SET name_ar = $1, sort_order = $2, is_active = $3
		 WHERE tenant_id = $4 AND id = $5 AND updated_at = $6
		 RETURNING id, tenant_id, name_ar, sort_order, is_active, updated_at`,
		nameAr, sortOrder, isActive, tenantID, id, expectedVersion,
	)
	if err != nil {
		return nil, err // sql.ErrNoRows means conflict
	}
	return &cat, nil
}

// SoftDelete sets is_active to false for the given category (tenant-scoped).
func (r *CategoryRepository) SoftDelete(tenantID uuid.UUID, id uuid.UUID) error {
	result, err := r.db.Exec(
		`UPDATE categories SET is_active = false WHERE tenant_id = $1 AND id = $2`,
		tenantID, id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
