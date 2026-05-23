import { ref, computed } from 'vue'
import type { CartItem, MenuItem } from '../types'

const items = ref<CartItem[]>([])

export function useCart() {
  const total = computed(() =>
    items.value.reduce((sum, item) => sum + item.price * item.quantity, 0)
  )

  const itemCount = computed(() =>
    items.value.reduce((sum, item) => sum + item.quantity, 0)
  )

  function addItem(menuItem: MenuItem) {
    const existing = items.value.find(i => i.menu_item_id === menuItem.id)
    if (existing) {
      existing.quantity++
    } else {
      items.value.push({
        menu_item_id: menuItem.id,
        name_ar: menuItem.name_ar,
        price: menuItem.price,
        quantity: 1,
      })
    }
  }

  function removeItem(menuItemId: string) {
    items.value = items.value.filter(i => i.menu_item_id !== menuItemId)
  }

  function incrementQty(menuItemId: string) {
    const item = items.value.find(i => i.menu_item_id === menuItemId)
    if (item) item.quantity++
  }

  function decrementQty(menuItemId: string) {
    const item = items.value.find(i => i.menu_item_id === menuItemId)
    if (item) {
      item.quantity--
      if (item.quantity <= 0) {
        removeItem(menuItemId)
      }
    }
  }

  function clear() {
    items.value = []
  }

  return {
    items,
    total,
    itemCount,
    addItem,
    removeItem,
    incrementQty,
    decrementQty,
    clear,
  }
}
