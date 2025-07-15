import path from 'node:path'
import { quasar, transformAssetUrls } from '@quasar/vite-plugin'
import vue from '@vitejs/plugin-vue'
import { visualizer } from 'rollup-plugin-visualizer'
import { defineConfig } from 'vite'

// Custom plugin for chunk size analysis and warnings
function chunkSizeAnalyzer() {
  return {
    name: 'chunk-size-analyzer',
    generateBundle(options, bundle) {
      const chunkSizeLimit = 500 * 1024 // 500KB in bytes
      const criticalChunkLimit = 300 * 1024 // 300KB for critical chunks

      Object.entries(bundle).forEach(([fileName, chunk]) => {
        if (chunk.type === 'chunk') {
          const size = Buffer.byteLength(chunk.code, 'utf8')
          const sizeKB = Math.round(size / 1024)

          // Check for oversized chunks
          if (size > chunkSizeLimit) {
            console.warn(`⚠️  Large chunk detected: ${fileName} (${sizeKB}KB)`)
            console.warn(`   Consider splitting this chunk further or lazy loading components`)
          }

          // Check for critical chunks that should be smaller
          if (chunk.isEntry && size > criticalChunkLimit) {
            console.warn(`⚠️  Large entry chunk: ${fileName} (${sizeKB}KB)`)
            console.warn(`   Entry chunks should be smaller for faster initial loading`)
          }

          // Log chunk information in development
          if (process.env.NODE_ENV === 'development') {
            console.log(`📦 Chunk: ${fileName} - ${sizeKB}KB`)
          }
        }
      })

      // Generate chunk size report
      const chunks = Object.entries(bundle)
        .filter(([, chunk]) => chunk.type === 'chunk')
        .map(([fileName, chunk]) => ({
          name: fileName,
          size: Buffer.byteLength(chunk.code, 'utf8'),
          isEntry: chunk.isEntry
        }))
        .sort((a, b) => b.size - a.size)

      const totalSize = chunks.reduce((sum, chunk) => sum + chunk.size, 0)

      console.log('\n📊 Bundle Analysis:')
      console.log(`Total bundle size: ${Math.round(totalSize / 1024)}KB`)
      console.log(`Number of chunks: ${chunks.length}`)
      console.log('\nLargest chunks:')
      chunks.slice(0, 5).forEach(chunk => {
        const sizeKB = Math.round(chunk.size / 1024)
        const type = chunk.isEntry ? '(entry)' : ''
        console.log(`  ${chunk.name}: ${sizeKB}KB ${type}`)
      })
    }
  }
}

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
    // Chunk size analyzer for build optimization
    chunkSizeAnalyzer(),
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
        // Advanced manual chunks for optimal caching and loading
        manualChunks: id => {
          // Route-based chunks for major pages
          if (
            id.includes('pages/PlayerUploadPage.vue') ||
            id.includes('components/InteractiveUploadLoader.vue')
          ) {
            return 'page-upload'
          }
          if (
            id.includes('pages/DatasetPage.vue') ||
            id.includes('components/PlayerDataTable.vue') ||
            id.includes('components/PlayerTableRow.vue')
          ) {
            return 'page-player-table'
          }
          if (id.includes('pages/TeamViewPage.vue') || id.includes('components/PitchDisplay.vue')) {
            return 'page-team-view'
          }
          if (
            id.includes('pages/PerformancePage.vue') ||
            id.includes('components/PerformanceMonitor.vue')
          ) {
            return 'page-performance'
          }

          // Vendor chunk splitting
          if (id.includes('node_modules')) {
            // Core Vue framework - highest priority, keep small
            if (
              id.includes('vue/dist/vue.esm') ||
              id.includes('vue-router/dist') ||
              id.includes('pinia/dist')
            ) {
              return 'vendor-vue-core'
            }

            // Quasar UI framework - split into smaller chunks to avoid large bundle
            if (
              id.includes('quasar/dist/quasar.esm.prod.js') ||
              id.includes('quasar/dist/quasar.common.js')
            ) {
              return 'vendor-quasar-core'
            }
            if (id.includes('quasar/src/components') || id.includes('quasar/dist/icon-set')) {
              return 'vendor-quasar-components'
            }
            if (id.includes('@quasar/extras')) {
              return 'vendor-quasar-extras'
            }

            // Charts and visualization - separate by library to reduce size
            if (id.includes('chart.js/dist') || id.includes('chart.js/auto')) {
              return 'vendor-chartjs-core'
            }
            if (id.includes('vue-chartjs')) {
              return 'vendor-vue-chartjs'
            }
            if (id.includes('chartjs-plugin')) {
              return 'vendor-chartjs-plugins'
            }

            // VueUse utilities - split by feature
            if (id.includes('@vueuse/core')) {
              return 'vendor-vueuse'
            }

            // CSS processing libraries
            if (id.includes('sass') || id.includes('postcss') || id.includes('autoprefixer')) {
              return 'vendor-css-processors'
            }

            // Development tools (should be excluded in production)
            if (id.includes('rollup-plugin-visualizer') || id.includes('@vue/devtools')) {
              return 'vendor-dev-tools'
            }

            // Other smaller vendor libraries
            if (id.includes('node_modules')) {
              // Group small utilities together
              if (
                id.includes('unique-slug') ||
                id.includes('fs-minipass') ||
                id.includes('biome')
              ) {
                return 'vendor-utils-small'
              }
              return 'vendor-misc'
            }
          }

          // Component-based chunks for heavy components
          if (
            id.includes('components/PlayerDetailDialog.vue') ||
            id.includes('components/player-details/')
          ) {
            return 'component-player-details'
          }
          if (
            id.includes('components/ExportOptionsDialog.vue') ||
            id.includes('utils/csvExport.js')
          ) {
            return 'component-export'
          }
          if (id.includes('components/ScatterPlotCard.vue') || id.includes('components/filters/')) {
            return 'component-charts-filters'
          }

          // Composables and utilities
          if (id.includes('composables/') || id.includes('utils/') || id.includes('services/')) {
            return 'shared-utilities'
          }

          // Store modules
          if (id.includes('stores/')) {
            return 'shared-stores'
          }
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
    // Optimize build performance and output with stricter chunk size limits
    chunkSizeWarningLimit: 500, // Reduced to encourage smaller, more focused chunks
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
      // Pre-bundle commonly used utilities with specific imports
      'chart.js/helpers',
      'chart.js/auto',
      'vue-chartjs',
      'chartjs-plugin-annotation'
    ],
    exclude: [
      '@vitejs/plugin-vue',
      // Exclude development-only dependencies
      'rollup-plugin-visualizer',
      '@vue/devtools-api'
    ],
    esbuildOptions: {
      target: 'es2020',
      format: 'esm',
      // Optimize for production builds
      treeShaking: true,
      // Enable more aggressive optimization
      minify: process.env.NODE_ENV === 'production',
      // Define globals for better optimization
      define: {
        'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development')
      }
    },
    // Force optimization of specific packages
    force: process.env.NODE_ENV === 'production'
  },
  // CSS optimization
  css: {
    devSourcemap: process.env.NODE_ENV === 'development'
  }
})
