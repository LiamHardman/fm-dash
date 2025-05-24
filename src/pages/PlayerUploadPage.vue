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
                        <li>Ensure the Go API (port 8091) is running.</li>
                        <li>
                            API loads <code>attribute_weights.json</code> and
                            <code>role_specific_overall_weights.json</code> from
                            its <code>public</code> folder.
                        </li>
                        <li>
                            Select an HTML file exported from Football Manager.
                        </li>
                        <li>
                            Click "Upload and Parse". Currency symbol will be
                            auto-detected.
                        </li>
                        <li>
                            Table shows players with FIFA-style stats, Overall
                            (best role), Age, etc.
                        </li>
                        <li>
                            Use filters: Name, Club, Position (short codes),
                            <strong>Role (specific to position)</strong>,
                            Nationality, Value ({{ detectedCurrencySymbol }}),
                            Media Handling, Personality, Age.
                        </li>
                        <li>
                            Click player row for detailed view (all role
                            overalls).
                        </li>
                        <li>Use "View Team Page" for team analysis.</li>
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
                            Client-side check for weight files pending...
                        </q-tooltip>
                    </q-btn>
                </q-card-section>
            </q-card>

            <PlayerFilters
                v-if="playerStore.currentDatasetId"
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
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
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
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
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
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
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
                        <q-card
                            class="text-center summary-card"
                            :class="
                                $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-2'
                            "
                            flat
                            bordered
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
                v-else-if="!loading && !playerStore.currentDatasetId"
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
            <q-card
                v-else-if="
                    !loading &&
                    playerStore.currentDatasetId &&
                    allPlayers.length === 0
                "
                class="q-pa-lg text-center no-data-card"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-grey-5'
                        : 'bg-grey-1 text-grey-7'
                "
                flat
                bordered
            >
                <q-icon name="sentiment_dissatisfied" size="4rem" />
                <div class="text-h6 q-mt-md">No Players Found</div>
                <div>
                    The uploaded file might not contain player data or an error
                    occurred during parsing.
                </div>
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

// This map is illustrative for client-side filtering if needed,
// but primary position filtering is now backend.
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
        const router = useRouter();
        const playerStore = usePlayerStore();
        const playerFile = ref(null);
        const filteredPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false);

        // Ensure allPlayers always returns an array and is reactive to store changes
        const allPlayers = computed(() => {
            // Directly use the store's allPlayers.value which is a shallowRef.
            // Vue's computed will re-evaluate when playerStore.allPlayers.value itself changes.
            return Array.isArray(playerStore.allPlayers)
                ? playerStore.allPlayers
                : [];
        });

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
        const roleSpecificOverallWeightsLoadedForFeedback = ref(false);

        const activeFilters = ref({}); // Stores the current filter values from PlayerFilters component

        // This computed property is for the PlayerDataTable's :is-goalkeeper-view prop.
        // It hints to the table if it should prioritize GK-specific columns.
        const isGoalkeeperView = computed(() => {
            // If a position filter is active and it's 'GK', consider it a goalkeeper view.
            // The actual filtering of players to GKs is done by the backend if 'GK' is selected.
            return activeFilters.value.position === "GK";
        });

        const loadJsonForFeedback = async (filePath, loadedFlagRef) => {
            try {
                const response = await fetch(filePath); // Assumes files are in public folder
                if (!response.ok)
                    throw new Error(`HTTP error! status: ${response.status}`);
                await response.json(); // Just to check if it's valid JSON
                loadedFlagRef.value = true;
            } catch (e) {
                console.warn(
                    `Client-side check: Failed to load ${filePath}:`,
                    e,
                );
                loadedFlagRef.value = true; // Still allow upload, backend has defaults
            }
        };

        onMounted(async () => {
            console.log(
                "PlayerUploadPage: Mounted. Current Dataset ID from store:",
                playerStore.currentDatasetId,
            );
            await loadJsonForFeedback(
                "/public/attribute_weights.json",
                attributeWeightsLoadedForFeedback,
            );
            await loadJsonForFeedback(
                "/public/role_specific_overall_weights.json",
                roleSpecificOverallWeightsLoadedForFeedback,
            );

            if (!playerStore.currentDatasetId) {
                await playerStore.loadFromSessionStorage();
            } else if (
                playerStore.allPlayers.length === 0 &&
                playerStore.currentDatasetId
            ) {
                await playerStore.fetchPlayersByDatasetId(
                    playerStore.currentDatasetId,
                );
                if (playerStore.allAvailableRoles.length === 0) {
                    // Fetch roles if not loaded
                    await playerStore.fetchAllAvailableRoles();
                }
            } else if (
                playerStore.currentDatasetId &&
                playerStore.allAvailableRoles.length === 0
            ) {
                await playerStore.fetchAllAvailableRoles();
            }
            // Initial application of any filters (usually none by default after fresh load)
            if (allPlayers.value.length > 0) {
                // Ensure allPlayers is populated before applying filters
                applyClientSideFilters(allPlayers.value, activeFilters.value);
            }
        });

        // This function applies client-side filters AFTER backend has potentially filtered by position/role
        const applyClientSideFilters = (playersToFilter, currentFilters) => {
            let tempPlayers = [...playersToFilter]; // Start with players already processed by backend (pos/role)

            // Apply client-side filters (name, club, age, value, personality, media handling)
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
            // Position and Role filtering are now primarily handled by the backend.
            // The `Overall` value is already adjusted by the backend based on role.
            // So, no explicit client-side position/role filtering on `parsedPositions` here.

            filteredPlayers.value = tempPlayers;
            console.log(
                `PlayerUploadPage: Client-side filters applied. Filtered players count: ${filteredPlayers.value.length}`,
            );
        };

        const uploadAndParse = async () => {
            if (!playerFile.value) {
                playerStore.error = "Please select an HTML file first.";
                return;
            }
            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                await playerStore.uploadPlayerFile(formData); // Corrected: Pass formData
                // The store now handles fetching players and roles.
                // The watcher on allPlayers will apply client-side filters.
                activeFilters.value = {}; // Reset client-side filter state for PlayerFilters component
                // PlayerFilters will emit its default values on next interaction or mount
            } catch (e) {
                // Error is set in the store
                console.error("Upload and Parse error in page:", e);
            }
        };

        const handleSort = (sortParams) => {
            // This function is kept for potential future client-side sorting enhancements
            // or if PlayerDataTable's internal sort needs to be overridden.
            // Currently, PlayerDataTable handles its own sorting based on the `filteredPlayers` prop.
            console.log(
                "PlayerUploadPage: Sort requested by PlayerDataTable:",
                sortParams,
            );
        };

        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        // This is the main function that reacts to filter changes from PlayerFilters.vue
        const handleFilterChanged = async (newFilters) => {
            console.log(
                `PlayerUploadPage: Filters changed from PlayerFilters.vue:`,
                JSON.parse(JSON.stringify(newFilters)),
            );
            activeFilters.value = newFilters; // Store the latest filters

            if (playerStore.currentDatasetId) {
                // Fetch data from backend with position and role filters
                // The backend will return players already filtered by position and with Overall adjusted for role.
                await playerStore.fetchPlayersByDatasetId(
                    playerStore.currentDatasetId,
                    newFilters.position, // Pass selected position
                    newFilters.role, // Pass selected role
                );
                // After fetch, allPlayers in the store is updated.
                // The watcher on `allPlayers` will then apply the *remaining* client-side filters.
            } else {
                // If no dataset is loaded, apply client-side filters to an empty list (or current local list if any)
                // This scenario should ideally not happen if PlayerFilters is only shown when data is present.
                applyClientSideFilters(allPlayers.value, newFilters);
            }
        };

        // Watch for changes in the store's allPlayers list.
        // This list is updated by fetchPlayersByDatasetId (which might include backend pos/role filtering).
        // After it's updated, apply the remaining client-side filters.
        watch(
            allPlayers,
            (newVal) => {
                console.log(
                    `PlayerUploadPage: Watcher for allPlayers (from store) triggered. New length: ${newVal?.length}`,
                );
                // Apply client-side filters to the (potentially backend-filtered) list
                applyClientSideFilters(newVal, activeFilters.value);
            },
            { immediate: true },
        ); // immediate: true to run on initial load if allPlayers is already populated

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
            roleSpecificOverallWeightsLoadedForFeedback,
            showUpgradeFinder,
            isGoalkeeperView,
            goToTeamView,
            currentDatasetId,
            detectedCurrencySymbol,
            handleFilterChanged, // Make sure this is returned
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

.instructions-card ol {
    padding-left: 20px; /* Standard padding for ordered lists */
    li {
        margin-bottom: 0.5em; /* Space between list items */
    }
}

.upload-card,
.summary-card,
.no-data-card {
    border-radius: 8px; /* Consistent border radius, use $generic-border-radius if defined */
}

.summary-cards .q-card {
    height: 100%; /* Make summary cards in a row take full height of their container */
}
/* Add any other specific styles for PlayerUploadPage here */
</style>
