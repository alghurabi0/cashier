<script setup lang="ts">
import { ref } from 'vue'
import { useManagement } from '../../composables/useManagement'
import CategoryFormDialog from './CategoryFormDialog.vue'
import type { Category } from '../../types'

const { categories, isLoading, createCategory, updateCategory, deleteCategory } = useManagement()

const showForm = ref(false)
const editingItem = ref<Category | null>(null)
const confirmDeleteId = ref<string | null>(null)

function openCreate() {
  editingItem.value = null
  showForm.value = true
}

function openEdit(item: Category) {
  editingItem.value = item
  showForm.value = true
}

async function onSave(data: { name_ar: string; sort_order: number }) {
  try {
    if (editingItem.value) {
      await updateCategory(editingItem.value.id, data.name_ar, data.sort_order)
    } else {
      await createCategory(data.name_ar, data.sort_order)
    }
    showForm.value = false
  } catch (err) {
    console.error('Save failed:', err)
  }
}

async function onConfirmDelete() {
  if (!confirmDeleteId.value) return
  try {
    await deleteCategory(confirmDeleteId.value)
    confirmDeleteId.value = null
  } catch (err) {
    console.error('Delete failed:', err)
  }
}
</script>

<template>
  <div class="category-manager">
    <div class="table-toolbar">
      <h2 class="section-title">إدارة الفئات</h2>
      <button class="btn btn-primary" @click="openCreate">+ إضافة فئة</button>
    </div>

    <div v-if="categories.length === 0 && !isLoading" class="empty-state">
      <span class="empty-icon">📁</span>
      <span class="empty-text">لا توجد فئات</span>
      <button class="btn btn-primary" @click="openCreate">أضف فئة جديدة</button>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th>الاسم</th>
          <th>ترتيب العرض</th>
          <th>الحالة</th>
          <th>إجراءات</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="cat in categories" :key="cat.id">
          <td class="cell-name">{{ cat.name_ar }}</td>
          <td>{{ cat.sort_order }}</td>
          <td>
            <span class="status-badge" :class="cat.is_active ? 'active' : 'inactive'">
              {{ cat.is_active ? 'فعال' : 'معطل' }}
            </span>
          </td>
          <td class="cell-actions">
            <button class="btn btn-ghost btn-icon" title="تعديل" @click="openEdit(cat)">✏️</button>
            <button class="btn btn-ghost btn-icon" title="حذف" @click="confirmDeleteId = cat.id">🗑️</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Create/Edit Dialog -->
    <CategoryFormDialog
      v-if="showForm"
      :editing="editingItem"
      @save="onSave"
      @cancel="showForm = false"
    />

    <!-- Delete Confirmation -->
    <div v-if="confirmDeleteId" class="modal-overlay" @click.self="confirmDeleteId = null">
      <div class="modal-content">
        <h3 class="modal-title">تأكيد الحذف</h3>
        <p class="text-muted">هل أنت متأكد من حذف هذه الفئة؟ سيتم حذف جميع المنتجات المرتبطة بها.</p>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="confirmDeleteId = null">إلغاء</button>
          <button class="btn btn-primary" style="background: var(--color-danger)" @click="onConfirmDelete">حذف</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.category-manager {
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

.status-badge {
  display: inline-flex;
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semi);
}

.status-badge.active {
  background: rgba(39, 174, 96, 0.15);
  color: var(--color-success);
}

.status-badge.inactive {
  background: rgba(231, 76, 60, 0.15);
  color: var(--color-danger);
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
