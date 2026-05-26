package service

import (
	"coffeeshop-pos/internal/model"
	"encoding/base64"
	"fmt"
	"log/slog"

	posSync "coffeeshop-pos/internal/sync"
	qrcode "github.com/skip2/go-qrcode"
)

// ManagementService is a Wails-bound service that proxies management
// operations to the central API. After each write, it triggers a local
// sync to refresh the SQLite cache.
type ManagementService struct {
	apiClient  *posSync.APIClient
	syncWorker *posSync.Worker
}

// NewManagementService creates a new ManagementService.
func NewManagementService(apiClient *posSync.APIClient, syncWorker *posSync.Worker) *ManagementService {
	return &ManagementService{
		apiClient:  apiClient,
		syncWorker: syncWorker,
	}
}

// refreshLocal triggers an immediate pull to refresh local SQLite data.
func (s *ManagementService) refreshLocal() {
	if err := s.syncWorker.PullAll(); err != nil {
		slog.Warn("management: failed to refresh local data after write", "error", err)
	}
}

// ── Inventory ──

// CreateInventoryItem creates a new raw material via the API.
func (s *ManagementService) CreateInventoryItem(nameAr, baseUnitAr string, stockQty, lowStockThreshold int, unitCost int64) (*model.InventoryItem, error) {
	result, err := s.apiClient.CreateInventoryItem(posSync.InventoryPayload{
		NameAr:            nameAr,
		BaseUnitAr:        baseUnitAr,
		StockQty:          stockQty,
		LowStockThreshold: lowStockThreshold,
		UnitCost:          unitCost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}
	slog.Info("management: created inventory item", "name", nameAr)
	s.refreshLocal()
	return result, nil
}

// UpdateInventoryItem updates a raw material via the API.
func (s *ManagementService) UpdateInventoryItem(id, nameAr, baseUnitAr string, stockQty, lowStockThreshold int, unitCost int64) (*model.InventoryItem, error) {
	result, err := s.apiClient.UpdateInventoryItem(id, posSync.InventoryPayload{
		NameAr:            nameAr,
		BaseUnitAr:        baseUnitAr,
		StockQty:          stockQty,
		LowStockThreshold: lowStockThreshold,
		UnitCost:          unitCost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}
	slog.Info("management: updated inventory item", "id", id)
	s.refreshLocal()
	return result, nil
}

// DeleteInventoryItem soft-deletes a raw material via the API.
func (s *ManagementService) DeleteInventoryItem(id string) error {
	if err := s.apiClient.DeleteInventoryItem(id); err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}
	slog.Info("management: deleted inventory item", "id", id)
	s.refreshLocal()
	return nil
}

// AdjustStock adjusts stock for a raw material via the API.
func (s *ManagementService) AdjustStock(inventoryItemID string, delta int, reason string) error {
	if err := s.apiClient.AdjustStock(posSync.StockAdjustPayload{
		InventoryItemID: inventoryItemID,
		Delta:           delta,
		Reason:          reason,
	}); err != nil {
		return fmt.Errorf("failed to adjust stock: %w", err)
	}
	slog.Info("management: adjusted stock", "item_id", inventoryItemID, "delta", delta, "reason", reason)
	s.refreshLocal()
	return nil
}

// ── Recipes ──

// RecipeIngredientInput is the frontend input for a single recipe ingredient.
type RecipeIngredientInput struct {
	InventoryItemID string `json:"inventory_item_id"`
	Quantity        int    `json:"quantity"`
}

// GetRecipe fetches the recipe for a menu item from the API.
func (s *ManagementService) GetRecipe(menuItemID string) ([]model.RecipeIngredientWithDetails, error) {
	result, err := s.apiClient.GetRecipe(menuItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}
	return result, nil
}

// SetRecipe sets/updates the recipe for a menu item via the API.
func (s *ManagementService) SetRecipe(menuItemID string, ingredients []RecipeIngredientInput) error {
	payloads := make([]posSync.RecipeIngredientPayload, len(ingredients))
	for i, ing := range ingredients {
		payloads[i] = posSync.RecipeIngredientPayload{
			InventoryItemID: ing.InventoryItemID,
			Quantity:        ing.Quantity,
		}
	}

	if err := s.apiClient.SetRecipe(menuItemID, payloads); err != nil {
		return fmt.Errorf("failed to set recipe: %w", err)
	}
	slog.Info("management: set recipe", "menu_item_id", menuItemID, "ingredients", len(ingredients))
	s.refreshLocal()
	return nil
}

// ── Menu Items ──

// CreateMenuItem creates a new menu item via the API.
func (s *ManagementService) CreateMenuItem(categoryID, nameAr string, price int64, costCalcMethod string, manualCostPrice int64, imagePath string) (*model.MenuItemWithCategory, error) {
	result, err := s.apiClient.CreateMenuItem(posSync.MenuItemPayload{
		CategoryID:      categoryID,
		NameAr:          nameAr,
		Price:           price,
		CostCalcMethod:  costCalcMethod,
		ManualCostPrice: manualCostPrice,
		ImagePath:       imagePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create menu item: %w", err)
	}
	slog.Info("management: created menu item", "name", nameAr)
	s.refreshLocal()
	return result, nil
}

// UpdateMenuItem updates a menu item via the API.
func (s *ManagementService) UpdateMenuItem(id, categoryID, nameAr string, price int64, costCalcMethod string, manualCostPrice int64, imagePath string) (*model.MenuItemWithCategory, error) {
	result, err := s.apiClient.UpdateMenuItem(id, posSync.MenuItemPayload{
		CategoryID:      categoryID,
		NameAr:          nameAr,
		Price:           price,
		CostCalcMethod:  costCalcMethod,
		ManualCostPrice: manualCostPrice,
		ImagePath:       imagePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update menu item: %w", err)
	}
	slog.Info("management: updated menu item", "id", id)
	s.refreshLocal()
	return result, nil
}

// DeleteMenuItem soft-deletes a menu item via the API.
func (s *ManagementService) DeleteMenuItem(id string) error {
	if err := s.apiClient.DeleteMenuItem(id); err != nil {
		return fmt.Errorf("failed to delete menu item: %w", err)
	}
	slog.Info("management: deleted menu item", "id", id)
	s.refreshLocal()
	return nil
}

// ── Categories ──

// CreateCategory creates a new category via the API.
func (s *ManagementService) CreateCategory(nameAr string, sortOrder int) (*model.Category, error) {
	result, err := s.apiClient.CreateCategory(posSync.CategoryPayload{
		NameAr:    nameAr,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	slog.Info("management: created category", "name", nameAr)
	s.refreshLocal()
	return result, nil
}

// UpdateCategory updates a category via the API.
func (s *ManagementService) UpdateCategory(id, nameAr string, sortOrder int) (*model.Category, error) {
	result, err := s.apiClient.UpdateCategory(id, posSync.CategoryPayload{
		NameAr:    nameAr,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	slog.Info("management: updated category", "id", id)
	s.refreshLocal()
	return result, nil
}

// DeleteCategory deletes a category via the API.
func (s *ManagementService) DeleteCategory(id string) error {
	if err := s.apiClient.DeleteCategory(id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	slog.Info("management: deleted category", "id", id)
	s.refreshLocal()
	return nil
}

// TriggerSync manually triggers a full sync (pull + push).
func (s *ManagementService) TriggerSync() error {
	slog.Info("management: manual sync triggered")
	if err := s.syncWorker.PullAll(); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}
	if err := s.syncWorker.PushOrders(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}
	return nil
}

// ── Tables ──

// ListTables fetches all restaurant tables from the API.
func (s *ManagementService) ListTables() ([]model.Table, error) {
	tables, err := s.apiClient.ListTables()
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	return tables, nil
}

// CreateTable creates a new restaurant table via the API.
func (s *ManagementService) CreateTable(number string) (*model.Table, error) {
	if number == "" {
		return nil, fmt.Errorf("table number is required")
	}
	result, err := s.apiClient.CreateTable(posSync.TablePayload{Number: number})
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	slog.Info("management: created table", "number", number)
	return result, nil
}

// DeleteTable deletes a restaurant table via the API.
func (s *ManagementService) DeleteTable(id string) error {
	if err := s.apiClient.DeleteTable(id); err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}
	slog.Info("management: deleted table", "id", id)
	return nil
}

// GetTableQRCode generates a QR code PNG image as a base64 data URL for a table's menu link.
func (s *ManagementService) GetTableQRCode(tableToken string, menuBaseURL string) (string, error) {
	if tableToken == "" {
		return "", fmt.Errorf("table token is required")
	}
	if menuBaseURL == "" {
		menuBaseURL = "http://localhost:5173"
	}

	menuURL := menuBaseURL + "?token=" + tableToken

	png, err := qrcode.Encode(menuURL, qrcode.Medium, 512)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	return dataURL, nil
}

// UploadMenuItemImage opens the API upload endpoint with a file path from the frontend.
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
