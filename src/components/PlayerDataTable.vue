<template>
    <div class="player-data-table">
        <div v-if="sortField" class="text-caption q-mb-sm q-pa-xs bg-grey-2">
            Current Sort: {{ sortField }} ({{ sortDirection }})
        </div>

        <q-card v-if="players.length === 0" class="q-pa-md text-center">
            <p class="text-grey-7">
                {{
                    loading
                        ? "Loading player data..."
                        : "No players match your search criteria."
                }}
            </p>
        </q-card>

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
                <template v-slot:header="props">
                    <q-tr :props="props">
                        <q-th
                            v-for="col in props.cols"
                            :key="col.name"
                            :props="props"
                            class="cursor-pointer"
                            @click="
                                sortTable(
                                    col.name,
                                    col.isFifaStat || col.isAttribute,
                                )
                            "
                        >
                            {{ col.label }}
                            <q-icon
                                v-if="sortField === col.name"
                                :name="
                                    sortDirection === 'asc'
                                        ? 'arrow_upward'
                                        : 'arrow_downward'
                                "
                                size="xs"
                                class="q-ml-xs"
                            />
                        </q-th>
                    </q-tr>
                </template>

                <template v-slot:body="props">
                    <q-tr
                        :props="props"
                        @click="onRowClick(props.row)"
                        class="cursor-pointer table-row-hover"
                    >
                        <q-td
                            v-for="col in props.cols"
                            :key="col.name"
                            :props="props"
                        >
                            <template v-if="col.isFifaStat">
                                <span
                                    :class="
                                        getFifaStatClass(props.row[col.field])
                                    "
                                    class="attribute-value fifa-stat-value"
                                >
                                    {{
                                        props.row[col.field] !== undefined
                                            ? props.row[col.field]
                                            : "-"
                                    }}
                                </span>
                            </template>
                            <template v-else-if="col.name === 'transfer_value'">
                                <span
                                    :class="getMoneyClass(props.row[col.field])"
                                    class="money-value"
                                >
                                    {{ props.row[col.field] || "-" }}
                                </span>
                            </template>
                            <template v-else-if="col.name === 'wage'">
                                <span
                                    :class="getMoneyClass(props.row[col.field])"
                                    class="money-value"
                                >
                                    {{ props.row[col.field] || "-" }}
                                </span>
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
                        :option-label="
                            (opt) => (opt === 0 ? 'All' : opt.toString())
                        "
                        borderless
                    />
                    <span class="q-ml-md text-caption">
                        {{
                            pagination.rowsPerPage === 0
                                ? 1
                                : (pagination.page - 1) *
                                      pagination.rowsPerPage +
                                  1
                        }}
                        -
                        {{
                            pagination.rowsPerPage === 0
                                ? players.length
                                : Math.min(
                                      pagination.page * pagination.rowsPerPage,
                                      players.length,
                                  )
                        }}
                        of {{ players.length }}
                    </span>
                </template>
            </q-table>
        </div>
    </div>
</template>

<script>
import { ref, computed, reactive, watch } from "vue";

export default {
    name: "PlayerDataTable",
    props: {
        players: {
            type: Array,
            required: true,
        },
        loading: {
            type: Boolean,
            default: false,
        },
    },
    emits: ["update:sort", "player-selected"], // Added 'player-selected'

    setup(props, { emit }) {
        const sortField = ref(null);
        const sortDirection = ref("asc");
        const rowsPerPageOptions = [10, 15, 20, 50, 0]; // 0 for 'All'
        const maxPagesToShow = 7;

        const pagination = reactive({
            sortBy: null,
            descending: false,
            page: 1,
            rowsPerPage: 15,
        });

        const pagesNumber = computed(() => {
            if (pagination.rowsPerPage === 0 || props.players.length === 0)
                return 1;
            return Math.ceil(props.players.length / pagination.rowsPerPage);
        });

        watch(
            () => props.players.length,
            () => {
                pagination.page = 1; // Reset to first page when player list changes
            },
        );

        watch(
            () => pagination.rowsPerPage,
            () => {
                // Adjust page number if it becomes invalid after changing rowsPerPage
                if (pagination.page > pagesNumber.value) {
                    pagination.page =
                        pagesNumber.value > 0 ? pagesNumber.value : 1;
                }
            },
        );

        // Column Definitions
        const baseColumns = [
            {
                name: "name",
                label: "Name",
                field: "name",
                sortable: true,
                align: "left",
            },
            {
                name: "position",
                label: "Position",
                field: "position",
                sortable: true,
                align: "left",
            }, // This is the original position string
            {
                name: "club",
                label: "Club",
                field: "club",
                sortable: true,
                align: "left",
            },
            {
                name: "transfer_value",
                label: "Transfer Value",
                field: "transfer_value", // Display field
                sortable: true,
                align: "right",
                sortField: "transferValueAmount", // Actual sort field
            },
            {
                name: "wage",
                label: "Salary", // Display label
                field: "wage", // Display field
                sortable: true,
                align: "right",
                sortField: "wageAmount", // Actual sort field
            },
        ];

        const fifaStatColumns = ref([
            {
                name: "PHY",
                label: "PHY",
                field: "PHY",
                sortable: true,
                align: "center",
                isFifaStat: true,
            },
            {
                name: "SHO",
                label: "SHO",
                field: "SHO",
                sortable: true,
                align: "center",
                isFifaStat: true,
            },
            {
                name: "PAS",
                label: "PAS",
                field: "PAS",
                sortable: true,
                align: "center",
                isFifaStat: true,
            },
            {
                name: "DRI",
                label: "DRI",
                field: "DRI",
                sortable: true,
                align: "center",
                isFifaStat: true,
            },
            {
                name: "DEF",
                label: "DEF",
                field: "DEF",
                sortable: true,
                align: "center",
                isFifaStat: true,
            },
            {
                name: "MEN",
                label: "MEN",
                field: "MEN",
                sortable: true,
                align: "center",
                isFifaStat: true,
            },
        ]);

        const allColumns = computed(() => [
            ...baseColumns,
            ...fifaStatColumns.value,
        ]);

        const isFifaStatColumn = (colName) => {
            return fifaStatColumns.value.some((col) => col.name === colName);
        };

        const getSortFieldKey = (colName) => {
            const baseCol = baseColumns.find((c) => c.name === colName);
            if (baseCol && baseCol.sortField) return baseCol.sortField;

            const fifaCol = fifaStatColumns.value.find(
                (c) => c.name === colName,
            );
            if (fifaCol) return fifaCol.field; // FIFA stats are direct properties now

            return colName; // Default to column name itself
        };

        const sortedPlayers = computed(() => {
            if (!sortField.value) return props.players;

            const fieldKey = getSortFieldKey(sortField.value);
            const direction = sortDirection.value;

            return [...props.players].sort((a, b) => {
                let valA = a[fieldKey];
                let valB = b[fieldKey];

                if (
                    (valA === null || valA === undefined) &&
                    (valB === null || valB === undefined)
                )
                    return 0;
                if (valA === null || valA === undefined)
                    return direction === "asc" ? 1 : -1;
                if (valB === null || valB === undefined)
                    return direction === "asc" ? -1 : 1;

                if (typeof valA === "number" && typeof valB === "number") {
                    return direction === "asc" ? valA - valB : valB - valA;
                }

                if (typeof valA === "string" && typeof valB === "string") {
                    valA = valA.toLowerCase();
                    valB = valB.toLowerCase();
                    if (valA < valB) return direction === "asc" ? -1 : 1;
                    if (valA > valB) return direction === "asc" ? 1 : -1;
                    return 0;
                }

                return 0;
            });
        });

        const displayedPlayers = computed(() => {
            if (pagination.rowsPerPage === 0) return sortedPlayers.value;

            const firstIndex = (pagination.page - 1) * pagination.rowsPerPage;
            const lastIndex = Math.min(
                firstIndex + pagination.rowsPerPage,
                sortedPlayers.value.length,
            );

            return sortedPlayers.value.slice(firstIndex, lastIndex);
        });

        // Styling Functions
        const getFifaStatClass = (value) => {
            if (value === null || value === undefined || value === "-")
                return "attribute-na";
            const numValue =
                typeof value === "number" ? value : parseInt(value, 10);
            if (isNaN(numValue)) return "attribute-na";
            if (numValue >= 90) return "attribute-elite";
            if (numValue >= 80) return "attribute-excellent";
            if (numValue >= 70) return "attribute-very-good";
            if (numValue >= 60) return "attribute-good";
            if (numValue >= 50) return "attribute-average";
            if (numValue >= 40) return "attribute-below-average";
            if (numValue >= 30) return "attribute-poor";
            return "attribute-very-poor";
        };

        const getMoneyClass = (value) => {
            // This function assumes `value` is the display string (e.g., "€1.5M")
            // For robust classification, it should parse this string into a number.
            // However, sorting uses pre-calculated `transferValueAmount` and `wageAmount`.
            // This is a simplified version for display styling.
            let amount = 0;
            if (typeof value === "string") {
                const cleaned = value.replace(/[^0-9.MKmk]/g, ""); // Keep M, K, m, k
                if (cleaned.toLowerCase().includes("m"))
                    amount = parseFloat(cleaned) * 1000000;
                else if (cleaned.toLowerCase().includes("k"))
                    amount = parseFloat(cleaned) * 1000;
                else amount = parseFloat(cleaned);
                if (isNaN(amount)) amount = 0;
            } else if (typeof value === "number") {
                amount = value; // If it's already a number (e.g. from direct binding if data changes)
            }

            if (amount >= 10000000) return "money-very-high";
            if (amount >= 1000000) return "money-high";
            if (amount >= 100000) return "money-medium-high";
            if (amount >= 10000) return "money-medium";
            if (amount > 0) return "money-low";
            return "money-na";
        };

        const onRequest = (requestProps) => {
            const { page, rowsPerPage, sortBy, descending } =
                requestProps.pagination;
            pagination.page = page;
            pagination.rowsPerPage = rowsPerPage;

            if (sortBy) {
                const newSortField = sortBy;
                const newSortDirection = descending ? "desc" : "asc";

                if (
                    sortField.value !== newSortField ||
                    sortDirection.value !== newSortDirection
                ) {
                    sortField.value = newSortField;
                    sortDirection.value = newSortDirection;

                    emit("update:sort", {
                        key: getSortFieldKey(sortField.value),
                        direction: sortDirection.value,
                        isAttribute: false,
                        isFifaStat: isFifaStatColumn(sortField.value),
                        displayField: sortField.value,
                    });
                }
            }
        };

        const onPageChange = (page) => {
            pagination.page = page;
        };

        const onRowsPerPageChange = (rowsPerPage) => {
            pagination.rowsPerPage = rowsPerPage;
            pagination.page = 1; // Reset to first page
        };

        const customSort = (rows, sortBy, descending) => {
            const fieldKey = getSortFieldKey(sortBy);
            const direction = descending ? "desc" : "asc";

            return [...rows].sort((a, b) => {
                let valA = a[fieldKey];
                let valB = b[fieldKey];

                if (
                    (valA === null || valA === undefined) &&
                    (valB === null || valB === undefined)
                )
                    return 0;
                if (valA === null || valA === undefined)
                    return direction === "asc" ? 1 : -1;
                if (valB === null || valB === undefined)
                    return direction === "asc" ? -1 : 1;

                if (typeof valA === "number" && typeof valB === "number") {
                    return direction === "asc" ? valA - valB : valB - valA;
                }
                if (typeof valA === "string" && typeof valB === "string") {
                    valA = valA.toLowerCase();
                    valB = valB.toLowerCase();
                    if (valA < valB) return direction === "asc" ? -1 : 1;
                    if (valA > valB) return direction === "asc" ? 1 : -1;
                    return 0;
                }
                return 0;
            });
        };

        const sortTable = (field, isFifaOrAttribute = false) => {
            const actualSortKey = getSortFieldKey(field);

            if (sortField.value === field) {
                sortDirection.value =
                    sortDirection.value === "asc" ? "desc" : "asc";
            } else {
                sortField.value = field;
                sortDirection.value = "asc";
            }

            pagination.sortBy = field;
            pagination.descending = sortDirection.value === "desc";

            emit("update:sort", {
                key: actualSortKey,
                direction: sortDirection.value,
                isAttribute: false,
                isFifaStat: isFifaStatColumn(field),
                displayField: field,
            });
        };

        // --- START: Row Click Handler ---
        const onRowClick = (player) => {
            // console.log('Player selected from table:', player);
            emit("player-selected", player); // Emit the full player object
        };
        // --- END: Row Click Handler ---

        return {
            sortField,
            sortDirection,
            pagination,
            pagesNumber,
            rowsPerPageOptions,
            maxPagesToShow,
            allColumns,
            sortedPlayers,
            displayedPlayers,
            isFifaStatColumn,
            getFifaStatClass,
            getMoneyClass,
            onRequest,
            onPageChange,
            onRowsPerPageChange,
            customSort,
            sortTable,
            onRowClick, // Expose the row click handler
        };
    },
};
</script>

<style scoped>
.player-data-table {
    width: 100%;
    overflow-x: auto; /* Ensures table is scrollable horizontally if needed */
}

:deep(.q-table th) {
    font-weight: 600;
    background-color: #f3f5f9;
    white-space: nowrap; /* Prevent header text wrapping */
}

:deep(.q-table td) {
    white-space: nowrap; /* Prevent cell text wrapping */
}

:deep(.q-table tr:nth-child(even)) {
    background-color: #f9fafb;
}

/* :deep(.q-table tr:hover) { background-color: #e5f1fb; } */ /* Default hover is fine, or use custom below */

:deep(.q-pagination .q-btn.q-btn--active) {
    background-color: var(--q-primary);
    color: white;
}

.attribute-value {
    /* General styling for stat values */
    display: inline-block;
    min-width: 30px; /* Adjusted min-width for 0-100 scale */
    text-align: center;
    font-weight: 600;
    padding: 2px 5px; /* Adjusted padding */
    border-radius: 3px;
    font-size: 0.85em; /* Slightly smaller font for denser table */
}

/* FIFA Stat Specific Styling (0-100 scale) */
.fifa-stat-value {
    font-size: 1.1em; /* Make FIFA stats a bit more prominent in the table */
    padding: 4px 8px;
}

.attribute-elite {
    /* 90-100 */
    background-color: #9c27b0; /* Accent color for elite */
    color: white;
}
.attribute-excellent {
    /* 80-89 */
    background-color: #20c997;
    color: white;
}
.attribute-very-good {
    /* 70-79 */
    background-color: #4dabf7;
    color: white;
}
.attribute-good {
    /* 60-69 */
    background-color: #82c91e;
    color: #212529;
}
.attribute-average {
    /* 50-59 */
    background-color: #ffc107; /* Brighter yellow for average */
    color: #212529;
}
.attribute-below-average {
    /* 40-49 */
    background-color: #fab005; /* Original average, now below average */
    color: #212529;
}
.attribute-poor {
    /* 30-39 */
    background-color: #ff922b;
    color: #212529;
}
.attribute-very-poor {
    /* < 30 */
    background-color: #fa5252;
    color: white;
}
.attribute-na {
    background-color: #e9ecef;
    color: #868e96;
}

/* Money value styling (remains the same) */
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

/* Style for clickable rows */
.table-row-hover:hover {
    background-color: #eef6ff !important; /* A light blueish hover, !important can help override Quasar's default even/odd row styling on hover */
    /* Or use a Quasar color: background-color: var(--q-primary-light) !important; */
}
</style>
