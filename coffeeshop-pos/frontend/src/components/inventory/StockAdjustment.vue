<script setup lang="ts">
import { ref } from 'vue'
import { useManagement } from '../../composables/useManagement'

const { inventoryItems, adjustStock, isLoading } = useManagement()

const selectedItemId = ref('')
const adjustType = ref<'delivery' | 'waste' | 'correction'>('delivery')
const quantity = ref(0)
const reason = ref('')
const successMessage = ref('')

const adjustTypes = [
  { id: 'delivery' as const, label: 'وارد (توريد)', icon: '📥' },
  { id: 'waste' as const, label: 'هالك', icon: '🗑️' },
  { id: 'correction' as const, label: 'تعديل', icon: '🔧' },
]

async function onSubmit() {
  if (!selectedItemId.value || quantity.value === 0) return

  let delta = quantity.value
  if (adjustType.value === 'waste') {
    delta = -Math.abs(delta)
  } else if (adjustType.value === 'delivery') {
    delta = Math.abs(delta)
  }

  const adjustReason = `${adjustTypes.find(t => t.id === adjustType.value)?.label}: ${reason.value || '—'}`

  try {
    await adjustStock(selectedItemId.value, delta, adjustReason)
    successMessage.value = 'تم تسجيل الحركة بنجاح ✓'
    quantity.value = 0
    reason.value = ''
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) {
    console.error('Stock adjustment failed:', err)
  }
}
</script>

<template>
  <div class="stock-adjustment">
    <div class="table-toolbar">
      <h2 class="section-title">حركة المخزون</h2>
    </div>

    <form class="adjust-form" @submit.prevent="onSubmit">
      <div class="form-group">
        <label class="form-label">المادة</label>
        <select v-model="selectedItemId" class="form-input" required>
          <option value="" disabled>اختر مادة...</option>
          <option v-for="item in inventoryItems" :key="item.id" :value="item.id">
            {{ item.name_ar }} ({{ item.stock_qty.toLocaleString() }} {{ item.base_unit_ar }})
          </option>
        </select>
      </div>

      <div class="form-group">
        <label class="form-label">النوع</label>
        <div class="type-buttons">
          <button
            v-for="type in adjustTypes"
            :key="type.id"
            type="button"
            class="type-btn"
            :class="{ active: adjustType === type.id }"
            @click="adjustType = type.id"
          >
            <span>{{ type.icon }}</span>
            {{ type.label }}
          </button>
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">الكمية</label>
          <input v-model.number="quantity" type="number" class="form-input" min="1" required />
        </div>
        <div class="form-group" style="flex: 2">
          <label class="form-label">السبب (اختياري)</label>
          <input v-model="reason" type="text" class="form-input" placeholder="مثال: توريد أسبوعي" />
        </div>
      </div>

      <button type="submit" class="btn btn-primary btn-lg" :disabled="isLoading || !selectedItemId || quantity === 0">
        {{ isLoading ? 'جاري التسجيل...' : 'تسجيل الحركة' }}
      </button>

      <div v-if="successMessage" class="success-message">
        {{ successMessage }}
      </div>
    </form>
  </div>
</template>

<style scoped>
.stock-adjustment {
  max-width: 600px;
}

.table-toolbar {
  margin-bottom: var(--gap-lg);
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
}

.adjust-form {
  display: flex;
  flex-direction: column;
  gap: var(--gap-lg);
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
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

select.form-input {
  cursor: pointer;
}

.type-buttons {
  display: flex;
  gap: var(--gap-sm);
}

.type-btn {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-sm) var(--gap-lg);
  border: 2px solid var(--color-border-light);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.type-btn:hover {
  border-color: var(--color-surface-3);
  color: var(--color-text);
}

.type-btn.active {
  border-color: var(--color-accent);
  background: rgba(233, 69, 96, 0.1);
  color: var(--color-accent);
}

.success-message {
  padding: var(--gap-md);
  background: rgba(39, 174, 96, 0.12);
  color: var(--color-success);
  border-radius: var(--radius-md);
  font-weight: var(--font-weight-semi);
  text-align: center;
  animation: slideUp var(--transition-base);
}
</style>
