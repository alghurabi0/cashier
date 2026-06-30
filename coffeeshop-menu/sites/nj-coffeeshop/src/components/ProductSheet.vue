<script setup lang="ts">
import { ref, watch } from 'vue'
import type { MenuItem } from '../types'
import { formatPrice } from '../types'

const props = defineProps<{
  item: MenuItem | null
  tr: Record<string, string>
}>()

const emit = defineEmits<{
  close: []
  add: [menuItemId: string, nameAr: string, price: number, qty: number]
}>()

const qty = ref(1)
watch(() => props.item, () => { qty.value = 1 })

function onAdd() {
  if (!props.item) return
  emit('add', props.item.id, props.item.name_ar, props.item.price, qty.value)
  emit('close')
}
</script>

<template>
  <Transition name="sheet">
    <div v-if="item" class="sheet-backdrop" @click.self="emit('close')">
      <div class="sheet">
        <div class="sheet-hero">
          <img :src="item.image_path || '/placeholder.svg'" :alt="item.name_ar" class="sheet-img" />
          <div class="sheet-img-overlay" />
          <button class="sheet-close" @click="emit('close')">✕</button>
          <div class="sheet-badge">{{ tr.featured }}</div>
        </div>

        <div class="sheet-body">
          <h2 class="sheet-name">{{ item.name_ar }}</h2>

          <div class="sheet-tags">
            <span class="tag">{{ tr.fresh }}</span>
            <span class="tag">{{ tr.special }}</span>
            <span class="tag">{{ tr.fast }}</span>
          </div>

          <div class="sheet-footer">
            <div class="qty-control">
              <button class="qty-btn" @click="qty = Math.max(1, qty - 1)">−</button>
              <span class="qty-num">{{ qty }}</span>
              <button class="qty-btn" @click="qty++">+</button>
            </div>
            <button class="add-to-cart-btn" @click="onAdd">
              <span>{{ tr.addtocart }}</span>
              <span class="btn-price">{{ formatPrice(item.price * qty) }} {{ tr.currency }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.sheet-backdrop {
  position: fixed; inset: 0; z-index: 200;
  background: rgba(0,0,0,0.78); backdrop-filter: blur(8px);
  display: flex; align-items: flex-end;
}

.sheet {
  width: 100%; background: #1a1a1a;
  border-radius: 28px 28px 0 0; overflow: hidden;
  max-height: 90dvh; display: flex; flex-direction: column;
}

.sheet-hero { position: relative; height: 270px; flex-shrink: 0; }
.sheet-img { width: 100%; height: 100%; object-fit: cover; }
.sheet-img-overlay {
  position: absolute; inset: 0;
  background: linear-gradient(to bottom, rgba(0,0,0,0.1) 0%, rgba(26,26,26,0.9) 100%);
}

.sheet-close {
  position: absolute; top: 14px; left: 14px;
  width: 36px; height: 36px; border-radius: 50%; border: none;
  background: rgba(0,0,0,0.6); color: #fff; font-size: 0.9rem;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  backdrop-filter: blur(4px); z-index: 2;
}

.sheet-badge {
  position: absolute; top: 14px; right: 14px;
  padding: 5px 12px;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d; font-size: 0.72rem; font-weight: 800;
  border-radius: 50px; z-index: 2;
}

.sheet-body {
  padding: 22px 20px 36px; overflow-y: auto;
  display: flex; flex-direction: column; gap: 14px;
  direction: rtl; text-align: right;
}

.sheet-name { font-size: 1.55rem; font-weight: 900; color: #f0e6d3; margin: 0; }

.sheet-tags { display: flex; gap: 8px; flex-wrap: wrap; }
.tag {
  padding: 5px 13px;
  background: rgba(201,168,76,0.09);
  border: 1px solid rgba(201,168,76,0.22);
  border-radius: 50px; font-size: 0.75rem; color: #c9a84c; font-weight: 700;
}

.sheet-footer { display: flex; align-items: center; gap: 12px; margin-top: 4px; }

.qty-control {
  display: flex; align-items: center;
  background: #111; border: 1px solid rgba(201,168,76,0.2);
  border-radius: 50px; overflow: hidden; flex-shrink: 0;
}

.qty-btn {
  width: 46px; height: 46px; border: none; background: transparent;
  color: #c9a84c; font-size: 1.5rem; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
}
.qty-btn:active { background: rgba(201,168,76,0.12); }
.qty-num { min-width: 34px; text-align: center; font-size: 1rem; font-weight: 800; color: #f0e6d3; }

.add-to-cart-btn {
  flex: 1; display: flex; align-items: center; justify-content: space-between;
  padding: 14px 22px;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  border: none; border-radius: 50px; color: #0d0d0d;
  font-family: 'Cairo', sans-serif; font-size: 0.98rem; font-weight: 800;
  cursor: pointer; box-shadow: 0 8px 28px rgba(201,168,76,0.4); transition: all 0.2s;
}
.add-to-cart-btn:active { transform: scale(0.97); }
.btn-price { font-size: 0.83rem; font-weight: 700; opacity: 0.8; }

.sheet-enter-active { transition: opacity 0.28s ease; }
.sheet-leave-active { transition: opacity 0.22s ease; }
.sheet-enter-active .sheet { transition: transform 0.35s cubic-bezier(0.32,0.72,0,1); }
.sheet-leave-active .sheet { transition: transform 0.25s ease-in; }
.sheet-enter-from, .sheet-leave-to { opacity: 0; }
.sheet-enter-from .sheet, .sheet-leave-to .sheet { transform: translateY(100%); }
</style>