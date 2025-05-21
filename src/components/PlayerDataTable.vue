<template>
    <div class="player-data-table-container">
        <div
            v-if="sortField"
            class="text-caption q-mb-sm q-pa-xs rounded-borders"
            :class="
                $q.dark.isActive
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
            :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1'"
            flat
            bordered
        >
            <p :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-7'">
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
            :class="$q.dark.isActive ? 'q-table--dark' : ''"
            table-header-class="player-table-header"
            dense
        >
            <template v-slot:header="props">
                <q-tr
                    :props="props"
                    :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-grey-2'"
                >
                    <q-th
                        v-for="col in props.cols"
                        :key="col.name"
                        :props="props"
                        class="cursor-pointer text-weight-bold"
                        @click="sortTable(col.name)"
                        :class="[
                            $q.dark.isActive ? 'text-grey-3' : 'text-grey-8',
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
                            $q.dark.isActive ? 'text-grey-4' : 'text-grey-9',
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
                                        $q.dark.isActive ? 'grey-6' : 'grey-7'
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
                    :color="$q.dark.isActive ? 'grey-6' : 'primary'"
                    :active-color="$q.dark.isActive ? 'primary' : 'primary'"
                    :text-color="$q.dark.isActive ? 'white' : 'primary'"
                    :active-text-color="$q.dark.isActive ? 'white' : 'white'"
                />
                <q-space />
                <span
                    class="text-caption q-mr-sm"
                    :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
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
                    :class="$q.dark.isActive ? 'text-grey-3 bg-grey-8' : ''"
                    :popup-content-class="
                        $q.dark.isActive
                            ? 'bg-grey-8 text-white'
                            : 'bg-white text-dark'
                    "
                />
                <span
                    class="q-ml-md text-caption"
                    :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
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
        const $q = useQuasar();
        const sortField = ref(null);
        const sortDirection = ref("asc");
        const rowsPerPageOptions = [10, 15, 20, 50, 0]; // 0 means 'All'
        const maxPagesToShow = 7;

        const pagination = reactive({
            sortBy: null,
            descending: false,
            page: 1,
            rowsPerPage: 15, // Default rows per page
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

        // Column definitions with adjusted styles for width and readability
        const nameColumnStyle =
            "min-width: 200px; width: 220px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const clubColumnStyle =
            "min-width: 180px; width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const positionColumnStyle =
            "min-width: 160px; width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const nationalityColumnStyle =
            "min-width: 150px; width: 160px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        const narrowColumnStyle =
            "min-width: 80px; width: 80px; text-align: center; white-space: nowrap;"; // For Age, Overall, FIFA stats
        const moneyColumnStyle =
            "min-width: 130px; width: 140px; text-align: right; white-space: nowrap;"; // For Value, Wage
        const defaultTextColumnStyle =
            "min-width: 140px; width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"; // For Personality, Media Handling

        const baseColumns = [
            {
                name: "name",
                label: "Name",
                field: "name",
                sortable: true,
                align: "left",
                style: nameColumnStyle,
                headerStyle: nameColumnStyle,
            },
            {
                name: "age",
                label: "Age",
                field: "age",
                sortable: true,
                align: "center",
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
            {
                name: "nationality_display",
                label: "Nationality",
                field: "nationality",
                sortable: true,
                align: "left",
                style: nationalityColumnStyle,
                headerStyle: nationalityColumnStyle,
            },
            {
                name: "position",
                label: "Position",
                field: "position",
                sortable: true,
                align: "left",
                style: positionColumnStyle,
                headerStyle: positionColumnStyle,
            },
            {
                name: "club",
                label: "Club",
                field: "club",
                sortable: true,
                align: "left",
                style: clubColumnStyle,
                headerStyle: clubColumnStyle,
            },
            {
                name: "media_handling",
                label: "Media Handling",
                field: "media_handling",
                sortable: true,
                align: "left",
                style: defaultTextColumnStyle,
                headerStyle: defaultTextColumnStyle,
            },
            {
                name: "personality",
                label: "Personality",
                field: "personality",
                sortable: true,
                align: "left",
                style: defaultTextColumnStyle,
                headerStyle: defaultTextColumnStyle,
            },
            {
                name: "transfer_value",
                label: "Value",
                field: "transfer_value",
                sortable: true,
                align: "right",
                sortField: "transferValueAmount",
                style: moneyColumnStyle,
                headerStyle: moneyColumnStyle,
            },
            {
                name: "wage",
                label: "Salary",
                field: "wage",
                sortable: true,
                align: "right",
                sortField: "wageAmount",
                style: moneyColumnStyle,
                headerStyle: moneyColumnStyle,
            },
            {
                name: "Overall",
                label: "Overall",
                field: "Overall",
                sortable: true,
                align: "center",
                isOverallStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
        ];

        const allFifaStatDefinitions = {
            GK: {
                name: "GK",
                label: "GK",
                field: "GK",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
            PHY: {
                name: "PHY",
                label: "PHY",
                field: "PHY",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
            PAS: {
                name: "PAS",
                label: "PAS",
                field: "PAS",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
            MEN: {
                name: "MEN",
                label: "MEN",
                field: "MEN",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
            DRI: {
                name: "DRI",
                label: "DRI",
                field: "DRI",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
            DEF: {
                name: "DEF",
                label: "DEF",
                field: "DEF",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
            SHO: {
                name: "SHO",
                label: "SHO",
                field: "SHO",
                sortable: true,
                align: "center",
                isFifaStat: true,
                style: narrowColumnStyle,
                headerStyle: narrowColumnStyle,
            },
        };

        const currentColumns = computed(() => {
            let fifaColumnsOrdered = [];
            if (props.isGoalkeeperView) {
                fifaColumnsOrdered = [
                    allFifaStatDefinitions.GK,
                    allFifaStatDefinitions.PHY,
                    allFifaStatDefinitions.PAS,
                    allFifaStatDefinitions.MEN,
                    allFifaStatDefinitions.DRI,
                    allFifaStatDefinitions.DEF,
                    allFifaStatDefinitions.SHO,
                ];
            } else {
                fifaColumnsOrdered = [
                    allFifaStatDefinitions.PHY,
                    allFifaStatDefinitions.SHO,
                    allFifaStatDefinitions.PAS,
                    allFifaStatDefinitions.DRI,
                    allFifaStatDefinitions.DEF,
                    allFifaStatDefinitions.MEN,
                ];
            }
            return [...baseColumns, ...fifaColumnsOrdered];
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

        const getMoneyClass = (v) => {
            let a = 0;
            if (typeof v === "string") {
                const c = v.replace(/[^0-9.MKmk]/g, "");
                if (c.toLowerCase().includes("m")) a = parseFloat(c) * 1000000;
                else if (c.toLowerCase().includes("k"))
                    a = parseFloat(c) * 1000;
                else a = parseFloat(c);
                if (isNaN(a)) a = 0;
            } else if (typeof v === "number") a = v;
            if (a >= 10000000) return "money-very-high";
            if (a >= 1000000) return "money-high";
            if (a >= 100000) return "money-medium-high";
            if (a >= 10000) return "money-medium";
            if (a > 0) return "money-low";
            return "money-na";
        };

        const onFlagError = (event) => {
            if (event.target) event.target.style.display = "none";
            const iconElement = event.target.nextElementSibling;
            if (iconElement && iconElement.tagName === "I") {
                iconElement.style.display = "inline-flex";
            }
        };

        const onRequest = (rp) => {
            const { page, rowsPerPage, sortBy, descending } = rp.pagination;
            pagination.page = page;
            pagination.rowsPerPage = rowsPerPage;
            if (sortBy) {
                const nF = sortBy;
                const nD = descending ? "desc" : "asc";
                if (sortField.value !== nF || sortDirection.value !== nD) {
                    sortField.value = nF;
                    sortDirection.value = nD;
                    emit("update:sort", {
                        key: getSortFieldKey(sortField.value),
                        direction: sortDirection.value,
                        isFifaStat: currentColumns.value.find(
                            (c) => c.name === nF,
                        )?.isFifaStat,
                        isOverallStat: currentColumns.value.find(
                            (c) => c.name === nF,
                        )?.isOverallStat,
                        displayField: sortField.value,
                    });
                }
            }
        };
        const onPageChange = (p) => {
            pagination.page = p;
        };
        const onRowsPerPageChange = (rpp) => {
            pagination.rowsPerPage = rpp;
            pagination.page = 1;
        };

        const customSort = (r, sB, d) => {
            const fK = getSortFieldKey(sB);
            const dir = d ? "desc" : "asc";
            return [...r].sort((a, b) => {
                let vA = a[fK];
                let vB = b[fK];
                if (
                    (vA === null || vA === undefined) &&
                    (vB === null || vB === undefined)
                )
                    return 0;
                if (vA === null || vA === undefined)
                    return dir === "asc" ? 1 : -1;
                if (vB === null || vB === undefined)
                    return dir === "asc" ? -1 : 1;
                if (typeof vA === "number" && typeof vB === "number")
                    return dir === "asc" ? vA - vB : vB - vA;
                if (typeof vA === "string" && typeof vB === "string") {
                    vA = vA.toLowerCase();
                    vB = vB.toLowerCase();
                    if (vA < vB) return dir === "asc" ? -1 : 1;
                    if (vA > vB) return dir === "asc" ? 1 : -1;
                    return 0;
                }
                return 0;
            });
        };
        const sortTable = (f) => {
            const aSK = getSortFieldKey(f);
            if (sortField.value === f)
                sortDirection.value =
                    sortDirection.value === "asc" ? "desc" : "asc";
            else {
                sortField.value = f;
                sortDirection.value = "asc";
            }
            pagination.sortBy = f;
            pagination.descending = sortDirection.value === "desc";
            emit("update:sort", {
                key: aSK,
                direction: sortDirection.value,
                isFifaStat: currentColumns.value.find((c) => c.name === f)
                    ?.isFifaStat,
                isOverallStat: currentColumns.value.find((c) => c.name === f)
                    ?.isOverallStat,
                displayField: f,
            });
        };
        const onRowClick = (p) => {
            emit("player-selected", p);
        };

        return {
            $q,
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
        font-size: 1rem; // Increased header font size
        padding: 12px 14px; // Slightly increased padding
        border-right: 0;
    }
    td {
        vertical-align: middle;
        padding: 10px 14px; // Slightly increased padding
        border-right: 0;
    }
    .table-cell-enhanced {
        font-size: 1.05rem; // Further increased cell font size
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
</style>
