<template>
  <q-page padding>
    <div class="q-pa-md">
      <h1 class="text-h4 text-center q-mb-lg">Football Manager HTML Player Parser</h1>
      
      <!-- Instructions Card -->
      <q-card class="q-mb-md bg-blue-1">
        <q-card-section>
          <div class="text-subtitle1 text-weight-bold">Instructions:</div>
          <ol class="q-ml-md">
            <li>Ensure the Go API (main.go) is running (usually on http://localhost:8080).</li>
            <li>Select an HTML file exported from Football Manager containing player data.</li>
            <li>Click "Upload and Parse" to process the file.</li>
            <li>Use search, sort, and pagination to explore the data.</li>
          </ol>
        </q-card-section>
      </q-card>

      <!-- File Upload Section -->
      <q-card class="q-mb-md">
        <q-card-section>
          <div class="text-subtitle1 q-mb-sm">Upload HTML File:</div>
          <q-file
            v-model="playerFile"
            label="Select HTML file"
            accept=".html"
            outlined
            counter
          >
            <template v-slot:prepend>
              <q-icon name="attach_file" />
            </template>
          </q-file>
          
          <q-btn
            class="q-mt-md full-width"
            color="primary"
            label="Upload and Parse"
            :loading="loading"
            :disable="!playerFile"
            @click="uploadAndParse"
          />
        </q-card-section>
      </q-card>

      <!-- Search Filters -->
      <q-card class="q-mb-md" v-if="allPlayers.length > 0">
        <q-card-section>
          <div class="text-subtitle1 q-mb-sm">Search Players</div>
          <div class="row q-col-gutter-md">
            <div class="col-12 col-sm-6 col-md-3">
              <q-input
                v-model="filters.name"
                label="Player Name"
                dense
                outlined
                clearable
                @update:model-value="handleSearch"
              />
            </div>
            <div class="col-12 col-sm-6 col-md-3">
              <q-input
                v-model="filters.club"
                label="Club"
                dense
                outlined
                clearable
                @update:model-value="handleSearch"
              />
            </div>
            <div class="col-12 col-sm-6 col-md-3">
              <q-input
                v-model="filters.transferValue"
                label="Transfer Value"
                dense
                outlined
                clearable
                placeholder="e.g., €1.5M, >1M, <500K"
                @update:model-value="handleSearch"
              />
            </div>
            <div class="col-12 col-sm-6 col-md-3 flex items-end">
              <div class="row q-col-gutter-sm full-width">
                <div class="col">
                  <q-btn
                    color="grey"
                    label="Clear Filters"
                    class="full-width"
                    @click="clearSearch"
                    :disable="!hasActiveFilters"
                  />
                </div>
              </div>
            </div>
          </div>
        </q-card-section>
      </q-card>

      <!-- Error Message -->
      <q-banner v-if="error" class="bg-negative text-white q-mb-md" rounded>
        {{ error }}
        <template v-slot:action>
          <q-btn flat color="white" label="Dismiss" @click="error = ''" />
        </template>
      </q-banner>

      <!-- Players Table with Stats -->
      <template v-if="allPlayers.length > 0">
        <div class="row q-col-gutter-md q-mb-md">
          <div class="col-12 col-md-3">
            <q-card class="text-center">
              <q-card-section>
                <div class="text-h6">{{ allPlayers.length }}</div>
                <div class="text-subtitle2">Total Players</div>
              </q-card-section>
            </q-card>
          </div>
          <div class="col-12 col-md-3">
            <q-card class="text-center">
              <q-card-section>
                <div class="text-h6">{{ filteredPlayers.length }}</div>
                <div class="text-subtitle2">Filtered Players</div>
              </q-card-section>
            </q-card>
          </div>
          <div class="col-12 col-md-3">
            <q-card class="text-center">
              <q-card-section>
                <div class="text-h6">{{ uniqueClubs }}</div>
                <div class="text-subtitle2">Unique Clubs</div>
              </q-card-section>
            </q-card>
          </div>
          <div class="col-12 col-md-3">
            <q-card class="text-center">
              <q-card-section>
                <div class="text-h6">{{ uniquePositions }}</div>
                <div class="text-subtitle2">Unique Positions</div>
              </q-card-section>
            </q-card>
          </div>
        </div>
        
        <PlayerDataTable
          :players="filteredPlayers"
          :loading="loading"
          @update:sort="handleSort"
        />
      </template>

      <!-- Empty State -->
      <q-card v-else-if="!loading" class="q-pa-lg text-center">
        <q-icon name="upload_file" size="4rem" color="grey-7" />
        <div class="text-h6 q-mt-md">No Player Data Yet</div>
        <div class="text-grey-7">Upload a file to see player data</div>
      </q-card>
    </div>
  </q-page>
</template>

<script>
import { ref, computed, reactive, watch } from 'vue'
import PlayerDataTable from '../components/PlayerDataTable.vue'
import playerService from '../services/playerService'

export default {
  name: 'PlayerUploadPage',
  components: {
    PlayerDataTable
  },
  
  setup() {
    const playerFile = ref(null)
    const loading = ref(false)
    const error = ref('')
    const allPlayers = ref([])
    
    // Sorting state
    const sortState = reactive({
      key: null,
      direction: 'asc',
      isAttribute: false,
      displayField: null
    })
    
    // Search filters
    const filters = reactive({
      name: '',
      club: '',
      transferValue: ''
    })
    
    // Check if any filters are active
    const hasActiveFilters = computed(() => {
      return filters.name !== '' || filters.club !== '' || filters.transferValue !== ''
    })
    
    // Calculate unique clubs
    const uniqueClubs = computed(() => {
      const clubs = new Set()
      allPlayers.value.forEach(player => {
        if (player.club) clubs.add(player.club)
      })
      return clubs.size
    })
    
    // Calculate unique positions
    const uniquePositions = computed(() => {
      const positions = new Set()
      allPlayers.value.forEach(player => {
        if (player.position) positions.add(player.position)
      })
      return positions.size
    })
    
    // Helper to parse monetary values (€1.5M, £500K, etc.) into integers for sorting
    const parseMonetaryValue = (valueStr) => {
      if (typeof valueStr !== 'string' || !valueStr) return 0
      
      // Remove any text after p/w (per week)
      const cleanedStr = valueStr.split(' p/w')[0]
      
      // Always log for debugging during this fix
      console.log(`PARSE: "${valueStr}" → "${cleanedStr}"`)
      
      // Extract multiplier suffix (M or K) before removing non-numeric chars
      // This ensures we catch the suffix even if it's not directly attached to the number
      let multiplier = 1
      const lowerStr = cleanedStr.toLowerCase()
      
      if (lowerStr.includes('m')) {
        multiplier = 1000000
      } else if (lowerStr.includes('k')) {
        multiplier = 1000
      }
      
      // Step 1: Remove all currency symbols (£, €, $) and other non-numeric chars except commas, periods, and digits
      let numStr = cleanedStr.replace(/[^0-9,.]/g, '')
      
      // Step 2: Handle numbers with commas
      // If there are commas in the number but no decimal points
      if (numStr.includes(',') && !numStr.includes('.')) {
        // Check if it's a thousands separator pattern (e.g., "1,234,567")
        // This heuristic checks if commas appear every 3 digits from the end
        const parts = numStr.split(',')
        let isThousandsSeparator = true
        
        // Check that most parts (except possibly the first) have length 3
        for (let i = 1; i < parts.length; i++) {
          if (parts[i].length !== 3) {
            isThousandsSeparator = false
            break
          }
        }
        
        if (isThousandsSeparator) {
          // If it looks like thousands separators, remove all commas
          numStr = numStr.replace(/,/g, '')
          console.log(`  Treating commas as thousands separators: "${numStr}"`)
        } else {
          // If not a clear thousands separator pattern, try treating the last comma as decimal
          numStr = numStr.replace(/,([^,]*)$/, '.$1')
          console.log(`  Treating last comma as decimal: "${numStr}"`)
        }
      } else if (numStr.includes(',')) {
        // If there are both commas and periods, assume comma is for thousands
        numStr = numStr.replace(/,/g, '')
        console.log(`  Mixed decimals and commas, treating commas as thousands: "${numStr}"`)
      }
      
      // Step 3: Parse the numeric value
      const numericValue = parseFloat(numStr)
      
      // Step 4: Calculate result (using Math.round to avoid floating point issues)
      const result = Math.round(isNaN(numericValue) ? 0 : numericValue * multiplier)
      
      // Always log the result for debugging
      console.log(`  Final result: "${valueStr}" → ${result} (multiplier: ${multiplier}, numericValue: ${numericValue})`)
      
      return result
    }
    
    // Filtered players based on search criteria
    const filteredPlayers = computed(() => {
      if (!allPlayers.value.length) return []
      
      const filtered = allPlayers.value.filter(player => {
        // Name filter
        const nameMatch = !filters.name ||
          (player.name && player.name.toLowerCase().includes(filters.name.toLowerCase()))
        
        // Club filter
        const clubMatch = !filters.club ||
          (player.club && player.club.toLowerCase().includes(filters.club.toLowerCase()))
        
        // Transfer value filter
        let transferValueMatch = true
        if (filters.transferValue) {
          // Value to compare against (the filter value)
          let compareValue = 0
          let operator = 'includes'  // default operator
          
          // Check for operators
          if (filters.transferValue.startsWith('>')) {
            operator = '>'
            compareValue = parseMonetaryValue(filters.transferValue.substring(1))
          } else if (filters.transferValue.startsWith('<')) {
            operator = '<'
            compareValue = parseMonetaryValue(filters.transferValue.substring(1))
          } else {
            // Text-based search
            operator = 'includes'
            const playerValueStr = String(player.transfer_value || '').toLowerCase()
            transferValueMatch = playerValueStr.includes(filters.transferValue.toLowerCase())
            return nameMatch && clubMatch && transferValueMatch
          }
          
          // Numeric comparison for > and < operators
          const playerValue = player.transferValueAmount || 0
          if (operator === '>') {
            transferValueMatch = playerValue > compareValue
          } else if (operator === '<') {
            transferValueMatch = playerValue < compareValue
          }
        }
        
        return nameMatch && clubMatch && transferValueMatch
      })
      
      // Sort the filtered players if a sort key is set
      if (sortState.key) {
        return sortPlayers([...filtered])
      }
      
      return filtered
    })
    
    // Process player data - convert string values to appropriate types
    const processPlayerData = (players) => {
      return players.map(player => {
        // Parse monetary values first to ensure they're integers
        const transferValue = parseMonetaryValue(player.transfer_value)
        const wageValue = parseMonetaryValue(player.wage)
        
        // Create a new player object with processed data
        const processedPlayer = {
          ...player,
          // Ensure age is a number
          age: parseInt(player.age, 10) || 0,
          // Store INTEGER values for sorting but keep display value
          transferValueAmount: transferValue,
          wageAmount: wageValue,
          // Keep other fields as they are
          attributes: { ...player.attributes }
        }
        
        // Process attributes - convert all numeric attributes to numbers
        if (processedPlayer.attributes) {
          Object.keys(processedPlayer.attributes).forEach(key => {
            const value = processedPlayer.attributes[key]
            if (value && !isNaN(parseInt(value, 10))) {
              processedPlayer.attributes[key] = parseInt(value, 10)
            }
          })
        }
        
        return processedPlayer
      })
    }
    
    // Debug function to log players
    const logPlayerData = () => {
      if (allPlayers.value.length > 0) {
        const player = allPlayers.value[0]
        console.log('First player:', player)
        console.log(`Transfer value: ${player.transfer_value} → ${player.transferValueAmount} (${typeof player.transferValueAmount})`)
        console.log(`Wage: ${player.wage} → ${player.wageAmount} (${typeof player.wageAmount})`)
        
        // Get second player for comparison if available
        if (allPlayers.value.length > 1) {
          const player2 = allPlayers.value[1]
          console.log('\nSecond player:', player2)
          console.log(`Transfer value: ${player2.transfer_value} → ${player2.transferValueAmount} (${typeof player2.transferValueAmount})`)
          console.log(`Wage: ${player2.wage} → ${player2.wageAmount} (${typeof player2.wageAmount})`)
        }
        
        // Log unique monetary values to check distribution
        const values = allPlayers.value
          .filter(p => p.transferValueAmount > 0)  // Filter out zero values
          .map(p => ({ 
            name: p.name, 
            display: p.transfer_value, 
            value: p.transferValueAmount 
          }))
          .sort((a, b) => b.value - a.value)  // Sort by value, high to low
          .slice(0, 10);  // Take top 10
        
        console.log('Top 10 transfer values:');
        values.forEach(v => console.log(`${v.name}: ${v.display} → ${v.value}`));
      }
    }
    
    // Upload and parse the player data
    const uploadAndParse = async () => {
      if (!playerFile.value) {
        error.value = 'Please select an HTML file first.'
        return
      }
      
      loading.value = true
      error.value = ''
      
      try {
        const formData = new FormData()
        formData.append('playerFile', playerFile.value)
        
        const response = await playerService.uploadPlayerFile(formData)
        
        // Process player data - convert strings to appropriate types
        const processedPlayers = processPlayerData(response)
        
        allPlayers.value = processedPlayers
        
        // Log for debugging
        logPlayerData()
        
        // Reset sort and filters on new upload
        sortState.key = null
        sortState.direction = 'asc'
        sortState.isAttribute = false
        sortState.displayField = null
        clearSearch()
        
      } catch (err) {
        error.value = `Failed to parse player data: ${err.message || 'Unknown error'}`
        allPlayers.value = []
      } finally {
        loading.value = false
      }
    }
    
    // Handle sorting
    const handleSort = (sortParams) => {
      // Update sort state with values from the event
      sortState.key = sortParams.key // The actual field to sort by (could be transferValueAmount)
      sortState.direction = sortParams.direction
      sortState.isAttribute = sortParams.isAttribute
      sortState.displayField = sortParams.displayField // For UI highlighting (might be transfer_value)
    }
    
    // Sort players based on current sort state
    const sortPlayers = (players) => {
      if (!sortState.key) return players
      
      // Log the sort operation for debugging
      console.log(`Sorting by: ${sortState.key} (${sortState.direction})`)
      
      // Double-check the top values for monetary fields
      if (sortState.key === 'transferValueAmount' || sortState.key === 'wageAmount') {
        const displayKey = sortState.key === 'transferValueAmount' ? 'transfer_value' : 'wage'
        const top5 = [...players]
          .filter(p => p[sortState.key] > 0)
          .sort((a, b) => b[sortState.key] - a[sortState.key])
          .slice(0, 5);
        
        console.log(`Top 5 ${sortState.key} values before sorting:`)
        top5.forEach(p => console.log(`${p.name}: ${p[displayKey]} → ${p[sortState.key]}`))
      }
      
      return players.sort((a, b) => {
        let valA, valB
        
        // Special handling for monetary fields (transferValueAmount and wageAmount)
        if (sortState.key === 'transferValueAmount' || sortState.key === 'wageAmount') {
          // These should be integers from processing
          valA = typeof a[sortState.key] === 'number' ? a[sortState.key] : 0
          valB = typeof b[sortState.key] === 'number' ? b[sortState.key] : 0
          
          // Occasionally log examples of the comparison for debugging
          if (Math.random() < 0.01) { // Log ~1% of comparisons
            const displayKey = sortState.key === 'transferValueAmount' ? 'transfer_value' : 'wage'
            console.log(`Comparing ${a.name} (${a[displayKey]} → ${valA}) vs ${b.name} (${b[displayKey]} → ${valB})`)
          }
          
          // Direct numeric comparison (faster for integers)
          return sortState.direction === 'asc' ? valA - valB : valB - valA
        }
        // Handle attribute fields
        else if (sortState.isAttribute) {
          valA = a.attributes ? a.attributes[sortState.key] : null
          valB = b.attributes ? b.attributes[sortState.key] : null
        }
        // Handle other fields
        else {
          valA = a[sortState.key]
          valB = b[sortState.key]
        }
        
        // Handle null/undefined values
        if (valA == null && valB == null) return 0
        if (valA == null) return sortState.direction === 'asc' ? 1 : -1
        if (valB == null) return sortState.direction === 'asc' ? -1 : 1
        
        // If both values are numbers, compare them directly
        if (typeof valA === 'number' && typeof valB === 'number') {
          return sortState.direction === 'asc' ? valA - valB : valB - valA
        }
        
        // Try to convert to number if they look like numbers
        if (!isNaN(parseFloat(valA)) && !isNaN(parseFloat(valB))) {
          const numA = parseFloat(valA)
          const numB = parseFloat(valB)
          return sortState.direction === 'asc' ? numA - numB : numB - numA
        }
        
        // String comparison
        if (typeof valA === 'string' && typeof valB === 'string') {
          valA = valA.toLowerCase()
          valB = valB.toLowerCase()
          if (valA < valB) return sortState.direction === 'asc' ? -1 : 1
          if (valA > valB) return sortState.direction === 'asc' ? 1 : -1
          return 0
        }
        
        // Fallback for mixed types
        const strA = String(valA).toLowerCase()
        const strB = String(valB).toLowerCase()
        if (strA < strB) return sortState.direction === 'asc' ? -1 : 1
        if (strA > strB) return sortState.direction === 'asc' ? 1 : -1
        return 0
      })
    }
    
    // Clear search filters
    const clearSearch = () => {
      filters.name = ''
      filters.club = ''
      filters.transferValue = ''
    }
    
    // Handle search (nothing needed here - computed property does the filtering)
    const handleSearch = () => {}
    
    return {
      playerFile,
      loading,
      error,
      allPlayers,
      filteredPlayers,
      uniqueClubs,
      uniquePositions,
      filters,
      hasActiveFilters,
      uploadAndParse,
      handleSort,
      handleSearch,
      clearSearch
    }
  }
}
</script>

<style>
.q-page {
  max-width: 1600px;
  margin: 0 auto;
}
</style>