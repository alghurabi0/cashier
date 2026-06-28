<script setup lang="ts">
import { useCart } from '../composables/useCart'
import { formatPrice } from '../types'

const { orderResult } = useCart()
</script>

<template>
  <div class="confirmation-view">
    <div class="confirm-content" v-if="orderResult">
      <div class="success-icon">✅</div>
      <h1 class="confirm-title">تم إرسال طلبك!</h1>
      <p class="confirm-subtitle">طلبك قيد التحضير</p>

      <div class="order-summary">
        <div class="summary-row">
          <span class="label">رقم الطلب</span>
          <span class="value">#{{ orderResult.order_number }}</span>
        </div>
        <div class="summary-row">
          <span class="label">الطاولة</span>
          <span class="value">🪑 {{ orderResult.table_number }}</span>
        </div>
        <div class="summary-row">
          <span class="label">المجموع</span>
          <span class="value total">{{ formatPrice(orderResult.total) }} <small>د.ع</small></span>
        </div>
      </div>

      <div class="items-list">
        <div v-for="item in orderResult.items" :key="item.id" class="item-row">
          <span class="item-qty">×{{ item.quantity }}</span>
          <span class="item-name">{{ item.name_ar_snapshot }}</span>
        </div>
      </div>

      <div class="waiting-indicator">
        <span class="waiting-icon">☕</span>
        <span>في انتظار تأكيد المقهى...</span>
      </div>
    </div>

    <div v-else class="confirm-content">
      <p class="text-muted">لا يوجد طلب</p>
    </div>
  </div>
</template>

<style scoped>
.confirmation-view {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--gap-xl);
}

.confirm-content {
  text-align: center;
  max-width: 400px;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-lg);
}

.success-icon {
  font-size: 4rem;
  animation: bounceIn 0.5s ease;
}

@keyframes bounceIn {
  0% { transform: scale(0); }
  60% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

.confirm-title {
  font-size: var(--font-size-2xl);
  font-weight: 800;
  color: var(--color-accent);
}

.confirm-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-muted);
}

.order-summary {
  width: 100%;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  padding: var(--gap-lg);
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.summary-row .label {
  color: var(--color-text-muted);
  font-weight: 600;
}

.summary-row .value {
  font-weight: 800;
}

.summary-row .value.total {
  font-size: var(--font-size-xl);
  color: var(--color-accent);
}

.summary-row .value.total small {
  font-size: var(--font-size-xs);
  opacity: 0.7;
}

.items-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
}

.item-row {
  display: flex;
  gap: var(--gap-sm);
  padding: var(--gap-sm) 0;
  border-bottom: 1px solid var(--color-border);
  font-size: var(--font-size-sm);
}

.item-qty {
  color: var(--color-accent);
  font-weight: 700;
  min-width: 28px;
}

.item-name {
  flex: 1;
  text-align: right;
}

.waiting-indicator {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-md) var(--gap-lg);
  background: rgba(212, 165, 116, 0.1);
  border-radius: var(--radius-full);
  color: var(--color-accent);
  font-weight: 600;
  animation: pulse 2s ease infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.7; }
  50% { opacity: 1; }
}

.waiting-icon {
  font-size: 1.3rem;
}
</style>
