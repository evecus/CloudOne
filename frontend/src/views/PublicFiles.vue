<template>
  <Layout>
    <div class="page-container">
      <div class="page-header">
        <div>
          <h2 class="page-title">{{ t.public }}</h2>
          <p class="page-sub">{{ origin }}/raw/path/to/file</p>
        </div>
        <div class="lang-mini">
          <button :class="{ active: lang === 'zh' }" @click="setLang('zh')">中文</button>
          <button :class="{ active: lang === 'en' }" @click="setLang('en')">EN</button>
        </div>
      </div>

      <div class="content">
        <div v-if="!files.length" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10"/>
            <line x1="2" y1="12" x2="22" y2="12"/>
            <path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/>
          </svg>
          <p>{{ lang === 'zh' ? '暂无公开文件' : 'No public files' }}</p>
        </div>

        <div v-else class="file-list">
          <div v-for="f in files" :key="f.path" class="file-card">
            <div class="file-icon-wrap" :class="f.is_dir ? 'folder' : 'file'">
              <svg v-if="f.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <div class="file-details">
              <p class="fname">{{ f.name }}</p>
              <p class="fpath">{{ origin }}/raw{{ f.path }}</p>
            </div>
            <a :href="`/raw${f.path}`" target="_blank" class="raw-btn">RAW</a>
          </div>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Layout from '../components/Layout.vue'
import { t, currentLang as lang, setLang } from '../i18n'

const files = ref([])
const origin = location.origin

async function load() {
  const res = await fetch('/public?path=/')
  const data = await res.json()
  files.value = data.files || []
}

onMounted(load)
</script>

<style scoped>
.page-container { flex: 1; display: flex; flex-direction: column; }
.page-header { padding: 20px 28px 16px; border-bottom: 1px solid var(--gray-100); background: white; display: flex; align-items: center; justify-content: space-between; }
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); }
.page-sub { font-size: 12px; color: var(--gray-400); font-family: 'JetBrains Mono', monospace; margin-top: 4px; }
.content { flex: 1; padding: 24px 28px; overflow: auto; }

.lang-mini { display: flex; border: 1px solid var(--gray-200); border-radius: 8px; overflow: hidden; }
.lang-mini button { padding: 6px 12px; border: none; background: transparent; font-size: 12px; font-family: inherit; cursor: pointer; color: var(--gray-500); transition: var(--transition); }
.lang-mini button.active { background: var(--blue-600); color: white; }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 300px; color: var(--gray-300); gap: 12px; }
.empty-state svg { width: 64px; height: 64px; }
.empty-state p { font-size: 16px; font-weight: 500; color: var(--gray-400); }

.file-list { display: flex; flex-direction: column; gap: 10px; }

.file-card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: white;
  border: 1px solid var(--gray-100);
  border-radius: var(--radius);
  padding: 14px 18px;
  box-shadow: var(--shadow-sm);
  transition: var(--transition);
}
.file-card:hover { box-shadow: var(--shadow); }

.file-icon-wrap { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.file-icon-wrap.folder { background: rgba(251,191,36,0.15); color: #F59E0B; }
.file-icon-wrap.file { background: var(--blue-50); color: var(--blue-500); }
.file-icon-wrap svg { width: 18px; height: 18px; }

.file-details { flex: 1; min-width: 0; }
.fname { font-size: 14px; font-weight: 600; color: var(--gray-700); }
.fpath { font-size: 11px; color: var(--gray-400); font-family: 'JetBrains Mono', monospace; margin-top: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.raw-btn {
  padding: 6px 14px;
  background: var(--gray-800);
  color: white;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
  text-decoration: none;
  font-family: 'JetBrains Mono', monospace;
  letter-spacing: 1px;
  transition: var(--transition);
}
.raw-btn:hover { background: var(--blue-600); }
</style>
