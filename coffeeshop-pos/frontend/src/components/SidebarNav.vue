<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  active: string
  pendingWebOrders: number
  userRole: string  // 'admin', 'cashier', or 'kitchen'
  userName: string
  kitchenModeEnabled: boolean
}>()

const emit = defineEmits<{
  navigate: [view: string]
  logout: []
}>()

const allNavItems = [
  { id: 'pos', icon: '📋', label: 'نقطة البيع', roles: ['admin', 'cashier', 'dev'] },
  { id: 'web-orders', icon: '🌐', label: 'طلبات الويب', roles: ['admin', 'cashier', 'dev'] },
  { id: 'kitchen', icon: '🍳', label: 'المطبخ', roles: ['admin', 'cashier', 'kitchen', 'dev'] },
  { id: 'order-history', icon: '📜', label: 'سجل الطلبات', roles: ['admin', 'dev'] },
  { id: 'reports', icon: '📊', label: 'التقارير', roles: ['admin', 'dev'] },
  { id: 'inventory', icon: '📦', label: 'المخزون', roles: ['admin', 'dev'] },
  { id: 'settings', icon: '⚙️', label: 'الإعدادات', roles: ['admin', 'dev'] },
]

const navItems = computed(() =>
  allNavItems.filter(item => {
    if (item.id === 'kitchen' && !props.kitchenModeEnabled) return false
    return item.roles.includes(props.userRole)
  })
)
</script>

<template>
  <nav class="sidebar-nav" style="--wails-draggable: drag">
    <div class="nav-items">
      <button
        v-for="item in navItems"
        :key="item.id"
        class="nav-item"
        :class="{ active: active === item.id }"
        :title="item.label"
        style="--wails-draggable: no-drag"
        @click="emit('navigate', item.id)"
      >
        <span class="nav-icon">
          {{ item.icon }}
          <span
            v-if="item.id === 'web-orders' && pendingWebOrders > 0"
            class="nav-badge"
          >
            {{ pendingWebOrders }}
          </span>
        </span>
        <span class="nav-label">{{ item.label }}</span>
      </button>
    </div>

    <div class="nav-footer" style="--wails-draggable: no-drag">
      <div class="user-info">
        <span class="user-name">{{ userName }}</span>
        <span class="user-role">{{ userRole === 'admin' ? 'مدير' : userRole === 'kitchen' ? 'مطبخ' : userRole === 'dev' ? 'مطور' : 'كاشير' }}</span>
      </div>
      <button class="nav-item logout-btn" title="تسجيل خروج" @click="emit('logout')">
        <span class="nav-icon">🔒</span>
        <span class="nav-label">خروج</span>
      </button>
    </div>
  </nav>
</template>

<style scoped>
.sidebar-nav {
  width: 72px;
  min-width: 72px;
  background: var(--color-surface);
  border-left: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 56px; /* offset for header */
  justify-content: space-between;
}

.nav-items {
  display: flex;
  flex-direction: column;
  gap: var(--gap-xs);
  padding: var(--gap-md) var(--gap-xs);
  width: 100%;
}

.nav-footer {
  padding: var(--gap-sm) var(--gap-xs) var(--gap-md);
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-xs);
  border-top: 1px solid var(--color-border);
}

.user-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
  padding: var(--gap-xs);
}

.user-name {
  font-size: 0.6rem;
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 64px;
}

.user-role {
  font-size: 0.55rem;
  color: var(--color-accent);
  font-weight: var(--font-weight-semi);
}

.logout-btn {
  color: var(--color-text-dim) !important;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: var(--gap-sm) var(--gap-xs);
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family);
  user-select: none;
}

.nav-item:hover {
  background: var(--color-surface-2);
  color: var(--color-text);
}

.nav-item.active {
  background: rgba(233, 69, 96, 0.12);
  color: var(--color-accent);
}

.nav-item.active::before {
  content: '';
  position: absolute;
  right: 0;
  width: 3px;
  height: 28px;
  background: var(--color-accent);
  border-radius: 2px 0 0 2px;
}

.nav-icon {
  font-size: 1.4rem;
  line-height: 1;
  position: relative;
}

.nav-badge {
  position: absolute;
  top: -6px;
  right: -8px;
  background: var(--color-danger);
  color: white;
  font-size: 0.6rem;
  font-weight: var(--font-weight-extra);
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 3px;
  animation: badgePop 0.3s ease;
}

@keyframes badgePop {
  0% { transform: scale(0); }
  70% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

.nav-label {
  font-size: 0.65rem;
  font-weight: var(--font-weight-semi);
  white-space: nowrap;
}
</style>
