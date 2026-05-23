<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Category } from '../../types'

const props = defineProps<{
  editing: Category | null
}>()

const emit = defineEmits<{
  save: [data: { name_ar: string; sort_order: number }]
  cancel: []
}>()

const nameAr = ref('')
const sortOrder = ref(0)

watch(() => props.editing, (item) => {
  if (item) {
    nameAr.value = item.name_ar
    sortOrder.value = item.sort_order
  } else {
    nameAr.value = ''
    sortOrder.value = 0
  }
}, { immediate: true })

function onSubmit() {
  if (!nameAr.value.trim()) return
  emit('save', {
    name_ar: nameAr.value.trim(),
    sort_order: sortOrder.value,
  })
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal-content">
      <h2 class="modal-title">{{ editing ? 'تعديل فئة' : 'إضافة فئة جديدة' }}</h2>

      <form class="form" @submit.prevent="onSubmit">
        <div class="form-group">
          <label class="form-label">اسم الفئة</label>
          <input
            v-model="nameAr"
            type="text"
            class="form-input"
            placeholder="مثال: مشروبات ساخنة"
            autofocus
          />
        </div>

        <div class="form-group">
          <label class="form-label">ترتيب العرض</label>
          <input
            v-model.number="sortOrder"
            type="number"
            class="form-input"
            min="0"
            style="width: 120px"
          />
          <p class="form-hint text-muted">الأصغر يظهر أولاً</p>
        </div>

        <div class="form-actions">
          <button type="button" class="btn btn-ghost" @click="emit('cancel')">إلغاء</button>
          <button type="submit" class="btn btn-primary" :disabled="!nameAr.trim()">
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
