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
      intro_video_url: string
    }
    created_at: string
  }
  users: User[]
  devices: Device[]
}

const detail = ref<TenantDetail | null>(null)
const activeTab = ref<'info' | 'users' | 'devices'>('info')

const videoUploading = ref(false)
const videoError = ref<string | null>(null)

const provisionCode = ref<string | null>(null)
const provisionExpiry = ref<string | null>(null)
const generatingCode = ref(false)

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

async function uploadVideo(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !detail.value) return

  videoUploading.value = true
  videoError.value = null

  try {
    const formData = new FormData()
    formData.append('file', file)

    const adminToken = localStorage.getItem('admin_token') || ''
    const uploadResp = await fetch(`${API_BASE}/api/v1/uploads?folder=tenant-videos`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${adminToken}` },
      body: formData,
    })

    if (!uploadResp.ok) {
      const body = await uploadResp.json().catch(() => ({}))
      throw new Error(body.error || `Upload failed (${uploadResp.status})`)
    }

    const uploadData = await uploadResp.json()
    const videoUrl = uploadData.data?.url || uploadData.url

    await put(`/api/v1/admin/tenants/${route.params.id}`, {
      settings: {
        ...detail.value.tenant.settings,
        intro_video_url: videoUrl,
      },
    })
    detail.value.tenant.settings.intro_video_url = videoUrl
  } catch (e: any) {
    videoError.value = e.message
  } finally {
    videoUploading.value = false
    input.value = ''
  }
}

async function removeVideo() {
  if (!detail.value) return
  try {
    await put(`/api/v1/admin/tenants/${route.params.id}`, {
      settings: {
        ...detail.value.tenant.settings,
        intro_video_url: '',
      },
    })
    detail.value.tenant.settings.intro_video_url = ''
  } catch {
    // handled by useApi
  }
}

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

async function generateProvisionCode() {
  if (!detail.value) return
  generatingCode.value = true
  try {
    const resp = await post<{ code: string; expires_at: string }>(
      `/api/v1/admin/tenants/${route.params.id}/provision-code`,
      {},
    )
    provisionCode.value = resp.code
    provisionExpiry.value = resp.expires_at
  } catch {
    // handled by useApi
  } finally {
    generatingCode.value = false
  }
}

function copyCode() {
  if (provisionCode.value) {
    navigator.clipboard.writeText(provisionCode.value)
  }
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

          <div class="setting-row" style="flex-direction: column; align-items: stretch; gap: var(--space-3);">
            <div>
              <div class="setting-label">فيديو شاشة الدخول</div>
              <div class="setting-desc">فيديو تعريفي يظهر خلف شاشة تسجيل الدخول في تطبيق نقطة البيع</div>
            </div>
            <div v-if="videoError" class="alert alert-error" style="margin: 0;">{{ videoError }}</div>
            <div v-if="detail.tenant.settings.intro_video_url" class="video-preview">
              <video :src="detail.tenant.settings.intro_video_url" class="video-thumb" muted playsinline controls></video>
              <div class="video-actions">
                <span class="video-url" :title="detail.tenant.settings.intro_video_url">{{ detail.tenant.settings.intro_video_url.split('/').pop() }}</span>
                <button class="btn btn-sm btn-danger" @click="removeVideo">حذف</button>
              </div>
            </div>
            <label class="upload-btn" :class="{ disabled: videoUploading }">
              <input type="file" accept="video/mp4" hidden @change="uploadVideo" :disabled="videoUploading" />
              {{ videoUploading ? 'جاري الرفع...' : (detail.tenant.settings.intro_video_url ? 'تغيير الفيديو' : 'رفع فيديو') }}
            </label>
          </div>

          <!-- Provision Code -->
          <div class="setting-row" style="flex-direction: column; align-items: stretch; gap: var(--space-3);">
            <div>
              <div class="setting-label">رمز الإعداد (Provisioning)</div>
              <div class="setting-desc">توليد رمز إعداد لمرة واحدة لربط تطبيق نقطة البيع بهذا المستأجر</div>
            </div>

            <div v-if="provisionCode" class="provision-result">
              <div class="provision-code-display">{{ provisionCode }}</div>
              <div class="provision-meta">
                <span>صالح حتى {{ formatDate(provisionExpiry!) }}</span>
                <button class="btn btn-sm btn-secondary" @click="copyCode">📋 نسخ</button>
              </div>
            </div>

            <button
              class="btn btn-primary"
              :disabled="generatingCode"
              @click="generateProvisionCode"
            >
              {{ generatingCode ? 'جاري التوليد...' : (provisionCode ? 'توليد رمز جديد' : 'توليد رمز إعداد') }}
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

.video-preview {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.video-thumb {
  width: 100%;
  max-height: 200px;
  border-radius: var(--radius-md);
  background: rgba(0, 0, 0, 0.3);
  object-fit: contain;
}

.video-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.video-url {
  font-size: var(--font-sm);
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  direction: ltr;
}

.upload-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--accent-bg);
  color: var(--accent);
  font-size: var(--font-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: center;
}

.upload-btn:hover {
  background: var(--accent);
  color: white;
}

.upload-btn.disabled {
  opacity: 0.5;
  pointer-events: none;
}

.provision-result {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-5);
  background: var(--accent-bg);
  border-radius: var(--radius-md);
  border: 1px dashed var(--accent);
}

.provision-code-display {
  font-size: 2.5rem;
  font-weight: 700;
  letter-spacing: 0.4em;
  color: var(--accent);
  font-family: 'SF Mono', 'Fira Code', monospace;
  direction: ltr;
  user-select: all;
}

.provision-meta {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-sm);
  color: var(--text-muted);
}
</style>
