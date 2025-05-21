<template>
    <div class="player-data-table-container">
        <div
            v-if="sortField"
            class="text-caption q-mb-sm q-pa-xs rounded-borders"
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

        <q-table
            v-else
            :rows="sortedPlayers"
            :columns="currentColumns"
            :loading="loading"
            row-key="name"
            :pagination.sync="pagination"
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
        >
            <template v-slot:header="props">
                <q-tr
                    :props="props"
                    :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-grey-2'"
                >
                    <q-th
                        v-for="col in props.cols"
                        :key="col.name"
                        :props="props"
                        class="cursor-pointer text-weight-bold"
                        @click="sortTable(col.name)"
                        :class="[
                            qInstance.dark.isActive
                                ? 'text-grey-3'
                                : 'text-grey-8',
                            col.headerClasses,
                        ]"
                        :style="col.headerStyle"
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
                            class="q-ml-xs sort-icon"
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
                                        props.row[col.field],
                                        100,
                                    )
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
                        <template
                            v-else-if="
                                col.name === 'transfer_value' ||
                                col.name === 'wage'
                            "
                        >
                            <span
                                :class="getMoneyClass(props.row[col.field])"
                                class="money-value"
                            >
                                {{ props.row[col.field] || "-" }}
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
                    class="text-caption q-mr-sm"
                    :class="
                        qInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'
                    "
                    >Rows per page:</span
                >
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
                    :class="
                        qInstance.dark.isActive ? 'text-grey-3 bg-grey-8' : ''
                    "
                    :popup-content-class="
                        qInstance.dark.isActive
                            ? 'bg-grey-8 text-white'
                            : 'bg-white text-dark'
                    "
                />
                <span
                    class="q-ml-md text-caption"
                    :class="
                        qInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'
                    "
                >
                    {{
                        pagination.rowsPerPage === 0
                            ? 1
                            : (pagination.page - 1) * pagination.rowsPerPage + 1
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
</template>

<script>
import { ref, computed, reactive, watch } from "vue";
import { useQuasar } from "quasar";

export default {
    name: "PlayerDataTable",
    props: {
        players: { type: Array, required: true },
        loading: { type: Boolean, default: false },
        isGoalkeeperView: { type: Boolean, default: false },
    },
    emits: ["update:sort", "player-selected"],

    setup(props, { emit }) {
        const $q = useQuasar(); // qInstance will be used in template
        const sortField = ref(null);
        const sortDirection = ref("asc");
        const rowsPerPageOptions = [10, 15, 20, 50, 0]; // 0 means 'All'
        const maxPagesToShow = 7; // For pagination controls

        const pagination = reactive({
            sortBy: null,
            descending: false,
            page: 1,
            rowsPerPage: 15, // Default rows per page
        });

        // Computed property for the total number of pages
        const pagesNumber = computed(() =>
            pagination.rowsPerPage === 0 || props.players.length === 0
                ? 1
                : Math.ceil(props.players.length / pagination.rowsPerPage),
        );

        // Watch for changes in player count to reset to page 1
        watch(
            () => props.players.length,
            () => {
                pagination.page = 1;
            },
        );
        // Adjust page number if rowsPerPage changes and current page becomes invalid
        watch(
            () => pagination.rowsPerPage,
            () => {
                if (pagination.page > pagesNumber.value)
                    pagination.page =
                        pagesNumber.value > 0 ? pagesNumber.value : 1;
            },
        );

        // Column style definitions with max-width to control column expansion
        const nameColumnStyle =
            "max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const ageColumnStyle =
            "max-width: 60px; text-align: center; white-space: nowrap;";
        const positionColumnStyle =
            "max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const clubColumnStyle =
            "max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const moneyColumnStyle =
            "max-width: 100px; text-align: right; white-space: nowrap;";
        const overallColumnStyle =
            "max-width: 70px; text-align: center; white-space: nowrap;";
        const fifaStatColumnStyle =
            "max-width: 60px; text-align: center; white-space: nowrap;";
        const textColumnStyle = // For Personality, Media Handling
            "max-width: 120px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const nationalityColumnStyle =
            "max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";

        // Base definitions for all columns
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
                sortField: "transferValueAmount", // Use numeric field for sorting
                style: moneyColumnStyle,
                headerStyle: moneyColumnStyle,
            },
            wage: {
                name: "wage",
                label: "Salary",
                field: "wage",
                sortable: true,
                align: "right",
                sortField: "wageAmount", // Use numeric field for sorting
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
                // Column for displaying flag and name
                name: "nationality_display",
                label: "Nationality",
                field: "nationality", // Sort by full name
                sortable: true,
                align: "left",
                style: nationalityColumnStyle,
                headerStyle: nationalityColumnStyle,
            },
        };

        // FIFA stat column definitions
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
            MEN: {
                name: "MEN",
                label: "MEN",
                field: "MEN",
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
        };

        // Computed property to determine which columns to display and in what order
        const currentColumns = computed(() => {
            // Adjusted order: Name, Nationality, then other base columns
            const newOrderBase = [
                baseColumnDefinitions.name,
                baseColumnDefinitions.nationality_display, // Moved Nationality here
                baseColumnDefinitions.age,
                baseColumnDefinitions.position,
                baseColumnDefinitions.club,
                baseColumnDefinitions.transfer_value,
                baseColumnDefinitions.wage,
                baseColumnDefinitions.Overall,
            ];

            let fifaColumnsInOrder = [];
            if (props.isGoalkeeperView) {
                fifaColumnsInOrder = [
                    allFifaStatDefinitions.GK,
                    allFifaStatDefinitions.PHY,
                    allFifaStatDefinitions.PAS,
                    allFifaStatDefinitions.MEN,
                    allFifaStatDefinitions.DRI,
                    allFifaStatDefinitions.DEF,
                    allFifaStatDefinitions.SHO,
                ];
            } else {
                fifaColumnsInOrder = [
                    allFifaStatDefinitions.PHY,
                    allFifaStatDefinitions.SHO,
                    allFifaStatDefinitions.PAS,
                    allFifaStatDefinitions.DRI,
                    allFifaStatDefinitions.DEF,
                    allFifaStatDefinitions.MEN,
                ];
            }

            // Nationality is no longer in trailingColumns as it's moved to newOrderBase
            const trailingColumns = [
                baseColumnDefinitions.personality,
                baseColumnDefinitions.media_handling,
            ];

            return [...newOrderBase, ...fifaColumnsInOrder, ...trailingColumns];
        });

        // Helper to get column label for display
        const getColumnLabel = (fieldName) => {
            const col = currentColumns.value.find((c) => c.name === fieldName);
            return col ? col.label : fieldName;
        };

        // Helper to get the actual field key for sorting (e.g., 'transferValueAmount' for 'Transfer Value')
        const getSortFieldKey = (colName) => {
            const colDef = currentColumns.value.find((c) => c.name === colName);
            return colDef?.sortField || colDef?.field || colName;
        };

        // Computed property for sorted players (client-side sorting)
        const sortedPlayers = computed(() => {
            if (!sortField.value) return props.players; // Return original if no sort field

            const fieldKey = getSortFieldKey(sortField.value);
            const direction = sortDirection.value;

            return [...props.players].sort((a, b) => {
                let vA = a[fieldKey];
                let vB = b[fieldKey];

                // Handle null or undefined values to sort them to the end
                if (
                    (vA === null || vA === undefined) &&
                    (vB === null || vB === undefined)
                )
                    return 0;
                if (vA === null || vA === undefined)
                    return direction === "asc" ? 1 : -1;
                if (vB === null || vB === undefined)
                    return direction === "asc" ? -1 : 1;

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
                return 0; // Fallback for other types
            });
        });

        // Function to get CSS class based on rating value (0-100 for FIFA/Overall, 1-20 for attributes)
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
            if (percentage >= 90) return "rating-tier-6"; // Elite
            if (percentage >= 80) return "rating-tier-5"; // Excellent
            if (percentage >= 70) return "rating-tier-4"; // Good
            if (percentage >= 55) return "rating-tier-3"; // Average
            if (percentage >= 40) return "rating-tier-2"; // Below Average
            return "rating-tier-1"; // Poor
        };

        // Function to get CSS class based on monetary value
        const getMoneyClass = (valueString) => {
            let amount = 0;
            if (typeof valueString === "string") {
                const cleaned = valueString.replace(/[^0-9.MKmk]/g, "");
                if (cleaned.toLowerCase().includes("m"))
                    amount = parseFloat(cleaned) * 1000000;
                else if (cleaned.toLowerCase().includes("k"))
                    amount = parseFloat(cleaned) * 1000;
                else amount = parseFloat(cleaned);
                if (isNaN(amount)) amount = 0;
            } else if (typeof valueString === "number") {
                // If already numeric (e.g. wageAmount)
                amount = valueString;
            }

            if (amount >= 10000000) return "money-very-high";
            if (amount >= 1000000) return "money-high";
            if (amount >= 100000) return "money-medium-high";
            if (amount >= 10000) return "money-medium";
            if (amount > 0) return "money-low";
            return "money-na";
        };

        // Handle flag image loading errors
        const onFlagError = (event) => {
            if (event.target) event.target.style.display = "none"; // Hide broken image
            const iconElement = event.target.nextElementSibling;
            if (iconElement && iconElement.tagName === "I") {
                iconElement.style.display = "inline-flex";
            }
        };

        const onRequest = (requestProp) => {
            const { page, rowsPerPage, sortBy, descending } =
                requestProp.pagination;
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
                        isFifaStat: currentColumns.value.find(
                            (c) => c.name === newSortField,
                        )?.isFifaStat,
                        isOverallStat: currentColumns.value.find(
                            (c) => c.name === newSortField,
                        )?.isOverallStat,
                        displayField: sortField.value,
                    });
                }
            }
        };
        const onPageChange = (newPage) => {
            pagination.page = newPage;
        };
        const onRowsPerPageChange = (newRowsPerPage) => {
            pagination.rowsPerPage = newRowsPerPage;
            pagination.page = 1;
        };

        const customSort = (rows, sortBy, descending) => {
            const fieldKey = getSortFieldKey(sortBy);
            const direction = descending ? "desc" : "asc";
            return [...rows].sort((a, b) => {
                let vA = a[fieldKey];
                let vB = b[fieldKey];
                if (
                    (vA === null || vA === undefined) &&
                    (vB === null || vB === undefined)
                )
                    return 0;
                if (vA === null || vA === undefined)
                    return direction === "asc" ? 1 : -1;
                if (vB === null || vB === undefined)
                    return direction === "asc" ? -1 : 1;
                if (typeof vA === "number" && typeof vB === "number")
                    return direction === "asc" ? vA - vB : vB - vA;
                if (typeof vA === "string" && typeof vB === "string") {
                    vA = vA.toLowerCase();
                    vB = vB.toLowerCase();
                    if (vA < vB) return direction === "asc" ? -1 : 1;
                    if (vA > vB) return direction === "asc" ? 1 : -1;
                    return 0;
                }
                return 0;
            });
        };
        const sortTable = (fieldName) => {
            const actualSortKey = getSortFieldKey(fieldName);
            if (sortField.value === fieldName) {
                sortDirection.value =
                    sortDirection.value === "asc" ? "desc" : "asc";
            } else {
                sortField.value = fieldName;
                sortDirection.value = "asc";
            }
            pagination.sortBy = fieldName;
            pagination.descending = sortDirection.value === "desc";

            emit("update:sort", {
                key: actualSortKey,
                direction: sortDirection.value,
                isFifaStat: currentColumns.value.find(
                    (c) => c.name === fieldName,
                )?.isFifaStat,
                isOverallStat: currentColumns.value.find(
                    (c) => c.name === fieldName,
                )?.isOverallStat,
                displayField: fieldName,
            });
        };
        const onRowClick = (player) => {
            emit("player-selected", player);
        };

        return {
            qInstance: $q,
            sortField,
            sortDirection,
            pagination,
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
        };
    },
};
</script>

<style lang="scss" scoped>
.player-data-table-container {
    width: 100%;
    overflow-x: auto;
}

.player-q-table {
    width: 100%;
    table-layout: auto;

    th .sort-icon {
        vertical-align: middle;
        margin-left: 4px;
    }

    &.q-table--dark {
        th {
            color: $grey-3;
            border-bottom-color: rgba(255, 255, 255, 0.15);
        }
        td {
            border-bottom-color: rgba(255, 255, 255, 0.1);
            color: $grey-4;
        }
        tr:last-child td {
            border-bottom: 0;
        }
        .q-table__bottom {
            border-top-color: rgba(255, 255, 255, 0.15);
        }
    }

    &:not(.q-table--dark) {
        th {
            background-color: #f0f4f8;
            color: $grey-8;
            border-bottom: 1px solid #dde2e6;
        }
        td {
            border-bottom: 1px solid #eef2f5;
            color: $grey-9;
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
    }
    td {
        vertical-align: middle;
        padding: 6px 10px;
        border-right: 0;
    }
    .table-cell-enhanced {
        font-size: 0.85rem;
    }
}

.table-row-hover {
    &:hover {
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.08) !important;
        }
        .body--light & {
            background-color: #e3f2fd !important;
        }
    }
}

.money-value {
    display: inline-block;
    font-weight: 500;
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 0.8rem;
}
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
        border: 1px solid rgba(255, 255, 255, 0.15);
    }
}

.flex.items-center .q-icon,
.flex.items-center img {
    flex-shrink: 0;
}

:deep(.q-table__bottom .q-select .q-field__native),
:deep(.q-table__bottom .q-select .q-field__input) {
    .body--dark & {
        color: $grey-3;
    }
}

.player-q-table td,
.player-q-table th {
    white-space: normal;
    word-break: break-word;
}
</style>
