<template>
  <div class="player-data-table">
    <!-- If no players to display -->
    <q-card v-if="players.length === 0" class="q-pa-md text-center">
      <p class="text-grey-7">
        {{ loading ? 'Loading player data...' : 'No players match your search criteria.' }}
      </p>
    </q-card>
    
    <!-- Players table with data -->
    <div v-else>
      <q-table
        :rows="players"
        :columns="allColumns"
        :loading="loading"
        row-key="name"
        :pagination="pagination"
        :rows-per-page-options="[10, 15, 20, 50, 0]"
        @request="onRequest"
        binary-state-sort
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
        
        <!-- Loading state overlay -->
        <template v-slot:loading>
          <q-inner-loading showing color="primary">
            <q-spinner size="50px" color="primary" />
          </q-inner-loading>
        </template>
        
        <!-- Pagination slot -->
        <template v-slot:bottom="props">
          <div class="row full-width q-py-sm">
            <div class="col-12 col-md-6 flex items-center justify-start q-px-md">
              <span class="text-subtitle2">
                Showing {{ Math.min(1, props.pagination.rowsPerPage * (props.pagination.page - 1) + 1) }} to 
                {{ Math.min(props.pagination.rowsPerPage * props.pagination.page, props.computedRowsNumber) }} 
                of {{ props.computedRowsNumber }} players
              </span>
            </div>
            <div class="col-12 col-md-6 flex items-center justify-end q-gutter-sm q-px-md">
              <q-select
                v-model="props.pagination.rowsPerPage"
                :options="[10, 15, 20, 50, 0]"
                label="Players per page"
                outlined
                dense
                options-dense
                emit-value
                map-options
                style="min-width: 150px"
                :option-label="opt => opt === 0 ? 'All' : opt.toString()"
              />
              <q-pagination
                v-model="props.pagination.page"
                :max="props.pagesNumber"
                :max-pages="6"
                boundary-links
                direction-links
              />
            </div>
          </div>
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
    
    // Pagination state
    const pagination = reactive({
      sortBy: null,
      descending: false,
      page: 1,
      rowsPerPage: 15,
      rowsNumber: computed(() => props.players.length)
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
      { name: 'transfer_value', label: 'Transfer Value', field: 'transfer_value', sortable: true, align: 'left' },
      { name: 'wage', label: 'Wage', field: 'wage', sortable: true, align: 'left' }
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
          align: 'left',
          isAttribute: true // Flag to identify attribute columns
        }))
    })
    
    // All columns (base + attributes)
    const allColumns = computed(() => [...baseColumns, ...attributeColumns.value])
    
    // Handle pagination request
    const onRequest = (props) => {
      pagination.page = props.pagination.page
      pagination.rowsPerPage = props.pagination.rowsPerPage
    }
    
    // Sort the table
    const sortTable = (field, isAttribute = false) => {
      if (sortField.value === field) {
        // Toggle direction
        sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
      } else {
        // New field
        sortField.value = field
        sortDirection.value = 'asc'
      }
      
      // Emit sort event to parent
      emit('update:sort', {
        key: field,
        direction: sortDirection.value,
        isAttribute
      })
    }
    
    return {
      sortField,
      sortDirection,
      pagination,
      baseColumns,
      attributeColumns,
      allColumns,
      onRequest,
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
</style>