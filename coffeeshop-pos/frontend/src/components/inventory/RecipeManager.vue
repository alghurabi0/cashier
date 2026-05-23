<script setup lang="ts">
import { ref, computed } from 'vue'
import { useManagement } from '../../composables/useManagement'
import RecipeDialog from './RecipeDialog.vue'
import type { MenuItem } from '../../types'

const { allMenuItems } = useManagement()

const selectedMenuItem = ref<MenuItem | null>(null)
const showRecipeDialog = ref(false)

// Group menu items by category
const menuItemsByCategory = computed(() => {
  const groups: Record<string, { categoryName: string; items: MenuItem[] }> = {}
  for (const item of allMenuItems.value) {
    const catId = item.category_id
    if (!groups[catId]) {
      groups[catId] = { categoryName: item.category_name_ar, items: [] }
    }
    groups[catId].items.push(item)
  }
  return Object.values(groups)
})

function openRecipe(item: MenuItem) {
  selectedMenuItem.value = item
  showRecipeDialog.value = true
}

function onRecipeSaved() {
  showRecipeDialog.value = false
}
</script>

<template>
  <div class="recipe-manager">
    <div class="table-toolbar">
      <h2 class="section-title">إدارة الوصفات</h2>
    </div>

    <div v-if="allMenuItems.length === 0" class="empty-state">
      <span class="empty-icon">📜</span>
      <span class="empty-text">لا توجد منتجات</span>
      <p class="text-muted text-sm">أضف منتجات من تبويب "قائمة المنتجات" أولاً، ثم ارجع هنا لربط كل منتج بمكوناته</p>
    </div>

    <div v-for="group in menuItemsByCategory" :key="group.categoryName" class="recipe-group">
      <h3 class="group-title">{{ group.categoryName }}</h3>
      <table class="data-table">
        <thead>
          <tr>
            <th>المنتج</th>
            <th>السعر</th>
            <th>التكلفة التلقائية</th>
            <th>حالة الوصفة</th>
            <th>إجراءات</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in group.items" :key="item.id">
            <td class="cell-name">{{ item.name_ar }}</td>
            <td>{{ item.price.toLocaleString() }} <small class="text-muted">د.ع</small></td>
            <td>
              <span v-if="item.cached_auto_cost > 0">
                {{ item.cached_auto_cost.toLocaleString() }} <small class="text-muted">د.ع</small>
              </span>
              <span v-else class="text-muted">—</span>
            </td>
            <td>
              <span v-if="item.cached_auto_cost > 0" class="recipe-status has-recipe">✅ وصفة موجودة</span>
              <span v-else class="recipe-status no-recipe">⚠️ لا يوجد</span>
            </td>
            <td>
              <button class="btn btn-ghost" @click="openRecipe(item)">
                {{ item.cached_auto_cost > 0 ? 'تعديل' : 'إضافة وصفة' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <RecipeDialog
      v-if="showRecipeDialog && selectedMenuItem"
      :menu-item="selectedMenuItem"
      @save="onRecipeSaved"
      @cancel="showRecipeDialog = false"
    />
  </div>
</template>

<style scoped>
.recipe-manager {
  display: flex;
  flex-direction: column;
  gap: var(--gap-lg);
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
}

.recipe-group {
  margin-bottom: var(--gap-md);
}

.group-title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semi);
  color: var(--color-accent);
  margin-bottom: var(--gap-sm);
  padding-bottom: var(--gap-xs);
  border-bottom: 1px solid var(--color-border);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: var(--gap-sm) var(--gap-md);
  text-align: right;
  border-bottom: 1px solid var(--color-border);
}

.data-table th {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  background: var(--color-surface);
}

.data-table tr:hover td {
  background: var(--color-surface-2);
}

.cell-name {
  font-weight: var(--font-weight-semi);
}

.recipe-status {
  font-size: var(--font-size-sm);
}

.recipe-status.has-recipe {
  color: var(--color-success);
}

.recipe-status.no-recipe {
  color: var(--color-warning);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-xl) 0;
  color: var(--color-text-dim);
}

.empty-icon {
  font-size: 3rem;
  opacity: 0.4;
}
</style>
