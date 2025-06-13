<template>
    <q-page class="performance-page">
        <!-- Error State -->
        <div v-if="pageLoadingError" class="error-container">
            <q-banner class="error-banner" rounded>
                <template v-slot:avatar>
                    <q-icon name="error" />
                </template>
                {{ pageLoadingError }}
                <q-btn flat color="white" label="Go to Upload Page" @click="router.push('/')" class="q-ml-md" />
            </q-banner>
        </div>

        <!-- No Data State -->
        <div v-else-if="!currentDatasetId || allPlayersData.length === 0" class="no-data-container">
             <q-banner class="no-data-banner">
                <template v-slot:avatar>
                    <q-icon name="warning" />
                </template>
                No player data found. Please upload a dataset first.
                <q-btn flat color="primary" label="Go to Upload Page" @click="router.push('/')" class="q-ml-md"/>
            </q-banner>
        </div>

        <!-- Main Content -->
        <div v-else class="main-content">
            <!-- New Styled Hero Header -->
            <div class="performance-hero-section">
                <div class="hero-content">
                    <div class="hero-left">
                        <div class="hero-title-line">
                            <q-icon name="trending_up" size="2.5rem" />
                            <h1 class="hero-title">Performance Leaders</h1>
                        </div>
                        <p class="hero-subtitle">
                            {{ formatNumber(filteredPlayers.length) }} players matching filters from {{ formatNumber(allPlayersData.length) }} total
                        </p>
                    </div>
                    <div class="hero-right">
                         <q-btn unelevated icon="share" label="Share" @click="shareDataset" class="share-btn-modern"/>
                    </div>
                </div>

                <!-- Filter Bar Integrated into Hero -->
                <div class="filter-bar">
                    <q-select
                        v-model="selectedDivisions"
                        :options="divisionOptions"
                        label="Filter by Division"
                        dense
                        outlined
                        multiple
                        use-chips
                        use-input
                        @filter="filterDivisionsFn"
                        class="division-filter"
                        dark
                        popup-content-class="bg-grey-10"
                    >
                         <template v-slot:no-option>
                            <q-item>
                                <q-item-section class="text-grey">No divisions found</q-item-section>
                            </q-item>
                        </template>
                    </q-select>
                    <div class="minutes-filter">
                        <div class="slider-label">Minimum Minutes Played</div>
                        <q-slider
                            v-model="sliderValue"
                            :min="0"
                            :max="maxMinutes"
                            :step="50"
                            label
                            :label-value="`${sliderValue}+ mins`"
                            label-always
                            class="q-mt-sm"
                            dark
                            color="light-blue-4"
                        />
                    </div>
                </div>
            </div>

             <!-- Tabbed Content Section -->
            <q-card class="tabs-card">
                <q-tabs
                    v-model="currentTab"
                    dense
                    class="text-grey"
                    active-color="primary"
                    indicator-color="primary"
                    align="justify"
                    narrow-indicator
                >
                    <q-tab name="attacking" icon="sports_soccer" label="Attacking" />
                    <q-tab name="passing" icon="swap_horiz" label="Passing" />
                    <q-tab name="defending" icon="shield" label="Defending" />
                    <q-tab name="goalkeeping" icon="sports_hockey" label="Goalkeeping" />
                </q-tabs>

                <q-separator />

                <q-tab-panels v-model="currentTab" animated>
                    <q-tab-panel name="attacking">
                        <div class="tab-content-layout">
                            <h2 class="category-title">Attacking Visualizations</h2>
                            <div class="charts-grid">
                                <ScatterPlotCard v-for="config in attackingCharts" :key="config.title" v-bind="config" :is-dark-mode="$q.dark.isActive" :all-players-data="filteredPlayers" />
                            </div>
                            <h2 class="category-title">Attacking Leaderboards</h2>
                            <div class="stats-grid">
                                <StatCard v-for="stat in attackingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                            </div>
                        </div>
                    </q-tab-panel>

                    <q-tab-panel name="passing">
                         <div class="tab-content-layout">
                            <h2 class="category-title">Passing Visualizations</h2>
                            <div class="charts-grid">
                                <ScatterPlotCard v-for="config in passingCharts" :key="config.title" v-bind="config" :is-dark-mode="$q.dark.isActive" :all-players-data="filteredPlayers" />
                            </div>
                            <h2 class="category-title">Passing Leaderboards</h2>
                            <div class="stats-grid">
                                <StatCard v-for="stat in passingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                            </div>
                        </div>
                    </q-tab-panel>

                    <q-tab-panel name="defending">
                        <div class="tab-content-layout">
                            <h2 class="category-title">Defending Visualizations</h2>
                            <div class="charts-grid">
                                <ScatterPlotCard v-for="config in defendingCharts" :key="config.title" v-bind="config" :is-dark-mode="$q.dark.isActive" :all-players-data="filteredPlayers" />
                            </div>
                            <h2 class="category-title">Defending Leaderboards</h2>
                            <div class="stats-grid">
                                <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                            </div>
                        </div>
                    </q-tab-panel>

                    <q-tab-panel name="goalkeeping">
                        <div class="tab-content-layout">
                            <h2 class="category-title">Goalkeeping Visualizations</h2>
                            <div class="charts-grid">
                                <ScatterPlotCard v-for="config in goalkeepingCharts" :key="config.title" v-bind="config" :is-dark-mode="$q.dark.isActive" :all-players-data="filteredPlayers" />
                            </div>
                            <h2 class="category-title">Goalkeeping Leaderboards</h2>
                            <div class="stats-grid">
                                <StatCard v-for="stat in goalkeepingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="openPlayerDetail" />
                            </div>
                        </div>
                    </q-tab-panel>
                </q-tab-panels>
            </q-card>
        </div>

        <!-- Player Detail Dialog -->
        <PlayerDetailDialog :player="playerForDetailView" :show="showPlayerDetailDialog" @close="showPlayerDetailDialog = false" :currency-symbol="detectedCurrencySymbol" :dataset-id="currentDatasetId" />
    </q-page>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useQuasar, debounce } from 'quasar';
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue';
import { usePlayerStore } from '../stores/playerStore';
import ScatterPlotCard from '../components/ScatterPlotCard.vue';
import StatCard from '../components/StatCard.vue';

const router = useRouter();
const route = useRoute();
const $q = useQuasar();
const playerStore = usePlayerStore();

// --- Reactive Data ---
const pageLoadingError = ref('');
const showPlayerDetailDialog = ref(false);
const playerForDetailView = ref(null);
const topPlayersByStat = ref({});
const currentTab = ref('attacking');

// --- Filter State with new defaults ---
const sliderValue = ref(1500);
const selectedMinutes = ref(1500);
const selectedDivisions = ref(['Premier League', 'Ligue 1 Uber Eats', 'Spanish First Division', 'Serie A', 'Bundesliga']);
const divisionOptions = ref([]);

// --- Computed Properties from Store ---
const allPlayersData = computed(() => playerStore.allPlayers);
const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol);
const currentDatasetId = computed(() => playerStore.currentDatasetId);

// --- Computed Properties for Filtering ---
const maxMinutes = computed(() => Math.max(2000, ...allPlayersData.value.map(p => getNumericValue(p.attributes?.Mins) || 0)));
const availableDivisions = computed(() => {
    const divisions = [...new Set(allPlayersData.value.map(p => getPlayerDivision(p)).filter(Boolean))].sort();
    // Ensure default selections are included if they exist in the data
    selectedDivisions.value = selectedDivisions.value.filter(d => divisions.includes(d));
    return divisions;
});
const filteredPlayers = computed(() => {
    return allPlayersData.value.filter(player => {
        const minutesPlayed = getNumericValue(player.attributes?.Mins) || 0;
        const division = getPlayerDivision(player);
        const matchesMinutes = minutesPlayed >= selectedMinutes.value;
        const matchesDivision = selectedDivisions.value.length === 0 || selectedDivisions.value.includes(division);
        return matchesMinutes && matchesDivision;
    });
});

// --- Data Definitions (Charts & Stats) ---
const scatterPlotConfigs = ref([
    { category: 'attacking', title: 'Shooting Performance', xAxisKey: 'xG/90', yAxisKey: 'Gls/90', xAxisLabel: 'Expected Goals per 90', yAxisLabel: 'Goals per 90', quadrantLabels: { topRight: ['Elite', 'Over-performing'], topLeft: ['Clinical', 'Over-performing'], bottomRight: ['Wasteful', 'Under-performing'], bottomLeft: ['Low Threat', 'Under-performing'] }},
    { category: 'attacking', title: 'Shooting Efficiency', xAxisKey: 'Shot/90', yAxisKey: 'Conv %', xAxisLabel: 'Shots per 90', yAxisLabel: 'Conversion %', quadrantLabels: { topRight: ['Elite Attacker', ''], topLeft: ['Selective Shooter', ''], bottomRight: ['Inefficient Volume', ''], bottomLeft: ['Limited Threat', ''] }},
    { category: 'passing', title: 'Creative Passing', xAxisKey: 'xA/90', yAxisKey: 'Asts/90', xAxisLabel: 'Expected Assists per 90', yAxisLabel: 'Assists per 90', quadrantLabels: { topRight: ['Elite Creator', ''], topLeft: ['Fortunate Creator', ''], bottomRight: ['Unlucky Creator', ''], bottomLeft: ['Limited Creator', ''] }},
    { category: 'passing', title: 'PASSING PROGRESSION', xAxisKey: 'Pr passes/90', yAxisKey: 'Pas %', xAxisLabel: 'PROGRESSIVE PASSES/90', yAxisLabel: 'PASS COMPLETION (%)', quadrantLabels: { topRight: ['Accurate passing', 'High volume'], topLeft: ['Accurate passing', 'Low volume'], bottomRight: ['Inaccurate passing', 'High volume'], bottomLeft: ['Inaccurate passing', 'Low volume'] }},
    { category: 'defending', title: 'Defensive Duels', xAxisKey: 'Tck/90', yAxisKey: 'Tck R', xAxisLabel: 'Tackles per 90', yAxisLabel: 'Tackle Success %', quadrantLabels: { topRight: ['Elite Ball-Winner', ''], topLeft: ['Conservative', ''], bottomRight: ['Reckless', ''], bottomLeft: ['Passive', ''] }},
    { category: 'defending', title: 'Pressing Efficiency', xAxisKey: 'Pres C/90', yAxisKey: 'Poss Won/90', xAxisLabel: 'Pressures Completed per 90', yAxisLabel: 'Possession Won per 90', quadrantLabels: { topRight: ['Effective Presser', ''], topLeft: ['Positional Winner', ''], bottomRight: ['Ineffective Presser', ''], bottomLeft: ['Low Activity', ''] }},
    { category: 'goalkeeping', title: 'Shot-Stopping', xAxisKey: 'Con/90', yAxisKey: 'Sv %', xAxisLabel: 'Goals Conceded per 90', yAxisLabel: 'Save Percentage', quadrantLabels: { topRight: ['Busy & Effective', ''], topLeft: ['Elite Goalkeeper', ''], bottomRight: ['Struggling', ''], bottomLeft: ['Protected', ''] }}
]);

const statCategories = {
    offensive: [{ key: 'Gls/90', name: 'Goals per 90' }, { key: 'xG/90', name: 'xG per 90' }, { key: 'Shot/90', name: 'Shots per 90' }, { key: 'Conv %', name: 'Conversion %' }],
    passing: [{ key: 'Asts/90', name: 'Assists per 90' }, { key: 'xA/90', name: 'xA per 90' }, { key: 'K Ps/90', name: 'Key Passes per 90' }, { key: 'Pas %', name: 'Pass Completion %' }],
    defensive: [{ key: 'Tck/90', name: 'Tackles per 90' }, { key: 'Int/90', name: 'Interceptions per 90' }, { key: 'Hdrs W/90', name: 'Headers Won per 90' }, { key: 'Pres C/90', name: 'Pressures Completed p90' }],
    goalkeeping: [{ key: 'Con/90', name: 'Goals Conceded p90' }, { key: 'xGP/90', name: 'xG Prevented p90' }, { key: 'Sv %', name: 'Save Percentage' }, { key: 'Clean Sheets', name: 'Clean Sheets' }]
};

// --- Computed properties for each tab ---
const attackingCharts = computed(() => scatterPlotConfigs.value.filter(c => c.category === 'attacking'));
const passingCharts = computed(() => scatterPlotConfigs.value.filter(c => c.category === 'passing'));
const defendingCharts = computed(() => scatterPlotConfigs.value.filter(c => c.category === 'defending'));
const goalkeepingCharts = computed(() => scatterPlotConfigs.value.filter(c => c.category === 'goalkeeping'));

const attackingStats = computed(() => statCategories.offensive);
const passingStats = computed(() => statCategories.passing);
const defendingStats = computed(() => statCategories.defensive);
const goalkeepingStats = computed(() => statCategories.goalkeeping);

const allStatsForCalculation = computed(() => [...attackingStats.value, ...passingStats.value, ...defendingStats.value, ...goalkeepingStats.value]);

// --- Helper Methods ---
const getPlayerName = (player) => player.name || player.Name || player.Player || 'Unknown Player';
const getPlayerDivision = (player) => player.division || player.Division || 'N/A';
const getNumericValue = (val) => {
    if (val === undefined || val === null || val === '-' || val === '') return null;
    const cleaned = String(val).replace(/,/g, '').replace(/%/g, '');
    const num = parseFloat(cleaned);
    return isNaN(num) ? null : num;
};

// --- Core Methods ---
const calculateTopPerformers = () => {
    const playersToProcess = filteredPlayers.value;
    const results = {};
    const uniqueStats = [...new Map(allStatsForCalculation.value.map(item => [item['key'], item])).values()];

    uniqueStats.forEach(stat => {
        const playersWithStat = playersToProcess.filter(player => {
            const numValue = getNumericValue(player.attributes?.[stat.key]);
            return numValue !== null && (stat.key === 'Con/90' ? numValue >= 0 : numValue > 0);
        });
        results[stat.key] = playersWithStat.sort((a, b) => {
            const valA = getNumericValue(a.attributes[stat.key]);
            const valB = getNumericValue(b.attributes[stat.key]);
            return stat.key === 'Con/90' ? valA - valB : valB - valA;
        }).slice(0, 10);
    });
    topPlayersByStat.value = results;
};

const formatNumber = (num) => new Intl.NumberFormat().format(num);
const openPlayerDetail = (player) => {
    playerForDetailView.value = player;
    showPlayerDetailDialog.value = true;
};

const shareDataset = () => {
    if (!currentDatasetId.value) return;
    const shareUrl = `${window.location.origin}/performance/${currentDatasetId.value}`;
    navigator.clipboard.writeText(shareUrl).then(() => {
        $q.notify({ message: 'Link copied to clipboard!', color: 'positive', icon: 'content_copy', position: 'top' });
    });
};

const initializeData = async () => {
    const targetDatasetId = route.params?.datasetId || route.query?.datasetId;
    if (targetDatasetId) {
        if (targetDatasetId !== playerStore.currentDatasetId) {
            await playerStore.fetchPlayersByDatasetId(targetDatasetId);
        }
    } else if (!currentDatasetId.value) {
        pageLoadingError.value = 'No dataset available. Please upload a dataset first.';
    }
};

const filterDivisionsFn = (val, update) => {
  update(() => {
    const needle = val.toLowerCase();
    divisionOptions.value = availableDivisions.value.filter(v => v.toLowerCase().indexOf(needle) > -1);
  })
};

// --- Watchers & Lifecycle ---
watch(sliderValue, debounce((newValue) => {
    selectedMinutes.value = newValue;
}, 300));

watch(filteredPlayers, () => {
    calculateTopPerformers();
}, { deep: true, immediate: true });

watch(availableDivisions, (newDivisions) => {
    divisionOptions.value = newDivisions;
}, { immediate: true });

onMounted(() => {
    initializeData();
});
</script>

<style lang="scss" scoped>
$primary-gradient: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
$card-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
$border-radius: 16px;

.performance-page {
    background-color: #f4f6f8;
    .body--dark & {
        background-color: #121212;
    }
}

.main-content {
    max-width: 1600px;
    margin: 0 auto;
    padding: 2rem;
}

// Hero Section Styling
.performance-hero-section {
    background: $primary-gradient;
    color: white;
    border-radius: $border-radius;
    padding: 2rem 2.5rem;
    margin-bottom: 2rem;
    box-shadow: $card-shadow;
    
    .hero-content {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 1.5rem;

        .hero-left {
            .hero-title-line {
                display: flex;
                align-items: center;
                gap: 1rem;
            }
            .hero-title {
                font-size: 2.25rem;
                font-weight: 700;
                margin: 0;
                line-height: 1.2;
            }
            .hero-subtitle {
                font-size: 1rem;
                margin: 0.5rem 0 0 0;
                color: rgba(255, 255, 255, 0.8);
            }
        }
        .share-btn-modern {
            background: rgba(255,255,255,0.15);
            color: white;
            font-weight: 600;
            border-radius: 8px;
            &:hover {
                background: rgba(255,255,255,0.25);
            }
        }
    }

    .filter-bar {
        display: grid;
        grid-template-columns: 1fr 2fr;
        gap: 2rem;
        align-items: center;
    }
    .minutes-filter .slider-label {
        font-size: 0.8rem;
        font-weight: 500;
        color: rgba(255, 255, 255, 0.8);
        margin-bottom: -0.5rem;
    }
}

// Tabs Styling
.tabs-card {
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    overflow: hidden;

    .body--dark & {
        background-color: #1e1e1e;
    }

    .q-tabs {
        padding: 0 1rem;
    }
    .q-tab {
        font-weight: 600;
    }
    .q-tab-panel {
        padding: 2rem;
    }
}

.tab-content-layout {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.category-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: #333;
    padding-bottom: 0.5rem;
    border-bottom: 2px solid #667eea;
    align-self: flex-start;

    .body--dark & {
        color: #f5f5f5;
    }
}

.charts-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
    gap: 24px;
}

.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 24px;
}

// Banners for error/no-data states
.error-container, .no-data-container {
    padding: 2rem;
}

.error-banner {
    background: linear-gradient(135deg, #d32f2f, #c62828);
    color: white;
}
.no-data-banner {
    background: #fff;
    border: 1px solid #ddd;
    box-shadow: none;
    .body--dark & {
        background-color: #2d2d2d;
        color: #f5f5f5;
        border-color: #424242;
    }
}
</style>
