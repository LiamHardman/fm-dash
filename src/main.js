import { createPinia } from 'pinia'
import { Notify, Quasar } from 'quasar'
import { createApp } from 'vue'
import router from './router'
import './css/app.scss' // Or '@/css/app.scss' if you prefer using the alias
// Import Quasar css
import 'quasar/dist/quasar.css'

// Import icon libraries
import '@quasar/extras/material-icons/material-icons.css'

import App from './App.vue'

async function initializeApp() {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(Quasar, {
    plugins: {
      Notify
    }, // import Quasar plugins as needed
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
  app.mount('#app')
}

// Initialize the app
initializeApp().catch(console.error)
