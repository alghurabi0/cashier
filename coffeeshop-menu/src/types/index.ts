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
  price: number
  image_path: string
  is_active: boolean
}

export interface CartItem {
  menu_item_id: string
  name_ar: string
  price: number
  quantity: number
}

export interface WebOrderResponse {
  id: string
  order_number: string
  source: string
  table_number: string
  status: string
  total: number
  items: {
    id: string
    name_ar_snapshot: string
    quantity: number
    unit_price: number
    line_total: number
  }[]
}

export function formatPrice(fils: number): string {
  return fils.toLocaleString('en-US')
}
