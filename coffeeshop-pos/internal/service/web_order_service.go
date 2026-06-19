package service

import (
	"coffeeshop-pos/internal/model"
	"encoding/json"
	"fmt"
	"log/slog"
	gosync "sync"
	"time"

	posSync "coffeeshop-pos/internal/sync"

	"github.com/jmoiron/sqlx"
)

// WebOrderService manages incoming web menu orders on the POS side.
// It is Wails-bound so the Vue frontend can call its methods.
type WebOrderService struct {
	db          *sqlx.DB
	apiClient   *posSync.APIClient
	configStore *ConfigStoreService
	mu          gosync.RWMutex
	fetched     bool

	// In-memory queues for web orders by status
	pendingOrders   []model.OrderWithItems
	acceptedOrders  []model.OrderWithItems
	completedOrders []model.OrderWithItems
}

// NewWebOrderService creates a new WebOrderService.
func NewWebOrderService(db *sqlx.DB, apiClient *posSync.APIClient, configStore *ConfigStoreService) *WebOrderService {
	return &WebOrderService{
		db:          db,
		apiClient:   apiClient,
		configStore: configStore,
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
		// TODO
		slog.Debug("web-orders: status update event received")
	}
}

// GetPendingOrders returns current pending web orders.
func (s *WebOrderService) GetPendingOrders() []model.OrderWithItems {
	s.ensureFetched()
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.OrderWithItems, len(s.pendingOrders))
	copy(result, s.pendingOrders)
	return result
}

// GetAcceptedOrders returns accepted (in-progress) web orders.
func (s *WebOrderService) GetAcceptedOrders() []model.OrderWithItems {
	s.ensureFetched()
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.OrderWithItems, len(s.acceptedOrders))
	copy(result, s.acceptedOrders)
	return result
}

// GetCompletedOrders returns completed web orders.
func (s *WebOrderService) GetCompletedOrders() []model.OrderWithItems {
	s.ensureFetched()
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.OrderWithItems, len(s.completedOrders))
	copy(result, s.completedOrders)
	return result
}

// ensureFetched guarantees that we've pulled the current web orders from the API at least once.
func (s *WebOrderService) ensureFetched() {
	s.mu.Lock()
	alreadyFetched := s.fetched
	s.mu.Unlock()

	if alreadyFetched {
		return
	}

	if err := s.FetchPendingFromAPI(); err != nil {
		slog.Warn("web-orders: failed to fetch initial pending orders from API", "error", err)
	}
}

// FetchPendingFromAPI pulls pending, accepted, and completed web orders for today from the central API.
// This is used on startup/refresh to populate the in-memory queue.
func (s *WebOrderService) FetchPendingFromAPI() error {
	// Call API to get today's orders
	today := time.Now().Format("2006-01-02")
	orders, err := s.apiClient.GetOrders(today, today)
	if err != nil {
		return err
	}

	var pending []model.OrderWithItems
	var accepted []model.OrderWithItems
	var completed []model.OrderWithItems

	for _, o := range orders {
		if o.Source == "web_menu" {
			switch o.Status {
			case "pending":
				pending = append(pending, o)
			case "accepted":
				accepted = append(accepted, o)
			case "completed":
				completed = append(completed, o)
			}
		}
	}

	s.mu.Lock()
	s.pendingOrders = pending
	s.acceptedOrders = accepted
	s.completedOrders = completed
	s.fetched = true
	s.mu.Unlock()

	return nil
}

// AcceptOrder accepts a pending order, triggers stock deduction, and updates the API.
// When kitchen mode is off, the order is also immediately completed.
func (s *WebOrderService) AcceptOrder(orderID string) error {
	// Update status on API to "accepted"
	if err := s.apiClient.UpdateOrderStatus(orderID, "accepted"); err != nil {
		return fmt.Errorf("failed to accept order: %w", err)
	}

	// If kitchen mode is off, immediately complete the order too
	skipKitchen := !s.configStore.IsKitchenModeEnabled()
	if skipKitchen {
		if err := s.apiClient.UpdateOrderStatus(orderID, "completed"); err != nil {
			return fmt.Errorf("failed to complete order: %w", err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Move from pending to accepted (or completed if skipping kitchen)
	for i, o := range s.pendingOrders {
		if o.ID == orderID {
			if skipKitchen {
				o.Status = "completed"
				s.completedOrders = append(s.completedOrders, o)
			} else {
				o.Status = "accepted"
				s.acceptedOrders = append(s.acceptedOrders, o)
			}
			s.pendingOrders = append(s.pendingOrders[:i], s.pendingOrders[i+1:]...)

			// Insert into local SQLite + recipe-based stock deduction
			s.insertLocalOrder(o)

			slog.Info("web-orders: order accepted", "order_id", orderID, "skip_kitchen", skipKitchen)
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

			// Sync local SQLite status so order history shows "completed"
			s.db.Exec(`UPDATE orders SET status = 'completed' WHERE id = ?`, orderID)

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
		 VALUES (?, ?, 'web_menu', ?, ?, ?, 'web', ?, 1)`,
		order.ID, order.OrderNumber, order.TableNumber, order.Status, order.Total, order.CreatedAt,
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
