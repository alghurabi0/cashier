<script setup lang="ts">
import { onMounted } from 'vue'
import { useOrderHistory } from '../composables/useOrderHistory'
import { useAuth } from '../composables/useAuth'
import { formatPrice } from '../types'

const {
  orders, selectedOrder, isLoading, dateFrom, dateTo, stats,
  initBindings, loadOrders, voidOrder,
} = useOrderHistory()

const { isAdmin } = useAuth()

onMounted(async () => {
  await initBindings()
  await loadOrders()
})

function onDateChange() {
  loadOrders()
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    completed: 'مكتمل',
    voided: 'ملغي',
    pending: 'معلق',
    accepted: 'مقبول',
  }
  return labels[status] || status
}

function sourceLabel(source: string): string {
  return source === 'web_menu' ? '🌐 ويب' : '📋 كاشير'
}
</script>

<template>
  <div class="history-view">
    <header class="view-header">
      <h1 class="view-title">📜 سجل الطلبات</h1>
      <div class="date-range">
        <label>من <input type="date" v-model="dateFrom" @change="onDateChange" /></label>
        <label>إلى <input type="date" v-model="dateTo" @change="onDateChange" /></label>
      </div>
    </header>

    <!-- Stats bar -->
    <div class="stats-bar">
      <div class="stat-card">
        <span class="stat-value">{{ stats.totalCount }}</span>
        <span class="stat-label">طلبات</span>
      </div>
      <div class="stat-card accent">
        <span class="stat-value">{{ formatPrice(stats.totalRevenue) }}</span>
        <span class="stat-label">المبيعات</span>
      </div>
      <div class="stat-card" v-if="stats.voidedCount > 0">
        <span class="stat-value danger">{{ stats.voidedCount }}</span>
        <span class="stat-label">ملغية</span>
      </div>
    </div>

    <div class="history-content">
      <!-- Order list -->
      <div class="order-list">
        <div v-if="isLoading" class="loading">جاري التحميل...</div>
        <div v-else-if="orders.length === 0" class="empty">لا توجد طلبات في هذه الفترة</div>
        <button
          v-for="order in orders"
          :key="order.id"
          class="order-row"
          :class="{ active: selectedOrder?.id === order.id, voided: order.status === 'voided' }"
          @click="selectedOrder = order"
        >
          <span class="order-num">{{ order.order_number }}</span>
          <span class="order-source">{{ sourceLabel(order.source) }}</span>
          <span class="order-total">{{ formatPrice(order.total) }}</span>
          <span class="order-status" :class="order.status">{{ statusLabel(order.status) }}</span>
        </button>
      </div>

      <!-- Order detail panel -->
      <div class="order-detail" v-if="selectedOrder">
        <div class="detail-header">
          <h2>{{ selectedOrder.order_number }}</h2>
          <span class="detail-status" :class="selectedOrder.status">{{ statusLabel(selectedOrder.status) }}</span>
        </div>

        <div class="detail-meta">
          <div class="meta-row">
            <span class="meta-label">المصدر</span>
            <span>{{ sourceLabel(selectedOrder.source) }}</span>
          </div>
          <div class="meta-row" v-if="selectedOrder.table_number">
            <span class="meta-label">الطاولة</span>
            <span>🪑 {{ selectedOrder.table_number }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">الوقت</span>
            <span>{{ selectedOrder.created_at }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">الدفع</span>
            <span>{{ selectedOrder.payment_method }}</span>
          </div>
        </div>

        <div class="detail-items">
          <div v-for="item in selectedOrder.items" :key="item.id" class="detail-line">
            <span class="line-qty">×{{ item.quantity }}</span>
            <span class="line-name">{{ item.name_ar_snapshot }}</span>
            <span class="line-total">{{ formatPrice(item.line_total) }}</span>
          </div>
        </div>

        <div class="detail-total">
          <span>المجموع</span>
          <span class="total-value">{{ formatPrice(selectedOrder.total) }} <small>د.ع</small></span>
        </div>

        <button
          v-if="isAdmin() && selectedOrder.status === 'completed'"
          class="btn btn-void"
          @click="voidOrder(selectedOrder.id)"
        >
          🗑️ إلغاء الطلب
        </button>
      </div>

      <div class="order-detail empty-detail" v-else>
        <span class="text-muted">اختر طلباً لعرض التفاصيل</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.history-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--gap-lg) var(--gap-xl);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.view-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
}

.date-range {
  display: flex;
  gap: var(--gap-md);
  align-items: center;
}

.date-range label {
  display: flex;
  align-items: center;
  gap: var(--gap-xs);
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.date-range input {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  padding: var(--gap-xs) var(--gap-sm);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
}

.stats-bar {
  display: flex;
  gap: var(--gap-md);
  padding: var(--gap-md) var(--gap-xl);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.stat-card {
  background: var(--color-surface);
  padding: var(--gap-sm) var(--gap-lg);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.stat-card.accent .stat-value {
  color: var(--color-accent);
}

.stat-value {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-extra);
  font-variant-numeric: tabular-nums;
}

.stat-value.danger {
  color: var(--color-danger);
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.history-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.order-list {
  width: 420px;
  min-width: 420px;
  border-left: 1px solid var(--color-border);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.order-row {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-md) var(--gap-lg);
  border: none;
  border-bottom: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text);
  cursor: pointer;
  font-family: var(--font-family);
  text-align: right;
  transition: background var(--transition-fast);
}

.order-row:hover {
  background: var(--color-surface);
}

.order-row.active {
  background: var(--color-surface-2);
  border-right: 3px solid var(--color-accent);
}

.order-row.voided {
  opacity: 0.55;
}

.order-num {
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  min-width: 50px;
}

.order-source {
  font-size: var(--font-size-xs);
}

.order-total {
  margin-right: auto;
  font-weight: var(--font-weight-semi);
  font-variant-numeric: tabular-nums;
}

.order-status {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.order-status.completed {
  background: rgba(92, 184, 92, 0.12);
  color: var(--color-success);
}

.order-status.voided {
  background: rgba(231, 76, 60, 0.12);
  color: var(--color-danger);
}

.order-detail {
  flex: 1;
  padding: var(--gap-xl);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--gap-lg);
}

.order-detail.empty-detail {
  align-items: center;
  justify-content: center;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-header h2 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-extra);
}

.detail-status {
  padding: var(--gap-xs) var(--gap-md);
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
}

.detail-status.completed {
  background: rgba(92, 184, 92, 0.12);
  color: var(--color-success);
}

.detail-status.voided {
  background: rgba(231, 76, 60, 0.12);
  color: var(--color-danger);
}

.detail-meta {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
}

.meta-row {
  display: flex;
  justify-content: space-between;
  font-size: var(--font-size-sm);
}

.meta-label {
  color: var(--color-text-muted);
}

.detail-items {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: var(--gap-md) 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.detail-line {
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
  font-variant-numeric: tabular-nums;
  color: var(--color-text-muted);
}

.detail-total {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
}

.total-value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-extra);
  color: var(--color-accent);
}

.total-value small {
  font-size: var(--font-size-xs);
  opacity: 0.7;
}

.btn-void {
  align-self: flex-start;
  padding: var(--gap-sm) var(--gap-lg);
  border: 1px solid var(--color-danger);
  border-radius: var(--radius-md);
  background: rgba(231, 76, 60, 0.08);
  color: var(--color-danger);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-void:hover {
  background: rgba(231, 76, 60, 0.18);
}

.loading, .empty {
  padding: var(--gap-xl);
  text-align: center;
  color: var(--color-text-dim);
}
</style>
