import { createRouter, createWebHistory } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
    },
    {
      path: '/tenants',
      name: 'tenants',
      component: () => import('../views/TenantsView.vue'),
    },
    {
      path: '/tenants/new',
      name: 'tenant-create',
      component: () => import('../views/CreateTenantView.vue'),
    },
    {
      path: '/tenants/:id',
      name: 'tenant-detail',
      component: () => import('../views/TenantDetailView.vue'),
    },
    {
      path: '/app-release',
      name: 'app-release',
      component: () => import('../views/AppReleaseView.vue'),
    },
  ],
})

// Auth guard
router.beforeEach((to) => {
  const { isAuthenticated } = useApi()
  if (!to.meta.public && !isAuthenticated()) {
    return { name: 'login' }
  }
})

export default router
