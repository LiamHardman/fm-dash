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
                            Media Handling, Personality, Age Range.
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
                            !roleSpecificOverallWeightsLoadedForFeedback ||
                            loading
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
                :is-loading="loading"
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
                        :disable="
                            allPlayers.length === 0 ||
                            !currentDatasetId ||
                            loading
                        "
                        class="q-px-lg"
                    />
                    <q-btn
                        color="secondary"
                        icon="find_replace"
                        label="Find Upgrades"
                        @click="showUpgradeFinder = true"
                        :disable="allPlayers.length === 0 || loading"
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
import { ref, computed, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { usePlayerStore } from "../stores/playerStore";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import UpgradeFinderDialog from "../components/UpgradeFinderDialog.vue";
import PlayerFilters from "../components/filters/PlayerFilters.vue";

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
        const roleSpecificOverallWeightsLoadedForFeedback = ref(false);

        const activeFilters = ref({});

        const isGoalkeeperView = computed(
            () => activeFilters.value.position === "GK",
        );

        const loadJsonForFeedback = async (filePath, loadedFlagRef) => {
            try {
                const response = await fetch(filePath);
                if (!response.ok)
                    throw new Error(`HTTP error! status: ${response.status}`);
                await response.json();
                loadedFlagRef.value = true;
            } catch (e) {
                console.warn(
                    `Client-side check: Failed to load ${filePath}:`,
                    e,
                );
                loadedFlagRef.value = true;
            }
        };

        onMounted(async () => {
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
                    activeFilters.value.position,
                    activeFilters.value.role,
                    activeFilters.value.ageRange,
                    activeFilters.value.transferValueRangeLocal,
                );
                if (playerStore.allAvailableRoles.length === 0) {
                    await playerStore.fetchAllAvailableRoles();
                }
            } else if (
                playerStore.currentDatasetId &&
                playerStore.allAvailableRoles.length === 0
            ) {
                await playerStore.fetchAllAvailableRoles();
            }
            if (allPlayers.value.length > 0) {
                applyClientSideFilters(allPlayers.value, activeFilters.value);
            }
        });

        const applyClientSideFilters = (playersToFilter, currentFilters) => {
            let tempPlayers = [...playersToFilter];

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
            // Age and Transfer Value range filters are now primarily handled by backend.
            // These client-side filters are only for additional refinement if needed,
            // or if backend filtering was not exhaustive for some edge cases.
            // For ageRange:
            if (
                currentFilters.ageRange &&
                typeof currentFilters.ageRange.min === "number" &&
                typeof currentFilters.ageRange.max === "number"
            ) {
                if (
                    currentFilters.ageRange.min >
                    playerStore.AGE_SLIDER_MIN_DEFAULT
                ) {
                    // Assuming AGE_SLIDER_MIN_DEFAULT is defined in store or locally
                    tempPlayers = tempPlayers.filter(
                        (p) => p.age >= currentFilters.ageRange.min,
                    );
                }
                if (
                    currentFilters.ageRange.max <
                    playerStore.AGE_SLIDER_MAX_DEFAULT
                ) {
                    // Assuming AGE_SLIDER_MAX_DEFAULT is defined in store or locally
                    tempPlayers = tempPlayers.filter(
                        (p) => p.age <= currentFilters.ageRange.max,
                    );
                }
            }

            // For transferValueRangeLocal:
            if (
                currentFilters.transferValueRangeLocal &&
                playerStore.transferValueRange &&
                typeof currentFilters.transferValueRangeLocal.min ===
                    "number" &&
                typeof currentFilters.transferValueRangeLocal.max === "number"
            ) {
                if (
                    currentFilters.transferValueRangeLocal.min >
                    playerStore.transferValueRange.min
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            (p.transferValueAmount || 0) >=
                            currentFilters.transferValueRangeLocal.min,
                    );
                }
                if (
                    currentFilters.transferValueRangeLocal.max <
                    playerStore.transferValueRange.max
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            (p.transferValueAmount || 0) <=
                            currentFilters.transferValueRangeLocal.max,
                    );
                }
            }
            filteredPlayers.value = tempPlayers;
        };

        const uploadAndParse = async () => {
            if (!playerFile.value) {
                playerStore.error = "Please select an HTML file first.";
                return;
            }
            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                await playerStore.uploadPlayerFile(formData);
                activeFilters.value = {};
            } catch (e) {
                console.error("Upload and Parse error in page:", e);
            }
        };

        const handleSort = (sortParams) => {
            console.log(
                "PlayerUploadPage: Sort requested by PlayerDataTable:",
                sortParams,
            );
        };

        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        const handleFilterChanged = async (newFilters) => {
            activeFilters.value = newFilters;
            if (playerStore.currentDatasetId) {
                await playerStore.fetchPlayersByDatasetId(
                    playerStore.currentDatasetId,
                    newFilters.position,
                    newFilters.role,
                    newFilters.ageRange,
                    newFilters.transferValueRangeLocal,
                );
            } else {
                applyClientSideFilters(allPlayers.value, newFilters);
            }
        };

        watch(
            allPlayers,
            (newVal) => {
                applyClientSideFilters(newVal, activeFilters.value);
            },
            { immediate: true },
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

        // Define default age slider values for comparison in applyClientSideFilters
        // These should match the initial values in PlayerFilters.vue
        playerStore.AGE_SLIDER_MIN_DEFAULT = 15; // Example, adjust if different
        playerStore.AGE_SLIDER_MAX_DEFAULT = 50; // Example, adjust if different

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
.instructions-card ol {
    padding-left: 20px;
    li {
        margin-bottom: 0.5em;
    }
}
.upload-card,
.summary-card,
.no-data-card {
    border-radius: 8px;
}
.summary-cards .q-card {
    height: 100%;
}
</style>
