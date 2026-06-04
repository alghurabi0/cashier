<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMenu } from '../composables/useMenu'
import { useCart } from '../composables/useCart'
import CategoryBar from '../components/CategoryBar.vue'
import MenuItemCard from '../components/MenuItemCard.vue'

const props = defineProps<{
  tableNumber: string
}>()

const emit = defineEmits<{
  'open-cart': []
}>()

const { categories, menuItems, isLoading, loadAll } = useMenu()
const { itemCount, addItem } = useCart()

const selectedCategoryId = ref<string | null>(null)

const filteredItems = computed(() => {
  if (!selectedCategoryId.value) return menuItems.value
  return menuItems.value.filter(i => i.category_id === selectedCategoryId.value)
})

onMounted(() => { loadAll() })

function onAddItem(menuItemId: string, nameAr: string, price: number) {
  addItem(menuItemId, nameAr, price)
}
</script>

<template>
  <div class="menu-view">
    <header class="menu-header">
      <div class="header-logo">
        <div class="logo-circle"><span class="logo-text">NJ</span></div>
        <div class="header-titles">
          <h1 class="shop-name">NJ Coffee</h1>
          <p class="shop-tagline">اختر ما يحلو لك ☕</p>
        </div>
      </div>
      <div class="table-badge">
        <span>🪑</span>
        <span>طاولة {{ tableNumber }}</span>
      </div>
    </header>

    <div class="menu-content">
      <CategoryBar
        :categories="categories"
        :selected-id="selectedCategoryId"
        @select="selectedCategoryId = $event"
      />
      <div v-if="isLoading" class="loading">
        <div class="loading-ring"></div>
        <span>جاري التحميل...</span>
      </div>
      <div v-else class="menu-grid">
        <MenuItemCard
          v-for="item in filteredItems"
          :key="item.id"
          :item="item"
          @add="onAddItem"
        />
      </div>
      <div v-if="!isLoading && filteredItems.length === 0" class="empty">
        لا توجد منتجات
      </div>
    </div>

    <button v-if="itemCount > 0" class="fab-cart" @click="emit('open-cart')">
      <span>🛒</span>
      <span class="fab-count">{{ itemCount }}</span>
      <span>عرض السلة</span>
    </button>
  </div>
</template>

<style scoped>
.menu-view { min-height: 100dvh; display: flex; flex-direction: column; background: var(--color-bg); }

.menu-header {
  padding: var(--gap-lg) var(--gap-lg) var(--gap-md);
  background: linear-gradient(160deg, #0d2918 0%, #0a1f12 100%);
  border-bottom: 1px solid var(--color-border-gold);
  display: flex; align-items: center; justify-content: space-between;
  position: sticky; top: 0; z-index: 50;
}
.header-logo { display: flex; align-items: center; gap: var(--gap-md); }
.logo-circle {
  width: 52px; height: 52px; border-radius: 50%;
  background: linear-gradient(135deg, #c8960a, #e0aa12);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 0 20px rgba(200,150,10,0.4); flex-shrink: 0;
}
.logo-text { font-size: 1.2rem; font-weight: 900; color: #0a1f12; letter-spacing: 1px; }
.header-titles { display: flex; flex-direction: column; }
.shop-name { font-size: 1.4rem; font-weight: 900; color: var(--color-accent); letter-spacing: 1px; line-height: 1.1; }
.shop-tagline { font-size: var(--font-size-sm); color: var(--color-text-muted); margin-top: 2px; }
.table-badge {
  display: flex; align-items: center; gap: var(--gap-xs);
  background: var(--color-accent-light); border: 1px solid var(--color-border-gold);
  padding: var(--gap-xs) var(--gap-md); border-radius: var(--radius-full);
  font-size: var(--font-size-sm); font-weight: 700; color: var(--color-accent);
}

.menu-content { flex: 1; padding: var(--gap-md) var(--gap-md) 100px; display: flex; flex-direction: column; gap: var(--gap-md); }
.menu-grid { display: flex; flex-direction: column; gap: var(--gap-md); }
@media (min-width: 600px) { .menu-grid { display: grid; grid-template-columns: repeat(2, 1fr); } }

.loading { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: var(--gap-md); padding: var(--gap-xl) 0; color: var(--color-text-muted); }
.loading-ring { width: 40px; height: 40px; border: 3px solid var(--color-surface-3); border-top-color: var(--color-accent); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty { text-align: center; padding: var(--gap-xl); color: var(--color-text-dim); font-size: var(--font-size-lg); }

.fab-cart {
  position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%);
  display: flex; align-items: center; gap: var(--gap-sm);
  padding: 14px 28px;
  background: linear-gradient(135deg, #c8960a, #e0aa12);
  color: #0a1f12; border: none; border-radius: var(--radius-full);
  font-family: var(--font-family); font-size: var(--font-size-md); font-weight: 800;
  cursor: pointer; box-shadow: 0 8px 32px rgba(200,150,10,0.45);
  animation: fabUp 0.3s ease; z-index: 100;
}
.fab-cart:active { transform: translateX(-50%) scale(0.95); }
@keyframes fabUp { from { transform: translateX(-50%) translateY(100px); opacity: 0; } to { transform: translateX(-50%) translateY(0); opacity: 1; } }
.fab-count { background: #0a1f12; color: var(--color-accent); width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: var(--font-size-sm); font-weight: 800; }
</style>