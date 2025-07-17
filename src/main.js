import { createPinia } from 'pinia'
import { Notify, Quasar } from 'quasar'
import { createApp } from 'vue'
import router from './router'
import './css/app.scss'

// Optimized CSS imports - only load what we need
import 'quasar/dist/quasar.css'
import '@quasar/extras/material-icons/material-icons.css'

import App from './App.vue'

// Create the app synchronously to avoid initialization issues
const app = createApp(App)
const pinia = createPinia()

// Configure Quasar with minimal settings first
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
    }
  }
})

app.use(pinia)
app.use(router)

// Mount the app
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
