<template>
    <div class="player-data-table-container">
        <div class="row items-center q-mb-sm table-controls-header">
            <div class="col">
                <div
                    v-if="sortField"
                    class="text-caption q-pa-xs rounded-borders sort-info-chip"
                >
                    Current Sort: {{ getColumnLabel(sortField) }} ({{
                        sortDirection === "asc" ? "Ascending" : "Descending"
                    }})
                </div>
            </div>
        </div>

        <q-card
            v-if="players.length === 0 && !loading"
            class="q-pa-md text-center no-players-card"
            flat
            bordered
        >
            <p class="no-players-text">
                No players match your search criteria.
            </p>
        </q-card>
        <q-card
            v-else-if="
                sortedPlayers.length === 0 && players.length > 0 && !loading
            "
            class="q-pa-md text-center no-players-card"
            flat
            bordered
        >
            <p class="no-players-text">
                No players to display with current sort (possibly all filtered
                out before slicing).
            </p>
        </q-card>

        <div v-else class="relative-position">
            <q-table
                :key="tableKey"
                :rows="sortedPlayers"
                :columns="currentColumns"
                :loading="loading"
                row-key="name"
                :pagination="pagination"
                @update:pagination="onPaginationUpdate"
                :rows-per-page-options="rowsPerPageOptions"
                @request="onRequest"
                :sort-method="customSort"
                binary-state-sort
                flat
                bordered
                class="player-q-table"
                :class="[
                    qInstance.dark.isActive ? 'q-table--dark' : '',
                    { 'table-sorting': isAsyncSorting }
                ]"
                table-header-class="player-table-header"
                dense
                virtual-scroll
                :virtual-scroll-item-size="32"
                :virtual-scroll-sticky-size-start="40"
                :virtual-scroll-sticky-size-end="55"
                style="height: 100%; min-height: 500px;"
                separator="horizontal"
                no-data-label="No players to display"
            >
            <template v-slot:header="props">
                <PlayerTableHeader 
                    :props="props"
                    :sort-field="sortField"
                    :sort-direction="sortDirection"
                    :is-async-sorting="isAsyncSorting"
                    :currency-symbol="currencySymbol"
                    @sort-table="sortTable"
                />
            </template>

            <template v-slot:body="props">
                <PlayerTableRow
                    :props="props"
                    :get-display-value="getDisplayValue"
                    :get-unified-rating-class="getUnifiedRatingClass"
                    :get-total-stats-rating-class="getTotalStatsRatingClass"
                    :get-value-score-class="getValueScoreClass"
                    :get-money-class="getMoneyClass"
                    :format-display-currency="formatDisplayCurrency"
                    :on-flag-error="onFlagError"
                    @row-click="onRowClick"
                    @right-click="onRightClick"
                    @club-click="onClubClick"
                />
            </template>

            <template v-slot:loading>
                <q-inner-loading showing color="primary">
                    <q-spinner size="50px" color="primary" />
                </q-inner-loading>
            </template>

            <template v-slot:pagination="scope">
                <PlayerTablePagination
                    :scope="scope"
                    :pages-number="pagesNumber"
                    :max-pages-to-show="maxPagesToShow"
                    :pagination-start-row="paginationStartRow"
                    :pagination-end-row="paginationEndRow"
                    :pagination-total-rows="paginationTotalRows"
                    :total-sorted-count="totalSortedCount"
                    :is-sliced="isSliced"
                    @page-change="onPageChange"
                />
            </template>
        </q-table>

        <PlayerTableContextMenu
            ref="contextMenu"
            :context-menu-player="contextMenuPlayer"
            :is-player-in-wishlist="contextMenuPlayer ? isPlayerInWishlist(contextMenuPlayer) : false"
            @add-to-wishlist="handleAddToWishlist"
            @remove-from-wishlist="handleRemoveFromWishlist"
            @player-details="handlePlayerDetails"
        />
    </div>
    </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { memoize } from '../composables/useMemoization'
import { usePlayerTableCells } from '../composables/usePlayerTableCells'
import { usePlayerTableColumns } from '../composables/usePlayerTableColumns'
import { useOptimizedSorting } from '../composables/useVirtualScrolling'
import { usePlayerCalculationWorker } from '../composables/useWebWorkers'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
import { useWishlistStore } from '../stores/wishlistStore'
import PlayerTableContextMenu from './player-table/PlayerTableContextMenu.vue'
import PlayerTableHeader from './player-table/PlayerTableHeader.vue'
import PlayerTablePagination from './player-table/PlayerTablePagination.vue'
import PlayerTableRow from './player-table/PlayerTableRow.vue'

const MAX_DISPLAY_PLAYERS = 1000

export default {
  name: 'PlayerDataTable',
  components: {
    PlayerTableHeader,
    PlayerTableRow,
    PlayerTablePagination,
    PlayerTableContextMenu,
  },
  props: {
    players: { type: Array, required: true },
    loading: { type: Boolean, default: false },
    isGoalkeeperView: { type: Boolean, default: false },
    currencySymbol: { type: String, default: '$' },
    filteredPlayerCount: { type: Number, default: 0 },
    showWishlistActions: { type: Boolean, default: false },
    datasetId: { type: String, default: null },
    showValueScore: { type: Boolean, default: false },
    defaultSortField: { type: String, default: 'Overall' },
    defaultSortDirection: { type: String, default: 'desc' },
  },
  emits: [
    'update:sort',
    'player-selected',
    'update:pagination',
    'team-selected',
    'remove-from-wishlist',
  ],

  setup(props, { emit }) {
    const $q = useQuasar()
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()
    const _router = useRouter()
    const _uiStore = useUiStore()

    // Initialize optimized sorting and web workers
    const { sortLargeArray, clearSortCache } = useOptimizedSorting()
    const { sortPlayers: sortPlayersWorker } = usePlayerCalculationWorker()

    // Cache generation key that changes when players data changes
    const cacheGeneration = ref(0)

    // Initialize composables
    const { currentColumns, getColumnLabel, getSortFieldKey } = usePlayerTableColumns(
      props.isGoalkeeperView,
      props.showValueScore
    )

    const {
      getUnifiedRatingClass,
      getTotalStatsRatingClass,
      getMoneyClass,
      getValueScoreClass,
      onFlagError,
      formatDisplayCurrency,
      getDisplayValue,
      getPlayerValue,
      clearCaches,
    } = usePlayerTableCells(props.isGoalkeeperView, props.currencySymbol, cacheGeneration)

    const contextMenu = ref(null)
    const sortField = ref(props.defaultSortField)
    const sortDirection = ref(props.defaultSortDirection)
    const rowsPerPageOptions = ref([10, 15, 20, 50, 0]) // Keep for internal logic, but selector is removed
    const maxPagesToShow = 7
    const totalSortedCount = ref(0)
    const isAsyncSorting = ref(false)
    const isSliced = ref(false)
    const currentSortController = ref(null) // For cancelling sort operations

    const pagination = ref({
      sortBy: props.defaultSortField,
      descending: props.defaultSortDirection === 'desc',
      page: 1,
      rowsPerPage: 50, // Default rows per page, even if selector is hidden
    })

    // Get current dataset ID - use prop first, then fallback to store
    const currentDatasetId = computed(() => props.datasetId || playerStore.currentDatasetId)

    watch(
      () => props.players,
      (_newPlayers, _oldPlayers) => {
        pagination.value.page = 1 // Reset to first page when player list changes
      },
      { deep: true }
    )

    const positionSortOrder = [
      'GK',
      'DR',
      'DC',
      'DL',
      'WBR',
      'WBL',
      'DM',
      'MR',
      'MC',
      'ML',
      'AMR',
      'AMC',
      'AML',
      'ST',
    ]

    // Memoized position index calculation (expensive string processing)
    const getPositionIndex = memoize(
      (positionString) => {
        if (!positionString || typeof positionString !== 'string') {
          return positionSortOrder.length + 2 // Place invalid/empty last
        }
        let processedString = positionString.toUpperCase()
        processedString = processedString.replace(/\bST\s*\(C\)/g, 'ST')
        processedString = processedString.replace(/\bM\s*\(C\)/g, 'MC')
        processedString = processedString.replace(/\bAM\s*\(C\)/g, 'AMC')
        processedString = processedString.replace(/\bDM\s*\(C\)/g, 'DM')
        processedString = processedString.replace(/\bD\s*\(C\)/g, 'DC')
        processedString = processedString.replace(/\bGK\s*\(C\)/g, 'GK')

        let minFoundIndex = positionSortOrder.length
        const sideMatch = processedString.match(/\(([^)]+)\)$/)
        let mainPart = processedString
        const sidesSpecified = []

        if (sideMatch?.[1]) {
          mainPart = processedString.substring(0, sideMatch.index).trim()
          const sideSpec = sideMatch[1]
          if (sideSpec.includes('R')) sidesSpecified.push('R')
          if (sideSpec.includes('L')) sidesSpecified.push('L')
        }

        mainPart = mainPart.replace(/\s*\(.*?\)\s*/g, '').trim()
        const basePositionCodes = mainPart
          .split(/[,/]/)
          .map((p) => p.trim())
          .filter((p) => p.length > 0)
        const rolesToEvaluate = new Set()

        for (const baseCode of basePositionCodes) {
          if (sidesSpecified.length > 0) {
            for (const side of sidesSpecified) {
              rolesToEvaluate.add(baseCode + side)
            }
          }
          rolesToEvaluate.add(baseCode)
        }

        if (rolesToEvaluate.size === 0 && positionString.trim() !== '') {
          rolesToEvaluate.add(processedString.replace(/\s*\(.*?\)\s*/g, '').trim())
        }
        if (rolesToEvaluate.size === 0) return positionSortOrder.length + 1

        for (const role of rolesToEvaluate) {
          const index = positionSortOrder.indexOf(role)
          if (index !== -1 && index < minFoundIndex) {
            minFoundIndex = index
          }
        }
        return minFoundIndex === positionSortOrder.length
          ? positionSortOrder.length + 1
          : minFoundIndex
      },
      {
        maxSize: 100, // Cache position calculations
        keyGenerator: (positionString) => positionString,
        cacheKey: 'positionIndex',
      }
    )

    const onPaginationUpdate = (newPagination) => {
      pagination.value = newPagination
    }

    // Cache for sorted results to avoid re-sorting on pagination changes
    const sortedPlayersCache = ref(null)
    const lastSortKey = ref('')

    // Separate reactive flag to prevent infinite loops
    const sortKeyChanged = ref('')
    const asyncSortTimeout = ref(null)

    const sortedPlayers = computed(() => {
      if (!props.players || props.players.length === 0) {
        totalSortedCount.value = 0
        isSliced.value = false
        sortedPlayersCache.value = null
        return []
      }

      const fieldKey = getSortFieldKey(sortField.value || 'Overall')
      const direction = sortDirection.value
      const currentSortKey = `${fieldKey}-${direction}-${props.players.length}-${cacheGeneration.value}`

      // Return cached result if sort parameters haven't changed
      if (sortedPlayersCache.value && lastSortKey.value === currentSortKey) {
        return sortedPlayersCache.value
      }

      // Custom sort function for complex sorting logic
      const customSortFn = (a, b, field, dir) => {
        // Use getPlayerValue to ensure GK stat mapping is applied for sorting
        let vA = getPlayerValue(a, field, sortField.value)
        let vB = getPlayerValue(b, field, sortField.value)
        const aIsNull = vA === null || vA === undefined
        const bIsNull = vB === null || vB === undefined

        if (aIsNull && bIsNull) return 0
        if (aIsNull) return dir === 'asc' ? 1 : -1
        if (bIsNull) return dir === 'asc' ? -1 : 1

        if (field === 'position') {
          const indexA = getPositionIndex(vA)
          const indexB = getPositionIndex(vB)
          return dir === 'asc' ? indexA - indexB : indexB - indexA
        }
        if (typeof vA === 'number' && typeof vB === 'number') {
          return dir === 'asc' ? vA - vB : vB - vA
        }
        if (typeof vA === 'string' && typeof vB === 'string') {
          vA = vA.toLowerCase()
          vB = vB.toLowerCase()
          if (vA < vB) return dir === 'asc' ? -1 : 1
          if (vA > vB) return dir === 'asc' ? 1 : -1
          return 0
        }
        return 0
      }

      // For small arrays, use synchronous sorting
      if (props.players.length <= 500) {
        const playersToSort = [...props.players]
        const fullSortedList = playersToSort.sort((a, b) => customSortFn(a, b, fieldKey, direction))

        totalSortedCount.value = fullSortedList.length
        let result
        let sliced = false

        if (fullSortedList.length > MAX_DISPLAY_PLAYERS) {
          sliced = true
          result = fullSortedList.slice(0, MAX_DISPLAY_PLAYERS)
        } else {
          result = fullSortedList
        }

        isSliced.value = sliced
        sortedPlayersCache.value = result
        lastSortKey.value = currentSortKey

        // Make sure async sorting flag is false for small datasets
        isAsyncSorting.value = false

        return result
      }

      // For large arrays, return unsorted data immediately and trigger async sort
      // This prevents blocking the UI thread on initial load
      if (sortKeyChanged.value !== currentSortKey) {
        sortKeyChanged.value = currentSortKey

        // Clear any existing timeout
        if (asyncSortTimeout.value) {
          clearTimeout(asyncSortTimeout.value)
        }

        // If this is the initial load, show unsorted data immediately
        if (!sortedPlayersCache.value) {
          const initialDisplay = [...props.players].slice(0, MAX_DISPLAY_PLAYERS)
          sortedPlayersCache.value = initialDisplay
          totalSortedCount.value = props.players.length
          isSliced.value = props.players.length > MAX_DISPLAY_PLAYERS
          lastSortKey.value = currentSortKey

          // Defer sorting to next tick to avoid blocking UI
          nextTick(() => {
            triggerAsyncSort(fieldKey, direction, customSortFn, currentSortKey)
          })

          return initialDisplay
        }

        // Debounce async sorting to prevent rapid successive calls
        asyncSortTimeout.value = setTimeout(() => {
          if (!isAsyncSorting.value && sortKeyChanged.value === currentSortKey) {
            triggerAsyncSort(fieldKey, direction, customSortFn, currentSortKey)
          }
        }, 100) // 100ms debounce
      }

      // CRITICAL: Always return the current cache during sorting to prevent layout shifts
      // This ensures the table maintains exactly the same data structure and length
      if (sortedPlayersCache.value && sortedPlayersCache.value.length > 0) {
        return sortedPlayersCache.value
      }
      // If no cache yet, create a stable fallback that matches the expected display size
      // This prevents any change in table height or structure during initial sort
      const targetLength = Math.min(MAX_DISPLAY_PLAYERS, props.players.length)
      const stableFallback = [...props.players].slice(0, targetLength)

      // Set initial cache to prevent further changes during sorting
      if (!sortedPlayersCache.value) {
        sortedPlayersCache.value = stableFallback
        totalSortedCount.value = props.players.length
        isSliced.value = props.players.length > MAX_DISPLAY_PLAYERS
      }

      return stableFallback
    })

    // Async sorting for large datasets using web workers
    const triggerAsyncSort = async (fieldKey, direction, customSortFn, sortKey) => {
      // Prevent concurrent async sorting operations
      if (isAsyncSorting.value) {
        return
      }

      // Cancel any previous sort operation
      if (currentSortController.value) {
        currentSortController.value.cancelled = true
      }

      // Create new controller for this sort operation
      const sortController = { cancelled: false }
      currentSortController.value = sortController

      // Set async sorting flag AFTER ensuring cache stability
      isAsyncSorting.value = true

      try {
        let fullSortedList
        const playerCount = props.players.length

        // Tiered sorting strategy based on dataset size
        if (playerCount >= 2000) {
          try {
            fullSortedList = await sortPlayersWorker(
              [...props.players],
              fieldKey,
              direction,
              sortField.value,
              props.isGoalkeeperView
            )
          } catch (_workerError) {
            // Fallback to main thread if Web Worker fails
            fullSortedList = await sortLargeArray(
              [...props.players],
              fieldKey,
              direction,
              customSortFn,
              2000
            )
          }
        } else {
          fullSortedList = await sortLargeArray(
            [...props.players],
            fieldKey,
            direction,
            customSortFn,
            2000
          )
        }

        // Check if this sort was cancelled
        if (sortController.cancelled) {
          return
        }

        totalSortedCount.value = fullSortedList.length
        let result
        let sliced = false

        if (fullSortedList.length > MAX_DISPLAY_PLAYERS) {
          sliced = true
          result = fullSortedList.slice(0, MAX_DISPLAY_PLAYERS)
        } else {
          result = fullSortedList
        }

        isSliced.value = sliced

        // Update cache in a way that doesn't trigger layout recalculation
        nextTick(() => {
          sortedPlayersCache.value = result
          lastSortKey.value = sortKey
        })
      } catch (_error) {
        if (sortController.cancelled) {
          return
        }

        // Show user-friendly error notification
        $q.notify({
          type: 'warning',
          message: 'Sorting was interrupted. Showing unsorted data.',
          position: 'top',
          timeout: 3000,
        })

        // Fallback to direct assignment if async sorting fails
        const fallbackResult = [...props.players].slice(0, MAX_DISPLAY_PLAYERS)
        nextTick(() => {
          sortedPlayersCache.value = fallbackResult
          totalSortedCount.value = props.players.length
          isSliced.value = props.players.length > MAX_DISPLAY_PLAYERS
          lastSortKey.value = sortKey
        })
      } finally {
        // Always clear the sorting flag and controller, regardless of outcome
        isAsyncSorting.value = false
        currentSortController.value = null
      }
    }

    // Cancel current async sort operation
    const _cancelAsyncSort = () => {
      if (currentSortController.value) {
        currentSortController.value.cancelled = true
      }
      isAsyncSorting.value = false
      currentSortController.value = null

      $q.notify({
        type: 'info',
        message: 'Sorting cancelled',
        position: 'top',
        timeout: 2000,
      })
    }

    const pagesNumber = computed(() => {
      if (
        !sortedPlayers.value ||
        sortedPlayers.value.length === 0 ||
        pagination.value.rowsPerPage === 0
      ) {
        return 1
      }
      return Math.ceil(sortedPlayers.value.length / pagination.value.rowsPerPage)
    })

    const paginationTotalRows = computed(() => sortedPlayers.value.length)
    const paginationStartRow = computed(() => {
      if (paginationTotalRows.value === 0) return 0
      return (pagination.value.page - 1) * pagination.value.rowsPerPage + 1
    })
    const paginationEndRow = computed(() => {
      if (paginationTotalRows.value === 0) return 0
      if (pagination.value.rowsPerPage === 0) return paginationTotalRows.value
      return Math.min(
        pagination.value.page * pagination.value.rowsPerPage,
        paginationTotalRows.value
      )
    })

    onMounted(() => {
      if (sortField.value) {
        emit('update:sort', {
          key: getSortFieldKey(sortField.value),
          direction: sortDirection.value,
          isFifaStat: currentColumns.value.find((c) => c.name === sortField.value)?.isFifaStat,
          isOverallStat: currentColumns.value.find((c) => c.name === sortField.value)
            ?.isOverallStat,
          isValueScore: currentColumns.value.find((c) => c.name === sortField.value)?.isValueScore,
          displayField: sortField.value,
        })
      }
    })

    const onRequest = (requestProp) => {
      const { page, sortBy, descending } = requestProp.pagination
      pagination.value.page = page

      if (
        sortBy &&
        (sortField.value !== sortBy || sortDirection.value !== (descending ? 'desc' : 'asc'))
      ) {
        sortField.value = sortBy
        sortDirection.value = descending ? 'desc' : 'asc'
        pagination.value.sortBy = sortBy
        pagination.value.descending = descending

        emit('update:sort', {
          key: getSortFieldKey(sortField.value),
          direction: sortDirection.value,
          isFifaStat: currentColumns.value.find((c) => c.name === sortBy)?.isFifaStat,
          isOverallStat: currentColumns.value.find((c) => c.name === sortBy)?.isOverallStat,
          isValueScore: currentColumns.value.find((c) => c.name === sortBy)?.isValueScore,
          displayField: sortBy,
        })
      }
      emit('update:pagination', { ...pagination.value })
    }

    const onPageChange = (newPage) => {
      pagination.value.page = newPage
    }

    const onRowsPerPageChange = (newRowsPerPage) => {
      pagination.value.rowsPerPage = newRowsPerPage
      pagination.value.page = 1
    }

    const customSort = (rows) => {
      // The actual sorting is now done in the `sortedPlayers` computed property.
      // QTable's `sort-method` is still needed, but we just return the rows as they are
      // because our computed property `sortedPlayers` (bound to :rows) already handles it.
      return rows
    }

    const sortTable = (fieldName) => {
      // Prevent rapid clicking during sort operations
      if (isAsyncSorting.value) {
        return
      }

      // Clear any pending async sorting
      if (asyncSortTimeout.value) {
        clearTimeout(asyncSortTimeout.value)
        asyncSortTimeout.value = null
      }

      // Clear sort cache when sort parameters change
      clearSortCache()
      // Don't clear sortedPlayersCache immediately to prevent layout shift
      // It will be updated when the new sort completes
      lastSortKey.value = ''
      sortKeyChanged.value = ''

      const actualSortKey = getSortFieldKey(fieldName)
      let newDirection
      if (sortField.value === fieldName) {
        newDirection = sortDirection.value === 'asc' ? 'desc' : 'asc'
      } else {
        const colDef = currentColumns.value.find((c) => c.name === fieldName)
        if (colDef && (colDef.isOverallStat || colDef.isFifaStat || colDef.isValueScore)) {
          newDirection = 'desc'
        } else {
          newDirection = 'asc'
        }
      }
      sortField.value = fieldName
      sortDirection.value = newDirection

      pagination.value.sortBy = fieldName
      pagination.value.descending = newDirection === 'desc'
      pagination.value.page = 1

      emit('update:sort', {
        key: actualSortKey,
        direction: newDirection,
        isFifaStat: currentColumns.value.find((c) => c.name === fieldName)?.isFifaStat,
        isOverallStat: currentColumns.value.find((c) => c.name === fieldName)?.isOverallStat,
        isValueScore: currentColumns.value.find((c) => c.name === fieldName)?.isValueScore,
        displayField: fieldName,
      })
    }

    const onRowClick = (player) => {
      emit('player-selected', player)
    }

    const onClubClick = (player) => {
      if (player.club && player.club.trim() !== '') {
        emit('team-selected', player.club)
      } else {
      }
    }

    const contextMenuPlayer = ref(null)

    const isPlayerInWishlist = (player) => {
      if (!player || !currentDatasetId.value) return false
      return wishlistStore.isInWishlist(currentDatasetId.value, player)
    }

    const handleAddToWishlist = async () => {
      if (contextMenuPlayer.value && currentDatasetId.value) {
        const success = await wishlistStore.addToWishlist(
          currentDatasetId.value,
          contextMenuPlayer.value
        )
        if (success) {
          $q.notify({
            type: 'positive',
            message: `${contextMenuPlayer.value.name} added to wishlist`,
            position: 'top',
            timeout: 2000,
          })
        } else {
          $q.notify({
            type: 'warning',
            message: `${contextMenuPlayer.value.name} is already in wishlist`,
            position: 'top',
            timeout: 2000,
          })
        }
      }
    }

    const handleRemoveFromWishlist = async () => {
      if (contextMenuPlayer.value && currentDatasetId.value) {
        const success = await wishlistStore.removeFromWishlist(
          currentDatasetId.value,
          contextMenuPlayer.value
        )
        if (success) {
          $q.notify({
            type: 'positive',
            message: `${contextMenuPlayer.value.name} removed from wishlist`,
            position: 'top',
            timeout: 2000,
          })
          if (props.showWishlistActions) {
            emit('remove-from-wishlist', contextMenuPlayer.value)
          }
        }
      }
    }

    const handlePlayerDetails = () => {
      if (contextMenuPlayer.value) {
        emit('player-selected', contextMenuPlayer.value)
      }
    }

    const onRightClick = (event, player) => {
      event.preventDefault()
      contextMenuPlayer.value = player
    }

    // Watch for changes in players prop and increment cache generation to force cache invalidation
    watch(
      () => props.players,
      () => {
        nextTick(() => {
          cacheGeneration.value++
          // Clear sorted players cache to force recalculation
          sortedPlayersCache.value = null
          lastSortKey.value = ''
        })
      },
      { deep: false }
    ) // shallow watch to detect reference changes

    // Clear memoization caches when component unmounts or when players prop changes significantly
    watch(
      () => props.players?.length,
      () => {
        // Clear all memoization caches when dataset changes
        clearCaches()
        getPositionIndex.clearCache()
      }
    )

    watch(
      () => props.isGoalkeeperView,
      () => {
        // Clear player value cache when view mode changes and increment generation
        nextTick(() => {
          cacheGeneration.value++
          clearCaches()
        })
      }
    )

    onUnmounted(() => {
      // Clear all caches on component cleanup
      clearCaches()
      getPositionIndex.clearCache()

      // Clear async sort timeout
      if (asyncSortTimeout.value) {
        clearTimeout(asyncSortTimeout.value)
      }
    })

    // Create a computed property for the table key
    const tableKey = computed(() => {
      return `table-${currentDatasetId.value || 'no-dataset'}-${sortField.value || 'Overall'}-${sortDirection.value || 'desc'}-${cacheGeneration.value || 0}`
    })

    return {
      qInstance: $q,
      cacheGeneration,
      sortField,
      sortDirection,
      pagination,
      onPaginationUpdate,
      pagesNumber,
      rowsPerPageOptions,
      maxPagesToShow,
      currentColumns,
      sortedPlayers,
      getColumnLabel,
      getUnifiedRatingClass,
      getTotalStatsRatingClass,
      getMoneyClass,
      getValueScoreClass,
      onFlagError,
      onRequest,
      onPageChange,
      onRowsPerPageChange,
      customSort,
      sortTable,
      onRowClick,
      onClubClick,
      formatDisplayCurrency,
      getDisplayValue,
      MAX_DISPLAY_PLAYERS,
      totalSortedCount,
      isSliced,
      isAsyncSorting,
      paginationTotalRows,
      paginationStartRow,
      paginationEndRow,
      contextMenu,
      contextMenuPlayer,
      isPlayerInWishlist,
      handleAddToWishlist,
      handleRemoveFromWishlist,
      handlePlayerDetails,
      onRightClick,
      tableKey,
    }
  },
}
</script>

<style lang="scss" scoped>
.player-data-table-container {
    background: transparent;
    
    // Ensure card sections have proper backgrounds
    .body--dark & {
        :deep(.q-card-section) {
            background: transparent !important;
            color: rgba(255, 255, 255, 0.9) !important;
        }
        
        :deep(.q-card-actions) {
            background: transparent !important;
            border-top: 1px solid rgba(255, 255, 255, 0.1) !important;
        }
    }
}

.table-controls-header {
    margin-bottom: 1rem;
    
    .sort-info-chip {
        background: rgba(46, 116, 181, 0.1);
        color: #2e74b5;
        font-weight: 600;
        padding: 0.5rem 1rem;
        border-radius: 8px;
        border: 1px solid rgba(46, 116, 181, 0.2);
        
        .body--dark & {
            background: rgba(96, 165, 250, 0.1);
            color: #60a5fa;
            border-color: rgba(96, 165, 250, 0.2);
        }
    }
}

.player-q-table {
    background: white;
    border-radius: 12px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06);
    border: 1px solid rgba(0, 0, 0, 0.06);
    overflow: hidden;
    
    .body--dark & {
        background: #1e293b;
        border: 1px solid rgba(255, 255, 255, 0.1);
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
    }
    
    .q-table__top {
        background: rgba(46, 116, 181, 0.03);
        border-bottom: 1px solid rgba(0, 0, 0, 0.06);
        
        .body--dark & {
            background: rgba(96, 165, 250, 0.05);
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
    }
    
    .q-table__bottom {
        background: rgba(46, 116, 181, 0.02);
        border-top: 1px solid rgba(0, 0, 0, 0.06);
        
        .body--dark & {
            background: rgba(96, 165, 250, 0.03);
            border-top: 1px solid rgba(255, 255, 255, 0.1);
        }
    }
}

.no-players-card {
    background: white !important;
    border: 1px solid rgba(46, 116, 181, 0.15) !important;
    border-radius: 12px !important;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06) !important;
    
    .body--dark & {
        background: #1e293b !important;
        border: 1px solid rgba(255, 255, 255, 0.1) !important;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3) !important;
    }
    
    .no-players-text {
        color: #64748b !important;
        font-size: 1.1rem;
        font-weight: 500;
        margin: 0;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.7) !important;
        }
    }
}



@keyframes sortProgress {
    0% {
        transform: translateX(-100%);
    }
    100% {
        transform: translateX(100%);
    }
}

@keyframes gentlePulse {
    0%, 100% {
        transform: scale(1);
        opacity: 1;
    }
    50% {
        transform: scale(1.02);
        opacity: 0.9;
    }
}

@keyframes subtleGlow {
    0%, 100% {
        opacity: 1;
    }
    50% {
        opacity: 0.8;
    }
}

@keyframes shimmer {
    0% {
        transform: translateX(-100%);
    }
    100% {
        transform: translateX(100%);
    }
}

.table-sorting {
    .modern-table-header {
        background: linear-gradient(180deg, rgba(46, 116, 181, 0.1) 0%, rgba(46, 116, 181, 0.15) 100%);
        
        .body--dark & {
            background: linear-gradient(180deg, rgba(255, 255, 255, 0.1) 0%, rgba(255, 255, 255, 0.15) 100%);
        }
    }
}

:deep(.q-table__bottom .q-select .q-field__native),
:deep(.q-table__bottom .q-select .q-field__input) {
    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

/* Fixed table layout for consistent column widths */
.fixed-table {
    table-layout: fixed;
    width: 100%;
}

.fixed-table .q-table__top {
    border-bottom: 1px solid rgba(0, 0, 0, 0.12);
}

.body--dark .fixed-table .q-table__top {
    border-color: rgba(255, 255, 255, 0.12);
}

// Full-screen table enhancements
.full-screen-table {
    height: 100% !important;
    
    .q-table {
        height: 100% !important;
        display: flex;
        flex-direction: column;
    }
    
    .q-table__container {
        flex: 1 !important;
        height: 100% !important;
        overflow: hidden;
    }
    
    .q-table__middle {
        flex: 1 !important;
        height: 100% !important;
        overflow: hidden;
    }
    
    .q-virtual-scroll {
        height: 100% !important;
    }
}

.player-data-table-container {
    height: 100%;
    display: flex;
    flex-direction: column;
    
    .relative-position {
        flex: 1;
        height: 100%;
        overflow: hidden;
    }
}
</style>