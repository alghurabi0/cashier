<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import type { Category, MenuItem } from '../types'
import { useCart } from '../composables/useCart'
import { useOrders } from '../composables/useOrders'
import HeaderBar from '../components/HeaderBar.vue'
import CategoryTabs from '../components/CategoryTabs.vue'
import MenuGrid from '../components/MenuGrid.vue'
import CartPanel from '../components/CartPanel.vue'
import CheckoutDialog from '../components/CheckoutDialog.vue'

// ── State ──
const categories = ref<Category[]>([])
const menuItems = ref<MenuItem[]>([])
const selectedCategoryId = ref<string | null>(null)
const showCheckout = ref(false)
const toastMessage = ref('')
const toastVisible = ref(false)
const syncStatus = ref<'online' | 'offline' | 'syncing'>('offline')

const { items: cartItems, total, itemCount, addItem, removeItem, incrementQty, decrementQty, clear: clearCart } = useCart()
const { todayOrders, addOrder } = useOrders()

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
}

// ── Data Loading ──
async function loadCategories() {
  if (!DataService) return
  try {
    const result = await DataService.GetCategories()
    categories.value = result || []
    syncStatus.value = 'online'
  } catch (err) {
    console.error('Failed to load categories:', err)
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

async function onConfirmOrder() {
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

    const order = await OrderService.CreateOrder(cartData, 'cash')
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

// ── Lifecycle ──
onMounted(async () => {
  await initBindings()
  await loadCategories()
  await loadMenuItems(null)
  await loadTodayOrders()
})
</script>

<template>
  <div class="pos-view">
    <HeaderBar
      :order-count="todayOrders.length"
      :sync-status="syncStatus"
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
