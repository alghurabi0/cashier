<script setup lang="ts">
import type { Category } from '../types'

defineProps<{
  categories: Category[]
  selectedId: string | null
}>()

const emit = defineEmits<{
  select: [id: string | null]
}>()

const emojiMap: Record<string, string> = {
  'مشروبات ساخنة': '☕',
  'شاي': '🍵',
  'لاتيه': '☕',
  'كابتشينو': '☕',
  'آيس درنك': '🧊',
  'آيس لاتيه': '🥛',
  'فرابتشينو': '🥤',
  'شيكات': '🥛',
  'موهيتو': '🍃',
  'آيس كريم': '🍦',
  'كريب': '🥞',
  'وافل': '🧇',
  'VIP': '⭐',
}

function getEmoji(name: string): string {
  return emojiMap[name] ?? '🍽️'
}
</script>

<template>
  <div class="category-bar">
    <button
      class="cat-btn"
      :class="{ active: selectedId === null }"
      @click="emit('select', null)"
    >
      <span class="cat-emoji">🏠</span>
      <span class="cat-label">الكل</span>
    </button>
    <button
      v-for="cat in categories"
      :key="cat.id"
      class="cat-btn"
      :class="{ active: selectedId === cat.id }"
      @click="emit('select', cat.id)"
    >
      <span class="cat-emoji">{{ getEmoji(cat.name_ar) }}</span>
      <span class="cat-label">{{ cat.name_ar }}</span>
    </button>
  </div>
</template>

<style scoped>
.category-bar {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding: 14px 14px;
  scrollbar-width: none;
  background: var(--color-bg);
  border-bottom: 1px solid rgba(201, 168, 76, 0.15);
  position: sticky;
  top: 0;
  z-index: 40;
}

.category-bar::-webkit-scrollbar { display: none; }

.cat-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 14px;
  border: 1px solid var(--color-border);
  border-radius: 50px;
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-family: var(--font-family);
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.cat-btn:hover {
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.cat-btn.active {
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d;
  border-color: transparent;
  box-shadow: 0 3px 12px rgba(201, 168, 76, 0.35);
}

.cat-emoji { font-size: 1.1rem; }

.cat-label {
  font-size: 0.72rem;
  font-weight: 700;
}
</style>