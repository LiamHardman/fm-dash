<template>
    <q-page padding class="dataset-page">
        <div class="q-pa-md">
            <div class="row items-center justify-between q-mb-lg">
                <q-btn
                    v-if="currentDatasetId"
                    flat
                    round
                    icon="share"
                    color="primary"
                    @click="shareDataset"
                    class="share-btn"
                    size="md"
                >
                    <q-tooltip>Share this dataset</q-tooltip>
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
                <!-- Dataset Overview Card -->
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

                <!-- Quick Actions Card -->
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

                <!-- Player Filters Card -->
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
                            :transfer-value-range="transferValueRange"
                            :currency-symbol="detectedCurrencySymbol"
                            :age-slider-min-default="AGE_SLIDER_MIN_DEFAULT"
                            :age-slider-max-default="AGE_SLIDER_MAX_DEFAULT"
                        />
                    </q-card-section>
                </q-card>

                <!-- Player Data Table -->
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

        // Filter states
        const nameFilter = ref("");
        const clubFilter = ref(null);
        const positionFilter = ref(null);
        const roleFilter = ref(null);
        const nationalityFilter = ref(null);
        const mediaHandlingFilter = ref([]);
        const personalityFilter = ref([]);
        const ageRangeFilter = ref({ min: 15, max: 50 });
        const transferValueRangeFilter = ref({ min: 0, max: 100000000 });
        const maxSalaryFilter = ref(null);
        // FIFA-style stat minimum filters
        const minOverallFilter = ref(0);
        const minPHYFilter = ref(0);
        const minSHOFilter = ref(0);
        const minPASFilter = ref(0);
        const minDRIFilter = ref(0);
        const minDEFFilter = ref(0);
        const minMENFilter = ref(0);
        const minGKFilter = ref(0);

        // FM Attribute filters - Technical
        const technicalAttributeKeys = [
            'crossing', 'dribbling', 'finishing', 'first_touch', 'free_kick_taking',
            'heading', 'long_shots', 'long_throws', 'marking', 'passing', 'penalty_taking',
            'tackling', 'technique', 'corners'
        ];

        // FM Attribute filters - Mental
        const mentalAttributeKeys = [
            'aggression', 'anticipation', 'bravery', 'composure', 'concentration',
            'decisions', 'determination', 'flair', 'leadership', 'off_the_ball',
            'positioning', 'teamwork', 'vision', 'work_rate'
        ];

        // FM Attribute filters - Physical
        const physicalAttributeKeys = [
            'acceleration', 'agility', 'balance', 'jumping_reach', 'natural_fitness',
            'pace', 'stamina', 'strength'
        ];

        // FM Attribute filters - Goalkeeping
        const goalkeeperAttributeKeys = [
            'aerial_reach', 'command_of_area', 'communication', 'eccentricity',
            'handling', 'kicking', 'one_on_ones', 'punching', 'reflexes',
            'rushing_out', 'tendency_to_punch', 'throwing'
        ];

        // Helper function to format attribute keys
        const formatAttrKey = (attr) => {
            return attr
                .split('_')
                .map((word, index) => index === 0 ? word.charAt(0).toUpperCase() + word.slice(1) : word.charAt(0).toUpperCase() + word.slice(1))
                .join('');
        };

        // All attribute keys combined
        const allAttributeKeys = [
            ...technicalAttributeKeys,
            ...mentalAttributeKeys,
            ...physicalAttributeKeys,
            ...goalkeeperAttributeKeys
        ];

        // Create reactive filters for all attributes
        const attributeFilters = {};
        allAttributeKeys.forEach(attr => {
            const filterKey = `min${formatAttrKey(attr)}Filter`;
            attributeFilters[filterKey] = ref(0);
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
        const transferValueRange = computed(
            () => playerStore.transferValueRange,
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
                positionFilter.value === "GK" ||
                roleFilter.value?.includes("Goalkeeper")
            );
        });

        const filteredPlayers = computed(() => {
            if (!Array.isArray(allPlayersData.value)) return [];

            return allPlayersData.value
                .filter((player) => {
                    // Name filter
                    if (
                        nameFilter.value &&
                        !player.name
                            .toLowerCase()
                            .includes(nameFilter.value.toLowerCase())
                    ) {
                        return false;
                    }

                    // Club filter
                    if (clubFilter.value && player.club !== clubFilter.value) {
                        return false;
                    }

                    // Position filter
                    if (positionFilter.value) {
                        const hasPosition = player.shortPositions?.includes(
                            positionFilter.value,
                        );
                        if (!hasPosition) return false;
                    }

                    // Role filter
                    if (roleFilter.value) {
                        const hasRole = player.roleSpecificOveralls?.some(
                            (role) => role.roleName === roleFilter.value,
                        );
                        if (!hasRole) return false;
                    }

                    // Nationality filter
                    if (
                        nationalityFilter.value &&
                        player.nationality !== nationalityFilter.value
                    ) {
                        return false;
                    }

                    // Media handling filter
                    if (
                        mediaHandlingFilter.value &&
                        mediaHandlingFilter.value.length > 0
                    ) {
                        if (!player.media_handling) return false;
                        const playerMediaHandlings = player.media_handling
                            .split(",")
                            .map((s) => s.trim());
                        const hasMediaHandling = mediaHandlingFilter.value.some(
                            (filter) => playerMediaHandlings.includes(filter),
                        );
                        if (!hasMediaHandling) return false;
                    }

                    // Personality filter
                    if (
                        personalityFilter.value &&
                        personalityFilter.value.length > 0
                    ) {
                        if (!player.personality) return false;
                        const hasPersonality = personalityFilter.value.includes(
                            player.personality,
                        );
                        if (!hasPersonality) return false;
                    }

                    // Age range filter
                    const playerAge = parseInt(player.age, 10) || 0;
                    if (
                        playerAge < ageRangeFilter.value.min ||
                        playerAge > ageRangeFilter.value.max
                    ) {
                        return false;
                    }

                    // Transfer value range filter
                    if (
                        player.transferValueAmount <
                            transferValueRangeFilter.value.min ||
                        player.transferValueAmount >
                            transferValueRangeFilter.value.max
                    ) {
                        return false;
                    }

                    // Max salary filter
                    if (
                        maxSalaryFilter.value !== null &&
                        player.wageAmount > maxSalaryFilter.value
                    ) {
                        return false;
                    }

                    // FIFA-style stat minimum filters
                    if (minOverallFilter.value > 0 && (player.Overall || 0) < minOverallFilter.value) {
                        return false;
                    }
                    if (minPHYFilter.value > 0 && (player.PHY || 0) < minPHYFilter.value) {
                        return false;
                    }
                    if (minSHOFilter.value > 0 && (player.SHO || 0) < minSHOFilter.value) {
                        return false;
                    }
                    if (minPASFilter.value > 0 && (player.PAS || 0) < minPASFilter.value) {
                        return false;
                    }
                    if (minDRIFilter.value > 0 && (player.DRI || 0) < minDRIFilter.value) {
                        return false;
                    }
                    if (minDEFFilter.value > 0 && (player.DEF || 0) < minDEFFilter.value) {
                        return false;
                    }
                    if (minMENFilter.value > 0 && (player.MEN || 0) < minMENFilter.value) {
                        return false;
                    }
                    if (minGKFilter.value > 0 && (player.GK || 0) < minGKFilter.value) {
                        return false;
                    }

                    // FM Attribute minimum filters
                    for (const attr of allAttributeKeys) {
                        const filterKey = `min${formatAttrKey(attr)}Filter`;
                        const attributeFilter = attributeFilters[filterKey];
                        if (attributeFilter && attributeFilter.value > 0) {
                            const playerAttrValue = parseInt(player[attr], 10) || 0;
                            if (playerAttrValue < attributeFilter.value) {
                                return false;
                            }
                        }
                    }

                    return true;
                })
                .map((player) => {
                    // If a role is selected, modify the player's overall to show role-specific rating
                    if (roleFilter.value && player.roleSpecificOveralls) {
                        // Debug logging - let's see what we're working with
                        console.log("=== ROLE FILTER DEBUG ===");
                        console.log("Player name:", player.name);
                        console.log(
                            "Player Overall (capital O):",
                            player.Overall,
                        );
                        console.log(
                            "Player overall (lowercase o):",
                            player.overall,
                        );
                        console.log("Selected role filter:", roleFilter.value);
                        console.log(
                            "roleSpecificOveralls type:",
                            typeof player.roleSpecificOveralls,
                        );
                        console.log(
                            "roleSpecificOveralls is array:",
                            Array.isArray(player.roleSpecificOveralls),
                        );
                        console.log(
                            "roleSpecificOveralls content:",
                            player.roleSpecificOveralls,
                        );

                        let roleSpecificOverall = null;

                        // Handle both array and object formats (as seen in TeamViewPage)
                        if (Array.isArray(player.roleSpecificOveralls)) {
                            console.log("Processing as array...");
                            // Array format: [{roleName: "DM - Anchor", score: 78}, ...]
                            const roleMatch = player.roleSpecificOveralls.find(
                                (rso) => rso.roleName === roleFilter.value,
                            );
                            console.log("Role match found:", roleMatch);
                            if (roleMatch) {
                                roleSpecificOverall = roleMatch.score;
                                console.log(
                                    "Role-specific overall from array:",
                                    roleSpecificOverall,
                                );
                            }
                        } else if (
                            typeof player.roleSpecificOveralls === "object"
                        ) {
                            console.log("Processing as object...");
                            // Object format: {"DM - Anchor": 78, "DM - Deep Lying Playmaker": 76, ...}
                            console.log(
                                "Available roles:",
                                Object.keys(player.roleSpecificOveralls),
                            );
                            roleSpecificOverall =
                                player.roleSpecificOveralls[roleFilter.value];
                            console.log(
                                "Role-specific overall from object:",
                                roleSpecificOverall,
                            );
                        }

                        // If we found a role-specific overall, use it
                        if (
                            roleSpecificOverall !== null &&
                            roleSpecificOverall !== undefined
                        ) {
                            console.log(
                                "✅ Using role-specific overall:",
                                roleSpecificOverall,
                            );
                            return {
                                ...player,
                                Overall: roleSpecificOverall, // Capital O to match PlayerDataTable
                            };
                        } else {
                            console.log(
                                "❌ No role match found, using original overall:",
                                player.Overall,
                            );
                        }
                        console.log("=== END DEBUG ===");
                    }
                    // Return original player if no role filter or no role match
                    return player;
                });
        });

        const fetchDataset = async (datasetId) => {
            pageLoading.value = true;
            pageLoadingError.value = "";
            try {
                await playerStore.fetchPlayersByDatasetId(datasetId);
                await playerStore.fetchAllAvailableRoles();
            } catch (err) {
                pageLoadingError.value = `Failed to load dataset: ${err.message || "Unknown server error"}.`;
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
                // Fallback for older browsers
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
            // Already on the dataset page showing all players
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
                // Open in new tab (since user requested new tab functionality)
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

        const handleFiltersChanged = (filters) => {
            nameFilter.value = filters.name;
            clubFilter.value = filters.club;
            positionFilter.value = filters.position;
            roleFilter.value = filters.role;
            nationalityFilter.value = filters.nationality;
            mediaHandlingFilter.value = filters.mediaHandling;
            personalityFilter.value = filters.personality;
            ageRangeFilter.value = filters.ageRange;
            transferValueRangeFilter.value = filters.transferValueRangeLocal;
            maxSalaryFilter.value = filters.maxSalary;
            // FIFA-style stat minimum filters
            minOverallFilter.value = filters.minOverall;
            minPHYFilter.value = filters.minPHY;
            minSHOFilter.value = filters.minSHO;
            minPASFilter.value = filters.minPAS;
            minDRIFilter.value = filters.minDRI;
            minDEFFilter.value = filters.minDEF;
            minMENFilter.value = filters.minMEN;
            minGKFilter.value = filters.minGK;
            
            // FM Attribute minimum filters
            allAttributeKeys.forEach(attr => {
                const filterKey = `min${formatAttrKey(attr)}Filter`;
                const filtersKey = `min${formatAttrKey(attr)}`;
                if (attributeFilters[filterKey] && filters[filtersKey] !== undefined) {
                    attributeFilters[filterKey].value = filters[filtersKey];
                }
            });
        };

        watch(
            () => route.params.datasetId,
            async (newId, oldId) => {
                if (newId && newId !== oldId) {
                    await fetchDataset(newId);
                }
            },
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
            transferValueRange,
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

.share-btn {
    min-width: 48px;
    min-height: 48px;
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
