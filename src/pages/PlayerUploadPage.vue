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
                        @click="playerStore.error = ''"
                /></template>
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

const shortToStandardizedLongPosMap = {
    GK: ["Goalkeeper"],
    DR: ["Right Back"],
    DC: ["Centre Back"],
    DL: ["Left Back"],
    WBR: ["Right Wing-Back"],
    WBL: ["Left Wing-Back"],
    DM: ["Centre Defensive Midfielder"],
    MR: ["Right Midfielder"],
    MC: ["Centre Midfielder"],
    ML: ["Left Midfielder"],
    AMR: ["Right Attacking Midfielder", "Right Winger"],
    AMC: ["Centre Attacking Midfielder"],
    AML: ["Left Attacking Midfielder", "Left Winger"],
    ST: ["Striker"],
};

export default {
    name: "PlayerUploadPage",
    components: {
        PlayerDataTable,
        PlayerDetailDialog,
        UpgradeFinderDialog,
        PlayerFilters,
    },
    setup() {
        console.log(`PlayerUploadPage: Setup function start.`);
        const router = useRouter();
        const playerStore = usePlayerStore();
        const playerFile = ref(null);
        const filteredPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false);

        // Ensure allPlayers always returns an array
        const allPlayers = computed(() =>
            Array.isArray(playerStore.allPlayers) ? playerStore.allPlayers : [],
        );

        const currentDatasetId = computed(() => playerStore.currentDatasetId);
        const detectedCurrencySymbol = computed(
            () => playerStore.detectedCurrencySymbol,
        );
        const loading = computed(() => playerStore.loading);
        const error = computed({
            get: () => playerStore.error,
            set: (value) => {
                playerStore.error = value;
            },
        });

        const uniqueClubsCount = computed(() => playerStore.uniqueClubs.length);
        const uniqueNationalitiesCount = computed(
            () => playerStore.uniqueNationalities.length,
        );
        const uniqueParsedPositionsCount = computed(
            () => playerStore.uniquePositionsCount,
        );

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

        const activeFilters = ref({});

        const isGoalkeeperView = computed(() => {
            if (!activeFilters.value.position) return false;
            const longNames =
                shortToStandardizedLongPosMap[activeFilters.value.position];
            return longNames ? longNames.includes("Goalkeeper") : false;
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

        onMounted(async () => {
            console.log(`PlayerUploadPage: Component mounted.`);
            console.time("PlayerUploadPage_onMounted_tasks_total");
            await loadJsonForFeedback(
                "/attribute_weights.json",
                attributeWeightsLoadedForFeedback,
                attributeWeightsErrorForFeedback,
            );
            await loadJsonForFeedback(
                "/role_specific_overall_weights.json",
                roleSpecificOverallWeightsLoadedForFeedback,
                roleSpecificOverallWeightsErrorForFeedback,
            );

            console.time("PlayerUploadPage_playerStore_loadFromSessionStorage");
            await playerStore.loadFromSessionStorage();
            console.timeEnd(
                "PlayerUploadPage_playerStore_loadFromSessionStorage",
            );

            console.timeEnd("PlayerUploadPage_onMounted_tasks_total");
        });

        const applyFiltersAndSort = () => {
            console.time("PlayerUploadPage_applyFiltersAndSort_total");
            // Ensure allPlayers.value is an array before proceeding
            if (
                !Array.isArray(allPlayers.value) ||
                allPlayers.value.length === 0
            ) {
                filteredPlayers.value = [];
                console.log(
                    `PlayerUploadPage: No players in allPlayers or not an array. Filtered list empty.`,
                );
                console.timeEnd("PlayerUploadPage_applyFiltersAndSort_total");
                return;
            }

            let tempPlayers = [...allPlayers.value];
            const currentFilters = activeFilters.value;

            console.log(
                `PlayerUploadPage: Applying filters:`,
                JSON.parse(JSON.stringify(currentFilters)),
            );
            console.log(
                `PlayerUploadPage: Starting with ${tempPlayers.length} players.`,
            );

            console.time("PlayerUploadPage_filtering_logic");
            if (currentFilters.name) {
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.name &&
                        p.name
                            .toLowerCase()
                            .includes(currentFilters.name.toLowerCase()),
                );
            }
            if (currentFilters.club) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.club === currentFilters.club,
                );
            }
            if (currentFilters.nationality) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.nationality === currentFilters.nationality,
                );
            }
            if (
                currentFilters.mediaHandling &&
                currentFilters.mediaHandling.length > 0
            ) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.media_handling) return false;
                    const playerStyles = p.media_handling
                        .split(",")
                        .map((s) => s.trim().toLowerCase());
                    const filterStylesLower = currentFilters.mediaHandling.map(
                        (s) => s.toLowerCase(),
                    );
                    return playerStyles.some((style) =>
                        filterStylesLower.includes(style),
                    );
                });
            }
            if (
                currentFilters.personality &&
                currentFilters.personality.length > 0
            ) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.personality) return false;
                    return currentFilters.personality.includes(p.personality);
                });
            }
            if (currentFilters.minAge !== null && currentFilters.minAge >= 0) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.age >= currentFilters.minAge,
                );
            }
            if (currentFilters.maxAge !== null && currentFilters.maxAge >= 0) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.age <= currentFilters.maxAge,
                );
            }
            if (currentFilters.position) {
                const targetLongNames =
                    shortToStandardizedLongPosMap[currentFilters.position];
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
                currentFilters.selectedTransferValue !== null &&
                playerStore.transferValueRange &&
                currentFilters.selectedTransferValue <
                    playerStore.transferValueRange.max
            ) {
                const threshold = currentFilters.selectedTransferValue;
                if (currentFilters.transferValueMode === "less") {
                    tempPlayers = tempPlayers.filter(
                        (p) => (p.transferValueAmount || 0) < threshold,
                    );
                } else if (currentFilters.transferValueMode === "more") {
                    tempPlayers = tempPlayers.filter(
                        (p) => (p.transferValueAmount || 0) > threshold,
                    );
                }
            }
            console.timeEnd("PlayerUploadPage_filtering_logic");

            console.log(
                `PlayerUploadPage: After filtering, ${tempPlayers.length} players remaining.`,
            );

            filteredPlayers.value = tempPlayers;
            console.log(
                `PlayerUploadPage: filteredPlayers updated. Length: ${filteredPlayers.value.length}`,
            );
            console.timeEnd("PlayerUploadPage_applyFiltersAndSort_total");
        };

        const uploadAndParse = async () => {
            console.time("PlayerUploadPage_uploadAndParse_function_total");
            if (!playerFile.value) {
                playerStore.error = "Please select an HTML file first.";
                console.warn(`PlayerUploadPage: Upload attempt without file.`);
                console.timeEnd(
                    "PlayerUploadPage_uploadAndParse_function_total",
                );
                return;
            }
            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                console.log(
                    `PlayerUploadPage: Calling playerStore.uploadPlayerFile.`,
                );
                await playerStore.uploadPlayerFile(formData);
                console.log(
                    `PlayerUploadPage: playerStore.uploadPlayerFile completed. allPlayers length: ${allPlayers.value.length}`,
                );
                // The watcher on allPlayers will trigger applyFiltersAndSort
            } catch (e) {
                console.error(
                    `PlayerUploadPage: Error during uploadAndParse:`,
                    e,
                );
            } finally {
                console.timeEnd(
                    "PlayerUploadPage_uploadAndParse_function_total",
                );
            }
        };

        const handleSort = (sortParams) => {
            console.log(
                `PlayerUploadPage: handleSort called with:`,
                JSON.parse(JSON.stringify(sortParams)),
            );
        };

        const handlePlayerSelected = (player) => {
            console.log(`PlayerUploadPage: Player selected:`, player.name);
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        const handleFilterChanged = (newFilters) => {
            console.log(
                `PlayerUploadPage: handleFilterChanged from PlayerFilters:`,
                JSON.parse(JSON.stringify(newFilters)),
            );
            activeFilters.value = newFilters;
            applyFiltersAndSort();
        };

        watch(
            allPlayers,
            (newVal, oldVal) => {
                console.log(
                    `PlayerUploadPage: Watcher for allPlayers triggered. New length: ${newVal?.length}, Old length: ${oldVal?.length}`,
                );
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
                playerStore.error =
                    "No data uploaded yet. Please upload a file first.";
            }
        };

        console.log(`PlayerUploadPage: Setup function end.`);
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
            uploadAndParse,
            handleSort,
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
