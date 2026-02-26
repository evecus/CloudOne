import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

const routes = [
  { path: '/login', component: () => import('./views/Login.vue'), meta: { public: true } },
  { path: '/', redirect: '/files' },
  { path: '/files', component: () => import('./views/Files.vue') },
  { path: '/shared', component: () => import('./views/Shared.vue') },
  { path: '/public-files', component: () => import('./views/PublicFiles.vue') },
  { path: '/webdav', component: () => import('./views/WebDAV.vue') },
  { path: '/pub', component: () => import('./views/PublicView.vue'), meta: { public: true } },
  { path: '/s/:code', component: () => import('./views/ShareView.vue'), meta: { public: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  // 已登录则跳过登录页，直接进文件页
  if (to.path === '/login' && auth.token) {
    return '/files'
  }

  // 未登录访问受保护路由，跳转到登录页
  if (!to.meta.public && !auth.token) {
    return '/login'
  }
})

export default router
