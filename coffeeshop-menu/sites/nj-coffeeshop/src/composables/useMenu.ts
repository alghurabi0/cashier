import { ref } from 'vue'
import { CashierMenu } from '@cashier/menu-sdk'
import type { Category, MenuItem, TenantInfo, TableInfo } from '@cashier/menu-sdk'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const sdk = new CashierMenu({ apiUrl: API_BASE })

const categories = ref<Category[]>([])
const menuItems = ref<MenuItem[]>([])
const tenantInfo = ref<TenantInfo | null>(null)
const tableInfo = ref<TableInfo | null>(null)
const isLoading = ref(false)
const error = ref<string | null>(null)

export function useMenu() {
  async function loadAll(token: string) {
    isLoading.value = true
    error.value = null
    try {
      const data = await sdk.load(token)
      tenantInfo.value = data.tenant
      tableInfo.value = data.table
      categories.value = data.categories || []
      menuItems.value = data.menu_items || []
    } catch (err: any) {
      error.value = err.message
    } finally {
      isLoading.value = false
    }
  }

  return {
    categories,
    menuItems,
    tenantInfo,
    tableInfo,
    isLoading,
    error,
    loadAll,
  }
}
