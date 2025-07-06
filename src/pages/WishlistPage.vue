<template>
    <q-page class="wishlist-page">
        <div class="q-pa-md">
            <q-banner
                v-if="!currentDatasetId"
                class="text-white bg-info q-mb-md"
                rounded
            >
                <template v-slot:avatar>
                    <q-icon name="info" />
                </template>
                No dataset loaded. Please upload a dataset first to use wishlists.
                <q-btn
                    flat
                    color="white"
                    label="Go to Upload Page"
                    @click="router.push('/upload')"
                    class="q-ml-md"
                />
            </q-banner>

            <div v-if="currentDatasetId">
                <!-- Hero Section -->
                <div class="hero-section">
                    <div class="hero-container">
                        <div class="hero-content">
                            <div class="hero-badge">
                                <q-icon name="favorite" size="1.2rem" />
                                <span>Wishlist</span>
                            </div>
                            <h1 class="hero-title">
                                Your Football Manager
                                <span class="gradient-text">Wishlist</span>
                            </h1>
                            <p class="hero-subtitle">
                                Keep track of players you're interested in scouting or signing. 
                                Your wishlist is automatically saved for this dataset.
                            </p>
                        </div>
                        <div class="hero-stats">
                            <div class="stat-card">
                                <div class="stat-number">{{ wishlistPlayers.length }}</div>
                                <div class="stat-label">Wishlisted Players</div>
                            </div>
                            <div class="stat-card">
                                <div class="stat-number">{{ uniqueClubsCount }}</div>
                                <div class="stat-label">Clubs</div>
                            </div>
                            <div class="stat-card">
                                <div class="stat-number">{{ uniqueNationalitiesCount }}</div>
                                <div class="stat-label">Nations</div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Action Buttons -->
                <div class="actions-section q-mb-md">
                    <q-btn
                        v-if="wishlistPlayers.length > 0"
                        unelevated
                        icon="clear_all"
                        label="Clear Wishlist"
                        color="negative"
                        @click="confirmClearWishlist"
                        class="q-mr-md"
                    >
                        <q-tooltip>Remove all players from wishlist</q-tooltip>
                    </q-btn>
                    <q-btn
                        unelevated
                        icon="arrow_back"
                        label="Back to Players"
                        color="primary"
                        @click="goToDataset"
                    >
                        <q-tooltip>Return to player dataset</q-tooltip>
                    </q-btn>
                </div>

                <!-- Wishlist Table -->
                <q-card
                    v-if="wishlistPlayers.length > 0"
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-sm">
                            Wishlisted Players ({{ wishlistPlayers.length }})
                        </div>
                        <PlayerDataTable
                            :players="wishlistPlayers"
                            :loading="wishlistStore.loading"
                            @player-selected="handlePlayerSelected"
                            @team-selected="handleTeamSelected"
                            :is-goalkeeper-view="false"
                            :currency-symbol="detectedCurrencySymbol"
                            :dataset-id="currentDatasetId"
                            :show-wishlist-actions="true"
                            @remove-from-wishlist="handleRemoveFromWishlist"
                        />
                    </q-card-section>
                </q-card>

                <!-- Empty State -->
                <q-card
                    v-else
                    class="q-pa-xl text-center"
                    :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-grey-1'"
                    flat
                    bordered
                >
                    <q-icon 
                        name="favorite_border" 
                        size="4rem" 
                        :color="quasarInstance.dark.isActive ? 'grey-6' : 'grey-5'"
                        class="q-mb-md"
                    />
                    <h3 
                        class="text-h5 q-mb-md" 
                        :class="quasarInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
                    >
                        Your wishlist is empty
                    </h3>
                    <p 
                        class="text-body1 q-mb-lg" 
                        :class="quasarInstance.dark.isActive ? 'text-grey-5' : 'text-grey-6'"
                    >
                        Start adding players to your wishlist by right-clicking on them in the player tables 
                        and selecting "Add to Wishlist".
                    </p>
                    <q-btn
                        unelevated
                        icon="search"
                        label="Browse Players"
                        color="primary"
                        @click="goToDataset"
                        size="lg"
                    />
                </q-card>
            </div>
        </div>

        <!-- Clear Wishlist Confirmation Dialog -->
        <q-dialog v-model="showClearDialog">
            <q-card style="min-width: 350px">
                <q-card-section>
                    <div class="text-h6">Clear Wishlist</div>
                </q-card-section>

                <q-card-section class="q-pt-none">
                    Are you sure you want to remove all {{ wishlistPlayers.length }} players from your wishlist? 
                    This action cannot be undone.
                </q-card-section>

                <q-card-actions align="right">
                    <q-btn flat label="Cancel" color="primary" v-close-popup />
                    <q-btn 
                        flat 
                        label="Clear Wishlist" 
                        color="negative" 
                        @click="clearWishlist"
                        v-close-popup 
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>

        <!-- Player Detail Dialog -->
        <PlayerDetailDialog
            :player="playerForDetailView"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
        />
    </q-page>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import { usePlayerStore } from '../stores/playerStore'
import { useWishlistStore } from '../stores/wishlistStore'

export default defineComponent({
  name: 'WishlistPage',
  components: {
    PlayerDataTable,
    PlayerDetailDialog
  },
  setup() {
    const router = useRouter()
    const quasarInstance = useQuasar()
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()
    const showClearDialog = ref(false)
    const showPlayerDetailDialog = ref(false)
    const playerForDetailView = ref(null)

    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)

    const wishlistPlayers = computed(() => {
      return wishlistStore.getWishlistForDataset(currentDatasetId.value)
    })

    const uniqueClubsCount = computed(() => {
      const clubs = new Set()
      for (const player of wishlistPlayers.value) {
        if (player.club) clubs.add(player.club)
      }
      return clubs.size
    })

    const uniqueNationalitiesCount = computed(() => {
      const nationalities = new Set()
      for (const player of wishlistPlayers.value) {
        if (player.nationality) nationalities.add(player.nationality)
      }
      return nationalities.size
    })

    // Initialize wishlist when component is mounted
    onMounted(async () => {
      if (currentDatasetId.value) {
        await wishlistStore.initializeWishlistForDataset(currentDatasetId.value)
      }
    })

    const goToDataset = () => {
      if (currentDatasetId.value) {
        router.push(`/dataset/${currentDatasetId.value}`)
      } else {
        router.push('/upload')
      }
    }

    const confirmClearWishlist = () => {
      showClearDialog.value = true
    }

    const clearWishlist = async () => {
      await wishlistStore.clearWishlistForDataset(currentDatasetId.value)
      quasarInstance.notify({
        type: 'positive',
        message: 'Wishlist cleared successfully',
        position: 'top'
      })
    }

    const handlePlayerSelected = player => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    const handleTeamSelected = team => {
      if (currentDatasetId.value) {
        const url = router.resolve({
          path: '/team-view',
          query: {
            datasetId: currentDatasetId.value,
            team: team
          }
        }).href
        const newWindow = window.open(url, '_blank')
        if (!newWindow) {
          console.error('Failed to open new window - likely blocked by popup blocker')
        } else {
        }
      } else {
      }
    }

    const handleRemoveFromWishlist = async player => {
      const success = await wishlistStore.removeFromWishlist(currentDatasetId.value, player)
      if (success) {
        quasarInstance.notify({
          type: 'positive',
          message: `${player.name} removed from wishlist`,
          position: 'top'
        })
      }
    }

    return {
      router,
      quasarInstance,
      wishlistStore,
      currentDatasetId,
      detectedCurrencySymbol,
      wishlistPlayers,
      uniqueClubsCount,
      uniqueNationalitiesCount,
      showClearDialog,
      showPlayerDetailDialog,
      playerForDetailView,
      goToDataset,
      confirmClearWishlist,
      clearWishlist,
      handlePlayerSelected,
      handleTeamSelected,
      handleRemoveFromWishlist
    }
  }
})
</script>

<style lang="scss" scoped>
.wishlist-page {
    min-height: 100vh;
}

.hero-section {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 24px;
    padding: 3rem 0;
    margin-bottom: 2rem;
    overflow: hidden;
    position: relative;

    &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(45deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0) 100%);
    }

    .body--dark & {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        
        &::before {
            background: linear-gradient(45deg, rgba(0,0,0,0.2) 0%, rgba(0,0,0,0) 100%);
        }
    }
}

.hero-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    position: relative;
    z-index: 1;

    @media (max-width: 768px) {
        flex-direction: column;
        text-align: center;
        gap: 2rem;
    }
}

.hero-content {
    flex: 1;
    max-width: 600px;
}

.hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 50px;
    color: white;
    font-size: 0.875rem;
    font-weight: 500;
    margin-bottom: 1.5rem;
    backdrop-filter: blur(10px);
}

.hero-title {
    font-size: 3rem;
    font-weight: 300;
    color: white;
    margin-bottom: 1rem;
    line-height: 1.1;

    @media (max-width: 768px) {
        font-size: 2.5rem;
    }
}

.gradient-text {
    background: linear-gradient(45deg, #ffd700, #ffed4e);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    font-weight: 400;
}

.hero-subtitle {
    font-size: 1.125rem;
    color: rgba(255, 255, 255, 0.9);
    line-height: 1.6;
    margin-bottom: 0;
}

.hero-stats {
    display: flex;
    gap: 1.5rem;

    @media (max-width: 768px) {
        justify-content: center;
    }
}

.stat-card {
    background: rgba(255, 255, 255, 0.15);
    padding: 1.5rem;
    border-radius: 16px;
    text-align: center;
    min-width: 120px;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
}

.stat-number {
    font-size: 2rem;
    font-weight: 600;
    color: white;
    margin-bottom: 0.5rem;
}

.stat-label {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.8);
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.actions-section {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
}
</style> 