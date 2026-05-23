package model

// Models for the POS desktop app. These mirror the API models but use
// string for UUIDs (SQLite stores UUIDs as TEXT).

// Category represents a menu category.
type Category struct {
	ID        string `db:"id"         json:"id"`
	NameAr    string `db:"name_ar"    json:"name_ar"`
	SortOrder int    `db:"sort_order" json:"sort_order"`
	IsActive  bool   `db:"is_active"  json:"is_active"`
}

// MenuItem represents a front-of-house product.
type MenuItem struct {
	ID              string `db:"id"                json:"id"`
	CategoryID      string `db:"category_id"       json:"category_id"`
	NameAr          string `db:"name_ar"            json:"name_ar"`
	Price           int64  `db:"price"              json:"price"`
	CostCalcMethod  string `db:"cost_calc_method"   json:"cost_calc_method"`
	ManualCostPrice int64  `db:"manual_cost_price"  json:"manual_cost_price"`
	CachedAutoCost  int64  `db:"cached_auto_cost"   json:"cached_auto_cost"`
	ImagePath       string `db:"image_path"         json:"image_path"`
	IsActive        bool   `db:"is_active"          json:"is_active"`
}

// MenuItemWithCategory extends MenuItem with the category name.
type MenuItemWithCategory struct {
	MenuItem
	CategoryNameAr string `db:"category_name_ar" json:"category_name_ar"`
}

// InventoryItem represents a back-of-house raw material.
type InventoryItem struct {
	ID                string `db:"id"                  json:"id"`
	NameAr            string `db:"name_ar"              json:"name_ar"`
	BaseUnitAr        string `db:"base_unit_ar"         json:"base_unit_ar"`
	StockQty          int    `db:"stock_qty"            json:"stock_qty"`
	LowStockThreshold int    `db:"low_stock_threshold"  json:"low_stock_threshold"`
	UnitCost          int64  `db:"unit_cost"            json:"unit_cost"`
	IsActive          bool   `db:"is_active"            json:"is_active"`
}

// RecipeIngredient links a MenuItem to an InventoryItem.
type RecipeIngredient struct {
	ID              string `db:"id"                json:"id"`
	MenuItemID      string `db:"menu_item_id"      json:"menu_item_id"`
	InventoryItemID string `db:"inventory_item_id" json:"inventory_item_id"`
	Quantity        int    `db:"quantity"           json:"quantity"`
}

// RecipeIngredientWithDetails includes inventory item info for display.
type RecipeIngredientWithDetails struct {
	RecipeIngredient
	InventoryNameAr string `db:"inventory_name_ar" json:"inventory_name_ar"`
	BaseUnitAr      string `db:"base_unit_ar"      json:"base_unit_ar"`
	UnitCost        int64  `db:"unit_cost"          json:"unit_cost"`
}

// Order represents a sales transaction in local SQLite.
type Order struct {
	ID            string `db:"id"             json:"id"`
	OrderNumber   string `db:"order_number"   json:"order_number"`
	Source        string `db:"source"         json:"source"`
	TableNumber   string `db:"table_number"   json:"table_number"`
	Status        string `db:"status"         json:"status"`
	Total         int64  `db:"total"          json:"total"`
	PaymentMethod string `db:"payment_method" json:"payment_method"`
	CreatedAt     string `db:"created_at"     json:"created_at"`
	Synced        bool   `db:"synced"         json:"synced"`
}

// OrderItem represents a single line in an order.
type OrderItem struct {
	ID             string `db:"id"               json:"id"`
	OrderID        string `db:"order_id"         json:"order_id"`
	MenuItemID     string `db:"menu_item_id"     json:"menu_item_id"`
	Quantity       int    `db:"quantity"          json:"quantity"`
	UnitPrice      int64  `db:"unit_price"       json:"unit_price"`
	LineTotal      int64  `db:"line_total"       json:"line_total"`
	NameArSnapshot string `db:"name_ar_snapshot" json:"name_ar_snapshot"`
}

// OrderWithItems bundles an order with its line items.
type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

// CartItem is the input from the Vue frontend when creating an order.
type CartItem struct {
	MenuItemID string `json:"menu_item_id"`
	NameAr     string `json:"name_ar"`
	Price      int64  `json:"price"`
	Quantity   int    `json:"quantity"`
}

// Table represents a restaurant table (from the API).
type Table struct {
	ID        string `json:"id"`
	Number    string `json:"number"`
	Token     string `json:"token"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

