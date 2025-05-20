<template>
    <q-page padding>
        <div class="q-pa-md">
            <h1 class="text-h4 text-center q-mb-lg">
                Football Manager HTML Player Parser
            </h1>

            <q-card class="q-mb-md bg-blue-1">
                <q-card-section>
                    <div class="text-subtitle1 text-weight-bold">
                        Instructions:
                    </div>
                    <ol class="q-ml-md">
                        <li>
                            Ensure the Go API (running on port 8091) is started.
                        </li>
                        <li>
                            The Go API will attempt to load
                            <code>attribute_weights.json</code> and
                            <code>role_specific_overall_weights.json</code> from
                            its <code>public</code> folder. If not found, it
                            uses internal defaults.
                        </li>
                        <li>
                            Select an HTML file exported from Football Manager.
                        </li>
                        <li>Click "Upload and Parse".</li>
                        <li>
                            The table will display players with pre-calculated
                            FIFA-style stats (PHY, SHO, etc.), parsed positions,
                            and Overall ratings (based on their best role).
                        </li>
                        <li>
                            Use filters for Name, Club, Position, Nationality,
                            and Transfer Value. Input fields are debounced for
                            performance.
                        </li>
                        <li>
                            Click on any player row for a detailed view, which
                            will show all calculated role-specific overalls
                            provided by the API.
                        </li>
                    </ol>
                </q-card-section>
            </q-card>

            <q-card class="q-mb-md">
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Upload HTML File:</div>
                    <q-file
                        v-model="playerFile"
                        label="Select HTML file"
                        accept=".html"
                        outlined
                        counter
                    >
                        <template v-slot:prepend
                            ><q-icon name="attach_file"
                        /></template>
                    </q-file>
                    <q-btn
                        class="q-mt-md full-width"
                        color="primary"
                        label="Upload and Parse"
                        :loading="loading"
                        :disable="
                            !playerFile ||
                            !attributeWeightsLoadedForFeedback ||
                            !roleSpecificOverallWeightsLoadedForFeedback
                        "
                        @click="uploadAndParse"
                    >
                        <q-tooltip
                            v-if="
                                !attributeWeightsLoadedForFeedback ||
                                !roleSpecificOverallWeightsLoadedForFeedback
                            "
                        >
                            Client-side check for weight files in public folder
                            pending... (Note: Go API has its own loading logic)
                        </q-tooltip>
                    </q-btn>
                </q-card-section>
            </q-card>

            <q-card class="q-mb-md" v-if="allPlayers.length > 0">
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Search Players</div>
                    <div class="row q-col-gutter-md">
                        <div class="col-12 col-sm-6 col-md-2">
                            <q-input
                                v-model="filters.name"
                                label="Player Name"
                                dense
                                outlined
                                clearable
                                @update:model-value="debouncedApplyFilters"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-2">
                            <q-input
                                v-model="filters.club"
                                label="Club"
                                dense
                                outlined
                                clearable
                                @update:model-value="debouncedApplyFilters"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3">
                            <q-select
                                v-model="filters.position"
                                :options="positionFilterOptions"
                                label="Position / Group"
                                dense
                                outlined
                                clearable
                                emit-value
                                map-options
                                @update:model-value="applyFiltersAndSort"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-2">
                            <q-input
                                v-model="filters.nationality"
                                label="Nationality"
                                dense
                                outlined
                                clearable
                                @update:model-value="debouncedApplyFilters"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3">
                            <q-input
                                v-model="filters.transferValue"
                                label="Transfer Value"
                                dense
                                outlined
                                clearable
                                placeholder="e.g., €1.5M, >1M, <500K"
                                @update:model-value="debouncedApplyFilters"
                            />
                        </div>
                        <div class="col-12 flex items-center q-mt-sm">
                            <q-btn
                                color="grey"
                                label="Clear All Filters"
                                class="full-width"
                                @click="clearAllFilters"
                                :disable="!hasActiveFilters"
                            />
                        </div>
                    </div>
                </q-card-section>
            </q-card>

            <q-banner
                v-if="error"
                class="bg-negative text-white q-mb-md"
                rounded
            >
                {{ error }}
                <template v-slot:action
                    ><q-btn
                        flat
                        color="white"
                        label="Dismiss"
                        @click="error = ''"
                /></template>
            </q-banner>
            <q-banner
                v-if="attributeWeightsErrorForFeedback"
                class="bg-warning text-dark q-mb-md"
                rounded
            >
                Client feedback: Could not load
                <code>public/attribute_weights.json</code>. The Go API will use
                its internal defaults if it also fails to load this file. Error:
                {{ attributeWeightsErrorForFeedback }}
            </q-banner>
            <q-banner
                v-if="roleSpecificOverallWeightsErrorForFeedback"
                class="bg-warning text-dark q-mb-md"
                rounded
            >
                Client feedback: Could not load
                <code>public/role_specific_overall_weights.json</code>. The Go
                API will use its internal defaults if it also fails to load this
                file. Error: {{ roleSpecificOverallWeightsErrorForFeedback }}
            </q-banner>

            <template v-if="allPlayers.length > 0">
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-2">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ allPlayers.length }}
                                </div>
                                <div class="text-subtitle2">
                                    Total Players
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-2">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ filteredPlayers.length }}
                                </div>
                                <div class="text-subtitle2">
                                    Filtered
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-2">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ uniqueClubsCount }}
                                </div>
                                <div class="text-subtitle2">
                                    Clubs
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ uniqueParsedPositionsCount }}
                                </div>
                                <div class="text-subtitle2">
                                    Positions
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ uniqueNationalitiesCount }}
                                </div>
                                <div class="text-subtitle2">
                                    Nationalities
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                </div>
                
                <!-- Upgrade Finder Button -->
                <div class="row justify-end q-mb-md">
                    <q-btn
                        color="primary"
                        icon="upgrade"
                        label="Find Upgrades"
                        @click="showUpgradeFinder = true"
                        :disable="filteredPlayers.length === 0"
                    />
                </div>
                
                <PlayerDataTable
                    :players="filteredPlayers"
                    :loading="loading"
                    @update:sort="handleSort"
                    @player-selected="handlePlayerSelected"
                />
            </template>

            <q-card v-else-if="!loading" class="q-pa-lg text-center">
                <q-icon name="upload_file" size="4rem" color="grey-7" />
                <div class="text-h6 q-mt-md">No Player Data Yet</div>
                <div class="text-grey-7">Upload a file to see player data</div>
            </q-card>
        </div>

        <PlayerDetailDialog
            :player="selectedPlayer"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
        />
        
        <!-- Upgrade Finder Dialog -->
        <UpgradeFinderDialog
            :show="showUpgradeFinder"
            :players="allPlayers"
            @close="showUpgradeFinder = false"
        />
    </q-page>
</template>

<script>
import { ref, computed, reactive, onMounted, watch } from "vue";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import UpgradeFinderDialog from "../components/UpgradeFinderDialog.vue";
import playerService from "../services/playerService";

// Position groups for filtering - this can remain client-side for filter generation
const positionGroups = {
    Goalkeepers: ["Goalkeeper"],
    Defenders: [
        "Sweeper",
        "Right Back",
        "Left Back",
        "Centre Back",
        "Right Wing-Back",
        "Left Wing-Back",
        "Centre Wing-Back",
    ],
    Midfielders: [
        "Right Defensive Midfielder",
        "Left Defensive Midfielder",
        "Centre Defensive Midfielder",
        "Right Midfielder",
        "Left Midfielder",
        "Centre Midfielder",
        "Right Attacking Midfielder",
        "Left Attacking Midfielder",
        "Centre Attacking Midfielder",
    ],
    Attackers: ["Striker", "Right Forward", "Left Forward", "Centre Forward"],
};

// Debounce utility function
function debounce(fn, delay) {
    let timeoutID = null;
    return function (...args) {
        clearTimeout(timeoutID);
        timeoutID = setTimeout(() => {
            fn.apply(this, args);
        }, delay);
    };
}

export default {
    name: "PlayerUploadPage",
    components: { PlayerDataTable, PlayerDetailDialog, UpgradeFinderDialog },
    setup() {
        const playerFile = ref(null);
        const loading = ref(false);
        const error = ref("");
        const allPlayers = ref([]);
        const filteredPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false);

        // Refs for client-side feedback about weight files, not for calculation
        const attributeWeightsLoadedForFeedback = ref(false);
        const attributeWeightsErrorForFeedback = ref("");
        const roleSpecificOverallWeightsLoadedForFeedback = ref(false);
        const roleSpecificOverallWeightsErrorForFeedback = ref("");

        const sortState = reactive({
            key: null,
            direction: "asc",
            isAttribute: false,
            displayField: null,
        });
        const filters = reactive({
            name: "",
            club: "",
            transferValue: "",
            position: null,
            nationality: "",
        });

        const hasActiveFilters = computed(
            () =>
                filters.name !== "" ||
                filters.club !== "" ||
                filters.transferValue !== "" ||
                filters.position !== null ||
                filters.nationality !== "",
        );
        const uniqueClubsCount = computed(
            () =>
                new Set(allPlayers.value.map((p) => p.club).filter(Boolean))
                    .size,
        );
        const uniqueParsedPositionsCount = computed(() => {
            const s = new Set();
            allPlayers.value.forEach((player) =>
                player.parsedPositions?.forEach((pos) => s.add(pos)),
            );
            return s.size;
        });
        const uniqueNationalitiesCount = computed(
            () =>
                new Set(
                    allPlayers.value.map((p) => p.nationality).filter(Boolean),
                ).size,
        );

        const parseMonetaryValueForFilter = (valueStr) => {
            if (typeof valueStr !== "string" || !valueStr) return 0;
            const cleanedStr = valueStr.split(" p/w")[0];
            let multiplier = 1;
            const lowerStr = cleanedStr.toLowerCase();
            if (lowerStr.includes("m")) multiplier = 1000000;
            else if (lowerStr.includes("k")) multiplier = 1000;
            let numStr = cleanedStr.replace(/[^0-9,.]/g, "");
            numStr = numStr.replace(/,/g, "");
            const value = parseFloat(numStr);
            return Math.round(isNaN(value) ? 0 : value * multiplier);
        };

        const loadJsonForFeedback = async (
            filePath,
            loadedFlagRef,
            errorRef,
        ) => {
            errorRef.value = "";
            try {
                const response = await fetch(filePath);
                if (!response.ok)
                    throw new Error(
                        `HTTP error! status: ${response.status} for ${filePath}`,
                    );
                await response.json();
                loadedFlagRef.value = true;
            } catch (e) {
                console.warn(
                    `Client-side check: Failed to load ${filePath}:`,
                    e,
                );
                errorRef.value =
                    e.message || `Unknown error loading ${filePath}.`;
                loadedFlagRef.value = true;
            }
        };

        onMounted(() => {
            loadJsonForFeedback(
                "/attribute_weights.json",
                attributeWeightsLoadedForFeedback,
                attributeWeightsErrorForFeedback,
            );
            loadJsonForFeedback(
                "/role_specific_overall_weights.json",
                roleSpecificOverallWeightsLoadedForFeedback,
                roleSpecificOverallWeightsErrorForFeedback,
            );
        });

        const processPlayersFromAPI = (playersData) => {
            return playersData.map((p) => ({
                ...p,
                age: parseInt(p.age, 10) || 0,
            }));
        };

        const positionFilterOptions = computed(() => {
            const options = [];
            Object.keys(positionGroups).forEach((group) => {
                options.push({ label: `${group} (Group)`, value: group });
            });
            const uniquePositions = new Set();
            allPlayers.value.forEach((player) => {
                player.parsedPositions?.forEach((pos) =>
                    uniquePositions.add(pos),
                );
            });
            Array.from(uniquePositions)
                .sort()
                .forEach((pos) => {
                    if (!positionGroups[pos]) {
                        options.push({ label: pos, value: pos });
                    }
                });
            return options;
        });

        const applyFiltersAndSort = () => {
            if (!allPlayers.value.length) {
                filteredPlayers.value = [];
                return;
            }
            let tempPlayers = [...allPlayers.value];

            if (filters.name) {
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.name &&
                        p.name
                            .toLowerCase()
                            .includes(filters.name.toLowerCase()),
                );
            }
            if (filters.club) {
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.club &&
                        p.club
                            .toLowerCase()
                            .includes(filters.club.toLowerCase()),
                );
            }
            if (filters.nationality) {
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.nationality &&
                        p.nationality
                            .toLowerCase()
                            .includes(filters.nationality.toLowerCase()),
                );
            }
            if (filters.position) {
                const selectedPosFilter = filters.position;
                if (positionGroups[selectedPosFilter]) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.positionGroups &&
                            p.positionGroups.includes(selectedPosFilter),
                    );
                } else {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.parsedPositions &&
                            p.parsedPositions.includes(selectedPosFilter),
                    );
                }
            }
            if (filters.transferValue) {
                let operator = "includes";
                let compareValueNum = 0;
                let filterStr = filters.transferValue;
                if (filterStr.startsWith(">")) {
                    operator = ">";
                    filterStr = filterStr.substring(1);
                } else if (filterStr.startsWith("<")) {
                    operator = "<";
                    filterStr = filterStr.substring(1);
                }
                if (operator === ">" || operator === "<") {
                    compareValueNum = parseMonetaryValueForFilter(filterStr);
                }
                tempPlayers = tempPlayers.filter((p) => {
                    const playerValueNum = p.transferValueAmount || 0;
                    const playerValueStr = String(
                        p.transfer_value || "",
                    ).toLowerCase();
                    if (operator === ">")
                        return playerValueNum > compareValueNum;
                    if (operator === "<")
                        return playerValueNum < compareValueNum;
                    return playerValueStr.includes(
                        filters.transferValue.toLowerCase(),
                    );
                });
            }

            if (sortState.key) {
                const sortKey = sortState.key;
                tempPlayers.sort((a, b) => {
                    let valA = a[sortKey];
                    let valB = b[sortKey];
                    if (valA == null && valB == null) return 0;
                    if (valA == null)
                        return sortState.direction === "asc" ? 1 : -1;
                    if (valB == null)
                        return sortState.direction === "asc" ? -1 : 1;
                    if (typeof valA === "number" && typeof valB === "number") {
                        return sortState.direction === "asc"
                            ? valA - valB
                            : valB - valA;
                    }
                    valA = String(valA).toLowerCase();
                    valB = String(valB).toLowerCase();
                    if (valA < valB)
                        return sortState.direction === "asc" ? -1 : 1;
                    if (valA > valB)
                        return sortState.direction === "asc" ? 1 : -1;
                    return 0;
                });
            }
            filteredPlayers.value = tempPlayers;
        };

        // Create a debounced version of the applyFiltersAndSort function
        const debouncedApplyFilters = debounce(applyFiltersAndSort, 300); // 300ms delay

        const uploadAndParse = async () => {
            if (!playerFile.value) {
                error.value = "Please select an HTML file first.";
                return;
            }
            loading.value = true;
            error.value = "";
            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                const playersDataFromApi =
                    await playerService.uploadPlayerFile(formData);
                allPlayers.value = processPlayersFromAPI(playersDataFromApi);
                sortState.key = null;
                applyFiltersAndSort();
            } catch (e) {
                error.value = `Failed to parse player data: ${e.message || "Unknown error"}`;
                allPlayers.value = [];
                filteredPlayers.value = [];
            } finally {
                loading.value = false;
            }
        };

        const handleSort = (sortParams) => {
            sortState.key = sortParams.key;
            sortState.direction = sortParams.direction;
            applyFiltersAndSort();
        };

        const clearAllFilters = () => {
            filters.name = "";
            filters.club = "";
            filters.transferValue = "";
            filters.position = null;
            filters.nationality = "";
            applyFiltersAndSort(); // Apply immediately after clearing
        };

        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        // Watchers
        // Watch allPlayers to initialize filteredPlayers or when data is loaded/cleared
        watch(
            () => allPlayers.value,
            () => {
                applyFiltersAndSort();
            },
            { deep: true },
        );

        // No longer need to watch individual filters directly as debouncedApplyFilters handles text inputs,
        // and q-select/clear button call applyFiltersAndSort directly.

        return {
            playerFile,
            loading,
            error,
            allPlayers,
            filteredPlayers,
            uniqueClubsCount,
            uniqueParsedPositionsCount,
            uniqueNationalitiesCount,
            filters,
            hasActiveFilters,
            positionFilterOptions,
            uploadAndParse,
            handleSort,
            debouncedApplyFilters, // Use this for text inputs
            applyFiltersAndSort, // Use this for q-select and clear button
            clearAllFilters,
            selectedPlayer,
            showPlayerDetailDialog,
            handlePlayerSelected,
            attributeWeightsLoadedForFeedback,
            attributeWeightsErrorForFeedback,
            roleSpecificOverallWeightsLoadedForFeedback,
            roleSpecificOverallWeightsErrorForFeedback,
            showUpgradeFinder,
        };
    },
};
</script>

<style>
.q-page {
    max-width: 1600px;
    margin: 0 auto;
}
</style>
