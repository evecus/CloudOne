<template>
  <Layout>
    <div class="page-container">

      <!-- 桌面端头部 -->
      <div class="page-header desktop-header">
        <div>
          <h2 class="page-title">{{ t.public }}</h2>
          <p class="page-sub">{{ origin }}/raw/path/to/file</p>
        </div>
        <button class="btn-settings" @click="showSettings=true" :title="t.settings">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
        </button>
      </div>

      <!-- 移动端头部 -->
      <div class="page-header mobile-header">
        <button class="mob-icon-btn" @click="showMobileNav=true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
        <h2 class="page-title">{{ t.public }}</h2>
      </div>

      <!-- 移动端导航抽屉 -->
      <teleport to="body">
        <div v-if="showMobileNav" class="mob-drawer-mask" @click="showMobileNav=false">
          <div class="mob-drawer-panel" @click.stop>
            <div class="mob-drawer-top">
              <svg viewBox="0 0 100 100" style="width:26px;height:26px"><defs><linearGradient id="pdlg" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" style="stop-color:#2563EB"/><stop offset="100%" style="stop-color:#38BDF8"/></linearGradient></defs><ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#pdlg)"/><circle cx="36" cy="56" r="16" fill="url(#pdlg)"/><circle cx="58" cy="50" r="20" fill="url(#pdlg)"/><polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/></svg>
              <span class="mob-brand-name">CloudOne</span>
            </div>
            <nav class="mob-drawer-nav">
              <router-link to="/files" class="mob-nav-item" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
                {{ t.myFiles }}
              </router-link>
              <router-link to="/shared" class="mob-nav-item" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
                {{ t.shared }}
              </router-link>
              <router-link to="/public-files" class="mob-nav-item mob-nav-active" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
                {{ t.public }}
              </router-link>
              <router-link to="/webdav" class="mob-nav-item" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12H3l9-9 9 9h-2"/><path d="M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"/><path d="M9 21v-6a2 2 0 012-2h2a2 2 0 012 2v6"/></svg>
                {{ t.webdav }}
              </router-link>
              <div class="mob-nav-divider"></div>
              <button class="mob-nav-item" @click="showSettings=true;showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
                {{ t.settings }}
              </button>
              <button class="mob-nav-item mob-nav-logout" @click="doLogout">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
                {{ t.logout }}
              </button>
            </nav>
          </div>
        </div>
      </teleport>

      <!--
        面包屑：仅在浏览公开文件夹内部时显示。
        顶层（currentPath="/"）是平铺视图，不显示面包屑。
      -->
      <div v-if="currentPath !== '/'" class="breadcrumb">
        <button class="crumb-home" @click="backToRoot">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>
          {{ t.public }}
        </button>
        <template v-for="(seg, i) in pathSegments" :key="i">
          <span class="crumb-sep">/</span>
          <button class="crumb-item" @click="navigateToIndex(i)">{{ seg }}</button>
        </template>
      </div>

      <!-- 顶层提示（平铺模式） -->
      <div v-if="currentPath === '/'" class="flat-hint">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px;flex-shrink:0;color:var(--blue-400)"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        <span>{{ lang === 'zh' ? '已直接显示所有公开的文件和文件夹，不显示上游目录' : 'All public files and folders are shown directly, without parent directories' }}</span>
      </div>

      <div class="content">
        <div v-if="loading" class="loading-state"><div class="spinner-lg"></div></div>
        <div v-else-if="!files.length" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
          <p>{{ lang === 'zh' ? '暂无公开文件' : 'No public files yet' }}</p>
          <span>{{ lang === 'zh' ? '在「我的文件」中将文件设为公开后将显示在这里' : 'Files marked public in My Files will appear here' }}</span>
        </div>
        <div v-else class="file-list">
          <div class="list-header">
            <span class="col-name">{{ t.name }}</span>
            <span class="col-path" v-if="currentPath === '/'">{{ lang === 'zh' ? '位置' : 'Location' }}</span>
            <span class="col-size">{{ t.size }}</span>
            <span class="col-date">{{ t.modified }}</span>
            <span class="col-actions"></span>
          </div>
          <div
            v-for="f in files" :key="f.path"
            class="file-row"
            :class="{ 'dir-row': f.is_dir }"
            @click="f.is_dir ? navigateTo(f.path) : null"
          >
            <div class="col-name">
              <div class="file-icon" :class="f.is_dir ? 'folder-icon' : 'file-icon-default'">
                <svg v-if="f.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </div>
              <span class="file-name">{{ f.name }}</span>
            </div>
            <!-- 顶层平铺时显示文件所在路径 -->
            <span class="col-path" v-if="currentPath === '/'">
              <span class="path-tag" :title="f.path">{{ parentDir(f.path) }}</span>
            </span>
            <span class="col-size">{{ f.is_dir ? '—' : formatSize(f.size) }}</span>
            <span class="col-date">{{ formatDate(f.mod_time) }}</span>
            <div class="col-actions" @click.stop>
              <template v-if="f.is_dir">
                <button class="act-btn" @click="navigateTo(f.path)" :title="lang === 'zh' ? '浏览' : 'Browse'">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
                </button>
              </template>
              <template v-else>
                <a :href="`/raw${f.path}`" target="_blank" class="act-btn" title="RAW">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                </a>
                <a :href="`/pub/dl?path=${encodeURIComponent(f.path)}`" :download="f.name" class="act-btn" :title="t.download">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                </a>
              </template>
              <button class="act-btn danger" @click="setPrivate(f)" :title="lang === 'zh' ? '取消公开' : 'Make Private'">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <SettingsModal v-model="showSettings" />
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Layout from '../components/Layout.vue'
import SettingsModal from '../components/SettingsModal.vue'
import api from '../api'
import { t, currentLang as lang } from '../i18n'

const files = ref([])
const currentPath = ref('/')
const loading = ref(false)
const origin = location.origin
const showSettings = ref(false)
const showMobileNav = ref(false)

const { logout } = useAuthStore()
const router = useRouter()
function doLogout() { logout(); router.push('/login') }

const pathSegments = computed(() => currentPath.value.split('/').filter(Boolean))

// 返回文件所在的父目录路径（用于顶层平铺时显示位置）
function parentDir(path) {
  const parts = path.split('/').filter(Boolean)
  if (parts.length <= 1) return '/'
  return '/' + parts.slice(0, -1).join('/')
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}
function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
}

async function load() {
  loading.value = true
  try {
    const res = await fetch(`/pub/list?path=${encodeURIComponent(currentPath.value)}`)
    const data = await res.json()
    files.value = data.files || []
  } catch {}
  loading.value = false
}

// 返回顶层（平铺视图）
function backToRoot() {
  currentPath.value = '/'
  load()
}

// 进入一个公开文件夹
function navigateTo(path) {
  currentPath.value = path
  load()
}

function navigateToIndex(i) {
  navigateTo('/' + pathSegments.value.slice(0, i + 1).join('/'))
}

async function setPrivate(f) {
  await api.put('/files/visibility', { path: f.path, is_public: false })
  // 取消公开后回到顶层（平铺视图会自动更新）
  currentPath.value = '/'
  load()
}

onMounted(load)
</script>

<style scoped>
.page-container { flex: 1; display: flex; flex-direction: column; }
.page-header { padding: 20px 28px 16px; border-bottom: 1px solid var(--gray-100); background: white; display: flex; align-items: center; justify-content: space-between; flex-shrink: 0; }
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); margin: 0; }
.page-sub { font-size: 12px; color: var(--gray-400); font-family: var(--editor-font, monospace); margin-top: 4px; }

.desktop-header { display: flex; }
.mobile-header  { display: none; }

.btn-settings { width: 36px; height: 36px; border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm); background: white; display: flex; align-items: center; justify-content: center; color: var(--gray-500); cursor: pointer; transition: var(--transition); }
.btn-settings svg { width: 17px; height: 17px; }
.btn-settings:hover { border-color: var(--blue-400); color: var(--blue-600); background: var(--blue-50); }

/* 平铺提示条 */
.flat-hint { display: flex; align-items: center; gap: 7px; font-size: 12px; color: var(--gray-500); background: var(--blue-50); border-bottom: 1px solid rgba(59,130,246,.1); padding: 8px 28px; }

/* 面包屑 */
.breadcrumb { display: flex; align-items: center; gap: 4px; flex-wrap: wrap; padding: 10px 28px 0; }
.crumb-home, .crumb-item { display: flex; align-items: center; gap: 5px; background: none; border: none; color: var(--blue-600); font-size: 13px; font-weight: 500; font-family: inherit; cursor: pointer; padding: 4px 8px; border-radius: 6px; transition: var(--transition); }
.crumb-home svg { width: 14px; height: 14px; }
.crumb-home:hover, .crumb-item:hover { background: var(--blue-50); }
.crumb-sep { color: var(--gray-300); font-size: 14px; }

.content { flex: 1; padding: 16px 28px 24px; overflow: auto; }
.loading-state { display: flex; justify-content: center; padding-top: 80px; }
.spinner-lg { width: 32px; height: 32px; border: 3px solid var(--gray-200); border-top-color: var(--blue-500); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 260px; color: var(--gray-300); gap: 10px; }
.empty-state svg { width: 56px; height: 56px; }
.empty-state p { font-size: 15px; font-weight: 500; color: var(--gray-400); }
.empty-state span { font-size: 13px; color: var(--gray-300); text-align: center; }

.file-list { background: white; border-radius: var(--radius-lg); box-shadow: var(--shadow-sm); overflow: hidden; border: 1px solid var(--gray-100); }
/* 顶层平铺：有位置列；子目录浏览：无位置列 */
.list-header { display: grid; grid-template-columns: 1.2fr 0.8fr 80px 120px 100px; padding: 10px 20px; background: var(--gray-50); border-bottom: 1px solid var(--gray-100); font-size: 11px; font-weight: 600; color: var(--gray-400); text-transform: uppercase; letter-spacing: 0.5px; align-items: center; }
.file-row { display: grid; grid-template-columns: 1.2fr 0.8fr 80px 120px 100px; padding: 11px 20px; border-bottom: 1px solid var(--gray-50); align-items: center; transition: var(--transition); }
.file-row:last-child { border-bottom: none; }
.dir-row { cursor: pointer; }
.dir-row:hover { background: var(--blue-50); }
.file-row:not(.dir-row):hover { background: var(--gray-50); }

.col-name { display: flex; align-items: center; gap: 10px; min-width: 0; }
.file-icon { width: 30px; height: 30px; border-radius: 7px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.file-icon.folder-icon { background: rgba(251,191,36,0.15); color: #F59E0B; }
.file-icon.file-icon-default { background: var(--blue-50); color: var(--blue-500); }
.file-icon svg { width: 15px; height: 15px; }
.file-name { font-size: 13px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: var(--gray-700); }

.col-path { min-width: 0; }
.path-tag { display: inline-block; max-width: 100%; font-family: 'JetBrains Mono', monospace; font-size: 11px; color: var(--gray-500); background: var(--gray-100); padding: 2px 7px; border-radius: 5px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.col-size, .col-date { font-size: 12px; color: var(--gray-400); }
.col-actions { display: flex; justify-content: flex-end; gap: 2px; }
.act-btn { width: 28px; height: 28px; border: none; background: none; cursor: pointer; border-radius: 6px; display: flex; align-items: center; justify-content: center; color: var(--gray-400); transition: var(--transition); text-decoration: none; }
.act-btn svg { width: 14px; height: 14px; }
.act-btn:hover { background: var(--blue-50); color: var(--blue-600); }
.act-btn.danger:hover { background: rgba(239,68,68,0.1); color: #EF4444; }

/* 移动端汉堡 */
.mob-icon-btn { width: 38px; height: 38px; border: none; background: none; color: var(--gray-600); display: flex; align-items: center; justify-content: center; border-radius: 8px; cursor: pointer; flex-shrink: 0; }
.mob-icon-btn svg { width: 22px; height: 22px; }
.mob-icon-btn:active { background: var(--gray-100); }

/* 抽屉 */
.mob-drawer-mask { position: fixed; inset: 0; background: rgba(15,23,42,.4); backdrop-filter: blur(2px); z-index: 200; display: flex; }
.mob-drawer-panel { width: 260px; max-width: 80vw; height: 100%; background: white; display: flex; flex-direction: column; box-shadow: 4px 0 24px rgba(15,23,42,.15); animation: slideInLeft .2s ease; }
@keyframes slideInLeft { from{transform:translateX(-100%)} to{transform:translateX(0)} }
.mob-drawer-top { padding: 24px 20px 16px; border-bottom: 1px solid var(--gray-100); display: flex; align-items: center; gap: 10px; }
.mob-brand-name { font-size: 18px; font-weight: 700; background: linear-gradient(135deg,#2563EB,#38BDF8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.mob-drawer-nav { flex: 1; padding: 12px 10px; display: flex; flex-direction: column; gap: 2px; overflow-y: auto; }
.mob-nav-item { display: flex; align-items: center; gap: 12px; padding: 13px 14px; border-radius: 10px; text-decoration: none; font-size: 15px; font-weight: 500; color: var(--gray-600); border: none; background: none; font-family: inherit; cursor: pointer; transition: background .15s, color .15s; text-align: left; width: 100%; }
.mob-nav-item svg { width: 20px; height: 20px; flex-shrink: 0; }
.mob-nav-item:hover, .mob-nav-item:active { background: var(--blue-50); color: var(--blue-600); }
.mob-nav-active { background: var(--blue-50); color: var(--blue-600); font-weight: 600; }
.mob-nav-logout { color: #EF4444; }
.mob-nav-logout:hover { background: rgba(239,68,68,.08) !important; color: #EF4444 !important; }
.mob-nav-divider { height: 1px; background: var(--gray-100); margin: 8px 4px; }

/* ─── 移动端适配 ─────────────────────────────────────────── */
@media (max-width: 768px) {
  .desktop-header { display: none; }
  .mobile-header  { display: flex; padding: 10px 14px; min-height: 52px; gap: 10px; }
  .flat-hint { padding: 7px 16px; font-size: 11px; }
  .breadcrumb { padding: 8px 16px 0; }
  .content { padding: 12px 16px 20px; }

  /* 隐藏位置列和日期列，只保留 名称 + 大小 + 操作 */
  .list-header { grid-template-columns: 1fr 60px 80px !important; padding: 8px 12px; }
  .file-row    { grid-template-columns: 1fr 60px 80px !important; padding: 10px 12px; }
  .list-header > .col-path,
  .file-row    > .col-path,
  .list-header > .col-date,
  .file-row    > .col-date { display: none; }

  .file-name { font-size: 13px; }
  .col-size  { font-size: 11px; }
  .act-btn   { width: 26px; height: 26px; }
}

@media (max-width: 480px) {
  /* 极小屏：只保留 名称 + 操作 */
  .list-header { grid-template-columns: 1fr 80px !important; }
  .file-row    { grid-template-columns: 1fr 80px !important; }
  .list-header > .col-size,
  .file-row    > .col-size { display: none; }
}
</style>
