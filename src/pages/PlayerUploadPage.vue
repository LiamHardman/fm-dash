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
              />
            </div>
            <div class="col-12 col-sm-6 col-md-3">
              <q-input
                v-model="filters.club"
                label="Club"
                dense
                outlined
                clearable
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
              />
            </div>
            <div class="col-12 col-sm-6 col-md-3 flex items-end">
              <div class="row q-col-gutter-sm full-width">
                <div class="col">
                  <q-btn
                    color="primary"
                    label="Search"
                    class="full-width"
                    @click="handleSearch"
                  />
                </div>
                <div class="col">
                  <q-btn
                    color="grey"
                    label="Clear"
                    class="full-width"
                    @click="clearSearch"
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

      <!-- Players Table -->
      <template v-if="allPlayers.length > 0">
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
import { ref, computed, reactive } from 'vue'
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
      isAttribute: false
    })
    
    // Search filters
    const filters = reactive({
      name: '',
      club: '',
      transferValue: ''
    })
    
    // Filtered players based on search criteria
    const filteredPlayers = computed(() => {
      if (!allPlayers.value.length) return []
      
      return allPlayers.value.filter(player => {
        // Name filter
        const nameMatch = !filters.name ||
          (player.name && player.name.toLowerCase().includes(filters.name.toLowerCase()))
        
        // Club filter
        const clubMatch = !filters.club ||
          (player.club && player.club.toLowerCase().includes(filters.club.toLowerCase()))
        
        // Transfer value filter
        let transferValueMatch = true
        if (filters.transferValue) {
          const playerValueStr = String(player.transfer_value || '').toLowerCase()
          if (filters.transferValue.startsWith('>')) {
            const val = parseFloat(filters.transferValue.substring(1).replace(/[^0-9.]/g, ''))
            const playerVal = parseFloat(String(player.transfer_value || '0').replace(/[^0-9.]/g, ''))
            transferValueMatch = playerVal > val
          } else if (filters.transferValue.startsWith('<')) {
            const val = parseFloat(filters.transferValue.substring(1).replace(/[^0-9.]/g, ''))
            const playerVal = parseFloat(String(player.transfer_value || '0').replace(/[^0-9.]/g, ''))
            transferValueMatch = playerVal < val
          } else {
            transferValueMatch = playerValueStr.includes(filters.transferValue.toLowerCase())
          }
        }
        
        return nameMatch && clubMatch && transferValueMatch
      })
    })
    
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
        
        allPlayers.value = response.map(p => ({
          ...p,
          attributes: p.attributes && typeof p.attributes === 'object' ? p.attributes : {}
        }))
        
        // Reset sort and filters on new upload
        sortState.key = null
        sortState.direction = 'asc'
        sortState.isAttribute = false
        clearSearch()
        
      } catch (err) {
        error.value = `Failed to parse player data: ${err.message || 'Unknown error'}`
        allPlayers.value = []
      } finally {
        loading.value = false
      }
    }
    
    // Handle sorting
    const handleSort = ({ key, isAttribute }) => {
      if (sortState.key === key) {
        // Toggle direction if same key
        sortState.direction = sortState.direction === 'asc' ? 'desc' : 'asc'
      } else {
        // New sort key
        sortState.key = key
        sortState.direction = 'asc'
        sortState.isAttribute = isAttribute
      }
      
      // Sort the players
      sortPlayers()
    }
    
    // Sort players based on current sort state
    const sortPlayers = () => {
      if (!sortState.key) return
      
      allPlayers.value.sort((a, b) => {
        let valA, valB
        
        if (sortState.isAttribute) {
          valA = a.attributes ? a.attributes[sortState.key] : null
          valB = b.attributes ? b.attributes[sortState.key] : null
        } else {
          valA = a[sortState.key]
          valB = b[sortState.key]
        }
        
        // Age sorting (numeric)
        if (sortState.key === 'age') {
          valA = parseInt(valA, 10) || 0
          valB = parseInt(valB, 10) || 0
        }
        // Transfer Value and Wage sorting (monetary)
        else if (sortState.key === 'transfer_value' || sortState.key === 'wage') {
          valA = parseMonetaryValue(String(valA))
          valB = parseMonetaryValue(String(valB))
        }
        // Numeric attribute sorting
        else if (sortState.isAttribute || (!isNaN(parseFloat(valA)) && !isNaN(parseFloat(valB)))) {
          const numA = parseFloat(valA)
          const numB = parseFloat(valB)
          if (!isNaN(numA) && !isNaN(numB)) {
            valA = numA
            valB = numB
          }
        }
        
        // Handle null/empty values
        if (valA == null || valA === '') {
          valA = sortState.direction === 'asc' ? Infinity : -Infinity
        }
        if (valB == null || valB === '') {
          valB = sortState.direction === 'asc' ? Infinity : -Infinity
        }
        
        // String normalization
        if (typeof valA === 'string' && typeof valB === 'string') {
          valA = valA.toLowerCase()
          valB = valB.toLowerCase()
        }
        
        // Compare and return sort order
        if (valA < valB) {
          return sortState.direction === 'asc' ? -1 : 1
        }
        if (valA > valB) {
          return sortState.direction === 'asc' ? 1 : -1
        }
        return 0
      })
    }
    
    // Helper to parse monetary values (€1.5M, £500K, etc.)
    const parseMonetaryValue = (valueStr) => {
      if (typeof valueStr !== 'string' || !valueStr) return 0
      
      let numStr = valueStr.replace(/[^0-9.,]/g, '')
      numStr = numStr.replace(',', '.')
      
      let multiplier = 1
      if (valueStr.toLowerCase().includes('m')) {
        multiplier = 1000000
      } else if (valueStr.toLowerCase().includes('k')) {
        multiplier = 1000
      }
      
      const cleanedNumStr = numStr.replace(/[^\d.]/g, '')
      const numericValue = parseFloat(cleanedNumStr)
      
      return isNaN(numericValue) ? 0 : numericValue * multiplier
    }
    
    // Clear search filters
    const clearSearch = () => {
      filters.name = ''
      filters.club = ''
      filters.transferValue = ''
    }
    
    // Handle search button click
    const handleSearch = () => {
      // The computed property will handle the filtering
    }
    
    return {
      playerFile,
      loading,
      error,
      allPlayers,
      filteredPlayers,
      filters,
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