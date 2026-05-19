import { createRouter, createWebHistory } from 'vue-router'

const LandingPage = () => import('../pages/LandingPage.vue')
const LandingPageTest = () => import('../pages/LandingPageTest.vue')
const LandingPageHeroTest = () => import('../pages/LandingPageHeroTest.vue')
const PlayerUploadPage = () => import('../pages/PlayerUploadPage.vue')
const TeamViewPage = () => import('../pages/TeamViewPage.vue')
const DatasetPage = () => import('../pages/DatasetPage.vue')
const NationsPage = () => import('../pages/NationsPage.vue')
const TeamsPage = () => import('../pages/TeamsPage.vue')
const LeaguesPage = () => import('../pages/LeaguesPage.vue')
const DocsPage = () => import('../pages/DocsPage.vue')
const WishlistPage = () => import('../pages/WishlistPage.vue')
const PerformancePage = () => import('../pages/PerformancePage.vue')
const CardDesignsPage = () => import('../pages/CardDesignsPage.vue')

const routes = [
  {
    path: '/',
    name: 'home',
    component: LandingPage,
  },
  {
    path: '/home-test',
    name: 'home-test',
    component: LandingPageTest,
  },
  {
    path: '/hero-test',
    name: 'hero-test',
    component: LandingPageHeroTest,
  },
  {
    path: '/upload',
    name: 'upload',
    component: PlayerUploadPage,
  },
  {
    path: '/dataset/:datasetId',
    name: 'dataset',
    component: DatasetPage,
    props: true,
  },
  {
    path: '/team-view',
    name: 'team-view',
    component: TeamViewPage,
  },
  {
    path: '/team-view/:datasetId',
    name: 'shared-dataset',
    component: TeamViewPage,
    props: true,
  },
  {
    path: '/nations',
    name: 'nations',
    component: NationsPage,
  },
  {
    path: '/nations/:datasetId',
    name: 'shared-nations',
    component: NationsPage,
    props: true,
  },
  {
    path: '/teams',
    name: 'teams',
    component: TeamsPage,
  },
  {
    path: '/teams/:datasetId',
    name: 'shared-teams',
    component: TeamsPage,
    props: true,
  },
  {
    path: '/leagues',
    name: 'leagues',
    component: LeaguesPage,
  },
  {
    path: '/leagues/:datasetId',
    name: 'shared-leagues',
    component: LeaguesPage,
    props: true,
  },
  {
    path: '/performance',
    name: 'performance',
    component: PerformancePage,
  },
  {
    path: '/performance/:datasetId',
    name: 'shared-performance',
    component: PerformancePage,
    props: true,
  },
  {
    path: '/wishlist',
    name: 'wishlist',
    component: WishlistPage,
  },
  {
    path: '/docs',
    name: 'docs',
    component: DocsPage,
  },
  {
    path: '/cards',
    name: 'cards',
    component: CardDesignsPage,
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
