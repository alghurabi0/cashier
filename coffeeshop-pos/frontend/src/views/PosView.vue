<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import type { Category, MenuItem, OrderWithItems } from '../types'
import { useCart } from '../composables/useCart'
import { useOrders } from '../composables/useOrders'
import HeaderBar from '../components/HeaderBar.vue'
import CategoryTabs from '../components/CategoryTabs.vue'
import MenuGrid from '../components/MenuGrid.vue'
import CartPanel from '../components/CartPanel.vue'
import CheckoutDialog from '../components/CheckoutDialog.vue'
import OrderQueuePanel from '../components/OrderQueuePanel.vue'

// ── State ──
const categories = ref<Category[]>([])
const menuItems = ref<MenuItem[]>([])
const selectedCategoryId = ref<string | null>(null)
const showCheckout = ref(false)
const toastMessage = ref('')
const toastVisible = ref(false)
const syncStatus = ref<'online' | 'offline' | 'syncing'>('offline')
const kitchenModeEnabled = ref(false)

const { items: cartItems, total, itemCount, addItem, removeItem, incrementQty, decrementQty, clear: clearCart } = useCart()
const { todayOrders, addOrder } = useOrders()

// ── Order queue state ──
const acceptedOrders = computed(() =>
  todayOrders.value.filter(o => o.status === 'accepted' && o.source === 'cashier')
)
const recentCompletedOrders = computed(() =>
  todayOrders.value.filter(o => o.status === 'completed' && o.source === 'cashier').slice(0, 5)
)

// ── Wails Bindings ──
let DataService: any = null
let OrderService: any = null
let ReceiptService: any = null

async function initBindings() {
  try {
    const dataMod = await import('../../bindings/coffeeshop-pos/internal/service/dataservice')
    DataService = dataMod
  } catch {
    console.warn('DataService bindings not available')
  }
  try {
    const orderMod = await import('../../bindings/coffeeshop-pos/internal/service/orderservice')
    OrderService = orderMod
  } catch {
    console.warn('OrderService bindings not available')
  }
  try {
    const receiptMod = await import('../../bindings/coffeeshop-pos/internal/service/receiptservice')
    ReceiptService = receiptMod
  } catch {
    console.warn('ReceiptService bindings not available')
  }
  // Load kitchen mode setting
  try {
    const configMod = await import('../../bindings/coffeeshop-pos/internal/service/configstoreservice')
    const kitchenVal = await configMod.Get('kitchen_mode_enabled')
    kitchenModeEnabled.value = kitchenVal === 'true'
  } catch { /* not available */ }
}

// ── Data Loading ──
async function loadCategories() {
  if (!DataService) return
  try {
    const result = await DataService.GetCategories()
    categories.value = result || []
  } catch (err) {
    console.error('Failed to load categories:', err)
  }
}

let SyncServiceMod: any = null
async function pollSyncStatus() {
  if (!SyncServiceMod) {
    try {
      SyncServiceMod = await import('../../bindings/coffeeshop-pos/internal/service/syncservice')
    } catch { return }
  }
  try {
    const status = await SyncServiceMod.GetSyncStatus()
    if (status.is_syncing) {
      syncStatus.value = 'syncing'
    } else {
      syncStatus.value = status.is_connected ? 'online' : 'offline'
    }
  } catch {
    syncStatus.value = 'offline'
  }
}

async function loadMenuItems(categoryId: string | null) {
  if (!DataService) return
  try {
    const result = await DataService.GetMenuItems(categoryId || '')
    menuItems.value = result || []
  } catch (err) {
    console.error('Failed to load menu items:', err)
  }
}

async function loadTodayOrders() {
  if (!OrderService) return
  try {
    const orders = await OrderService.GetTodayOrders()
    if (orders) todayOrders.value = orders
  } catch { /* silently ignore */ }
}

// ── Category Selection ──
function onCategorySelect(id: string | null) {
  selectedCategoryId.value = id
}

watch(selectedCategoryId, (id) => {
  loadMenuItems(id)
})

// ── Cart Actions ──
function onAddToCart(item: MenuItem) {
  addItem(item)
}

function onCheckout() {
  if (cartItems.value.length === 0) return
  showCheckout.value = true
}

async function onConfirmOrder(tableNumber: string) {
  if (!OrderService) {
    showToast('خطأ: خدمة الطلبات غير متوفرة')
    return
  }
  try {
    const cartData = cartItems.value.map(item => ({
      menu_item_id: item.menu_item_id,
      name_ar: item.name_ar,
      price: item.price,
      quantity: item.quantity,
    }))

    const order = await OrderService.CreateOrder(cartData, 'cash', tableNumber || '')
    if (order) {
      addOrder(order)
      if (ReceiptService) {
        try { await ReceiptService.PrintReceipt(order) }
        catch (err) { console.warn('Receipt printing failed:', err) }
      }
      showCheckout.value = false
      clearCart()
      showToast(`✓ تم الطلب ${order.order_number}`)
      loadMenuItems(selectedCategoryId.value)
    }
  } catch (err) {
    console.error('Failed to create order:', err)
    showToast('خطأ في إنشاء الطلب')
  }
}

// ── Toast ──
function showToast(message: string) {
  toastMessage.value = message
  toastVisible.value = true
  setTimeout(() => { toastVisible.value = false }, 3000)
}

// ── Polling ──
let queuePoll: ReturnType<typeof setInterval> | null = null
let statusPoll: ReturnType<typeof setInterval> | null = null

// ── Lifecycle ──
onMounted(async () => {
  await initBindings()
  await loadCategories()
  await loadMenuItems(null)
  await loadTodayOrders()
  await pollSyncStatus()

  queuePoll = setInterval(loadTodayOrders, 5000)
  statusPoll = setInterval(pollSyncStatus, 5000)
})

onUnmounted(() => {
  if (queuePoll) { clearInterval(queuePoll); queuePoll = null }
  if (statusPoll) { clearInterval(statusPoll); statusPoll = null }
})
</script>

<template>
  <div class="pos-view">
    <HeaderBar
      :order-count="acceptedOrders.length"
      :sync-status="syncStatus"
      :kitchen-mode-enabled="kitchenModeEnabled"
    />

    <div class="pos-body">
      <div class="pos-main">
        <CategoryTabs
          :categories="categories"
          :selected-id="selectedCategoryId"
          @select="onCategorySelect"
        />
        <MenuGrid
          :items="menuItems"
          @add-to-cart="onAddToCart"
        />
      </div>

      <CartPanel
        :items="cartItems"
        :total="total"
        :item-count="itemCount"
        @increment="incrementQty"
        @decrement="decrementQty"
        @remove="removeItem"
        @checkout="onCheckout"
      />
    </div>

    <OrderQueuePanel
      :accepted-orders="acceptedOrders"
      :completed-orders="recentCompletedOrders"
      :kitchen-mode-enabled="kitchenModeEnabled"
    />

    <CheckoutDialog
      v-if="showCheckout"
      :items="cartItems"
      :total="total"
      @confirm="onConfirmOrder"
      @cancel="showCheckout = false"
    />

    <Transition name="toast">
      <div v-if="toastVisible" class="toast">
        {{ toastMessage }}
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.pos-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.pos-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.pos-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.toast-enter-active {
  animation: slideUpFade var(--transition-base);
}
.toast-leave-active {
  animation: slideUpFade var(--transition-base) reverse;
}
</style>
