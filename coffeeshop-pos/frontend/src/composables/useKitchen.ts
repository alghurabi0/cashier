import { ref, computed } from 'vue'
import type { OrderWithItems } from '../types'

let OrderService: any = null
let WebOrderService: any = null

const acceptedOrders = ref<OrderWithItems[]>([])
const isLoading = ref(false)

export function useKitchen() {
  async function initBindings() {
    try {
      OrderService = await import('../../bindings/coffeeshop-pos/internal/service/orderservice')
    } catch {
      console.warn('OrderService bindings not available')
    }
    try {
      WebOrderService = await import('../../bindings/coffeeshop-pos/internal/service/weborderservice')
    } catch {
      console.warn('WebOrderService bindings not available')
    }
  }

  const orderCount = computed(() => acceptedOrders.value.length)

  async function loadOrders() {
    if (!OrderService) return
    try {
      const result = await OrderService.GetAcceptedOrders()
      acceptedOrders.value = result || []
    } catch (err) {
      console.error('Failed to load accepted orders:', err)
    }
  }

  let pollInterval: ReturnType<typeof setInterval> | null = null

  function startPolling(intervalMs = 2000) {
    if (pollInterval) return
    pollInterval = setInterval(() => {
      loadOrders()
    }, intervalMs)
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  async function completeOrder(orderID: string, source: string) {
    isLoading.value = true
    try {
      if (source === 'web_menu' && WebOrderService) {
        await WebOrderService.CompleteOrder(orderID)
      } else if (OrderService) {
        await OrderService.CompleteCashierOrder(orderID)
      }
      await loadOrders()
    } catch (err) {
      console.error('Failed to complete order:', err)
    } finally {
      isLoading.value = false
    }
  }

  return {
    acceptedOrders,
    orderCount,
    isLoading,
    initBindings,
    loadOrders,
    startPolling,
    stopPolling,
    completeOrder,
  }
}
