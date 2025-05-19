import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { quasar } from '@quasar/vite-plugin'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    quasar()
  ],
  server: {
    port: 3000,
    proxy: {
      '/upload': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})