<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMenu } from '../composables/useMenu'
import { useCart } from '../composables/useCart'
import CategoryBar from '../components/CategoryBar.vue'
import MenuItemCard from '../components/MenuItemCard.vue'
import ProductSheet from '../components/ProductSheet.vue'
import { t, languages, detectLang } from '../i18n'
import type { Lang } from '../i18n'
import type { MenuItem } from '../types'

const props = defineProps<{ tableNumber: string }>()
const emit = defineEmits<{ 'open-cart': [] }>()

const { categories, menuItems, isLoading, loadAll } = useMenu()
const { itemCount, addItem } = useCart()

const selectedCategoryId = ref<string | null>(null)
const showAll = ref(false)
const selectedItem = ref<MenuItem | null>(null)
const showLangPicker = ref(false)
const lang = ref<Lang>(detectLang())

const tr = computed(() => t[lang.value])
const dir = computed(() => languages.find(l => l.code === lang.value)?.dir ?? 'rtl')

watch(lang, (l) => {
  document.documentElement.lang = l
  document.documentElement.dir = dir.value
})

const selectedCategoryName = computed(() => {
  if (!selectedCategoryId.value) return tr.value.fullmenu
  return categories.value.find(c => c.id === selectedCategoryId.value)?.name_ar ?? tr.value.fullmenu
})

const filteredItems = computed(() => {
  if (!selectedCategoryId.value) return menuItems.value
  return menuItems.value.filter(i => i.category_id === selectedCategoryId.value)
})

const visibleItems = computed(() =>
  showAll.value ? filteredItems.value : filteredItems.value.slice(0, 5)
)

const hasMore = computed(() => filteredItems.value.length > 5 && !showAll.value)

onMounted(() => {
  loadAll()
  document.documentElement.lang = lang.value
  document.documentElement.dir = dir.value
})

function onSelectCategory(id: string | null) {
  selectedCategoryId.value = id
  showAll.value = false
}

function onSheetAdd(menuItemId: string, nameAr: string, price: number, qty: number) {
  for (let i = 0; i < qty; i++) addItem(menuItemId, nameAr, price)
}

function selectLang(l: Lang) {
  lang.value = l
  showLangPicker.value = false
}

function openWhatsapp() {
  const num = import.meta.env.VITE_WHATSAPP
  if (num) window.open(`https://wa.me/${num}`, '_blank')
}
</script>

<template>
  <div class="menu-view" :dir="dir">

    <header class="top-header">
      <div class="header-left">
        <div class="header-icon-wrap">🍽️</div>
        <div class="header-brand">
          <span class="brand-main">NJ</span>
          <span class="brand-sub">COFFEE</span>
        </div>
      </div>
      <div class="table-chip">
        <span>🪑</span>
        <span>{{ tr.table }} {{ tableNumber }}</span>
      </div>
    </header>

    <section class="hero-section">
      <div class="hero-card">
        <img class="hero-img" src="https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?w=900&q=85" alt="NJ Coffee" />
        <div class="hero-overlay" />
        <div class="hero-content">
          <p class="hero-eyebrow">✦ {{ tr.welcome }} ✦</p>
          <h1 class="hero-title">{{ tr.tagline }}</h1>
          <p class="hero-tagline">{{ tr.sub }}</p>
          <p class="hero-sub">{{ tr.herosub }}</p>
        </div>
        <div class="hero-dots">
          <span class="dot active" />
          <span class="dot" />
          <span class="dot" />
        </div>
      </div>
    </section>

    <CategoryBar
      :categories="categories"
      :selected-id="selectedCategoryId"
      :all-label="tr.all"
      @select="onSelectCategory"
    />

    <main class="menu-content">
      <div class="section-header" v-if="!isLoading">
        <div class="section-line" />
        <span class="section-title">
          <span class="section-star">✦</span>
          {{ selectedCategoryName }}
          <span class="section-star">✦</span>
        </span>
        <div class="section-line" />
      </div>

      <div v-if="isLoading" class="loading-state">
        <div class="loading-ring" />
        <span>{{ tr.loading }}</span>
      </div>

      <div v-else-if="filteredItems.length > 0" class="items-list">
        <MenuItemCard
          v-for="item in visibleItems"
          :key="item.id"
          :item="item"
          :lang="lang"
          @click="selectedItem = item"
        />
        <button v-if="hasMore" class="show-more-btn" @click="showAll = true">
          <span>{{ tr.showmore }}</span>
          <span class="more-count">({{ filteredItems.length - 5 }})</span>
        </button>
      </div>

      <div v-else class="empty-state">
        <span class="empty-icon">🍃</span>
        <p>{{ tr.empty }}</p>
      </div>
    </main>

    <nav class="bottom-nav">
      <button class="nav-btn nav-btn-cart" @click="emit('open-cart')">
        <div class="nav-icon-wrap">
          <span class="nav-icon">🛒</span>
          <span v-if="itemCount > 0" class="nav-badge">{{ itemCount }}</span>
        </div>
        <span class="nav-label">{{ tr.cart }}</span>
      </button>
      <button class="nav-btn" @click="showLangPicker = true">
        <span class="nav-icon">🌐</span>
        <span class="nav-label">{{ tr.language }}</span>
      </button>
      <button class="nav-btn" @click="openWhatsapp">
        <span class="nav-icon">📞</span>
        <span class="nav-label">{{ tr.contact }}</span>
      </button>
    </nav>

    <Transition name="sheet">
      <div v-if="showLangPicker" class="lang-backdrop" @click.self="showLangPicker = false">
        <div class="lang-sheet">
          <div class="lang-handle" />
          <h3 class="lang-title">{{ tr.language }}</h3>
          <div class="lang-list">
            <button
              v-for="l in languages"
              :key="l.code"
              class="lang-btn"
              :class="{ active: lang === l.code }"
              @click="selectLang(l.code)"
            >
              <span class="lang-label">{{ l.label }}</span>
              <span v-if="lang === l.code" class="lang-check">✓</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <ProductSheet
      :item="selectedItem"
      :tr="tr"
      :lang="lang"
      @close="selectedItem = null"
      @add="onSheetAdd"
    />

  </div>
</template>

<style scoped>
.menu-view {
  min-height: 100dvh;
  background: #0e0e0e;
  color: #f0e6d3;
  font-family: 'Cairo', sans-serif;
  display: flex;
  flex-direction: column;
  padding-bottom: 80px;
}

.top-header {
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: space-between;
  padding: 13px 18px;
  background: rgba(14,14,14,0.96);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(201,168,76,0.14);
}

.header-left { display: flex; align-items: center; gap: 10px; }

.header-icon-wrap {
  width: 38px; height: 38px; border-radius: 10px;
  background: rgba(201,168,76,0.12);
  border: 1px solid rgba(201,168,76,0.28);
  display: flex; align-items: center; justify-content: center;
  font-size: 1.1rem;
}

.header-brand { display: flex; flex-direction: column; line-height: 1; }
.brand-main { font-size: 1.15rem; font-weight: 900; color: #c9a84c; letter-spacing: 3px; }
.brand-sub { font-size: 0.55rem; font-weight: 600; color: rgba(201,168,76,0.55); letter-spacing: 4px; }

.table-chip {
  display: flex; align-items: center; gap: 5px;
  padding: 6px 13px;
  background: rgba(201,168,76,0.1);
  border: 1px solid rgba(201,168,76,0.25);
  border-radius: 50px;
  font-size: 0.75rem; font-weight: 700; color: #c9a84c;
}

.hero-section { padding: 14px 14px 0; }

.hero-card {
  position: relative; height: 215px; border-radius: 22px; overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.6), 0 0 0 1px rgba(201,168,76,0.1);
}

.hero-img {
  position: absolute; inset: 0; width: 100%; height: 100%;
  object-fit: cover; object-position: center 35%; filter: brightness(0.72);
}

.hero-overlay {
  position: absolute; inset: 0;
  background: linear-gradient(160deg, rgba(14,14,14,0.1) 0%, rgba(14,14,14,0.45) 50%, rgba(14,14,14,0.93) 100%);
}

.hero-content {
  position: relative; z-index: 2; height: 100%;
  display: flex; flex-direction: column; justify-content: flex-end;
  padding: 0 20px 18px; gap: 2px;
}

.hero-eyebrow { font-size: 0.68rem; color: #c9a84c; font-weight: 700; letter-spacing: 2px; margin: 0; }
.hero-title { font-size: 2rem; font-weight: 900; color: #f0e6d3; margin: 0; line-height: 1.1; }
.hero-tagline { font-size: 0.92rem; color: #c9a84c; margin: 0; font-weight: 700; }
.hero-sub { font-size: 0.7rem; color: rgba(240,230,211,0.55); margin: 0; }

.hero-dots { position: absolute; bottom: 14px; right: 18px; display: flex; gap: 5px; z-index: 3; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: rgba(255,255,255,0.28); }
.dot.active { background: #c9a84c; width: 18px; border-radius: 3px; }

.section-header { display: flex; align-items: center; gap: 10px; padding: 16px 16px 6px; }
.section-line { flex: 1; height: 1px; background: linear-gradient(to left, rgba(201,168,76,0.3), transparent); }
.section-line:first-child { background: linear-gradient(to right, rgba(201,168,76,0.3), transparent); }
.section-title { font-size: 0.9rem; font-weight: 800; color: #f0e6d3; white-space: nowrap; display: flex; align-items: center; gap: 6px; }
.section-star { color: #c9a84c; font-size: 0.6rem; }

.menu-content { flex: 1; display: flex; flex-direction: column; }
.items-list { display: flex; flex-direction: column; gap: 10px; padding: 6px 14px 16px; }

.show-more-btn {
  display: flex; align-items: center; justify-content: center; gap: 6px;
  margin: 6px auto 0; padding: 13px 36px;
  background: transparent; border: 1px solid rgba(201,168,76,0.35);
  border-radius: 50px; color: #c9a84c;
  font-family: 'Cairo', sans-serif; font-size: 0.9rem; font-weight: 700; cursor: pointer;
}
.show-more-btn:hover { background: rgba(201,168,76,0.08); border-color: #c9a84c; }
.more-count { opacity: 0.7; font-size: 0.8rem; }

.loading-state { display: flex; flex-direction: column; align-items: center; gap: 14px; padding: 60px 0; color: #666; }
.loading-ring { width: 44px; height: 44px; border: 3px solid rgba(201,168,76,0.15); border-top-color: #c9a84c; border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.empty-state { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 60px 20px; color: #666; text-align: center; }
.empty-icon { font-size: 2.5rem; }

.bottom-nav {
  position: fixed; bottom: 0; left: 0; right: 0; z-index: 60;
  display: flex; align-items: center; justify-content: space-around;
  padding: 10px 16px calc(10px + env(safe-area-inset-bottom));
  background: rgba(14,14,14,0.98); backdrop-filter: blur(20px);
  border-top: 1px solid rgba(201,168,76,0.12);
}

.nav-btn {
  position: relative; display: flex; flex-direction: column; align-items: center; gap: 3px;
  background: none; border: none; cursor: pointer; color: #555;
  font-family: 'Cairo', sans-serif; padding: 6px 16px; border-radius: 14px; min-width: 64px; transition: all 0.2s;
}
.nav-btn:hover { color: #c9a84c; background: rgba(201,168,76,0.07); }
.nav-btn-cart { color: #c9a84c; }
.nav-icon-wrap { position: relative; display: inline-flex; }
.nav-icon { font-size: 1.25rem; }
.nav-label { font-size: 0.68rem; font-weight: 700; }
.nav-badge {
  position: absolute; top: -5px; left: -5px;
  background: #c9a84c; color: #0d0d0d;
  font-size: 0.6rem; font-weight: 900;
  min-width: 17px; height: 17px; border-radius: 50px;
  display: flex; align-items: center; justify-content: center; padding: 0 3px;
}

.lang-backdrop {
  position: fixed; inset: 0; z-index: 200;
  background: rgba(0,0,0,0.7); backdrop-filter: blur(6px);
  display: flex; align-items: flex-end;
}

.lang-sheet {
  width: 100%; background: #1a1a1a;
  border-radius: 24px 24px 0 0; padding: 16px 20px 40px;
}

.lang-handle {
  width: 40px; height: 4px; border-radius: 2px;
  background: rgba(255,255,255,0.15); margin: 0 auto 20px;
}

.lang-title { font-size: 1rem; font-weight: 800; color: #f0e6d3; text-align: center; margin: 0 0 16px; }
.lang-list { display: flex; flex-direction: column; gap: 8px; }

.lang-btn {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 18px; background: #222;
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 14px; color: #f0e6d3;
  font-family: 'Cairo', sans-serif; font-size: 1rem; font-weight: 700; cursor: pointer; transition: all 0.2s;
}
.lang-btn.active { background: rgba(201,168,76,0.12); border-color: rgba(201,168,76,0.4); color: #c9a84c; }
.lang-check { color: #c9a84c; font-size: 1.1rem; }

.sheet-enter-active { transition: opacity 0.25s ease; }
.sheet-leave-active { transition: opacity 0.2s ease; }
.sheet-enter-from, .sheet-leave-to { opacity: 0; }
</style>