import path from 'node:path'
import { quasar, transformAssetUrls } from '@quasar/vite-plugin'
import vue from '@vitejs/plugin-vue'
import { visualizer } from 'rollup-plugin-visualizer'
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
    }),
    // Bundle analyzer - only in development or when explicitly requested
    (process.env.ANALYZE === 'true' || process.env.NODE_ENV === 'development') &&
      visualizer({
        filename: 'dist/stats.html',
        open: process.env.ANALYZE === 'true',
        gzipSize: true,
        brotliSize: true,
        template: 'treemap' // Better visualization
      })
  ].filter(Boolean),
  build: {
    // Optimize build output with advanced chunking
    rollupOptions: {
      output: {
        // Manual chunks for better caching and loading
        manualChunks: {
          // Core Vue framework
          'vue-core': ['vue', 'vue-router', 'pinia'],
          // UI framework
          'ui-framework': ['quasar'],
          // Charts and visualization
          charts: ['chart.js', 'vue-chartjs', 'chartjs-plugin-annotation'],
          // Utilities and composables
          utils: ['@vueuse/core'],
          // Large vendor libraries
          'vendor-large': []
        },
        chunkFileNames: chunkInfo => {
          const facadeModuleId = chunkInfo.facadeModuleId
            ? chunkInfo.facadeModuleId.split('/').pop().replace('.vue', '')
            : 'chunk'
          return `js/${facadeModuleId}-[hash].js`
        },
        assetFileNames: assetInfo => {
          const info = assetInfo.name.split('.')
          const ext = info[info.length - 1]
          if (/png|jpe?g|svg|gif|tiff|bmp|ico/i.test(ext)) {
            return `images/[name]-[hash][extname]`
          }
          if (/css/i.test(ext)) {
            return `css/[name]-[hash][extname]`
          }
          return `assets/[name]-[hash][extname]`
        }
      }
    },
    // Optimize build performance and output
    chunkSizeWarningLimit: 800, // Reduced from 1000 to encourage smaller chunks
    sourcemap: process.env.NODE_ENV === 'development',
    cssCodeSplit: true,
    assetsInlineLimit: 2048, // Reduced from 4096 to avoid large inline assets
    // Enable advanced minification
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: process.env.NODE_ENV === 'production',
        drop_debugger: true,
        pure_funcs: ['console.log', 'console.info'], // Remove specific console methods
        passes: 2 // Multiple compression passes
      },
      mangle: {
        safari10: true // Safari 10+ compatibility
      }
    },
    // CommonJS to ESM conversion optimization
    commonjsOptions: {
      include: [/node_modules/],
      transformMixedEsModules: true
    },
    // Target modern browsers for smaller bundle
    target: 'es2020'
  },
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
  },
  define: {
    // Make environment variables available to the frontend
    __GA_TRACKING_ID__: JSON.stringify(process.env.VITE_GA_TRACKING_ID || 'G-QYG3QS5C5Y'),
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0'),
    __BUILD_DATE__: JSON.stringify(new Date().toISOString()),
    // Define global to prevent require() errors
    global: 'globalThis'
  },
  // Alternative way to expose env vars (prefixed with VITE_)
  envPrefix: ['VITE_'],
  // Enhanced dependency optimization
  optimizeDeps: {
    include: [
      'vue',
      'vue-router',
      'pinia',
      'quasar',
      '@vueuse/core',
      // Pre-bundle commonly used utilities
      'chart.js/auto',
      'vue-chartjs'
    ],
    exclude: ['@vitejs/plugin-vue'],
    esbuildOptions: {
      target: 'es2020',
      format: 'esm',
      // Optimize for production builds
      treeShaking: true
    },
    // Force optimization of specific packages
    force: process.env.NODE_ENV === 'production'
  },
  // CSS optimization
  css: {
    devSourcemap: process.env.NODE_ENV === 'development'
  }
})
