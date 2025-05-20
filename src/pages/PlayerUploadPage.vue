<template>
    <q-page padding>
        <div class="q-pa-md">
            <h1 class="text-h4 text-center q-mb-lg">
                Football Manager HTML Player Parser
            </h1>

            <q-card class="q-mb-md bg-blue-1">
                <q-card-section>
                    <div class="text-subtitle1 text-weight-bold">
                        Instructions:
                    </div>
                    <ol class="q-ml-md">
                        <li>Ensure the Go API is running.</li>
                        <li>Select an HTML file. Click "Upload and Parse".</li>
                        <li>
                            The table will now include Nationality and Overall
                            rating columns.
                        </li>
                        <li>
                            Use filters for Name, Club, Position, Nationality,
                            and Transfer Value.
                        </li>
                        <li>Click on any player row for a detailed view.</li>
                    </ol>
                </q-card-section>
            </q-card>

            <q-card class="q-mb-md">
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Upload HTML File:</div>
                    <q-file
                        v-model="playerFile"
                        label="Select HTML file"
                        accept=".html"
                        outlined
                        counter
                    >
                        <template v-slot:prepend
                            ><q-icon name="attach_file"
                        /></template>
                    </q-file>
                    <q-btn
                        class="q-mt-md full-width"
                        color="primary"
                        label="Upload and Parse"
                        :loading="loading"
                        :disable="!playerFile"
                        @click="uploadAndParse"
                    />
                </q-card-section>
            </q-card>

            <q-card class="q-mb-md" v-if="allPlayers.length > 0">
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Search Players</div>
                    <div class="row q-col-gutter-md">
                        <div class="col-12 col-sm-6 col-md-2">
                            <q-input
                                v-model="filters.name"
                                label="Player Name"
                                dense
                                outlined
                                clearable
                                @update:model-value="handleSearch"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-2">
                            <q-input
                                v-model="filters.club"
                                label="Club"
                                dense
                                outlined
                                clearable
                                @update:model-value="handleSearch"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3">
                            <q-select
                                v-model="filters.position"
                                :options="positionFilterOptions"
                                label="Position / Group"
                                dense
                                outlined
                                clearable
                                emit-value
                                map-options
                                @update:model-value="handleSearch"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-2">
                            <q-input
                                v-model="filters.nationality"
                                label="Nationality"
                                dense
                                outlined
                                clearable
                                @update:model-value="handleSearch"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3">
                            <q-input
                                v-model="filters.transferValue"
                                label="Transfer Value"
                                dense
                                outlined
                                clearable
                                placeholder="e.g., €1.5M, >1M, <500K"
                                @update:model-value="handleSearch"
                            />
                        </div>
                        <div class="col-12 flex items-center q-mt-sm">
                            <q-btn
                                color="grey"
                                label="Clear All Filters"
                                class="full-width"
                                @click="clearAllFilters"
                                :disable="!hasActiveFilters"
                            />
                        </div>
                    </div>
                </q-card-section>
            </q-card>

            <q-banner
                v-if="error"
                class="bg-negative text-white q-mb-md"
                rounded
            >
                {{ error }}
                <template v-slot:action
                    ><q-btn
                        flat
                        color="white"
                        label="Dismiss"
                        @click="error = ''"
                /></template>
            </q-banner>

            <template v-if="allPlayers.length > 0">
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-2">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ allPlayers.length }}
                                </div>
                                <div class="text-subtitle2">
                                    Total Players
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-2">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ filteredPlayers.length }}
                                </div>
                                <div class="text-subtitle2">
                                    Filtered
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-2">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ uniqueClubsCount }}
                                </div>
                                <div class="text-subtitle2">
                                    Clubs
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ uniqueParsedPositionsCount }}
                                </div>
                                <div class="text-subtitle2">
                                    Positions
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center"
                            ><q-card-section
                                ><div class="text-h6">
                                    {{ uniqueNationalitiesCount }}
                                </div>
                                <div class="text-subtitle2">
                                    Nationalities
                                </div></q-card-section
                            ></q-card
                        >
                    </div>
                </div>
                <PlayerDataTable
                    :players="filteredPlayers"
                    :loading="loading"
                    @update:sort="handleSort"
                    @player-selected="handlePlayerSelected"
                />
            </template>

            <q-card v-else-if="!loading" class="q-pa-lg text-center">
                <q-icon name="upload_file" size="4rem" color="grey-7" />
                <div class="text-h6 q-mt-md">No Player Data Yet</div>
                <div class="text-grey-7">Upload a file to see player data</div>
            </q-card>
        </div>

        <PlayerDetailDialog
            :player="selectedPlayer"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
        />
    </q-page>
</template>

<script>
import { ref, computed, reactive } from "vue";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import playerService from "../services/playerService";

// Position Parsing Logic (assumed to be correct from previous steps)
const positionRoleMap = {
    GK: "Goalkeeper",
    SW: "Sweeper",
    D: "Defender",
    WB: "Wing-Back",
    DM: "Defensive Midfielder",
    M: "Midfielder",
    AM: "Attacking Midfielder",
    ST: "Striker",
    F: "Forward",
};
const positionSideMap = { R: "Right", L: "Left", C: "Centre" };
const standardizedPositionNameMap = {
    "Goalkeeper (Centre)": "Goalkeeper",
    Goalkeeper: "Goalkeeper",
    "Sweeper (Centre)": "Sweeper",
    Sweeper: "Sweeper",
    "Defender (Right)": "Right Back",
    "Defender (Left)": "Left Back",
    "Defender (Centre)": "Centre Back",
    "Wing-Back (Right)": "Right Wing-Back",
    "Wing-Back (Left)": "Left Wing-Back",
    "Wing-Back (Centre)": "Centre Wing-Back",
    "Defensive Midfielder (Right)": "Right Defensive Midfielder",
    "Defensive Midfielder (Left)": "Left Defensive Midfielder",
    "Defensive Midfielder (Centre)": "Centre Defensive Midfielder",
    "Midfielder (Right)": "Right Midfielder",
    "Midfielder (Left)": "Left Midfielder",
    "Midfielder (Centre)": "Centre Midfielder",
    "Attacking Midfielder (Right)": "Right Attacking Midfielder",
    "Attacking Midfielder (Left)": "Left Attacking Midfielder",
    "Attacking Midfielder (Centre)": "Centre Attacking Midfielder",
    "Striker (Centre)": "Striker",
    Striker: "Striker",
    "Forward (Right)": "Right Forward",
    "Forward (Left)": "Left Forward",
    "Forward (Centre)": "Centre Forward",
};
const positionGroups = {
    Goalkeepers: ["Goalkeeper"],
    Defenders: [
        "Sweeper",
        "Right Back",
        "Left Back",
        "Centre Back",
        "Right Wing-Back",
        "Left Wing-Back",
        "Centre Wing-Back",
    ],
    Midfielders: [
        "Right Defensive Midfielder",
        "Left Defensive Midfielder",
        "Centre Defensive Midfielder",
        "Right Midfielder",
        "Left Midfielder",
        "Centre Midfielder",
        "Right Attacking Midfielder",
        "Left Attacking Midfielder",
        "Centre Attacking Midfielder",
    ],
    Attackers: ["Striker", "Right Forward", "Left Forward", "Centre Forward"],
};
function parsePlayerPositions(positionStr) {
    if (!positionStr || typeof positionStr !== "string") return [];
    const parsedPositions = new Set();
    const parts = positionStr
        .split(",")
        .map((p) => p.trim())
        .filter((p) => p);
    parts.forEach((part) => {
        const match = part.match(/^([A-Z]+)(?:\s*\(([A-Z]+)\))?$/);
        if (!match) return;
        const roleKey = match[1];
        const sideKeysStr = match[2];
        const roleFullName = positionRoleMap[roleKey];
        if (!roleFullName) return;
        if (sideKeysStr) {
            for (const sideKey of sideKeysStr) {
                const sideFullName = positionSideMap[sideKey];
                if (sideFullName) {
                    const detailedName = `${roleFullName} (${sideFullName})`;
                    parsedPositions.add(
                        standardizedPositionNameMap[detailedName] ||
                            detailedName,
                    );
                }
            }
        } else {
            if (["D", "DM", "M", "AM", "ST", "F", "WB"].includes(roleKey)) {
                const detailedName = `${roleFullName} (Centre)`;
                parsedPositions.add(
                    standardizedPositionNameMap[detailedName] || detailedName,
                );
            } else if (["GK", "SW"].includes(roleKey)) {
                parsedPositions.add(
                    standardizedPositionNameMap[roleFullName] || roleFullName,
                );
            } else {
                parsedPositions.add(roleFullName);
            }
        }
    });
    return Array.from(parsedPositions);
}
function getPlayerPositionGroups(parsedPositionsArray) {
    const groups = new Set();
    if (!parsedPositionsArray || parsedPositionsArray.length === 0) return [];
    parsedPositionsArray.forEach((pos) => {
        for (const groupName in positionGroups) {
            if (positionGroups[groupName].includes(pos)) {
                groups.add(groupName);
            }
        }
    });
    return Array.from(groups);
}

export default {
    name: "PlayerUploadPage",
    components: { PlayerDataTable, PlayerDetailDialog },
    setup() {
        const playerFile = ref(null);
        const loading = ref(false);
        const error = ref("");
        const allPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);

        const sortState = reactive({
            key: null,
            direction: "asc",
            isAttribute: false,
            displayField: null,
        });
        const filters = reactive({
            name: "",
            club: "",
            transferValue: "",
            position: null,
            nationality: "",
        });

        const hasActiveFilters = computed(
            () =>
                filters.name !== "" ||
                filters.club !== "" ||
                filters.transferValue !== "" ||
                filters.position !== null ||
                filters.nationality !== "",
        );
        const uniqueClubsCount = computed(
            () =>
                new Set(allPlayers.value.map((p) => p.club).filter(Boolean))
                    .size,
        );
        const uniqueParsedPositionsCount = computed(() => {
            const p = new Set();
            allPlayers.value.forEach((player) =>
                player.parsedPositions?.forEach((pos) => p.add(pos)),
            );
            return p.size;
        });
        const uniqueNationalitiesCount = computed(
            () =>
                new Set(
                    allPlayers.value.map((p) => p.nationality).filter(Boolean),
                ).size,
        );

        const parseMonetaryValue = (valueStr) => {
            if (typeof valueStr !== "string" || !valueStr) return 0;
            const c = valueStr.split(" p/w")[0];
            let m = 1;
            const l = c.toLowerCase();
            if (l.includes("m")) m = 1000000;
            else if (l.includes("k")) m = 1000;
            let n = c.replace(/[^0-9,.]/g, "");
            if (n.includes(",") && !n.includes(".")) {
                const p = n.split(",");
                let i = true;
                for (let k = 1; k < p.length; k++) {
                    if (p[k].length !== 3) {
                        i = false;
                        break;
                    }
                }
                if (i) n = n.replace(/,/g, "");
                else n = n.replace(/,([^,]*)$/, ".$1");
            } else if (n.includes(",")) {
                n = n.replace(/,/g, "");
            }
            const v = parseFloat(n);
            return Math.round(isNaN(v) ? 0 : v * m);
        };
        const fifaStatCategories = {
            PHY: ["Acc", "Pac", "Bal", "Jum", "Nat", "Sta", "Str"],
            SHO: ["Fin", "Lon", "Pen", "Tec", "Cmp", "OtB", "Hea"],
            PAS: ["Tec", "Cor", "Cro", "Fre", "L Th", "Pas", "Vis"],
            DRI: ["Dri", "Fir", "Tec", "Bal", "Agi", "Fla"],
            DEF: ["Pos", "Tck", "Hea", "Mar", "Cnt", "Ant"],
            MEN: [
                "Agg",
                "Ant",
                "Bra",
                "Cmp",
                "Dec",
                "Det",
                "Ldr",
                "OtB",
                "Pos",
                "Tea",
                "Vis",
                "Wor",
            ],
        };
        const calculateFifaStat = (attributes, categoryName) => {
            const s = fifaStatCategories[categoryName];
            if (!s || s.length === 0) return 0;
            let sum = 0;
            let count = 0;
            s.forEach((n) => {
                const v = attributes[n];
                if (v !== undefined && v !== null && !isNaN(parseInt(v, 10))) {
                    sum += parseInt(v, 10);
                    count++;
                }
            });
            if (count === 0) return 0;
            return Math.round((sum * 5) / count);
        };

        const processPlayerData = (players) => {
            return players.map((player) => {
                const transferValue = parseMonetaryValue(player.transfer_value);
                const wageValue = parseMonetaryValue(player.wage);
                const numericAttributes = {};
                if (player.attributes) {
                    Object.keys(player.attributes).forEach((key) => {
                        const value = player.attributes[key];
                        numericAttributes[key] =
                            value && !isNaN(parseInt(value, 10))
                                ? parseInt(value, 10)
                                : 0;
                    });
                }
                const parsedPlayerPositions = parsePlayerPositions(
                    player.position,
                );
                const playerPosGroups = getPlayerPositionGroups(
                    parsedPlayerPositions,
                );

                const phy = calculateFifaStat(numericAttributes, "PHY");
                const sho = calculateFifaStat(numericAttributes, "SHO");
                const pas = calculateFifaStat(numericAttributes, "PAS");
                const dri = calculateFifaStat(numericAttributes, "DRI");
                const def = calculateFifaStat(numericAttributes, "DEF");
                const men = calculateFifaStat(numericAttributes, "MEN");

                // START: Calculate Overall Rating
                const fifaStatsForOverall = [phy, sho, pas, dri, def, men];
                const overallSum = fifaStatsForOverall.reduce(
                    (acc, current) => acc + current,
                    0,
                );
                const overall =
                    fifaStatsForOverall.length > 0
                        ? Math.round(overallSum / fifaStatsForOverall.length)
                        : 0;
                // END: Calculate Overall Rating

                return {
                    ...player,
                    age: parseInt(player.age, 10) || 0,
                    transferValueAmount: transferValue,
                    wageAmount: wageValue,
                    attributes: numericAttributes,
                    PHY: phy,
                    SHO: sho,
                    PAS: pas,
                    DRI: dri,
                    DEF: def,
                    MEN: men,
                    Overall: overall, // Add Overall rating to player object
                    parsedPositions: parsedPlayerPositions,
                    positionGroups: playerPosGroups,
                };
            });
        };

        const positionFilterOptions = computed(() => {
            const o = [];
            Object.keys(positionGroups).forEach((g) => {
                o.push({ label: `${g} (Group)`, value: g });
            });
            const s = new Set();
            allPlayers.value.forEach((p) => {
                p.parsedPositions?.forEach((pos) => s.add(pos));
            });
            Array.from(s)
                .sort()
                .forEach((pos) => {
                    if (!positionGroups[pos]) {
                        o.push({ label: pos, value: pos });
                    }
                });
            return o;
        });

        const filteredPlayers = computed(() => {
            if (!allPlayers.value.length) return [];
            let tempPlayers = [...allPlayers.value];
            if (filters.name)
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.name &&
                        p.name
                            .toLowerCase()
                            .includes(filters.name.toLowerCase()),
                );
            if (filters.club)
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.club &&
                        p.club
                            .toLowerCase()
                            .includes(filters.club.toLowerCase()),
                );
            if (filters.transferValue) {
                let o = "includes";
                let c = 0;
                let f = filters.transferValue;
                if (f.startsWith(">")) {
                    o = ">";
                    f = f.substring(1);
                } else if (f.startsWith("<")) {
                    o = "<";
                    f = f.substring(1);
                }
                if (o !== "includes") c = parseMonetaryValue(f);
                tempPlayers = tempPlayers.filter((p) => {
                    const v = p.transferValueAmount || 0;
                    if (o === ">") return v > c;
                    if (o === "<") return v < c;
                    return String(p.transfer_value || "")
                        .toLowerCase()
                        .includes(filters.transferValue.toLowerCase());
                });
            }
            if (filters.position) {
                const s = filters.position;
                if (positionGroups[s]) {
                    tempPlayers = tempPlayers.filter(
                        (p) => p.positionGroups && p.positionGroups.includes(s),
                    );
                } else {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.parsedPositions && p.parsedPositions.includes(s),
                    );
                }
            }
            if (filters.nationality)
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.nationality &&
                        p.nationality
                            .toLowerCase()
                            .includes(filters.nationality.toLowerCase()),
                );
            if (sortState.key) return sortPlayersLogic([...tempPlayers]);
            return tempPlayers;
        });

        const sortPlayersLogic = (playersToSort) => {
            if (!sortState.key) return playersToSort;
            const k = sortState.isAttribute
                ? sortState.key
                : allPlayers.value.length > 0 &&
                    Object.keys(allPlayers.value[0]).includes(
                        sortState.key + "Amount",
                    )
                  ? sortState.key + "Amount"
                  : sortState.key;
            return [...playersToSort].sort((a, b) => {
                let vA, vB;
                if (sortState.isAttribute) {
                    vA = a.attributes ? a.attributes[k] : null;
                    vB = b.attributes ? b.attributes[k] : null;
                } else {
                    vA = a[k];
                    vB = b[k];
                }
                if (vA == null && vB == null) return 0;
                if (vA == null) return sortState.direction === "asc" ? 1 : -1;
                if (vB == null) return sortState.direction === "asc" ? -1 : 1;
                if (typeof vA === "number" && typeof vB === "number") {
                    return sortState.direction === "asc" ? vA - vB : vB - vA;
                }
                vA = String(vA).toLowerCase();
                vB = String(vB).toLowerCase();
                if (vA < vB) return sortState.direction === "asc" ? -1 : 1;
                if (vA > vB) return sortState.direction === "asc" ? 1 : -1;
                return 0;
            });
        };
        const uploadAndParse = async () => {
            if (!playerFile.value) {
                error.value = "Please select an HTML file first.";
                return;
            }
            loading.value = true;
            error.value = "";
            try {
                const f = new FormData();
                f.append("playerFile", playerFile.value);
                const r = await playerService.uploadPlayerFile(f);
                allPlayers.value = processPlayerData(r);
                sortState.key = null;
                sortState.direction = "asc";
                clearAllFilters();
            } catch (e) {
                error.value = `Failed to parse player data: ${e.message || "Unknown error"}`;
                allPlayers.value = [];
            } finally {
                loading.value = false;
            }
        };
        const handleSort = (sortParams) => {
            sortState.key = sortParams.key;
            sortState.direction = sortParams.direction;
            sortState.isAttribute = sortParams.isAttribute;
            sortState.displayField = sortParams.displayField;
        };
        const clearAllFilters = () => {
            filters.name = "";
            filters.club = "";
            filters.transferValue = "";
            filters.position = null;
            filters.nationality = "";
        };
        const handleSearch = () => {
            /* Reactive filtering */
        };
        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        return {
            playerFile,
            loading,
            error,
            allPlayers,
            filteredPlayers,
            uniqueClubsCount,
            uniqueParsedPositionsCount,
            uniqueNationalitiesCount,
            filters,
            hasActiveFilters,
            positionFilterOptions,
            uploadAndParse,
            handleSort,
            handleSearch,
            clearAllFilters,
            selectedPlayer,
            showPlayerDetailDialog,
            handlePlayerSelected,
        };
    },
};
</script>

<style>
.q-page {
    max-width: 1600px;
    margin: 0 auto;
}
</style>
