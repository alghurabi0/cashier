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
  background: #1e1e1e;
  border: 1px solid rgba(255,255,255,0.04);
  border-radius: 12px;
  position: relative;
  transition: background 0.15s ease;
}

.cart-item:hover {
  background: #242424;
  border-color: rgba(201,168,76,0.12);
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
  color: #e8dcc8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-unit {
  font-size: 0.7rem;
  color: #555;
}

.item-controls {
  display: flex;
  align-items: center;
  gap: 4px;
  background: #111;
  border-radius: 8px;
  padding: 3px;
}

.qty-btn {
  width: 26px;
  height: 26px;
  border: none;
  background: transparent;
  color: #888;
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
  background: #222;
  color: #c9a84c;
}

.qty-val {
  min-width: 24px;
  text-align: center;
  font-size: 0.85rem;
  font-weight: 800;
  color: #f0e6d3;
  font-variant-numeric: tabular-nums;
}

.item-total {
  font-size: 0.9rem;
  font-weight: 800;
  color: #c9a84c;
  font-variant-numeric: tabular-nums;
  min-width: 56px;
  text-align: left;
}

.remove-btn {
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  color: #444;
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
.remove-btn:hover { color: #e74c3c; background: rgba(231,76,60,0.12); }
</style>