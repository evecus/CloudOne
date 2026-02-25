<template>
  <Layout>
    <div class="page-container">
      <div class="page-header">
        <div>
          <h2 class="page-title">{{ t.public }}</h2>
          <p class="page-sub">{{ origin }}/raw/path/to/file</p>
        </div>
        <button class="btn-settings" @click="showSettings = true" :title="t.settings">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
        </button>
      </div>

      <div v-if="currentPath !== '/'" class="breadcrumb">
        <button class="crumb-home" @click="navigateTo('/')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>
          {{ t.home }}
        </button>
        <template v-for="(seg, i) in pathSegments" :key="i">
          <span class="crumb-sep">/</span>
          <button class="crumb-item" @click="navigateToIndex(i)">{{ seg }}</button>
        </template>
      </div>

      <div class="content">
        <div v-if="loading" class="loading-state"><div class="spinner-lg"></div></div>

        <div v-else-if="!files.length" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
          <p>{{ lang === 'zh' ? '此目录下暂无公开文件' : 'No public files in this directory' }}</p>
          <span>{{ lang === 'zh' ? '在「我的文件」中将文件设为公开后将显示在这里' : 'Files marked public in My Files will appear here' }}</span>
        </div>

        <div v-else class="file-list">
          <div class="list-header">
            <span class="col-name">{{ t.name }}</span>
            <span class="col-size">{{ t.size }}</span>
            <span class="col-date">{{ t.modified }}</span>
            <span class="col-actions"></span>
          </div>
          <div v-for="f in files" :key="f.path" class="file-row" @click="f.is_dir ? navigateTo(f.path) : null" :class="{ 'dir-row': f.is_dir }">
            <div class="col-name">
              <div class="file-icon" :class="f.is_dir ? 'folder-icon' : 'file-icon-default'">
                <svg v-if="f.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </div>
              <span class="file-name">{{ f.name }}</span>
              <span v-if="f.is_public" class="badge-pub">{{ lang === 'zh' ? '公开' : 'Public' }}</span>
              <span v-else-if="f.is_dir" class="badge-partial">{{ lang === 'zh' ? '含公开文件' : 'Has public' }}</span>
            </div>
            <span class="col-size">{{ f.is_dir ? '—' : formatSize(f.size) }}</span>
            <span class="col-date">{{ formatDate(f.mod_time) }}</span>
            <div class="col-actions" @click.stop>
              <template v-if="f.is_dir">
                <button class="act-btn" @click="navigateTo(f.path)" :title="lang === 'zh' ? '浏览' : 'Browse'">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
                </button>
              </template>
              <template v-else>
                <!-- RAW：直接查看文件内容（类似 GitHub raw） -->
                <a :href="`/raw${f.path}`" target="_blank" class="act-btn" title="RAW">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                </a>
                <!-- 下载：强制下载 -->
                <a :href="`/pub/dl?path=${encodeURIComponent(f.path)}`" :download="f.name" class="act-btn" :title="t.download">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                </a>
              </template>
              <!-- 取消公开 -->
              <button class="act-btn danger" @click="setPrivate(f)" :title="lang === 'zh' ? '取消公开' : 'Make Private'">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
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
import Layout from '../components/Layout.vue'
import SettingsModal from '../components/SettingsModal.vue'
import api from '../api'
import { t, currentLang as lang } from '../i18n'

const files = ref([])
const currentPath = ref('/')
const loading = ref(false)
const origin = location.origin
const showSettings = ref(false)

const pathSegments = computed(() => currentPath.value.split('/').filter(Boolean))

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B','KB','MB','GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}
function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { year:'numeric', month:'short', day:'numeric' })
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

function navigateTo(path) { currentPath.value = path; load() }
function navigateToIndex(i) { navigateTo('/' + pathSegments.value.slice(0, i + 1).join('/')) }

async function setPrivate(f) {
  await api.put('/files/visibility', { path: f.path, is_public: false })
  load()
}

onMounted(load)
</script>

<style scoped>
.page-container { flex: 1; display: flex; flex-direction: column; }
.page-header { padding: 20px 28px 16px; border-bottom: 1px solid var(--gray-100); background: white; display: flex; align-items: center; justify-content: space-between; flex-shrink: 0; }
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); }
.page-sub { font-size: 12px; color: var(--gray-400); font-family: var(--editor-font, monospace); margin-top: 4px; }

.btn-settings { width: 36px; height: 36px; border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm); background: white; display: flex; align-items: center; justify-content: center; color: var(--gray-500); cursor: pointer; transition: var(--transition); }
.btn-settings svg { width: 17px; height: 17px; }
.btn-settings:hover { border-color: var(--blue-400); color: var(--blue-600); background: var(--blue-50); }

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
.empty-state span { font-size: 13px; color: var(--gray-300); }

.file-list { background: white; border-radius: var(--radius-lg); box-shadow: var(--shadow-sm); overflow: hidden; border: 1px solid var(--gray-100); }
.list-header { display: grid; grid-template-columns: 1fr 90px 130px 110px; padding: 10px 20px; background: var(--gray-50); border-bottom: 1px solid var(--gray-100); font-size: 11px; font-weight: 600; color: var(--gray-400); text-transform: uppercase; letter-spacing: 0.5px; }
.file-row { display: grid; grid-template-columns: 1fr 90px 130px 110px; padding: 11px 20px; border-bottom: 1px solid var(--gray-50); align-items: center; transition: var(--transition); }
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
.badge-pub { padding: 2px 7px; background: rgba(16,185,129,0.12); color: #10B981; border-radius: 20px; font-size: 10px; font-weight: 600; flex-shrink: 0; }
.badge-partial { padding: 2px 7px; background: rgba(245,158,11,0.12); color: #D97706; border-radius: 20px; font-size: 10px; font-weight: 600; flex-shrink: 0; }

.col-size, .col-date { font-size: 12px; color: var(--gray-400); }
.col-actions { display: flex; justify-content: flex-end; gap: 2px; }
.act-btn { width: 28px; height: 28px; border: none; background: none; cursor: pointer; border-radius: 6px; display: flex; align-items: center; justify-content: center; color: var(--gray-400); transition: var(--transition); text-decoration: none; }
.act-btn svg { width: 14px; height: 14px; }
.act-btn:hover { background: var(--blue-50); color: var(--blue-600); }
.act-btn.danger:hover { background: rgba(239,68,68,0.1); color: #EF4444; }
</style>
