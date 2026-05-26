<script setup lang="ts">
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
  <button class="menu-card" :class="{ 'has-image': item.image_path }" @click="onAdd">
    <div v-if="item.image_path" class="card-image">
      <img :src="item.image_path" :alt="item.name_ar" loading="lazy" />
    </div>
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
  padding: 0;
  display: flex;
  align-items: center;
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: right;
  font-family: var(--font-family);
  color: var(--color-text);
  width: 100%;
  overflow: hidden;
}

.menu-card:not(.has-image) {
  padding: var(--gap-lg);
}

.menu-card:active {
  transform: scale(0.97);
}

.menu-card:hover {
  border-color: var(--color-accent);
  background: var(--color-surface-2);
}

.card-image {
  width: 72px;
  height: 72px;
  flex-shrink: 0;
  overflow: hidden;
}

.card-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
  flex: 1;
  padding: var(--gap-md) var(--gap-lg);
}

.has-image .card-body {
  padding-right: var(--gap-md);
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
  padding: var(--gap-md);
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
