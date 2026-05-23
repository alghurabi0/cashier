<script setup lang="ts">
import type { CartItem } from '../types'
import { formatPrice } from '../types'

defineProps<{
  items: CartItem[]
  total: number
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal-content checkout-modal">
      <h2 class="modal-title">تأكيد الطلب</h2>

      <div class="order-summary">
        <div
          v-for="item in items"
          :key="item.menu_item_id"
          class="summary-item"
        >
          <span class="summary-name">{{ item.name_ar }}</span>
          <span class="summary-qty">×{{ item.quantity }}</span>
          <span class="summary-price">{{ formatPrice(item.price * item.quantity) }}</span>
        </div>
      </div>

      <div class="summary-divider"></div>

      <div class="summary-total">
        <span class="total-label">المجموع</span>
        <span class="total-value">{{ formatPrice(total) }} <small>د.ع</small></span>
      </div>

      <div class="payment-section">
        <span class="payment-label text-muted text-sm">طريقة الدفع</span>
        <div class="payment-methods">
          <button class="payment-btn active">
            <span class="payment-icon">💵</span>
            <span>نقدي</span>
          </button>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn btn-ghost" @click="emit('cancel')">إلغاء</button>
        <button class="btn btn-primary btn-lg" @click="emit('confirm')">
          ✓ تأكيد ودفع
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.checkout-modal {
  min-width: 450px;
}

.modal-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  margin-bottom: var(--gap-lg);
  text-align: center;
}

.order-summary {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
  max-height: 300px;
  overflow-y: auto;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-sm) 0;
}

.summary-name {
  flex: 1;
  font-weight: var(--font-weight-semi);
}

.summary-qty {
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.summary-price {
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  min-width: 70px;
  text-align: left;
}

.summary-divider {
  height: 1px;
  background: var(--color-border-light);
  margin: var(--gap-md) 0;
}

.summary-total {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: var(--gap-lg);
}

.summary-total .total-label {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semi);
}

.summary-total .total-value {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-extra);
  color: var(--color-accent);
}

.summary-total .total-value small {
  font-size: var(--font-size-sm);
  opacity: 0.7;
}

.payment-section {
  margin-bottom: var(--gap-lg);
}

.payment-label {
  display: block;
  margin-bottom: var(--gap-sm);
}

.payment-methods {
  display: flex;
  gap: var(--gap-sm);
}

.payment-btn {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-md) var(--gap-lg);
  border: 2px solid var(--color-border-light);
  border-radius: var(--radius-md);
  background: var(--color-surface-2);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semi);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.payment-btn.active {
  border-color: var(--color-accent);
  background: rgba(233, 69, 96, 0.1);
}

.payment-icon {
  font-size: var(--font-size-xl);
}

.modal-actions {
  display: flex;
  gap: var(--gap-md);
  justify-content: flex-end;
}

.modal-actions .btn-lg {
  flex: 1;
}
</style>
