<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { post, setToken, loading, error } = useApi()

const username = ref('')
const password = ref('')

async function handleLogin() {
  try {
    const result = await post<{ token: string }>('/api/v1/auth/login', {
      username: username.value,
      password: password.value,
    })
    setToken(result.token)
    router.push('/')
  } catch {
    // error is set by useApi
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <div class="login-logo">☕</div>
        <h1>لوحة الإدارة</h1>
        <p>إدارة المستأجرين والأجهزة والإعدادات</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div v-if="error" class="alert alert-error">{{ error }}</div>

        <div class="input-group">
          <label for="username">اسم المستخدم</label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="input"
            placeholder="superadmin@platform"
            autocomplete="username"
            required
          />
        </div>

        <div class="input-group">
          <label for="password">كلمة المرور</label>
          <input
            id="password"
            v-model="password"
            type="password"
            class="input"
            placeholder="••••••••"
            autocomplete="current-password"
            required
          />
        </div>

        <button type="submit" class="btn btn-primary btn-lg" style="width: 100%" :disabled="loading">
          <span v-if="loading" class="spinner" style="width: 18px; height: 18px;"></span>
          <span v-else>تسجيل الدخول</span>
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background:
    radial-gradient(ellipse at top right, rgba(99, 102, 241, 0.08) 0%, transparent 50%),
    radial-gradient(ellipse at bottom left, rgba(139, 92, 246, 0.06) 0%, transparent 50%),
    var(--bg-primary);
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  padding: var(--space-10);
  box-shadow: var(--shadow-lg);
}

.login-header {
  text-align: center;
  margin-bottom: var(--space-8);
}

.login-logo {
  font-size: 3rem;
  margin-bottom: var(--space-4);
}

.login-header h1 {
  font-size: var(--font-2xl);
  font-weight: 700;
  margin-bottom: var(--space-2);
}

.login-header p {
  font-size: var(--font-sm);
  color: var(--text-secondary);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}
</style>
