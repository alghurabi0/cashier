import { ref } from 'vue'
import type { Category, MenuItem } from '../types'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const categories = ref<Category[]>([])
const menuItems = ref<MenuItem[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)

export function useMenu() {
  async function loadCategories() {
    try {
      const res = await fetch(`${API_BASE}/api/v1/categories`)
      if (!res.ok) throw new Error('Failed to load categories')
      const json = await res.json()
      categories.value = (json.data || []).filter((c: Category) => c.is_active)
    } catch (err: any) {
      error.value = err.message
    }
  }

  async function loadMenuItems() {
    try {
      const res = await fetch(`${API_BASE}/api/v1/menu-items`)
      if (!res.ok) throw new Error('Failed to load menu items')
      const json = await res.json()
      menuItems.value = (json.data || []).filter((m: MenuItem) => m.is_active)
      console.log(menuItems);
    } catch (err: any) {
      error.value = err.message
    }
  }

  async function loadAll() {
    isLoading.value = true
    await Promise.all([loadCategories(), loadMenuItems()])
    isLoading.value = false
  }

  return {
    categories,
    menuItems,
    isLoading,
    error,
    loadAll,
  }
}
