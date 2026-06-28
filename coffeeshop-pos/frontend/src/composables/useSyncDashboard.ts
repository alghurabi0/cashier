import { ref, onMounted, onUnmounted } from 'vue'
import type { SyncStatusSnapshot, SyncLogEntry, FailedOrderInfo } from '../types'

let SyncService: any = null

export function useSyncDashboard() {
  const syncStatus = ref<SyncStatusSnapshot | null>(null)
  const failedOrders = ref<FailedOrderInfo[]>([])
  const isLoading = ref(false)
  const actionMessage = ref('')
  let pollTimer: ReturnType<typeof setInterval> | null = null

  async function initBindings() {
    try {
      SyncService = await import('../../bindings/coffeeshop-pos/internal/service/syncservice')
    } catch {
      console.warn('SyncService bindings not available')
    }
  }

  async function refresh() {
    if (!SyncService) return
    try {
      const status = await SyncService.GetSyncStatus()
      syncStatus.value = status
      // Also refresh failed orders if there are any
      if (status && status.failed_orders > 0) {
        const orders = await SyncService.GetFailedOrders()
        failedOrders.value = orders || []
      } else {
        failedOrders.value = []
      }
    } catch (err) {
      console.error('Failed to refresh sync status:', err)
    }
  }

  async function triggerSync() {
    if (!SyncService) return
    isLoading.value = true
    actionMessage.value = ''
    try {
      await SyncService.TriggerSync()
      actionMessage.value = '✓ تمت المزامنة بنجاح'
      await refresh()
    } catch (err: any) {
      actionMessage.value = '✕ فشلت المزامنة: ' + (err.message || err)
    } finally {
      isLoading.value = false
      setTimeout(() => { actionMessage.value = '' }, 4000)
    }
  }

  async function triggerFullSync() {
    if (!SyncService) return
    isLoading.value = true
    actionMessage.value = ''
    try {
      await SyncService.TriggerFullSync()
      actionMessage.value = '✓ تمت المزامنة الكاملة بنجاح'
      await refresh()
    } catch (err: any) {
      actionMessage.value = '✕ فشلت المزامنة الكاملة: ' + (err.message || err)
    } finally {
      isLoading.value = false
      setTimeout(() => { actionMessage.value = '' }, 4000)
    }
  }

  async function retryFailed() {
    if (!SyncService) return
    isLoading.value = true
    actionMessage.value = ''
    try {
      const count = await SyncService.RetryFailedOrders()
      actionMessage.value = `✓ تم إعادة ${count} طلب للمزامنة`
      await refresh()
    } catch (err: any) {
      actionMessage.value = '✕ فشلت إعادة المحاولة: ' + (err.message || err)
    } finally {
      isLoading.value = false
      setTimeout(() => { actionMessage.value = '' }, 4000)
    }
  }

  async function resetSyncState() {
    if (!SyncService) return
    isLoading.value = true
    actionMessage.value = ''
    try {
      await SyncService.ResetSyncState()
      actionMessage.value = '✓ تم إعادة تعيين حالة المزامنة'
      await refresh()
    } catch (err: any) {
      actionMessage.value = '✕ فشلت إعادة التعيين: ' + (err.message || err)
    } finally {
      isLoading.value = false
      setTimeout(() => { actionMessage.value = '' }, 4000)
    }
  }

  function startPolling(intervalMs = 5000) {
    stopPolling()
    pollTimer = setInterval(refresh, intervalMs)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  return {
    syncStatus,
    failedOrders,
    isLoading,
    actionMessage,
    initBindings,
    refresh,
    triggerSync,
    triggerFullSync,
    retryFailed,
    resetSyncState,
    startPolling,
    stopPolling,
  }
}
