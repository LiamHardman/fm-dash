import { createPinia } from 'pinia'
import { Notify, Quasar } from 'quasar'
import { createApp } from 'vue'
import router from './router'
import './css/app.scss'

// Optimized CSS imports - only load what we need
import 'quasar/dist/quasar.css'
import '@quasar/extras/material-icons/material-icons.css'

import App from './App.vue'
import { loadDevLibraries, preloadCriticalLibraries } from './utils/libraryOptimizer.js'

async function initializeApp() {
  // Preload critical libraries for better performance
  await preloadCriticalLibraries()

  const app = createApp(App)
  const pinia = createPinia()

  // Configure Quasar with optimized settings
  app.use(Quasar, {
    plugins: {
      Notify
    },
    config: {
      brand: {
        primary: '#1976D2',
        secondary: '#26A69A',
        accent: '#9C27B0',
        dark: '#1d1d1d',
        positive: '#21BA45',
        negative: '#C10015',
        info: '#31CCEC',
        warning: '#F2C037'
      },
      // Optimize Quasar for production
      ...(process.env.NODE_ENV === 'production' && {
        ripple: false, // Disable ripple effect in production for better performance
        loadingBar: { skipHijack: true } // Skip loading bar hijacking
      })
    }
  })

  app.use(pinia)
  app.use(router)

  // Load development-only libraries conditionally
  if (process.env.NODE_ENV === 'development') {
    loadDevLibraries()
      .then(devTools => {
        if (devTools.length > 0) {
          console.log('🛠️ Development tools loaded:', devTools.length)
        }
      })
      .catch(error => {
        console.warn('⚠️ Some development tools failed to load:', error.message)
      })
  }

  app.mount('#app')

  // Performance monitoring in production
  if (process.env.NODE_ENV === 'production' && 'performance' in window) {
    // Log initial load performance
    window.addEventListener('load', () => {
      const perfData = performance.getEntriesByType('navigation')[0]
      if (perfData) {
        console.log('📊 App Load Performance:', {
          domContentLoaded: Math.round(
            perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart
          ),
          loadComplete: Math.round(perfData.loadEventEnd - perfData.loadEventStart),
          totalTime: Math.round(perfData.loadEventEnd - perfData.fetchStart)
        })
      }
    })
  }
}

// Initialize the app with error handling
initializeApp().catch(error => {
  console.error('❌ Failed to initialize application:', error)

  // Graceful error handling
  if (typeof window !== 'undefined') {
    const errorMessage =
      process.env.NODE_ENV === 'development'
        ? `Application failed to start: ${error.message}`
        : 'Application failed to start. Please refresh the page.'

    if (window.alert) {
      window.alert(errorMessage)
    }

    // Try to show a basic error message in the DOM
    const errorDiv = document.createElement('div')
    errorDiv.innerHTML = `
      <div style="
        position: fixed; 
        top: 50%; 
        left: 50%; 
        transform: translate(-50%, -50%);
        background: #f44336; 
        color: white; 
        padding: 20px; 
        border-radius: 8px;
        font-family: Arial, sans-serif;
        text-align: center;
        z-index: 9999;
      ">
        <h3>Application Error</h3>
        <p>${errorMessage}</p>
        <button onclick="window.location.reload()" style="
          background: white; 
          color: #f44336; 
          border: none; 
          padding: 10px 20px; 
          border-radius: 4px; 
          cursor: pointer;
          margin-top: 10px;
        ">
          Reload Page
        </button>
      </div>
    `
    document.body.appendChild(errorDiv)
  }

  throw error
})
