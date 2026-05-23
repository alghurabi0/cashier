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
      <h2 class="cart-title">السلة</h2>
      <span v-if="itemCount > 0" class="badge">{{ itemCount }}</span>
    </div>

    <div class="cart-items" v-if="items.length > 0">
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
      <span class="empty-icon">🛒</span>
      <span class="empty-text">السلة فارغة</span>
      <span class="empty-hint text-muted text-sm">اضغط على منتج لإضافته</span>
    </div>

    <div class="cart-footer" v-if="items.length > 0">
      <div class="cart-total-row">
        <span class="total-label">المجموع</span>
        <span class="total-value">{{ formatPrice(total) }} <small>د.ع</small></span>
      </div>
      <button class="btn btn-primary btn-lg checkout-btn" @click="emit('checkout')">
        إتمام الطلب
      </button>
    </div>
  </aside>
</template>

<style scoped>
.cart-panel {
  display: flex;
  flex-direction: column;
  width: 360px;
  min-width: 360px;
  background: var(--color-cart-bg);
  border-right: 1px solid var(--color-border);
  height: 100%;
}

.cart-header {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-lg);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.cart-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
}

.cart-items {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-sm);
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
}

.cart-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--gap-sm);
  color: var(--color-text-dim);
}

.cart-empty .empty-icon {
  font-size: 2.5rem;
  opacity: 0.3;
}

.cart-empty .empty-text {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semi);
}

.cart-footer {
  padding: var(--gap-lg);
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
}

.cart-total-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: var(--gap-md);
}

.total-label {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semi);
  color: var(--color-text-muted);
}

.total-value {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-extra);
  color: var(--color-accent);
}

.total-value small {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-normal);
  opacity: 0.7;
}

.checkout-btn {
  width: 100%;
  font-size: var(--font-size-lg);
  padding: var(--gap-md);
}
</style>
