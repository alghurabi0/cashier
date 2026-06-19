<script setup lang="ts">
import type { OrderWithItems } from '../types'
import { formatPrice } from '../types'

defineProps<{
  acceptedOrders: OrderWithItems[]
  completedOrders: OrderWithItems[]
  kitchenModeEnabled: boolean
}>()

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr.replace(' ', 'T'))
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const mins = Math.floor(diffMs / 60000)
  if (mins < 1) return 'الآن'
  if (mins < 60) return `${mins} د`
  const hours = Math.floor(mins / 60)
  return `${hours} س`
}
</script>

<template>
  <div class="order-queue-panel" v-if="acceptedOrders.length > 0 || completedOrders.length > 0">
    <!-- Accepted (in-progress) -->
    <div v-if="kitchenModeEnabled && acceptedOrders.length > 0" class="queue-section">
      <div class="section-header preparing">
        <span class="section-icon">⏳</span>
        <span class="section-title">قيد التحضير</span>
        <span class="section-count">{{ acceptedOrders.length }}</span>
      </div>
      <div class="queue-items">
        <div v-for="order in acceptedOrders" :key="order.id" class="queue-card preparing">
          <span class="queue-number">{{ order.order_number }}</span>
          <span class="queue-time">{{ timeAgo(order.created_at) }}</span>
          <span class="queue-total">{{ formatPrice(order.total) }}</span>
        </div>
      </div>
    </div>

    <!-- Recently completed -->
    <div v-if="completedOrders.length > 0" class="queue-section">
      <div class="section-header completed">
        <span class="section-icon">✅</span>
        <span class="section-title">مكتملة</span>
      </div>
      <div class="queue-items">
        <div v-for="order in completedOrders.slice(0, 5)" :key="order.id" class="queue-card completed">
          <span class="queue-number">{{ order.order_number }}</span>
          <span class="queue-time">{{ timeAgo(order.created_at) }}</span>
          <span class="queue-total">{{ formatPrice(order.total) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.order-queue-panel {
  border-top: 1px solid var(--color-border);
  padding: var(--gap-sm) var(--gap-md);
  display: flex;
  gap: var(--gap-lg);
  flex-shrink: 0;
  overflow-x: auto;
  background: var(--color-surface);
}

.queue-section {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  min-width: 0;
}

.section-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  min-width: 56px;
  flex-shrink: 0;
}

.section-icon {
  font-size: 1rem;
}

.section-title {
  font-size: 0.6rem;
  font-weight: var(--font-weight-bold);
  color: var(--color-text-muted);
  white-space: nowrap;
}

.section-count {
  font-size: 0.6rem;
  font-weight: var(--font-weight-extra);
  background: var(--color-accent);
  color: white;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

.queue-items {
  display: flex;
  gap: var(--gap-xs);
  overflow-x: auto;
}

.queue-card {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-xs) var(--gap-sm);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  white-space: nowrap;
  flex-shrink: 0;
}

.queue-card.preparing {
  background: rgba(243, 156, 18, 0.1);
  border: 1px solid rgba(243, 156, 18, 0.25);
}

.queue-card.completed {
  background: rgba(92, 184, 92, 0.08);
  border: 1px solid rgba(92, 184, 92, 0.2);
  opacity: 0.7;
}

.queue-number {
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
}

.queue-time {
  color: var(--color-text-dim);
}

.queue-total {
  font-weight: var(--font-weight-semi);
  font-variant-numeric: tabular-nums;
}
</style>
