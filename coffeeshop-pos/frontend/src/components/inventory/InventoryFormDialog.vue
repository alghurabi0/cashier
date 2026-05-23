<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { InventoryItem } from '../../types'

const props = defineProps<{
  editing: InventoryItem | null
}>()

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
  emit('save', {
    name_ar: nameAr.value.trim(),
    base_unit_ar: baseUnitAr.value.trim(),
    stock_qty: stockQty.value,
    low_stock_threshold: lowThreshold.value,
    unit_cost: unitCost.value,
  })
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal-content form-dialog">
      <h2 class="modal-title">{{ editing ? 'تعديل مادة' : 'إضافة مادة جديدة' }}</h2>

      <form @submit.prevent="onSubmit" class="form">
        <div class="form-group">
          <label class="form-label">الاسم</label>
          <input v-model="nameAr" type="text" class="form-input" placeholder="مثال: بن إسبريسو" required />
        </div>

        <div class="form-group">
          <label class="form-label">وحدة القياس</label>
          <input v-model="baseUnitAr" type="text" class="form-input" placeholder="مثال: غرام، مل" required />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">الكمية الأولية</label>
            <input v-model.number="stockQty" type="number" class="form-input" min="0" />
          </div>
          <div class="form-group">
            <label class="form-label">حد التنبيه</label>
            <input v-model.number="lowThreshold" type="number" class="form-input" min="0" />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">تكلفة الوحدة (د.ع)</label>
          <input v-model.number="unitCost" type="number" class="form-input" min="0" />
        </div>

        <div class="form-actions">
          <button type="button" class="btn btn-ghost" @click="emit('cancel')">إلغاء</button>
          <button type="submit" class="btn btn-primary">{{ editing ? 'حفظ التعديلات' : 'إضافة' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.form-dialog {
  min-width: 420px;
}

.form {
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
  margin-top: var(--gap-lg);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
}

.form-row {
  display: flex;
  gap: var(--gap-md);
}

.form-row .form-group {
  flex: 1;
}

.form-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  color: var(--color-text-muted);
}

.form-input {
  padding: var(--gap-sm) var(--gap-md);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  transition: border-color var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.form-actions {
  display: flex;
  gap: var(--gap-md);
  justify-content: flex-end;
  margin-top: var(--gap-sm);
}
</style>
