import { ref, computed } from 'vue'

let OrderService: any = null

export interface HistoryOrder {
  id: string
  order_number: string
  source: string
  table_number: string
  status: string
  total: number
  payment_method: string
  created_at: string
  items: HistoryOrderItem[]
}

export interface HistoryOrderItem {
  id: string
  name_ar_snapshot: string
  quantity: number
  unit_price: number
  line_total: number
}

const orders = ref<HistoryOrder[]>([])
const selectedOrder = ref<HistoryOrder | null>(null)
const isLoading = ref(false)
const dateFrom = ref(new Date().toISOString().split('T')[0])
const dateTo = ref(new Date().toISOString().split('T')[0])

export function useOrderHistory() {
  async function initBindings() {
    try {
      OrderService = await import('../../bindings/coffeeshop-pos/internal/service/orderservice')
    } catch {
      console.warn('OrderService bindings not available')
    }
  }

  const stats = computed(() => {
    const validOrders = orders.value.filter(o => o.status !== 'voided')
    const voidedOrders = orders.value.filter(o => o.status === 'voided')
    return {
      totalCount: validOrders.length,
      totalRevenue: validOrders.reduce((sum, o) => sum + o.total, 0),
      voidedCount: voidedOrders.length,
    }
  })

  async function loadOrders() {
    if (!OrderService) return
    isLoading.value = true
    try {
      const result = await OrderService.GetOrdersByDateRange(dateFrom.value, dateTo.value)
      orders.value = result || []
    } catch (err) {
      console.error('Failed to load order history:', err)
      orders.value = []
    } finally {
      isLoading.value = false
    }
  }

  async function voidOrder(orderID: string) {
    if (!OrderService) return
    try {
      await OrderService.VoidOrder(orderID)
      await loadOrders()
      selectedOrder.value = null
    } catch (err) {
      console.error('Failed to void order:', err)
    }
  }

  return {
    orders,
    selectedOrder,
    isLoading,
    dateFrom,
    dateTo,
    stats,
    initBindings,
    loadOrders,
    voidOrder,
  }
}
