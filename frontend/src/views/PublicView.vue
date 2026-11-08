<template>
  <div class="pub-page">
    <!-- Header -->
    <div class="pub-header">
      <div class="pub-brand">
        <div class="pub-brand-icon">
          <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="pslg" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#2563EB"/>
                <stop offset="100%" style="stop-color:#38BDF8"/>
              </linearGradient>
            </defs>
            <ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#pslg)"/>
            <circle cx="36" cy="56" r="16" fill="url(#pslg)"/>
            <circle cx="58" cy="50" r="20" fill="url(#pslg)"/>
            <polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/>
          </svg>
        </div>
        <span class="pub-brand-name">CloudOne</span>
      </div>
      <div class="pub-header-center">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
        <span>{{ lang === 'zh' ? '公开文件' : 'Public Files' }}</span>
      </div>
      <div class="pub-header-right">
        <span class="raw-hint">{{ origin }}/raw/...</span>
      </div>
    </div>

    <!-- Content -->
    <div class="pub-content">
      <!-- 面包屑 -->
      <div v-if="currentPath !== '/'" class="pub-breadcrumb">
        <button @click="navigateTo('/')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>
          {{ lang === 'zh' ? '根目录' : 'Root' }}
        </button>
        <template v-for="(seg, i) in pathSegments" :key="i">
          <span class="crumb-sep">/</span>
          <button @click="navigateToIndex(i)">{{ seg }}</button>
        </template>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="pub-loading">
        <div class="spinner"></div>
      </div>

      <!-- Empty -->
      <div v-else-if="!files.length" class="pub-empty">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
        <p>{{ lang === 'zh' ? '此目录下暂无公开文件' : 'No public files here' }}</p>
      </div>

      <!-- File List -->
      <div v-else class="pub-list">
        <!-- Directory rows -->
        <div v-for="f in files" :key="f.path" class="pub-row" :class="{ 'pub-row-dir': f.is_dir }" @click="f.is_dir ? navigateTo(f.path) : null">
          <div class="pub-row-icon" :class="f.is_dir ? 'folder' : 'file'">
            <svg v-if="f.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
          </div>
          <div class="pub-row-info">
            <p class="pub-row-name">{{ f.name }}</p>
            <p class="pub-row-meta">{{ f.is_dir ? (f.is_public ? (lang === 'zh' ? '公开文件夹 · 点击浏览' : 'Public folder · Click to browse') : (lang === 'zh' ? '文件夹（含公开文件）· 点击浏览' : 'Folder with public files · Browse')) : formatSize(f.size) }}</p>
          </div>
          <div class="pub-row-actions" @click.stop>
            <template v-if="!f.is_dir">
              <!-- RAW 按钮 -->
              <a :href="`/raw${f.path}`" target="_blank" class="pub-btn pub-btn-raw" title="RAW">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                <span>RAW</span>
              </a>
              <!-- 下载按钮 -->
              <a :href="`/pub/dl?path=${encodeURIComponent(f.path)}`" :download="f.name" class="pub-btn pub-btn-dl">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                <span>{{ lang === 'zh' ? '下载' : 'Download' }}</span>
              </a>
            </template>
            <template v-else>
              <span class="pub-enter-hint">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
              </span>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { currentLang as lang } from '../i18n'

const files = ref([])
const currentPath = ref('/')
const loading = ref(false)
const origin = location.origin

const pathSegments = computed(() => currentPath.value.split('/').filter(Boolean))

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B','KB','MB','GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
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

onMounted(load)
</script>

<style scoped>
.pub-page { min-height: 100vh; background: var(--gray-50); }

.pub-header {
  padding: 16px 32px;
  background: white;
  border-bottom: 1px solid var(--gray-100);
  box-shadow: var(--shadow-sm);
  display: flex; align-items: center;
  gap: 16px;
}
.pub-brand { display: flex; align-items: center; gap: 10px; }
.pub-brand-icon { width: 30px; height: 30px; }
.pub-brand-name { font-size: 16px; font-weight: 700; background: linear-gradient(135deg,#2563EB,#38BDF8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }

.pub-header-center { display: flex; align-items: center; gap: 7px; flex: 1; justify-content: center; }
.pub-header-center svg { width: 18px; height: 18px; color: var(--blue-500); }
.pub-header-center span { font-size: 15px; font-weight: 600; color: var(--gray-700); }

.pub-header-right {}
.raw-hint { font-size: 11px; color: var(--gray-400); font-family: var(--editor-font, monospace); background: var(--gray-100); padding: 3px 10px; border-radius: 20px; }

.pub-content { max-width: 800px; margin: 0 auto; padding: 28px 20px; }

.pub-breadcrumb { display: flex; align-items: center; gap: 4px; flex-wrap: wrap; margin-bottom: 20px; }
.pub-breadcrumb button { display: flex; align-items: center; gap: 5px; background: none; border: none; color: var(--blue-600); font-size: 13px; font-weight: 500; font-family: inherit; cursor: pointer; padding: 4px 8px; border-radius: 6px; transition: var(--transition); }
.pub-breadcrumb button svg { width: 14px; height: 14px; }
.pub-breadcrumb button:hover { background: var(--blue-50); }
.crumb-sep { color: var(--gray-300); font-size: 14px; }

.pub-loading { display: flex; justify-content: center; padding: 80px; }
.spinner { width: 32px; height: 32px; border: 3px solid var(--gray-200); border-top-color: var(--blue-500); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.pub-empty { display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 80px 0; color: var(--gray-300); }
.pub-empty svg { width: 60px; height: 60px; }
.pub-empty p { font-size: 15px; color: var(--gray-400); font-weight: 500; }

.pub-list {
  background: white;
  border-radius: var(--radius-lg);
  border: 1px solid var(--gray-100);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.pub-row {
  display: flex; align-items: center; gap: 14px;
  padding: 14px 20px;
  border-bottom: 1px solid var(--gray-50);
  transition: var(--transition);
}
.pub-row:last-child { border-bottom: none; }
.pub-row-dir { cursor: pointer; }
.pub-row-dir:hover { background: var(--blue-50); }
.pub-row:not(.pub-row-dir):hover { background: var(--gray-50); }

.pub-row-icon { width: 38px; height: 38px; border-radius: 9px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.pub-row-icon.folder { background: rgba(251,191,36,0.15); color: #F59E0B; }
.pub-row-icon.file { background: var(--blue-50); color: var(--blue-500); }
.pub-row-icon svg { width: 18px; height: 18px; }

.pub-row-info { flex: 1; min-width: 0; }
.pub-row-name { font-size: 14px; font-weight: 600; color: var(--gray-700); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pub-row-meta { font-size: 11px; color: var(--gray-400); margin-top: 2px; }

.pub-row-actions { display: flex; align-items: center; gap: 8px; }

.pub-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 6px 14px; border-radius: var(--radius-sm);
  font-size: 12px; font-weight: 600; font-family: inherit;
  cursor: pointer; border: none; text-decoration: none;
  transition: var(--transition);
}
.pub-btn svg { width: 13px; height: 13px; }

.pub-btn-raw {
  background: var(--gray-100); color: var(--gray-600);
  border: 1.5px solid var(--gray-200);
}
.pub-btn-raw:hover { background: var(--gray-800); color: white; border-color: var(--gray-800); }

.pub-btn-dl {
  background: var(--primary-gradient); color: white;
  box-shadow: var(--primary-shadow);
}
.pub-btn-dl:hover { transform: translateY(-1px); box-shadow: var(--primary-shadow-hover); }

.pub-enter-hint { color: var(--gray-300); display: flex; }
.pub-enter-hint svg { width: 18px; height: 18px; }

@media (max-width: 600px) {
  .pub-header { padding: 12px 16px; gap: 10px; flex-wrap: wrap; }
  .pub-header-right { display: none; }
  .pub-header-center { justify-content: flex-start; }
  .pub-content { padding: 16px 14px; }
  .pub-row { padding: 12px 14px; gap: 10px; }
  .pub-row-name { font-size: 13px; }
  .pub-btn { padding: 6px 10px; font-size: 11px; }
  .pub-btn span { display: none; } /* 只显示图标 */
  .pub-btn { padding: 7px; }
  .pub-btn svg { width: 16px; height: 16px; }
  .pub-row-icon { width: 32px; height: 32px; border-radius: 8px; }
}
</style>
