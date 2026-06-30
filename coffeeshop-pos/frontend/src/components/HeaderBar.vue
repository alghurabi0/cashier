<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

defineProps<{
  orderCount: number
  syncStatus: 'online' | 'offline' | 'syncing'
  kitchenModeEnabled: boolean
}>()

const currentTime = ref('')
const currentDate = ref('')

function updateTime() {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
  currentDate.value = now.toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long' })
}

let timer: ReturnType<typeof setInterval>
onMounted(() => { updateTime(); timer = setInterval(updateTime, 1000) })
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <header class="header-bar" style="--wails-draggable: drag">
    <div class="header-right">
      <div class="logo">
        <span class="logo-icon">☕</span>
        <div class="logo-texts">
          <span class="logo-name">NJ COFFEE</span>
          <span class="logo-sub">نظام نقطة البيع</span>
        </div>
      </div>
    </div>

    <div class="header-center">
      <span class="clock">{{ currentTime }}</span>
      <span class="date">{{ currentDate }}</span>
    </div>

    <div class="header-left" style="--wails-draggable: no-drag">
      <div class="order-chip" v-if="kitchenModeEnabled && orderCount > 0">
        <span class="order-chip-dot"></span>
        <span>{{ orderCount }} طلب نشط</span>
      </div>
      <div class="sync-pill" :class="syncStatus">
        <span class="sync-dot"></span>
        <span>{{ syncStatus === 'online' ? 'متصل' : syncStatus === 'syncing' ? 'مزامنة' : 'غير متصل' }}</span>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--gap-lg);
  background: var(--color-cart-bg);
  border-bottom: 1px solid var(--color-border-light);
  height: 58px;
  flex-shrink: 0;
  position: relative;
}

/* Logo */
.logo {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo-icon {
  font-size: 1.5rem;
  filter: drop-shadow(0 0 6px var(--color-accent-glow));
}

.logo-texts {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.logo-name {
  font-size: 1rem;
  font-weight: 800;
  background: linear-gradient(135deg, var(--color-accent), var(--color-accent-hover));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: 0.08em;
}

.logo-sub {
  font-size: 0.62rem;
  color: var(--color-text-dim);
  font-weight: 600;
  letter-spacing: 0.03em;
}

/* Center clock */
.header-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
}

.clock {
  font-size: 1.1rem;
  font-weight: 800;
  color: var(--color-text);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.date {
  font-size: 0.65rem;
  color: var(--color-text-dim);
  font-weight: 600;
}

/* Left */
.header-left {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
}

.order-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  background: var(--color-accent-glow);
  border: 1px solid var(--color-accent-glow);
  color: var(--color-accent);
  font-size: 0.75rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 999px;
}

.order-chip-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-accent);
  animation: pulse 1.5s infinite;
}

.sync-pill {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 600;
  background: var(--color-surface);
  color: var(--color-text-dim);
  border: 1px solid var(--color-border);
}

.sync-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.sync-pill.online .sync-dot { background: var(--color-success); box-shadow: 0 0 6px rgba(39,174,96,0.5); }
.sync-pill.online { color: var(--color-success); }
.sync-pill.offline .sync-dot { background: var(--color-danger); }
.sync-pill.offline { color: var(--color-danger); }
.sync-pill.syncing .sync-dot { background: var(--color-warning); animation: pulse 1s infinite; }
.sync-pill.syncing { color: var(--color-warning); }
</style>