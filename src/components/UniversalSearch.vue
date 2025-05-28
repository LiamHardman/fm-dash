<template>
  <div class="universal-search">
    <q-input
      v-model="searchQuery"
      filled
      dense
      :placeholder="hasDatasetId ? 'Search players, teams, leagues, nations...' : 'Upload a dataset first to search'"
      :disable="!hasDatasetId"
      class="search-input"
      @input="onSearchInput"
      @keyup="onKeyUp"
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
      v-if="showResults && (results.length > 0 || isLoading)"
      class="search-results"
      flat
      bordered
    >
      <q-card-section v-if="isLoading" class="text-center">
        <q-spinner size="sm" />
        <div class="text-caption q-mt-xs">Searching...</div>
      </q-card-section>
      
      <q-list v-else-if="results.length > 0" separator>
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
            <q-item-label caption>{{ result.description }}</q-item-label>
          </q-item-section>
          
          <q-item-section side>
            <q-chip :color="getResultColor(result.type)" text-color="white" size="sm">
              {{ result.type }}
            </q-chip>
          </q-item-section>
        </q-item>
      </q-list>
      
      <q-card-section v-else class="text-center text-grey-6">
        <div class="text-caption">No results found</div>
      </q-card-section>
    </q-card>
  </div>
</template>

<script>
import { defineComponent, ref, computed, watch, nextTick, onMounted } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import { debounce } from '../utils/debounce'

export default defineComponent({
  name: 'UniversalSearch',
  setup() {
    console.log('🔍 UniversalSearch component is being created!');
    const playerStore = usePlayerStore()
    const searchQuery = ref('')
    const results = ref([])
    const isLoading = ref(false)
    const searchInput = ref(null)
    
    console.log('UniversalSearch: Component setup started')
    console.log('UniversalSearch: Current dataset ID:', playerStore.currentDatasetId)
    
    const showSearch = computed(() => {
      const shouldShow = !!playerStore.currentDatasetId
      console.log('UniversalSearch: showSearch computed:', shouldShow, 'datasetId:', playerStore.currentDatasetId)
      return shouldShow
    })
    const showResults = computed(() => searchQuery.value.length > 0)
    const hasDatasetId = computed(() => !!playerStore.currentDatasetId)
    
    const searchAPI = async (query) => {
      console.log('🔍 searchAPI called with:', {
        query: query.trim(),
        datasetId: playerStore.currentDatasetId,
        queryLength: query.trim().length,
        hasDatasetId: !!playerStore.currentDatasetId
      });
      
      if (!query.trim()) {
        console.log('🔍 Search skipped: empty query')
        return []
      }
      
      if (!playerStore.currentDatasetId) {
        console.log('🔍 Search skipped: no dataset ID available')
        console.log('🔍 Player store current state:', {
          currentDatasetId: playerStore.currentDatasetId,
          hasPlayers: playerStore.allPlayers?.length > 0
        })
        return []
      }
      
      const url = `/api/search/${playerStore.currentDatasetId}?q=${encodeURIComponent(query)}`
      console.log('🔍 Making search request to:', url)
      
      try {
        const response = await fetch(url)
        console.log('🔍 Search response status:', response.status)
        
        if (response.ok) {
          const data = await response.json()
          console.log('🔍 Search results received:', data.length, 'items')
          return data
        } else {
          const errorText = await response.text()
          console.error('🔍 Search API error:', response.status, errorText)
        }
      } catch (error) {
        console.error('🔍 Search network error:', error)
      }
      return []
    }
    
    const debouncedSearch = debounce(async (query) => {
      console.log('🔍 debouncedSearch called with query:', query)
      
      if (!query.trim()) {
        console.log('🔍 debouncedSearch: clearing results for empty query')
        results.value = []
        isLoading.value = false
        return
      }
      
      console.log('🔍 debouncedSearch: starting search...')
      isLoading.value = true
      results.value = await searchAPI(query)
      isLoading.value = false
      console.log('🔍 debouncedSearch: search completed')
    }, 300)
    
    const onSearchInput = () => {
      console.log('🔍 onSearchInput called with query:', searchQuery.value)
      console.log('🔍 onSearchInput: hasDatasetId:', !!playerStore.currentDatasetId)
      debouncedSearch(searchQuery.value)
    }
    
    const onKeyUp = () => {
      console.log('🔍 onKeyUp called with query:', searchQuery.value)
      console.log('🔍 onKeyUp: hasDatasetId:', !!playerStore.currentDatasetId)
      debouncedSearch(searchQuery.value)
    }
    
    const clearSearch = () => {
      searchQuery.value = ''
      results.value = []
      isLoading.value = false
    }
    
    const getResultIcon = (type) => {
      switch (type) {
        case 'player': return 'person'
        case 'team': return 'groups'
        case 'league': return 'emoji_events'
        case 'nation': return 'flag'
        default: return 'search'
      }
    }
    
    const getResultColor = (type) => {
      switch (type) {
        case 'player': return 'blue'
        case 'team': return 'green'
        case 'league': return 'orange'
        case 'nation': return 'red'
        default: return 'grey'
      }
    }
    
    const handleResultClick = (result) => {
      // Generate proper URL with current dataset ID
      let url = result.url
      
      // Replace empty dataset ID placeholder with actual dataset ID
      if (url.includes('/dataset/')) {
        url = url.replace('/dataset/', `/dataset/${playerStore.currentDatasetId}`)
      }
      
      // Handle different result types with appropriate navigation
      if (result.type === 'player') {
        // Navigate to dataset page with player search filter
        url = `/dataset/${playerStore.currentDatasetId}?search=${encodeURIComponent(result.name)}`
      } else if (result.type === 'team') {
        // Navigate to dataset page with team filter
        url = `/dataset/${playerStore.currentDatasetId}?team=${encodeURIComponent(result.name)}`
      } else if (result.type === 'league') {
        // Navigate to leagues page with league filter
        url = `/leagues?league=${encodeURIComponent(result.name)}`
      } else if (result.type === 'nation') {
        // Navigate to nations page with nation filter  
        url = `/nations?nation=${encodeURIComponent(result.name)}`
      }
      
      // Use router navigation instead of opening new tab for better UX
      // TODO: Implement router navigation when router is available
      // For now, navigate in the same tab
      window.location.href = url
      clearSearch()
    }
    
    // Focus search input when dataset changes
    watch(() => playerStore.currentDatasetId, (newId, oldId) => {
      console.log('UniversalSearch: Dataset ID changed from', oldId, 'to', newId)
      if (newId) {
        nextTick(() => {
          if (searchInput.value) {
            searchInput.value.focus()
          }
        })
      }
    })
    
    // Watch searchQuery changes to debug v-model
    watch(searchQuery, (newQuery, oldQuery) => {
      console.log('🔍 searchQuery watcher:', { newQuery, oldQuery })
    })
    
    // Add onMounted to see if component is being created
    onMounted(() => {
      console.log('UniversalSearch: Component mounted, dataset ID:', playerStore.currentDatasetId)
    })
    
    return {
      searchQuery,
      results,
      isLoading,
      showSearch,
      showResults,
      searchInput,
      onSearchInput,
      onKeyUp,
      clearSearch,
      getResultIcon,
      getResultColor,
      handleResultClick,
      hasDatasetId
    }
  }
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
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }
  
  .search-result-item {
    &:hover {
      background-color: rgba(0, 0, 0, 0.05);
      
      .body--dark & {
        background-color: rgba(255, 255, 255, 0.05);
      }
    }
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