// src/router/index.js
import { createRouter, createWebHistory } from "vue-router";

// Lazy load components
const PlayerUploadPage = () => import("../pages/PlayerUploadPage.vue");
const TeamViewPage = () => import("../pages/TeamViewPage.vue");

const routes = [
  {
    path: "/",
    name: "home",
    component: PlayerUploadPage,
  },
  // Add new route for the Team View page
  {
    path: "/team-view",
    name: "team-view",
    component: TeamViewPage,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
