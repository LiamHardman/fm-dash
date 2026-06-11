<template>
    <q-page class="teams-view-page">
        <div class="main-content">
            <!-- Error Banner -->
            <q-banner
                v-if="pageLoadingError"
                class="error-banner"
                rounded
            >
                <template v-slot:avatar>
                    <q-icon name="error" />
                </template>
                {{ pageLoadingError }}
                <q-btn
                    flat
                    color="white"
                    label="Go to Upload Page"
                    @click="router.push('/')"
                    class="q-ml-md"
                />
            </q-banner>

            <!-- Share Button - Modern Design -->
            <div v-if="!pageLoadingError && currentDatasetId" class="share-section">
                <q-btn
                    unelevated
                    icon-right="share"
                    label="Share Dataset"
                    color="primary"
                    @click="shareDataset"
                    class="share-btn-modern"
                    size="md"
                >
                    <q-tooltip>Share this dataset with others</q-tooltip>
                </q-btn>
            </div>

            <!-- Teams Overview -->
            <div v-if="!pageLoadingError && !pageLoading" class="teams-overview">
                <q-card class="teams-overview-card">
                    <q-card-section class="empty-state-content">
                        <div class="empty-state-icon">
                            <q-icon name="sports_soccer" size="4rem" />
                        </div>
                        <h3 class="empty-state-title">Top Teams Overview</h3>
                        <p class="empty-state-description">
                            Click on any team below to view detailed tactical analysis, formation optimization, and squad insights.
                        </p>
                        <q-btn
                            color="primary"
                            unelevated
                            label="Browse Dataset"
                            @click="router.push(`/dataset/${currentDatasetId}`)"
                            v-if="currentDatasetId"
                            class="empty-state-btn"
                        />
                    </q-card-section>
                </q-card>
            </div>

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
            <q-card
                v-if="!pageLoading && !pageLoadingError && teamsData.length > 0"
                class="teams-overview-card"
            >
                <q-card-section>
                    <div class="card-header">
                        <h3 class="card-title">
                            <q-icon name="sports_soccer" class="card-icon" />
                            Top Teams Overview
                        </h3>
                        <p class="card-subtitle">Select a team to analyze player talents and squad composition</p>
                    </div>
                    

                    
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
                </q-card-section>
            </q-card>

            <!-- Additional Banners -->
            <q-banner
                v-else-if="!pageLoading && !pageLoadingError && teamsData.length > 0"
                class="info-banner"
            >
                <template v-slot:avatar>
                    <q-icon name="info" />
                </template>
                Please select a team to view its players and analyze formations.
            </q-banner>
            
            <q-banner
                v-else-if="!pageLoading && !teamsLoading && teamsData.length === 0 && !pageLoadingError"
                class="warning-banner"
            >
                <template v-slot:avatar>
                    <q-icon name="warning" />
                </template>
                No player data available. Please upload a player file on the main page first.
                <q-btn
                    flat
                    color="primary"
                    label="Go to Upload Page"
                    @click="router.push('/')"
                    class="q-ml-md"
                />
            </q-banner>
        </div>

    </q-page>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TeamLogo from '../components/TeamLogo.vue'
import { fetchTopTeams } from '../services/playerService'
import { usePlayerStore } from '../stores/playerStore'

export default {
  name: 'TeamsPage',
  components: { TeamLogo },
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
// Variables
$border-radius: 12px;
$border-radius-small: 8px;
$card-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
$card-shadow-hover: 0 4px 20px rgba(0, 0, 0, 0.12);

// Base Page Layout
.teams-view-page {
    min-height: 100vh;
    background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
    
    .body--dark & {
        background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
    }
}

.main-content {
    max-width: 1400px;
    margin: 0 auto;
    padding: 2rem;
    
    @media (max-width: 768px) {
        padding: 1rem;
    }
}

// Banners
.error-banner {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
    color: white;
    margin-bottom: 2rem;
    border: none;
    box-shadow: $card-shadow;
}

.info-banner {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    margin-bottom: 2rem;
    border: none;
    box-shadow: $card-shadow;
}

.warning-banner {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
    color: white;
    margin-bottom: 2rem;
    border: none;
    box-shadow: $card-shadow;
}

// Share Section
.share-section {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 2rem;
    
    .share-btn-modern {
        border-radius: $border-radius;
        font-weight: 500;
        text-transform: none;
        box-shadow: $card-shadow;
        transition: all 0.2s ease;
        
        &:hover {
            transform: translateY(-1px);
            box-shadow: $card-shadow-hover;
        }
    }
}

// Empty State
.empty-state {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 60vh;
    
    .empty-state-card {
        max-width: 500px;
        width: 100%;
        border-radius: $border-radius;
        box-shadow: $card-shadow;
        background: white;
        
        .body--dark & {
            background: #1e293b;
        }
        
        .empty-state-content {
            text-align: center;
            padding: 3rem 2rem;
            
            .empty-state-icon {
                margin-bottom: 1.5rem;
                color: #64748b;
                
                .body--dark & {
                    color: #94a3b8;
                }
            }
            
            .empty-state-title {
                font-size: 1.5rem;
                font-weight: 600;
                margin-bottom: 1rem;
                color: #1e293b;
                
                .body--dark & {
                    color: #f1f5f9;
                }
            }
            
            .empty-state-description {
                color: #64748b;
                margin-bottom: 2rem;
                line-height: 1.6;
                
                .body--dark & {
                    color: #94a3b8;
                }
            }
            
            .empty-state-btn {
                border-radius: $border-radius;
                font-weight: 500;
                text-transform: none;
            }
        }
    }
}

// Loading States
.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    gap: 1rem;
    
    .loading-text {
        font-size: 1.1rem;
        color: #64748b;
        font-weight: 500;
        
        .body--dark & {
            color: #94a3b8;
        }
    }
}

// Team Dashboard
.team-dashboard {
    .team-hero-section {
        background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
        border-radius: $border-radius;
        margin-bottom: 2rem;
        padding: 2rem;
        color: white;
        box-shadow: $card-shadow;
        
        .team-hero-content {
            .team-primary-info {
                display: flex;
                justify-content: space-between;
                align-items: flex-start;
                gap: 2rem;
                
                @media (max-width: 768px) {
                    flex-direction: column;
                    gap: 1rem;
                }
                
                .team-name-section {
                    flex: 1;
                    
                    .team-name-hero {
                        font-size: 2.5rem;
                        font-weight: 700;
                        margin-bottom: 0.5rem;
                        line-height: 1.2;
                        
                        @media (max-width: 768px) {
                            font-size: 2rem;
                        }
                    }
                    
                    .team-division-hero {
                        display: flex;
                        align-items: center;
                        gap: 0.5rem;
                        font-size: 1.1rem;
                        color: #cbd5e1;
                        
                        .player-count {
                            margin-left: 1rem;
                            padding-left: 1rem;
                            border-left: 1px solid #475569;
                        }
                    }
                }
                
                .team-rating-display {
                    display: grid;
                    grid-template-columns: repeat(4, 1fr);
                    gap: 1rem;
                    min-width: 300px;
                    
                    @media (max-width: 768px) {
                        grid-template-columns: repeat(2, 1fr);
                        min-width: auto;
                        width: 100%;
                    }
                    
                    .rating-item {
                        text-align: center;
                        padding: 1rem;
                        background: rgba(255, 255, 255, 0.1);
                        border-radius: $border-radius-small;
                        
                        .rating-label {
                            font-size: 0.9rem;
                            color: #cbd5e1;
                            margin-bottom: 0.5rem;
                        }
                        
                        .rating-value {
                            font-size: 1.5rem;
                            font-weight: 700;
                        }
                    }
                }
            }
        }
    }
}

// Teams Overview Card
.teams-overview-card {
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    background: white;
    
    .body--dark & {
        background: #1e293b;
    }
    
    .card-header {
        margin-bottom: 2rem;
        
        .card-title {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            font-size: 1.5rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: #1e293b;
            
            .body--dark & {
                color: #f1f5f9;
            }
            
            .card-icon {
                color: var(--accent);
            }
        }
        
        .card-subtitle {
            color: #64748b;
            margin: 0;
            
            .body--dark & {
                color: #94a3b8;
            }
        }
    }
    
    .modern-filter-section {
        margin-bottom: 2rem;
        
        .team-select {
            max-width: 400px;
        }
    }
    
    .teams-list {
                    .team-row {
                display: grid;
                grid-template-columns: auto 1fr auto auto;
                gap: 1rem;
                align-items: center;
                padding: 1rem;
                border-radius: $border-radius-small;
                cursor: pointer;
                transition: all 0.2s ease;
                border: 1px solid rgba(0, 0, 0, 0.06);
                margin-bottom: 0.5rem;
            
            .body--dark & {
                border-color: rgba(255, 255, 255, 0.1);
            }
            
            &:hover {
                transform: translateY(-1px);
                box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
                background: var(--accent-soft);
                
                .body--dark & {
                    background: var(--accent-soft);
                }
            }
            
            .team-logo-container {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 32px;
                
                .team-logo {
                    border-radius: 4px;
                    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
                }
            }
            
            .team-info {
                .team-name {
                    font-weight: 700;
                    font-size: 1rem;
                    color: #1e293b;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.9);
                    }
                }
                
                .team-division {
                    font-size: 0.9rem;
                    color: #64748b;
                    margin-bottom: 0.25rem;
                    
                    .body--dark & {
                        color: #94a3b8;
                    }
                }
                
                .player-count {
                    font-size: 0.85rem;
                    color: #64748b;
                    margin-top: 0.2rem;
                    
                    .body--dark & {
                        color: rgba(255, 255, 255, 0.6);
                    }
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
                                color: #d1d5db;
                                
                                .body--dark & {
                                    color: rgba(255, 255, 255, 0.3);
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    
    .show-more-btn {
        border-radius: $border-radius-small;
        font-weight: 500;
        text-transform: none;
    }
}

// Players Table Card
.players-table-card {
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    background: white;
    
    .body--dark & {
        background: #1e293b;
    }
    
    .card-header {
        margin-bottom: 1.5rem;
        
        .card-title {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: #1e293b;
            
            .body--dark & {
                color: #f1f5f9;
            }
            
            .card-icon {
                color: var(--accent);
            }
        }
        
        .card-subtitle {
            color: #64748b;
            margin: 0;
            
            .body--dark & {
                color: #94a3b8;
            }
        }
    }
    
    .table-container {
        .modern-table {
            border-radius: $border-radius-small;
            overflow: hidden;
        }
        
        .no-players-banner {
            text-align: center;
            padding: 3rem 2rem;
            color: #64748b;
            
            .body--dark & {
                color: #94a3b8;
            }
            
            h4 {
                margin: 1rem 0 0.5rem 0;
                color: #1e293b;
                
                .body--dark & {
                    color: #f1f5f9;
                }
            }
            
            p {
                margin: 0;
            }
        }
    }
}

// Rating Tier Classes
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
    color: #9ca3af !important;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.5) !important;
    }
}

// Responsive Design
@media (max-width: 768px) {
    .main-content {
        padding: 1rem;
    }
    
    .teams-list .team-row {
        grid-template-columns: auto 1fr auto;
        gap: 0.75rem;
        
        .team-section-ratings {
            display: none;
        }
    }
}
</style> 
