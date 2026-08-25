<template>
  <div class="universal-search">
    <q-input
      v-model="searchQuery"
      filled
      dense
      :placeholder="hasDatasetId ? 'Search players, teams, leagues, nations...' : 'Upload a dataset first to search'"
      :disable="!hasDatasetId"
      class="search-input"
      @keyup.escape="clearSearch"
      ref="searchInput"
    >
      <template v-slot:prepend>
        <q-icon name="search" />
      </template>
      <template v-slot:append>
        <q-btn
          v-if="searchQuery"
          flat
          round
          dense
          icon="clear"
          @click="clearSearch"
          size="sm"
        />
      </template>
    </q-input>
    
    <q-card
      v-if="showResults"
      class="search-results"
      flat
      bordered
    >
      <q-card-section v-if="isLoading" class="text-center">
        <q-spinner size="sm" />
        <div class="text-caption q-mt-xs">Searching...</div>
      </q-card-section>
      
      <q-list v-else-if="results && results.length > 0" separator>
        <q-item
          v-for="result in results"
          :key="`${result.type}-${result.id}`"
          clickable
          @click="handleResultClick(result)"
          class="search-result-item"
        >
          <q-item-section avatar>
            <q-icon :name="getResultIcon(result.type)" :color="getResultColor(result.type)" />
          </q-item-section>
          
          <q-item-section>
            <q-item-label>{{ result.name }}</q-item-label>
            <q-item-label caption>{{ result.description || result.subText }}</q-item-label>
          </q-item-section>
          
          <q-item-section side class="search-result-side">
            <!-- Show overall rating for players with the same styling as PlayerDataTable -->
            <div v-if="result.type === 'player' && result.overall" class="player-rating-container">
              <span 
                :class="getUnifiedRatingClass(result.overall, 100)"
                class="attribute-value fifa-stat-value search-result-rating"
              >
                {{ result.overall }}
              </span>
            </div>
            <q-chip v-else :color="getResultColor(result.type)" text-color="white" size="sm">
              {{ result.type }}
            </q-chip>
          </q-item-section>
        </q-item>
      </q-list>
      
      <q-card-section v-else class="text-center text-grey-6">
        <div class="text-caption">No results found</div>
      </q-card-section>
    </q-card>

    <!-- Player Detail Dialog -->
    <PlayerDetailDialog
      :player="playerForDetailView"
      :show="showPlayerDetailDialog"
      @close="showPlayerDetailDialog = false"
      :currency-symbol="detectedCurrencySymbol"
      :dataset-id="currentDatasetId"
    />
  </div>
</template>

<script>
import { computed, defineComponent, nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/playerStore'
import { debounce } from '../utils/debounce'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

export default defineComponent({
  name: 'UniversalSearch',
  components: {
    PlayerDetailDialog,
  },
  setup() {
    const router = useRouter()
    const playerStore = usePlayerStore()
    const searchQuery = ref('')
    const results = ref([])
    const isLoading = ref(false)
    const searchInput = ref(null)
    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)

    const showResults = computed(() => searchQuery.value.length > 0)
    const hasDatasetId = computed(() => !!playerStore.currentDatasetId)
    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol || '$')

    // Request cancellation support
    let currentSearchController = null
    let currentSearchId = 0

    const searchAPI = async (query, signal) => {
      if (!query.trim() || !playerStore.currentDatasetId) {
        return []
      }

      const url = `/api/search/${playerStore.currentDatasetId}?q=${encodeURIComponent(query)}`

      // Performance debugging
      const searchStart = performance.now()
      console.log(`🔍 Starting search for "${query}"`)

      try {
        const fetchStart = performance.now()
        const response = await fetch(url, { signal })
        const fetchTime = performance.now() - fetchStart

        if (response.ok) {
          const parseStart = performance.now()
          const data = await response.json()
          const parseTime = performance.now() - parseStart
          const totalTime = performance.now() - searchStart

          console.log(`✅ Search completed for "${query}":`, {
            totalTime: `${totalTime.toFixed(1)}ms`,
            fetchTime: `${fetchTime.toFixed(1)}ms`,
            parseTime: `${parseTime.toFixed(1)}ms`,
            resultCount: Array.isArray(data) ? data.length : 0,
          })

          // Ensure we always return an array
          return Array.isArray(data) ? data : []
        }
        return []
      } catch (error) {
        if (error.name === 'AbortError') {
          console.log(`⏹️ Search aborted for "${query}"`)
          return []
        }
        console.error(`❌ Search failed for "${query}":`, error)
        return []
      }
    }

    // Create stable debounced function with cancellation support
    const debouncedSearchFn = debounce(async (query) => {
      // Generate unique ID for this search
      const searchId = ++currentSearchId

      // Cancel previous request if it exists
      if (currentSearchController) {
        currentSearchController.abort()
      }

      if (!query.trim()) {
        results.value = []
        isLoading.value = false
        currentSearchController = null
        return
      }

      // Create new AbortController for this request
      currentSearchController = new AbortController()
      const signal = currentSearchController.signal

      // Only set loading if this is still the latest search
      if (searchId === currentSearchId) {
        isLoading.value = true
      }

      try {
        const searchResults = await searchAPI(query, signal)

        // Only update results if this is still the latest search and wasn't aborted
        if (searchId === currentSearchId && !signal.aborted) {
          results.value = Array.isArray(searchResults) ? searchResults : []
        }
      } catch (error) {
        if (error.name !== 'AbortError') {
          // Only update results if this is still the latest search
          if (searchId === currentSearchId) {
            results.value = []
          }
        }
      } finally {
        // Only clear loading state if this is still the latest search
        if (searchId === currentSearchId) {
          isLoading.value = false
          currentSearchController = null
        }
      }
    }, 150)

    // Watch searchQuery and trigger debounced search
    watch(searchQuery, (newQuery) => {
      debouncedSearchFn(newQuery)
    })

    const clearSearch = () => {
      searchQuery.value = ''
      results.value = []
      isLoading.value = false
    }

    const getResultIcon = (type) => {
      switch (type) {
        case 'player':
          return 'person'
        case 'team':
          return 'groups'
        case 'league':
          return 'emoji_events'
        case 'nation':
          return 'flag'
        default:
          return 'search'
      }
    }

    const getResultColor = (type) => {
      switch (type) {
        case 'player':
          return 'blue'
        case 'team':
          return 'green'
        case 'league':
          return 'orange'
        case 'nation':
          return 'red'
        default:
          return 'grey'
      }
    }

    // Unified rating class function (same as used in PlayerDataTable)
    const getUnifiedRatingClass = (value, maxScale) => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue) || value === null || value === undefined || value === '-')
        return 'rating-na'
      const percentage = (numValue / maxScale) * 100
      if (percentage >= 90) return 'rating-tier-6'
      if (percentage >= 80) return 'rating-tier-5'
      if (percentage >= 70) return 'rating-tier-4'
      if (percentage >= 55) return 'rating-tier-3'
      if (percentage >= 40) return 'rating-tier-2'
      return 'rating-tier-1'
    }

    // Create a fast lookup cache for players by name
    const playerNameCache = computed(() => {
      if (!playerStore.allPlayers) return new Map()

      const cache = new Map()
      for (const player of playerStore.allPlayers) {
        if (player.name) {
          cache.set(player.name.toLowerCase(), player)
        }
      }
      return cache
    })

    const findPlayerByName = (playerName) => {
      return playerNameCache.value.get(playerName.toLowerCase())
    }

    const handleResultClick = (result) => {
      if (result.type === 'player') {
        // Find the full player object and open detail dialog
        const player = findPlayerByName(result.name)
        if (player) {
          playerForDetailView.value = player
          showPlayerDetailDialog.value = true
        } else {
          // Fallback: navigate to dataset page with search filter
          router.push({
            path: `/dataset/${playerStore.currentDatasetId}`,
            query: { search: result.name },
          })
        }
      } else if (result.type === 'team') {
        // Navigate to team view page
        const url = router.resolve({
          path: '/team-view',
          query: {
            datasetId: playerStore.currentDatasetId,
            team: result.name,
          },
        }).href
        window.open(url, '_blank')
      } else if (result.type === 'league') {
        // Navigate to leagues page with league filter
        router.push({
          path: `/leagues/${playerStore.currentDatasetId}`,
          query: { league: result.name },
        })
      } else if (result.type === 'nation') {
        // Navigate to nations page with nation filter
        router.push({
          path: `/nations/${playerStore.currentDatasetId}`,
          query: { nation: result.name },
        })
      }

      clearSearch()
    }

    // Focus search input when dataset changes
    watch(
      () => playerStore.currentDatasetId,
      (newId) => {
        if (newId) {
          nextTick(() => {
            if (searchInput.value) {
              searchInput.value.focus()
            }
          })
        }
      }
    )

    return {
      searchQuery,
      results,
      isLoading,
      showResults,
      searchInput,
      clearSearch,
      getResultIcon,
      getResultColor,
      getUnifiedRatingClass,
      handleResultClick,
      hasDatasetId,
      playerForDetailView,
      showPlayerDetailDialog,
      currentDatasetId,
      detectedCurrencySymbol,
    }
  },
})
</script>

<style lang="scss" scoped>
.universal-search {
  position: relative;
  width: 300px;
  
  .search-input {
    width: 100%;
  }
  
  .search-results {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    z-index: 1000;
    max-height: 400px;
    overflow-y: auto;
    margin-top: 4px;
    background: var(--surface-card);
    border-color: var(--surface-border);
    box-shadow: var(--shadow-3);
  }

  .search-result-item {
    &:hover {
      background-color: var(--surface-raised);
    }
  }

  .search-result-side {
    flex: 0 0 auto;
    align-items: center;
  }

  .player-rating-container {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .search-result-rating {
    font-size: 0.9rem !important;
    padding: 2px 6px !important;
    min-width: 28px !important;
    border-radius: 4px !important;
    border: 1px solid var(--surface-border);
  }
}

@media (max-width: 768px) {
  .universal-search {
    width: 250px;
  }
}

@media (max-width: 480px) {
  .universal-search {
    width: 200px;
  }
}
</style>