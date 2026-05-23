import { ref } from 'vue'
import type { OrderWithItems } from '../types'

const todayOrders = ref<OrderWithItems[]>([])
const lastOrder = ref<OrderWithItems | null>(null)

export function useOrders() {
  function setTodayOrders(orders: OrderWithItems[]) {
    todayOrders.value = orders
  }

  function addOrder(order: OrderWithItems) {
    todayOrders.value.unshift(order)
    lastOrder.value = order
  }

  function clearLastOrder() {
    lastOrder.value = null
  }

  return {
    todayOrders,
    lastOrder,
    setTodayOrders,
    addOrder,
    clearLastOrder,
  }
}
