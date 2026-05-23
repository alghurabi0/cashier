import { ref, computed } from 'vue'

let WebOrderService: any = null

const pendingOrders = ref<OrderWithItems[]>([])
const acceptedOrders = ref<OrderWithItems[]>([])
const completedOrders = ref<OrderWithItems[]>([])
const isLoading = ref(false)
const soundEnabled = ref(true)
let notificationAudio: HTMLAudioElement | null = null

export interface OrderWithItems {
  id: string
  order_number: string
  source: string
  table_number: string
  status: string
  total: number
  payment_method: string
  created_at: string
  items: OrderItem[]
}

export interface OrderItem {
  id: string
  order_id: string
  menu_item_id: string
  quantity: number
  unit_price: number
  line_total: number
  name_ar_snapshot: string
}

export function useWebOrders() {
  async function initBindings() {
    try {
      WebOrderService = await import('../../bindings/coffeeshop-pos/internal/service/weborderservice')
    } catch {
      console.warn('WebOrderService bindings not available')
    }

    // Initialize audio
    notificationAudio = new Audio('/notification.mp3')
    notificationAudio.volume = 0.7
  }

  const pendingCount = computed(() => pendingOrders.value.length)

  async function loadOrders() {
    if (!WebOrderService) return
    try {
      const pending = await WebOrderService.GetPendingOrders()
      pendingOrders.value = pending || []

      const accepted = await WebOrderService.GetAcceptedOrders()
      acceptedOrders.value = accepted || []

      const completed = await WebOrderService.GetCompletedOrders()
      completedOrders.value = completed || []
    } catch (err) {
      console.error('Failed to load web orders:', err)
    }
  }

  // Poll for new orders (SSE pushes to Go, but we need to poll Go → Vue)
  let pollInterval: ReturnType<typeof setInterval> | null = null

  function startPolling(intervalMs = 2000) {
    if (pollInterval) return
    pollInterval = setInterval(async () => {
      const prevCount = pendingOrders.value.length
      await loadOrders()
      const newCount = pendingOrders.value.length

      // Play sound if new orders arrived
      if (newCount > prevCount && soundEnabled.value && notificationAudio) {
        notificationAudio.play().catch(() => {})
      }
    }, intervalMs)
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  function toggleSound() {
    soundEnabled.value = !soundEnabled.value
  }

  async function acceptOrder(orderID: string) {
    if (!WebOrderService) return
    isLoading.value = true
    try {
      await WebOrderService.AcceptOrder(orderID)
      await loadOrders()
    } catch (err) {
      console.error('Failed to accept order:', err)
    } finally {
      isLoading.value = false
    }
  }

  async function rejectOrder(orderID: string) {
    if (!WebOrderService) return
    isLoading.value = true
    try {
      await WebOrderService.RejectOrder(orderID)
      await loadOrders()
    } catch (err) {
      console.error('Failed to reject order:', err)
    } finally {
      isLoading.value = false
    }
  }

  async function completeOrder(orderID: string) {
    if (!WebOrderService) return
    isLoading.value = true
    try {
      await WebOrderService.CompleteOrder(orderID)
      await loadOrders()
    } catch (err) {
      console.error('Failed to complete order:', err)
    } finally {
      isLoading.value = false
    }
  }

  return {
    pendingOrders,
    acceptedOrders,
    completedOrders,
    pendingCount,
    isLoading,
    soundEnabled,
    initBindings,
    loadOrders,
    startPolling,
    stopPolling,
    toggleSound,
    acceptOrder,
    rejectOrder,
    completeOrder,
  }
}
