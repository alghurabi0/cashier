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

// Create inserts an order and its items in a transaction.
// Assigns a server-side order_number (ORD-YYYYMMDD-NNN) and deducts stock via recipes.
func (r *OrderRepository) Create(req model.CreateOrderRequest) (*model.OrderWithItems, error) {
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
		// Already exists — return the existing order
		return r.FindByID(req.ID)
	}

	// Generate server-side order number: ORD-YYYYMMDD-NNN
	orderNumber, err := r.generateOrderNumber(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	// Determine created_at: use client timestamp if provided, else server time
	createdAt := req.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	// Insert order
	var order model.Order
	err = tx.Get(&order,
		`INSERT INTO orders (id, order_number, source, table_number, status, total, payment_method, created_at)
		 VALUES ($1, $2, $3, $4, 'completed', $5, $6, $7)
		 RETURNING id, order_number, source, table_number, status, total, payment_method, created_at`,
		req.ID, orderNumber, req.Source, req.TableNumber, req.Total, req.PaymentMethod, createdAt,
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
			`INSERT INTO order_items (id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot`,
			itemReq.ID, req.ID, &menuItemID, itemReq.Quantity, itemReq.UnitPrice, itemReq.LineTotal, itemReq.NameArSnapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
		items = append(items, item)
	}

	// Deduct stock via recipe engine (best-effort — don't fail the order)
	for _, itemReq := range req.Items {
		r.deductStockForItem(tx, itemReq.MenuItemID, itemReq.Quantity)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &model.OrderWithItems{Order: order, Items: items}, nil
}

// FindByID returns an order with its items.
func (r *OrderRepository) FindByID(id uuid.UUID) (*model.OrderWithItems, error) {
	var order model.Order
	err := r.db.Get(&order,
		`SELECT id, order_number, source, table_number, status, total, payment_method, created_at
		 FROM orders WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	var items []model.OrderItem
	err = r.db.Select(&items,
		`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot
		 FROM order_items WHERE order_id = $1 ORDER BY name_ar_snapshot ASC`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}

	return &model.OrderWithItems{Order: order, Items: items}, nil
}

// generateOrderNumber creates a sequential daily number: ORD-YYYYMMDD-NNN
func (r *OrderRepository) generateOrderNumber(tx *sqlx.Tx) (string, error) {
	today := time.Now().Format("20060102")
	prefix := "ORD-" + today + "-"

	var count int
	err := tx.Get(&count,
		`SELECT COUNT(*) FROM orders WHERE order_number LIKE $1`,
		prefix+"%",
	)
	if err != nil {
		return "", fmt.Errorf("failed to count today's orders: %w", err)
	}

	return fmt.Sprintf("%s%03d", prefix, count+1), nil
}

// deductStockForItem looks up the recipe for a menu item and deducts stock.
// Best-effort: logs errors but doesn't fail the transaction.
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
		return // No recipe = no deduction
	}

	for _, ing := range ingredients {
		deductAmount := orderQty * ing.Quantity
		tx.Exec(
			`UPDATE inventory_items SET stock_qty = stock_qty - $1 WHERE id = $2`,
			deductAmount, ing.InventoryItemID,
		)
	}
}

// UpdateStatus updates an order's status with transition validation.
func (r *OrderRepository) UpdateStatus(id uuid.UUID, newStatus string) (*model.Order, error) {
	// Get current status
	var currentStatus string
	err := r.db.Get(&currentStatus, `SELECT status FROM orders WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Validate transition
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
		`UPDATE orders SET status = $1 WHERE id = $2
		 RETURNING id, order_number, source, table_number, status, total, payment_method, created_at`,
		newStatus, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}
	return &order, nil
}

// CreateWebOrder creates an order from the web menu with server-side price resolution.
func (r *OrderRepository) CreateWebOrder(tableNumber string, items []model.WebOrderItemInput) (*model.OrderWithItems, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate order number
	orderNumber, err := r.generateOrderNumber(tx)
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

	// Resolve prices and build temporary order items
	for _, item := range items {
		menuItemID, err := uuid.Parse(item.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("invalid menu_item_id: %s", item.MenuItemID)
		}

		// Get current price and name from database
		var menuItem struct {
			Price  int64  `db:"price"`
			NameAr string `db:"name_ar"`
		}
		err = tx.Get(&menuItem,
			`SELECT price, name_ar FROM menu_items WHERE id = $1 AND is_active = true`,
			menuItemID)
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

	// Insert order first to satisfy foreign key constraints
	var order model.Order
	err = tx.Get(&order,
		`INSERT INTO orders (id, order_number, source, table_number, status, total, payment_method, created_at)
		 VALUES ($1, $2, 'web_menu', $3, 'pending', $4, 'web', now())
		 RETURNING id, order_number, source, table_number, status, total, payment_method, created_at`,
		orderID, orderNumber, tableNumber, totalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	var orderItems []model.OrderItem
	for _, ri := range resolvedItems {
		var oi model.OrderItem
		// Create a copy of the UUID to take its pointer
		mID := ri.menuItemID
		err = tx.Get(&oi,
			`INSERT INTO order_items (id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot`,
			ri.itemID, order.ID, &mID, ri.quantity, ri.price, ri.lineTotal, ri.nameArSnapshot)
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

// ListByDateRange returns all orders between from and to dates (inclusive).
func (r *OrderRepository) ListByDateRange(from, to time.Time) ([]model.OrderWithItems, error) {
	var orders []model.Order
	err := r.db.Select(&orders,
		`SELECT id, order_number, source, table_number, status, total, payment_method, created_at
		 FROM orders
		 WHERE created_at >= $1 AND created_at < $2
		 ORDER BY created_at DESC`,
		from, to.Add(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	var result []model.OrderWithItems
	for _, order := range orders {
		var items []model.OrderItem
		err := r.db.Select(&items,
			`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot
			 FROM order_items WHERE order_id = $1 ORDER BY name_ar_snapshot ASC`,
			order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch order items: %w", err)
		}
		result = append(result, model.OrderWithItems{Order: order, Items: items})
	}

	return result, nil
}

