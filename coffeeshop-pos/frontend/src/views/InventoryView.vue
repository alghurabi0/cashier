<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useManagement } from '../composables/useManagement'
import CategoryManager from '../components/inventory/CategoryManager.vue'
import MenuItemManager from '../components/inventory/MenuItemManager.vue'
import InventoryTable from '../components/inventory/InventoryTable.vue'
import RecipeManager from '../components/inventory/RecipeManager.vue'
import StockAdjustment from '../components/inventory/StockAdjustment.vue'

const activeTab = ref<'categories' | 'menu-items' | 'materials' | 'recipes' | 'stock'>('materials')

const { initBindings, loadCategories, loadInventoryItems, loadMenuItems } = useManagement()

const tabs = [
  { id: 'materials' as const,   label: 'المواد الخام',      icon: '🧪' },
  { id: 'recipes' as const,     label: 'الوصفات',           icon: '📜' },
  { id: 'stock' as const,       label: 'حركة المخزون',      icon: '📊' },
  { id: 'menu-items' as const,  label: 'قائمة المنتجات',    icon: '🍕' },
  { id: 'categories' as const,  label: 'الفئات',            icon: '📁' },
]

onMounted(async () => {
  await initBindings()
  await Promise.all([loadCategories(), loadInventoryItems(), loadMenuItems()])
})
</script>

<template>
  <div class="inv-view">
    <header class="inv-header">
      <div class="header-left">
        <span class="header-icon">📦</span>
        <div>
          <h1 class="header-title">إدارة المخزون</h1>
          <p class="header-sub">المواد الخام • الوصفات • حركة المخزون</p>
        </div>
      </div>
    </header>

    <div class="inv-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="inv-tab"
        :class="{ active: activeTab === tab.id }"
        @click="activeTab = tab.id"
      >
        <span>{{ tab.icon }}</span>
        {{ tab.label }}
      </button>
    </div>

    <div class="inv-content">
      <CategoryManager  v-if="activeTab === 'categories'" />
      <MenuItemManager  v-else-if="activeTab === 'menu-items'" />
      <InventoryTable   v-else-if="activeTab === 'materials'" />
      <RecipeManager    v-else-if="activeTab === 'recipes'" />
      <StockAdjustment  v-else-if="activeTab === 'stock'" />
    </div>
  </div>
</template>

<style scoped>
.inv-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background: var(--color-bg);
}

.inv-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: var(--color-cart-bg);
  border-bottom: 1px solid var(--color-border-light);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-icon { font-size: 1.6rem; }

.header-title {
  font-size: 1.1rem;
  font-weight: 800;
  color: var(--color-text);
}

.header-sub {
  font-size: 0.7rem;
  color: var(--color-text-dim);
  margin-top: 2px;
}

.inv-tabs {
  display: flex;
  gap: 6px;
  padding: 12px 24px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.inv-tabs::-webkit-scrollbar { display: none; }

.inv-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border: 1px solid var(--color-border-light);
  border-radius: 999px;
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-family: inherit;
  font-size: 0.82rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.18s ease;
}

.inv-tab:hover {
  border-color: var(--color-accent-glow);
  color: var(--color-accent);
}

.inv-tab.active {
  background: linear-gradient(135deg, var(--color-accent), var(--color-accent-hover));
  border-color: var(--color-accent);
  color: #0d0d0d;
  box-shadow: var(--shadow-glow);
}

.inv-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
</style>