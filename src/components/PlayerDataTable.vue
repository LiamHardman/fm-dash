<template>
    <div class="player-data-table-container">
        <div class="row items-center q-mb-sm table-controls-header">
            <div class="col">
                <div
                    v-if="sortField"
                    class="text-caption q-pa-xs rounded-borders sort-info-chip"
                    :class="
                        qInstance.dark.isActive
                            ? 'bg-grey-8 text-grey-4'
                            : 'bg-grey-2 text-grey-7'
                    "
                >
                    Current Sort: {{ getColumnLabel(sortField) }} ({{
                        sortDirection === "asc" ? "Ascending" : "Descending"
                    }})
                </div>
            </div>
        </div>

        <q-card
            v-if="players.length === 0 && !loading"
            class="q-pa-md text-center"
            :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-grey-1'"
            flat
            bordered
        >
            <p :class="qInstance.dark.isActive ? 'text-grey-5' : 'text-grey-7'">
                No players match your search criteria.
            </p>
        </q-card>
        <q-card
            v-else-if="
                sortedPlayers.length === 0 && players.length > 0 && !loading
            "
            class="q-pa-md text-center"
            :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-grey-1'"
            flat
            bordered
        >
            <p :class="qInstance.dark.isActive ? 'text-grey-5' : 'text-grey-7'">
                No players to display with current sort (possibly all filtered
                out before slicing).
            </p>
        </q-card>

        <div v-else class="relative-position">
            <q-table
                :key="`table-${currentDatasetId}-${sortField}-${sortDirection}`"
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
                style="height: 70vh; min-height: 400px; max-height: 800px;"
                separator="horizontal"
                no-data-label="No players to display"
            >
            <template v-slot:header="props">
                <q-tr
                    :props="props"
                    :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-grey-2'"
                    class="modern-table-header"
                >
                    <q-th
                        v-for="col in props.cols"
                        :key="col.name"
                        :props="props"
                        class="text-weight-bold modern-header-cell"
                        :class="[
                            qInstance.dark.isActive
                                ? 'text-grey-3'
                                : 'text-grey-8',
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
                            qInstance.dark.isActive
                                ? 'text-grey-4'
                                : 'text-grey-9',
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
                    :color="qInstance.dark.isActive ? 'grey-6' : 'primary'"
                    :active-color="
                        qInstance.dark.isActive ? 'primary' : 'primary'
                    "
                    :text-color="qInstance.dark.isActive ? 'white' : 'primary'"
                    :active-text-color="
                        qInstance.dark.isActive ? 'white' : 'white'
                    "
                />
                <q-space />
                <span
                    class="q-ml-md text-caption"
                    :class="
                        qInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'
                    "
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
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from "vue";
import { useQuasar } from "quasar";
import { usePlayerStore } from "../stores/playerStore";
import { useWishlistStore } from "../stores/wishlistStore";
import { formatCurrency } from "../utils/currencyUtils";
import { useOptimizedSorting } from "../composables/useVirtualScrolling";
import { memoize, memoizedComputed, useMemoizedComputeds } from "../composables/useMemoization";
import { usePlayerCalculationWorker } from "../composables/useWebWorkers";

const MAX_DISPLAY_PLAYERS = 1000;

export default {
    name: "PlayerDataTable",
    props: {
        players: { type: Array, required: true },
        loading: { type: Boolean, default: false },
        isGoalkeeperView: { type: Boolean, default: false },
        currencySymbol: { type: String, default: "$" },
        filteredPlayerCount: { type: Number, default: 0 },
        showWishlistActions: { type: Boolean, default: false },
    },
    emits: ["update:sort", "player-selected", "update:pagination", "team-selected", "remove-from-wishlist"],

    setup(props, { emit }) {
        console.log(`PlayerDataTable: Setup function start.`); // Static label
        const $q = useQuasar();
        const playerStore = usePlayerStore();
        const wishlistStore = useWishlistStore();
        
        // Initialize optimized sorting and web workers
        const { sortLargeArray, clearSortCache } = useOptimizedSorting();
        const { sortPlayers: sortPlayersWorker, getPendingTaskCount, terminateWorker } = usePlayerCalculationWorker();
        
        const contextMenu = ref(null);
        const sortField = ref("Overall");
        const sortDirection = ref("desc");
        const rowsPerPageOptions = ref([10, 15, 20, 50, 0]); // Keep for internal logic, but selector is removed
        const maxPagesToShow = 7;
        const totalSortedCount = ref(0);
        const isAsyncSorting = ref(false);
        const isSliced = ref(false);
        const currentSortController = ref(null); // For cancelling sort operations
        
        // Cache generation key that changes when players data changes
        const cacheGeneration = ref(0);

        const pagination = ref({
            sortBy: "Overall",
            descending: true,
            page: 1,
            rowsPerPage: 50, // Default rows per page, even if selector is hidden
        });

        // Get current dataset ID
        const currentDatasetId = computed(() => playerStore.currentDatasetId);

        watch(
            () => props.players,
            (newPlayers, oldPlayers) => {
                console.log(
                    `PlayerDataTable: props.players changed. New length: ${newPlayers?.length}, Old length: ${oldPlayers?.length}`,
                );
                pagination.value.page = 1; // Reset to first page when player list changes
            },
            { deep: true },
        );

        const positionSortOrder = [
            "GK",
            "DR",
            "DC",
            "DL",
            "WBR",
            "WBL",
            "DM",
            "MR",
            "MC",
            "ML",
            "AMR",
            "AMC",
            "AML",
            "ST",
        ];

        // Memoized position index calculation (expensive string processing)
        const getPositionIndex = memoize((positionString) => {
            if (!positionString || typeof positionString !== "string") {
                return positionSortOrder.length + 2; // Place invalid/empty last
            }
            let processedString = positionString.toUpperCase();
            processedString = processedString.replace(/\bST\s*\(C\)/g, "ST");
            processedString = processedString.replace(/\bM\s*\(C\)/g, "MC");
            processedString = processedString.replace(/\bAM\s*\(C\)/g, "AMC");
            processedString = processedString.replace(/\bDM\s*\(C\)/g, "DM");
            processedString = processedString.replace(/\bD\s*\(C\)/g, "DC");
            processedString = processedString.replace(/\bGK\s*\(C\)/g, "GK");

            let minFoundIndex = positionSortOrder.length;
            const sideMatch = processedString.match(/\(([^)]+)\)$/);
            let mainPart = processedString;
            let sidesSpecified = [];

            if (sideMatch && sideMatch[1]) {
                mainPart = processedString.substring(0, sideMatch.index).trim();
                const sideSpec = sideMatch[1];
                if (sideSpec.includes("R")) sidesSpecified.push("R");
                if (sideSpec.includes("L")) sidesSpecified.push("L");
            }

            mainPart = mainPart.replace(/\s*\(.*?\)\s*/g, "").trim();
            const basePositionCodes = mainPart
                .split(/[,/]/)
                .map((p) => p.trim())
                .filter((p) => p.length > 0);
            const rolesToEvaluate = new Set();

            for (const baseCode of basePositionCodes) {
                if (sidesSpecified.length > 0) {
                    for (const side of sidesSpecified) {
                        rolesToEvaluate.add(baseCode + side);
                    }
                }
                rolesToEvaluate.add(baseCode);
            }

            if (rolesToEvaluate.size === 0 && positionString.trim() !== "") {
                rolesToEvaluate.add(
                    processedString.replace(/\s*\(.*?\)\s*/g, "").trim(),
                );
            }
            if (rolesToEvaluate.size === 0) return positionSortOrder.length + 1;

            for (const role of rolesToEvaluate) {
                const index = positionSortOrder.indexOf(role);
                if (index !== -1 && index < minFoundIndex) {
                    minFoundIndex = index;
                }
            }
            return minFoundIndex === positionSortOrder.length
                ? positionSortOrder.length + 1
                : minFoundIndex;
        }, {
            maxSize: 100, // Cache position calculations
            keyGenerator: (positionString) => positionString,
            cacheKey: 'positionIndex'
        });

        const onPaginationUpdate = (newPagination) => {
            console.log(
                `PlayerDataTable: onPaginationUpdate triggered. New pagination:`,
                JSON.parse(JSON.stringify(newPagination)),
            );
            pagination.value = newPagination;
        };

        // Column definitions with fixed widths to prevent layout shifts
        const nameColumnStyle =
            "width: 200px; min-width: 200px; max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const ageColumnStyle =
            "width: 60px; min-width: 60px; max-width: 60px; text-align: center; white-space: nowrap;";
        const positionColumnStyle =
            "width: 150px; min-width: 150px; max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const clubColumnStyle =
            "width: 180px; min-width: 180px; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const moneyColumnStyle =
            "width: 110px; min-width: 110px; max-width: 110px; text-align: right; white-space: nowrap;";
        const overallColumnStyle =
            "width: 70px; min-width: 70px; max-width: 70px; text-align: center; white-space: nowrap;";
        const fifaStatColumnStyle =
            "width: 60px; min-width: 60px; max-width: 60px; text-align: center; white-space: nowrap;";
        const textColumnStyle =
            "width: 120px; min-width: 120px; max-width: 120px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const nationalityColumnStyle =
            "width: 150px; min-width: 150px; max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";

        const baseColumnDefinitions = {
            name: {
                name: "name",
                label: "Name",
                field: "name",
                sortable: true,
                align: "left",
                style: nameColumnStyle,
                headerStyle: nameColumnStyle,
            },
            age: {
                name: "age",
                label: "Age",
                field: "age",
                sortable: true,
                align: "center",
                style: ageColumnStyle,
                headerStyle: ageColumnStyle,
            },
            position: {
                name: "position",
                label: "Position",
                field: "position",
                sortable: true,
                align: "left",
                style: positionColumnStyle,
                headerStyle: positionColumnStyle,
            },
            club: {
                name: "club",
                label: "Club",
                field: "club",
                sortable: true,
                align: "left",
                style: clubColumnStyle,
                headerStyle: clubColumnStyle,
            },
            transfer_value: {
                name: "transfer_value",
                label: "Value",
                field: "transfer_value",
                sortable: true,
                align: "right",
                sortField: "transferValueAmount",
                style: moneyColumnStyle,
                headerStyle: moneyColumnStyle,
            },
            wage: {
                name: "wage",
                label: "Salary",
                field: "wage",
                sortable: true,
                align: "right",
                sortField: "wageAmount",
                style: moneyColumnStyle,
                headerStyle: moneyColumnStyle,
            },
            Overall: {
                name: "Overall",
                label: "Overall",
                field: "Overall",
                sortable: true,
                align: "center",
                isOverallStat: true,
                style: overallColumnStyle,
                headerStyle: overallColumnStyle,
            },
            personality: {
                name: "personality",
                label: "Personality",
                field: "personality",
                sortable: true,
                align: "left",
                style: textColumnStyle,
                headerStyle: textColumnStyle,
            },
            media_handling: {
                name: "media_handling",
                label: "Media Desc.",
                field: "media_handling",
                sortable: true,
                align: "left",
                style: textColumnStyle,
                headerStyle: textColumnStyle,
            },
            nationality_display: {
                name: "nationality_display",
                label: "Nationality",
                field: "nationality",
                sortable: true,
                align: "left",
                style: nationalityColumnStyle,
                headerStyle: nationalityColumnStyle,
            },
        };

        const allFifaStatDefinitions = {
            GK: {
                name: "GK",
                label: "GK",
                field: "GK",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            DIV: {
                name: "DIV",
                label: "DIV",
                field: "DIV",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            HAN: {
                name: "HAN",
                label: "HAN",
                field: "HAN",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            REF: {
                name: "REF",
                label: "REF",
                field: "REF",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            KIC: {
                name: "KIC",
                label: "KIC",
                field: "KIC",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            SPD: {
                name: "SPD",
                label: "SPD",
                field: "SPD",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            POS: {
                name: "POS",
                label: "POS",
                field: "POS",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            PAC: {
                name: "PAC",
                label: "PAC",
                field: "PAC",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            SHO: {
                name: "SHO",
                label: "SHO",
                field: "SHO",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            PAS: {
                name: "PAS",
                label: "PAS",
                field: "PAS",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            DRI: {
                name: "DRI",
                label: "DRI",
                field: "DRI",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            DEF: {
                name: "DEF",
                label: "DEF",
                field: "DEF",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
            PHY: {
                name: "PHY",
                label: "PHY",
                field: "PHY",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: fifaStatColumnStyle,
                headerStyle: fifaStatColumnStyle,
            },
        };

        // Memoized computed for columns (recalculated when isGoalkeeperView changes)
        const currentColumns = memoizedComputed(() => {
            const newOrderBase = [
                baseColumnDefinitions.name,
                baseColumnDefinitions.nationality_display,
                baseColumnDefinitions.age,
                baseColumnDefinitions.position,
                baseColumnDefinitions.club,
                baseColumnDefinitions.transfer_value,
                baseColumnDefinitions.wage,
                baseColumnDefinitions.Overall,
            ];
            let fifaColumnsInOrder = props.isGoalkeeperView
                ? [
                      allFifaStatDefinitions.DIV,
                      allFifaStatDefinitions.HAN,
                      allFifaStatDefinitions.REF,
                      allFifaStatDefinitions.KIC,
                      allFifaStatDefinitions.SPD,
                      allFifaStatDefinitions.POS,
                  ]
                : [
                      allFifaStatDefinitions.PAC,
                      allFifaStatDefinitions.SHO,
                      allFifaStatDefinitions.PAS,
                      allFifaStatDefinitions.DRI,
                      allFifaStatDefinitions.DEF,
                      allFifaStatDefinitions.PHY,
                  ];
            const trailingColumns = [
                baseColumnDefinitions.personality,
                baseColumnDefinitions.media_handling,
            ];
            return [...newOrderBase, ...fifaColumnsInOrder, ...trailingColumns];
        }, {
            dependencies: [() => props.isGoalkeeperView]
        });

        const getColumnLabel = (fieldName) => {
            const col = currentColumns.value.find((c) => c.name === fieldName);
            return col ? col.label : fieldName;
        };

        const getSortFieldKey = (colName) => {
            const colDef = currentColumns.value.find((c) => c.name === colName);
            return colDef?.sortField || colDef?.field || colName;
        };

        // Cache for sorted results to avoid re-sorting on pagination changes
        const sortedPlayersCache = ref(null);
        const lastSortKey = ref('');

        // Separate reactive flag to prevent infinite loops
        const sortKeyChanged = ref('');
        const asyncSortTimeout = ref(null);

        const sortedPlayers = computed(() => {
            if (!props.players || props.players.length === 0) {
                totalSortedCount.value = 0;
                isSliced.value = false;
                sortedPlayersCache.value = null;
                return [];
            }

            const fieldKey = getSortFieldKey(sortField.value || "Overall");
            const direction = sortDirection.value;
            const currentSortKey = `${fieldKey}-${direction}-${props.players.length}-${cacheGeneration.value}`;

            // Return cached result if sort parameters haven't changed
            if (sortedPlayersCache.value && lastSortKey.value === currentSortKey) {
                return sortedPlayersCache.value;
            }

            // Custom sort function for complex sorting logic
            const customSortFn = (a, b, field, dir) => {
                // Use getPlayerValue to ensure GK stat mapping is applied for sorting
                let vA = getPlayerValue(a, field, sortField.value);
                let vB = getPlayerValue(b, field, sortField.value);
                const aIsNull = vA === null || vA === undefined;
                const bIsNull = vB === null || vB === undefined;

                if (aIsNull && bIsNull) return 0;
                if (aIsNull) return dir === "asc" ? 1 : -1;
                if (bIsNull) return dir === "asc" ? -1 : 1;

                if (field === "position") {
                    const indexA = getPositionIndex(vA);
                    const indexB = getPositionIndex(vB);
                    return dir === "asc" ? indexA - indexB : indexB - indexA;
                }
                if (typeof vA === "number" && typeof vB === "number") {
                    return dir === "asc" ? vA - vB : vB - vA;
                }
                if (typeof vA === "string" && typeof vB === "string") {
                    vA = vA.toLowerCase();
                    vB = vB.toLowerCase();
                    if (vA < vB) return dir === "asc" ? -1 : 1;
                    if (vA > vB) return dir === "asc" ? 1 : -1;
                    return 0;
                }
                return 0;
            };

            // For small arrays, use synchronous sorting
            if (props.players.length <= 500) {
                const playersToSort = [...props.players];
                const fullSortedList = playersToSort.sort((a, b) => customSortFn(a, b, fieldKey, direction));
                
                totalSortedCount.value = fullSortedList.length;
                let result;
                let sliced = false;
                
                if (fullSortedList.length > MAX_DISPLAY_PLAYERS) {
                    sliced = true;
                    result = fullSortedList.slice(0, MAX_DISPLAY_PLAYERS);
                } else {
                    result = fullSortedList;
                }
                
                isSliced.value = sliced;
                sortedPlayersCache.value = result;
                lastSortKey.value = currentSortKey;
                
                // Make sure async sorting flag is false for small datasets
                isAsyncSorting.value = false;
                
                return result;
            }

            // For large arrays, mark that we need async sorting but don't trigger it here
            if (sortKeyChanged.value !== currentSortKey) {
                sortKeyChanged.value = currentSortKey;
                
                // Clear any existing timeout
                if (asyncSortTimeout.value) {
                    clearTimeout(asyncSortTimeout.value);
                }
                
                // Debounce async sorting to prevent rapid successive calls
                asyncSortTimeout.value = setTimeout(() => {
                    if (!isAsyncSorting.value && sortKeyChanged.value === currentSortKey) {
                        triggerAsyncSort(fieldKey, direction, customSortFn, currentSortKey);
                    }
                }, 100); // 100ms debounce
            }

            // CRITICAL: Always return the current cache during sorting to prevent layout shifts
            // This ensures the table maintains exactly the same data structure and length
            if (sortedPlayersCache.value && sortedPlayersCache.value.length > 0) {
                return sortedPlayersCache.value;
            } else {
                // If no cache yet, create a stable fallback that matches the expected display size
                // This prevents any change in table height or structure during initial sort
                const targetLength = Math.min(MAX_DISPLAY_PLAYERS, props.players.length);
                const stableFallback = [...props.players].slice(0, targetLength);
                
                // Set initial cache to prevent further changes during sorting
                if (!sortedPlayersCache.value) {
                    sortedPlayersCache.value = stableFallback;
                    totalSortedCount.value = props.players.length;
                    isSliced.value = props.players.length > MAX_DISPLAY_PLAYERS;
                }
                
                return stableFallback;
            }
        });

        // Async sorting for large datasets using web workers
        const triggerAsyncSort = async (fieldKey, direction, customSortFn, sortKey) => {
            // Prevent concurrent async sorting operations
            if (isAsyncSorting.value) {
                console.log('Async sorting already in progress, skipping');
                return;
            }
            
            // Cancel any previous sort operation
            if (currentSortController.value) {
                currentSortController.value.cancelled = true;
            }
            
            // Create new controller for this sort operation
            const sortController = { cancelled: false };
            currentSortController.value = sortController;
            
            // Set async sorting flag AFTER ensuring cache stability
            isAsyncSorting.value = true;
            
            try {
                let fullSortedList;
                const playerCount = props.players.length;
                
                // Tiered sorting strategy based on dataset size
                if (playerCount >= 2000) {
                    // Use Web Workers for large datasets (2000+)
                    console.log(`Using Web Worker for large dataset: ${playerCount} players`);
                    try {
                        fullSortedList = await sortPlayersWorker(
                            [...props.players],
                            fieldKey,
                            direction,
                            sortField.value,
                            props.isGoalkeeperView
                        );
                    } catch (workerError) {
                        console.warn('Web Worker failed, falling back to main thread:', workerError);
                        // Fallback to main thread if Web Worker fails
                        fullSortedList = await sortLargeArray(
                            [...props.players], 
                            fieldKey, 
                            direction, 
                            customSortFn,
                            2000
                        );
                    }
                } else {
                    // Use optimized main thread sorting for medium datasets (500-2000)
                    console.log(`Using optimized main thread sorting for medium dataset: ${playerCount} players`);
                    fullSortedList = await sortLargeArray(
                        [...props.players], 
                        fieldKey, 
                        direction, 
                        customSortFn,
                        2000
                    );
                }

                // Check if this sort was cancelled
                if (sortController.cancelled) {
                    console.log('Async sort was cancelled');
                    return;
                }

                totalSortedCount.value = fullSortedList.length;
                let result;
                let sliced = false;
                
                if (fullSortedList.length > MAX_DISPLAY_PLAYERS) {
                    sliced = true;
                    result = fullSortedList.slice(0, MAX_DISPLAY_PLAYERS);
                } else {
                    result = fullSortedList;
                }
                
                isSliced.value = sliced;
                
                // Update cache in a way that doesn't trigger layout recalculation
                nextTick(() => {
                    sortedPlayersCache.value = result;
                    lastSortKey.value = sortKey;
                });
                
                console.log(`Async sort completed successfully for ${fullSortedList.length} players`);
            } catch (error) {
                if (sortController.cancelled) {
                    console.log('Async sort was cancelled during execution');
                    return;
                }
                
                console.error('Error during async sorting:', error.message || error);
                
                // Show user-friendly error notification
                $q.notify({
                    type: 'warning',
                    message: 'Sorting was interrupted. Showing unsorted data.',
                    position: 'top',
                    timeout: 3000,
                });
                
                // Fallback to direct assignment if async sorting fails
                const fallbackResult = [...props.players].slice(0, MAX_DISPLAY_PLAYERS);
                nextTick(() => {
                    sortedPlayersCache.value = fallbackResult;
                    totalSortedCount.value = props.players.length;
                    isSliced.value = props.players.length > MAX_DISPLAY_PLAYERS;
                    lastSortKey.value = sortKey;
                });
            } finally {
                // Always clear the sorting flag and controller, regardless of outcome
                isAsyncSorting.value = false;
                currentSortController.value = null;
            }
        };

        // Cancel current async sort operation
        const cancelAsyncSort = () => {
            if (currentSortController.value) {
                currentSortController.value.cancelled = true;
                console.log('User cancelled async sort operation');
            }
            isAsyncSorting.value = false;
            currentSortController.value = null;
            
            $q.notify({
                type: 'info',
                message: 'Sorting cancelled',
                position: 'top',
                timeout: 2000,
            });
        };

        const pagesNumber = computed(() => {
            if (
                !sortedPlayers.value ||
                sortedPlayers.value.length === 0 ||
                pagination.value.rowsPerPage === 0
            ) {
                return 1;
            }
            return Math.ceil(
                sortedPlayers.value.length / pagination.value.rowsPerPage,
            );
        });

        const paginationTotalRows = computed(() => sortedPlayers.value.length);
        const paginationStartRow = computed(() => {
            if (paginationTotalRows.value === 0) return 0;
            return (
                (pagination.value.page - 1) * pagination.value.rowsPerPage + 1
            );
        });
        const paginationEndRow = computed(() => {
            if (paginationTotalRows.value === 0) return 0;
            if (pagination.value.rowsPerPage === 0)
                return paginationTotalRows.value;
            return Math.min(
                pagination.value.page * pagination.value.rowsPerPage,
                paginationTotalRows.value,
            );
        });

        onMounted(() => {
            console.log(`PlayerDataTable: Component mounted.`);
            if (sortField.value) {
                emit("update:sort", {
                    key: getSortFieldKey(sortField.value),
                    direction: sortDirection.value,
                    isFifaStat: currentColumns.value.find(
                        (c) => c.name === sortField.value,
                    )?.isFifaStat,
                    isOverallStat: currentColumns.value.find(
                        (c) => c.name === sortField.value,
                    )?.isOverallStat,
                    displayField: sortField.value,
                });
            }
        });

        // Memoized rating class calculation (called frequently in table rendering)
        const getUnifiedRatingClass = memoize((value, maxScale) => {
            const numValue = parseInt(value, 10);
            if (
                isNaN(numValue) ||
                value === null ||
                value === undefined ||
                value === "-"
            )
                return "rating-na";
            const percentage = (numValue / maxScale) * 100;
            if (percentage >= 90) return "rating-tier-6";
            if (percentage >= 80) return "rating-tier-5";
            if (percentage >= 70) return "rating-tier-4";
            if (percentage >= 55) return "rating-tier-3";
            if (percentage >= 40) return "rating-tier-2";
            return "rating-tier-1";
        }, {
            maxSize: 200, // Cache up to 200 different rating calculations
            keyGenerator: (value, maxScale) => `${value}-${maxScale}`,
            cacheKey: 'unifiedRatingClass'
        });

        const getMoneyClass = (numericAmount) => {
            if (numericAmount === null || numericAmount === undefined)
                return "money-na";
            if (numericAmount >= 10000000) return "money-very-high";
            if (numericAmount >= 1000000) return "money-high";
            if (numericAmount >= 100000) return "money-medium-high";
            if (numericAmount >= 10000) return "money-medium";
            if (numericAmount > 0) return "money-low";
            return "money-na";
        };

        const onFlagError = (event) => {
            if (event.target) event.target.style.display = "none";
            const placeholderIcon = event.target.nextElementSibling;
            if (
                placeholderIcon &&
                placeholderIcon.classList.contains("q-icon")
            ) {
                placeholderIcon.style.display = "inline-flex";
            }
        };

        const onRequest = (requestProp) => {
            console.log(
                `PlayerDataTable: onRequest triggered. Props:`,
                JSON.parse(JSON.stringify(requestProp)),
            );
            const { page, sortBy, descending } = requestProp.pagination;
            pagination.value.page = page;

            if (
                sortBy &&
                (sortField.value !== sortBy ||
                    sortDirection.value !== (descending ? "desc" : "asc"))
            ) {
                sortField.value = sortBy;
                sortDirection.value = descending ? "desc" : "asc";
                pagination.value.sortBy = sortBy;
                pagination.value.descending = descending;

                emit("update:sort", {
                    key: getSortFieldKey(sortField.value),
                    direction: sortDirection.value,
                    isFifaStat: currentColumns.value.find(
                        (c) => c.name === sortBy,
                    )?.isFifaStat,
                    isOverallStat: currentColumns.value.find(
                        (c) => c.name === sortBy,
                    )?.isOverallStat,
                    displayField: sortBy,
                });
            }
            emit("update:pagination", { ...pagination.value });
        };

        const onPageChange = (newPage) => {
            console.log(`PlayerDataTable: onPageChange. New page: ${newPage}`);
            pagination.value.page = newPage;
        };

        const onRowsPerPageChange = (newRowsPerPage) => {
            console.log(
                `PlayerDataTable: onRowsPerPageChange. New rowsPerPage: ${newRowsPerPage}`,
            );
            pagination.value.rowsPerPage = newRowsPerPage;
            pagination.value.page = 1;
        };

        const customSort = (rows) => {
            // The actual sorting is now done in the `sortedPlayers` computed property.
            // QTable's `sort-method` is still needed, but we just return the rows as they are
            // because our computed property `sortedPlayers` (bound to :rows) already handles it.
            return rows;
        };

        const sortTable = (fieldName) => {
            console.time("PlayerDataTable: sortTable_execution");
            
            // Prevent rapid clicking during sort operations
            if (isAsyncSorting.value) {
                console.log('Sort already in progress, ignoring click');
                return;
            }
            
            // Clear any pending async sorting
            if (asyncSortTimeout.value) {
                clearTimeout(asyncSortTimeout.value);
                asyncSortTimeout.value = null;
            }
            
            // Clear sort cache when sort parameters change
            clearSortCache();
            // Don't clear sortedPlayersCache immediately to prevent layout shift
            // It will be updated when the new sort completes
            lastSortKey.value = '';
            sortKeyChanged.value = '';
            
            const actualSortKey = getSortFieldKey(fieldName);
            let newDirection;
            if (sortField.value === fieldName) {
                newDirection = sortDirection.value === "asc" ? "desc" : "asc";
            } else {
                const colDef = currentColumns.value.find(
                    (c) => c.name === fieldName,
                );
                if (colDef && (colDef.isOverallStat || colDef.isFifaStat)) {
                    newDirection = "desc";
                } else {
                    newDirection = "asc";
                }
            }
            sortField.value = fieldName;
            sortDirection.value = newDirection;

            pagination.value.sortBy = fieldName;
            pagination.value.descending = newDirection === "desc";
            pagination.value.page = 1;

            emit("update:sort", {
                key: actualSortKey,
                direction: newDirection,
                isFifaStat: currentColumns.value.find(
                    (c) => c.name === fieldName,
                )?.isFifaStat,
                isOverallStat: currentColumns.value.find(
                    (c) => c.name === fieldName,
                )?.isOverallStat,
                displayField: fieldName,
            });
            console.timeEnd("PlayerDataTable: sortTable_execution");
        };

        const onRowClick = (player) => {
            emit("player-selected", player);
        };

        const onClubClick = (player) => {
            console.log('onClubClick called with player:', player);
            console.log('Club name:', player.club);
            if (player.club && player.club.trim() !== '') {
                console.log('Emitting team-selected event with club:', player.club);
                emit("team-selected", player.club);
            } else {
                console.log('Club name is empty or invalid');
            }
        };

        const formatDisplayCurrency = (numericAmount, originalDisplayValue) => {
            return formatCurrency(
                numericAmount,
                props.currencySymbol,
                originalDisplayValue,
            );
        };

        // GK stat mapping for both display and sorting consistency
        const gkStatMapping = {
            'PAC': 'DIV',  // Diving -> Pace
            'SHO': 'HAN',  // Handling -> Shooting  
            'PAS': 'KIC',  // Kicking -> Passing
            'DRI': 'REF',  // Reflexes -> Dribbling
            'DEF': 'SPD',  // Speed -> Defending
            'PHY': 'POS'   // Positioning -> Physical
        };

        // Memoized player value getter (called frequently during sorting and rendering)
        const getPlayerValue = (player, fieldKey, columnName = null) => {
            // For non-goalkeeper view, map GK stats to standard FIFA stats if the player is a goalkeeper
            if (!props.isGoalkeeperView && player.position && player.position.includes('GK')) {
                const mappedStat = gkStatMapping[columnName || fieldKey];
                if (mappedStat && player[mappedStat] !== undefined) {
                    return player[mappedStat];
                }
            }
            
            // Default behavior - use the field key
            const value = player[fieldKey];
            
            return value;
        };

        // Memoized version for non-Overall fields only
        const getPlayerValueMemoized = memoize((player, fieldKey, columnName = null) => {
            return getPlayerValue(player, fieldKey, columnName);
        }, {
            maxSize: 1000,
            keyGenerator: (player, fieldKey, columnName) => {
                // Try to use the player's UID for cache key
                let playerUID = player.UID || player.uid;
                
                // Enhanced debugging to understand UID field issues
                if (!playerUID) {
                    // Check if this is one of the first few players to avoid spam
                    const isFirstFewPlayers = props.players.indexOf(player) < 3;
                    if (isFirstFewPlayers) {
                        console.log('Player missing UID:', player.name);
                        console.log('Available fields:', Object.keys(player));
                        console.log('UID field check:', {
                            'player.UID': player.UID,
                            'player.uid': player.uid,
                            'player.Uid': player.Uid,
                            'player.id': player.id,
                            'player.ID': player.ID,
                            'player.playerId': player.playerId,
                            'player.player_id': player.player_id
                        });
                        
                        // Check if there's any field that looks like an ID
                        const possibleIdFields = Object.keys(player).filter(key => 
                            key.toLowerCase().includes('id') || 
                            key.toLowerCase().includes('uid') ||
                            key.toLowerCase() === 'unique'
                        );
                        console.log('Possible ID fields found:', possibleIdFields);
                        
                        if (possibleIdFields.length > 0) {
                            possibleIdFields.forEach(field => {
                                console.log(`${field}:`, player[field]);
                            });
                        }
                    }
                }
                
                // If no UID available or UID is empty, create a composite unique key
                if (!playerUID || playerUID === '') {
                    playerUID = `${player.name || 'unknown'}-${player.club || 'unknown'}-${player.age || 'unknown'}-${player.position || 'unknown'}`;
                }
                
                return `gen${cacheGeneration.value}-${playerUID}-${fieldKey}-${columnName || ''}`;
            },
            cacheKey: 'playerValue'
        });

        const getDisplayValue = (player, col) => {
            // For Overall field, always use non-memoized version to ensure reactivity
            if (col.field === 'Overall' || col.name === 'Overall') {
                return getPlayerValue(player, col.field, col.name);
            }
            // For other fields, use memoized version for performance
            return getPlayerValueMemoized(player, col.field, col.name);
        };

        const contextMenuPlayer = ref(null);
        
        const isPlayerInWishlist = (player) => {
            if (!player || !currentDatasetId.value) return false;
            return wishlistStore.isInWishlist(currentDatasetId.value, player);
        };

        const handleAddToWishlist = async () => {
            if (contextMenuPlayer.value && currentDatasetId.value) {
                const success = await wishlistStore.addToWishlist(currentDatasetId.value, contextMenuPlayer.value);
                if (success) {
                    $q.notify({
                        type: 'positive',
                        message: `${contextMenuPlayer.value.name} added to wishlist`,
                        position: 'top',
                        timeout: 2000,
                    });
                } else {
                    $q.notify({
                        type: 'warning',
                        message: `${contextMenuPlayer.value.name} is already in wishlist`,
                        position: 'top',
                        timeout: 2000,
                    });
                }
            }
        };

        const handleRemoveFromWishlist = async () => {
            if (contextMenuPlayer.value && currentDatasetId.value) {
                const success = await wishlistStore.removeFromWishlist(currentDatasetId.value, contextMenuPlayer.value);
                if (success) {
                    $q.notify({
                        type: 'positive',
                        message: `${contextMenuPlayer.value.name} removed from wishlist`,
                        position: 'top',
                        timeout: 2000,
                    });
                    if (props.showWishlistActions) {
                        emit("remove-from-wishlist", contextMenuPlayer.value);
                    }
                }
            }
        };

        const handlePlayerDetails = () => {
            if (contextMenuPlayer.value) {
                emit("player-selected", contextMenuPlayer.value);
            }
        };

        const onRightClick = (event, player) => {
            event.preventDefault();
            contextMenuPlayer.value = player;
        };

        // Watch for changes in players prop and increment cache generation to force cache invalidation
        watch(() => props.players, () => {
            cacheGeneration.value++;
            // Clear sorted players cache to force recalculation
            sortedPlayersCache.value = null;
            lastSortKey.value = '';
        }, { deep: false }); // shallow watch to detect reference changes

        // Clear memoization caches when component unmounts or when players prop changes significantly
        watch(() => props.players?.length, () => {
            // Clear all memoization caches when dataset changes
            getUnifiedRatingClass.clearCache();
            getPlayerValueMemoized.clearCache();
            getPositionIndex.clearCache();
        });

        watch(() => props.isGoalkeeperView, () => {
            // Clear player value cache when view mode changes and increment generation
            cacheGeneration.value++;
            getPlayerValueMemoized.clearCache();
        });

        onUnmounted(() => {
            // Clear all caches on component cleanup
            getUnifiedRatingClass.clearCache();
            getPlayerValueMemoized.clearCache();
            getPositionIndex.clearCache();
            
            // Clear async sort timeout
            if (asyncSortTimeout.value) {
                clearTimeout(asyncSortTimeout.value);
            }
        });

        console.log(`PlayerDataTable: Setup function end.`);
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
        };
    },
};
</script>

<style lang="scss" scoped>
.player-data-table-container {
    background: white;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
    /* Ensure absolutely stable dimensions */
    width: 100%;
    min-width: 100%;
    max-width: 100%;
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.02);
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
    }
}

.table-controls-header {
    padding: 0 4px 8px 4px;
    align-items: center;
}

.sort-info-chip {
    display: inline-block;
    margin-right: 16px;
    padding: 6px 10px;
    font-size: 0.8rem;
}

.player-q-table {
    width: 100%;
    table-layout: fixed; /* CHANGED: from auto to fixed */
    
    /* Disable layout-affecting transitions to prevent shifts during sorting */
    th, td, tr {
        transition: none !important;
    }
    
    /* Preserve necessary loading animations */
    .q-spinner, .sorting-spinner, .q-linear-progress {
        transition: unset !important;
        animation-duration: unset !important;
    }
    
    /* Enforce strict column width control */
    th, td {
        white-space: nowrap !important;
        word-break: keep-all !important;
        /* Prevent any width changes */
        flex-shrink: 0 !important;
        flex-grow: 0 !important;
    }
    
    /* Apply ellipsis only to specific long-text columns by targeting their style attribute */
    th[style*="text-overflow: ellipsis"], 
    td[style*="text-overflow: ellipsis"] {
        overflow: hidden !important;
        text-overflow: ellipsis !important;
    }
    
    /* Override any Quasar default behaviors that might affect width */
    .q-table__container .q-table {
        table-layout: fixed !important;
    }
    
    &.table-sorting {
        /* Prevent any layout shifts during sorting */
        overflow: visible; /* Changed from hidden to prevent layout changes */
        
        /* Minimal visual feedback without affecting layout */
        .modern-header-cell.sorting-in-progress {
            position: relative;
            /* Remove opacity changes that can affect layout */
        }
        
        /* Keep the body stable during sorting */
        .q-table__middle {
            overflow: visible;
        }
    }

    th .sort-icon {
        vertical-align: middle;
        margin-left: 4px;
        /* Ensure icon doesn't affect layout */
        position: absolute;
        right: 8px;
        top: 50%;
        transform: translateY(-50%);
    }

    .sorting-spinner {
        animation: gentlePulse 1.5s infinite ease-in-out;
        /* Position spinner absolutely to prevent layout shifts */
        position: absolute !important;
        right: 8px;
        top: 50%;
        transform: translateY(-50%);
        margin: 0 !important;
    }

    // Dark mode specific styles
    &.q-table--dark {
        th {
            color: $grey-3; // Make sure $grey-3 is defined or use a Quasar variable
            border-bottom-color: rgba(255, 255, 255, 0.15);
        }

        td {
            border-bottom-color: rgba(255, 255, 255, 0.1);
            color: $grey-4; // Make sure $grey-4 is defined or use a Quasar variable
        }

        tr:last-child td {
            border-bottom: 0;
        }

        .q-table__bottom {
            border-top-color: rgba(255, 255, 255, 0.15);
        }
    }

    // Light mode specific styles (default, but explicitly stated for clarity)
    &:not(.q-table--dark) {
        th {
            background-color: #f0f4f8;
            color: $grey-8; // Make sure $grey-8 is defined or use a Quasar variable
            border-bottom: 1px solid #dde2e6;
        }

        td {
            border-bottom: 1px solid #eef2f5;
            color: $grey-9; // Make sure $grey-9 is defined or use a Quasar variable
        }

        tr:last-child td {
            border-bottom: 0;
        }

        .q-table__bottom {
            border-top: 1px solid #dde2e6;
        }
    }

    th {
        font-weight: 600;
        font-size: 0.8rem;
        padding: 8px 32px 8px 10px; /* Add right padding for sort icon */
        border-right: 0;
        /* Remove transition to prevent layout shifts during sorting */
        /* white-space: nowrap; // This is now handled by inline styles for specific columns */
        /* overflow: hidden; // This is now handled by inline styles for specific columns */
        /* text-overflow: ellipsis; // This is now handled by inline styles for specific columns */
        
        /* Ensure consistent height and width during sorting */
        min-height: 40px;
        box-sizing: border-box;
        position: relative; /* For absolutely positioned sort icons */
        
        &.sorting-in-progress {
            position: relative;
            background: linear-gradient(90deg, 
                transparent 0%, 
                rgba(25, 118, 210, 0.03) 50%, 
                transparent 100%);
            animation: subtleGlow 2s infinite ease-in-out;
            
            .body--dark & {
                background: linear-gradient(90deg, 
                    transparent 0%, 
                    rgba(255, 255, 255, 0.02) 50%, 
                    transparent 100%);
            }
        }
        
        &.active-sort {
            background-color: rgba(25, 118, 210, 0.05);
            
            .body--dark & {
                background-color: rgba(255, 255, 255, 0.05);
            }
        }
    }

    td {
        vertical-align: middle;
        padding: 6px 10px;
        border-right: 0;
        /* white-space: nowrap; // This is now handled by inline styles for specific columns */
        /* overflow: hidden; // This is now handled by inline styles for specific columns */
        /* text-overflow: ellipsis; // This is now handled by inline styles for specific columns */
        
        /* Ensure consistent height during sorting */
        min-height: 32px;
        box-sizing: border-box;
    }

    .table-cell-enhanced {
        font-size: 0.85rem;
    }

    // Subtle row shimmer during sorting
    .row-shimmer {
        position: relative;
        
        &::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, 
                transparent 0%, 
                rgba(25, 118, 210, 0.02) 50%, 
                transparent 100%);
            animation: shimmer 3s infinite ease-in-out;
            pointer-events: none;
            
            .body--dark & {
                background: linear-gradient(90deg, 
                    transparent 0%, 
                    rgba(255, 255, 255, 0.015) 50%, 
                    transparent 100%);
            }
        }
    }
}

/*
  This rule was likely causing the issue with text-overflow: ellipsis.
  The inline styles for columns like 'name', 'position', and 'club' set 'white-space: nowrap',
  which is necessary for ellipsis. This global rule would override it.
  If you need word wrapping for other columns, apply 'white-space: normal' and 'word-break: break-word'
  more selectively, perhaps through a class on those specific column definitions.
*/
.player-q-table td,
.player-q-table th {
    /* white-space: normal; */ /* CHANGED: Commented out to allow inline styles to take effect for ellipsis */
    word-break: break-word; /* Keep this if you want long words without spaces to break */
}

.table-row-hover {
    &:hover {
        .body--dark & {
            // Ensure .body--dark is a class on your body/html tag when dark mode is active
            background-color: rgba(255, 255, 255, 0.08) !important;
        }

        .body--light & {
            // Ensure .body--light is a class on your body/html tag when light mode is active
            background-color: #e3f2fd !important;
        }
    }
}

// Money value styling
.money-value {
    display: inline-block;
    font-weight: 500;
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 0.8rem;
}

// Specific color classes for money values (can be expanded)
.money-very-high {
    color: #1b5e20;
    font-weight: 700;
}
.money-high {
    color: #2e7d32;
}
.money-medium-high {
    color: #4caf50;
}
.money-medium {
    color: #757575;
}
.money-low {
    color: #9e9e9e;
}
.money-na {
    color: #bdbdbd;
}

.body--dark {
    // Ensure .body--dark is a class on your body/html tag
    .money-very-high {
        color: #a5d6a7;
        font-weight: 700;
    }
    .money-high {
        color: #81c784;
    }
    .money-medium-high {
        color: #66bb6a;
    }
    .money-medium {
        color: #b0bec5;
    }
    .money-low {
        color: #90a4ae;
    }
    .money-na {
        color: #78909c;
    }
}

.nationality-flag {
    border: 1px solid rgba(0, 0, 0, 0.15);
    object-fit: cover;
    margin-right: 8px;
    width: 20px !important;
    height: 13px !important;
    flex-shrink: 0;

    .body--dark & {
        // Ensure .body--dark is a class on your body/html tag
        border: 1px solid rgba(255, 255, 255, 0.15);
    }
}

.nationality-flag-placeholder {
    margin-right: 8px;
    width: 20px;
    height: 13px;
    flex-shrink: 0;
}

.nationality-cell {
    width: 100%;
    overflow: hidden;
    
    .nationality-text {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        flex: 1;
        min-width: 0; /* Allow text to shrink */
    }
}

.flex.items-center .q-icon,
.flex.items-center img {
    flex-shrink: 0;
}

// Ensure select in pagination is styled for dark mode (though selector is removed, :deep might affect other selects if not scoped)
:deep(.q-table__bottom .q-select .q-field__native),
:deep(.q-table__bottom .q-select .q-field__input) {
    .body--dark & {
        // Ensure .body--dark is a class on your body/html tag
        color: $grey-3; // Make sure $grey-3 is defined or use a Quasar variable
    }
}

.club-link {
    color: inherit;
    text-decoration: none;
    cursor: pointer;
    
    &:hover {
        text-decoration: underline;
    }
}

// Modern Table Enhancements
.modern-table-header {
    background: linear-gradient(180deg, rgba(240, 244, 248, 0.8) 0%, rgba(240, 244, 248, 1) 100%);
    
    .body--dark & {
        background: linear-gradient(180deg, rgba(66, 66, 66, 0.8) 0%, rgba(66, 66, 66, 1) 100%);
    }
}

.modern-header-cell {
    &:hover {
        background: rgba(25, 118, 210, 0.08);
        
        .body--dark & {
            background: rgba(144, 202, 249, 0.08);
        }
    }
    
    &.active-sort {
        background: rgba(25, 118, 210, 0.12);
        color: #1976d2;
        
        .body--dark & {
            background: rgba(144, 202, 249, 0.12);
            color: #90caf9;
        }
    }
}

.modern-table-row {
    /* Remove transition to prevent layout shifts */
    
    &:hover {
        /* Remove transform to prevent layout shifts */
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            box-shadow: 0 2px 8px rgba(255, 255, 255, 0.1);
        }
    }
}

.modern-stat-badge {
    padding: 3px 8px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    text-align: center;
    min-width: 32px;
    display: inline-block;
    /* Remove transition to prevent layout shifts */
    
    &:hover {
        /* Remove transform to prevent layout shifts */
    }
}

// Enhanced sort info chip
.sort-info-chip {
    background: linear-gradient(135deg, rgba(25, 118, 210, 0.1) 0%, rgba(25, 118, 210, 0.05) 100%);
    border: 1px solid rgba(25, 118, 210, 0.2);
    color: #1976d2;
    
    .body--dark & {
        background: linear-gradient(135deg, rgba(144, 202, 249, 0.1) 0%, rgba(144, 202, 249, 0.05) 100%);
        border: 1px solid rgba(144, 202, 249, 0.2);
        color: #90caf9;
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
        transform: scale(1.05);
        opacity: 0.8;
    }
}

@keyframes subtleGlow {
    0%, 100% {
        opacity: 1;
    }
    50% {
        opacity: 0.7;
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
</style>
