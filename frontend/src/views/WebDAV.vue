<template>
  <Layout>
    <div class="webdav-page">

      <!-- ===== 桌面端顶部栏 ===== -->
      <div class="page-header">
        <h2 class="page-title">{{ t.webdav }}</h2>
      </div>

      <!-- ===== 移动端顶部栏 ===== -->
      <div class="mobile-header">
        <button class="mob-icon-btn" @click="showMobileNav = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
        <span class="mob-title">{{ t.webdav }}</span>
        <div style="width:40px"></div>
      </div>

      <!-- 移动端导航抽屉 -->
      <teleport to="body">
        <div v-if="showMobileNav" class="mob-drawer-mask" @click="showMobileNav = false">
          <div class="mob-drawer-left" @click.stop>
            <div class="mob-drawer-header">
              <div class="mob-drawer-brand">
                <svg viewBox="0 0 100 100" style="width:28px;height:28px"><defs><linearGradient id="wdlg" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" style="stop-color:#2563EB"/><stop offset="100%" style="stop-color:#38BDF8"/></linearGradient></defs><ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#wdlg)"/><circle cx="36" cy="56" r="16" fill="url(#wdlg)"/><circle cx="58" cy="50" r="20" fill="url(#wdlg)"/><polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/></svg>
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
              <router-link to="/public-files" class="mob-nav-item" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
                {{ t.public }}
              </router-link>
              <router-link to="/webdav" class="mob-nav-item active" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12H3l9-9 9 9h-2"/><path d="M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"/><path d="M9 21v-6a2 2 0 012-2h2a2 2 0 012 2v6"/></svg>
                {{ t.webdav }}
              </router-link>
              <div class="mob-drawer-divider"></div>
              <router-link to="/settings" class="mob-nav-item" @click="showMobileNav = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
                {{ t.settings }}
              </router-link>
            </nav>
          </div>
        </div>
      </teleport>

      <!-- ===== 未启用提示 ===== -->
      <div v-if="!settings.webdav_enabled" class="disabled-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 12H3l9-9 9 9h-2"/><path d="M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"/></svg>
        <p>{{ t.webdavDisabled }}</p>
        <button class="btn-go-settings" @click="showSettings = true">{{ t.settings }}</button>
      </div>

      <!-- ===== 已启用内容 ===== -->
      <div v-else class="webdav-content">

        <!-- 连接信息卡片 -->
        <div class="info-card">
          <h3 class="card-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            {{ t.webdavConnInfo }}
          </h3>
          <div class="info-rows">
            <div class="info-row">
              <span class="info-label">{{ t.webdavAddress }}</span>
              <div class="info-value-wrap">
                <code class="info-value">{{ davUrl }}</code>
                <button class="copy-btn" @click="copyText(davUrl)" :class="{ copied: copied }">
                  <svg v-if="!copied" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
                  {{ copied ? t.webdavCopied : '' }}
                </button>
              </div>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t.webdavUsername }}</span>
              <code class="info-value">{{ settings.webdav_username || auth.user?.username }}</code>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t.webdavPasswordLabel }}</span>
              <span class="info-value info-tip">{{ t.webdavPasswordValue }}</span>
            </div>
          </div>
        </div>

        <!-- 客户端连接方式 -->
        <div class="info-card">
          <h3 class="card-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
            {{ t.webdavClients }}
          </h3>
          <ul class="client-list">
            <li>🍎 {{ t.webdavMacOS }}</li>
            <li>🪟 {{ t.webdavWindows }}</li>
            <li>📱 {{ t.webdavMobile }}</li>
            <li>☁️ {{ t.webdavRclone }}</li>
          </ul>
        </div>

        <!-- 文件浏览器 -->
        <div class="file-card">
          <div class="file-card-header">
            <h3 class="card-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
              {{ t.webdavFileBrowser }}
            </h3>
            <!-- 面包屑 -->
            <div class="breadcrumb">
              <button class="crumb-btn" @click="navigate('/')">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>
              </button>
              <template v-for="(seg, i) in pathSegments" :key="i">
                <span class="crumb-sep">›</span>
                <button class="crumb-btn" @click="navigateToIndex(i)">{{ seg }}</button>
              </template>
            </div>
          </div>

          <div v-if="loading" class="file-loading"><div class="spinner"></div></div>
          <div v-else-if="!files.length" class="file-empty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
            <p>{{ t.noFiles }}</p>
          </div>
          <div v-else class="file-list">
            <div class="file-list-header">
              <span>{{ t.name }}</span>
              <span class="col-size">{{ t.size }}</span>
              <span class="col-date">{{ t.modified }}</span>
            </div>
            <div
              v-for="file in files"
              :key="file.path"
              class="file-row"
              @click="handleClick(file)"
            >
              <div class="file-name-cell">
                <div class="file-icon" :class="file.is_dir ? 'folder' : getExt(file.name)">
                  <svg v-if="file.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                </div>
                <span class="file-name">{{ file.name }}</span>
              </div>
              <span class="col-size">{{ file.is_dir ? '—' : formatSize(file.size) }}</span>
              <span class="col-date">{{ formatDate(file.mod_time) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <SettingsModal v-model="showSettings" @update:modelValue="onSettingsClose" />
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Layout from '../components/Layout.vue'
import SettingsModal from '../components/SettingsModal.vue'
import api from '../api'
import { t } from '../i18n'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const settings = ref({ webdav_enabled: false, webdav_sub_path: '', webdav_username: '', webdav_password: '' })
const files = ref([])
const currentPath = ref('/')
const loading = ref(false)
const copied = ref(false)
const showMobileNav = ref(false)
const showSettings = ref(false)

async function onSettingsClose(val) {
  if (!val) {
    // 弹窗关闭时重新加载WebDAV设置
    await loadSettings()
  }
}

const davUrl = computed(() => `${window.location.origin}/dav/`)

const pathSegments = computed(() => {
  return currentPath.value.split('/').filter(Boolean)
})

async function loadSettings() {
  const { data } = await api.get('/webdav/settings')
  settings.value = data
  if (data.webdav_enabled) {
    loadFiles('/')
  }
}

async function loadFiles(path) {
  loading.value = true
  try {
    const subPath = settings.value.webdav_sub_path || 'webdav'
    const apiPath = path === '/' ? '/' + subPath : '/' + subPath + path
    const { data } = await api.get('/files', { params: { path: apiPath } })
    // 后端返回 { files: [...], path: "..." } 或直接数组，兼容两种
    const list = Array.isArray(data) ? data : (data.files || [])
    files.value = list.map(f => ({
      ...f,
      path: f.path.replace(new RegExp('^/' + subPath), '') || '/'
    }))
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
      : currentPath.value + '/' + file.name
    loadFiles(newPath)
  }
}

async function copyText(text) {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
    } else {
      // HTTP环境fallback
      const el = document.createElement('textarea')
      el.value = text
      el.style.position = 'fixed'
      el.style.opacity = '0'
      document.body.appendChild(el)
      el.focus()
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
    }
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  } catch (e) {
    console.error('复制失败', e)
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

onMounted(loadSettings)
</script>

<style scoped>
.webdav-page { flex: 1; display: flex; flex-direction: column; overflow: hidden; }

/* 桌面端顶部栏 */
.page-header {
  padding: 20px 28px 16px;
  border-bottom: 1px solid var(--gray-100);
  background: white;
}
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); }

/* 移动端顶部栏 */
.mobile-header { display: none; }

.disabled-state {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 16px;
  color: var(--gray-400);
}
.disabled-state svg { width: 56px; height: 56px; }
.disabled-state p { font-size: 15px; font-weight: 500; color: var(--gray-500); }
.btn-go-settings {
  padding: 10px 24px; background: var(--primary-gradient);
  color: white; border: none; border-radius: var(--radius-sm);
  font-size: 14px; font-weight: 600; cursor: pointer;
  text-decoration: none; display: inline-block;
  box-shadow: var(--primary-shadow);
}

.webdav-content {
  flex: 1; overflow-y: auto;
  padding: 24px 28px; display: flex; flex-direction: column; gap: 20px;
}

/* 信息卡片 */
.info-card {
  background: white; border: 1px solid var(--gray-100);
  border-radius: var(--radius-lg); padding: 20px 24px;
  box-shadow: var(--shadow-sm);
}
.card-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 14px; font-weight: 600; color: var(--gray-700);
  margin-bottom: 16px;
}
.card-title svg { width: 16px; height: 16px; color: var(--blue-500); }

.info-rows { display: flex; flex-direction: column; gap: 12px; }
.info-row { display: flex; align-items: flex-start; gap: 12px; flex-wrap: wrap; }
.info-label {
  min-width: 80px; font-size: 12px; font-weight: 600;
  color: var(--gray-400); padding-top: 2px; flex-shrink: 0;
}
.info-value-wrap { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; flex-wrap: wrap; }
.info-value {
  font-family: 'JetBrains Mono', monospace; font-size: 13px;
  color: var(--blue-700); background: var(--blue-50);
  padding: 3px 10px; border-radius: 6px; word-break: break-all;
}
.info-tip { font-family: inherit; font-size: 13px; color: var(--gray-500); background: var(--gray-50); }

.copy-btn {
  display: flex; align-items: center; gap: 5px;
  padding: 4px 10px; border: 1.5px solid var(--gray-200);
  border-radius: 6px; background: white; cursor: pointer;
  font-size: 12px; font-weight: 500; color: var(--gray-500);
  transition: var(--transition); white-space: nowrap;
}
.copy-btn svg { width: 13px; height: 13px; }
.copy-btn:hover { border-color: var(--blue-400); color: var(--blue-600); }
.copy-btn.copied { border-color: #10B981; color: #10B981; }

.client-list {
  list-style: none; display: flex; flex-direction: column; gap: 10px;
  padding: 0; margin: 0;
}
.client-list li { font-size: 13px; color: var(--gray-600); line-height: 1.5; }

/* 文件浏览卡片 */
.file-card {
  background: white; border: 1px solid var(--gray-100);
  border-radius: var(--radius-lg); overflow: hidden;
  box-shadow: var(--shadow-sm);
}
.file-card-header {
  padding: 16px 20px 14px;
  border-bottom: 1px solid var(--gray-100);
  display: flex; align-items: center; justify-content: space-between;
  gap: 12px; flex-wrap: wrap;
}
.file-card-header .card-title { margin-bottom: 0; }

.breadcrumb { display: flex; align-items: center; gap: 2px; flex-wrap: wrap; }
.crumb-btn {
  display: flex; align-items: center; padding: 3px 6px;
  border: none; background: none; font-size: 12px;
  color: var(--blue-600); font-weight: 500; cursor: pointer;
  border-radius: 4px; font-family: inherit;
}
.crumb-btn:hover { background: var(--blue-50); }
.crumb-sep { font-size: 12px; color: var(--gray-300); }

.file-loading { display: flex; justify-content: center; padding: 40px; }
.spinner { width: 28px; height: 28px; border: 3px solid var(--gray-200); border-top-color: var(--blue-500); border-radius: 50%; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.file-empty {
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; padding: 48px 20px; gap: 12px; color: var(--gray-300);
}
.file-empty svg { width: 48px; height: 48px; }
.file-empty p { font-size: 14px; color: var(--gray-400); }

.file-list-header {
  display: grid; grid-template-columns: 1fr 80px 140px;
  padding: 8px 16px; font-size: 12px; font-weight: 600;
  color: var(--gray-400); background: var(--gray-50);
  border-bottom: 1px solid var(--gray-100);
}
.file-row {
  display: grid; grid-template-columns: 1fr 80px 140px;
  padding: 10px 16px; border-bottom: 1px solid var(--gray-50);
  cursor: pointer; transition: background 0.1s; align-items: center;
}
.file-row:last-child { border-bottom: none; }
.file-row:hover { background: var(--blue-50); }

.file-name-cell { display: flex; align-items: center; gap: 10px; min-width: 0; }
.file-icon {
  width: 30px; height: 30px; border-radius: 7px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.file-icon svg { width: 15px; height: 15px; }
.file-icon.folder { background: rgba(245,158,11,.15); color: #F59E0B; }
.file-icon.img    { background: rgba(16,185,129,.12); color: #10B981; }
.file-icon.video  { background: rgba(139,92,246,.12); color: #8B5CF6; }
.file-icon.audio  { background: rgba(244,63,94,.12); color: #F43F5E; }
.file-icon.archive { background: rgba(245,158,11,.12); color: #F59E0B; }
.file-icon.pdf    { background: rgba(239,68,68,.12); color: #EF4444; }
.file-icon.code,
.file-icon.text   { background: rgba(59,130,246,.12); color: #3B82F6; }
.file-icon.default { background: var(--gray-100); color: var(--gray-400); }

.file-name { font-size: 13px; font-weight: 500; color: var(--gray-700); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.file-row:hover .file-name { color: var(--blue-600); }
.col-size, .col-date { font-size: 12px; color: var(--gray-400); }

/* 移动端抽屉（复用 Files.vue 样式变量） */
.mob-drawer-mask {
  position: fixed; inset: 0; background: rgba(15,23,42,.4);
  backdrop-filter: blur(3px); z-index: 400; display: flex;
}
.mob-drawer-left {
  width: 260px; background: white; height: 100%;
  box-shadow: 4px 0 24px rgba(15,23,42,.15);
  display: flex; flex-direction: column;
  animation: slideInLeft .2s cubic-bezier(.4,0,.2,1);
}
@keyframes slideInLeft { from { transform: translateX(-100%); } to { transform: translateX(0); } }
.mob-drawer-header { padding: 20px 16px 14px; border-bottom: 1px solid var(--gray-100); }
.mob-drawer-brand { display: flex; align-items: center; gap: 10px; }
.mob-drawer-brand span { font-size: 17px; font-weight: 700; background: var(--primary-gradient); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.mob-drawer-nav { flex: 1; padding: 12px 10px; display: flex; flex-direction: column; gap: 2px; overflow-y: auto; }
.mob-nav-item {
  display: flex; align-items: center; gap: 12px; padding: 13px 14px;
  border-radius: 10px; text-decoration: none; font-size: 15px; font-weight: 500;
  color: var(--gray-600); border: none; background: none;
  font-family: inherit; cursor: pointer; transition: background .15s, color .15s; text-align: left;
}
.mob-nav-item svg { width: 20px; height: 20px; flex-shrink: 0; }
.mob-nav-item:hover, .mob-nav-item:active { background: var(--blue-50); color: var(--blue-600); }
.mob-nav-item.active { background: var(--blue-50); color: var(--blue-600); font-weight: 600; }
.mob-drawer-divider { height: 1px; background: var(--gray-100); margin: 8px 4px; }

@media (max-width: 768px) {
  .page-header { display: none; }

  .mobile-header {
    display: flex; align-items: center; justify-content: space-between;
    padding: 10px 12px; background: white;
    border-bottom: 1px solid var(--gray-100);
    position: sticky; top: 0; z-index: 10;
  }
  .mob-icon-btn {
    width: 40px; height: 40px; border: none; background: none;
    cursor: pointer; display: flex; align-items: center; justify-content: center;
    color: var(--gray-600); border-radius: 8px;
  }
  .mob-icon-btn svg { width: 20px; height: 20px; }
  .mob-title { font-size: 16px; font-weight: 700; color: var(--gray-800); }

  .webdav-content { padding: 16px 12px; gap: 14px; }
  .info-card { padding: 16px; }
  .info-row { flex-direction: column; gap: 6px; }
  .info-label { min-width: unset; }
  .info-value-wrap { width: 100%; }

  .file-card-header { padding: 12px 14px 10px; flex-direction: column; align-items: flex-start; gap: 8px; }
  .file-list-header { grid-template-columns: 1fr 60px; }
  .file-list-header .col-date { display: none; }
  .file-row { grid-template-columns: 1fr 60px; padding: 10px 14px; }
  .file-row .col-date { display: none; }
  .file-name { font-size: 14px; }
}

@media (max-width: 480px) {
  .file-list-header { grid-template-columns: 1fr; }
  .file-list-header .col-size { display: none; }
  .file-row { grid-template-columns: 1fr; }
  .file-row .col-size { display: none; }
}
</style>
