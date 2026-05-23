<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
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
  <div class="web-orders-view">
    <header class="view-header">
      <h1 class="view-title">🌐 طلبات القائمة الإلكترونية</h1>
      <button
        class="btn btn-ghost sound-toggle"
        :class="{ muted: !soundEnabled }"
        @click="toggleSound"
        :title="soundEnabled ? 'كتم الصوت' : 'تشغيل الصوت'"
      >
        {{ soundEnabled ? '🔊' : '🔇' }}
      </button>
    </header>

    <div class="orders-columns">
      <!-- Pending -->
      <div class="order-column pending-col">
        <div class="column-header">
          <span class="column-icon">⏳</span>
          <span class="column-title">بانتظار</span>
          <span class="column-count" v-if="pendingOrders.length">{{ pendingOrders.length }}</span>
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
            لا توجد طلبات
          </div>
        </div>
      </div>

      <!-- Accepted -->
      <div class="order-column accepted-col">
        <div class="column-header">
          <span class="column-icon">✅</span>
          <span class="column-title">مقبولة</span>
          <span class="column-count" v-if="acceptedOrders.length">{{ acceptedOrders.length }}</span>
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
            لا توجد طلبات مقبولة
          </div>
        </div>
      </div>

      <!-- Completed -->
      <div class="order-column completed-col">
        <div class="column-header">
          <span class="column-icon">📜</span>
          <span class="column-title">المكتملة</span>
          <span class="column-count" v-if="completedOrders.length">{{ completedOrders.length }}</span>
        </div>
        <div class="column-body">
          <WebOrderCard
            v-for="order in completedOrders"
            :key="order.id"
            :order="order"
            status="completed"
          />
          <div v-if="completedOrders.length === 0" class="empty-col">
            لا توجد طلبات مكتملة
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

.sound-toggle {
  font-size: 1.3rem;
  border-radius: var(--radius-full);
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.sound-toggle.muted {
  opacity: 0.5;
}

.orders-columns {
  display: flex;
  flex: 1;
  gap: var(--gap-md);
  padding: var(--gap-lg);
  overflow: hidden;
}

.order-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  overflow: hidden;
}

.column-header {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
  padding: var(--gap-md) var(--gap-lg);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.column-icon {
  font-size: var(--font-size-lg);
}

.column-title {
  font-weight: var(--font-weight-bold);
  font-size: var(--font-size-md);
}

.column-count {
  background: var(--color-accent);
  color: white;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  margin-right: auto;
}

.column-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-md);
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.empty-col {
  text-align: center;
  padding: var(--gap-xl);
  color: var(--color-text-dim);
  font-size: var(--font-size-sm);
}
</style>
