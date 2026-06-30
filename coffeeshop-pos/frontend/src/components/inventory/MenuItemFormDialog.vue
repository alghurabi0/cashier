<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Dialogs } from '@wailsio/runtime'
import type { MenuItem } from '../../types'
import { useManagement } from '../../composables/useManagement'

const { categories, uploadMenuItemImage } = useManagement()

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
const selectedImageFilePath = ref('')
const selectedImageFileName = ref('')
const isUploadingImage = ref(false)
const uploadError = ref<string | null>(null)

onMounted(() => {
  if (props.editing) {
    categoryId.value = props.editing.category_id
    nameAr.value = props.editing.name_ar
    price.value = props.editing.price
    costCalcMethod.value = props.editing.cost_calc_method
    manualCostPrice.value = props.editing.manual_cost_price
    imagePath.value = props.editing.image_path || ''
  } else if (categories.value.length > 0) {
    categoryId.value = categories.value[0].id
  }
})

function getFileName(filePath: string): string {
  return filePath.split(/[\\/]/).pop() || filePath
}

async function chooseImage() {
  uploadError.value = null
  const selected = await Dialogs.OpenFile({
    CanChooseFiles: true,
    CanChooseDirectories: false,
    AllowsMultipleSelection: false,
    Title: 'اختر صورة المنتج',
    ButtonText: 'اختيار',
    Filters: [
      { DisplayName: 'Images', Pattern: '*.jpg;*.jpeg;*.png;*.webp' },
    ],
  })
  const filePath = Array.isArray(selected) ? selected[0] : selected
  if (!filePath) return

  if (!/\.(jpe?g|png|webp)$/i.test(filePath)) {
    uploadError.value = 'نوع الصورة غير مدعوم. اختر صورة JPG أو PNG أو WebP.'
    return
  }

  selectedImageFilePath.value = filePath
  selectedImageFileName.value = getFileName(filePath)
}

function removeImage() {
  imagePath.value = ''
  selectedImageFilePath.value = ''
  selectedImageFileName.value = ''
  uploadError.value = null
}

function cancelDialog() {
  if (isUploadingImage.value) return
  emit('cancel')
}

async function onSubmit() {
  if (!nameAr.value.trim() || !categoryId.value) return
  uploadError.value = null

  let finalImagePath = imagePath.value
  if (selectedImageFilePath.value) {
    isUploadingImage.value = true
    try {
      finalImagePath = await uploadMenuItemImage(selectedImageFilePath.value)
      imagePath.value = finalImagePath
      selectedImageFilePath.value = ''
      selectedImageFileName.value = ''
    } catch (err) {
      uploadError.value = 'يتطلب رفع الصور اتصالاً بالإنترنت. تأكد من الاتصال ثم حاول مرة أخرى.'
      console.error('Image upload failed:', err)
      return
    } finally {
      isUploadingImage.value = false
    }
  }

  emit('save', {
    category_id: categoryId.value,
    name_ar: nameAr.value.trim(),
    price: price.value,
    cost_calc_method: costCalcMethod.value,
    manual_cost_price: manualCostPrice.value,
    image_path: finalImagePath
  })
}
</script>

<template>
  <div class="overlay" @click.self="cancelDialog">
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
        </div>
        <div class="field">
          <label>صورة المنتج</label>
          <div class="image-picker">
            <div class="image-preview" :class="{ empty: !imagePath && !selectedImageFileName }">
              <img v-if="imagePath && !selectedImageFileName" :src="imagePath" :alt="nameAr || 'صورة المنتج'" />
              <div v-else class="image-placeholder">
                <svg viewBox="0 0 64 64" aria-hidden="true">
                  <path d="M18 44h30a10 10 0 0 0 1.2-19.9A17 17 0 0 0 16.6 19 12.5 12.5 0 0 0 18 44Z" />
                  <path d="M32 24v16" />
                  <path d="m25 31 7-7 7 7" />
                </svg>
                <span>{{ selectedImageFileName || 'لم يتم اختيار صورة' }}</span>
              </div>
            </div>
            <div class="image-actions">
              <button type="button" class="secondary-btn" @click="chooseImage" :disabled="isUploadingImage">
                {{ imagePath || selectedImageFileName ? 'تغيير الصورة' : 'اختيار صورة' }}
              </button>
              <button
                v-if="imagePath || selectedImageFileName"
                type="button"
                class="danger-btn"
                @click="removeImage"
                :disabled="isUploadingImage"
              >
                إزالة
              </button>
            </div>
          </div>
          <div v-if="uploadError" class="upload-error">
            <svg viewBox="0 0 64 64" aria-hidden="true">
              <path d="M18 44h30a10 10 0 0 0 1.2-19.9A17 17 0 0 0 16.6 19 12.5 12.5 0 0 0 18 44Z" />
              <path d="M20 52 48 16" />
            </svg>
            <span>{{ uploadError }}</span>
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
          <button type="button" class="cancel-btn" @click="cancelDialog" :disabled="isUploadingImage">إلغاء</button>
          <button type="submit" class="save-btn" :disabled="isUploadingImage">
            {{ isUploadingImage ? 'جاري رفع الصورة...' : (editing ? 'حفظ التعديلات' : 'إضافة') }}
          </button>
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
  background: var(--color-surface);
  border: 1px solid var(--color-border-light);
  border-radius: 18px;
  padding: 24px;
  width: min(540px, 92vw);
  max-height: 92vh;
  overflow-y: auto;
  box-shadow: var(--shadow-lg);
  animation: slideUp 0.2s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(14px); }
  to   { opacity: 1; transform: translateY(0); }
}

.dialog-title { font-size: 1rem; font-weight: 800; color: var(--color-text); margin-bottom: 20px; }

.form { display: flex; flex-direction: column; gap: 14px; }

.field { display: flex; flex-direction: column; gap: 5px; }
.field label { font-size: 0.75rem; font-weight: 700; color: var(--color-text-muted); }
.field input, .field select {
  padding: 10px 12px;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  color: var(--color-text);
  font-family: inherit;
  font-size: 0.9rem;
  transition: border-color 0.15s;
}
.field input:focus, .field select:focus { outline: none; border-color: var(--color-accent); }

.field-row { display: flex; gap: 12px; }
.field-row .field { flex: 1; }

.image-picker {
  display: grid;
  grid-template-columns: 132px 1fr;
  gap: 12px;
  align-items: stretch;
}

.image-preview {
  position: relative;
  min-height: 104px;
  overflow: hidden;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-bg);
}

.image-preview.empty {
  border-style: dashed;
  border-color: var(--color-border-light);
}

.image-preview img {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 104px;
  object-fit: cover;
}

.image-placeholder {
  min-height: 104px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px;
  color: #777;
  text-align: center;
  font-size: 0.78rem;
  overflow-wrap: anywhere;
}

.image-placeholder svg,
.upload-error svg {
  width: 34px;
  height: 34px;
  fill: none;
  stroke: currentColor;
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.image-actions {
  display: flex;
  flex-wrap: wrap;
  align-content: center;
  gap: 8px;
}

.secondary-btn,
.danger-btn {
  min-height: 38px;
  padding: 8px 14px;
  border-radius: 10px;
  font-family: inherit;
  font-size: 0.82rem;
  font-weight: 800;
  cursor: pointer;
}

.secondary-btn {
  border: 1px solid var(--color-border-light);
  background: var(--color-border-light);
  color: var(--color-accent-hover);
}

.danger-btn {
  border: 1px solid rgba(231,76,60,0.28);
  background: rgba(231,76,60,0.08);
  color: #ff8c7f;
}

.upload-error {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
  padding: 10px 12px;
  border: 1px solid rgba(231,76,60,0.26);
  border-radius: 10px;
  background: rgba(231,76,60,0.08);
  color: #ffb0a8;
  font-size: 0.8rem;
  line-height: 1.5;
}

.upload-error svg {
  flex: 0 0 auto;
  width: 42px;
  height: 42px;
}

.form-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 4px; }

.cancel-btn {
  padding: 10px 18px; border: 1px solid var(--color-border);
  border-radius: 10px; background: transparent; color: var(--color-text-muted);
  font-family: inherit; font-size: 0.88rem; font-weight: 700; cursor: pointer;
}

.save-btn {
  padding: 10px 24px; border: none;
  border-radius: 10px; background: linear-gradient(135deg, var(--color-accent), var(--color-accent-hover));
  color: #0d0d0d; font-family: inherit; font-size: 0.88rem; font-weight: 800; cursor: pointer;
}

.save-btn:hover { filter: brightness(1.08); }

button:disabled {
  opacity: 0.58;
  cursor: not-allowed;
  transform: none;
}

@media (max-width: 560px) {
  .image-picker {
    grid-template-columns: 1fr;
  }

  .image-actions {
    justify-content: flex-start;
  }
}
</style>
