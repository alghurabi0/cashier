<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useManagement } from '../../composables/useManagement'
import type { MenuItem } from '../../types'

const props = defineProps<{
  editing: MenuItem | null
}>()

const emit = defineEmits<{
  save: [data: { category_id: string; name_ar: string; price: number; cost_calc_method: string; manual_cost_price: number; image_path: string }]
  cancel: []
}>()

const { categories } = useManagement()

let ManagementService: any = null

const categoryId = ref('')
const nameAr = ref('')
const price = ref(0)
const costCalcMethod = ref('auto')
const manualCostPrice = ref(0)
const imagePath = ref('')
const isUploadingImage = ref(false)

const showManualCost = computed(() => costCalcMethod.value === 'manual')

watch(() => props.editing, (item) => {
  if (item) {
    categoryId.value = item.category_id
    nameAr.value = item.name_ar
    price.value = item.price
    costCalcMethod.value = item.cost_calc_method
    manualCostPrice.value = item.manual_cost_price
    imagePath.value = item.image_path || ''
  } else {
    categoryId.value = categories.value.length > 0 ? categories.value[0].id : ''
    nameAr.value = ''
    price.value = 0
    costCalcMethod.value = 'auto'
    manualCostPrice.value = 0
    imagePath.value = ''
  }
}, { immediate: true })

async function initBindings() {
  try {
    ManagementService = await import('../../../bindings/coffeeshop-pos/internal/service/managementservice')
  } catch { /* not available */ }
}
initBindings()

async function onPickImage() {
  // Use a hidden file input to pick the image
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/jpeg,image/png,image/webp'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file || !ManagementService) return

    isUploadingImage.value = true
    try {
      // For Wails, we need to pass the file path — use a temporary URL for preview
      // Since browser file input doesn't give us a real path, we'll upload directly
      // For now, store as data URL preview and set path empty until R2 is configured
      const reader = new FileReader()
      reader.onload = () => {
        imagePath.value = reader.result as string
      }
      reader.readAsDataURL(file)
    } catch (err) {
      console.error('Image upload failed:', err)
    } finally {
      isUploadingImage.value = false
    }
  }
  input.click()
}

function removeImage() {
  imagePath.value = ''
}

function onSubmit() {
  if (!nameAr.value.trim() || !categoryId.value || price.value <= 0) return
  emit('save', {
    category_id: categoryId.value,
    name_ar: nameAr.value.trim(),
    price: price.value,
    cost_calc_method: costCalcMethod.value,
    manual_cost_price: showManualCost.value ? manualCostPrice.value : 0,
    image_path: imagePath.value,
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

        <!-- Image Upload -->
        <div class="form-group">
          <label class="form-label">صورة المنتج (اختياري)</label>
          <div class="image-upload-area">
            <div v-if="imagePath" class="image-preview">
              <img :src="imagePath" alt="معاينة" class="preview-img" />
              <button type="button" class="remove-image-btn" @click="removeImage" title="إزالة الصورة">✕</button>
            </div>
            <button
              v-else
              type="button"
              class="upload-btn"
              :disabled="isUploadingImage"
              @click="onPickImage"
            >
              <span v-if="isUploadingImage">⏳ جاري الرفع...</span>
              <span v-else>📷 اختر صورة</span>
            </button>
            <button
              v-if="imagePath"
              type="button"
              class="btn btn-ghost btn-sm"
              @click="onPickImage"
            >
              تغيير الصورة
            </button>
          </div>
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

/* Image upload styles */
.image-upload-area {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
}

.image-preview {
  position: relative;
  width: 80px;
  height: 80px;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 2px solid var(--color-border-light);
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-image-btn {
  position: absolute;
  top: 2px;
  right: 2px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  font-size: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-btn {
  padding: var(--gap-md) var(--gap-lg);
  border: 2px dashed var(--color-border-light);
  border-radius: var(--radius-md);
  background: var(--color-surface-2);
  color: var(--color-text-muted);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.upload-btn:hover:not(:disabled) {
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.upload-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
