<template>
  <div class="share-page">
    <div class="share-header">
      <div class="brand">
        <div class="brand-icon">
          <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="slg3" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#2563EB"/>
                <stop offset="100%" style="stop-color:#38BDF8"/>
              </linearGradient>
            </defs>
            <ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#slg3)"/>
            <circle cx="36" cy="56" r="16" fill="url(#slg3)"/>
            <circle cx="58" cy="50" r="20" fill="url(#slg3)"/>
            <polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/>
          </svg>
        </div>
        <span>CloudOne</span>
      </div>
    </div>

    <div class="share-content">
      <div v-if="loading" class="loading-state">
        <div class="spinner-lg"></div>
      </div>

      <div v-else-if="error" class="error-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
        <p>{{ error }}</p>
      </div>

      <template v-else-if="data">
        <div class="share-card">

          <!-- 单文件下载视图（直接分享文件 或 从目录点进来的文件） -->
          <div v-if="fileView" class="share-file">
            <button v-if="data.is_dir" class="back-btn" @click="fileView = null">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px"><polyline points="15 18 9 12 15 6"/></svg>
              返回
            </button>
            <div class="file-preview-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <h2>{{ fileView.name }}</h2>
            <a :href="fileView.downloadUrl" class="download-btn">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              下载
            </a>
          </div>

          <!-- 目录视图 -->
          <div v-else-if="data.is_dir" class="share-dir">
            <div class="breadcrumb">
              <button class="crumb-btn" @click="navigateTo(null)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>
                {{ rootName }}
              </button>
              <template v-for="(seg, i) in subPath" :key="i">
                <span class="crumb-sep">/</span>
                <button class="crumb-btn" @click="navigateTo(subPath.slice(0, i+1).join('/'))">{{ seg }}</button>
              </template>
            </div>

            <div v-if="dirLoading" class="loading-state" style="padding:40px"><div class="spinner-lg"></div></div>
            <div v-else class="file-list">
              <div v-if="!currentFiles.length" class="empty-dir">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>
                <span>空文件夹</span>
              </div>
              <div
                v-for="f in currentFiles" :key="f.path"
                class="file-item clickable"
                @click="f.is_dir ? enterDir(f) : openFile(f)"
              >
                <svg v-if="f.is_dir" viewBox="0 0 24 24" fill="currentColor" class="ficon folder"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="ficon"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                <span class="fname">{{ f.name }}</span>
                <span v-if="f.is_dir" class="fdir-arrow">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><polyline points="9 18 15 12 9 6"/></svg>
                </span>
                <span v-else class="fsize">{{ formatSize(f.size) }}</span>
              </div>
            </div>
          </div>

          <!-- 直接分享的单文件 -->
          <div v-else class="share-file">
            <div class="file-preview-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <h2>{{ fileName }}</h2>
            <a :href="`/api/s/${route.params.code}/download`" class="download-btn">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              下载
            </a>
          </div>

        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const data = ref(null)
const loading = ref(true)
const error = ref('')

const subPath = ref([])
const currentFiles = ref([])
const dirLoading = ref(false)

// 目录内点击文件后切换到单文件下载视图
const fileView = ref(null) // { name, downloadUrl }

const rootName = computed(() => {
  if (!data.value) return ''
  const p = data.value.file_path || ''
  return p.split('/').filter(Boolean).pop() || p
})

const fileName = computed(() => {
  if (!data.value) return ''
  const p = data.value.file_path || ''
  return p.split('/').filter(Boolean).pop() || p
})

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B','KB','MB','GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

// 目录内点击文件，切换到下载视图
function openFile(f) {
  const sp = [...subPath.value, f.name].join('/')
  fileView.value = {
    name: f.name,
    downloadUrl: `/api/s/${route.params.code}/download?subpath=${encodeURIComponent(sp)}`
  }
}

async function enterDir(f) {
  subPath.value = [...subPath.value, f.name]
  await loadDir()
}

async function navigateTo(relPath) {
  subPath.value = relPath ? relPath.split('/').filter(Boolean) : []
  await loadDir()
}

async function loadDir() {
  dirLoading.value = true
  try {
    const sp = subPath.value.join('/')
    const url = sp
      ? `/api/s/${route.params.code}?subpath=${encodeURIComponent(sp)}`
      : `/api/s/${route.params.code}`
    const res = await fetch(url)
    if (!res.ok) throw new Error()
    const json = await res.json()
    currentFiles.value = json.files || []
  } catch {
    currentFiles.value = []
  }
  dirLoading.value = false
}

onMounted(async () => {
  try {
    const res = await fetch(`/api/s/${route.params.code}`)
    if (!res.ok) throw new Error('Not found')
    data.value = await res.json()
    if (data.value.is_dir) {
      currentFiles.value = data.value.files || []
    }
  } catch {
    error.value = 'Share link not found or expired.'
  }
  loading.value = false
})
</script>

<style scoped>
.share-page { min-height: 100vh; background: var(--gray-50); }
.share-header { padding: 20px 32px; background: white; border-bottom: 1px solid var(--gray-100); box-shadow: var(--shadow-sm); }
.brand { display: flex; align-items: center; gap: 12px; }
.brand-icon { width: 36px; height: 36px; }
.brand span { font-size: 18px; font-weight: 700; background: linear-gradient(135deg, #2563EB, #38BDF8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.share-content { max-width: 640px; margin: 48px auto; padding: 0 20px; }
.loading-state { display: flex; justify-content: center; padding: 80px; }
.spinner-lg { width: 32px; height: 32px; border: 3px solid var(--gray-200); border-top-color: var(--blue-500); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-state { display: flex; flex-direction: column; align-items: center; gap: 16px; padding: 80px 0; color: var(--gray-300); }
.error-state svg { width: 64px; height: 64px; }
.error-state p { font-size: 16px; color: var(--gray-400); }
.share-card { background: white; border-radius: 20px; padding: 32px 40px 40px; box-shadow: var(--shadow); border: 1px solid var(--gray-100); }
.breadcrumb { display: flex; align-items: center; gap: 4px; flex-wrap: wrap; margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid var(--gray-100); }
.crumb-btn { display: flex; align-items: center; gap: 4px; background: none; border: none; color: var(--blue-600); font-size: 13px; font-weight: 500; font-family: inherit; cursor: pointer; padding: 3px 7px; border-radius: 6px; transition: background .15s; }
.crumb-btn:hover { background: var(--blue-50); }
.crumb-sep { color: var(--gray-300); font-size: 13px; }
.file-list { display: flex; flex-direction: column; gap: 2px; }
.file-item { display: flex; align-items: center; gap: 12px; padding: 10px 12px; border-radius: 8px; transition: background .15s; }
.file-item.clickable { cursor: pointer; }
.file-item.clickable:hover { background: var(--blue-50); }
.ficon { width: 18px; height: 18px; flex-shrink: 0; color: var(--gray-400); }
.ficon.folder { color: #F59E0B; }
.fname { flex: 1; font-size: 14px; color: var(--gray-700); }
.fsize { font-size: 12px; color: var(--gray-400); }
.fdir-arrow { color: var(--gray-300); display: flex; align-items: center; }
.empty-dir { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 40px; color: var(--gray-300); }
.empty-dir svg { width: 40px; height: 40px; }
.empty-dir span { font-size: 13px; color: var(--gray-400); }
.share-file { text-align: center; }
.back-btn { display: flex; align-items: center; gap: 5px; background: none; border: none; color: var(--blue-600); font-size: 13px; font-family: inherit; cursor: pointer; padding: 4px 8px; border-radius: 6px; margin-bottom: 24px; transition: background .15s; }
.back-btn:hover { background: var(--blue-50); }
.file-preview-icon { width: 80px; height: 80px; background: var(--blue-50); border-radius: 20px; display: flex; align-items: center; justify-content: center; margin: 0 auto 20px; color: var(--blue-500); }
.file-preview-icon svg { width: 40px; height: 40px; }
.share-file h2 { font-size: 22px; font-weight: 600; color: var(--gray-800); margin-bottom: 24px; }
.download-btn { display: inline-flex; align-items: center; gap: 8px; padding: 14px 28px; background: var(--primary-gradient); color: white; border-radius: var(--radius); text-decoration: none; font-size: 15px; font-weight: 600; box-shadow: var(--primary-shadow); transition: var(--transition); }
.download-btn svg { width: 18px; height: 18px; }
.download-btn:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(37,99,235,0.35); }

@media (max-width: 600px) {
  .share-header { padding: 14px 20px; }
  .share-content { margin: 24px auto; padding: 0 14px; }
  .share-card { padding: 24px 20px 28px; border-radius: 16px; }
  .share-file h2 { font-size: 17px; word-break: break-all; }
  .download-btn { width: 100%; justify-content: center; padding: 16px; font-size: 16px; border-radius: 14px; }
  .file-preview-icon { width: 64px; height: 64px; border-radius: 16px; }
  .file-preview-icon svg { width: 32px; height: 32px; }
  .file-item { padding: 12px 10px; }
  .fname { font-size: 13px; }
}
</style>
