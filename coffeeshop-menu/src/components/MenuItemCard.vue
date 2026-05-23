<script setup lang="ts">
import { computed } from 'vue'
import type { MenuItem } from '../types'
import { formatPrice } from '../types'

const props = defineProps<{
  item: MenuItem
}>()

const emit = defineEmits<{
  add: [menuItemId: string, nameAr: string, price: number]
}>()

function onAdd() {
  emit('add', props.item.id, props.item.name_ar, props.item.price)
}
</script>

<template>
  <button class="menu-card" @click="onAdd">
    <div class="card-body">
      <h3 class="card-name">{{ item.name_ar }}</h3>
      <span class="card-price">{{ formatPrice(item.price) }} <small>د.ع</small></span>
    </div>
    <div class="card-action">
      <span class="add-icon">+</span>
    </div>
  </button>
</template>

<style scoped>
.menu-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--gap-lg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: right;
  font-family: var(--font-family);
  color: var(--color-text);
  width: 100%;
}

.menu-card:active {
  transform: scale(0.97);
}

.menu-card:hover {
  border-color: var(--color-accent);
  background: var(--color-surface-2);
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
}

.card-name {
  font-size: var(--font-size-md);
  font-weight: 700;
}

.card-price {
  font-size: var(--font-size-lg);
  font-weight: 800;
  color: var(--color-accent);
}

.card-price small {
  font-size: var(--font-size-xs);
  font-weight: 600;
  opacity: 0.7;
}

.card-action {
  flex-shrink: 0;
}

.add-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-accent);
  color: var(--color-bg);
  font-size: 1.3rem;
  font-weight: 800;
  transition: all var(--transition-fast);
}

.menu-card:hover .add-icon {
  transform: scale(1.1);
}
</style>
