export type { Category, MenuItem } from '@cashier/menu-sdk'

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
