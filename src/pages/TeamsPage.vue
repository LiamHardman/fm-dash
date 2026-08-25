<template>
    <q-page class="teams-view-page">
        <div class="page-container">
            <PageHeader
                title="Teams"
                subtitle="Browse top teams and dive into tactical analysis, formation optimization, and squad insights."
                icon="sports_soccer"
            >
                <template #actions>
                    <q-btn
                        v-if="!pageLoadingError && currentDatasetId"
                        unelevated
                        icon-right="share"
                        label="Share Dataset"
                        color="primary"
                        @click="shareDataset"
                        class="share-btn-modern"
                    >
                        <q-tooltip>Share this dataset with others</q-tooltip>
                    </q-btn>
                </template>
            </PageHeader>

            <!-- Error State -->
            <EmptyState
                v-if="pageLoadingError"
                icon="error"
                title="Couldn't load teams"
                :description="pageLoadingError"
            >
                <template #actions>
                    <q-btn unelevated color="primary" label="Go to Upload Page" @click="router.push('/')" />
                </template>
            </EmptyState>

            <!-- Teams Overview -->
            <EmptyState
                v-if="!pageLoadingError && !pageLoading"
                icon="sports_soccer"
                title="Top Teams Overview"
                description="Click on any team below to view detailed tactical analysis, formation optimization, and squad insights."
            >
                <template #actions>
                    <q-btn
                        v-if="currentDatasetId"
                        color="primary"
                        unelevated
                        label="Browse Dataset"
                        @click="router.push(`/dataset/${currentDatasetId}`)"
                    />
                </template>
            </EmptyState>

            <!-- Loading States -->
            <div v-if="pageLoading" class="loading-state">
                <q-spinner-orbit color="primary" size="4em" />
                <div class="loading-text">Loading teams overview...</div>
            </div>

            <div v-else-if="teamsLoading" class="loading-state">
                <q-spinner-cube color="primary" size="3em" />
                <div class="loading-text">Loading top teams...</div>
            </div>

            <!-- Teams Overview Card -->
            <SectionCard
                v-if="!pageLoading && !pageLoadingError && teamsData.length > 0"
                title="Top Teams Overview"
                icon="sports_soccer"
                class="teams-list-card"
            >
                <p class="card-subtitle">Select a team to analyze player talents and squad composition</p>

                <div class="teams-list">
                    <div
                        v-for="team in (showAllTeams ? teamsData : teamsData.slice(0, 20)).filter(t => t && t.name)"
                        :key="team.name || 'unknown'"
                        class="team-row"
                        @click="selectTeam(team.name)"
                    >
                        <div class="team-logo-container">
                            <TeamLogo
                                :team-name="team.name"
                                :size="32"
                                class="team-logo"
                            />
                        </div>
                        <div class="team-info">
                            <div class="team-name">{{ team.name }}</div>
                            <div class="team-division">{{ team.division }}</div>
                            <div class="player-count">{{ team.playerCount }} players</div>
                        </div>
                        <div class="team-section-ratings">
                            <div
                                class="section-ratings-large"
                                :title="`${team.name} - ATT: ${team.attRating}, MID: ${team.midRating}, DEF: ${team.defRating}`"
                            >
                                <div class="section-rating-large att">
                                    <span class="section-label-large">ATT</span>
                                    <span
                                        class="section-value-large"
                                        :class="getOverallClass(team.attRating)"
                                    >
                                        {{ team.attRating }}
                                    </span>
                                </div>
                                <div class="section-rating-large mid">
                                    <span class="section-label-large">MID</span>
                                    <span
                                        class="section-value-large"
                                        :class="getOverallClass(team.midRating)"
                                    >
                                        {{ team.midRating }}
                                    </span>
                                </div>
                                <div class="section-rating-large def">
                                    <span class="section-label-large">DEF</span>
                                    <span
                                        class="section-value-large"
                                        :class="getOverallClass(team.defRating)"
                                    >
                                        {{ team.defRating }}
                                    </span>
                                </div>
                            </div>
                        </div>
                        <div class="team-overall">
                            <div class="team-rating" :title="`${team.name} - Overall: ${team.bestOverall}`">
                                <div
                                    class="highest-overall-large"
                                    :class="getOverallClass(team.bestOverall)"
                                >
                                    {{ team.bestOverall > 0 ? team.bestOverall + ' AVG' : 'N/A' }}
                                </div>
                                <div class="star-rating-large">
                                    <span
                                        v-for="star in 5"
                                        :key="star"
                                        class="star-large"
                                        :class="getStarClass(team.bestOverall, star)"
                                    >
                                        ★
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Show More Button -->
                <div v-if="!showAllTeams && teamsData.length > 20" class="text-center q-mt-md">
                    <q-btn
                        flat
                        color="primary"
                        @click="showAllTeams = true"
                        class="show-more-btn"
                    >
                        Show All Teams
                        <q-icon name="expand_more" class="q-ml-sm" />
                    </q-btn>
                </div>
            </SectionCard>

            <!-- Additional Banners -->
            <EmptyState
                v-else-if="!pageLoading && !pageLoadingError && teamsData.length > 0"
                icon="info"
                title="Select a team"
                description="Please select a team to view its players and analyze formations."
            />

            <EmptyState
                v-else-if="!pageLoading && !teamsLoading && teamsData.length === 0 && !pageLoadingError"
                icon="warning"
                title="No player data available"
                description="Please upload a player file on the main page first."
            >
                <template #actions>
                    <q-btn unelevated color="primary" label="Go to Upload Page" @click="router.push('/')" />
                </template>
            </EmptyState>
        </div>

    </q-page>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import EmptyState from '../components/layout/EmptyState.vue'
import PageHeader from '../components/layout/PageHeader.vue'
import SectionCard from '../components/layout/SectionCard.vue'
import TeamLogo from '../components/TeamLogo.vue'
import { fetchTopTeams } from '../services/playerService'
import { usePlayerStore } from '../stores/playerStore'

export default {
  name: 'TeamsPage',
  components: { TeamLogo, PageHeader, SectionCard, EmptyState },
  setup() {
    const quasarInstance = useQuasar()
    const router = useRouter()
    const route = useRoute()
    const playerStore = usePlayerStore()

    const pageLoading = ref(true)
    const pageLoadingError = ref('')

    const currentDatasetId = computed(() => playerStore.currentDatasetId)

    // Teams data
    const teamsData = ref([])
    const teamsLoading = ref(false)
    const showAllTeams = ref(false)

    const loadTopTeams = async (datasetId = currentDatasetId.value) => {
      if (!datasetId) {
        teamsData.value = []
        return
      }

      teamsLoading.value = true
      try {
        const teams = await fetchTopTeams(datasetId, 100)
        teamsData.value = teams
      } catch (error) {
        console.error('Error loading top teams:', error)
        pageLoadingError.value = `Failed to load teams data: ${error.message || 'Unknown server error'}. Please try uploading again.`
        quasarInstance.notify({
          type: 'negative',
          message: 'Failed to load teams data',
          position: 'top',
        })
      } finally {
        teamsLoading.value = false
      }
    }

    const selectTeam = (teamName) => {
      // Navigate to team-view page with team filter
      const datasetId = currentDatasetId.value || sessionStorage.getItem('currentDatasetId')
      if (datasetId) {
        router.push({
          path: '/team-view',
          query: {
            datasetId: datasetId,
            team: teamName,
          },
        })
      }
    }

    const handleTeamSelected = (teamName) => {
      // Navigate to team view or handle team selection
      console.log('Team selected:', teamName)
    }

    const getOverallClass = (overall) => {
      if (overall === null || overall === undefined || overall === 0) return 'rating-na'
      const numericOverall = Number(overall)
      if (Number.isNaN(numericOverall)) return 'rating-na'

      if (numericOverall >= 90) return 'rating-tier-6'
      if (numericOverall >= 80) return 'rating-tier-5'
      if (numericOverall >= 70) return 'rating-tier-4'
      if (numericOverall >= 55) return 'rating-tier-3'
      if (numericOverall >= 40) return 'rating-tier-2'
      return 'rating-tier-1'
    }

    const getStarClass = (overall, starPosition) => {
      if (!overall || overall === 0) return 'empty'

      const starRating = getStarRating(overall)

      if (starPosition <= Math.floor(starRating)) {
        return 'filled'
      }
      if (starPosition === Math.floor(starRating) + 1 && starRating % 1 === 0.5) {
        return 'half'
      }
      return 'empty'
    }

    const getStarRating = (overall) => {
      if (!overall || overall === 0) return 0

      if (overall >= 85) return 5
      if (overall >= 82) return 4.5
      if (overall >= 78) return 4
      if (overall >= 74) return 3.5
      if (overall >= 70) return 3
      if (overall >= 67) return 2.5
      if (overall >= 64) return 2
      if (overall >= 60) return 1.5
      if (overall >= 55) return 1
      if (overall >= 50) return 0.5
      return 0
    }

    const shareDataset = () => {
      if (currentDatasetId.value) {
        const shareUrl = `${window.location.origin}/teams?datasetId=${currentDatasetId.value}`
        if (navigator.share) {
          navigator.share({
            title: 'FM24 Teams Dataset',
            text: 'Check out this Football Manager 2024 teams dataset!',
            url: shareUrl,
          })
        } else {
          navigator.clipboard.writeText(shareUrl).then(() => {
            quasarInstance.notify({
              type: 'positive',
              message: 'Share link copied to clipboard!',
              position: 'top',
            })
          })
        }
      }
    }

    onMounted(async () => {
      const datasetIdFromQuery = route.query.datasetId
      const datasetIdFromRoute = route.params.datasetId
      const _teamFromQuery = route.query.team
      const finalDatasetId =
        datasetIdFromRoute || datasetIdFromQuery || sessionStorage.getItem('currentDatasetId')

      if (finalDatasetId) {
        if (
          datasetIdFromQuery &&
          datasetIdFromQuery !== sessionStorage.getItem('currentDatasetId')
        ) {
          sessionStorage.setItem('currentDatasetId', datasetIdFromQuery)
        } else if (!datasetIdFromQuery && sessionStorage.getItem('currentDatasetId')) {
          router.replace({ query: { datasetId: finalDatasetId } })
        }
        playerStore.setCurrentDatasetId(finalDatasetId)
        await loadTopTeams(finalDatasetId)
      } else {
        pageLoadingError.value =
          'No player dataset ID found. Please upload a file on the main page.'
      }
      pageLoading.value = false
    })

    watch(
      () => route.query.datasetId,
      async (newId, oldId) => {
        if (newId && newId !== oldId) {
          sessionStorage.setItem('currentDatasetId', newId)
          playerStore.setCurrentDatasetId(newId)
          teamsData.value = []
          pageLoadingError.value = ''
          pageLoading.value = true
          await loadTopTeams(newId)
          pageLoading.value = false
        }
      }
    )

    return {
      selectTeam,
      pageLoading,
      pageLoadingError,
      getOverallClass,
      getStarClass,
      getStarRating,
      quasarInstance,
      router,
      currentDatasetId,
      shareDataset,
      teamsData,
      teamsLoading,
      showAllTeams,
      handleTeamSelected,
    }
  },
}
</script>

<style lang="scss" scoped>
.teams-view-page {
    min-height: 100vh;
    background: var(--surface-page);
}

.page-container {
    max-width: var(--content-max-width);
    margin: 0 auto;
    padding: var(--page-gutter);

    @media (max-width: 768px) {
        padding: var(--page-gutter-sm);
    }
}

// Loading States
.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 40vh;
    gap: 1rem;

    .loading-text {
        font-size: 1.1rem;
        color: var(--text-secondary);
        font-weight: 500;
    }
}

// Teams Overview Card
.teams-list-card {
    .card-subtitle {
        color: var(--text-secondary);
        margin: 0 0 var(--section-gap) 0;
    }

    .teams-list {
        .team-row {
            display: grid;
            grid-template-columns: auto 1fr auto auto;
            gap: var(--density-gap);
            align-items: center;
            padding: var(--density-cell-padding);
            border-radius: var(--radius-sm);
            cursor: pointer;
            transition: all 0.2s ease;
            border: 1px solid var(--surface-border);
            margin-bottom: var(--density-gap);

            &:hover {
                transform: var(--lift-sm);
                box-shadow: var(--shadow-2);
                background: var(--accent-soft);
            }

            .team-logo-container {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 32px;

                .team-logo {
                    border-radius: 4px;
                    box-shadow: var(--shadow-1);
                }
            }

            .team-info {
                .team-name {
                    font-weight: 700;
                    font-size: 1rem;
                    color: var(--text-primary);
                }

                .team-division {
                    font-size: 0.9rem;
                    color: var(--text-secondary);
                    margin-bottom: 0.25rem;
                }

                .player-count {
                    font-size: 0.85rem;
                    color: var(--text-secondary);
                    margin-top: 0.2rem;
                }
            }

            .team-section-ratings {
                .section-ratings-large {
                    display: flex;
                    gap: 0.5rem;

                    .section-rating-large {
                        display: flex;
                        flex-direction: column;
                        align-items: center;

                        .section-label-large {
                            font-size: 0.7rem;
                            font-weight: 600;
                            text-transform: uppercase;
                            letter-spacing: 0.5px;
                            margin-bottom: 0.2rem;
                            color: var(--text-secondary);
                        }

                        .section-value-large {
                            font-size: 0.8rem;
                            font-weight: 700;
                            padding: 0.2rem 0.4rem;
                            border-radius: 4px;
                            min-width: 24px;
                            text-align: center;
                        }
                    }
                }
            }

            .team-overall {
                text-align: center;

                .team-rating {
                    .highest-overall-large {
                        font-size: 0.9rem;
                        font-weight: 700;
                        margin-bottom: 0.3rem;
                    }

                    .star-rating-large {
                        .star-large {
                            font-size: 0.8rem;
                            margin: 0 0.05rem;

                            &.filled {
                                color: #fbbf24;
                            }

                            &.half {
                                color: #fbbf24;
                                opacity: 0.5;
                            }

                            &.empty {
                                color: var(--text-muted);
                            }
                        }
                    }
                }
            }
        }
    }

    .show-more-btn {
        font-weight: 500;
        text-transform: none;
    }
}

// Rating Tier Classes (semantic quality-tier colors, shared convention across the app)
.rating-tier-6 {
    background-color: #10b981 !important;
    color: white !important;
}

.rating-tier-5 {
    background-color: #34d399 !important;
    color: white !important;
}

.rating-tier-4 {
    background-color: #fbbf24 !important;
    color: white !important;
}

.rating-tier-3 {
    background-color: #f59e0b !important;
    color: white !important;
}

.rating-tier-2 {
    background-color: #ef4444 !important;
    color: white !important;
}

.rating-tier-1 {
    background-color: #991b1b !important;
    color: white !important;
}

.rating-na {
    color: var(--text-muted) !important;
}

// Responsive Design
@media (max-width: 768px) {
    .teams-list-card .teams-list .team-row {
        grid-template-columns: auto 1fr auto;
        gap: 0.75rem;

        .team-section-ratings {
            display: none;
        }
    }
}
</style>
