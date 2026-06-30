package service

import (
	"coffeeshop-pos/internal/model"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// OrderService is a Wails-bound service for creating and managing orders.
// All exported methods are callable from the Vue frontend.
type OrderService struct {
	db          *sqlx.DB
	configStore *ConfigStoreService
}

const orderSelectCols = `id, COALESCE(order_number,'') AS order_number, source, COALESCE(table_number,'') AS table_number, status, total, COALESCE(payment_method,'') AS payment_method, COALESCE(device_id,'') AS device_id, created_at, updated_at, synced`

// NewOrderService creates a new OrderService.
func NewOrderService(db *sqlx.DB, configStore *ConfigStoreService) *OrderService {
	return &OrderService{db: db, configStore: configStore}
}

// CreateOrder is called from the frontend when the cashier confirms checkout.
// 1. Generates UUID + local daily order number
// 2. Inserts order + order_items into SQLite with status 'accepted' (kitchen will complete)
// 3. Runs the recipe engine to deduct local inventory
// 4. Returns the created order
func (s *OrderService) CreateOrder(items []model.CartItem, paymentMethod string, tableNumber string) (*model.OrderWithItems, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}
	if paymentMethod == "" {
		paymentMethod = "cash"
	}

	// Generate UUIDs
	orderID := uuid.New().String()
	now := time.Now()
	createdAt := now.Format("2006-01-02 15:04:05")

	// Calculate total and build order items
	var total int64
	var orderItems []model.OrderItem
	for _, item := range items {
		lineTotal := item.Price * int64(item.Quantity)
		total += lineTotal

		orderItems = append(orderItems, model.OrderItem{
			ID:             uuid.New().String(),
			OrderID:        orderID,
			MenuItemID:     item.MenuItemID,
			Quantity:        item.Quantity,
			UnitPrice:      item.Price,
			LineTotal:      lineTotal,
			NameArSnapshot: item.NameAr,
		})
	}

	// Generate local daily order number
	orderNumber, err := s.generateDailyOrderNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	// Begin transaction
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Determine initial status based on kitchen mode
	initialStatus := "completed"
	if s.configStore.IsKitchenModeEnabled() {
		initialStatus = "accepted"
	}

	_, err = tx.Exec(
		`INSERT INTO orders (id, order_number, source, table_number, status, total, payment_method, created_at, updated_at, synced)
		 VALUES (?, ?, 'cashier', ?, ?, ?, ?, ?, ?, 0)`,
		orderID, orderNumber, tableNumber, initialStatus, total, paymentMethod, createdAt, createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	for _, item := range orderItems {
		_, err = tx.Exec(
			`INSERT INTO order_items (id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.OrderID, item.MenuItemID, item.Quantity,
			item.UnitPrice, item.LineTotal, item.NameArSnapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Recipe engine: deduct local inventory
	for _, item := range items {
		s.deductLocalStock(tx, item.MenuItemID, item.Quantity)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("order created", "order_id", orderID, "order_number", orderNumber, "total", total, "table", tableNumber)

	order := model.Order{
		ID:            orderID,
		OrderNumber:   orderNumber,
		Source:        "cashier",
		TableNumber:   tableNumber,
		Status:        initialStatus,
		Total:         total,
		PaymentMethod: paymentMethod,
		CreatedAt:     createdAt,
		Synced:        0,
	}

	return &model.OrderWithItems{Order: order, Items: orderItems}, nil
}

// CompleteCashierOrder marks a cashier order as completed (called from kitchen).
func (s *OrderService) CompleteCashierOrder(orderID string) (*model.OrderWithItems, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	// Check current status
	var currentStatus string
	err := s.db.Get(&currentStatus, `SELECT status FROM orders WHERE id = ?`, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if currentStatus != "accepted" {
		return nil, fmt.Errorf("can only complete accepted orders (current status: %s)", currentStatus)
	}

	_, err = s.db.Exec(
		`UPDATE orders SET status = 'completed' WHERE id = ?`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to complete order: %w", err)
	}

	// Fetch the updated order with items
	var order model.Order
	err = s.db.Get(&order,
		`SELECT ` + orderSelectCols + `
		 FROM orders WHERE id = ?`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch completed order: %w", err)
	}

	var items []model.OrderItem
	err = s.db.Select(&items,
		`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot
		 FROM order_items WHERE order_id = ?
		 ORDER BY name_ar_snapshot ASC`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}

	slog.Info("cashier order completed", "order_id", orderID, "order_number", order.OrderNumber)
	return &model.OrderWithItems{Order: order, Items: items}, nil
}

// GetAcceptedOrders returns all orders with status 'accepted' from today (both cashier and web sources).
// Used by the Kitchen Display to show orders waiting to be prepared.
func (s *OrderService) GetAcceptedOrders() ([]model.OrderWithItems, error) {
	today := time.Now().Format("2006-01-02")

	var orders []model.Order
	err := s.db.Select(&orders,
		`SELECT `+orderSelectCols+`
		 FROM orders
		 WHERE status = 'accepted' AND date(created_at) = ?
		 ORDER BY datetime(created_at) ASC`,
		today,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accepted orders: %w", err)
	}

	var result []model.OrderWithItems
	for _, order := range orders {
		var items []model.OrderItem
		err := s.db.Select(&items,
			`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot
			 FROM order_items WHERE order_id = ?
			 ORDER BY name_ar_snapshot ASC`,
			order.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch order items: %w", err)
		}
		result = append(result, model.OrderWithItems{Order: order, Items: items})
	}

	return result, nil
}

// GetTodayOrders returns all orders created today.
func (s *OrderService) GetTodayOrders() ([]model.OrderWithItems, error) {
	today := time.Now().Format("2006-01-02")

	var orders []model.Order
	err := s.db.Select(&orders,
		`SELECT `+orderSelectCols+`
		 FROM orders
		 WHERE date(created_at) = ?
		 ORDER BY datetime(created_at) DESC`,
		today,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch today's orders: %w", err)
	}

	var result []model.OrderWithItems
	for _, order := range orders {
		var items []model.OrderItem
		err := s.db.Select(&items,
			`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot
			 FROM order_items WHERE order_id = ?
			 ORDER BY name_ar_snapshot ASC`,
			order.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch order items: %w", err)
		}
		result = append(result, model.OrderWithItems{Order: order, Items: items})
	}

	return result, nil
}

// GetTodayOrderCount returns the count of today's orders (for display in header).
func (s *OrderService) GetTodayOrderCount() (int, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := s.db.Get(&count, `SELECT COUNT(*) FROM orders WHERE created_at LIKE ?`, today+"%")
	if err != nil {
		return 0, fmt.Errorf("failed to count today's orders: %w", err)
	}
	return count, nil
}

// generateDailyOrderNumber returns a local daily counter like #001, #002.
func (s *OrderService) generateDailyOrderNumber() (string, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := s.db.Get(&count, `SELECT COUNT(*) FROM orders WHERE created_at LIKE ?`, today+"%")
	if err != nil {
		return "", fmt.Errorf("failed to count today's orders: %w", err)
	}
	return fmt.Sprintf("#%03d", count+1), nil
}

// deductLocalStock looks up the recipe for a menu item and deducts stock locally.
// Best-effort: logs errors but doesn't fail the order.
func (s *OrderService) deductLocalStock(tx *sqlx.Tx, menuItemID string, orderQty int) {
	type recipeRow struct {
		InventoryItemID string `db:"inventory_item_id"`
		Quantity        int    `db:"quantity"`
	}

	var ingredients []recipeRow
	err := tx.Select(&ingredients,
		`SELECT inventory_item_id, quantity FROM recipe_ingredients WHERE menu_item_id = ?`,
		menuItemID,
	)
	if err != nil || len(ingredients) == 0 {
		return
	}

	for _, ing := range ingredients {
		deductAmount := orderQty * ing.Quantity
		_, err := tx.Exec(
			`UPDATE inventory_items SET stock_qty = stock_qty - ? WHERE id = ?`,
			deductAmount, ing.InventoryItemID,
		)
		if err != nil {
			slog.Warn("failed to deduct local stock", "inventory_item_id", ing.InventoryItemID, "error", err)
		}
	}
}

// GetOrdersByDateRange returns orders between from and to dates (inclusive).
// Dates should be in "YYYY-MM-DD" format.
func (s *OrderService) GetOrdersByDateRange(from, to string) ([]model.OrderWithItems, error) {
	if from == "" || to == "" {
		return nil, fmt.Errorf("from and to dates are required")
	}

	var orders []model.Order
	err := s.db.Select(&orders,
		`SELECT `+orderSelectCols+`
		 FROM orders
		 WHERE datetime(created_at) >= datetime(?)
		   AND datetime(created_at) < datetime(?, '+1 day')
		 ORDER BY datetime(created_at) DESC`,
		from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	var result []model.OrderWithItems
	for _, order := range orders {
		var items []model.OrderItem
		err := s.db.Select(&items,
			`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot
			 FROM order_items WHERE order_id = ?
			 ORDER BY name_ar_snapshot ASC`,
			order.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch order items: %w", err)
		}
		result = append(result, model.OrderWithItems{Order: order, Items: items})
	}

	return result, nil
}

// VoidOrder marks a completed order as voided. No stock reversal.
func (s *OrderService) VoidOrder(orderID string) (*model.Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	// Check current status
	var currentStatus string
	err := s.db.Get(&currentStatus, `SELECT status FROM orders WHERE id = ?`, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if currentStatus == "voided" {
		return nil, fmt.Errorf("order is already voided")
	}
	if currentStatus != "completed" {
		return nil, fmt.Errorf("can only void completed orders (current status: %s)", currentStatus)
	}

	_, err = s.db.Exec(
		`UPDATE orders SET status = 'voided', synced = 0 WHERE id = ?`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to void order: %w", err)
	}

	var order model.Order
	err = s.db.Get(&order,
		`SELECT ` + orderSelectCols + `
		 FROM orders WHERE id = ?`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch voided order: %w", err)
	}

	slog.Info("order voided", "order_id", orderID, "order_number", order.OrderNumber)
	return &order, nil
}

