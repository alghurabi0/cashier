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
	deviceID   string
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

// SetDeviceID sets the device ID sent via X-Device-ID header.
func (c *APIClient) SetDeviceID(deviceID string) {
	c.deviceID = deviceID
}

// SetToken sets the JWT auth token for subsequent requests.
func (c *APIClient) SetToken(token string) {
	c.token = token
}

// SetBaseURL updates the base URL for the API client.
func (c *APIClient) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimRight(baseURL, "/")
}

// LoginResponse contains the data returned after a successful login.
type LoginResponse struct {
	Token  string `json:"token"`
	Tenant struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		Settings struct {
			KitchenModeEnabled     bool   `json:"kitchen_mode_enabled"`
			ConflictResolutionMode string `json:"conflict_resolution_mode"`
			MenuURL                string `json:"menu_url"`
			IntroVideoURL          string `json:"intro_video_url"`
		} `json:"settings"`
	} `json:"tenant"`
}

// ProvisionResponse contains the data returned after provisioning with a code.
type ProvisionResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
	Tenant   struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		Settings struct {
			KitchenModeEnabled     bool   `json:"kitchen_mode_enabled"`
			ConflictResolutionMode string `json:"conflict_resolution_mode"`
			MenuURL                string `json:"menu_url"`
			IntroVideoURL          string `json:"intro_video_url"`
		} `json:"settings"`
	} `json:"tenant"`
}

// Login authenticates with the API and stores the returned token.
// Username should be in the format "user@tenant-slug".
func (c *APIClient) Login(username, password string) (*LoginResponse, error) {
	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	resp, err := c.doRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data LoginResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}

	c.token = result.Data.Token
	return &result.Data, nil
}

// Provision calls POST /api/v1/provision with a setup code.
func (c *APIClient) Provision(code string) (*ProvisionResponse, error) {
	body := fmt.Sprintf(`{"code":%q}`, code)
	resp, err := c.doRequest("POST", "/api/v1/provision", strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("provision failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("provision failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ProvisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode provision response: %w", err)
	}

	c.token = result.Token
	return &result, nil
}

// GetCategories fetches all active categories from the API.
func (c *APIClient) GetCategories() ([]model.Category, error) {
	return fetchList[model.Category](c, "/api/v1/categories")
}

// GetCategoriesSince fetches categories modified since the given timestamp.
func (c *APIClient) GetCategoriesSince(since string) ([]model.Category, error) {
	return fetchList[model.Category](c, "/api/v1/categories?since="+since)
}

// GetMenuItems fetches all active menu items from the API.
func (c *APIClient) GetMenuItems() ([]model.MenuItemWithCategory, error) {
	return fetchList[model.MenuItemWithCategory](c, "/api/v1/menu-items")
}

// GetMenuItemsSince fetches menu items modified since the given timestamp.
func (c *APIClient) GetMenuItemsSince(since string) ([]model.MenuItemWithCategory, error) {
	return fetchList[model.MenuItemWithCategory](c, "/api/v1/menu-items?since="+since)
}

// GetInventoryItems fetches all active inventory items from the API.
func (c *APIClient) GetInventoryItems() ([]model.InventoryItem, error) {
	return fetchList[model.InventoryItem](c, "/api/v1/inventory")
}

// GetInventoryItemsSince fetches inventory items modified since the given timestamp.
func (c *APIClient) GetInventoryItemsSince(since string) ([]model.InventoryItem, error) {
	return fetchList[model.InventoryItem](c, "/api/v1/inventory?since="+since)
}

// GetRecipe fetches the recipe for a specific menu item.
func (c *APIClient) GetRecipe(menuItemID string) ([]model.RecipeIngredientWithDetails, error) {
	return fetchList[model.RecipeIngredientWithDetails](c, "/api/v1/menu-items/"+menuItemID+"/recipe")
}

// GetAllRecipes fetches all recipe ingredients in bulk.
func (c *APIClient) GetAllRecipes() ([]model.RecipeIngredientWithDetails, error) {
	return fetchList[model.RecipeIngredientWithDetails](c, "/api/v1/recipes")
}

// GetAllRecipesSince fetches recipe ingredients for menu items modified since the given timestamp.
func (c *APIClient) GetAllRecipesSince(since string) ([]model.RecipeIngredientWithDetails, error) {
	return fetchList[model.RecipeIngredientWithDetails](c, "/api/v1/recipes?since="+since)
}

// GetOrders fetches orders from the central API for a date range.
func (c *APIClient) GetOrders(from, to string) ([]model.OrderWithItems, error) {
	return fetchList[model.OrderWithItems](c, "/api/v1/orders?from="+from+"&to="+to)
}

// GetOrdersSince fetches orders modified since the given timestamp (for delta sync).
func (c *APIClient) GetOrdersSince(since string) ([]model.OrderWithItems, error) {
	return fetchList[model.OrderWithItems](c, "/api/v1/orders?since="+since)
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
	return writeJSONWithHeaders[T](c, method, path, payload, expectedStatus, nil)
}

// writeJSONWithHeaders sends a JSON request with optional extra headers.
func writeJSONWithHeaders[T any](c *APIClient, method, path string, payload any, expectedStatus int, headers map[string]string) (*T, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.doRequestWithHeaders(method, path, bytes.NewReader(jsonBytes), headers)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &SyncConflictError{StatusCode: 409, Body: string(bodyBytes)}
	}

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

// SyncConflictError represents a 409 Conflict response from the API.
type SyncConflictError struct {
	StatusCode int
	Body       string
}

func (e *SyncConflictError) Error() string {
	return fmt.Sprintf("conflict (409): %s", e.Body)
}

// putJSONVersioned sends a PUT with an X-Expected-Version header for optimistic concurrency.
func putJSONVersioned[T any](c *APIClient, path string, payload any, baseVersion string) (*T, error) {
	headers := map[string]string{"X-Expected-Version": baseVersion}
	return writeJSONWithHeaders[T](c, "PUT", path, payload, http.StatusOK, headers)
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
	ID                string `json:"id,omitempty"`
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

func (c *APIClient) UpdateInventoryItemVersioned(id string, p InventoryPayload, baseVersion string) (*model.InventoryItem, error) {
	return putJSONVersioned[model.InventoryItem](c, "/api/v1/inventory/"+id, p, baseVersion)
}

func (c *APIClient) DeleteInventoryItem(id string) error {
	return c.doDelete("/api/v1/inventory/" + id)
}

// StockAdjustPayload is the JSON body for POST /api/v1/inventory/adjust.
type StockAdjustPayload struct {
	ID              string `json:"id,omitempty"`
	InventoryItemID string `json:"inventory_item_id"`
	Delta           int    `json:"delta"`
	Reason          string `json:"reason_ar"`
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
	ID              string `json:"id,omitempty"`
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

func (c *APIClient) UpdateMenuItemVersioned(id string, p MenuItemPayload, baseVersion string) (*model.MenuItemWithCategory, error) {
	return putJSONVersioned[model.MenuItemWithCategory](c, "/api/v1/menu-items/"+id, p, baseVersion)
}

func (c *APIClient) DeleteMenuItem(id string) error {
	return c.doDelete("/api/v1/menu-items/" + id)
}

// ── Category Management ──

// CategoryPayload is the JSON body for creating/updating a category.
type CategoryPayload struct {
	ID        string `json:"id,omitempty"`
	NameAr    string `json:"name_ar"`
	SortOrder int    `json:"sort_order"`
}

func (c *APIClient) CreateCategory(p CategoryPayload) (*model.Category, error) {
	return postJSON[model.Category](c, "/api/v1/categories", p)
}

func (c *APIClient) UpdateCategory(id string, p CategoryPayload) (*model.Category, error) {
	return putJSON[model.Category](c, "/api/v1/categories/"+id, p)
}

func (c *APIClient) UpdateCategoryVersioned(id string, p CategoryPayload, baseVersion string) (*model.Category, error) {
	return putJSONVersioned[model.Category](c, "/api/v1/categories/"+id, p, baseVersion)
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
	return c.doRequestWithHeaders(method, path, body, nil)
}

func (c *APIClient) doRequestWithHeaders(method, path string, body io.Reader, extraHeaders map[string]string) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.deviceID != "" {
		req.Header.Set("X-Device-ID", c.deviceID)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
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
	ID     string `json:"id,omitempty"`
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
		URL  string `json:"url"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}
	url := result.Data.URL
	if url == "" {
		url = result.URL
	}
	if url == "" {
		return "", fmt.Errorf("upload response did not include an image URL")
	}
	return url, nil
}
