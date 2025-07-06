<template>
    <q-layout view="hHh lpR fFf">
        <q-header
            flat
            class="app-header"
        >
            <q-toolbar class="header-toolbar">
                <q-toolbar-title class="header-title">
                    <router-link to="/" class="app-title-link">FM-Dash</router-link>
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
                        <router-link to="/team-view" class="nav-link">Team View</router-link>
                        <router-link to="/performance" class="nav-link">Performance</router-link>
                        <router-link to="/nations" class="nav-link">Nations</router-link>
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
    </q-layout>
</template>

<script>
import { computed, defineComponent, onMounted, ref } from 'vue'
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
    SettingsModal
  },
  setup() {
    const uiStore = useUiStore()
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()

    // Initialize analytics with automatic page view tracking
    const _analytics = useAnalytics()

    // Settings modal state
    const showSettingsModal = ref(false)

    onMounted(() => {
      uiStore.initSettings() // Initialize all settings including the new rating calculation setting
    })

    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const wishlistCount = computed(() => wishlistStore.getWishlistCount(currentDatasetId.value))

    return {
      isDarkModeActive: uiStore.isDarkModeActive,
      toggleDarkMode: uiStore.toggleDarkMode,
      currentDatasetId,
      wishlistCount,
      showSettingsModal
    }
  }
})
</script>

<style lang="scss" scoped>
.app-header {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-bottom: 1px solid rgba(26, 35, 126, 0.1);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.05);
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        backdrop-filter: blur(10px);
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
    font-weight: 300;
    font-size: 1.5rem;
    letter-spacing: 2px;
    text-transform: uppercase;
    transition: opacity 0.2s ease;
    
    &:hover {
        opacity: 0.7;
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.nav-links {
    display: flex;
    gap: 2rem;
    margin-left: 3rem;
}

.search-container {
    margin-left: 2rem;
    margin-right: 1rem;
}

.nav-link {
    text-decoration: none;
    color: #666;
    font-weight: 400;
    font-size: 0.9rem;
    letter-spacing: 0.5px;
    padding: 0.5rem 0;
    position: relative;
    transition: color 0.2s ease;
    display: flex;
    align-items: center;
    
    &:hover {
        color: #1a237e;
    }
    
    &.router-link-active {
        color: #1a237e;
        
        &::after {
            content: '';
            position: absolute;
            bottom: 0;
            left: 0;
            right: 0;
            height: 2px;
            background: #1a237e;
        }
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
        
        &:hover {
            color: rgba(255, 255, 255, 0.9);
        }
        
        &.router-link-active {
            color: rgba(255, 255, 255, 0.9);
            
            &::after {
                background: rgba(255, 255, 255, 0.9);
            }
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
        background: #FFDD00;
        border: 1px solid #000000;
        border-radius: 8px;
        height: 36px;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0 1rem;
        text-decoration: none;
        color: #000000;
        font-size: 14px;
        font-weight: 500;
        transition: all 0.2s ease;
        white-space: nowrap;
        
        &:hover {
            background: #FFE55C;
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
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
