package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"coffeeshop-pos/internal/model"

	"github.com/jmoiron/sqlx"
)

type Worker struct {
	client         *APIClient
	db             *sqlx.DB
	mu             sync.Mutex
	Status         *SyncStatus
	triggerCh      chan struct{}
	OnMenuPulled   func(imageURLs []string) // called after menu items are synced
}

// NewWorker creates a new sync worker.
func NewWorker(client *APIClient, db *sqlx.DB) *Worker {
	status := NewSyncStatus()
	status.SetAuditLogger(NewAuditLogger(db))

	return &Worker{
		client:    client,
		db:        db,
		Status:    status,
		triggerCh: make(chan struct{}, 1), // buffered so TriggerPull never blocks
	}
}

// TriggerPull signals the sync loop to run an immediate pull cycle.
// Non-blocking: if a trigger is already pending, the call is a no-op (debounce).
func (w *Worker) TriggerPull() {
	select {
	case w.triggerCh <- struct{}{}:
		slog.Debug("sync: immediate pull triggered via SSE")
	default:
		// Already triggered, skip (debounce)
	}
}

// checkHealth probes the API health endpoint to determine connectivity.
func (w *Worker) checkHealth() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(w.client.baseURL + "/api/v1/health")
	if err != nil {
		w.Status.SetConnected(false, err.Error())
		w.Status.Log("health", "connection", "error", err.Error(), 0)
		return false
	}
	defer resp.Body.Close()

	connected := resp.StatusCode == http.StatusOK
	if connected {
		w.Status.SetConnected(true, "")
		w.Status.Log("health", "connection", "ok", "connected", 0)
	} else {
		errMsg := fmt.Sprintf("health check returned status %d", resp.StatusCode)
		w.Status.SetConnected(false, errMsg)
		w.Status.Log("health", "connection", "error", errMsg, 0)
	}
	return connected
}

// refreshOrderCounts queries SQLite for pending/failed order counts and updates status.
func (w *Worker) refreshOrderCounts() {
	var pending, failed int
	w.db.Get(&pending, `SELECT COUNT(*) FROM orders WHERE synced = 0`)
	w.db.Get(&failed, `SELECT COUNT(*) FROM orders WHERE synced = 2`)
	w.Status.SetOrderCounts(pending, failed)
}

// Start begins the sync loop. It runs an initial full sync, then syncs
// periodically at the given interval. Stops when the context is cancelled.
func (w *Worker) Start(ctx context.Context, intervalSeconds int) {
	slog.Info("sync worker starting", "interval_seconds", intervalSeconds)

	// Initial health check + sync
	online := w.checkHealth()
	if online {
		w.Status.SetSyncing(true)
		if err := w.PullAll(); err != nil {
			slog.Error("initial sync failed", "error", err)
			w.Status.RecordError()
		} else {
			w.Status.RecordPullSuccess()
		}
		w.Status.SetSyncing(false)
	}

	// Push any orders created while offline (always try)
	if err := w.PushOrders(); err != nil {
		slog.Error("initial order push failed", "error", err)
	}
	w.refreshOrderCounts()

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("sync worker stopping")
			return
		case <-ticker.C:
			w.runSyncCycle()
		case <-w.triggerCh:
			slog.Info("sync: running immediate cycle (SSE trigger)")
			w.runSyncCycle()
		}
	}
}

// runSyncCycle performs a single pull + push cycle.
func (w *Worker) runSyncCycle() {
	online := w.checkHealth()
	w.Status.SetSyncing(true)

	if online {
		if err := w.PullAll(); err != nil {
			slog.Error("sync failed", "error", err)
			w.Status.RecordError()
		} else {
			w.Status.RecordPullSuccess()
		}
	} else {
		slog.Debug("sync: skipping pull — API offline")
		w.Status.Log("pull", "all", "skipped", "API offline", 0)
	}

	// Push admin changes (categories, menu items, inventory, etc.)
	if err := w.PushChangeLog(); err != nil {
		slog.Error("change log push failed", "error", err)
	}

	if err := w.PushOrders(); err != nil {
		slog.Error("order push failed", "error", err)
	} else {
		w.Status.RecordPushSuccess()
	}

	w.refreshOrderCounts()
	w.Status.SetSyncing(false)
}

// PullAll syncs all tables from the API to local SQLite.
// Uses delta sync (timestamp-based) when possible, falling back to full sync on first run.
func (w *Worker) PullAll() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	slog.Debug("sync: pulling data from API")

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
	if err := w.pullTables(); err != nil {
		return fmt.Errorf("sync tables: %w", err)
	}
	if err := w.pullOrders(); err != nil {
		return fmt.Errorf("sync orders: %w", err)
	}

	// Update global sync timestamp (backwards compat with GetLastSyncTime)
	now := time.Now().UTC().Format(time.RFC3339)
	w.db.Exec(`INSERT OR REPLACE INTO sync_meta (table_name, last_synced_at) VALUES ('all', ?)`, now)

	slog.Debug("sync: pull complete")
	return nil
}

// hasPendingChange checks if there's a pending (status=0) change_log entry for the given entity.
// If true, we skip overwriting this entity during pull to preserve local-first writes.
func (w *Worker) hasPendingChange(entityType, entityID string) bool {
	var count int
	w.db.Get(&count, `SELECT COUNT(*) FROM change_log WHERE entity_type = ? AND entity_id = ? AND status = 0`, entityType, entityID)
	return count > 0
}

// getLastSyncedAt returns the per-table last sync timestamp, or "" if never synced.
func (w *Worker) getLastSyncedAt(tableName string) string {
	var ts string
	err := w.db.Get(&ts, `SELECT last_synced_at FROM sync_meta WHERE table_name = ?`, tableName)
	if err != nil || ts == "" {
		return ""
	}
	return ts
}

// updateLastSyncedAt stores the current UTC time as the per-table last sync timestamp.
func (w *Worker) updateLastSyncedAt(tableName string) {
	now := time.Now().UTC().Format(time.RFC3339)
	w.db.Exec(`INSERT OR REPLACE INTO sync_meta (table_name, last_synced_at) VALUES (?, ?)`, tableName, now)
	w.Status.SetTableSyncTime(tableName, now)
}

func (w *Worker) pullCategories() error {
	since := w.getLastSyncedAt("categories")

	var categories []struct {
		ID        string `json:"id"`
		NameAr    string `json:"name_ar"`
		SortOrder int    `json:"sort_order"`
		IsActive  bool   `json:"is_active"`
		UpdatedAt string `json:"updated_at"`
	}

	var err error
	if since != "" {
		// Delta sync
		raw, fetchErr := w.client.GetCategoriesSince(since)
		if fetchErr != nil {
			return fetchErr
		}
		for _, c := range raw {
			categories = append(categories, struct {
				ID        string `json:"id"`
				NameAr    string `json:"name_ar"`
				SortOrder int    `json:"sort_order"`
				IsActive  bool   `json:"is_active"`
				UpdatedAt string `json:"updated_at"`
			}{c.ID, c.NameAr, c.SortOrder, c.IsActive, c.UpdatedAt})
		}
	} else {
		// Full sync
		raw, fetchErr := w.client.GetCategories()
		if fetchErr != nil {
			return fetchErr
		}
		for _, c := range raw {
			categories = append(categories, struct {
				ID        string `json:"id"`
				NameAr    string `json:"name_ar"`
				SortOrder int    `json:"sort_order"`
				IsActive  bool   `json:"is_active"`
				UpdatedAt string `json:"updated_at"`
			}{c.ID, c.NameAr, c.SortOrder, c.IsActive, c.UpdatedAt})
		}
	}
	_ = err

	if len(categories) == 0 {
		slog.Debug("sync: categories up to date")
		w.updateLastSyncedAt("categories")
		w.Status.Log("pull", "categories", "ok", "up to date", 0)
		return nil
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, c := range categories {
		// Skip entities with pending local changes to avoid overwriting un-pushed edits
		if w.hasPendingChange("category", c.ID) {
			continue
		}
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO categories (id, name_ar, sort_order, is_active, updated_at)
			 VALUES (?, ?, ?, ?, ?)`,
			c.ID, c.NameAr, c.SortOrder, c.IsActive, c.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("upsert category: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	w.updateLastSyncedAt("categories")
	mode := "full"
	if since != "" { mode = "delta" }
	w.Status.Log("pull", "categories", "ok", mode, len(categories))
	slog.Debug("sync: pulled categories", "count", len(categories), "delta", since != "")
	return nil
}

func (w *Worker) pullMenuItems() error {
	since := w.getLastSyncedAt("menu_items")

	var items []struct {
		ID              string `json:"id"`
		CategoryID      string `json:"category_id"`
		NameAr          string `json:"name_ar"`
		Price           int64  `json:"price"`
		CostCalcMethod  string `json:"cost_calc_method"`
		ManualCostPrice int64  `json:"manual_cost_price"`
		CachedAutoCost  int64  `json:"cached_auto_cost"`
		ImagePath       string `json:"image_path"`
		IsActive        bool   `json:"is_active"`
		UpdatedAt       string `json:"updated_at"`
	}

	if since != "" {
		raw, err := w.client.GetMenuItemsSince(since)
		if err != nil {
			return err
		}
		for _, r := range raw {
			items = append(items, struct {
				ID              string `json:"id"`
				CategoryID      string `json:"category_id"`
				NameAr          string `json:"name_ar"`
				Price           int64  `json:"price"`
				CostCalcMethod  string `json:"cost_calc_method"`
				ManualCostPrice int64  `json:"manual_cost_price"`
				CachedAutoCost  int64  `json:"cached_auto_cost"`
				ImagePath       string `json:"image_path"`
				IsActive        bool   `json:"is_active"`
				UpdatedAt       string `json:"updated_at"`
			}{r.ID, r.CategoryID, r.NameAr, r.Price, r.CostCalcMethod, r.ManualCostPrice, r.CachedAutoCost, r.ImagePath, r.IsActive, r.UpdatedAt})
		}
	} else {
		raw, err := w.client.GetMenuItems()
		if err != nil {
			return err
		}
		for _, r := range raw {
			items = append(items, struct {
				ID              string `json:"id"`
				CategoryID      string `json:"category_id"`
				NameAr          string `json:"name_ar"`
				Price           int64  `json:"price"`
				CostCalcMethod  string `json:"cost_calc_method"`
				ManualCostPrice int64  `json:"manual_cost_price"`
				CachedAutoCost  int64  `json:"cached_auto_cost"`
				ImagePath       string `json:"image_path"`
				IsActive        bool   `json:"is_active"`
				UpdatedAt       string `json:"updated_at"`
			}{r.ID, r.CategoryID, r.NameAr, r.Price, r.CostCalcMethod, r.ManualCostPrice, r.CachedAutoCost, r.ImagePath, r.IsActive, r.UpdatedAt})
		}
	}

	if len(items) == 0 {
		slog.Debug("sync: menu items up to date")
		w.updateLastSyncedAt("menu_items")
		w.Status.Log("pull", "menu_items", "ok", "up to date", 0)
		return nil
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		if w.hasPendingChange("menu_item", item.ID) {
			continue
		}
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO menu_items
			 (id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.CategoryID, item.NameAr, item.Price,
			item.CostCalcMethod, item.ManualCostPrice, item.CachedAutoCost,
			item.ImagePath, item.IsActive, item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("upsert menu item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	w.updateLastSyncedAt("menu_items")
	mode := "full"
	if since != "" { mode = "delta" }
	w.Status.Log("pull", "menu_items", "ok", mode, len(items))
	slog.Debug("sync: pulled menu items", "count", len(items), "delta", since != "")

	if w.OnMenuPulled != nil {
		urls := make([]string, len(items))
		for i, item := range items {
			urls[i] = item.ImagePath
		}
		w.OnMenuPulled(urls)
	}

	return nil
}

func (w *Worker) pullInventory() error {
	since := w.getLastSyncedAt("inventory_items")

	var items []struct {
		ID                string `json:"id"`
		NameAr            string `json:"name_ar"`
		BaseUnitAr        string `json:"base_unit_ar"`
		StockQty          int    `json:"stock_qty"`
		LowStockThreshold int    `json:"low_stock_threshold"`
		UnitCost          int64  `json:"unit_cost"`
		IsActive          bool   `json:"is_active"`
		UpdatedAt         string `json:"updated_at"`
	}

	if since != "" {
		raw, err := w.client.GetInventoryItemsSince(since)
		if err != nil {
			return err
		}
		for _, r := range raw {
			items = append(items, struct {
				ID                string `json:"id"`
				NameAr            string `json:"name_ar"`
				BaseUnitAr        string `json:"base_unit_ar"`
				StockQty          int    `json:"stock_qty"`
				LowStockThreshold int    `json:"low_stock_threshold"`
				UnitCost          int64  `json:"unit_cost"`
				IsActive          bool   `json:"is_active"`
				UpdatedAt         string `json:"updated_at"`
			}{r.ID, r.NameAr, r.BaseUnitAr, r.StockQty, r.LowStockThreshold, r.UnitCost, r.IsActive, r.UpdatedAt})
		}
	} else {
		raw, err := w.client.GetInventoryItems()
		if err != nil {
			return err
		}
		for _, r := range raw {
			items = append(items, struct {
				ID                string `json:"id"`
				NameAr            string `json:"name_ar"`
				BaseUnitAr        string `json:"base_unit_ar"`
				StockQty          int    `json:"stock_qty"`
				LowStockThreshold int    `json:"low_stock_threshold"`
				UnitCost          int64  `json:"unit_cost"`
				IsActive          bool   `json:"is_active"`
				UpdatedAt         string `json:"updated_at"`
			}{r.ID, r.NameAr, r.BaseUnitAr, r.StockQty, r.LowStockThreshold, r.UnitCost, r.IsActive, r.UpdatedAt})
		}
	}

	if len(items) == 0 {
		slog.Debug("sync: inventory items up to date")
		w.updateLastSyncedAt("inventory_items")
		w.Status.Log("pull", "inventory_items", "ok", "up to date", 0)
		return nil
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		if w.hasPendingChange("inventory_item", item.ID) {
			continue
		}
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO inventory_items
			 (id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.NameAr, item.BaseUnitAr, item.StockQty,
			item.LowStockThreshold, item.UnitCost, item.IsActive, item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("upsert inventory item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	w.updateLastSyncedAt("inventory_items")
	mode := "full"
	if since != "" { mode = "delta" }
	w.Status.Log("pull", "inventory_items", "ok", mode, len(items))
	slog.Debug("sync: pulled inventory items", "count", len(items), "delta", since != "")
	return nil
}

func (w *Worker) pullRecipes() error {
	since := w.getLastSyncedAt("recipe_ingredients")

	var ingredients []struct {
		ID              string `json:"id"`
		MenuItemID      string `json:"menu_item_id"`
		InventoryItemID string `json:"inventory_item_id"`
		Quantity        int    `json:"quantity"`
		UpdatedAt       string `json:"updated_at"`
	}

	if since != "" {
		raw, err := w.client.GetAllRecipesSince(since)
		if err != nil {
			return err
		}
		for _, r := range raw {
			ingredients = append(ingredients, struct {
				ID              string `json:"id"`
				MenuItemID      string `json:"menu_item_id"`
				InventoryItemID string `json:"inventory_item_id"`
				Quantity        int    `json:"quantity"`
				UpdatedAt       string `json:"updated_at"`
			}{r.ID, r.MenuItemID, r.InventoryItemID, r.Quantity, r.UpdatedAt})
		}
	} else {
		raw, err := w.client.GetAllRecipes()
		if err != nil {
			return err
		}
		for _, r := range raw {
			ingredients = append(ingredients, struct {
				ID              string `json:"id"`
				MenuItemID      string `json:"menu_item_id"`
				InventoryItemID string `json:"inventory_item_id"`
				Quantity        int    `json:"quantity"`
				UpdatedAt       string `json:"updated_at"`
			}{r.ID, r.MenuItemID, r.InventoryItemID, r.Quantity, r.UpdatedAt})
		}
	}

	if len(ingredients) == 0 {
		slog.Debug("sync: recipes up to date")
		w.updateLastSyncedAt("recipe_ingredients")
		w.Status.Log("pull", "recipe_ingredients", "ok", "up to date", 0)
		return nil
	}

	// Group ingredients by menu_item_id for delete-and-replace
	byMenuItem := make(map[string][]struct {
		ID              string `json:"id"`
		MenuItemID      string `json:"menu_item_id"`
		InventoryItemID string `json:"inventory_item_id"`
		Quantity        int    `json:"quantity"`
		UpdatedAt       string `json:"updated_at"`
	})
	for _, ing := range ingredients {
		byMenuItem[ing.MenuItemID] = append(byMenuItem[ing.MenuItemID], ing)
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for menuItemID, ings := range byMenuItem {
		// Delete existing recipe for this menu item
		tx.Exec(`DELETE FROM recipe_ingredients WHERE menu_item_id = ?`, menuItemID)

		for _, ing := range ings {
			_, err := tx.Exec(
				`INSERT INTO recipe_ingredients (id, menu_item_id, inventory_item_id, quantity, updated_at)
				 VALUES (?, ?, ?, ?, ?)`,
				ing.ID, ing.MenuItemID, ing.InventoryItemID, ing.Quantity, ing.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("insert recipe ingredient: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	w.updateLastSyncedAt("recipe_ingredients")
	mode := "full"
	if since != "" { mode = "delta" }
	w.Status.Log("pull", "recipe_ingredients", "ok", mode, len(ingredients))
	slog.Debug("sync: pulled recipes", "menu_items", len(byMenuItem), "total_ingredients", len(ingredients), "delta", since != "")
	return nil
}

func (w *Worker) pullTables() error {
	tables, err := w.client.ListTables()
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		w.Status.Log("pull", "tables", "ok", "up to date", 0)
		return nil
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, t := range tables {
		// Skip tables with pending local changes
		if w.hasPendingChange("table", t.ID) {
			continue
		}
		_, err := tx.Exec(
			`INSERT OR REPLACE INTO tables (id, number, token, is_active, synced, created_at)
			 VALUES (?, ?, ?, ?, 1, ?)`,
			t.ID, t.Number, t.Token, t.IsActive, t.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("upsert table: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	w.Status.Log("pull", "tables", "ok", "full", len(tables))
	slog.Debug("sync: pulled tables", "count", len(tables))
	return nil
}

// pullOrders fetches orders from the API and upserts into local SQLite.
// Uses delta sync via sync_meta. Orders already existing locally (from local creation)
// are updated for status changes only. New orders (from other POS terminals) are inserted.
func (w *Worker) pullOrders() error {
	since := w.getLastSyncedAt("orders")

	var orders []model.OrderWithItems
	var err error

	if since != "" {
		orders, err = w.client.GetOrdersSince(since)
	} else {
		// First sync: pull today's orders
		today := time.Now().Format("2006-01-02")
		orders, err = w.client.GetOrders(today, today)
	}
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		w.Status.Log("pull", "orders", "ok", "up to date", 0)
		return nil
	}

	tx, err := w.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	inserted := 0
	updated := 0

	for _, o := range orders {
		// Check if order already exists locally
		var existingID string
		existsErr := tx.Get(&existingID, `SELECT id FROM orders WHERE id = ?`, o.ID)

		if existsErr != nil {
			// Order doesn't exist locally — insert it (from another POS terminal)
			_, err := tx.Exec(
				`INSERT OR IGNORE INTO orders (id, order_number, source, table_number, status, total, payment_method, device_id, created_at, updated_at, synced)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
				o.ID, o.OrderNumber, o.Source, o.TableNumber, o.Status,
				o.Total, o.PaymentMethod, o.DeviceID,
				o.CreatedAt, o.UpdatedAt,
			)
			if err != nil {
				slog.Warn("sync: failed to insert pulled order", "id", o.ID, "error", err)
				continue
			}

			// Insert order items
			for _, item := range o.Items {
				tx.Exec(
					`INSERT OR IGNORE INTO order_items (id, order_id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot)
					 VALUES (?, ?, ?, ?, ?, ?, ?)`,
					item.ID, o.ID, item.MenuItemID, item.Quantity,
					item.UnitPrice, item.LineTotal, item.NameArSnapshot,
				)
			}
			inserted++
		} else {
			// Order exists — update status if changed (e.g., accepted → completed)
			tx.Exec(
				`UPDATE orders SET status = ?, updated_at = ? WHERE id = ? AND status != ?`,
				o.Status, o.UpdatedAt, o.ID, o.Status,
			)
			updated++
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	w.updateLastSyncedAt("orders")
	mode := "full"
	if since != "" {
		mode = "delta"
	}
	w.Status.Log("pull", "orders", "ok", mode, len(orders))
	slog.Debug("sync: pulled orders", "total", len(orders), "inserted", inserted, "updated", updated, "delta", since != "")
	return nil
}

// PushChangeLog processes the change_log queue and pushes pending admin changes to the API.
// Uses the same backoff/dead-letter pattern as PushOrders.
func (w *Worker) PushChangeLog() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	type changeEntry struct {
		ID          int    `db:"id"`
		EntityType  string `db:"entity_type"`
		EntityID    string `db:"entity_id"`
		Operation   string `db:"operation"`
		Payload     string `db:"payload"`
		BaseVersion string `db:"base_version"`
		RetryCount  int    `db:"retry_count"`
		LastRetryAt string `db:"last_retry_at"`
	}

	var entries []changeEntry
	err := w.db.Select(&entries,
		`SELECT id, entity_type, entity_id, operation, payload, base_version, retry_count, last_retry_at
		 FROM change_log WHERE status = 0
		 ORDER BY created_at ASC`)
	if err != nil {
		return fmt.Errorf("fetch pending changes: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	now := time.Now()
	pushed := 0
	failed := 0

	for _, entry := range entries {
		// Check backoff
		if entry.RetryCount > 0 && entry.LastRetryAt != "" {
			lastRetry, parseErr := time.Parse(time.RFC3339, entry.LastRetryAt)
			if parseErr == nil {
				if now.Sub(lastRetry) < backoffDuration(entry.RetryCount) {
					continue // skip, still in backoff
				}
			}
		}

		pushErr := w.dispatchChange(entry.EntityType, entry.EntityID, entry.Operation, entry.Payload, entry.BaseVersion)

		// LWW conflict resolution: on 409, log conflict and retry without version header
		if pushErr != nil {
			if _, ok := pushErr.(*SyncConflictError); ok {
				slog.Warn("sync: conflict detected, applying LWW (force overwrite)",
					"entity_type", entry.EntityType, "entity_id", entry.EntityID)
				w.Status.RecordConflict(entry.EntityType, entry.EntityID, "lww")

				// Retry without version header → unconditional overwrite
				pushErr = w.dispatchChange(entry.EntityType, entry.EntityID, entry.Operation, entry.Payload, "")
			}
		}

		if pushErr != nil {
			newRetry := entry.RetryCount + 1
			retryAt := now.UTC().Format(time.RFC3339)

			if newRetry >= maxRetries {
				w.db.Exec(
					`UPDATE change_log SET status = 2, retry_count = ?, last_retry_at = ?, sync_error = ? WHERE id = ?`,
					newRetry, retryAt, pushErr.Error(), entry.ID,
				)
				slog.Warn("sync: change dead-lettered",
					"id", entry.ID, "type", entry.EntityType, "op", entry.Operation, "error", pushErr)
				w.Status.Log("push", "change_log", "error",
					fmt.Sprintf("dead-lettered %s/%s: %s", entry.EntityType, entry.Operation, pushErr.Error()), 0)
			} else {
				w.db.Exec(
					`UPDATE change_log SET retry_count = ?, last_retry_at = ?, sync_error = ? WHERE id = ?`,
					newRetry, retryAt, pushErr.Error(), entry.ID,
				)
				slog.Warn("sync: change push failed, will retry",
					"id", entry.ID, "type", entry.EntityType, "retry", newRetry, "error", pushErr)
			}
			failed++
			continue
		}

		// Success: mark as synced
		w.db.Exec(`UPDATE change_log SET status = 1 WHERE id = ?`, entry.ID)
		pushed++
	}

	if pushed > 0 || failed > 0 {
		w.Status.Log("push", "change_log", "ok",
			fmt.Sprintf("%d pushed, %d failed", pushed, failed), pushed)
		slog.Info("sync: pushed changes", "pushed", pushed, "failed", failed)
	}

	return nil
}

// dispatchChange routes a change_log entry to the correct API call.
// baseVersion is passed for optimistic concurrency (empty = no version check).
func (w *Worker) dispatchChange(entityType, entityID, operation, payloadJSON, baseVersion string) error {
	switch entityType {
	case "category":
		return w.dispatchCategoryChange(entityID, operation, payloadJSON, baseVersion)
	case "menu_item":
		return w.dispatchMenuItemChange(entityID, operation, payloadJSON, baseVersion)
	case "inventory_item":
		return w.dispatchInventoryItemChange(entityID, operation, payloadJSON, baseVersion)
	case "stock_adjustment":
		return w.dispatchStockAdjustmentChange(entityID, payloadJSON)
	case "recipe":
		return w.dispatchRecipeChange(entityID, payloadJSON)
	case "table":
		return w.dispatchTableChange(entityID, operation, payloadJSON)
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}
}

func (w *Worker) dispatchCategoryChange(entityID, operation, payloadJSON, baseVersion string) error {
	switch operation {
	case "create":
		var p CategoryPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return fmt.Errorf("unmarshal category payload: %w", err)
		}
		p.ID = entityID // Send client-generated UUID
		_, err := w.client.CreateCategory(p)
		return err
	case "update":
		var p CategoryPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return fmt.Errorf("unmarshal category payload: %w", err)
		}
		if baseVersion != "" {
			_, err := w.client.UpdateCategoryVersioned(entityID, p, baseVersion)
			return err
		}
		_, err := w.client.UpdateCategory(entityID, p)
		return err
	case "delete":
		return w.client.DeleteCategory(entityID)
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

func (w *Worker) dispatchMenuItemChange(entityID, operation, payloadJSON, baseVersion string) error {
	switch operation {
	case "create":
		var p MenuItemPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return fmt.Errorf("unmarshal menu item payload: %w", err)
		}
		p.ID = entityID
		_, err := w.client.CreateMenuItem(p)
		return err
	case "update":
		var p MenuItemPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return fmt.Errorf("unmarshal menu item payload: %w", err)
		}
		if baseVersion != "" {
			_, err := w.client.UpdateMenuItemVersioned(entityID, p, baseVersion)
			return err
		}
		_, err := w.client.UpdateMenuItem(entityID, p)
		return err
	case "delete":
		return w.client.DeleteMenuItem(entityID)
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

func (w *Worker) dispatchInventoryItemChange(entityID, operation, payloadJSON, baseVersion string) error {
	switch operation {
	case "create":
		var p InventoryPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return fmt.Errorf("unmarshal inventory payload: %w", err)
		}
		p.ID = entityID
		_, err := w.client.CreateInventoryItem(p)
		return err
	case "update":
		var p InventoryPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return fmt.Errorf("unmarshal inventory payload: %w", err)
		}
		if baseVersion != "" {
			_, err := w.client.UpdateInventoryItemVersioned(entityID, p, baseVersion)
			return err
		}
		_, err := w.client.UpdateInventoryItem(entityID, p)
		return err
	case "delete":
		return w.client.DeleteInventoryItem(entityID)
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

func (w *Worker) dispatchStockAdjustmentChange(entityID, payloadJSON string) error {
	var p StockAdjustPayload
	if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
		return fmt.Errorf("unmarshal stock adjust payload: %w", err)
	}
	p.ID = entityID
	return w.client.AdjustStock(p)
}

func (w *Worker) dispatchRecipeChange(menuItemID, payloadJSON string) error {
	var wrapper struct {
		MenuItemID  string                    `json:"menu_item_id"`
		Ingredients []RecipeIngredientPayload `json:"ingredients"`
	}
	if err := json.Unmarshal([]byte(payloadJSON), &wrapper); err != nil {
		return fmt.Errorf("unmarshal recipe payload: %w", err)
	}
	return w.client.SetRecipe(menuItemID, wrapper.Ingredients)
}

func (w *Worker) dispatchTableChange(entityID, operation, payloadJSON string) error {
	switch operation {
	case "create":
		var p TablePayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return fmt.Errorf("unmarshal table payload: %w", err)
		}
		p.ID = entityID
		result, err := w.client.CreateTable(p)
		if err != nil {
			return err
		}
		// Reconcile: update local row with server-generated token
		if result != nil && result.Token != "" {
			w.db.Exec(`UPDATE tables SET token = ?, synced = 1 WHERE id = ?`, result.Token, entityID)
			slog.Info("sync: table token reconciled", "id", entityID, "token", result.Token)
		}
		return nil
	case "delete":
		return w.client.DeleteTable(entityID)
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

// GetFailedChanges returns all dead-lettered change_log entries.
type FailedChangeInfo struct {
	ID          int    `json:"id" db:"id"`
	EntityType  string `json:"entity_type" db:"entity_type"`
	EntityID    string `json:"entity_id" db:"entity_id"`
	Operation   string `json:"operation" db:"operation"`
	RetryCount  int    `json:"retry_count" db:"retry_count"`
	SyncError   string `json:"sync_error" db:"sync_error"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	LastRetryAt string `json:"last_retry_at" db:"last_retry_at"`
}

func (w *Worker) GetFailedChanges() ([]FailedChangeInfo, error) {
	var changes []FailedChangeInfo
	err := w.db.Select(&changes,
		`SELECT id, entity_type, entity_id, operation, retry_count, sync_error, created_at, last_retry_at
		 FROM change_log WHERE status = 2
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("fetch failed changes: %w", err)
	}
	return changes, nil
}

// RetryFailedChanges resets all dead-lettered changes (status=2) back to pending (status=0).
func (w *Worker) RetryFailedChanges() (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	result, err := w.db.Exec(
		`UPDATE change_log SET status = 0, retry_count = 0, last_retry_at = '', sync_error = '' WHERE status = 2`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to retry failed changes: %w", err)
	}

	affected, _ := result.RowsAffected()
	count := int(affected)
	w.Status.Log("retry", "change_log", "ok", fmt.Sprintf("reset %d dead-lettered changes", count), count)
	slog.Info("sync: retried failed changes", "count", count)
	return count, nil
}

// GetChangeLogStats returns pending/synced/failed counts for the sync dashboard.
type ChangeLogStats struct {
	Pending int `json:"pending" db:"pending"`
	Synced  int `json:"synced" db:"synced"`
	Failed  int `json:"failed" db:"failed"`
}

func (w *Worker) GetChangeLogStats() ChangeLogStats {
	var stats ChangeLogStats
	w.db.Get(&stats.Pending, `SELECT COUNT(*) FROM change_log WHERE status = 0`)
	w.db.Get(&stats.Synced, `SELECT COUNT(*) FROM change_log WHERE status = 1`)
	w.db.Get(&stats.Failed, `SELECT COUNT(*) FROM change_log WHERE status = 2`)
	return stats
}

// maxRetries is the maximum number of push attempts before dead-lettering an order.
const maxRetries = 10

// backoffDuration returns the backoff delay for a given retry count.
// 5s → 10s → 20s → 40s → 80s → 160s → 300s (capped at 5 min).
func backoffDuration(retryCount int) time.Duration {
	d := time.Duration(5<<uint(retryCount)) * time.Second
	if d > 5*time.Minute {
		d = 5 * time.Minute
	}
	return d
}

// PushOrders sends unsynced local orders to the central API with retry logic.
// Orders with synced=0 are candidates. Exponential backoff is applied per retry.
// After maxRetries failures, orders are dead-lettered (synced=2).
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
		RetryCount    int    `db:"retry_count"`
		LastRetryAt   string `db:"last_retry_at"`
	}

	var orders []orderRow
	err := w.db.Select(&orders,
		`SELECT id, order_number, source, table_number, total, payment_method,
		        created_at, retry_count, last_retry_at
		 FROM orders WHERE synced = 0
		 ORDER BY created_at ASC`)
	if err != nil {
		return fmt.Errorf("fetch unsynced orders: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	now := time.Now()
	pushed := 0
	skipped := 0
	failed := 0

	for _, order := range orders {
		// Check backoff: skip if not enough time has elapsed since last retry
		if order.RetryCount > 0 && order.LastRetryAt != "" {
			lastRetry, parseErr := time.Parse(time.RFC3339, order.LastRetryAt)
			if parseErr == nil {
				requiredBackoff := backoffDuration(order.RetryCount)
				if now.Sub(lastRetry) < requiredBackoff {
					skipped++
					continue
				}
			}
		}

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
		err := w.db.Select(&items,
			`SELECT id, menu_item_id, quantity, unit_price, line_total, name_ar_snapshot
			 FROM order_items WHERE order_id = ?`, order.ID)
		if err != nil {
			slog.Warn("sync: failed to fetch order items", "order_id", order.ID, "error", err)
			continue
		}

		// Format CreatedAt to RFC3339 for the central API
		createdAtStr := order.CreatedAt
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", order.CreatedAt, time.Local); err == nil {
			createdAtStr = t.Format(time.RFC3339)
		}

		// Build push payload
		payload := OrderPushPayload{
			ID:            order.ID,
			Source:        order.Source,
			TableNumber:   order.TableNumber,
			Total:         order.Total,
			PaymentMethod: order.PaymentMethod,
			CreatedAt:     createdAtStr,
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
		result, pushErr := w.client.PushOrder(payload)
		if pushErr != nil {
			newRetry := order.RetryCount + 1
			retryAt := now.UTC().Format(time.RFC3339)

			if newRetry >= maxRetries {
				// Dead letter: mark as permanently failed
				w.db.Exec(
					`UPDATE orders SET synced = 2, retry_count = ?, last_retry_at = ?, sync_error = ? WHERE id = ?`,
					newRetry, retryAt, pushErr.Error(), order.ID,
				)
				slog.Warn("sync: order dead-lettered after max retries",
					"order_id", order.ID, "retries", newRetry, "error", pushErr)
				w.Status.Log("push", "orders", "error",
					fmt.Sprintf("dead-lettered: %s", pushErr.Error()), 0)
			} else {
				// Record failure for retry with backoff
				w.db.Exec(
					`UPDATE orders SET retry_count = ?, last_retry_at = ?, sync_error = ? WHERE id = ?`,
					newRetry, retryAt, pushErr.Error(), order.ID,
				)
				slog.Warn("sync: order push failed, will retry",
					"order_id", order.ID, "retry", newRetry, "next_backoff", backoffDuration(newRetry), "error", pushErr)
				w.Status.Log("retry", "orders", "error",
					fmt.Sprintf("retry %d/%d: %s", newRetry, maxRetries, pushErr.Error()), 0)
			}
			failed++
			continue
		}

		// Success: mark as synced and update server-assigned order number
		serverOrderNumber := result.OrderNumber
		if serverOrderNumber == "" {
			serverOrderNumber = order.OrderNumber
		}

		_, err = w.db.Exec(
			`UPDATE orders SET synced = 1, order_number = ?, retry_count = 0, sync_error = '' WHERE id = ?`,
			serverOrderNumber, order.ID,
		)
		if err != nil {
			slog.Warn("sync: failed to mark order as synced", "order_id", order.ID, "error", err)
		} else {
			slog.Info("sync: pushed order", "order_id", order.ID, "server_order_number", serverOrderNumber)
		}
		pushed++
	}

	if pushed > 0 || failed > 0 {
		w.Status.Log("push", "orders", "ok",
			fmt.Sprintf("%d pushed, %d failed, %d skipped (backoff)", pushed, failed, skipped), pushed)
	}

	return nil
}

// FailedOrderInfo holds details about a dead-lettered order for the dashboard.
type FailedOrderInfo struct {
	ID            string `json:"id" db:"id"`
	OrderNumber   string `json:"order_number" db:"order_number"`
	Total         int64  `json:"total" db:"total"`
	RetryCount    int    `json:"retry_count" db:"retry_count"`
	SyncError     string `json:"sync_error" db:"sync_error"`
	CreatedAt     string `json:"created_at" db:"created_at"`
	LastRetryAt   string `json:"last_retry_at" db:"last_retry_at"`
}

// ResetSyncMeta clears all per-table sync timestamps, forcing a full re-sync on the next cycle.
func (w *Worker) ResetSyncMeta() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := w.db.Exec(`DELETE FROM sync_meta`)
	if err != nil {
		return fmt.Errorf("failed to reset sync_meta: %w", err)
	}

	w.Status.Log("reset", "sync_meta", "ok", "all timestamps cleared", 0)
	slog.Info("sync: sync_meta reset — next cycle will be a full sync")
	return nil
}

// RetryFailedOrders resets all dead-lettered orders (synced=2) back to pending (synced=0).
func (w *Worker) RetryFailedOrders() (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	result, err := w.db.Exec(
		`UPDATE orders SET synced = 0, retry_count = 0, last_retry_at = '', sync_error = '' WHERE synced = 2`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to retry failed orders: %w", err)
	}

	affected, _ := result.RowsAffected()
	count := int(affected)

	w.Status.Log("retry", "orders", "ok", fmt.Sprintf("reset %d dead-lettered orders", count), count)
	slog.Info("sync: retried failed orders", "count", count)
	return count, nil
}

// GetFailedOrders returns all dead-lettered orders with their error details.
func (w *Worker) GetFailedOrders() ([]FailedOrderInfo, error) {
	var orders []FailedOrderInfo
	err := w.db.Select(&orders,
		`SELECT id, order_number, total, retry_count, sync_error, created_at, last_retry_at
		 FROM orders WHERE synced = 2
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("fetch failed orders: %w", err)
	}
	return orders, nil
}
