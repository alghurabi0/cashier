<script setup lang="ts">
import type { MenuItem } from '../types'
import { formatPrice } from '../types'

defineProps<{
  item: MenuItem
}>()
</script>

<template>
  <div class="menu-card">
    <div class="card-image-wrap">
      <img class="card-image" :src="item.image_path || '/placeholder.svg'" :alt="item.name_ar" loading="lazy" />
    </div>
    <div class="card-body">
      <h3 class="card-name">{{ item.name_ar }}</h3>
      <span class="card-price">{{ formatPrice(item.price) }} <small>د.ع</small></span>
    </div>
    <div class="card-action">
      <div class="view-btn"><span>+</span></div>
    </div>
  </div>
</template>

<style scoped>
.menu-card {
  display: flex; align-items: center; gap: 14px;
  padding: 11px 13px; background: #1c1c1c;
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 18px; cursor: pointer;
  direction: rtl; text-align: right;
  transition: all 0.2s ease; position: relative; overflow: hidden;
}

.menu-card::before {
  content: ''; position: absolute; top: 0; right: 0;
  width: 60px; height: 100%;
  background: linear-gradient(to left, rgba(201,168,76,0.04), transparent);
  pointer-events: none;
}

.menu-card:active { transform: scale(0.975); border-color: rgba(201,168,76,0.3); }

.card-image-wrap {
  flex-shrink: 0; width: 88px; height: 88px;
  border-radius: 14px; overflow: hidden; background: #111;
}

.card-image { width: 100%; height: 100%; object-fit: cover; display: block; transition: transform 0.4s ease; }
.menu-card:active .card-image { transform: scale(1.05); }

.card-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }

.card-name {
  font-size: 0.96rem; font-weight: 800; color: #f0e6d3;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin: 0;
}

.card-price { font-size: 1rem; font-weight: 800; color: #c9a84c; margin-top: 3px; }
.card-price small { font-size: 0.63rem; font-weight: 600; opacity: 0.75; margin-right: 2px; }

.card-action { flex-shrink: 0; }

.view-btn {
  width: 40px; height: 40px; border-radius: 50%;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d; display: flex; align-items: center; justify-content: center;
  font-size: 1.5rem; font-weight: 300;
  box-shadow: 0 4px 14px rgba(201,168,76,0.35); transition: all 0.2s;
}

.menu-card:active .view-btn { transform: scale(0.9); }
</style>