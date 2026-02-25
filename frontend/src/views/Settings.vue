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

        <!-- Profile -->
        <div class="settings-card">
          <h4 class="card-title">{{ t.profile }}</h4>
          <div class="form-grid">
            <div class="field">
              <label>{{ t.username }}</label>
              <input v-model="userForm.username" type="text" />
            </div>
            <div class="field">
              <label>{{ t.email }}</label>
              <input v-model="userForm.email" type="email" />
            </div>
            <div class="field">
              <label>{{ t.password }}</label>
              <input v-model="userForm.password" type="password" :placeholder="lang === 'zh' ? '留空则不修改' : 'Leave blank to keep'" />
            </div>
          </div>
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
import { ref, onMounted } from 'vue'
import Layout from '../components/Layout.vue'
import api from '../api'
import { t, currentLang as lang, setLang } from '../i18n'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const userForm = ref({ username: '', email: '', password: '' })
const settingsForm = ref({ storageDir: '' })
const savedUser = ref(false)
const savedSettings = ref(false)
const currentTheme = ref(localStorage.getItem('theme') || 'blue')

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
    userForm.value.email = auth.user.email
  }
  const { data } = await api.get('/settings')
  settingsForm.value.storageDir = data.storage_dir
}

async function saveUser() {
  const payload = {
    username: userForm.value.username,
    email: userForm.value.email,
  }
  if (userForm.value.password) payload.password = userForm.value.password
  const { data } = await api.put('/user', payload)
  auth.user = data
  userForm.value.password = ''
  savedUser.value = true
  setTimeout(() => savedUser.value = false, 2000)
}

async function saveSettings() {
  await api.put('/settings', { storage_dir: settingsForm.value.storageDir, lang: lang.value })
  savedSettings.value = true
  setTimeout(() => savedSettings.value = false, 2000)
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
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  background: white;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
  position: relative;
}

.theme-swatch:hover {
  border-color: var(--blue-400);
  background: var(--blue-50);
}

.theme-swatch.active {
  border-color: var(--blue-500);
  background: var(--blue-50);
  box-shadow: 0 0 0 3px var(--blue-100);
}

.swatch-preview {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(0,0,0,0.15);
}

.swatch-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--gray-700);
  flex: 1;
}

.swatch-check {
  font-size: 12px;
  font-weight: 700;
  color: var(--blue-600);
}
</style>
