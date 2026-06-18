<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { CartItem } from '../types'
import { formatPrice } from '../types'

interface Table {
  id: string
  number: string
  token: string
  is_active: boolean
}

defineProps<{
  items: CartItem[]
  total: number
}>()

const emit = defineEmits<{
  confirm: [tableNumber: string]
  cancel: []
}>()

const tables = ref<Table[]>([])
const selectedTable = ref('')
let ManagementService: any = null

onMounted(async () => {
  try {
    ManagementService = await import('../../bindings/coffeeshop-pos/internal/service/managementservice')
    const result = await ManagementService.ListTables()
    tables.value = result || []
  } catch {
    // Tables not available
  }
})

function selectTable(number: string) {
  selectedTable.value = selectedTable.value === number ? '' : number
}
</script>

<template>
  <div class="overlay" @click.self="emit('cancel')">
    <div class="dialog">

      <!-- Header -->
      <div class="dialog-header">
        <span class="dialog-icon">🧾</span>
        <h2 class="dialog-title">تأكيد الطلب</h2>
      </div>

      <!-- Items -->
      <div class="items-list">
        <div v-for="item in items" :key="item.menu_item_id" class="item-row">
          <span class="item-qty">×{{ item.quantity }}</span>
          <span class="item-name">{{ item.name_ar }}</span>
          <span class="item-price">{{ formatPrice(item.price * item.quantity) }}</span>
        </div>
      </div>

      <!-- Total -->
      <div class="total-row">
        <span class="total-label">المجموع الكلي</span>
        <span class="total-val">{{ formatPrice(total) }} <small>د.ع</small></span>
      </div>

      <!-- Table Selection -->
      <div v-if="tables.length > 0" class="section">
        <div class="section-label">🪑 الطاولة <span class="optional">(اختياري)</span></div>
        <div class="table-grid">
          <button
            v-for="table in tables"
            :key="table.id"
            class="table-btn"
            :class="{ active: selectedTable === table.number }"
            @click="selectTable(table.number)"
          >
            {{ table.number }}
          </button>
        </div>
      </div>

      <!-- Payment -->
      <div class="section">
        <div class="section-label">💳 طريقة الدفع</div>
        <div class="payment-row">
          <button class="pay-btn active">
            <span>💵</span>
            <span>نقدي</span>
          </button>
        </div>
      </div>

      <!-- Actions -->
      <div class="actions">
        <button class="cancel-btn" @click="emit('cancel')">إلغاء</button>
        <button class="confirm-btn" @click="emit('confirm', selectedTable)">
          ✓ تأكيد الطلب
        </button>
      </div>

    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.7);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: #161616;
  border: 1px solid rgba(201,168,76,0.2);
  border-radius: 20px;
  padding: 28px;
  min-width: 420px;
  max-width: 500px;
  width: 90vw;
  box-shadow: 0 20px 60px rgba(0,0,0,0.6), 0 0 40px rgba(201,168,76,0.06);
  animation: slideUp 0.2s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* Header */
.dialog-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.dialog-icon { font-size: 1.4rem; }

.dialog-title {
  font-size: 1.1rem;
  font-weight: 800;
  color: #f0e6d3;
}

/* Items */
.items-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
  max-height: 200px;
  overflow-y: auto;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
  font-size: 0.88rem;
}

.item-qty {
  color: #c9a84c;
  font-weight: 800;
  min-width: 28px;
  font-size: 0.82rem;
}

.item-name {
  flex: 1;
  color: #d4c8b0;
}

.item-price {
  color: #888;
  font-variant-numeric: tabular-nums;
  font-size: 0.82rem;
}

/* Total */
.total-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding: 14px 0;
  border-top: 1px solid rgba(201,168,76,0.15);
  border-bottom: 1px solid rgba(201,168,76,0.15);
  margin-bottom: 18px;
}

.total-label {
  font-size: 0.9rem;
  color: #888;
  font-weight: 600;
}

.total-val {
  font-size: 1.6rem;
  font-weight: 800;
  color: #c9a84c;
  font-variant-numeric: tabular-nums;
}

.total-val small {
  font-size: 0.7rem;
  opacity: 0.7;
  margin-right: 3px;
}

/* Sections */
.section {
  margin-bottom: 18px;
}

.section-label {
  font-size: 0.78rem;
  font-weight: 700;
  color: #666;
  margin-bottom: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.optional {
  font-weight: 400;
  color: #444;
}

/* Table */
.table-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.table-btn {
  padding: 7px 16px;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 999px;
  background: #222;
  color: #888;
  font-family: inherit;
  font-size: 0.82rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}

.table-btn:hover {
  border-color: rgba(201,168,76,0.35);
  color: #c9a84c;
}

.table-btn.active {
  background: rgba(201,168,76,0.15);
  border-color: #c9a84c;
  color: #c9a84c;
}

/* Payment */
.payment-row { display: flex; gap: 8px; }

.pay-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: 1px solid rgba(201,168,76,0.3);
  border-radius: 10px;
  background: rgba(201,168,76,0.08);
  color: #c9a84c;
  font-family: inherit;
  font-size: 0.88rem;
  font-weight: 700;
  cursor: pointer;
}

/* Actions */
.actions {
  display: flex;
  gap: 10px;
  margin-top: 4px;
}

.cancel-btn {
  padding: 12px 20px;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  background: transparent;
  color: #666;
  font-family: inherit;
  font-size: 0.88rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}

.cancel-btn:hover {
  background: #222;
  color: #999;
}

.confirm-btn {
  flex: 1;
  padding: 14px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d;
  font-family: inherit;
  font-size: 1rem;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.18s ease;
  box-shadow: 0 4px 20px rgba(201,168,76,0.3);
}

.confirm-btn:hover {
  filter: brightness(1.08);
  box-shadow: 0 6px 28px rgba(201,168,76,0.45);
}

.confirm-btn:active {
  transform: scale(0.98);
}
</style>