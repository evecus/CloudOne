import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
  },
  server: {
    proxy: {
      '/api': 'http://localhost:6677',
      '/raw': 'http://localhost:6677',
      '/s': 'http://localhost:6677',
      '/public': 'http://localhost:6677',
    }
  }
})
