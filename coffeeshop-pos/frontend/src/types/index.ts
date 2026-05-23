// TypeScript interfaces mirroring the Go models exposed via Wails bindings.

export interface Category {
  id: string
  name_ar: string
  sort_order: number
  is_active: boolean
}

export interface MenuItem {
  id: string
  category_id: string
  name_ar: string
  price: number          // fils
  cost_calc_method: string
  manual_cost_price: number
  cached_auto_cost: number
  image_path: string
  is_active: boolean
  category_name_ar: string
}

export interface InventoryItem {
  id: string
  name_ar: string
  base_unit_ar: string
  stock_qty: number
  low_stock_threshold: number
  unit_cost: number
  is_active: boolean
}

export interface CartItem {
  menu_item_id: string
  name_ar: string
  price: number          // fils per unit
  quantity: number
}

export interface OrderItem {
  id: string
  order_id: string
  menu_item_id: string
  quantity: number
  unit_price: number
  line_total: number
  name_ar_snapshot: string
}

export interface Order {
  id: string
  order_number: string
  source: string
  table_number: string
  status: string
  total: number
  payment_method: string
  created_at: string
  synced: boolean
}

export interface OrderWithItems extends Order {
  items: OrderItem[]
}

/** Format a fils amount as a readable price string (e.g. 3500 → "3,500") */
export function formatPrice(fils: number): string {
  return fils.toLocaleString('en-US')
}

// ── Management Types ──

export interface RecipeIngredient {
  menu_item_id: string
  inventory_item_id: string
  quantity: number
}

export interface RecipeIngredientWithDetails extends RecipeIngredient {
  inventory_name_ar: string
  base_unit_ar: string
  unit_cost: number
}

export interface RecipeIngredientInput {
  inventory_item_id: string
  quantity: number
}

export interface InventoryFormInput {
  name_ar: string
  base_unit_ar: string
  stock_qty: number
  low_stock_threshold: number
  unit_cost: number
}

export interface StockAdjustmentInput {
  inventory_item_id: string
  delta: number
  reason: string
}

export interface MenuItemFormInput {
  category_id: string
  name_ar: string
  price: number
  cost_calc_method: string
  manual_cost_price: number
}

