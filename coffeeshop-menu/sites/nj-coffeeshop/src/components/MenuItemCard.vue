<script setup lang="ts">
import type { MenuItem } from '../types'
import { formatPrice } from '../types'
import type { Lang } from '../i18n'

const props = defineProps<{
  item: MenuItem
  lang: Lang
}>()

function getImage(nameAr: string): string {
  const n = nameAr.toLowerCase()
  if (n.includes('فرابتشينو')) return 'https://images.unsplash.com/photo-1461023058943-07fcbe16d735?w=200&q=75'
  if (n.includes('موهيتو')) return 'https://images.unsplash.com/photo-1551538827-9c037cb4f32a?w=200&q=75'
  if (n.includes('شيك') || n.includes('ميلك')) return 'https://images.unsplash.com/photo-1572490122747-3968b75cc699?w=200&q=75'
  if (n.includes('ايس كريم') || n.includes('آيس كريم')) return 'https://images.unsplash.com/photo-1497034825429-c343d7c6a68f?w=200&q=75'
  if (n.includes('كريب')) return 'https://images.unsplash.com/photo-1519676867240-f03562e64548?w=200&q=75'
  if (n.includes('وافل')) return 'https://images.unsplash.com/photo-1562376552-0d160a2f238d?w=200&q=75'
  if (n.includes('ايس') || n.includes('آيس')) return 'https://images.unsplash.com/photo-1461023058943-07fcbe16d735?w=200&q=75'
  if (n.includes('شاي')) return 'https://images.unsplash.com/photo-1556679343-c7306c1976bc?w=200&q=75'
  if (n.includes('ماتشا') || n.includes('ماشا')) return 'https://images.unsplash.com/photo-1582793988951-9aed5509eb97?w=200&q=75'
  if (n.includes('لاتيه')) return 'https://images.unsplash.com/photo-1561882468-9110d70d0fb8?w=200&q=75'
  if (n.includes('كابتشينو')) return 'https://images.unsplash.com/photo-1572442388796-11668a67e53d?w=200&q=75'
  if (n.includes('موكا') || n.includes('شوكولا')) return 'https://images.unsplash.com/photo-1542990253-0d0f5be5f0ed?w=200&q=75'
  if (n.includes('زعفران') || n.includes('حليب')) return 'https://images.unsplash.com/photo-1563805042-7684c019e1cb?w=200&q=75'
  if (n.includes('اسبريسو') || n.includes('امريكانو')) return 'https://images.unsplash.com/photo-1510591509098-f4fdc6d0ff04?w=200&q=75'
  if (n.includes('vip')) return 'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?w=200&q=75'
  return 'https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?w=200&q=75'
}

const descriptions: Record<Lang, Record<string, string>> = {
  ar: {
    'فرابتشينو': 'مشروب مثلج مخفوق بالكريمة',
    'موهيتو': 'منعش بالنعناع والليمون',
    'شيك': 'ميلك شيك كريمي فاخر',
    'آيس كريم': 'آيس كريم طازج لذيذ',
    'كريب': 'كريب طري بحشوة مميزة',
    'وافل': 'وافل مقرمش بالمكملات',
    'لاتيه': 'لاتيه كريمي بنكهة رائعة',
    'كابتشينو': 'كابتشينو إيطالي أصيل',
    'موكا': 'قهوة بالشوكولاتة الداكنة',
    'شاي': 'شاي فاخر من أجود الأنواع',
    'امريكانو': 'أمريكانو غني بالنكهة',
    'اسبريسو': 'إسبريسو مركز عطري',
    'آيس': 'مشروب بارد منعش',
    'default': 'مشروب مميز من مجموعتنا',
  },
  en: {
    'فرابتشينو': 'Blended iced coffee with cream',
    'موهيتو': 'Fresh mint & lemon refresher',
    'شيك': 'Creamy premium milkshake',
    'آيس كريم': 'Fresh daily ice cream',
    'كريب': 'Soft crepe with special filling',
    'وافل': 'Crispy Belgian waffle',
    'لاتيه': 'Creamy latte art coffee',
    'كابتشينو': 'Authentic Italian cappuccino',
    'موكا': 'Rich dark chocolate coffee',
    'شاي': 'Premium tea blend',
    'امريكانو': 'Bold & smooth americano',
    'اسبريسو': 'Concentrated golden espresso',
    'آيس': 'Refreshing iced drink',
    'default': 'Special drink from our collection',
  },
  fa: {
    'فرابتشينو': 'نوشیدنی یخ با خامه',
    'موهيتو': 'موهیتو با نعناع و لیمو',
    'شيك': 'میلک شیک خامه‌ای',
    'آيس كريم': 'بستنی تازه روزانه',
    'كريب': 'کرپ نرم با مواد ویژه',
    'وافل': 'وافل بلژیکی ترد',
    'لاتيه': 'لاته خامه‌ای با آرت',
    'كابتشينو': 'کاپوچینو ایتالیایی اصیل',
    'موكا': 'قهوه با شکلات تلخ',
    'شاي': 'چای ممتاز',
    'امريكانو': 'آمریکانو قوی',
    'اسبريسو': 'اسپرسو غلیظ طلایی',
    'آيس': 'نوشیدنی خنک',
    'default': 'نوشیدنی ویژه از مجموعه ما',
  },
  zh: {
    'فرابتشينو': '奶油冰咖啡',
    'موهيتو': '薄荷柠檬清爽饮',
    'شيك': '奶油奶昔',
    'آيس كريم': '新鲜每日冰淇淋',
    'كريب': '软可丽饼配特色馅料',
    'وافل': '酥脆比利时华夫饼',
    'لاتيه': '拿铁咖啡艺术',
    'كابتشينو': '正宗意式卡布奇诺',
    'موكا': '浓郁黑巧克力咖啡',
    'شاي': '优质茶饮',
    'امريكانو': '浓郁美式咖啡',
    'اسبريسو': '浓缩金色意式咖啡',
    'آيس': '清爽冰饮',
    'default': '我们精选特色饮品',
  },
  tr: {
    'فرابتشينو': 'Kremalı buz kahvesi',
    'موهيتو': 'Nane limon serinletici',
    'شيك': 'Kremali milkshake',
    'آيس كريم': 'Taze günlük dondurma',
    'كريب': 'Özel dolgulu yumuşak krep',
    'وافل': 'Çıtır Belçika waffle',
    'لاتيه': 'Kremali latte art kahve',
    'كابتشينو': 'Otantik İtalyan cappuccino',
    'موكا': 'Zengin çikolatalı kahve',
    'شاي': 'Premium çay',
    'امريكانو': 'Güçlü americano',
    'اسبريسو': 'Yoğun altın espresso',
    'آيس': 'Serinletici buzlu içecek',
    'default': 'Koleksiyonumuzdan özel içecek',
  },
  fr: {
    'فرابتشينو': 'Café glacé mixé à la crème',
    'موهيتو': 'Mojito menthe citron',
    'شيك': 'Milkshake crémeux premium',
    'آيس كريم': 'Glace fraîche du jour',
    'كريب': 'Crêpe moelleuse garnie',
    'وافل': 'Gaufre belge croustillante',
    'لاتيه': 'Latte art crémeux',
    'كابتشينو': 'Cappuccino italien authentique',
    'موكا': 'Café chocolat noir intense',
    'شاي': 'Thé premium',
    'امريكانو': 'Américano corsé',
    'اسبريسو': 'Espresso concentré doré',
    'آيس': 'Boisson glacée rafraîchissante',
    'default': 'Boisson spéciale de notre collection',
  },
}

function getDescription(nameAr: string): string {
  const n = nameAr.toLowerCase()
  const d = descriptions[props.lang]
  if (n.includes('فرابتشينو')) return d['فرابتشينو']
  if (n.includes('موهيتو')) return d['موهيتو']
  if (n.includes('شيك') || n.includes('ميلك')) return d['شيك']
  if (n.includes('آيس كريم') || n.includes('ايس كريم')) return d['آيس كريم']
  if (n.includes('كريب')) return d['كريب']
  if (n.includes('وافل')) return d['وافل']
  if (n.includes('لاتيه')) return d['لاتيه']
  if (n.includes('كابتشينو')) return d['كابتشينو']
  if (n.includes('موكا') || n.includes('شوكولا')) return d['موكا']
  if (n.includes('شاي')) return d['شاي']
  if (n.includes('امريكانو')) return d['امريكانو']
  if (n.includes('اسبريسو')) return d['اسبريسو']
  if (n.includes('آيس') || n.includes('ايس')) return d['آيس']
  return d['default']
}
</script>

<template>
  <div class="menu-card">
    <div class="card-image-wrap">
      <img class="card-image" :src="getImage(item.name_ar)" :alt="item.name_ar" loading="lazy" />
    </div>
    <div class="card-body">
      <h3 class="card-name">{{ item.name_ar }}</h3>
      <p class="card-desc">{{ getDescription(item.name_ar) }}</p>
      <span class="card-price">{{ formatPrice(item.price) }} <small>د.ع</small></span>
    </div>
    <div class="card-action">
      <div class="view-btn"><span>+</span></div>
    </div>
  </div>
</template>

<style scoped>
.menu-card {
  display: flex; align-items: center; gap: 14px;
  padding: 11px 13px; background: #1c1c1c;
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 18px; cursor: pointer;
  direction: rtl; text-align: right;
  transition: all 0.2s ease; position: relative; overflow: hidden;
}

.menu-card::before {
  content: ''; position: absolute; top: 0; right: 0;
  width: 60px; height: 100%;
  background: linear-gradient(to left, rgba(201,168,76,0.04), transparent);
  pointer-events: none;
}

.menu-card:active { transform: scale(0.975); border-color: rgba(201,168,76,0.3); }

.card-image-wrap {
  flex-shrink: 0; width: 88px; height: 88px;
  border-radius: 14px; overflow: hidden; background: #111;
}

.card-image { width: 100%; height: 100%; object-fit: cover; display: block; transition: transform 0.4s ease; }
.menu-card:active .card-image { transform: scale(1.05); }

.card-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }

.card-name {
  font-size: 0.96rem; font-weight: 800; color: #f0e6d3;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin: 0;
}

.card-desc {
  font-size: 0.72rem; color: #666;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin: 0;
}

.card-price { font-size: 1rem; font-weight: 800; color: #c9a84c; margin-top: 3px; }
.card-price small { font-size: 0.63rem; font-weight: 600; opacity: 0.75; margin-right: 2px; }

.card-action { flex-shrink: 0; }

.view-btn {
  width: 40px; height: 40px; border-radius: 50%;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d; display: flex; align-items: center; justify-content: center;
  font-size: 1.5rem; font-weight: 300;
  box-shadow: 0 4px 14px rgba(201,168,76,0.35); transition: all 0.2s;
}

.menu-card:active .view-btn { transform: scale(0.9); }
</style>