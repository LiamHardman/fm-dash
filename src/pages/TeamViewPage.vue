// src/pages/TeamViewPage.vue
<template>
    <q-page padding class="team-view-page">
        <div class="q-pa-md">
            <h1
                class="text-h4 text-center q-mb-lg page-title"
                :class="
                    quasarInstance.dark.isActive ? 'text-grey-2' : 'text-grey-9'
                "
            >
                Team Analysis
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
                    <div class="text-subtitle1 q-mb-sm">Select Team</div>
                    <q-select
                        v-model="selectedTeamName"
                        :options="teamOptions"
                        label="Search and Select Team"
                        outlined
                        dense
                        use-input
                        hide-selected
                        fill-input
                        input-debounce="300"
                        @filter="filterTeamOptions"
                        @update:model-value="loadTeamPlayers"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : 'bg-white text-dark'
                        "
                        clearable
                        @clear="clearTeamSelection"
                        :disable="pageLoading || allPlayersData.length === 0"
                    >
                        <template v-slot:no-option>
                            <q-item>
                                <q-item-section class="text-grey">
                                    No teams found.
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
                    Loading player data from server...
                </div>
            </div>
            <div v-else-if="loadingTeam" class="text-center q-my-xl">
                <q-spinner-dots color="primary" size="2em" />
                <div
                    class="q-mt-sm text-caption"
                    :class="
                        quasarInstance.dark.isActive
                            ? 'text-grey-5'
                            : 'text-grey-7'
                    "
                >
                    Loading team data...
                </div>
            </div>

            <div v-if="!pageLoading && !pageLoadingError">
                <div v-if="selectedTeamName && !loadingTeam">
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
                                Team Formation & Best XI
                            </div>
                            <div class="row q-col-gutter-md items-start">
                                <div class="col-12 col-md-4">
                                    <q-select
                                        v-model="selectedFormationKey"
                                        :options="formationOptions"
                                        label="Select Formation"
                                        outlined
                                        dense
                                        emit-value
                                        map-options
                                        :label-color="
                                            quasarInstance.dark.isActive
                                                ? 'grey-4'
                                                : ''
                                        "
                                        :popup-content-class="
                                            quasarInstance.dark.isActive
                                                ? 'bg-grey-8 text-white'
                                                : 'bg-white text-dark'
                                        "
                                    />
                                    <div
                                        v-if="bestTeamAverageOverall !== null"
                                        class="q-mt-md text-subtitle1"
                                        :class="
                                            quasarInstance.dark.isActive
                                                ? 'text-grey-3'
                                                : 'text-grey-8'
                                        "
                                    >
                                        Best XI Average Overall:
                                        <span
                                            class="text-weight-bold attribute-value"
                                            :class="
                                                getOverallClass(
                                                    bestTeamAverageOverall,
                                                )
                                            "
                                        >
                                            {{ bestTeamAverageOverall }}
                                        </span>
                                    </div>
                                    <q-banner
                                        v-if="calculationMessage"
                                        class="q-mt-sm"
                                        :class="calculationMessageClass"
                                    >
                                        {{ calculationMessage }}
                                    </q-banner>
                                    
                                    <!-- Compact Squad Depth -->
                                    <div 
                                        v-if="selectedFormationKey && Object.keys(squadComposition).length > 0"
                                        class="q-mt-md"
                                    >
                                        <div class="text-subtitle2 text-weight-bold q-mb-sm">Squad Depth</div>
                                        <div class="compact-squad-depth">
                                            <div
                                                v-for="slot in currentFormationLayout.flatMap(
                                                    (row) => row.positions,
                                                )"
                                                :key="slot.id"
                                                class="depth-position-compact"
                                            >
                                                <div class="position-label">
                                                    {{
                                                        getSlotDisplayName(
                                                            slot,
                                                            currentFormationLayout.flatMap(
                                                                (r) => r.positions,
                                                            ),
                                                        )
                                                    }}
                                                </div>
                                                <div 
                                                    v-if="squadComposition[slot.id] && squadComposition[slot.id].length > 0"
                                                    class="depth-players-compact"
                                                >
                                                    <div
                                                        v-for="(playerEntry, index) in squadComposition[slot.id].slice(0, 3)"
                                                        :key="playerEntry.player.name + '-' + slot.id + '-' + index"
                                                        class="depth-player-compact"
                                                        :class="{ 'starter': index === 0, 'backup': index > 0 }"
                                                        @click="handlePlayerSelectedFromTeam(playerEntry.player)"
                                                    >
                                                        <span class="player-rank-compact">{{ index + 1 }}.</span>
                                                        <span class="player-name-compact">{{ playerEntry.player.name }}</span>
                                                        <span 
                                                            class="overall-compact"
                                                            :class="getOverallClass(playerEntry.overallInRole)"
                                                        >
                                                            {{ playerEntry.overallInRole }}
                                                        </span>
                                                    </div>
                                                </div>
                                                <div 
                                                    v-else
                                                    class="no-players-compact"
                                                >
                                                    No players
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div class="col-12 col-md-8">
                                    <PitchDisplay
                                        :formation="currentFormationLayout"
                                        :players="bestTeamPlayersForPitch"
                                        @player-click="
                                            handlePlayerSelectedFromTeam
                                        "
                                        @player-moved="handlePlayerMovedOnPitch"
                                    />
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>

                    <q-card
                        class="q-mb-md"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-9'
                                : 'bg-white'
                        "
                    >
                        <q-card-section>
                            <div class="text-h6 q-mb-sm">
                                Players in {{ selectedTeamName }} ({{
                                    teamPlayers.length
                                }})
                            </div>
                            <PlayerDataTable
                                v-if="teamPlayers.length > 0"
                                :players="teamPlayers"
                                :loading="false"
                                @player-selected="handlePlayerSelectedFromTeam"
                                :is-goalkeeper-view="teamIsGoalkeeperView"
                                :currency-symbol="detectedCurrencySymbol"
                                table-style="max-height: 400px;"
                                class="team-player-table"
                            />
                            <q-banner
                                v-else
                                class="text-center"
                                :class="
                                    quasarInstance.dark.isActive
                                        ? 'bg-grey-8 text-grey-4'
                                        : 'bg-grey-2 text-grey-7'
                                "
                            >
                                No players found for this team with the current
                                data.
                            </q-banner>
                        </q-card-section>
                    </q-card>

                </div>
                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingTeam &&
                        allPlayersData.length > 0 &&
                        !selectedTeamName
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
                    Please select a team to view its players and analyze
                    formations.
                </q-banner>
                <q-banner
                    v-else-if="
                        !pageLoading &&
                        !loadingTeam &&
                        allPlayersData.length === 0 &&
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
                    No player data available. Please upload a player file on the
                    main page first.
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
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import PitchDisplay from "../components/PitchDisplay.vue";
import { formations, getFormationLayout } from "../utils/formations";
// Currency utils are not directly used here for formatting,
// but PlayerDataTable and PlayerDetailDialog will use them with the passed symbol.

const fmSlotRoleMatcher = {
    GK: ["Goalkeeper"],
    "D (R)": ["Defender (Right)", "Right Back"],
    "D (L)": ["Defender (Left)", "Left Back"],
    "D (C)": ["Defender (Centre)", "Centre Back"],
    "WB (R)": ["Wing-Back (Right)", "Right Wing-Back"],
    "WB (L)": ["Wing-Back (Left)", "Left Wing-Back"],
    "DM (C)": ["Defensive Midfielder (Centre)", "Centre Defensive Midfielder"],
    "M (R)": ["Midfielder (Right)", "Right Midfielder"],
    "M (L)": ["Midfielder (Left)", "Left Midfielder"],
    "M (C)": ["Midfielder (Centre)", "Centre Midfielder"],
    "AM (R)": [
        "Attacking Midfielder (Right)",
        "Right Attacking Midfielder",
        "Winger (Right)",
    ],
    "AM (L)": [
        "Attacking Midfielder (Left)",
        "Left Attacking Midfielder",
        "Winger (Left)",
    ],
    "AM (C)": ["Attacking Midfielder (Centre)", "Centre Attacking Midfielder"],
    "ST (C)": ["Striker (Centre)", "Striker"],
};

export default {
    name: "TeamViewPage",
    components: { PlayerDataTable, PlayerDetailDialog, PitchDisplay },
    setup() {
        const quasarInstance = useQuasar();
        const router = useRouter();
        const route = useRoute();
        const playerStore = usePlayerStore();

        const selectedTeamName = ref(null);
        const teamOptions = ref([]);
        const allTeamNamesCache = ref([]);
        const teamPlayers = ref([]);
        const loadingTeam = ref(false);
        const pageLoading = ref(true);
        const pageLoadingError = ref("");
        
        // Computed properties from store
        const allPlayersData = computed(() => playerStore.allPlayers);
        const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol);

        const selectedFormationKey = ref(null);

        const squadComposition = ref({});

        const bestTeamAverageOverall = ref(null);
        const calculationMessage = ref("");
        const calculationMessageClass = ref("");

        const playerForDetailView = ref(null);
        const showPlayerDetailDialog = ref(false);

        // Map position names to their short codes, more specific for each side
        const fmMatcherToRoleKeyPrefix = {
            GOALKEEPER: "GK",
            SWEEPER: "DC",
            "DEFENDER (RIGHT)": "DR",
            "RIGHT BACK": "DR",
            "DEFENDER (LEFT)": "DL",
            "LEFT BACK": "DL",
            "DEFENDER (CENTRE)": "DC",
            "CENTRE BACK": "DC",
            "WING-BACK (RIGHT)": "WBR",
            "RIGHT WING-BACK": "WBR",
            "WING-BACK (LEFT)": "WBL",
            "LEFT WING-BACK": "WBL",
            "DEFENSIVE MIDFIELDER (CENTRE)": "DM",
            "CENTRE DEFENSIVE MIDFIELDER": "DM",
            "MIDFIELDER (RIGHT)": "MR",
            "RIGHT MIDFIELDER": "MR",
            "MIDFIELDER (LEFT)": "ML",
            "LEFT MIDFIELDER": "ML",
            "MIDFIELDER (CENTRE)": "MC",
            "CENTRE MIDFIELDER": "MC",
            "ATTACKING MIDFIELDER (RIGHT)": "AMR",
            "RIGHT ATTACKING MIDFIELDER": "AMR",
            "WINGER (RIGHT)": "AMR",
            "ATTACKING MIDFIELDER (LEFT)": "AML",
            "LEFT ATTACKING MIDFIELDER": "AML",
            "WINGER (LEFT)": "AML",
            "ATTACKING MIDFIELDER (CENTRE)": "AMC",
            "CENTRE ATTACKING MIDFIELDER": "AMC",
            "STRIKER (CENTRE)": "ST",
            STRIKER: "ST",
        };
        
        // For handling combined positions like D/WB(R)
        // The first position is the PREFERRED position, others are fallbacks
        const positionSideMap = {
            // Map FM formation slots to possible shortPositions (in strict priority order)
            "D (R)": ["DR"],                     // Right defender should ONLY be DR
            "D (L)": ["DL"],                     // Left defender should ONLY be DL
            "D (C)": ["DC"],                     // Center defender should ONLY be DC
            "WB (R)": ["WBR"],                   // Right wing back should ONLY be WBR
            "WB (L)": ["WBL"],                   // Left wing back should ONLY be WBL
            "DM (C)": ["DM"],                    // Defensive mid should ONLY be DM
            "M (R)": ["MR"],                     // Right mid should ONLY be MR
            "M (L)": ["ML"],                     // Left mid should ONLY be ML
            "M (C)": ["MC"],                     // Center mid should ONLY be MC
            "AM (R)": ["AMR"],                   // Right attacking mid should ONLY be AMR
            "AM (L)": ["AML"],                   // Left attacking mid should ONLY be AML
            "AM (C)": ["AMC"],                   // Center attacking mid should ONLY be AMC
            "ST (C)": ["ST"],                    // Striker should ONLY be ST
            "GK": ["GK"]                         // Goalkeeper is always GK
        };
        
        // Secondary fallback map - only used if no players are found for a position
        const fallbackPositionMap = {
            "D (R)": ["DR", "WBR", "MR"],
            "D (L)": ["DL", "WBL", "ML"],
            "D (C)": ["DC", "DM"],
            "WB (R)": ["WBR", "DR", "MR"],
            "WB (L)": ["WBL", "DL", "ML"],
            "DM (C)": ["DM", "DC", "MC"],
            "M (R)": ["MR", "WBR", "AMR"],
            "M (L)": ["ML", "WBL", "AML"],
            "M (C)": ["MC", "DM"],
            "AM (R)": ["AMR", "MR"],
            "AM (L)": ["AML", "ML"],
            "AM (C)": ["AMC", "MC"],
            "ST (C)": ["ST", "AMC"],
            "GK": ["GK"]
        };

        const fetchPlayersAndCurrency = async (datasetId) => {
            pageLoading.value = true;
            pageLoadingError.value = "";
            try {
                await playerStore.fetchPlayersByDatasetId(datasetId);
                // The store now handles all data processing and storage
            } catch (err) {
                pageLoadingError.value = `Failed to load player data: ${err.message || "Unknown server error"}. Please try uploading again.`;
            } finally {
                pageLoading.value = false;
            }
        };

        onMounted(async () => {
            const datasetIdFromQuery = route.query.datasetId;
            const teamFromQuery = route.query.team;
            let finalDatasetId =
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
                    // If loading from session, ensure query param is updated for consistency/bookmarking
                    router.replace({ query: { datasetId: finalDatasetId } });
                }
                await fetchPlayersAndCurrency(finalDatasetId);
            } else {
                pageLoadingError.value =
                    "No player dataset ID found. Please upload a file on the main page.";
                pageLoading.value = false;
            }

            if (!pageLoadingError.value && allPlayersData.value.length > 0) {
                populateTeamFilterOptions();
                
                // If a team was specified in the query params, select it
                if (teamFromQuery && teamFromQuery.trim() !== '') {
                    selectedTeamName.value = teamFromQuery;
                    loadTeamPlayers();
                } else if (selectedTeamName.value) {
                    // If a team was previously selected (e.g. from state restoration)
                    loadTeamPlayers();
                }
            }
        });

        const populateTeamFilterOptions = () => {
            if (!allPlayersData.value || allPlayersData.value.length === 0) {
                allTeamNamesCache.value = [];
                teamOptions.value = [];
                return;
            }
            const uniqueTeams = new Set();
            allPlayersData.value.forEach((player) => {
                if (player.club && player.club.trim() !== "") {
                    uniqueTeams.add(player.club);
                }
            });
            allTeamNamesCache.value = Array.from(uniqueTeams).sort();
            teamOptions.value = allTeamNamesCache.value;
        };

        const filterTeamOptions = (val, update) => {
            if (val === "") {
                update(() => {
                    teamOptions.value = allTeamNamesCache.value;
                });
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                teamOptions.value = allTeamNamesCache.value.filter(
                    (team) => team.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        const loadTeamPlayers = () => {
            if (!selectedTeamName.value) {
                teamPlayers.value = [];
                squadComposition.value = {};
                bestTeamAverageOverall.value = null;
                calculationMessage.value = "";
                selectedFormationKey.value = null;
                return;
            }
            loadingTeam.value = true;
            setTimeout(() => {
                // Simulate async operation if needed
                if (Array.isArray(allPlayersData.value)) {
                    teamPlayers.value = allPlayersData.value.filter(
                        (p) => p.club === selectedTeamName.value,
                    );
                } else {
                    teamPlayers.value = [];
                }
                
                // Auto-select the best formation for this team
                if (teamPlayers.value.length > 0) {
                    const bestFormation = calculateBestFormationForTeam();
                    if (bestFormation) {
                        selectedFormationKey.value = bestFormation;
                        calculationMessage.value = `Auto-selected best formation: ${formations[bestFormation].name}. Calculating Best XI...`;
                        calculationMessageClass.value = quasarInstance.dark.isActive
                            ? "bg-info text-white"
                            : "bg-blue-2 text-primary";
                    } else {
                        selectedFormationKey.value = null;
                        squadComposition.value = {};
                        bestTeamAverageOverall.value = null;
                        calculationMessage.value = "No suitable formation found for this team.";
                        calculationMessageClass.value = quasarInstance.dark.isActive
                            ? "text-grey-5"
                            : "text-grey-7";
                    }
                } else {
                    selectedFormationKey.value = null;
                    squadComposition.value = {};
                    bestTeamAverageOverall.value = null;
                    calculationMessage.value = "No players found for this team.";
                    calculationMessageClass.value = quasarInstance.dark.isActive
                        ? "text-grey-5"
                        : "text-grey-7";
                }
                
                loadingTeam.value = false;
            }, 200);
        };

        const clearTeamSelection = () => {
            selectedTeamName.value = null;
            teamPlayers.value = [];
            selectedFormationKey.value = null;
            squadComposition.value = {};
            bestTeamAverageOverall.value = null;
            calculationMessage.value = "";
        };

        const formationOptions = computed(() => {
            return Object.keys(formations).map((key) => ({
                label: formations[key].name,
                value: key,
            }));
        });

        const currentFormationLayout = computed(() => {
            if (!selectedFormationKey.value) {
                return [];
            }
            return getFormationLayout(selectedFormationKey.value) || [];
        });

        const bestTeamPlayersForPitch = computed(() => {
            const starters = {};
            if (
                !squadComposition.value ||
                Object.keys(squadComposition.value).length === 0
            ) {
                return starters;
            }
            for (const slotId in squadComposition.value) {
                if (
                    squadComposition.value[slotId] &&
                    squadComposition.value[slotId].length > 0
                ) {
                    const starterEntry = squadComposition.value[slotId][0];
                    // Use the role-specific score for this position, not their global Overall
                    // Add the exactMatch flag to display if the player is in their natural position
                    starters[slotId] = {
                        ...starterEntry.player,                    // Spread all player properties
                        Overall: starterEntry.overallInRole,       // Use position-specific rating
                        exactPositionMatch: starterEntry.exactMatch // Pass this to the pitch display
                    };
                } else {
                    starters[slotId] = null; // No player for this slot
                }
            }
            return starters;
        });

        const teamIsGoalkeeperView = computed(() => {
            // This computed property is for the PlayerDataTable on this page.
            // It should reflect if the *selected team* has goalkeepers,
            // rather than a global filter.
            return teamPlayers.value.some((p) =>
                p.positionGroups?.includes("Goalkeepers"),
            );
        });

        const handlePlayerSelectedFromTeam = (player) => {
            playerForDetailView.value = player;
            showPlayerDetailDialog.value = true;
        };

        const getOverallClass = (overall) => {
            if (overall === null || overall === undefined) return "rating-na";
            const numericOverall = Number(overall);
            if (isNaN(numericOverall)) return "rating-na";

            if (numericOverall >= 90) return "rating-tier-6";
            if (numericOverall >= 80) return "rating-tier-5";
            if (numericOverall >= 70) return "rating-tier-4";
            if (numericOverall >= 55) return "rating-tier-3";
            if (numericOverall >= 40) return "rating-tier-2";
            return "rating-tier-1";
        };

        const getPlayerOverallForRole = (player, slotFormationRole) => {
            if (!player || !slotFormationRole) return 0;

            let bestScoreForRole = 0;
            let matchType = "none"; // For debugging: tracks how the match was found
            
            if (!player.roleSpecificOveralls) {
                return 0; // No role overalls available
            }
            
            // Check if roleSpecificOveralls exists in either array or object format
            const hasRoleOveralls = Array.isArray(player.roleSpecificOveralls) 
                ? player.roleSpecificOveralls.length > 0
                : Object.keys(player.roleSpecificOveralls).length > 0;
            
            if (!hasRoleOveralls) {
                return 0; // No role overalls available
            }
            
            // Get the required positions for this slot (strict matching)
            const upperSlotRoleOriginal = slotFormationRole.toUpperCase();
            const requiredPositions = positionSideMap[upperSlotRoleOriginal] || [];
            
            // 1. STRICT MATCHING: Player must have the EXACT position to play here
            if (player.shortPositions && player.shortPositions.length > 0) {
                // Check if player has ANY of the required positions
                const exactPositionMatches = player.shortPositions.filter(pos => 
                    requiredPositions.includes(pos)
                );
                
                if (exactPositionMatches.length > 0) {
                    // Perfect position match! Find the best role score
                    matchType = "exact";
                    
                    // Find best score from roleSpecificOveralls - handle both array and object formats
                    if (Array.isArray(player.roleSpecificOveralls)) {
                        player.roleSpecificOveralls.forEach(rso => {
                            const rsoBasePosition = rso.roleName
                                .split(" - ")[0] // "DC" from "DC - BPD"
                                .trim();
                            
                            // Check if this role's position is one of the player's exact positions
                            if (exactPositionMatches.includes(rsoBasePosition)) {
                                bestScoreForRole = Math.max(
                                    bestScoreForRole,
                                    rso.score,
                                );
                            }
                        });
                    } else {
                        // Object format
                        Object.entries(player.roleSpecificOveralls).forEach(([roleName, score]) => {
                            const rsoBasePosition = roleName
                                .split(" - ")[0] // "DC" from "DC - BPD"
                                .trim();
                            
                            // Check if this role's position is one of the player's exact positions
                            if (exactPositionMatches.includes(rsoBasePosition)) {
                                bestScoreForRole = Math.max(
                                    bestScoreForRole,
                                    score,
                                );
                            }
                        });
                    }
                    
                    // If we have an exact position match but no specific role score,
                    // give them a baseline score
                    if (bestScoreForRole === 0) {
                        bestScoreForRole = MIN_SUITABILITY_THRESHOLD;
                    }
                    
                    // Add a small preference boost just for sorting purposes
                    // (we'll store the original score in a separate property)
                }
            }
            
            // Skip fallbacks if we found an exact match
            if (bestScoreForRole > 0) {
                // For debugging
                //console.log(`Exact match for ${player.name} in ${slotFormationRole}: score=${bestScoreForRole}`);
                return bestScoreForRole;
            }
            
            // 2. FALLBACK MATCHING: If no exact match, try fallback positions
            const fallbackPositions = fallbackPositionMap[upperSlotRoleOriginal] || [];
            
            if (player.shortPositions && player.shortPositions.length > 0) {
                // Check if player has ANY of the fallback positions
                const fallbackMatches = player.shortPositions.filter(pos => 
                    fallbackPositions.includes(pos)
                );
                
                if (fallbackMatches.length > 0) {
                    // Fallback position match - these will be scored lower
                    matchType = "fallback";
                    
                    // Find best score from roleSpecificOveralls with fallback positions
                    if (Array.isArray(player.roleSpecificOveralls)) {
                        player.roleSpecificOveralls.forEach(rso => {
                            const rsoBasePosition = rso.roleName
                                .split(" - ")[0] // "DC" from "DC - BPD"
                                .trim();
                            
                            if (fallbackMatches.includes(rsoBasePosition)) {
                                bestScoreForRole = Math.max(
                                    bestScoreForRole,
                                    rso.score,
                                );
                            }
                        });
                    } else {
                        // Object format
                        Object.entries(player.roleSpecificOveralls).forEach(([roleName, score]) => {
                            const rsoBasePosition = roleName
                                .split(" - ")[0] // "DC" from "DC - BPD"
                                .trim();
                            
                            if (fallbackMatches.includes(rsoBasePosition)) {
                                bestScoreForRole = Math.max(
                                    bestScoreForRole,
                                    score,
                                );
                            }
                        });
                    }
                    
                    // If we have a fallback position match but no specific role score,
                    // give them a minimal score
                    if (bestScoreForRole === 0) {
                        bestScoreForRole = MIN_SUITABILITY_THRESHOLD - 10; // Lower threshold for fallbacks
                    }
                    
                    // Note: Original score is preserved, we'll just use the exactMatch flag for sorting
                }
            }
            
            // 3. LAST RESORT: If still no match, use the old FM matcher approach
            if (bestScoreForRole === 0) {
                const upperSlotRole = slotFormationRole.toUpperCase();
                const fmPositionMatchers = fmSlotRoleMatcher[upperSlotRole] || [upperSlotRole];
                
                // Convert detailed positions to base role key prefixes
                const targetRoleKeyPrefixes = fmPositionMatchers
                    .map(matcher => fmMatcherToRoleKeyPrefix[matcher.toUpperCase()])
                    .filter(prefix => !!prefix)
                    .reduce((acc, val) => (acc.includes(val) ? acc : [...acc, val]), []);
                
                // Check roleSpecificOveralls against these prefixes
                if (Array.isArray(player.roleSpecificOveralls)) {
                    player.roleSpecificOveralls.forEach(rso => {
                        const rsoBasePosition = rso.roleName
                            .split(" - ")[0] // "DC" from "DC - BPD"
                            .trim();
                        
                        if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
                            matchType = "legacy";
                            bestScoreForRole = Math.max(
                                bestScoreForRole,
                                rso.score,
                            );
                        }
                    });
                } else if (player.roleSpecificOveralls) {
                    Object.entries(player.roleSpecificOveralls).forEach(([roleName, score]) => {
                        const rsoBasePosition = roleName
                            .split(" - ")[0]
                            .trim();
                        
                        if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
                            matchType = "legacy";
                            bestScoreForRole = Math.max(
                                bestScoreForRole,
                                score,
                            );
                        }
                    });
                }
                
                // Legacy matches will be sorted last by using the exactMatch flag
            }
            
            // For debugging
            //if (bestScoreForRole > 0) {
            //    console.log(`${matchType} match for ${player.name} in ${slotFormationRole}: score=${bestScoreForRole}`);
            //}
            
            return bestScoreForRole;
        };

        const MIN_SUITABILITY_THRESHOLD = 40;

        const getSlotDisplayName = (slot, allSlots) => {
            const roleCounts = allSlots.reduce((acc, s) => {
                acc[s.role] = (acc[s.role] || 0) + 1;
                return acc;
            }, {});

            if (roleCounts[slot.role] > 1) {
                // If multiple slots have the same base role (e.g., two "ST (C)"),
                // use the more specific ID (like "STCL", "STCR").
                // Extract the prefix from ID, e.g., "STCL" from "STCL_41212N"
                return slot.id.split("_")[0];
            }
            return slot.role; // Otherwise, use the general role name like "AM (C)"
        };

        const calculateBestFormationForTeam = () => {
            if (teamPlayers.value.length === 0) {
                return null;
            }

            let bestFormationKey = null;
            let bestAverageOverall = 0;

            // Test each formation to find the one with highest average overall
            Object.keys(formations).forEach(formationKey => {
                const formationLayoutForCalc = getFormationLayout(formationKey);
                if (!formationLayoutForCalc) return;

                const formationSlots = formationLayoutForCalc.flatMap(row => row.positions);
                const tempSquadComposition = {};
                
                // Initialize slots
                formationSlots.forEach(slot => {
                    tempSquadComposition[slot.id] = [];
                });

                // Calculate player scores for each position in this formation
                const allPotentialPlayerAssignments = [];
                formationSlots.forEach(slot => {
                    teamPlayers.value.forEach(player => {
                        const overallInRole = getPlayerOverallForRole(player, slot.role);
                        
                        if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
                            const slotPositions = positionSideMap[slot.role.toUpperCase()] || [];
                            const playerPositions = player.shortPositions || [];
                            const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos));
                            
                            if (isExactMatch || overallInRole >= MIN_SUITABILITY_THRESHOLD) {
                                const assignment = {
                                    player,
                                    slotId: slot.id,
                                    slotRole: slot.role,
                                    overallInRole: overallInRole,
                                    sortScore: overallInRole,
                                    exactMatch: isExactMatch
                                };
                                
                                if (isExactMatch) {
                                    assignment.sortScore += 10000;
                                } else {
                                    assignment.sortScore -= 5000;
                                }
                                
                                allPotentialPlayerAssignments.push(assignment);
                            }
                        }
                    });
                });

                // Sort assignments by sort score
                allPotentialPlayerAssignments.sort((a, b) => b.sortScore - a.sortScore);

                const assignedPlayersToSlots = new Set();

                // Fill starting XI for this formation
                formationSlots.forEach(slot => {
                    for (const assignment of allPotentialPlayerAssignments) {
                        if (
                            assignment.slotId === slot.id &&
                            !assignedPlayersToSlots.has(assignment.player.name)
                        ) {
                            tempSquadComposition[slot.id].push({
                                player: assignment.player,
                                overallInRole: assignment.overallInRole,
                                exactMatch: assignment.exactMatch
                            });
                            assignedPlayersToSlots.add(assignment.player.name);
                            break;
                        }
                    }
                });

                // Calculate average overall for this formation
                let sumOfStartersOverall = 0;
                let startersCount = 0;
                Object.values(tempSquadComposition).forEach(slotPlayers => {
                    if (slotPlayers && slotPlayers.length > 0) {
                        sumOfStartersOverall += slotPlayers[0].overallInRole;
                        startersCount++;
                    }
                });

                if (startersCount > 0) {
                    const averageOverall = sumOfStartersOverall / startersCount;
                    if (averageOverall > bestAverageOverall) {
                        bestAverageOverall = averageOverall;
                        bestFormationKey = formationKey;
                    }
                }
            });

            return bestFormationKey;
        };

        const calculateBestTeamAndDepth = () => {
            if (!selectedFormationKey.value || teamPlayers.value.length === 0) {
                squadComposition.value = {};
                bestTeamAverageOverall.value = null;
                calculationMessage.value = selectedFormationKey.value
                    ? "No players in the selected team."
                    : "Select a formation.";
                calculationMessageClass.value = "bg-warning text-dark";
                return;
            }

            calculationMessage.value = "Calculating best team and depth...";
            calculationMessageClass.value = quasarInstance.dark.isActive
                ? "bg-info text-white"
                : "bg-blue-2 text-primary";

            const tempSquadComposition = {};
            const formationLayoutForCalc = getFormationLayout(
                selectedFormationKey.value,
            );
            if (!formationLayoutForCalc) {
                calculationMessage.value = "Invalid formation selected.";
                calculationMessageClass.value = "bg-negative text-white";
                return;
            }

            const formationSlots = formationLayoutForCalc.flatMap(
                (row) => row.positions,
            );

            formationSlots.forEach((slot) => {
                tempSquadComposition[slot.id] = [];
            });

            // ENHANCEMENT: First, compute all player scores for all positions
            // and check which players can play in which positions
            const playerPositionMap = new Map(); // Maps player name to positions they can play
            
            teamPlayers.value.forEach(player => {
                const playablePositions = [];
                if (player.shortPositions && player.shortPositions.length > 0) {
                    playablePositions.push(...player.shortPositions);
                }
                playerPositionMap.set(player.name, playablePositions);
                
                // Debug: Log player positions
                //console.log(`${player.name} positions: ${playablePositions.join(', ')}`);
            });

            // Calculate player scores for each position
            const allPotentialPlayerAssignments = [];
            formationSlots.forEach((slot) => {
                teamPlayers.value.forEach((player) => {
                    const overallInRole = getPlayerOverallForRole(
                        player,
                        slot.role, // Use the general role from formation (e.g., "ST (C)")
                    );
                    
                    // Only include players who meet the threshold and are properly positioned
                    if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
                        // Get the compatible positions for this slot
                        const slotPositions = positionSideMap[slot.role.toUpperCase()] || [];
                        const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || [];
                        
                        // STRICT POSITION CHECKING: Check if player can play in this position
                        // For this to be true, the player MUST have one of the required positions 
                        // in their shortPositions array
                        
                        const playerPositions = playerPositionMap.get(player.name) || [];
                        
                        // For first XI and depth chart, we ONLY want players who can ACTUALLY play the position
                        // isExactMatch means player has the EXACT position for this slot
                        const isExactMatch = playerPositions.some(pos => 
                            slotPositions.includes(pos)
                        );
                        
                        // We won't use fallback positions at all for squad depth chart
                        // This ensures only properly positioned players are shown
                        const canPlayInPosition = isExactMatch;
                        
                        // Only add if player can actually play this position and meets minimum quality
                        if (canPlayInPosition && overallInRole >= MIN_SUITABILITY_THRESHOLD) {
                            // Strict position filtering: 
                            // 1. For first team selection, we want EXACT position matches only unless
                            //    there are no players for a position
                            // 2. For depth, we can be more flexible
                            
                            // Store the original role score and position match info
                            const assignment = {
                                player,
                                slotId: slot.id,
                                slotRole: slot.role,
                                overallInRole: overallInRole,  // Store original score for display
                                sortScore: overallInRole,      // Will be used for sorting
                                exactMatch: isExactMatch       // Flag for UI display
                            };
                            
                            // Adjust sort score (but not display score) based on position match
                            if (isExactMatch) {
                                // Huge boost to ensure exact matches are picked first
                                assignment.sortScore += 10000; 
                            } else {
                                // Penalty for out-of-position players
                                // They'll only be selected if no exact matches are available
                                assignment.sortScore -= 5000; 
                            }
                            
                            allPotentialPlayerAssignments.push(assignment);
                        }
                    }
                });
            });

            // Sort assignments by the sort score, which already includes position match bonus
            allPotentialPlayerAssignments.sort((a, b) => {
                return b.sortScore - a.sortScore;
            });

            const assignedPlayersToSlots = new Set();

            for (let depthIndex = 0; depthIndex < 3; depthIndex++) {
                // First pass: fill positions with exact matches
                formationSlots.forEach((slot) => {
                    if (tempSquadComposition[slot.id].length === depthIndex) {
                        // If this slot needs a player at current depth
                        for (const assignment of allPotentialPlayerAssignments) {
                            if (
                                assignment.slotId === slot.id &&
                                assignment.exactMatch && // Only use exact matches in first pass
                                !assignedPlayersToSlots.has(
                                    assignment.player.name,
                                )
                            ) {
                                // Check if this player is already a starter in *another* slot if we are filling backups
                                let alreadyStarterElsewhere = false;
                                if (depthIndex > 0) {
                                    // Only check for backups
                                    for (const sId in tempSquadComposition) {
                                        if (
                                            tempSquadComposition[sId].length >
                                                0 &&
                                            tempSquadComposition[sId][0].player
                                                .name === assignment.player.name
                                        ) {
                                            alreadyStarterElsewhere = true;
                                            break;
                                        }
                                    }
                                }

                                if (!alreadyStarterElsewhere) {
                                    tempSquadComposition[slot.id].push({
                                        player: assignment.player,
                                        overallInRole: assignment.overallInRole,
                                        exactMatch: assignment.exactMatch
                                    });
                                    assignedPlayersToSlots.add(
                                        assignment.player.name,
                                    );
                                    break; // Move to next slot for this depth level
                                }
                            }
                        }
                    }
                });
                
                // Second pass: fill remaining positions with fallback matches
                formationSlots.forEach((slot) => {
                    if (tempSquadComposition[slot.id].length === depthIndex) {
                        // If this slot still needs a player after the first pass
                        for (const assignment of allPotentialPlayerAssignments) {
                            if (
                                assignment.slotId === slot.id &&
                                !assignedPlayersToSlots.has(
                                    assignment.player.name,
                                )
                            ) {
                                // Check if this player is already a starter in *another* slot if we are filling backups
                                let alreadyStarterElsewhere = false;
                                if (depthIndex > 0) {
                                    // Only check for backups
                                    for (const sId in tempSquadComposition) {
                                        if (
                                            tempSquadComposition[sId].length >
                                                0 &&
                                            tempSquadComposition[sId][0].player
                                                .name === assignment.player.name
                                        ) {
                                            alreadyStarterElsewhere = true;
                                            break;
                                        }
                                    }
                                }

                                if (!alreadyStarterElsewhere) {
                                    tempSquadComposition[slot.id].push({
                                        player: assignment.player,
                                        overallInRole: assignment.overallInRole,
                                        exactMatch: assignment.exactMatch
                                    });
                                    assignedPlayersToSlots.add(
                                        assignment.player.name,
                                    );
                                    break; // Move to next slot for this depth level
                                }
                            }
                        }
                    }
                });
            }

            // Ensure each slot in tempSquadComposition is sorted by overallInRole descending
            for (const slotId in tempSquadComposition) {
                tempSquadComposition[slotId].sort(
                    (a, b) => b.overallInRole - a.overallInRole,
                );
            }
            
            // Check if any positions have no players assigned at all
            // In that case, try to find any player who can play there as a fallback
            for (const slot of formationSlots) {
                if (tempSquadComposition[slot.id].length === 0) {
                    console.log(`No exact position matches found for ${slot.role}, trying fallbacks`);
                    
                    // Get fallback positions for this slot
                    const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || [];
                    
                    // Find any players who can play in fallback positions
                    const fallbackAssignments = [];
                    
                    teamPlayers.value.forEach(player => {
                        if (!assignedPlayersToSlots.has(player.name)) {
                            const playerPositions = player.shortPositions || [];
                            
                            // Check if player can play any fallback position
                            const canPlayFallback = playerPositions.some(pos => 
                                fallbackPositions.includes(pos)
                            );
                            
                            if (canPlayFallback) {
                                const overallInRole = getPlayerOverallForRole(player, slot.role);
                                if (overallInRole >= MIN_SUITABILITY_THRESHOLD - 10) {
                                    fallbackAssignments.push({
                                        player,
                                        overallInRole,
                                        exactMatch: false
                                    });
                                }
                            }
                        }
                    });
                    
                    // Sort fallbacks by score
                    fallbackAssignments.sort((a, b) => b.overallInRole - a.overallInRole);
                    
                    // Add best fallback if available
                    if (fallbackAssignments.length > 0) {
                        const bestFallback = fallbackAssignments[0];
                        tempSquadComposition[slot.id].push(bestFallback);
                        assignedPlayersToSlots.add(bestFallback.player.name);
                    }
                }
            }

            squadComposition.value = tempSquadComposition;

            let sumOfStartersOverall = 0;
            let startersCount = 0;
            Object.values(squadComposition.value).forEach((slotPlayers) => {
                if (slotPlayers && slotPlayers.length > 0) {
                    sumOfStartersOverall += slotPlayers[0].overallInRole;
                    startersCount++;
                }
            });

            if (startersCount > 0) {
                bestTeamAverageOverall.value = Math.round(
                    sumOfStartersOverall / startersCount,
                );
                calculationMessage.value = `Best XI & Depth calculated. Average Overall: ${bestTeamAverageOverall.value}.`;
                calculationMessageClass.value = quasarInstance.dark.isActive
                    ? "bg-positive text-white"
                    : "bg-green-2 text-positive";
            } else {
                bestTeamAverageOverall.value = 0;
                calculationMessage.value =
                    "Could not assign any suitable players to form a Best XI.";
                calculationMessageClass.value = quasarInstance.dark.isActive
                    ? "bg-negative text-white"
                    : "bg-red-2 text-negative";
            }
        };

        watch(selectedFormationKey, (newKey) => {
            if (newKey && selectedTeamName.value) {
                calculateBestTeamAndDepth();
            } else {
                squadComposition.value = {};
                bestTeamAverageOverall.value = null;
                calculationMessage.value = "Select a team and formation.";
                calculationMessageClass.value = quasarInstance.dark.isActive
                    ? "text-grey-5"
                    : "text-grey-7";
            }
        });

        const handlePlayerMovedOnPitch = (moveData) => {
            const { player, fromSlotId, toSlotId, toSlotRole } = moveData;

            const currentStarters = JSON.parse(
                JSON.stringify(bestTeamPlayersForPitch.value),
            );
            const playerToMoveFullData = allPlayersData.value.find(
                (p) => p.name === player.name,
            );

            if (!playerToMoveFullData) return;

            // Calculate the role-specific rating for this player in the new position
            const overallInNewRole = getPlayerOverallForRole(
                playerToMoveFullData,
                toSlotRole,
            );
            
            // Check if player is in their natural position in the new slot
            const playerPositions = playerToMoveFullData.shortPositions || [];
            const slotPositions = positionSideMap[toSlotRole.toUpperCase()] || [];
            const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos));
            
            const playerCurrentlyInTargetSlotFullData = currentStarters[
                toSlotId
            ]
                ? allPlayersData.value.find(
                      (p) => p.name === currentStarters[toSlotId].name,
                  )
                : null;

            // Update target slot with role-specific rating and position match info
            currentStarters[toSlotId] = {
                ...playerToMoveFullData,
                Overall: overallInNewRole,           // Role-specific rating for the position
                exactPositionMatch: isExactMatch     // Position match flag for UI
            };

            // Update original slot
            if (playerCurrentlyInTargetSlotFullData && fromSlotId) {
                const originalRoleOfFromSlot = currentFormationLayout.value
                    .flatMap((r) => r.positions)
                    .find((p) => p.id === fromSlotId)?.role;
                    
                if (originalRoleOfFromSlot) {
                    // Calculate role-specific rating for the player in the original slot
                    const overallInOldRole = getPlayerOverallForRole(
                        playerCurrentlyInTargetSlotFullData,
                        originalRoleOfFromSlot,
                    );
                    
                    // Check if player is in their natural position in the original slot
                    const playerPositions = playerCurrentlyInTargetSlotFullData.shortPositions || [];
                    const slotPositions = positionSideMap[originalRoleOfFromSlot.toUpperCase()] || [];
                    const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos));
                    
                    currentStarters[fromSlotId] = {
                        ...playerCurrentlyInTargetSlotFullData,
                        Overall: overallInOldRole,        // Role-specific rating
                        exactPositionMatch: isExactMatch  // Position match flag
                    };
                } else {
                    currentStarters[fromSlotId] = null;
                }
            } else if (fromSlotId) {
                currentStarters[fromSlotId] = null;
            }

            // To make PitchDisplay update, we need to change the object reference
            // or ensure its internal properties are reactive.
            // This is a simplified visual swap; it doesn't formally update squadComposition.
            // For a temporary visual update of the pitch:
            const newPitchState = { ...currentStarters };
            // This assignment might not be enough if PitchDisplay relies on squadComposition.
            // A better way would be to have a local ref for pitch display players.
            // For now, we'll log and message that depth isn't updated.
            console.log("Visual swap on pitch:", newPitchState);

            let sumOfDisplayedOveralls = 0;
            let countOfDisplayedOveralls = 0;
            Object.values(newPitchState).forEach((p) => {
                if (p && typeof p.Overall === "number") {
                    // p.Overall is now the position-specific rating
                    sumOfDisplayedOveralls += p.Overall;
                    countOfDisplayedOveralls++;
                }
            });
            bestTeamAverageOverall.value =
                countOfDisplayedOveralls > 0
                    ? Math.round(
                          sumOfDisplayedOveralls / countOfDisplayedOveralls,
                      )
                    : 0;

            calculationMessage.value = `Team visually adjusted. New Avg Overall: ${bestTeamAverageOverall.value}. (Depth chart not updated by drag & drop).`;
            calculationMessageClass.value = quasarInstance.dark.isActive
                ? "bg-info text-white"
                : "bg-blue-2 text-primary";

            // To actually make PitchDisplay update from this drag-drop,
            // bestTeamPlayersForPitch would need to be made writable or a separate ref used.
            // For now, this is a visual indication of the swap's effect on average overall.
            // The actual `bestTeamPlayersForPitch` computed will still be based on `squadComposition`.
            // To truly reflect the drag-drop, `squadComposition` itself would need to be modified.
        };

        watch(
            () => allPlayersData.value,
            (newVal) => {
                if (pageLoading.value) return; // Don't run if initial load is happening
                if (newVal && newVal.length > 0) {
                    populateTeamFilterOptions();
                    if (selectedTeamName.value) loadTeamPlayers(); // Reload team if already selected
                } else if (!pageLoadingError.value) {
                    // Only clear if no error
                    clearTeamSelection();
                    allTeamNamesCache.value = [];
                    teamOptions.value = [];
                }
            },
            { deep: true }, // deep might be intensive if allPlayersData is huge
        );

        watch(
            () => route.query.datasetId,
            async (newId, oldId) => {
                if (newId && newId !== oldId) {
                    sessionStorage.setItem("currentDatasetId", newId);
                    await fetchPlayersAndCurrency(newId); // Use combined fetch
                    clearTeamSelection(); // Reset team selection as data has changed
                    if (
                        !pageLoadingError.value &&
                        allPlayersData.value.length > 0
                    ) {
                        populateTeamFilterOptions();
                    }
                }
            },
        );

        watch(
            () => route.query.team,
            (newTeam) => {
                if (newTeam && newTeam.trim() !== '' && newTeam !== selectedTeamName.value) {
                    selectedTeamName.value = newTeam;
                    loadTeamPlayers();
                }
            },
        );

        return {
            allPlayersData,
            selectedTeamName,
            teamOptions,
            filterTeamOptions,
            loadTeamPlayers,
            clearTeamSelection,
            teamPlayers,
            loadingTeam,
            pageLoading,
            pageLoadingError,
            selectedFormationKey,
            formationOptions,
            currentFormationLayout,
            squadComposition,
            bestTeamPlayersForPitch,
            bestTeamAverageOverall,
            calculationMessage,
            calculationMessageClass,
            playerForDetailView,
            showPlayerDetailDialog,
            handlePlayerSelectedFromTeam,
            teamIsGoalkeeperView,
            getOverallClass,
            getSlotDisplayName,
            handlePlayerMovedOnPitch,
            quasarInstance,
            router,
            detectedCurrencySymbol, // Expose currency symbol
        };
    },
};
</script>

<style lang="scss" scoped>
.team-view-page {
    max-width: 1600px;
    margin: 0 auto;
}

.page-title {
    // Standard title styling
}

.filter-card,
.q-card {
    // General card styling for this page
    border-radius: $generic-border-radius;
}

.team-player-table {
    :deep(.q-table__container) {
        max-height: 450px; // Or your desired height
        overflow-y: auto;
    }
    // Sticky header for the team player table
    :deep(th) {
        position: sticky;
        top: 0;
        z-index: 1; // Ensure header is above scrolling content
    }
    .body--dark & :deep(th) {
        background-color: $grey-9 !important; // Dark mode header background
    }
    .body--light & :deep(th) {
        background-color: $grey-2 !important; // Light mode header background
    }
}

.attribute-value {
    display: inline-block;
    min-width: 28px; // Ensure some width for small numbers
    text-align: center;
    font-weight: 600;
    padding: 2px 5px;
    border-radius: 4px;
    line-height: 1.3;
    font-size: 0.8em; // Slightly smaller for badges
}

.overall-badge {
    font-size: 0.85em;
    padding: 2px 4px;
}

.depth-player-item {
    padding-top: 4px;
    padding-bottom: 4px;
    min-height: auto;
    transition: background-color 0.2s ease;

    .player-rank {
        font-size: 0.8em;
        color: $grey-6;
        .body--dark & {
            color: $grey-5;
        }
        font-weight: bold;
        min-width: 18px; // Space for "1.", "2.", "3."
        text-align: right;
        margin-right: 4px;
    }
    .player-name {
        font-size: 0.85em;
        font-weight: 500;
    }
    .backup-label {
        font-size: 0.7em;
        font-style: italic;
    }

    &.starter-highlight {
        background-color: rgba($positive, 0.1); // Light green for starters
        .body--dark & {
            background-color: rgba($positive, 0.2);
        }
        .player-name {
            font-weight: 700; // Bolder name for starters
        }
    }
    &:hover {
        background-color: rgba($primary, 0.05);
        .body--dark & {
            background-color: rgba($primary, 0.15);
        }
    }
}

.q-list--separator > .q-item:not(:first-child):before {
    background: rgba(128, 128, 128, 0.15); // Lighter separator
    .body--dark & {
        background: rgba(255, 255, 255, 0.1);
    }
}

// Position match indicators
.position-match-indicator {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 4px;
    flex-shrink: 0;
    
    &.exact-match {
        background-color: #4caf50; // Green for natural position
        box-shadow: 0 0 2px rgba(76, 175, 80, 0.7);
    }
    
    &.off-position {
        background-color: #ff9800; // Orange for off position
        box-shadow: 0 0 2px rgba(255, 152, 0, 0.7);
    }
}

.d-flex {
    display: flex !important;
}

.align-items-center {
    align-items: center !important;
}

.q-mr-xs {
    margin-right: 4px !important;
}

// Compact Squad Depth Styles
.compact-squad-depth {
    max-height: 500px;
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 8px;
    width: 100%;
    
    .depth-position-compact {
        padding: 8px;
        border-radius: 6px;
        border: 1px solid rgba(0, 0, 0, 0.1);
        min-height: 100px;
        
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.08);
            border-color: rgba(255, 255, 255, 0.1);
        }
        
        .body--light & {
            background-color: rgba(0, 0, 0, 0.05);
            border-color: rgba(0, 0, 0, 0.1);
        }
    }
    
    .position-label {
        font-size: 0.75rem;
        font-weight: 700;
        margin-bottom: 6px;
        text-align: center;
        color: $grey-7;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        
        .body--dark & {
            color: $grey-3;
        }
    }
    
    .depth-players-compact {
        display: flex;
        flex-direction: column;
        gap: 3px;
    }
    
    .depth-player-compact {
        display: flex;
        align-items: center;
        gap: 4px;
        padding: 3px 6px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.75rem;
        min-height: 22px;
        
        &.starter {
            font-weight: 700;
            background-color: rgba($positive, 0.15);
            border: 1px solid rgba($positive, 0.3);
            
            .body--dark & {
                background-color: rgba($positive, 0.25);
                border-color: rgba($positive, 0.4);
            }
        }
        
        &.backup {
            font-weight: 500;
            background-color: rgba($grey-5, 0.1);
            
            .body--dark & {
                background-color: rgba($grey-5, 0.15);
            }
        }
        
        &:hover {
            background-color: rgba($primary, 0.15);
            transform: translateY(-1px);
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            
            .body--dark & {
                background-color: rgba($primary, 0.25);
            }
        }
    }
    
    .player-rank-compact {
        font-size: 0.65rem;
        font-weight: bold;
        min-width: 14px;
        color: $grey-6;
        
        .body--dark & {
            color: $grey-4;
        }
    }
    
    .player-name-compact {
        flex: 1;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        font-size: 0.72rem;
    }
    
    .overall-compact {
        font-size: 0.7rem;
        font-weight: bold;
        padding: 2px 4px;
        border-radius: 3px;
        min-width: 24px;
        text-align: center;
        border: 1px solid rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            border-color: rgba(255, 255, 255, 0.1);
        }
    }
    
    .no-players-compact {
        font-size: 0.7rem;
        color: $grey-6;
        font-style: italic;
        text-align: center;
        padding: 8px;
        
        .body--dark & {
            color: $grey-5;
        }
    }
}

// Ensure global rating tier colors are applied if not already via app.scss
.rating-tier-6 {
    /* styles from app.scss */
}
.rating-tier-5 {
    /* styles from app.scss */
}
// ... etc. for all tiers
</style>
