package service

import (
	"coffeeshop-pos/internal/model"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	posSync "coffeeshop-pos/internal/sync"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/jmoiron/sqlx"
)

// ManagementService is a Wails-bound service for admin operations.
// All writes go to local SQLite first, then get queued for async push to the API.
type ManagementService struct {
	db         *sqlx.DB
	apiClient  *posSync.APIClient
	syncWorker *posSync.Worker
}

// NewManagementService creates a new ManagementService.
func NewManagementService(db *sqlx.DB, apiClient *posSync.APIClient, syncWorker *posSync.Worker) *ManagementService {
	return &ManagementService{
		db:         db,
		apiClient:  apiClient,
		syncWorker: syncWorker,
	}
}

// queueChange inserts a change_log entry for async push to the API.
// baseVersion is the entity's updated_at at the time of edit (for optimistic concurrency).
// Pass "" for create operations or when version tracking is not applicable.
func (s *ManagementService) queueChange(entityType, entityID, operation string, payload any, baseVersion string) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		slog.Error("management: failed to marshal change_log payload", "error", err)
		return
	}
	_, err = s.db.Exec(
		`INSERT INTO change_log (entity_type, entity_id, operation, payload, base_version) VALUES (?, ?, ?, ?, ?)`,
		entityType, entityID, operation, string(payloadJSON), baseVersion,
	)
	if err != nil {
		slog.Error("management: failed to insert change_log entry", "error", err, "entity_type", entityType)
	}
}

// getEntityVersion reads the current updated_at of an entity from SQLite.
// Used to capture the base version before local updates for optimistic concurrency.
func (s *ManagementService) getEntityVersion(table, id string) string {
	var version string
	s.db.Get(&version, `SELECT COALESCE(updated_at, '') FROM `+table+` WHERE id = ?`, id)
	return version
}

// ── Inventory ──

// CreateInventoryItem creates a new raw material locally and queues for API push.
func (s *ManagementService) CreateInventoryItem(nameAr, baseUnitAr string, stockQty, lowStockThreshold int, unitCost int64) (*model.InventoryItem, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`INSERT INTO inventory_items (id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 1, ?)`,
		id, nameAr, baseUnitAr, stockQty, lowStockThreshold, unitCost, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory item locally: %w", err)
	}

	s.queueChange("inventory_item", id, "create", posSync.InventoryPayload{
		NameAr:            nameAr,
		BaseUnitAr:        baseUnitAr,
		StockQty:          stockQty,
		LowStockThreshold: lowStockThreshold,
		UnitCost:          unitCost,
	}, "")

	slog.Info("management: created inventory item locally", "id", id, "name", nameAr)
	return &model.InventoryItem{
		ID: id, NameAr: nameAr, BaseUnitAr: baseUnitAr,
		StockQty: stockQty, LowStockThreshold: lowStockThreshold,
		UnitCost: unitCost, IsActive: true, UpdatedAt: now,
	}, nil
}

// UpdateInventoryItem updates a raw material locally and queues for API push.
func (s *ManagementService) UpdateInventoryItem(id, nameAr, baseUnitAr string, stockQty, lowStockThreshold int, unitCost int64) (*model.InventoryItem, error) {
	baseVersion := s.getEntityVersion("inventory_items", id)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`UPDATE inventory_items SET name_ar = ?, base_unit_ar = ?, stock_qty = ?, low_stock_threshold = ?, unit_cost = ?, updated_at = ? WHERE id = ?`,
		nameAr, baseUnitAr, stockQty, lowStockThreshold, unitCost, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory item locally: %w", err)
	}

	s.queueChange("inventory_item", id, "update", posSync.InventoryPayload{
		NameAr:            nameAr,
		BaseUnitAr:        baseUnitAr,
		StockQty:          stockQty,
		LowStockThreshold: lowStockThreshold,
		UnitCost:          unitCost,
	}, baseVersion)

	slog.Info("management: updated inventory item locally", "id", id)
	return &model.InventoryItem{
		ID: id, NameAr: nameAr, BaseUnitAr: baseUnitAr,
		StockQty: stockQty, LowStockThreshold: lowStockThreshold,
		UnitCost: unitCost, IsActive: true, UpdatedAt: now,
	}, nil
}

// DeleteInventoryItem soft-deletes a raw material locally and queues for API push.
func (s *ManagementService) DeleteInventoryItem(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`UPDATE inventory_items SET is_active = 0, updated_at = ? WHERE id = ?`, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete inventory item locally: %w", err)
	}

	s.queueChange("inventory_item", id, "delete", nil, "")
	slog.Info("management: deleted inventory item locally", "id", id)
	return nil
}

// AdjustStock adjusts stock for a raw material locally and queues for API push.
func (s *ManagementService) AdjustStock(inventoryItemID string, delta int, reason string) error {
	adjID := uuid.New().String()

	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert local adjustment
	_, err = tx.Exec(
		`INSERT INTO stock_adjustments (id, inventory_item_id, delta, reason_ar, synced) VALUES (?, ?, ?, ?, 0)`,
		adjID, inventoryItemID, delta, reason,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock adjustment: %w", err)
	}

	// Update stock qty
	_, err = tx.Exec(
		`UPDATE inventory_items SET stock_qty = stock_qty + ?, updated_at = ? WHERE id = ?`,
		delta, time.Now().UTC().Format(time.RFC3339), inventoryItemID,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock qty: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	s.queueChange("stock_adjustment", adjID, "create", posSync.StockAdjustPayload{
		InventoryItemID: inventoryItemID,
		Delta:           delta,
		Reason:          reason,
	}, "")

	slog.Info("management: adjusted stock locally", "item_id", inventoryItemID, "delta", delta)
	return nil
}

// ── Recipes ──

// RecipeIngredientInput is the frontend input for a single recipe ingredient.
type RecipeIngredientInput struct {
	InventoryItemID string `json:"inventory_item_id"`
	Quantity        int    `json:"quantity"`
}

// GetRecipe fetches the recipe for a menu item from local SQLite.
func (s *ManagementService) GetRecipe(menuItemID string) ([]model.RecipeIngredientWithDetails, error) {
	var ingredients []model.RecipeIngredientWithDetails
	err := s.db.Select(&ingredients,
		`SELECT ri.id, ri.menu_item_id, ri.inventory_item_id, ri.quantity,
		        ii.name_ar AS inventory_name_ar, ii.base_unit_ar, ii.unit_cost
		 FROM recipe_ingredients ri
		 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		 WHERE ri.menu_item_id = ?`,
		menuItemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}
	return ingredients, nil
}

// SetRecipe sets/updates the recipe for a menu item locally and queues for API push.
func (s *ManagementService) SetRecipe(menuItemID string, ingredients []RecipeIngredientInput) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing recipe
	tx.Exec(`DELETE FROM recipe_ingredients WHERE menu_item_id = ?`, menuItemID)

	now := time.Now().UTC().Format(time.RFC3339)
	for _, ing := range ingredients {
		id := uuid.New().String()
		_, err := tx.Exec(
			`INSERT INTO recipe_ingredients (id, menu_item_id, inventory_item_id, quantity, updated_at)
			 VALUES (?, ?, ?, ?, ?)`,
			id, menuItemID, ing.InventoryItemID, ing.Quantity, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert recipe ingredient: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Queue the full recipe as a single change_log entry
	payloads := make([]posSync.RecipeIngredientPayload, len(ingredients))
	for i, ing := range ingredients {
		payloads[i] = posSync.RecipeIngredientPayload{
			InventoryItemID: ing.InventoryItemID,
			Quantity:        ing.Quantity,
		}
	}
	s.queueChange("recipe", menuItemID, "update", map[string]any{
		"menu_item_id": menuItemID,
		"ingredients":  payloads,
	}, "")

	slog.Info("management: set recipe locally", "menu_item_id", menuItemID, "ingredients", len(ingredients))
	return nil
}

// ── Menu Items ──

// CreateMenuItem creates a new menu item locally and queues for API push.
func (s *ManagementService) CreateMenuItem(categoryID, nameAr string, price int64, costCalcMethod string, manualCostPrice int64, imagePath string) (*model.MenuItemWithCategory, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	if costCalcMethod == "" {
		costCalcMethod = "auto"
	}

	_, err := s.db.Exec(
		`INSERT INTO menu_items (id, category_id, name_ar, price, cost_calc_method, manual_cost_price, cached_auto_cost, image_path, is_active, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 0, ?, 1, ?)`,
		id, categoryID, nameAr, price, costCalcMethod, manualCostPrice, imagePath, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu item locally: %w", err)
	}

	s.queueChange("menu_item", id, "create", posSync.MenuItemPayload{
		CategoryID:      categoryID,
		NameAr:          nameAr,
		Price:           price,
		CostCalcMethod:  costCalcMethod,
		ManualCostPrice: manualCostPrice,
		ImagePath:       imagePath,
	}, "")

	// Fetch category name for response
	var categoryNameAr string
	s.db.Get(&categoryNameAr, `SELECT name_ar FROM categories WHERE id = ?`, categoryID)

	slog.Info("management: created menu item locally", "id", id, "name", nameAr)
	return &model.MenuItemWithCategory{
		MenuItem: model.MenuItem{
			ID: id, CategoryID: categoryID, NameAr: nameAr, Price: price,
			CostCalcMethod: costCalcMethod, ManualCostPrice: manualCostPrice,
			ImagePath: imagePath, IsActive: true, UpdatedAt: now,
		},
		CategoryNameAr: categoryNameAr,
	}, nil
}

// UpdateMenuItem updates a menu item locally and queues for API push.
func (s *ManagementService) UpdateMenuItem(id, categoryID, nameAr string, price int64, costCalcMethod string, manualCostPrice int64, imagePath string) (*model.MenuItemWithCategory, error) {
	baseVersion := s.getEntityVersion("menu_items", id)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`UPDATE menu_items SET category_id = ?, name_ar = ?, price = ?, cost_calc_method = ?, manual_cost_price = ?, image_path = ?, updated_at = ? WHERE id = ?`,
		categoryID, nameAr, price, costCalcMethod, manualCostPrice, imagePath, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update menu item locally: %w", err)
	}

	s.queueChange("menu_item", id, "update", posSync.MenuItemPayload{
		CategoryID:      categoryID,
		NameAr:          nameAr,
		Price:           price,
		CostCalcMethod:  costCalcMethod,
		ManualCostPrice: manualCostPrice,
		ImagePath:       imagePath,
	}, baseVersion)

	var categoryNameAr string
	s.db.Get(&categoryNameAr, `SELECT name_ar FROM categories WHERE id = ?`, categoryID)

	slog.Info("management: updated menu item locally", "id", id)
	return &model.MenuItemWithCategory{
		MenuItem: model.MenuItem{
			ID: id, CategoryID: categoryID, NameAr: nameAr, Price: price,
			CostCalcMethod: costCalcMethod, ManualCostPrice: manualCostPrice,
			ImagePath: imagePath, IsActive: true, UpdatedAt: now,
		},
		CategoryNameAr: categoryNameAr,
	}, nil
}

// DeleteMenuItem soft-deletes a menu item locally and queues for API push.
func (s *ManagementService) DeleteMenuItem(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`UPDATE menu_items SET is_active = 0, updated_at = ? WHERE id = ?`, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete menu item locally: %w", err)
	}

	s.queueChange("menu_item", id, "delete", nil, "")
	slog.Info("management: deleted menu item locally", "id", id)
	return nil
}

// ── Categories ──

// CreateCategory creates a new category locally and queues for API push.
func (s *ManagementService) CreateCategory(nameAr string, sortOrder int) (*model.Category, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`INSERT INTO categories (id, name_ar, sort_order, is_active, updated_at)
		 VALUES (?, ?, ?, 1, ?)`,
		id, nameAr, sortOrder, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create category locally: %w", err)
	}

	s.queueChange("category", id, "create", posSync.CategoryPayload{
		NameAr:    nameAr,
		SortOrder: sortOrder,
	}, "")

	slog.Info("management: created category locally", "id", id, "name", nameAr)
	return &model.Category{
		ID: id, NameAr: nameAr, SortOrder: sortOrder, IsActive: true, UpdatedAt: now,
	}, nil
}

// UpdateCategory updates a category locally and queues for API push.
func (s *ManagementService) UpdateCategory(id, nameAr string, sortOrder int) (*model.Category, error) {
	baseVersion := s.getEntityVersion("categories", id)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`UPDATE categories SET name_ar = ?, sort_order = ?, updated_at = ? WHERE id = ?`,
		nameAr, sortOrder, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update category locally: %w", err)
	}

	s.queueChange("category", id, "update", posSync.CategoryPayload{
		NameAr:    nameAr,
		SortOrder: sortOrder,
	}, baseVersion)

	slog.Info("management: updated category locally", "id", id)
	return &model.Category{
		ID: id, NameAr: nameAr, SortOrder: sortOrder, IsActive: true, UpdatedAt: now,
	}, nil
}

// DeleteCategory deletes a category locally and queues for API push.
func (s *ManagementService) DeleteCategory(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`UPDATE categories SET is_active = 0, updated_at = ? WHERE id = ?`, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete category locally: %w", err)
	}

	s.queueChange("category", id, "delete", nil, "")
	slog.Info("management: deleted category locally", "id", id)
	return nil
}

// TriggerSync manually triggers a full sync (pull + push).
func (s *ManagementService) TriggerSync() error {
	slog.Info("management: manual sync triggered")
	if err := s.syncWorker.PullAll(); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}
	if err := s.syncWorker.PushChangeLog(); err != nil {
		return fmt.Errorf("change push failed: %w", err)
	}
	if err := s.syncWorker.PushOrders(); err != nil {
		return fmt.Errorf("order push failed: %w", err)
	}
	return nil
}

// ── Tables ──

// ListTables fetches all tables from local SQLite.
func (s *ManagementService) ListTables() ([]model.Table, error) {
	var tables []model.Table
	err := s.db.Select(&tables,
		`SELECT id, number, token, is_active, synced, created_at FROM tables WHERE is_active = 1 ORDER BY number ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	return tables, nil
}

// CreateTable creates a new table locally with a temporary token and queues for API push.
func (s *ManagementService) CreateTable(number string) (*model.Table, error) {
	if number == "" {
		return nil, fmt.Errorf("table number is required")
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`INSERT INTO tables (id, number, token, is_active, synced, created_at) VALUES (?, ?, '', 1, 0, ?)`,
		id, number, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create table locally: %w", err)
	}

	s.queueChange("table", id, "create", posSync.TablePayload{Number: number}, "")

	slog.Info("management: created table locally (pending sync for token)", "id", id, "number", number)
	return &model.Table{
		ID: id, Number: number, Token: "", IsActive: true, Synced: 0, CreatedAt: now,
	}, nil
}

// DeleteTable soft-deletes a table locally and queues for API push.
func (s *ManagementService) DeleteTable(id string) error {
	_, err := s.db.Exec(`UPDATE tables SET is_active = 0 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete table locally: %w", err)
	}

	s.queueChange("table", id, "delete", nil, "")
	slog.Info("management: deleted table locally", "id", id)
	return nil
}

// GetTableQRCode generates a QR code for a table's web menu link.
// Returns an error message if the table hasn't been synced yet (no real token).
func (s *ManagementService) GetTableQRCode(tableToken string, menuBaseURL string) (string, error) {
	if tableToken == "" {
		return "", fmt.Errorf("هذه الطاولة لم تتم مزامنتها بعد. يرجى الانتظار حتى اكتمال المزامنة للحصول على رمز QR")
	}
	if menuBaseURL == "" {
		return "", fmt.Errorf("لم يتم تعيين رابط القائمة. اذهب إلى الإعدادات وأدخل رابط القائمة أولاً")
	}

	menuURL := menuBaseURL + "?token=" + tableToken

	png, err := qrcode.Encode(menuURL, qrcode.Medium, 512)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	return dataURL, nil
}

// UploadMenuItemImage uploads an image file to the API (network required).
func (s *ManagementService) UploadMenuItemImage(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path is required")
	}
	url, err := s.apiClient.UploadImage(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}
	slog.Info("management: uploaded menu item image", "url", url)
	return url, nil
}
