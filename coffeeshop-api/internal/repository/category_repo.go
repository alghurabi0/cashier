package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"

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

// FindAll returns all active categories sorted by sort_order.
func (r *CategoryRepository) FindAll() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Select(&categories,
		`SELECT id, name_ar, sort_order, is_active
		 FROM categories
		 WHERE is_active = true
		 ORDER BY sort_order ASC, name_ar ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	return categories, nil
}

// FindByID returns a single category by ID.
func (r *CategoryRepository) FindByID(id uuid.UUID) (*model.Category, error) {
	var cat model.Category
	err := r.db.Get(&cat,
		`SELECT id, name_ar, sort_order, is_active
		 FROM categories
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return &cat, nil
}

// Create inserts a new category and returns it with the generated ID.
func (r *CategoryRepository) Create(nameAr string, sortOrder int) (*model.Category, error) {
	var cat model.Category
	err := r.db.Get(&cat,
		`INSERT INTO categories (name_ar, sort_order)
		 VALUES ($1, $2)
		 RETURNING id, name_ar, sort_order, is_active`,
		nameAr, sortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return &cat, nil
}

// Update modifies an existing category.
func (r *CategoryRepository) Update(id uuid.UUID, nameAr string, sortOrder int, isActive bool) (*model.Category, error) {
	var cat model.Category
	err := r.db.Get(&cat,
		`UPDATE categories
		 SET name_ar = $1, sort_order = $2, is_active = $3
		 WHERE id = $4
		 RETURNING id, name_ar, sort_order, is_active`,
		nameAr, sortOrder, isActive, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return &cat, nil
}

// SoftDelete sets is_active to false for the given category.
func (r *CategoryRepository) SoftDelete(id uuid.UUID) error {
	result, err := r.db.Exec(
		`UPDATE categories SET is_active = false WHERE id = $1`,
		id,
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
