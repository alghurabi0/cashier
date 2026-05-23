<script setup lang="ts">
import { useCart } from '../composables/useCart'
import { formatPrice } from '../types'

defineProps<{
  token: string
}>()

const emit = defineEmits<{
  back: []
  submitted: []
}>()

const { items, total, isSubmitting, orderError, incrementQty, decrementQty, removeItem, submitOrder } = useCart()

async function onSubmit(token: string) {
  await submitOrder(token)
  if (!orderError.value) {
    emit('submitted')
  }
}
</script>

<template>
  <div class="cart-view">
    <header class="cart-header">
      <button class="back-btn" @click="emit('back')">→ العودة</button>
      <h2 class="cart-title">🛒 السلة</h2>
    </header>

    <div class="cart-items" v-if="items.length > 0">
      <div v-for="item in items" :key="item.menu_item_id" class="cart-item">
        <div class="item-info">
          <span class="item-name">{{ item.name_ar }}</span>
          <span class="item-price text-muted">{{ formatPrice(item.price) }} د.ع</span>
        </div>
        <div class="item-controls">
          <button class="qty-btn" @click="decrementQty(item.menu_item_id)">−</button>
          <span class="qty-value">{{ item.quantity }}</span>
          <button class="qty-btn" @click="incrementQty(item.menu_item_id)">+</button>
        </div>
        <span class="item-total">{{ formatPrice(item.price * item.quantity) }}</span>
      </div>
    </div>

    <div v-else class="cart-empty">
      <span class="empty-icon">🛒</span>
      <span>السلة فارغة</span>
      <button class="btn btn-primary" @click="emit('back')">تصفح القائمة</button>
    </div>

    <div class="cart-footer" v-if="items.length > 0">
      <div class="total-row">
        <span class="total-label">المجموع</span>
        <span class="total-value">{{ formatPrice(total) }} <small>د.ع</small></span>
      </div>

      <div v-if="orderError" class="error-message">
        {{ orderError }}
      </div>

      <button
        class="btn btn-primary btn-lg submit-btn"
        :disabled="isSubmitting"
        @click="onSubmit(token)"
      >
        {{ isSubmitting ? 'جاري الإرسال...' : '📤 أرسل الطلب' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.cart-view {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
}

.cart-header {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-lg);
  border-bottom: 1px solid var(--color-border);
}

.back-btn {
  background: none;
  border: none;
  color: var(--color-accent);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  font-weight: 700;
  cursor: pointer;
}

.cart-title {
  font-size: var(--font-size-xl);
  font-weight: 800;
}

.cart-items {
  flex: 1;
  padding: var(--gap-lg);
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.cart-item {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-md);
  background: var(--color-surface);
  border-radius: var(--radius-md);
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-name {
  font-weight: 700;
}

.item-price {
  font-size: var(--font-size-sm);
}

.item-controls {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
}

.qty-btn {
  width: 32px;
  height: 32px;
  border: 1px solid var(--color-border);
  border-radius: 50%;
  background: var(--color-surface-2);
  color: var(--color-text);
  font-size: 1.1rem;
  font-weight: 800;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  font-family: var(--font-family);
}

.qty-btn:active {
  transform: scale(0.9);
}

.qty-value {
  min-width: 28px;
  text-align: center;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.item-total {
  font-weight: 800;
  color: var(--color-accent);
  min-width: 60px;
  text-align: left;
  font-variant-numeric: tabular-nums;
}

.cart-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--gap-lg);
  color: var(--color-text-dim);
  font-size: var(--font-size-lg);
}

.empty-icon {
  font-size: 3rem;
  opacity: 0.3;
}

.cart-footer {
  padding: var(--gap-lg);
  border-top: 1px solid var(--color-border);
  background: var(--color-surface);
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.total-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.total-label {
  font-size: var(--font-size-lg);
  font-weight: 700;
}

.total-value {
  font-size: var(--font-size-2xl);
  font-weight: 800;
  color: var(--color-accent);
}

.total-value small {
  font-size: var(--font-size-sm);
  opacity: 0.7;
}

.error-message {
  background: rgba(231, 76, 60, 0.12);
  color: var(--color-danger);
  padding: var(--gap-sm) var(--gap-md);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  text-align: center;
}

.submit-btn {
  width: 100%;
}
</style>
