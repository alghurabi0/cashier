// ─── Types ───

export interface TenantInfo {
  name: string
  slug: string
}

export interface TableInfo {
  number: string
}

export interface Category {
  id: string
  name_ar: string
  image_url?: string
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

export interface MenuData {
  tenant: TenantInfo
  table: TableInfo
  categories: Category[]
  menu_items: MenuItem[]
}

export interface OrderItem {
  menu_item_id: string
  quantity: number
}

export interface OrderResult {
  id: string
  order_number: string
  table_number: string
  status: string
  total: number
  items: Array<{
    menu_item_id: string
    name_ar: string
    quantity: number
    unit_price: number
    total: number
  }>
}

// ─── SDK Class ───

export interface CashierMenuOptions {
  /** The base URL of the coffeeshop API (e.g. "https://api.example.com") */
  apiUrl: string
}

/**
 * CashierMenu SDK — connects any static menu site to the coffeeshop API.
 *
 * Usage:
 * ```ts
 * import { CashierMenu } from '@cashier/menu-sdk'
 *
 * const menu = new CashierMenu({ apiUrl: 'https://api.example.com' })
 * const data = await menu.load(token)
 * // data.tenant, data.table, data.categories, data.menu_items
 *
 * await menu.submitOrder(token, [{ menu_item_id: '...', quantity: 2 }])
 * ```
 */
export class CashierMenu {
  private apiUrl: string

  constructor(options: CashierMenuOptions) {
    this.apiUrl = options.apiUrl.replace(/\/$/, '') // strip trailing slash
  }

  /**
   * Load all menu data for a given table token.
   * Returns tenant info, table number, categories, and menu items.
   * Throws on invalid token or network error.
   */
  async load(token: string): Promise<MenuData> {
    const res = await fetch(`${this.apiUrl}/api/v1/public/menu?token=${encodeURIComponent(token)}`)
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body.error || `Failed to load menu (${res.status})`)
    }
    return res.json()
  }

  /**
   * Submit a web order for a given table token.
   * Returns the created order with order number, items, and total.
   * Throws on validation error or network error.
   */
  async submitOrder(token: string, items: OrderItem[]): Promise<OrderResult> {
    const res = await fetch(`${this.apiUrl}/api/v1/web-orders?token=${encodeURIComponent(token)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items }),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body.error || `Failed to submit order (${res.status})`)
    }
    const json = await res.json()
    return json.data ?? json
  }

  /**
   * Helper: extract the table token from the current page URL.
   * Returns null if no token is found.
   */
  static getTokenFromURL(): string | null {
    return new URLSearchParams(window.location.search).get('token')
  }
}
