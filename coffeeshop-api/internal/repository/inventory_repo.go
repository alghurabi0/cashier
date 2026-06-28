package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// InventoryRepository handles database operations for inventory items.
type InventoryRepository struct {
	db *sqlx.DB
}

// NewInventoryRepository creates a new InventoryRepository.
func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

const invSelectCols = `id, tenant_id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active, updated_at`

// FindAll returns all active inventory items for a tenant sorted by name.
func (r *InventoryRepository) FindAll(tenantID uuid.UUID) ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	err := r.db.Select(&items,
		`SELECT `+invSelectCols+`
		 FROM inventory_items
		 WHERE tenant_id = $1 AND is_active = true
		 ORDER BY name_ar ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory items: %w", err)
	}
	return items, nil
}

// FindAllSince returns all inventory items (including inactive) modified since the given time.
func (r *InventoryRepository) FindAllSince(tenantID uuid.UUID, since time.Time) ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	err := r.db.Select(&items,
		`SELECT `+invSelectCols+`
		 FROM inventory_items
		 WHERE tenant_id = $1 AND updated_at > $2
		 ORDER BY name_ar ASC`,
		tenantID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory items since %v: %w", since, err)
	}
	return items, nil
}

// FindByID returns a single inventory item by ID (tenant-scoped).
func (r *InventoryRepository) FindByID(tenantID uuid.UUID, id uuid.UUID) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.Get(&item,
		`SELECT `+invSelectCols+`
		 FROM inventory_items
		 WHERE tenant_id = $1 AND id = $2`,
		tenantID, id,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found: %w", err)
	}
	return &item, nil
}

// Create inserts a new inventory item under a tenant.
func (r *InventoryRepository) Create(tenantID uuid.UUID, nameAr, baseUnitAr string, stockQty, lowStockThreshold int, unitCost int64, clientID *uuid.UUID) (*model.InventoryItem, error) {
	var item model.InventoryItem
	if clientID != nil {
		err := r.db.Get(&item,
			`INSERT INTO inventory_items (id, tenant_id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING `+invSelectCols,
			*clientID, tenantID, nameAr, baseUnitAr, stockQty, lowStockThreshold, unitCost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create inventory item: %w", err)
		}
	} else {
		err := r.db.Get(&item,
			`INSERT INTO inventory_items (tenant_id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING `+invSelectCols,
			tenantID, nameAr, baseUnitAr, stockQty, lowStockThreshold, unitCost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create inventory item: %w", err)
		}
	}
	return &item, nil
}

// Update modifies an existing inventory item (excluding stock_qty, tenant-scoped).
func (r *InventoryRepository) Update(tenantID uuid.UUID, id uuid.UUID, nameAr, baseUnitAr string, lowStockThreshold int, unitCost int64, isActive bool) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.Get(&item,
		`UPDATE inventory_items
		 SET name_ar = $1, base_unit_ar = $2, low_stock_threshold = $3, unit_cost = $4, is_active = $5
		 WHERE tenant_id = $6 AND id = $7
		 RETURNING `+invSelectCols,
		nameAr, baseUnitAr, lowStockThreshold, unitCost, isActive, tenantID, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}
	return &item, nil
}

// UpdateWithVersion modifies an inventory item only if its updated_at matches expectedVersion.
func (r *InventoryRepository) UpdateWithVersion(tenantID uuid.UUID, id uuid.UUID, nameAr, baseUnitAr string, lowStockThreshold int, unitCost int64, isActive bool, expectedVersion time.Time) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.Get(&item,
		`UPDATE inventory_items
		 SET name_ar = $1, base_unit_ar = $2, low_stock_threshold = $3, unit_cost = $4, is_active = $5
		 WHERE tenant_id = $6 AND id = $7 AND updated_at = $8
		 RETURNING `+invSelectCols,
		nameAr, baseUnitAr, lowStockThreshold, unitCost, isActive, tenantID, id, expectedVersion,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// SoftDelete sets is_active to false for the given inventory item (tenant-scoped).
func (r *InventoryRepository) SoftDelete(tenantID uuid.UUID, id uuid.UUID) error {
	result, err := r.db.Exec(
		`UPDATE inventory_items SET is_active = false WHERE tenant_id = $1 AND id = $2`,
		tenantID, id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("inventory item not found")
	}
	return nil
}

// AdjustStock atomically inserts a stock adjustment record and updates the
// inventory item's stock_qty in a single transaction (tenant-scoped).
func (r *InventoryRepository) AdjustStock(tenantID uuid.UUID, inventoryItemID uuid.UUID, delta int, reasonAr string, clientID *uuid.UUID) (*model.StockAdjustment, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert adjustment record
	var adjustment model.StockAdjustment
	if clientID != nil {
		err = tx.Get(&adjustment,
			`INSERT INTO stock_adjustments (id, tenant_id, inventory_item_id, delta, reason_ar)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, tenant_id, inventory_item_id, delta, reason_ar, created_at`,
			*clientID, tenantID, inventoryItemID, delta, reasonAr,
		)
	} else {
		err = tx.Get(&adjustment,
			`INSERT INTO stock_adjustments (tenant_id, inventory_item_id, delta, reason_ar)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, tenant_id, inventory_item_id, delta, reason_ar, created_at`,
			tenantID, inventoryItemID, delta, reasonAr,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to insert stock adjustment: %w", err)
	}

	// Update stock_qty on the inventory item
	result, err := tx.Exec(
		`UPDATE inventory_items SET stock_qty = stock_qty + $1 WHERE tenant_id = $2 AND id = $3`,
		delta, tenantID, inventoryItemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update stock qty: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("inventory item not found")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &adjustment, nil
}
