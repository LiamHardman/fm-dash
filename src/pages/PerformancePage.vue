<template>
    <q-page class="performance-page full-height">
        <!-- Top Bar -->
        <div class="top-bar" v-if="!pageLoadingError && currentDatasetId && allPlayersData.length > 0">
            <div class="top-bar-content">
                <!-- Left section: Page info -->
                <div class="dataset-info">
                    <div class="dataset-title">
                        <q-icon name="trending_up" size="1.2rem" class="q-mr-xs" />
                        Performance Leaders
                    </div>
                    <div class="dataset-stats">
                        <span class="stat-item">Top Performers by Category</span>
                        <span class="stat-separator">•</span>
                        <span class="stat-item">{{ formatNumber(allPlayersData.length) }} Players</span>
                    </div>
                </div>

                <!-- Right section: Share button -->
                <div class="top-bar-controls">
                    <q-btn
                        v-if="currentDatasetId"
                        flat
                        dense
                        icon="share"
                        @click="shareDataset"
                        class="share-btn"
                        size="sm"
                    >
                        <q-tooltip>Share Performance Data</q-tooltip>
                    </q-btn>
                </div>
            </div>
        </div>

        <!-- Error State -->
        <div v-if="pageLoadingError" class="error-container">
            <q-banner class="error-banner" rounded>
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
        </div>

        <!-- No Data State -->
        <div v-if="!pageLoadingError && (!currentDatasetId || allPlayersData.length === 0)" class="no-data-container">
            <q-banner class="no-data-banner">
                <template v-slot:avatar>
                    <q-icon name="warning" />
                </template>
                No player data found. Please upload a dataset first.
                <q-btn
                    flat
                    color="primary"
                    label="Go to Upload Page"
                    @click="router.push('/')"
                    class="q-ml-md"
                />
            </q-banner>
        </div>

        <!-- Performance Categories -->
        <div v-if="!pageLoadingError && currentDatasetId && allPlayersData.length > 0" class="performance-content">
            <div class="categories-container">
                <!-- General Stats -->
                <div class="category-section">
                    <h2 class="category-title">
                        <q-icon name="assessment" class="q-mr-sm" />
                        General
                    </h2>
                    <div class="stats-grid">
                        <div v-for="stat in generalStats" :key="stat.key" class="stat-card">
                            <q-card flat bordered class="full-height">
                                <q-card-section class="stat-header">
                                    <div class="stat-name">{{ stat.name }}</div>
                                </q-card-section>
                                <q-card-section class="stat-players">
                                    <div v-if="topPlayersByStat[stat.key] && topPlayersByStat[stat.key].length > 0">
                                        <q-list separator dense>
                                            <q-item 
                                                v-for="(player, index) in topPlayersByStat[stat.key]" 
                                                :key="player.id || index"
                                                clickable
                                                @click="openPlayerDetail(player)"
                                                class="player-item"
                                            >
                                                <q-item-section avatar>
                                                    <div class="rank-badge">{{ index + 1 }}</div>
                                                </q-item-section>
                                                <q-item-section>
                                                    <q-item-label class="player-name">{{ getPlayerName(player) }}</q-item-label>
                                                    <q-item-label caption>{{ getPlayerClub(player) }}</q-item-label>
                                                </q-item-section>
                                                <q-item-section side>
                                                    <div class="stat-value">{{ formatStatValue(player.attributes[stat.key]) }}</div>
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                    </div>
                                    <div v-else class="no-data-message">
                                        No data available
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>
                    </div>
                </div>

                <!-- Offensive Stats -->
                <div class="category-section">
                    <h2 class="category-title">
                        <q-icon name="sports_soccer" class="q-mr-sm" />
                        Offensive
                    </h2>
                    <div class="stats-grid">
                        <div v-for="stat in offensiveStats" :key="stat.key" class="stat-card">
                            <q-card flat bordered class="full-height">
                                <q-card-section class="stat-header">
                                    <div class="stat-name">{{ stat.name }}</div>
                                </q-card-section>
                                <q-card-section class="stat-players">
                                    <div v-if="topPlayersByStat[stat.key] && topPlayersByStat[stat.key].length > 0">
                                        <q-list separator dense>
                                            <q-item 
                                                v-for="(player, index) in topPlayersByStat[stat.key]" 
                                                :key="player.id || index"
                                                clickable
                                                @click="openPlayerDetail(player)"
                                                class="player-item"
                                            >
                                                <q-item-section avatar>
                                                    <div class="rank-badge">{{ index + 1 }}</div>
                                                </q-item-section>
                                                <q-item-section>
                                                    <q-item-label class="player-name">{{ getPlayerName(player) }}</q-item-label>
                                                    <q-item-label caption>{{ getPlayerClub(player) }}</q-item-label>
                                                </q-item-section>
                                                <q-item-section side>
                                                    <div class="stat-value">{{ formatStatValue(player.attributes[stat.key]) }}</div>
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                    </div>
                                    <div v-else class="no-data-message">
                                        No data available
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>
                    </div>
                </div>

                <!-- Passing Stats -->
                <div class="category-section">
                    <h2 class="category-title">
                        <q-icon name="swap_horiz" class="q-mr-sm" />
                        Passing
                    </h2>
                    <div class="stats-grid">
                        <div v-for="stat in passingStats" :key="stat.key" class="stat-card">
                            <q-card flat bordered class="full-height">
                                <q-card-section class="stat-header">
                                    <div class="stat-name">{{ stat.name }}</div>
                                </q-card-section>
                                <q-card-section class="stat-players">
                                    <div v-if="topPlayersByStat[stat.key] && topPlayersByStat[stat.key].length > 0">
                                        <q-list separator dense>
                                            <q-item 
                                                v-for="(player, index) in topPlayersByStat[stat.key]" 
                                                :key="player.id || index"
                                                clickable
                                                @click="openPlayerDetail(player)"
                                                class="player-item"
                                            >
                                                <q-item-section avatar>
                                                    <div class="rank-badge">{{ index + 1 }}</div>
                                                </q-item-section>
                                                <q-item-section>
                                                    <q-item-label class="player-name">{{ getPlayerName(player) }}</q-item-label>
                                                    <q-item-label caption>{{ getPlayerClub(player) }}</q-item-label>
                                                </q-item-section>
                                                <q-item-section side>
                                                    <div class="stat-value">{{ formatStatValue(player.attributes[stat.key]) }}</div>
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                    </div>
                                    <div v-else class="no-data-message">
                                        No data available
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>
                    </div>
                </div>

                <!-- Defensive Stats -->
                <div class="category-section">
                    <h2 class="category-title">
                        <q-icon name="shield" class="q-mr-sm" />
                        Defensive
                    </h2>
                    <div class="stats-grid">
                        <div v-for="stat in defensiveStats" :key="stat.key" class="stat-card">
                            <q-card flat bordered class="full-height">
                                <q-card-section class="stat-header">
                                    <div class="stat-name">{{ stat.name }}</div>
                                </q-card-section>
                                <q-card-section class="stat-players">
                                    <div v-if="topPlayersByStat[stat.key] && topPlayersByStat[stat.key].length > 0">
                                        <q-list separator dense>
                                            <q-item 
                                                v-for="(player, index) in topPlayersByStat[stat.key]" 
                                                :key="player.id || index"
                                                clickable
                                                @click="openPlayerDetail(player)"
                                                class="player-item"
                                            >
                                                <q-item-section avatar>
                                                    <div class="rank-badge">{{ index + 1 }}</div>
                                                </q-item-section>
                                                <q-item-section>
                                                    <q-item-label class="player-name">{{ getPlayerName(player) }}</q-item-label>
                                                    <q-item-label caption>{{ getPlayerClub(player) }}</q-item-label>
                                                </q-item-section>
                                                <q-item-section side>
                                                    <div class="stat-value">{{ formatStatValue(player.attributes[stat.key]) }}</div>
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                    </div>
                                    <div v-else class="no-data-message">
                                        No data available
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>
                    </div>
                </div>

                <!-- Goalkeeping Stats -->
                <div class="category-section">
                    <h2 class="category-title">
                        <q-icon name="sports_hockey" class="q-mr-sm" />
                        Goalkeeping
                    </h2>
                    <div class="stats-grid">
                        <div v-for="stat in goalkeepingStats" :key="stat.key" class="stat-card">
                            <q-card flat bordered class="full-height">
                                <q-card-section class="stat-header">
                                    <div class="stat-name">{{ stat.name }}</div>
                                </q-card-section>
                                <q-card-section class="stat-players">
                                    <div v-if="topPlayersByStat[stat.key] && topPlayersByStat[stat.key].length > 0">
                                        <q-list separator dense>
                                            <q-item 
                                                v-for="(player, index) in topPlayersByStat[stat.key]" 
                                                :key="player.id || index"
                                                clickable
                                                @click="openPlayerDetail(player)"
                                                class="player-item"
                                            >
                                                <q-item-section avatar>
                                                    <div class="rank-badge">{{ index + 1 }}</div>
                                                </q-item-section>
                                                <q-item-section>
                                                    <q-item-label class="player-name">{{ getPlayerName(player) }}</q-item-label>
                                                    <q-item-label caption>{{ getPlayerClub(player) }}</q-item-label>
                                                </q-item-section>
                                                <q-item-section side>
                                                    <div class="stat-value">{{ formatStatValue(player.attributes[stat.key]) }}</div>
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                    </div>
                                    <div v-else class="no-data-message">
                                        No data available
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Player Detail Dialog -->
        <PlayerDetailDialog
            :player="playerForDetailView"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
            :currency-symbol="detectedCurrencySymbol"
            :dataset-id="currentDatasetId"
        />
    </q-page>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useQuasar } from 'quasar'
import PlayerDetailDialog from '../components/PlayerDetailDialog.vue'
import { usePlayerStore } from '../stores/playerStore'

const router = useRouter()
const route = useRoute()
const $q = useQuasar()
const playerStore = usePlayerStore()

// Reactive data
const pageLoadingError = ref('')
const showPlayerDetailDialog = ref(false)
const playerForDetailView = ref(null)
const topPlayersByStat = ref({})

// Computed properties from store
const allPlayersData = computed(() => playerStore.allPlayers)
const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
const currentDatasetId = computed(() => playerStore.currentDatasetId)

// Performance stat categories
const generalStats = [
    { key: 'Av Rat', name: 'Average Rating' },
    { key: 'Apps', name: 'Appearances' },
    { key: 'Mins', name: 'Minutes Played' },
    { key: 'Clean Sheets', name: 'Clean Sheets' }
]

const offensiveStats = [
    { key: 'Gls/90', name: 'Goals per 90' },
    { key: 'xG/90', name: 'Expected Goals per 90' },
    { key: 'NP-xG/90', name: 'Non-Penalty xG per 90' },
    { key: 'Shot/90', name: 'Shots per 90' },
    { key: 'ShT/90', name: 'Shots on Target per 90' },
    { key: 'Conv %', name: 'Conversion %' },
    { key: 'Drb/90', name: 'Dribbles per 90' }
]

const passingStats = [
    { key: 'Asts/90', name: 'Assists per 90' },
    { key: 'xA/90', name: 'Expected Assists per 90' },
    { key: 'Ch C/90', name: 'Chances Created per 90' },
    { key: 'K Ps/90', name: 'Key Passes per 90' },
    { key: 'Ps C/90', name: 'Passes Completed per 90' },
    { key: 'Ps A/90', name: 'Pass Attempts per 90' },
    { key: 'Pas %', name: 'Pass Completion %' },
    { key: 'Pr passes/90', name: 'Progressive Passes per 90' },
    { key: 'Cr C/90', name: 'Crosses Completed per 90' },
    { key: 'CRS A/90', name: 'Crosses Attempted per 90' },
    { key: 'Cr C/A', name: 'Cross Completion %' },
    { key: 'Poss Lost/90', name: 'Possession Lost per 90' }
]

const defensiveStats = [
    { key: 'Tck/90', name: 'Tackles per 90' },
    { key: 'Tck R', name: 'Tackle Ratio %' },
    { key: 'Int/90', name: 'Interceptions per 90' },
    { key: 'Clr/90', name: 'Clearances per 90' },
    { key: 'Blk/90', name: 'Blocks per 90' },
    { key: 'Hdrs W/90', name: 'Headers Won per 90' },
    { key: 'Pres C/90', name: 'Pressures Completed per 90' },
    { key: 'Poss Won/90', name: 'Possession Won per 90' },
    { key: 'Fls', name: 'Fouls' },
    { key: 'FA', name: 'Fouls Against' }
]

const goalkeepingStats = [
    { key: 'Con/90', name: 'Goals Conceded per 90' },
    { key: 'Cln/90', name: 'Clean Sheets per 90' },
    { key: 'xGP/90', name: 'Expected Goals Prevented per 90' },
    { key: 'Sv %', name: 'Save Percentage' }
]

// Helper methods
const getPlayerName = (player) => {
    return player.name || player.Name || player.Player || 'Unknown Player'
}

const getPlayerClub = (player) => {
    return player.club || player.Club || 'Unknown Club'
}

// Methods
const calculateTopPerformers = () => {
    console.log('🔍 Calculating performance stats for', allPlayersData.value.length, 'players')
    
    const allStats = [...generalStats, ...offensiveStats, ...passingStats, ...defensiveStats, ...goalkeepingStats]
    const results = {}
    
    allStats.forEach(stat => {
        // Filter players who have this stat
        const playersWithStat = allPlayersData.value.filter(player => {
            const value = player.attributes?.[stat.key]
            if (value === undefined || value === null || value === '-' || value === '') {
                return false
            }
            
            // Handle comma-separated values and convert to number
            const cleanValue = value.toString().replace(/,/g, '').replace(/%/g, '')
            const numValue = parseFloat(cleanValue)
            
            // For defensive stats that should be lower (like Con/90), include all valid positive values
            // For most other stats, only include values > 0
            if (stat.key === 'Con/90') {
                return !isNaN(numValue) && numValue >= 0
            } else {
                return !isNaN(numValue) && numValue > 0
            }
        })
        
        console.log(`Found ${playersWithStat.length} players with valid ${stat.key} data`)
        
        // Sort by stat value - ascending for "lower is better" stats, descending for others
        const sortedPlayers = playersWithStat.sort((a, b) => {
            const getNumericValue = (val) => {
                const cleaned = val.toString().replace(/,/g, '').replace(/%/g, '')
                return parseFloat(cleaned)
            }
            
            const valueA = getNumericValue(a.attributes[stat.key])
            const valueB = getNumericValue(b.attributes[stat.key])
            
            // For goals conceded, lower is better
            if (stat.key === 'Con/90') {
                return valueA - valueB
            } else {
                return valueB - valueA
            }
        })
        
        results[stat.key] = sortedPlayers.slice(0, 10)
    })
    
    topPlayersByStat.value = results
}

const formatStatValue = (value) => {
    if (value === undefined || value === null || value === '-' || value === '') {
        return 'N/A'
    }
    
    // Convert to string to handle both number and string values
    const stringValue = value.toString()
    
    // Handle comma-separated values (like minutes)
    const cleanValue = stringValue.replace(/,/g, '')
    
    // Convert to number
    const numValue = parseFloat(cleanValue)
    if (isNaN(numValue)) {
        return 'N/A'
    }
    
    // Special formatting for different stat types
    if (stringValue.includes('%') || stringValue.includes('Sv %') || stringValue.includes('Conv %') || stringValue.includes('Pas %') || stringValue.includes('Tck R') || stringValue.includes('Cr C/A')) {
        return numValue.toFixed(1) + '%'
    }
    
    // For large numbers like minutes, add commas
    if (numValue >= 1000 && Number.isInteger(numValue)) {
        return numValue.toLocaleString()
    }
    
    // Round to 2 decimal places for most stats
    if (numValue % 1 === 0) {
        return numValue.toString()
    } else {
        return numValue.toFixed(2)
    }
}

const formatNumber = (num) => {
    return new Intl.NumberFormat().format(num)
}

const openPlayerDetail = (player) => {
    playerForDetailView.value = player
    showPlayerDetailDialog.value = true
}

const shareDataset = () => {
    if (!currentDatasetId.value) return
    
    const shareUrl = `${window.location.origin}/performance/${currentDatasetId.value}`
    
    if (navigator.share) {
        navigator.share({
            title: 'FM Performance Data',
            text: 'Check out these top performance statistics!',
            url: shareUrl
        }).catch(err => console.log('Error sharing:', err))
    } else {
        navigator.clipboard.writeText(shareUrl).then(() => {
            $q.notify({
                message: 'Link copied to clipboard!',
                color: 'positive',
                icon: 'content_copy',
                position: 'top'
            })
        }).catch(err => {
            console.error('Failed to copy link:', err)
            $q.notify({
                message: 'Failed to copy link',
                color: 'negative',
                icon: 'error',
                position: 'top'
            })
        })
    }
}

const initializeData = () => {
    // Check if we have a dataset from route params
    const datasetIdFromRoute = route.params?.datasetId
    const datasetIdFromQuery = route.query?.datasetId
    
    // If we have a dataset ID from route/query, use it
    if (datasetIdFromRoute || datasetIdFromQuery) {
        const targetDatasetId = datasetIdFromRoute || datasetIdFromQuery
        if (targetDatasetId !== sessionStorage.getItem('currentDatasetId')) {
            sessionStorage.setItem('currentDatasetId', targetDatasetId)
            playerStore.fetchPlayersByDatasetId(targetDatasetId)
        }
    } else if (!currentDatasetId.value) {
        // No dataset loaded, show error
        pageLoadingError.value = 'No dataset available. Please upload a dataset first.'
        return
    }
    
    // If we have players data, calculate top performers
    if (allPlayersData.value.length > 0) {
        calculateTopPerformers()
    }
}

// Watchers
watch(allPlayersData, (newPlayers) => {
    if (newPlayers.length > 0) {
        calculateTopPerformers()
        pageLoadingError.value = ''
    }
}, { deep: true })

// Lifecycle
onMounted(() => {
    initializeData()
})
</script>

<style scoped>
.performance-page {
    background: var(--q-color-background);
}

.top-bar {
    background: var(--q-color-surface);
    border-bottom: 1px solid var(--q-color-separator);
    padding: 16px 24px;
    position: sticky;
    top: 0;
    z-index: 100;
}

.top-bar-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    max-width: 1400px;
    margin: 0 auto;
}

.dataset-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.dataset-title {
    font-size: 1.2rem;
    font-weight: 600;
    color: var(--q-color-on-surface);
    display: flex;
    align-items: center;
}

.dataset-stats {
    color: var(--q-color-on-surface-variant);
    font-size: 0.875rem;
}

.stat-item {
    font-weight: 500;
}

.stat-separator {
    opacity: 0.5;
    margin: 0 8px;
}

.top-bar-controls {
    display: flex;
    align-items: center;
    gap: 8px;
}

.share-btn {
    border-radius: 8px;
    width: 36px;
    height: 36px;
}

.error-container,
.no-data-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 400px;
    padding: 48px 24px;
}

.error-banner,
.no-data-banner {
    max-width: 600px;
}

.performance-content {
    padding: 24px;
    max-width: 1400px;
    margin: 0 auto;
}

.categories-container {
    display: flex;
    flex-direction: column;
    gap: 48px;
}

.category-section {
    width: 100%;
}

.category-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--q-color-on-surface);
    margin: 0 0 24px 0;
    display: flex;
    align-items: center;
}

.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 24px;
}

.stat-card {
    height: 100%;
}

.stat-header {
    background: var(--q-color-primary);
    color: white;
    padding: 12px 16px;
}

.stat-name {
    font-weight: 600;
    font-size: 0.9rem;
}

.stat-players {
    padding: 0;
}

.player-item {
    transition: background-color 0.2s ease;
}

.player-item:hover {
    background: var(--q-color-surface-variant);
}

.rank-badge {
    background: var(--q-color-primary);
    color: white;
    border-radius: 50%;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 600;
}

.player-name {
    font-weight: 500;
    color: var(--q-color-on-surface);
}

.stat-value {
    font-weight: 600;
    color: var(--q-color-primary);
    font-size: 0.9rem;
}

.no-data-message {
    padding: 16px;
    text-align: center;
    color: var(--q-color-on-surface-variant);
    font-style: italic;
}

@media (max-width: 768px) {
    .top-bar-content {
        flex-direction: column;
        gap: 16px;
        align-items: stretch;
    }
    
    .stats-grid {
        grid-template-columns: 1fr;
    }
    
    .performance-content {
        padding: 16px;
    }
}
</style> 