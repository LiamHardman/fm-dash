<template>
    <q-page class="wishlist-page">
        <div class="page-container">
            <EmptyState
                v-if="!currentDatasetId"
                icon="folder_off"
                title="No dataset loaded"
                description="Please upload a dataset first to use wishlists."
            >
                <template #actions>
                    <q-btn
                        unelevated
                        color="primary"
                        icon="upload"
                        label="Go to Upload Page"
                        @click="router.push('/upload')"
                    />
                </template>
            </EmptyState>

            <div v-if="currentDatasetId">
                <PageHeader
                    title="Wishlist"
                    subtitle="Keep track of players you're interested in scouting or signing. Your wishlist is automatically saved for this dataset."
                    icon="favorite"
                >
                    <template #actions>
                        <q-btn
                            v-if="wishlistPlayers.length > 0"
                            unelevated
                            icon="clear_all"
                            label="Clear Wishlist"
                            color="negative"
                            @click="confirmClearWishlist"
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
                    </template>
                </PageHeader>

                <!-- Stats -->
                <div v-if="wishlistPlayers.length > 0" class="wishlist-stats">
                    <StatTile icon="favorite" label="Wishlisted Players" :value="wishlistPlayers.length" />
                    <StatTile icon="shield" label="Clubs" :value="uniqueClubsCount" />
                    <StatTile icon="flag" label="Nations" :value="uniqueNationalitiesCount" />
                </div>

                <!-- Wishlist Table -->
                <SectionCard
                    v-if="wishlistPlayers.length > 0"
                    :title="`Wishlisted Players (${wishlistPlayers.length})`"
                    icon="list"
                >
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
                </SectionCard>

                <!-- Empty State -->
                <EmptyState
                    v-else
                    icon="favorite_border"
                    title="Your wishlist is empty"
                    description="Start adding players to your wishlist by right-clicking on them in the player tables and selecting &quot;Add to Wishlist&quot;."
                >
                    <template #actions>
                        <q-btn
                            unelevated
                            icon="search"
                            label="Browse Players"
                            color="primary"
                            @click="goToDataset"
                            size="lg"
                        />
                    </template>
                </EmptyState>
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
import EmptyState from '../components/layout/EmptyState.vue'
import PageHeader from '../components/layout/PageHeader.vue'
import SectionCard from '../components/layout/SectionCard.vue'
import StatTile from '../components/layout/StatTile.vue'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import { usePlayerStore } from '../stores/playerStore'
import { useWishlistStore } from '../stores/wishlistStore'

export default defineComponent({
  name: 'WishlistPage',
  components: {
    PlayerDataTable,
    PlayerDetailDialog,
    PageHeader,
    SectionCard,
    EmptyState,
    StatTile,
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

.page-container {
    max-width: var(--content-max-width);
    margin: 0 auto;
    padding: var(--page-gutter);
}

.wishlist-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--section-gap);
    margin-bottom: var(--section-gap);
}

@media (max-width: 768px) {
    .page-container {
        padding: var(--page-gutter-sm);
    }
}
</style>
