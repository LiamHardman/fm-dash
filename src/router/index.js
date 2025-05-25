// src/router/index.js
import { createRouter, createWebHistory } from "vue-router";

// Lazy load components
const LandingPage = () => import("../pages/LandingPage.vue");
const PlayerUploadPage = () => import("../pages/PlayerUploadPage.vue");
const TeamViewPage = () => import("../pages/TeamViewPage.vue");
const DocsPage = () => import("../pages/DocsPage.vue");

const routes = [
  {
    path: "/",
    name: "home",
    component: LandingPage,
  },
  {
    path: "/upload",
    name: "upload",
    component: PlayerUploadPage,
  },
  {
    path: "/team-view",
    name: "team-view",
    component: TeamViewPage,
  },
  {
    path: "/docs",
    name: "docs",
    component: DocsPage,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
