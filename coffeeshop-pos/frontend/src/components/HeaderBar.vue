<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

defineProps<{
  orderCount: number
  syncStatus: 'online' | 'offline' | 'syncing'
}>()

const currentTime = ref('')

function updateTime() {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('ar-IQ', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

let timer: ReturnType<typeof setInterval>

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<template>
  <header class="header-bar" style="--wails-draggable: drag">
    <div class="header-right">
      <div class="header-logo">
        <span class="logo-icon">☕</span>
        <span class="logo-text">المقهى</span>
      </div>
    </div>

    <div class="header-center">
      <span class="header-clock">{{ currentTime }}</span>
    </div>

    <div class="header-left" style="--wails-draggable: no-drag">
      <div class="header-stat">
        <span class="stat-label">الطلبات</span>
        <span class="badge">{{ orderCount }}</span>
      </div>
      <div class="sync-indicator" :class="syncStatus">
        <span class="sync-dot"></span>
        <span class="sync-label">{{ syncStatus === 'online' ? 'متصل' : syncStatus === 'syncing' ? 'مزامنة...' : 'غير متصل' }}</span>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--gap-sm) var(--gap-lg);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  height: 56px;
  flex-shrink: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
}

.header-logo {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
}

.logo-icon {
  font-size: var(--font-size-xl);
}

.logo-text {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-extra);
  background: linear-gradient(135deg, var(--color-accent), #ff8a65);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}

.header-clock {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semi);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--gap-lg);
}

.header-stat {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.sync-indicator {
  display: flex;
  align-items: center;
  gap: var(--gap-xs);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  background: var(--color-surface-2);
}

.sync-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  transition: background var(--transition-fast);
}

.sync-indicator.online .sync-dot {
  background: var(--color-success);
  box-shadow: 0 0 6px var(--color-success-glow);
}

.sync-indicator.offline .sync-dot {
  background: var(--color-danger);
}

.sync-indicator.syncing .sync-dot {
  background: var(--color-warning);
  animation: pulse 1.5s infinite;
}

.sync-label {
  color: var(--color-text-muted);
}
</style>
