<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { get, loading } = useApi()

interface TenantWithCounts {
  id: string
  name: string
  slug: string
  is_active: boolean
  user_count: number
  device_count: number
  created_at: string
}

const tenants = ref<TenantWithCounts[]>([])
const search = ref('')

const filteredTenants = computed(() => {
  if (!search.value) return tenants.value
  const q = search.value.toLowerCase()
  return tenants.value.filter(t =>
    t.name.toLowerCase().includes(q) || t.slug.toLowerCase().includes(q)
  )
})

onMounted(async () => {
  try {
    tenants.value = await get<TenantWithCounts[]>('/api/v1/admin/tenants')
  } catch {
    // handled by useApi
  }
})

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ar-IQ', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">المستأجرون</h1>
        <p class="page-subtitle">إدارة جميع المقاهي المسجلة في المنصة</p>
      </div>
      <router-link to="/tenants/new" class="btn btn-primary">
        ➕ مستأجر جديد
      </router-link>
    </div>

    <!-- Search -->
    <div style="margin-bottom: var(--space-6);">
      <input
        v-model="search"
        class="input"
        placeholder="🔍 بحث بالاسم أو المعرّف..."
        style="max-width: 400px;"
      />
    </div>

    <div v-if="loading" style="display: flex; justify-content: center; padding: 4rem;">
      <div class="spinner"></div>
    </div>

    <div v-else-if="filteredTenants.length === 0" class="empty-state">
      <div class="empty-icon">🏪</div>
      <p>لا يوجد مستأجرون بعد</p>
      <router-link to="/tenants/new" class="btn btn-primary">إنشاء أول مستأجر</router-link>
    </div>

    <div v-else class="card" style="padding: 0; overflow: hidden;">
      <table class="data-table">
        <thead>
          <tr>
            <th>الاسم</th>
            <th>المعرّف</th>
            <th>الحالة</th>
            <th>المستخدمون</th>
            <th>الأجهزة</th>
            <th>تاريخ الإنشاء</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="tenant in filteredTenants"
            :key="tenant.id"
            style="cursor: pointer;"
            @click="router.push(`/tenants/${tenant.id}`)"
          >
            <td style="font-weight: 600;">{{ tenant.name }}</td>
            <td>
              <code style="font-size: var(--font-xs); color: var(--text-muted); direction: ltr; display: inline-block;">
                {{ tenant.slug }}
              </code>
            </td>
            <td>
              <span :class="['badge', tenant.is_active ? 'badge-success' : 'badge-danger']">
                {{ tenant.is_active ? 'نشط' : 'معلّق' }}
              </span>
            </td>
            <td>{{ tenant.user_count }}</td>
            <td>{{ tenant.device_count }}</td>
            <td style="color: var(--text-secondary); font-size: var(--font-sm);">
              {{ formatDate(tenant.created_at) }}
            </td>
            <td>
              <span style="color: var(--text-muted);">←</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
