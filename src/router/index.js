// src/router/index.js
import { createRouter, createWebHistory } from "vue-router";
import PlayerUploadPage from "../pages/PlayerUploadPage.vue";
// Import the new TeamViewPage
import TeamViewPage from "../pages/TeamViewPage.vue";

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
    // You might want to pass props or handle data fetching here if needed
    // For now, we assume TeamViewPage can access `allPlayers` if passed via router state
    // or if it fetches/receives it through another mechanism (e.g., a global store)
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
