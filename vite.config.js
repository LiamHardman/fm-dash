import path from 'node:path'
import { quasar, transformAssetUrls } from '@quasar/vite-plugin'
import vue from '@vitejs/plugin-vue'
import { visualizer } from 'rollup-plugin-visualizer'
import { defineConfig } from 'vite'
import viteCompression from 'vite-plugin-compression'

// Custom plugin for chunk size analysis and warnings
function chunkSizeAnalyzer() {
  return {
    name: 'chunk-size-analyzer',
    generateBundle(_options, bundle) {
      const chunkSizeLimit = 500 * 1024 // 500KB in bytes
      const criticalChunkLimit = 300 * 1024 // 300KB for critical chunks

      Object.entries(bundle).forEach(([fileName, chunk]) => {
        if (chunk.type === 'chunk') {
          const size = Buffer.byteLength(chunk.code, 'utf8')
          const sizeKB = Math.round(size / 1024)

          // Check for oversized chunks
          if (size > chunkSizeLimit) {
            console.warn(`⚠️  Large chunk detected: ${fileName} (${sizeKB}KB)`)
            console.warn('   Consider splitting this chunk further or lazy loading components')
          }

          // Check for critical chunks that should be smaller
          if (chunk.isEntry && size > criticalChunkLimit) {
            console.warn(`⚠️  Large entry chunk: ${fileName} (${sizeKB}KB)`)
            console.warn('   Entry chunks should be smaller for faster initial loading')
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
          isEntry: chunk.isEntry,
        }))
        .sort((a, b) => b.size - a.size)

      const totalSize = chunks.reduce((sum, chunk) => sum + chunk.size, 0)

      console.log('\n📊 Bundle Analysis:')
      console.log(`Total bundle size: ${Math.round(totalSize / 1024)}KB`)
      console.log(`Number of chunks: ${chunks.length}`)
      console.log('\nLargest chunks:')
      chunks.slice(0, 5).forEach((chunk) => {
        const sizeKB = Math.round(chunk.size / 1024)
        const type = chunk.isEntry ? '(entry)' : ''
        console.log(`  ${chunk.name}: ${sizeKB}KB ${type}`)
      })
    },
  }
}

export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', '@quasar/extras', 'quasar'],
    // Ensure proper module resolution
    esbuildOptions: {
      target: 'esnext',
      // Add tree-shaking optimization
      treeShaking: true,
    },
    // Pre-bundle heavy dependencies for faster dev server startup
    entries: ['src/main.js', 'src/stores/playerStore.js', 'src/composables/useApi.js'],
  },
  plugins: [
    vue({
      template: { transformAssetUrls },
    }),
    quasar({
      sassVariables: '@/quasar-variables.scss',
    }),
    // Chunk size analyzer for build optimization
    chunkSizeAnalyzer(),
    // Pre-compress build output to .br files so the server can skip on-the-fly compression.
    // Only takes effect if the serving layer prefers a pre-built .br over compressing itself —
    // see docs/PERFORMANCE_FIXES_2026-07-08.md #10 for the nginx follow-up check.
    viteCompression({
      algorithm: 'brotliCompress',
      ext: '.br',
      threshold: 1024,
      deleteOriginFile: false,
    }),
    // Bundle analyzer - only in development or when explicitly requested
    (process.env.ANALYZE === 'true' || process.env.NODE_ENV === 'development') &&
      visualizer({
        filename: 'dist/stats.html',
        open: process.env.ANALYZE === 'true',
        gzipSize: true,
        brotliSize: true,
        template: 'treemap', // Better visualization
      }),
  ].filter(Boolean),
  build: {
    // Optimize build output with advanced chunking
    rollupOptions: {
      // Ensure proper module resolution
      preserveEntrySignatures: 'strict',
      // Handle ES module initialization issues
      onwarn(warning, warn) {
        // Suppress circular dependency warnings for now
        if (warning.code === 'CIRCULAR_DEPENDENCY') {
          return
        }
        warn(warning)
      },
      output: {
        // Keep manual chunking limited to third-party libraries. Local source
        // modules have enough cross-route dependencies that forcing them into
        // named chunks can create ESM evaluation cycles in production builds.
        manualChunks: (id) => {
          if (id.includes('node_modules')) {
            // Core Vue framework - highest priority, keep small
            if (
              id.includes('vue/dist/vue.esm') ||
              id.includes('vue-router/dist') ||
              id.includes('pinia/dist')
            ) {
              return 'vendor-vue-core'
            }

            // Quasar UI framework - keep together to prevent initialization issues
            if (id.includes('quasar') || id.includes('@quasar/extras')) {
              return 'vendor-quasar'
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
              if (id.includes('biome')) {
                return 'vendor-utils-small'
              }
              return 'vendor-misc'
            }
          }
        },
        chunkFileNames: (chunkInfo) => {
          const facadeModuleId = chunkInfo.facadeModuleId
            ? chunkInfo.facadeModuleId.split('/').pop().replace('.vue', '')
            : 'chunk'
          return `js/${facadeModuleId}-[hash].js`
        },
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name.split('.')
          const ext = info[info.length - 1]
          if (/png|jpe?g|svg|gif|tiff|bmp|ico/i.test(ext)) {
            return 'images/[name]-[hash][extname]'
          }
          if (/css/i.test(ext)) {
            return 'css/[name]-[hash][extname]'
          }
          return 'assets/[name]-[hash][extname]'
        },
      },
    },
    // Optimize build performance and output with stricter chunk size limits
    chunkSizeWarningLimit: 500, // Reduced to encourage smaller, more focused chunks
    sourcemap: process.env.NODE_ENV === 'development',
    cssCodeSplit: true,
    assetsInlineLimit: 2048, // Reduced from 4096 to avoid large inline assets
    // Use esbuild for much faster minification (5-10x faster than terser)
    minify: 'esbuild',
    // Optional: esbuild minify options for production
    ...(process.env.NODE_ENV === 'production' && {
      esbuild: {
        drop: ['console', 'debugger'],
      },
    }),
    // CommonJS to ESM conversion optimization
    commonjsOptions: {
      include: [/node_modules/],
      transformMixedEsModules: true,
    },
    // Target modern browsers for smaller bundle
    target: 'es2020',
  },
  server: {
    port: 3000,
    proxy: {
      // Proxy for the file upload endpoint
      '/upload': {
        target: 'http://localhost:8091', // Your Go backend URL
        changeOrigin: true, // Recommended for most cases
        // secure: false, // If your backend is HTTP and Vite is HTTPS (dev only)
      },
      // Proxy for all other API calls (e.g., /api/players, /api/roles)
      '/api': {
        target: 'http://localhost:8091', // Your Go backend URL
        changeOrigin: true,
        // secure: false,
        // Optional: rewrite path if your Go API doesn't expect /api prefix
        // rewrite: (path) => path.replace(/^\/api/, '')
        // However, your Go routes are already /api/players and /api/roles, so no rewrite needed here.
      },
      '/debug': {
        target: 'http://localhost:8091',
        changeOrigin: true,
      },
    },
  },
  define: {
    // Make environment variables available to the frontend
    __GA_TRACKING_ID__: JSON.stringify(process.env.VITE_GA_TRACKING_ID || 'G-QYG3QS5C5Y'),
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0'),
    __BUILD_DATE__: JSON.stringify(new Date().toISOString()),
    // Define global to prevent require() errors
    global: 'globalThis',
  },
  // Alternative way to expose env vars (prefixed with VITE_)
  envPrefix: ['VITE_'],

  // CSS optimization
  css: {
    devSourcemap: process.env.NODE_ENV === 'development',
  },
})
