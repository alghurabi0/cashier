package sync

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// Worker runs in the background and syncs data from the central API to local SQLite.
type Worker struct {
	client *APIClient
	db     *sqlx.DB
	mu     sync.Mutex // protects concurrent pull/push operations
}

// NewWorker creates a new sync worker.
func NewWorker(client *APIClient, db *sqlx.DB) *Worker {
	return &Worker{
		client: client,
		db:     db,
	}
}


// Start begins the sync loop. It runs an initial full sync, then syncs
// periodically at the given interval. Stops when the context is cancelled.
func (w *Worker) Start(ctx context.Context, intervalSeconds int) {
	slog.Info("sync worker starting", "interval_seconds", intervalSeconds)

	// Initial sync
	if err := w.PullAll(); err != nil {
		slog.Error("initial sync failed", "error", err)
	}
	// Push any orders created while offline
	if err := w.PushOrders(); err != nil {
		slog.Error("initial order push failed", "error", err)
	}

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("sync worker stopping")
			return
		case <-ticker.C:
			if err := w.PullAll(); err != nil {
				slog.Error("sync failed", "error", err)
			}
			if err := w.PushOrders(); err != nil {
				slog.Error("order push failed", "error", err)
			}
		}
	}
}

// PullAll performs a full sync of all tables from the API to local SQLite.
func (w *Worker) PullAll() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	slog.Debug("sync: pulling all data from API")

	if err := w.pullCategories(); err != nil {
		return fmt.Errorf("sync categories: %w", err)
	}
	if err := w.pullMenuItems(); err != nil {
		return fmt.Errorf("sync menu items: %w", err)
	}
	if err := w.pullInventory(); err != nil {
		return fmt.Errorf("sync inventory: %w", err)
	}
	if err := w.pullRecipes(); err != nil {
		return fmt.Errorf("sync recipes: %w", err)
	}

	// Update sync timestamp
	now := time.Now().UTC().Format(time.RFC3339)
	w.db.Exec(`INSERT OR REPLACE INTO sync_meta (table_name, last_synced_at) VALUES ('all', ?)`, now)

	slog.Debug("sync: pull complete")
	return nil
}

func (w *Worker) pullCategories() error {
	categories, err := w.client.GetCategories()
	if err != nil {
		return err
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, c := range categories {
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO categories (id, name_ar, sort_order, is_active)
			 VALUES (?, ?, ?, ?)`,
			c.ID, c.NameAr, c.SortOrder, c.IsActive,
		)
		if err != nil {
			return fmt.Errorf("upsert category: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	slog.Debug("sync: pulled categories", "count", len(categories))
	return nil
}

func (w *Worker) pullMenuItems() error {
	items, err := w.client.GetMenuItems()
	if err != nil {
		return err
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO menu_items
			 (id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.CategoryID, item.NameAr, item.Price,
			item.CostCalcMethod, item.ManualCostPrice, item.CachedAutoCost,
			item.ImagePath, item.IsActive,
		)
		if err != nil {
			return fmt.Errorf("upsert menu item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	slog.Debug("sync: pulled menu items", "count", len(items))
	return nil
}

func (w *Worker) pullInventory() error {
	items, err := w.client.GetInventoryItems()
	if err != nil {
		return err
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO inventory_items
			 (id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.NameAr, item.BaseUnitAr, item.StockQty,
			item.LowStockThreshold, item.UnitCost, item.IsActive,
		)
		if err != nil {
			return fmt.Errorf("upsert inventory item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	slog.Debug("sync: pulled inventory items", "count", len(items))
	return nil
}

func (w *Worker) pullRecipes() error {
	// Get all menu item IDs from local DB
	var menuItemIDs []string
	if err := w.db.Select(&menuItemIDs, `SELECT id FROM menu_items WHERE is_active = 1`); err != nil {
		return fmt.Errorf("fetch menu item IDs: %w", err)
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	totalCount := 0
	for _, menuItemID := range menuItemIDs {
		ingredients, err := w.client.GetRecipe(menuItemID)
		if err != nil {
			slog.Warn("sync: failed to fetch recipe", "menu_item_id", menuItemID, "error", err)
			continue
		}

		// Clear existing recipe for this menu item
		tx.Exec(`DELETE FROM recipe_ingredients WHERE menu_item_id = ?`, menuItemID)

		for _, ing := range ingredients {
			_, err := tx.Exec(
				`INSERT INTO recipe_ingredients (id, menu_item_id, inventory_item_id, quantity)
				 VALUES (?, ?, ?, ?)`,
				ing.ID, ing.MenuItemID, ing.InventoryItemID, ing.Quantity,
			)
			if err != nil {
				return fmt.Errorf("insert recipe ingredient: %w", err)
			}
		}
		totalCount += len(ingredients)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	slog.Debug("sync: pulled recipes", "menu_items", len(menuItemIDs), "total_ingredients", totalCount)
	return nil
}

// PushOrders sends all unsynced local orders to the central API.
func (w *Worker) PushOrders() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	type orderRow struct {
		ID            string `db:"id"`
		OrderNumber   string `db:"order_number"`
		Source        string `db:"source"`
		TableNumber   string `db:"table_number"`
		Total         int64  `db:"total"`
		PaymentMethod string `db:"payment_method"`
		CreatedAt     string `db:"created_at"`
	}

	var orders []orderRow
	err := w.db.Select(&orders, `SELECT id, order_number, source, table_number, total, payment_method, created_at FROM orders WHERE synced = 0`)
	if err != nil {
		return fmt.Errorf("fetch unsynced orders: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	slog.Debug("sync: pushing orders", "count", len(orders))

	for _, order := range orders {
		// Fetch order items
		type itemRow struct {
			ID             string `db:"id"`
			MenuItemID     string `db:"menu_item_id"`
			Quantity       int    `db:"quantity"`
			UnitPrice      int64  `db:"unit_price"`
			LineTotal      int64  `db:"line_total"`
			NameArSnapshot string `db:"name_ar_snapshot"`
		}

		var items []itemRow
		err := w.db.Select(&items, `SELECT id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot FROM order_items WHERE order_id = ?`, order.ID)
		if err != nil {
			slog.Warn("sync: failed to fetch order items", "order_id", order.ID, "error", err)
			continue
		}

		// Build push payload
		payload := OrderPushPayload{
			ID:            order.ID,
			Source:        order.Source,
			TableNumber:   order.TableNumber,
			Total:         order.Total,
			PaymentMethod: order.PaymentMethod,
			CreatedAt:     order.CreatedAt,
		}
		for _, item := range items {
			payload.Items = append(payload.Items, OrderItemPush{
				ID:             item.ID,
				MenuItemID:     item.MenuItemID,
				Quantity:       item.Quantity,
				UnitPrice:      item.UnitPrice,
				LineTotal:      item.LineTotal,
				NameArSnapshot: item.NameArSnapshot,
			})
		}

		// Push to API
		result, err := w.client.PushOrder(payload)
		if err != nil {
			slog.Warn("sync: failed to push order", "order_id", order.ID, "error", err)
			continue // Will retry on next cycle
		}

		// Mark as synced and update server-assigned order number
		serverOrderNumber := result.OrderNumber
		if serverOrderNumber == "" {
			serverOrderNumber = order.OrderNumber
		}

		_, err = w.db.Exec(
			`UPDATE orders SET synced = 1, order_number = ? WHERE id = ?`,
			serverOrderNumber, order.ID,
		)
		if err != nil {
			slog.Warn("sync: failed to mark order as synced", "order_id", order.ID, "error", err)
		} else {
			slog.Info("sync: pushed order", "order_id", order.ID, "server_order_number", serverOrderNumber)
		}
	}

	return nil
}

