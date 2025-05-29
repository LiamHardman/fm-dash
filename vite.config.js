import path from 'path'
import { quasar, transformAssetUrls } from '@quasar/vite-plugin'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  plugins: [
    vue({
      template: { transformAssetUrls }
    }),
    quasar({
      sassVariables: '@/quasar-variables.scss'
    })
  ],
  server: {
    port: 3000,
    proxy: {
      // Proxy for the file upload endpoint
      '/upload': {
        target: 'http://localhost:8091', // Your Go backend URL
        changeOrigin: true // Recommended for most cases
        // secure: false, // If your backend is HTTP and Vite is HTTPS (dev only)
      },
      // Proxy for all other API calls (e.g., /api/players, /api/roles)
      '/api': {
        target: 'http://localhost:8091', // Your Go backend URL
        changeOrigin: true
        // secure: false,
        // Optional: rewrite path if your Go API doesn't expect /api prefix
        // rewrite: (path) => path.replace(/^\/api/, '')
        // However, your Go routes are already /api/players and /api/roles, so no rewrite needed here.
      }
    }
  }
})
