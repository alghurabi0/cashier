<script setup lang="ts">
import { ref } from 'vue'
import { useManagement } from '../../composables/useManagement'
import { formatPrice } from '../../types'
import InventoryFormDialog from './InventoryFormDialog.vue'
import type { InventoryItem } from '../../types'

const { inventoryItems, isLoading, createInventoryItem, updateInventoryItem, deleteInventoryItem } = useManagement()

const showForm = ref(false)
const editingItem = ref<InventoryItem | null>(null)
const confirmDeleteId = ref<string | null>(null)

function openCreate() {
  editingItem.value = null
  showForm.value = true
}

function openEdit(item: InventoryItem) {
  editingItem.value = item
  showForm.value = true
}

async function onSave(data: { name_ar: string; base_unit_ar: string; stock_qty: number; low_stock_threshold: number; unit_cost: number }) {
  try {
    if (editingItem.value) {
      await updateInventoryItem(editingItem.value.id, data.name_ar, data.base_unit_ar, data.stock_qty, data.low_stock_threshold, data.unit_cost)
    } else {
      await createInventoryItem(data.name_ar, data.base_unit_ar, data.stock_qty, data.low_stock_threshold, data.unit_cost)
    }
    showForm.value = false
  } catch (err) {
    console.error('Save failed:', err)
  }
}

async function onConfirmDelete() {
  if (!confirmDeleteId.value) return
  try {
    await deleteInventoryItem(confirmDeleteId.value)
    confirmDeleteId.value = null
  } catch (err) {
    console.error('Delete failed:', err)
  }
}

function stockStatus(item: InventoryItem): 'ok' | 'low' | 'critical' {
  if (item.low_stock_threshold <= 0) return 'ok'
  if (item.stock_qty <= 0) return 'critical'
  if (item.stock_qty <= item.low_stock_threshold) return 'low'
  return 'ok'
}
</script>

<template>
  <div class="inventory-table">
    <div class="table-toolbar">
      <h2 class="section-title">المواد الخام</h2>
      <button class="btn btn-primary" @click="openCreate">+ إضافة مادة</button>
    </div>

    <div v-if="inventoryItems.length === 0 && !isLoading" class="empty-state">
      <span class="empty-icon">🧪</span>
      <span class="empty-text">لا توجد مواد خام</span>
      <button class="btn btn-primary" @click="openCreate">أضف مادة جديدة</button>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th>الاسم</th>
          <th>الوحدة</th>
          <th>المخزون</th>
          <th>حد التنبيه</th>
          <th>تكلفة الوحدة</th>
          <th>إجراءات</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in inventoryItems" :key="item.id">
          <td class="cell-name">{{ item.name_ar }}</td>
          <td class="cell-unit">{{ item.base_unit_ar }}</td>
          <td>
            <span class="stock-badge" :class="stockStatus(item)">
              {{ item.stock_qty.toLocaleString() }}
            </span>
          </td>
          <td class="text-muted">{{ item.low_stock_threshold.toLocaleString() }}</td>
          <td>{{ formatPrice(item.unit_cost) }} <small class="text-muted">د.ع</small></td>
          <td class="cell-actions">
            <button class="btn btn-ghost btn-icon" title="تعديل" @click="openEdit(item)">✏️</button>
            <button class="btn btn-ghost btn-icon" title="حذف" @click="confirmDeleteId = item.id">🗑️</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Create/Edit Dialog -->
    <InventoryFormDialog
      v-if="showForm"
      :editing="editingItem"
      @save="onSave"
      @cancel="showForm = false"
    />

    <!-- Delete Confirmation -->
    <div v-if="confirmDeleteId" class="modal-overlay" @click.self="confirmDeleteId = null">
      <div class="modal-content">
        <h3 class="modal-title">تأكيد الحذف</h3>
        <p class="text-muted">هل أنت متأكد من حذف هذه المادة؟</p>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="confirmDeleteId = null">إلغاء</button>
          <button class="btn btn-primary" style="background: var(--color-danger)" @click="onConfirmDelete">حذف</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--gap-lg);
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

.cell-unit {
  color: var(--color-text-muted);
}

.cell-actions {
  display: flex;
  gap: var(--gap-xs);
}

.stock-badge {
  display: inline-flex;
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  font-size: var(--font-size-sm);
}

.stock-badge.ok {
  background: rgba(39, 174, 96, 0.15);
  color: var(--color-success);
}

.stock-badge.low {
  background: rgba(243, 156, 18, 0.15);
  color: var(--color-warning);
}

.stock-badge.critical {
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
