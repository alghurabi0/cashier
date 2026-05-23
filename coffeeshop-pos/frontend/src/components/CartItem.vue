<script setup lang="ts">
import type { CartItem } from '../types'
import { formatPrice } from '../types'

defineProps<{
  item: CartItem
}>()

const emit = defineEmits<{
  increment: [menuItemId: string]
  decrement: [menuItemId: string]
  remove: [menuItemId: string]
}>()
</script>

<template>
  <div class="cart-item">
    <div class="cart-item-info">
      <span class="cart-item-name">{{ item.name_ar }}</span>
      <span class="cart-item-price text-muted text-sm">
        {{ formatPrice(item.price) }} × {{ item.quantity }}
      </span>
    </div>
    <div class="cart-item-controls">
      <div class="qty-controls">
        <button class="btn btn-icon btn-ghost" @click="emit('decrement', item.menu_item_id)">−</button>
        <span class="qty-value">{{ item.quantity }}</span>
        <button class="btn btn-icon btn-ghost" @click="emit('increment', item.menu_item_id)">+</button>
      </div>
      <span class="cart-item-total">{{ formatPrice(item.price * item.quantity) }}</span>
    </div>
    <button class="remove-btn" @click="emit('remove', item.menu_item_id)" title="حذف">✕</button>
  </div>
</template>

<style scoped>
.cart-item {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-md);
  background: var(--color-surface);
  border-radius: var(--radius-md);
  position: relative;
  transition: background var(--transition-fast);
}

.cart-item:hover {
  background: var(--color-surface-2);
}

.cart-item-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.cart-item-name {
  font-weight: var(--font-weight-semi);
  font-size: var(--font-size-md);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cart-item-controls {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
}

.qty-controls {
  display: flex;
  align-items: center;
  gap: var(--gap-xs);
  background: var(--color-bg);
  border-radius: var(--radius-sm);
  padding: 2px;
}

.qty-value {
  min-width: 28px;
  text-align: center;
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
}

.cart-item-total {
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  min-width: 60px;
  text-align: left;
  color: var(--color-accent);
}

.remove-btn {
  position: absolute;
  top: 4px;
  left: 4px;
  width: 20px;
  height: 20px;
  border: none;
  background: none;
  color: var(--color-text-dim);
  cursor: pointer;
  font-size: var(--font-size-xs);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: all var(--transition-fast);
}

.cart-item:hover .remove-btn {
  opacity: 1;
}

.remove-btn:hover {
  color: var(--color-danger);
  background: rgba(231, 76, 60, 0.15);
}
</style>
