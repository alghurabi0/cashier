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
    <div class="item-info">
      <span class="item-name">{{ item.name_ar }}</span>
      <span class="item-unit">{{ formatPrice(item.price) }} / قطعة</span>
    </div>
    <div class="item-controls">
      <button class="qty-btn" @click="emit('decrement', item.menu_item_id)">−</button>
      <span class="qty-val">{{ item.quantity }}</span>
      <button class="qty-btn" @click="emit('increment', item.menu_item_id)">+</button>
    </div>
    <span class="item-total">{{ formatPrice(item.price * item.quantity) }}</span>
    <button class="remove-btn" @click="emit('remove', item.menu_item_id)">✕</button>
  </div>
</template>

<style scoped>
.cart-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  position: relative;
  transition: background 0.15s ease;
}

.cart-item:hover {
  background: var(--color-surface-2);
  border-color: var(--color-border-light);
}

.item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-name {
  font-size: 0.88rem;
  font-weight: 700;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-unit {
  font-size: 0.7rem;
  color: var(--color-text-dim);
}

.item-controls {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--color-cart-bg);
  border-radius: 8px;
  padding: 3px;
}

.qty-btn {
  width: 26px;
  height: 26px;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  font-size: 1rem;
  font-family: inherit;
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.qty-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-accent);
}

.qty-val {
  min-width: 24px;
  text-align: center;
  font-size: 0.85rem;
  font-weight: 800;
  color: var(--color-text);
  font-variant-numeric: tabular-nums;
}

.item-total {
  font-size: 0.9rem;
  font-weight: 800;
  color: var(--color-accent);
  font-variant-numeric: tabular-nums;
  min-width: 56px;
  text-align: left;
}

.remove-btn {
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  color: var(--color-text-dim);
  cursor: pointer;
  font-size: 0.7rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.cart-item:hover .remove-btn { opacity: 1; }
.remove-btn:hover { color: var(--color-danger); background: rgba(231,76,60,0.12); }
</style>