<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuth, type LocalUser } from '../composables/useAuth'

const { loginUser, error, isLoading, listUsers } = useAuth()

const users = ref<LocalUser[]>([])
const selectedUser = ref<LocalUser | null>(null)
const pinDigits = ref<string[]>([])
const shaking = ref(false)
const MAX_PIN = 4

const roleLabels: Record<string, string> = {
  admin: 'مدير',
  cashier: 'كاشير',
  kitchen: 'مطبخ',
  dev: 'مطوّر',
}

const roleColors: Record<string, string> = {
  admin: '#c8960a',
  cashier: '#5cb85c',
  kitchen: '#e67e22',
  dev: '#9b59b6',
}

function userInitials(name: string): string {
  const words = name.trim().split(/\s+/)
  if (words.length >= 2) return words[0][0] + words[1][0]
  return name.slice(0, 2)
}

function selectUser(user: LocalUser) {
  selectedUser.value = user
  pinDigits.value = []
  error.value = null
}

function goBack() {
  selectedUser.value = null
  pinDigits.value = []
  error.value = null
}

function onDigit(d: string) {
  if (pinDigits.value.length >= MAX_PIN) return
  pinDigits.value.push(d)
}

function onBackspace() {
  pinDigits.value.pop()
}

async function onSubmit() {
  if (!selectedUser.value || pinDigits.value.length < 4) return
  const pin = pinDigits.value.join('')
  try {
    await loginUser(selectedUser.value.id, pin)
  } catch {
    shaking.value = true
    setTimeout(() => {
      shaking.value = false
      pinDigits.value = []
    }, 600)
  }
}

function onKeydown(e: KeyboardEvent) {
  if (!selectedUser.value) return
  if (e.key >= '0' && e.key <= '9') onDigit(e.key)
  else if (e.key === 'Backspace') onBackspace()
  else if (e.key === 'Enter') onSubmit()
}

onMounted(async () => {
  users.value = await listUsers()
})

const videoSrc = ref('')
onMounted(async () => {
  try {
    const ConfigStoreService = await import('../../bindings/coffeeshop-pos/internal/service/configstoreservice')
    const url = await ConfigStoreService.Get('intro_video_url')
    if (url) videoSrc.value = url
  } catch { /* no config available */ }
})
</script>

<template>
  <div class="login-screen" @keydown="onKeydown" tabindex="0">
    <video v-if="videoSrc" class="bg-video" autoplay muted loop playsinline>
      <source :src="videoSrc" type="video/mp4" />
    </video>
    <div class="bg-overlay"></div>

    <!-- Step 1: User Selection -->
    <div v-if="!selectedUser" class="login-card user-select-card">
      <div class="login-header">
        <p class="login-subtitle">اختر حسابك</p>
      </div>

      <div class="user-grid">
        <button
          v-for="user in users"
          :key="user.id"
          class="user-tile"
          @click="selectUser(user)"
        >
          <div class="user-avatar" :style="{ background: roleColors[user.role] || '#5cb85c' }">
            {{ userInitials(user.name_ar) }}
          </div>
          <span class="user-tile-name">{{ user.name_ar }}</span>
          <span class="user-tile-role">{{ roleLabels[user.role] || user.role }}</span>
        </button>
      </div>
    </div>

    <!-- Step 2: PIN Entry -->
    <div v-else class="login-card" :class="{ shake: shaking }">
      <div class="login-header">
        <button class="back-btn" @click="goBack">→ رجوع</button>
        <div class="user-avatar user-avatar-lg" :style="{ background: roleColors[selectedUser.role] || '#5cb85c' }">
          {{ userInitials(selectedUser.name_ar) }}
        </div>
        <h1 class="login-title">{{ selectedUser.name_ar }}</h1>
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

.user-select-card {
  min-width: 360px;
  max-width: 520px;
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
  width: 100%;
}

.back-btn {
  align-self: flex-end;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.6);
  font-family: var(--font-family);
  font-size: 0.85rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: all 0.15s ease;
}

.back-btn:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}

.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 16px;
  width: 100%;
}

.user-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px 12px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: var(--font-family);
}

.user-tile:hover {
  background: rgba(255, 255, 255, 0.14);
  border-color: rgba(255, 255, 255, 0.25);
  transform: translateY(-2px);
}

.user-avatar {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  font-weight: 800;
  color: #fff;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.user-avatar-lg {
  width: 72px;
  height: 72px;
  font-size: 1.5rem;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4);
}

.user-tile-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: #ffffff;
}

.user-tile-role {
  font-size: 0.7rem;
  color: rgba(255, 255, 255, 0.5);
}

.login-title {
  font-size: 1.6rem;
  font-weight: 900;
  color: #ffffff;
  letter-spacing: 1px;
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
