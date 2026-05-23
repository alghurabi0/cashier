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
  <div class="category-bar">
    <button
      class="cat-btn"
      :class="{ active: selectedId === null }"
      @click="emit('select', null)"
    >
      الكل
    </button>
    <button
      v-for="cat in categories"
      :key="cat.id"
      class="cat-btn"
      :class="{ active: selectedId === cat.id }"
      @click="emit('select', cat.id)"
    >
      {{ cat.name_ar }}
    </button>
  </div>
</template>

<style scoped>
.category-bar {
  display: flex;
  gap: var(--gap-sm);
  overflow-x: auto;
  padding: var(--gap-sm) 0;
  scrollbar-width: none;
}
.category-bar::-webkit-scrollbar {
  display: none;
}

.cat-btn {
  white-space: nowrap;
  padding: var(--gap-sm) var(--gap-lg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: 700;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.cat-btn:hover {
  color: var(--color-text);
  border-color: var(--color-surface-3);
}

.cat-btn.active {
  background: var(--color-accent);
  color: var(--color-bg);
  border-color: var(--color-accent);
}
</style>
