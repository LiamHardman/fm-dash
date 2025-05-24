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
                            <br />
                            <span
                                class="text-caption"
                                :class="
                                    $q.dark.isActive
                                        ? 'text-grey-5'
                                        : 'text-grey-7'
                                "
                            >
                                Maximum file size: 15MB (approx. 10,000
                                players).
                            </span>
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
                        <template v-slot:hint> Max file size: 15MB </template>
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
                <div class="row justify-between items-center q-mb-md q-mt-md">
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
                    :filtered-player-count="filteredPlayers.length"
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
import { useQuasar } from "quasar";
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
        const $q = useQuasar();
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

            if (
                currentFilters.ageRange &&
                typeof currentFilters.ageRange.min === "number" &&
                typeof currentFilters.ageRange.max === "number"
            ) {
                if (
                    currentFilters.ageRange.min >
                    playerStore.AGE_SLIDER_MIN_DEFAULT
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) => p.age >= currentFilters.ageRange.min,
                    );
                }
                if (
                    currentFilters.ageRange.max <
                    playerStore.AGE_SLIDER_MAX_DEFAULT
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) => p.age <= currentFilters.ageRange.max,
                    );
                }
            }

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
                // Clear any previous specific error notifications if successful
                if (!playerStore.error) {
                    $q.notify({
                        type: "positive",
                        message: "File uploaded and parsed successfully!",
                        position: "top",
                        timeout: 3000,
                    });
                }
            } catch (e) {
                console.error("Upload and Parse error in page:", e);
                // Error is already set in the store by uploadPlayerFile action
                // The q-banner will display it.
                // We can still show a Quasar notification for immediate feedback if desired,
                // especially for the 413 error, though the store now handles the message.
                if (
                    e.status === 413 ||
                    (e.message && e.message.includes("File too large"))
                ) {
                    $q.notify({
                        type: "negative",
                        message: playerStore.error, // Use the message from the store
                        position: "top",
                        timeout: 5000,
                        actions: [
                            {
                                label: "Dismiss",
                                color: "white",
                                handler: () => {
                                    /* ... */
                                },
                            },
                        ],
                    });
                } else if (playerStore.error) {
                    // For other errors set by the store
                    $q.notify({
                        type: "negative",
                        message: playerStore.error,
                        position: "top",
                        timeout: 5000,
                        actions: [
                            {
                                label: "Dismiss",
                                color: "white",
                                handler: () => {
                                    /* ... */
                                },
                            },
                        ],
                    });
                }
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

        return {
            $q,
            playerFile,
            playerStore,
            loading,
            error,
            allPlayers,
            filteredPlayers,
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
.no-data-card {
    border-radius: 8px;
}
</style>
