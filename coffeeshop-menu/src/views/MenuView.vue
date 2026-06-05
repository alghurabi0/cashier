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

onMounted(() => {
  loadAll()
})

function onAddItem(menuItemId: string, nameAr: string, price: number) {
  addItem(menuItemId, nameAr, price)
}
</script>

<template>
  <div class="menu-view">
    <header class="menu-header">
      <div class="header-logo">
        <div class="logo-circle">
          <span class="logo-text">NJ</span>
        </div>
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

    <button
      v-if="itemCount > 0"
      class="fab-cart"
      @click="emit('open-cart')"
    >
      <span>🛒</span>
      <span class="fab-count">{{ itemCount }}</span>
      <span>عرض السلة</span>
    </button>
  </div>
</template>

<style scoped>
.menu-view {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
}

.menu-header {
  padding: var(--gap-lg);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 50;
  box-shadow: 0 2px 12px rgba(139, 94, 60, 0.08);
}

.header-logo {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
}

.logo-circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--color-accent);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-text {
  font-size: 1.1rem;
  font-weight: 900;
  color: #ffffff;
  letter-spacing: 1px;
}

.header-titles {
  display: flex;
  flex-direction: column;
}

.shop-name {
  font-size: 1.3rem;
  font-weight: 900;
  color: var(--color-accent);
  line-height: 1.1;
}

.shop-tagline {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin-top: 2px;
}

.table-badge {
  display: flex;
  align-items: center;
  gap: var(--gap-xs);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  padding: var(--gap-xs) var(--gap-md);
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: 700;
  color: var(--color-accent);
}

.menu-content {
  flex: 1;
  padding: var(--gap-md) var(--gap-md) 100px;
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--gap-md);
}

@media (min-width: 600px) {
  .menu-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 900px) {
  .menu-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--gap-md);
  padding: var(--gap-xl) 0;
  color: var(--color-text-muted);
}

.loading-ring {
  width: 40px;
  height: 40px;
  border: 3px solid var(--color-surface-3);
  border-top-color: var(--color-accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty {
  text-align: center;
  padding: var(--gap-xl);
  color: var(--color-text-dim);
  font-size: var(--font-size-lg);
}

.fab-cart {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: 14px 28px;
  background: var(--color-accent);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-full);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 8px 24px rgba(139, 94, 60, 0.4);
  animation: fabSlideUp 0.3s ease;
  z-index: 100;
}

.fab-cart:active {
  transform: translateX(-50%) scale(0.95);
}

@keyframes fabSlideUp {
  from { transform: translateX(-50%) translateY(100px); opacity: 0; }
  to   { transform: translateX(-50%) translateY(0); opacity: 1; }
}

.fab-count {
  background: #ffffff;
  color: var(--color-accent);
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-sm);
  font-weight: 800;
}
</style>