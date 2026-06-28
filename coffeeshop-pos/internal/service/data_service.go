package service

import (
	"coffeeshop-pos/internal/model"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/jmoiron/sqlx"
)

var imageCache sync.Map

// DataService is a Wails-bound service that exposes local SQLite data
// to the frontend. All methods are exported and callable from JavaScript.
type DataService struct {
	db *sqlx.DB
}

// NewDataService creates a new DataService.
func NewDataService(db *sqlx.DB) *DataService {
	return &DataService{db: db}
}

// GetCategories returns all active categories from local SQLite.
func (s *DataService) GetCategories() ([]model.Category, error) {
	var categories []model.Category
	err := s.db.Select(&categories,
		`SELECT id, name_ar, sort_order, is_active
		 FROM categories WHERE is_active = 1
		 ORDER BY sort_order ASC, name_ar ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	return categories, nil
}

// GetMenuItems returns active menu items from local SQLite.
// If categoryID is non-empty, filters by category.
func (s *DataService) GetMenuItems(categoryID string) ([]model.MenuItemWithCategory, error) {
	var items []model.MenuItemWithCategory

	if categoryID != "" {
		err := s.db.Select(&items,
			`SELECT mi.id, mi.category_id, mi.name_ar, mi.price,
			        mi.cost_calc_method, mi.manual_cost_price, mi.cached_auto_cost,
			        mi.image_path, mi.is_active,
			        c.name_ar AS category_name_ar
			 FROM menu_items mi
			 JOIN categories c ON c.id = mi.category_id
			 WHERE mi.is_active = 1 AND mi.category_id = ?
			 ORDER BY mi.name_ar ASC`,
			categoryID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch menu items: %w", err)
		}
		return items, nil
	}

	err := s.db.Select(&items,
		`SELECT mi.id, mi.category_id, mi.name_ar, mi.price,
		        mi.cost_calc_method, mi.manual_cost_price, mi.cached_auto_cost,
		        mi.image_path, mi.is_active,
		        c.name_ar AS category_name_ar
		 FROM menu_items mi
		 JOIN categories c ON c.id = mi.category_id
		 WHERE mi.is_active = 1
		 ORDER BY mi.name_ar ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch menu items: %w", err)
	}
	return items, nil
}

// GetInventoryItems returns all active inventory items from local SQLite.
func (s *DataService) GetInventoryItems() ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	err := s.db.Select(&items,
		`SELECT id, name_ar, base_unit_ar, stock_qty, low_stock_threshold, unit_cost, is_active
		 FROM inventory_items WHERE is_active = 1
		 ORDER BY name_ar ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory items: %w", err)
	}
	return items, nil
}

// GetImageDataURI fetches an image URL and returns it as a base64 data URI.
// Results are cached in memory so each URL is fetched only once.
func (s *DataService) GetImageDataURI(imageURL string) (string, error) {
	if imageURL == "" {
		return "", nil
	}

	if cached, ok := imageCache.Load(imageURL); ok {
		return cached.(string), nil
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		slog.Warn("failed to fetch image", "url", imageURL, "error", err)
		return "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image fetch returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}

	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(body))
	imageCache.Store(imageURL, dataURI)

	return dataURI, nil
}

// GetLastSyncTime returns the last successful sync timestamp from sync_meta.
func (s *DataService) GetLastSyncTime() string {
	var lastSync string
	err := s.db.Get(&lastSync, `SELECT last_synced_at FROM sync_meta WHERE table_name = 'all'`)
	if err != nil {
		return ""
	}
	return lastSync
}
