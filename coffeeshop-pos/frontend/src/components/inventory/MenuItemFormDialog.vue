<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { MenuItem } from '../../types'
import { useManagement } from '../../composables/useManagement'

const { categories } = useManagement()

const props = defineProps<{ editing: MenuItem | null }>()
const emit = defineEmits<{
  save: [data: { category_id: string; name_ar: string; price: number; cost_calc_method: string; manual_cost_price: number; image_path: string }]
  cancel: []
}>()

const categoryId = ref('')
const nameAr = ref('')
const price = ref(0)
const costCalcMethod = ref('auto')
const manualCostPrice = ref(0)
const imagePath = ref('')

onMounted(() => {
  if (props.editing) {
    categoryId.value = props.editing.category_id
    nameAr.value = props.editing.name_ar
    price.value = props.editing.price
    costCalcMethod.value = props.editing.cost_calc_method
    manualCostPrice.value = props.editing.manual_cost_price
    imagePath.value = props.editing.image_path
  } else if (categories.value.length > 0) {
    categoryId.value = categories.value[0].id
  }
})

function onSubmit() {
  if (!nameAr.value.trim() || !categoryId.value) return
  emit('save', {
    category_id: categoryId.value,
    name_ar: nameAr.value.trim(),
    price: price.value,
    cost_calc_method: costCalcMethod.value,
    manual_cost_price: manualCostPrice.value,
    image_path: imagePath.value
  })
}
</script>

<template>
  <div class="overlay" @click.self="emit('cancel')">
    <div class="dialog">
      <h2 class="dialog-title">{{ editing ? '✏️ تعديل منتج' : '➕ إضافة منتج جديد' }}</h2>

      <form @submit.prevent="onSubmit" class="form">
        <div class="field">
          <label>الفئة</label>
          <select v-model="categoryId" required>
            <option disabled value="">اختر الفئة</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name_ar }}
            </option>
          </select>
        </div>
        <div class="field">
          <label>اسم المنتج</label>
          <input v-model="nameAr" type="text" placeholder="مثال: كابتشينو" required />
        </div>
        <div class="field-row">
          <div class="field">
            <label>السعر (د.ع)</label>
            <input v-model.number="price" type="number" min="0" required />
          </div>
          <div class="field">
            <label>مسار الصورة (اختياري)</label>
            <input v-model="imagePath" type="text" placeholder="/images/cup.png" />
          </div>
        </div>
        <div class="field-row">
          <div class="field">
            <label>طريقة حساب التكلفة</label>
            <select v-model="costCalcMethod" required>
              <option value="auto">تلقائي (من المكونات)</option>
              <option value="manual">يدوي</option>
            </select>
          </div>
          <div class="field" v-if="costCalcMethod === 'manual'">
            <label>التكلفة اليدوية (د.ع)</label>
            <input v-model.number="manualCostPrice" type="number" min="0" />
          </div>
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
.field input, .field select {
  padding: 10px 12px;
  background: #222;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 10px;
  color: #f0e6d3;
  font-family: inherit;
  font-size: 0.9rem;
  transition: border-color 0.15s;
}
.field input:focus, .field select:focus { outline: none; border-color: #c9a84c; }

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