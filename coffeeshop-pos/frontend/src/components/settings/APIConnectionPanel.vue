<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfigStore } from '../../composables/useConfigStore'

const { setupConnection, getConnection, error, isLoading } = useConfigStore()

const apiURL = ref('')
const username = ref('')
const password = ref('')
const connected = ref(false)
const showForm = ref(false)
const successMsg = ref('')

async function loadConnection() {
  const conn = await getConnection()
  if (conn) {
    apiURL.value = conn.api_url || ''
    username.value = conn.username || ''
    password.value = ''
    connected.value = !!conn.api_url
  }
}

async function onSave() {
  successMsg.value = ''
  try {
    await setupConnection(apiURL.value, username.value, password.value)
    connected.value = true
    showForm.value = false
    successMsg.value = '✅ تم الاتصال بنجاح'
    setTimeout(() => { successMsg.value = '' }, 3000)
  } catch {
    // error is handled by composable
  }
}

onMounted(loadConnection)
</script>

<template>
  <div class="api-connection">
    <div class="panel-header">
      <h2 class="panel-title">🔗 اتصال الخادم</h2>
      <span v-if="connected" class="status-badge connected">متصل</span>
      <span v-else class="status-badge disconnected">غير متصل</span>
    </div>

    <div v-if="connected && !showForm" class="connection-info">
      <div class="info-row">
        <span class="text-muted">رابط الخادم</span>
        <code dir="ltr">{{ apiURL }}</code>
      </div>
      <div class="info-row">
        <span class="text-muted">المستخدم</span>
        <span>{{ username }}</span>
      </div>
      <button class="btn-outline" @click="showForm = true">تعديل</button>
    </div>

    <div v-if="!connected || showForm" class="edit-form">
      <div class="form-group">
        <label class="form-label">رابط الخادم (API URL)</label>
        <input v-model="apiURL" type="url" class="form-input" dir="ltr" placeholder="http://localhost:8080" />
      </div>
      <div class="form-group">
        <label class="form-label">اسم المستخدم</label>
        <input v-model="username" type="text" class="form-input" dir="ltr" placeholder="admin" />
      </div>
      <div class="form-group">
        <label class="form-label">كلمة المرور</label>
        <input v-model="password" type="password" class="form-input" dir="ltr" placeholder="••••••••" />
      </div>

      <div v-if="error" class="form-error">{{ error }}</div>
      <div v-if="successMsg" class="form-success">{{ successMsg }}</div>

      <div class="form-actions">
        <button v-if="showForm" class="btn-outline" @click="showForm = false">إلغاء</button>
        <button class="btn-primary" :disabled="isLoading || !apiURL || !username || !password" @click="onSave">
          {{ isLoading ? 'جاري الاتصال...' : '💾 حفظ واتصال' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.api-connection {
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.panel-header {
  display: flex;
  align-items: center;
  gap: var(--gap-sm);
}

.panel-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
}

.status-badge {
  font-size: var(--font-size-xs);
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: var(--font-weight-semi);
}

.status-badge.connected {
  background: rgba(46, 204, 113, 0.15);
  color: #2ecc71;
}

.status-badge.disconnected {
  background: rgba(231, 76, 60, 0.15);
  color: #e74c3c;
}

.connection-info {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--gap-xs) 0;
}

.info-row code {
  font-size: var(--font-size-xs);
  background: var(--color-surface);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.form-input {
  padding: var(--gap-xs) var(--gap-sm);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.form-error {
  color: var(--color-danger);
  font-size: var(--font-size-sm);
}

.form-success {
  color: #2ecc71;
  font-size: var(--font-size-sm);
}

.form-actions {
  display: flex;
  gap: var(--gap-sm);
  justify-content: flex-end;
}

.btn-primary {
  padding: var(--gap-xs) var(--gap-md);
  background: var(--color-accent);
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-bg);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-primary:hover:not(:disabled) { filter: brightness(1.1); }
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }

.btn-outline {
  padding: var(--gap-xs) var(--gap-md);
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-outline:hover {
  background: var(--color-surface);
}
</style>
