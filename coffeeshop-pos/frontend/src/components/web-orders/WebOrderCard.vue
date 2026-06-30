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
  if (mins < 60) return `منذ ${mins} دقيقة`
  const hours = Math.floor(mins / 60)
  return `منذ ${hours} ساعة`
}
</script>

<template>
  <div class="order-card" :class="status">

    <div class="card-header">
      <div class="order-num">#{{ order.order_number }}</div>
      <div class="table-badge">🪑 طاولة {{ order.table_number }}</div>
      <div class="order-time">{{ timeAgo(order.created_at) }}</div>
    </div>

    <div class="card-items">
      <div v-for="item in order.items" :key="item.id" class="item-row">
        <span class="item-qty">×{{ item.quantity }}</span>
        <span class="item-name">{{ item.name_ar_snapshot }}</span>
        <span class="item-price">{{ formatPrice(item.line_total) }}</span>
      </div>
    </div>

    <div class="card-footer">
      <div class="total-wrap">
        <span class="total-label">المجموع</span>
        <span class="total-amount">{{ formatPrice(order.total) }} <small>د.ع</small></span>
      </div>

      <div class="actions" v-if="status === 'pending'">
        <button class="action-btn reject-btn" @click="emit('reject', order.id)">✗ رفض</button>
        <button class="action-btn accept-btn" @click="emit('accept', order.id)">✓ قبول</button>
      </div>

      <div class="actions" v-else-if="status === 'accepted'">
        <button class="action-btn complete-btn" @click="emit('complete', order.id)">✓ اكتمل</button>
      </div>
    </div>

  </div>
</template>

<style scoped>
.order-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 14px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: all 0.2s ease;
}

.order-card.pending {
  border-right: 3px solid var(--color-warning);
  animation: pulseGlow 2.5s infinite;
}

.order-card.accepted {
  border-right: 3px solid var(--color-success);
}

.order-card.completed {
  opacity: 0.55;
  border-right: 3px solid var(--color-text-dim);
}

@keyframes pulseGlow {
  0%, 100% { box-shadow: 0 0 0 0 rgba(243,156,18,0); }
  50% { box-shadow: 0 0 16px 2px rgba(243,156,18,0.12); }
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.order-num {
  font-size: 1.05rem;
  font-weight: 900;
  color: var(--color-accent);
  letter-spacing: 1px;
}

.table-badge {
  background: var(--color-border-light);
  border: 1px solid var(--color-border-light);
  padding: 3px 10px;
  border-radius: 50px;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--color-accent);
}

.order-time {
  margin-right: auto;
  font-size: 0.68rem;
  color: var(--color-text-dim);
}

.card-items {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.item-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.82rem;
}

.item-qty {
  color: var(--color-accent);
  font-weight: 800;
  min-width: 28px;
}

.item-name {
  flex: 1;
  color: var(--color-text);
}

.item-price {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.total-wrap {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.total-label {
  font-size: 0.65rem;
  color: var(--color-text-dim);
}

.total-amount {
  font-size: 1.05rem;
  font-weight: 900;
  color: var(--color-text);
}

.total-amount small {
  font-size: 0.65rem;
  opacity: 0.6;
  margin-right: 2px;
}

.actions {
  display: flex;
  gap: 6px;
}

.action-btn {
  padding: 7px 16px;
  border: none;
  border-radius: 8px;
  font-family: 'Cairo', sans-serif;
  font-size: 0.8rem;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.2s;
}

.accept-btn {
  background: var(--color-success);
  color: #fff;
}
.accept-btn:hover { filter: brightness(1.15); }

.reject-btn {
  background: rgba(231,76,60,0.12);
  color: var(--color-danger);
  border: 1px solid rgba(231,76,60,0.25);
}
.reject-btn:hover { background: rgba(231,76,60,0.2); }

.complete-btn {
  background: linear-gradient(135deg, var(--color-accent), var(--color-accent-hover));
  color: #0d0d0d;
  box-shadow: 0 4px 14px var(--color-accent-glow);
}
.complete-btn:hover { filter: brightness(1.1); }
</style>