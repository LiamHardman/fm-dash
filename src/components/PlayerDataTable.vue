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
                                    col.isFifaStat || col.isOverallStat,
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
                            <template
                                v-if="col.isFifaStat || col.isOverallStat"
                            >
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
                                        class="q-mr-xs"
                                        style="
                                            border: 1px solid #ccc;
                                            object-fit: cover;
                                        "
                                        @error="onFlagError($event, props.row)"
                                    />
                                    <q-icon
                                        v-else
                                        name="flag"
                                        size="xs"
                                        color="grey-6"
                                        class="q-mr-xs"
                                    />
                                    <span>{{
                                        props.row.nationality || "-"
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
                    <q-inner-loading showing color="primary"
                        ><q-spinner size="50px" color="primary"
                    /></q-inner-loading>
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
        players: { type: Array, required: true },
        loading: { type: Boolean, default: false },
    },
    emits: ["update:sort", "player-selected"],

    setup(props, { emit }) {
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

        const baseColumns = [
            {
                name: "name",
                label: "Name",
                field: "name",
                sortable: true,
                align: "left",
            },
            {
                name: "age",
                label: "Age",
                field: "age", // Assumes player.age is available and numeric
                sortable: true,
                align: "center",
            },
            {
                name: "nationality_display",
                label: "Nationality",
                field: "nationality",
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
            {
                name: "club",
                label: "Club",
                field: "club",
                sortable: true,
                align: "left",
            },
            {
                name: "media_handling", // New column
                label: "Media Handling",
                field: "media_handling", // Assumes player.media_handling from API
                sortable: true,
                align: "left",
            },
            {
                name: "personality", // New column
                label: "Personality",
                field: "personality", // Assumes player.personality from API
                sortable: true,
                align: "left",
            },
            {
                name: "transfer_value",
                label: "Transfer Value",
                field: "transfer_value",
                sortable: true,
                align: "right",
                sortField: "transferValueAmount",
            },
            {
                name: "wage",
                label: "Salary",
                field: "wage",
                sortable: true,
                align: "right",
                sortField: "wageAmount",
            },
            {
                name: "Overall",
                label: "Overall",
                field: "Overall",
                sortable: true,
                align: "center",
                isOverallStat: true,
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

        const getSortFieldKey = (colName) => {
            const colDef = allColumns.value.find((c) => c.name === colName);
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

        const getFifaStatClass = (v) => {
            if (v === null || v === undefined || v === "-")
                return "attribute-na";
            const n = typeof v === "number" ? v : parseInt(v, 10);
            if (isNaN(n)) return "attribute-na";
            if (n >= 90) return "attribute-elite";
            if (n >= 80) return "attribute-excellent";
            if (n >= 70) return "attribute-very-good";
            if (n >= 60) return "attribute-good";
            if (n >= 50) return "attribute-average";
            if (n >= 40) return "attribute-below-average";
            if (n >= 30) return "attribute-poor";
            return "attribute-very-poor";
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

        const onFlagError = (event, player) => {
            event.target.style.display = "none";
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
                        isFifaStat: allColumns.value.find((c) => c.name === nF)
                            ?.isFifaStat,
                        isOverallStat: allColumns.value.find(
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
                isFifaStat: allColumns.value.find((c) => c.name === f)
                    ?.isFifaStat,
                isOverallStat: allColumns.value.find((c) => c.name === f)
                    ?.isOverallStat,
                displayField: f,
            });
        };
        const onRowClick = (p) => {
            emit("player-selected", p);
        };

        return {
            sortField,
            sortDirection,
            pagination,
            pagesNumber,
            rowsPerPageOptions,
            maxPagesToShow,
            allColumns,
            sortedPlayers,
            getFifaStatClass,
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

<style scoped>
.player-data-table {
    width: 100%;
    overflow-x: auto;
}
:deep(.q-table th) {
    font-weight: 600;
    background-color: #f3f5f9;
    white-space: nowrap;
}
:deep(.q-table td) {
    white-space: nowrap;
    vertical-align: middle;
} /* Align items vertically in cells */
:deep(.q-table tr:nth-child(even)) {
    background-color: #f9fafb;
}
.table-row-hover:hover {
    background-color: #eef6ff !important;
}
:deep(.q-pagination .q-btn.q-btn--active) {
    background-color: var(--q-primary);
    color: white;
}
.attribute-value {
    display: inline-block;
    min-width: 30px;
    text-align: center;
    font-weight: 600;
    padding: 2px 5px;
    border-radius: 3px;
    font-size: 0.85em;
}
.fifa-stat-value {
    font-size: 1.1em;
    padding: 4px 8px;
}
.attribute-elite {
    background-color: #9c27b0;
    color: white;
}
.attribute-excellent {
    background-color: #20c997;
    color: white;
}
.attribute-very-good {
    background-color: #4dabf7;
    color: white;
}
.attribute-good {
    background-color: #82c91e;
    color: #212529;
}
.attribute-average {
    background-color: #ffc107;
    color: #212529;
}
.attribute-below-average {
    background-color: #fab005;
    color: #212529;
}
.attribute-poor {
    background-color: #ff922b;
    color: #212529;
}
.attribute-very-poor {
    background-color: #fa5252;
    color: white;
}
.attribute-na {
    background-color: #e9ecef;
    color: #868e96;
}
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
/* Ensure flag and text are aligned in the nationality cell */
.flex.items-center .q-icon,
.flex.items-center img {
    flex-shrink: 0; /* Prevent icon/image from shrinking */
}
</style>
