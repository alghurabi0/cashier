<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import SidebarNav from './components/SidebarNav.vue'
import StatusBar from './components/StatusBar.vue'
import SetupScreen from './views/SetupScreen.vue'
import LoginScreen from './views/LoginScreen.vue'
import PosView from './views/PosView.vue'
import WebOrdersView from './views/WebOrdersView.vue'
import KitchenView from './views/KitchenView.vue'
import OrderHistoryView from './views/OrderHistoryView.vue'
import ReportsView from './views/ReportsView.vue'
import InventoryView from './views/InventoryView.vue'
import SettingsView from './views/SettingsView.vue'
import { useAuth } from './composables/useAuth'
import { useConfigStore } from './composables/useConfigStore'
import { useWebOrders } from './composables/useWebOrders'

const activeView = ref('pos')
const lastSyncTime = ref('')
const kitchenModeEnabled = ref(false)
let DataService: any = null
let ConfigStoreService: any = null
let syncTimePoll: ReturnType<typeof setInterval> | null = null

async function pollSyncTime() {
  if (!DataService) {
    try {
      DataService = await import('../bindings/coffeeshop-pos/internal/service/dataservice')
    } catch { return }
  }
  try {
    const raw = await DataService.GetLastSyncTime()
    if (raw) {
      const d = new Date(raw)
      lastSyncTime.value = d.toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
    }
  } catch { /* ignore */ }
}

const { currentUser, initBindings: initAuth, checkExistingSession, logout } = useAuth()
const { isSetup, initBindings: initConfig, checkSetup } = useConfigStore()
const { pendingCount, initBindings: initWebOrders, startPolling, stopPolling } = useWebOrders()

const isLoggedIn = computed(() => currentUser.value !== null)
const userRole = computed(() => currentUser.value?.role || 'cashier')
const userName = computed(() => currentUser.value?.name_ar || '')

// App state: 'loading' → 'setup' → 'login' → 'ready'
const appState = ref<'loading' | 'setup' | 'login' | 'ready'>('loading')

onMounted(async () => {
  await initConfig()
  await initAuth()

  // Check if API connection is configured
  const configured = await checkSetup()
  if (!configured) {
    appState.value = 'setup'
    return
  }

  // Load kitchen mode setting
  try {
    ConfigStoreService = await import('../bindings/coffeeshop-pos/internal/service/configstoreservice')
    const kitchenVal = await ConfigStoreService.Get('kitchen_mode_enabled')
    kitchenModeEnabled.value = kitchenVal === 'true'
  } catch { /* not available */ }

  // API is configured — check for existing PIN session
  await checkExistingSession()
  if (currentUser.value) {
    appState.value = 'ready'
    activeView.value = (currentUser.value.role === 'kitchen' && kitchenModeEnabled.value) ? 'kitchen' : 'pos'
    await initWebOrders()
    startPolling(2000)
    // Start sync time polling
    await pollSyncTime()
    syncTimePoll = setInterval(pollSyncTime, 30000)
  } else {
    appState.value = 'login'
  }
})

// Called when setup wizard completes successfully
async function onSetupComplete() {
  await initAuth()
  appState.value = 'login'
}

// Called when PIN login succeeds (handled by useAuth reactivity)
// We watch isLoggedIn to transition from login → ready
import { watch } from 'vue'
watch(isLoggedIn, async (loggedIn) => {
  if (loggedIn && appState.value === 'login') {
    appState.value = 'ready'
    activeView.value = (currentUser.value?.role === 'kitchen' && kitchenModeEnabled.value) ? 'kitchen' : 'pos'
    await initWebOrders()
    startPolling(2000)
    await pollSyncTime()
    syncTimePoll = setInterval(pollSyncTime, 30000)
  }
})

async function onLogout() {
  stopPolling()
  if (syncTimePoll) { clearInterval(syncTimePoll); syncTimePoll = null }
  await logout()
  activeView.value = 'pos'
  appState.value = 'login'
}
</script>

<template>
  <!-- Loading state -->
  <div v-if="appState === 'loading'" class="app-loading">
    <span class="loading-icon">☕</span>
    <span class="loading-text">جاري التحميل...</span>
  </div>

  <!-- First-time setup wizard -->
  <SetupScreen v-else-if="appState === 'setup'" @complete="onSetupComplete" />

  <!-- PIN login -->
  <LoginScreen v-else-if="appState === 'login'" />

  <!-- Main app -->
  <div v-else class="app-layout">
    <SidebarNav
      :active="activeView"
      :pending-web-orders="pendingCount"
      :user-role="userRole"
      :user-name="userName"
      :kitchen-mode-enabled="kitchenModeEnabled"
      @navigate="activeView = $event"
      @logout="onLogout"
    />

    <div class="app-content">
      <div class="app-main">
        <PosView v-if="activeView === 'pos'" />
        <WebOrdersView v-else-if="activeView === 'web-orders'" />
        <KitchenView v-else-if="activeView === 'kitchen'" />
        <OrderHistoryView v-else-if="activeView === 'order-history'" />
        <ReportsView v-else-if="activeView === 'reports'" />
        <InventoryView v-else-if="activeView === 'inventory'" />
        <SettingsView v-else-if="activeView === 'settings'" />
      </div>
      <StatusBar :last-sync-time="lastSyncTime" />
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.app-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.app-main {
  flex: 1;
  overflow: hidden;
}

.app-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: var(--gap-md);
}

.loading-icon {
  font-size: 3rem;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.95); }
}

.loading-text {
  color: var(--color-text-muted);
  font-size: var(--font-size-md);
}
</style>
