<template>
  <Layout>
    <div class="files-page">
      <!-- Header -->
      <div class="page-header">
        <div class="breadcrumb">
          <button class="crumb-home" @click="navigate('/')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
            {{ t.home }}
          </button>
          <template v-for="(seg, i) in pathSegments" :key="i">
            <span class="crumb-sep">/</span>
            <button class="crumb-item" @click="navigateToIndex(i)">{{ seg }}</button>
          </template>
        </div>

        <div class="header-actions">
          <div class="lang-mini">
            <button :class="{ active: lang === 'zh' }" @click="setLang('zh')">中文</button>
            <button :class="{ active: lang === 'en' }" @click="setLang('en')">EN</button>
          </div>
          <button class="btn-action" @click="showUpload = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            {{ t.upload }}
          </button>
          <button class="btn-action" @click="showMkdir = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
            {{ t.newFolder }}
          </button>
          <button class="btn-action" @click="showCreate = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
            {{ t.newFile }}
          </button>
        </div>
      </div>

      <!-- Drop zone + file list -->
      <div
        class="drop-zone"
        :class="{ dragging }"
        @dragover.prevent="dragging = true"
        @dragleave="dragging = false"
        @drop.prevent="handleDrop"
      >
        <div v-if="loading" class="loading-state">
          <div class="spinner-lg"></div>
        </div>

        <div v-else-if="!files.length" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>
          <p>{{ t.noFiles }}</p>
          <span>{{ t.uploadDrop }}</span>
        </div>

        <div v-else class="file-table">
          <div class="file-header">
            <span class="col-name">{{ t.name }}</span>
            <span class="col-size">{{ t.size }}</span>
            <span class="col-date">{{ t.modified }}</span>
            <span class="col-vis"></span>
            <span class="col-actions"></span>
          </div>

          <transition-group name="slide-up" tag="div">
            <div
              v-for="file in files"
              :key="file.path"
              class="file-row"
              :class="{ selected: selected.includes(file.path) }"
              @click="handleClick(file)"
              @contextmenu.prevent="openCtx($event, file)"
            >
              <div class="col-name">
                <div class="file-icon" :class="file.is_dir ? 'folder-icon' : getExt(file.name)">
                  <svg v-if="file.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                </div>
                <span class="file-name">{{ file.name }}</span>
                <span v-if="file.is_public" class="badge-public">{{ t.isPublic }}</span>
              </div>
              <span class="col-size">{{ file.is_dir ? '—' : formatSize(file.size) }}</span>
              <span class="col-date">{{ formatDate(file.mod_time) }}</span>
              <span class="col-vis"></span>
              <div class="col-actions" @click.stop>
                <div class="action-menu">
                  <button class="act-btn" @click.stop="downloadFile(file)" v-if="!file.is_dir" :title="t.download">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  </button>
                  <button class="act-btn" @click.stop="shareFile(file)" :title="t.share">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
                  </button>
                  <button class="act-btn" @click.stop="togglePublic(file)" :title="file.is_public ? t.setPrivate : t.setPublic">
                    <svg v-if="file.is_public" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
                    <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
                  </button>
                  <button class="act-btn danger" @click.stop="confirmDelete(file)" :title="t.delete">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/></svg>
                  </button>
                </div>
              </div>
            </div>
          </transition-group>
        </div>
      </div>
    </div>

    <!-- Upload Modal -->
    <teleport to="body">
      <div v-if="showUpload" class="modal-bg" @click.self="showUpload = false">
        <div class="modal">
          <h3>{{ t.upload }}</h3>
          <div class="upload-area" @click="$refs.fileInput.click()" @dragover.prevent @drop.prevent="dropUpload">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            <p>{{ t.uploadDrop }}</p>
          </div>
          <input ref="fileInput" type="file" multiple @change="uploadFiles" style="display:none" />
          <div v-if="uploadProgress.length" class="upload-list">
            <div v-for="(u, i) in uploadProgress" :key="i" class="upload-item">
              <span>{{ u.name }}</span>
              <span :class="u.done ? 'done' : 'pending'">{{ u.done ? '✓' : '...' }}</span>
            </div>
          </div>
          <button class="modal-close" @click="showUpload = false">{{ t.cancel }}</button>
        </div>
      </div>

      <!-- Mkdir Modal -->
      <div v-if="showMkdir" class="modal-bg" @click.self="showMkdir = false">
        <div class="modal">
          <h3>{{ t.newFolder }}</h3>
          <div class="field">
            <label>{{ t.folderName }}</label>
            <input v-model="newDirName" type="text" @keyup.enter="doMkdir" ref="mkdirInput" />
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showMkdir = false">{{ t.cancel }}</button>
            <button class="btn-primary-sm" @click="doMkdir">{{ t.create }}</button>
          </div>
        </div>
      </div>

      <!-- Create File Modal -->
      <div v-if="showCreate" class="modal-bg" @click.self="showCreate = false">
        <div class="modal modal-lg">
          <h3>{{ t.createFileTitle }}</h3>
          <div class="field">
            <label>{{ t.fileName }}</label>
            <input v-model="newFile.name" type="text" />
          </div>
          <div class="field">
            <label>{{ t.fileContent }}</label>
            <textarea v-model="newFile.content" rows="8"></textarea>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showCreate = false">{{ t.cancel }}</button>
            <button class="btn-primary-sm" @click="doCreateFile">{{ t.create }}</button>
          </div>
        </div>
      </div>

      <!-- Delete Confirm -->
      <div v-if="deleteTarget" class="modal-bg" @click.self="deleteTarget = null">
        <div class="modal">
          <h3>{{ t.confirmDelete }}</h3>
          <p class="modal-desc">{{ t.deleteWarning }}</p>
          <div class="modal-actions">
            <button class="btn-ghost" @click="deleteTarget = null">{{ t.cancel }}</button>
            <button class="btn-danger" @click="doDelete">{{ t.confirm }}</button>
          </div>
        </div>
      </div>

      <!-- Share Modal -->
      <div v-if="shareResult" class="modal-bg" @click.self="shareResult = null">
        <div class="modal">
          <h3>{{ t.shareTitle }}</h3>
          <div class="share-link-box">
            <input readonly :value="shareUrl" class="share-input" />
            <button class="btn-copy" @click="copyShare">
              {{ copied ? t.linkCopied : t.copyLink }}
            </button>
          </div>
          <p class="share-raw">{{ t.rawLink }}: <a :href="rawShareUrl" target="_blank">{{ rawShareUrl }}</a></p>
          <button class="modal-close" @click="shareResult = null">{{ t.cancel }}</button>
        </div>
      </div>
    </teleport>
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import Layout from '../components/Layout.vue'
import api from '../api'
import { t, currentLang as lang, setLang } from '../i18n'

const currentPath = ref('/')
const files = ref([])
const loading = ref(false)
const dragging = ref(false)
const selected = ref([])

const showUpload = ref(false)
const showMkdir = ref(false)
const showCreate = ref(false)
const mkdirInput = ref(null)
const fileInput = ref(null)

const newDirName = ref('')
const newFile = ref({ name: '', content: '' })
const deleteTarget = ref(null)
const shareResult = ref(null)
const copied = ref(false)
const uploadProgress = ref([])

const pathSegments = computed(() => {
  return currentPath.value.split('/').filter(Boolean)
})

const shareUrl = computed(() =>
  shareResult.value ? `${location.origin}/s/${shareResult.value.code}` : ''
)
const rawShareUrl = computed(() =>
  shareResult.value ? `${location.origin}/s/${shareResult.value.code}/raw` : ''
)

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/files', { params: { path: currentPath.value } })
    files.value = data.files || []
  } catch {}
  loading.value = false
}

function navigate(path) {
  currentPath.value = path
  load()
}

function navigateToIndex(i) {
  const segs = pathSegments.value.slice(0, i + 1)
  navigate('/' + segs.join('/'))
}

function handleClick(file) {
  if (file.is_dir) navigate(file.path)
}

function getExt(name) {
  const ext = name.split('.').pop()?.toLowerCase()
  if (['jpg','jpeg','png','gif','webp','svg'].includes(ext)) return 'img'
  if (['mp4','mkv','avi','mov'].includes(ext)) return 'video'
  if (['mp3','flac','wav','aac'].includes(ext)) return 'audio'
  if (['zip','tar','gz','rar','7z'].includes(ext)) return 'archive'
  if (['pdf'].includes(ext)) return 'pdf'
  if (['md','txt','log'].includes(ext)) return 'text'
  if (['js','ts','py','go','java','sh','rs'].includes(ext)) return 'code'
  return 'default'
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B','KB','MB','GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString(lang.value === 'zh' ? 'zh-CN' : 'en-US', {
    year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
  })
}

function downloadFile(file) {
  const url = `/api/files/download?path=${encodeURIComponent(file.path)}`
  const a = document.createElement('a')
  a.href = url
  const token = localStorage.getItem('token')
  // Use fetch for auth
  fetch(url, { headers: { Authorization: `Bearer ${token}` } })
    .then(r => r.blob())
    .then(blob => {
      const url2 = URL.createObjectURL(blob)
      a.href = url2
      a.download = file.name
      a.click()
      URL.revokeObjectURL(url2)
    })
}

async function togglePublic(file) {
  await api.put('/files/visibility', { path: file.path, is_public: !file.is_public })
  file.is_public = !file.is_public
}

async function shareFile(file) {
  const { data } = await api.post('/share', { path: file.path, is_dir: file.is_dir })
  shareResult.value = data
}

async function copyShare() {
  await navigator.clipboard.writeText(shareUrl.value)
  copied.value = true
  setTimeout(() => copied.value = false, 2000)
}

function confirmDelete(file) {
  deleteTarget.value = file
}

async function doDelete() {
  await api.delete('/files', { params: { path: deleteTarget.value.path } })
  deleteTarget.value = null
  load()
}

async function doMkdir() {
  if (!newDirName.value) return
  const path = currentPath.value.replace(/\/$/, '') + '/' + newDirName.value
  await api.post('/files/mkdir', { path })
  newDirName.value = ''
  showMkdir.value = false
  load()
}

async function doCreateFile() {
  if (!newFile.value.name) return
  const path = currentPath.value.replace(/\/$/, '') + '/' + newFile.value.name
  await api.post('/files/create', { path, content: newFile.value.content })
  newFile.value = { name: '', content: '' }
  showCreate.value = false
  load()
}

async function uploadFiles(e) {
  const fs = Array.from(e.target.files)
  uploadProgress.value = fs.map(f => ({ name: f.name, done: false }))
  const form = new FormData()
  fs.forEach(f => form.append('files', f))
  await api.post('/files/upload', form, { params: { path: currentPath.value } })
  uploadProgress.value.forEach(u => u.done = true)
  setTimeout(() => { showUpload.value = false; uploadProgress.value = []; load() }, 800)
}

async function dropUpload(e) {
  const fs = Array.from(e.dataTransfer.files)
  uploadProgress.value = fs.map(f => ({ name: f.name, done: false }))
  const form = new FormData()
  fs.forEach(f => form.append('files', f))
  await api.post('/files/upload', form, { params: { path: currentPath.value } })
  uploadProgress.value.forEach(u => u.done = true)
  setTimeout(() => { showUpload.value = false; uploadProgress.value = []; load() }, 800)
}

async function handleDrop(e) {
  dragging.value = false
  const fs = Array.from(e.dataTransfer.files)
  if (!fs.length) return
  const form = new FormData()
  fs.forEach(f => form.append('files', f))
  await api.post('/files/upload', form, { params: { path: currentPath.value } })
  load()
}

onMounted(load)
</script>

<style scoped>
.files-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.page-header {
  padding: 20px 28px 16px;
  border-bottom: 1px solid var(--gray-100);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: white;
  flex-shrink: 0;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.crumb-home, .crumb-item {
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  color: var(--blue-600);
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: var(--transition);
}

.crumb-home svg { width: 15px; height: 15px; }
.crumb-home:hover, .crumb-item:hover { background: var(--blue-50); }

.crumb-sep {
  color: var(--gray-300);
  font-size: 14px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.lang-mini {
  display: flex;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  overflow: hidden;
  margin-right: 8px;
}

.lang-mini button {
  padding: 6px 12px;
  border: none;
  background: transparent;
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  color: var(--gray-500);
  transition: var(--transition);
}

.lang-mini button.active {
  background: var(--blue-600);
  color: white;
}

.btn-action {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  background: white;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  color: var(--gray-600);
  cursor: pointer;
  transition: var(--transition);
}

.btn-action svg { width: 16px; height: 16px; }
.btn-action:hover { border-color: var(--blue-400); color: var(--blue-600); background: var(--blue-50); }

.drop-zone {
  flex: 1;
  overflow: auto;
  padding: 20px 28px;
  transition: background 0.2s;
}

.drop-zone.dragging {
  background: rgba(37,99,235,0.04);
  box-shadow: inset 0 0 0 2px rgba(37,99,235,0.3);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  color: var(--gray-300);
  gap: 12px;
}

.empty-state svg { width: 64px; height: 64px; }
.empty-state p { font-size: 16px; font-weight: 500; color: var(--gray-400); }
.empty-state span { font-size: 13px; color: var(--gray-300); }

.loading-state {
  display: flex;
  justify-content: center;
  padding-top: 80px;
}

.spinner-lg {
  width: 32px;
  height: 32px;
  border: 3px solid var(--gray-200);
  border-top-color: var(--blue-500);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.file-table {
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
  border: 1px solid var(--gray-100);
}

.file-header {
  display: grid;
  grid-template-columns: 1fr 100px 160px 60px 140px;
  padding: 12px 20px;
  background: var(--gray-50);
  border-bottom: 1px solid var(--gray-100);
  font-size: 12px;
  font-weight: 600;
  color: var(--gray-400);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.file-row {
  display: grid;
  grid-template-columns: 1fr 100px 160px 60px 140px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--gray-50);
  cursor: pointer;
  transition: var(--transition);
  align-items: center;
}

.file-row:last-child { border-bottom: none; }
.file-row:hover { background: var(--blue-50); }

.col-name {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.file-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.file-icon.folder-icon { background: rgba(251,191,36,0.15); color: #F59E0B; }
.file-icon.img { background: rgba(16,185,129,0.12); color: #10B981; }
.file-icon.video { background: rgba(139,92,246,0.12); color: #8B5CF6; }
.file-icon.audio { background: rgba(244,63,94,0.12); color: #F43F5E; }
.file-icon.archive { background: rgba(245,158,11,0.12); color: #F59E0B; }
.file-icon.pdf { background: rgba(239,68,68,0.12); color: #EF4444; }
.file-icon.code { background: rgba(59,130,246,0.12); color: #3B82F6; }
.file-icon.text, .file-icon.default { background: var(--gray-100); color: var(--gray-400); }

.file-icon svg { width: 16px; height: 16px; }

.file-name {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--gray-700);
}

.badge-public {
  padding: 2px 8px;
  background: rgba(16,185,129,0.12);
  color: #10B981;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.col-size, .col-date {
  font-size: 13px;
  color: var(--gray-400);
}

.col-actions {
  display: flex;
  justify-content: flex-end;
}

.action-menu {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
}

.file-row:hover .action-menu { opacity: 1; }

.act-btn {
  width: 30px;
  height: 30px;
  border: none;
  background: none;
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gray-400);
  transition: var(--transition);
}

.act-btn svg { width: 15px; height: 15px; }
.act-btn:hover { background: var(--gray-100); color: var(--blue-600); }
.act-btn.danger:hover { background: rgba(239,68,68,0.1); color: #EF4444; }

/* Modals */
.modal-bg {
  position: fixed;
  inset: 0;
  background: rgba(15,23,42,0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: white;
  border-radius: 20px;
  padding: 32px;
  width: 440px;
  max-width: 90vw;
  box-shadow: var(--shadow-lg);
  animation: modalIn 0.2s cubic-bezier(0.4,0,0.2,1);
}

.modal-lg { width: 560px; }

@keyframes modalIn {
  from { opacity: 0; transform: scale(0.95) translateY(8px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--gray-800);
  margin-bottom: 20px;
}

.modal-desc {
  font-size: 14px;
  color: var(--gray-500);
  margin-bottom: 24px;
}

.field {
  margin-bottom: 16px;
}

.field label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--gray-600);
  margin-bottom: 8px;
}

.field input, .field textarea {
  width: 100%;
  padding: 10px 14px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-family: inherit;
  color: var(--gray-800);
  outline: none;
  resize: vertical;
  transition: var(--transition);
}

.field input:focus, .field textarea:focus {
  border-color: var(--blue-500);
  box-shadow: 0 0 0 3px rgba(59,130,246,0.1);
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 24px;
}

.btn-ghost {
  padding: 10px 20px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--gray-600);
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: var(--transition);
}

.btn-ghost:hover { background: var(--gray-50); }

.btn-primary-sm {
  padding: 10px 20px;
  background: linear-gradient(135deg, #2563EB, #0EA5E9);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: var(--transition);
  box-shadow: 0 4px 12px rgba(37,99,235,0.25);
}

.btn-primary-sm:hover { transform: translateY(-1px); box-shadow: 0 6px 20px rgba(37,99,235,0.3); }

.btn-danger {
  padding: 10px 20px;
  background: #EF4444;
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: var(--transition);
}

.modal-close {
  margin-top: 16px;
  padding: 8px 16px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--gray-500);
  font-size: 14px;
  font-family: inherit;
  cursor: pointer;
  width: 100%;
  transition: var(--transition);
}

.modal-close:hover { background: var(--gray-50); }

.upload-area {
  border: 2px dashed var(--gray-200);
  border-radius: var(--radius);
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: var(--transition);
  margin-bottom: 16px;
}

.upload-area:hover { border-color: var(--blue-400); background: var(--blue-50); }
.upload-area svg { width: 40px; height: 40px; color: var(--gray-300); margin-bottom: 12px; }
.upload-area p { font-size: 14px; color: var(--gray-400); }

.upload-list {
  margin-bottom: 16px;
  max-height: 160px;
  overflow: auto;
}

.upload-item {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
  color: var(--gray-600);
  border-bottom: 1px solid var(--gray-50);
}

.upload-item .done { color: #10B981; }
.upload-item .pending { color: var(--gray-400); }

.share-link-box {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.share-input {
  flex: 1;
  padding: 10px 14px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-family: 'JetBrains Mono', monospace;
  color: var(--gray-600);
  background: var(--gray-50);
  outline: none;
}

.btn-copy {
  padding: 10px 16px;
  background: var(--blue-600);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  white-space: nowrap;
  transition: var(--transition);
}

.share-raw {
  font-size: 12px;
  color: var(--gray-400);
  margin-bottom: 16px;
}

.share-raw a { color: var(--blue-500); text-decoration: none; }
</style>
