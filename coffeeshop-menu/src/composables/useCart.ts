import { ref, computed } from 'vue'
import type { CartItem, WebOrderResponse } from '../types'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const items = ref<CartItem[]>([])
const isSubmitting = ref(false)
const orderResult = ref<WebOrderResponse | null>(null)
const orderError = ref<string | null>(null)

export function useCart() {
  const total = computed(() =>
    items.value.reduce((sum, item) => sum + item.price * item.quantity, 0)
  )

  const itemCount = computed(() =>
    items.value.reduce((sum, item) => sum + item.quantity, 0)
  )

  function addItem(menuItemId: string, nameAr: string, price: number) {
    const existing = items.value.find(i => i.menu_item_id === menuItemId)
    if (existing) {
      existing.quantity++
    } else {
      items.value.push({
        menu_item_id: menuItemId,
        name_ar: nameAr,
        price,
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
    if (item && item.quantity > 1) {
      item.quantity--
    } else if (item) {
      removeItem(menuItemId)
    }
  }

  function clearCart() {
    items.value = []
  }

  async function submitOrder(token: string) {
    isSubmitting.value = true
    orderError.value = null

    const body = {
      items: items.value.map(i => ({
        menu_item_id: i.menu_item_id,
        quantity: i.quantity,
      })),
    }

    try {
      const res = await fetch(`${API_BASE}/api/v1/web-orders?token=${token}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || 'Failed to submit order')
      }

      orderResult.value = await res.json()
      clearCart()
    } catch (err: any) {
      orderError.value = err.message
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    items,
    total,
    itemCount,
    isSubmitting,
    orderResult,
    orderError,
    addItem,
    removeItem,
    incrementQty,
    decrementQty,
    clearCart,
    submitOrder,
  }
}
