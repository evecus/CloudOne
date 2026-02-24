<template>
  <Layout>
    <div class="page-container">
      <div class="page-header">
        <h2 class="page-title">{{ t.shared }}</h2>
        <div class="lang-mini">
          <button :class="{ active: lang === 'zh' }" @click="setLang('zh')">中文</button>
          <button :class="{ active: lang === 'en' }" @click="setLang('en')">EN</button>
        </div>
      </div>

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
                <a :href="`/s/${s.code}`" target="_blank" class="share-url">
                  {{ origin }}/s/{{ s.code }}
                </a>
              </div>
            </div>
            <div class="share-actions">
              <button class="act-btn" @click="copyLink(s)" :title="t.copyLink">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                <span v-if="copied === s.id">✓</span>
              </button>
              <button class="act-btn danger" @click="deleteShare(s.id)" :title="t.delete">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/></svg>
              </button>
            </div>
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

const shares = ref([])
const copied = ref(null)
const origin = location.origin

async function load() {
  const { data } = await api.get('/share')
  shares.value = data || []
}

async function deleteShare(id) {
  await api.delete(`/share/${id}`)
  load()
}

async function copyLink(s) {
  await navigator.clipboard.writeText(`${origin}/s/${s.code}`)
  copied.value = s.id
  setTimeout(() => copied.value = null, 2000)
}

function formatDate(d) {
  return new Date(d).toLocaleDateString(lang.value === 'zh' ? 'zh-CN' : 'en-US', {
    year: 'numeric', month: 'short', day: 'numeric'
  })
}

onMounted(load)
</script>

<style scoped>
.page-container { flex: 1; display: flex; flex-direction: column; }
.page-header {
  padding: 20px 28px 16px;
  border-bottom: 1px solid var(--gray-100);
  background: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.page-title { font-size: 20px; font-weight: 700; color: var(--gray-800); }
.content { flex: 1; padding: 24px 28px; overflow: auto; }

.lang-mini { display: flex; border: 1px solid var(--gray-200); border-radius: 8px; overflow: hidden; }
.lang-mini button { padding: 6px 12px; border: none; background: transparent; font-size: 12px; font-family: inherit; cursor: pointer; color: var(--gray-500); transition: var(--transition); }
.lang-mini button.active { background: var(--blue-600); color: white; }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 300px; color: var(--gray-300); gap: 12px; }
.empty-state svg { width: 64px; height: 64px; }
.empty-state p { font-size: 16px; font-weight: 500; color: var(--gray-400); }

.share-list { display: flex; flex-direction: column; gap: 12px; }

.share-card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: white;
  border: 1px solid var(--gray-100);
  border-radius: var(--radius);
  padding: 16px 20px;
  transition: var(--transition);
  box-shadow: var(--shadow-sm);
}

.share-card:hover { box-shadow: var(--shadow); border-color: var(--gray-200); }

.share-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--blue-50);
  color: var(--blue-500);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.share-icon svg { width: 20px; height: 20px; }

.share-info { flex: 1; min-width: 0; }
.share-path { font-size: 14px; font-weight: 600; color: var(--gray-700); margin-bottom: 4px; }
.share-date { font-size: 12px; color: var(--gray-400); margin-bottom: 6px; }

.share-url {
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
  color: var(--blue-500);
  text-decoration: none;
  background: var(--blue-50);
  padding: 3px 8px;
  border-radius: 6px;
}

.share-actions { display: flex; gap: 4px; }

.act-btn {
  width: 34px;
  height: 34px;
  border: none;
  background: var(--gray-50);
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: var(--gray-400);
  font-size: 12px;
  transition: var(--transition);
}

.act-btn svg { width: 15px; height: 15px; }
.act-btn:hover { background: var(--blue-50); color: var(--blue-600); }
.act-btn.danger:hover { background: rgba(239,68,68,0.1); color: #EF4444; }
</style>
