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

function openCreate() { editingItem.value = null; showForm.value = true }
function openEdit(item: InventoryItem) { editingItem.value = item; showForm.value = true }

async function onSave(data: any) {
  try {
    if (editingItem.value) {
      await updateInventoryItem(editingItem.value.id, data.name_ar, data.base_unit_ar, data.stock_qty, data.low_stock_threshold, data.unit_cost)
    } else {
      await createInventoryItem(data.name_ar, data.base_unit_ar, data.stock_qty, data.low_stock_threshold, data.unit_cost)
    }
    showForm.value = false
  } catch (err) { console.error(err) }
}

async function onConfirmDelete() {
  if (!confirmDeleteId.value) return
  try { await deleteInventoryItem(confirmDeleteId.value); confirmDeleteId.value = null }
  catch (err) { console.error(err) }
}

function stockStatus(item: InventoryItem): 'ok' | 'low' | 'critical' {
  if (item.low_stock_threshold <= 0) return 'ok'
  if (item.stock_qty <= 0) return 'critical'
  if (item.stock_qty <= item.low_stock_threshold) return 'low'
  return 'ok'
}
</script>

<template>
  <div class="inv-table">
    <div class="toolbar">
      <h2 class="section-title">🧪 المواد الخام</h2>
      <button class="add-btn" @click="openCreate">+ إضافة مادة</button>
    </div>

    <div v-if="inventoryItems.length === 0 && !isLoading" class="empty">
      <span>🧪</span>
      <p>لا توجد مواد خام</p>
      <button class="add-btn" @click="openCreate">أضف مادة جديدة</button>
    </div>

    <table v-else class="table">
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
          <td class="name-cell">{{ item.name_ar }}</td>
          <td class="muted">{{ item.base_unit_ar }}</td>
          <td>
            <span class="stock-chip" :class="stockStatus(item)">
              {{ item.stock_qty.toLocaleString() }}
              <span v-if="stockStatus(item) === 'low'"> ⚠️</span>
              <span v-if="stockStatus(item) === 'critical'"> 🔴</span>
            </span>
          </td>
          <td class="muted">{{ item.low_stock_threshold.toLocaleString() }}</td>
          <td class="gold">{{ formatPrice(item.unit_cost) }} <small>د.ع</small></td>
          <td class="actions-cell">
            <button class="icon-btn" @click="openEdit(item)">✏️</button>
            <button class="icon-btn danger" @click="confirmDeleteId = item.id">🗑️</button>
          </td>
        </tr>
      </tbody>
    </table>

    <InventoryFormDialog v-if="showForm" :editing="editingItem" @save="onSave" @cancel="showForm = false" />

    <div v-if="confirmDeleteId" class="overlay" @click.self="confirmDeleteId = null">
      <div class="confirm-dialog">
        <h3>تأكيد الحذف</h3>
        <p>هل أنت متأكد من حذف هذه المادة؟</p>
        <div class="confirm-actions">
          <button class="cancel-btn" @click="confirmDeleteId = null">إلغاء</button>
          <button class="delete-btn" @click="onConfirmDelete">حذف</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.inv-table { display: flex; flex-direction: column; gap: 16px; }

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title { font-size: 1rem; font-weight: 800; color: #f0e6d3; }

.add-btn {
  padding: 8px 18px;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d;
  border: none;
  border-radius: 10px;
  font-family: inherit;
  font-size: 0.85rem;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.18s ease;
}

.add-btn:hover { filter: brightness(1.08); }

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px 0;
  color: #444;
  font-size: 0.9rem;
}

.empty span { font-size: 2.5rem; opacity: 0.3; }

.table {
  width: 100%;
  border-collapse: collapse;
  background: #1a1a1a;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(255,255,255,0.05);
}

.table th {
  padding: 12px 16px;
  text-align: right;
  font-size: 0.75rem;
  font-weight: 700;
  color: #555;
  background: #161616;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.table td {
  padding: 12px 16px;
  text-align: right;
  border-bottom: 1px solid rgba(255,255,255,0.04);
  font-size: 0.88rem;
}

.table tr:last-child td { border-bottom: none; }
.table tr:hover td { background: rgba(201,168,76,0.04); }

.name-cell { font-weight: 700; color: #e8dcc8; }
.muted { color: #666; }
.gold { color: #c9a84c; font-weight: 700; }
.gold small { font-size: 0.65rem; opacity: 0.7; }

.stock-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.stock-chip.ok { background: rgba(39,174,96,0.12); color: #27ae60; }
.stock-chip.low { background: rgba(243,156,18,0.12); color: #f39c12; }
.stock-chip.critical { background: rgba(231,76,60,0.12); color: #e74c3c; }

.actions-cell { display: flex; gap: 4px; }

.icon-btn {
  width: 32px; height: 32px;
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 8px;
  background: #222;
  cursor: pointer;
  font-size: 0.85rem;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s ease;
}

.icon-btn:hover { background: #2a2a2a; border-color: rgba(201,168,76,0.3); }
.icon-btn.danger:hover { background: rgba(231,76,60,0.12); border-color: rgba(231,76,60,0.3); }

/* Overlay */
.overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.7);
  backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000;
}

.confirm-dialog {
  background: #161616;
  border: 1px solid rgba(231,76,60,0.3);
  border-radius: 16px;
  padding: 24px;
  min-width: 320px;
  display: flex; flex-direction: column; gap: 12px;
}

.confirm-dialog h3 { font-size: 1rem; font-weight: 800; color: #f0e6d3; }
.confirm-dialog p { font-size: 0.85rem; color: #666; }

.confirm-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 8px; }

.cancel-btn {
  padding: 8px 16px; border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px; background: transparent; color: #666;
  font-family: inherit; font-size: 0.85rem; font-weight: 700; cursor: pointer;
}

.delete-btn {
  padding: 8px 16px; border: none;
  border-radius: 8px; background: #e74c3c; color: white;
  font-family: inherit; font-size: 0.85rem; font-weight: 700; cursor: pointer;
}
</style>