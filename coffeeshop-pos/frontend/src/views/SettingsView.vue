<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useManagement } from '../composables/useManagement'
import { useAuth } from '../composables/useAuth'
import UserManagementPanel from '../components/settings/UserManagementPanel.vue'
import TableManagementPanel from '../components/settings/TableManagementPanel.vue'
import APIConnectionPanel from '../components/settings/APIConnectionPanel.vue'

const { initBindings, triggerSync, isLoading } = useManagement()
const { currentUser } = useAuth()

const isDevUser = computed(() => currentUser.value?.role === 'dev' || currentUser.value?.role === 'admin')

const shopName = ref('المقهى')
const syncInterval = ref(30)
const syncMessage = ref('')
const menuBaseURL = ref('')
const menuURLSaved = ref(false)

let ConfigStoreService: any = null

async function onSync() {
  syncMessage.value = ''
  try {
    await triggerSync()
    syncMessage.value = '✓ تمت المزامنة بنجاح'
    setTimeout(() => { syncMessage.value = '' }, 3000)
  } catch (err: any) {
    syncMessage.value = '✕ فشلت المزامنة: ' + (err.message || err)
  }
}

onMounted(async () => {
  await initBindings()
  try {
    ConfigStoreService = await import('../../bindings/coffeeshop-pos/internal/service/configstoreservice')
    const savedURL = await ConfigStoreService.Get('menu_base_url')
    if (savedURL) menuBaseURL.value = savedURL
  } catch { /* not available */ }
})

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
    </header>

    <div class="settings-content">
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

      <div class="settings-section">
        <h2 class="settings-section-title">المزامنة</h2>
        <div class="form-group">
          <label class="form-label">فاصل المزامنة (ثانية)</label>
          <input v-model.number="syncInterval" type="number" class="form-input" min="10" max="300" style="width: 120px" />
        </div>

        <button class="btn btn-primary" :disabled="isLoading" @click="onSync">
          {{ isLoading ? 'جاري المزامنة...' : '🔄 مزامنة الآن' }}
        </button>

        <div v-if="syncMessage" class="sync-feedback" :class="{ success: syncMessage.startsWith('✓'), error: syncMessage.startsWith('✕') }">
          {{ syncMessage }}
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
}

.settings-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--gap-xl);
  max-width: 600px;
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

.sync-feedback {
  padding: var(--gap-sm) var(--gap-md);
  border-radius: var(--radius-sm);
  font-weight: var(--font-weight-semi);
  font-size: var(--font-size-sm);
}

.sync-feedback.success {
  background: rgba(39, 174, 96, 0.12);
  color: var(--color-success);
}

.sync-feedback.error {
  background: rgba(231, 76, 60, 0.12);
  color: var(--color-danger);
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
</style>
