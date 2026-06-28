<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useManagement } from '../composables/useManagement'
import { useAuth } from '../composables/useAuth'
import UserManagementPanel from '../components/settings/UserManagementPanel.vue'
import TableManagementPanel from '../components/settings/TableManagementPanel.vue'
import APIConnectionPanel from '../components/settings/APIConnectionPanel.vue'
import SyncDashboardPanel from '../components/settings/SyncDashboardPanel.vue'

const { initBindings, isLoading } = useManagement()
const { currentUser } = useAuth()

const isDevUser = computed(() => currentUser.value?.role === 'dev' || currentUser.value?.role === 'admin')

// Sub-navigation within settings
const activeTab = ref<'general' | 'sync'>('general')

const shopName = ref('المقهى')
const menuBaseURL = ref('')
const menuURLSaved = ref(false)
const kitchenModeEnabled = ref(false)

let ConfigStoreService: any = null

onMounted(async () => {
  await initBindings()
  try {
    ConfigStoreService = await import('../../bindings/coffeeshop-pos/internal/service/configstoreservice')
    const savedURL = await ConfigStoreService.Get('menu_base_url')
    if (savedURL) menuBaseURL.value = savedURL
    // Load kitchen mode setting
    const kitchenVal = await ConfigStoreService.Get('kitchen_mode_enabled')
    kitchenModeEnabled.value = kitchenVal === 'true'
  } catch { /* not available */ }
})

async function toggleKitchenMode() {
  if (!ConfigStoreService) return
  try {
    const newVal = !kitchenModeEnabled.value
    await ConfigStoreService.SetKitchenModeEnabled(newVal)
    kitchenModeEnabled.value = newVal
  } catch (err) {
    console.error('Failed to toggle kitchen mode:', err)
  }
}

async function saveMenuBaseURL() {
  if (!ConfigStoreService) return
  try {
    await ConfigStoreService.Set('menu_base_url', menuBaseURL.value.trim())
    menuURLSaved.value = true
    setTimeout(() => { menuURLSaved.value = false }, 2000)
  } catch (err) {
    console.error('Failed to save menu base URL:', err)
  }
}
</script>

<template>
  <div class="settings-view">
    <header class="view-header">
      <h1 class="view-title">⚙️ الإعدادات</h1>
      <!-- Sub-navigation tabs -->
      <nav class="settings-tabs">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'general' }"
          @click="activeTab = 'general'"
        >
          🏪 عام
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'sync' }"
          @click="activeTab = 'sync'"
        >
          📡 المزامنة
        </button>
      </nav>
    </header>

    <!-- General Settings Tab -->
    <div v-if="activeTab === 'general'" class="settings-content">
      <div class="settings-section">
        <h2 class="settings-section-title">معلومات المقهى</h2>
        <div class="form-group">
          <label class="form-label">اسم المقهى</label>
          <input v-model="shopName" type="text" class="form-input" />
          <p class="form-hint text-muted text-sm">يظهر على الإيصالات وشريط العنوان</p>
        </div>
      </div>

      <div class="settings-divider"></div>

      <!-- API Connection -->
      <div class="settings-section">
        <APIConnectionPanel />
      </div>

      <div class="settings-divider"></div>

      <!-- Workflow Settings -->
      <div class="settings-section">
        <h2 class="settings-section-title">🔄 سير العمل</h2>
        <div class="form-group">
          <div class="toggle-row">
            <div class="toggle-info">
              <span class="form-label">وضع المطبخ</span>
              <p class="form-hint text-muted text-sm">
                عند التفعيل، تمر الطلبات بمرحلة التحضير في المطبخ قبل اكتمالها.
                عند التعطيل، تُعتبر الطلبات مكتملة مباشرة عند إنشائها أو قبولها.
              </p>
            </div>
            <button
              class="toggle-btn"
              :class="{ active: kitchenModeEnabled }"
              @click="toggleKitchenMode"
            >
              <span class="toggle-knob"></span>
            </button>
          </div>
        </div>
      </div>

      <div class="settings-divider"></div>

      <!-- User Management -->
      <div class="settings-section">
        <UserManagementPanel />
      </div>

      <div class="settings-divider"></div>

      <!-- Table Management -->
      <div class="settings-section">
        <TableManagementPanel />
      </div>

      <!-- Dev Settings (role-guarded) -->
      <template v-if="isDevUser">
        <div class="settings-divider"></div>
        <div class="settings-section">
          <h2 class="settings-section-title">🔧 إعدادات المطور</h2>
          <div class="form-group">
            <label class="form-label">رابط القائمة الإلكترونية (Menu Base URL)</label>
            <div class="input-with-action">
              <input
                v-model="menuBaseURL"
                type="url"
                class="form-input"
                placeholder="https://menu.example.com"
              />
              <button class="btn btn-primary btn-sm" @click="saveMenuBaseURL">
                {{ menuURLSaved ? '✓ تم الحفظ' : 'حفظ' }}
              </button>
            </div>
            <p class="form-hint text-muted text-sm">يُستخدم لتوليد رموز QR للطاولات</p>
          </div>
        </div>
      </template>

      <div class="settings-divider"></div>

      <div class="settings-section">
        <h2 class="settings-section-title">حول التطبيق</h2>
        <div class="about-info">
          <div class="about-row">
            <span class="text-muted">الإصدار</span>
            <span>1.0.0-alpha</span>
          </div>
          <div class="about-row">
            <span class="text-muted">النظام</span>
            <span>Wails v3 + Vue 3</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Sync Dashboard Tab -->
    <div v-else-if="activeTab === 'sync'" class="settings-content">
      <SyncDashboardPanel />
    </div>
  </div>
</template>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.view-header {
  padding: var(--gap-lg) var(--gap-xl);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.view-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  margin-bottom: var(--gap-md);
}

/* Settings sub-tabs */
.settings-tabs {
  display: flex;
  gap: var(--gap-xs);
}

.tab-btn {
  padding: var(--gap-sm) var(--gap-lg);
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all 0.2s;
}

.tab-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-text);
}

.tab-btn.active {
  background: var(--color-accent);
  color: white;
}

.settings-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-xl);
  max-width: 700px;
}

.settings-section {
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.settings-section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  margin-bottom: var(--gap-xs);
}

.settings-divider {
  height: 1px;
  background: var(--color-border);
  margin: var(--gap-xl) 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
}

.form-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  color: var(--color-text-muted);
}

.form-input {
  padding: var(--gap-sm) var(--gap-md);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  max-width: 400px;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.form-hint {
  margin-top: 2px;
}

.input-with-action {
  display: flex;
  gap: var(--gap-sm);
  align-items: center;
}

.about-info {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
}

.about-row {
  display: flex;
  justify-content: space-between;
  padding: var(--gap-sm) 0;
  border-bottom: 1px solid var(--color-border);
}

/* Toggle */
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--gap-lg);
}

.toggle-info {
  flex: 1;
}

.toggle-btn {
  position: relative;
  width: 52px;
  height: 28px;
  border-radius: 14px;
  border: none;
  background: var(--color-surface-2);
  cursor: pointer;
  transition: background 0.25s ease;
  flex-shrink: 0;
}

.toggle-btn.active {
  background: var(--color-success, #27ae60);
}

.toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: white;
  transition: transform 0.25s ease;
  box-shadow: 0 1px 4px rgba(0,0,0,0.3);
}

.toggle-btn.active .toggle-knob {
  transform: translateX(24px);
}
</style>
