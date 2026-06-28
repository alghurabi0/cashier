import { ref, computed } from 'vue'
import { CashierMenu } from '@cashier/menu-sdk'
import type { CartItem, WebOrderResponse } from '../types'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const sdk = new CashierMenu({ apiUrl: API_BASE })

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

    try {
      const result = await sdk.submitOrder(
        token,
        items.value.map(i => ({
          menu_item_id: i.menu_item_id,
          quantity: i.quantity,
        }))
      )
      orderResult.value = result as unknown as WebOrderResponse
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
