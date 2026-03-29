<template>
  <div class="editor-page">
    <!-- 顶部导航栏 -->
    <div class="editor-header">
      <div class="editor-header-left">
        <button class="btn-back" @click="goBack" :title="t.back">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </button>
        <div class="editor-breadcrumb">
          <span class="editor-filename">{{ filename }}</span>
          <span class="editor-filepath">{{ parentPath }}</span>
        </div>
      </div>
      <div class="editor-header-right">
        <span v-if="saved" class="save-badge saved">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
          {{ lang === 'zh' ? '已保存' : 'Saved' }}
        </span>
        <span v-if="saveError" class="save-badge error">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          {{ saveError }}
        </span>
        <button v-if="mode === 'text' || forceText" class="btn-copy-all" @click="copyAll" :title="lang==='zh'?'复制全部内容':'Copy all'">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
          </svg>
          <span class="btn-copy-label">{{ copyDone ? (lang==='zh'?'已复制':'Copied!') : (lang==='zh'?'复制全部':'Copy All') }}</span>
        </button>
        <button v-if="mode === 'text' || forceText" class="btn-save" @click="doSave" :disabled="saving">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z"/>
            <polyline points="17 21 17 13 7 13 7 21"/>
            <polyline points="7 3 7 8 15 8"/>
          </svg>
          {{ saving ? (lang === 'zh' ? '保存中…' : 'Saving…') : t.saveFile }}
        </button>
      </div>
    </div>

    <!-- 主体区域 -->
    <div class="editor-body">
      <!-- 加载中 -->
      <div v-if="loading" class="editor-state">
        <div class="spinner"></div>
        <p>{{ lang === 'zh' ? '正在加载文件…' : 'Loading file…' }}</p>
      </div>

      <!-- 加载失败 -->
      <div v-else-if="loadError" class="editor-state editor-state-error">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        <p>{{ loadError }}</p>
        <button class="btn-ghost-sm" @click="goBack">{{ lang === 'zh' ? '返回' : 'Go Back' }}</button>
      </div>

      <!-- 图片预览 -->
      <div v-else-if="mode === 'image'" class="preview-img-wrap">
        <img :src="previewUrl" :alt="filename" class="preview-img" />
      </div>

      <!-- 文本编辑器 -->
      <div v-else-if="mode === 'text' || forceText" class="editor-cm-wrap">
        <p v-if="contentError" class="edit-error">{{ contentError }}</p>
        <CodeEditor v-else v-model="content" :filename="filename" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CodeEditor from '../components/CodeEditor.vue'
import { useAuthStore } from '../stores/auth'
import api from '../api'
import { t, currentLang as lang } from '../i18n'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

// 从路由参数解析文件路径（/edit/root/1.txt → /root/1.txt）
const filePath = computed(() => {
  const match = route.params.pathMatch
  if (!match) return '/'
  const joined = Array.isArray(match) ? match.join('/') : match
  return '/' + joined
})

const filename = computed(() => filePath.value.split('/').filter(Boolean).pop() || '')
const parentPath = computed(() => {
  const parts = filePath.value.split('/').filter(Boolean)
  parts.pop()
  return parts.length ? '/' + parts.join('/') : '/'
})

// 文件类型检测（与 Files.vue 保持一致）
const TEXT_EXTS  = new Set(['txt','md','markdown','dockerfile','makefile','log','service','timer','mount','target','ini','asc','keys','readme','man','networks','hosts','hostname','fstab','group','timezone','crontab','conf','cfg','env','yaml','yml','toml','json','jsonc','json5','html','htm','xml','svg','css','scss','sass','less','js','mjs','cjs','ts','tsx','jsx','vue','svelte','astro','py','go','java','rs','c','cpp','cc','h','hpp','sh','mod','bash','zsh','fish','ps1','bat','cmd','rb','php','pl','lua','r','swift','kt','cs','vb','sql','graphql','proto','csv','tsv','tex','rst','adoc','dart','scala','clj','cljs','erl','hrl','ex','exs','sol','gd','glsl','hlsl','wgsl','properties','gradle','editorconfig','browserslistrc','lock','strings','plist','xcconfig','ejs','pug','jade','hbs','handlebars','twig','liquid','mjml','ipynb','bib','diff','patch','ignore','procfile','htpasswd','htaccess','inf','reg','awk','sls','tf','tfvars','j2','repo','rules','dts','dtsi','config','inc','po','pot','msg'])
const IMAGE_EXTS = new Set(['jpg','jpeg','png','gif','webp','bmp','ico','tiff','tif','avif','svg'])

function getMode(name) {
  if (!name) return 'text'
  const lower = name.toLowerCase()
  const base = lower.split('/').pop()
  if (['dockerfile','makefile','gnumakefile'].includes(base)) return 'text'
  if (!name.includes('.')) return TEXT_EXTS.has(lower) ? 'text' : 'unknown'
  const ext = lower.split('.').pop()
  if (IMAGE_EXTS.has(ext)) return 'image'
  if (TEXT_EXTS.has(ext)) return 'text'
  return 'unknown'
}

const mode      = ref('text')
const forceText = ref(false)
const loading   = ref(true)
const loadError = ref('')
const content   = ref('')
const contentError = ref('')
const previewUrl   = ref('')
const saving    = ref(false)
const saved     = ref(false)
const saveError = ref('')
const copyDone  = ref(false)
let copyTimer   = null

async function copyAll() {
  try {
    await navigator.clipboard.writeText(content.value)
    copyDone.value = true
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => { copyDone.value = false }, 2000)
  } catch {
    // clipboard API 不可用时 fallback
    const ta = document.createElement('textarea')
    ta.value = content.value
    ta.style.cssText = 'position:fixed;opacity:0'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    copyDone.value = true
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => { copyDone.value = false }, 2000)
  }
}

let savedTimer = null

async function loadFile() {
  loading.value = true
  loadError.value = ''
  contentError.value = ''
  previewUrl.value = ''

  const m = getMode(filename.value)
  mode.value = m

  if (m === 'image') {
    const token = localStorage.getItem('token')
    try {
      const resp = await fetch(`/api/files/download?path=${encodeURIComponent(filePath.value)}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!resp.ok) throw new Error(resp.statusText)
      previewUrl.value = URL.createObjectURL(await resp.blob())
    } catch (e) {
      loadError.value = e.message || 'Failed to load image'
    }
    loading.value = false
    return
  }

  if (m === 'unknown') {
    // 未知类型：调后端 detect 判断是否为文本
    try {
      const { data } = await api.get('/files/detect', { params: { path: filePath.value } })
      if (data.text) {
        forceText.value = true
        // 继续往下读内容
      } else {
        loadError.value = lang.value === 'zh'
          ? `无法打开该文件：二进制格式（${data.mime || '未知类型'}），不支持在线编辑`
          : `Cannot open file: binary format (${data.mime || 'unknown type'}), not supported for editing`
        loading.value = false
        return
      }
    } catch (e) {
      loadError.value = e.response?.data?.error || (lang.value === 'zh' ? '检测文件类型失败' : 'Failed to detect file type')
      loading.value = false
      return
    }
  }

  // text mode（包括 forceText 的 unknown 文件）
  try {
    const { data } = await api.get('/files/content', { params: { path: filePath.value } })
    content.value = data.content
  } catch (e) {
    contentError.value = e.response?.data?.error || (lang.value === 'zh' ? '无法读取文件内容' : 'Cannot read file content')
  }
  loading.value = false
}

async function doSave() {
  if (saving.value) return
  saving.value = true
  saved.value = false
  saveError.value = ''
  try {
    await api.put('/files/content', { path: filePath.value, content: content.value })
    saved.value = true
    clearTimeout(savedTimer)
    savedTimer = setTimeout(() => { saved.value = false }, 2500)
  } catch (e) {
    saveError.value = e.response?.data?.error || (lang.value === 'zh' ? '保存失败' : 'Save failed')
    setTimeout(() => { saveError.value = '' }, 3000)
  } finally {
    saving.value = false
  }
}

function doDownload() {
  const token = localStorage.getItem('token')
  const a = document.createElement('a')
  a.href = `/api/files/download?path=${encodeURIComponent(filePath.value)}`
  // 带 token 下载
  fetch(a.href, { headers: { Authorization: `Bearer ${token}` } })
    .then(r => r.blob())
    .then(blob => {
      const url = URL.createObjectURL(blob)
      a.href = url
      a.download = filename.value
      a.click()
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    })
}

function goBack() {
  // 返回到文件列表对应目录
  const urlPath = parentPath.value === '/' ? '/files' : '/files' + parentPath.value
  router.push(urlPath)
}

// Ctrl+S / Cmd+S 快捷键保存
function onKeydown(e) {
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    if (mode.value === 'text' || forceText.value) doSave()
  }
}

onMounted(() => {
  loadFile()
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  clearTimeout(savedTimer)
  clearTimeout(copyTimer)
})
</script>

<style scoped>
.editor-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
  background: var(--gray-50, #F8FAFC);
  overflow: hidden;
}

/* ── 顶部栏 ───────────────────────────────────────── */
.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 16px;
  height: 56px;
  min-height: 56px;
  background: #fff;
  border-bottom: 1px solid var(--gray-200, #E2E8F0);
  box-shadow: 0 1px 3px rgba(15,23,42,0.06);
  z-index: 10;
}

.editor-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.editor-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.btn-back {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 8px;
  border: 1px solid var(--gray-200, #E2E8F0);
  background: #fff;
  color: var(--gray-600, #475569);
  cursor: pointer;
  flex-shrink: 0;
  transition: background .15s, border-color .15s, color .15s;
}
.btn-back:hover {
  background: var(--gray-50, #F8FAFC);
  border-color: var(--blue-300, #93C5FD);
  color: var(--blue-600, #2563EB);
}
.btn-back svg { width: 18px; height: 18px; }

.editor-breadcrumb {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.editor-filename {
  font-size: 15px;
  font-weight: 600;
  color: var(--gray-900, #0F172A);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.editor-filepath {
  font-size: 12px;
  color: var(--gray-400, #94A3B8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-copy-all {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 13px;
  border-radius: 8px;
  border: 1px solid var(--gray-300, #CBD5E1);
  background: #fff;
  color: var(--gray-700, #334155);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background .15s, border-color .15s, color .15s;
}
.btn-copy-all:hover { background: var(--gray-50); border-color: var(--blue-300, #93C5FD); color: var(--blue-600, #2563EB); }
.btn-copy-all svg { width: 15px; height: 15px; flex-shrink: 0; }
.btn-save {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border-radius: 8px;
  border: none;
  background: var(--blue-600, #2563EB);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background .15s, opacity .15s;
}
.btn-save:hover { background: var(--blue-700, #1D4ED8); }
.btn-save:disabled { opacity: .6; cursor: not-allowed; }
.btn-save svg { width: 15px; height: 15px; }

.btn-ghost-sm {
  padding: 6px 13px;
  border-radius: 7px;
  border: 1px solid var(--gray-300, #CBD5E1);
  background: #fff;
  color: var(--gray-700, #334155);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background .15s, border-color .15s;
}
.btn-ghost-sm:hover { background: var(--gray-50, #F8FAFC); border-color: var(--blue-300, #93C5FD); }

.save-badge {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 6px;
}
.save-badge svg { width: 13px; height: 13px; }
.save-badge.saved  { color: #16A34A; background: #F0FDF4; }
.save-badge.error  { color: #DC2626; background: #FEF2F2; }

/* ── 主体 ─────────────────────────────────────────── */
.editor-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 16px;
  overflow: hidden;
}

.editor-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--gray-500, #64748B);
  text-align: center;
}
.editor-state svg { width: 48px; height: 48px; opacity: .5; }
.editor-state p   { font-size: 15px; margin: 0; }
.editor-state-error svg { color: var(--red-400, #F87171); opacity: 1; }
.editor-state-error p   { color: var(--gray-700, #334155); }

.spinner {
  width: 32px; height: 32px;
  border: 3px solid var(--blue-100, #DBEAFE);
  border-top-color: var(--blue-500, #3B82F6);
  border-radius: 50%;
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* 图片预览 */
.preview-img-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: auto;
  background: var(--gray-100, #F1F5F9);
  border-radius: 10px;
}
.preview-img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 6px;
}

/* 不支持类型 */
.unsupported-wrap {
  gap: 12px;
}
.unsupported-ext {
  font-size: 28px;
  font-weight: 700;
  color: var(--gray-300, #CBD5E1);
  font-family: 'JetBrains Mono', monospace;
}

/* 编辑器容器 */
.editor-cm-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  border-radius: 10px;
}
.edit-error {
  background: #FEF2F2;
  color: #DC2626;
  border: 1px solid #FECACA;
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 13px;
  margin: 0 0 10px;
}

/* ── 移动端 ─────────────────────────────────────── */
@media (max-width: 640px) {
  .editor-header { height: 52px; min-height: 52px; padding: 0 12px; gap: 8px; }
  .editor-filename { font-size: 14px; }
  .editor-filepath { display: none; }
  .btn-save { padding: 7px 12px; font-size: 13px; }
  .btn-save svg { display: none; }
  .btn-copy-all { padding: 7px 10px; }
  .btn-copy-label { display: none; }
  .editor-body { padding: 10px; }
}
</style>
