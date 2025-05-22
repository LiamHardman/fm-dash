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

                    <q-card
                        :class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-9'
                                : 'bg-white'
                        "
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
                                        v-if="bestTeamOverall !== null"
                                        class="q-mt-md text-subtitle1"
                                        :class="
                                            quasarInstance.dark.isActive
                                                ? 'text-grey-3'
                                                : 'text-grey-8'
                                        "
                                    >
                                        Best Possible Team Overall:
                                        <span
                                            class="text-weight-bold attribute-value"
                                            :class="
                                                getOverallClass(bestTeamOverall)
                                            "
                                        >
                                            {{ bestTeamOverall }}
                                        </span>
                                    </div>
                                    <q-banner
                                        v-if="calculationMessage"
                                        class="q-mt-sm"
                                        :class="calculationMessageClass"
                                    >
                                        {{ calculationMessage }}
                                    </q-banner>
                                </div>
                                <div class="col-12 col-md-8">
                                    <PitchDisplay
                                        :formation="currentFormationLayout"
                                        :players="bestTeamPlayers"
                                        @player-click="
                                            handlePlayerSelectedFromTeam
                                        "
                                        @player-moved="handlePlayerMovedOnPitch"
                                    />
                                </div>
                            </div>
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
        />
    </q-page>
</template>

<script>
import { ref, computed, onMounted, watch } from "vue";
import { useQuasar } from "quasar";
import { useRouter, useRoute } from "vue-router";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import PitchDisplay from "../components/PitchDisplay.vue";
import { formations, getFormationLayout } from "../utils/formations";
import playerService from "../services/playerService";

// Mapping from formation slot roles (which are now more FM-like from formations.js)
// to the detailed position strings found in player.parsedPositions or player.roleSpecificOveralls.
// This map helps bridge the gap between formation slot roles and player's specific position capabilities.
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

        const allPlayersData = ref([]);
        const selectedTeamName = ref(null);
        const teamOptions = ref([]);
        const allTeamNamesCache = ref([]);
        const teamPlayers = ref([]);
        const loadingTeam = ref(false);
        const pageLoading = ref(true);
        const pageLoadingError = ref("");

        const selectedFormationKey = ref(null);
        const bestTeamPlayers = ref({});
        const bestTeamOverall = ref(null);
        const calculationMessage = ref("");
        const calculationMessageClass = ref("");

        const playerForDetailView = ref(null);
        const showPlayerDetailDialog = ref(false);

        // NEW: Map to convert descriptive position names (from fmSlotRoleMatcher)
        // to the abbreviated prefixes used in player.roleSpecificOveralls (e.g., "MC", "DC").
        // Keys should be uppercase as they are looked up with .toUpperCase().
        const fmMatcherToRoleKeyPrefix = {
            GOALKEEPER: "GK",
            SWEEPER: "DC", // Sweepers often use DC role specifics
            "DEFENDER (RIGHT)": "DR/L",
            "RIGHT BACK": "DR/L",
            "DEFENDER (LEFT)": "DR/L",
            "LEFT BACK": "DR/L",
            "DEFENDER (CENTRE)": "DC",
            "CENTRE BACK": "DC",
            "WING-BACK (RIGHT)": "WBR/L",
            "RIGHT WING-BACK": "WBR/L",
            "WING-BACK (LEFT)": "WBR/L",
            "LEFT WING-BACK": "WBR/L",
            "DEFENSIVE MIDFIELDER (CENTRE)": "DM",
            "CENTRE DEFENSIVE MIDFIELDER": "DM",
            "MIDFIELDER (RIGHT)": "MR/L",
            "RIGHT MIDFIELDER": "MR/L",
            "MIDFIELDER (LEFT)": "MR/L",
            "LEFT MIDFIELDER": "MR/L",
            "MIDFIELDER (CENTRE)": "MC",
            "CENTRE MIDFIELDER": "MC",
            "ATTACKING MIDFIELDER (RIGHT)": "AMR/L",
            "RIGHT ATTACKING MIDFIELDER": "AMR/L",
            "WINGER (RIGHT)": "AMR/L",
            "ATTACKING MIDFIELDER (LEFT)": "AMR/L",
            "LEFT ATTACKING MIDFIELDER": "AMR/L",
            "WINGER (LEFT)": "AMR/L",
            "ATTACKING MIDFIELDER (CENTRE)": "AMC",
            "CENTRE ATTACKING MIDFIELDER": "AMC",
            "STRIKER (CENTRE)": "ST",
            STRIKER: "ST",
        };

        const fetchPlayers = async (datasetId) => {
            pageLoading.value = true;
            pageLoadingError.value = "";
            allPlayersData.value = [];
            try {
                const players =
                    await playerService.getPlayersByDatasetId(datasetId);
                allPlayersData.value = players.map((p) => ({
                    ...p,
                    age: parseInt(p.age, 10) || 0,
                }));
            } catch (err) {
                pageLoadingError.value = `Failed to load player data: ${err.message || "Unknown server error"}. Please try uploading again.`;
                allPlayersData.value = [];
            } finally {
                pageLoading.value = false;
            }
        };

        onMounted(async () => {
            const datasetIdFromQuery = route.query.datasetId;
            const datasetIdFromSession =
                sessionStorage.getItem("currentDatasetId");
            let finalDatasetId = null;

            if (datasetIdFromQuery) {
                finalDatasetId = datasetIdFromQuery;
                if (datasetIdFromQuery !== datasetIdFromSession) {
                    sessionStorage.setItem(
                        "currentDatasetId",
                        datasetIdFromQuery,
                    );
                }
            } else if (datasetIdFromSession) {
                finalDatasetId = datasetIdFromSession;
                router.replace({ query: { datasetId: finalDatasetId } });
            }

            if (finalDatasetId) {
                await fetchPlayers(finalDatasetId);
            } else {
                pageLoadingError.value =
                    "No player dataset ID found. Please upload a file on the main page.";
                pageLoading.value = false;
            }

            if (!pageLoadingError.value && allPlayersData.value.length > 0) {
                populateTeamFilterOptions();
                if (selectedTeamName.value) {
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
                bestTeamPlayers.value = {};
                bestTeamOverall.value = null;
                calculationMessage.value = "";
                selectedFormationKey.value = null;
                return;
            }
            loadingTeam.value = true;
            setTimeout(() => {
                if (Array.isArray(allPlayersData.value)) {
                    teamPlayers.value = allPlayersData.value.filter(
                        (p) => p.club === selectedTeamName.value,
                    );
                } else {
                    teamPlayers.value = [];
                }
                loadingTeam.value = false;
                selectedFormationKey.value = null; // Reset formation when team changes
                bestTeamPlayers.value = {};
                bestTeamOverall.value = null;
                calculationMessage.value =
                    "Select a formation to calculate Best XI.";
                calculationMessageClass.value = quasarInstance.dark.isActive
                    ? "text-grey-5"
                    : "text-grey-7";
            }, 200);
        };

        const clearTeamSelection = () => {
            selectedTeamName.value = null;
            teamPlayers.value = [];
            selectedFormationKey.value = null;
            bestTeamPlayers.value = {};
            bestTeamOverall.value = null;
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

        const teamIsGoalkeeperView = computed(() => {
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
            if (overall >= 90) return "rating-tier-6";
            if (overall >= 80) return "rating-tier-5";
            if (overall >= 70) return "rating-tier-4";
            if (overall >= 60) return "rating-tier-3";
            if (overall >= 50) return "rating-tier-2";
            return "rating-tier-1";
        };

        // Updated getPlayerOverallForRole function
        const getPlayerOverallForRole = (player, slotFormationRole) => {
            if (!player || !slotFormationRole) return 0;

            let scoreToReturn = 0;
            if (
                player.roleSpecificOveralls &&
                player.roleSpecificOveralls.length > 0
            ) {
                const upperSlotRole = slotFormationRole.toUpperCase();
                // Get the general position matchers (e.g., "Midfielder (Centre)")
                const fmPositionMatchers = fmSlotRoleMatcher[upperSlotRole] || [
                    upperSlotRole,
                ];
                // Convert these to role key prefixes (e.g., "MC")
                const targetRoleKeyPrefixes = fmPositionMatchers
                    .map(
                        (matcher) =>
                            fmMatcherToRoleKeyPrefix[matcher.toUpperCase()],
                    )
                    .filter((prefix) => !!prefix) // Filter out any undefined if a matcher wasn't found
                    .reduce(
                        (acc, val) => (acc.includes(val) ? acc : [...acc, val]),
                        [],
                    ); // Unique prefixes

                player.roleSpecificOveralls.forEach((rso) => {
                    // rso.roleName is like "MC - AP - Attack" or "MC - Generic"
                    const rsoBasePosition = rso.roleName
                        .split(" - ")[0]
                        .trim()
                        .toUpperCase(); // Extracts "MC"
                    if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
                        scoreToReturn = Math.max(scoreToReturn, rso.score);
                    }
                });
            }
            return scoreToReturn;
        };

        const MIN_SUITABILITY_THRESHOLD = 40;

        const calculateBestTeam = () => {
            if (!selectedFormationKey.value || teamPlayers.value.length === 0) {
                bestTeamPlayers.value = {};
                bestTeamOverall.value = null;
                calculationMessage.value = selectedFormationKey.value
                    ? "No players in the selected team."
                    : "Select a formation.";
                calculationMessageClass.value = "bg-warning text-dark";
                return;
            }

            calculationMessage.value = "Calculating best team...";
            calculationMessageClass.value = quasarInstance.dark.isActive
                ? "bg-info text-white"
                : "bg-blue-2 text-primary";

            const tempBestTeamPlayers = {};
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
            let availablePlayers = [...teamPlayers.value];
            const assignedPlayerNames = new Set();
            let playersAssignedCount = 0;

            formationSlots.forEach((slot) => {
                tempBestTeamPlayers[slot.id] = null;
            });

            const gkSlot = formationSlots.find(
                (slot) => slot.role.toUpperCase() === "GK",
            );
            if (gkSlot) {
                const goalkeepers = availablePlayers
                    .filter(
                        (p) =>
                            p.parsedPositions?.includes("Goalkeeper") &&
                            !assignedPlayerNames.has(p.name),
                    )
                    .map((p) => ({
                        player: p,
                        score: getPlayerOverallForRole(p, gkSlot.role),
                    }))
                    .filter((p) => p.score >= MIN_SUITABILITY_THRESHOLD)
                    .sort((a, b) => b.score - a.score);

                if (goalkeepers.length > 0) {
                    const bestGk = goalkeepers[0];
                    tempBestTeamPlayers[gkSlot.id] = {
                        ...bestGk.player,
                        Overall: bestGk.score,
                    };
                    assignedPlayerNames.add(bestGk.player.name);
                    playersAssignedCount++;
                }
            }

            const playerScoresForSlots = [];
            availablePlayers.forEach((player) => {
                if (assignedPlayerNames.has(player.name)) return;
                formationSlots.forEach((slot) => {
                    if (tempBestTeamPlayers[slot.id]) return;
                    const score = getPlayerOverallForRole(player, slot.role);
                    if (score >= MIN_SUITABILITY_THRESHOLD) {
                        playerScoresForSlots.push({
                            playerId: player.name,
                            playerObj: player,
                            slotId: slot.id,
                            slotRole: slot.role,
                            score: score,
                            originalOverall: player.Overall,
                        });
                    }
                });
            });

            playerScoresForSlots.sort((a, b) => {
                if (b.score !== a.score) return b.score - a.score;
                return (b.originalOverall || 0) - (a.originalOverall || 0);
            });

            playerScoresForSlots.forEach((entry) => {
                if (
                    !tempBestTeamPlayers[entry.slotId] &&
                    !assignedPlayerNames.has(entry.playerId)
                ) {
                    tempBestTeamPlayers[entry.slotId] = {
                        ...entry.playerObj,
                        Overall: entry.score,
                    };
                    assignedPlayerNames.add(entry.playerId);
                    playersAssignedCount++;
                }
            });

            bestTeamPlayers.value = tempBestTeamPlayers;

            if (playersAssignedCount > 0) {
                let sumOfDisplayedOveralls = 0;
                Object.values(bestTeamPlayers.value).forEach((playerInSlot) => {
                    if (
                        playerInSlot &&
                        typeof playerInSlot.Overall === "number"
                    ) {
                        sumOfDisplayedOveralls += playerInSlot.Overall;
                    }
                });
                bestTeamOverall.value =
                    playersAssignedCount > 0
                        ? Math.round(
                              sumOfDisplayedOveralls / playersAssignedCount,
                          )
                        : 0;
                calculationMessage.value = `Best XI calculated with ${playersAssignedCount} players. Average Overall: ${bestTeamOverall.value}.`;
                calculationMessageClass.value = quasarInstance.dark.isActive
                    ? "bg-positive text-white"
                    : "bg-green-2 text-positive";
            } else {
                bestTeamOverall.value = 0;
                calculationMessage.value =
                    "Could not assign any suitable players.";
                calculationMessageClass.value = quasarInstance.dark.isActive
                    ? "bg-negative text-white"
                    : "bg-red-2 text-negative";
            }
        };

        watch(selectedFormationKey, (newKey) => {
            if (newKey) {
                calculateBestTeam();
            } else {
                bestTeamPlayers.value = {};
                bestTeamOverall.value = null;
                calculationMessage.value = "Select a formation.";
                calculationMessageClass.value = quasarInstance.dark.isActive
                    ? "text-grey-5"
                    : "text-grey-7";
            }
        });

        const handlePlayerMovedOnPitch = (moveData) => {
            const { player, fromSlotId, toSlotId, toSlotRole } = moveData;
            const newBestTeam = { ...bestTeamPlayers.value };

            const movedPlayerFullObject = teamPlayers.value.find(
                (p) => p.name === player.name,
            );
            if (!movedPlayerFullObject) return;

            const existingPlayerInTargetSlotData = newBestTeam[toSlotId];

            newBestTeam[toSlotId] = {
                ...movedPlayerFullObject,
                Overall: getPlayerOverallForRole(
                    movedPlayerFullObject,
                    toSlotRole,
                ),
            };

            if (existingPlayerInTargetSlotData && fromSlotId) {
                const fromSlotDefinition = currentFormationLayout.value
                    .flatMap((row) => row.positions)
                    .find((p) => p.id === fromSlotId);
                if (fromSlotDefinition) {
                    const existingPlayerFullObject = teamPlayers.value.find(
                        (p) => p.name === existingPlayerInTargetSlotData.name,
                    );
                    if (existingPlayerFullObject) {
                        newBestTeam[fromSlotId] = {
                            ...existingPlayerFullObject,
                            Overall: getPlayerOverallForRole(
                                existingPlayerFullObject,
                                fromSlotDefinition.role,
                            ),
                        };
                    } else {
                        newBestTeam[fromSlotId] = null;
                    }
                } else {
                    newBestTeam[fromSlotId] = null;
                }
            } else if (fromSlotId) {
                newBestTeam[fromSlotId] = null;
            }

            bestTeamPlayers.value = newBestTeam;

            let sumOfDisplayedOveralls = 0;
            let countOfDisplayedOveralls = 0;
            Object.values(bestTeamPlayers.value).forEach((playerInSlot) => {
                if (playerInSlot && typeof playerInSlot.Overall === "number") {
                    sumOfDisplayedOveralls += playerInSlot.Overall;
                    countOfDisplayedOveralls++;
                }
            });
            bestTeamOverall.value =
                countOfDisplayedOveralls > 0
                    ? Math.round(
                          sumOfDisplayedOveralls / countOfDisplayedOveralls,
                      )
                    : 0;
            calculationMessage.value = `Team adjusted. New Average Overall: ${bestTeamOverall.value}.`;
            calculationMessageClass.value = quasarInstance.dark.isActive
                ? "bg-info text-white"
                : "bg-blue-2 text-primary";
        };

        watch(
            () => allPlayersData.value,
            (newVal) => {
                if (pageLoading.value) return;
                if (newVal && newVal.length > 0) {
                    populateTeamFilterOptions();
                    if (selectedTeamName.value) loadTeamPlayers();
                } else if (!pageLoadingError.value) {
                    clearTeamSelection();
                    allTeamNamesCache.value = [];
                    teamOptions.value = [];
                }
            },
            { deep: true },
        );

        watch(
            () => route.query.datasetId,
            async (newId, oldId) => {
                if (newId && newId !== oldId) {
                    sessionStorage.setItem("currentDatasetId", newId);
                    await fetchPlayers(newId);
                    clearTeamSelection();
                    if (
                        !pageLoadingError.value &&
                        allPlayersData.value.length > 0
                    ) {
                        populateTeamFilterOptions();
                    }
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
            bestTeamPlayers,
            bestTeamOverall,
            calculationMessage,
            calculationMessageClass,
            playerForDetailView,
            showPlayerDetailDialog,
            handlePlayerSelectedFromTeam,
            teamIsGoalkeeperView,
            getOverallClass,
            handlePlayerMovedOnPitch,
            quasarInstance,
            router,
        };
    },
};
</script>

<style lang="scss" scoped>
.team-view-page {
    max-width: 1400px;
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
        max-height: 450px; // Ensure this is effective
        overflow-y: auto;
    }
    :deep(th) {
        position: sticky;
        top: 0;
        z-index: 1;
        // background-color will be handled by quasar dark mode or specific classes
    }
    .body--dark & :deep(th) {
        background-color: $grey-9 !important; // Dark mode header
    }
    .body--light & :deep(th) {
        background-color: $grey-2 !important; // Light mode header
    }
}

.attribute-value {
    display: inline-block;
    min-width: 32px;
    text-align: center;
    font-weight: 600;
    padding: 3px 6px;
    border-radius: 4px;
    line-height: 1.4;
    font-size: 0.9em; // Base size
}

// Rating tier classes are globally defined in app.scss
// Ensure they are correctly applied. Example:
.rating-tier-6 {
    /* styles from app.scss */
}
.rating-tier-5 {
    /* styles from app.scss */
}
// ... and so on for all tiers and rating-na
</style>
