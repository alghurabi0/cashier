<script setup lang="ts">
import type { OrderWithItems } from '../../composables/useWebOrders'
import { formatPrice } from '../../types'

const props = defineProps<{
  order: OrderWithItems
  status: 'pending' | 'accepted' | 'completed' | 'rejected'
}>()

const emit = defineEmits<{
  accept: [orderID: string]
  reject: [orderID: string]
  complete: [orderID: string]
}>()

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const mins = Math.floor(diffMs / 60000)
  if (mins < 1) return 'الآن'
  if (mins < 60) return `${mins} دقيقة`
  const hours = Math.floor(mins / 60)
  return `${hours} ساعة`
}
</script>

<template>
  <div class="web-order-card" :class="status">
    <div class="order-header">
      <span class="order-number">#{{ order.order_number }}</span>
      <span class="table-badge">🪑 طاولة {{ order.table_number }}</span>
    </div>

    <div class="order-time">{{ timeAgo(order.created_at) }}</div>

    <div class="order-items">
      <div v-for="item in order.items" :key="item.id" class="order-line">
        <span class="line-qty">×{{ item.quantity }}</span>
        <span class="line-name">{{ item.name_ar_snapshot }}</span>
        <span class="line-total">{{ formatPrice(item.line_total) }}</span>
      </div>
    </div>

    <div class="order-footer">
      <span class="order-total">{{ formatPrice(order.total) }} <small>د.ع</small></span>

      <div class="order-actions" v-if="status === 'pending'">
        <button class="btn btn-accept" @click="emit('accept', order.id)" title="قبول">✓ قبول</button>
        <button class="btn btn-reject" @click="emit('reject', order.id)" title="رفض">✗ رفض</button>
      </div>

      <div class="order-actions" v-else-if="status === 'accepted'">
        <button class="btn btn-complete" @click="emit('complete', order.id)" title="اكتمل">✓ اكتمل</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.web-order-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--gap-md);
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
  transition: all var(--transition-fast);
}

.web-order-card.pending {
  border-right: 3px solid var(--color-warning);
  animation: pulseGlow 2s infinite;
}

.web-order-card.accepted {
  border-right: 3px solid var(--color-success);
}

.web-order-card.completed {
  opacity: 0.7;
  border-right: 3px solid var(--color-text-dim);
}

@keyframes pulseGlow {
  0%, 100% { box-shadow: 0 0 0 0 rgba(243, 156, 18, 0); }
  50% { box-shadow: 0 0 12px 2px rgba(243, 156, 18, 0.15); }
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-number {
  font-weight: var(--font-weight-extra);
  font-size: var(--font-size-lg);
  font-variant-numeric: tabular-nums;
}

.table-badge {
  background: var(--color-surface-2);
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
}

.order-time {
  font-size: var(--font-size-xs);
  color: var(--color-text-dim);
}

.order-items {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--gap-sm) 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.order-line {
  display: flex;
  align-items: baseline;
  gap: var(--gap-sm);
  font-size: var(--font-size-sm);
}

.line-qty {
  color: var(--color-accent);
  font-weight: var(--font-weight-bold);
  min-width: 28px;
}

.line-name {
  flex: 1;
}

.line-total {
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-total {
  font-weight: var(--font-weight-extra);
  font-size: var(--font-size-lg);
}

.order-total small {
  font-size: var(--font-size-xs);
  opacity: 0.7;
}

.order-actions {
  display: flex;
  gap: var(--gap-sm);
}

.btn {
  padding: var(--gap-xs) var(--gap-md);
  border: none;
  border-radius: var(--radius-sm);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-accept {
  background: var(--color-success);
  color: white;
}

.btn-accept:hover {
  filter: brightness(1.1);
}

.btn-reject {
  background: var(--color-surface-2);
  color: var(--color-danger);
}

.btn-reject:hover {
  background: rgba(231, 76, 60, 0.15);
}

.btn-complete {
  background: var(--color-accent);
  color: white;
}

.btn-complete:hover {
  filter: brightness(1.1);
}
</style>
