package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// MenuItemRepository handles database operations for menu items.
type MenuItemRepository struct {
	db *sqlx.DB
}

// NewMenuItemRepository creates a new MenuItemRepository.
func NewMenuItemRepository(db *sqlx.DB) *MenuItemRepository {
	return &MenuItemRepository{db: db}
}

// FindAll returns all active menu items with their category name.
// Optionally filters by category_id.
func (r *MenuItemRepository) FindAll(categoryID *uuid.UUID) ([]model.MenuItemWithCategory, error) {
	var items []model.MenuItemWithCategory

	if categoryID != nil {
		err := r.db.Select(&items,
			`SELECT mi.id, mi.category_id, mi.name_ar, mi.price,
			        mi.cost_calc_method, mi.manual_cost_price, mi.cached_auto_cost,
			        mi.image_path, mi.is_active,
			        c.name_ar AS category_name_ar
			 FROM menu_items mi
			 JOIN categories c ON c.id = mi.category_id
			 WHERE mi.is_active = true AND mi.category_id = $1
			 ORDER BY mi.name_ar ASC`,
			*categoryID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch menu items: %w", err)
		}
		return items, nil
	}

	err := r.db.Select(&items,
		`SELECT mi.id, mi.category_id, mi.name_ar, mi.price,
		        mi.cost_calc_method, mi.manual_cost_price, mi.cached_auto_cost,
		        mi.image_path, mi.is_active,
		        c.name_ar AS category_name_ar
		 FROM menu_items mi
		 JOIN categories c ON c.id = mi.category_id
		 WHERE mi.is_active = true
		 ORDER BY mi.name_ar ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch menu items: %w", err)
	}
	return items, nil
}

// FindByID returns a single menu item by ID with its category name.
func (r *MenuItemRepository) FindByID(id uuid.UUID) (*model.MenuItemWithCategory, error) {
	var item model.MenuItemWithCategory
	err := r.db.Get(&item,
		`SELECT mi.id, mi.category_id, mi.name_ar, mi.price,
		        mi.cost_calc_method, mi.manual_cost_price, mi.cached_auto_cost,
		        mi.image_path, mi.is_active,
		        c.name_ar AS category_name_ar
		 FROM menu_items mi
		 JOIN categories c ON c.id = mi.category_id
		 WHERE mi.id = $1`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("menu item not found: %w", err)
	}
	return &item, nil
}

// Create inserts a new menu item and returns it.
func (r *MenuItemRepository) Create(item *model.MenuItem) (*model.MenuItem, error) {
	var created model.MenuItem
	err := r.db.Get(&created,
		`INSERT INTO menu_items (category_id, name_ar, price, cost_calc_method, manual_cost_price, image_path)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active`,
		item.CategoryID, item.NameAr, item.Price, item.CostCalcMethod, item.ManualCostPrice, item.ImagePath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu item: %w", err)
	}
	return &created, nil
}

// Update modifies an existing menu item.
func (r *MenuItemRepository) Update(item *model.MenuItem) (*model.MenuItem, error) {
	var updated model.MenuItem
	err := r.db.Get(&updated,
		`UPDATE menu_items
		 SET category_id = $1, name_ar = $2, price = $3, cost_calc_method = $4,
		     manual_cost_price = $5, image_path = $6, is_active = $7
		 WHERE id = $8
		 RETURNING id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active`,
		item.CategoryID, item.NameAr, item.Price, item.CostCalcMethod,
		item.ManualCostPrice, item.ImagePath, item.IsActive, item.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update menu item: %w", err)
	}
	return &updated, nil
}

// SoftDelete sets is_active to false for the given menu item.
func (r *MenuItemRepository) SoftDelete(id uuid.UUID) error {
	result, err := r.db.Exec(
		`UPDATE menu_items SET is_active = false WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete menu item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("menu item not found")
	}

	return nil
}

// UpdateCachedAutoCost updates the cached_auto_cost field on a menu item.
// Called after recipe changes or inventory unit_cost changes.
func (r *MenuItemRepository) UpdateCachedAutoCost(id uuid.UUID, cost int64) error {
	result, err := r.db.Exec(
		`UPDATE menu_items SET cached_auto_cost = $1 WHERE id = $2`,
		cost, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update cached auto cost: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("menu item not found")
	}
	return nil
}
