<template>
    <q-dialog
        :model-value="show"
        @update:model-value="$emit('close')"
        persistent
        maximized
        transition-show="slide-up"
        transition-hide="slide-down"
    >
        <q-card
            class="upgrade-finder-dialog"
            :class="$q.dark.isActive ? 'bg-dark' : 'bg-grey-1'"
        >
            <q-card-section
                class="row items-center q-pb-none"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-10 text-white'
                        : 'bg-primary text-white'
                "
            >
                <q-icon name="manage_search" size="md" class="q-mr-sm" />
                <div class="text-h6">Upgrade Finder</div>
                <q-space />
                <q-btn
                    icon="close"
                    flat
                    round
                    dense
                    v-close-popup
                    @click="$emit('close')"
                />
            </q-card-section>

            <q-card-section class="q-pt-md">
                <div class="row q-col-gutter-x-md q-col-gutter-y-sm q-mb-md">
                    <div class="col-12 col-md-6 col-lg-3">
                        <q-select
                            v-model="teamName"
                            :options="teamOptions"
                            label="Team Name"
                            outlined
                            dense
                            use-input
                            hide-selected
                            fill-input
                            input-debounce="300"
                            @filter="filterTeams"
                            :rules="[(val) => !!val || 'Team name is required']"
                            clearable
                            @clear="
                                teamName = null;
                                selectedTeamPlayer = null;
                                teamPlayersForSelection = [];
                            "
                            :label-color="$q.dark.isActive ? 'grey-4' : ''"
                            :input-class="$q.dark.isActive ? 'text-grey-3' : ''"
                            :popup-content-class="
                                $q.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        >
                            <template v-slot:no-option>
                                <q-item>
                                    <q-item-section class="text-grey">
                                        No results
                                    </q-item-section>
                                </q-item>
                            </template>
                        </q-select>
                    </div>

                    <div class="col-12 col-md-6 col-lg-3">
                        <q-select
                            v-model="selectedPosition"
                            :options="positionFilterOptions"
                            label="Position / Group"
                            dense
                            outlined
                            emit-value
                            map-options
                            :rules="[(val) => !!val || 'Position is required']"
                            clearable
                            @clear="
                                selectedPosition = null;
                                selectedTeamPlayer = null;
                                teamPlayersForSelection = [];
                            "
                            :label-color="$q.dark.isActive ? 'grey-4' : ''"
                            :popup-content-class="
                                $q.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        />
                    </div>

                    <div class="col-12 col-md-6 col-lg-3">
                        <q-select
                            v-model="selectedTeamPlayer"
                            :options="teamPlayersForSelection"
                            label="Select Player for Upgrade Base"
                            option-label="name"
                            option-value="name"
                            map-options
                            emit-value
                            dense
                            outlined
                            clearable
                            :disable="
                                !teamName ||
                                !selectedPosition ||
                                teamPlayersForSelection.length === 0
                            "
                            :hint="
                                selectedTeamPlayer
                                    ? `Base Overall: ${getBaseOverallFromSelectedPlayer()}`
                                    : 'Select a player to set base overall'
                            "
                            :label-color="$q.dark.isActive ? 'grey-4' : ''"
                            :popup-content-class="
                                $q.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        >
                            <template v-slot:option="scope">
                                <q-item
                                    v-bind="scope.itemProps"
                                    :dark="$q.dark.isActive"
                                >
                                    <q-item-section>
                                        <q-item-label>{{
                                            scope.opt.name
                                        }}</q-item-label>
                                        <q-item-label caption
                                            >Overall:
                                            {{
                                                scope.opt.Overall
                                            }}</q-item-label
                                        >
                                    </q-item-section>
                                </q-item>
                            </template>
                            <template v-slot:no-option>
                                <q-item :dark="$q.dark.isActive">
                                    <q-item-section class="text-grey">
                                        {{
                                            teamName && selectedPosition
                                                ? "No players in this team/position"
                                                : "Select team and position first"
                                        }}
                                    </q-item-section>
                                </q-item>
                            </template>
                        </q-select>
                    </div>

                    <div class="col-12 col-md-6 col-lg-3">
                        <div>
                            <div
                                class="text-subtitle2 q-mb-xs"
                                :class="
                                    $q.dark.isActive
                                        ? 'text-grey-4'
                                        : 'text-grey-8'
                                "
                            >
                                Upgrade By: {{ upgradeByValue }}
                            </div>
                            <q-slider
                                v-model="upgradeByValue"
                                :min="-10"
                                :max="10"
                                :step="1"
                                label
                                label-always
                                color="primary"
                                :dark="$q.dark.isActive"
                                :disable="!selectedTeamPlayer"
                            />
                        </div>
                    </div>
                </div>

                <div class="row q-col-gutter-x-md q-col-gutter-y-sm q-mb-md">
                    <div class="col-12 col-md-6 col-lg-3">
                        <div
                            class="q-mb-xs text-subtitle2"
                            :class="
                                $q.dark.isActive ? 'text-grey-4' : 'text-grey-8'
                            "
                        >
                            Maximum Age:
                            {{
                                maxAgeFilter === ageSliderMax
                                    ? "Any"
                                    : maxAgeFilter
                            }}
                            <q-btn
                                flat
                                dense
                                icon="clear"
                                size="sm"
                                @click="maxAgeFilter = ageSliderMax"
                                v-if="maxAgeFilter < ageSliderMax"
                                class="q-ml-xs"
                                round
                                :text-color="
                                    $q.dark.isActive ? 'grey-5' : 'grey-7'
                                "
                            >
                                <q-tooltip>Clear age filter (Any)</q-tooltip>
                            </q-btn>
                        </div>
                        <q-slider
                            v-model="maxAgeFilter"
                            :min="ageSliderMin"
                            :max="ageSliderMax"
                            :step="1"
                            label
                            color="secondary"
                            :dark="$q.dark.isActive"
                        />
                    </div>

                    <div class="col-12 col-md-6 col-lg-3">
                        <div
                            class="q-mb-xs text-subtitle2"
                            :class="
                                $q.dark.isActive ? 'text-grey-4' : 'text-grey-8'
                            "
                        >
                            Max Transfer Value:
                            {{
                                maxTransferValueFilter ===
                                computedMaxSliderTransferValue
                                    ? "Any"
                                    : formattedMaxTransferValueLabel
                            }}
                            <q-btn
                                flat
                                dense
                                icon="clear"
                                size="sm"
                                @click="
                                    maxTransferValueFilter =
                                        computedMaxSliderTransferValue
                                "
                                v-if="
                                    maxTransferValueFilter <
                                        computedMaxSliderTransferValue &&
                                    props.players &&
                                    props.players.length > 0
                                "
                                class="q-ml-xs"
                                round
                                :text-color="
                                    $q.dark.isActive ? 'grey-5' : 'grey-7'
                                "
                            >
                                <q-tooltip>Clear value filter (Any)</q-tooltip>
                            </q-btn>
                        </div>
                        <q-slider
                            v-model="maxTransferValueFilter"
                            :min="computedMinSliderTransferValue"
                            :max="computedMaxSliderTransferValue"
                            :step="computedStepSliderTransferValue"
                            label
                            :label-value="formattedMaxTransferValueLabel"
                            color="secondary"
                            :dark="$q.dark.isActive"
                            :disable="
                                !props.players || props.players.length === 0
                            "
                        />
                    </div>
                </div>

                <div class="row q-col-gutter-md">
                    <div class="col-12">
                        <q-btn
                            color="primary"
                            icon="search"
                            label="Find Upgrades"
                            class="full-width q-py-sm"
                            @click="findUpgrades"
                            :loading="loading"
                            :disable="
                                !teamName ||
                                !selectedPosition ||
                                !selectedTeamPlayer ||
                                loading
                            "
                        />
                    </div>
                </div>
            </q-card-section>

            <q-card-section v-if="showResults" class="q-mt-md results-section">
                <q-separator :dark="$q.dark.isActive" class="q-mb-md" />
                <div
                    class="text-h6 q-mb-md"
                    :class="$q.dark.isActive ? 'text-grey-2' : 'text-grey-9'"
                >
                    Results
                </div>

                <div v-if="selectedTeamPlayerObject" class="q-mb-lg">
                    <div
                        class="text-subtitle1 q-mb-sm"
                        :class="
                            $q.dark.isActive ? 'text-grey-3' : 'text-grey-8'
                        "
                    >
                        Baseline Player:
                    </div>
                    <q-card
                        flat
                        bordered
                        :class="
                            $q.dark.isActive
                                ? 'bg-grey-8 text-grey-3'
                                : 'bg-blue-grey-1 text-blue-grey-10'
                        "
                    >
                        <q-item>
                            <q-item-section avatar>
                                <q-avatar>
                                    <img
                                        v-if="
                                            selectedTeamPlayerObject.nationality_iso
                                        "
                                        :src="`https://flagcdn.com/w40/${selectedTeamPlayerObject.nationality_iso.toLowerCase()}.png`"
                                        :alt="
                                            selectedTeamPlayerObject.nationality ||
                                            'Flag'
                                        "
                                        class="nationality-flag-dialog"
                                    />
                                    <q-icon v-else name="person" />
                                </q-avatar>
                            </q-item-section>
                            <q-item-section>
                                <q-item-label class="text-weight-bold">{{
                                    selectedTeamPlayerObject.name
                                }}</q-item-label>
                                <q-item-label
                                    caption
                                    :class="
                                        $q.dark.isActive
                                            ? 'text-grey-5'
                                            : 'text-blue-grey-7'
                                    "
                                >
                                    {{ selectedTeamPlayerObject.position }} |
                                    Age: {{ selectedTeamPlayerObject.age }} |
                                    Club: {{ selectedTeamPlayerObject.club }}
                                </q-item-label>
                            </q-item-section>
                            <q-item-section side top>
                                <q-item-label
                                    caption
                                    :class="
                                        $q.dark.isActive
                                            ? 'text-grey-5'
                                            : 'text-blue-grey-7'
                                    "
                                    >Overall</q-item-label
                                >
                                <div
                                    class="attribute-value fifa-stat-value text-h6"
                                    :class="
                                        getUnifiedRatingClass(
                                            selectedTeamPlayerObject.Overall,
                                            100,
                                        )
                                    "
                                >
                                    {{ selectedTeamPlayerObject.Overall }}
                                </div>
                            </q-item-section>
                        </q-item>
                        <q-card-section
                            class="q-pt-none"
                            :class="
                                $q.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-blue-grey-8'
                            "
                        >
                            Target Overall for Upgrades:
                            <span
                                class="text-weight-bold attribute-value"
                                :class="
                                    getUnifiedRatingClass(
                                        targetOverallForSearch,
                                        100,
                                    )
                                "
                            >
                                {{ targetOverallForSearch }}
                            </span>
                            (Base {{ getBaseOverallFromSelectedPlayer() }} +
                            Upgrade By {{ upgradeByValue }})
                        </q-card-section>
                    </q-card>
                </div>

                <div
                    class="text-subtitle1 q-mb-sm"
                    v-if="upgradePlayers.length > 0"
                    :class="$q.dark.isActive ? 'text-grey-3' : 'text-grey-8'"
                >
                    Potential upgrades ({{ upgradePlayers.length }} players
                    found):
                </div>

                <PlayerDataTable
                    v-if="upgradePlayers.length > 0"
                    :players="upgradePlayers"
                    :loading="loading"
                    @player-selected="handlePlayerSelectedForDetailView"
                    :is-goalkeeper-view="upgradeFinderIsGoalkeeperView"
                />

                <q-banner
                    v-else-if="showResults && !loading && !initialLoad"
                    class="q-mt-md"
                    :class="
                        $q.dark.isActive
                            ? 'bg-blue-grey-8 text-blue-grey-2'
                            : 'bg-info text-white'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="info" />
                    </template>
                    No upgrades found matching all criteria. Try adjusting
                    filters.
                </q-banner>
                <q-banner
                    v-else-if="
                        showResults &&
                        !loading &&
                        initialLoad &&
                        !selectedTeamPlayer
                    "
                    class="q-mt-md"
                    :class="
                        $q.dark.isActive
                            ? 'bg-orange-9 text-white'
                            : 'bg-amber text-dark'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="warning" />
                    </template>
                    Please select a team, position, and a player from that team
                    to serve as the upgrade baseline.
                </q-banner>
            </q-card-section>
            <q-inner-loading :showing="loading">
                <q-spinner-gears size="50px" color="primary" />
            </q-inner-loading>
        </q-card>
    </q-dialog>

    <PlayerDetailDialog
        :player="playerForDetailView"
        :show="showPlayerDetailDialog"
        @close="showPlayerDetailDialog = false"
    />
</template>

<script>
// Script content remains the same as your provided version
import { ref, computed, onMounted, watch } from "vue";
import { useQuasar } from "quasar";
import PlayerDataTable from "./PlayerDataTable.vue";
import PlayerDetailDialog from "./PlayerDetailDialog.vue";

const formatMonetaryValue = (value) => {
    if (value === null || value === undefined) return "Any";
    if (value >= 1000000) return `€${(value / 1000000).toFixed(1)}M`;
    if (value >= 1000) return `€${Math.round(value / 1000)}K`;
    return `€${value}`;
};

export default {
    name: "UpgradeFinderDialog",
    components: { PlayerDataTable, PlayerDetailDialog },
    props: {
        show: { type: Boolean, default: false },
        players: { type: Array, required: true },
    },
    emits: ["close"],
    setup(props, { emit }) {
        const $q = useQuasar();
        const teamName = ref(null);
        const teamOptions = ref([]);
        const allTeamNamesCache = ref([]);

        const selectedPosition = ref(null);
        const selectedTeamPlayer = ref(null);
        const teamPlayersForSelection = ref([]);

        const upgradeByValue = ref(1);

        const ageSliderMin = 15;
        const ageSliderMax = 50;
        const maxAgeFilter = ref(ageSliderMax);

        const maxTransferValueFilter = ref(null);
        const dynamicMinTransferValue = ref(0);
        const dynamicMaxTransferValue = ref(100000000);

        const loading = ref(false);
        const showResults = ref(false);
        const initialLoad = ref(true);

        const upgradePlayers = ref([]);
        const playerForDetailView = ref(null);
        const showPlayerDetailDialog = ref(false);

        const populateAllTeamNames = () => {
            if (!props || !props.players) {
                allTeamNamesCache.value = [];
                teamOptions.value = [];
                return;
            }
            const uniqueTeams = new Set();
            props.players.forEach((player) => {
                if (player.club && player.club.trim() !== "") {
                    uniqueTeams.add(player.club);
                }
            });
            allTeamNamesCache.value = Array.from(uniqueTeams).sort();
            teamOptions.value = allTeamNamesCache.value;
        };

        const updateTransferValueSliderBounds = () => {
            if (!props || !props.players) {
                dynamicMinTransferValue.value = 0;
                dynamicMaxTransferValue.value = 100000000;
                maxTransferValueFilter.value = dynamicMaxTransferValue.value;
                return;
            }

            if (props.players.length === 0) {
                dynamicMinTransferValue.value = 0;
                dynamicMaxTransferValue.value = 100000000;
                maxTransferValueFilter.value = dynamicMaxTransferValue.value;
                return;
            }

            let minVal = Infinity;
            let maxVal = 0;
            props.players.forEach((p) => {
                if (typeof p.transferValueAmount === "number") {
                    minVal = Math.min(minVal, p.transferValueAmount);
                    maxVal = Math.max(maxVal, p.transferValueAmount);
                }
            });

            dynamicMinTransferValue.value =
                minVal === Infinity ? 0 : Math.max(0, minVal);
            dynamicMaxTransferValue.value =
                maxVal === 0 && minVal === Infinity ? 100000000 : maxVal;

            if (
                maxTransferValueFilter.value === null ||
                maxTransferValueFilter.value > dynamicMaxTransferValue.value ||
                maxTransferValueFilter.value < dynamicMinTransferValue.value
            ) {
                maxTransferValueFilter.value = dynamicMaxTransferValue.value;
            }
        };

        onMounted(() => {
            populateAllTeamNames();
            updateTransferValueSliderBounds();
            maxAgeFilter.value = ageSliderMax;
            if (props && props.players && props.players.length > 0) {
                maxTransferValueFilter.value = dynamicMaxTransferValue.value;
            } else {
                maxTransferValueFilter.value = 100000000;
            }
        });

        watch(
            () => props.players,
            (newPlayers) => {
                populateAllTeamNames();
                updateTransferValueSliderBounds();
                if (newPlayers && newPlayers.length > 0) {
                    maxTransferValueFilter.value =
                        dynamicMaxTransferValue.value;
                } else {
                    allTeamNamesCache.value = [];
                    teamOptions.value = [];
                    dynamicMinTransferValue.value = 0;
                    dynamicMaxTransferValue.value = 100000000;
                    maxTransferValueFilter.value =
                        dynamicMaxTransferValue.value;
                }
            },
            { immediate: true, deep: true },
        );

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
            Attackers: [
                "Striker",
                "Right Forward",
                "Left Forward",
                "Centre Forward",
            ],
        };

        const positionFilterOptions = computed(() => {
            const options = [];
            Object.keys(positionGroups).forEach((group) => {
                options.push({ label: `${group} (Group)`, value: group });
            });
            const uniquePositions = new Set();
            if (props.players) {
                props.players.forEach((player) => {
                    player.parsedPositions?.forEach((pos) =>
                        uniquePositions.add(pos),
                    );
                });
            }
            Array.from(uniquePositions)
                .sort()
                .forEach((pos) => {
                    if (!positionGroups[pos]) {
                        // Add only if not already covered by a group entry (avoids duplicates if a single position is also a group name)
                        options.push({ label: pos, value: pos });
                    }
                });
            return options.sort((a, b) => a.label.localeCompare(b.label));
        });

        const filterTeams = (val, update) => {
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

        watch([teamName, selectedPosition], () => {
            selectedTeamPlayer.value = null;
            if (
                teamName.value &&
                selectedPosition.value &&
                props.players &&
                props.players.length > 0
            ) {
                teamPlayersForSelection.value = props.players
                    .filter((player) => {
                        if (player.club !== teamName.value) return false;
                        const isGroup =
                            !!positionGroups[selectedPosition.value];
                        if (isGroup) {
                            return (
                                player.positionGroups &&
                                player.positionGroups.includes(
                                    selectedPosition.value,
                                )
                            );
                        } else {
                            return (
                                player.parsedPositions &&
                                player.parsedPositions.includes(
                                    selectedPosition.value,
                                )
                            );
                        }
                    })
                    .sort((a, b) => (b.Overall || 0) - (a.Overall || 0));
            } else {
                teamPlayersForSelection.value = [];
            }
        });

        const getBaseOverallFromSelectedPlayer = () => {
            if (!selectedTeamPlayer.value) return null;
            const player = teamPlayersForSelection.value.find(
                (p) => p.name === selectedTeamPlayer.value,
            );
            return player ? player.Overall : null;
        };

        const selectedTeamPlayerObject = computed(() => {
            if (!selectedTeamPlayer.value) return null;
            return (
                teamPlayersForSelection.value.find(
                    (p) => p.name === selectedTeamPlayer.value,
                ) || null
            );
        });

        const targetOverallForSearch = computed(() => {
            const base = getBaseOverallFromSelectedPlayer();
            if (base === null) return null;
            return base + upgradeByValue.value;
        });

        const computedMinSliderTransferValue = computed(
            () => dynamicMinTransferValue.value,
        );
        const computedMaxSliderTransferValue = computed(
            () => dynamicMaxTransferValue.value,
        );

        const computedStepSliderTransferValue = computed(() => {
            const range =
                computedMaxSliderTransferValue.value -
                computedMinSliderTransferValue.value;
            if (range <= 0) return 10000;
            if (range < 100000) return 5000;
            if (range < 1000000) return 25000;
            if (range < 10000000) return 100000;
            if (range < 50000000) return 250000;
            return 500000;
        });

        const formattedMaxTransferValueLabel = computed(() => {
            if (
                maxTransferValueFilter.value ===
                computedMaxSliderTransferValue.value
            )
                return "Any";
            return formatMonetaryValue(maxTransferValueFilter.value);
        });

        const findUpgrades = async () => {
            if (!selectedTeamPlayer.value) {
                upgradePlayers.value = [];
                showResults.value = true;
                initialLoad.value = false;
                return;
            }
            if (!props || !props.players) {
                loading.value = false;
                return;
            }

            loading.value = true;
            showResults.value = true;
            initialLoad.value = false;

            const baseOverall = getBaseOverallFromSelectedPlayer();
            if (baseOverall === null) {
                loading.value = false;
                upgradePlayers.value = [];
                return;
            }

            const targetOverall = baseOverall + upgradeByValue.value;
            const currentMaxTransferValue = maxTransferValueFilter.value;
            const currentMaxAge = maxAgeFilter.value;

            await new Promise((resolve) => setTimeout(resolve, 300));

            try {
                upgradePlayers.value = props.players
                    .filter((player) => {
                        if (player.club === teamName.value) return false;
                        if (
                            player.transfer_value &&
                            player.transfer_value.toLowerCase() ===
                                "not for sale"
                        )
                            return false;

                        let matchesPosition = false;
                        const isGroup =
                            !!positionGroups[selectedPosition.value];
                        if (isGroup) {
                            matchesPosition =
                                player.positionGroups &&
                                player.positionGroups.includes(
                                    selectedPosition.value,
                                );
                        } else {
                            matchesPosition =
                                player.parsedPositions &&
                                player.parsedPositions.includes(
                                    selectedPosition.value,
                                );
                        }
                        if (!matchesPosition) return false;

                        if ((player.Overall || 0) < targetOverall) return false;
                        if (
                            currentMaxAge < ageSliderMax &&
                            (player.age || 0) > currentMaxAge
                        )
                            return false;
                        if (
                            currentMaxTransferValue <
                                computedMaxSliderTransferValue.value &&
                            (player.transferValueAmount || 0) >
                                currentMaxTransferValue
                        )
                            return false;

                        return true;
                    })
                    .sort((a, b) => (b.Overall || 0) - (a.Overall || 0));
            } catch (error) {
                console.error("Error finding upgrades:", error);
            } finally {
                loading.value = false;
            }
        };

        const handlePlayerSelectedForDetailView = (player) => {
            playerForDetailView.value = player;
            showPlayerDetailDialog.value = true;
        };

        const getUnifiedRatingClass = (value, maxScale) => {
            // Copied from PlayerDetailDialog for consistency
            const numValue = parseInt(value, 10);
            if (
                isNaN(numValue) ||
                value === null ||
                value === undefined ||
                value === "-"
            )
                return "rating-na";
            const percentage = (numValue / maxScale) * 100;
            if (percentage >= 90) return "rating-tier-6";
            if (percentage >= 80) return "rating-tier-5";
            if (percentage >= 70) return "rating-tier-4";
            if (percentage >= 55) return "rating-tier-3";
            if (percentage >= 40) return "rating-tier-2";
            return "rating-tier-1";
        };

        const upgradeFinderIsGoalkeeperView = computed(() => {
            return (
                selectedPosition.value === "Goalkeepers" ||
                selectedPosition.value === "Goalkeeper"
            );
        });

        watch(
            () => props.show,
            (newValue) => {
                if (!newValue) {
                    teamName.value = null;
                    selectedPosition.value = null;
                    selectedTeamPlayer.value = null;
                    teamPlayersForSelection.value = [];
                    upgradeByValue.value = 1;
                    maxAgeFilter.value = ageSliderMax;

                    if (props && props.players && props.players.length > 0) {
                        maxTransferValueFilter.value =
                            computedMaxSliderTransferValue.value;
                    } else {
                        maxTransferValueFilter.value = 100000000;
                    }

                    showResults.value = false;
                    upgradePlayers.value = [];
                    loading.value = false;
                    initialLoad.value = true;
                } else {
                    populateAllTeamNames();
                    updateTransferValueSliderBounds();
                    maxAgeFilter.value = ageSliderMax;
                    maxTransferValueFilter.value =
                        computedMaxSliderTransferValue.value;
                }
            },
        );

        return {
            $q,
            teamName,
            teamOptions,
            filterTeams,
            selectedPosition,
            positionFilterOptions,
            selectedTeamPlayer,
            teamPlayersForSelection,
            getBaseOverallFromSelectedPlayer,
            selectedTeamPlayerObject,
            targetOverallForSearch,
            upgradeByValue,
            maxAgeFilter,
            ageSliderMin,
            ageSliderMax,
            maxTransferValueFilter,
            computedMinSliderTransferValue,
            computedMaxSliderTransferValue,
            computedStepSliderTransferValue,
            formattedMaxTransferValueLabel,
            loading,
            showResults,
            initialLoad,
            upgradePlayers,
            findUpgrades,
            getUnifiedRatingClass,
            playerForDetailView,
            showPlayerDetailDialog,
            handlePlayerSelectedForDetailView,
            props,
            upgradeFinderIsGoalkeeperView,
        };
    },
};
</script>

<style lang="scss" scoped>
.upgrade-finder-dialog {
    // Main dialog card already gets bg-dark or bg-grey-1 from template
}

.results-section {
    // Ensure this section also adapts if needed, though q-card usually handles it
}

// Baseline Player Card Styling
.q-card.bg-blue-grey-1 {
    // Light mode specific
    // Already has text-blue-grey-10 from template for text
}
.q-card.bg-grey-8 {
    // Dark mode specific
    // Already has text-grey-3 from template for text
}

.nationality-flag-dialog {
    border: 1px solid rgba(0, 0, 0, 0.2);
    .body--dark & {
        border: 1px solid rgba(255, 255, 255, 0.2);
    }
}

// Attribute value styling (uses global .attribute-value and .rating-tier-* classes)
.attribute-value {
    // Base styling is global in app.scss
}
.fifa-stat-value {
    // Specifics for FIFA stats if needed, often just inherits from attribute-value
}

.q-banner {
    border-radius: $generic-border-radius;
}

.text-subtitle2 {
    font-size: 0.875rem;
    font-weight: 400;
    line-height: 1.25rem;
    letter-spacing: 0.0178571429em;
    // Color is applied dynamically in template
}

// Ensure q-select dropdowns are styled for dark mode
:deep(.q-menu) {
    .body--dark & {
        background-color: $grey-8 !important;
        color: $grey-3 !important;
        .q-item--active,
        .q-item.q-router-link--active {
            background-color: rgba(255, 255, 255, 0.1) !important;
        }
        .q-item:hover {
            background-color: rgba(255, 255, 255, 0.05) !important;
        }
    }
}
</style>
