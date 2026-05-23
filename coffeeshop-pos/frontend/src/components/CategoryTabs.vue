<script setup lang="ts">
import type { Category } from '../types'

defineProps<{
  categories: Category[]
  selectedId: string | null
}>()

const emit = defineEmits<{
  select: [id: string | null]
}>()
</script>

<template>
  <div class="category-tabs">
    <button
      class="tab"
      :class="{ active: selectedId === null }"
      @click="emit('select', null)"
    >
      الكل
    </button>
    <button
      v-for="cat in categories"
      :key="cat.id"
      class="tab"
      :class="{ active: selectedId === cat.id }"
      @click="emit('select', cat.id)"
    >
      {{ cat.name_ar }}
    </button>
  </div>
</template>

<style scoped>
.category-tabs {
  display: flex;
  gap: var(--gap-sm);
  padding: var(--gap-md) var(--gap-lg);
  overflow-x: auto;
  flex-shrink: 0;
  scrollbar-width: none;
}

.category-tabs::-webkit-scrollbar {
  display: none;
}

.tab {
  padding: var(--gap-sm) var(--gap-lg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-full);
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semi);
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition-fast);
  user-select: none;
}

.tab:hover {
  background: var(--color-surface-2);
  color: var(--color-text);
  border-color: var(--color-surface-3);
}

.tab.active {
  background: var(--color-accent);
  color: white;
  border-color: var(--color-accent);
  box-shadow: var(--shadow-glow);
}
</style>
