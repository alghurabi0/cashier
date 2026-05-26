<script setup lang="ts">
import { ref, onMounted } from 'vue'
import MenuView from './views/MenuView.vue'
import CartView from './views/CartView.vue'
import ConfirmationView from './views/ConfirmationView.vue'

type ViewState = 'menu' | 'cart' | 'confirmation' | 'invalid'

const currentView = ref<ViewState>('menu')
const tableToken = ref('')
const tableNumber = ref('')

onMounted(() => {
  const params = new URLSearchParams(window.location.search)
  const token = params.get('token')
  if (!token) {
    currentView.value = 'invalid'
    return
  }
  tableToken.value = token
  // Extract table number from token context (we'll show it after menu loads)
  tableNumber.value = '—'
})
</script>

<template>
  <div class="app">
    <!-- Invalid token / no token -->
    <div v-if="currentView === 'invalid'" class="invalid-view">
      <div class="invalid-content">
        <span class="invalid-icon">⚠️</span>
        <h1>رابط غير صالح</h1>
        <p class="text-muted">يرجى مسح رمز QR من على الطاولة</p>
      </div>
    </div>

    <!-- Menu -->
    <MenuView
      v-else-if="currentView === 'menu'"
      :table-number="tableNumber"
      @open-cart="currentView = 'cart'"
    />

    <!-- Cart -->
    <CartView
      v-else-if="currentView === 'cart'"
      :token="tableToken"
      @back="currentView = 'menu'"
      @submitted="currentView = 'confirmation'"
    />

    <!-- Confirmation -->
    <ConfirmationView v-else-if="currentView === 'confirmation'" />
  </div>
</template>

<style scoped>
.app {
  min-height: 100dvh;
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

.invalid-icon {
  font-size: 4rem;
}

.invalid-content h1 {
  font-size: var(--font-size-2xl);
  font-weight: 800;
}
</style>
