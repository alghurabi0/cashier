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

function getEmoji(nameAr: string): string {
  const n = nameAr
  if (n.includes('فرابتشينو')) return '🥤'
  if (n.includes('موهيتو')) return '🍃'
  if (n.includes('ميلك شيك') || n.includes('شيك')) return '🥛'
  if (n.includes('ايس كريم') || n.includes('سكويات') || n.includes('أفوكاتو')) return '🍦'
  if (n.includes('كريب')) return '🥞'
  if (n.includes('وافل')) return '🧇'
  if (n.includes('صحن فواكه')) return '🍓'
  if (n.includes('ايس') || n.includes('كلد') || n.includes('بارد') || n.includes('إسبانيش')) return '🧊'
  if (n.includes('شاي')) return '🍵'
  if (n.includes('لاتيه')) return '☕'
  if (n.includes('كابتشينو')) return '☕'
  if (n.includes('موكا') || n.includes('جوكليت') || n.includes('شوكولا') || n.includes('هوت')) return '🍫'
  if (n.includes('قهوة') || n.includes('اسبريسو') || n.includes('امريكانو') || n.includes('كورتادو')) return '☕'
  if (n.includes('زعفران') || n.includes('حليب')) return '🥛'
  if (n.includes('VIP') || n.includes('إن جي') || n.includes('ماشا') || n.includes('معجون')) return '⭐'
  return '☕'
}
</script>

<template>
  <div class="menu-card" @click="onAdd">
    <div class="card-emoji-area">
      <span class="card-emoji">{{ getEmoji(item.name_ar) }}</span>
    </div>
    <div class="card-footer">
      <div class="card-info">
        <h3 class="card-name">{{ item.name_ar }}</h3>
        <span class="card-price">{{ formatPrice(item.price) }} <small>د.ع</small></span>
      </div>
      <button class="add-btn" @click.stop="onAdd">+</button>
    </div>
  </div>
</template>

<style scoped>
.menu-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 8px rgba(139, 94, 60, 0.08);
}

.menu-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(139, 94, 60, 0.15);
  border-color: var(--color-border-brown);
}

.menu-card:active {
  transform: scale(0.97);
}

.card-emoji-area {
  background: var(--color-surface-2);
  height: 110px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-emoji {
  font-size: 3rem;
}

.card-footer {
  padding: 10px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.card-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.card-name {
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-price {
  font-size: 0.9rem;
  font-weight: 800;
  color: var(--color-accent);
}

.card-price small {
  font-size: 0.65rem;
  font-weight: 600;
  opacity: 0.8;
}

.add-btn {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: none;
  background: var(--color-accent);
  color: #ffffff;
  font-size: 1.4rem;
  font-weight: 900;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s ease;
  box-shadow: 0 3px 10px rgba(139, 94, 60, 0.3);
}

.add-btn:hover {
  background: var(--color-accent-hover);
  transform: scale(1.1);
}
</style>