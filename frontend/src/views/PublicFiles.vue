<template>
  <Layout>
    <div class="page-container">

      <div class="page-header desktop-header">
        <div>
          <h2 class="page-title">{{ t.public }}</h2>
          <p class="page-sub">{{ origin }}/raw/path/to/file</p>
        </div>
        <button class="btn-settings" @click="showSettings=true" :title="t.settings">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
        </button>
      </div>

      <div class="mobile-header">
        <button class="mob-icon-btn" @click="showMobileNav = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
        <span class="mob-title">{{ t.public }}</span>
        <button class="mob-icon-btn" @click="showSettings = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
        </button>
      </div>

      <teleport to="body">
        <div v-if="showMobileNav" class="mob-drawer-mask" @click="showMobileNav = false">
          <div class="mob-drawer-left" @click.stop>
            <div class="mob-drawer-header">
              <div class="mob-drawer-brand">
                 <svg viewBox="0 0 100 100" style="width:28px;height:28px"><defs><linearGradient id="pblg" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" style="stop-color:#2563EB"/><stop offset="100%" style="stop-color:#38BDF8"/></linearGradient></defs><ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#pblg)"/><circle cx="36" cy="56" r="16" fill="url(#pblg)"/><circle cx="58" cy="50" r="20" fill="url(#pblg)"/><polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/></svg>
                 <span>CloudOne</span>
              </div>
            </div>
            <nav class="mob-drawer-nav">
              <router-link to="/files" class="mob-nav-item" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
                {{ t.myFiles }}
              </router-link>
              <router-link to="/shared" class="mob-nav-item" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
                {{ t.shared }}
              </router-link>
              <router-link to="/public-files" class="mob-nav-item mob-nav-active" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
                {{ t.public }}
              </router-link>
              <router-link to="/webdav" class="mob-nav-item" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12H3l9-9 9 9h-2"/><path d="M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"/><path d="M9 21v-6a2 2 0 012-2h2a2 2 0 012 2v6"/></svg>
                {{ t.webdav }}
              </router-link>
              <div class="mob-nav-divider"></div>
              <router-link to="/settings" class="mob-nav-item" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
                {{ t.settings }}
              </router-link>
            </nav>
          </div>
        </div>
      </teleport>

      <div class="breadcrumb">
        <button class="crumb-item" @click="navigate('/')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>
        </button>
        <template v-for="(seg, i) in pathSegments" :key="i">
          <span class="crumb-sep">›</span>
          <button class="crumb-item" @click="navigateToIndex(i)">{{ seg }}</button>
        </template>
      </div>

      <div class="content">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
        </div>
        <div v-else-if="!files.length" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
          <p>{{ t.noFiles }}</p>
        </div>
        <div v-else class="list-wrapper">
          <div class="list-header">
            <span>{{ t.name }}</span>
            <span class="col-size">{{ t.size }}</span>
            <span class="col-date">{{ t.modified }}</span>
            <span class="col-actions"></span>
          </div>
          <div v-for="file in files" :key="file.path" class="list-item" @click="handleClick(file)">
            <div class="name-cell">
              <div class="file-icon" :class="file.is_dir ? 'folder' : getExt(file.name)">
                <svg v-if="file.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </div>
              <span class="file-name" :title="file.name">{{ file.name }}</span>
            </div>
            <span class="col-size">{{ file.is_dir ? '—' : formatSize(file.size) }}</span>
            <span class="col-date">{{ formatDate(file.mod_time) }}</span>
            <div class="col-actions" @click.stop>
              <button class="action-btn" @click="copyLink(file)" title="复制链接">
                 <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg>
              </button>
              <a v-if="!file.is_dir" :href="getDownloadUrl(file)" class="action-btn" :download="file.name" title="下载">
                 <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>

    <PublicSettingsModal v-model="showSettings" />
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import Layout from '../components/Layout.vue'
import PublicSettingsModal from '../components/PublicSettingsModal.vue'
import api from '../api'
import { t } from '../i18n'

const files = ref([])
const currentPath = ref('/')
const loading = ref(false)
const showSettings = ref(false)
const showMobileNav = ref(false)

const origin = window.location.origin
const lang = window.localStorage.getItem('lang') || 'zh'

const pathSegments = computed(() => {
  return currentPath.value.split('/').filter(Boolean)
})

async function loadFiles(path) {
  loading.value = true
  try {
    const { data } = await api.get('/public/files', { params: { path } })
    files.value = data || []
    currentPath.value = path
  } catch (e) {
    files.value = []
  }
  loading.value = false
}

function navigate(path) {
  loadFiles(path)
}

function navigateToIndex(i) {
  const segs = pathSegments.value.slice(0, i + 1)
  loadFiles('/' + segs.join('/'))
}

function handleClick(file) {
  if (file.is_dir) {
    const newPath = currentPath.value === '/' 
      ? '/' + file.name 
      : currentPath.value.endsWith('/') ? currentPath.value + file.name : currentPath.value + '/' + file.name
    loadFiles(newPath)
  }
}

function getDownloadUrl(file) {
  const fullPath = currentPath.value === '/' ? '/' + file.name : currentPath.value + '/' + file.name
  return `${origin}/raw${fullPath}`
}

async function copyLink(file) {
  const url = getDownloadUrl(file)
  try {
    await navigator.clipboard.writeText(url)
    alert(lang === 'zh' ? '链接已复制到剪贴板' : 'Link copied to clipboard')
  } catch (err) {
    const el = document.createElement('textarea')
    el.value = url
    document.body.appendChild(el)
    el.select()
    document.execCommand('copy')
    document.body.removeChild(el)
    alert(lang === 'zh' ? '链接已复制到剪贴板' : 'Link copied to clipboard')
  }
}

function getExt(name) {
  const ext = name?.split('.').pop()?.toLowerCase()
  if (['jpg','jpeg','png','gif','webp','svg'].includes(ext)) return 'img'
  if (['mp4','mkv','avi','mov'].includes(ext)) return 'video'
  if (['mp3','flac','wav','aac'].includes(ext)) return 'audio'
  if (['zip','tar','gz','rar','7z'].includes(ext)) return 'archive'
  if (['pdf'].includes(ext)) return 'pdf'
  if (['md','txt','log','ini','conf','yaml','yml','toml','json','sh'].includes(ext)) return 'text'
  if (['js','ts','py','go','java','rs','c','cpp'].includes(ext)) return 'code'
  return 'default'
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const u = ['B','KB','MB','GB'], i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + u[i]
}

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { year:'numeric', month:'short', day:'numeric', hour:'2-digit', minute:'2-digit' })
}

onMounted(() => loadFiles('/'))
</script>

<style scoped>
.page-container { flex: 1; display: flex; flex-direction: column; height: 100%; background: #fff; overflow: hidden; }

/* 头部 */
.page-header { padding: 20px 28px 16px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--gray-100); background: white; }
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); }
.page-sub { font-size: 12px; color: var(--gray-400); font-family: 'JetBrains Mono', monospace; margin-top: 2px; }

.btn-settings { width: 36px; height: 36px; border-radius: 10px; border: 1.5px solid var(--gray-100); background: white; color: var(--gray-500); cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all .2s; }
.btn-settings:hover { border-color: var(--blue-400); color: var(--blue-500); background: var(--blue-50); }
.btn-settings svg { width: 18px; height: 18px; }

.mobile-header { display: none; }

/* 面包屑 */
.breadcrumb { padding: 12px 28px; display: flex; align-items: center; gap: 4px; border-bottom: 1px solid var(--gray-50); background: #fafafa; }
.crumb-item { padding: 4px 8px; border: none; background: none; font-size: 13px; color: var(--blue-600); font-weight: 500; cursor: pointer; border-radius: 6px; transition: background .2s; }
.crumb-item:hover { background: var(--blue-50); }
.crumb-sep { color: var(--gray-300); font-size: 14px; }

/* 内容区 */
.content { flex: 1; overflow-y: auto; padding: 0 28px 40px; }
.loading-state { display: flex; justify-content: center; padding: 60px; }
.spinner { width: 32px; height: 32px; border: 3px solid var(--gray-100); border-top-color: var(--blue-500); border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 80px 20px; color: var(--gray-300); }
.empty-state svg { width: 64px; height: 64px; margin-bottom: 16px; stroke-width: 1; }
.empty-state p { font-size: 15px; color: var(--gray-400); }

/* 列表样式 */
.list-wrapper { width: 100%; }
.list-header { display: grid; grid-template-columns: 1fr 100px 180px 100px; padding: 12px 16px; font-size: 12px; font-weight: 600; color: var(--gray-400); border-bottom: 1px solid var(--gray-100); position: sticky; top: 0; background: white; z-index: 1; }
.list-item { display: grid; grid-template-columns: 1fr 100px 180px 100px; padding: 10px 16px; border-bottom: 1px solid var(--gray-50); cursor: pointer; transition: background .1s; align-items: center; border-radius: 8px; margin: 2px 0; }
.list-item:hover { background: var(--blue-50); }

.name-cell { display: flex; align-items: center; gap: 12px; min-width: 0; }
.file-icon { width: 32px; height: 32px; border-radius: 8px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.file-icon svg { width: 16px; height: 16px; }
.file-icon.folder { background: rgba(245,158,11,.15); color: #F59E0B; }
.file-icon.img { background: rgba(16,185,129,.12); color: #10B981; }
.file-icon.video { background: rgba(139,92,246,.12); color: #8B5CF6; }
.file-icon.audio { background: rgba(244,63,94,.12); color: #F43F5E; }
.file-icon.archive { background: rgba(245,158,11,.12); color: #F59E0B; }
.file-icon.pdf { background: rgba(239,68,68,.12); color: #EF4444; }
.file-icon.code, .file-icon.text { background: rgba(59,130,246,.12); color: #3B82F6; }
.file-icon.default { background: var(--gray-100); color: var(--gray-400); }

.file-name { font-size: 14px; font-weight: 500; color: var(--gray-700); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.list-item:hover .file-name { color: var(--blue-600); }
.col-size, .col-date { font-size: 13px; color: var(--gray-400); }

.col-actions { display: flex; align-items: center; justify-content: flex-end; gap: 6px; opacity: 0; transition: opacity .2s; }
.list-item:hover .col-actions { opacity: 1; }
.action-btn { width: 32px; height: 32px; border: none; background: white; color: var(--gray-500); border-radius: 8px; display: flex; align-items: center; justify-content: center; cursor: pointer; transition: all .2s; border: 1px solid var(--gray-100); }
.action-btn:hover { background: var(--blue-600); color: white; border-color: var(--blue-600); }
.action-btn svg { width: 14px; height: 14px; }

/* 移动端抽屉样式 */
.mob-drawer-mask { position: fixed; inset: 0; background: rgba(15,23,42,.4); backdrop-filter: blur(4px); z-index: 500; display: flex; }
.mob-drawer-left { width: 280px; background: white; height: 100%; box-shadow: 4px 0 24px rgba(0,0,0,.1); display: flex; flex-direction: column; animation: slideInLeft .3s ease-out; }
@keyframes slideInLeft { from { transform: translateX(-100%); } to { transform: translateX(0); } }
.mob-drawer-header { padding: 24px 20px; border-bottom: 1px solid var(--gray-100); }
.mob-drawer-brand { display: flex; align-items: center; gap: 12px; }
.mob-drawer-brand span { font-size: 18px; font-weight: 800; background: var(--primary-gradient); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.mob-drawer-nav { flex: 1; padding: 12px; display: flex; flex-direction: column; gap: 4px; overflow-y: auto; }
.mob-nav-item { display: flex; align-items: center; gap: 12px; padding: 12px 16px; border-radius: 12px; text-decoration: none; font-size: 15px; font-weight: 500; color: var(--gray-600); transition: all .2s; }
.mob-nav-item svg { width: 20px; height: 20px; flex-shrink: 0; }
.mob-nav-item:hover, .mob-nav-item:active { background: var(--blue-50); color: var(--blue-600); }
.mob-nav-active { background: var(--blue-50); color: var(--blue-600); font-weight: 600; }
.mob-nav-divider { height: 1px; background: var(--gray-100); margin: 8px 4px; }

/* 移动端适配 */
@media (max-width: 768px) {
  .desktop-header { display: none; }
  .mobile-header { display: flex; padding: 10px 14px; min-height: 52px; gap: 10px; align-items: center; justify-content: space-between; background: white; border-bottom: 1px solid var(--gray-100); position: sticky; top: 0; z-index: 10; }
  .mob-icon-btn { width: 40px; height: 40px; border: none; background: none; display: flex; align-items: center; justify-content: center; color: var(--gray-600); border-radius: 10px; }
  .mob-icon-btn svg { width: 20px; height: 20px; }
  .mob-title { font-size: 16px; font-weight: 700; color: var(--gray-800); }

  .breadcrumb { padding: 8px 16px; background: white; }
  .content { padding: 12px 16px 20px; }

  .list-header { grid-template-columns: 1fr 60px 40px; }
  .list-header .col-date { display: none; }
  .list-item { grid-template-columns: 1fr 60px 40px; padding: 10px 12px; }
  .list-item .col-date { display: none; }
  .col-actions { opacity: 1; }
  .action-btn { border: none; background: transparent; }
  .action-btn[download] { display: none; }
}

@media (max-width: 480px) {
  .list-header .col-size { display: none; }
  .list-header { grid-template-columns: 1fr 40px; }
  .list-item .col-size { display: none; }
  .list-item { grid-template-columns: 1fr 40px; }
}
</style>
