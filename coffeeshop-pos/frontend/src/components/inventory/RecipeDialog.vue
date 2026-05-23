<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useManagement } from '../../composables/useManagement'
import type { MenuItem, RecipeIngredientWithDetails } from '../../types'

const props = defineProps<{
  menuItem: MenuItem
}>()

const emit = defineEmits<{
  save: []
  cancel: []
}>()

const { inventoryItems, getRecipe, setRecipe, isLoading } = useManagement()

interface IngredientRow {
  inventory_item_id: string
  quantity: number
}

const ingredients = ref<IngredientRow[]>([])

// Auto-cost preview
const autoCost = computed(() => {
  let total = 0
  for (const ing of ingredients.value) {
    const invItem = inventoryItems.value.find(i => i.id === ing.inventory_item_id)
    if (invItem) {
      total += ing.quantity * invItem.unit_cost
    }
  }
  return total
})

// Available inventory items not yet added
const availableItems = computed(() => {
  const usedIds = new Set(ingredients.value.map(i => i.inventory_item_id))
  return inventoryItems.value.filter(i => !usedIds.has(i.id))
})

function addIngredient() {
  if (availableItems.value.length === 0) return
  ingredients.value.push({
    inventory_item_id: availableItems.value[0].id,
    quantity: 1,
  })
}

function removeIngredient(index: number) {
  ingredients.value.splice(index, 1)
}

function getItemName(id: string): string {
  const item = inventoryItems.value.find(i => i.id === id)
  return item?.name_ar ?? '—'
}

function getItemUnit(id: string): string {
  const item = inventoryItems.value.find(i => i.id === id)
  return item?.base_unit_ar ?? ''
}

async function onSave() {
  try {
    await setRecipe(props.menuItem.id, ingredients.value)
    emit('save')
  } catch (err) {
    console.error('Failed to save recipe:', err)
  }
}

onMounted(async () => {
  try {
    const existing = await getRecipe(props.menuItem.id)
    if (existing && existing.length > 0) {
      ingredients.value = existing.map((e: RecipeIngredientWithDetails) => ({
        inventory_item_id: e.inventory_item_id,
        quantity: e.quantity,
      }))
    }
  } catch {
    // No recipe yet, start empty
  }
})
</script>

<template>
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal-content recipe-dialog">
      <h2 class="modal-title">وصفة: {{ menuItem.name_ar }}</h2>

      <div class="ingredients-list">
        <div v-if="ingredients.length === 0" class="empty-hint text-muted">
          لا توجد مكونات. اضغط "إضافة مكون" لتبدأ.
        </div>

        <div v-for="(ing, idx) in ingredients" :key="idx" class="ingredient-row">
          <select v-model="ing.inventory_item_id" class="form-input ingredient-select">
            <option v-for="item in inventoryItems" :key="item.id" :value="item.id">
              {{ item.name_ar }} ({{ item.base_unit_ar }})
            </option>
          </select>
          <input
            v-model.number="ing.quantity"
            type="number"
            class="form-input ingredient-qty"
            min="1"
            placeholder="الكمية"
          />
          <span class="ingredient-unit text-muted">{{ getItemUnit(ing.inventory_item_id) }}</span>
          <button class="btn btn-icon btn-ghost" @click="removeIngredient(idx)" title="حذف">✕</button>
        </div>
      </div>

      <button
        class="btn btn-ghost add-ingredient-btn"
        @click="addIngredient"
        :disabled="availableItems.length === 0"
      >
        + إضافة مكون
      </button>

      <div class="cost-preview" v-if="ingredients.length > 0">
        <span class="cost-label">التكلفة التلقائية:</span>
        <span class="cost-value">{{ autoCost.toLocaleString() }} <small>د.ع</small></span>
      </div>

      <div class="form-actions">
        <button class="btn btn-ghost" @click="emit('cancel')">إلغاء</button>
        <button class="btn btn-primary" :disabled="isLoading" @click="onSave">
          {{ isLoading ? 'جاري الحفظ...' : 'حفظ الوصفة' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.recipe-dialog {
  min-width: 500px;
}

.ingredients-list {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
  margin: var(--gap-lg) 0 var(--gap-md);
  max-height: 300px;
  overflow-y: auto;
}

.ingredient-row {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-sm);
  background: var(--color-surface-2);
  border-radius: var(--radius-sm);
}

.ingredient-select {
  flex: 1;
  min-width: 0;
}

.ingredient-qty {
  width: 80px;
  text-align: center;
}

.ingredient-unit {
  min-width: 40px;
  font-size: var(--font-size-sm);
}

.form-input {
  padding: var(--gap-sm) var(--gap-md);
  background: var(--color-bg);
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

.add-ingredient-btn {
  align-self: flex-start;
  margin-bottom: var(--gap-md);
}

.cost-preview {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding: var(--gap-md);
  background: var(--color-surface);
  border-radius: var(--radius-md);
  margin-bottom: var(--gap-lg);
}

.cost-label {
  font-weight: var(--font-weight-semi);
  color: var(--color-text-muted);
}

.cost-value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-extra);
  color: var(--color-success);
}

.cost-value small {
  font-size: var(--font-size-sm);
  opacity: 0.7;
}

.empty-hint {
  text-align: center;
  padding: var(--gap-lg);
}

.form-actions {
  display: flex;
  gap: var(--gap-md);
  justify-content: flex-end;
}
</style>
