// src/pages/PlayerUploadPage.vue
<template>
    <q-page padding>
        <div class="q-pa-md">
            <h1
                class="text-h4 text-center q-mb-lg page-title"
                :class="$q.dark.isActive ? 'text-grey-2' : 'text-grey-9'"
            >
                Football Manager HTML Player Parser
            </h1>

            <q-card
                class="q-mb-md instructions-card"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-grey-3'
                        : 'bg-blue-grey-1 text-blue-grey-10'
                "
            >
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
                            uses internal defaults. (Note: Role keys in
                            <code>role_specific_overall_weights.json</code> now
                            expect full names like "Centre Back - Ball Playing
                            Defender").
                        </li>
                        <li>
                            Select an HTML file exported from Football Manager.
                        </li>
                        <li>
                            Click "Upload and Parse". The data will be stored
                            temporarily on the server. The app will detect the
                            currency symbol (e.g., €, $, £) from the data.
                        </li>
                        <li>
                            The table will display players with pre-calculated
                            FIFA-style stats (PHY, SHO, etc.), parsed positions,
                            Overall ratings (based on their best role), Age,
                            Media Handling, and Personality. Goalkeeping (GK)
                            stats will appear if filtering for GKs. Monetary
                            values will use the detected currency symbol.
                        </li>
                        <li>
                            Use filters for Name, Club (searchable dropdown),
                            Position (now short names like GK, DC, ST, sorted
                            GK-ST), Nationality (searchable dropdown), Transfer
                            Value (text input, slider, and mode using the
                            detected currency), Media Handling (multi-select),
                            Personality (multi-select), and Age range. Input
                            fields are debounced for performance.
                        </li>
                        <li>
                            Click on any player row for a detailed view, which
                            will show all calculated role-specific overalls (now
                            with full role names) and specific goalkeeping
                            attributes if applicable. Player positions will now
                            be displayed as short names (e.g., GK, DC, ST).
                            Monetary values will use the detected currency
                            symbol.
                        </li>
                        <li>
                            Use the "View Team Page" button to navigate to the
                            team analysis section using the uploaded data and
                            currency.
                        </li>
                    </ol>
                </q-card-section>
            </q-card>

            <q-card
                class="q-mb-md upload-card"
                :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'"
            >
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Upload HTML File:</div>
                    <q-file
                        v-model="playerFile"
                        label="Select HTML file"
                        accept=".html"
                        outlined
                        counter
                        :label-color="$q.dark.isActive ? 'grey-4' : ''"
                        :input-class="$q.dark.isActive ? 'text-grey-3' : ''"
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

            <PlayerFilters
                v-if="allPlayers.length > 0"
                :players="allPlayers"
                :currency-symbol="detectedCurrencySymbol"
                :transfer-value-range="playerStore.transferValueRange"
                :unique-clubs="playerStore.uniqueClubs"
                :unique-nationalities="playerStore.uniqueNationalities"
                :unique-media-handlings="playerStore.uniqueMediaHandlings"
                :unique-personalities="playerStore.uniquePersonalities"
                @filter-changed="handleFilterChanged"
            />

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
                <div class="row q-col-gutter-md q-mb-md summary-cards">
                    <div class="col-12 col-md-2">
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
                        >
                            <q-card-section>
                                <div class="text-h6">
                                    {{ allPlayers.length }}
                                </div>
                                <div class="text-subtitle2">Total Players</div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-2">
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
                        >
                            <q-card-section>
                                <div class="text-h6">
                                    {{ filteredPlayers.length }}
                                </div>
                                <div class="text-subtitle2">Filtered</div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-2">
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
                        >
                            <q-card-section>
                                <div class="text-h6">
                                    {{ uniqueClubsCount }}
                                </div>
                                <div class="text-subtitle2">Clubs</div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
                        >
                            <q-card-section>
                                <div class="text-h6">
                                    {{ uniqueParsedPositionsCount }}
                                </div>
                                <div class="text-subtitle2">Positions</div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
                        >
                            <q-card-section>
                                <div class="text-h6">
                                    {{ uniqueNationalitiesCount }}
                                </div>
                                <div class="text-subtitle2">Nationalities</div>
                            </q-card-section>
                        </q-card>
                    </div>
                </div>

                <div class="row justify-between items-center q-mb-md">
                    <q-btn
                        color="info"
                        icon="groups"
                        label="View Team Page"
                        @click="goToTeamView"
                        :disable="allPlayers.length === 0 || !currentDatasetId"
                        class="q-px-lg"
                    />
                    <q-btn
                        color="secondary"
                        icon="find_replace"
                        label="Find Upgrades"
                        @click="showUpgradeFinder = true"
                        :disable="allPlayers.length === 0"
                        class="q-px-lg"
                    />
                </div>

                <PlayerDataTable
                    :players="filteredPlayers"
                    :loading="loading"
                    @update:sort="handleSort"
                    @player-selected="handlePlayerSelected"
                    :is-goalkeeper-view="isGoalkeeperView"
                    :currency-symbol="detectedCurrencySymbol"
                />
            </template>

            <q-card
                v-else-if="!loading"
                class="q-pa-lg text-center no-data-card"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-grey-5'
                        : 'bg-grey-1 text-grey-7'
                "
                flat
                bordered
            >
                <q-icon name="upload_file" size="4rem" />
                <div class="text-h6 q-mt-md">No Player Data Yet</div>
                <div>Upload a file to see player data</div>
            </q-card>
        </div>

        <PlayerDetailDialog
            :player="selectedPlayer"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
            :currency-symbol="detectedCurrencySymbol"
        />

        <UpgradeFinderDialog
            :show="showUpgradeFinder"
            :players="allPlayers"
            @close="showUpgradeFinder = false"
            :currency-symbol="detectedCurrencySymbol"
        />
    </q-page>
</template>

<script>
import { ref, computed, reactive, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { usePlayerStore } from "../stores/playerStore";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import UpgradeFinderDialog from "../components/UpgradeFinderDialog.vue";
import PlayerFilters from "../components/filters/PlayerFilters.vue";
import { formatCurrency, parseCurrencyString } from "../utils/currencyUtils";

// MODIFIED: Ordered short positions for filter dropdown
const orderedShortPositions = [
    "GK",
    "DR",
    "DC",
    "DL",
    "WBR",
    "WBL",
    "DM",
    "MR",
    "MC",
    "ML",
    "AMR",
    "AMC",
    "AML",
    "ST",
];

// MODIFIED: Mapping from short position codes (filter values) to standardized full names (for matching player.parsedPositions)
const shortToStandardizedLongPosMap = {
    GK: ["Goalkeeper"],
    DR: ["Right Back"],
    DC: ["Centre Back"],
    DL: ["Left Back"],
    WBR: ["Right Wing-Back"],
    WBL: ["Left Wing-Back"],
    DM: ["Centre Defensive Midfielder"], // Assuming DM primarily maps to Centre DM for filtering
    MR: ["Right Midfielder"],
    MC: ["Centre Midfielder"],
    ML: ["Left Midfielder"],
    AMR: ["Right Attacking Midfielder", "Right Winger"], // Include winger if it's a distinct parsedPosition
    AMC: ["Centre Attacking Midfielder"],
    AML: ["Left Attacking Midfielder", "Left Winger"], // Include winger
    ST: ["Striker"],
};
// Note: The `positionGroups` definition from the original file is not directly used for this specific filter logic,
// but the `shortToStandardizedLongPosMap` should align with how `parsedPositions` are generated in Go.

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
    components: { 
        PlayerDataTable, 
        PlayerDetailDialog, 
        UpgradeFinderDialog,
        PlayerFilters
    },
    setup() {
        const router = useRouter();
        const playerStore = usePlayerStore();
        const playerFile = ref(null);
        const filteredPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false);
        
        // Derived properties from store
        const allPlayers = computed(() => playerStore.allPlayers);
        const currentDatasetId = computed(() => playerStore.currentDatasetId);
        const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol);
        const loading = computed(() => playerStore.loading);
        const error = computed(() => playerStore.error);
        
        const uniqueClubsCount = computed(() => playerStore.uniqueClubs.length);
        const uniqueNationalitiesCount = computed(() => playerStore.uniqueNationalities.length);
        const uniqueParsedPositionsCount = computed(() => playerStore.uniquePositionsCount);

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

        const transferValueTextInput = ref("");

        const filters = reactive({
            name: "",
            club: null,
            selectedTransferValue: null,
            transferValueMode: "less",
            position: null, // Will hold short position code, e.g., "DC"
            nationality: null,
            mediaHandling: [],
            personality: [],
            minAge: null,
            maxAge: null,
        });

        const clubOptions = ref([]);
        const nationalityOptions = ref([]);
        const mediaHandlingOptions = ref([]);
        const personalityOptions = ref([]);

        const allUniqueClubs = ref([]);
        const allUniqueNationalities = ref([]);
        const allUniqueMediaHandlings = ref([]);
        const allUniquePersonalities = ref([]);

        const isGoalkeeperView = computed(() => {
            // Check if the selected short position filter implies a goalkeeper view
            if (!filters.position) return false; // No position filter active
            const longNames = shortToStandardizedLongPosMap[filters.position];
            return longNames ? longNames.includes("Goalkeeper") : false;
        });

        const hasActiveFilters = computed(
            () =>
                filters.name !== "" ||
                filters.club !== null ||
                filters.selectedTransferValue !== null ||
                filters.position !== null ||
                filters.nationality !== null ||
                (Array.isArray(filters.mediaHandling) &&
                    filters.mediaHandling.length > 0) ||
                (Array.isArray(filters.personality) &&
                    filters.personality.length > 0) ||
                filters.minAge !== null ||
                filters.maxAge !== null,
        );

        // These are now replaced by the computed properties from playerStore below

        const formatSliderValueWithCurrency = (value) => {
            if (value === null || value === undefined) return "";
            return formatCurrency(value, detectedCurrencySymbol.value);
        };

        const transferValueSliderStep = computed(() => {
            const range = playerStore.transferValueRange.max - playerStore.transferValueRange.min;
            if (range <= 0) return 10000;
            if (range < 50000) return 1000;
            if (range < 250000) return 5000;
            if (range < 1000000) return 10000;
            if (range < 10000000) return 50000;
            if (range < 50000000) return 100000;
            return 250000;
        });

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
            // Load player data from session storage via store
            playerStore.loadFromSessionStorage();
        });

        const processPlayersFromAPI = (playersData) => {
            return playersData.map((p) => ({
                ...p,
                age: parseInt(p.age, 10) || 0,
            }));
        };

        const fetchPlayersByDatasetId = async (datasetId) => {
            if (!datasetId) return;
            loading.value = true;
            error.value = "";
            try {
                const response =
                    await playerService.getPlayersByDatasetId(datasetId);
                allPlayers.value = processPlayersFromAPI(response.players);
                detectedCurrencySymbol.value = response.currencySymbol || "$";
                sessionStorage.setItem(
                    "detectedCurrencySymbol",
                    detectedCurrencySymbol.value,
                );
            } catch (e) {
                error.value = `Failed to fetch player data for dataset ${datasetId}: ${e.message || "Unknown error"}`;
                allPlayers.value = [];
                currentDatasetId.value = null;
                detectedCurrencySymbol.value = "$";
                sessionStorage.removeItem("currentDatasetId");
                sessionStorage.removeItem("detectedCurrencySymbol");
            } finally {
                loading.value = false;
            }
        };

        // MODIFIED: positionFilterOptions to use short names and defined order
        const positionFilterOptions = computed(() => {
            const options = [{ label: "Any Position", value: null }];
            orderedShortPositions.forEach((shortPos) => {
                // Check if any player can actually play this short position based on the map
                // This ensures only relevant short positions are shown if data is sparse,
                // though for a full dataset, all should be relevant.
                // For simplicity now, we'll include all defined short positions.
                // A more dynamic approach would be to check if shortToStandardizedLongPosMap[shortPos]
                // has any corresponding player.parsedPositions in the current allPlayers.value.
                options.push({ label: shortPos, value: shortPos });
            });
            return options;
        });

        // This function is no longer needed as the playerStore manages all options
        const updateDropdownOptionsAndSliderBounds = () => {
            // Initial value for selectedTransferValue if null
            if (filters.selectedTransferValue === null) {
                filters.selectedTransferValue = playerStore.transferValueRange.max;
                transferValueTextInput.value = formatSliderValueWithCurrency(
                    playerStore.transferValueRange.max
                );
            } else {
                // Ensure selectedTransferValue is within current range
                filters.selectedTransferValue = Math.max(
                    playerStore.transferValueRange.min,
                    Math.min(
                        filters.selectedTransferValue,
                        playerStore.transferValueRange.max
                    )
                );
            }
        };

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
                    (p) => p.club === filters.club,
                );
            }
            if (filters.nationality) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.nationality === filters.nationality,
                );
            }

            if (filters.mediaHandling && filters.mediaHandling.length > 0) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.media_handling) return false;
                    const playerStyles = p.media_handling
                        .split(",")
                        .map((s) => s.trim().toLowerCase());
                    const filterStylesLower = filters.mediaHandling.map((s) =>
                        s.toLowerCase(),
                    );
                    return playerStyles.some((style) =>
                        filterStylesLower.includes(style),
                    );
                });
            }

            if (filters.personality && filters.personality.length > 0) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.personality) return false;
                    return filters.personality.includes(p.personality);
                });
            }

            if (filters.minAge !== null && filters.minAge >= 0) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.age >= filters.minAge,
                );
            }
            if (filters.maxAge !== null && filters.maxAge >= 0) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.age <= filters.maxAge,
                );
            }

            // MODIFIED: Position filtering logic
            if (filters.position) {
                const targetLongNames =
                    shortToStandardizedLongPosMap[filters.position];
                if (targetLongNames && targetLongNames.length > 0) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.parsedPositions &&
                            p.parsedPositions.some((pp) =>
                                targetLongNames.includes(pp),
                            ),
                    );
                }
            }

            if (
                filters.selectedTransferValue !== null &&
                filters.selectedTransferValue < playerStore.transferValueRange.max
            ) {
                const threshold = filters.selectedTransferValue;
                if (filters.transferValueMode === "less") {
                    tempPlayers = tempPlayers.filter(
                        (p) => (p.transferValueAmount || 0) < threshold,
                    );
                } else if (filters.transferValueMode === "more") {
                    tempPlayers = tempPlayers.filter(
                        (p) => (p.transferValueAmount || 0) > threshold,
                    );
                }
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

        const debouncedApplyFilters = debounce(applyFiltersAndSort, 300);

        const updateNumericValueFromTextInput = () => {
            const numericValue = parseCurrencyString(
                transferValueTextInput.value,
            );
            if (numericValue !== null) {
                const clampedValue = Math.max(
                    playerStore.transferValueRange.min,
                    Math.min(numericValue, playerStore.transferValueRange.max),
                );
                if (filters.selectedTransferValue !== clampedValue) {
                    filters.selectedTransferValue = clampedValue;
                }
            } else if (transferValueTextInput.value.trim() === "") {
                if (
                    filters.selectedTransferValue !==
                    playerStore.transferValueRange.max
                ) {
                    filters.selectedTransferValue =
                        playerStore.transferValueRange.max;
                }
            }
            applyFiltersAndSort();
        };
        const debouncedUpdateNumericValueFromTextInput = debounce(
            updateNumericValueFromTextInput,
            400,
        );

        watch(
            () => filters.selectedTransferValue,
            (newValue) => {
                const currentTextParsed = parseCurrencyString(
                    transferValueTextInput.value,
                );
                if (
                    currentTextParsed !== newValue ||
                    newValue === playerStore.transferValueRange.max
                ) {
                    transferValueTextInput.value =
                        newValue === playerStore.transferValueRange.max &&
                        newValue !== null
                            ? ""
                            : formatSliderValueWithCurrency(newValue);
                }
            },
        );

        const uploadAndParse = async () => {
            if (!playerFile.value) {
                playerStore.error = "Please select an HTML file first.";
                return;
            }
            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                await playerStore.uploadPlayerFile(formData);
                sortState.key = null;
            } catch (e) {
                // Error handling is already done in the store
                console.error(e);
            }
        };

        const handleSort = (sortParams) => {
            sortState.key = sortParams.key;
            sortState.direction = sortParams.direction;
            applyFiltersAndSort();
        };

        const clearAllFilters = () => {
            filters.name = "";
            filters.club = null;
            filters.selectedTransferValue = playerStore.transferValueRange.max;
            filters.transferValueMode = "less";
            transferValueTextInput.value = "";
            filters.position = null;
            filters.nationality = null;
            filters.mediaHandling = [];
            filters.personality = [];
            filters.minAge = null;
            filters.maxAge = null;
            applyFiltersAndSort();
        };

        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        const filterClubOptions = (val, update) => {
            if (val === "") {
                update(() => {
                    clubOptions.value = allUniqueClubs.value;
                });
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                clubOptions.value = allUniqueClubs.value.filter(
                    (v) => v.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        const filterNationalityOptions = (val, update) => {
            if (val === "") {
                update(() => {
                    nationalityOptions.value = allUniqueNationalities.value;
                });
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                nationalityOptions.value = allUniqueNationalities.value.filter(
                    (v) => v.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        watch(
            () => allPlayers.value,
            (newPlayers) => {
                updateDropdownOptionsAndSliderBounds();
                if (!newPlayers || newPlayers.length === 0) {
                    filters.selectedTransferValue = playerStore.transferValueRange.max;
                    transferValueTextInput.value = "";
                }
                applyFiltersAndSort();
            },
            { deep: true, immediate: true },
        );

        const goToTeamView = () => {
            if (playerStore.currentDatasetId) {
                router.push({
                    name: "team-view",
                    query: { datasetId: playerStore.currentDatasetId },
                });
            } else {
                playerStore.error = "No data uploaded yet. Please upload a file first.";
            }
        };

        // Move these up before the watch statement that uses them
        // Derived properties from store

        // We're using the existing filters reactive object declared earlier

        const handleFilterChanged = (newFilters) => {
            Object.assign(filters, newFilters);
            applyFiltersAndSort();
        };

        return {
            playerFile,
            playerStore,
            loading,
            error,
            allPlayers,
            filteredPlayers,
            uniqueClubsCount,
            uniqueParsedPositionsCount,
            uniqueNationalitiesCount,
            filters,
            uploadAndParse,
            handleSort,
            applyFiltersAndSort,
            selectedPlayer,
            showPlayerDetailDialog,
            handlePlayerSelected,
            attributeWeightsLoadedForFeedback,
            attributeWeightsErrorForFeedback,
            roleSpecificOverallWeightsLoadedForFeedback,
            roleSpecificOverallWeightsErrorForFeedback,
            showUpgradeFinder,
            isGoalkeeperView,
            goToTeamView,
            currentDatasetId,
            detectedCurrencySymbol,
            handleFilterChanged,
        };
    },
};
</script>

<style lang="scss" scoped>
.q-page {
    max-width: 1600px;
    margin: 0 auto;
    padding-top: 24px;
    padding-bottom: 24px;
}

.page-title {
    // Already has text-center
}

.instructions-card {
    border-radius: $generic-border-radius;
    ol {
        padding-left: 20px;
        li {
            margin-bottom: 0.5em;
        }
    }
}

.upload-card,
.filter-card,
.summary-card,
.no-data-card {
    border-radius: $generic-border-radius;
}

.summary-cards .q-card {
    height: 100%;
}

:deep(.q-field__native),
:deep(.q-field__label) {
    color: currentColor;
}
</style>
