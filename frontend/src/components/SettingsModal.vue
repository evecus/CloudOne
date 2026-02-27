<template>
  <teleport to="body">
    <div v-if="modelValue" class="modal-bg" @click.self="$emit('update:modelValue', false)">
      <div class="settings-modal">
        <!-- Header -->
        <div class="sm-header">
          <div class="sm-header-left">
            <div class="sm-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="3"/>
                <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/>
              </svg>
            </div>
            <span class="sm-title">{{ lang === 'zh' ? '设置' : 'Settings' }}</span>
          </div>
          <button class="sm-close" @click="$emit('update:modelValue', false)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <!-- Body -->
        <div class="sm-body">

          <!-- 账户信息 -->
          <div class="sm-card">
            <h4 class="sm-card-title">{{ lang === 'zh' ? '账户信息' : 'Account' }}</h4>
            <div class="sm-fields">
              <div class="sm-field">
                <label>{{ lang === 'zh' ? '用户名' : 'Username' }}</label>
                <input v-model="userForm.username" type="text" />
              </div>
              <div class="sm-field">
                <label>{{ lang === 'zh' ? '新密码' : 'New Password' }}</label>
                <input v-model="userForm.password" type="password" :placeholder="lang === 'zh' ? '留空则不修改' : 'Leave blank to keep'" />
              </div>
            </div>
            <div class="sm-actions">
              <button class="sm-btn" @click="saveUser">
                <svg v-if="savedUser" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                <span>{{ savedUser ? (lang === 'zh' ? '已保存' : 'Saved!') : (lang === 'zh' ? '保存' : 'Save') }}</span>
              </button>
            </div>
          </div>

          <!-- 存储目录 -->
          <div class="sm-card">
            <h4 class="sm-card-title">{{ lang === 'zh' ? '存储目录' : 'Storage Directory' }}</h4>
            <div class="sm-field">
              <input v-model="storageForm.dir" type="text" :class="{ changed: storageForm.dir !== originalStorageDir }" />
            </div>
            <div v-if="storageForm.dir !== originalStorageDir" class="sm-warning">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
              <span>{{ lang === 'zh' ? '更换存储目录后，当前目录下所有文件将不再显示，所有分享链接和公开设置将失效！' : 'Changing storage dir will hide all current files and invalidate all shares and public settings!' }}</span>
            </div>
            <div class="sm-actions">
              <button class="sm-btn" :class="{ danger: storageForm.dir !== originalStorageDir }" @click="saveStorage">
                <svg v-if="savedStorage" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                <span>{{ savedStorage ? (lang === 'zh' ? '已保存' : 'Saved!') : (storageForm.dir !== originalStorageDir ? (lang === 'zh' ? '⚠ 确认更换' : '⚠ Confirm Change') : (lang === 'zh' ? '保存' : 'Save')) }}</span>
              </button>
            </div>
          </div>

          <!-- 界面主题 -->
          <div class="sm-card">
            <h4 class="sm-card-title">{{ lang === 'zh' ? '界面主题' : 'Theme' }}</h4>
            <p class="sm-card-desc">{{ lang === 'zh' ? '选择您喜欢的主题色彩，点击保存生效' : 'Choose a color theme and click Save to apply' }}</p>
            <div class="theme-grid">
              <button v-for="theme in themes" :key="theme.id" class="theme-swatch" :class="{ active: pendingTheme === theme.id }" @click="pendingTheme = theme.id" :title="theme.label">
                <span class="swatch-dot" :style="{ background: theme.gradient }"></span>
                <span class="swatch-name">{{ theme.label }}</span>
                <span v-if="pendingTheme === theme.id" class="swatch-check">✓</span>
              </button>
            </div>
            <div class="sm-actions">
              <button class="sm-btn" @click="saveTheme">
                <svg v-if="savedTheme" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                <span>{{ savedTheme ? (lang === 'zh' ? '已应用' : 'Applied!') : (lang === 'zh' ? '应用主题' : 'Apply Theme') }}</span>
              </button>
            </div>
          </div>

          <!-- 字体设置 -->
          <div class="sm-card">
            <h4 class="sm-card-title">{{ lang === 'zh' ? '字体设置' : 'Font Settings' }}</h4>
            <div class="sm-fields">
              <div class="sm-field">
                <label>{{ lang === 'zh' ? '界面字体' : 'UI Font' }}</label>
                <select v-model="fontForm.ui">
                  <option v-for="f in uiFonts" :key="f.id" :value="f.id">{{ f.label }}</option>
                </select>
                <p class="font-preview" :style="{ fontFamily: getFontFamily(fontForm.ui, 'ui') }">{{ lang === 'zh' ? '字体预览 AaBbCc 你好世界' : 'Font preview AaBbCc Hello World' }}</p>
              </div>
              <div class="sm-field">
                <label>{{ lang === 'zh' ? '编辑器字体' : 'Editor Font' }}</label>
                <select v-model="fontForm.editor">
                  <option v-for="f in editorFonts" :key="f.id" :value="f.id">{{ f.label }}</option>
                </select>
                <p class="font-preview mono" :style="{ fontFamily: getFontFamily(fontForm.editor, 'editor') }">const hello = "world"; // 字体预览</p>
              </div>
            </div>
            <div class="sm-actions">
              <button class="sm-btn" @click="saveFont">
                <svg v-if="savedFont" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                <span>{{ savedFont ? (lang === 'zh' ? '已应用' : 'Applied!') : (lang === 'zh' ? '应用字体' : 'Apply Font') }}</span>
              </button>
            </div>
          </div>

          <!-- WebDAV -->
          <div class="sm-card">
            <h4 class="sm-card-title">WebDAV</h4>
            <div class="sm-fields">
              <div class="sm-field">
                <label style="display:flex;align-items:center;gap:10px;cursor:pointer">
                  <span>{{ lang === 'zh' ? '启用 WebDAV' : 'Enable WebDAV' }}</span>
                  <div class="toggle-wrap" @click="webdavForm.enabled = !webdavForm.enabled">
                    <div class="toggle" :class="{ on: webdavForm.enabled }">
                      <div class="toggle-knob"></div>
                    </div>
                  </div>
                </label>
              </div>
              <div class="sm-field">
                <label>{{ lang === 'zh' ? '存储子目录（相对于存储根目录）' : 'Storage subdirectory (relative to root)' }}</label>
                <input v-model="webdavForm.subPath" type="text" placeholder="webdav" />
              </div>
              <div class="sm-field">
                <label>{{ lang === 'zh' ? 'WebDAV 独立用户名（留空则使用 CloudOne 账户）' : 'WebDAV username (blank = use CloudOne account)' }}</label>
                <input v-model="webdavForm.username" type="text" :placeholder="lang === 'zh' ? '留空则使用 CloudOne 账户名' : 'Leave blank to use account username'" />
              </div>
              <div class="sm-field">
                <label>{{ lang === 'zh' ? 'WebDAV 独立密码（留空则使用账户密码）' : 'WebDAV password (blank = use account password)' }}</label>
                <input v-model="webdavForm.password" type="password" :placeholder="lang === 'zh' ? '留空则使用账户密码' : 'Leave blank to use account password'" />
              </div>
            </div>
            <div class="sm-actions">
              <button class="sm-btn" @click="saveWebDAV">
                <svg v-if="savedWebDAV" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                <span>{{ savedWebDAV ? (lang === 'zh' ? '已保存' : 'Saved!') : (lang === 'zh' ? '保存' : 'Save') }}</span>
              </button>
            </div>
          </div>

          <!-- 语言 -->
          <div class="sm-card">
            <h4 class="sm-card-title">{{ lang === 'zh' ? '语言 / Language' : 'Language / 语言' }}</h4>
            <div class="lang-btns">
              <button :class="{ active: lang === 'zh' }" @click="switchLang('zh')">中文</button>
              <button :class="{ active: lang === 'en' }" @click="switchLang('en')">English</button>
            </div>
          </div>

        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '../api'
import { currentLang as lang, setLang } from '../i18n'
import { useAuthStore } from '../stores/auth'

const props = defineProps({ modelValue: Boolean })
const emit = defineEmits(['update:modelValue', 'storage-changed'])
const auth = useAuthStore()

// User form
const userForm = ref({ username: '', password: '' })
const savedUser = ref(false)

// Storage form
const storageForm = ref({ dir: '' })
const originalStorageDir = ref('')
const savedStorage = ref(false)

// Theme
const themes = [
  { id: 'blue',   label: lang.value === 'zh' ? '蓝白' : 'Blue',   gradient: 'linear-gradient(135deg,#2563EB,#0EA5E9)' },
  { id: 'rose',   label: lang.value === 'zh' ? '玫瑰粉' : 'Rose Pink',    gradient: 'linear-gradient(135deg,#E11D48,#FB7185)' },
  { id: 'forest', label: lang.value === 'zh' ? '森绿' : 'Forest Green',   gradient: 'linear-gradient(135deg,#16A34A,#4ADE80)' },
  { id: 'amber',  label: lang.value === 'zh' ? '暖橙' : 'Warm Amber',     gradient: 'linear-gradient(135deg,#D97706,#FBBF24)' },
  { id: 'violet', label: lang.value === 'zh' ? '紫罗兰' : 'Violet',       gradient: 'linear-gradient(135deg,#7C3AED,#A78BFA)' },
  { id: 'teal',   label: lang.value === 'zh' ? '碳岩青' : 'Teal',         gradient: 'linear-gradient(135deg,#0D9488,#2DD4BF)' },
  { id: 'slate',  label: lang.value === 'zh' ? '石墨灰' : 'Slate Gray',   gradient: 'linear-gradient(135deg,#334155,#64748B)' },
]
const pendingTheme = ref(localStorage.getItem('theme') || 'blue')
const savedTheme = ref(false)

// Font
const uiFonts = [
  { id: 'sora',     label: 'Sora（默认）',     family: "'Sora', -apple-system, sans-serif" },
  { id: 'inter',    label: 'Inter',            family: "'Inter', -apple-system, sans-serif" },
  { id: 'roboto',   label: 'Roboto',           family: "'Roboto', -apple-system, sans-serif" },
  { id: 'noto',     label: 'Noto Sans SC',     family: "'Noto Sans SC', -apple-system, sans-serif" },
  { id: 'lato',     label: 'Lato',             family: "'Lato', -apple-system, sans-serif" },
  { id: 'system',   label: lang.value === 'zh' ? '系统默认' : 'System Default', family: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif" },
]
const editorFonts = [
  { id: 'jetbrains', label: 'JetBrains Mono（默认）', family: "'JetBrains Mono', 'Courier New', monospace" },
  { id: 'firacode',  label: 'Fira Code',              family: "'Fira Code', 'Courier New', monospace" },
  { id: 'sourcecp',  label: 'Source Code Pro',        family: "'Source Code Pro', 'Courier New', monospace" },
  { id: 'ibmplex',   label: 'IBM Plex Mono',          family: "'IBM Plex Mono', 'Courier New', monospace" },
  { id: 'consolas',  label: 'Consolas',               family: "Consolas, 'Courier New', monospace" },
  { id: 'courier',   label: 'Courier New',            family: "'Courier New', monospace" },
]
const fontForm = ref({
  ui: localStorage.getItem('uiFont') || 'sora',
  editor: localStorage.getItem('editorFont') || 'jetbrains',
})
const savedFont = ref(false)

function getFontFamily(id, type) {
  const list = type === 'ui' ? uiFonts : editorFonts
  return list.find(f => f.id === id)?.family || ''
}

function switchLang(l) {
  setLang(l)
}

// WebDAV
const webdavForm = ref({ enabled: false, subPath: '', username: '', password: '' })
const savedWebDAV = ref(false)

async function loadWebDAV() {
  try {
    const { data } = await api.get('/webdav/settings')
    webdavForm.value.enabled = data.webdav_enabled
    webdavForm.value.subPath = data.webdav_sub_path || ''
    webdavForm.value.username = data.webdav_username || ''
    webdavForm.value.password = ''
  } catch {}
}

async function saveWebDAV() {
  await api.put('/webdav/settings', {
    webdav_enabled: webdavForm.value.enabled,
    webdav_sub_path: webdavForm.value.subPath,
    webdav_username: webdavForm.value.username,
    webdav_password: webdavForm.value.password,
  })
  savedWebDAV.value = true
  setTimeout(() => savedWebDAV.value = false, 2000)
}

function applyFont(uiId, editorId) {
  const uiFamily = getFontFamily(uiId, 'ui')
  const editorFamily = getFontFamily(editorId, 'editor')
  if (uiFamily) document.documentElement.style.setProperty('--ui-font', uiFamily)
  if (editorFamily) document.documentElement.style.setProperty('--editor-font', editorFamily)
}

async function loadData() {
  if (auth.user) {
    userForm.value.username = auth.user.username
  }
  try {
    const { data } = await api.get('/settings')
    storageForm.value.dir = data.storage_dir
    originalStorageDir.value = data.storage_dir

    // 主题：优先服务端，回退 localStorage
    const theme = data.ui_theme || localStorage.getItem('theme') || 'blue'
    pendingTheme.value = theme
    document.documentElement.setAttribute('data-theme', theme === 'blue' ? '' : theme)
    localStorage.setItem('theme', theme)

    // 字体：优先服务端，回退 localStorage
    const uiFont = data.ui_font || localStorage.getItem('uiFont') || 'sora'
    const editorFont = data.editor_font || localStorage.getItem('editorFont') || 'jetbrains'
    fontForm.value.ui = uiFont
    fontForm.value.editor = editorFont
    applyFont(uiFont, editorFont)
    localStorage.setItem('uiFont', uiFont)
    localStorage.setItem('editorFont', editorFont)
  } catch {}
  await loadWebDAV()
}

async function saveUser() {
  const payload = { username: userForm.value.username }
  if (userForm.value.password) payload.password = userForm.value.password
  const { data } = await api.put('/user', payload)
  auth.user = data
  userForm.value.password = ''
  savedUser.value = true
  setTimeout(() => savedUser.value = false, 2000)
}

async function saveStorage() {
  await api.put('/settings', { storage_dir: storageForm.value.dir, lang: lang.value })
  originalStorageDir.value = storageForm.value.dir
  savedStorage.value = true
  // 关闭设置界面，触发文件列表刷新
  setTimeout(() => {
    savedStorage.value = false
    emit('update:modelValue', false)
    emit('storage-changed')
  }, 800)
}

async function saveTheme() {
  const id = pendingTheme.value
  // 立即本地生效
  document.documentElement.setAttribute('data-theme', id === 'blue' ? '' : id)
  localStorage.setItem('theme', id)
  // 同步到服务端（所有设备同步）
  try { await api.put('/settings', { ui_theme: id }) } catch {}
  savedTheme.value = true
  setTimeout(() => savedTheme.value = false, 2000)
}

async function saveFont() {
  const { ui, editor } = fontForm.value
  // 立即本地生效
  applyFont(ui, editor)
  localStorage.setItem('uiFont', ui)
  localStorage.setItem('editorFont', editor)
  // 同步到服务端（所有设备同步）
  try { await api.put('/settings', { ui_font: ui, editor_font: editor }) } catch {}
  savedFont.value = true
  setTimeout(() => savedFont.value = false, 2000)
}

watch(() => props.modelValue, (v) => {
  if (v) loadData()
})
</script>

<style scoped>
.modal-bg {
  position: fixed; inset: 0;
  background: rgba(15,23,42,0.45);
  backdrop-filter: blur(5px);
  display: flex; align-items: center; justify-content: center;
  z-index: 200;
}

.settings-modal {
  background: white;
  border-radius: 20px;
  width: 580px;
  max-width: 94vw;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 24px 80px rgba(0,0,0,0.18);
  animation: smIn 0.22s cubic-bezier(0.4,0,0.2,1);
}
@keyframes smIn { from { opacity:0; transform: scale(0.94) translateY(12px); } to { opacity:1; transform:scale(1) translateY(0); } }

.sm-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--gray-100);
  flex-shrink: 0;
}
.sm-header-left { display: flex; align-items: center; gap: 10px; }
.sm-icon { width: 32px; height: 32px; background: var(--blue-50); border-radius: 8px; display: flex; align-items: center; justify-content: center; color: var(--blue-600); }
.sm-icon svg { width: 16px; height: 16px; }
.sm-title { font-size: 17px; font-weight: 700; color: var(--gray-800); }

.sm-close {
  width: 32px; height: 32px;
  border: none; background: #FEE2E2; color: #EF4444;
  border-radius: 8px; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: var(--transition);
}
.sm-close:hover { background: #FECACA; }
.sm-close svg { width: 15px; height: 15px; }

.sm-body {
  flex: 1; overflow-y: auto; padding: 20px 24px 24px;
  display: flex; flex-direction: column; gap: 16px;
}

.sm-card {
  background: var(--gray-50);
  border: 1px solid var(--gray-100);
  border-radius: var(--radius);
  padding: 18px 20px;
}

.sm-card-title { font-size: 13px; font-weight: 700; color: var(--gray-700); margin-bottom: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
.sm-card-desc { font-size: 12px; color: var(--gray-400); margin-bottom: 12px; margin-top: -8px; }

.sm-fields { display: flex; flex-direction: column; gap: 12px; }
.sm-field { display: flex; flex-direction: column; gap: 6px; }
.sm-field label { font-size: 12px; font-weight: 600; color: var(--gray-500); }
.sm-field input, .sm-field select {
  padding: 9px 12px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  font-size: 13px; font-family: inherit;
  color: var(--gray-800); background: white; outline: none;
  transition: var(--transition);
}
.sm-field input:focus, .sm-field select:focus { border-color: var(--blue-500); box-shadow: 0 0 0 3px rgba(59,130,246,0.1); }
.sm-field input.changed { border-color: #F59E0B; background: #FFFBEB; }

.font-preview { font-size: 13px; color: var(--gray-500); padding: 6px 10px; background: white; border-radius: 6px; border: 1px solid var(--gray-200); margin-top: 4px; }
.font-preview.mono { font-size: 12px; }

.sm-warning {
  display: flex; align-items: flex-start; gap: 8px;
  padding: 10px 12px; background: #FFF7ED;
  border: 1px solid #FED7AA; border-radius: 8px;
  margin: 8px 0;
  color: #92400E; font-size: 12px; line-height: 1.5;
}
.sm-warning svg { width: 16px; height: 16px; color: #F59E0B; flex-shrink: 0; margin-top: 1px; }

.sm-actions { display: flex; justify-content: flex-end; margin-top: 14px; }
.sm-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 20px;
  background: var(--primary-gradient); color: white;
  border: none; border-radius: var(--radius-sm);
  font-size: 13px; font-weight: 600; font-family: inherit;
  cursor: pointer; transition: var(--transition);
  box-shadow: var(--primary-shadow);
}
.sm-btn:hover { transform: translateY(-1px); }
.sm-btn.danger { background: linear-gradient(135deg, #DC2626, #EF4444); box-shadow: 0 4px 12px rgba(220,38,38,0.3); }
.sm-btn svg { width: 14px; height: 14px; }

/* Theme grid */
.theme-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); gap: 8px; }
.theme-swatch {
  display: flex; align-items: center; gap: 8px;
  padding: 9px 12px; border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm); background: white;
  cursor: pointer; transition: var(--transition); font-family: inherit;
}
.theme-swatch:hover { border-color: var(--blue-400); }
.theme-swatch.active { border-color: var(--blue-500); background: var(--blue-50); box-shadow: 0 0 0 3px var(--blue-100); }
.swatch-dot { width: 18px; height: 18px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 2px 5px rgba(0,0,0,0.15); }
.swatch-name { font-size: 12px; font-weight: 500; color: var(--gray-700); flex: 1; }
.swatch-check { font-size: 11px; font-weight: 700; color: var(--blue-600); }

/* Lang */
.lang-btns { display: flex; gap: 8px; }
.lang-btns button {
  flex: 1; padding: 10px;
  border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm);
  background: white; font-size: 14px; font-weight: 500;
  font-family: inherit; cursor: pointer; color: var(--gray-500);
  transition: var(--transition);
}
.lang-btns button.active { background: var(--primary-gradient); color: white; border-color: transparent; }
.lang-btns button:hover:not(.active) { border-color: var(--blue-400); color: var(--blue-600); }

/* Toggle switch */
.toggle-wrap { display: inline-flex; cursor: pointer; }
.toggle {
  width: 40px; height: 22px; border-radius: 11px;
  background: var(--gray-200); transition: background 0.2s;
  position: relative; flex-shrink: 0;
}
.toggle.on { background: var(--blue-500); }
.toggle-knob {
  position: absolute; top: 3px; left: 3px;
  width: 16px; height: 16px; border-radius: 50%;
  background: white; box-shadow: 0 1px 3px rgba(0,0,0,0.2);
  transition: transform 0.2s;
}
.toggle.on .toggle-knob { transform: translateX(18px); }

@media (max-width: 600px) {
  /* 弹窗从底部弹出，宽度撑满 */
  .modal-bg { align-items: flex-end; }
  .settings-modal {
    width: 100% !important;
    max-width: 100% !important;
    border-radius: 20px 20px 0 0 !important;
    /* dvh 自动排除浏览器 UI，所有浏览器一致；vh 降级兼容 */
    max-height: 92dvh !important;
  }
  .sm-header { padding: 14px 16px 12px; }
  /* iOS safe area：底部避开 home 条 */
  .sm-body { padding: 14px 16px calc(16px + env(safe-area-inset-bottom, 0px)); }
  .sm-section { padding: 14px 16px; gap: 10px; }
  .sm-section-title { font-size: 13px; }
  .sm-btn { padding: 10px 16px; font-size: 13px; }
  /* 防止 iOS 输入框自动缩放 */
  input, textarea { font-size: 16px !important; }
}
</style>
