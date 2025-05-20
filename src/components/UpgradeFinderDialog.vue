<template>
    <q-dialog
        :model-value="show"
        @update:model-value="$emit('close')"
        persistent
        maximized
        transition-show="slide-up"
        transition-hide="slide-down"
    >
        <q-card class="upgrade-finder-dialog">
            <q-card-section class="row items-center q-pb-none">
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

            <q-card-section class="q-pt-sm">
                <p class="text-body1 q-mb-md">
                    Search for players that would be an upgrade to your current team.
                </p>

                <div class="row q-col-gutter-md q-mb-md">
                    <!-- Team Name Searchable Dropdown -->
                    <div class="col-12 col-sm-6 col-md-4">
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

                    <!-- Position Selector -->
                    <div class="col-12 col-sm-6 col-md-4">
                        <q-select
                            v-model="selectedPosition"
                            :options="positionFilterOptions"
                            label="Position / Group"
                            dense
                            outlined
                            emit-value
                            map-options
                            :rules="[(val) => !!val || 'Position is required']"
                        />
                    </div>

                    <!-- Minimum Rating Slider -->
                    <div class="col-12 col-sm-6 col-md-4">
                        <div>
                            <div class="text-subtitle2 q-mb-xs">Minimum Rating: {{ minRating }}</div>
                            <q-slider
                                v-model="minRating"
                                :min="60"
                                :max="99"
                                :step="1"
                                label
                                label-always
                                color="primary"
                            />
                        </div>
                    </div>
                </div>

                <div class="row q-col-gutter-md">
                    <div class="col-12">
                        <q-btn
                            color="primary"
                            label="Find Upgrades"
                            class="full-width"
                            @click="findUpgrades"
                            :loading="loading"
                            :disable="!teamName || !selectedPosition || loading"
                        />
                    </div>
                </div>
            </q-card-section>

            <!-- Results Section -->
            <q-card-section v-if="showResults">
                <div class="text-h6 q-mb-md">Results</div>

                <!-- Team Players -->
                <div class="text-subtitle1 q-mb-sm" v-if="teamPlayers.length > 0">
                    Current {{ teamName }} players in {{ selectedPositionLabel }}:
                </div>
                <q-list bordered separator v-if="teamPlayers.length > 0" class="q-mb-lg">
                    <q-item v-for="player in teamPlayers" :key="player.name">
                        <q-item-section>
                            <q-item-label>{{ player.name }}</q-item-label>
                            <q-item-label caption>
                                {{ player.position }} | Overall: {{ player.Overall }}
                            </q-item-label>
                        </q-item-section>
                        <q-item-section side>
                            <div
                                class="attribute-value fifa-stat-value"
                                :class="getFifaStatClass(player.Overall)"
                            >
                                {{ player.Overall }}
                            </div>
                        </q-item-section>
                    </q-item>
                </q-list>

                <!-- No Team Players Message -->
                <q-banner v-else class="bg-warning text-dark q-mb-lg">
                    No players found for team "{{ teamName }}" in position "{{ selectedPositionLabel }}".
                </q-banner>

                <!-- Upgrade Players -->
                <div class="text-subtitle1 q-mb-sm" v-if="upgradePlayers.length > 0">
                    Potential upgrades ({{ upgradePlayers.length }} players found):
                </div>

                <!-- Upgrades Table -->
                <PlayerDataTable
                    v-if="upgradePlayers.length > 0"
                    :players="upgradePlayers"
                    :loading="loading"
                    @player-selected="handlePlayerSelected"
                />

                <!-- No Upgrades Message -->
                <q-banner v-else-if="showResults && !loading" class="bg-blue-1 text-dark">
                    No upgrades found for the selected criteria. Try lowering the minimum rating or changing the position.
                </q-banner>
            </q-card-section>
        </q-card>
    </q-dialog>

    <!-- Player Detail Dialog -->
    <PlayerDetailDialog
        :player="selectedPlayer"
        :show="showPlayerDetailDialog"
        @close="showPlayerDetailDialog = false"
    />
</template>

<script>
import { ref, computed, onMounted, watch } from "vue";
import PlayerDataTable from "./PlayerDataTable.vue";
import PlayerDetailDialog from "./PlayerDetailDialog.vue";

export default {
    name: "UpgradeFinderDialog",
    components: { PlayerDataTable, PlayerDetailDialog },
    props: {
        show: { type: Boolean, default: false },
        players: { type: Array, required: true },
    },
    emits: ["close"],
    setup(props, { emit }) {
        const teamName = ref("");
        const teamOptions = ref([]);
        const selectedPosition = ref(null);
        const minRating = ref(70);
        const loading = ref(false);
        const showResults = ref(false);
        const teamPlayers = ref([]);
        const upgradePlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        
        // Get unique team names from the players array
        const getAllTeams = () => {
            const uniqueTeams = new Set();
            props.players.forEach(player => {
                if (player.club && player.club.trim() !== '') {
                    uniqueTeams.add(player.club);
                }
            });
            return Array.from(uniqueTeams).sort();
        };

        // Position groups from PlayerUploadPage for filtering
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

        const positionFilterOptions = computed(() => {
            const options = [];
            Object.keys(positionGroups).forEach((group) => {
                options.push({ label: `${group} (Group)`, value: group });
            });
            const uniquePositions = new Set();
            props.players.forEach((player) => {
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

        const selectedPositionLabel = computed(() => {
            const option = positionFilterOptions.value.find(
                (opt) => opt.value === selectedPosition.value
            );
            return option ? option.label : selectedPosition.value;
        });

        // Filter teams based on user input
        const filterTeams = (val, update) => {
            if (val === '') {
                update(() => {
                    teamOptions.value = getAllTeams();
                });
                return;
            }

            update(() => {
                const needle = val.toLowerCase();
                teamOptions.value = getAllTeams().filter(
                    team => team.toLowerCase().indexOf(needle) > -1
                );
            });
        };

        // Initialize team options when dialog is opened
        watch(() => props.show, (newVal) => {
            if (newVal) {
                teamOptions.value = getAllTeams();
            }
        });

        // Also initialize on component mount
        onMounted(() => {
            teamOptions.value = getAllTeams();
        });

        const findUpgrades = async () => {
            loading.value = true;
            showResults.value = true;

            try {
                // Find players from the specified team in the selected position
                teamPlayers.value = props.players.filter((player) => {
                    if (!player.club || player.club !== teamName.value) {
                        return false;
                    }

                    if (positionGroups[selectedPosition.value]) {
                        return player.positionGroups && player.positionGroups.includes(selectedPosition.value);
                    } else {
                        return player.parsedPositions && player.parsedPositions.includes(selectedPosition.value);
                    }
                });

                // Find players better than the team's best player in that position
                const teamMaxRating = Math.max(...teamPlayers.value.map(p => p.Overall || 0), 0);
                
                // Find upgrade players - better than team's players and meet minimum rating
                upgradePlayers.value = props.players.filter((player) => {
                    // Exclude players from the same team
                    if (player.club && player.club === teamName.value) {
                        return false;
                    }

                    // Check if player plays in the selected position
                    let matchesPosition = false;
                    if (positionGroups[selectedPosition.value]) {
                        matchesPosition = player.positionGroups && player.positionGroups.includes(selectedPosition.value);
                    } else {
                        matchesPosition = player.parsedPositions && player.parsedPositions.includes(selectedPosition.value);
                    }

                    // Player must match position, exceed team's best player, and meet minimum rating
                    return matchesPosition && 
                           player.Overall >= Math.max(teamMaxRating, minRating.value);
                });

                // Sort upgrades by overall rating descending
                upgradePlayers.value.sort((a, b) => (b.Overall || 0) - (a.Overall || 0));
            } catch (error) {
                console.error("Error finding upgrades:", error);
            } finally {
                loading.value = false;
            }
        };

        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        const getFifaStatClass = (v) => {
            if (v === null || v === undefined || v === "-")
                return "attribute-na";
            const n = typeof v === "number" ? v : parseInt(v, 10);
            if (isNaN(n)) return "attribute-na";
            if (n >= 90) return "attribute-elite";
            if (n >= 80) return "attribute-excellent";
            if (n >= 70) return "attribute-very-good";
            if (n >= 60) return "attribute-good";
            if (n >= 50) return "attribute-average";
            if (n >= 40) return "attribute-below-average";
            if (n >= 30) return "attribute-poor";
            return "attribute-very-poor";
        };

        // Reset results when dialog is closed
        watch(() => props.show, (newValue) => {
            if (!newValue) {
                showResults.value = false;
                teamPlayers.value = [];
                upgradePlayers.value = [];
            }
        });

        return {
            teamName,
            teamOptions,
            filterTeams,
            selectedPosition,
            selectedPositionLabel,
            minRating,
            loading,
            showResults,
            teamPlayers,
            upgradePlayers,
            positionFilterOptions,
            findUpgrades,
            getFifaStatClass,
            selectedPlayer,
            showPlayerDetailDialog,
            handlePlayerSelected
        };
    },
};
</script>

<style scoped>
.upgrade-finder-dialog {
    border-radius: 8px;
    min-height: 80vh;
}

.attribute-value {
    display: inline-block;
    min-width: 30px;
    text-align: center;
    font-weight: 600;
    padding: 2px 5px;
    border-radius: 3px;
    font-size: 0.85em;
}
.fifa-stat-value {
    font-size: 1.1em;
    padding: 4px 8px;
}
.attribute-elite {
    background-color: #9c27b0;
    color: white;
}
.attribute-excellent {
    background-color: #20c997;
    color: white;
}
.attribute-very-good {
    background-color: #4dabf7;
    color: white;
}
.attribute-good {
    background-color: #82c91e;
    color: #212529;
}
.attribute-average {
    background-color: #ffc107;
    color: #212529;
}
.attribute-below-average {
    background-color: #fab005;
    color: #212529;
}
.attribute-poor {
    background-color: #ff922b;
    color: #212529;
}
.attribute-very-poor {
    background-color: #fa5252;
    color: white;
}
.attribute-na {
    background-color: #e9ecef;
    color: #868e96;
}
</style>