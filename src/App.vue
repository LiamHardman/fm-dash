<template>
    <q-layout view="hHh lpR fFf">
        <q-header
            flat
            class="app-header"
        >
            <q-toolbar class="header-toolbar">
                <q-toolbar-title class="header-title">
                    <router-link to="/" class="app-title-link">
                        <q-icon name="sports_soccer" class="brand-icon" />FM-Dash
                    </router-link>
                </q-toolbar-title>
                
                <div class="nav-links">
                    <!-- Always show Upload and Docs -->
                    <router-link to="/upload" class="nav-link">Upload</router-link>
                    <router-link to="/docs" class="nav-link">Docs</router-link>
                    
                    <!-- Only show these links when data is uploaded -->
                    <template v-if="currentDatasetId">
                        <router-link 
                            :to="`/dataset/${currentDatasetId}`" 
                            class="nav-link"
                        >
                            Players
                        </router-link>
                        <!-- <router-link to="/team-view" class="nav-link">Team View</router-link> -->
                        <router-link to="/performance" class="nav-link">Performance</router-link>
                        <router-link to="/nations" class="nav-link">Nations</router-link>
                        <router-link to="/teams" class="nav-link">Teams</router-link>
                        <router-link to="/leagues" class="nav-link">Leagues</router-link>
                        <router-link 
                            to="/wishlist" 
                            class="nav-link wishlist-link"
                        >
                            <q-icon name="favorite" size="1rem" class="q-mr-xs" />
                            Wishlist
                            <q-badge 
                                v-if="wishlistCount > 0" 
                                :label="wishlistCount" 
                                color="positive" 
                                class="q-ml-xs"
                            />
                        </router-link>
                    </template>
                </div>

                <!-- Universal Search Component - only show when data is uploaded -->
                <div v-if="currentDatasetId" class="search-container">
                    <UniversalSearch />
                </div>
                
                <q-space />
                
                <!-- Always show these buttons -->
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

        <q-page-container>
            <router-view />
        </q-page-container>

        <q-footer class="app-footer">
            <div class="footer-content">
                <p>&copy; {{ new Date().getFullYear() }} Liam Hardman.</p>
            </div>
        </q-footer>

        <!-- Settings Modal -->
        <SettingsModal v-model="showSettingsModal" />
        
        <!-- First Time Tutorial Modal -->
        <FirstTimeTutorialModal v-model="showTutorialModal" />
    </q-layout>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref } from 'vue'
import FirstTimeTutorialModal from './components/FirstTimeTutorialModal.vue'
import SettingsModal from './components/SettingsModal.vue'
import UniversalSearch from './components/UniversalSearch.vue'
import { useAnalytics } from './composables/useAnalytics'
import { usePlayerStore } from './stores/playerStore'
import { useUiStore } from './stores/uiStore'
import { useWishlistStore } from './stores/wishlistStore'

export default defineComponent({
  name: 'App',
  components: {
    UniversalSearch,
    SettingsModal,
    FirstTimeTutorialModal,
  },
  setup() {
    const $q = useQuasar()
    const uiStore = useUiStore()
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()

    // Initialize analytics with automatic page view tracking
    const _analytics = useAnalytics()

    // Settings modal state
    const showSettingsModal = ref(false)

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
    })

    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const wishlistCount = computed(() => wishlistStore.getWishlistCount(currentDatasetId.value))

    return {
      isDarkModeActive: uiStore.isDarkModeActive,
      toggleDarkMode: uiStore.toggleDarkMode,
      currentDatasetId,
      wishlistCount,
      showSettingsModal,
      showTutorialModal,
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
    padding: 0 2rem;
    min-height: 60px;
}

.header-title {
    flex: 0 0 auto;
}

.app-title-link {
    text-decoration: none;
    color: #1a237e;
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

    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.brand-icon {
    font-size: 1.3rem;
    opacity: 0.85;
}

.nav-links {
    display: flex;
    gap: 0.25rem;
    margin-left: 3rem;
}

.search-container {
    margin-left: 2rem;
    margin-right: 1rem;
}

.nav-link {
    text-decoration: none;
    color: #555;
    font-weight: 400;
    font-size: 0.9rem;
    letter-spacing: 0.5px;
    padding: 0.4rem 0.75rem;
    border-radius: 20px;
    position: relative;
    transition: color 0.2s ease, background-color 0.2s ease;
    display: flex;
    align-items: center;

    &:hover {
        color: #1a237e;
        background-color: rgba(26, 35, 126, 0.06);
    }

    &.router-link-active {
        color: #1a237e;
        background-color: rgba(26, 35, 126, 0.1);
        font-weight: 500;
    }

    .body--dark & {
        color: rgba(255, 255, 255, 0.65);

        &:hover {
            color: rgba(255, 255, 255, 0.9);
            background-color: rgba(255, 255, 255, 0.08);
        }

        &.router-link-active {
            color: rgba(255, 255, 255, 0.95);
            background-color: rgba(255, 255, 255, 0.12);
            font-weight: 500;
        }
    }
}

.wishlist-link {
    .q-icon {
        transition: color 0.2s ease;
    }
}

.dark-mode-btn {
    color: #666;
    
    &:hover {
        color: #1a237e;
        background: rgba(26, 35, 126, 0.05);
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
        
        &:hover {
            color: rgba(255, 255, 255, 0.9);
            background: rgba(255, 255, 255, 0.1);
        }
    }
}

.settings-btn {
    color: #666;
    margin-right: 0.5rem;
    
    &:hover {
        color: #1a237e;
        background: rgba(26, 35, 126, 0.05);
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
        
        &:hover {
            color: rgba(255, 255, 255, 0.9);
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
        border: 1.5px solid #1a237e;
        border-radius: 8px;
        height: 36px;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0 1rem;
        text-decoration: none;
        color: #1a237e;
        font-size: 13px;
        font-weight: 500;
        transition: all 0.2s ease;
        white-space: nowrap;

        &:hover {
            background: rgba(26, 35, 126, 0.08);
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(26, 35, 126, 0.15);
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

.app-footer {
    background: transparent;
    border-top: 1px solid rgba(26, 35, 126, 0.1);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.02);
        border-top: 1px solid rgba(255, 255, 255, 0.1);
    }
}

.footer-content {
    padding: 1rem 2rem;
    text-align: center;
    
    p {
        margin: 0;
        color: #666;
        font-size: 0.85rem;
        font-weight: 300;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.6);
        }
    }
}

@media (max-width: 768px) {
    .header-toolbar {
        padding: 0 1rem;
        min-height: 56px;
    }
    
    .app-title-link {
        font-size: 1.2rem;
        letter-spacing: 1px;
    }
    
    .nav-links {
        display: none;
    }
    
    .search-container {
        margin-left: 1rem;
        margin-right: 0.5rem;
    }
    
    .bmc-button-wrapper {
        margin-right: 0.5rem;
        
        .bmc-button {
            height: 32px !important;
            font-size: 12px !important;
        }
    }
    
    .footer-content {
        padding: 1rem;
        
        p {
            font-size: 0.8rem;
        }
    }
}
</style>
