<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { InventoryItem } from '../../types'

const props = defineProps<{ editing: InventoryItem | null }>()
const emit = defineEmits<{
  save: [data: { name_ar: string; base_unit_ar: string; stock_qty: number; low_stock_threshold: number; unit_cost: number }]
  cancel: []
}>()

const nameAr = ref('')
const baseUnitAr = ref('')
const stockQty = ref(0)
const lowThreshold = ref(0)
const unitCost = ref(0)

onMounted(() => {
  if (props.editing) {
    nameAr.value = props.editing.name_ar
    baseUnitAr.value = props.editing.base_unit_ar
    stockQty.value = props.editing.stock_qty
    lowThreshold.value = props.editing.low_stock_threshold
    unitCost.value = props.editing.unit_cost
  }
})

function onSubmit() {
  if (!nameAr.value.trim() || !baseUnitAr.value.trim()) return
  emit('save', { name_ar: nameAr.value.trim(), base_unit_ar: baseUnitAr.value.trim(), stock_qty: stockQty.value, low_stock_threshold: lowThreshold.value, unit_cost: unitCost.value })
}
</script>

<template>
  <div class="overlay" @click.self="emit('cancel')">
    <div class="dialog">
      <h2 class="dialog-title">{{ editing ? '✏️ تعديل مادة' : '➕ إضافة مادة جديدة' }}</h2>

      <form @submit.prevent="onSubmit" class="form">
        <div class="field">
          <label>اسم المادة</label>
          <input v-model="nameAr" type="text" placeholder="مثال: بن إسبريسو" required />
        </div>
        <div class="field">
          <label>وحدة القياس</label>
          <input v-model="baseUnitAr" type="text" placeholder="مثال: غرام، مل، كيلو" required />
        </div>
        <div class="field-row">
          <div class="field">
            <label>الكمية الأولية</label>
            <input v-model.number="stockQty" type="number" min="0" />
          </div>
          <div class="field">
            <label>حد التنبيه</label>
            <input v-model.number="lowThreshold" type="number" min="0" />
          </div>
        </div>
        <div class="field">
          <label>تكلفة الوحدة (د.ع)</label>
          <input v-model.number="unitCost" type="number" min="0" />
        </div>
        <div class="form-actions">
          <button type="button" class="cancel-btn" @click="emit('cancel')">إلغاء</button>
          <button type="submit" class="save-btn">{{ editing ? 'حفظ التعديلات' : 'إضافة' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.7);
  backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000;
}

.dialog {
  background: #161616;
  border: 1px solid rgba(201,168,76,0.2);
  border-radius: 18px;
  padding: 24px;
  min-width: 400px;
  box-shadow: 0 20px 50px rgba(0,0,0,0.5);
  animation: slideUp 0.2s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(14px); }
  to   { opacity: 1; transform: translateY(0); }
}

.dialog-title { font-size: 1rem; font-weight: 800; color: #f0e6d3; margin-bottom: 20px; }

.form { display: flex; flex-direction: column; gap: 14px; }

.field { display: flex; flex-direction: column; gap: 5px; }
.field label { font-size: 0.75rem; font-weight: 700; color: #666; }
.field input {
  padding: 10px 12px;
  background: #222;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 10px;
  color: #f0e6d3;
  font-family: inherit;
  font-size: 0.9rem;
  transition: border-color 0.15s;
}
.field input:focus { outline: none; border-color: #c9a84c; }

.field-row { display: flex; gap: 12px; }
.field-row .field { flex: 1; }

.form-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 4px; }

.cancel-btn {
  padding: 10px 18px; border: 1px solid rgba(255,255,255,0.08);
  border-radius: 10px; background: transparent; color: #666;
  font-family: inherit; font-size: 0.88rem; font-weight: 700; cursor: pointer;
}

.save-btn {
  padding: 10px 24px; border: none;
  border-radius: 10px; background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d; font-family: inherit; font-size: 0.88rem; font-weight: 800; cursor: pointer;
}

.save-btn:hover { filter: brightness(1.08); }
</style>