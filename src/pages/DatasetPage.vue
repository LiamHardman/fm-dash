<template>
    <q-page padding class="dataset-page">
        <div class="q-pa-md">
            <div class="row items-center justify-between q-mb-lg">
                <q-btn
                    v-if="currentDatasetId"
                    unelevated
                    icon="share"
                    label="Share Dataset"
                    color="positive"
                    @click="shareDataset"
                    class="share-btn-enhanced"
                    size="md"
                >
                    <q-tooltip>Copy shareable link to clipboard</q-tooltip>
                </q-btn>
            </div>

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
                    Loading dataset...
                </div>
            </div>

            <div v-if="!pageLoading && !pageLoadingError">
                <q-card
                    class="q-mb-md"
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-md">Dataset Overview</div>
                        <div class="row q-col-gutter-md">
                            <div class="col-12 col-sm-6 col-md-3">
                                <q-card flat bordered class="stats-card">
                                    <q-card-section class="text-center">
                                        <div class="text-h4 text-primary">
                                            {{ allPlayersData.length }}
                                        </div>
                                        <div class="text-subtitle2">
                                            Total Players
                                        </div>
                                    </q-card-section>
                                </q-card>
                            </div>
                            <div class="col-12 col-sm-6 col-md-3">
                                <q-card flat bordered class="stats-card">
                                    <q-card-section class="text-center">
                                        <div class="text-h4 text-secondary">
                                            {{ uniqueClubs.length }}
                                        </div>
                                        <div class="text-subtitle2">Teams</div>
                                    </q-card-section>
                                </q-card>
                            </div>
                            <div class="col-12 col-sm-6 col-md-3">
                                <q-card flat bordered class="stats-card">
                                    <q-card-section class="text-center">
                                        <div class="text-h4 text-accent">
                                            {{ uniqueNationalities.length }}
                                        </div>
                                        <div class="text-subtitle2">
                                            Nationalities
                                        </div>
                                    </q-card-section>
                                </q-card>
                            </div>
                            <div class="col-12 col-sm-6 col-md-3">
                                <q-card flat bordered class="stats-card">
                                    <q-card-section class="text-center">
                                        <div class="text-h4 text-positive">
                                            {{ detectedCurrencySymbol }}
                                        </div>
                                        <div class="text-subtitle2">
                                            Currency
                                        </div>
                                    </q-card-section>
                                </q-card>
                            </div>
                        </div>
                    </q-card-section>
                </q-card>

                <q-card
                    class="q-mb-md"
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-md">Quick Actions</div>
                        <div class="row q-col-gutter-md">
                            <div class="col-12 col-sm-4">
                                <q-btn
                                    color="primary"
                                    label="View All Players"
                                    icon="group"
                                    @click="viewAllPlayers"
                                    class="full-width"
                                    size="lg"
                                />
                            </div>
                            <div class="col-12 col-sm-4">
                                <q-btn
                                    color="secondary"
                                    label="Team Analysis"
                                    icon="sports_soccer"
                                    @click="viewTeamAnalysis"
                                    class="full-width"
                                    size="lg"
                                />
                            </div>
                            <div class="col-12 col-sm-4">
                                <q-btn
                                    color="accent"
                                    label="Find Upgrades"
                                    icon="find_replace"
                                    @click="showUpgradeFinder = true"
                                    :disable="allPlayersData.length === 0"
                                    class="full-width"
                                    size="lg"
                                />
                            </div>
                        </div>
                    </q-card-section>
                </q-card>

                <q-card
                    v-if="allPlayersData.length > 0"
                    class="q-mb-md"
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <PlayerFilters
                            @filter-changed="handleFiltersChanged"
                            :all-available-roles="allAvailableRoles"
                            :unique-clubs="uniqueClubs"
                            :unique-nationalities="uniqueNationalities"
                            :unique-media-handlings="uniqueMediaHandlings"
                            :unique-personalities="uniquePersonalities"
                            :transfer-value-range="transferValueRangeForFilters"
                            :initial-dataset-range="
                                initialDatasetTransferValueRangeForFilters
                            "
                            :salary-range="salaryRangeForFilters"
                            :currency-symbol="detectedCurrencySymbol"
                            :age-slider-min-default="AGE_SLIDER_MIN_DEFAULT"
                            :age-slider-max-default="AGE_SLIDER_MAX_DEFAULT"
                            :is-loading="loading"
                        />
                    </q-card-section>
                </q-card>

                <q-card
                    v-if="allPlayersData.length > 0"
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <div class="text-h6 q-mb-sm">
                            Players ({{ filteredPlayers.length }})
                        </div>
                        <PlayerDataTable
                            :players="filteredPlayers"
                            :loading="loading"
                            @player-selected="handlePlayerSelected"
                            @team-selected="handleTeamSelected"
                            :is-goalkeeper-view="isGoalkeeperView"
                            :currency-symbol="detectedCurrencySymbol"
                        />
                    </q-card-section>
                </q-card>

                <q-banner
                    v-else-if="!pageLoading && !pageLoadingError"
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
                    No player data found for this dataset.
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
        <UpgradeFinderDialog
            :show="showUpgradeFinder"
            :players="allPlayersData"
            @close="showUpgradeFinder = false"
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
import PlayerFilters from "../components/filters/PlayerFilters.vue";
import UpgradeFinderDialog from "../components/UpgradeFinderDialog.vue";

// Define FM attribute keys for filtering (raw keys as used in player.attributes)
const rawTechnicalAttributeKeysConst = [
    "Cor",
    "Cro",
    "Dri",
    "Fin",
    "Fir",
    "Fre",
    "Hea",
    "Lon",
    "L Th",
    "Mar",
    "Pas",
    "Pen",
    "Tck",
    "Tec",
];
const rawMentalAttributeKeysConst = [
    "Agg",
    "Ant",
    "Bra",
    "Cmp",
    "Cnt",
    "Dec",
    "Det",
    "Fla",
    "Ldr",
    "OtB",
    "Pos",
    "Tea",
    "Vis",
    "Wor",
];
const rawPhysicalAttributeKeysConst = [
    "Acc",
    "Agi",
    "Bal",
    "Jum",
    "Nat",
    "Pac",
    "Sta",
    "Str",
];
const rawGoalkeeperAttributeKeysConst = [
    "Aer",
    "Cmd",
    "Com",
    "Ecc",
    "Han",
    "Kic",
    "1v1",
    "Pun",
    "Ref",
    "TRO",
    "Thr",
];

const allRawFmAttributeKeys = [
    ...rawTechnicalAttributeKeysConst,
    ...rawMentalAttributeKeysConst,
    ...rawPhysicalAttributeKeysConst,
    ...rawGoalkeeperAttributeKeysConst,
];

// Helper to create filter keys like 'minCor', 'minLTh' (matching PlayerFilters.vue's formatAttrKey for consistency)
const formatFilterKeyPrefix = (attrKey) => {
    return attrKey.replace(/\s+/g, "").replace(/\(|\)/g, "");
};

export default {
    name: "DatasetPage",
    components: {
        PlayerDataTable,
        PlayerDetailDialog,
        PlayerFilters,
        UpgradeFinderDialog,
    },
    setup() {
        const quasarInstance = useQuasar();
        const router = useRouter();
        const route = useRoute();
        const playerStore = usePlayerStore();

        const pageLoading = ref(true);
        const pageLoadingError = ref("");
        const playerForDetailView = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false);

        // Centralized filter state for this page
        const currentFilters = ref({
            name: "",
            club: null,
            position: null,
            role: null,
            nationality: null,
            mediaHandling: [],
            personality: [],
            ageRange: {
                min: playerStore.AGE_SLIDER_MIN_DEFAULT,
                max: playerStore.AGE_SLIDER_MAX_DEFAULT,
            },
            transferValueRangeLocal: {
                min:
                    playerStore.initialDatasetTransferValueRange?.value?.min ||
                    0,
                max:
                    playerStore.initialDatasetTransferValueRange?.value?.max ||
                    100000000,
                userSet: false,
            },
            maxSalary: playerStore.salaryRange?.value?.max || 1000000,
            minOverall: 0,
            minPHY: 0,
            minSHO: 0,
            minPAS: 0,
            minDRI: 0,
            minDEF: 0,
            minMEN: 0,
            minGK: 0,
        });

        // Initialize FM attribute filters in currentFilters
        allRawFmAttributeKeys.forEach((attrKey) => {
            currentFilters.value[`min${formatFilterKeyPrefix(attrKey)}`] = 0;
        });

        // Computed properties from store
        const allPlayersData = computed(() => playerStore.allPlayers);
        const detectedCurrencySymbol = computed(
            () => playerStore.detectedCurrencySymbol,
        );
        const currentDatasetId = computed(() => playerStore.currentDatasetId);
        const loading = computed(() => playerStore.loading);
        const uniqueClubs = computed(() => playerStore.uniqueClubs);
        const uniqueNationalities = computed(
            () => playerStore.uniqueNationalities,
        );
        const uniqueMediaHandlings = computed(
            () => playerStore.uniqueMediaHandlings,
        );
        const uniquePersonalities = computed(
            () => playerStore.uniquePersonalities,
        );

        // For PlayerFilters component props - ensure safe access with fallbacks
        const transferValueRangeForFilters = computed(
            () =>
                playerStore.currentDatasetTransferValueRange.value || {
                    min: 0,
                    max: 100000000,
                },
        );
        const initialDatasetTransferValueRangeForFilters = computed(
            () =>
                playerStore.initialDatasetTransferValueRange.value || {
                    min: 0,
                    max: 100000000,
                },
        );
        const salaryRangeForFilters = computed(
            () => playerStore.salaryRange.value || { min: 0, max: 1000000 },
        );

        const allAvailableRoles = computed(() => playerStore.allAvailableRoles);
        const AGE_SLIDER_MIN_DEFAULT = computed(
            () => playerStore.AGE_SLIDER_MIN_DEFAULT,
        );
        const AGE_SLIDER_MAX_DEFAULT = computed(
            () => playerStore.AGE_SLIDER_MAX_DEFAULT,
        );

        const isGoalkeeperView = computed(() => {
            return (
                currentFilters.value.position === "GK" ||
                currentFilters.value.role?.includes("Goalkeeper")
            );
        });

        const filteredPlayers = computed(() => {
            if (!Array.isArray(allPlayersData.value)) return [];

            return allPlayersData.value
                .filter((player) => {
                    // Name filter
                    if (
                        currentFilters.value.name &&
                        !player.name
                            .toLowerCase()
                            .includes(currentFilters.value.name.toLowerCase())
                    ) {
                        return false;
                    }

                    // Club filter
                    if (
                        currentFilters.value.club &&
                        player.club !== currentFilters.value.club
                    ) {
                        return false;
                    }

                    // Position filter
                    if (currentFilters.value.position) {
                        const hasPosition = player.shortPositions?.includes(
                            currentFilters.value.position,
                        );
                        if (!hasPosition) return false;
                    }

                    // Role filter
                    if (currentFilters.value.role) {
                        const hasRole = player.roleSpecificOveralls?.some(
                            (role) =>
                                role.roleName === currentFilters.value.role,
                        );
                        if (!hasRole) return false;
                    }

                    // Nationality filter
                    if (
                        currentFilters.value.nationality &&
                        player.nationality !== currentFilters.value.nationality
                    ) {
                        return false;
                    }

                    // Media handling filter
                    if (
                        currentFilters.value.mediaHandling &&
                        currentFilters.value.mediaHandling.length > 0
                    ) {
                        if (!player.media_handling) return false;
                        const playerMediaHandlings = player.media_handling
                            .split(",")
                            .map((s) => s.trim());
                        const hasMediaHandling =
                            currentFilters.value.mediaHandling.some((filter) =>
                                playerMediaHandlings.includes(filter),
                            );
                        if (!hasMediaHandling) return false;
                    }

                    // Personality filter
                    if (
                        currentFilters.value.personality &&
                        currentFilters.value.personality.length > 0
                    ) {
                        if (!player.personality) return false;
                        const hasPersonality =
                            currentFilters.value.personality.includes(
                                player.personality,
                            );
                        if (!hasPersonality) return false;
                    }

                    // Age range filter
                    const playerAge = parseInt(player.age, 10) || 0;
                    if (
                        playerAge < currentFilters.value.ageRange.min ||
                        playerAge > currentFilters.value.ageRange.max
                    ) {
                        return false;
                    }

                    // Transfer value range filter
                    if (
                        player.transferValueAmount <
                            currentFilters.value.transferValueRangeLocal.min ||
                        player.transferValueAmount >
                            currentFilters.value.transferValueRangeLocal.max
                    ) {
                        return false;
                    }

                    // Max salary filter
                    if (
                        currentFilters.value.maxSalary !== null &&
                        player.wageAmount > currentFilters.value.maxSalary
                    ) {
                        return false;
                    }

                    // FIFA-style stat minimum filters
                    if (
                        currentFilters.value.minOverall > 0 &&
                        (player.Overall || 0) < currentFilters.value.minOverall
                    )
                        return false;
                    if (
                        currentFilters.value.minPHY > 0 &&
                        (player.PHY || 0) < currentFilters.value.minPHY
                    )
                        return false;
                    if (
                        currentFilters.value.minSHO > 0 &&
                        (player.SHO || 0) < currentFilters.value.minSHO
                    )
                        return false;
                    if (
                        currentFilters.value.minPAS > 0 &&
                        (player.PAS || 0) < currentFilters.value.minPAS
                    )
                        return false;
                    if (
                        currentFilters.value.minDRI > 0 &&
                        (player.DRI || 0) < currentFilters.value.minDRI
                    )
                        return false;
                    if (
                        currentFilters.value.minDEF > 0 &&
                        (player.DEF || 0) < currentFilters.value.minDEF
                    )
                        return false;
                    if (
                        currentFilters.value.minMEN > 0 &&
                        (player.MEN || 0) < currentFilters.value.minMEN
                    )
                        return false;
                    if (
                        currentFilters.value.minGK > 0 &&
                        (player.GK || 0) < currentFilters.value.minGK
                    )
                        return false;

                    // FM Attribute minimum filters
                    for (const rawAttrKey of allRawFmAttributeKeys) {
                        const filterKeyForVal = `min${formatFilterKeyPrefix(rawAttrKey)}`;
                        const minVal = currentFilters.value[filterKeyForVal];

                        if (minVal > 0) {
                            const playerAttrStr = player.attributes[rawAttrKey]; // FM attributes are in player.attributes as strings
                            const playerAttrVal =
                                parseInt(playerAttrStr, 10) || 0;
                            if (playerAttrVal < minVal) {
                                return false;
                            }
                        }
                    }

                    return true;
                })
                .map((player) => {
                    if (
                        currentFilters.value.role &&
                        player.roleSpecificOveralls
                    ) {
                        let roleSpecificOverall = null;
                        if (Array.isArray(player.roleSpecificOveralls)) {
                            const roleMatch = player.roleSpecificOveralls.find(
                                (rso) =>
                                    rso.roleName === currentFilters.value.role,
                            );
                            if (roleMatch)
                                roleSpecificOverall = roleMatch.score;
                        } else if (
                            typeof player.roleSpecificOveralls === "object"
                        ) {
                            roleSpecificOverall =
                                player.roleSpecificOveralls[
                                    currentFilters.value.role
                                ];
                        }

                        if (
                            roleSpecificOverall !== null &&
                            roleSpecificOverall !== undefined
                        ) {
                            return { ...player, Overall: roleSpecificOverall };
                        }
                    }
                    return player;
                });
        });

        const fetchDataset = async (datasetId) => {
            pageLoading.value = true;
            pageLoadingError.value = "";
            try {
                await playerStore.fetchPlayersByDatasetId(datasetId);
                await playerStore.fetchAllAvailableRoles();

                // Safely access store values and provide defaults
                const initTvRange =
                    playerStore.initialDatasetTransferValueRange.value;
                const initSalaryRange = playerStore.salaryRange.value;

                currentFilters.value.transferValueRangeLocal = {
                    min:
                        initTvRange && initTvRange.min !== undefined
                            ? initTvRange.min
                            : 0,
                    max:
                        initTvRange && initTvRange.max !== undefined
                            ? initTvRange.max
                            : 100000000,
                    userSet: false,
                };
                currentFilters.value.maxSalary =
                    initSalaryRange && initSalaryRange.max !== undefined
                        ? initSalaryRange.max
                        : 1000000;
            } catch (err) {
                pageLoadingError.value = `Failed to load dataset: ${err.message || "Unknown server error"}.`;
                playerStore.resetState();
            } finally {
                pageLoading.value = false;
            }
        };

        onMounted(async () => {
            const datasetIdFromRoute = route.params.datasetId;
            if (datasetIdFromRoute) {
                await fetchDataset(datasetIdFromRoute);
            } else {
                pageLoadingError.value = "No dataset ID provided in URL.";
                pageLoading.value = false;
            }
        });

        const shareDataset = async () => {
            if (!currentDatasetId.value) return;
            const shareUrl = `${window.location.origin}/dataset/${currentDatasetId.value}`;
            try {
                await navigator.clipboard.writeText(shareUrl);
                quasarInstance.notify({
                    message: "Dataset link copied to clipboard!",
                    color: "positive",
                    icon: "check_circle",
                    position: "top",
                    timeout: 2000,
                });
            } catch (err) {
                const textArea = document.createElement("textarea");
                textArea.value = shareUrl;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand("copy");
                document.body.removeChild(textArea);
                quasarInstance.notify({
                    message: "Dataset link copied to clipboard!",
                    color: "positive",
                    icon: "check_circle",
                    position: "top",
                    timeout: 2000,
                });
            }
        };

        const viewAllPlayers = () => {
            /* Already on this page */
        };
        const viewTeamAnalysis = () => {
            if (currentDatasetId.value) {
                router.push(`/team-view?datasetId=${currentDatasetId.value}`);
            }
        };

        const handlePlayerSelected = (player) => {
            playerForDetailView.value = player;
            showPlayerDetailDialog.value = true;
        };

        const handleTeamSelected = (teamName) => {
            if (currentDatasetId.value) {
                const url = router.resolve({
                    path: "/team-view",
                    query: {
                        datasetId: currentDatasetId.value,
                        team: teamName,
                    },
                }).href;
                window.open(url, "_blank");
            }
        };

        const handleFiltersChanged = (filtersFromChild) => {
            const newTransferRange = filtersFromChild.transferValueRangeLocal;
            const oldTransferRange =
                currentFilters.value.transferValueRangeLocal;

            currentFilters.value = {
                ...currentFilters.value,
                ...filtersFromChild,
            };

            if (
                newTransferRange &&
                oldTransferRange &&
                (newTransferRange.min !== oldTransferRange.min ||
                    newTransferRange.max !== oldTransferRange.max)
            ) {
                currentFilters.value.transferValueRangeLocal.userSet = true;
            }
        };

        watch(
            () => route.params.datasetId,
            async (newId, oldId) => {
                if (newId && newId !== oldId) {
                    await fetchDataset(newId);
                }
            },
        );

        watch(
            () => playerStore.initialDatasetTransferValueRange.value,
            (newRange) => {
                if (
                    newRange &&
                    !currentFilters.value.transferValueRangeLocal.userSet
                ) {
                    // Check userSet flag
                    currentFilters.value.transferValueRangeLocal = {
                        min: newRange.min !== undefined ? newRange.min : 0,
                        max:
                            newRange.max !== undefined
                                ? newRange.max
                                : 100000000,
                        userSet: false, // Keep userSet as false when programmatically updating
                    };
                }
            },
            { immediate: true, deep: true },
        );

        watch(
            () => playerStore.salaryRange.value,
            (newRange) => {
                const currentMaxSalary = currentFilters.value.maxSalary;
                const storeMaxSalary = playerStore.salaryRange.value?.max;

                // Only update if maxSalary hasn't been manually set by user OR is still at its initial large default OR matches the current store max
                if (
                    newRange &&
                    (currentMaxSalary === 1000000 ||
                        currentMaxSalary === null ||
                        currentMaxSalary === storeMaxSalary)
                ) {
                    if (newRange.max !== undefined) {
                        currentFilters.value.maxSalary = newRange.max;
                    } else {
                        currentFilters.value.maxSalary = 1000000;
                    }
                }
            },
            { immediate: true, deep: true },
        );

        return {
            pageLoading,
            pageLoadingError,
            allPlayersData,
            detectedCurrencySymbol,
            currentDatasetId,
            loading,
            uniqueClubs,
            uniqueNationalities,
            uniqueMediaHandlings,
            uniquePersonalities,
            transferValueRangeForFilters,
            initialDatasetTransferValueRangeForFilters,
            salaryRangeForFilters,
            allAvailableRoles,
            AGE_SLIDER_MIN_DEFAULT,
            AGE_SLIDER_MAX_DEFAULT,
            isGoalkeeperView,
            filteredPlayers,
            playerForDetailView,
            showPlayerDetailDialog,
            showUpgradeFinder,
            shareDataset,
            viewAllPlayers,
            viewTeamAnalysis,
            handlePlayerSelected,
            handleTeamSelected,
            handleFiltersChanged,
            quasarInstance,
            router,
            currentFilters,
        };
    },
};
</script>

<style lang="scss" scoped>
.dataset-page {
    max-width: 1600px;
    margin: 0 auto;
}

.page-title {
    margin: 0;
}

.share-btn-enhanced {
    font-weight: 600;
    border-radius: 6px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: all 0.2s ease;
    min-width: 140px;
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
    }
    
    .body--dark & {
        box-shadow: 0 2px 4px rgba(255, 255, 255, 0.1);
        
        &:hover {
            box-shadow: 0 4px 8px rgba(255, 255, 255, 0.15);
        }
    }
}

.stats-card {
    height: 100%;
    transition: transform 0.2s ease;

    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }

    .body--dark & {
        background-color: rgba(255, 255, 255, 0.05);

        &:hover {
            box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
        }
    }
}

.q-card {
    border-radius: $generic-border-radius;
}
</style>
