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
  if (n.includes('سكويات') || n.includes('أفوكاتو')) return '🍦'
  if (n.includes('كريب')) return '🥞'
  if (n.includes('وافل')) return '🧇'
  if (n.includes('صحن فواكه')) return '🍓'
  if (n.includes('ايس') || n.includes('كلد')) return '🧊'
  if (n.includes('شاي')) return '🍵'
  if (n.includes('لاتيه') || n.includes('كابتشينو')) return '☕'
  if (n.includes('موكا') || n.includes('جوكليت') || n.includes('شوكولا') || n.includes('هوت')) return '🍫'
  if (n.includes('قهوة') || n.includes('اسبريسو') || n.includes('امريكانو') || n.includes('كورتادو')) return '☕'
  if (n.includes('إن جي') || n.includes('ماشا') || n.includes('VIP')) return '⭐'
  return '☕'
}

function getCategoryClass(nameAr: string): string {
  const n = nameAr
  if (n.includes('ايس') || n.includes('كلد') || n.includes('موهيتو') || n.includes('فرابتشينو')) return 'cold'
  if (n.includes('كريب') || n.includes('وافل') || n.includes('سكويات') || n.includes('أفوكاتو') || n.includes('صحن')) return 'dessert'
  if (n.includes('VIP') || n.includes('إن جي') || n.includes('ماشا') || n.includes('معجون')) return 'vip'
  return 'hot'
}
</script>

<template>
  <button class="menu-card" :class="getCategoryClass(item.name_ar)" @click="onAdd">
    <div class="card-emoji-wrap">
      <span class="card-emoji">{{ getEmoji(item.name_ar) }}</span>
      <div class="emoji-glow"></div>
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
  border-radius: var(--radius-xl);
  padding: 14px 14px 14px 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: right;
  font-family: var(--font-family);
  color: var(--color-text);
  width: 100%;
  position: relative;
  overflow: hidden;
}
.menu-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(200,150,10,0.04) 0%, transparent 60%);
  pointer-events: none;
}
.menu-card:hover {
  border-color: var(--color-border-gold);
  background: var(--color-surface-2);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.3);
}
.menu-card:active { transform: scale(0.97); }

.card-emoji-wrap {
  position: relative;
  width: 60px; height: 60px;
  border-radius: 14px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
}
.menu-card.hot .card-emoji-wrap   { background: rgba(200,150,10,0.08); border-color: rgba(200,150,10,0.2); }
.menu-card.cold .card-emoji-wrap  { background: rgba(30,120,180,0.1);  border-color: rgba(30,120,180,0.2); }
.menu-card.dessert .card-emoji-wrap { background: rgba(160,80,180,0.08); border-color: rgba(160,80,180,0.2); }
.menu-card.vip .card-emoji-wrap   { background: rgba(200,150,10,0.15); border-color: rgba(200,150,10,0.4); }

.card-emoji { font-size: 1.9rem; position: relative; z-index: 2; }

.emoji-glow {
  position: absolute; inset: 0; border-radius: 14px;
  opacity: 0; transition: opacity 0.2s;
}
.menu-card.hot .emoji-glow    { background: radial-gradient(circle at 50% 80%, rgba(200,150,10,0.3), transparent 70%); }
.menu-card.cold .emoji-glow   { background: radial-gradient(circle at 50% 80%, rgba(30,120,220,0.3), transparent 70%); }
.menu-card.dessert .emoji-glow { background: radial-gradient(circle at 50% 80%, rgba(180,80,200,0.3), transparent 70%); }
.menu-card.vip .emoji-glow    { background: radial-gradient(circle at 50% 80%, rgba(200,150,10,0.4), transparent 70%); }
.menu-card:hover .emoji-glow { opacity: 1; }

.card-body { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 0; }
.card-name { font-size: 0.95rem; font-weight: 700; color: var(--color-text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.menu-card.vip .card-name { color: var(--color-accent); }
.card-price { font-size: 1rem; font-weight: 800; color: var(--color-accent); }
.card-price small { font-size: 0.65rem; font-weight: 600; opacity: 0.75; }

.card-action { flex-shrink: 0; }
.add-icon {
  display: flex; align-items: center; justify-content: center;
  width: 38px; height: 38px; border-radius: 50%;
  background: linear-gradient(135deg, #c8960a, #e0aa12);
  color: #0a1f12; font-size: 1.5rem; font-weight: 900;
  transition: all 0.2s ease;
  box-shadow: 0 3px 12px rgba(200,150,10,0.3);
}
.menu-card:hover .add-icon { transform: rotate(90deg) scale(1.1); }
</style>