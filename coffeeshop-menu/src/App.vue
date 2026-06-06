<script setup lang="ts">
import { ref, onMounted } from 'vue'
import MenuView from './views/MenuView.vue'
import CartView from './views/CartView.vue'
import ConfirmationView from './views/ConfirmationView.vue'

type ViewState = 'splash' | 'menu' | 'cart' | 'confirmation' | 'invalid'

const currentView = ref<ViewState>('splash')
const tableToken = ref('')
const tableNumber = ref('')
const videoRef = ref<HTMLVideoElement | null>(null)

const particles = Array.from({ length: 22 }, (_, i) => ({
  id: i,
  x: Math.random() * 100,
  size: 10 + Math.random() * 12,
  delay: Math.random() * 7,
  duration: 7 + Math.random() * 8,
  rotate: Math.random() * 360,
  type: i % 3 === 0 ? '🍫' : '🫘',
}))

onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  const token = params.get('token')
  if (!token) {
    currentView.value = 'invalid'
    return
  }
  tableToken.value = token
  tableNumber.value = '—'
  setTimeout(() => {
    if (videoRef.value) videoRef.value.playbackRate = 0.8
  }, 100)
})

function onVideoEnd() {
  currentView.value = 'menu'
}
</script>

<template>
  <div class="app">

    <div v-if="currentView === 'splash'" class="splash-screen" @click="currentView = 'menu'">
      <div class="particles-layer">
        <div
          v-for="p in particles"
          :key="p.id"
          class="particle"
          :style="{
            left: p.x + '%',
            fontSize: p.size + 'px',
            animationDelay: p.delay + 's',
            animationDuration: p.duration + 's',
          }"
        >{{ p.type }}</div>
      </div>

      <video ref="videoRef" class="splash-video" autoplay muted playsinline @ended="onVideoEnd">
        <source src="/intro.mp4" type="video/mp4" />
      </video>
      <div class="color-filter"></div>
      <div class="edge-top"></div>
      <div class="edge-bottom"></div>
      <div class="edge-sides"></div>
      <div class="splash-skip">اضغط للتخطي</div>
    </div>

    <div v-else-if="currentView === 'invalid'" class="invalid-view">
      <div class="invalid-content">
        <span class="invalid-icon">⚠️</span>
        <h1>رابط غير صالح</h1>
        <p class="text-muted">يرجى مسح رمز QR من على الطاولة</p>
      </div>
    </div>

    <MenuView v-else-if="currentView === 'menu'" :table-number="tableNumber" @open-cart="currentView = 'cart'" />
    <CartView v-else-if="currentView === 'cart'" :token="tableToken" @back="currentView = 'menu'" @submitted="currentView = 'confirmation'" />
    <ConfirmationView v-else-if="currentView === 'confirmation'" />
  </div>
</template>

<style scoped>
.app { min-height: 100dvh; }

.splash-screen {
  position: fixed;
  inset: 0;
  /* نفس لون خلفية الفيديو — أخضر داكن زيتوني */
  background: radial-gradient(ellipse at center, #2e3d1a 0%, #1e2d10 50%, #131f0a 100%);
  z-index: 999;
  cursor: pointer;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.particles-layer {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
}

.particle {
  position: absolute;
  bottom: -40px;
  opacity: 0;
  animation: floatUp linear infinite;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,0.4));
}

@keyframes floatUp {
  0%   { bottom: -40px; opacity: 0; transform: rotate(0deg) translateX(0); }
  10%  { opacity: 0.5; }
  50%  { transform: rotate(180deg) translateX(15px); }
  90%  { opacity: 0.3; }
  100% { bottom: 105%; opacity: 0; transform: rotate(360deg) translateX(-10px); }
}

.splash-video {
  position: relative;
  z-index: 2;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scale(0.98);
  transform-origin: center center;
}

.color-filter {
  position: absolute;
  inset: 0;
  background: rgba(100, 130, 50, 0.08);
  pointer-events: none;
  z-index: 3;
}

.edge-top {
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 20%;
  background: linear-gradient(to bottom, #1e2d10 0%, transparent 100%);
  pointer-events: none;
  z-index: 4;
}

.edge-bottom {
  position: absolute;
  bottom: 0; left: 0; right: 0;
  height: 22%;
  background: linear-gradient(to top, #1e2d10 0%, transparent 100%);
  pointer-events: none;
  z-index: 4;
}

.edge-sides {
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at center, transparent 40%, rgba(19, 31, 10, 0.9) 100%);
  pointer-events: none;
  z-index: 4;
}

.splash-skip {
  position: absolute;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  color: rgba(255, 230, 180, 0.85);
  font-size: 0.85rem;
  font-family: var(--font-family);
  background: rgba(0,0,0,0.3);
  padding: 8px 20px;
  border-radius: 999px;
  backdrop-filter: blur(8px);
  z-index: 5;
}

.invalid-view {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.invalid-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-lg);
}

.invalid-icon { font-size: 4rem; }

.invalid-content h1 {
  font-size: var(--font-size-2xl);
  font-weight: 800;
}
</style>