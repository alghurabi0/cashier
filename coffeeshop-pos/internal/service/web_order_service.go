package service

import (
	"coffeeshop-pos/internal/model"
	"encoding/json"
	"fmt"
	"log/slog"
	gosync "sync"

	posSync "coffeeshop-pos/internal/sync"

	"github.com/jmoiron/sqlx"
)

// WebOrderService manages incoming web menu orders on the POS side.
// It is Wails-bound so the Vue frontend can call its methods.
type WebOrderService struct {
	db        *sqlx.DB
	apiClient *posSync.APIClient
	mu        gosync.RWMutex

	// In-memory queues for web orders by status
	pendingOrders   []model.OrderWithItems
	acceptedOrders  []model.OrderWithItems
	completedOrders []model.OrderWithItems
}

// NewWebOrderService creates a new WebOrderService.
func NewWebOrderService(db *sqlx.DB, apiClient *posSync.APIClient) *WebOrderService {
	return &WebOrderService{
		db:        db,
		apiClient: apiClient,
	}
}

// HandleSSEEvent processes an SSE event from the API.
// Called by the SSE client callback — not exposed to frontend.
func (s *WebOrderService) HandleSSEEvent(event posSync.SSEEvent) {
	switch event.Type {
	case "new_order":
		var order model.OrderWithItems
		if err := json.Unmarshal(event.Data, &order); err != nil {
			slog.Warn("web-orders: failed to parse new_order event", "error", err)
			return
		}
		s.mu.Lock()
		s.pendingOrders = append(s.pendingOrders, order)
		s.mu.Unlock()
		slog.Info("web-orders: new order received", "order_number", order.OrderNumber, "table", order.TableNumber)
	case "order_status":
		// Another POS client may have updated the status
		slog.Debug("web-orders: status update event received")
	}
}

// GetPendingOrders returns current pending web orders.
func (s *WebOrderService) GetPendingOrders() []model.OrderWithItems {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.OrderWithItems, len(s.pendingOrders))
	copy(result, s.pendingOrders)
	return result
}

// GetAcceptedOrders returns accepted (in-progress) web orders.
func (s *WebOrderService) GetAcceptedOrders() []model.OrderWithItems {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.OrderWithItems, len(s.acceptedOrders))
	copy(result, s.acceptedOrders)
	return result
}

// GetCompletedOrders returns completed web orders.
func (s *WebOrderService) GetCompletedOrders() []model.OrderWithItems {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.OrderWithItems, len(s.completedOrders))
	copy(result, s.completedOrders)
	return result
}

// AcceptOrder accepts a pending order, triggers stock deduction, and updates the API.
func (s *WebOrderService) AcceptOrder(orderID string) error {
	// Update status on API
	if err := s.apiClient.UpdateOrderStatus(orderID, "accepted"); err != nil {
		return fmt.Errorf("failed to accept order: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Move from pending to accepted
	for i, o := range s.pendingOrders {
		if o.ID == orderID {
			o.Status = "accepted"
			s.acceptedOrders = append(s.acceptedOrders, o)
			s.pendingOrders = append(s.pendingOrders[:i], s.pendingOrders[i+1:]...)

			// Insert into local SQLite + recipe-based stock deduction
			s.insertLocalOrder(o)

			slog.Info("web-orders: order accepted", "order_id", orderID)
			return nil
		}
	}

	return fmt.Errorf("order not found in pending queue: %s", orderID)
}

// RejectOrder rejects a pending order and updates the API.
func (s *WebOrderService) RejectOrder(orderID string) error {
	if err := s.apiClient.UpdateOrderStatus(orderID, "rejected"); err != nil {
		return fmt.Errorf("failed to reject order: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from pending
	for i, o := range s.pendingOrders {
		if o.ID == orderID {
			s.pendingOrders = append(s.pendingOrders[:i], s.pendingOrders[i+1:]...)
			slog.Info("web-orders: order rejected", "order_id", orderID)
			return nil
		}
	}

	return fmt.Errorf("order not found in pending queue: %s", orderID)
}

// CompleteOrder marks an accepted order as completed and updates the API.
func (s *WebOrderService) CompleteOrder(orderID string) error {
	if err := s.apiClient.UpdateOrderStatus(orderID, "completed"); err != nil {
		return fmt.Errorf("failed to complete order: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, o := range s.acceptedOrders {
		if o.ID == orderID {
			o.Status = "completed"
			s.completedOrders = append(s.completedOrders, o)
			s.acceptedOrders = append(s.acceptedOrders[:i], s.acceptedOrders[i+1:]...)
			slog.Info("web-orders: order completed", "order_id", orderID)
			return nil
		}
	}

	return fmt.Errorf("order not found in accepted queue: %s", orderID)
}

// insertLocalOrder saves a web order to local SQLite and performs recipe-based stock deduction.
func (s *WebOrderService) insertLocalOrder(order model.OrderWithItems) {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Warn("web-orders: failed to begin local tx", "error", err)
		return
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.Exec(
		`INSERT OR IGNORE INTO orders (id, order_number, source, table_number, status, total, payment_method, created_at, synced)
		 VALUES (?, ?, 'web_menu', ?, 'accepted', ?, 'web', ?, 1)`,
		order.ID, order.OrderNumber, order.TableNumber, order.Total, order.CreatedAt,
	)
	if err != nil {
		slog.Warn("web-orders: failed to insert local order", "error", err)
		return
	}

	// Insert items
	for _, item := range order.Items {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO order_items (id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			item.ID, order.ID, item.MenuItemID, item.Quantity, item.UnitPrice, item.LineTotal, item.NameArSnapshot,
		)
		if err != nil {
			slog.Warn("web-orders: failed to insert order item", "error", err)
		}

		// Recipe-based stock deduction
		s.deductStockForItem(tx, item.MenuItemID, item.Quantity)
	}

	if err := tx.Commit(); err != nil {
		slog.Warn("web-orders: failed to commit local order", "error", err)
	}
}

// deductStockForItem deducts local inventory based on recipe.
func (s *WebOrderService) deductStockForItem(tx *sqlx.Tx, menuItemID string, orderQty int) {
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
		tx.Exec(
			`UPDATE inventory_items SET stock_qty = stock_qty - ? WHERE id = ?`,
			deductAmount, ing.InventoryItemID,
		)
	}
}
