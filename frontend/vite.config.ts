import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    proxy: {
      // `go run ./cmd` listens on 8085. Under docker compose the API is on
      // 8005 instead: start with API_TARGET=http://localhost:8005.
      '/api': {
        target: process.env.API_TARGET ?? 'http://localhost:8085',
        changeOrigin: true
      }
    }
  }
})
