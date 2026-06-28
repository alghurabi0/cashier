<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { get, loading } = useApi()

interface PlatformStats {
  total_tenants: number
  active_tenants: number
  total_users: number
  total_devices: number
  total_orders: number
  today_orders: number
}

const stats = ref<PlatformStats | null>(null)

onMounted(async () => {
  try {
    stats.value = await get<PlatformStats>('/api/v1/admin/stats')
  } catch {
    // handled by useApi
  }
})

const statCards = [
  { key: 'total_tenants', label: 'إجمالي المستأجرين', icon: '🏪', color: '#6366f1' },
  { key: 'active_tenants', label: 'المستأجرون النشطون', icon: '✅', color: '#10b981' },
  { key: 'total_users', label: 'إجمالي المستخدمين', icon: '👥', color: '#8b5cf6' },
  { key: 'total_devices', label: 'الأجهزة المسجلة', icon: '🖥️', color: '#f59e0b' },
  { key: 'total_orders', label: 'إجمالي الطلبات', icon: '📋', color: '#ec4899' },
  { key: 'today_orders', label: 'طلبات اليوم', icon: '🔥', color: '#ef4444' },
]
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">الرئيسية</h1>
        <p class="page-subtitle">نظرة عامة على المنصة</p>
      </div>
    </div>

    <div v-if="loading" style="display: flex; justify-content: center; padding: 4rem;">
      <div class="spinner"></div>
    </div>

    <div v-else-if="stats" class="stats-grid">
      <div v-for="card in statCards" :key="card.key" class="stat-card">
        <div class="stat-icon" :style="{ background: card.color + '18', color: card.color }">
          {{ card.icon }}
        </div>
        <div class="stat-value">{{ (stats as any)[card.key] }}</div>
        <div class="stat-label">{{ card.label }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-5);
}
</style>
