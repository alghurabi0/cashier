<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useManagement } from '../../composables/useManagement'
import type { MenuItem } from '../../types'

const props = defineProps<{
  editing: MenuItem | null
}>()

const emit = defineEmits<{
  save: [data: { category_id: string; name_ar: string; price: number; cost_calc_method: string; manual_cost_price: number }]
  cancel: []
}>()

const { categories } = useManagement()

const categoryId = ref('')
const nameAr = ref('')
const price = ref(0)
const costCalcMethod = ref('auto')
const manualCostPrice = ref(0)

const showManualCost = computed(() => costCalcMethod.value === 'manual')

watch(() => props.editing, (item) => {
  if (item) {
    categoryId.value = item.category_id
    nameAr.value = item.name_ar
    price.value = item.price
    costCalcMethod.value = item.cost_calc_method
    manualCostPrice.value = item.manual_cost_price
  } else {
    categoryId.value = categories.value.length > 0 ? categories.value[0].id : ''
    nameAr.value = ''
    price.value = 0
    costCalcMethod.value = 'auto'
    manualCostPrice.value = 0
  }
}, { immediate: true })

function onSubmit() {
  if (!nameAr.value.trim() || !categoryId.value || price.value <= 0) return
  emit('save', {
    category_id: categoryId.value,
    name_ar: nameAr.value.trim(),
    price: price.value,
    cost_calc_method: costCalcMethod.value,
    manual_cost_price: showManualCost.value ? manualCostPrice.value : 0,
  })
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal-content">
      <h2 class="modal-title">{{ editing ? 'تعديل منتج' : 'إضافة منتج جديد' }}</h2>

      <form class="form" @submit.prevent="onSubmit">
        <div class="form-group">
          <label class="form-label">اسم المنتج</label>
          <input
            v-model="nameAr"
            type="text"
            class="form-input"
            placeholder="مثال: لاتيه"
            autofocus
          />
        </div>

        <div class="form-group">
          <label class="form-label">الفئة</label>
          <select v-model="categoryId" class="form-input">
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name_ar }}
            </option>
          </select>
          <p v-if="categories.length === 0" class="form-hint" style="color: var(--color-danger)">
            ⚠️ أضف فئة من تبويب "الفئات" أولاً
          </p>
        </div>

        <div class="form-group">
          <label class="form-label">السعر (د.ع)</label>
          <input
            v-model.number="price"
            type="number"
            class="form-input"
            min="0"
            step="100"
            style="width: 180px"
          />
        </div>

        <div class="form-group">
          <label class="form-label">طريقة حساب التكلفة</label>
          <select v-model="costCalcMethod" class="form-input" style="width: 200px">
            <option value="auto">تلقائي (من الوصفة)</option>
            <option value="manual">يدوي</option>
          </select>
        </div>

        <div v-if="showManualCost" class="form-group">
          <label class="form-label">التكلفة اليدوية (د.ع)</label>
          <input
            v-model.number="manualCostPrice"
            type="number"
            class="form-input"
            min="0"
            step="100"
            style="width: 180px"
          />
        </div>

        <div class="form-actions">
          <button type="button" class="btn btn-ghost" @click="emit('cancel')">إلغاء</button>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="!nameAr.trim() || !categoryId || price <= 0"
          >
            {{ editing ? 'حفظ التعديل' : 'إضافة' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
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

.form-hint {
  font-size: var(--font-size-xs);
}

.form-actions {
  display: flex;
  gap: var(--gap-md);
  justify-content: flex-end;
  margin-top: var(--gap-sm);
}
</style>
