<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useManagement } from '../composables/useManagement'
import CategoryManager from '../components/inventory/CategoryManager.vue'
import MenuItemManager from '../components/inventory/MenuItemManager.vue'
import InventoryTable from '../components/inventory/InventoryTable.vue'
import RecipeManager from '../components/inventory/RecipeManager.vue'
import StockAdjustment from '../components/inventory/StockAdjustment.vue'

const activeTab = ref<'categories' | 'menu-items' | 'materials' | 'recipes' | 'stock'>('categories')

const { initBindings, loadCategories, loadInventoryItems, loadMenuItems } = useManagement()

const tabs = [
  { id: 'categories' as const, label: 'الفئات', icon: '📁' },
  { id: 'menu-items' as const, label: 'قائمة المنتجات', icon: '🍕' },
  { id: 'materials' as const, label: 'المواد الخام', icon: '🧪' },
  { id: 'recipes' as const, label: 'الوصفات', icon: '📜' },
  { id: 'stock' as const, label: 'حركة المخزون', icon: '📊' },
]

onMounted(async () => {
  await initBindings()
  await Promise.all([
    loadCategories(),
    loadInventoryItems(),
    loadMenuItems(),
  ])
})
</script>

<template>
  <div class="inventory-view">
    <header class="view-header">
      <h1 class="view-title">📦 إدارة القائمة والمخزون</h1>
    </header>

    <div class="view-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="view-tab"
        :class="{ active: activeTab === tab.id }"
        @click="activeTab = tab.id"
      >
        <span class="tab-icon">{{ tab.icon }}</span>
        {{ tab.label }}
      </button>
    </div>

    <div class="view-content">
      <CategoryManager v-if="activeTab === 'categories'" />
      <MenuItemManager v-else-if="activeTab === 'menu-items'" />
      <InventoryTable v-else-if="activeTab === 'materials'" />
      <RecipeManager v-else-if="activeTab === 'recipes'" />
      <StockAdjustment v-else-if="activeTab === 'stock'" />
    </div>
  </div>
</template>

<style scoped>
.inventory-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.view-header {
  padding: var(--gap-lg) var(--gap-xl);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.view-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
}

.view-tabs {
  display: flex;
  gap: var(--gap-sm);
  padding: var(--gap-md) var(--gap-xl);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
  overflow-x: auto;
}

.view-tab {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-sm) var(--gap-lg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-full);
  background: transparent;
  color: var(--color-text-muted);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semi);
  cursor: pointer;
  transition: all var(--transition-fast);
  user-select: none;
  white-space: nowrap;
}

.view-tab:hover {
  background: var(--color-surface-2);
  color: var(--color-text);
}

.view-tab.active {
  background: var(--color-accent);
  color: white;
  border-color: var(--color-accent);
}

.tab-icon {
  font-size: var(--font-size-sm);
}

.view-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-lg) var(--gap-xl);
}
</style>
