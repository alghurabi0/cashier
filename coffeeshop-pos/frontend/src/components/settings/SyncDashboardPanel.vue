<script setup lang="ts">
import { onMounted, onUnmounted, computed, ref } from 'vue'
import { useSyncDashboard } from '../../composables/useSyncDashboard'
import { formatPrice } from '../../types'

const {
  syncStatus,
  failedOrders,
  isLoading,
  actionMessage,
  initBindings,
  refresh,
  triggerSync,
  triggerFullSync,
  retryFailed,
  resetSyncState,
  startPolling,
  stopPolling,
} = useSyncDashboard()

const showResetConfirm = ref(false)
const showFailedDetails = ref(false)

// ── Computed helpers ──

const connectionClass = computed(() => {
  if (!syncStatus.value) return 'unknown'
  if (syncStatus.value.is_syncing) return 'syncing'
  return syncStatus.value.is_connected ? 'online' : 'offline'
})

const connectionLabel = computed(() => {
  if (!syncStatus.value) return '...'
  if (syncStatus.value.is_syncing) return '⏳ جاري المزامنة...'
  return syncStatus.value.is_connected ? '🟢 متصل' : '🔴 غير متصل'
})

const tableSyncEntries = computed(() => {
  if (!syncStatus.value?.table_sync_times) return []
  const labels: Record<string, string> = {
    categories: 'الأصناف',
    menu_items: 'المنتجات',
    inventory_items: 'المخزون',
    recipe_ingredients: 'الوصفات',
    all: 'الكل',
  }
  return Object.entries(labels).map(([key, label]) => ({
    key,
    label,
    time: syncStatus.value?.table_sync_times?.[key] || '',
    hasData: !!syncStatus.value?.table_sync_times?.[key],
  }))
})

const reversedLogs = computed(() => {
  if (!syncStatus.value?.recent_logs) return []
  return [...syncStatus.value.recent_logs].reverse()
})

function formatTime(iso: string): string {
  if (!iso) return '—'
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch {
    return iso
  }
}

function formatDateTime(iso: string): string {
  if (!iso) return '—'
  try {
    const d = new Date(iso)
    return d.toLocaleDateString('ar-IQ', { day: 'numeric', month: 'short' }) +
      ' ' + d.toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
  } catch {
    return iso
  }
}

const opLabels: Record<string, string> = {
  pull: 'سحب',
  push: 'دفع',
  health: 'فحص',
  retry: 'إعادة',
  reset: 'إعادة تعيين',
}

const entityLabels: Record<string, string> = {
  categories: 'الأصناف',
  menu_items: 'المنتجات',
  inventory_items: 'المخزون',
  recipe_ingredients: 'الوصفات',
  orders: 'الطلبات',
  connection: 'الاتصال',
  sync_meta: 'حالة المزامنة',
  all: 'الكل',
}

function onResetConfirm() {
  showResetConfirm.value = false
  resetSyncState()
}

onMounted(async () => {
  await initBindings()
  await refresh()
  startPolling(5000)
})

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <div class="sync-dashboard">
    <h2 class="section-title">📡 لوحة المزامنة</h2>

    <!-- A. Connection Status Banner -->
    <div class="conn-banner" :class="connectionClass">
      <span class="conn-label">{{ connectionLabel }}</span>
      <span v-if="syncStatus?.last_health_check_at" class="conn-time">
        آخر فحص: {{ formatTime(syncStatus.last_health_check_at) }}
      </span>
      <span v-if="!syncStatus?.is_connected && syncStatus?.last_connect_error" class="conn-error">
        {{ syncStatus.last_connect_error }}
      </span>
    </div>

    <!-- B. Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-value" :class="{ pulse: (syncStatus?.pending_orders || 0) > 0 }">
          {{ syncStatus?.pending_orders ?? '—' }}
        </div>
        <div class="stat-label">طلبات معلّقة</div>
      </div>
      <div class="stat-card" :class="{ danger: (syncStatus?.failed_orders || 0) > 0 }">
        <div class="stat-value">{{ syncStatus?.failed_orders ?? '—' }}</div>
        <div class="stat-label">طلبات فاشلة</div>
        <button
          v-if="(syncStatus?.failed_orders || 0) > 0"
          class="stat-action"
          :disabled="isLoading"
          @click="retryFailed"
        >
          ♻️ إعادة المحاولة
        </button>
      </div>
      <div class="stat-card" :class="{ warn: (syncStatus?.consecutive_errors || 0) > 0 }">
        <div class="stat-value">{{ syncStatus?.consecutive_errors ?? 0 }}</div>
        <div class="stat-label">أخطاء متتالية</div>
      </div>
    </div>

    <!-- C. Per-Table Sync Timestamps -->
    <div class="table-section">
      <h3 class="sub-title">حالة المزامنة لكل جدول</h3>
      <table class="sync-table">
        <thead>
          <tr>
            <th>الجدول</th>
            <th>آخر مزامنة</th>
            <th>الحالة</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in tableSyncEntries" :key="entry.key">
            <td>{{ entry.label }}</td>
            <td class="mono">{{ entry.hasData ? formatDateTime(entry.time) : '—' }}</td>
            <td>
              <span v-if="entry.hasData" class="status-icon ok">✅</span>
              <span v-else class="status-icon warn">⚠️</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- D. Controls -->
    <div class="controls-section">
      <h3 class="sub-title">التحكم</h3>
      <div class="controls-row">
        <button class="btn btn-primary" :disabled="isLoading" @click="triggerSync">
          🔄 مزامنة الآن
        </button>
        <button class="btn btn-secondary" :disabled="isLoading" @click="triggerFullSync">
          🔄 مزامنة كاملة
        </button>
        <button
          v-if="(syncStatus?.failed_orders || 0) > 0"
          class="btn btn-warn"
          :disabled="isLoading"
          @click="retryFailed"
        >
          ♻️ إعادة محاولة الفاشلة
        </button>
        <button class="btn btn-danger-outline" :disabled="isLoading" @click="showResetConfirm = true">
          🗑️ إعادة تعيين
        </button>
      </div>

      <!-- Action feedback -->
      <div v-if="actionMessage" class="action-msg" :class="{ success: actionMessage.startsWith('✓'), error: actionMessage.startsWith('✕') }">
        {{ actionMessage }}
      </div>

      <!-- Reset confirmation -->
      <div v-if="showResetConfirm" class="confirm-box">
        <p>⚠️ سيؤدي هذا إلى حذف جميع بيانات المزامنة المحلية. الدورة القادمة ستكون مزامنة كاملة.</p>
        <div class="confirm-actions">
          <button class="btn btn-danger" @click="onResetConfirm">تأكيد</button>
          <button class="btn btn-secondary" @click="showResetConfirm = false">إلغاء</button>
        </div>
      </div>
    </div>

    <!-- E. Sync Activity Log -->
    <div class="log-section">
      <h3 class="sub-title">سجل النشاط <span class="log-count">({{ reversedLogs.length }})</span></h3>
      <div class="log-scroll">
        <table class="log-table">
          <thead>
            <tr>
              <th>الوقت</th>
              <th>العملية</th>
              <th>الكيان</th>
              <th>الحالة</th>
              <th>التفاصيل</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(log, i) in reversedLogs" :key="i" :class="'log-' + log.status">
              <td class="mono">{{ formatTime(log.time) }}</td>
              <td>{{ opLabels[log.operation] || log.operation }}</td>
              <td>{{ entityLabels[log.entity] || log.entity }}</td>
              <td>
                <span v-if="log.status === 'ok'" class="badge ok">✅</span>
                <span v-else-if="log.status === 'error'" class="badge error">❌</span>
                <span v-else class="badge skip">⏭️</span>
              </td>
              <td class="log-msg">{{ log.message }}</td>
            </tr>
            <tr v-if="reversedLogs.length === 0">
              <td colspan="5" class="empty-log">لا توجد سجلات بعد</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- F. Failed Orders Detail -->
    <div v-if="(syncStatus?.failed_orders || 0) > 0" class="failed-section">
      <h3 class="sub-title clickable" @click="showFailedDetails = !showFailedDetails">
        ❌ الطلبات الفاشلة ({{ syncStatus?.failed_orders }})
        <span class="chevron" :class="{ open: showFailedDetails }">▸</span>
      </h3>
      <div v-if="showFailedDetails" class="failed-list">
        <div v-for="order in failedOrders" :key="order.id" class="failed-card">
          <div class="failed-header">
            <span class="failed-id">{{ order.order_number || order.id.substring(0, 8) }}</span>
            <span class="failed-total">{{ formatPrice(order.total) }} د.ع</span>
            <span class="failed-retries">{{ order.retry_count }}/10 محاولات</span>
          </div>
          <div class="failed-time">
            {{ formatDateTime(order.created_at) }}
            <span v-if="order.last_retry_at"> • آخر محاولة: {{ formatDateTime(order.last_retry_at) }}</span>
          </div>
          <div class="failed-error">{{ order.sync_error }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sync-dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--gap-lg);
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  margin-bottom: var(--gap-xs);
}

.sub-title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semi);
  margin-bottom: var(--gap-sm);
  color: var(--color-text-muted);
}

/* Connection Banner */
.conn-banner {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-md) var(--gap-lg);
  border-radius: var(--radius-md);
  font-weight: var(--font-weight-semi);
  transition: background 0.3s;
}
.conn-banner.online { background: rgba(39, 174, 96, 0.1); color: #27ae60; }
.conn-banner.offline { background: rgba(231, 76, 60, 0.1); color: #e74c3c; }
.conn-banner.syncing { background: rgba(243, 156, 18, 0.1); color: #f39c12; }
.conn-banner.unknown { background: var(--color-surface-2); color: var(--color-text-muted); }
.conn-label { font-size: var(--font-size-md); }
.conn-time { font-size: var(--font-size-sm); opacity: 0.7; }
.conn-error { font-size: var(--font-size-sm); margin-inline-start: auto; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--gap-md);
}
.stat-card {
  background: var(--color-surface-2);
  border-radius: var(--radius-md);
  padding: var(--gap-md);
  text-align: center;
  border: 1px solid var(--color-border-light);
  transition: border-color 0.3s;
}
.stat-card.danger { border-color: #e74c3c; }
.stat-card.warn { border-color: #f39c12; }
.stat-value {
  font-size: 2rem;
  font-weight: var(--font-weight-bold);
  line-height: 1;
  margin-bottom: var(--gap-xs);
}
.stat-value.pulse { animation: pulseValue 1.5s ease-in-out infinite; }
@keyframes pulseValue {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}
.stat-action {
  margin-top: var(--gap-sm);
  padding: var(--gap-xs) var(--gap-sm);
  border: 1px solid #e74c3c;
  border-radius: var(--radius-sm);
  background: transparent;
  color: #e74c3c;
  font-size: var(--font-size-xs);
  cursor: pointer;
  font-family: var(--font-family);
}
.stat-action:hover { background: rgba(231, 76, 60, 0.1); }

/* Sync Table */
.sync-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}
.sync-table th {
  text-align: right;
  padding: var(--gap-sm);
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-muted);
  font-weight: var(--font-weight-semi);
}
.sync-table td {
  padding: var(--gap-sm);
  border-bottom: 1px solid var(--color-border-light);
}
.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: var(--font-size-xs); }
.status-icon { font-size: 0.9rem; }

/* Controls */
.controls-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--gap-sm);
}
.btn {
  padding: var(--gap-sm) var(--gap-md);
  border-radius: var(--radius-sm);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary { background: var(--color-accent); color: white; }
.btn-primary:hover:not(:disabled) { filter: brightness(1.1); }
.btn-secondary { background: var(--color-surface-2); color: var(--color-text); border-color: var(--color-border-light); }
.btn-secondary:hover:not(:disabled) { background: var(--color-surface-3, var(--color-surface-2)); }
.btn-warn { background: rgba(243, 156, 18, 0.15); color: #f39c12; border-color: #f39c12; }
.btn-warn:hover:not(:disabled) { background: rgba(243, 156, 18, 0.25); }
.btn-danger-outline { background: transparent; color: #e74c3c; border-color: #e74c3c; }
.btn-danger-outline:hover:not(:disabled) { background: rgba(231, 76, 60, 0.1); }
.btn-danger { background: #e74c3c; color: white; }

.action-msg {
  padding: var(--gap-sm) var(--gap-md);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  margin-top: var(--gap-sm);
}
.action-msg.success { background: rgba(39, 174, 96, 0.12); color: var(--color-success); }
.action-msg.error { background: rgba(231, 76, 60, 0.12); color: var(--color-danger); }

.confirm-box {
  margin-top: var(--gap-sm);
  padding: var(--gap-md);
  background: rgba(231, 76, 60, 0.08);
  border: 1px solid rgba(231, 76, 60, 0.3);
  border-radius: var(--radius-sm);
}
.confirm-box p { margin: 0 0 var(--gap-sm); font-size: var(--font-size-sm); color: #e74c3c; }
.confirm-actions { display: flex; gap: var(--gap-sm); }

/* Activity Log */
.log-scroll {
  max-height: 280px;
  overflow-y: auto;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-sm);
}
.log-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-xs);
}
.log-table th {
  text-align: right;
  padding: var(--gap-xs) var(--gap-sm);
  background: var(--color-surface-2);
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-muted);
  font-weight: var(--font-weight-semi);
  position: sticky;
  top: 0;
}
.log-table td {
  padding: var(--gap-xs) var(--gap-sm);
  border-bottom: 1px solid var(--color-border-light);
  vertical-align: top;
}
.log-table tr.log-error td { color: #e74c3c; }
.log-table tr.log-skipped td { color: #f39c12; }
.log-msg {
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.badge {
  font-size: 0.75rem;
}
.empty-log { text-align: center; color: var(--color-text-muted); padding: var(--gap-lg) !important; }
.log-count { font-weight: normal; font-size: var(--font-size-xs); }

/* Failed Orders */
.failed-section {
  border: 1px solid rgba(231, 76, 60, 0.3);
  border-radius: var(--radius-md);
  padding: var(--gap-md);
  background: rgba(231, 76, 60, 0.04);
}
.clickable { cursor: pointer; user-select: none; }
.chevron {
  display: inline-block;
  transition: transform 0.2s;
  margin-inline-start: var(--gap-xs);
}
.chevron.open { transform: rotate(90deg); }
.failed-list {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
  margin-top: var(--gap-sm);
}
.failed-card {
  background: var(--color-surface-2);
  border-radius: var(--radius-sm);
  padding: var(--gap-sm) var(--gap-md);
  border-right: 3px solid #e74c3c;
}
.failed-header {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  font-weight: var(--font-weight-semi);
}
.failed-id { font-family: 'SF Mono', monospace; }
.failed-total { color: var(--color-accent); }
.failed-retries { font-size: var(--font-size-xs); color: #e74c3c; margin-inline-start: auto; }
.failed-time { font-size: var(--font-size-xs); color: var(--color-text-muted); margin-top: 2px; }
.failed-error {
  font-size: var(--font-size-xs);
  color: #e74c3c;
  margin-top: var(--gap-xs);
  background: rgba(231, 76, 60, 0.08);
  padding: var(--gap-xs) var(--gap-sm);
  border-radius: var(--radius-sm);
  font-family: 'SF Mono', monospace;
  word-break: break-all;
}
</style>
