<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { get, put, loading, error } = useApi()

interface AppRelease {
  version: string
  release_notes: string
  download_url_win: string
  download_url_mac: string
  updated_at: string
}

const release = ref<AppRelease | null>(null)
const form = ref({
  version: '',
  release_notes: '',
  download_url_win: '',
  download_url_mac: '',
})
const saving = ref(false)
const saved = ref(false)

onMounted(async () => {
  try {
    release.value = await get<AppRelease>('/api/v1/admin/app-release')
    form.value = {
      version: release.value.version,
      release_notes: release.value.release_notes,
      download_url_win: release.value.download_url_win,
      download_url_mac: release.value.download_url_mac,
    }
  } catch {
    // handled by useApi
  }
})

async function saveRelease() {
  saving.value = true
  saved.value = false
  try {
    release.value = await put<AppRelease>('/api/v1/admin/app-release', form.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 3000)
  } catch {
    // handled by useApi
  } finally {
    saving.value = false
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
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">إدارة تحديث التطبيق</h1>
        <p class="page-subtitle">تحكم بإصدار تطبيق نقطة البيع وروابط التحميل</p>
      </div>
    </div>

    <div v-if="loading && !release" style="display: flex; justify-content: center; padding: 4rem;">
      <div class="spinner"></div>
    </div>

    <div v-else-if="error && !release" class="alert alert-error">{{ error }}</div>

    <div v-else-if="release" class="card release-card">
      <div class="current-version">
        <span class="version-label">الإصدار الحالي</span>
        <span class="version-value">{{ release.version }}</span>
        <span v-if="release.updated_at" class="version-date">آخر تحديث: {{ formatDate(release.updated_at) }}</span>
      </div>

      <form class="release-form" @submit.prevent="saveRelease">
        <div class="input-group">
          <label>رقم الإصدار</label>
          <input v-model="form.version" class="input" placeholder="1.0.0" dir="ltr" required />
        </div>

        <div class="input-group">
          <label>ملاحظات الإصدار</label>
          <textarea v-model="form.release_notes" class="input textarea" placeholder="ما الجديد في هذا الإصدار..." rows="3"></textarea>
        </div>

        <div class="input-group">
          <label>رابط تحميل Windows</label>
          <input v-model="form.download_url_win" class="input" placeholder="https://..." dir="ltr" type="url" />
        </div>

        <div class="input-group">
          <label>رابط تحميل macOS</label>
          <input v-model="form.download_url_mac" class="input" placeholder="https://..." dir="ltr" type="url" />
        </div>

        <div class="form-actions">
          <button type="submit" class="btn btn-primary" :disabled="saving || !form.version">
            {{ saving ? 'جاري الحفظ...' : 'حفظ التغييرات' }}
          </button>
          <span v-if="saved" class="save-success">تم الحفظ بنجاح ✓</span>
        </div>
      </form>

      <div class="info-box">
        <strong>كيف يعمل التحديث التلقائي:</strong>
        <p>تطبيق نقطة البيع يتحقق تلقائياً من هذا الإصدار عند التشغيل. إذا كان الإصدار هنا أحدث من إصدار التطبيق المثبت، يُعرض إشعار تحديث للمستخدم مع رابط التحميل المناسب لنظام التشغيل.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.release-card {
  max-width: 600px;
}

.current-version {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-5);
  margin-bottom: var(--space-6);
  background: var(--accent-bg);
  border-radius: var(--radius-md);
}

.version-label {
  font-size: var(--font-sm);
  color: var(--text-muted);
}

.version-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--accent);
  font-family: 'SF Mono', 'Fira Code', monospace;
  direction: ltr;
}

.version-date {
  font-size: var(--font-sm);
  color: var(--text-secondary);
}

.release-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.textarea {
  resize: vertical;
  min-height: 80px;
}

.form-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-top: var(--space-2);
}

.save-success {
  color: var(--success);
  font-size: var(--font-sm);
  font-weight: 500;
}

.info-box {
  margin-top: var(--space-6);
  padding: var(--space-4);
  background: rgba(255, 255, 255, 0.03);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  font-size: var(--font-sm);
  color: var(--text-secondary);
  line-height: 1.6;
}

.info-box strong {
  display: block;
  margin-bottom: var(--space-2);
  color: var(--text-primary);
}
</style>
