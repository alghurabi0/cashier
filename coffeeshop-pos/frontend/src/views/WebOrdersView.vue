<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useWebOrders } from '../composables/useWebOrders'
import WebOrderCard from '../components/web-orders/WebOrderCard.vue'

const {
  pendingOrders,
  acceptedOrders,
  completedOrders,
  soundEnabled,
  initBindings,
  loadOrders,
  startPolling,
  stopPolling,
  toggleSound,
  acceptOrder,
  rejectOrder,
  completeOrder,
} = useWebOrders()

const kitchenModeEnabled = ref(false)

onMounted(async () => {
  await initBindings()
  await loadOrders()
  startPolling(2000)
  // Load kitchen mode setting
  try {
    const configMod = await import('../../bindings/coffeeshop-pos/internal/service/configstoreservice')
    const kitchenVal = await configMod.Get('kitchen_mode_enabled')
    kitchenModeEnabled.value = kitchenVal === 'true'
  } catch { /* not available */ }
})

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <div class="web-orders-view">
    <header class="view-header">
      <div class="header-left">
        <div class="header-icon">🌐</div>
        <div>
          <h1 class="view-title">طلبات القائمة الإلكترونية</h1>
          <p class="view-sub">تحديث تلقائي كل ثانيتين</p>
        </div>
      </div>
      <div class="header-right">
        <div class="stats-chips">
          <div class="stat-chip pending-chip">
            <span class="chip-dot" />
            <span>بانتظار</span>
            <strong>{{ pendingOrders.length }}</strong>
          </div>
          <div class="stat-chip accepted-chip" v-if="kitchenModeEnabled">
            <span class="chip-dot" />
            <span>مقبولة</span>
            <strong>{{ acceptedOrders.length }}</strong>
          </div>
          <div class="stat-chip completed-chip">
            <span class="chip-dot" />
            <span>مكتملة</span>
            <strong>{{ completedOrders.length }}</strong>
          </div>
        </div>
        <button class="sound-btn" :class="{ muted: !soundEnabled }" @click="toggleSound">
          {{ soundEnabled ? '🔊' : '🔇' }}
        </button>
      </div>
    </header>

    <div class="orders-columns">
      <!-- Pending -->
      <div class="order-column pending-col">
        <div class="column-header">
          <div class="col-indicator pending-ind" />
          <span class="column-title">⏳ بانتظار</span>
          <span class="column-count pending-count" v-if="pendingOrders.length">{{ pendingOrders.length }}</span>
        </div>
        <div class="column-body">
          <WebOrderCard
            v-for="order in pendingOrders"
            :key="order.id"
            :order="order"
            status="pending"
            @accept="acceptOrder"
            @reject="rejectOrder"
          />
          <div v-if="pendingOrders.length === 0" class="empty-col">
            <span class="empty-icon">☕</span>
            <p>لا توجد طلبات جديدة</p>
          </div>
        </div>
      </div>

      <!-- Accepted (only show when kitchen mode is enabled) -->
      <div class="order-column accepted-col" v-if="kitchenModeEnabled">
        <div class="column-header">
          <div class="col-indicator accepted-ind" />
          <span class="column-title">✅ مقبولة</span>
          <span class="column-count accepted-count" v-if="acceptedOrders.length">{{ acceptedOrders.length }}</span>
        </div>
        <div class="column-body">
          <WebOrderCard
            v-for="order in acceptedOrders"
            :key="order.id"
            :order="order"
            status="accepted"
            @complete="completeOrder"
          />
          <div v-if="acceptedOrders.length === 0" class="empty-col">
            <span class="empty-icon">🔄</span>
            <p>لا توجد طلبات مقبولة</p>
          </div>
        </div>
      </div>

      <!-- Completed -->
      <div class="order-column completed-col">
        <div class="column-header">
          <div class="col-indicator completed-ind" />
          <span class="column-title">📜 المكتملة</span>
          <span class="column-count completed-count" v-if="completedOrders.length">{{ completedOrders.length }}</span>
        </div>
        <div class="column-body">
          <WebOrderCard
            v-for="order in completedOrders"
            :key="order.id"
            :order="order"
            status="completed"
          />
          <div v-if="completedOrders.length === 0" class="empty-col">
            <span class="empty-icon">✨</span>
            <p>لا توجد طلبات مكتملة</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.web-orders-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background: var(--color-bg);
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border-light);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-icon {
  width: 42px; height: 42px;
  border-radius: 12px;
  background: var(--color-accent-glow);
  border: 1px solid var(--color-border-light);
  display: flex; align-items: center; justify-content: center;
  font-size: 1.2rem;
}

.view-title {
  font-size: 1.1rem;
  font-weight: 800;
  color: var(--color-text);
  margin: 0;
}

.view-sub {
  font-size: 0.7rem;
  color: var(--color-text-dim);
  margin: 2px 0 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.stats-chips {
  display: flex;
  gap: 8px;
}

.stat-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 50px;
  font-size: 0.78rem;
  font-weight: 700;
  border: 1px solid transparent;
}

.chip-dot {
  width: 7px; height: 7px;
  border-radius: 50%;
}

.pending-chip { background: rgba(243,156,18,0.1); border-color: rgba(243,156,18,0.25); color: var(--color-warning); }
.pending-chip .chip-dot { background: var(--color-warning); }
.accepted-chip { background: rgba(39,174,96,0.1); border-color: rgba(39,174,96,0.25); color: var(--color-success); }
.accepted-chip .chip-dot { background: var(--color-success); }
.completed-chip { background: var(--color-accent-glow); border-color: var(--color-border-light); color: var(--color-accent); }
.completed-chip .chip-dot { background: var(--color-accent); }

.sound-btn {
  width: 38px; height: 38px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  font-size: 1.1rem;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.2s;
}
.sound-btn:hover { border-color: var(--color-border-light); }
.sound-btn.muted { opacity: 0.4; }

.orders-columns {
  display: flex;
  flex: 1;
  gap: 12px;
  padding: 14px;
  overflow: hidden;
}

.order-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--color-surface);
  border-radius: 16px;
  border: 1px solid var(--color-border);
  overflow: hidden;
}

.column-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.col-indicator {
  width: 4px; height: 20px;
  border-radius: 2px;
  flex-shrink: 0;
}

.pending-ind { background: var(--color-warning); }
.accepted-ind { background: var(--color-success); }
.completed-ind { background: var(--color-accent); }

.column-title {
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--color-text);
  flex: 1;
}

.column-count {
  font-size: 0.7rem;
  font-weight: 900;
  padding: 2px 9px;
  border-radius: 50px;
}

.pending-count { background: rgba(243,156,18,0.15); color: var(--color-warning); }
.accepted-count { background: rgba(39,174,96,0.15); color: var(--color-success); }
.completed-count { background: var(--color-accent-glow); color: var(--color-accent); }

.column-body {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  scrollbar-width: thin;
  scrollbar-color: var(--color-accent-glow) transparent;
}

.empty-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 20px;
  color: var(--color-text-dim);
  text-align: center;
}

.empty-icon { font-size: 2rem; opacity: 0.5; }
.empty-col p { font-size: 0.82rem; margin: 0; }
</style>