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
                :initial-dataset-range="initialDatasetTransferRange"
                :salary-range="salaryRange"
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

        <q-dialog v-model="showFileSizeLimitModal" persistent>
            <q-card
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-white'
                        : 'bg-white text-dark'
                "
            >
                <q-card-section class="row items-center">
                    <q-avatar
                        icon="warning"
                        color="negative"
                        text-color="white"
                    />
                    <span class="q-ml-sm text-subtitle1">File Too Large</span>
                </q-card-section>

                <q-card-section class="q-pt-none">
                    Please ensure your HTML export contains 10,000 players or
                    less. (Max file size: 15MB)
                </q-card-section>

                <q-card-actions align="right">
                    <q-btn
                        flat
                        label="OK"
                        color="primary"
                        v-close-popup
                        @click="showFileSizeLimitModal = false"
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>
    </q-page>
</template>

<script>
import { ref, computed, onMounted, watch } from "vue";
import { useRouter, useRoute } from "vue-router"; // Added useRoute
import { useQuasar, Notify } from "quasar";
import { usePlayerStore } from "../stores/playerStore";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import UpgradeFinderDialog from "../components/UpgradeFinderDialog.vue";
import PlayerFilters from "../components/filters/PlayerFilters.vue";

const MAX_FILE_SIZE_BYTES = 15 * 1024 * 1024;

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
        const route = useRoute(); // Added useRoute
        const playerStore = usePlayerStore();
        const $q = useQuasar();
        const playerFile = ref(null);
        const filteredPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false);
        const showFileSizeLimitModal = ref(false);

        const initialDatasetTransferRange = ref({ min: 0, max: 100000000 }); // Default

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

        const salaryRange = computed(() => {
            if (!allPlayers.value || allPlayers.value.length === 0) {
                return { min: 0, max: 1000000 };
            }
            const salaries = allPlayers.value
                .filter((p) => typeof p.wageAmount === "number")
                .map((p) => p.wageAmount);
            if (salaries.length === 0) return { min: 0, max: 1000000 };
            const min = Math.min(...salaries);
            let max = Math.max(...salaries);
            if (min >= max) max = min + 10000; // Ensure max is always greater than min
            return { min, max };
        });

        const attributeWeightsLoadedForFeedback = ref(false);
        const roleSpecificOverallWeightsLoadedForFeedback = ref(false);

        const activeFilters = ref({});

        const isGoalkeeperView = computed(
            () => activeFilters.value.position === "GK",
        );

        const setInitialDatasetRange = () => {
            if (
                Array.isArray(playerStore.allPlayers) &&
                playerStore.allPlayers.length > 0
            ) {
                const values = playerStore.allPlayers
                    .filter((p) => typeof p.transferValueAmount === "number")
                    .map((p) => p.transferValueAmount);
                if (values.length > 0) {
                    const minVal = Math.min(0, ...values);
                    let maxVal = Math.max(...values);
                    if (minVal >= maxVal && values.length > 0)
                        maxVal = minVal + 50000; // Ensure max > min if values exist
                    else if (
                        minVal === 0 &&
                        maxVal === 0 &&
                        values.some((v) => v === 0)
                    )
                        maxVal = 50000;
                    initialDatasetTransferRange.value = {
                        min: minVal,
                        max: maxVal,
                    };
                } else {
                    // No numeric transfer values found
                    initialDatasetTransferRange.value = {
                        min: 0,
                        max: 100000000,
                    };
                }
            } else {
                // No players in store
                initialDatasetTransferRange.value = { min: 0, max: 100000000 };
            }
            console.log(
                "PlayerUploadPage: Initial dataset transfer range set:",
                JSON.parse(JSON.stringify(initialDatasetTransferRange.value)),
            );
        };

        const fetchAndSetInitialRange = async (datasetId) => {
            if (
                !playerStore.currentDatasetId ||
                playerStore.currentDatasetId !== datasetId ||
                playerStore.allPlayers.length === 0
            ) {
                await playerStore.fetchPlayersByDatasetId(datasetId); // This will populate allPlayers
            }
            if (
                playerStore.allAvailableRoles.length === 0 &&
                playerStore.currentDatasetId
            ) {
                await playerStore.fetchAllAvailableRoles();
            }
            setInitialDatasetRange(); // Set the range after allPlayers is populated
        };

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

            const storedDatasetId = sessionStorage.getItem("currentDatasetId");
            if (storedDatasetId) {
                playerStore.currentDatasetId = storedDatasetId;
                playerStore.detectedCurrencySymbol =
                    sessionStorage.getItem("detectedCurrencySymbol") || "$";
                await fetchAndSetInitialRange(storedDatasetId);
            } else {
                // If no dataset ID, ensure initial range is default
                setInitialDatasetRange();
            }

            if (allPlayers.value.length > 0) {
                applyClientSideFilters(allPlayers.value, activeFilters.value);
            }
        });

        watch(
            () => route.query.datasetId,
            async (newId, oldId) => {
                if (newId && newId !== oldId) {
                    sessionStorage.setItem("currentDatasetId", newId);
                    playerStore.currentDatasetId = newId;
                    await fetchAndSetInitialRange(newId);
                    activeFilters.value = {}; // Reset filters on dataset change
                }
            },
        );

        const applyClientSideFilters = (playersToFilter, currentFilters) => {
            if (!Array.isArray(playersToFilter)) {
                filteredPlayers.value = [];
                return;
            }

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
                        (p) =>
                            (parseInt(p.age, 10) || 0) >=
                            currentFilters.ageRange.min,
                    );
                }
                if (
                    currentFilters.ageRange.max <
                    playerStore.AGE_SLIDER_MAX_DEFAULT
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            (parseInt(p.age, 10) || 0) <=
                            currentFilters.ageRange.max,
                    );
                }
            }

            if (
                currentFilters.transferValueRangeLocal &&
                initialDatasetTransferRange.value && // Use the stable initial range for comparison logic
                typeof currentFilters.transferValueRangeLocal.min ===
                    "number" &&
                typeof currentFilters.transferValueRangeLocal.max === "number"
            ) {
                if (
                    currentFilters.transferValueRangeLocal.min >
                    initialDatasetTransferRange.value.min // Compare against true min
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            (p.transferValueAmount || 0) >=
                            currentFilters.transferValueRangeLocal.min,
                    );
                }
                if (
                    currentFilters.transferValueRangeLocal.max <
                    initialDatasetTransferRange.value.max // Compare against true max
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            (p.transferValueAmount || 0) <=
                            currentFilters.transferValueRangeLocal.max,
                    );
                }
            }

            // Max salary filter - only apply if not at max value (which means "Any")
            if (
                currentFilters.maxSalary !== null &&
                currentFilters.maxSalary !== undefined &&
                typeof currentFilters.maxSalary === "number" &&
                currentFilters.maxSalary < salaryRange.value.max
            ) {
                tempPlayers = tempPlayers.filter(
                    (p) => (p.wageAmount || 0) <= currentFilters.maxSalary,
                );
            }

            filteredPlayers.value = tempPlayers;
        };

        const uploadAndParse = async () => {
            if (!playerFile.value) {
                playerStore.error = "Please select an HTML file first.";
                return;
            }
            if (playerFile.value.size > MAX_FILE_SIZE_BYTES) {
                showFileSizeLimitModal.value = true;
                return;
            }
            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                await playerStore.uploadPlayerFile(formData); // This now fetches players and roles
                setInitialDatasetRange(); // Set the initial range AFTER allPlayers is populated by upload
                activeFilters.value = {};
                if (!playerStore.error) {
                    Notify.create({
                        type: "positive",
                        message: "File uploaded and parsed successfully!",
                        position: "top",
                        timeout: 3000,
                    });
                }
            } catch (e) {
                console.error("Upload and Parse error in page:", e);
                if (playerStore.error) {
                    Notify.create({
                        type: "negative",
                        message: playerStore.error,
                        position: "top",
                        timeout: 5000,
                        actions: [{ label: "Dismiss", color: "white" }],
                    });
                } else {
                    Notify.create({
                        type: "negative",
                        message: `Upload failed: ${e.message}`,
                        position: "top",
                        timeout: 5000,
                        actions: [{ label: "Dismiss", color: "white" }],
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
                // The store's fetchPlayersByDatasetId is called, which then populates allPlayers.
                // The watcher on allPlayers will then call applyClientSideFilters.
                // No need to directly call applyClientSideFilters here if fetch is happening.
                await playerStore.fetchPlayersByDatasetId(
                    playerStore.currentDatasetId,
                    newFilters.position,
                    newFilters.role,
                    newFilters.ageRange,
                    newFilters.transferValueRangeLocal,
                    newFilters.maxSalary,
                );
            } else {
                applyClientSideFilters(allPlayers.value, newFilters);
            }
        };

        watch(
            allPlayers,
            (newVal) => {
                // This watcher ensures client-side filtering is applied whenever allPlayers changes
                // (e.g., after initial load, after upload, or after backend filter fetch)
                applyClientSideFilters(newVal, activeFilters.value);
                if (
                    newVal &&
                    newVal.length > 0 &&
                    initialDatasetTransferRange.value.max === 100000000
                ) {
                    // If players are loaded and initial range hasn't been properly set, set it.
                    // This can happen if onMounted finishes before playerStore.allPlayers is populated from session.
                    setInitialDatasetRange();
                }
            },
            { immediate: true, deep: true }, // deep: true might be heavy if allPlayers is huge. Consider alternatives if performance issues.
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
            showFileSizeLimitModal,
            initialDatasetTransferRange, // Expose for PlayerFilters
            salaryRange, // Expose for PlayerFilters
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
    // Styling for the main page title if needed
}

.instructions-card {
    ol {
        padding-left: 20px;
        li {
            margin-bottom: 0.5em;
        }
    }
}

.upload-card,
.no-data-card {
    border-radius: 8px;
}

.body--dark .q-field__bottom {
    color: rgba(255, 255, 255, 0.5) !important;
}
</style>
