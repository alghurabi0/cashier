package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// OrderRepository handles database operations for orders.
type OrderRepository struct {
	db *sqlx.DB
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

const orderSelectCols = `id, tenant_id, device_id, order_number, source, table_number, status, total, payment_method, created_at, updated_at`
const orderItemSelectCols = `id, tenant_id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot`

// Create inserts an order and its items in a transaction (tenant-scoped).
// Assigns a server-side order_number and deducts stock via recipes.
func (r *OrderRepository) Create(tenantID uuid.UUID, deviceID *uuid.UUID, req model.CreateOrderRequest) (*model.OrderWithItems, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check for duplicate (idempotent push)
	var existingCount int
	if err := tx.Get(&existingCount, `SELECT COUNT(*) FROM orders WHERE id = $1`, req.ID); err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if existingCount > 0 {
		return r.FindByID(tenantID, req.ID)
	}

	// Generate server-side order number (tenant-scoped)
	orderNumber, err := r.generateOrderNumber(tx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	// Determine created_at
	createdAt := req.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	// Insert order
	var order model.Order
	err = tx.Get(&order,
		`INSERT INTO orders (id, tenant_id, device_id, order_number, source, table_number, status, total, payment_method, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'completed', $7, $8, $9)
		 RETURNING `+orderSelectCols,
		req.ID, tenantID, deviceID, orderNumber, req.Source, req.TableNumber, req.Total, req.PaymentMethod, createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	var items []model.OrderItem
	for _, itemReq := range req.Items {
		var item model.OrderItem
		menuItemID := itemReq.MenuItemID
		err = tx.Get(&item,
			`INSERT INTO order_items (id, tenant_id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING `+orderItemSelectCols,
			itemReq.ID, tenantID, req.ID, &menuItemID, itemReq.Quantity, itemReq.UnitPrice, itemReq.LineTotal, itemReq.NameArSnapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
		items = append(items, item)
	}

	// Deduct stock via recipe engine (best-effort)
	for _, itemReq := range req.Items {
		r.deductStockForItem(tx, itemReq.MenuItemID, itemReq.Quantity)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &model.OrderWithItems{Order: order, Items: items}, nil
}

// FindByID returns an order with its items (tenant-scoped).
func (r *OrderRepository) FindByID(tenantID uuid.UUID, id uuid.UUID) (*model.OrderWithItems, error) {
	var order model.Order
	err := r.db.Get(&order,
		`SELECT `+orderSelectCols+` FROM orders WHERE tenant_id = $1 AND id = $2`,
		tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	var items []model.OrderItem
	err = r.db.Select(&items,
		`SELECT `+orderItemSelectCols+`
		 FROM order_items WHERE order_id = $1 ORDER BY name_ar_snapshot ASC`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}

	return &model.OrderWithItems{Order: order, Items: items}, nil
}

// generateOrderNumber creates a sequential daily number: ORD-YYYYMMDD-NNN (tenant-scoped).
func (r *OrderRepository) generateOrderNumber(tx *sqlx.Tx, tenantID uuid.UUID) (string, error) {
	today := time.Now().Format("20060102")
	prefix := "ORD-" + today + "-"

	var count int
	err := tx.Get(&count,
		`SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND order_number LIKE $2`,
		tenantID, prefix+"%",
	)
	if err != nil {
		return "", fmt.Errorf("failed to count today's orders: %w", err)
	}

	return fmt.Sprintf("%s%03d", prefix, count+1), nil
}

// deductStockForItem looks up the recipe for a menu item and deducts stock.
func (r *OrderRepository) deductStockForItem(tx *sqlx.Tx, menuItemID uuid.UUID, orderQty int) {
	type recipeRow struct {
		InventoryItemID string `db:"inventory_item_id"`
		Quantity        int    `db:"quantity"`
	}

	var ingredients []recipeRow
	err := tx.Select(&ingredients,
		`SELECT inventory_item_id, quantity FROM recipe_ingredients WHERE menu_item_id = $1`,
		menuItemID,
	)
	if err != nil || len(ingredients) == 0 {
		return
	}

	for _, ing := range ingredients {
		deductAmount := orderQty * ing.Quantity
		tx.Exec(
			`UPDATE inventory_items SET stock_qty = stock_qty - $1 WHERE id = $2`,
			deductAmount, ing.InventoryItemID,
		)
	}
}

// UpdateStatus updates an order's status with transition validation (tenant-scoped).
func (r *OrderRepository) UpdateStatus(tenantID uuid.UUID, id uuid.UUID, newStatus string) (*model.Order, error) {
	var currentStatus string
	err := r.db.Get(&currentStatus,
		`SELECT status FROM orders WHERE tenant_id = $1 AND id = $2`,
		tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	valid := false
	switch currentStatus {
	case "pending":
		valid = newStatus == "accepted" || newStatus == "rejected"
	case "accepted":
		valid = newStatus == "completed"
	}
	if !valid {
		return nil, fmt.Errorf("invalid status transition: %s → %s", currentStatus, newStatus)
	}

	var order model.Order
	err = r.db.Get(&order,
		`UPDATE orders SET status = $1 WHERE tenant_id = $2 AND id = $3
		 RETURNING `+orderSelectCols,
		newStatus, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}
	return &order, nil
}

// CreateWebOrder creates an order from the web menu with server-side price resolution (tenant-scoped).
func (r *OrderRepository) CreateWebOrder(tenantID uuid.UUID, tableNumber string, items []model.WebOrderItemInput) (*model.OrderWithItems, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	orderNumber, err := r.generateOrderNumber(tx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	orderID := uuid.New()
	var totalAmount int64

	type resolvedItem struct {
		itemID         uuid.UUID
		menuItemID     uuid.UUID
		quantity       int
		price          int64
		lineTotal      int64
		nameArSnapshot string
	}
	var resolvedItems []resolvedItem

	for _, item := range items {
		menuItemID, err := uuid.Parse(item.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("invalid menu_item_id: %s", item.MenuItemID)
		}

		var menuItem struct {
			Price  int64  `db:"price"`
			NameAr string `db:"name_ar"`
		}
		err = tx.Get(&menuItem,
			`SELECT price, name_ar FROM menu_items WHERE tenant_id = $1 AND id = $2 AND is_active = true`,
			tenantID, menuItemID)
		if err != nil {
			return nil, fmt.Errorf("menu item not found: %s", item.MenuItemID)
		}

		lineTotal := menuItem.Price * int64(item.Quantity)
		totalAmount += lineTotal

		resolvedItems = append(resolvedItems, resolvedItem{
			itemID:         uuid.New(),
			menuItemID:     menuItemID,
			quantity:       item.Quantity,
			price:          menuItem.Price,
			lineTotal:      lineTotal,
			nameArSnapshot: menuItem.NameAr,
		})
	}

	var order model.Order
	err = tx.Get(&order,
		`INSERT INTO orders (id, tenant_id, order_number, source, table_number, status, total, payment_method, created_at)
		 VALUES ($1, $2, $3, 'web_menu', $4, 'pending', $5, 'web', now())
		 RETURNING `+orderSelectCols,
		orderID, tenantID, orderNumber, tableNumber, totalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	var orderItems []model.OrderItem
	for _, ri := range resolvedItems {
		var oi model.OrderItem
		mID := ri.menuItemID
		err = tx.Get(&oi,
			`INSERT INTO order_items (id, tenant_id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING `+orderItemSelectCols,
			ri.itemID, tenantID, order.ID, &mID, ri.quantity, ri.price, ri.lineTotal, ri.nameArSnapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
		orderItems = append(orderItems, oi)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &model.OrderWithItems{Order: order, Items: orderItems}, nil
}

// ListByDateRange returns all orders between from and to dates (tenant-scoped).
func (r *OrderRepository) ListByDateRange(tenantID uuid.UUID, from, to time.Time) ([]model.OrderWithItems, error) {
	var orders []model.Order
	err := r.db.Select(&orders,
		`SELECT `+orderSelectCols+`
		 FROM orders
		 WHERE tenant_id = $1 AND created_at >= $2 AND created_at < $3
		 ORDER BY created_at DESC`,
		tenantID, from, to.Add(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	var result []model.OrderWithItems
	for _, order := range orders {
		var items []model.OrderItem
		err := r.db.Select(&items,
			`SELECT `+orderItemSelectCols+`
			 FROM order_items WHERE order_id = $1 ORDER BY name_ar_snapshot ASC`,
			order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch order items: %w", err)
		}
		result = append(result, model.OrderWithItems{Order: order, Items: items})
	}

	return result, nil
}

// FindAllSince returns all orders (with items) modified since the given time (tenant-scoped).
// Used for delta-sync by POS terminals to pull orders from other devices.
func (r *OrderRepository) FindAllSince(tenantID uuid.UUID, since time.Time) ([]model.OrderWithItems, error) {
	var orders []model.Order
	err := r.db.Select(&orders,
		`SELECT `+orderSelectCols+`
		 FROM orders
		 WHERE tenant_id = $1 AND updated_at > $2
		 ORDER BY updated_at ASC`,
		tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders since: %w", err)
	}

	var result []model.OrderWithItems
	for _, order := range orders {
		var items []model.OrderItem
		err := r.db.Select(&items,
			`SELECT `+orderItemSelectCols+`
			 FROM order_items WHERE order_id = $1 ORDER BY name_ar_snapshot ASC`,
			order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch order items: %w", err)
		}
		result = append(result, model.OrderWithItems{Order: order, Items: items})
	}

	return result, nil
}
