package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"
	"time"

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

const menuItemSelectCols = `mi.id, mi.tenant_id, mi.category_id, mi.name_ar, mi.price,
		        mi.cost_calc_method, mi.manual_cost_price, mi.cached_auto_cost,
		        mi.image_path, mi.is_active, mi.updated_at,
		        c.name_ar AS category_name_ar`

// FindAll returns all active menu items for a tenant with their category name.
// Optionally filters by category_id.
func (r *MenuItemRepository) FindAll(tenantID uuid.UUID, categoryID *uuid.UUID) ([]model.MenuItemWithCategory, error) {
	var items []model.MenuItemWithCategory

	if categoryID != nil {
		err := r.db.Select(&items,
			`SELECT `+menuItemSelectCols+`
			 FROM menu_items mi
			 JOIN categories c ON c.id = mi.category_id
			 WHERE mi.tenant_id = $1 AND mi.is_active = true AND mi.category_id = $2
			 ORDER BY mi.name_ar ASC`,
			tenantID, *categoryID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch menu items: %w", err)
		}
		return items, nil
	}

	err := r.db.Select(&items,
		`SELECT `+menuItemSelectCols+`
		 FROM menu_items mi
		 JOIN categories c ON c.id = mi.category_id
		 WHERE mi.tenant_id = $1 AND mi.is_active = true
		 ORDER BY mi.name_ar ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch menu items: %w", err)
	}
	return items, nil
}

// FindAllSince returns all menu items (including inactive) modified since the given time.
func (r *MenuItemRepository) FindAllSince(tenantID uuid.UUID, since time.Time) ([]model.MenuItemWithCategory, error) {
	var items []model.MenuItemWithCategory
	err := r.db.Select(&items,
		`SELECT `+menuItemSelectCols+`
		 FROM menu_items mi
		 JOIN categories c ON c.id = mi.category_id
		 WHERE mi.tenant_id = $1 AND mi.updated_at > $2
		 ORDER BY mi.name_ar ASC`,
		tenantID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch menu items since %v: %w", since, err)
	}
	return items, nil
}

// FindByID returns a single menu item by ID (tenant-scoped).
func (r *MenuItemRepository) FindByID(tenantID uuid.UUID, id uuid.UUID) (*model.MenuItemWithCategory, error) {
	var item model.MenuItemWithCategory
	err := r.db.Get(&item,
		`SELECT `+menuItemSelectCols+`
		 FROM menu_items mi
		 JOIN categories c ON c.id = mi.category_id
		 WHERE mi.tenant_id = $1 AND mi.id = $2`,
		tenantID, id,
	)
	if err != nil {
		return nil, fmt.Errorf("menu item not found: %w", err)
	}
	return &item, nil
}

// Create inserts a new menu item and returns it.
func (r *MenuItemRepository) Create(tenantID uuid.UUID, item *model.MenuItem) (*model.MenuItem, error) {
	var created model.MenuItem
	if item.ID != uuid.Nil {
		err := r.db.Get(&created,
			`INSERT INTO menu_items (id, tenant_id, category_id, name_ar, price, cost_calc_method, manual_cost_price, image_path)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING id, tenant_id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active, updated_at`,
			item.ID, tenantID, item.CategoryID, item.NameAr, item.Price, item.CostCalcMethod, item.ManualCostPrice, item.ImagePath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create menu item: %w", err)
		}
	} else {
		err := r.db.Get(&created,
			`INSERT INTO menu_items (tenant_id, category_id, name_ar, price, cost_calc_method, manual_cost_price, image_path)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING id, tenant_id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active, updated_at`,
			tenantID, item.CategoryID, item.NameAr, item.Price, item.CostCalcMethod, item.ManualCostPrice, item.ImagePath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create menu item: %w", err)
		}
	}
	return &created, nil
}

// Update modifies an existing menu item (tenant-scoped).
func (r *MenuItemRepository) Update(tenantID uuid.UUID, item *model.MenuItem) (*model.MenuItem, error) {
	var updated model.MenuItem
	err := r.db.Get(&updated,
		`UPDATE menu_items
		 SET category_id = $1, name_ar = $2, price = $3, cost_calc_method = $4,
		     manual_cost_price = $5, image_path = $6, is_active = $7
		 WHERE tenant_id = $8 AND id = $9
		 RETURNING id, tenant_id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active, updated_at`,
		item.CategoryID, item.NameAr, item.Price, item.CostCalcMethod,
		item.ManualCostPrice, item.ImagePath, item.IsActive, tenantID, item.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update menu item: %w", err)
	}
	return &updated, nil
}

// UpdateWithVersion modifies a menu item only if its updated_at matches expectedVersion.
func (r *MenuItemRepository) UpdateWithVersion(tenantID uuid.UUID, item *model.MenuItem, expectedVersion time.Time) (*model.MenuItem, error) {
	var updated model.MenuItem
	err := r.db.Get(&updated,
		`UPDATE menu_items
		 SET category_id = $1, name_ar = $2, price = $3, cost_calc_method = $4,
		     manual_cost_price = $5, image_path = $6, is_active = $7
		 WHERE tenant_id = $8 AND id = $9 AND updated_at = $10
		 RETURNING id, tenant_id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active, updated_at`,
		item.CategoryID, item.NameAr, item.Price, item.CostCalcMethod,
		item.ManualCostPrice, item.ImagePath, item.IsActive, tenantID, item.ID, expectedVersion,
	)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// SoftDelete sets is_active to false for the given menu item (tenant-scoped).
func (r *MenuItemRepository) SoftDelete(tenantID uuid.UUID, id uuid.UUID) error {
	result, err := r.db.Exec(
		`UPDATE menu_items SET is_active = false WHERE tenant_id = $1 AND id = $2`,
		tenantID, id,
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
