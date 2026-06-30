<script setup lang="ts">
import { ref } from 'vue'
import { useConfigStore } from '../composables/useConfigStore'

const emit = defineEmits<{ complete: [] }>()
const { setupConnection, provisionWithCode, error, isLoading } = useConfigStore()

const mode = ref<'provision' | 'manual'>('provision')
const apiURL = ref('http://localhost:8080')
const provisionCode = ref('')
const username = ref('')
const password = ref('')
const shaking = ref(false)

function shake() {
  shaking.value = true
  setTimeout(() => { shaking.value = false }, 600)
}

async function onProvision() {
  if (!apiURL.value || !provisionCode.value) return
  try {
    await provisionWithCode(apiURL.value, provisionCode.value.toUpperCase())
    emit('complete')
  } catch {
    shake()
  }
}

async function onManualSubmit() {
  if (!apiURL.value || !username.value || !password.value) return
  try {
    await setupConnection(apiURL.value, username.value, password.value)
    emit('complete')
  } catch {
    shake()
  }
}
</script>

<template>
  <div class="setup-screen">
    <div class="setup-card" :class="{ shake: shaking }">
      <div class="setup-header">
        <span class="setup-icon">☕</span>
        <h1 class="setup-title">إعداد نقطة البيع</h1>
        <p class="setup-subtitle">اتصل بالخادم المركزي للبدء</p>
      </div>

      <!-- Mode toggle -->
      <div class="mode-toggle">
        <button
          class="mode-btn"
          :class="{ active: mode === 'provision' }"
          @click="mode = 'provision'"
        >
          رمز الإعداد
        </button>
        <button
          class="mode-btn"
          :class="{ active: mode === 'manual' }"
          @click="mode = 'manual'"
        >
          إعداد يدوي
        </button>
      </div>

      <!-- Provision code mode -->
      <form v-if="mode === 'provision'" class="setup-form" @submit.prevent="onProvision">
        <div class="form-group">
          <label class="form-label">رابط الخادم (API URL)</label>
          <input
            v-model="apiURL"
            type="url"
            class="form-input"
            placeholder="http://localhost:8080"
            dir="ltr"
            autocomplete="url"
          />
        </div>

        <div class="form-group">
          <label class="form-label">رمز الإعداد</label>
          <input
            v-model="provisionCode"
            type="text"
            class="form-input provision-code-input"
            placeholder="مثال: NJ3K7X"
            dir="ltr"
            autocomplete="off"
            maxlength="8"
          />
        </div>

        <div v-if="error" class="setup-error">{{ error }}</div>

        <button
          type="submit"
          class="btn-connect"
          :disabled="isLoading || !apiURL || !provisionCode"
        >
          {{ isLoading ? 'جاري التفعيل...' : '🔗 تفعيل' }}
        </button>
      </form>

      <!-- Manual mode -->
      <form v-else class="setup-form" @submit.prevent="onManualSubmit">
        <div class="form-group">
          <label class="form-label">رابط الخادم (API URL)</label>
          <input
            v-model="apiURL"
            type="url"
            class="form-input"
            placeholder="http://localhost:8080"
            dir="ltr"
            autocomplete="url"
          />
        </div>

        <div class="form-group">
          <label class="form-label">اسم المستخدم</label>
          <input
            v-model="username"
            type="text"
            class="form-input"
            placeholder="admin@tenant-slug"
            dir="ltr"
            autocomplete="username"
          />
        </div>

        <div class="form-group">
          <label class="form-label">كلمة المرور</label>
          <input
            v-model="password"
            type="password"
            class="form-input"
            placeholder="••••••••"
            dir="ltr"
            autocomplete="current-password"
          />
        </div>

        <div v-if="error" class="setup-error">{{ error }}</div>

        <button
          type="submit"
          class="btn-connect"
          :disabled="isLoading || !apiURL || !username || !password"
        >
          {{ isLoading ? 'جاري الاتصال...' : '🔗 اتصال' }}
        </button>
      </form>

      <p class="setup-hint">
        💡 احصل على رمز الإعداد من مزود الخدمة
      </p>
    </div>
  </div>
</template>

<style scoped>
.setup-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: var(--color-bg);
}

.setup-card {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-lg);
  padding: var(--gap-xl);
  max-width: 420px;
  width: 100%;
}

.setup-card.shake {
  animation: shakeAnim 0.5s ease;
}

@keyframes shakeAnim {
  0%, 100% { transform: translateX(0); }
  15% { transform: translateX(-12px); }
  30% { transform: translateX(10px); }
  45% { transform: translateX(-8px); }
  60% { transform: translateX(6px); }
  75% { transform: translateX(-4px); }
}

.setup-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-sm);
}

.setup-icon {
  font-size: 3.5rem;
}

.setup-title {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-extra);
}

.setup-subtitle {
  color: var(--color-text-muted);
  font-size: var(--font-size-md);
}

.mode-toggle {
  display: flex;
  gap: 2px;
  background: var(--color-surface);
  border-radius: var(--radius-md);
  padding: 2px;
  width: 100%;
}

.mode-btn {
  flex: 1;
  padding: var(--gap-xs) var(--gap-sm);
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-muted);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.mode-btn.active {
  background: var(--color-accent);
  color: var(--color-bg);
}

.setup-form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
  text-align: right;
}

.form-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  color: var(--color-text-muted);
}

.form-input {
  padding: var(--gap-sm) var(--gap-md);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  transition: border-color var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.provision-code-input {
  text-align: center;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.3em;
  text-transform: uppercase;
}

.setup-error {
  color: var(--color-danger);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
  text-align: center;
}

.btn-connect {
  padding: var(--gap-sm) var(--gap-lg);
  background: var(--color-accent);
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-bg);
  font-family: var(--font-family);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-bold);
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-top: var(--gap-sm);
}

.btn-connect:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-connect:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.setup-hint {
  font-size: var(--font-size-xs);
  color: var(--color-text-dim);
  margin-top: var(--gap-sm);
}
</style>
