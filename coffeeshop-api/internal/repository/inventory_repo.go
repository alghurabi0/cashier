package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"

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

// FindAll returns all active inventory items sorted by name.
func (r *InventoryRepository) FindAll() ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	err := r.db.Select(&items,
		`SELECT id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active
		 FROM inventory_items
		 WHERE is_active = true
		 ORDER BY name_ar ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory items: %w", err)
	}
	return items, nil
}

// FindByID returns a single inventory item by ID.
func (r *InventoryRepository) FindByID(id uuid.UUID) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.Get(&item,
		`SELECT id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active
		 FROM inventory_items
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("inventory item not found: %w", err)
	}
	return &item, nil
}

// Create inserts a new inventory item and returns it.
func (r *InventoryRepository) Create(nameAr, baseUnitAr string, stockQty, lowStockThreshold int, unitCost int64) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.Get(&item,
		`INSERT INTO inventory_items (name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active`,
		nameAr, baseUnitAr, stockQty, lowStockThreshold, unitCost,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}
	return &item, nil
}

// Update modifies an existing inventory item (excluding stock_qty).
func (r *InventoryRepository) Update(id uuid.UUID, nameAr, baseUnitAr string, lowStockThreshold int, unitCost int64, isActive bool) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.Get(&item,
		`UPDATE inventory_items
		 SET name_ar = $1, base_unit_ar = $2, low_stock_threshold = $3, unit_cost = $4, is_active = $5
		 WHERE id = $6
		 RETURNING id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active`,
		nameAr, baseUnitAr, lowStockThreshold, unitCost, isActive, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}
	return &item, nil
}

// SoftDelete sets is_active to false for the given inventory item.
func (r *InventoryRepository) SoftDelete(id uuid.UUID) error {
	result, err := r.db.Exec(
		`UPDATE inventory_items SET is_active = false WHERE id = $1`,
		id,
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
// inventory item's stock_qty in a single transaction.
func (r *InventoryRepository) AdjustStock(inventoryItemID uuid.UUID, delta int, reasonAr string) (*model.StockAdjustment, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert adjustment record
	var adjustment model.StockAdjustment
	err = tx.Get(&adjustment,
		`INSERT INTO stock_adjustments (inventory_item_id, delta, reason_ar)
		 VALUES ($1, $2, $3)
		 RETURNING id, inventory_item_id, delta, reason_ar, created_at`,
		inventoryItemID, delta, reasonAr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert stock adjustment: %w", err)
	}

	// Update stock_qty on the inventory item
	result, err := tx.Exec(
		`UPDATE inventory_items SET stock_qty = stock_qty + $1 WHERE id = $2`,
		delta, inventoryItemID,
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
