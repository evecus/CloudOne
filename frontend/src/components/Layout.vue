<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="sidebar-top">
        <router-link to="/files" class="brand">
          <div class="brand-icon">
            <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
              <defs>
                <linearGradient id="slg" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" style="stop-color:#2563EB"/>
                  <stop offset="100%" style="stop-color:#0EA5E9"/>
                </linearGradient>
              </defs>
              <ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#slg)"/>
              <circle cx="36" cy="56" r="16" fill="url(#slg)"/>
              <circle cx="58" cy="50" r="20" fill="url(#slg)"/>
              <polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/>
            </svg>
          </div>
          <span class="brand-name">CloudOne</span>
        </router-link>
      </div>

      <nav class="nav">
        <router-link to="/files" class="nav-item" :class="{ active: $route.path.startsWith('/files') }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
          {{ t.myFiles }}
        </router-link>
        <router-link to="/shared" class="nav-item" :class="{ active: $route.path === '/shared' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
          {{ t.shared }}
        </router-link>
        <router-link to="/public-files" class="nav-item" :class="{ active: $route.path === '/public-files' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
          {{ t.public }}
        </router-link>
        <router-link to="/webdav" class="nav-item" :class="{ active: $route.path === '/webdav' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12H3l9-9 9 9h-2"/><path d="M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"/><path d="M9 21v-6a2 2 0 012-2h2a2 2 0 012 2v6"/></svg>
          {{ t.webdav }}
        </router-link>
      </nav>

      <div class="sidebar-bottom">
        <div class="user-info">
          <div class="avatar">{{ auth.user?.username?.[0]?.toUpperCase() || '?' }}</div>
          <span class="user-name">{{ auth.user?.username }}</span>
          <button class="icon-btn" @click="showSettings = true" :title="t.settings">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
          </button>
        </div>
      </div>
      <SettingsModal v-model="showSettings" />
    </aside>

    <main class="main"><slot /></main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { t } from '../i18n'
import api from '../api'
import SettingsModal from './SettingsModal.vue'

const router = useRouter()
const auth = useAuthStore()
const showSettings = ref(false)

function doLogout() { auth.logout(); router.push('/login') }
</script>

<style scoped>
.layout { display: flex; height: 100vh; overflow: hidden; background: var(--gray-50); }

.sidebar { width: 200px; background: white; border-right: 1px solid var(--gray-100); display: flex; flex-direction: column; flex-shrink: 0; box-shadow: 2px 0 12px rgba(0,0,0,0.03); }

.sidebar-top { padding: 20px 16px 16px; }

.brand { display: flex; align-items: center; justify-content: center; gap: 10px; text-decoration: none; }

.brand-icon { width: 36px; height: 36px; filter: drop-shadow(0 2px 8px rgba(37,99,235,0.25)); }

.brand-name { font-size: 17px; font-weight: 700; background: var(--primary-gradient); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }

.nav { flex: 1; padding: 8px 10px; display: flex; flex-direction: column; gap: 2px; }

.nav-item { display: flex; align-items: center; justify-content: center; gap: 10px; padding: 10px; border-radius: var(--radius-sm); text-decoration: none; color: var(--gray-500); font-size: 13px; font-weight: 500; transition: var(--transition); }

.nav-item svg { width: 17px; height: 17px; flex-shrink: 0; }

.nav-item:hover { background: var(--blue-50); color: var(--blue-600); }

.nav-item.active { background: var(--primary-gradient-soft); color: var(--blue-600); font-weight: 600; }

.sidebar-bottom { padding: 14px; border-top: 1px solid var(--gray-100); }

.user-info { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }

.avatar { width: 34px; height: 34px; border-radius: 50%; background: var(--primary-gradient); color: white; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 600; flex-shrink: 0; }

.user-name { font-size: 12px; font-weight: 600; color: var(--gray-700); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.icon-btn { width: 34px; height: 34px; border-radius: var(--radius-sm); border: none; background: transparent; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--gray-400); transition: var(--transition); text-decoration: none; }

.icon-btn svg { width: 17px; height: 17px; }

.icon-btn:hover { background: var(--gray-100); color: var(--gray-700); }

.main { flex: 1; overflow: auto; display: flex; flex-direction: column; }

/* 移动端隐藏侧边栏，main 全屏 */
@media (max-width: 768px) {
  .sidebar { display: none; }
  .layout { flex-direction: column; }
  .main { flex: 1; overflow: auto; }
}
</style>
