import { createRouter, createWebHistory } from 'vue-router'
import PlayerUploadPage from '../pages/PlayerUploadPage.vue'

const routes = [
  {
    path: '/',
    name: 'home',
    component: PlayerUploadPage
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
