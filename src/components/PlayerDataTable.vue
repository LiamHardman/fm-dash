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
            <q-btn
                flat
                dense
                round
                icon="view_column"
                class="customize-columns-btn"
                @click="showColumnCustomizeDialog = true"
            >
                <q-tooltip>Customize columns</q-tooltip>
            </q-btn>
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
                :columns="displayedColumns"
                :loading="loading"
                :row-key="(row) => row.uid ?? row.UID ?? row.name"
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
                virtual-scroll
                :virtual-scroll-item-size="virtualScrollItemSize"
                :virtual-scroll-sticky-size-start="48"
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
                    :column-max-values="columnMaxValues"
                    @row-click="onRowClick"
                    @right-click="onRightClick"
                    @club-click="onClubClick"
                    @division-click="onDivisionClick"
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
            :is-player-in-shortlist="contextMenuPlayer ? isPlayerInShortlist(contextMenuPlayer) : false"
            :is-player-in-comparison="contextMenuPlayer ? comparisonStore.isInComparison(currentDatasetId, contextMenuPlayer) : false"
            :has-match-filters="Object.keys(matchFilters || {}).length > 0"
            @add-to-shortlist="handleAddToShortlist"
            @remove-from-shortlist="handleRemoveFromShortlist"
            @player-details="handlePlayerDetails"
            @add-to-comparison="handleAddToComparison"
            @remove-from-comparison="handleRemoveFromComparison"
            @why-this-player="handleWhyThisPlayer"
        />

        <CustomizeListDialog
            v-model="showColumnCustomizeDialog"
            title="Customize columns"
            hint="Reorder or hide player table columns. The Name column always stays first and can't be hidden."
            :items="columnPickerItems"
            :hidden-ids="uiStore.playerTableHiddenColumns"
            :order-ids="uiStore.playerTableColumnOrder"
            @update:hidden="uiStore.setPlayerTableHiddenColumns"
            @update:order="uiStore.setPlayerTableColumnOrder"
            @reset="resetColumnCustomization"
        />
    </div>
    </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerSorting } from '../composables/usePlayerSorting'
import { usePlayerTableCells } from '../composables/usePlayerTableCells'
import { usePlayerTableColumns } from '../composables/usePlayerTableColumns'
import { useComparisonStore } from '../stores/comparisonStore'
import { usePlayerStore } from '../stores/playerStore'
import { useShortlistStore } from '../stores/shortlistStore'
import { useUiStore } from '../stores/uiStore'
import { explainPlayerMatch } from '../utils/savedSearch'
import CustomizeListDialog from './layout/CustomizeListDialog.vue'
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
    CustomizeListDialog,
  },
  props: {
    players: { type: Array, required: true },
    loading: { type: Boolean, default: false },
    isGoalkeeperView: { type: Boolean, default: false },
    currencySymbol: { type: String, default: '$' },
    filteredPlayerCount: { type: Number, default: 0 },
    showWishlistActions: { type: Boolean, default: false },
    datasetId: { type: String, default: null },
    matchFilters: { type: Object, default: () => ({}) },
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
    const shortlistStore = useShortlistStore()
    const comparisonStore = useComparisonStore()
    const router = useRouter()
    const uiStore = useUiStore()

    // Reactive CA display preference from settings
    const showCA = computed(() => uiStore.showCA)

    // Estimated row height (px) fed to q-table's virtual scroller. This is a
    // rendering-correctness companion to the CSS --density-row-height/
    // --density-cell-padding tokens (see PlayerTableRow.vue/PlayerTableHeader.vue):
    // virtual-scroll-item-size is a plain number prop, not a CSS var, so it has to
    // be kept in sync with density here or the virtual scroller's row-position
    // math drifts from the actual (density-aware) rendered row height once rows
    // shrink in compact mode.
    const virtualScrollItemSize = computed(() => (uiStore.density === 'compact' ? 50 : 68))

    // Initialize optimized sorting and web workers

    // Cache generation key that changes when players data changes
    const cacheGeneration = ref(0)

    // Initialize composables
    const { currentColumns, getColumnLabel, getSortFieldKey } = usePlayerTableColumns(
      () => props.isGoalkeeperView,
      () => props.showValueScore,
      showCA
    )

    // Column visibility/order customization (ticket 11 of the UI redesign map).
    // 'name' is never hideable/orderable -- it's excluded from the picker and
    // always forced first in displayedColumns regardless of stored order.
    // "Pinning" a column is just moving it to the front of the order list.
    const showColumnCustomizeDialog = ref(false)

    const columnPickerItems = computed(() =>
      currentColumns.value
        .filter((c) => c.name !== 'name')
        .map((c) => ({ id: c.name, label: c.label, icon: 'view_column' }))
    )

    const displayedColumns = computed(() => {
      const nameCol = currentColumns.value.find((c) => c.name === 'name')
      const rest = currentColumns.value.filter((c) => c.name !== 'name')
      const hidden = new Set(uiStore.playerTableHiddenColumns)
      const order = uiStore.playerTableColumnOrder

      let visible = rest.filter((c) => !hidden.has(c.name))
      if (order.length) {
        visible = [...visible].sort((a, b) => {
          const ai = order.indexOf(a.name)
          const bi = order.indexOf(b.name)
          return (
            (ai === -1 ? Number.POSITIVE_INFINITY : ai) -
            (bi === -1 ? Number.POSITIVE_INFINITY : bi)
          )
        })
      }

      return nameCol ? [nameCol, ...visible] : visible
    })

    const resetColumnCustomization = () => {
      uiStore.setPlayerTableHiddenColumns([])
      uiStore.setPlayerTableColumnOrder([])
    }

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
    } = usePlayerTableCells(
      () => props.isGoalkeeperView,
      () => props.currencySymbol,
      cacheGeneration
    )

    const contextMenu = ref(null)
    const sortField = ref(props.defaultSortField)
    const sortDirection = ref(props.defaultSortDirection)
    const rowsPerPageOptions = ref([10, 15, 20, 50, 0]) // Keep for internal logic, but selector is removed
    const maxPagesToShow = 7
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
      { deep: false }
    )

    const onPaginationUpdate = (newPagination) => {
      pagination.value = newPagination
    }

    const sortSpecification = computed(() => ({
      field: getSortFieldKey(sortField.value || 'Overall'),
      direction: sortDirection.value,
      goalkeeper: props.isGoalkeeperView,
    }))
    const { sortedPlayers, isAsyncSorting, totalSortedCount, isSliced } = usePlayerSorting(
      () => props.players || [],
      sortSpecification,
      MAX_DISPLAY_PLAYERS
    )

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
      }
    }

    const onDivisionClick = (division) => {
      if (division && division.trim() !== '') {
        router.push({
          name: 'leagues',
          query: { league: division, datasetId: currentDatasetId.value },
        })
      }
    }

    const contextMenuPlayer = ref(null)

    const isPlayerInShortlist = (player) => {
      if (!player || !currentDatasetId.value) return false
      return shortlistStore.activeList?.items?.some(
        (item) =>
          item.playerRef.datasetId === currentDatasetId.value &&
          Number(item.playerRef.playerUid) === Number(player.uid ?? player.UID)
      )
    }

    const handleAddToShortlist = async () => {
      if (contextMenuPlayer.value && currentDatasetId.value) {
        const success = await shortlistStore.addPlayer(
          currentDatasetId.value,
          contextMenuPlayer.value
        )
        if (success) {
          $q.notify({
            type: 'positive',
            message: `${contextMenuPlayer.value.name} added to Shortlist`,
            position: 'top',
            timeout: 2000,
          })
        } else {
          $q.notify({
            type: 'warning',
            message: `${contextMenuPlayer.value.name} is already in this Shortlist`,
            position: 'top',
            timeout: 2000,
          })
        }
      }
    }

    const handleRemoveFromShortlist = async () => {
      if (contextMenuPlayer.value && currentDatasetId.value && shortlistStore.activeList) {
        const item = shortlistStore.activeList.items.find(
          (candidate) =>
            candidate.playerRef.datasetId === currentDatasetId.value &&
            Number(candidate.playerRef.playerUid) ===
              Number(contextMenuPlayer.value.uid ?? contextMenuPlayer.value.UID)
        )
        if (item) {
          await shortlistStore.removeItem(item)
          $q.notify({
            type: 'positive',
            message: `${contextMenuPlayer.value.name} removed from Shortlist`,
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

    const handleWhyThisPlayer = () => {
      if (!contextMenuPlayer.value) return
      $q.notify({
        type: 'info',
        icon: 'help_outline',
        message: `${contextMenuPlayer.value.name}: ${explainPlayerMatch(props.matchFilters, contextMenuPlayer.value).join(' · ')}`,
        timeout: 7000,
        multiLine: true,
      })
    }

    const handleAddToComparison = () => {
      if (!contextMenuPlayer.value || !currentDatasetId.value) return
      const added = comparisonStore.addToComparison(currentDatasetId.value, contextMenuPlayer.value)
      if (added) {
        $q.notify({
          type: 'positive',
          message: `${contextMenuPlayer.value.name} added to comparison`,
          timeout: 1500,
        })
      } else if (
        comparisonStore.getCount(currentDatasetId.value) >= comparisonStore.MAX_COMPARISON_PLAYERS
      ) {
        $q.notify({ type: 'warning', message: 'Comparison is full (max 4 players)', timeout: 1500 })
      }
    }

    const handleRemoveFromComparison = () => {
      if (!contextMenuPlayer.value || !currentDatasetId.value) return
      comparisonStore.removeFromComparison(currentDatasetId.value, contextMenuPlayer.value)
      $q.notify({
        type: 'info',
        message: `${contextMenuPlayer.value.name} removed from comparison`,
        timeout: 1500,
      })
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
    })

    // Create a computed property for the table key
    const tableKey = computed(() => {
      return `table-${currentDatasetId.value || 'no-dataset'}-${sortField.value || 'Overall'}-${sortDirection.value || 'desc'}-${cacheGeneration.value || 0}`
    })

    // Dynamic max values for gauge columns where the scale is data-relative (e.g. TotalStats)
    const columnMaxValues = computed(() => {
      if (!props.players || props.players.length === 0) return {}
      let maxTotal = 0
      for (const p of props.players) {
        const v = p.totalStats
        if (typeof v === 'number' && v > maxTotal) maxTotal = v
      }
      return { TotalStats: maxTotal || 1 }
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
      comparisonStore,
      isPlayerInShortlist,
      handleAddToShortlist,
      handleRemoveFromShortlist,
      handlePlayerDetails,
      handleWhyThisPlayer,
      handleAddToComparison,
      handleRemoveFromComparison,
      onRightClick,
      onDivisionClick,
      tableKey,
      columnMaxValues,
      uiStore,
      virtualScrollItemSize,
      displayedColumns,
      showColumnCustomizeDialog,
      columnPickerItems,
      resetColumnCustomization,
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
        background: var(--accent-soft);
        color: var(--accent);
        font-weight: 600;
        padding: 0.5rem 1rem;
        border-radius: 8px;
        border: 1px solid var(--accent-soft-strong);
    }
}

.player-q-table {
    background: var(--surface-card);
    border-radius: 12px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06);
    border: 1px solid rgba(0, 0, 0, 0.06);
    overflow: hidden;

    .body--dark & {
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
    background: var(--surface-card) !important;
    border: 1px solid var(--accent-soft-strong) !important;
    border-radius: 12px !important;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06) !important;

    .body--dark & {
        border: 1px solid rgba(255, 255, 255, 0.1) !important;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3) !important;
    }

    .no-players-text {
        color: var(--text-secondary) !important;
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
