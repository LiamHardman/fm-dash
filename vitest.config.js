import { quasar } from '@quasar/vite-plugin'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [
    vue(),
    quasar({
      sassVariables: 'src/quasar-variables.scss'
    })
  ],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['src/test-setup.js'],
    include: ['src/**/*.{test,spec}.{js,ts,vue}'],
    exclude: ['node_modules', 'dist', 'src/api/**'],
    testTimeout: 5000,
    hookTimeout: 5000
  },
  resolve: {
    alias: {
      '@': new URL('./src', import.meta.url).pathname
    }
  }
})
