<template>
    <q-layout view="hHh lpR fFf">
        <q-header flat class="app-header">
            <q-toolbar class="header-toolbar">
                <q-btn
                    flat
                    round
                    icon="menu"
                    @click="toggleSidebar"
                    class="sidebar-toggle-btn"
                    aria-label="Toggle navigation"
                />
                <q-toolbar-title class="header-title">
                    <router-link to="/" class="app-title-link">
                        <q-icon name="sports_soccer" class="brand-icon" />FM-Dash
                    </router-link>
                </q-toolbar-title>

                <!-- Universal Search Component - only show when data is uploaded -->
                <div v-if="currentDatasetId" class="search-container">
                    <UniversalSearch />
                </div>
                <DatasetSwitcher />

                <q-space />

                <!-- Buy me a coffee button -->
                <div class="bmc-button-wrapper">
                    <a
                        href="https://www.buymeacoffee.com/LiamHardman"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="bmc-button"
                    >
                        <span class="bmc-text">☕ Buy me a coffee</span>
                    </a>
                </div>

                <!-- Settings button -->
                <q-btn
                    flat
                    round
                    icon="settings"
                    @click="showSettingsModal = true"
                    class="settings-btn"
                >
                    <q-tooltip>Settings</q-tooltip>
                </q-btn>

                <!-- Dark mode toggle -->
                <q-btn
                    flat
                    round
                    :icon="$q.dark.isActive ? 'light_mode' : 'dark_mode'"
                    @click="toggleDarkMode"
                    class="dark-mode-btn"
                >
                    <q-tooltip>{{ $q.dark.isActive ? 'Light Mode' : 'Dark Mode' }}</q-tooltip>
                </q-btn>
            </q-toolbar>
        </q-header>

        <!-- Persistent sidebar navigation: embedded panel above 768px (collapsible to
             an icon rail via sidebarMini), overlay drawer below it (toggled by the
             header's menu button via sidebarOpen). Replaces the old flat top-nav
             link list, which had grown to 8+ links with no room to grow further. -->
        <q-drawer
            v-model="sidebarOpen"
            show-if-above
            :mini="sidebarMini"
            :width="248"
            :mini-width="72"
            :breakpoint="SIDEBAR_BREAKPOINT"
            bordered
            class="app-sidebar"
        >
            <div class="sidebar-inner">
                <q-list padding class="sidebar-nav-list">
                    <q-item
                        v-for="item in visibleNavItems"
                        :key="item.id"
                        clickable
                        v-ripple
                        :to="item.to"
                        :exact="item.exact"
                        active-class="sidebar-nav-item--active"
                        class="sidebar-nav-item"
                        @click="sidebarOpen = false"
                    >
                        <q-item-section avatar>
                            <q-icon :name="item.icon" />
                        </q-item-section>
                        <q-item-section v-if="!sidebarMini">
                            {{ item.label }}
                            <q-badge
                                v-if="item.badge"
                                :label="item.badge"
                                color="positive"
                                class="q-ml-xs"
                            />
                        </q-item-section>
                        <q-tooltip v-if="sidebarMini" anchor="center right" self="center left">
                            {{ item.label }}
                        </q-tooltip>
                    </q-item>
                </q-list>

                <div class="sidebar-spacer" />

                <div class="sidebar-footer">
                    <q-btn
                        flat
                        dense
                        no-caps
                        :icon="sidebarMini ? 'tune' : undefined"
                        :label="sidebarMini ? '' : 'Customize'"
                        class="sidebar-footer-btn"
                        @click="showCustomizeDialog = true"
                    >
                        <q-tooltip v-if="sidebarMini">Customize navigation</q-tooltip>
                    </q-btn>
                    <q-btn
                        flat
                        dense
                        round
                        :icon="sidebarMini ? 'chevron_right' : 'chevron_left'"
                        class="sidebar-collapse-btn gt-xs"
                        aria-label="Collapse sidebar"
                        @click="sidebarMini = !sidebarMini"
                    />
                </div>
            </div>
        </q-drawer>

        <q-page-container>
            <ErrorBoundary><router-view /></ErrorBoundary>
        </q-page-container>

        <!-- Settings Modal -->
        <SettingsModal v-model="showSettingsModal" />

        <!-- First Time Tutorial Modal -->
        <FirstTimeTutorialModal v-model="showTutorialModal" />

        <!-- Sidebar navigation customization (hide/reorder items) -->
        <CustomizeListDialog
            v-model="showCustomizeDialog"
            title="Customize navigation"
            hint="Reorder or hide sidebar items. Dataset-only items stay hidden until a dataset is loaded, regardless of this setting."
            :items="allNavItemsMeta"
            :hidden-ids="uiStore.sidebarHiddenItems"
            :order-ids="uiStore.sidebarItemOrder"
            @update:hidden="uiStore.setSidebarHiddenItems"
            @update:order="uiStore.setSidebarItemOrder"
            @reset="() => { uiStore.setSidebarHiddenItems([]); uiStore.setSidebarItemOrder([]) }"
        />

        <!-- FM-Dash Chatbot — dataset-scoped, only mounted once a dataset is loaded -->
        <ChatWidget v-if="currentDatasetId" />
    </q-layout>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref } from 'vue'
import ChatWidget from './components/ChatWidget.vue'
import DatasetSwitcher from './components/DatasetSwitcher.vue'
import ErrorBoundary from './components/ErrorBoundary.vue'
import FirstTimeTutorialModal from './components/FirstTimeTutorialModal.vue'
import CustomizeListDialog from './components/layout/CustomizeListDialog.vue'
import SettingsModal from './components/SettingsModal.vue'
import UniversalSearch from './components/UniversalSearch.vue'
import { useAnalytics } from './composables/useAnalytics'
import playerService from './services/playerService'
import { usePlayerStore } from './stores/playerStore'
import { useSavedSearchStore } from './stores/savedSearchStore'
import { useShortlistStore } from './stores/shortlistStore'
import { useUiStore } from './stores/uiStore'
import { useWeightsStore } from './stores/weightsStore'

export default defineComponent({
  name: 'App',
  components: {
    UniversalSearch,
    SettingsModal,
    FirstTimeTutorialModal,
    CustomizeListDialog,
    ErrorBoundary,
    ChatWidget,
    DatasetSwitcher,
  },
  setup() {
    const $q = useQuasar()
    const uiStore = useUiStore()
    const playerStore = usePlayerStore()
    const shortlistStore = useShortlistStore()
    const savedSearchStore = useSavedSearchStore()
    const weightsStore = useWeightsStore()

    // Initialize analytics with automatic page view tracking
    const _analytics = useAnalytics()

    // Settings modal state
    const showSettingsModal = ref(false)

    // Sidebar customization dialog state
    const showCustomizeDialog = ref(false)

    // Sidebar state: sidebarOpen controls the overlay drawer below the 768px
    // breakpoint; sidebarMini controls the icon-rail collapse above it.
    const sidebarOpen = ref(false)
    const sidebarMini = ref(false)

    // Must match the q-drawer's own :breakpoint="768" below -- using Quasar's
    // $q.screen.lt.sm (600px) here instead left a 600-767px gap where the
    // drawer was already in mobile-overlay mode (only responds to
    // sidebarOpen) but this toggled sidebarMini instead, so the hamburger
    // button did nothing visible in that width range.
    const SIDEBAR_BREAKPOINT = 768

    const toggleSidebar = () => {
      if ($q.screen.width < SIDEBAR_BREAKPOINT) {
        sidebarOpen.value = !sidebarOpen.value
      } else {
        sidebarMini.value = !sidebarMini.value
      }
    }

    // Tutorial modal state
    const showTutorialModal = computed({
      get: () => uiStore.showFirstTimeTutorial,
      set: (value) => {
        if (value) {
          uiStore.showTutorial()
        } else {
          uiStore.hideTutorial()
        }
      },
    })

    onMounted(() => {
      // Expose Quasar instance globally for the UI store
      window.$q = $q

      uiStore.initSettings() // Initialize all settings including the new rating calculation setting
      shortlistStore.load()
      savedSearchStore.load()

      weightsStore.load()
      // Re-apply the saved active weight profile to the backend -- the API's
      // in-memory weights reset to defaults on restart, but the browser's
      // choice of active profile should survive that.
      if (weightsStore.activeProfile) {
        playerService
          .updateConfig({ attributeWeights: weightsStore.activeProfile.weights })
          .catch(() => {})
      }
    })

    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const shortlistCount = computed(() => shortlistStore.activeList?.items?.length || 0)

    // Full nav metadata. requiresDataset items only render once a dataset is
    // loaded; `to` is computed per-item since Players needs the current
    // dataset id in its path.
    const allNavItemsMeta = computed(() => [
      { id: 'home', label: 'Home', icon: 'home', to: '/', exact: true },
      { id: 'upload', label: 'Upload', icon: 'upload', to: '/upload' },
      { id: 'progression', label: 'Progression', icon: 'trending_up', to: '/progression' },
      { id: 'saved-searches', label: 'Saved searches', icon: 'bookmarks', to: '/saved-searches' },
      { id: 'save-import', label: 'Save Import', icon: 'science', to: '/save-import' },
      { id: 'docs', label: 'Docs', icon: 'menu_book', to: '/docs' },
      {
        id: 'players',
        label: 'Players',
        icon: 'groups',
        to: currentDatasetId.value ? `/dataset/${currentDatasetId.value}` : '/upload',
        requiresDataset: true,
      },
      {
        id: 'performance',
        label: 'Performance',
        icon: 'leaderboard',
        to: currentDatasetId.value ? `/performance/${currentDatasetId.value}` : '/upload',
        requiresDataset: true,
      },
      {
        id: 'nations',
        label: 'Nations',
        icon: 'flag',
        to: `/nations/${currentDatasetId.value}`,
        requiresDataset: true,
      },
      {
        id: 'teams',
        label: 'Teams',
        icon: 'shield',
        to: `/teams/${currentDatasetId.value}`,
        requiresDataset: true,
      },
      {
        id: 'leagues',
        label: 'Leagues',
        icon: 'emoji_events',
        to: `/leagues/${currentDatasetId.value}`,
        requiresDataset: true,
      },
      {
        id: 'shortlists',
        label: 'Shortlists',
        icon: 'favorite',
        to: '/shortlists',
        badge: shortlistCount.value > 0 ? shortlistCount.value : null,
      },
      {
        id: 'scouting-book',
        label: 'Scouting Book',
        icon: 'menu_book',
        to: '/scouting-book',
        requiresDataset: true,
      },
    ])

    const visibleNavItems = computed(() => {
      const hidden = new Set(uiStore.sidebarHiddenItems)
      const order = uiStore.sidebarItemOrder
      let items = allNavItemsMeta.value.filter(
        (item) => (!item.requiresDataset || currentDatasetId.value) && !hidden.has(item.id)
      )
      if (order.length) {
        items = [...items].sort((a, b) => {
          const ai = order.indexOf(a.id)
          const bi = order.indexOf(b.id)
          return (ai === -1 ? Infinity : ai) - (bi === -1 ? Infinity : bi)
        })
      }
      return items
    })

    return {
      uiStore,
      isDarkModeActive: uiStore.isDarkModeActive,
      toggleDarkMode: uiStore.toggleDarkMode,
      currentDatasetId,
      shortlistCount,
      showSettingsModal,
      showTutorialModal,
      showCustomizeDialog,
      sidebarOpen,
      sidebarMini,
      toggleSidebar,
      SIDEBAR_BREAKPOINT,
      allNavItemsMeta,
      visibleNavItems,
    }
  },
})
</script>

<style lang="scss" scoped>
.app-header {
    background: rgba(255, 255, 255, 0.96);
    backdrop-filter: blur(14px);
    box-shadow: 0 2px 16px rgba(26, 35, 126, 0.08);

    .body--dark & {
        background: rgba(16, 16, 28, 0.88);
        box-shadow: 0 2px 16px rgba(0, 0, 0, 0.35);
        backdrop-filter: blur(14px);
    }
}

.header-toolbar {
    padding: 0 1.25rem;
    min-height: 60px;
}

.sidebar-toggle-btn {
    color: var(--text-secondary);
    margin-right: 0.25rem;
}

.header-title {
    flex: 0 0 auto;
}

.app-title-link {
    text-decoration: none;
    color: var(--text-primary);
    font-weight: 400;
    font-size: 1.5rem;
    letter-spacing: 2px;
    text-transform: uppercase;
    transition: opacity 0.2s ease;
    display: flex;
    align-items: center;
    gap: 0.45rem;

    &:hover {
        opacity: 0.7;
    }
}

.brand-icon {
    font-size: 1.3rem;
    opacity: 0.85;
}

.search-container {
    margin-left: 1.5rem;
    margin-right: 1rem;
}

.dark-mode-btn {
    color: var(--text-secondary);

    &:hover {
        color: var(--accent);
        background: var(--accent-soft);
    }

    .body--dark & {
        &:hover {
            background: rgba(255, 255, 255, 0.1);
        }
    }
}

.settings-btn {
    color: var(--text-secondary);
    margin-right: 0.5rem;

    &:hover {
        color: var(--accent);
        background: var(--accent-soft);
    }

    .body--dark & {
        &:hover {
            background: rgba(255, 255, 255, 0.1);
        }
    }
}

.bmc-button-wrapper {
    margin-right: 1rem;
    display: flex;
    align-items: center;

    .bmc-button {
        background: transparent;
        border: 1.5px solid var(--accent);
        border-radius: 8px;
        height: 36px;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0 1rem;
        text-decoration: none;
        color: var(--accent);
        font-size: 13px;
        font-weight: 500;
        transition: all 0.2s ease;
        white-space: nowrap;

        &:hover {
            background: var(--accent-soft);
            transform: translateY(-1px);
            box-shadow: 0 2px 8px var(--accent-soft-strong);
        }

        .body--dark & {
            border-color: rgba(255, 255, 255, 0.45);
            color: rgba(255, 255, 255, 0.85);

            &:hover {
                background: rgba(255, 255, 255, 0.1);
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
            }
        }

        .bmc-text {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
    }
}

// ─── Sidebar ──────────────────────────────────────────────────────────────

.app-sidebar {
    background: var(--surface-card);

    :deep(.q-drawer__content) {
        display: flex;
        flex-direction: column;
    }
}

.sidebar-inner {
    display: flex;
    flex-direction: column;
    height: 100%;
}

.sidebar-nav-list {
    flex: 0 0 auto;
}

.sidebar-nav-item {
    border-radius: var(--radius-sm);
    margin: 0 0.5rem 0.15rem 0.5rem;
    color: var(--text-secondary);

    .q-icon {
        color: var(--text-secondary);
    }

    &:hover {
        background: var(--accent-soft);
        color: var(--text-primary);
    }

    &--active {
        background: var(--accent-soft-strong);
        color: var(--accent);
        font-weight: 600;

        .q-icon {
            color: var(--accent);
        }
    }
}

.sidebar-spacer {
    flex: 1 1 auto;
}

.sidebar-footer {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.6rem;
    border-top: 1px solid var(--surface-border);
}

.sidebar-footer-btn {
    color: var(--text-secondary);
}

// ─── Responsive header ───────────────────────────────────────────────────
// Ported forward from the pre-redesign top-nav's own 768px media query (see
// git history), which shrank the header's chrome on narrow viewports. The
// old rule also hid `.nav-links`/showed a `.mobile-menu-btn` -- both no
// longer exist now that nav lives entirely in the q-drawer sidebar (which
// has its own :breakpoint="768" prop) -- but the header padding/title/
// bmc-button sizing adjustments still apply since those elements are still
// in the top bar and were otherwise left un-migrated when the flat nav was
// replaced.
@media (max-width: 768px) {
    .header-toolbar {
        padding: 0 0.75rem;
        min-height: 56px;
    }

    .app-title-link {
        font-size: 1.15rem;
        letter-spacing: 1px;
    }

    .search-container {
        margin-left: 0.75rem;
        margin-right: 0.5rem;
    }

    .bmc-button-wrapper {
        margin-right: 0.25rem;

        .bmc-button {
            height: 32px;
            padding: 0 0.6rem;
            font-size: 12px;
        }
    }
}
</style>
