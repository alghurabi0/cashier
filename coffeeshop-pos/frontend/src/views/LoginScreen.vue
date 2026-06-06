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
  if (e.key >= '0' && e.key <= '9') onDigit(e.key)
  else if (e.key === 'Backspace') onBackspace()
  else if (e.key === 'Enter') onSubmit()
}
</script>

<template>
  <div class="login-screen" @keydown="onKeydown" tabindex="0">
    <video class="bg-video" autoplay muted loop playsinline>
      <source src="/bg-video.mp4" type="video/mp4" />
    </video>
    <div class="bg-overlay"></div>

    <div class="login-card" :class="{ shake: shaking }">
      <div class="login-header">
        <div class="logo-circle">NJ</div>
        <h1 class="login-title">NJ Coffee</h1>
        <p class="login-subtitle">أدخل رمز PIN للدخول</p>
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
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  overflow: hidden;
  outline: none;
}

.bg-video {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  z-index: 0;
}

.bg-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  z-index: 1;
}

.login-card {
  position: relative;
  z-index: 2;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  padding: 40px 48px;
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 24px;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.4);
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
  gap: 10px;
}

.logo-circle {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #c8960a, #e0aa12);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
  font-weight: 900;
  color: #0a1f12;
  box-shadow: 0 4px 20px rgba(200, 150, 10, 0.5);
}

.login-title {
  font-size: 2rem;
  font-weight: 900;
  color: #ffffff;
  letter-spacing: 2px;
}

.login-subtitle {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.9rem;
}

.pin-dots {
  display: flex;
  gap: 12px;
}

.pin-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.4);
  transition: all 0.15s ease;
}

.pin-dot.filled {
  background: #c8960a;
  border-color: #c8960a;
  transform: scale(1.2);
  box-shadow: 0 0 10px rgba(200, 150, 10, 0.6);
}

.login-error {
  color: #ff6b6b;
  font-size: 0.85rem;
}

.pin-pad {
  display: grid;
  grid-template-columns: repeat(3, 64px);
  gap: 10px;
}

.pad-btn {
  width: 64px;
  height: 56px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  font-family: var(--font-family);
  font-size: 1.3rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.pad-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.pad-btn:active {
  transform: scale(0.92);
}

.pad-backspace {
  color: rgba(255, 255, 255, 0.7);
}

.pad-enter {
  background: linear-gradient(135deg, #c8960a, #e0aa12);
  color: #0a1f12;
  border-color: transparent;
  font-size: 1.4rem;
  font-weight: 900;
}

.pad-enter:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pad-enter:hover:not(:disabled) {
  filter: brightness(1.1);
}
</style>