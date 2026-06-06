<script setup lang="ts">
import type { MenuItem } from '../types'
import { formatPrice } from '../types'

const props = defineProps<{ item: MenuItem }>()
const emit = defineEmits<{ add: [menuItemId: string, nameAr: string, price: number] }>()

function onAdd() {
  emit('add', props.item.id, props.item.name_ar, props.item.price)
}

function getImage(nameAr: string): string {
  const n = nameAr
  if (n.includes('فرابتشينو')) return 'https://images.unsplash.com/photo-1572490122747-3968b75cc699?w=200&q=80'
  if (n.includes('موهيتو')) return 'https://images.unsplash.com/photo-1556679343-c7306c1976bc?w=200&q=80'
  if (n.includes('ميلك شيك') || n.includes('شيك')) return 'https://images.unsplash.com/photo-1563805042-7684c019e1cb?w=200&q=80'
  if (n.includes('وافل')) return 'https://images.unsplash.com/photo-1562376552-0d160a2f238d?w=200&q=80'
  if (n.includes('كريب')) return 'https://images.unsplash.com/photo-1519676867240-f03562e64548?w=200&q=80'
  if (n.includes('سكويات') || n.includes('أفوكاتو') || n.includes('ايس كريم')) return 'https://images.unsplash.com/photo-1497034825429-c343d7c6a68f?w=200&q=80'
  if (n.includes('شاي')) return 'https://images.unsplash.com/photo-1597481499750-3e6b22637e12?w=200&q=80'
  if (n.includes('ايس') || n.includes('إسبانيش') || n.includes('كلد')) return 'https://images.unsplash.com/photo-1517701604599-bb29b565090c?w=200&q=80'
  if (n.includes('موكا') || n.includes('جوكليت') || n.includes('شوكولا') || n.includes('هوت')) return 'https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=200&q=80'
  if (n.includes('لاتيه')) return 'https://images.unsplash.com/photo-1561882468-9110d70b1f3a?w=200&q=80'
  if (n.includes('كابتشينو')) return 'https://images.unsplash.com/photo-1534778101976-62847782c213?w=200&q=80'
  if (n.includes('صحن فواكه')) return 'https://images.unsplash.com/photo-1490474418585-ba9bad8fd0ea?w=200&q=80'
  if (n.includes('VIP') || n.includes('إن جي') || n.includes('ماشا') || n.includes('معجون')) return 'https://images.unsplash.com/photo-1509042239860-f550ce710b93?w=200&q=80'
  return 'https://images.unsplash.com/photo-1509042239860-f550ce710b93?w=200&q=80'
}

function getDescription(nameAr: string): string {
  const n = nameAr
  if (n.includes('اسبريسو سينكل')) return 'جرعة اسبريسو مركزة'
  if (n.includes('اسبريسو دبل')) return 'جرعتان من الاسبريسو'
  if (n.includes('امريكانو')) return 'اسبريسو وماء ساخن'
  if (n.includes('كورتادو')) return 'اسبريسو مع حليب ساخن'
  if (n.includes('لاتيه')) return 'اسبريسو مع حليب مبخر'
  if (n.includes('كابتشينو')) return 'اسبريسو مع فوم الحليب'
  if (n.includes('موكا')) return 'اسبريسو وشوكولاتة وحليب'
  if (n.includes('فرابتشينو')) return 'مشروب مثلج كريمي'
  if (n.includes('موهيتو')) return 'نعناع طازج وليمون منعش'
  if (n.includes('ميلك شيك') || n.includes('شيك')) return 'شيك كريمي بارد'
  if (n.includes('وافل')) return 'وافل طازج هش ولذيذ'
  if (n.includes('كريب')) return 'كريب طري بحشوة شهية'
  if (n.includes('شاي')) return 'شاي طازج عطري'
  return 'من أفضل المكونات الطازجة'
}
</script>

<template>
  <div class="menu-card" @click="onAdd">
    <img :src="getImage(item.name_ar)" :alt="item.name_ar" class="card-img" loading="lazy" />
    <div class="card-body">
      <h3 class="card-name">{{ item.name_ar }}</h3>
      <p class="card-desc">{{ getDescription(item.name_ar) }}</p>
      <span class="card-price">{{ formatPrice(item.price) }} <small>د.ع</small></span>
    </div>
    <button class="add-btn" @click.stop="onAdd">+</button>
  </div>
</template>

<style scoped>
.menu-card {
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.menu-card:hover {
  border-color: var(--color-accent);
  background: var(--color-surface-2);
  box-shadow: 0 4px 20px rgba(201, 168, 76, 0.12);
}

.menu-card:active { transform: scale(0.98); }

.card-img {
  width: 90px;
  height: 90px;
  border-radius: var(--radius-md);
  object-fit: cover;
  flex-shrink: 0;
  border: 1px solid var(--color-border);
}

.card-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.card-name {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-desc {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-price {
  font-size: 0.95rem;
  font-weight: 800;
  color: var(--color-accent);
  margin-top: 2px;
}

.card-price small {
  font-size: 0.65rem;
  opacity: 0.8;
}

.add-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: var(--color-accent);
  color: #0d0d0d;
  font-size: 1.4rem;
  font-weight: 900;
  cursor: pointer;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  box-shadow: 0 3px 10px rgba(201, 168, 76, 0.3);
}

.add-btn:hover {
  background: var(--color-accent-hover);
  transform: scale(1.1);
}
</style>