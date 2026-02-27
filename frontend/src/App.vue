<template>
  <router-view v-slot="{ Component }">
    <transition name="fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'
import api from './api'

const auth = useAuthStore()

const uiFontMap = {
  sora:   "'Sora', -apple-system, sans-serif",
  inter:  "'Inter', -apple-system, sans-serif",
  roboto: "'Roboto', -apple-system, sans-serif",
  noto:   "'Noto Sans SC', -apple-system, sans-serif",
  lato:   "'Lato', -apple-system, sans-serif",
  system: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
}
const editorFontMap = {
  jetbrains: "'JetBrains Mono', 'Courier New', monospace",
  firacode:  "'Fira Code', 'Courier New', monospace",
  sourcecp:  "'Source Code Pro', 'Courier New', monospace",
  ibmplex:   "'IBM Plex Mono', 'Courier New', monospace",
  consolas:  "Consolas, 'Courier New', monospace",
  courier:   "'Courier New', monospace",
}

function applyTheme(id) {
  document.documentElement.setAttribute('data-theme', id === 'blue' ? '' : (id || ''))
}
function applyFonts(uiId, editorId) {
  document.documentElement.style.setProperty('--ui-font', uiFontMap[uiId] || uiFontMap.sora)
  document.documentElement.style.setProperty('--editor-font', editorFontMap[editorId] || editorFontMap.jetbrains)
}

onMounted(async () => {
  // 第一步：先用 localStorage 立即渲染，避免页面闪烁
  applyTheme(localStorage.getItem('theme') || 'blue')
  applyFonts(localStorage.getItem('uiFont') || 'sora', localStorage.getItem('editorFont') || 'jetbrains')

  if (auth.token) {
    try { await auth.fetchUser() } catch {}
    // 第二步：登录后从服务端拉最新设置，覆盖本地缓存（实现跨设备同步）
    try {
      const { data } = await api.get('/settings')
      const theme = data.ui_theme || localStorage.getItem('theme') || 'blue'
      const uiFont = data.ui_font || localStorage.getItem('uiFont') || 'sora'
      const editorFont = data.editor_font || localStorage.getItem('editorFont') || 'jetbrains'
      applyTheme(theme)
      applyFonts(uiFont, editorFont)
      localStorage.setItem('theme', theme)
      localStorage.setItem('uiFont', uiFont)
      localStorage.setItem('editorFont', editorFont)
    } catch {}
  }
})
</script>
