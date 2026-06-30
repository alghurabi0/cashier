<script setup lang="ts">
import type { CartItem as CartItemType } from '../types'
import { formatPrice } from '../types'
import CartItemComponent from './CartItem.vue'

defineProps<{
  items: CartItemType[]
  total: number
  itemCount: number
}>()

const emit = defineEmits<{
  increment: [menuItemId: string]
  decrement: [menuItemId: string]
  remove: [menuItemId: string]
  checkout: []
}>()
</script>

<template>
  <aside class="cart-panel">
    <div class="cart-header">
      <div class="cart-title-row">
        <span class="cart-icon">🛒</span>
        <h2 class="cart-title">السلة</h2>
        <span v-if="itemCount > 0" class="cart-badge">{{ itemCount }}</span>
      </div>
    </div>

    <div class="cart-body">
      <div v-if="items.length > 0" class="cart-items">
        <CartItemComponent
          v-for="item in items"
          :key="item.menu_item_id"
          :item="item"
          @increment="emit('increment', $event)"
          @decrement="emit('decrement', $event)"
          @remove="emit('remove', $event)"
        />
      </div>

      <div v-else class="cart-empty">
        <div class="empty-icon">🛒</div>
        <span class="empty-text">السلة فارغة</span>
        <span class="empty-hint">اضغط على منتج لإضافته</span>
      </div>
    </div>

    <div class="cart-footer" v-if="items.length > 0">
      <div class="total-box">
        <span class="total-label">المجموع</span>
        <span class="total-value">{{ formatPrice(total) }} <small>د.ع</small></span>
      </div>
      <button class="checkout-btn" @click="emit('checkout')">
        ✓ إتمام الطلب
      </button>
    </div>
  </aside>
</template>

<style scoped>
.cart-panel {
  display: flex;
  flex-direction: column;
  width: 320px;
  min-width: 320px;
  background: var(--color-cart-bg);
  border-right: 1px solid var(--color-border-light);
  height: 100%;
}

.cart-header {
  padding: 16px 16px 12px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.cart-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cart-icon { font-size: 1.1rem; }

.cart-title {
  font-size: 1rem;
  font-weight: 800;
  color: var(--color-text);
}

.cart-badge {
  background: var(--color-accent);
  color: #0d0d0d;
  font-size: 0.7rem;
  font-weight: 800;
  padding: 2px 7px;
  border-radius: 999px;
  margin-right: auto;
}

.cart-body {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.cart-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cart-empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 0;
}

.empty-icon {
  font-size: 2.5rem;
  opacity: 0.15;
}

.empty-text {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--color-text-dim);
}

.empty-hint {
  font-size: 0.72rem;
  color: var(--color-text-dim);
}

.cart-footer {
  padding: 14px 16px;
  border-top: 1px solid var(--color-border-light);
  flex-shrink: 0;
}

.total-box {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 12px;
}

.total-label {
  font-size: 0.82rem;
  color: var(--color-text-muted);
  font-weight: 600;
}

.total-value {
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--color-accent);
  font-variant-numeric: tabular-nums;
}

.total-value small {
  font-size: 0.65rem;
  opacity: 0.7;
  margin-right: 2px;
}

.checkout-btn {
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, var(--color-accent), var(--color-accent-hover));
  color: #0d0d0d;
  font-family: inherit;
  font-size: 0.95rem;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.18s ease;
  box-shadow: 0 4px 16px var(--color-accent-glow);
}

.checkout-btn:hover {
  filter: brightness(1.08);
  box-shadow: 0 6px 24px var(--color-accent-glow);
}

.checkout-btn:active { transform: scale(0.98); }
</style>