<script setup lang="ts">
import { ref } from 'vue'
import { useAuth } from '../composables/useAuth'

const { login, error, isLoading } = useAuth()

const pinDigits = ref<string[]>([])
const shaking = ref(false)
const MAX_PIN = 6

function onDigit(d: string) {
  if (pinDigits.value.length >= MAX_PIN) return
  pinDigits.value.push(d)
  // Auto-submit when 4+ digits entered and user presses enter via keypad
}

function onBackspace() {
  pinDigits.value.pop()
}

async function onSubmit() {
  if (pinDigits.value.length < 4) return
  const pin = pinDigits.value.join('')
  try {
    await login(pin)
  } catch {
    shaking.value = true
    setTimeout(() => {
      shaking.value = false
      pinDigits.value = []
    }, 600)
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key >= '0' && e.key <= '9') {
    onDigit(e.key)
  } else if (e.key === 'Backspace') {
    onBackspace()
  } else if (e.key === 'Enter') {
    onSubmit()
  }
}
</script>

<template>
  <div class="login-screen" @keydown="onKeydown" tabindex="0" ref="loginRef">
    <div class="login-card" :class="{ shake: shaking }">
      <div class="login-header">
        <span class="login-icon">☕</span>
        <h1 class="login-title">المقهى</h1>
        <p class="login-subtitle">أدخل رمز PIN</p>
      </div>

      <div class="pin-dots">
        <span
          v-for="i in MAX_PIN"
          :key="i"
          class="pin-dot"
          :class="{ filled: i <= pinDigits.length }"
        />
      </div>

      <div v-if="error" class="login-error">{{ error }}</div>

      <div class="pin-pad">
        <button v-for="d in ['1','2','3','4','5','6','7','8','9']" :key="d" class="pad-btn" @click="onDigit(d)">{{ d }}</button>
        <button class="pad-btn pad-backspace" @click="onBackspace">⌫</button>
        <button class="pad-btn" @click="onDigit('0')">0</button>
        <button class="pad-btn pad-enter" @click="onSubmit" :disabled="pinDigits.length < 4 || isLoading">
          {{ isLoading ? '...' : '✓' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: var(--color-bg);
  outline: none;
}

.login-card {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-lg);
  padding: var(--gap-xl);
}

.login-card.shake {
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

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-sm);
}

.login-icon {
  font-size: 3.5rem;
}

.login-title {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-extra);
}

.login-subtitle {
  color: var(--color-text-muted);
  font-size: var(--font-size-md);
}

.pin-dots {
  display: flex;
  gap: var(--gap-md);
}

.pin-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid var(--color-surface-3);
  transition: all var(--transition-fast);
}

.pin-dot.filled {
  background: var(--color-accent);
  border-color: var(--color-accent);
  transform: scale(1.15);
}

.login-error {
  color: var(--color-danger);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semi);
}

.pin-pad {
  display: grid;
  grid-template-columns: repeat(3, 64px);
  gap: var(--gap-sm);
}

.pad-btn {
  width: 64px;
  height: 56px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: 1.3rem;
  font-weight: var(--font-weight-bold);
  cursor: pointer;
  transition: all var(--transition-fast);
  user-select: none;
}

.pad-btn:active {
  transform: scale(0.92);
  background: var(--color-surface-2);
}

.pad-btn:hover {
  border-color: var(--color-surface-3);
}

.pad-backspace {
  font-size: 1.1rem;
  color: var(--color-text-muted);
}

.pad-enter {
  background: var(--color-accent);
  color: var(--color-bg);
  border-color: var(--color-accent);
  font-size: 1.4rem;
}

.pad-enter:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pad-enter:hover:not(:disabled) {
  filter: brightness(1.1);
}
</style>
