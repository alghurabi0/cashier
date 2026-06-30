<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from './composables/useApi'

const route = useRoute()
const router = useRouter()
const { isAuthenticated, clearToken } = useApi()

const showSidebar = computed(() => route.name !== 'login' && isAuthenticated())

function logout() {
  clearToken()
  router.push('/login')
}

const navItems = [
  { path: '/', label: 'الرئيسية', icon: '📊' },
  { path: '/tenants', label: 'المستأجرون', icon: '🏪' },
  { path: '/tenants/new', label: 'مستأجر جديد', icon: '➕' },
  { path: '/app-release', label: 'تحديث التطبيق', icon: '🔄' },
]
</script>

<template>
  <div class="app-layout" :class="{ 'has-sidebar': showSidebar }">
    <!-- Sidebar -->
    <aside v-if="showSidebar" class="sidebar">
      <div class="sidebar-header">
        <div class="logo">☕</div>
        <div class="logo-text">لوحة الإدارة</div>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-label">{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <button class="btn btn-secondary btn-sm" style="width: 100%" @click="logout">
          تسجيل الخروج
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
}

.app-layout.has-sidebar .main-content {
  margin-right: var(--sidebar-width);
}

/* Sidebar */
.sidebar {
  position: fixed;
  top: 0;
  right: 0;
  width: var(--sidebar-width);
  height: 100vh;
  background: var(--bg-sidebar);
  border-left: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  z-index: 100;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-6);
  border-bottom: 1px solid var(--border-color);
}

.logo {
  font-size: 1.8rem;
}

.logo-text {
  font-size: var(--font-lg);
  font-weight: 700;
  color: var(--text-primary);
}

.sidebar-nav {
  flex: 1;
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-decoration: none;
  transition: all var(--transition-fast);
  font-size: var(--font-sm);
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--accent-bg);
  color: var(--accent);
  font-weight: 500;
}

.nav-icon {
  font-size: 1.1rem;
  width: 24px;
  text-align: center;
}

.sidebar-footer {
  padding: var(--space-4);
  border-top: 1px solid var(--border-color);
}

/* Main Content */
.main-content {
  flex: 1;
  padding: var(--space-8);
  min-height: 100vh;
}
</style>
