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
                <q-tr
                    :props="props"
                    class="modern-table-header"
                >
                    <q-th
                        v-for="col in props.cols"
                        :key="col.name"
                        :props="props"
                        class="text-weight-bold modern-header-cell"
                        :class="[
                            col.headerClasses,
                            { 'active-sort': sortField === col.name },
                            { 'cursor-pointer': true },
                            { 'sorting-in-progress': isAsyncSorting && sortField === col.name }
                        ]"
                        :style="col.headerStyle"
                        @click="sortTable(col.name)"
                    >
                        <span
                            v-if="
                                col.name === 'transfer_value' ||
                                col.name === 'wage'
                            "
                        >
                            {{ col.label }} ({{ currencySymbol }})
                        </span>
                        <span v-else>
                            {{ col.label }}
                        </span>
                        <q-icon
                            v-if="sortField === col.name && !isAsyncSorting"
                            :name="
                                sortDirection === 'asc'
                                    ? 'arrow_upward'
                                    : 'arrow_downward'
                            "
                            size="xs"
                            class="q-ml-xs sort-icon"
                        />
                        <!-- Subtle sorting indicator -->
                        <q-spinner-dots
                            v-if="sortField === col.name && isAsyncSorting"
                            size="xs"
                            color="primary"
                            class="q-ml-xs sorting-spinner"
                        />
                    </q-th>
                </q-tr>
            </template>

            <template v-slot:body="props">
                <q-tr
                    :props="props"
                    @click="onRowClick(props.row)"
                    @contextmenu="onRightClick($event, props.row)"
                    class="cursor-pointer table-row-hover modern-table-row"
                >
                    <q-td
                        v-for="col in props.cols"
                        :key="col.name"
                        :props="props"
                        :class="[
                            col.classes,
                            'table-cell-enhanced',
                        ]"
                        :style="col.style"
                    >
                        <template v-if="col.isFifaStat || col.isOverallStat">
                            <span
                                :class="
                                    getUnifiedRatingClass(
                                        getDisplayValue(props.row, col),
                                        100,
                                    )
                                "
                                class="attribute-value fifa-stat-value modern-stat-badge"
                            >
                                {{
                                    getDisplayValue(props.row, col) !== undefined
                                        ? getDisplayValue(props.row, col)
                                        : "-"
                                }}
                            </span>
                        </template>
                        <template v-else-if="col.isValueScore">
                            <span
                                :class="getValueScoreClass(props.row.valueScore)"
                                class="attribute-value value-score-value modern-stat-badge"
                            >
                                {{ 
                                    props.row.valueScore !== undefined && props.row.valueScore !== null
                                        ? Math.round(props.row.valueScore)
                                        : "-"
                                }}
                            </span>
                        </template>
                        <template
                            v-else-if="
                                col.name === 'transfer_value' ||
                                col.name === 'wage'
                            "
                        >
                            <span
                                :class="
                                    getMoneyClass(
                                        props.row[col.sortField || col.field],
                                    )
                                "
                                class="money-value"
                            >
                                {{
                                    formatDisplayCurrency(
                                        props.row[col.sortField || col.field],
                                        props.row[col.field],
                                    )
                                }}
                            </span>
                        </template>
                        <template
                            v-else-if="col.name === 'nationality_display'"
                        >
                            <div class="flex items-center no-wrap nationality-cell">
                                <img
                                    v-if="props.row.nationality_iso"
                                    :src="`https://flagcdn.com/w20/${props.row.nationality_iso.toLowerCase()}.png`"
                                    :alt="props.row.nationality || 'Flag'"
                                    width="20"
                                    height="13"
                                    class="nationality-flag flex-shrink-0"
                                    @error="onFlagError($event, props.row)"
                                />
                                <q-icon
                                    v-else
                                    name="flag"
                                    size="xs"
                                    :color="
                                        qInstance.dark.isActive
                                            ? 'grey-6'
                                            : 'grey-7'
                                    "
                                    class="nationality-flag-placeholder flex-shrink-0"
                                />
                                <span class="nationality-text">{{ props.row.nationality || "-" }}</span>
                            </div>
                        </template>
                        <template v-else-if="col.name === 'club'">
                            <div class="club-cell">
                                <span 
                                    class="club-link"
                                    @click.stop="onClubClick(props.row)"
                                    :title="`View ${props.row[col.field]} team page`"
                                >{{
                                    props.row[col.field] !== undefined &&
                                    props.row[col.field] !== null
                                        ? props.row[col.field]
                                        : "-"
                                }}</span>
                            </div>
                        </template>
                        <template v-else>
                            <span>{{
                                props.row[col.field] !== undefined &&
                                props.row[col.field] !== null
                                    ? props.row[col.field]
                                    : "-"
                            }}</span>
                        </template>
                    </q-td>
                </q-tr>
            </template>

            <template v-slot:loading>
                <q-inner-loading showing color="primary">
                    <q-spinner size="50px" color="primary" />
                </q-inner-loading>
            </template>

            <template v-slot:pagination="scope">
                <q-pagination
                    v-model="scope.pagination.page"
                    :max="pagesNumber"
                    :max-pages="maxPagesToShow"
                    boundary-links
                    direction-links
                    @update:model-value="onPageChange"
                    color="primary"
                    active-color="primary"
                    text-color="primary"
                    active-text-color="white"
                />
                <q-space />
                <span
                    class="q-ml-md text-caption pagination-info"
                >
                    {{ paginationStartRow }} - {{ paginationEndRow }} of
                    {{ paginationTotalRows }}
                    <span v-if="isSliced" class="text-italic q-ml-xs"
                        >(from {{ totalSortedCount }} total sorted)</span
                    >
                </span>
            </template>
        </q-table>
    </div>

        <!-- Context Menu -->
        <q-menu 
            ref="contextMenu"
            touch-position 
            context-menu
            :offset="[10, 10]"
        >
            <q-list dense style="min-width: 180px">
                <q-item 
                    clickable 
                    v-close-popup 
                    @click="handleAddToWishlist"
                    v-if="contextMenuPlayer && !isPlayerInWishlist(contextMenuPlayer)"
                >
                    <q-item-section avatar>
                        <q-icon name="favorite_border" color="positive" />
                    </q-item-section>
                    <q-item-section>Add to Wishlist</q-item-section>
                </q-item>
                
                <q-item 
                    clickable 
                    v-close-popup 
                    @click="handleRemoveFromWishlist"
                    v-if="contextMenuPlayer && isPlayerInWishlist(contextMenuPlayer)"
                >
                    <q-item-section avatar>
                        <q-icon name="favorite" color="negative" />
                    </q-item-section>
                    <q-item-section>Remove from Wishlist</q-item-section>
                </q-item>
                
                <q-separator />
                
                <q-item 
                    clickable 
                    v-close-popup 
                    @click="handlePlayerDetails"
                    v-if="contextMenuPlayer"
                >
                    <q-item-section avatar>
                        <q-icon name="info" color="info" />
                    </q-item-section>
                    <q-item-section>View Details</q-item-section>
                </q-item>
            </q-list>
        </q-menu>
    </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { memoize } from '../composables/useMemoization'
import { useOptimizedSorting } from '../composables/useVirtualScrolling'
import { usePlayerCalculationWorker } from '../composables/useWebWorkers'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
import { useWishlistStore } from '../stores/wishlistStore'
import { formatCurrency } from '../utils/currencyUtils'

const MAX_DISPLAY_PLAYERS = 1000

export default {
  name: 'PlayerDataTable',
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
    defaultSortDirection: { type: String, default: 'desc' }
  },
  emits: [
    'update:sort',
    'player-selected',
    'update:pagination',
    'team-selected',
    'remove-from-wishlist'
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

    const contextMenu = ref(null)
    const sortField = ref(props.defaultSortField)
    const sortDirection = ref(props.defaultSortDirection)
    const rowsPerPageOptions = ref([10, 15, 20, 50, 0]) // Keep for internal logic, but selector is removed
    const maxPagesToShow = 7
    const totalSortedCount = ref(0)
    const isAsyncSorting = ref(false)
    const isSliced = ref(false)
    const currentSortController = ref(null) // For cancelling sort operations

    // Cache generation key that changes when players data changes
    const cacheGeneration = ref(0)

    const pagination = ref({
      sortBy: props.defaultSortField,
      descending: props.defaultSortDirection === 'desc',
      page: 1,
      rowsPerPage: 50 // Default rows per page, even if selector is hidden
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
      'ST'
    ]

    // Memoized position index calculation (expensive string processing)
    const getPositionIndex = memoize(
      positionString => {
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
          .map(p => p.trim())
          .filter(p => p.length > 0)
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
        keyGenerator: positionString => positionString,
        cacheKey: 'positionIndex'
      }
    )

    const onPaginationUpdate = newPagination => {
      pagination.value = newPagination
    }

    // Column definitions with fixed widths to prevent layout shifts
    const nameColumnStyle =
      'width: 200px; min-width: 200px; max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'
    const ageColumnStyle =
      'width: 60px; min-width: 60px; max-width: 60px; text-align: center; white-space: nowrap;'
    const positionColumnStyle =
      'width: 150px; min-width: 150px; max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'
    const clubColumnStyle =
      'width: 180px; min-width: 180px; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'
    const moneyColumnStyle =
      'width: 110px; min-width: 110px; max-width: 110px; text-align: right; white-space: nowrap;'
    const overallColumnStyle =
      'width: 70px; min-width: 70px; max-width: 70px; text-align: center; white-space: nowrap;'
    const fifaStatColumnStyle =
      'width: 60px; min-width: 60px; max-width: 60px; text-align: center; white-space: nowrap;'
    const textColumnStyle =
      'width: 120px; min-width: 120px; max-width: 120px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'
    const nationalityColumnStyle =
      'width: 150px; min-width: 150px; max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'

    const baseColumnDefinitions = {
      name: {
        name: 'name',
        label: 'Name',
        field: 'name',
        sortable: true,
        align: 'left',
        style: nameColumnStyle,
        headerStyle: nameColumnStyle
      },
      age: {
        name: 'age',
        label: 'Age',
        field: 'age',
        sortable: true,
        align: 'center',
        style: ageColumnStyle,
        headerStyle: ageColumnStyle
      },
      position: {
        name: 'position',
        label: 'Position',
        field: 'position',
        sortable: true,
        align: 'left',
        style: positionColumnStyle,
        headerStyle: positionColumnStyle
      },
      club: {
        name: 'club',
        label: 'Club',
        field: 'club',
        sortable: true,
        align: 'left',
        style: clubColumnStyle,
        headerStyle: clubColumnStyle
      },
      transfer_value: {
        name: 'transfer_value',
        label: 'Value',
        field: 'transfer_value',
        sortable: true,
        align: 'right',
        sortField: 'transferValueAmount',
        style: moneyColumnStyle,
        headerStyle: moneyColumnStyle
      },
      wage: {
        name: 'wage',
        label: 'Salary',
        field: 'wage',
        sortable: true,
        align: 'right',
        sortField: 'wageAmount',
        style: moneyColumnStyle,
        headerStyle: moneyColumnStyle
      },
      Overall: {
        name: 'Overall',
        label: 'Overall',
        field: 'Overall',
        sortable: true,
        align: 'center',
        isOverallStat: true,
        style: overallColumnStyle,
        headerStyle: overallColumnStyle
      },
      valueScore: {
        name: 'valueScore',
        label: 'Value Score',
        field: 'valueScore',
        sortable: true,
        align: 'center',
        isValueScore: true,
        style: overallColumnStyle,
        headerStyle: overallColumnStyle
      },
      personality: {
        name: 'personality',
        label: 'Personality',
        field: 'personality',
        sortable: true,
        align: 'left',
        style: textColumnStyle,
        headerStyle: textColumnStyle
      },
      media_handling: {
        name: 'media_handling',
        label: 'Media Desc.',
        field: 'media_handling',
        sortable: true,
        align: 'left',
        style: textColumnStyle,
        headerStyle: textColumnStyle
      },
      nationality_display: {
        name: 'nationality_display',
        label: 'Nationality',
        field: 'nationality',
        sortable: true,
        align: 'left',
        style: nationalityColumnStyle,
        headerStyle: nationalityColumnStyle
      }
    }

    const allFifaStatDefinitions = {
      GK: {
        name: 'GK',
        label: 'GK',
        field: 'GK',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      DIV: {
        name: 'DIV',
        label: 'DIV',
        field: 'DIV',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      HAN: {
        name: 'HAN',
        label: 'HAN',
        field: 'HAN',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      REF: {
        name: 'REF',
        label: 'REF',
        field: 'REF',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      KIC: {
        name: 'KIC',
        label: 'KIC',
        field: 'KIC',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      SPD: {
        name: 'SPD',
        label: 'SPD',
        field: 'SPD',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      POS: {
        name: 'POS',
        label: 'POS',
        field: 'POS',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      PAC: {
        name: 'PAC',
        label: 'PAC',
        field: 'PAC',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      SHO: {
        name: 'SHO',
        label: 'SHO',
        field: 'SHO',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      PAS: {
        name: 'PAS',
        label: 'PAS',
        field: 'PAS',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      DRI: {
        name: 'DRI',
        label: 'DRI',
        field: 'DRI',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      DEF: {
        name: 'DEF',
        label: 'DEF',
        field: 'DEF',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      },
      PHY: {
        name: 'PHY',
        label: 'PHY',
        field: 'PHY',
        sortable: true,
        align: 'center',
        isFifaStat: true,
        style: fifaStatColumnStyle,
        headerStyle: fifaStatColumnStyle
      }
    }

    // Regular computed for columns (removed memoization to fix reactivity issues)
    const currentColumns = computed(() => {
      const newOrderBase = [
        baseColumnDefinitions.name,
        baseColumnDefinitions.nationality_display,
        baseColumnDefinitions.age,
        baseColumnDefinitions.position,
        baseColumnDefinitions.club,
        baseColumnDefinitions.transfer_value,
        baseColumnDefinitions.wage,
        baseColumnDefinitions.Overall
      ]

      // Add value score column if enabled
      if (props.showValueScore) {
        newOrderBase.push(baseColumnDefinitions.valueScore)
      }

      const fifaColumnsInOrder = props.isGoalkeeperView
        ? [
            allFifaStatDefinitions.DIV,
            allFifaStatDefinitions.HAN,
            allFifaStatDefinitions.REF,
            allFifaStatDefinitions.KIC,
            allFifaStatDefinitions.SPD,
            allFifaStatDefinitions.POS
          ]
        : [
            allFifaStatDefinitions.PAC,
            allFifaStatDefinitions.SHO,
            allFifaStatDefinitions.PAS,
            allFifaStatDefinitions.DRI,
            allFifaStatDefinitions.DEF,
            allFifaStatDefinitions.PHY
          ]

      const trailingColumns = [
        baseColumnDefinitions.personality,
        baseColumnDefinitions.media_handling
      ]
      return [...newOrderBase, ...fifaColumnsInOrder, ...trailingColumns]
    })

    const getColumnLabel = fieldName => {
      const col = currentColumns.value.find(c => c.name === fieldName)
      return col ? col.label : fieldName
    }

    const getSortFieldKey = colName => {
      const colDef = currentColumns.value.find(c => c.name === colName)
      return colDef?.sortField || colDef?.field || colName
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
          timeout: 3000
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
        timeout: 2000
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
          isFifaStat: currentColumns.value.find(c => c.name === sortField.value)?.isFifaStat,
          isOverallStat: currentColumns.value.find(c => c.name === sortField.value)?.isOverallStat,
          isValueScore: currentColumns.value.find(c => c.name === sortField.value)?.isValueScore,
          displayField: sortField.value
        })
      }
    })

    // Memoized rating class calculation (called frequently in table rendering)
    const getUnifiedRatingClass = memoize(
      (value, maxScale) => {
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
      },
      {
        maxSize: 200, // Cache up to 200 different rating calculations
        keyGenerator: (value, maxScale) => `${value}-${maxScale}`,
        cacheKey: 'unifiedRatingClass'
      }
    )

    const getMoneyClass = numericAmount => {
      if (numericAmount === null || numericAmount === undefined) return 'money-na'
      return 'money-uniform'
    }

    const getValueScoreClass = valueScore => {
      if (valueScore === null || valueScore === undefined) return 'rating-na'
      const score = Number(valueScore)
      if (Number.isNaN(score)) return 'rating-na'

      if (score >= 80) return 'rating-tier-6' // Excellent value - highest tier
      if (score >= 60) return 'rating-tier-5' // Great value
      if (score >= 40) return 'rating-tier-4' // Good value
      if (score >= 20) return 'rating-tier-3' // Fair value
      if (score >= 0) return 'rating-tier-2' // Poor value
      return 'rating-na'
    }

    const onFlagError = event => {
      if (event.target) event.target.style.display = 'none'
      const placeholderIcon = event.target.nextElementSibling
      if (placeholderIcon?.classList.contains('q-icon')) {
        placeholderIcon.style.display = 'inline-flex'
      }
    }

    const onRequest = requestProp => {
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
          isFifaStat: currentColumns.value.find(c => c.name === sortBy)?.isFifaStat,
          isOverallStat: currentColumns.value.find(c => c.name === sortBy)?.isOverallStat,
          isValueScore: currentColumns.value.find(c => c.name === sortBy)?.isValueScore,
          displayField: sortBy
        })
      }
      emit('update:pagination', { ...pagination.value })
    }

    const onPageChange = newPage => {
      pagination.value.page = newPage
    }

    const onRowsPerPageChange = newRowsPerPage => {
      pagination.value.rowsPerPage = newRowsPerPage
      pagination.value.page = 1
    }

    const customSort = rows => {
      // The actual sorting is now done in the `sortedPlayers` computed property.
      // QTable's `sort-method` is still needed, but we just return the rows as they are
      // because our computed property `sortedPlayers` (bound to :rows) already handles it.
      return rows
    }

    const sortTable = fieldName => {
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
        const colDef = currentColumns.value.find(c => c.name === fieldName)
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
        isFifaStat: currentColumns.value.find(c => c.name === fieldName)?.isFifaStat,
        isOverallStat: currentColumns.value.find(c => c.name === fieldName)?.isOverallStat,
        isValueScore: currentColumns.value.find(c => c.name === fieldName)?.isValueScore,
        displayField: fieldName
      })
    }

    const onRowClick = player => {
      emit('player-selected', player)
    }

    const onClubClick = player => {
      if (player.club && player.club.trim() !== '') {
        emit('team-selected', player.club)
      } else {
      }
    }

    const formatDisplayCurrency = (numericAmount, originalDisplayValue) => {
      return formatCurrency(numericAmount, props.currencySymbol, originalDisplayValue)
    }

    // GK stat mapping for both display and sorting consistency
    const gkStatMapping = {
      PAC: 'DIV', // Diving -> Pace
      SHO: 'HAN', // Handling -> Shooting
      PAS: 'KIC', // Kicking -> Passing
      DRI: 'REF', // Reflexes -> Dribbling
      DEF: 'SPD', // Speed -> Defending
      PHY: 'POS' // Positioning -> Physical
    }

    // Memoized player value getter (called frequently during sorting and rendering)
    const getPlayerValue = (player, fieldKey, columnName = null) => {
      // For non-goalkeeper view, map GK stats to standard FIFA stats if the player is a goalkeeper
      if (!props.isGoalkeeperView && player.position && player.position.includes('GK')) {
        const mappedStat = gkStatMapping[columnName || fieldKey]
        if (mappedStat && player[mappedStat] !== undefined) {
          return player[mappedStat]
        }
      }

      // Default behavior - use the field key
      const value = player[fieldKey]

      return value
    }

    // Memoized version for non-Overall fields only
    const getPlayerValueMemoized = memoize(
      (player, fieldKey, columnName = null) => {
        return getPlayerValue(player, fieldKey, columnName)
      },
      {
        maxSize: 1000,
        keyGenerator: (player, fieldKey, columnName) => {
          // Try to use the player's UID for cache key
          let playerUID = player.UID || player.uid

          // If no UID available or UID is empty, create a composite unique key
          if (!playerUID || playerUID === '') {
            playerUID = `${player.name || 'unknown'}-${player.club || 'unknown'}-${player.age || 'unknown'}-${player.position || 'unknown'}`
          }

          return `gen${cacheGeneration.value}-${playerUID}-${fieldKey}-${columnName || ''}`
        },
        cacheKey: 'playerValue'
      }
    )

    const getDisplayValue = (player, col) => {
      // For Overall field, always use non-memoized version to ensure reactivity
      if (col.field === 'Overall' || col.name === 'Overall') {
        return getPlayerValue(player, col.field, col.name)
      }
      // For other fields, use memoized version for performance
      return getPlayerValueMemoized(player, col.field, col.name)
    }

    const contextMenuPlayer = ref(null)

    const isPlayerInWishlist = player => {
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
            timeout: 2000
          })
        } else {
          $q.notify({
            type: 'warning',
            message: `${contextMenuPlayer.value.name} is already in wishlist`,
            position: 'top',
            timeout: 2000
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
            timeout: 2000
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
        cacheGeneration.value++
        // Clear sorted players cache to force recalculation
        sortedPlayersCache.value = null
        lastSortKey.value = ''
      },
      { deep: false }
    ) // shallow watch to detect reference changes

    // Clear memoization caches when component unmounts or when players prop changes significantly
    watch(
      () => props.players?.length,
      () => {
        // Clear all memoization caches when dataset changes
        getUnifiedRatingClass.clearCache()
        getPlayerValueMemoized.clearCache()
        getPositionIndex.clearCache()
      }
    )

    watch(
      () => props.isGoalkeeperView,
      () => {
        // Clear player value cache when view mode changes and increment generation
        cacheGeneration.value++
        getPlayerValueMemoized.clearCache()
      }
    )

    onUnmounted(() => {
      // Clear all caches on component cleanup
      getUnifiedRatingClass.clearCache()
      getPlayerValueMemoized.clearCache()
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
      tableKey
    }
  }
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

.modern-table-header {
    background: linear-gradient(180deg, rgba(46, 116, 181, 0.08) 0%, rgba(46, 116, 181, 0.12) 100%);
    
    .body--dark & {
        background: linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.12) 100%);
    }
}

.modern-header-cell {
    color: #1e293b !important;
    font-weight: 700 !important;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-size: 0.8rem !important;
    padding: 1rem 0.75rem !important;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.9) !important;
    }

    &:hover {
        background: rgba(46, 116, 181, 0.15) !important;
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.15) !important;
        }
    }
    
    &.active-sort {
        background: rgba(46, 116, 181, 0.2) !important;
        color: #2e74b5 !important;
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.2) !important;
            color: #60a5fa !important;
        }
        
        .sort-icon {
            color: #2e74b5 !important;
            
            .body--dark & {
                color: #60a5fa !important;
            }
        }
    }
    
    &.sorting-in-progress {
        background: rgba(46, 116, 181, 0.15) !important;
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.15) !important;
        }
        
        .sorting-spinner {
            color: #2e74b5 !important;
            
            .body--dark & {
                color: #60a5fa !important;
            }
        }
    }
}

.modern-table-row {
    transition: all 0.2s ease;
    
    &:hover {
        background: rgba(46, 116, 181, 0.04) !important;
        box-shadow: 0 2px 8px rgba(46, 116, 181, 0.15);
        transform: translateY(-1px);
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.04) !important;
            box-shadow: 0 2px 8px rgba(255, 255, 255, 0.15);
        }
    }
    
    &:nth-child(even) {
        background: rgba(46, 116, 181, 0.02);
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.02);
        }
        
        &:hover {
            background: rgba(46, 116, 181, 0.06) !important;
            
            .body--dark & {
                background: rgba(255, 255, 255, 0.06) !important;
            }
        }
    }
}

.table-cell-enhanced {
    color: #334155 !important;
    font-weight: 500;
    padding: 0.75rem !important;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.85) !important;
    }
}

.modern-stat-badge {
    padding: 4px 10px;
    border-radius: 8px;
    font-size: 0.8rem;
    font-weight: 700;
    text-align: center;
    min-width: 36px;
    display: inline-block;
    border: 1px solid transparent;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    
    .body--dark & {
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
    }
}

.money-value {
    font-weight: 500;
}

.money-uniform {
    color: #334155;
    .body--dark & {
        color: rgba(255, 255, 255, 0.85);
    }
}

.money-na {
    color: #9ca3af;
    .body--dark & {
        color: #6b7280;
    }
}

.nationality-flag {
    border: 1px solid rgba(0, 0, 0, 0.15);
    object-fit: cover;
    margin-right: 8px;
    width: 20px !important;
    height: 13px !important;
    flex-shrink: 0;
    border-radius: 3px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);

    .body--dark & {
        border: 1px solid rgba(255, 255, 255, 0.15);
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
    }
}

.nationality-flag-placeholder {
    margin-right: 8px;
    width: 20px;
    height: 13px;
    flex-shrink: 0;
    color: #9ca3af;
    
    .body--dark & {
        color: #6b7280;
    }
}

.nationality-cell {
    width: 100%;
    overflow: hidden;
    
    .nationality-text {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        flex: 1;
        min-width: 0;
        font-weight: 500;
    }
}

.club-cell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.club-cell .club-link {
    cursor: pointer;
    color: inherit;
    text-decoration: none;
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.club-cell .club-link:hover {
    text-decoration: underline;
}

.body--dark .club-cell .club-link:hover {
    color: #81C784;
}

.body--light .club-cell .club-link:hover {
    color: #2E7D32;
}

.pagination-info {
    color: #64748b !important;
    font-weight: 500;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7) !important;
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