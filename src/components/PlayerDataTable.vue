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
import { formatCurrency } from "../utils/currencyUtils"; // Import the utility

export default {
    name: "PlayerDataTable",
    props: {
        players: { type: Array, required: true },
        loading: { type: Boolean, default: false },
        isGoalkeeperView: { type: Boolean, default: false },
        currencySymbol: { type: String, default: "$" }, // New prop
    },
    emits: ["update:sort", "player-selected"],

    setup(props, { emit }) {
        const $q = useQuasar();
        const sortField = ref(null);
        const sortDirection = ref("asc");
        const rowsPerPageOptions = [10, 15, 20, 50, 0];
        const maxPagesToShow = 7;

        const pagination = reactive({
            sortBy: null,
            descending: false,
            page: 1,
            rowsPerPage: 15,
        });

        const pagesNumber = computed(() =>
            pagination.rowsPerPage === 0 || props.players.length === 0
                ? 1
                : Math.ceil(props.players.length / pagination.rowsPerPage),
        );

        watch(
            () => props.players.length,
            () => {
                pagination.page = 1;
            },
        );
        watch(
            () => pagination.rowsPerPage,
            () => {
                if (pagination.page > pagesNumber.value)
                    pagination.page =
                        pagesNumber.value > 0 ? pagesNumber.value : 1;
            },
        );

        const nameColumnStyle =
            "max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const ageColumnStyle =
            "max-width: 60px; text-align: center; white-space: nowrap;";
        const positionColumnStyle =
            "max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const clubColumnStyle =
            "max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const moneyColumnStyle =
            "max-width: 110px; text-align: right; white-space: nowrap;"; // Slightly wider for currency
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
                label: "Value", // Symbol will be added in header template
                field: "transfer_value", // This is the original string for display via formatCurrency
                sortable: true,
                align: "right",
                sortField: "transferValueAmount", // Use numeric field for actual sorting
                style: moneyColumnStyle,
                headerStyle: moneyColumnStyle,
            },
            wage: {
                name: "wage",
                label: "Salary", // Symbol will be added in header template
                field: "wage", // This is the original string for display via formatCurrency
                sortable: true,
                align: "right",
                sortField: "wageAmount", // Use numeric field for actual sorting
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
                      allFifaStatDefinitions.GK,
                      allFifaStatDefinitions.PHY,
                      allFifaStatDefinitions.PAS,
                      allFifaStatDefinitions.MEN,
                      allFifaStatDefinitions.DRI,
                      allFifaStatDefinitions.DEF,
                      allFifaStatDefinitions.SHO,
                  ]
                : [
                      allFifaStatDefinitions.PHY,
                      allFifaStatDefinitions.SHO,
                      allFifaStatDefinitions.PAS,
                      allFifaStatDefinitions.DRI,
                      allFifaStatDefinitions.DEF,
                      allFifaStatDefinitions.MEN,
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
            if (!sortField.value) return props.players;

            const fieldKey = getSortFieldKey(sortField.value);
            const direction = sortDirection.value;

            return [...props.players].sort((a, b) => {
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
            return "money-na"; // For 0 or unparsed
        };

        const onFlagError = (event) => {
            if (event.target) event.target.style.display = "none";
            const iconElement = event.target.nextElementSibling;
            if (iconElement && iconElement.tagName === "I")
                iconElement.style.display = "inline-flex";
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
            /* Uses sortedPlayers computed */ return rows;
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

        // Method to format currency for display using the utility
        // It now takes the numeric amount first, then the original display string as a fallback.
        const formatDisplayCurrency = (numericAmount, originalDisplayValue) => {
            // The backend's `transferValueAmount` or `wageAmount` is the single numeric value.
            // The `originalDisplayValue` (e.g., player.transfer_value) is the string from HTML "£85M - £99M"
            // We prioritize formatting the numericAmount.
            return formatCurrency(
                numericAmount,
                props.currencySymbol,
                originalDisplayValue,
            );
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
            formatDisplayCurrency, // Expose the new formatting method
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
    table-layout: auto; // Allow table to determine column widths

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
            background-color: #f0f4f8; // Lighter header for light mode
            color: $grey-8;
            border-bottom: 1px solid #dde2e6;
        }
        td {
            border-bottom: 1px solid #eef2f5; // Lighter cell borders
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
        font-size: 0.8rem; // Slightly smaller header text
        padding: 8px 10px; // Consistent padding
        border-right: 0; // Remove vertical borders between headers
    }
    td {
        vertical-align: middle;
        padding: 6px 10px; // Consistent padding
        border-right: 0; // Remove vertical borders between cells
    }
    .table-cell-enhanced {
        font-size: 0.85rem; // Slightly larger cell text
    }
}

.table-row-hover {
    &:hover {
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.08) !important;
        }
        .body--light & {
            background-color: #e3f2fd !important; // Light blue hover for light mode
        }
    }
}

// Monetary value styling
.money-value {
    display: inline-block;
    font-weight: 500;
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 0.8rem; // Slightly smaller money values
}
// Specific money classes (colors defined in app.scss or here if preferred)
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

// Ensure icons and images in flex containers don't shrink unexpectedly
.flex.items-center .q-icon,
.flex.items-center img {
    flex-shrink: 0;
}

// Dark mode select in pagination
:deep(.q-table__bottom .q-select .q-field__native),
:deep(.q-table__bottom .q-select .q-field__input) {
    .body--dark & {
        color: $grey-3;
    }
}

// Allow text wrapping in cells
.player-q-table td,
.player-q-table th {
    white-space: normal; // Allow wrapping
    word-break: break-word; // Break long words
}
</style>
