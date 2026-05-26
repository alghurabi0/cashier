package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"coffeeshop-pos/internal/model"
)

// APIClient is a thin HTTP client for the central coffeeshop-api.
type APIClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewAPIClient creates a new API client.
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetToken sets the JWT auth token for subsequent requests.
func (c *APIClient) SetToken(token string) {
	c.token = token
}

// SetBaseURL updates the base URL for the API client.
func (c *APIClient) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimRight(baseURL, "/")
}

// Login authenticates with the API and stores the returned token.
func (c *APIClient) Login(username, password string) error {
	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	resp, err := c.doRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.token = result.Data.Token
	return nil
}

// GetCategories fetches all active categories from the API.
func (c *APIClient) GetCategories() ([]model.Category, error) {
	return fetchList[model.Category](c, "/api/v1/categories")
}

// GetMenuItems fetches all active menu items from the API.
func (c *APIClient) GetMenuItems() ([]model.MenuItemWithCategory, error) {
	return fetchList[model.MenuItemWithCategory](c, "/api/v1/menu-items")
}

// GetInventoryItems fetches all active inventory items from the API.
func (c *APIClient) GetInventoryItems() ([]model.InventoryItem, error) {
	return fetchList[model.InventoryItem](c, "/api/v1/inventory")
}

// GetRecipe fetches the recipe for a specific menu item.
func (c *APIClient) GetRecipe(menuItemID string) ([]model.RecipeIngredientWithDetails, error) {
	return fetchList[model.RecipeIngredientWithDetails](c, "/api/v1/menu-items/"+menuItemID+"/recipe")
}

// GetOrders fetches orders from the central API for a date range.
func (c *APIClient) GetOrders(from, to string) ([]model.OrderWithItems, error) {
	return fetchList[model.OrderWithItems](c, "/api/v1/orders?from="+from+"&to="+to)
}

// fetchList is a generic helper for fetching list endpoints.
func fetchList[T any](c *APIClient, path string) ([]T, error) {
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data []T `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// ── Generic write helpers ──

// postJSON sends a POST request and decodes a single-item response.
func postJSON[T any](c *APIClient, path string, payload any) (*T, error) {
	return writeJSON[T](c, "POST", path, payload, http.StatusCreated)
}

// putJSON sends a PUT request and decodes a single-item response.
func putJSON[T any](c *APIClient, path string, payload any) (*T, error) {
	return writeJSON[T](c, "PUT", path, payload, http.StatusOK)
}

func writeJSON[T any](c *APIClient, method, path string, payload any, expectedStatus int) (*T, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.doRequest(method, path, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data T `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Data, nil
}

// doDelete sends a DELETE request and returns an error if it fails.
func (c *APIClient) doDelete(path string) error {
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ── Inventory Management ──

// InventoryPayload is the JSON body for creating/updating an inventory item.
type InventoryPayload struct {
	NameAr            string `json:"name_ar"`
	BaseUnitAr        string `json:"base_unit_ar"`
	StockQty          int    `json:"stock_qty"`
	LowStockThreshold int    `json:"low_stock_threshold"`
	UnitCost          int64  `json:"unit_cost"`
}

func (c *APIClient) CreateInventoryItem(p InventoryPayload) (*model.InventoryItem, error) {
	return postJSON[model.InventoryItem](c, "/api/v1/inventory", p)
}

func (c *APIClient) UpdateInventoryItem(id string, p InventoryPayload) (*model.InventoryItem, error) {
	return putJSON[model.InventoryItem](c, "/api/v1/inventory/"+id, p)
}

func (c *APIClient) DeleteInventoryItem(id string) error {
	return c.doDelete("/api/v1/inventory/" + id)
}

// StockAdjustPayload is the JSON body for POST /api/v1/inventory/adjust.
type StockAdjustPayload struct {
	InventoryItemID string `json:"inventory_item_id"`
	Delta           int    `json:"delta"`
	Reason          string `json:"reason"`
}

func (c *APIClient) AdjustStock(p StockAdjustPayload) error {
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	resp, err := c.doRequest("POST", "/api/v1/inventory/adjust", bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("adjust stock request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ── Recipe Management ──

// RecipeIngredientPayload is a single ingredient for SetRecipe.
type RecipeIngredientPayload struct {
	InventoryItemID string `json:"inventory_item_id"`
	Quantity        int    `json:"quantity"`
}

func (c *APIClient) SetRecipe(menuItemID string, ingredients []RecipeIngredientPayload) error {
	payload := map[string]any{"ingredients": ingredients}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe: %w", err)
	}
	resp, err := c.doRequest("PUT", "/api/v1/menu-items/"+menuItemID+"/recipe", bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("set recipe request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ── Menu Item Management ──

// MenuItemPayload is the JSON body for creating/updating a menu item.
type MenuItemPayload struct {
	CategoryID      string `json:"category_id"`
	NameAr          string `json:"name_ar"`
	Price           int64  `json:"price"`
	CostCalcMethod  string `json:"cost_calc_method"`
	ManualCostPrice int64  `json:"manual_cost_price"`
	ImagePath       string `json:"image_path"`
}

func (c *APIClient) CreateMenuItem(p MenuItemPayload) (*model.MenuItemWithCategory, error) {
	return postJSON[model.MenuItemWithCategory](c, "/api/v1/menu-items", p)
}

func (c *APIClient) UpdateMenuItem(id string, p MenuItemPayload) (*model.MenuItemWithCategory, error) {
	return putJSON[model.MenuItemWithCategory](c, "/api/v1/menu-items/"+id, p)
}

func (c *APIClient) DeleteMenuItem(id string) error {
	return c.doDelete("/api/v1/menu-items/" + id)
}

// ── Category Management ──

// CategoryPayload is the JSON body for creating/updating a category.
type CategoryPayload struct {
	NameAr    string `json:"name_ar"`
	SortOrder int    `json:"sort_order"`
}

func (c *APIClient) CreateCategory(p CategoryPayload) (*model.Category, error) {
	return postJSON[model.Category](c, "/api/v1/categories", p)
}

func (c *APIClient) UpdateCategory(id string, p CategoryPayload) (*model.Category, error) {
	return putJSON[model.Category](c, "/api/v1/categories/"+id, p)
}

func (c *APIClient) DeleteCategory(id string) error {
	return c.doDelete("/api/v1/categories/" + id)
}

// OrderPushPayload is the JSON body sent to POST /api/v1/orders.
type OrderPushPayload struct {
	ID            string          `json:"id"`
	Source        string          `json:"source"`
	TableNumber   string          `json:"table_number"`
	Total         int64           `json:"total"`
	PaymentMethod string          `json:"payment_method"`
	Items         []OrderItemPush `json:"items"`
	CreatedAt     string          `json:"created_at"`
}

// OrderItemPush is a single line item in an OrderPushPayload.
type OrderItemPush struct {
	ID             string `json:"id"`
	MenuItemID     string `json:"menu_item_id"`
	Quantity       int    `json:"quantity"`
	UnitPrice      int64  `json:"unit_price"`
	LineTotal      int64  `json:"line_total"`
	NameArSnapshot string `json:"name_ar_snapshot"`
}

// OrderPushResponse is the server's response after accepting an order.
type OrderPushResponse struct {
	OrderNumber string `json:"order_number"`
}

// PushOrder sends a local order to the central API.
// Returns the server-assigned order_number on success.
func (c *APIClient) PushOrder(payload OrderPushPayload) (*OrderPushResponse, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order: %w", err)
	}

	resp, err := c.doRequest("POST", "/api/v1/orders", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to push order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data struct {
			OrderNumber string `json:"order_number"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &OrderPushResponse{OrderNumber: result.Data.OrderNumber}, nil
}

func (c *APIClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

// ── Web Order Methods ──

// UpdateOrderStatus updates an order's status via PUT /api/v1/orders/{id}/status.
func (c *APIClient) UpdateOrderStatus(orderID string, status string) error {
	body := map[string]string{"status": status}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	resp, err := c.doRequest("PUT", "/api/v1/orders/"+orderID+"/status", bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update order status failed (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// ── Table Management ──

// TablePayload is the body for creating a table.
type TablePayload struct {
	Number string `json:"number"`
}

// ListTables fetches all tables from the API.
func (c *APIClient) ListTables() ([]model.Table, error) {
	return fetchList[model.Table](c, "/api/v1/tables")
}

// CreateTable creates a new table via the API.
func (c *APIClient) CreateTable(payload TablePayload) (*model.Table, error) {
	return postJSON[model.Table](c, "/api/v1/tables", payload)
}

// DeleteTable deletes a table via the API.
func (c *APIClient) DeleteTable(id string) error {
	return c.doDelete("/api/v1/tables/" + id)
}

// UploadImage uploads an image file to the API and returns the public URL.
func (c *APIClient) UploadImage(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/uploads", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}
	return result.URL, nil
}
