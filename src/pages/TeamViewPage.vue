// src/pages/TeamViewPage.vue
<template>
    <q-page padding class="team-view-page">
        <div class="q-pa-md">
            <h1
                class="text-h4 text-center q-mb-lg page-title"
                :class="$q.dark.isActive ? 'text-grey-2' : 'text-grey-9'"
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
                    @click="$router.push('/')"
                    class="q-ml-md"
                />
            </q-banner>

            <q-card
                v-if="!pageLoadingError"
                class="q-mb-md filter-card"
                :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'"
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
                        :label-color="$q.dark.isActive ? 'grey-4' : ''"
                        :popup-content-class="
                            $q.dark.isActive
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
                    :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-7'"
                >
                    Loading player data from server...
                </div>
            </div>
            <div v-else-if="loadingTeam" class="text-center q-my-xl">
                <q-spinner-dots color="primary" size="2em" />
                <div
                    class="q-mt-sm text-caption"
                    :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-7'"
                >
                    Loading team data...
                </div>
            </div>

            <div v-if="!pageLoading && !pageLoadingError">
                <div v-if="selectedTeamName && !loadingTeam">
                    <q-card
                        class="q-mb-md"
                        :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'"
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
                                    $q.dark.isActive
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
                        :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'"
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
                                        @update:model-value="calculateBestTeam"
                                        :label-color="
                                            $q.dark.isActive ? 'grey-4' : ''
                                        "
                                        :popup-content-class="
                                            $q.dark.isActive
                                                ? 'bg-grey-8 text-white'
                                                : 'bg-white text-dark'
                                        "
                                    />
                                    <div
                                        v-if="bestTeamOverall !== null"
                                        class="q-mt-md text-subtitle1"
                                        :class="
                                            $q.dark.isActive
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
                        $q.dark.isActive
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
                        $q.dark.isActive
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
                        @click="$router.push('/')"
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

export default {
    name: "TeamViewPage",
    components: { PlayerDataTable, PlayerDetailDialog, PitchDisplay },
    setup() {
        const $q = useQuasar();
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
                console.log(
                    "TeamViewPage: Successfully fetched players for dataset:",
                    datasetId,
                );
            } catch (err) {
                console.error(
                    "TeamViewPage: Error fetching players by dataset ID:",
                    err,
                );
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
                console.warn(
                    "TeamViewPage: No datasetId in query or sessionStorage.",
                );
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
                selectedFormationKey.value = null;
                bestTeamPlayers.value = {};
                bestTeamOverall.value = null;
                calculationMessage.value = "";
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

        // Corrected: Removed .reverse()
        // formations.js now defines layout with attackers first (top of pitch)
        // and GK last (bottom of pitch). PitchDisplay renders rows top-to-bottom.
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

        const calculateBestTeam = () => {
            if (!selectedFormationKey.value || teamPlayers.value.length === 0) {
                bestTeamPlayers.value = {};
                bestTeamOverall.value = null;
                calculationMessage.value =
                    "Please select a formation and ensure team players are loaded.";
                calculationMessageClass.value = "bg-warning text-dark";
                return;
            }

            calculationMessage.value = "Calculating best team...";
            calculationMessageClass.value = $q.dark.isActive
                ? "bg-info text-white"
                : "bg-blue-2 text-primary";
            bestTeamPlayers.value = {};
            bestTeamOverall.value = null;

            // Use the layout directly from formations.js (attackers first, GK last)
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

            let currentTeam = {};
            let playersAssigned = 0;

            availablePlayers.sort(
                (a, b) => (b.Overall || 0) - (a.Overall || 0),
            );

            const assignedPlayerNames = new Set();

            const gkSlot = formationSlots.find(
                (slot) => slot.role.toUpperCase() === "GK",
            );
            if (gkSlot) {
                const bestGk = availablePlayers
                    .filter(
                        (p) =>
                            p.parsedPositions?.includes("Goalkeeper") &&
                            !assignedPlayerNames.has(p.name),
                    )
                    .sort((a, b) => (b.Overall || 0) - (a.Overall || 0))[0];
                if (bestGk) {
                    currentTeam[gkSlot.id] = bestGk;
                    assignedPlayerNames.add(bestGk.name);
                    playersAssigned++;
                }
            }

            for (const slot of formationSlots) {
                if (slot.role.toUpperCase() === "GK" && currentTeam[slot.id]) {
                    continue;
                }
                if (currentTeam[slot.id]) {
                    continue;
                }

                let bestPlayerForSlot = null;
                let highestSuitabilityScore = -1;

                for (const player of availablePlayers) {
                    if (assignedPlayerNames.has(player.name)) {
                        continue;
                    }

                    let suitabilityScore = 0;
                    const slotRoleKey = slot.role.toUpperCase();

                    const roleSpecificMatch = player.roleSpecificOveralls?.find(
                        (rso) => {
                            const parts = rso.roleName.split(" - ");
                            const roleFromFile =
                                parts.length > 1 ? parts[1].toUpperCase() : "";
                            return roleFromFile === slotRoleKey;
                        },
                    );

                    if (roleSpecificMatch) {
                        suitabilityScore = roleSpecificMatch.score;
                    } else {
                        const playerPositions =
                            player.parsedPositions?.map((p) =>
                                p.toUpperCase(),
                            ) || [];
                        if (playerPositions.includes(slotRoleKey)) {
                            suitabilityScore = player.Overall || 0;
                        } else {
                            const firstTwoSlot = slotRoleKey.substring(0, 2);
                            if (
                                playerPositions.some((pp) =>
                                    pp.startsWith(firstTwoSlot),
                                )
                            ) {
                                suitabilityScore = (player.Overall || 0) * 0.8;
                            }
                        }
                    }

                    if (suitabilityScore > highestSuitabilityScore) {
                        highestSuitabilityScore = suitabilityScore;
                        bestPlayerForSlot = player;
                    }
                }

                if (bestPlayerForSlot) {
                    currentTeam[slot.id] = bestPlayerForSlot;
                    assignedPlayerNames.add(bestPlayerForSlot.name);
                    playersAssigned++;
                }
            }

            bestTeamPlayers.value = currentTeam;
            if (playersAssigned > 0) {
                let sumOfAssignedOveralls = 0;
                let countOfAssignedOveralls = 0;
                Object.values(currentTeam).forEach((player) => {
                    if (player && typeof player.Overall === "number") {
                        sumOfAssignedOveralls += player.Overall;
                        countOfAssignedOveralls++;
                    }
                });
                bestTeamOverall.value =
                    countOfAssignedOveralls > 0
                        ? Math.round(
                              sumOfAssignedOveralls / countOfAssignedOveralls,
                          )
                        : 0;

                calculationMessage.value = `Best XI calculated with ${playersAssigned} players. Average Overall: ${bestTeamOverall.value}.`;
                calculationMessageClass.value = $q.dark.isActive
                    ? "bg-positive text-white"
                    : "bg-green-2 text-positive";
            } else {
                bestTeamOverall.value = 0;
                calculationMessage.value =
                    "Could not assign any players to the formation.";
                calculationMessageClass.value = $q.dark.isActive
                    ? "bg-negative text-white"
                    : "bg-red-2 text-negative";
            }
        };

        watch(
            () => allPlayersData.value,
            (newVal) => {
                if (pageLoading.value) return;

                if (newVal && newVal.length > 0) {
                    populateTeamFilterOptions();
                    if (selectedTeamName.value) {
                        loadTeamPlayers();
                    }
                } else if (!pageLoadingError.value) {
                    allTeamNamesCache.value = [];
                    teamOptions.value = [];
                    selectedTeamName.value = null;
                    teamPlayers.value = [];
                    bestTeamPlayers.value = {};
                    bestTeamOverall.value = null;
                    calculationMessage.value = "";
                }
            },
            { deep: true },
        );

        watch(
            () => route.query.datasetId,
            async (newId, oldId) => {
                if (newId && newId !== oldId) {
                    console.log(
                        "TeamViewPage: datasetId in query changed, re-fetching.",
                        newId,
                    );
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
            calculateBestTeam,
            calculationMessage,
            calculationMessageClass,
            playerForDetailView,
            showPlayerDetailDialog,
            handlePlayerSelectedFromTeam,
            teamIsGoalkeeperView,
            getOverallClass,
            $q,
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
        max-height: 450px;
        overflow-y: auto;
    }
    :deep(th) {
        position: sticky;
        top: 0;
        z-index: 1;
    }
    .body--dark & :deep(th) {
        background-color: $grey-9 !important;
    }
    .body--light & :deep(th) {
        background-color: $grey-2 !important;
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
    font-size: 0.9em;
}

// Rating tier classes are assumed to be globally defined or imported
.rating-tier-6 {
    background-color: #7e57c2;
    color: white !important;
    font-weight: 700;
    border: 1px solid #5e35b1;
    .body--dark & {
        background-color: #9575cd;
        border-color: #7e57c2;
    }
}
.rating-tier-5 {
    background-color: #26a69a;
    color: white !important;
    .body--dark & {
        background-color: #00897b;
    }
}
.rating-tier-4 {
    background-color: #66bb6a;
    color: white !important;
    .body--dark & {
        background-color: #4caf50;
    }
}
.rating-tier-3 {
    background-color: #42a5f5;
    color: white !important;
    .body--dark & {
        background-color: #2196f3;
    }
}
.rating-tier-2 {
    background-color: #ffa726;
    color: #333333 !important;
    .body--dark & {
        background-color: #fb8c00;
        color: white !important;
    }
}
.rating-tier-1 {
    background-color: #ef5350;
    color: white !important;
    .body--dark & {
        background-color: #e53935;
    }
}
.rating-na {
    background-color: $grey-4;
    color: $grey-8 !important;
    font-weight: normal;
    border: 1px solid $grey-5;
    .body--dark & {
        background-color: $grey-8;
        color: $grey-5 !important;
        border: 1px solid $grey-7;
    }
}
</style>
