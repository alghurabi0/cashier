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
  'مشروبات باردة': '🧊',
  'فرابتشينو': '🥤',
  'شيكات': '🥛',
  'موهيتو': '🍃',
  'حلويات': '🍰',
  'آيس كريم': '🍦',
  'كريب': '🥞',
  'وافل': '🧇',
  'Drinks': '🥤',
}

function getEmoji(name: string) {
  return emojiMap[name] ?? '☕'
}
</script>

<template>
  <div class="cat-bar">
    <button class="cat-tab" :class="{ active: selectedId === null }" @click="emit('select', null)">
      <span class="cat-emoji">🏠</span>
      <span class="cat-label">الكل</span>
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
</template>

<style scoped>
.cat-bar {
  display: flex;
  gap: 8px;
  padding: 10px var(--gap-lg);
  overflow-x: auto;
  flex-shrink: 0;
  scrollbar-width: none;
  border-bottom: 1px solid rgba(255,255,255,0.04);
}

.cat-bar::-webkit-scrollbar { display: none; }

.cat-tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border: 1px solid rgba(201,168,76,0.15);
  border-radius: 999px;
  background: #1a1a1a;
  color: #666;
  font-family: var(--font-family);
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  transition: all 0.18s ease;
  min-width: 58px;
  user-select: none;
}

.cat-tab:hover {
  border-color: rgba(201,168,76,0.35);
  color: #c9a84c;
  background: #222;
}

.cat-tab.active {
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  border-color: #c9a84c;
  color: #0d0d0d;
  box-shadow: 0 4px 16px rgba(201,168,76,0.35);
}

.cat-emoji {
  font-size: 1.2rem;
  line-height: 1;
}

.cat-label {
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1;
}

.cat-tab.active .cat-label {
  color: #0d0d0d;
  font-weight: 800;
}
</style>