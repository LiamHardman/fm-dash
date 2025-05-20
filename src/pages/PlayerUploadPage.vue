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
                        <li>Place <code>attribute_weights.json</code> in your project's <code>public</code> folder.</li>
                        <li>Select an HTML file. Click "Upload and Parse".</li>
                        <li>
                            The table will now include Nationality and Overall
                            rating columns. The summarized stats (PHY, SHO, etc.)
                            will be calculated using weights from the JSON file.
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
                        :disable="!playerFile || !weightsLoaded"
                        @click="uploadAndParse"
                    >
                         <q-tooltip v-if="!weightsLoaded">Waiting for attribute weights to load...</q-tooltip>
                    </q-btn>
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
             <q-banner v-if="weightsError" class="bg-warning text-dark q-mb-md" rounded>
                Could not load attribute weights from <code>attribute_weights.json</code>. Using default calculation.
                Error: {{ weightsError }}
                <template v-slot:action>
                    <q-btn flat color="dark" label="Retry" @click="loadAttributeWeights" :loading="loadingWeights" />
                </template>
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
import { ref, computed, reactive, onMounted } from "vue"; // Import onMounted
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import playerService from "../services/playerService";

// Position Parsing Logic (remains the same as your provided code)
const positionRoleMap = {
    GK: "Goalkeeper", SW: "Sweeper", D: "Defender", WB: "Wing-Back", DM: "Defensive Midfielder", M: "Midfielder", AM: "Attacking Midfielder", ST: "Striker", F: "Forward",
};
const positionSideMap = { R: "Right", L: "Left", C: "Centre" };
const standardizedPositionNameMap = {
    "Goalkeeper (Centre)": "Goalkeeper", Goalkeeper: "Goalkeeper", "Sweeper (Centre)": "Sweeper", Sweeper: "Sweeper", "Defender (Right)": "Right Back", "Defender (Left)": "Left Back", "Defender (Centre)": "Centre Back", "Wing-Back (Right)": "Right Wing-Back", "Wing-Back (Left)": "Left Wing-Back", "Wing-Back (Centre)": "Centre Wing-Back", "Defensive Midfielder (Right)": "Right Defensive Midfielder", "Defensive Midfielder (Left)": "Left Defensive Midfielder", "Defensive Midfielder (Centre)": "Centre Defensive Midfielder", "Midfielder (Right)": "Right Midfielder", "Midfielder (Left)": "Left Midfielder", "Midfielder (Centre)": "Centre Midfielder", "Attacking Midfielder (Right)": "Right Attacking Midfielder", "Attacking Midfielder (Left)": "Left Attacking Midfielder", "Attacking Midfielder (Centre)": "Centre Attacking Midfielder", "Striker (Centre)": "Striker", Striker: "Striker", "Forward (Right)": "Right Forward", "Forward (Left)": "Left Forward", "Forward (Centre)": "Centre Forward",
};
const positionGroups = {
    Goalkeepers: ["Goalkeeper"], Defenders: ["Sweeper", "Right Back", "Left Back", "Centre Back", "Right Wing-Back", "Left Wing-Back", "Centre Wing-Back"], Midfielders: ["Right Defensive Midfielder", "Left Defensive Midfielder", "Centre Defensive Midfielder", "Right Midfielder", "Left Midfielder", "Centre Midfielder", "Right Attacking Midfielder", "Left Attacking Midfielder", "Centre Attacking Midfielder"], Attackers: ["Striker", "Right Forward", "Left Forward", "Centre Forward"],
};
function parsePlayerPositions(positionStr) {
    if (!positionStr || typeof positionStr !== "string") return [];
    const parsedPositions = new Set();
    const parts = positionStr.split(",").map((p) => p.trim()).filter((p) => p);
    parts.forEach((part) => {
        const match = part.match(/^([A-Z]+)(?:\s*\(([A-Z]+)\))?$/);
        if (!match) return;
        const roleKey = match[1]; const sideKeysStr = match[2]; const roleFullName = positionRoleMap[roleKey];
        if (!roleFullName) return;
        if (sideKeysStr) {
            for (const sideKey of sideKeysStr) {
                const sideFullName = positionSideMap[sideKey];
                if (sideFullName) {
                    const detailedName = `${roleFullName} (${sideFullName})`;
                    parsedPositions.add(standardizedPositionNameMap[detailedName] || detailedName);
                }
            }
        } else {
            if (["D", "DM", "M", "AM", "ST", "F", "WB"].includes(roleKey)) {
                const detailedName = `${roleFullName} (Centre)`;
                parsedPositions.add(standardizedPositionNameMap[detailedName] || detailedName);
            } else if (["GK", "SW"].includes(roleKey)) {
                parsedPositions.add(standardizedPositionNameMap[roleFullName] || roleFullName);
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
            if (positionGroups[groupName].includes(pos)) groups.add(groupName);
        }
    });
    return Array.from(groups);
}

// Default weights in case the JSON file fails to load
const defaultFifaStatWeights = {
    PHY: { Acc: 7, Pac: 6, Str: 5, Sta: 4, Nat: 3, Bal: 2, Jum: 1 },
    SHO: { Fin: 7, OtB: 6, Cmp: 5, Tec: 4, Hea: 3, Lon: 2, Pen: 1 },
    PAS: { Pas: 7, Vis: 6, Tec: 5, Cro: 4, Fre: 3, Cor: 2, "L Th": 1 },
    DRI: { Dri: 6, Fir: 5, Tec: 4, Agi: 3, Bal: 2, Fla: 1 },
    DEF: { Tck: 6, Mar: 5, Hea: 4, Pos: 3, Cnt: 2, Ant: 1 },
    MEN: { Wor: 11, Dec: 10, Tea: 9, Det: 8, Bra: 7, Ldr: 6, Vis: 5, Agg: 4, OtB: 3, Pos: 2, Ant: 1 }
};


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

        // START: Reactive variable for weights and loading state
        const fifaStatWeights = ref({}); // Will hold weights from JSON
        const weightsLoaded = ref(false);
        const weightsError = ref("");
        const loadingWeights = ref(false);
        // END: Reactive variable for weights

        const sortState = reactive({ key: null, direction: "asc", isAttribute: false, displayField: null });
        const filters = reactive({ name: "", club: "", transferValue: "", position: null, nationality: "" });

        const hasActiveFilters = computed(() => filters.name !== "" || filters.club !== "" || filters.transferValue !== "" || filters.position !== null || filters.nationality !== "");
        const uniqueClubsCount = computed(() => new Set(allPlayers.value.map((p) => p.club).filter(Boolean)).size);
        const uniqueParsedPositionsCount = computed(() => {
            const p = new Set();
            allPlayers.value.forEach((player) => player.parsedPositions?.forEach((pos) => p.add(pos)));
            return p.size;
        });
        const uniqueNationalitiesCount = computed(() => new Set(allPlayers.value.map((p) => p.nationality).filter(Boolean)).size);

        const parseMonetaryValue = (valueStr) => {
            if (typeof valueStr !== "string" || !valueStr) return 0;
            const c = valueStr.split(" p/w")[0]; // Remove " p/w" if present
            let m = 1; // Multiplier
            const l = c.toLowerCase();
            if (l.includes("m")) m = 1000000;
            else if (l.includes("k")) m = 1000;
            
            // Clean the string to keep only numbers, decimal point, and comma for thousands
            let n = c.replace(/[^0-9,.]/g, "");

            // Handle European-style decimal (comma) and thousands (dot) if necessary,
            // or assume standard US/UK style. This simplified version assumes US/UK.
            // If commas are used as thousand separators:
            n = n.replace(/,/g, ""); // Remove all commas
            
            const v = parseFloat(n);
            return Math.round(isNaN(v) ? 0 : v * m);
        };
        
        // START: Updated calculateFifaStat function to use reactive weights
        const calculateFifaStat = (attributes, categoryName) => {
            // Use loaded weights, or fallback to default if not loaded or error
            const currentWeights = weightsLoaded.value && Object.keys(fifaStatWeights.value).length > 0 
                                   ? fifaStatWeights.value 
                                   : defaultFifaStatWeights;

            const categoryWeights = currentWeights[categoryName];
            if (!categoryWeights) {
                // console.warn(`No weights defined for category: ${categoryName}`);
                return 0; // Return 0 if no weights for category (even in default)
            }

            let weightedSum = 0;
            let totalWeightOfPresentAttributes = 0;

            // Iterate over the attributes defined in our weights configuration for this category
            for (const attrName in categoryWeights) {
                // Check if the player actually has this attribute and it's a number
                if (attributes.hasOwnProperty(attrName)) {
                    const attrValue = parseInt(attributes[attrName], 10);
                    const attrWeight = categoryWeights[attrName];

                    // Ensure the attribute value is a valid number (typically 0-20 for FM stats)
                    if (!isNaN(attrValue) && attrValue >= 0 && attrValue <= 20) {
                        weightedSum += attrValue * attrWeight;
                        totalWeightOfPresentAttributes += attrWeight;
                    }
                }
            }

            if (totalWeightOfPresentAttributes === 0) {
                return 0; // Avoid division by zero if no relevant attributes are present or valid
            }

            // Calculate the weighted average (on a 0-20 scale)
            const weightedAverage = weightedSum / totalWeightOfPresentAttributes;
            
            // Scale to 0-100 (FIFA-like) and round
            return Math.round(weightedAverage * 5.25);
        };
        // END: Updated calculateFifaStat function

        // START: Function to load weights from JSON
        const loadAttributeWeights = async () => {
            loadingWeights.value = true;
            weightsError.value = "";
            try {
                // Assuming attribute_weights.json is in the public folder
                const response = await fetch('/attribute_weights.json'); // Path relative to the public folder
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                fifaStatWeights.value = data;
                weightsLoaded.value = true;
                // console.log("Attribute weights loaded successfully:", data);
            } catch (e) {
                console.error("Failed to load attribute weights:", e);
                weightsError.value = e.message || "Unknown error loading weights.";
                fifaStatWeights.value = { ...defaultFifaStatWeights }; // Fallback to defaults on error
                weightsLoaded.value = false; // Indicate that loading failed, but we have defaults
            } finally {
                loadingWeights.value = false;
            }
        };

        // Load weights when the component is mounted
        onMounted(() => {
            loadAttributeWeights();
        });
        // END: Function to load weights

        const processPlayerData = (players) => {
            return players.map((player) => {
                const transferValue = parseMonetaryValue(player.transfer_value);
                const wageValue = parseMonetaryValue(player.wage);
                
                // Ensure attributes are numeric for calculation
                const numericAttributes = {};
                if (player.attributes) {
                    Object.keys(player.attributes).forEach((key) => {
                        const value = player.attributes[key];
                        // Default to 0 if attribute is not a number or missing,
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

                // Calculate FIFA-style stats using the new weighted method
                const phy = calculateFifaStat(numericAttributes, "PHY");
                const sho = calculateFifaStat(numericAttributes, "SHO");
                const pas = calculateFifaStat(numericAttributes, "PAS");
                const dri = calculateFifaStat(numericAttributes, "DRI");
                const def = calculateFifaStat(numericAttributes, "DEF");
                const men = calculateFifaStat(numericAttributes, "MEN");

                const fifaStatsForOverall = [phy, sho, pas, dri, def, men];
                const validStats = fifaStatsForOverall.filter(s => typeof s === 'number' && !isNaN(s));
                const overallSum = validStats.reduce((acc, current) => acc + current, 0);
                const overall = validStats.length > 0 ? Math.round(overallSum / validStats.length) : 0;
                
                return {
                    ...player,
                    age: parseInt(player.age, 10) || 0,
                    transferValueAmount: transferValue,
                    wageAmount: wageValue,
                    attributes: numericAttributes, // Store the numeric attributes
                    PHY: phy,
                    SHO: sho,
                    PAS: pas,
                    DRI: dri,
                    DEF: def,
                    MEN: men,
                    Overall: overall, 
                    parsedPositions: parsedPlayerPositions,
                    positionGroups: playerPosGroups,
                };
            });
        };

        const positionFilterOptions = computed(() => {
            const o = [];
            Object.keys(positionGroups).forEach((g) => o.push({ label: `${g} (Group)`, value: g }));
            const s = new Set();
            allPlayers.value.forEach((p) => p.parsedPositions?.forEach((pos) => s.add(pos)));
            Array.from(s).sort().forEach((pos) => { if (!positionGroups[pos]) o.push({ label: pos, value: pos }); });
            return o;
        });

        const filteredPlayers = computed(() => {
            if (!allPlayers.value.length) return [];
            let tempPlayers = [...allPlayers.value];
            if (filters.name) tempPlayers = tempPlayers.filter((p) => p.name && p.name.toLowerCase().includes(filters.name.toLowerCase()));
            if (filters.club) tempPlayers = tempPlayers.filter((p) => p.club && p.club.toLowerCase().includes(filters.club.toLowerCase()));
            if (filters.transferValue) {
                let o = "includes"; let c = 0; let f = filters.transferValue;
                if (f.startsWith(">")) { o = ">"; f = f.substring(1); } 
                else if (f.startsWith("<")) { o = "<"; f = f.substring(1); }
                if (o !== "includes") c = parseMonetaryValue(f);
                tempPlayers = tempPlayers.filter((p) => {
                    const v = p.transferValueAmount || 0;
                    if (o === ">") return v > c; if (o === "<") return v < c;
                    return String(p.transfer_value || "").toLowerCase().includes(filters.transferValue.toLowerCase());
                });
            }
            if (filters.position) {
                const s = filters.position;
                if (positionGroups[s]) tempPlayers = tempPlayers.filter((p) => p.positionGroups && p.positionGroups.includes(s));
                else tempPlayers = tempPlayers.filter((p) => p.parsedPositions && p.parsedPositions.includes(s));
            }
            if (filters.nationality) tempPlayers = tempPlayers.filter((p) => p.nationality && p.nationality.toLowerCase().includes(filters.nationality.toLowerCase()));
            if (sortState.key) return sortPlayersLogic([...tempPlayers]);
            return tempPlayers;
        });

        const sortPlayersLogic = (playersToSort) => {
            if (!sortState.key) return playersToSort;
            const k = sortState.isAttribute ? sortState.key : allPlayers.value.length > 0 && Object.keys(allPlayers.value[0]).includes(sortState.key + "Amount") ? sortState.key + "Amount" : sortState.key;
            return [...playersToSort].sort((a, b) => {
                let vA, vB;
                if (sortState.isAttribute) { vA = a.attributes ? a.attributes[k] : null; vB = b.attributes ? b.attributes[k] : null; } 
                else { vA = a[k]; vB = b[k]; }
                if (vA == null && vB == null) return 0;
                if (vA == null) return sortState.direction === "asc" ? 1 : -1;
                if (vB == null) return sortState.direction === "asc" ? -1 : 1;
                if (typeof vA === "number" && typeof vB === "number") return sortState.direction === "asc" ? vA - vB : vB - vA;
                vA = String(vA).toLowerCase(); vB = String(vB).toLowerCase();
                if (vA < vB) return sortState.direction === "asc" ? -1 : 1;
                if (vA > vB) return sortState.direction === "asc" ? 1 : -1;
                return 0;
            });
        };
        const uploadAndParse = async () => {
            if (!playerFile.value) { error.value = "Please select an HTML file first."; return; }
            if (!weightsLoaded.value && weightsError.value) { // If loading failed and we are using defaults, allow parsing.
                 // console.warn("Parsing with default weights due to loading error.");
            } else if (!weightsLoaded.value) { // Still loading, or failed silently before defaults were set
                error.value = "Attribute weights are not yet loaded. Please wait or check console for errors.";
                return;
            }

            loading.value = true; error.value = "";
            try {
                const f = new FormData(); f.append("playerFile", playerFile.value);
                const r = await playerService.uploadPlayerFile(f);
                allPlayers.value = processPlayerData(r);
                sortState.key = null; sortState.direction = "asc"; clearAllFilters();
            } catch (e) {
                error.value = `Failed to parse player data: ${e.message || "Unknown error"}`;
                allPlayers.value = [];
            } finally { loading.value = false; }
        };
        const handleSort = (sortParams) => {
            sortState.key = sortParams.key; sortState.direction = sortParams.direction;
            sortState.isAttribute = sortParams.isAttribute; sortState.displayField = sortParams.displayField;
        };
        const clearAllFilters = () => {
            filters.name = ""; filters.club = ""; filters.transferValue = ""; filters.position = null; filters.nationality = "";
        };
        const handleSearch = () => { /* Reactive filtering, computed property handles it */ };
        const handlePlayerSelected = (player) => { selectedPlayer.value = player; showPlayerDetailDialog.value = true; };

        return {
            playerFile, loading, error, allPlayers, filteredPlayers, uniqueClubsCount, uniqueParsedPositionsCount, uniqueNationalitiesCount,
            filters, hasActiveFilters, positionFilterOptions, uploadAndParse, handleSort, handleSearch, clearAllFilters,
            selectedPlayer, showPlayerDetailDialog, handlePlayerSelected,
            weightsLoaded, weightsError, loadingWeights, loadAttributeWeights // Expose new refs and method
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
