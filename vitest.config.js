import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { quasar } from '@quasar/vite-plugin'

export default defineConfig({
  plugins: [
    vue(),
    quasar({
      sassVariables: 'src/quasar-variables.scss'
    })
  ],
  test: {
    environment: 'happy-dom',
    globals: true,
    setupFiles: [],
    include: ['src/**/*.{test,spec}.{js,ts,vue}'],
    exclude: ['node_modules', 'dist', 'src/api/**']
  },
  resolve: {
    alias: {
      '@': '/src'
    }
  }
}) 