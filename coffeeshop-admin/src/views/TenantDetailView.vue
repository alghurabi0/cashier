<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../composables/useApi'

const route = useRoute()
const { get, put, post, loading, error } = useApi()

interface User {
  id: string
  username: string
  role: string
  created_at: string
}

interface Device {
  id: string
  device_name: string
  device_type: string
  is_active: boolean
  last_seen_at: string | null
  created_at: string
}

interface TenantDetail {
  tenant: {
    id: string
    name: string
    slug: string
    is_active: boolean
    settings: {
      kitchen_mode_enabled: boolean
      conflict_resolution_mode: string
    }
    created_at: string
  }
  users: User[]
  devices: Device[]
}

const detail = ref<TenantDetail | null>(null)
const activeTab = ref<'info' | 'users' | 'devices'>('info')

// Add user form
const showAddUser = ref(false)
const newUsername = ref('')
const newPassword = ref('')
const newRole = ref('cashier')
const addUserError = ref<string | null>(null)

onMounted(async () => {
  await loadTenant()
})

async function loadTenant() {
  try {
    const tenantDetail = await get<TenantDetail>(`/api/v1/admin/tenants/${route.params.id}`)
    detail.value = {
      ...tenantDetail,
      users: tenantDetail.users ?? [],
      devices: tenantDetail.devices ?? [],
    }
  } catch {
    // handled
  }
}

async function toggleActive() {
  if (!detail.value) return
  const newStatus = !detail.value.tenant.is_active
  try {
    await put(`/api/v1/admin/tenants/${route.params.id}`, { is_active: newStatus })
    detail.value.tenant.is_active = newStatus
  } catch {
    // handled
  }
}

async function toggleKitchenMode() {
  if (!detail.value) return
  const newVal = !detail.value.tenant.settings.kitchen_mode_enabled
  try {
    await put(`/api/v1/admin/tenants/${route.params.id}`, {
      settings: {
        ...detail.value.tenant.settings,
        kitchen_mode_enabled: newVal,
      },
    })
    detail.value.tenant.settings.kitchen_mode_enabled = newVal
  } catch {
    // handled
  }
}

async function toggleConflictMode() {
  if (!detail.value) return
  const newVal = detail.value.tenant.settings.conflict_resolution_mode === 'last-write-wins'
    ? 'manual'
    : 'last-write-wins'
  try {
    await put(`/api/v1/admin/tenants/${route.params.id}`, {
      settings: {
        ...detail.value.tenant.settings,
        conflict_resolution_mode: newVal,
      },
    })
    detail.value.tenant.settings.conflict_resolution_mode = newVal
  } catch {
    // handled
  }
}

async function addUser() {
  addUserError.value = null
  try {
    await post(`/api/v1/admin/tenants/${route.params.id}/users`, {
      username: newUsername.value,
      password: newPassword.value,
      role: newRole.value,
    })
    showAddUser.value = false
    newUsername.value = ''
    newPassword.value = ''
    newRole.value = 'cashier'
    await loadTenant()
  } catch (e: any) {
    addUserError.value = e.message
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ar-IQ', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatLastSeen(iso: string | null): string {
  if (!iso) return 'لم يظهر بعد'
  return formatDate(iso)
}
</script>

<template>
  <div>
    <div v-if="loading && !detail" style="display: flex; justify-content: center; padding: 4rem;">
      <div class="spinner"></div>
    </div>

    <div v-else-if="error && !detail" class="alert alert-error">{{ error }}</div>

    <template v-else-if="detail">
      <!-- Header -->
      <div class="page-header">
        <div>
          <div style="display: flex; align-items: center; gap: var(--space-3);">
            <h1 class="page-title">{{ detail.tenant.name }}</h1>
            <span :class="['badge', detail.tenant.is_active ? 'badge-success' : 'badge-danger']">
              {{ detail.tenant.is_active ? 'نشط' : 'معلّق' }}
            </span>
          </div>
          <p class="page-subtitle">
            <code style="direction: ltr; display: inline-block;">{{ detail.tenant.slug }}</code>
            · أُنشئ {{ formatDate(detail.tenant.created_at) }}
          </p>
        </div>
        <button
          class="btn"
          :class="detail.tenant.is_active ? 'btn-danger' : 'btn-primary'"
          @click="toggleActive"
        >
          {{ detail.tenant.is_active ? '⏸️ تعليق' : '▶️ تفعيل' }}
        </button>
      </div>

      <!-- Tabs -->
      <div class="tabs">
        <button
          v-for="tab in [
            { key: 'info', label: 'المعلومات', icon: '⚙️' },
            { key: 'users', label: `المستخدمون (${detail.users.length})`, icon: '👥' },
            { key: 'devices', label: `الأجهزة (${detail.devices.length})`, icon: '🖥️' },
          ]"
          :key="tab.key"
          class="tab-btn"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key as any"
        >
          {{ tab.icon }} {{ tab.label }}
        </button>
      </div>

      <!-- Info Tab -->
      <div v-if="activeTab === 'info'" class="card" style="max-width: 600px;">
        <h3 style="margin-bottom: var(--space-5); font-size: var(--font-lg);">الإعدادات</h3>

        <div class="settings-list">
          <div class="setting-row">
            <div>
              <div class="setting-label">وضع المطبخ</div>
              <div class="setting-desc">إرسال الطلبات إلى شاشة المطبخ</div>
            </div>
            <button class="toggle-btn" :class="{ on: detail.tenant.settings.kitchen_mode_enabled }" @click="toggleKitchenMode">
              <span class="toggle-knob"></span>
            </button>
          </div>

          <div class="setting-row">
            <div>
              <div class="setting-label">حل التعارضات</div>
              <div class="setting-desc">
                {{ detail.tenant.settings.conflict_resolution_mode === 'last-write-wins' ? 'الكتابة الأخيرة تفوز (تلقائي)' : 'يدوي' }}
              </div>
            </div>
            <button class="btn btn-sm btn-secondary" @click="toggleConflictMode">
              تبديل
            </button>
          </div>
        </div>
      </div>

      <!-- Users Tab -->
      <div v-if="activeTab === 'users'">
        <div style="margin-bottom: var(--space-4);">
          <button class="btn btn-primary btn-sm" @click="showAddUser = !showAddUser">
            {{ showAddUser ? 'إلغاء' : '➕ إضافة مستخدم' }}
          </button>
        </div>

        <!-- Add User Form -->
        <div v-if="showAddUser" class="card" style="max-width: 500px; margin-bottom: var(--space-5);">
          <form @submit.prevent="addUser" class="add-user-form">
            <div v-if="addUserError" class="alert alert-error">{{ addUserError }}</div>
            <div class="input-group">
              <label>اسم المستخدم</label>
              <input v-model="newUsername" class="input" placeholder="اسم المستخدم" required />
            </div>
            <div class="input-group">
              <label>كلمة المرور</label>
              <input v-model="newPassword" type="password" class="input" placeholder="6 أحرف على الأقل" minlength="6" required />
            </div>
            <div class="input-group">
              <label>الدور</label>
              <select v-model="newRole" class="input">
                <option value="cashier">كاشير</option>
                <option value="admin">مدير</option>
              </select>
            </div>
            <button type="submit" class="btn btn-primary" :disabled="loading">إضافة</button>
          </form>
        </div>

        <div v-if="detail.users.length === 0" class="empty-state">
          <p>لا يوجد مستخدمون</p>
        </div>
        <div v-else class="card" style="padding: 0; overflow: hidden;">
          <table class="data-table">
            <thead>
              <tr>
                <th>اسم المستخدم</th>
                <th>الدور</th>
                <th>تاريخ الإنشاء</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in detail.users" :key="user.id">
                <td style="font-weight: 500;">{{ user.username }}</td>
                <td>
                  <span :class="['badge', user.role === 'admin' ? 'badge-warning' : 'badge-success']">
                    {{ user.role === 'admin' ? 'مدير' : 'كاشير' }}
                  </span>
                </td>
                <td style="color: var(--text-secondary); font-size: var(--font-sm);">{{ formatDate(user.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Devices Tab -->
      <div v-if="activeTab === 'devices'">
        <div v-if="detail.devices.length === 0" class="empty-state">
          <div class="empty-icon">🖥️</div>
          <p>لا توجد أجهزة مسجلة بعد</p>
        </div>
        <div v-else class="card" style="padding: 0; overflow: hidden;">
          <table class="data-table">
            <thead>
              <tr>
                <th>اسم الجهاز</th>
                <th>النوع</th>
                <th>الحالة</th>
                <th>آخر ظهور</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="device in detail.devices" :key="device.id">
                <td style="font-weight: 500;">{{ device.device_name }}</td>
                <td>
                  <span class="badge badge-success">
                    {{ device.device_type === 'pos' ? 'نقطة بيع' : 'شاشة مطبخ' }}
                  </span>
                </td>
                <td>
                  <span :class="['badge', device.is_active ? 'badge-success' : 'badge-danger']">
                    {{ device.is_active ? 'نشط' : 'غير نشط' }}
                  </span>
                </td>
                <td style="color: var(--text-secondary); font-size: var(--font-sm);">
                  {{ formatLastSeen(device.last_seen_at) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.tabs {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-6);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: var(--space-3);
}

.tab-btn {
  padding: var(--space-3) var(--space-4);
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-family: var(--font-family);
  font-size: var(--font-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab-btn:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.04);
}

.tab-btn.active {
  color: var(--accent);
  background: var(--accent-bg);
  font-weight: 500;
}

.settings-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--border-color);
}

.setting-row:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.setting-label {
  font-weight: 500;
  margin-bottom: var(--space-1);
}

.setting-desc {
  font-size: var(--font-sm);
  color: var(--text-muted);
}

/* Toggle Button */
.toggle-btn {
  width: 48px;
  height: 26px;
  border-radius: 13px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  position: relative;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.toggle-btn.on {
  background: var(--accent);
}

.toggle-knob {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: white;
  transition: transform var(--transition-fast);
}

.toggle-btn.on .toggle-knob {
  transform: translateX(-22px);
}

.add-user-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
</style>
