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
                        <li>
                            Ensure the Go API (main.go) is running (usually on
                            http://localhost:8091).
                        </li>
                        <li>
                            Select an HTML file exported from Football Manager
                            containing player data.
                        </li>
                        <li>Click "Upload and Parse".</li>
                        <li>
                            Use the new "Position / Group" filter along with
                            other filters.
                        </li>
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
                        <template v-slot:prepend>
                            <q-icon name="attach_file" />
                        </template>
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
                        <div class="col-12 col-sm-6 col-md-3">
                            <q-input
                                v-model="filters.name"
                                label="Player Name"
                                dense
                                outlined
                                clearable
                                @update:model-value="handleSearch"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3">
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
                <template v-slot:action>
                    <q-btn
                        flat
                        color="white"
                        label="Dismiss"
                        @click="error = ''"
                    />
                </template>
            </q-banner>

            <template v-if="allPlayers.length > 0">
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-3">
                        <q-card class="text-center">
                            <q-card-section>
                                <div class="text-h6">
                                    {{ allPlayers.length }}
                                </div>
                                <div class="text-subtitle2">Total Players</div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center">
                            <q-card-section>
                                <div class="text-h6">
                                    {{ filteredPlayers.length }}
                                </div>
                                <div class="text-subtitle2">
                                    Filtered Players
                                </div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center">
                            <q-card-section>
                                <div class="text-h6">
                                    {{ uniqueClubsCount }}
                                </div>
                                <div class="text-subtitle2">Unique Clubs</div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center">
                            <q-card-section>
                                <div class="text-h6">
                                    {{ uniqueParsedPositionsCount }}
                                </div>
                                <div class="text-subtitle2">
                                    Unique Positions
                                </div>
                            </q-card-section>
                        </q-card>
                    </div>
                </div>

                <PlayerDataTable
                    :players="filteredPlayers"
                    :loading="loading"
                    @update:sort="handleSort"
                />
            </template>

            <q-card v-else-if="!loading" class="q-pa-lg text-center">
                <q-icon name="upload_file" size="4rem" color="grey-7" />
                <div class="text-h6 q-mt-md">No Player Data Yet</div>
                <div class="text-grey-7">Upload a file to see player data</div>
            </q-card>
        </div>
    </q-page>
</template>

<script>
import { ref, computed, reactive } from "vue";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import playerService from "../services/playerService";

// --- START: Position Parsing and Mapping Logic ---
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

// Defines the standardized specific position names we want to use
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
    "Wing-Back (Centre)": "Centre Wing-Back", // Though (C) is rare for WB
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
    "Forward (Centre)": "Centre Forward", // If F appears
};

// Position Groups for filtering
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

/**
 * Parses a raw Football Manager position string into an array of standardized position names.
 * Example: "M (C), AM (RC)" -> ["Centre Midfielder", "Right Attacking Midfielder", "Centre Attacking Midfielder"]
 * @param {string} positionStr - The raw position string from the player data.
 * @returns {string[]} An array of standardized position names.
 */
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

        const roleKey = match[1]; // e.g., M, AM
        const sideKeysStr = match[2]; // e.g., C, RC, or undefined

        const roleFullName = positionRoleMap[roleKey];
        if (!roleFullName) return;

        if (sideKeysStr) {
            // Sides are specified e.g. (C), (RL), (RLC)
            for (const sideKey of sideKeysStr) {
                // Iterate over R, L, C
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
            // No sides specified, assume central for most roles or use base role
            if (["D", "DM", "M", "AM", "ST", "F", "WB"].includes(roleKey)) {
                // WB also often implies central if no side
                const detailedName = `${roleFullName} (Centre)`;
                parsedPositions.add(
                    standardizedPositionNameMap[detailedName] || detailedName,
                );
            } else if (["GK", "SW"].includes(roleKey)) {
                parsedPositions.add(
                    standardizedPositionNameMap[roleFullName] || roleFullName,
                );
            } else {
                parsedPositions.add(roleFullName); // Fallback for unknown role structure
            }
        }
    });
    return Array.from(parsedPositions);
}

/**
 * Determines which position groups a player belongs to based on their parsed positions.
 * @param {string[]} parsedPositionsArray - Array of standardized position names.
 * @returns {string[]} An array of group names (e.g., ["Defenders", "Midfielders"]).
 */
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
// --- END: Position Parsing and Mapping Logic ---

export default {
    name: "PlayerUploadPage",
    components: {
        PlayerDataTable,
    },

    setup() {
        const playerFile = ref(null);
        const loading = ref(false);
        const error = ref("");
        const allPlayers = ref([]);

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
            position: null, // Changed from '' to null for q-select clearable
        });

        const hasActiveFilters = computed(() => {
            return (
                filters.name !== "" ||
                filters.club !== "" ||
                filters.transferValue !== "" ||
                filters.position !== null
            );
        });

        const uniqueClubsCount = computed(() => {
            const clubs = new Set();
            allPlayers.value.forEach((player) => {
                if (player.club) clubs.add(player.club);
            });
            return clubs.size;
        });

        // Updated to count unique PARSED positions
        const uniqueParsedPositionsCount = computed(() => {
            const positions = new Set();
            allPlayers.value.forEach((player) => {
                player.parsedPositions?.forEach((pos) => positions.add(pos));
            });
            return positions.size;
        });

        const parseMonetaryValue = (valueStr) => {
            // (Monetary parsing logic - kept from previous version, assumed correct)
            if (typeof valueStr !== "string" || !valueStr) return 0;
            const cleanedStr = valueStr.split(" p/w")[0];
            let multiplier = 1;
            const lowerStr = cleanedStr.toLowerCase();
            if (lowerStr.includes("m")) multiplier = 1000000;
            else if (lowerStr.includes("k")) multiplier = 1000;
            let numStr = cleanedStr.replace(/[^0-9,.]/g, "");
            if (numStr.includes(",") && !numStr.includes(".")) {
                const parts = numStr.split(",");
                let isThousandsSeparator = true;
                for (let i = 1; i < parts.length; i++) {
                    if (parts[i].length !== 3) {
                        isThousandsSeparator = false;
                        break;
                    }
                }
                if (isThousandsSeparator) numStr = numStr.replace(/,/g, "");
                else numStr = numStr.replace(/,([^,]*)$/, ".$1");
            } else if (numStr.includes(",")) {
                numStr = numStr.replace(/,/g, "");
            }
            const numericValue = parseFloat(numStr);
            return Math.round(
                isNaN(numericValue) ? 0 : numericValue * multiplier,
            );
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
            const statNames = fifaStatCategories[categoryName];
            if (!statNames || statNames.length === 0) return 0;
            let sum = 0;
            let count = 0;
            statNames.forEach((statName) => {
                const value = attributes[statName];
                if (
                    value !== undefined &&
                    value !== null &&
                    !isNaN(parseInt(value, 10))
                ) {
                    sum += parseInt(value, 10);
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

                // Parse positions and determine groups
                const parsedPlayerPositions = parsePlayerPositions(
                    player.position,
                );
                const playerPosGroups = getPlayerPositionGroups(
                    parsedPlayerPositions,
                );

                return {
                    ...player,
                    age: parseInt(player.age, 10) || 0,
                    transferValueAmount: transferValue,
                    wageAmount: wageValue,
                    attributes: numericAttributes,
                    PHY: calculateFifaStat(numericAttributes, "PHY"),
                    SHO: calculateFifaStat(numericAttributes, "SHO"),
                    PAS: calculateFifaStat(numericAttributes, "PAS"),
                    DRI: calculateFifaStat(numericAttributes, "DRI"),
                    DEF: calculateFifaStat(numericAttributes, "DEF"),
                    MEN: calculateFifaStat(numericAttributes, "MEN"),
                    parsedPositions: parsedPlayerPositions, // Add parsed positions to player object
                    positionGroups: playerPosGroups, // Add position groups to player object
                };
            });
        };

        // Options for the position filter dropdown
        const positionFilterOptions = computed(() => {
            const options = [];
            // Add groups first
            Object.keys(positionGroups).forEach((groupName) => {
                options.push({
                    label: `${groupName} (Group)`,
                    value: groupName,
                });
            });

            // Add unique specific positions found in the data
            const specificPlayerPositions = new Set();
            allPlayers.value.forEach((player) => {
                player.parsedPositions?.forEach((pos) =>
                    specificPlayerPositions.add(pos),
                );
            });

            Array.from(specificPlayerPositions)
                .sort()
                .forEach((pos) => {
                    // Avoid adding if a group has the same name (unlikely here)
                    if (!positionGroups[pos]) {
                        options.push({ label: pos, value: pos });
                    }
                });
            return options;
        });

        const filteredPlayers = computed(() => {
            if (!allPlayers.value.length) return [];
            let tempPlayers = [...allPlayers.value];

            // Apply text filters
            if (filters.name) {
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.name &&
                        p.name
                            .toLowerCase()
                            .includes(filters.name.toLowerCase()),
                );
            }
            if (filters.club) {
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.club &&
                        p.club
                            .toLowerCase()
                            .includes(filters.club.toLowerCase()),
                );
            }

            // Apply Transfer Value Filter
            if (filters.transferValue) {
                let operator = "includes";
                let compareValueNum = 0;
                let filterValStr = filters.transferValue;

                if (filterValStr.startsWith(">")) {
                    operator = ">";
                    filterValStr = filterValStr.substring(1);
                } else if (filterValStr.startsWith("<")) {
                    operator = "<";
                    filterValStr = filterValStr.substring(1);
                }

                if (operator !== "includes")
                    compareValueNum = parseMonetaryValue(filterValStr);

                tempPlayers = tempPlayers.filter((p) => {
                    const playerValueNum = p.transferValueAmount || 0;
                    if (operator === ">")
                        return playerValueNum > compareValueNum;
                    if (operator === "<")
                        return playerValueNum < compareValueNum;
                    // 'includes' logic for string search on original transfer_value
                    return String(p.transfer_value || "")
                        .toLowerCase()
                        .includes(filters.transferValue.toLowerCase());
                });
            }

            // Apply Position Filter
            if (filters.position) {
                const selectedFilter = filters.position;
                if (positionGroups[selectedFilter]) {
                    // It's a group filter
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.positionGroups &&
                            p.positionGroups.includes(selectedFilter),
                    );
                } else {
                    // It's a specific position filter
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.parsedPositions &&
                            p.parsedPositions.includes(selectedFilter),
                    );
                }
            }

            // Apply sorting
            if (sortState.key) {
                return sortPlayersLogic([...tempPlayers]); // Use a separate sorting function
            }
            return tempPlayers;
        });

        const sortPlayersLogic = (playersToSort) => {
            // Renamed to avoid conflict
            if (!sortState.key) return playersToSort;
            return [...playersToSort].sort((a, b) => {
                let valA, valB;
                if (
                    ["PHY", "SHO", "PAS", "DRI", "DEF", "MEN"].includes(
                        sortState.key,
                    ) ||
                    ["transferValueAmount", "wageAmount", "age"].includes(
                        sortState.key,
                    )
                ) {
                    valA = a[sortState.key];
                    valB = b[sortState.key];
                } else {
                    // name, club, position (original string)
                    valA = String(a[sortState.key] || "").toLowerCase();
                    valB = String(b[sortState.key] || "").toLowerCase();
                }

                if (valA == null && valB == null) return 0;
                if (valA == null) return sortState.direction === "asc" ? 1 : -1;
                if (valB == null) return sortState.direction === "asc" ? -1 : 1;

                if (typeof valA === "number" && typeof valB === "number") {
                    return sortState.direction === "asc"
                        ? valA - valB
                        : valB - valA;
                }
                if (valA < valB) return sortState.direction === "asc" ? -1 : 1;
                if (valA > valB) return sortState.direction === "asc" ? 1 : -1;
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
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                const response = await playerService.uploadPlayerFile(formData);
                allPlayers.value = processPlayerData(response);
                // console.log("Processed players (first 3):", allPlayers.value.slice(0,3).map(p => ({name: p.name, pos: p.position, parsed: p.parsedPositions, groups: p.positionGroups })));

                sortState.key = null;
                sortState.direction = "asc";
                clearAllFilters(); // Clear all filters on new upload
            } catch (err) {
                error.value = `Failed to parse player data: ${err.message || "Unknown error"}`;
                allPlayers.value = [];
            } finally {
                loading.value = false;
            }
        };

        const handleSort = (sortParams) => {
            sortState.key = sortParams.key;
            sortState.direction = sortParams.direction;
            sortState.isAttribute = sortParams.isAttribute; // Though less relevant now for FIFA stats
            sortState.displayField = sortParams.displayField;
        };

        const clearAllFilters = () => {
            filters.name = "";
            filters.club = "";
            filters.transferValue = "";
            filters.position = null; // Reset position filter
        };

        const handleSearch = () => {
            /* Filtering is reactive via computed props */
        };

        return {
            playerFile,
            loading,
            error,
            allPlayers,
            filteredPlayers,
            uniqueClubsCount,
            uniqueParsedPositionsCount, // Updated stat counts
            filters,
            hasActiveFilters,
            positionFilterOptions, // Added for position filter
            uploadAndParse,
            handleSort,
            handleSearch,
            clearAllFilters, // Updated clear function
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
