// src/router/index.js
import { createRouter, createWebHistory } from "vue-router";

// Lazy load components
const LandingPage = () => import("../pages/LandingPage.vue");
const PlayerUploadPage = () => import("../pages/PlayerUploadPage.vue");
const TeamViewPage = () => import("../pages/TeamViewPage.vue");
const DatasetPage = () => import("../pages/DatasetPage.vue");
const NationsPage = () => import("../pages/NationsPage.vue");
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
    path: "/dataset/:datasetId",
    name: "dataset",
    component: DatasetPage,
    props: true,
  },
  {
    path: "/team-view",
    name: "team-view",
    component: TeamViewPage,
  },
  {
    path: "/team-view/:datasetId",
    name: "shared-dataset",
    component: TeamViewPage,
    props: true,
  },
  {
    path: "/nations",
    name: "nations",
    component: NationsPage,
  },
  {
    path: "/nations/:datasetId",
    name: "shared-nations",
    component: NationsPage,
    props: true,
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
