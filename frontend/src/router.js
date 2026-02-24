import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

const routes = [
  { path: '/login', component: () => import('./views/Login.vue'), meta: { public: true } },
  { path: '/', redirect: '/files' },
  { path: '/files', component: () => import('./views/Files.vue') },
  { path: '/shared', component: () => import('./views/Shared.vue') },
  { path: '/public-files', component: () => import('./views/PublicFiles.vue'), meta: { public: true } },
  { path: '/settings', component: () => import('./views/Settings.vue') },
  { path: '/s/:code', component: () => import('./views/ShareView.vue'), meta: { public: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  // Check setup status
  if (to.path !== '/login') {
    try {
      const { data } = await fetch('/api/auth/status').then(r => r.json())
      if (!data?.setup && !data) {
        // fetch failed - just continue
      }
    } catch {}
  }

  if (!to.meta.public && !auth.token) {
    return '/login'
  }
})

export default router
