<script setup lang="ts">
import type { MenuItem } from '../types'
import { formatPrice } from '../types'

defineProps<{
  item: MenuItem
}>()

const emit = defineEmits<{
  add: [item: MenuItem]
}>()
</script>

<template>
  <button class="menu-item-card" :class="{ 'has-image': item.image_path }" @click="emit('add', item)">
    <img v-if="item.image_path" :src="item.image_path" :alt="item.name_ar" class="card-bg-image" loading="lazy" />
    <div class="card-body">
      <span class="item-name">{{ item.name_ar }}</span>
      <span class="item-price">{{ formatPrice(item.price) }} <small>د.ع</small></span>
    </div>
    <div class="card-add-hint">
      <span>+</span>
    </div>
  </button>
</template>

<style scoped>
.menu-item-card {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: var(--gap-lg);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
  min-height: 120px;
  font-family: var(--font-family);
  text-align: right;
  position: relative;
  overflow: hidden;
  user-select: none;
}

.menu-item-card.has-image {
  min-height: 140px;
  padding: 0;
}

.card-bg-image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  z-index: 0;
}

.has-image .card-body {
  position: relative;
  z-index: 2;
  background: linear-gradient(to top, rgba(0,0,0,0.75) 0%, rgba(0,0,0,0.3) 60%, transparent 100%);
  padding: var(--gap-lg);
  margin-top: auto;
}

.has-image .item-name {
  color: #fff;
  text-shadow: 0 1px 3px rgba(0,0,0,0.5);
}

.has-image .item-price {
  color: #ffcc80;
  text-shadow: 0 1px 3px rgba(0,0,0,0.5);
}

.menu-item-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, transparent 60%, var(--color-accent-glow));
  opacity: 0;
  transition: opacity var(--transition-fast);
  z-index: 1;
}

.has-image::before {
  display: none;
}

.menu-item-card:hover {
  border-color: var(--color-accent);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.menu-item-card:hover::before {
  opacity: 1;
}

.menu-item-card:active {
  transform: translateY(0) scale(0.98);
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
  position: relative;
  z-index: 1;
}

.item-name {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  line-height: 1.3;
}

.item-price {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-extra);
  color: var(--color-accent);
}

.item-price small {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-normal);
  opacity: 0.7;
}

.card-add-hint {
  position: absolute;
  bottom: var(--gap-sm);
  left: var(--gap-sm);
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--color-surface-2);
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-lg);
  opacity: 0;
  transition: all var(--transition-fast);
  z-index: 3;
}

.menu-item-card:hover .card-add-hint {
  opacity: 1;
  background: var(--color-accent);
  color: white;
}
</style>

