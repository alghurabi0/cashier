<script setup lang="ts">
import { ref, computed } from 'vue'
import { useManagement } from '../../composables/useManagement'
import { formatPrice } from '../../types'
import MenuItemFormDialog from './MenuItemFormDialog.vue'
import type { MenuItem } from '../../types'

const { allMenuItems, isLoading, createMenuItem, updateMenuItem, deleteMenuItem } = useManagement()

const showForm = ref(false)
const editingItem = ref<MenuItem | null>(null)
const confirmDeleteId = ref<string | null>(null)

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

function openCreate() {
  editingItem.value = null
  showForm.value = true
}

function openEdit(item: MenuItem) {
  editingItem.value = item
  showForm.value = true
}

async function onSave(data: { category_id: string; name_ar: string; price: number; cost_calc_method: string; manual_cost_price: number; image_path: string }) {
  try {
    if (editingItem.value) {
      await updateMenuItem(editingItem.value.id, data.category_id, data.name_ar, data.price, data.cost_calc_method, data.manual_cost_price, data.image_path || '')
    } else {
      await createMenuItem(data.category_id, data.name_ar, data.price, data.cost_calc_method, data.manual_cost_price, data.image_path || '')
    }
    showForm.value = false
  } catch (err) {
    console.error('Save failed:', err)
  }
}

async function onConfirmDelete() {
  if (!confirmDeleteId.value) return
  try {
    await deleteMenuItem(confirmDeleteId.value)
    confirmDeleteId.value = null
  } catch (err) {
    console.error('Delete failed:', err)
  }
}

function costMethodLabel(method: string): string {
  return method === 'manual' ? 'يدوي' : 'تلقائي'
}
</script>

<template>
  <div class="menu-item-manager">
    <div class="table-toolbar">
      <h2 class="section-title">قائمة المنتجات</h2>
      <button class="btn btn-primary" @click="openCreate">+ إضافة منتج</button>
    </div>

    <div v-if="allMenuItems.length === 0 && !isLoading" class="empty-state">
      <span class="empty-icon">🍕</span>
      <span class="empty-text">لا توجد منتجات</span>
      <p class="text-muted text-sm">أضف فئة أولاً من تبويب "الفئات"، ثم أضف منتجاتك هنا</p>
      <button class="btn btn-primary" @click="openCreate">أضف منتج جديد</button>
    </div>

    <div v-for="group in menuItemsByCategory" :key="group.categoryName" class="item-group">
      <h3 class="group-title">{{ group.categoryName }}</h3>
      <table class="data-table">
        <thead>
          <tr>
            <th>المنتج</th>
            <th>السعر</th>
            <th>طريقة التكلفة</th>
            <th>التكلفة</th>
            <th>إجراءات</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in group.items" :key="item.id">
            <td class="cell-name">{{ item.name_ar }}</td>
            <td>{{ formatPrice(item.price) }} <small class="text-muted">د.ع</small></td>
            <td>
              <span class="cost-method-badge" :class="item.cost_calc_method">
                {{ costMethodLabel(item.cost_calc_method) }}
              </span>
            </td>
            <td>
              <span v-if="item.cost_calc_method === 'manual' && item.manual_cost_price > 0">
                {{ formatPrice(item.manual_cost_price) }} <small class="text-muted">د.ع</small>
              </span>
              <span v-else-if="item.cached_auto_cost > 0">
                {{ formatPrice(item.cached_auto_cost) }} <small class="text-muted">د.ع</small>
              </span>
              <span v-else class="text-muted">—</span>
            </td>
            <td class="cell-actions">
              <button class="btn btn-ghost btn-icon" title="تعديل" @click="openEdit(item)">✏️</button>
              <button class="btn btn-ghost btn-icon" title="حذف" @click="confirmDeleteId = item.id">🗑️</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Dialog -->
    <MenuItemFormDialog
      v-if="showForm"
      :editing="editingItem"
      @save="onSave"
      @cancel="showForm = false"
    />

    <!-- Delete Confirmation -->
    <div v-if="confirmDeleteId" class="modal-overlay" @click.self="confirmDeleteId = null">
      <div class="modal-content">
        <h3 class="modal-title">تأكيد الحذف</h3>
        <p class="text-muted">هل أنت متأكد من حذف هذا المنتج؟</p>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="confirmDeleteId = null">إلغاء</button>
          <button class="btn btn-primary" style="background: var(--color-danger)" @click="onConfirmDelete">حذف</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.menu-item-manager {
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

.item-group {
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
  padding: var(--gap-md);
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

.cell-actions {
  display: flex;
  gap: var(--gap-xs);
}

.cost-method-badge {
  display: inline-flex;
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semi);
}

.cost-method-badge.auto {
  background: rgba(52, 152, 219, 0.15);
  color: #3498db;
}

.cost-method-badge.manual {
  background: rgba(243, 156, 18, 0.15);
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

.empty-text {
  font-size: var(--font-size-lg);
}

.modal-actions {
  display: flex;
  gap: var(--gap-md);
  justify-content: flex-end;
  margin-top: var(--gap-lg);
}
</style>
