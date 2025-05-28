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

        <q-table
            v-else
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
            :class="qInstance.dark.isActive ? 'q-table--dark' : ''"
            table-header-class="player-table-header"
            dense
            virtual-scroll
            :virtual-scroll-item-size="30"
            :virtual-scroll-sticky-size-start="32"
            :virtual-scroll-sticky-size-end="55"
            style="height: 70vh"
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
                        class="cursor-pointer text-weight-bold modern-header-cell"
                        @click="sortTable(col.name)"
                        :class="[
                            qInstance.dark.isActive
                                ? 'text-grey-3'
                                : 'text-grey-8',
                            col.headerClasses,
                            { 'active-sort': sortField === col.name }
                        ]"
                        :style="col.headerStyle"
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
                            v-if="sortField === col.name"
                            :name="
                                sortDirection === 'asc'
                                    ? 'arrow_upward'
                                    : 'arrow_downward'
                            "
                            size="xs"
                            class="q-ml-xs sort-icon"
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
                            <div class="flex items-center no-wrap">
                                <img
                                    v-if="props.row.nationality_iso"
                                    :src="`https://flagcdn.com/w20/${props.row.nationality_iso.toLowerCase()}.png`"
                                    :alt="props.row.nationality || 'Flag'"
                                    width="20"
                                    height="13"
                                    class="q-mr-xs nationality-flag"
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
                                    class="q-mr-xs"
                                />
                                <span>{{ props.row.nationality || "-" }}</span>
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
import { ref, computed, watch, onMounted } from "vue";
import { useQuasar } from "quasar";
import { usePlayerStore } from "../stores/playerStore";
import { useWishlistStore } from "../stores/wishlistStore";
import { formatCurrency } from "../utils/currencyUtils";

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
        
        
        const contextMenu = ref(null);
        const sortField = ref("Overall");
        const sortDirection = ref("desc");
        const rowsPerPageOptions = ref([10, 15, 20, 50, 0]); // Keep for internal logic, but selector is removed
        const maxPagesToShow = 7;
        const totalSortedCount = ref(0);
        const isSliced = ref(false);

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

        const getPositionIndex = (positionString) => {
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
        };

        const onPaginationUpdate = (newPagination) => {
            console.log(
                `PlayerDataTable: onPaginationUpdate triggered. New pagination:`,
                JSON.parse(JSON.stringify(newPagination)),
            );
            pagination.value = newPagination;
        };

        // Column definitions
        const nameColumnStyle =
            "max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const ageColumnStyle =
            "max-width: 60px; text-align: center; white-space: nowrap;";
        const positionColumnStyle =
            "max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const clubColumnStyle =
            "max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const moneyColumnStyle =
            "max-width: 110px; text-align: right; white-space: nowrap;";
        const overallColumnStyle =
            "max-width: 70px; text-align: center; white-space: nowrap;";
        const fifaStatColumnStyle =
            "max-width: 60px; text-align: center; white-space: nowrap;";
        const textColumnStyle =
            "max-width: 120px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const nationalityColumnStyle =
            "max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";

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

        const currentColumns = computed(() => {
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
        });

        const getColumnLabel = (fieldName) => {
            const col = currentColumns.value.find((c) => c.name === fieldName);
            return col ? col.label : fieldName;
        };

        const getSortFieldKey = (colName) => {
            const colDef = currentColumns.value.find((c) => c.name === colName);
            return colDef?.sortField || colDef?.field || colName;
        };

        const sortedPlayers = computed(() => {
            if (!props.players || props.players.length === 0) {
                totalSortedCount.value = 0;
                isSliced.value = false;
                return [];
            }

            const fieldKey = getSortFieldKey(sortField.value || "Overall");
            const direction = sortDirection.value;

            // Sort the players array - create a copy to avoid mutations
            const playersToSort = [...props.players];
            const fullSortedList = playersToSort.sort((a, b) => {
                // Use getPlayerValue to ensure GK stat mapping is applied for sorting
                let vA = getPlayerValue(a, fieldKey, sortField.value);
                let vB = getPlayerValue(b, fieldKey, sortField.value);
                const aIsNull = vA === null || vA === undefined;
                const bIsNull = vB === null || vB === undefined;

                if (aIsNull && bIsNull) return 0;
                if (aIsNull) return direction === "asc" ? 1 : -1;
                if (bIsNull) return direction === "asc" ? -1 : 1;

                if (fieldKey === "position") {
                    const indexA = getPositionIndex(vA);
                    const indexB = getPositionIndex(vB);
                    return direction === "asc" ? indexA - indexB : indexB - indexA;
                }
                if (typeof vA === "number" && typeof vB === "number") {
                    return direction === "asc" ? vA - vB : vB - vA;
                }
                if (typeof vA === "string" && typeof vB === "string") {
                    vA = vA.toLowerCase();
                    vB = vB.toLowerCase();
                    if (vA < vB) return direction === "asc" ? -1 : 1;
                    if (vA > vB) return direction === "asc" ? 1 : -1;
                    return 0;
                }
                return 0;
            });

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
            return result;
        });

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

        const getUnifiedRatingClass = (value, maxScale) => {
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
        };

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

        // Get the actual value for sorting or display, with GK mapping applied
        const getPlayerValue = (player, fieldKey, columnName = null) => {
            // For non-goalkeeper view, map GK stats to standard FIFA stats if the player is a goalkeeper
            if (!props.isGoalkeeperView && player.position && player.position.includes('GK')) {
                const mappedStat = gkStatMapping[columnName || fieldKey];
                if (mappedStat && player[mappedStat] !== undefined) {
                    return player[mappedStat];
                }
            }
            
            // Default behavior - use the field key
            return player[fieldKey];
        };

        const getDisplayValue = (player, col) => {
            return getPlayerValue(player, col.field, col.name);
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

        console.log(`PlayerDataTable: Setup function end.`);
        return {
            qInstance: $q,
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
    width: 100%;
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

    th .sort-icon {
        vertical-align: middle;
        margin-left: 4px;
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
        padding: 8px 10px;
        border-right: 0;
        transition: all 0.2s ease;
        /* white-space: nowrap; // This is now handled by inline styles for specific columns */
        /* overflow: hidden; // This is now handled by inline styles for specific columns */
        /* text-overflow: ellipsis; // This is now handled by inline styles for specific columns */
    }

    td {
        vertical-align: middle;
        padding: 6px 10px;
        border-right: 0;
        /* white-space: nowrap; // This is now handled by inline styles for specific columns */
        /* overflow: hidden; // This is now handled by inline styles for specific columns */
        /* text-overflow: ellipsis; // This is now handled by inline styles for specific columns */
    }

    .table-cell-enhanced {
        font-size: 0.85rem;
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

    .body--dark & {
        // Ensure .body--dark is a class on your body/html tag
        border: 1px solid rgba(255, 255, 255, 0.15);
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
    transition: all 0.2s ease;
    
    &:hover {
        transform: scale(1.001);
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
    transition: transform 0.2s ease;
    
    &:hover {
        transform: scale(1.05);
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

// Enhanced table container
.player-data-table-container {
    background: white;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.02);
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
    }
}
</style>
