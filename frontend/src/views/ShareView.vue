<template>
  <div class="share-page">
    <div class="share-header">
      <div class="brand">
        <div class="brand-icon">
          <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="slg3" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#2563EB"/>
                <stop offset="100%" style="stop-color:#38BDF8"/>
              </linearGradient>
            </defs>
            <ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#slg3)"/>
            <circle cx="36" cy="56" r="16" fill="url(#slg3)"/>
            <circle cx="58" cy="50" r="20" fill="url(#slg3)"/>
            <polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/>
          </svg>
        </div>
        <span>CloudOne</span>
      </div>
    </div>

    <div class="share-content">
      <div v-if="loading" class="loading-state">
        <div class="spinner-lg"></div>
      </div>

      <div v-else-if="error" class="error-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
        <p>{{ error }}</p>
      </div>

      <template v-else-if="data">
        <div class="share-card">
          <div v-if="data.files" class="share-dir">
            <h2>{{ data.share?.file_path }}</h2>
            <div class="file-list">
              <div v-for="f in data.files" :key="f.path" class="file-item">
                <svg v-if="f.is_dir" viewBox="0 0 24 24" fill="currentColor" class="ficon folder"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="ficon"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                <span class="fname">{{ f.name }}</span>
                <span class="fsize">{{ f.is_dir ? '' : formatSize(f.size) }}</span>
              </div>
            </div>
          </div>

          <div v-else class="share-file">
            <div class="file-preview-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <h2>{{ data.path?.split('/').pop() }}</h2>
            <p class="file-path">{{ data.path }}</p>
            <a :href="`/s/${$route.params.code}/raw`" class="download-btn">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              Download
            </a>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const data = ref(null)
const loading = ref(true)
const error = ref('')

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B','KB','MB','GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

onMounted(async () => {
  try {
    const res = await fetch(`/s/${route.params.code}`)
    if (!res.ok) throw new Error('Not found')
    data.value = await res.json()
  } catch (e) {
    error.value = 'Share link not found or expired.'
  }
  loading.value = false
})
</script>

<style scoped>
.share-page { min-height: 100vh; background: var(--gray-50); }
.share-header {
  padding: 20px 32px;
  background: white;
  border-bottom: 1px solid var(--gray-100);
  box-shadow: var(--shadow-sm);
}
.brand { display: flex; align-items: center; gap: 12px; }
.brand-icon { width: 36px; height: 36px; }
.brand span { font-size: 18px; font-weight: 700; background: linear-gradient(135deg, #2563EB, #38BDF8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }

.share-content { max-width: 640px; margin: 48px auto; padding: 0 20px; }

.loading-state { display: flex; justify-content: center; padding: 80px; }
.spinner-lg { width: 32px; height: 32px; border: 3px solid var(--gray-200); border-top-color: var(--blue-500); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.error-state { display: flex; flex-direction: column; align-items: center; gap: 16px; padding: 80px 0; color: var(--gray-300); }
.error-state svg { width: 64px; height: 64px; }
.error-state p { font-size: 16px; color: var(--gray-400); }

.share-card {
  background: white;
  border-radius: 20px;
  padding: 40px;
  box-shadow: var(--shadow);
  border: 1px solid var(--gray-100);
}

.share-dir h2 { font-size: 20px; font-weight: 600; color: var(--gray-700); margin-bottom: 20px; }
.file-list { display: flex; flex-direction: column; gap: 2px; }
.file-item { display: flex; align-items: center; gap: 12px; padding: 10px 12px; border-radius: 8px; }
.file-item:hover { background: var(--gray-50); }
.ficon { width: 18px; height: 18px; flex-shrink: 0; color: var(--gray-400); }
.ficon.folder { color: #F59E0B; }
.fname { flex: 1; font-size: 14px; color: var(--gray-700); }
.fsize { font-size: 12px; color: var(--gray-400); }

.share-file { text-align: center; }
.file-preview-icon { width: 80px; height: 80px; background: var(--blue-50); border-radius: 20px; display: flex; align-items: center; justify-content: center; margin: 0 auto 20px; color: var(--blue-500); }
.file-preview-icon svg { width: 40px; height: 40px; }
.share-file h2 { font-size: 22px; font-weight: 600; color: var(--gray-800); margin-bottom: 8px; }
.file-path { font-size: 13px; color: var(--gray-400); font-family: 'JetBrains Mono', monospace; margin-bottom: 24px; }

.download-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 14px 28px;
  background: linear-gradient(135deg, #2563EB, #0EA5E9);
  color: white;
  border-radius: var(--radius);
  text-decoration: none;
  font-size: 15px;
  font-weight: 600;
  box-shadow: 0 4px 16px rgba(37,99,235,0.3);
  transition: var(--transition);
}

.download-btn svg { width: 18px; height: 18px; }
.download-btn:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(37,99,235,0.35); }
</style>
