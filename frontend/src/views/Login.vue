<template>
  <div class="auth-page">
    <!-- Background mesh -->
    <div class="mesh"></div>
    <div class="mesh-2"></div>

    <div class="auth-card">
      <!-- Lang switch top-right -->
      <div class="lang-switch">
        <button :class="{ active: lang === 'zh' }" @click="setLang('zh')">中文</button>
        <button :class="{ active: lang === 'en' }" @click="setLang('en')">EN</button>
      </div>

      <!-- Logo centered -->
      <div class="logo-area">
        <div class="logo-icon">
          <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="lg1" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" class="login-grad-1"/>
                <stop offset="100%" class="login-grad-2"/>
              </linearGradient>
            </defs>
            <ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#lg1)"/>
            <circle cx="36" cy="56" r="16" fill="url(#lg1)"/>
            <circle cx="58" cy="50" r="20" fill="url(#lg1)"/>
            <polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/>
          </svg>
        </div>
        <h1 class="logo-name">CloudOne</h1>
        <p class="logo-tag">{{ t.tagline }}</p>
      </div>

      <transition name="slide-up" mode="out-in">
        <div v-if="mode === 'setup'" key="setup">
          <div class="form-header">
            <h2 class="form-title">{{ t.setup }}</h2>
            <p class="form-desc">{{ t.setupDesc }}</p>
          </div>
          <form @submit.prevent="doSetup">
            <div class="field">
              <label>{{ t.username }}</label>
              <input v-model="form.username" type="text" required autocomplete="username" />
            </div>
            <div class="field">
              <label>{{ t.password }}</label>
              <input v-model="form.password" type="password" required autocomplete="new-password" />
            </div>
            <p v-if="error" class="error">{{ error }}</p>
            <button type="submit" class="btn-primary" :disabled="loading">
              <span v-if="!loading">{{ t.createAccount }}</span>
              <span v-else class="spinner"></span>
            </button>
          </form>
        </div>

        <div v-else key="login">
          <div class="form-header">
            <h2 class="form-title">{{ t.login }}</h2>
            <p class="form-desc">{{ t.loginDesc }}</p>
          </div>
          <form @submit.prevent="doLogin">
            <div class="field">
              <label>{{ t.username }}</label>
              <input v-model="form.username" type="text" required autocomplete="username" />
            </div>
            <div class="field">
              <label>{{ t.password }}</label>
              <input v-model="form.password" type="password" required autocomplete="current-password" />
            </div>
            <p v-if="error" class="error">{{ error }}</p>
            <button type="submit" class="btn-primary" :disabled="loading">
              <span v-if="!loading">{{ t.login }}</span>
              <span v-else class="spinner"></span>
            </button>
          </form>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { t, currentLang as lang, setLang } from '../i18n'

const router = useRouter()
const auth = useAuthStore()
const mode = ref('login')
const form = ref({ username: '', password: '' })
const error = ref('')
const loading = ref(false)

onMounted(async () => {
  const res = await fetch('/api/auth/status')
  const data = await res.json()
  if (!data.setup) mode.value = 'setup'
})

async function doLogin() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(form.value.username, form.value.password)
    router.push('/files')
  } catch (e) {
    error.value = e.response?.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}

async function doSetup() {
  loading.value = true
  error.value = ''
  try {
    await auth.setup(form.value.username, form.value.password)
    router.push('/files')
  } catch (e) {
    error.value = e.response?.data?.error || 'Setup failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #EFF6FF 0%, #F0F9FF 50%, #DBEAFE 100%);
  position: relative;
  overflow: hidden;
}

.mesh {
  position: absolute;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(37,99,235,0.12) 0%, transparent 70%);
  top: -200px;
  right: -200px;
  animation: pulse 8s ease-in-out infinite;
}

.mesh-2 {
  position: absolute;
  width: 400px;
  height: 400px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(56,189,248,0.1) 0%, transparent 70%);
  bottom: -100px;
  left: -100px;
  animation: pulse 10s ease-in-out infinite reverse;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.1); opacity: 0.7; }
}

.auth-card {
  background: rgba(255,255,255,0.85);
  backdrop-filter: blur(24px);
  border: 1px solid rgba(37,99,235,0.12);
  border-radius: 24px;
  padding: 40px 48px 48px;
  width: 420px;
  box-shadow: 0 32px 80px rgba(37,99,235,0.1), 0 8px 24px rgba(0,0,0,0.06);
  position: relative;
  z-index: 1;
}

@media (max-width: 480px) {
  .auth-page { align-items: flex-start; padding: 0; }
  .auth-card {
    width: 100%;
    min-height: 100vh;
    border-radius: 0;
    padding: 40px 24px 32px;
    border: none;
    box-shadow: none;
    background: white;
  }
  .mesh, .mesh-2 { display: none; }
  .logo-area { margin-bottom: 28px; }
  .logo-icon { width: 44px; height: 44px; }
  .logo-name { font-size: 20px; }
  .field input { padding: 14px 16px; font-size: 16px; }
  .btn-primary { padding: 16px; font-size: 16px; }
}

/* Lang switch — absolute top-right */
.lang-switch {
  position: absolute;
  top: 18px;
  right: 18px;
  display: flex;
  gap: 4px;
}

.lang-switch button {
  padding: 3px 10px;
  border-radius: 20px;
  border: 1px solid var(--gray-200);
  background: transparent;
  font-size: 11px;
  font-family: inherit;
  cursor: pointer;
  color: var(--gray-500);
  transition: var(--transition);
}

.lang-switch button.active {
  background: var(--blue-600);
  color: white;
  border-color: var(--blue-600);
}

/* Logo — centered column */
.logo-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  margin-bottom: 32px;
  text-align: center;
}

.logo-icon {
  width: 52px;
  height: 52px;
  filter: drop-shadow(0 4px 12px rgba(37,99,235,0.3));
  margin-bottom: 4px;
}

.logo-name {
  font-size: 24px;
  font-weight: 700;
  background: var(--primary-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  line-height: 1;
}

.logo-tag {
  font-size: 13px;
  color: var(--gray-400);
  margin-top: 0;
}

/* Form header — centered */
.form-header {
  text-align: center;
  margin-bottom: 28px;
}

.form-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--gray-800);
  margin-bottom: 6px;
}

.form-desc {
  font-size: 14px;
  color: var(--gray-400);
  margin-bottom: 0;
}

.field {
  margin-bottom: 20px;
}

.field label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--gray-600);
  margin-bottom: 8px;
}

.field input {
  width: 100%;
  padding: 12px 16px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-family: inherit;
  background: var(--white);
  color: var(--gray-800);
  transition: var(--transition);
  outline: none;
}

.field input:focus {
  border-color: var(--blue-500);
  box-shadow: 0 0 0 3px rgba(59,130,246,0.12);
}

.btn-primary {
  width: 100%;
  padding: 14px;
  background: var(--primary-gradient);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: var(--transition);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 8px;
  box-shadow: 0 4px 16px rgba(37,99,235,0.3);
}

.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 8px 24px rgba(37,99,235,0.35); }
.btn-primary:active { transform: translateY(0); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; transform: none; }

.error {
  color: #EF4444;
  font-size: 13px;
  margin-bottom: 12px;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  display: inline-block;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* Logo SVG 渐变跟主题色走 */
.login-grad-1 { stop-color: var(--blue-600); }
.login-grad-2 { stop-color: var(--sky-500); }
</style>
