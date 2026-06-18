<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

defineProps<{
  orderCount: number
  syncStatus: 'online' | 'offline' | 'syncing'
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
      <div class="order-chip" v-if="orderCount > 0">
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
  background: #111;
  border-bottom: 1px solid rgba(201,168,76,0.15);
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
  filter: drop-shadow(0 0 6px rgba(201,168,76,0.5));
}

.logo-texts {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.logo-name {
  font-size: 1rem;
  font-weight: 800;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: 0.08em;
}

.logo-sub {
  font-size: 0.62rem;
  color: #555;
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
  color: #e8dcc8;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.date {
  font-size: 0.65rem;
  color: #555;
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
  background: rgba(201,168,76,0.12);
  border: 1px solid rgba(201,168,76,0.3);
  color: #c9a84c;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 999px;
}

.order-chip-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #c9a84c;
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
  background: #1a1a1a;
  color: #555;
  border: 1px solid rgba(255,255,255,0.05);
}

.sync-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.sync-pill.online .sync-dot { background: #27ae60; box-shadow: 0 0 6px rgba(39,174,96,0.5); }
.sync-pill.online { color: #27ae60; }
.sync-pill.offline .sync-dot { background: #e74c3c; }
.sync-pill.offline { color: #e74c3c; }
.sync-pill.syncing .sync-dot { background: #f39c12; animation: pulse 1s infinite; }
.sync-pill.syncing { color: #f39c12; }
</style>