import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:8765', changeOrigin: true, ws: true },
      '/ws': { target: 'ws://localhost:8765', ws: true },
    },
  },
  build: {
    outDir: '../backend/frontend_dist',
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus', '@element-plus/icons-vue'],
          'vue-vendor': ['vue', 'vue-router', 'pinia'],
          'axios-vendor': ['axios'],
          'cytoscape': ['cytoscape'],
        },
      },
    },
  },
})
