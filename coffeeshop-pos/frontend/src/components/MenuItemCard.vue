<script setup lang="ts">
import type { MenuItem } from '../types'
import { formatPrice } from '../types'

defineProps<{
  item: MenuItem
}>()

const emit = defineEmits<{
  add: [item: MenuItem]
}>()
</script>

<template>
  <button class="item-card" @click="emit('add', item)">
    <img v-if="item.image_path" :src="item.image_path" :alt="item.name_ar" class="card-img" loading="lazy" />
    <div class="card-overlay" v-if="item.image_path"></div>
    <div class="card-body">
      <span class="item-name">{{ item.name_ar }}</span>
      <span class="item-price">{{ formatPrice(item.price) }} <small>د.ع</small></span>
    </div>
    <div class="add-circle">+</div>
  </button>
</template>

<style scoped>
.item-card {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: 14px;
  background: #1a1a1a;
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 16px;
  cursor: pointer;
  min-height: 130px;
  font-family: var(--font-family);
  text-align: right;
  overflow: hidden;
  transition: all 0.2s ease;
  user-select: none;
}

.item-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(201,168,76,0.06) 0%, transparent 60%);
  opacity: 0;
  transition: opacity 0.2s;
}

.item-card:hover {
  border-color: rgba(201,168,76,0.4);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.4);
}

.item-card:hover::before { opacity: 1; }
.item-card:active { transform: scale(0.97); }

.card-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  z-index: 0;
  transition: transform 0.3s ease;
}

.item-card:hover .card-img { transform: scale(1.05); }

.card-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.8) 0%, rgba(0,0,0,0.2) 60%, transparent 100%);
  z-index: 1;
}

.card-body {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.item-name {
  font-size: 0.92rem;
  font-weight: 700;
  color: #f0e6d3;
  line-height: 1.3;
}

.item-price {
  font-size: 1rem;
  font-weight: 800;
  color: #c9a84c;
  line-height: 1;
}

.item-price small {
  font-size: 0.65rem;
  font-weight: 600;
  opacity: 0.75;
  margin-right: 2px;
}

.add-circle {
  position: absolute;
  bottom: 10px;
  left: 10px;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
  font-weight: 900;
  z-index: 3;
  opacity: 0;
  transform: scale(0.7);
  transition: all 0.2s ease;
}

.item-card:hover .add-circle {
  opacity: 1;
  transform: scale(1);
}
</style>