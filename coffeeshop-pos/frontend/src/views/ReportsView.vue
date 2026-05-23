<script setup lang="ts">
import { onMounted } from 'vue'
import { useReports } from '../composables/useReports'
import { formatPrice } from '../types'

const { report, isLoading, dateFrom, dateTo, initBindings, loadReport } = useReports()

onMounted(async () => {
  await initBindings()
  await loadReport()
})

function onDateChange() {
  loadReport()
}

function sourceLabel(source: string): string {
  const labels: Record<string, string> = {
    cashier: '📋 كاشير',
    web_menu: '🌐 ويب',
  }
  return labels[source] || source
}
</script>

<template>
  <div class="reports-view">
    <header class="view-header">
      <h1 class="view-title">📊 التقارير</h1>
      <div class="date-range">
        <label>من <input type="date" v-model="dateFrom" @change="onDateChange" /></label>
        <label>إلى <input type="date" v-model="dateTo" @change="onDateChange" /></label>
      </div>
    </header>

    <div v-if="isLoading" class="loading">جاري التحميل...</div>

    <div v-else-if="report" class="report-content">
      <!-- Summary cards -->
      <div class="summary-cards">
        <div class="summary-card">
          <span class="card-icon">💰</span>
          <span class="card-value accent">{{ formatPrice(report.total_sales) }}</span>
          <span class="card-label">إجمالي المبيعات</span>
        </div>
        <div class="summary-card">
          <span class="card-icon">📦</span>
          <span class="card-value">{{ formatPrice(report.total_recipe_cost) }}</span>
          <span class="card-label">تكلفة الوصفات</span>
        </div>
        <div class="summary-card">
          <span class="card-icon">📈</span>
          <span class="card-value" :class="report.gross_profit >= 0 ? 'success' : 'danger'">
            {{ formatPrice(report.gross_profit) }}
          </span>
          <span class="card-label">صافي الربح</span>
        </div>
        <div class="summary-card">
          <span class="card-icon">%</span>
          <span class="card-value" :class="report.profit_margin >= 0 ? 'success' : 'danger'">
            {{ report.profit_margin.toFixed(1) }}%
          </span>
          <span class="card-label">هامش الربح</span>
        </div>
      </div>

      <!-- Order stats -->
      <div class="order-stats">
        <div class="stat-item">
          <span class="stat-num">{{ report.order_count }}</span>
          <span class="stat-text">طلبات</span>
        </div>
        <div class="stat-item" v-if="report.voided_count > 0">
          <span class="stat-num danger">{{ report.voided_count }}</span>
          <span class="stat-text">ملغية</span>
        </div>
      </div>

      <div class="report-sections">
        <!-- Sales by source -->
        <div class="report-section" v-if="Object.keys(report.sales_by_source).length > 0">
          <h3 class="section-title">المبيعات حسب المصدر</h3>
          <div class="source-bars">
            <div v-for="(amount, source) in report.sales_by_source" :key="source" class="source-bar">
              <span class="source-name">{{ sourceLabel(source as string) }}</span>
              <div class="bar-track">
                <div
                  class="bar-fill"
                  :style="{ width: (report.total_sales > 0 ? (amount / report.total_sales * 100) : 0) + '%' }"
                />
              </div>
              <span class="source-amount">{{ formatPrice(amount) }}</span>
            </div>
          </div>
        </div>

        <!-- Top items -->
        <div class="report-section" v-if="report.top_items.length > 0">
          <h3 class="section-title">الأكثر مبيعاً</h3>
          <div class="top-items">
            <div v-for="(item, idx) in report.top_items" :key="idx" class="top-item">
              <span class="top-rank">{{ idx + 1 }}</span>
              <span class="top-name">{{ item.name_ar }}</span>
              <span class="top-qty">×{{ item.total_qty }}</span>
              <span class="top-revenue">{{ formatPrice(item.total_revenue) }}</span>
            </div>
          </div>
        </div>

        <!-- Daily breakdown -->
        <div class="report-section" v-if="report.daily_breakdown.length > 1">
          <h3 class="section-title">التفصيل اليومي</h3>
          <div class="daily-table">
            <div class="daily-header">
              <span>التاريخ</span>
              <span>الطلبات</span>
              <span>المبيعات</span>
            </div>
            <div v-for="day in report.daily_breakdown" :key="day.date" class="daily-row">
              <span>{{ day.date }}</span>
              <span>{{ day.order_count }}</span>
              <span class="accent">{{ formatPrice(day.total_sales) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty">لا توجد بيانات</div>
  </div>
</template>

<style scoped>
.reports-view {
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

.report-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-xl);
  display: flex;
  flex-direction: column;
  gap: var(--gap-xl);
}

.summary-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--gap-md);
}

.summary-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--gap-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-sm);
  text-align: center;
}

.card-icon {
  font-size: 1.5rem;
}

.card-value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-extra);
  font-variant-numeric: tabular-nums;
}

.card-value.accent { color: var(--color-accent); }
.card-value.success { color: var(--color-success); }
.card-value.danger { color: var(--color-danger); }

.card-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.order-stats {
  display: flex;
  gap: var(--gap-lg);
}

.stat-item {
  display: flex;
  align-items: baseline;
  gap: var(--gap-xs);
}

.stat-num {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-extra);
}

.stat-num.danger { color: var(--color-danger); }

.stat-text {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.report-sections {
  display: flex;
  flex-direction: column;
  gap: var(--gap-xl);
}

.section-title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-bold);
  margin-bottom: var(--gap-md);
}

/* Sales by source */
.source-bars {
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.source-bar {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
}

.source-name {
  min-width: 80px;
  font-size: var(--font-size-sm);
}

.bar-track {
  flex: 1;
  height: 20px;
  background: var(--color-surface);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--color-accent), rgba(233, 69, 96, 0.7));
  border-radius: var(--radius-full);
  transition: width var(--transition-base);
  min-width: 4px;
}

.source-amount {
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  min-width: 80px;
  text-align: left;
}

/* Top items */
.top-items {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.top-item {
  display: flex;
  align-items: baseline;
  gap: var(--gap-md);
  padding: var(--gap-sm) var(--gap-md);
  background: var(--color-surface);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
}

.top-rank {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-surface-2);
  border-radius: 50%;
  font-weight: var(--font-weight-extra);
  font-size: var(--font-size-xs);
  color: var(--color-accent);
}

.top-name {
  flex: 1;
}

.top-qty {
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.top-revenue {
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
}

/* Daily breakdown */
.daily-table {
  display: flex;
  flex-direction: column;
}

.daily-header, .daily-row {
  display: grid;
  grid-template-columns: 1fr 80px 100px;
  padding: var(--gap-sm) var(--gap-md);
  font-size: var(--font-size-sm);
}

.daily-header {
  color: var(--color-text-muted);
  font-weight: var(--font-weight-bold);
  border-bottom: 1px solid var(--color-border);
}

.daily-row {
  border-bottom: 1px solid var(--color-border);
}

.daily-row .accent {
  color: var(--color-accent);
  font-weight: var(--font-weight-bold);
}

.loading, .empty {
  padding: var(--gap-xl);
  text-align: center;
  color: var(--color-text-dim);
}
</style>
