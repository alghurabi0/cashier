<script setup lang="ts">
import type { MenuItem } from '../types'
import MenuItemCard from './MenuItemCard.vue'

defineProps<{
  items: MenuItem[]
}>()

const emit = defineEmits<{
  addToCart: [item: MenuItem]
}>()
</script>

<template>
  <div class="menu-grid-wrapper">
    <div v-if="items.length === 0" class="empty-state">
      <span class="empty-icon">📋</span>
      <span class="empty-text">لا توجد منتجات</span>
    </div>
    <div v-else class="menu-grid">
      <MenuItemCard
        v-for="item in items"
        :key="item.id"
        :item="item"
        @add="emit('addToCart', item)"
      />
    </div>
  </div>
</template>

<style scoped>
.menu-grid-wrapper {
  flex: 1;
  overflow-y: auto;
  padding: 0 var(--gap-lg) var(--gap-lg);
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: var(--gap-md);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: var(--gap-md);
  color: var(--color-text-dim);
}

.empty-icon {
  font-size: 3rem;
  opacity: 0.4;
}

.empty-text {
  font-size: var(--font-size-lg);
}
</style>
