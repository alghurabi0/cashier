<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMenu } from '../composables/useMenu'
import { useCart } from '../composables/useCart'
import CategoryBar from '../components/CategoryBar.vue'
import MenuItemCard from '../components/MenuItemCard.vue'
import { formatPrice } from '../types'

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
      <div class="header-top">
        <h1 class="shop-name">☕ المقهى</h1>
        <span class="table-info">🪑 طاولة {{ tableNumber }}</span>
      </div>
      <p class="shop-tagline">اختر ما يحلو لك</p>
    </header>

    <div class="menu-content">
      <CategoryBar
        :categories="categories"
        :selected-id="selectedCategoryId"
        @select="selectedCategoryId = $event"
      />

      <div v-if="isLoading" class="loading">
        <span class="loading-spinner">⏳</span>
        جاري التحميل...
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

    <!-- Floating Cart Button -->
    <button
      v-if="itemCount > 0"
      class="fab-cart"
      @click="emit('open-cart')"
    >
      <span class="fab-icon">🛒</span>
      <span class="fab-count">{{ itemCount }}</span>
      <span class="fab-label">عرض السلة</span>
    </button>
  </div>
</template>

<style scoped>
.menu-view {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
}

.menu-header {
  padding: var(--gap-xl) var(--gap-lg) var(--gap-lg);
  background: linear-gradient(135deg, var(--color-surface-2), var(--color-surface));
  border-bottom: 1px solid var(--color-border);
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.shop-name {
  font-size: var(--font-size-2xl);
  font-weight: 800;
}

.table-info {
  background: var(--color-surface-3);
  padding: var(--gap-xs) var(--gap-md);
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-accent);
}

.shop-tagline {
  color: var(--color-text-muted);
  font-size: var(--font-size-md);
  margin-top: var(--gap-xs);
}

.menu-content {
  flex: 1;
  padding: var(--gap-lg);
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.menu-grid {
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

@media (min-width: 600px) {
  .menu-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
  }
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--gap-sm);
  padding: var(--gap-xl);
  color: var(--color-text-muted);
}

.loading-spinner {
  font-size: 1.5rem;
  animation: spin 1.5s linear infinite;
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

/* Floating Action Button */
.fab-cart {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-md) var(--gap-xl);
  background: var(--color-accent);
  color: var(--color-bg);
  border: none;
  border-radius: var(--radius-full);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 8px 32px rgba(212, 165, 116, 0.35);
  animation: fabSlideUp 0.3s ease;
  z-index: 100;
}

.fab-cart:active {
  transform: translateX(-50%) scale(0.95);
}

@keyframes fabSlideUp {
  from {
    transform: translateX(-50%) translateY(100px);
    opacity: 0;
  }
  to {
    transform: translateX(-50%) translateY(0);
    opacity: 1;
  }
}

.fab-icon {
  font-size: 1.2rem;
}

.fab-count {
  background: var(--color-bg);
  color: var(--color-accent);
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-sm);
}
</style>
