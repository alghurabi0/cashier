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
  { id: 'delivery' as const,   label: 'وارد',    icon: '📥', color: '#27ae60' },
  { id: 'waste' as const,      label: 'هالك',    icon: '🗑️', color: '#e74c3c' },
  { id: 'correction' as const, label: 'تعديل',   icon: '🔧', color: '#f39c12' },
]

async function onSubmit() {
  if (!selectedItemId.value || quantity.value === 0) return
  let delta = quantity.value
  if (adjustType.value === 'waste') delta = -Math.abs(delta)
  else if (adjustType.value === 'delivery') delta = Math.abs(delta)
  const label = adjustTypes.find(t => t.id === adjustType.value)?.label
  try {
    await adjustStock(selectedItemId.value, delta, `${label}: ${reason.value || '—'}`)
    successMessage.value = '✓ تم تسجيل الحركة بنجاح'
    quantity.value = 0
    reason.value = ''
    setTimeout(() => { successMessage.value = '' }, 3000)
  } catch (err) { console.error(err) }
}
</script>

<template>
  <div class="stock-adj">
    <h2 class="section-title">📊 حركة المخزون</h2>

    <form class="adj-form" @submit.prevent="onSubmit">
      <div class="field">
        <label>المادة</label>
        <select v-model="selectedItemId" required>
          <option value="" disabled>اختر مادة...</option>
          <option v-for="item in inventoryItems" :key="item.id" :value="item.id">
            {{ item.name_ar }} — {{ item.stock_qty.toLocaleString() }} {{ item.base_unit_ar }}
          </option>
        </select>
      </div>

      <div class="field">
        <label>نوع الحركة</label>
        <div class="type-row">
          <button
            v-for="type in adjustTypes"
            :key="type.id"
            type="button"
            class="type-btn"
            :class="{ active: adjustType === type.id }"
            :style="adjustType === type.id ? `--active-color: ${type.color}` : ''"
            @click="adjustType = type.id"
          >
            {{ type.icon }} {{ type.label }}
          </button>
        </div>
      </div>

      <div class="field-row">
        <div class="field">
          <label>الكمية</label>
          <input v-model.number="quantity" type="number" min="1" required />
        </div>
        <div class="field" style="flex:2">
          <label>السبب (اختياري)</label>
          <input v-model="reason" type="text" placeholder="مثال: توريد أسبوعي" />
        </div>
      </div>

      <button type="submit" class="submit-btn" :disabled="isLoading || !selectedItemId || quantity === 0">
        {{ isLoading ? 'جاري التسجيل...' : '✓ تسجيل الحركة' }}
      </button>

      <div v-if="successMessage" class="success">{{ successMessage }}</div>
    </form>
  </div>
</template>

<style scoped>
.stock-adj { max-width: 580px; display: flex; flex-direction: column; gap: 20px; }

.section-title { font-size: 1rem; font-weight: 800; color: var(--color-text); }

.adj-form { display: flex; flex-direction: column; gap: 16px; }

.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 0.75rem; font-weight: 700; color: var(--color-text-muted); }

.field input, .field select {
  padding: 10px 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  color: var(--color-text);
  font-family: inherit;
  font-size: 0.9rem;
}

.field input:focus, .field select:focus { outline: none; border-color: var(--color-accent); }

.field-row { display: flex; gap: 12px; }
.field-row .field { flex: 1; }

.type-row { display: flex; gap: 8px; }

.type-btn {
  flex: 1; padding: 10px;
  border: 1px solid var(--color-border);
  border-radius: 10px; background: var(--color-surface);
  color: var(--color-text-muted); font-family: inherit;
  font-size: 0.85rem; font-weight: 700; cursor: pointer;
  transition: all 0.15s ease;
}

.type-btn:hover { border-color: var(--color-accent-glow); color: var(--color-text-dim); }

.type-btn.active {
  border-color: var(--active-color, #c9a84c);
  background: color-mix(in srgb, var(--active-color, #c9a84c) 12%, transparent);
  color: var(--active-color, #c9a84c);
}

.submit-btn {
  padding: 13px;
  border: none; border-radius: 12px;
  background: linear-gradient(135deg, var(--color-accent), var(--color-accent-hover));
  color: #0d0d0d; font-family: inherit;
  font-size: 0.95rem; font-weight: 800; cursor: pointer;
  transition: all 0.18s ease;
}

.submit-btn:hover:not(:disabled) { filter: brightness(1.08); }
.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.success {
  padding: 12px; border-radius: 10px;
  background: rgba(39,174,96,0.12); color: var(--color-success);
  font-weight: 700; text-align: center;
  font-size: 0.9rem;
}
</style>