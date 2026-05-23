import { ref } from 'vue'
import type { Category, InventoryItem, MenuItem, RecipeIngredientWithDetails } from '../types'

// Reactive state for management views
const inventoryItems = ref<InventoryItem[]>([])
const allMenuItems = ref<MenuItem[]>([])
const categories = ref<Category[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)

let ManagementService: any = null
let DataService: any = null

export function useManagement() {
  async function initBindings() {
    try {
      ManagementService = await import('../../bindings/coffeeshop-pos/internal/service/managementservice')
    } catch {
      console.warn('ManagementService bindings not available')
    }
    try {
      DataService = await import('../../bindings/coffeeshop-pos/internal/service/dataservice')
    } catch {
      console.warn('DataService bindings not available')
    }
  }

  // ── Data Loading ──

  async function loadCategories() {
    if (!DataService) return
    try {
      const result = await DataService.GetCategories()
      categories.value = result || []
    } catch (err: any) {
      console.error('Failed to load categories:', err)
      error.value = err.message
    }
  }

  async function loadInventoryItems() {
    if (!DataService) return
    try {
      const result = await DataService.GetInventoryItems()
      inventoryItems.value = result || []
    } catch (err: any) {
      console.error('Failed to load inventory:', err)
      error.value = err.message
    }
  }

  async function loadMenuItems() {
    if (!DataService) return
    try {
      const result = await DataService.GetMenuItems('')
      allMenuItems.value = result || []
    } catch (err: any) {
      console.error('Failed to load menu items:', err)
    }
  }

  // ── Category CRUD ──

  async function createCategory(nameAr: string, sortOrder: number) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.CreateCategory(nameAr, sortOrder)
      await loadCategories()
      await loadMenuItems() // refresh category names in menu items
    } finally {
      isLoading.value = false
    }
  }

  async function updateCategory(id: string, nameAr: string, sortOrder: number) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.UpdateCategory(id, nameAr, sortOrder)
      await loadCategories()
      await loadMenuItems()
    } finally {
      isLoading.value = false
    }
  }

  async function deleteCategory(id: string) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.DeleteCategory(id)
      await loadCategories()
      await loadMenuItems()
    } finally {
      isLoading.value = false
    }
  }

  // ── Menu Item CRUD ──

  async function createMenuItem(categoryId: string, nameAr: string, price: number, costCalcMethod: string, manualCostPrice: number) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.CreateMenuItem(categoryId, nameAr, price, costCalcMethod, manualCostPrice)
      await loadMenuItems()
    } finally {
      isLoading.value = false
    }
  }

  async function updateMenuItem(id: string, categoryId: string, nameAr: string, price: number, costCalcMethod: string, manualCostPrice: number) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.UpdateMenuItem(id, categoryId, nameAr, price, costCalcMethod, manualCostPrice)
      await loadMenuItems()
    } finally {
      isLoading.value = false
    }
  }

  async function deleteMenuItem(id: string) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.DeleteMenuItem(id)
      await loadMenuItems()
    } finally {
      isLoading.value = false
    }
  }

  // ── Inventory CRUD ──

  async function createInventoryItem(nameAr: string, baseUnitAr: string, stockQty: number, lowThreshold: number, unitCost: number) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.CreateInventoryItem(nameAr, baseUnitAr, stockQty, lowThreshold, unitCost)
      await loadInventoryItems()
    } finally {
      isLoading.value = false
    }
  }

  async function updateInventoryItem(id: string, nameAr: string, baseUnitAr: string, stockQty: number, lowThreshold: number, unitCost: number) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.UpdateInventoryItem(id, nameAr, baseUnitAr, stockQty, lowThreshold, unitCost)
      await loadInventoryItems()
    } finally {
      isLoading.value = false
    }
  }

  async function deleteInventoryItem(id: string) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.DeleteInventoryItem(id)
      await loadInventoryItems()
    } finally {
      isLoading.value = false
    }
  }

  async function adjustStock(itemId: string, delta: number, reason: string) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.AdjustStock(itemId, delta, reason)
      await loadInventoryItems()
    } finally {
      isLoading.value = false
    }
  }

  // ── Recipes ──

  async function getRecipe(menuItemId: string): Promise<RecipeIngredientWithDetails[]> {
    if (!ManagementService) throw new Error('Management service unavailable')
    const result = await ManagementService.GetRecipe(menuItemId)
    return result || []
  }

  async function setRecipe(menuItemId: string, ingredients: { inventory_item_id: string; quantity: number }[]) {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.SetRecipe(menuItemId, ingredients)
      await loadMenuItems()
    } finally {
      isLoading.value = false
    }
  }

  // ── Sync ──

  async function triggerSync() {
    if (!ManagementService) throw new Error('Management service unavailable')
    isLoading.value = true
    try {
      await ManagementService.TriggerSync()
      await loadCategories()
      await loadInventoryItems()
      await loadMenuItems()
    } finally {
      isLoading.value = false
    }
  }

  return {
    inventoryItems,
    allMenuItems,
    categories,
    isLoading,
    error,
    initBindings,
    loadCategories,
    loadInventoryItems,
    loadMenuItems,
    createCategory,
    updateCategory,
    deleteCategory,
    createMenuItem,
    updateMenuItem,
    deleteMenuItem,
    createInventoryItem,
    updateInventoryItem,
    deleteInventoryItem,
    adjustStock,
    getRecipe,
    setRecipe,
    triggerSync,
  }
}
