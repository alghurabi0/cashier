<script setup lang="ts">
import { ref, watch } from 'vue'
import type { MenuItem } from '../types'
import { formatPrice } from '../types'
import type { Lang } from '../i18n'

const props = defineProps<{
  item: MenuItem | null
  tr: Record<string, string>
  lang: Lang
}>()

const emit = defineEmits<{
  close: []
  add: [menuItemId: string, nameAr: string, price: number, qty: number]
}>()

const qty = ref(1)
watch(() => props.item, () => { qty.value = 1 })

function getImage(nameAr: string): string {
  const n = nameAr.toLowerCase()
  if (n.includes('فرابتشينو')) return 'https://images.unsplash.com/photo-1461023058943-07fcbe16d735?w=600&q=85'
  if (n.includes('موهيتو')) return 'https://images.unsplash.com/photo-1551538827-9c037cb4f32a?w=600&q=85'
  if (n.includes('شيك') || n.includes('ميلك')) return 'https://images.unsplash.com/photo-1572490122747-3968b75cc699?w=600&q=85'
  if (n.includes('ايس كريم') || n.includes('آيس كريم')) return 'https://images.unsplash.com/photo-1497034825429-c343d7c6a68f?w=600&q=85'
  if (n.includes('كريب')) return 'https://images.unsplash.com/photo-1519676867240-f03562e64548?w=600&q=85'
  if (n.includes('وافل')) return 'https://images.unsplash.com/photo-1562376552-0d160a2f238d?w=600&q=85'
  if (n.includes('ايس') || n.includes('آيس')) return 'https://images.unsplash.com/photo-1461023058943-07fcbe16d735?w=600&q=85'
  if (n.includes('شاي')) return 'https://images.unsplash.com/photo-1556679343-c7306c1976bc?w=600&q=85'
  if (n.includes('ماتشا') || n.includes('ماشا')) return 'https://images.unsplash.com/photo-1582793988951-9aed5509eb97?w=600&q=85'
  if (n.includes('لاتيه')) return 'https://images.unsplash.com/photo-1561882468-9110d70d0fb8?w=600&q=85'
  if (n.includes('كابتشينو')) return 'https://images.unsplash.com/photo-1572442388796-11668a67e53d?w=600&q=85'
  if (n.includes('موكا') || n.includes('شوكولا')) return 'https://images.unsplash.com/photo-1542990253-0d0f5be5f0ed?w=600&q=85'
  if (n.includes('زعفران') || n.includes('حليب')) return 'https://images.unsplash.com/photo-1563805042-7684c019e1cb?w=600&q=85'
  if (n.includes('اسبريسو') || n.includes('امريكانو')) return 'https://images.unsplash.com/photo-1510591509098-f4fdc6d0ff04?w=600&q=85'
  if (n.includes('vip')) return 'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?w=600&q=85'
  return 'https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?w=600&q=85'
}

const descriptions: Record<Lang, Record<string, string>> = {
  ar: {
    'فرابتشينو': 'مشروب مثلج مخفوق بعناية، مزيج من القهوة الطازجة والكريمة الناعمة',
    'موهيتو': 'مشروب منعش بارد من النعناع الطازج والليمون الحامض',
    'شيك': 'ميلك شيك كريمي فاخر من أجود أنواع الآيس كريم',
    'آيس كريم': 'آيس كريم طازج يومياً من ألبان طبيعية عالية الجودة',
    'كريب': 'كريب فرنسي طري محضّر لحظياً بحشوة مميزة',
    'وافل': 'وافل بلجيكي ذهبي مقرمش مع الكريمة والعسل',
    'لاتيه': 'لاتيه كريمي ناعم مع إسبريسو مزدوج ورسمة فاخرة',
    'كابتشينو': 'كابتشينو إيطالي أصيل بطبقة رغوة كثيفة ومخملية',
    'موكا': 'قهوة موكا غنية بالشوكولاتة الداكنة والإسبريسو العميق',
    'شاي': 'شاي فاخر من أجود المزارع العالمية مع نعناع طازج',
    'امريكانو': 'أمريكانو كلاسيكي قوي من إسبريسو مزدوج',
    'اسبريسو': 'إسبريسو مركّز بطبقة كريما ذهبية أصيلة',
    'آيس': 'مشروب بارد منعش على ثلج صافٍ',
    'default': 'مشروب مميز محضّر من أجود المكونات الطازجة',
  },
  en: {
    'فرابتشينو': 'Blended iced coffee with fresh cream and cocoa',
    'موهيتو': 'Refreshing cold drink with fresh mint and lemon',
    'شيك': 'Premium creamy milkshake with finest ice cream',
    'آيس كريم': 'Fresh daily ice cream from natural dairy',
    'كريب': 'Soft French crepe prepared fresh with special filling',
    'وافل': 'Golden crispy Belgian waffle with cream and honey',
    'لاتيه': 'Creamy smooth latte with double espresso and latte art',
    'كابتشينو': 'Authentic Italian cappuccino with thick velvety foam',
    'موكا': 'Rich mocha with dark chocolate and deep espresso',
    'شاي': 'Premium tea from the world finest farms',
    'امريكانو': 'Classic bold americano from double espresso',
    'اسبريسو': 'Concentrated espresso with golden crema',
    'آيس': 'Refreshing iced drink served on clear ice',
    'default': 'Special drink crafted from the finest fresh ingredients',
  },
  fa: {
    'فرابتشينو': 'قهوه یخ مخلوط با خامه تازه',
    'موهيتو': 'نوشیدنی خنک با نعناع و لیمو',
    'شيك': 'میلک شیک خامه‌ای از بهترین بستنی',
    'آيس كريم': 'بستنی تازه روزانه از لبنیات طبیعی',
    'كريب': 'کرپ فرانسوی نرم با مواد ویژه',
    'وافل': 'وافل بلژیکی طلایی با خامه و عسل',
    'لاتيه': 'لاته خامه‌ای با دابل اسپرسو',
    'كابتشينو': 'کاپوچینو ایتالیایی اصیل با کف مخملی',
    'موكا': 'موکا غنی با شکلات تلخ و اسپرسو',
    'شاي': 'چای ممتاز از بهترین مزارع جهان',
    'امريكانو': 'آمریکانو کلاسیک قوی',
    'اسبريسو': 'اسپرسو غلیظ با کرما طلایی',
    'آيس': 'نوشیدنی خنک و گوارا',
    'default': 'نوشیدنی ویژه از بهترین مواد تازه',
  },
  zh: {
    'فرابتشينو': '奶油冰咖啡混合饮品',
    'موهيتو': '薄荷柠檬清爽冷饮',
    'شيك': '顶级奶油奶昔',
    'آيس كريم': '天然乳制品新鲜冰淇淋',
    'كريب': '新鲜法式软可丽饼配特色馅料',
    'وافل': '金黄酥脆比利时华夫饼配奶油蜂蜜',
    'لاتيه': '奶油顺滑拿铁咖啡艺术',
    'كابتشينو': '正宗意式卡布奇诺配丝绒泡沫',
    'موكا': '浓郁黑巧克力摩卡咖啡',
    'شاي': '来自世界最佳茶园的优质茶',
    'امريكانو': '经典浓郁美式咖啡',
    'اسبريسو': '金色克丽玛浓缩意式咖啡',
    'آيس': '清澈冰块上的清爽冰饮',
    'default': '精选最优质新鲜食材特调饮品',
  },
  tr: {
    'فرابتشينو': 'Taze krema ile karıştırılmış buz kahvesi',
    'موهيتو': 'Taze nane ve limonlu soğuk içecek',
    'شيك': 'En iyi dondurmayla kremali milkshake',
    'آيس كريم': 'Doğal sütten günlük taze dondurma',
    'كريب': 'Özel dolgulu taze Fransız krebi',
    'وافل': 'Krema ve ballı altın çıtır Belçika waffle',
    'لاتيه': 'Çift espressolu kremali latte art',
    'كابتشينو': 'Kadifemsi köpüklü otantik İtalyan cappuccino',
    'موكا': 'Bitter çikolata ve derin espressolu zengin moka',
    'شاي': 'Dünyanın en iyi çiftliklerinden premium çay',
    'امريكانو': 'Çift espressodan klasik bold americano',
    'اسبريسو': 'Altın kremalı konsantre espresso',
    'آيس': 'Berrak buzda sergilenen serinletici içecek',
    'default': 'En taze malzemelerden özel içecek',
  },
  fr: {
    'فرابتشينو': 'Café glacé mixé à la crème fraîche',
    'موهيتو': 'Boisson fraîche à la menthe et citron',
    'شيك': 'Milkshake crémeux premium',
    'آيس كريم': 'Glace fraîche du jour au lait naturel',
    'كريب': 'Crêpe française moelleuse garnie',
    'وافل': 'Gaufre belge dorée croustillante avec crème et miel',
    'لاتيه': 'Latte art crémeux avec double espresso',
    'كابتشينو': 'Cappuccino italien authentique à mousse veloutée',
    'موكا': 'Moka riche au chocolat noir et espresso profond',
    'شاي': 'Thé premium des meilleures fermes du monde',
    'امريكانو': 'Américano classique corsé',
    'اسبريسو': 'Espresso concentré à la crema dorée',
    'آيس': 'Boisson glacée rafraîchissante sur glace claire',
    'default': 'Boisson spéciale des meilleurs ingrédients frais',
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

function onAdd() {
  if (!props.item) return
  emit('add', props.item.id, props.item.name_ar, props.item.price, qty.value)
  emit('close')
}
</script>

<template>
  <Transition name="sheet">
    <div v-if="item" class="sheet-backdrop" @click.self="emit('close')">
      <div class="sheet">
        <div class="sheet-hero">
          <img :src="getImage(item.name_ar)" :alt="item.name_ar" class="sheet-img" />
          <div class="sheet-img-overlay" />
          <button class="sheet-close" @click="emit('close')">✕</button>
          <div class="sheet-badge">{{ tr.featured }}</div>
        </div>

        <div class="sheet-body">
          <h2 class="sheet-name">{{ item.name_ar }}</h2>
          <p class="sheet-desc">{{ getDescription(item.name_ar) }}</p>

          <div class="sheet-tags">
            <span class="tag">{{ tr.fresh }}</span>
            <span class="tag">{{ tr.special }}</span>
            <span class="tag">{{ tr.fast }}</span>
          </div>

          <div class="sheet-footer">
            <div class="qty-control">
              <button class="qty-btn" @click="qty = Math.max(1, qty - 1)">−</button>
              <span class="qty-num">{{ qty }}</span>
              <button class="qty-btn" @click="qty++">+</button>
            </div>
            <button class="add-to-cart-btn" @click="onAdd">
              <span>{{ tr.addtocart }}</span>
              <span class="btn-price">{{ formatPrice(item.price * qty) }} {{ tr.currency }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.sheet-backdrop {
  position: fixed; inset: 0; z-index: 200;
  background: rgba(0,0,0,0.78); backdrop-filter: blur(8px);
  display: flex; align-items: flex-end;
}

.sheet {
  width: 100%; background: #1a1a1a;
  border-radius: 28px 28px 0 0; overflow: hidden;
  max-height: 90dvh; display: flex; flex-direction: column;
}

.sheet-hero { position: relative; height: 270px; flex-shrink: 0; }
.sheet-img { width: 100%; height: 100%; object-fit: cover; }
.sheet-img-overlay {
  position: absolute; inset: 0;
  background: linear-gradient(to bottom, rgba(0,0,0,0.1) 0%, rgba(26,26,26,0.9) 100%);
}

.sheet-close {
  position: absolute; top: 14px; left: 14px;
  width: 36px; height: 36px; border-radius: 50%; border: none;
  background: rgba(0,0,0,0.6); color: #fff; font-size: 0.9rem;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  backdrop-filter: blur(4px); z-index: 2;
}

.sheet-badge {
  position: absolute; top: 14px; right: 14px;
  padding: 5px 12px;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d; font-size: 0.72rem; font-weight: 800;
  border-radius: 50px; z-index: 2;
}

.sheet-body {
  padding: 22px 20px 36px; overflow-y: auto;
  display: flex; flex-direction: column; gap: 14px;
  direction: rtl; text-align: right;
}

.sheet-name { font-size: 1.55rem; font-weight: 900; color: #f0e6d3; margin: 0; }
.sheet-desc { font-size: 0.88rem; color: #888; line-height: 1.75; margin: 0; }

.sheet-tags { display: flex; gap: 8px; flex-wrap: wrap; }
.tag {
  padding: 5px 13px;
  background: rgba(201,168,76,0.09);
  border: 1px solid rgba(201,168,76,0.22);
  border-radius: 50px; font-size: 0.75rem; color: #c9a84c; font-weight: 700;
}

.sheet-footer { display: flex; align-items: center; gap: 12px; margin-top: 4px; }

.qty-control {
  display: flex; align-items: center;
  background: #111; border: 1px solid rgba(201,168,76,0.2);
  border-radius: 50px; overflow: hidden; flex-shrink: 0;
}

.qty-btn {
  width: 46px; height: 46px; border: none; background: transparent;
  color: #c9a84c; font-size: 1.5rem; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
}
.qty-btn:active { background: rgba(201,168,76,0.12); }
.qty-num { min-width: 34px; text-align: center; font-size: 1rem; font-weight: 800; color: #f0e6d3; }

.add-to-cart-btn {
  flex: 1; display: flex; align-items: center; justify-content: space-between;
  padding: 14px 22px;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  border: none; border-radius: 50px; color: #0d0d0d;
  font-family: 'Cairo', sans-serif; font-size: 0.98rem; font-weight: 800;
  cursor: pointer; box-shadow: 0 8px 28px rgba(201,168,76,0.4); transition: all 0.2s;
}
.add-to-cart-btn:active { transform: scale(0.97); }
.btn-price { font-size: 0.83rem; font-weight: 700; opacity: 0.8; }

.sheet-enter-active { transition: opacity 0.28s ease; }
.sheet-leave-active { transition: opacity 0.22s ease; }
.sheet-enter-active .sheet { transition: transform 0.35s cubic-bezier(0.32,0.72,0,1); }
.sheet-leave-active .sheet { transition: transform 0.25s ease-in; }
.sheet-enter-from, .sheet-leave-to { opacity: 0; }
.sheet-enter-from .sheet, .sheet-leave-to .sheet { transform: translateY(100%); }
</style>