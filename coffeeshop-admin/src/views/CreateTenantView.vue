<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { post, loading, error } = useApi()

const shopName = ref('')
const slug = ref('')
const adminUsername = ref('')
const adminPassword = ref('')
const slugManuallyEdited = ref(false)
const success = ref(false)

interface CreateResult {
  tenant: { id: string; name: string; slug: string }
  token: string
  user: { username: string }
}

const createdResult = ref<CreateResult | null>(null)
const showPassword = ref(false)
const copied = ref(false)

// Auto-generate slug from shop name
watch(shopName, (name) => {
  if (!slugManuallyEdited.value) {
    slug.value = name
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^\u0621-\u064Aa-z0-9-]/g, '')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '')
  }
})

function onSlugInput() {
  slugManuallyEdited.value = true
}

const apiUrl = computed(() => import.meta.env.VITE_API_URL || 'http://localhost:8080')

async function handleCreate() {
  try {
    createdResult.value = await post<CreateResult>('/api/v1/tenants', {
      name: shopName.value,
      slug: slug.value,
      admin_username: adminUsername.value,
      admin_password: adminPassword.value,
    })
    success.value = true
  } catch {
    // error is set by useApi
  }
}

function copySetupInfo() {
  if (!createdResult.value) return
  const info = `
معلومات إعداد نقطة البيع
═══════════════════════
اسم المقهى: ${createdResult.value.tenant.name}
رابط الخادم: ${apiUrl.value}
اسم المستخدم: ${createdResult.value.user.username}@${createdResult.value.tenant.slug}
كلمة المرور: ${adminPassword.value}
  `.trim()
  navigator.clipboard.writeText(info)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

function resetForm() {
  shopName.value = ''
  slug.value = ''
  adminUsername.value = ''
  adminPassword.value = ''
  slugManuallyEdited.value = false
  success.value = false
  createdResult.value = null
}
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ success ? '✅ تم الإنشاء بنجاح' : '➕ مستأجر جديد' }}</h1>
        <p class="page-subtitle">{{ success ? 'معلومات الإعداد جاهزة للمشاركة مع العميل' : 'إنشاء مقهى جديد على المنصة' }}</p>
      </div>
    </div>

    <!-- Success: Setup Instructions Card -->
    <div v-if="success && createdResult" class="setup-card">
      <div class="setup-header">
        <span class="setup-icon">🎉</span>
        <div>
          <h2>تم إنشاء "{{ createdResult.tenant.name }}" بنجاح!</h2>
          <p>شارك المعلومات التالية مع العميل لإعداد نقطة البيع</p>
        </div>
      </div>

      <div class="setup-info">
        <div class="info-row">
          <span class="info-label">رابط الخادم</span>
          <code class="info-value ltr">{{ apiUrl }}</code>
        </div>
        <div class="info-row">
          <span class="info-label">اسم المستخدم</span>
          <code class="info-value ltr">{{ createdResult.user.username }}@{{ createdResult.tenant.slug }}</code>
        </div>
        <div class="info-row">
          <span class="info-label">كلمة المرور</span>
          <div style="display: flex; align-items: center; gap: var(--space-2);">
            <code class="info-value ltr">{{ showPassword ? adminPassword : '••••••••' }}</code>
            <button class="btn btn-sm btn-secondary" @click="showPassword = !showPassword">
              {{ showPassword ? 'إخفاء' : 'إظهار' }}
            </button>
          </div>
        </div>
      </div>

      <div class="setup-actions">
        <button class="btn btn-primary" @click="copySetupInfo">
          {{ copied ? '✅ تم النسخ!' : '📋 نسخ معلومات الإعداد' }}
        </button>
        <button class="btn btn-secondary" @click="resetForm">
          ➕ إنشاء مستأجر آخر
        </button>
        <router-link :to="`/tenants/${createdResult.tenant.id}`" class="btn btn-secondary">
          عرض التفاصيل ←
        </router-link>
      </div>
    </div>

    <!-- Create Form -->
    <div v-else class="card" style="max-width: 600px;">
      <form @submit.prevent="handleCreate" class="create-form">
        <div v-if="error" class="alert alert-error">{{ error }}</div>

        <div class="input-group">
          <label for="shopName">اسم المقهى</label>
          <input id="shopName" v-model="shopName" class="input" placeholder="مثال: مقهى النخبة" required />
        </div>

        <div class="input-group">
          <label for="slug">المعرّف (Slug)</label>
          <input
            id="slug"
            v-model="slug"
            class="input"
            placeholder="nj-coffee"
            style="direction: ltr; text-align: left;"
            @input="onSlugInput"
            required
          />
          <small style="color: var(--text-muted); font-size: var(--font-xs);">
            يُستخدم في تسجيل الدخول: username@{{ slug || 'slug' }}
          </small>
        </div>

        <div class="input-group">
          <label for="adminUser">اسم مستخدم المدير</label>
          <input id="adminUser" v-model="adminUsername" class="input" placeholder="admin" required />
        </div>

        <div class="input-group">
          <label for="adminPass">كلمة مرور المدير</label>
          <input id="adminPass" v-model="adminPassword" type="password" class="input" placeholder="6 أحرف على الأقل" minlength="6" required />
        </div>

        <button type="submit" class="btn btn-primary btn-lg" style="width: 100%; margin-top: var(--space-4);" :disabled="loading">
          <span v-if="loading" class="spinner" style="width: 18px; height: 18px;"></span>
          <span v-else>إنشاء المستأجر</span>
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.create-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.setup-card {
  background: var(--bg-card);
  border: 1px solid var(--border-active);
  border-radius: var(--radius-xl);
  padding: var(--space-8);
  box-shadow: var(--shadow-glow);
  max-width: 700px;
}

.setup-header {
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
  margin-bottom: var(--space-8);
}

.setup-icon {
  font-size: 2.5rem;
}

.setup-header h2 {
  font-size: var(--font-xl);
  font-weight: 700;
  margin-bottom: var(--space-1);
}

.setup-header p {
  color: var(--text-secondary);
  font-size: var(--font-sm);
}

.setup-info {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--space-5);
  margin-bottom: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.info-label {
  font-size: var(--font-sm);
  color: var(--text-secondary);
  font-weight: 500;
  min-width: 120px;
}

.info-value {
  font-size: var(--font-sm);
  background: rgba(99, 102, 241, 0.08);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  color: var(--accent-hover);
  font-family: monospace;
}

.ltr {
  direction: ltr;
  display: inline-block;
}

.setup-actions {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}
</style>
