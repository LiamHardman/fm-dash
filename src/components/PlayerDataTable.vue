<template>
  <div class="player-data-table">
    <!-- Debug info -->
    <div v-if="sortField" class="text-caption q-mb-sm q-pa-xs bg-grey-2">
      Current Sort: {{ sortField }} ({{ sortDirection }})
    </div>
    
    <!-- If no players to display -->
    <q-card v-if="players.length === 0" class="q-pa-md text-center">
      <p class="text-grey-7">
        {{ loading ? 'Loading player data...' : 'No players match your search criteria.' }}
      </p>
    </q-card>
    
    <!-- Players table with data -->
    <div v-else>
      <q-table
        :rows="sortedPlayers"
        :columns="allColumns"
        :loading="loading"
        row-key="name"
        :pagination.sync="pagination" 
        :rows-per-page-options="rowsPerPageOptions"
        @request="onRequest"
        :sort-method="customSort"
        binary-state-sort
        flat
        bordered
      >
        <!-- Custom header slots for sortable columns -->
        <template v-slot:header="props">
          <q-tr :props="props">
            <q-th
              v-for="col in props.cols"
              :key="col.name"
              :props="props"
              class="cursor-pointer"
              @click="sortTable(col.name, col.isAttribute)"
            >
              {{ col.label }}
              <q-icon
                v-if="sortField === col.name"
                :name="sortDirection === 'asc' ? 'arrow_upward' : 'arrow_downward'"
                size="xs"
                class="q-ml-xs"
              />
            </q-th>
          </q-tr>
        </template>
        
        <!-- Custom body cell rendering for different data types -->
        <template v-slot:body-cell="props">
          <q-td :props="props">
            <template v-if="isAttributeColumn(props.col.name)">
              <!-- For player attributes (stats), apply color coding based on value -->
              <span
                :class="getAttributeClass(props.value)"
                class="attribute-value"
              >
                {{ props.value || '-' }}
              </span>
            </template>
            <template v-else-if="props.col.name === 'transfer_value'">
              <!-- For transfer values with money formatting -->
              <span 
                :class="getMoneyClass(props.value)"
                class="money-value"
              >
                {{ props.value || '-' }}
              </span>
            </template>
            <template v-else-if="props.col.name === 'wage'">
              <!-- For wage values with money formatting -->
              <span 
                :class="getMoneyClass(props.value)"
                class="money-value"
              >
                {{ props.value || '-' }}
              </span>
            </template>
            <template v-else>
              <!-- Default rendering for other columns -->
              <span>{{ props.value || '-' }}</span>
            </template>
          </q-td>
        </template>
        
        <!-- Loading state overlay -->
        <template v-slot:loading>
          <q-inner-loading showing color="primary">
            <q-spinner size="50px" color="primary" />
          </q-inner-loading>
        </template>
        
        <!-- Pagination controls -->
        <template v-slot:pagination="scope">
          <q-pagination
            v-model="scope.pagination.page"
            :max="pagesNumber"
            :max-pages="maxPagesToShow"
            boundary-links
            direction-links
            @update:model-value="onPageChange"
          />
          <q-space />
          <span class="text-caption q-mr-sm">Rows per page:</span>
          <q-select
            v-model="pagination.rowsPerPage"
            :options="rowsPerPageOptions"
            dense
            options-dense
            emit-value
            map-options
            style="min-width: 100px"
            @update:model-value="onRowsPerPageChange"
            :option-label="opt => opt === 0 ? 'All' : opt.toString()"
            borderless
          />
          <span class="q-ml-md text-caption">
            {{ (pagination.page - 1) * pagination.rowsPerPage + 1 }}
            -
            {{ Math.min(pagination.page * pagination.rowsPerPage, players.length) }}
            of {{ players.length }}
          </span>
        </template>
      </q-table>
    </div>
  </div>
</template>

<script>
import { ref, computed, reactive, watch } from 'vue'

export default {
  name: 'PlayerDataTable',
  props: {
    players: {
      type: Array,
      required: true
    },
    loading: {
      type: Boolean,
      default: false
    }
  },
  emits: ['update:sort'],
  
  setup(props, { emit }) {
    // Sorting state
    const sortField = ref(null)
    const sortDirection = ref('asc')
    
    // Pagination settings
    const rowsPerPageOptions = [10, 15, 20, 50, 0]
    const maxPagesToShow = 7
    
    // Pagination state
    const pagination = reactive({
      sortBy: null,
      descending: false,
      page: 1,
      rowsPerPage: 15
    })
    
    // Calculate total number of pages
    const pagesNumber = computed(() => {
      if (pagination.rowsPerPage === 0) return 1
      return Math.ceil(props.players.length / pagination.rowsPerPage)
    })
    
    // Compute the displayed players based on current pagination
    const displayedPlayers = computed(() => {
      if (pagination.rowsPerPage === 0) return sortedPlayers.value
      
      const firstIndex = (pagination.page - 1) * pagination.rowsPerPage
      const lastIndex = Math.min(firstIndex + pagination.rowsPerPage, props.players.length)
      
      // Log the paginated players for debugging
      const paginatedPlayers = sortedPlayers.value.slice(firstIndex, lastIndex)
      
      // If sorting by money fields, log the values on the current page
      if (sortField.value === 'transfer_value' || sortField.value === 'wage') {
        console.log(`Current page (${pagination.page}) player values:`)
        paginatedPlayers.forEach(player => {
          const field = sortField.value === 'transfer_value' ? 'transferValueAmount' : 'wageAmount'
          console.log(`${player.name}: ${player[sortField.value]} → ${player[field]}`)
        })
      }
      
      return paginatedPlayers
    })
    
    // Reset pagination when players change
    watch(() => props.players.length, () => {
      pagination.page = 1
    })
    
    // Base columns (always present)
    const baseColumns = [
      { name: 'name', label: 'Name', field: 'name', sortable: true, align: 'left' },
      { name: 'position', label: 'Position', field: 'position', sortable: true, align: 'left' },
      { name: 'age', label: 'Age', field: 'age', sortable: true, align: 'left' },
      { name: 'club', label: 'Club', field: 'club', sortable: true, align: 'left' },
      { 
        name: 'transfer_value', 
        label: 'Transfer Value', 
        field: 'transfer_value', 
        sortable: true, 
        align: 'right',
        sortField: 'transferValueAmount' // Field to use for sorting
      },
      { 
        name: 'wage', 
        label: 'Wage', 
        field: 'wage', 
        sortable: true, 
        align: 'right',
        sortField: 'wageAmount' // Field to use for sorting
      }
    ]
    
    // Attribute columns (dynamic based on player data)
    const attributeColumns = computed(() => {
      if (!props.players.length) return []
      
      // Get all unique attribute keys from the first player
      const player = props.players[0]
      if (!player.attributes || typeof player.attributes !== 'object') {
        return []
      }
      
      return Object.keys(player.attributes)
        .sort() // Sort attribute names alphabetically
        .map(attr => ({
          name: attr,
          label: attr,
          field: row => row.attributes?.[attr] || '-',
          sortable: true,
          align: 'center', // Center alignment for attribute values
          isAttribute: true // Flag to identify attribute columns
        }))
    })
    
    // All columns (base + attributes)
    const allColumns = computed(() => [...baseColumns, ...attributeColumns.value])
    
    // Check if a column is an attribute column
    const isAttributeColumn = (colName) => {
      return attributeColumns.value.some(col => col.name === colName)
    }
    
    // Get the actual field to sort by from a column name
    const getSortField = (colName) => {
      // For monetary fields, use their numeric equivalents
      if (colName === 'transfer_value') return 'transferValueAmount'
      if (colName === 'wage') return 'wageAmount'
      return colName
    }
    
    // Sorted players computed property - handles sorting directly
    const sortedPlayers = computed(() => {
      if (!sortField.value) return props.players
      
      const field = getSortField(sortField.value)
      const direction = sortDirection.value
      const isAttr = isAttributeColumn(sortField.value)
      
      console.log(`Sorting players by ${field} (${direction}, isAttribute: ${isAttr})`)
      
      // Debug example values for monetary sorts
      if (field === 'transferValueAmount' || field === 'wageAmount') {
        // Get all players with the field and log the top 10
        const valuableList = [...props.players]
          .filter(p => p[field] > 0)
          .sort((a, b) => b[field] - a[field])
          .slice(0, 10)
        
        console.log(`Top 10 ${field} values in dataset:`)
        const displayField = field === 'transferValueAmount' ? 'transfer_value' : 'wage'
        valuableList.forEach(player => {
          console.log(`${player.name}: ${player[displayField]} → ${player[field]} (${typeof player[field]})`)
        })
      }
      
      const sortedList = [...props.players].sort((a, b) => {
        let valA, valB
        
        // Get values based on sort field
        if (isAttr) {
          // Attribute field (from the attributes object)
          valA = a.attributes?.[field]
          valB = b.attributes?.[field]
        } else if (field === 'transferValueAmount' || field === 'wageAmount') {
          // Use the numeric amount fields for money values
          valA = a[field] || 0 // Ensure we have a number, default to 0
          valB = b[field] || 0
          
          // Debug comparison for monetary values
          if (Math.random() < 0.05) { // Only log ~5% of comparisons to avoid console spam
            console.log(`Comparing: ${a.name} (${valA}) vs ${b.name} (${valB})`)
          }
        } else {
          // Regular field
          valA = a[field]
          valB = b[field]
        }
        
        // Handle null/empty values
        if (valA == null && valB == null) return 0
        if (valA == null) return direction === 'asc' ? 1 : -1
        if (valB == null) return direction === 'asc' ? -1 : 1
        
        // Direct number comparison for numeric fields
        if (typeof valA === 'number' && typeof valB === 'number') {
          return direction === 'asc' ? valA - valB : valB - valA
        }
        
        // String comparison for text fields
        if (typeof valA === 'string' && typeof valB === 'string') {
          // For numeric strings, try to compare as numbers first
          const numA = parseFloat(valA)
          const numB = parseFloat(valB)
          if (!isNaN(numA) && !isNaN(numB)) {
            return direction === 'asc' ? numA - numB : numB - numA
          }
          
          // Otherwise do string comparison
          valA = valA.toLowerCase()
          valB = valB.toLowerCase()
          if (valA < valB) return direction === 'asc' ? -1 : 1
          if (valA > valB) return direction === 'asc' ? 1 : -1
          return 0
        }
        
        // Fallback comparison
        return 0
      })
      
      // Log top results after sorting
      if (field === 'transferValueAmount' || field === 'wageAmount') {
        console.log(`Top 10 ${field} values AFTER sorting (${direction}):`)
        const displayField = field === 'transferValueAmount' ? 'transfer_value' : 'wage'
        const toShow = direction === 'asc' ? sortedList.slice(0, 10) : sortedList.slice(0, 10)
        toShow.forEach(player => {
          console.log(`${player.name}: ${player[displayField]} → ${player[field]}`)
        })
      }
      
      return sortedList
    })
    
    // Get CSS class for attribute values based on their value
    const getAttributeClass = (value) => {
      if (value === null || value === undefined || value === '-') {
        return 'attribute-na'
      }
      
      // Convert to number if it's a string
      const numValue = typeof value === 'number' ? value : parseInt(value, 10)
      
      // Return class based on attribute value
      if (isNaN(numValue)) return 'attribute-na'
      
      if (numValue >= 18) return 'attribute-excellent'
      if (numValue >= 15) return 'attribute-very-good'
      if (numValue >= 12) return 'attribute-good'
      if (numValue >= 9) return 'attribute-average'
      if (numValue >= 6) return 'attribute-poor'
      return 'attribute-very-poor'
    }
    
    // Get CSS class for money values based on amount
    const getMoneyClass = (value) => {
      if (value === null || value === undefined || value === '-') {
        return 'money-na'
      }
      
      // Parse the monetary value
      const amount = parseMonetaryValue(value)
      
      // Log the parsed amount for debugging
      console.log(`Money class: ${value} → ${amount}`)
      
      // Return class based on monetary value
      if (amount >= 10000000) return 'money-very-high'  // 10M+
      if (amount >= 1000000) return 'money-high'        // 1M+
      if (amount >= 100000) return 'money-medium-high'  // 100K+
      if (amount >= 10000) return 'money-medium'        // 10K+
      if (amount > 0) return 'money-low'                // > 0
      return 'money-na'                                 // 0 or invalid
    }
    
    // Helper to parse monetary values (€1.5M, £500K, etc.)
    const parseMonetaryValue = (valueStr) => {
      if (typeof valueStr !== 'string' || !valueStr) return 0
      
      // Remove any text after p/w (per week)
      const cleanedStr = valueStr.split(' p/w')[0]
      
      // Enhanced debug logging for ALL monetary values
      console.log(`Parsing monetary value: "${valueStr}" → "${cleanedStr}"`)
      
      // Determine multiplier based on suffix
      let multiplier = 1
      if (cleanedStr.toLowerCase().includes('m')) {
        multiplier = 1000000
      } else if (cleanedStr.toLowerCase().includes('k')) {
        multiplier = 1000
      }
      
      // Special handling for values with commas (e.g., £350,000)
      // Remove all currency symbols and non-numeric chars except commas and periods
      let numStr = cleanedStr.replace(/[^0-9,.]/g, '')
      
      // Handle comma-separated thousands (e.g., convert £350,000 to 350000)
      // First check if it looks like a thousands separator comma pattern
      if (numStr.includes(',') && !numStr.includes('.')) {
        // Replace all commas with nothing (i.e., remove them)
        numStr = numStr.replace(/,/g, '')
      } else {
        // If it's likely a decimal comma, replace with period
        numStr = numStr.replace(',', '.')
      }
      
      // Parse the numeric value
      const numericValue = parseFloat(numStr)
      
      // Calculate result
      const result = Math.round(isNaN(numericValue) ? 0 : numericValue * multiplier)
      
      // Debug output for all cases
      console.log(`Parse result: "${valueStr}" → ${result} (multiplier: ${multiplier}, numericValue: ${numericValue})`)
      
      return result
    }
    
    // Handle pagination request
    const onRequest = (props) => {
      console.log(`Pagination request: page ${props.pagination.page}, rows ${props.pagination.rowsPerPage}`)
      pagination.page = props.pagination.page
      pagination.rowsPerPage = props.pagination.rowsPerPage
      
      // Log page change for debugging
      if (sortField.value === 'transfer_value' || sortField.value === 'wage') {
        console.log(`Changing to page ${pagination.page} with ${pagination.rowsPerPage} rows per page`)
      }
    }
    
    // Handle page change
    const onPageChange = (page) => {
      console.log(`Page changed to: ${page}`)
      pagination.page = page
    }
    
    // Handle rows per page change
    const onRowsPerPageChange = (rowsPerPage) => {
      console.log(`Rows per page changed to: ${rowsPerPage}`)
      pagination.rowsPerPage = rowsPerPage
      pagination.page = 1 // Reset to first page when changing rows per page
    }
    
    // Sort the table - handles sorting internally and notifies parent
    // Custom sort method for q-table
    const customSort = (rows, sortBy, descending) => {
      console.log(`Q-Table sort request: ${sortBy}, descending: ${descending}`)
      
      // Use our custom sort logic
      const field = getSortField(sortBy)
      const isAttr = isAttributeColumn(sortBy)
      const direction = descending ? 'desc' : 'asc'
      
      console.log(`Custom sorting by ${field} (${direction})`)
      
      return rows.sort((a, b) => {
        let valA, valB
        
        // Get values based on sort field
        if (isAttr) {
          valA = a.attributes?.[field]
          valB = b.attributes?.[field]
        } else if (field === 'transferValueAmount' || field === 'wageAmount') {
          valA = typeof a[field] === 'number' ? a[field] : 0
          valB = typeof b[field] === 'number' ? b[field] : 0
          
          // Log a sample comparison
          if (Math.random() < 0.01) {
            const displayKey = field === 'transferValueAmount' ? 'transfer_value' : 'wage'
            console.log(`Q-Table comparing: ${a.name} (${a[displayKey]} → ${valA}) vs ${b.name} (${b[displayKey]} → ${valB})`)
          }
        } else {
          valA = a[field]
          valB = b[field]
        }
        
        // Handle null/undefined values
        if (valA == null && valB == null) return 0
        if (valA == null) return direction === 'asc' ? 1 : -1
        if (valB == null) return direction === 'asc' ? -1 : 1
        
        // Direct number comparison for numeric fields
        if (typeof valA === 'number' && typeof valB === 'number') {
          return direction === 'asc' ? valA - valB : valB - valA
        }
        
        // Try to convert to number if they look like numbers
        if (!isNaN(parseFloat(valA)) && !isNaN(parseFloat(valB))) {
          const numA = parseFloat(valA)
          const numB = parseFloat(valB)
          return direction === 'asc' ? numA - numB : numB - numA
        }
        
        // String comparison
        if (typeof valA === 'string' && typeof valB === 'string') {
          valA = valA.toLowerCase()
          valB = valB.toLowerCase()
          if (valA < valB) return direction === 'asc' ? -1 : 1
          if (valA > valB) return direction === 'asc' ? 1 : -1
          return 0
        }
        
        return 0
      })
    }
    
    const sortTable = (field, isAttribute = false) => {
      // For monetary fields, use their actual sort field
      const actualField = getSortField(field)
      
      // Update sorting direction
      if (sortField.value === field) {
        // Toggle direction if same field
        sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
      } else {
        // New field
        sortField.value = field
        sortDirection.value = 'asc'
      }
      
      // Update pagination sort props to trigger the q-table's internal sort
      pagination.sortBy = field
      pagination.descending = sortDirection.value === 'desc'
      
      // Log the sort request for debugging
      console.log(`Sort request: ${field} → ${actualField} (${sortDirection.value}, isAttribute: ${isAttribute})`)
      
      // Emit an event for the parent component to use
      emit('update:sort', {
        key: actualField,
        direction: sortDirection.value,
        isAttribute: isAttribute,
        displayField: field // Original field name for UI
      })
    }
    
    return {
      sortField,
      sortDirection,
      pagination,
      pagesNumber,
      rowsPerPageOptions,
      maxPagesToShow,
      baseColumns,
      attributeColumns,
      allColumns,
      sortedPlayers,
      displayedPlayers,
      isAttributeColumn,
      getAttributeClass,
      getMoneyClass,
      parseMonetaryValue,
      onRequest,
      onPageChange,
      onRowsPerPageChange,
      customSort,
      sortTable
    }
  }
}
</script>

<style scoped>
.player-data-table {
  width: 100%;
  overflow-x: auto;
}

/* Make the table headers more prominent */
:deep(.q-table th) {
  font-weight: 600;
  background-color: #f3f5f9;
}

/* Alternate row colors for better readability */
:deep(.q-table tr:nth-child(even)) {
  background-color: #f9fafb;
}

/* Hover effect on rows */
:deep(.q-table tr:hover) {
  background-color: #e5f1fb;
}

/* Style the pagination controls */
:deep(.q-pagination .q-btn.q-btn--active) {
  background-color: var(--q-primary);
  color: white;
}

/* Attribute value color coding (Football Manager style) */
.attribute-value {
  display: inline-block;
  min-width: 22px;
  text-align: center;
  font-weight: 600;
  padding: 1px 4px;
  border-radius: 3px;
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

/* Money value styling */
.money-value {
  display: inline-block;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: 3px;
}

.money-very-high {
  color: #2b8a3e;
  font-weight: 700;
}

.money-high {
  color: #2b8a3e;
}

.money-medium-high {
  color: #5c940d;
}

.money-medium {
  color: #212529;
}

.money-low {
  color: #495057;
}

.money-na {
  color: #868e96;
}
</style>