import { createRouter, createWebHistory } from 'vue-router'

const LandingPage = () => import('../pages/LandingPage.vue')
const PlayerUploadPage = () => import('../pages/PlayerUploadPage.vue')
const TeamViewPage = () => import('../pages/TeamViewPage.vue')
const DatasetPage = () => import('../pages/DatasetPage.vue')
const NationsPage = () => import('../pages/NationsPage.vue')
const TeamsPage = () => import('../pages/TeamsPage.vue')
const LeaguesPage = () => import('../pages/LeaguesPage.vue')
const DocsPage = () => import('../pages/DocsPage.vue')
const ShortlistsPage = () => import('../pages/ShortlistsPage.vue')
const SavedSearchesPage = () => import('../pages/SavedSearchesPage.vue')
const PerformancePage = () => import('../pages/PerformancePage.vue')
const CardDesignsPage = () => import('../pages/CardDesignsPage.vue')
const ProgressionPage = () => import('../pages/ProgressionPage.vue')
const SaveImportPage = () => import('../pages/SaveImportPage.vue')
const SaveAnalysisPage = () => import('../pages/SaveAnalysisPage.vue')
const ScoutingBookPage = () => import('../pages/ScoutingBookPage.vue')

const routes = [
  {
    path: '/',
    name: 'home',
    component: LandingPage,
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
    redirect: () =>
      sessionStorage.getItem('currentDatasetId')
        ? `/nations/${sessionStorage.getItem('currentDatasetId')}`
        : '/upload',
  },
  {
    path: '/nations/:datasetId',
    name: 'shared-nations',
    component: NationsPage,
    props: true,
  },
  {
    path: '/teams',
    redirect: () =>
      sessionStorage.getItem('currentDatasetId')
        ? `/teams/${sessionStorage.getItem('currentDatasetId')}`
        : '/upload',
  },
  {
    path: '/teams/:datasetId',
    name: 'shared-teams',
    component: TeamsPage,
    props: true,
  },
  {
    path: '/leagues',
    redirect: () =>
      sessionStorage.getItem('currentDatasetId')
        ? `/leagues/${sessionStorage.getItem('currentDatasetId')}`
        : '/upload',
  },
  {
    path: '/leagues/:datasetId',
    name: 'shared-leagues',
    component: LeaguesPage,
    props: true,
  },
  {
    path: '/performance',
    redirect: () =>
      sessionStorage.getItem('currentDatasetId')
        ? `/performance/${sessionStorage.getItem('currentDatasetId')}`
        : '/upload',
  },
  {
    path: '/performance/:datasetId',
    name: 'shared-performance',
    component: PerformancePage,
    props: true,
  },
  {
    path: '/shortlists',
    name: 'shortlists',
    component: ShortlistsPage,
  },
  {
    path: '/saved-searches',
    name: 'saved-searches',
    component: SavedSearchesPage,
  },
  {
    path: '/wishlist',
    redirect: '/shortlists',
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
  {
    path: '/progression',
    name: 'progression',
    component: ProgressionPage,
  },
  {
    path: '/save-import',
    name: 'save-import',
    component: SaveImportPage,
  },
  {
    path: '/save_analysis',
    name: 'save-analysis',
    component: SaveAnalysisPage,
  },
  {
    path: '/save_analysis/:datasetId',
    name: 'shared-save-analysis',
    component: SaveAnalysisPage,
    props: true,
  },
  {
    path: '/scouting-book',
    name: 'scouting-book',
    component: ScoutingBookPage,
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
