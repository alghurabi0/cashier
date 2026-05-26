<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useKitchen } from '../composables/useKitchen'
import { formatPrice } from '../types'

const {
  acceptedOrders,
  orderCount,
  isLoading,
  initBindings,
  loadOrders,
  startPolling,
  stopPolling,
  completeOrder,
} = useKitchen()

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

function sourceLabel(source: string): string {
  return source === 'web_menu' ? '🌐 ويب' : '📋 كاشير'
}

function sourceClass(source: string): string {
  return source === 'web_menu' ? 'source-web' : 'source-cashier'
}

onMounted(async () => {
  await initBindings()
  await loadOrders()
  startPolling(2000)
})

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <div class="kitchen-view">
    <header class="kitchen-header">
      <div class="header-right">
        <span class="header-icon">🍳</span>
        <h1 class="header-title">المطبخ</h1>
        <span v-if="orderCount > 0" class="order-badge">{{ orderCount }}</span>
      </div>
      <span class="header-hint text-muted" v-if="orderCount === 0">لا توجد طلبات بانتظار التحضير</span>
    </header>

    <div class="kitchen-body">
      <div class="orders-grid">
        <div
          v-for="order in acceptedOrders"
          :key="order.id"
          class="kitchen-card"
          :class="sourceClass(order.source)"
        >
          <!-- Card Header -->
          <div class="card-top">
            <span class="card-number">#{{ order.order_number }}</span>
            <span class="card-source" :class="sourceClass(order.source)">{{ sourceLabel(order.source) }}</span>
          </div>

          <!-- Table + Time -->
          <div class="card-meta">
            <span v-if="order.table_number" class="meta-table">🪑 طاولة {{ order.table_number }}</span>
            <span class="meta-time">⏱ {{ timeAgo(order.created_at) }}</span>
          </div>

          <!-- Items -->
          <div class="card-items">
            <div v-for="item in order.items" :key="item.id" class="item-row">
              <span class="item-qty">×{{ item.quantity }}</span>
              <span class="item-name">{{ item.name_ar_snapshot }}</span>
            </div>
          </div>

          <!-- Footer -->
          <div class="card-footer">
            <span class="card-total">{{ formatPrice(order.total) }} <small>د.ع</small></span>
            <button
              class="btn-ready"
              :disabled="isLoading"
              @click="completeOrder(order.id, order.source)"
            >
              ✓ جاهز
            </button>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="acceptedOrders.length === 0 && !isLoading" class="empty-state">
        <span class="empty-icon">👨‍🍳</span>
        <span class="empty-text">لا توجد طلبات بانتظار التحضير</span>
        <span class="empty-hint text-muted">ستظهر الطلبات الجديدة هنا تلقائياً</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.kitchen-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background: var(--color-bg);
}

.kitchen-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--gap-lg) var(--gap-xl);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
}

.header-icon {
  font-size: 1.8rem;
}

.header-title {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-extra);
}

.order-badge {
  background: var(--color-accent);
  color: white;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-extra);
  min-width: 28px;
  height: 28px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
  animation: badgePop 0.3s ease;
}

@keyframes badgePop {
  0% { transform: scale(0); }
  70% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

.header-hint {
  font-size: var(--font-size-sm);
}

.kitchen-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-lg);
}

/* Responsive grid — optimized for tablets */
.orders-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--gap-lg);
}

.kitchen-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl, 16px);
  padding: var(--gap-lg);
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
  transition: all var(--transition-fast);
  animation: cardSlideIn 0.3s ease;
}

@keyframes cardSlideIn {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.kitchen-card.source-web {
  border-right: 4px solid var(--color-info, #3498db);
}

.kitchen-card.source-cashier {
  border-right: 4px solid var(--color-accent);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-number {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-extra);
  font-variant-numeric: tabular-nums;
}

.card-source {
  padding: 4px 12px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
}

.card-source.source-web {
  background: rgba(52, 152, 219, 0.12);
  color: var(--color-info, #3498db);
}

.card-source.source-cashier {
  background: rgba(233, 69, 96, 0.12);
  color: var(--color-accent);
}

.card-meta {
  display: flex;
  gap: var(--gap-md);
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.meta-table {
  background: var(--color-surface-2);
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-weight: var(--font-weight-semi);
}

.card-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: var(--gap-md) 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.item-row {
  display: flex;
  align-items: baseline;
  gap: var(--gap-sm);
  font-size: var(--font-size-lg);
}

.item-qty {
  color: var(--color-accent);
  font-weight: var(--font-weight-extra);
  min-width: 36px;
  font-size: var(--font-size-xl);
}

.item-name {
  flex: 1;
  font-weight: var(--font-weight-semi);
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-total {
  font-weight: var(--font-weight-extra);
  font-size: var(--font-size-lg);
  font-variant-numeric: tabular-nums;
}

.card-total small {
  font-size: var(--font-size-xs);
  opacity: 0.7;
}

.btn-ready {
  padding: var(--gap-md) var(--gap-xl);
  border: none;
  border-radius: var(--radius-lg);
  background: var(--color-success);
  color: white;
  font-family: var(--font-family);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-extra);
  cursor: pointer;
  transition: all var(--transition-fast);
  min-width: 120px;
  text-align: center;
}

.btn-ready:hover {
  filter: brightness(1.1);
  transform: scale(1.02);
}

.btn-ready:active {
  transform: scale(0.97);
}

.btn-ready:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  gap: var(--gap-md);
  text-align: center;
}

.empty-icon {
  font-size: 5rem;
  opacity: 0.3;
}

.empty-text {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-dim);
}

.empty-hint {
  font-size: var(--font-size-md);
}
</style>
