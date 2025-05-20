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

                <template v-slot:body-cell="props">
                    <q-td :props="props">
                        <template v-if="props.col.isFifaStat">
                            <span
                                :class="getFifaStatClass(props.value)"
                                class="attribute-value fifa-stat-value"
                            >
                                {{
                                    props.value !== undefined
                                        ? props.value
                                        : "-"
                                }}
                            </span>
                        </template>
                        <template
                            v-else-if="props.col.name === 'transfer_value'"
                        >
                            <span
                                :class="getMoneyClass(props.value)"
                                class="money-value"
                            >
                                {{ props.value || "-" }}
                            </span>
                        </template>
                        <template v-else-if="props.col.name === 'wage'">
                            <span
                                :class="getMoneyClass(props.value)"
                                class="money-value"
                            >
                                {{ props.value || "-" }}
                            </span>
                        </template>
                        <template v-else>
                            <span>{{
                                props.value !== undefined &&
                                props.value !== null
                                    ? props.value
                                    : "-"
                            }}</span>
                        </template>
                    </q-td>
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
    emits: ["update:sort"],

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
                if (pagination.page > pagesNumber.value) {
                    pagination.page =
                        pagesNumber.value > 0 ? pagesNumber.value : 1;
                }
            },
        );

        // --- START: Column Definitions ---
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
            },
            // { name: 'age', label: 'Age', field: 'age', sortable: true, align: 'left' }, // Age was not requested
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
                field: "transfer_value", // This is the display field
                sortable: true,
                align: "right",
                sortField: "transferValueAmount", // Field to use for actual sorting
            },
            {
                name: "wage", // 'Wage' is used in main.go, 'Salary' in user request. Assuming 'Wage' is correct from code.
                label: "Salary", // Display label as 'Salary'
                field: "wage", // This is the display field
                sortable: true,
                align: "right",
                sortField: "wageAmount", // Field to use for actual sorting
            },
        ];

        // FIFA Stat columns (now static)
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
        // --- END: Column Definitions ---

        // Check if a column is a FIFA stat column (used for styling/sorting)
        const isFifaStatColumn = (colName) => {
            return fifaStatColumns.value.some((col) => col.name === colName);
        };

        // Get the actual field to sort by from a column name
        const getSortFieldKey = (colName) => {
            const baseCol = baseColumns.find((c) => c.name === colName);
            if (baseCol && baseCol.sortField) return baseCol.sortField; // e.g. transferValueAmount

            const fifaCol = fifaStatColumns.value.find(
                (c) => c.name === colName,
            );
            if (fifaCol) return fifaCol.field; // e.g. PHY

            return colName; // Default to column name itself
        };

        const sortedPlayers = computed(() => {
            if (!sortField.value) return props.players;

            const fieldKey = getSortFieldKey(sortField.value);
            const direction = sortDirection.value;
            const isFifa = isFifaStatColumn(sortField.value);

            // console.log(`Sorting players by ${fieldKey} (${direction}), isFifaStat: ${isFifa}`);

            const sortedList = [...props.players].sort((a, b) => {
                let valA, valB;

                // Get values based on sort field
                // FIFA stats are now direct properties on the player object from PlayerUploadPage
                valA = a[fieldKey];
                valB = b[fieldKey];

                // Handle null/empty/undefined values
                if (
                    (valA === null || valA === undefined) &&
                    (valB === null || valB === undefined)
                )
                    return 0;
                if (valA === null || valA === undefined)
                    return direction === "asc" ? 1 : -1;
                if (valB === null || valB === undefined)
                    return direction === "asc" ? -1 : 1;

                // Direct number comparison
                if (typeof valA === "number" && typeof valB === "number") {
                    return direction === "asc" ? valA - valB : valB - valA;
                }

                // String comparison for text fields (e.g., name, club, position)
                if (typeof valA === "string" && typeof valB === "string") {
                    valA = valA.toLowerCase();
                    valB = valB.toLowerCase();
                    if (valA < valB) return direction === "asc" ? -1 : 1;
                    if (valA > valB) return direction === "asc" ? 1 : -1;
                    return 0;
                }

                // Fallback comparison (should ideally not be reached if data is clean)
                const strA = String(valA).toLowerCase();
                const strB = String(valB).toLowerCase();
                if (strA < strB) return direction === "asc" ? -1 : 1;
                if (strA > strB) return direction === "asc" ? 1 : -1;
                return 0;
            });
            return sortedList;
        });

        const displayedPlayers = computed(() => {
            if (pagination.rowsPerPage === 0) return sortedPlayers.value; // Show all if rowsPerPage is 0

            const firstIndex = (pagination.page - 1) * pagination.rowsPerPage;
            // Ensure lastIndex does not exceed the length of sortedPlayers
            const lastIndex = Math.min(
                firstIndex + pagination.rowsPerPage,
                sortedPlayers.value.length,
            );

            return sortedPlayers.value.slice(firstIndex, lastIndex);
        });

        // --- START: Styling Functions ---
        // New styling for FIFA stats (0-100 scale)
        const getFifaStatClass = (value) => {
            if (value === null || value === undefined || value === "-")
                return "attribute-na";
            const numValue =
                typeof value === "number" ? value : parseInt(value, 10);
            if (isNaN(numValue)) return "attribute-na";

            if (numValue >= 90) return "attribute-elite"; // 90-100
            if (numValue >= 80) return "attribute-excellent"; // 80-89
            if (numValue >= 70) return "attribute-very-good"; // 70-79
            if (numValue >= 60) return "attribute-good"; // 60-69
            if (numValue >= 50) return "attribute-average"; // 50-59
            if (numValue >= 40) return "attribute-below-average"; // 40-49
            if (numValue >= 30) return "attribute-poor"; // 30-39
            return "attribute-very-poor"; // < 30
        };

        // Existing styling for money values
        const getMoneyClass = (value) => {
            if (value === null || value === undefined || value === "-")
                return "money-na";
            // Assuming parseMonetaryValue is available or value is already numeric
            // For simplicity, let's assume PlayerUploadPage has already converted these to numbers if needed for sorting
            // Here, we'd typically parse it if it's still a string.
            // const amount = parseMonetaryValue(value); // This function would be needed here if not pre-processed
            // For now, we'll rely on the `transferValueAmount` and `wageAmount` for actual numeric value
            // and this function might need adjustment if `value` is the raw string.
            // Let's assume `value` for display is the string, and we need to parse for class.
            // This is a simplified version, real parsing should be robust.
            let amount = 0;
            if (typeof value === "string") {
                const cleaned = value.replace(/[^0-9.MKmk]/g, "");
                if (cleaned.toLowerCase().includes("m"))
                    amount = parseFloat(cleaned) * 1000000;
                else if (cleaned.toLowerCase().includes("k"))
                    amount = parseFloat(cleaned) * 1000;
                else amount = parseFloat(cleaned);
            } else if (typeof value === "number") {
                amount = value;
            }

            if (amount >= 10000000) return "money-very-high";
            if (amount >= 1000000) return "money-high";
            if (amount >= 100000) return "money-medium-high";
            if (amount >= 10000) return "money-medium";
            if (amount > 0) return "money-low";
            return "money-na";
        };
        // --- END: Styling Functions ---

        const onRequest = (requestProps) => {
            const { page, rowsPerPage, sortBy, descending } =
                requestProps.pagination;
            // console.log(`Q-Table onRequest: page ${page}, rows ${rowsPerPage}, sortBy ${sortBy}, descending ${descending}`);

            pagination.page = page;
            pagination.rowsPerPage = rowsPerPage;

            // If sortBy is provided by q-table, update our local sort state
            // This happens when user clicks q-table's native sort icons (if not overridden by custom header)
            if (sortBy) {
                const newSortField = sortBy;
                const newSortDirection = descending ? "desc" : "asc";

                if (
                    sortField.value !== newSortField ||
                    sortDirection.value !== newSortDirection
                ) {
                    sortField.value = newSortField;
                    sortDirection.value = newSortDirection;
                    // console.log(`Sort updated by q-table internal: ${sortField.value} (${sortDirection.value})`);

                    // Emit update to parent if needed, or handle sorting directly via `sortedPlayers` computed prop
                    emit("update:sort", {
                        key: getSortFieldKey(sortField.value),
                        direction: sortDirection.value,
                        isAttribute: false, // Determine this based on column def if necessary
                        isFifaStat: isFifaStatColumn(sortField.value),
                        displayField: sortField.value,
                    });
                }
            }
        };

        const onPageChange = (page) => {
            // console.log(`Page changed to: ${page}`);
            pagination.page = page;
        };

        const onRowsPerPageChange = (rowsPerPage) => {
            // console.log(`Rows per page changed to: ${rowsPerPage}`);
            pagination.rowsPerPage = rowsPerPage;
            pagination.page = 1; // Reset to first page
        };

        // This method is called by q-table if `sort-method` prop is used.
        // It should return the sorted array.
        // Our `sortedPlayers` computed property already handles the sorting,
        // so this customSort can just return the already sorted `sortedPlayers.value`.
        // However, q-table expects this to take `rows, sortBy, descending`.
        const customSort = (rows, sortBy, descending) => {
            // console.log(`Q-Table customSort called: sortBy ${sortBy}, descending ${descending}. Rows count: ${rows.length}`);
            // Update internal sort state if q-table tries to sort
            if (sortBy) {
                if (
                    sortField.value !== sortBy ||
                    sortDirection.value !== (descending ? "desc" : "asc")
                ) {
                    sortField.value = sortBy;
                    sortDirection.value = descending ? "desc" : "asc";
                    // console.log(`Sort state updated from customSort: ${sortField.value} (${sortDirection.value})`);
                    // Emit to parent to synchronize if PlayerUploadPage manages the primary sort state
                    emit("update:sort", {
                        key: getSortFieldKey(sortField.value),
                        direction: sortDirection.value,
                        isAttribute: false, // Simplified, adjust if needed
                        isFifaStat: isFifaStatColumn(sortField.value),
                        displayField: sortField.value,
                    });
                }
            }
            // The actual sorting is done by the `sortedPlayers` computed property,
            // which reacts to `sortField` and `sortDirection`.
            // `rows` passed here would be the unsorted `props.players`.
            // We need to sort `rows` based on `sortBy` and `descending` for q-table.
            // Or, ensure q-table uses `sortedPlayers`.
            // For simplicity with the current setup, let `sortedPlayers` handle it.
            // If q-table's `sort-method` is used, it expects this function to perform the sort.
            // Let's re-implement the sort here based on sortBy and descending for q-table's expectation.

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

        // This is triggered by clicking custom header cells
        const sortTable = (field, isFifaOrAttribute = false) => {
            const actualSortKey = getSortFieldKey(field);

            if (sortField.value === field) {
                sortDirection.value =
                    sortDirection.value === "asc" ? "desc" : "asc";
            } else {
                sortField.value = field; // This should be the display field name (e.g. 'PHY', 'name')
                sortDirection.value = "asc";
            }

            // Update pagination sort props to potentially trigger q-table's internal sort if not fully overridden
            pagination.sortBy = field; // Use display field for q-table's `sortBy` state
            pagination.descending = sortDirection.value === "desc";

            // console.log(`Sort request from header click: ${field} → actual key ${actualSortKey} (${sortDirection.value})`);

            emit("update:sort", {
                key: actualSortKey, // The actual data key for sorting (e.g. PHY, transferValueAmount)
                direction: sortDirection.value,
                isAttribute: false, // For simplicity, can be refined if original attributes are still used
                isFifaStat: isFifaStatColumn(field), // Check if it's a FIFA stat
                displayField: field, // Original field name for UI
            });
        };

        return {
            sortField,
            sortDirection,
            pagination,
            pagesNumber,
            rowsPerPageOptions,
            maxPagesToShow,
            allColumns,
            sortedPlayers, // Use this for the table rows
            displayedPlayers, // This is now correctly paginating sortedPlayers
            isFifaStatColumn,
            getFifaStatClass,
            getMoneyClass,
            onRequest,
            onPageChange,
            onRowsPerPageChange,
            customSort, // Provide this to q-table
            sortTable,
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

:deep(.q-table tr:hover) {
    background-color: #e5f1fb;
}

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
    /* Can add specific styles if different from generic .attribute-value */
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
</style>
