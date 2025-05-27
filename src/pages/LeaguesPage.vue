<template>
    <q-page padding class="leagues-page">
        <div class="q-pa-md">
            <h1
                class="text-h4 text-center q-mb-lg page-title"
                :class="
                    quasarInstance.dark.isActive ? 'text-grey-2' : 'text-grey-9'
                "
            >
                Leagues Analysis
            </h1>

            <q-banner
                v-if="pageLoadingError"
                class="text-white bg-negative q-mb-md"
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

            <q-card
                v-if="!pageLoadingError"
                class="q-mb-md filter-card"
                :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
            >
                <q-card-section>
                    <div class="row items-center justify-between q-mb-sm">
                        <div class="text-subtitle1">Select League</div>
                        <q-btn
                            v-if="currentDatasetId"
                            unelevated
                            icon="share"
                            label="Share Dataset"
                            color="positive"
                            @click="shareDataset"
                            class="share-btn-enhanced"
                            size="sm"
                        >
                            <q-tooltip>Copy shareable link to clipboard</q-tooltip>
                        </q-btn>
                    </div>
                    <q-select
                        v-model="selectedLeagueName"
                        :options="leagueOptions"
                        label="Search and Select League"
                        outlined
                        dense
                        use-input
                        hide-selected
                        fill-input
                        input-debounce="300"
                        @filter="filterLeagueOptions"
                        @update:model-value="loadLeagueTeams"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : 'bg-white text-dark'
                        "
                        clearable
                        @clear="clearLeagueSelection"
                        :disable="pageLoading || allLeaguesData.length === 0"
                    >
                        <template v-slot:no-option>
                            <q-item>
                                <q-item-section class="text-grey">
                                    No leagues found.
                                </q-item-section>
                            </q-item>
                        </template>
                    </q-select>
                </q-card-section>
            </q-card>

            <div v-if="pageLoading" class="text-center q-my-xl">
                <q-spinner-dots color="primary" size="3em" />
                <div
                    class="q-mt-md text-caption"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'text-grey-5'
                            : 'text-grey-7'
                    "
                >
                    Loading leagues data from server...
                </div>
            </div>
            <div v-else-if="loadingLeague" class="text-center q-my-xl">
                <q-spinner-dots color="primary" size="2em" />
                <div
                    class="q-mt-sm text-caption"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'text-grey-5'
                            : 'text-grey-7'
                    "
                >
                    Loading league teams...
                </div>
            </div>

            <div v-if="!pageLoading && !pageLoadingError">
                <!-- Leagues Overview Card -->
                <q-card
                    v-if="!selectedLeagueName && !loadingLeague && allLeaguesData.length > 0"
                    class="q-mb-md"
                    :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-md">Leagues Overview</div>
                        <div class="leagues-list">
                            <div
                                v-for="league in displayedLeagues"
                                :key="league.name"
                                class="league-row"
                                @click="selectLeague(league.name)"
                            >
                                <div class="league-info">
                                    <div class="league-name">{{ league.name }}</div>
                                    <div class="team-count">{{ league.teamCount }} teams</div>
                                    <div class="player-count">{{ league.playerCount }} players</div>
                                </div>
                                <div class="league-section-ratings">
                                    <div 
                                        v-if="league.bestOverall > 0"
                                        class="section-ratings-large"
                                    >
                                        <div class="section-rating-large att">
                                            <span class="section-label-large">ATT</span>
                                            <span 
                                                class="section-value-large"
                                                :class="getOverallClass(league.attRating)"
                                            >
                                                {{ league.attRating }}
                                            </span>
                                        </div>
                                        <div class="section-rating-large mid">
                                            <span class="section-label-large">MID</span>
                                            <span 
                                                class="section-value-large"
                                                :class="getOverallClass(league.midRating)"
                                            >
                                                {{ league.midRating }}
                                            </span>
                                        </div>
                                        <div class="section-rating-large def">
                                            <span class="section-label-large">DEF</span>
                                            <span 
                                                class="section-value-large"
                                                :class="getOverallClass(league.defRating)"
                                            >
                                                {{ league.defRating }}
                                            </span>
                                        </div>
                                    </div>
                                    <div 
                                        v-else
                                        class="no-rating-message"
                                    >
                                        Incomplete Teams
                                    </div>
                                </div>
                                <div class="league-overall">
                                    <div class="league-rating">
                                        <div 
                                            class="highest-overall-large"
                                            :class="getOverallClass(league.bestOverall)"
                                        >
                                            {{ league.bestOverall > 0 ? league.bestOverall + ' AVG' : 'N/A' }}
                                        </div>
                                        <div class="star-rating-large">
                                            <span
                                                v-for="star in 5"
                                                :key="star"
                                                class="star-large"
                                                :class="getStarClass(league.bestOverall, star)"
                                            >
                                                ★
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <!-- Show More Button -->
                        <div v-if="!showAllLeagues && allLeaguesData.length > INITIAL_LEAGUES_LIMIT" class="text-center q-mt-md">
                            <q-btn
                                flat
                                color="primary"
                                @click="showAllLeagues = true"
                                class="show-more-btn"
                            >
                                Show All Leagues
                                <q-icon name="expand_more" class="q-ml-sm" />
                            </q-btn>
                        </div>
                    </q-card-section>
                </q-card>

                <!-- Selected League Teams -->
                <div v-if="selectedLeagueName && !loadingLeague && leagueTeams.length > 0">
                    <q-card
                        :class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-9'
                                : 'bg-white'
                        "
                        class="q-mb-md"
                    >
                        <q-card-section>
                            <div class="text-h6 q-mb-md">
                                {{ selectedLeagueName }} - Teams ({{ leagueTeams.length }})
                            </div>
                            <div class="teams-list">
                                <div
                                    v-for="team in leagueTeams"
                                    :key="team.name"
                                    class="team-row"
                                    @click="handleTeamSelected(team)"
                                >
                                    <div class="team-info">
                                        <div class="team-name">{{ team.name }}</div>
                                        <div class="team-player-count">{{ team.playerCount }} players</div>
                                    </div>
                                    <div class="team-section-ratings">
                                        <div 
                                            v-if="team.bestOverall > 0"
                                            class="section-ratings-large"
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
                                        <div 
                                            v-else
                                            class="no-rating-message"
                                        >
                                            Incomplete Squad
                                        </div>
                                    </div>
                                    <div class="team-overall">
                                        <div class="team-rating">
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
                        </q-card-section>
                    </q-card>
                </div>

                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingLeague &&
                        allLeaguesData.length > 0 &&
                        !selectedLeagueName
                    "
                    class="text-center q-mt-lg"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'bg-blue-grey-8 text-blue-grey-2'
                            : 'bg-blue-1 text-primary'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="info" />
                    </template>
                    Select a league to view its teams and their ratings.
                </q-banner>
                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingLeague &&
                        allLeaguesData.length === 0 &&
                        !pageLoadingError
                    "
                    class="text-center q-mt-lg"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'bg-red-9 text-red-2'
                            : 'bg-red-1 text-negative'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="warning" />
                    </template>
                    No league data available. Please upload a player file with Division information on the main page first.
                    <q-btn
                        flat
                        color="primary"
                        label="Go to Upload Page"
                        @click="router.push('/')"
                        class="q-ml-md"
                    />
                </q-banner>
            </div>
        </div>
        <PlayerDetailDialog
            :player="playerForDetailView"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
            :currency-symbol="detectedCurrencySymbol"
        />
    </q-page>
</template>

<script>
import { ref, computed, onMounted, watch } from "vue";
import { useQuasar } from "quasar";
import { useRouter, useRoute } from "vue-router";
import { usePlayerStore } from "../stores/playerStore";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import { debounce } from "../utils/debounce";

export default {
    name: "LeaguesPage",
    components: { PlayerDetailDialog },
    setup() {
        const quasarInstance = useQuasar();
        const router = useRouter();
        const route = useRoute();
        const playerStore = usePlayerStore();

        const selectedLeagueName = ref(null);
        const leagueOptions = ref([]);
        const allLeagueNamesCache = ref([]);
        const allLeaguesData = ref([]);
        const leagueTeams = ref([]);
        const loadingLeague = ref(false);
        const pageLoading = ref(true);
        const pageLoadingError = ref("");
        
        // Pagination for leagues
        const showAllLeagues = ref(false);
        const INITIAL_LEAGUES_LIMIT = 30;
        
        // Computed properties from store
        const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol);
        const currentDatasetId = computed(() => playerStore.currentDatasetId);

        // Limited leagues for initial rendering
        const displayedLeagues = computed(() => {
            if (!allLeaguesData.value || allLeaguesData.value.length === 0) return [];
            
            if (!showAllLeagues.value && allLeaguesData.value.length > INITIAL_LEAGUES_LIMIT) {
                return allLeaguesData.value.slice(0, INITIAL_LEAGUES_LIMIT);
            }
            
            return allLeaguesData.value;
        });

        const playerForDetailView = ref(null);
        const showPlayerDetailDialog = ref(false);

        const fetchLeaguesAndCurrency = async (datasetId) => {
            pageLoading.value = true;
            pageLoadingError.value = "";
            try {
                const response = await fetch(`/api/leagues/${datasetId}`);
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                const leaguesData = await response.json();
                allLeaguesData.value = leaguesData || [];
                
                // Also fetch currency symbol from player store
                await playerStore.fetchPlayersByDatasetId(datasetId);
            } catch (err) {
                pageLoadingError.value = `Failed to load leagues data: ${err.message || "Unknown server error"}. Please try uploading again.`;
            } finally {
                pageLoading.value = false;
            }
        };

        const fetchTeamsForLeague = async (datasetId, leagueName) => {
            loadingLeague.value = true;
            try {
                const response = await fetch(`/api/teams/${datasetId}/${encodeURIComponent(leagueName)}`);
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                const teamsData = await response.json();
                leagueTeams.value = teamsData || [];
            } catch (err) {
                console.error("Error fetching teams data:", err);
                leagueTeams.value = [];
            } finally {
                loadingLeague.value = false;
            }
        };

        onMounted(async () => {
            const datasetIdFromQuery = route.query.datasetId;
            const datasetIdFromRoute = route.params.datasetId;
            const leagueFromQuery = route.query.league;
            let finalDatasetId =
                datasetIdFromRoute ||
                datasetIdFromQuery ||
                sessionStorage.getItem("currentDatasetId");

            if (finalDatasetId) {
                if (
                    datasetIdFromQuery &&
                    datasetIdFromQuery !==
                        sessionStorage.getItem("currentDatasetId")
                ) {
                    sessionStorage.setItem(
                        "currentDatasetId",
                        datasetIdFromQuery,
                    );
                } else if (
                    !datasetIdFromQuery &&
                    sessionStorage.getItem("currentDatasetId")
                ) {
                    router.replace({ query: { datasetId: finalDatasetId } });
                }
                await fetchLeaguesAndCurrency(finalDatasetId);
            } else {
                pageLoadingError.value =
                    "No player dataset ID found. Please upload a file on the main page.";
                pageLoading.value = false;
            }

            if (!pageLoadingError.value && allLeaguesData.value.length > 0) {
                populateLeagueFilterOptions();
                
                if (leagueFromQuery && leagueFromQuery.trim() !== '') {
                    selectedLeagueName.value = leagueFromQuery;
                    loadLeagueTeams();
                }
            }
        });

        const populateLeagueFilterOptions = () => {
            if (!allLeaguesData.value || allLeaguesData.value.length === 0) {
                allLeagueNamesCache.value = [];
                leagueOptions.value = [];
                return;
            }
            const leagueNames = allLeaguesData.value.map(league => league.name).sort();
            allLeagueNamesCache.value = leagueNames;
            leagueOptions.value = leagueNames;
        };

        const filterLeagueOptions = (val, update) => {
            if (val === "") {
                update(() => {
                    leagueOptions.value = allLeagueNamesCache.value;
                });
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                leagueOptions.value = allLeagueNamesCache.value.filter(
                    (league) => league.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        const selectLeague = (leagueName) => {
            selectedLeagueName.value = leagueName;
            loadLeagueTeams();
        };

        const loadLeagueTeams = async () => {
            if (!selectedLeagueName.value) {
                leagueTeams.value = [];
                return;
            }
            
            const datasetId = currentDatasetId.value || sessionStorage.getItem("currentDatasetId");
            if (datasetId) {
                await fetchTeamsForLeague(datasetId, selectedLeagueName.value);
            }
        };

        const clearLeagueSelection = () => {
            selectedLeagueName.value = null;
            leagueTeams.value = [];
        };

        const handleTeamSelected = (team) => {
            // Navigate to the team-view page with team filtering
            const datasetId = currentDatasetId.value || sessionStorage.getItem("currentDatasetId");
            if (datasetId) {
                router.push({
                    path: '/team-view',
                    query: { 
                        datasetId: datasetId,
                        team: team.name 
                    }
                });
            }
        };

        const getOverallClass = (overall) => {
            if (overall === null || overall === undefined || overall === 0) return "rating-na";
            const numericOverall = Number(overall);
            if (isNaN(numericOverall)) return "rating-na";

            if (numericOverall >= 90) return "rating-tier-6";
            if (numericOverall >= 80) return "rating-tier-5";
            if (numericOverall >= 70) return "rating-tier-4";
            if (numericOverall >= 55) return "rating-tier-3";
            if (numericOverall >= 40) return "rating-tier-2";
            return "rating-tier-1";
        };

        const getStarClass = (overall, starPosition) => {
            if (!overall || overall === 0) return "star-empty";
            
            const starRating = getStarRating(overall);
            
            if (starPosition <= Math.floor(starRating)) {
                return "star-full";
            } else if (starPosition === Math.floor(starRating) + 1 && starRating % 1 === 0.5) {
                return "star-half";
            } else {
                return "star-empty";
            }
        };

        const getStarRating = (overall) => {
            if (!overall || overall === 0) return 0;
            
            if (overall >= 85) return 5;
            if (overall >= 82) return 4.5;
            if (overall >= 78) return 4;
            if (overall >= 74) return 3.5;
            if (overall >= 70) return 3;
            if (overall >= 67) return 2.5;
            if (overall >= 64) return 2;
            if (overall >= 60) return 1.5;
            if (overall >= 55) return 1;
            if (overall >= 50) return 0.5;
            return 0;
        };

        const shareDataset = async () => {
            if (!currentDatasetId.value) return;
            
            const shareUrl = `${window.location.origin}/leagues/${currentDatasetId.value}`;
            
            try {
                await navigator.clipboard.writeText(shareUrl);
                quasarInstance.notify({
                    message: 'Link copied to clipboard!',
                    color: 'positive',
                    icon: 'check_circle',
                    position: 'top',
                    timeout: 2000
                });
            } catch (err) {
                const textArea = document.createElement('textarea');
                textArea.value = shareUrl;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
                
                quasarInstance.notify({
                    message: 'Link copied to clipboard!',
                    color: 'positive',
                    icon: 'check_circle',
                    position: 'top',
                    timeout: 2000
                });
            }
        };

        watch(
            () => route.query.datasetId,
            async (newId, oldId) => {
                if (newId && newId !== oldId) {
                    sessionStorage.setItem("currentDatasetId", newId);
                    await fetchLeaguesAndCurrency(newId);
                    clearLeagueSelection();
                    if (
                        !pageLoadingError.value &&
                        allLeaguesData.value.length > 0
                    ) {
                        populateLeagueFilterOptions();
                    }
                }
            },
        );

        watch(
            () => route.query.league,
            (newLeague) => {
                if (newLeague && newLeague.trim() !== '' && newLeague !== selectedLeagueName.value) {
                    selectedLeagueName.value = newLeague;
                    loadLeagueTeams();
                }
            },
        );

        return {
            allLeaguesData,
            selectedLeagueName,
            leagueOptions,
            filterLeagueOptions,
            loadLeagueTeams,
            clearLeagueSelection,
            selectLeague,
            leagueTeams,
            loadingLeague,
            pageLoading,
            pageLoadingError,
            playerForDetailView,
            showPlayerDetailDialog,
            handleTeamSelected,
            getOverallClass,
            getStarClass,
            getStarRating,
            quasarInstance,
            router,
            detectedCurrencySymbol,
            currentDatasetId,
            shareDataset,
            displayedLeagues,
            showAllLeagues,
            INITIAL_LEAGUES_LIMIT,
        };
    },
};
</script>

<style lang="scss" scoped>
.leagues-page {
    max-width: 1600px;
    margin: 0 auto;
}

.page-title {
    // Standard title styling
}

.filter-card,
.q-card {
    border-radius: $generic-border-radius;
}

.leagues-list,
.teams-list {
    max-height: 600px;
    overflow-y: auto;
}

.league-row,
.team-row {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-radius: 6px;
    border: 1px solid rgba(0, 0, 0, 0.1);
    margin-bottom: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }
    
    .body--dark & {
        background-color: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.1);
        
        &:hover {
            background-color: rgba(255, 255, 255, 0.08);
            box-shadow: 0 2px 8px rgba(255, 255, 255, 0.1);
        }
    }
    
    .body--light & {
        background-color: rgba(0, 0, 0, 0.02);
        border-color: rgba(0, 0, 0, 0.1);
        
        &:hover {
            background-color: rgba(0, 0, 0, 0.05);
        }
    }
}

.league-info,
.team-info {
    flex-shrink: 0;
    min-width: 150px;
}

.league-name,
.team-name {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 2px;
    
    .body--dark & {
        color: $grey-2;
    }
    
    .body--light & {
        color: $grey-8;
    }
}

.team-count,
.player-count,
.team-player-count {
    font-size: 0.85rem;
    color: $grey-6;
    
    .body--dark & {
        color: $grey-4;
    }
}

.league-section-ratings,
.team-section-ratings {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 0 16px;
}

.section-ratings-large {
    display: flex;
    gap: 20px;
    align-items: center;
}

.section-rating-large {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    
    .section-label-large {
        font-size: 0.8rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.8px;
    }
    
    .section-value-large {
        font-size: 1.1rem;
        font-weight: bold;
        padding: 4px 8px;
        border-radius: 6px;
        min-width: 32px;
        text-align: center;
        border: 1px solid rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            border-color: rgba(255, 255, 255, 0.1);
        }
    }
    
    &.att .section-label-large {
        color: #F44336; // Red for attack
        
        .body--dark & {
            color: #FF5722;
        }
    }
    
    &.mid .section-label-large {
        color: #2196F3; // Blue for midfield
        
        .body--dark & {
            color: #03A9F4;
        }
    }
    
    &.def .section-label-large {
        color: #4CAF50; // Green for defense
        
        .body--dark & {
            color: #8BC34A;
        }
    }
}

.no-rating-message {
    font-size: 0.9rem;
    color: $grey-6;
    font-style: italic;
    text-align: center;
    
    .body--dark & {
        color: $grey-5;
    }
}

.league-overall,
.team-overall {
    flex-shrink: 0;
    margin-left: 16px;
}

.league-rating,
.team-rating {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    min-width: 120px;
}

.highest-overall-large {
    font-weight: bold;
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 1rem;
    text-align: center;
    min-width: 60px;
}

.star-rating-large {
    display: flex;
    gap: 2px;
    font-size: 1.2rem;
    line-height: 1;
}

.star-large {
    transition: color 0.2s ease;
    
    &.star-full {
        color: #FFD700; // Gold
    }
    
    &.star-half {
        background: linear-gradient(90deg, #FFD700 50%, transparent 50%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        
        // Fallback for browsers that don't support background-clip: text
        @supports not (-webkit-background-clip: text) {
            color: #FFD700;
        }
    }
    
    &.star-empty {
        color: #E0E0E0;
        
        .body--dark & {
            color: #424242;
        }
    }
}

.share-btn-enhanced {
    font-weight: 600;
    border-radius: 6px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: all 0.2s ease;
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
    }
    
    .body--dark & {
        box-shadow: 0 2px 4px rgba(255, 255, 255, 0.1);
        
        &:hover {
            box-shadow: 0 4px 8px rgba(255, 255, 255, 0.15);
        }
    }
}

.show-more-btn {
    font-weight: 500;
    border-radius: 8px;
    padding: 8px 24px;
    transition: all 0.2s ease;
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            box-shadow: 0 2px 8px rgba(255, 255, 255, 0.1);
        }
    }
}
</style>