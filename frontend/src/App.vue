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

onMounted(async () => {
  // 应用主题
  const theme = localStorage.getItem('theme') || 'blue'
  document.documentElement.setAttribute('data-theme', theme)
  // 应用字体
  const uiFont = localStorage.getItem('uiFont') || 'sora'
  const editorFont = localStorage.getItem('editorFont') || 'jetbrains'
  document.documentElement.style.setProperty('--ui-font', uiFontMap[uiFont] || uiFontMap.sora)
  document.documentElement.style.setProperty('--editor-font', editorFontMap[editorFont] || editorFontMap.jetbrains)

  if (auth.token) {
    try { await auth.fetchUser() } catch {}
  }
})
</script>
