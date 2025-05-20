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
                        <li>Click "Upload and Parse" to process the file.</li>
                        <li>
                            The table will now display core player info and
                            FIFA-style aggregated stats.
                        </li>
                        <li>
                            Use search, sort, and pagination to explore the
                            data.
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
                        <div class="col-12 col-sm-6 col-md-3 flex items-end">
                            <div class="row q-col-gutter-sm full-width">
                                <div class="col">
                                    <q-btn
                                        color="grey"
                                        label="Clear Filters"
                                        class="full-width"
                                        @click="clearSearch"
                                        :disable="!hasActiveFilters"
                                    />
                                </div>
                            </div>
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
                                <div class="text-h6">{{ uniqueClubs }}</div>
                                <div class="text-subtitle2">Unique Clubs</div>
                            </q-card-section>
                        </q-card>
                    </div>
                    <div class="col-12 col-md-3">
                        <q-card class="text-center">
                            <q-card-section>
                                <div class="text-h6">{{ uniquePositions }}</div>
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
import { ref, computed, reactive } from "vue"; // Removed 'watch' as it's not used directly here now
import PlayerDataTable from "../components/PlayerDataTable.vue";
import playerService from "../services/playerService";

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

        // Sorting state
        const sortState = reactive({
            key: null,
            direction: "asc",
            isAttribute: false, // Will be true for FIFA stats as well for styling/sorting logic
            displayField: null,
        });

        // Search filters
        const filters = reactive({
            name: "",
            club: "",
            transferValue: "",
        });

        // Check if any filters are active
        const hasActiveFilters = computed(() => {
            return (
                filters.name !== "" ||
                filters.club !== "" ||
                filters.transferValue !== ""
            );
        });

        // Calculate unique clubs
        const uniqueClubs = computed(() => {
            const clubs = new Set();
            allPlayers.value.forEach((player) => {
                if (player.club) clubs.add(player.club);
            });
            return clubs.size;
        });

        // Calculate unique positions
        const uniquePositions = computed(() => {
            const positions = new Set();
            allPlayers.value.forEach((player) => {
                if (player.position) positions.add(player.position);
            });
            return positions.size;
        });

        // Helper to parse monetary values (€1.5M, £500K, etc.) into integers for sorting
        const parseMonetaryValue = (valueStr) => {
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

        // --- START: FIFA Stats Calculation Logic ---
        const fifaStatCategories = {
            PHY: ["Acc", "Pac", "Bal", "Jum", "Nat", "Sta", "Str"],
            SHO: ["Fin", "Lon", "Pen", "Tec", "Cmp", "OtB", "Hea"], // Tec is also in PAS & DRI
            PAS: ["Tec", "Cor", "Cro", "Fre", "L Th", "Pas", "Vis"], // Tec is also in SHO & DRI, Vis in MEN
            DRI: ["Dri", "Fir", "Tec", "Bal", "Agi", "Fla"], // Tec is also in SHO & PAS, Bal in PHY
            DEF: ["Pos", "Tck", "Hea", "Mar", "Cnt", "Ant"], // Hea is also in SHO, Pos & Ant in MEN
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
            ], // Ant, Cmp, OtB, Pos, Vis are in other categories
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
            // Multiply by 5 (0-20 to 0-100), then divide by count for mean
            return Math.round((sum * 5) / count);
        };
        // --- END: FIFA Stats Calculation Logic ---

        // Process player data - convert string values to appropriate types and calculate FIFA stats
        const processPlayerData = (players) => {
            return players.map((player) => {
                const transferValue = parseMonetaryValue(player.transfer_value);
                const wageValue = parseMonetaryValue(player.wage);

                // Ensure original attributes are numbers
                const numericAttributes = {};
                if (player.attributes) {
                    Object.keys(player.attributes).forEach((key) => {
                        const value = player.attributes[key];
                        if (value && !isNaN(parseInt(value, 10))) {
                            numericAttributes[key] = parseInt(value, 10);
                        } else {
                            numericAttributes[key] = 0; // Default to 0 if not a number or missing
                        }
                    });
                }

                const processedPlayer = {
                    ...player,
                    age: parseInt(player.age, 10) || 0,
                    transferValueAmount: transferValue,
                    wageAmount: wageValue,
                    attributes: numericAttributes, // Use the numerically converted attributes
                    // Calculate and add FIFA stats
                    PHY: calculateFifaStat(numericAttributes, "PHY"),
                    SHO: calculateFifaStat(numericAttributes, "SHO"),
                    PAS: calculateFifaStat(numericAttributes, "PAS"),
                    DRI: calculateFifaStat(numericAttributes, "DRI"),
                    DEF: calculateFifaStat(numericAttributes, "DEF"),
                    MEN: calculateFifaStat(numericAttributes, "MEN"),
                };
                return processedPlayer;
            });
        };

        // Filtered players based on search criteria
        const filteredPlayers = computed(() => {
            if (!allPlayers.value.length) return [];

            const filtered = allPlayers.value.filter((player) => {
                const nameMatch =
                    !filters.name ||
                    (player.name &&
                        player.name
                            .toLowerCase()
                            .includes(filters.name.toLowerCase()));
                const clubMatch =
                    !filters.club ||
                    (player.club &&
                        player.club
                            .toLowerCase()
                            .includes(filters.club.toLowerCase()));

                let transferValueMatch = true;
                if (filters.transferValue) {
                    let compareValue = 0;
                    let operator = "includes";
                    if (filters.transferValue.startsWith(">")) {
                        operator = ">";
                        compareValue = parseMonetaryValue(
                            filters.transferValue.substring(1),
                        );
                    } else if (filters.transferValue.startsWith("<")) {
                        operator = "<";
                        compareValue = parseMonetaryValue(
                            filters.transferValue.substring(1),
                        );
                    } else {
                        operator = "includes";
                        const playerValueStr = String(
                            player.transfer_value || "",
                        ).toLowerCase();
                        transferValueMatch = playerValueStr.includes(
                            filters.transferValue.toLowerCase(),
                        );
                        return nameMatch && clubMatch && transferValueMatch;
                    }
                    const playerValue = player.transferValueAmount || 0;
                    if (operator === ">")
                        transferValueMatch = playerValue > compareValue;
                    else if (operator === "<")
                        transferValueMatch = playerValue < compareValue;
                }
                return nameMatch && clubMatch && transferValueMatch;
            });

            if (sortState.key) {
                return sortPlayers([...filtered]);
            }
            return filtered;
        });

        // Upload and parse the player data
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
                const processedPlayers = processPlayerData(response);
                allPlayers.value = processedPlayers;

                // Log first player to check FIFA stats
                // if (processedPlayers.length > 0) {
                //   console.log("First processed player with FIFA stats:", processedPlayers[0]);
                // }

                sortState.key = null;
                sortState.direction = "asc";
                sortState.isAttribute = false;
                sortState.displayField = null;
                clearSearch();
            } catch (err) {
                error.value = `Failed to parse player data: ${err.message || "Unknown error"}`;
                allPlayers.value = [];
            } finally {
                loading.value = false;
            }
        };

        // Handle sorting
        const handleSort = (sortParams) => {
            sortState.key = sortParams.key;
            sortState.direction = sortParams.direction;
            sortState.isAttribute = sortParams.isAttribute;
            sortState.displayField = sortParams.displayField;
        };

        // Sort players based on current sort state
        const sortPlayers = (playersToSort) => {
            if (!sortState.key) return playersToSort;

            // console.log(`Sorting by: ${sortState.key} (${sortState.direction}), isAttribute: ${sortState.isAttribute}`);

            return playersToSort.sort((a, b) => {
                let valA, valB;

                // Check if the sort key is one of the FIFA stats or a direct player property
                if (
                    ["PHY", "SHO", "PAS", "DRI", "DEF", "MEN"].includes(
                        sortState.key,
                    )
                ) {
                    valA = a[sortState.key]; // These are now direct properties on the player object
                    valB = b[sortState.key];
                } else if (
                    sortState.isAttribute &&
                    a.attributes &&
                    b.attributes
                ) {
                    // Fallback for original attributes if any were kept
                    valA = a.attributes[sortState.key];
                    valB = b.attributes[sortState.key];
                } else {
                    // Direct properties like name, age, transferValueAmount, etc.
                    valA = a[sortState.key];
                    valB = b[sortState.key];
                }

                // Handle null/undefined values
                if (valA == null && valB == null) return 0;
                if (valA == null) return sortState.direction === "asc" ? 1 : -1;
                if (valB == null) return sortState.direction === "asc" ? -1 : 1;

                // Numeric comparison
                if (typeof valA === "number" && typeof valB === "number") {
                    return sortState.direction === "asc"
                        ? valA - valB
                        : valB - valA;
                }

                // String comparison (fallback)
                valA = String(valA).toLowerCase();
                valB = String(valB).toLowerCase();
                if (valA < valB) return sortState.direction === "asc" ? -1 : 1;
                if (valA > valB) return sortState.direction === "asc" ? 1 : -1;
                return 0;
            });
        };

        // Clear search filters
        const clearSearch = () => {
            filters.name = "";
            filters.club = "";
            filters.transferValue = "";
        };

        // Handle search (nothing needed here - computed property does the filtering)
        const handleSearch = () => {};

        return {
            playerFile,
            loading,
            error,
            allPlayers,
            filteredPlayers,
            uniqueClubs,
            uniquePositions,
            filters,
            hasActiveFilters,
            uploadAndParse,
            handleSort,
            handleSearch,
            clearSearch,
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
