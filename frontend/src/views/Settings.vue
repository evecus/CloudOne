<template>
  <Layout>
    <div class="page-container">
      <div class="page-header">
        <h2 class="page-title">{{ t.settings }}</h2>
        <div class="lang-mini">
          <button :class="{ active: lang === 'zh' }" @click="switchLang('zh')">中文</button>
          <button :class="{ active: lang === 'en' }" @click="switchLang('en')">EN</button>
        </div>
      </div>

      <div class="content">
        <!-- Logo section -->
        <div class="settings-logo">
          <div class="big-icon">
            <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
              <defs>
                <linearGradient id="slg2" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" style="stop-color:#2563EB"/>
                  <stop offset="100%" style="stop-color:#38BDF8"/>
                </linearGradient>
              </defs>
              <ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#slg2)"/>
              <circle cx="36" cy="56" r="16" fill="url(#slg2)"/>
              <circle cx="58" cy="50" r="20" fill="url(#slg2)"/>
              <polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/>
            </svg>
          </div>
          <div>
            <h3>CloudOne</h3>
            <p>v1.0.0</p>
          </div>
        </div>

        <!-- Token 过期提示 -->
        <div v-if="auth.isTokenExpiringSoon" class="expiry-banner">
          <span>⚠️ {{ lang === 'zh' ? '登录状态将在 1 小时内过期，建议重新登录。' : 'Your session expires in less than 1 hour. Consider re-logging in.' }}</span>
        </div>

        <!-- Profile -->
        <div class="settings-card">
          <h4 class="card-title">{{ t.profile }}</h4>
          <div class="form-grid">
            <div class="field">
              <label>{{ t.username }}</label>
              <input v-model="userForm.username" type="text" />
            </div>
            <div class="field">
              <label>{{ t.password }}</label>
              <input
                v-model="userForm.password"
                type="password"
                :placeholder="lang === 'zh' ? '留空则不修改' : 'Leave blank to keep'"
              />
            </div>
            <div class="field">
              <label>{{ lang === 'zh' ? '确认新密码' : 'Confirm new password' }}</label>
              <input
                v-model="userForm.passwordConfirm"
                type="password"
                :placeholder="lang === 'zh' ? '再次输入新密码' : 'Re-enter new password'"
              />
            </div>
          </div>
          <p v-if="userError" class="field-error">{{ userError }}</p>
          <button class="btn-save" @click="saveUser">
            <span v-if="!savedUser">{{ t.save }}</span>
            <span v-else>✓ {{ t.saved }}</span>
          </button>
        </div>

        <!-- Storage -->
        <div class="settings-card">
          <h4 class="card-title">{{ t.storageDir }}</h4>
          <div class="field">
            <input v-model="settingsForm.storageDir" type="text" />
          </div>
          <button class="btn-save" @click="saveSettings">
            <span v-if="!savedSettings">{{ t.save }}</span>
            <span v-else>✓ {{ t.saved }}</span>
          </button>
        </div>

        <!-- WebDAV -->
        <div class="settings-card">
          <h4 class="card-title">{{ t.webdavSettings }}</h4>

          <div class="webdav-toggle-row">
            <div class="webdav-toggle-info">
              <span class="toggle-label">{{ t.webdavEnabled }}</span>
              <span class="toggle-desc">{{ t.webdavEnabledDesc }}</span>
            </div>
            <button class="toggle-btn" :class="{ on: webdavForm.enabled }" @click="webdavForm.enabled = !webdavForm.enabled">
              <span class="toggle-knob"></span>
            </button>
          </div>

          <div v-if="webdavForm.enabled" class="webdav-fields">
            <div class="field">
              <label>{{ t.webdavSubPath }}</label>
              <p class="field-desc">{{ t.webdavSubPathDesc }}</p>
              <input v-model="webdavForm.subPath" type="text" :placeholder="t.webdavSubPathPlaceholder" />
            </div>
            <div class="field">
              <label>{{ t.webdavPassword }}</label>
              <p class="field-desc">{{ t.webdavPasswordDesc }}</p>
              <!-- 密码字段永远不回显，始终为空，表示"不修改" -->
              <input
                v-model="webdavForm.password"
                type="password"
                :placeholder="webdavHasPassword
                  ? (lang === 'zh' ? '留空则保持原密码不变' : 'Leave blank to keep current password')
                  : (lang === 'zh' ? '设置 WebDAV 密码' : 'Set WebDAV password')"
                autocomplete="new-password"
              />
              <p v-if="webdavHasPassword" class="field-hint">
                {{ lang === 'zh' ? '✓ 已设置密码（加密存储）' : '✓ Password is set (encrypted)' }}
              </p>
            </div>
            <div class="webdav-addr-preview">
              <span class="addr-label">{{ t.webdavAddress }}</span>
              <code class="addr-value">{{ davUrl }}</code>
            </div>
          </div>

          <button class="btn-save" @click="saveWebDAV">
            <span v-if="!savedWebDAV">{{ t.save }}</span>
            <span v-else>✓ {{ t.saved }}</span>
          </button>
        </div>

        <!-- SSH -->
        <div class="settings-card">
          <h4 class="card-title">{{ lang === 'zh' ? 'SSH 连接' : 'SSH Connection' }}</h4>
          <p class="card-desc">{{ lang === 'zh' ? '配置 SSH 终端连接信息，密码和私钥加密保存' : 'Configure SSH terminal connection. Credentials are encrypted.' }}</p>

          <div class="ssh-fields">
            <div class="ssh-row">
              <div class="field ssh-field-host">
                <label>{{ lang === 'zh' ? '主机地址' : 'Host' }}</label>
                <input v-model="sshForm.host" type="text" placeholder="127.0.0.1" />
              </div>
              <div class="field ssh-field-port">
                <label>{{ lang === 'zh' ? '端口' : 'Port' }}</label>
                <input v-model.number="sshForm.port" type="number" placeholder="22" min="1" max="65535" />
              </div>
              <div class="field ssh-field-user">
                <label>{{ lang === 'zh' ? '用户名' : 'Username' }}</label>
                <input v-model="sshForm.user" type="text" placeholder="root" autocomplete="off" />
              </div>
            </div>

            <div class="field">
              <label>{{ lang === 'zh' ? '认证方式' : 'Auth Type' }}</label>
              <div class="ssh-auth-tabs">
                <button
                  class="ssh-auth-tab"
                  :class="{ active: sshForm.authType === 'password' }"
                  @click="sshForm.authType = 'password'">
                  {{ lang === 'zh' ? '密码' : 'Password' }}
                </button>
                <button
                  class="ssh-auth-tab"
                  :class="{ active: sshForm.authType === 'key' }"
                  @click="sshForm.authType = 'key'">
                  {{ lang === 'zh' ? '私钥' : 'Private Key' }}
                </button>
              </div>
            </div>

            <div v-if="sshForm.authType === 'password'" class="field">
              <label>{{ lang === 'zh' ? '密码' : 'Password' }}</label>
              <input
                v-model="sshForm.password"
                type="password"
                :placeholder="sshHasPassword
                  ? (lang === 'zh' ? '留空则保持原密码不变' : 'Leave blank to keep current')
                  : (lang === 'zh' ? '输入 SSH 密码' : 'Enter SSH password')"
                autocomplete="new-password"
              />
              <p v-if="sshHasPassword" class="field-hint">✓ {{ lang === 'zh' ? '已设置密码（加密存储）' : 'Password is set (encrypted)' }}</p>
            </div>

            <div v-if="sshForm.authType === 'key'" class="field">
              <label>{{ lang === 'zh' ? '私钥内容' : 'Private Key' }}</label>
              <textarea
                v-model="sshForm.privateKey"
                class="ssh-key-textarea"
                :placeholder="sshHasKey
                  ? (lang === 'zh' ? '留空则保持原私钥不变' : 'Leave blank to keep current key')
                  : (lang === 'zh' ? '粘贴 PEM 格式私钥（-----BEGIN ...）' : 'Paste PEM private key (-----BEGIN ...)')"
                autocomplete="off"
                spellcheck="false"
              ></textarea>
              <p v-if="sshHasKey" class="field-hint">✓ {{ lang === 'zh' ? '已设置私钥（加密存储）' : 'Private key is set (encrypted)' }}</p>
            </div>
          </div>

          <button class="btn-save" @click="saveSSH">
            <span v-if="!savedSSH">{{ t.save }}</span>
            <span v-else>✓ {{ t.saved }}</span>
          </button>
        </div>

        <!-- Theme -->
        <div class="settings-card">
          <h4 class="card-title">{{ lang === 'zh' ? '界面主题' : 'Theme' }}</h4>
          <p class="card-desc">{{ lang === 'zh' ? '选择您喜欢的主题色彩' : 'Choose your preferred color theme' }}</p>
          <div class="theme-grid">
            <button
              v-for="theme in themes"
              :key="theme.id"
              class="theme-swatch"
              :class="{ active: currentTheme === theme.id }"
              @click="applyTheme(theme.id)"
              :title="theme.label"
            >
              <span class="swatch-preview" :style="{ background: theme.gradient }"></span>
              <span class="swatch-name">{{ theme.label }}</span>
              <span v-if="currentTheme === theme.id" class="swatch-check">✓</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Layout from '../components/Layout.vue'
import api from '../api'
import { t, currentLang as lang, setLang } from '../i18n'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const userForm = ref({ username: '', password: '', passwordConfirm: '' })
const settingsForm = ref({ storageDir: '' })
const savedUser = ref(false)
const savedSettings = ref(false)
const savedWebDAV = ref(false)
const userError = ref('')
const currentTheme = ref(localStorage.getItem('theme') || 'blue')

// WebDAV：只记录是否有密码，不存储密码内容
const webdavHasPassword = ref(false)
const webdavForm = ref({ enabled: false, subPath: '', password: '' })
const davUrl = computed(() => `${window.location.origin}/dav/`)

// SSH
const savedSSH = ref(false)
const sshHasPassword = ref(false)
const sshHasKey = ref(false)
const sshForm = ref({ host: '', port: 22, user: '', authType: 'password', password: '', privateKey: '' })

const themes = [
  { id: 'blue',   label: lang.value === 'zh' ? '蓝白（默认）' : 'Blue (Default)', gradient: 'linear-gradient(135deg, #2563EB, #0EA5E9)' },
  { id: 'rose',   label: lang.value === 'zh' ? '玫瑰粉' : 'Rose Pink',   gradient: 'linear-gradient(135deg, #E11D48, #FB7185)' },
  { id: 'forest', label: lang.value === 'zh' ? '森绿' : 'Forest Green', gradient: 'linear-gradient(135deg, #16A34A, #4ADE80)' },
  { id: 'amber',  label: lang.value === 'zh' ? '暖橙' : 'Warm Amber',   gradient: 'linear-gradient(135deg, #D97706, #FBBF24)' },
  { id: 'violet', label: lang.value === 'zh' ? '紫罗兰' : 'Violet',      gradient: 'linear-gradient(135deg, #7C3AED, #A78BFA)' },
  { id: 'teal',   label: lang.value === 'zh' ? '碳岩青' : 'Teal',        gradient: 'linear-gradient(135deg, #0D9488, #2DD4BF)' },
  { id: 'slate',  label: lang.value === 'zh' ? '石墨灰' : 'Slate Gray',  gradient: 'linear-gradient(135deg, #334155, #64748B)' },
]

function applyTheme(id) {
  currentTheme.value = id
  localStorage.setItem('theme', id)
  document.documentElement.setAttribute('data-theme', id)
}

function switchLang(l) {
  setLang(l)
}

async function load() {
  if (auth.user) {
    userForm.value.username = auth.user.username
  }
  const { data } = await api.get('/settings')
  settingsForm.value.storageDir = data.storage_dir

  const { data: wdata } = await api.get('/webdav/settings')
  webdavForm.value = {
    enabled: wdata.webdav_enabled,
    subPath: wdata.webdav_sub_path,
    password: '', // 永不回显密码
  }
  webdavHasPassword.value = wdata.webdav_has_password

  // SSH
  const { data: sdata } = await api.get('/ssh/settings')
  sshForm.value = {
    host: sdata.ssh_host || '127.0.0.1',
    port: sdata.ssh_port || 22,
    user: sdata.ssh_user || 'root',
    authType: sdata.ssh_auth_type || 'password',
    password: '',
    privateKey: '',
  }
  sshHasPassword.value = sdata.ssh_has_password
  sshHasKey.value = sdata.ssh_has_key
}

async function saveWebDAV() {
  await api.put('/webdav/settings', {
    webdav_enabled: webdavForm.value.enabled,
    webdav_sub_path: webdavForm.value.subPath,
    webdav_password: webdavForm.value.password, // 空字符串=不修改
  })
  webdavForm.value.password = '' // 保存后清空输入框
  savedWebDAV.value = true
  // 重新加载，更新 has_password 状态
  const { data: wdata } = await api.get('/webdav/settings')
  webdavHasPassword.value = wdata.webdav_has_password
  setTimeout(() => savedWebDAV.value = false, 2000)
}

async function saveUser() {
  userError.value = ''
  // 如果填写了新密码，校验两次输入是否一致
  if (userForm.value.password) {
    if (userForm.value.password !== userForm.value.passwordConfirm) {
      userError.value = lang.value === 'zh' ? '两次输入的密码不一致' : 'Passwords do not match'
      return
    }
    if (userForm.value.password.length < 6) {
      userError.value = lang.value === 'zh' ? '密码至少 6 位' : 'Password must be at least 6 characters'
      return
    }
  }

  const payload = { username: userForm.value.username }
  if (userForm.value.password) payload.password = userForm.value.password

  const { data } = await api.put('/user', payload)
  auth.user = data.user
  // 后端下发了新 token（密码已修改），前端更新存储
  if (data.token) {
    auth.updateToken(data.token)
  }
  userForm.value.password = ''
  userForm.value.passwordConfirm = ''
  savedUser.value = true
  setTimeout(() => savedUser.value = false, 2000)
}

async function saveSettings() {
  await api.put('/settings', { storage_dir: settingsForm.value.storageDir, lang: lang.value })
  savedSettings.value = true
  setTimeout(() => savedSettings.value = false, 2000)
}

async function saveSSH() {
  const payload = {
    ssh_host:     sshForm.value.host || '127.0.0.1',
    ssh_port:     sshForm.value.port || 22,
    ssh_user:     sshForm.value.user || 'root',
    ssh_auth_type: sshForm.value.authType,
  }
  if (sshForm.value.authType === 'password' && sshForm.value.password) {
    payload.ssh_password = sshForm.value.password
  }
  if (sshForm.value.authType === 'key' && sshForm.value.privateKey) {
    payload.ssh_private_key = sshForm.value.privateKey
  }
  await api.put('/ssh/settings', payload)
  sshForm.value.password = ''
  sshForm.value.privateKey = ''
  savedSSH.value = true
  const { data: sdata } = await api.get('/ssh/settings')
  sshHasPassword.value = sdata.ssh_has_password
  sshHasKey.value = sdata.ssh_has_key
  setTimeout(() => savedSSH.value = false, 2000)
}

onMounted(load)
</script>

<style scoped>
.page-container { flex: 1; display: flex; flex-direction: column; }
.page-header { padding: 20px 28px 16px; border-bottom: 1px solid var(--gray-100); background: white; display: flex; align-items: center; justify-content: space-between; }
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); }
.content { flex: 1; padding: 28px; overflow: auto; max-width: 680px; }

.lang-mini { display: flex; border: 1px solid var(--gray-200); border-radius: 8px; overflow: hidden; }
.lang-mini button { padding: 6px 12px; border: none; background: transparent; font-size: 12px; font-family: inherit; cursor: pointer; color: var(--gray-500); transition: var(--transition); }
.lang-mini button.active { background: var(--blue-600); color: white; }

/* 过期提示横幅 */
.expiry-banner {
  background: #FEF3C7;
  border: 1px solid #F59E0B;
  border-radius: var(--radius-sm);
  padding: 10px 16px;
  margin-bottom: 20px;
  font-size: 13px;
  color: #92400E;
}

.settings-logo {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 32px;
  padding: 24px;
  background: var(--primary-gradient-soft);
  border-radius: var(--radius-lg);
  border: 1px solid var(--blue-100);
}

.big-icon { width: 72px; height: 72px; filter: drop-shadow(0 4px 16px rgba(37,99,235,0.25)); }
.settings-logo h3 { font-size: 24px; font-weight: 700; background: var(--primary-gradient); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.settings-logo p { font-size: 13px; color: var(--gray-400); margin-top: 4px; }

.settings-card {
  background: white;
  border: 1px solid var(--gray-100);
  border-radius: var(--radius-lg);
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: var(--shadow-sm);
}

.card-title { font-size: 15px; font-weight: 600; color: var(--gray-700); margin-bottom: 20px; }

.form-grid { display: grid; gap: 16px; }

.field { margin-bottom: 0; }
.field label { display: block; font-size: 13px; font-weight: 500; color: var(--gray-600); margin-bottom: 8px; }
.field input { width: 100%; padding: 10px 14px; border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm); font-size: 14px; font-family: inherit; color: var(--gray-800); outline: none; transition: var(--transition); }
.field input:focus { border-color: var(--blue-500); box-shadow: 0 0 0 3px rgba(59,130,246,0.1); }

.field-error { color: #EF4444; font-size: 12px; margin-top: 8px; }
.field-hint { color: #16A34A; font-size: 12px; margin-top: 6px; }

.btn-save {
  margin-top: 20px;
  padding: 10px 24px;
  background: var(--primary-gradient);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: var(--transition);
  box-shadow: var(--primary-shadow);
}
.btn-save:hover { transform: translateY(-1px); box-shadow: var(--primary-shadow-hover); }

.card-desc { font-size: 13px; color: var(--gray-400); margin-bottom: 16px; margin-top: -12px; }

.theme-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 10px;
}
.theme-swatch {
  display: flex; align-items: center; gap: 10px; padding: 10px 14px;
  border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm);
  background: white; cursor: pointer; transition: var(--transition); font-family: inherit; position: relative;
}
.theme-swatch:hover { border-color: var(--blue-400); background: var(--blue-50); }
.theme-swatch.active { border-color: var(--blue-500); background: var(--blue-50); box-shadow: 0 0 0 3px var(--blue-100); }
.swatch-preview { width: 20px; height: 20px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 2px 6px rgba(0,0,0,0.15); }
.swatch-name { font-size: 13px; font-weight: 500; color: var(--gray-700); flex: 1; }
.swatch-check { font-size: 12px; font-weight: 700; color: var(--blue-600); }

.webdav-toggle-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 16px; flex-wrap: wrap; }
.webdav-toggle-info { flex: 1; min-width: 0; }
.toggle-label { font-size: 14px; font-weight: 600; color: var(--gray-700); display: block; }
.toggle-desc { font-size: 12px; color: var(--gray-400); margin-top: 3px; display: block; }
.toggle-btn { width: 44px; height: 24px; border-radius: 12px; background: var(--gray-200); border: none; cursor: pointer; position: relative; transition: background 0.2s; flex-shrink: 0; }
.toggle-btn.on { background: var(--blue-500); }
.toggle-knob { position: absolute; top: 3px; left: 3px; width: 18px; height: 18px; border-radius: 50%; background: white; transition: transform 0.2s; box-shadow: 0 1px 4px rgba(0,0,0,.2); }
.toggle-btn.on .toggle-knob { transform: translateX(20px); }
.webdav-fields { display: flex; flex-direction: column; gap: 16px; margin-bottom: 4px; }
.field-desc { font-size: 12px; color: var(--gray-400); margin: -4px 0 8px; }
.webdav-addr-preview { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; padding: 10px 14px; background: var(--blue-50); border: 1px solid var(--blue-100); border-radius: var(--radius-sm); }
.addr-label { font-size: 12px; font-weight: 600; color: var(--gray-500); flex-shrink: 0; }
.addr-value { font-family: 'JetBrains Mono', monospace; font-size: 12px; color: var(--blue-700); word-break: break-all; }

@media (max-width: 768px) {
  .page-header { padding: 14px 16px; }
  .content { padding: 16px; }
  .settings-card { padding: 16px; }
  .theme-grid { grid-template-columns: repeat(2, 1fr); }
  .webdav-addr-preview { flex-direction: column; align-items: flex-start; gap: 4px; }
}

/* SSH */
.ssh-fields { display: flex; flex-direction: column; gap: 16px; margin-bottom: 4px; }
.ssh-row { display: flex; gap: 12px; }
.ssh-field-host { flex: 3; }
.ssh-field-port { flex: 1; min-width: 80px; }
.ssh-field-user { flex: 2; }
.ssh-auth-tabs { display: flex; gap: 6px; }
.ssh-auth-tab { padding: 6px 16px; border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm); background: white; font-size: 13px; font-weight: 500; color: var(--gray-600); cursor: pointer; transition: var(--transition); }
.ssh-auth-tab.active { border-color: var(--blue-500); color: var(--blue-600); background: var(--blue-50); }
.ssh-key-textarea { width: 100%; height: 120px; padding: 10px 14px; border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm); font-size: 12px; font-family: 'JetBrains Mono', monospace; color: var(--gray-800); outline: none; resize: vertical; transition: var(--transition); }
.ssh-key-textarea:focus { border-color: var(--blue-500); box-shadow: 0 0 0 3px rgba(59,130,246,0.1); }
@media (max-width: 600px) {
  .ssh-row { flex-direction: column; gap: 12px; }
}
</style>
