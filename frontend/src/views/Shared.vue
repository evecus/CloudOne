<template>
  <Layout>
    <div class="page-container">

      <!-- 桌面端头部 -->
      <div class="page-header desktop-header">
        <h2 class="page-title">{{ t.shared }}</h2>
        <div style="display:flex;gap:6px">
          <button class="btn-settings" @click="showSettings=true" :title="t.settings">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
          </button>
          <button class="btn-settings" @click="doLogout" :title="t.logout">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
          </button>
        </div>
      </div>

      <!-- 移动端头部 -->
      <div class="page-header mobile-header">
        <button class="mob-icon-btn" @click="showMobileNav=true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
        <h2 class="page-title">{{ t.shared }}</h2>
      </div>

      <!-- 移动端导航抽屉 -->
      <teleport to="body">
        <div v-if="showMobileNav" class="mob-drawer-mask" @click="showMobileNav=false">
          <div class="mob-drawer-panel" @click.stop>
            <div class="mob-drawer-top">
              <svg viewBox="0 0 100 100" style="width:26px;height:26px"><defs><linearGradient id="sdlg" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" style="stop-color:#2563EB"/><stop offset="100%" style="stop-color:#38BDF8"/></linearGradient></defs><ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#sdlg)"/><circle cx="36" cy="56" r="16" fill="url(#sdlg)"/><circle cx="58" cy="50" r="20" fill="url(#sdlg)"/><polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/></svg>
              <span class="mob-brand-name">CloudOne</span>
            </div>
            <nav class="mob-drawer-nav">
              <router-link to="/files" class="mob-nav-item" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
                {{ t.myFiles }}
              </router-link>
              <router-link to="/shared" class="mob-nav-item mob-nav-active" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
                {{ t.shared }}
              </router-link>
              <router-link to="/public-files" class="mob-nav-item" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
                {{ t.public }}
              </router-link>
              <router-link to="/webdav" class="mob-nav-item" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12H3l9-9 9 9h-2"/><path d="M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"/><path d="M9 21v-6a2 2 0 012-2h2a2 2 0 012 2v6"/></svg>
                {{ t.webdav }}
              </router-link>
              <div class="mob-nav-divider"></div>
              <button class="mob-nav-item" @click="showSettings=true;showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
                {{ t.settings }}
              </button>
              <button class="mob-nav-item mob-nav-logout" @click="doLogout">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
                {{ t.logout }}
              </button>
            </nav>
          </div>
        </div>
      </teleport>

      <div class="content">
        <div v-if="!shares.length" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/>
            <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/>
          </svg>
          <p>{{ t.noShares }}</p>
        </div>
        <div v-else class="share-list">
          <div v-for="s in shares" :key="s.id" class="share-card">
            <div class="share-icon">
              <svg v-if="s.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <div class="share-info">
              <p class="share-path">{{ s.file_path }}</p>
              <p class="share-date">{{ formatDate(s.created_at) }}</p>
              <div class="share-links">
                <a :href="`/s/${s.code}`" target="_blank" class="share-url">{{ origin }}/s/{{ s.code }}</a>
              </div>
            </div>
            <div class="share-actions">
              <button class="act-btn" @click="copyLink(s)" :title="t.copyLink">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
              </button>
              <button class="act-btn danger" @click="confirmCancel(s)" :title="lang==='zh'?'取消分享':'Unshare'">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/>
                  <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/>
                  <line x1="2" y1="2" x2="22" y2="22"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <teleport to="body">
      <div v-if="cancelTarget" class="modal-bg" @click.self="cancelTarget=null">
        <div class="modal">
          <h3>{{ lang==='zh'?'取消分享':'Unshare' }}</h3>
          <p class="modal-desc">{{ lang==='zh'?`确定取消分享 "${cancelTarget.file_path}" 吗？`:`Unshare "${cancelTarget.file_path}"?` }}</p>
          <div class="modal-actions">
            <button class="btn-ghost" @click="cancelTarget=null">{{ lang==='zh'?'取消':'Cancel' }}</button>
            <button class="btn-danger" @click="doCancel">{{ lang==='zh'?'确认取消分享':'Confirm' }}</button>
          </div>
        </div>
      </div>
      <div v-if="toastMsg" class="toast-tip">{{ toastMsg }}</div>
    </teleport>

    <SettingsModal v-model="showSettings" />
  </Layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Layout from '../components/Layout.vue'
import SettingsModal from '../components/SettingsModal.vue'
import api from '../api'
import { t, currentLang as lang } from '../i18n'

const showSettings = ref(false)
const showMobileNav = ref(false)
const { logout } = useAuthStore()
const router = useRouter()
function doLogout() { logout(); router.push('/login') }

const shares = ref([])
const cancelTarget = ref(null)
const origin = location.origin
const toastMsg = ref('')
let toastTimer = null

function showToast(msg) {
  toastMsg.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 2500)
}
async function load() {
  const { data } = await api.get('/share')
  shares.value = data || []
}
function confirmCancel(s) { cancelTarget.value = s }
async function doCancel() {
  await api.delete(`/share/${cancelTarget.value.id}`)
  cancelTarget.value = null; load()
}
async function copyLink(s) {
  copyText(`${origin}/s/${s.code}`)
  showToast(lang.value === 'zh' ? '链接已复制到剪切板' : 'Link copied to clipboard')
}
function copyText(text) {
  if (navigator.clipboard && window.isSecureContext) { navigator.clipboard.writeText(text).catch(() => fallbackCopy(text)) }
  else { fallbackCopy(text) }
}
function fallbackCopy(text) {
  const el = document.createElement('textarea')
  el.value = text; el.style.cssText = 'position:fixed;top:-9999px;left:-9999px'
  document.body.appendChild(el); el.select(); document.execCommand('copy'); document.body.removeChild(el)
}
function formatDate(d) {
  return new Date(d).toLocaleDateString(lang.value === 'zh' ? 'zh-CN' : 'en-US', { year:'numeric', month:'short', day:'numeric' })
}
onMounted(load)
</script>

<style scoped>
.page-container { flex: 1; display: flex; flex-direction: column; }
.page-header { padding: 20px 28px 16px; border-bottom: 1px solid var(--gray-100); background: white; display: flex; align-items: center; justify-content: space-between; }
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); margin: 0; }
.content { flex: 1; padding: 24px 28px; overflow: auto; }

.desktop-header { display: flex; }
.mobile-header  { display: none; }

.btn-settings { width: 36px; height: 36px; border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm); background: white; display: flex; align-items: center; justify-content: center; color: var(--gray-500); cursor: pointer; transition: var(--transition); }
.btn-settings svg { width: 17px; height: 17px; }
.btn-settings:hover { border-color: var(--blue-400); color: var(--blue-600); background: var(--blue-50); }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 300px; color: var(--gray-300); gap: 12px; }
.empty-state svg { width: 64px; height: 64px; }
.empty-state p { font-size: 16px; font-weight: 500; color: var(--gray-400); }

.share-list { display: flex; flex-direction: column; gap: 12px; }
.share-card { display: flex; align-items: center; gap: 16px; background: white; border: 1px solid var(--gray-100); border-radius: var(--radius); padding: 16px 20px; transition: var(--transition); box-shadow: var(--shadow-sm); }
.share-card:hover { box-shadow: var(--shadow); border-color: var(--gray-200); }
.share-icon { width: 40px; height: 40px; border-radius: 10px; background: var(--blue-50); color: var(--blue-500); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.share-icon svg { width: 20px; height: 20px; }
.share-info { flex: 1; min-width: 0; }
.share-path { font-size: 14px; font-weight: 600; color: var(--gray-700); margin-bottom: 4px; }
.share-date { font-size: 12px; color: var(--gray-400); margin-bottom: 6px; }
.share-url { font-size: 12px; font-family: 'JetBrains Mono', monospace; color: var(--blue-500); text-decoration: none; background: var(--blue-50); padding: 3px 8px; border-radius: 6px; }
.share-actions { display: flex; gap: 4px; }
.act-btn { width: 34px; height: 34px; border: none; background: var(--gray-50); border-radius: 8px; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--gray-400); transition: var(--transition); }
.act-btn svg { width: 15px; height: 15px; }
.act-btn:hover { background: var(--blue-50); color: var(--blue-600); }
.act-btn.danger:hover { background: rgba(239,68,68,0.1); color: #EF4444; }

.modal-bg { position: fixed; inset: 0; background: rgba(15,23,42,.45); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: white; border-radius: 20px; padding: 32px; width: 400px; max-width: 90vw; box-shadow: 0 20px 60px rgba(0,0,0,.15); animation: modalIn .2s cubic-bezier(.4,0,.2,1); }
@keyframes modalIn { from{opacity:0;transform:scale(.95) translateY(8px)} to{opacity:1;transform:scale(1) translateY(0)} }
.modal h3 { font-size: 18px; font-weight: 600; color: var(--gray-800); margin-bottom: 12px; }
.modal-desc { font-size: 14px; color: var(--gray-500); margin-bottom: 24px; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }
.btn-ghost { padding: 9px 18px; border: 1.5px solid var(--gray-200); border-radius: var(--radius-sm); background: transparent; color: var(--gray-600); font-size: 14px; font-weight: 500; font-family: inherit; cursor: pointer; }
.btn-ghost:hover { background: var(--gray-50); }
.btn-danger { padding: 9px 18px; background: #EF4444; color: white; border: none; border-radius: var(--radius-sm); font-size: 14px; font-weight: 600; font-family: inherit; cursor: pointer; }

/* 移动端汉堡 */
.mob-icon-btn { width: 38px; height: 38px; border: none; background: none; color: var(--gray-600); display: flex; align-items: center; justify-content: center; border-radius: 8px; cursor: pointer; flex-shrink: 0; }
.mob-icon-btn svg { width: 22px; height: 22px; }
.mob-icon-btn:active { background: var(--gray-100); }

/* 抽屉 */
.mob-drawer-mask { position: fixed; inset: 0; background: rgba(15,23,42,.4); backdrop-filter: blur(2px); z-index: 200; display: flex; }
.mob-drawer-panel { width: 260px; max-width: 80vw; height: 100%; background: white; display: flex; flex-direction: column; box-shadow: 4px 0 24px rgba(15,23,42,.15); animation: slideInLeft .2s ease; }
@keyframes slideInLeft { from{transform:translateX(-100%)} to{transform:translateX(0)} }
.mob-drawer-top { padding: 24px 20px 16px; border-bottom: 1px solid var(--gray-100); display: flex; align-items: center; gap: 10px; }
.mob-brand-name { font-size: 18px; font-weight: 700; background: linear-gradient(135deg,#2563EB,#38BDF8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.mob-drawer-nav { flex: 1; padding: 12px 10px; display: flex; flex-direction: column; gap: 2px; overflow-y: auto; }
.mob-nav-item { display: flex; align-items: center; gap: 12px; padding: 13px 14px; border-radius: 10px; text-decoration: none; font-size: 15px; font-weight: 500; color: var(--gray-600); border: none; background: none; font-family: inherit; cursor: pointer; transition: background .15s, color .15s; text-align: left; width: 100%; }
.mob-nav-item svg { width: 20px; height: 20px; flex-shrink: 0; }
.mob-nav-item:hover, .mob-nav-item:active { background: var(--blue-50); color: var(--blue-600); }
.mob-nav-active { background: var(--blue-50); color: var(--blue-600); font-weight: 600; }
.mob-nav-logout { color: #EF4444; }
.mob-nav-logout:hover { background: rgba(239,68,68,.08) !important; color: #EF4444 !important; }
.mob-nav-divider { height: 1px; background: var(--gray-100); margin: 8px 4px; }

@media (max-width: 768px) {
  .desktop-header { display: none; }
  .mobile-header  { display: flex; padding: 10px 14px; min-height: 52px; gap: 10px; }
  .content { padding: 16px; }
  .share-card { padding: 14px 16px; gap: 12px; }
  .share-path { font-size: 13px; }
  .share-url { font-size: 11px; word-break: break-all; white-space: normal; }
  .modal { width: 100% !important; max-width: 100% !important; border-radius: 20px 20px 0 0 !important; padding: 24px 20px 36px !important; }
  .modal-bg { align-items: flex-end; }
}
</style>
