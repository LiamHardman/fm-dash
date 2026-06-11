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
                <PageHero
                    rounded
                    icon="favorite"
                    badge="Wishlist"
                    title="Your Football Manager"
                    highlight="Wishlist"
                    subtitle="Keep track of players you're interested in scouting or signing. Your wishlist is automatically saved for this dataset."
                >
                    <template #side>
                        <div class="hero-stats">
                            <div class="hero-stat">
                                <div class="hero-stat-number">{{ wishlistPlayers.length }}</div>
                                <div class="hero-stat-label">Wishlisted Players</div>
                            </div>
                            <div class="hero-stat">
                                <div class="hero-stat-number">{{ uniqueClubsCount }}</div>
                                <div class="hero-stat-label">Clubs</div>
                            </div>
                            <div class="hero-stat">
                                <div class="hero-stat-number">{{ uniqueNationalitiesCount }}</div>
                                <div class="hero-stat-label">Nations</div>
                            </div>
                        </div>
                    </template>
                </PageHero>

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
import PageHero from '../components/PageHero.vue'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import { usePlayerStore } from '../stores/playerStore'
import { useWishlistStore } from '../stores/wishlistStore'

export default defineComponent({
  name: 'WishlistPage',
  components: {
    PageHero,
    PlayerDataTable,
    PlayerDetailDialog,
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
        position: 'top',
      })
    }

    const handlePlayerSelected = (player) => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    const handleTeamSelected = (team) => {
      if (currentDatasetId.value) {
        const url = router.resolve({
          path: '/team-view',
          query: {
            datasetId: currentDatasetId.value,
            team: team,
          },
        }).href
        const newWindow = window.open(url, '_blank')
        if (!newWindow) {
        } else {
        }
      } else {
      }
    }

    const handleRemoveFromWishlist = async (player) => {
      const success = await wishlistStore.removeFromWishlist(currentDatasetId.value, player)
      if (success) {
        quasarInstance.notify({
          type: 'positive',
          message: `${player.name} removed from wishlist`,
          position: 'top',
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
      handleRemoveFromWishlist,
    }
  },
})
</script>

<style lang="scss" scoped>
.wishlist-page {
    min-height: 100vh;
}

// Hero styling comes from the shared PageHero component.

.actions-section {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
}
</style> 