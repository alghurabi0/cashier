<script setup lang="ts">
import type { Category } from '../types'

defineProps<{
  categories: Category[]
  selectedId: string | null
  allLabel: string
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

function getEmoji(nameAr: string): string {
  return emojiMap[nameAr] ?? '☕'
}
</script>

<template>
  <div class="category-bar-wrap">
    <div class="category-bar">
      <button
        class="cat-tab"
        :class="{ active: selectedId === null }"
        @click="emit('select', null)"
      >
        <span class="cat-emoji">🏠</span>
        <span class="cat-label">{{ allLabel }}</span>
      </button>

      <button
        v-for="cat in categories"
        :key="cat.id"
        class="cat-tab"
        :class="{ active: selectedId === cat.id }"
        @click="emit('select', cat.id)"
      >
        <span class="cat-emoji">{{ getEmoji(cat.name_ar) }}</span>
        <span class="cat-label">{{ cat.name_ar }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.category-bar-wrap {
  position: sticky;
  top: 64px;
  z-index: 40;
  background: #0e0e0e;
  padding: 12px 0 10px;
  border-bottom: 1px solid rgba(201,168,76,0.1);
}

.category-bar {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding: 0 14px;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
}

.category-bar::-webkit-scrollbar { display: none; }

.cat-tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 9px 14px;
  border: 1px solid rgba(201,168,76,0.15);
  border-radius: 50px;
  background: #1a1a1a;
  color: #666;
  font-family: 'Cairo', sans-serif;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  flex-shrink: 0;
  min-width: 60px;
}

.cat-tab:hover {
  border-color: rgba(201,168,76,0.35);
  color: #c9a84c;
}

.cat-tab.active {
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  border-color: #c9a84c;
  color: #0d0d0d;
  box-shadow: 0 4px 18px rgba(201,168,76,0.4);
  transform: translateY(-1px);
}

.cat-emoji { font-size: 1.3rem; line-height: 1; }

.cat-label { font-size: 0.65rem; font-weight: 700; line-height: 1; }

.cat-tab.active .cat-label { color: #0d0d0d; font-weight: 800; }
</style>