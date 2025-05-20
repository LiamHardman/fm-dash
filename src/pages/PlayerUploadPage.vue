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
                        <li>Place <code>attribute_weights.json</code> (for PHY, SHO, etc.) and <code>role_specific_overall_weights.json</code> (for Overall rating) in your project's <code>public</code> folder.</li>
                        <li>Select an HTML file. Click "Upload and Parse".</li>
                        <li>
                            The table will now include Nationality and Overall
                            rating columns. The Overall rating is now calculated based on role-specific attribute weights.
                        </li>
                        <li>
                            Use filters for Name, Club, Position, Nationality,
                            and Transfer Value.
                        </li>
                        <li>Click on any player row for a detailed view, which will show all calculated role-specific overalls.</li>
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
                        :disable="!playerFile || !attributeWeightsLoaded || !roleSpecificOverallWeightsLoaded"
                        @click="uploadAndParse"
                    >
                         <q-tooltip v-if="!attributeWeightsLoaded || !roleSpecificOverallWeightsLoaded">Waiting for attribute weights to load...</q-tooltip>
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
             <q-banner v-if="attributeWeightsError" class="bg-warning text-dark q-mb-md" rounded>
                Could not load <code>attribute_weights.json</code> (for PHY, SHO etc.). Using default calculation.
                Error: {{ attributeWeightsError }}
                <template v-slot:action>
                    <q-btn flat color="dark" label="Retry" @click="loadAttributeWeights" :loading="loadingAttributeWeights" />
                </template>
            </q-banner>
            <q-banner v-if="roleSpecificOverallWeightsError" class="bg-warning text-dark q-mb-md" rounded>
                Could not load <code>role_specific_overall_weights.json</code>. Overall rating might be 0 or based on defaults.
                Error: {{ roleSpecificOverallWeightsError }}
                <template v-slot:action>
                    <q-btn flat color="dark" label="Retry" @click="loadRoleSpecificOverallWeights" :loading="loadingRoleSpecificOverallWeights" />
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
import { ref, computed, reactive, onMounted } from "vue";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import playerService from "../services/playerService";

// Position Parsing Logic
const positionRoleMap = { GK: "Goalkeeper", SW: "Sweeper", D: "Defender", WB: "Wing-Back", DM: "Defensive Midfielder", M: "Midfielder", AM: "Attacking Midfielder", ST: "Striker", F: "Forward" };
const positionSideMap = { R: "Right", L: "Left", C: "Centre" };
const standardizedPositionNameMap = { "Goalkeeper (Centre)": "Goalkeeper", Goalkeeper: "Goalkeeper", "Sweeper (Centre)": "Sweeper", Sweeper: "Sweeper", "Defender (Right)": "Right Back", "Defender (Left)": "Left Back", "Defender (Centre)": "Centre Back", "Wing-Back (Right)": "Right Wing-Back", "Wing-Back (Left)": "Left Wing-Back", "Wing-Back (Centre)": "Centre Wing-Back", "Defensive Midfielder (Right)": "Right Defensive Midfielder", "Defensive Midfielder (Left)": "Left Defensive Midfielder", "Defensive Midfielder (Centre)": "Centre Defensive Midfielder", "Midfielder (Right)": "Right Midfielder", "Midfielder (Left)": "Left Midfielder", "Midfielder (Centre)": "Centre Midfielder", "Attacking Midfielder (Right)": "Right Attacking Midfielder", "Attacking Midfielder (Left)": "Left Attacking Midfielder", "Attacking Midfielder (Centre)": "Centre Attacking Midfielder", "Striker (Centre)": "Striker", Striker: "Striker", "Forward (Right)": "Right Forward", "Forward (Left)": "Left Forward", "Forward (Centre)": "Centre Forward" };
const positionGroups = { Goalkeepers: ["Goalkeeper"], Defenders: ["Sweeper", "Right Back", "Left Back", "Centre Back", "Right Wing-Back", "Left Wing-Back", "Centre Wing-Back"], Midfielders: ["Right Defensive Midfielder", "Left Defensive Midfielder", "Centre Defensive Midfielder", "Right Midfielder", "Left Midfielder", "Centre Midfielder", "Right Attacking Midfielder", "Left Attacking Midfielder", "Centre Attacking Midfielder"], Attackers: ["Striker", "Right Forward", "Left Forward", "Centre Forward"] };

function parsePlayerPositions(positionStr) {
    if (!positionStr || typeof positionStr !== "string") return [];
    const finalPositions = new Set();
    const mainParts = positionStr.split(',').map(p => p.trim()).filter(p => p);
    mainParts.forEach(part => {
        const sideMatch = part.match(/^(.*?)(\s*\(([A-Z]+)\))?$/);
        let rolesStringSegment = part; 
        let explicitSidesArray = null;
        if (sideMatch && sideMatch[2]) { 
            rolesStringSegment = sideMatch[1].trim(); 
            explicitSidesArray = sideMatch[3].split(''); 
        }
        const individualRoleKeys = rolesStringSegment.split('/').map(r => r.trim()).filter(r => r);
        individualRoleKeys.forEach(roleKey => {
            const roleFullName = positionRoleMap[roleKey]; 
            if (roleFullName) {
                const sidesToIterate = explicitSidesArray || ['C'];
                sidesToIterate.forEach(sideKey => {
                    const sideFullName = positionSideMap[sideKey]; 
                    if (sideFullName) {
                        const detailedName = `${roleFullName} (${sideFullName})`;
                        finalPositions.add(standardizedPositionNameMap[detailedName] || detailedName);
                    }
                });
            } else if (standardizedPositionNameMap[roleKey]) {
                 finalPositions.add(roleKey);
            }
        });
    });
    return Array.from(finalPositions);
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

// Default weights for attribute_weights.json (PHY, SHO, etc.)
const defaultAttributeWeights = { PHY: { Acc: 7, Pac: 6, Str: 5, Sta: 4, Nat: 3, Bal: 2, Jum: 1 }, SHO: { Fin: 7, OtB: 6, Cmp: 5, Tec: 4, Hea: 3, Lon: 2, Pen: 1 }, PAS: { Pas: 7, Vis: 6, Tec: 5, Cro: 4, Fre: 3, Cor: 2, "L Th": 1 }, DRI: { Dri: 6, Fir: 5, Tec: 4, Agi: 3, Bal: 2, Fla: 1 }, DEF: { Tck: 6, Mar: 5, Hea: 4, Pos: 3, Cnt: 2, Ant: 1 }, MEN: { Wor: 11, Dec: 10, Tea: 9, Det: 8, Bra: 7, Ldr: 6, Vis: 5, Agg: 4, OtB: 3, Pos: 2, Ant: 1 } };

// Default weights for role_specific_overall_weights.json
const defaultRoleSpecificOverallWeights = {
  "DC - BPD": {"Cor": 5, "Cro": 1, "Dri": 40, "Fin": 10, "Fir": 35, "Fre": 10, "Hea": 55, "Lon": 10, "Tea": 20, "L Th": 0, "Mar": 55, "Pas": 55, "Pen": 10, "Tck": 40, "Tec": 35, "Agg": 40, "Ant": 50, "Bra": 30, "Cmp": 80, "Cnt": 50, "Dec": 50, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 55, "Vis": 50, "Wor": 55, "Acc": 90, "Agi": 60, "Bal": 35, "Jum": 65, "Nat": 10, "Pac": 90, "Sta": 30, "Str": 50 },
  "DC - CD": {"Cor": 10, "Cro": 10, "Dri": 30, "Fin": 10, "Fir": 30, "Fre": 5, "Hea": 60, "Lon": 0, "L Th": 0, "Mar": 70, "Pas": 40, "Pen": 0, "Tck": 70, "Tec": 30, "Agg": 60, "Ant": 65, "Bra": 50, "Cmp": 80, "Cnt": 65, "Dec": 65, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 65, "Tea": 20, "Vis": 30, "Wor": 60, "Acc": 60, "Agi": 30, "Bal": 30, "Jum": 65, "Nat": 10, "Pac": 70, "Sta": 40, "Str": 60 },
  "DC - Generic": {"Cor": 10, "Cro": 10, "Dri": 10, "Fin": 10, "Fir": 20, "Fre": 10, "Hea": 50, "Lon": 10, "L Th": 10, "Mar": 80, "Pas": 20, "Pen": 10, "Tck": 50, "Tec": 10, "Agg": 0, "Ant": 50, "Bra": 20, "Cmp": 20, "Cnt": 40, "Dec": 10, "Det": 0, "Fla": 0, "Ldr": 20, "OtB": 10, "Pos": 80, "Tea": 10, "Vis": 10, "Wor": 20, "Acc": 60, "Agi": 60, "Bal": 20, "Jum": 60, "Nat": 0, "Pac": 50, "Sta": 30, "Str": 60 },
  "DC - WCB": {"Cor": 5, "Cro": 30, "Dri": 40, "Fin": 10, "Fir": 50, "Fre": 10, "Hea": 50, "Lon": 10, "L Th": 0, "Mar": 40, "Pas": 50, "Pen": 10, "Tck": 30, "Tec": 30, "Agg": 50, "Ant": 60, "Bra": 30, "Cmp": 65, "Cnt": 50, "Dec": 40, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 10, "Pos": 60, "Tea": 20, "Vis": 60, "Wor": 60, "Acc": 75, "Agi": 65, "Bal": 30, "Jum": 35, "Nat": 30, "Pac": 75, "Sta": 60, "Str": 45 },
  "DR/L - CWB": {"Cor": 10, "Cro": 45, "Dri": 50, "Fin": 20, "Fir": 50, "Fre": 10, "Hea": 15, "Lon": 20, "L Th": 5, "Mar": 25, "Pas": 40, "Pen": 10, "Tck": 30, "Tec": 30, "Agg": 15, "Ant": 40, "Bra": 20, "Cmp": 25, "Cnt": 45, "Dec": 55, "Det": 0, "Fla": 20, "Ldr": 10, "OtB": 10, "Pos": 45, "Tea": 20, "Vis": 35, "Wor": 50, "Acc": 85, "Agi": 60, "Bal": 45, "Jum": 20, "Nat": 30, "Pac": 80, "Sta": 75, "Str": 35 },
  "DR/L - FB": {"Cor": 10, "Cro": 35, "Dri": 30, "Fin": 15, "Fir": 30, "Fre": 5, "Hea": 30, "Lon": 10, "L Th": 10, "Mar": 40, "Pas": 55, "Pen": 5, "Tck": 40, "Tec": 20, "Agg": 15, "Ant": 35, "Bra": 30, "Cmp": 25, "Cnt": 35, "Dec": 50, "Det": 20, "Fla": 15, "Ldr": 15, "OtB": 15, "Pos": 50, "Tea": 55, "Vis": 30, "Wor": 60, "Acc": 50, "Agi": 35, "Bal": 35, "Jum": 20, "Nat": 20, "Pac": 65, "Sta": 65, "Str": 30 },
  "DR/L - Generic": {"Cor": 10, "Cro": 20, "Dri": 10, "Fin": 10, "Fir": 30, "Fre": 10, "Hea": 20, "Lon": 10, "Mar": 30, "Pas": 20, "Pen": 10, "Tck": 40, "Tec": 20, "Agg": 0, "Ant": 30, "Bra": 20, "Cmp": 20, "Cnt": 40, "Dec": 70, "Det": 0, "Fla": 0, "Ldr": 0, "OtB": 10, "Pos": 40, "Vis": 20, "Wor": 20, "Acc": 70, "Agi": 60, "Bal": 20, "Jum": 20, "Nat": 0, "Pac": 50, "Sta": 60, "Str": 40, "Tea": 20, "L Th": 0 },
  "DR/L - IWB": {"Cor": 10, "Cro": 10, "Dri": 35, "Fin": 25, "Fir": 55, "Fre": 5, "Hea": 30, "Lon": 35, "L Th": 10, "Mar": 50, "Pas": 55, "Pen": 5, "Tck": 45, "Tec": 55, "Agg": 40, "Ant": 65, "Bra": 35, "Cmp": 60, "Cnt": 50, "Dec": 55, "Det": 5, "Fla": 5, "Ldr": 5, "OtB": 50, "Pos": 50, "Tea": 20, "Vis": 35, "Wor": 70, "Acc": 65, "Agi": 50, "Bal": 35, "Jum": 30, "Nat": 25, "Pac": 45, "Sta": 70, "Str": 35 },
  "DR/L - WB": {"Cor": 1, "Cro": 50, "Dri": 35, "Fin": 10, "Fir": 35, "Fre": 5, "Hea": 20, "Lon": 10, "L Th": 1, "Mar": 30, "Pas": 35, "Pen": 10, "Tck": 35, "Tec": 30, "Agg": 0, "Ant": 35, "Bra": 20, "Cmp": 25, "Cnt": 35, "Dec": 65, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 0, "Pos": 40, "Tea": 20, "Vis": 20, "Wor": 30, "Acc": 65, "Agi": 55, "Bal": 20, "Jum": 30, "Nat": 0, "Pac": 50, "Sta": 65, "Str": 45 },
  "WBR/L - CWB - Su": {"Cor": 10, "Cro": 45, "Dri": 50, "Fin": 20, "Fir": 50, "Fre": 10, "Hea": 15, "Lon": 20, "L Th": 5, "Mar": 25, "Pas": 40, "Pen": 10, "Tck": 30, "Tec": 30, "Agg": 15, "Ant": 40, "Bra": 20, "Cmp": 25, "Cnt": 45, "Dec": 55, "Det": 0, "Fla": 20, "Ldr": 10, "OtB": 10, "Pos": 45, "Tea": 20, "Vis": 35, "Wor": 50, "Acc": 85, "Agi": 60, "Bal": 45, "Jum": 20, "Nat": 30, "Pac": 80, "Sta": 75, "Str": 35 },
  "WBR/L - Generic": {"Cor": 10, "Cro": 30, "Dri": 20, "Fin": 10, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 10, "Mar": 20, "Pas": 30, "Pen": 10, "Tck": 30, "Tec": 30, "Agg": 0, "Ant": 30, "Bra": 10, "Cmp": 20, "Cnt": 30, "Dec": 50, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 10, "Pos": 50, "Tea": 20, "Vis": 40, "Wor": 40, "Acc": 80, "Agi": 50, "Bal": 20, "Jum": 10, "Nat": 0, "Pac": 60, "Sta": 70, "Str": 40 },
  "WBR/L - IWB - At": {"Cor": 30, "Cro": 25, "Dri": 55, "Fin": 20, "Fir": 30, "Fre": 10, "Hea": 20, "Lon": 35, "L Th": 30, "Mar": 45, "Pas": 45, "Pen": 10, "Tck": 35, "Tec": 45, "Agg": 45, "Ant": 45, "Bra": 20, "Cmp": 30, "Cnt": 45, "Dec": 45, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 70, "Pos": 30, "Tea": 10, "Vis": 25, "Wor": 90, "Acc": 100, "Agi": 60, "Bal": 25, "Jum": 40, "Nat": 10, "Pac": 90, "Sta": 100, "Str": 25 },
  "WBR/L - WB - Su": {"Cor": 0, "Cro": 50, "Dri": 35, "Fin": 10, "Fir": 35, "Fre": 5, "Hea": 20, "Lon": 10, "L Th": 0, "Mar": 30, "Pas": 35, "Pen": 10, "Tck": 35, "Tec": 30, "Agg": 0, "Ant": 35, "Bra": 20, "Cmp": 25, "Cnt": 35, "Dec": 65, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 0, "Pos": 40, "Tea": 20, "Vis": 20, "Wor": 30, "Acc": 65, "Agi": 55, "Bal": 20, "Jum": 30, "Nat": 0, "Pac": 50, "Sta": 65, "Str": 45 },
  "DM - DLP - De": {"Cor": 15, "Cro": 10, "Dri": 30, "Fin": 15, "Fir": 60, "Fre": 0, "Hea": 10, "Lon": 20, "L Th": 0, "Mar": 25, "Pas": 55, "Pen": 15, "Tck": 40, "Tec": 55, "Agg": 20, "Ant": 55, "Bra": 10, "Cmp": 35, "Cnt": 35, "Dec": 75, "Det": 0, "Fla": 20, "Ldr": 10, "OtB": 55, "Pos": 70, "Tea": 25, "Vis": 55, "Wor": 40, "Acc": 40, "Agi": 20, "Bal": 10, "Jum": 0, "Nat": 30, "Pac": 25, "Sta": 40, "Str": 35 },
  "DM - DLP - Su": {"Cor": 15, "Cro": 10, "Dri": 30, "Fin": 15, "Fir": 60, "Fre": 0, "Hea": 10, "Lon": 20, "L Th": 0, "Mar": 25, "Pas": 65, "Pen": 15, "Tck": 30, "Tec": 60, "Agg": 20, "Ant": 55, "Bra": 10, "Cmp": 35, "Cnt": 35, "Dec": 75, "Det": 0, "Fla": 20, "Ldr": 10, "OtB": 55, "Pos": 70, "Tea": 25, "Vis": 55, "Wor": 35, "Acc": 35, "Agi": 20, "Bal": 10, "Jum": 0, "Nat": 30, "Pac": 30, "Sta": 40, "Str": 35 },
  "DM - DM - De": {"Cor": 20, "Cro": 20, "Dri": 30, "Fin": 15, "Fir": 55, "Fre": 5, "Hea": 15, "Lon": 50, "L Th": 5, "Mar": 55, "Pas": 50, "Pen": 5, "Tck": 45, "Tec": 50, "Agg": 55, "Ant": 60, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 70, "Det": 10, "Fla": 10, "Ldr": 20, "OtB": 50, "Pos": 35, "Tea": 50, "Vis": 70, "Wor": 70, "Acc": 60, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 45, "Sta": 70, "Str": 45 },
  "DM - DM - Su": {"Cor": 20, "Cro": 35, "Dri": 35, "Fin": 25, "Fir": 55, "Fre": 5, "Hea": 15, "Lon": 50, "L Th": 5, "Mar": 15, "Pas": 65, "Pen": 5, "Tck": 35, "Tec": 55, "Agg": 30, "Ant": 55, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 70, "Det": 10, "Fla": 10, "Ldr": 20, "OtB": 50, "Pos": 35, "Tea": 50, "Vis": 75, "Wor": 70, "Acc": 55, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 50, "Sta": 70, "Str": 45 },
  "DM - Generic": {"Cor": 10, "Cro": 10, "Dri": 20, "Fin": 20, "Fir": 40, "Fre": 10, "Hea": 10, "Lon": 30, "L Th": 10, "Mar": 30, "Pas": 40, "Pen": 10, "Tck": 70, "Tec": 30, "Agg": 0, "Ant": 50, "Bra": 10, "Cmp": 20, "Cnt": 30, "Dec": 80, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 10, "Pos": 50, "Tea": 20, "Vis": 40, "Wor": 60, "Acc": 60, "Agi": 20, "Bal": 10, "Jum": 0, "Nat": 40, "Pac": 40, "Sta": 40, "Str": 50 },
  "DM - RPM - Su": {"Cor": 10, "Cro": 10, "Dri": 45, "Fin": 20, "Fir": 50, "Fre": 30, "Hea": 10, "Lon": 40, "L Th": 5, "Mar": 20, "Pas": 65, "Pen": 10, "Tck": 35, "Tec": 50, "Agg": 50, "Ant": 55, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 65, "Det": 20, "Fla": 50, "Ldr": 10, "OtB": 40, "Pos": 65, "Tea": 10, "Vis": 55, "Wor": 90, "Acc": 65, "Agi": 45, "Bal": 35, "Jum": 15, "Nat": 10, "Pac": 70, "Sta": 70, "Str": 35 },
  "DM - VOL - Su": {"Cor": 20, "Cro": 20, "Dri": 40, "Fin": 35, "Fir": 50, "Fre": 5, "Hea": 15, "Lon": 55, "L Th": 5, "Mar": 55, "Pas": 40, "Pen": 5, "Tck": 40, "Tec": 55, "Agg": 55, "Ant": 60, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 70, "Det": 10, "Fla": 30, "Ldr": 20, "OtB": 35, "Pos": 55, "Tea": 50, "Vis": 50, "Wor": 70, "Acc": 50, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 60, "Sta": 70, "Str": 55 },
  "MC - AP - Att": {"Cor": 30, "Cro": 35, "Dri": 55, "Fin": 35, "Fir": 65, "Fre": 15, "Hea": 15, "Lon": 30, "L Th": 5, "Mar": 15, "Pas": 55, "Pen": 15, "Tck": 20, "Tec": 65, "Agg": 25, "Ant": 70, "Bra": 20, "Cmp": 55, "Cnt": 30, "Dec": 50, "Det": 15, "Fla": 65, "Ldr": 15, "OtB": 60, "Pos": 35, "Tea": 45, "Vis": 65, "Wor": 25, "Acc": 75, "Agi": 75, "Bal": 45, "Jum": 15, "Nat": 0, "Pac": 50, "Sta": 50, "Str": 20 },
  "MC - BBM - Su": {"Cor": 10, "Cro": 10, "Dri": 50, "Fin": 40, "Fir": 55, "Fre": 20, "Hea": 10, "Lon": 50, "L Th": 5, "Mar": 20, "Pas": 55, "Pen": 10, "Tck": 35, "Tec": 60, "Agg": 60, "Ant": 50, "Bra": 30, "Cmp": 55, "Cnt": 45, "Dec": 60, "Det": 20, "Fla": 30, "Ldr": 10, "OtB": 50, "Pos": 60, "Tea": 40, "Vis": 45, "Wor": 65, "Acc": 65, "Agi": 45, "Bal": 45, "Jum": 15, "Nat": 10, "Pac": 55, "Sta": 70, "Str": 40 },
  "MC - BWM - De": {"Cor": 20, "Cro": 20, "Dri": 30, "Fin": 15, "Fir": 50, "Fre": 5, "Hea": 15, "Lon": 45, "L Th": 5, "Mar": 60, "Pas": 45, "Pen": 5, "Tck": 60, "Tec": 50, "Agg": 55, "Ant": 60, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 70, "Det": 10, "Fla": 10, "Ldr": 20, "OtB": 35, "Pos": 55, "Tea": 50, "Vis": 50, "Wor": 80, "Acc": 45, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 55, "Sta": 70, "Str": 55 },
  "MC - CM - Att": {"Cor": 20, "Cro": 35, "Dri": 50, "Fin": 25, "Fir": 50, "Fre": 5, "Hea": 15, "Lon": 50, "L Th": 5, "Mar": 15, "Pas": 50, "Pen": 5, "Tck": 30, "Tec": 55, "Agg": 30, "Ant": 55, "Bra": 30, "Cmp": 55, "Cnt": 50, "Dec": 75, "Det": 10, "Fla": 10, "Ldr": 20, "OtB": 55, "Pos": 35, "Tea": 50, "Vis": 65, "Wor": 65, "Acc": 55, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 50, "Sta": 75, "Str": 40 },
  "MC - CM - De": {"Cor": 20, "Cro": 20, "Dri": 30, "Fin": 15, "Fir": 55, "Fre": 5, "Hea": 15, "Lon": 50, "L Th": 5, "Mar": 55, "Pas": 50, "Pen": 5, "Tck": 45, "Tec": 50, "Agg": 55, "Ant": 60, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 70, "Det": 10, "Fla": 10, "Ldr": 20, "OtB": 50, "Pos": 35, "Tea": 50, "Vis": 70, "Wor": 70, "Acc": 60, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 45, "Sta": 70, "Str": 45 },
  "MC - CM - Su": {"Cor": 20, "Cro": 35, "Dri": 35, "Fin": 25, "Fir": 55, "Fre": 5, "Hea": 15, "Lon": 50, "L Th": 5, "Mar": 15, "Pas": 65, "Pen": 5, "Tck": 35, "Tec": 55, "Agg": 30, "Ant": 55, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 70, "Det": 10, "Fla": 10, "Ldr": 20, "OtB": 50, "Pos": 35, "Tea": 50, "Vis": 75, "Wor": 70, "Acc": 55, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 50, "Sta": 70, "Str": 45 },
  "MC - Generic": {"Cor": 10, "Cro": 10, "Dri": 20, "Fin": 20, "Fir": 60, "Fre": 10, "Hea": 10, "Lon": 30, "L Th": 10, "Mar": 30, "Pas": 60, "Pen": 10, "Tck": 30, "Tec": 40, "Agg": 0, "Ant": 30, "Bra": 10, "Cmp": 30, "Cnt": 20, "Dec": 70, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 30, "Pos": 30, "Tea": 20, "Vis": 60, "Wor": 30, "Acc": 60, "Agi": 60, "Bal": 20, "Jum": 10, "Nat": 0, "Pac": 50, "Sta": 60, "Str": 40 },
  "MC - MEZ - At": {"Cor": 10, "Cro": 35, "Dri": 55, "Fin": 40, "Fir": 60, "Fre": 10, "Hea": 15, "Lon": 35, "L Th": 10, "Mar": 25, "Pas": 45, "Pen": 10, "Tck": 20, "Tec": 55, "Agg": 5, "Ant": 30, "Bra": 10, "Cmp": 40, "Cnt": 15, "Dec": 50, "Det": 0, "Fla": 5, "Ldr": 5, "OtB": 35, "Pos": 30, "Tea": 25, "Vis": 55, "Wor": 45, "Acc": 70, "Agi": 55, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 55, "Sta": 55, "Str": 25 },
  "MC - MEZ - Su": {"Cor": 10, "Cro": 30, "Dri": 45, "Fin": 30, "Fir": 60, "Fre": 10, "Hea": 15, "Lon": 35, "L Th": 10, "Mar": 25, "Pas": 50, "Pen": 10, "Tck": 25, "Tec": 55, "Agg": 5, "Ant": 30, "Bra": 10, "Cmp": 35, "Cnt": 15, "Dec": 55, "Det": 0, "Fla": 5, "Ldr": 5, "OtB": 35, "Pos": 30, "Tea": 25, "Vis": 55, "Wor": 45, "Acc": 65, "Agi": 50, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 55, "Sta": 55, "Str": 30 },
  "MC - RPM - Su": {"Cor": 10, "Cro": 10, "Dri": 50, "Fin": 20, "Fir": 50, "Fre": 30, "Hea": 10, "Lon": 40, "L Th": 5, "Mar": 20, "Pas": 70, "Pen": 10, "Tck": 35, "Tec": 50, "Agg": 50, "Ant": 55, "Bra": 30, "Cmp": 60, "Cnt": 50, "Dec": 65, "Det": 20, "Fla": 50, "Ldr": 10, "OtB": 40, "Pos": 65, "Tea": 10, "Vis": 60, "Wor": 80, "Acc": 70, "Agi": 45, "Bal": 35, "Jum": 15, "Nat": 10, "Pac": 60, "Sta": 65, "Str": 35 },
  "MR/L - IW - At": {"Cor": 30, "Cro": 55, "Dri": 65, "Fin": 35, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 30, "Mar": 35, "Pas": 50, "Pen": 15, "Tck": 35, "Tec": 50, "Agg": 35, "Ant": 35, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 10, "Vis": 30, "Wor": 60, "Acc": 90, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "MR/L - IW - Su": {"Cor": 30, "Cro": 25, "Dri": 50, "Fin": 45, "Fir": 55, "Fre": 10, "Hea": 10, "Lon": 40, "L Th": 0, "Mar": 20, "Pas": 60, "Pen": 15, "Tck": 10, "Tec": 55, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 60, "Wor": 55, "Acc": 90, "Agi": 55, "Bal": 20, "Jum": 10, "Nat": 10, "Pac": 75, "Sta": 70, "Str": 35 },
  "MR/L - WG - At": {"Cor": 30, "Cro": 65, "Dri": 55, "Fin": 15, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 0, "Mar": 35, "Pas": 50, "Pen": 15, "Tck": 35, "Tec": 50, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 35, "Wor": 40, "Acc": 100, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "MR/L - WG - Su": {"Cor": 30, "Cro": 70, "Dri": 45, "Fin": 15, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 0, "Mar": 40, "Pas": 60, "Pen": 15, "Tck": 40, "Tec": 55, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 55, "Wor": 40, "Acc": 100, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "MR/L Generic": {"Cor": 10, "Cro": 50, "Dri": 30, "Fin": 20, "Fir": 40, "Fre": 10, "Hea": 10, "Lon": 20, "L Th": 10, "Mar": 10, "Pas": 30, "Pen": 10, "Tck": 20, "Tec": 40, "Agg": 0, "Ant": 30, "Bra": 10, "Cmp": 30, "Cnt": 20, "Dec": 50, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 20, "Pos": 10, "Tea": 20, "Vis": 30, "Wor": 30, "Acc": 80, "Agi": 60, "Bal": 20, "Jum": 10, "Nat": 0, "Pac": 60, "Sta": 50, "Str": 30 },
  "AMC - AM - At": {"Cor": 20, "Cro": 35, "Dri": 65, "Fin": 40, "Fir": 55, "Fre": 5, "Hea": 15, "Lon": 55, "L Th": 5, "Mar": 15, "Pas": 50, "Pen": 5, "Tck": 25, "Tec": 55, "Agg": 30, "Ant": 55, "Bra": 30, "Cmp": 55, "Cnt": 50, "Dec": 75, "Det": 10, "Fla": 20, "Ldr": 20, "OtB": 55, "Pos": 35, "Tea": 50, "Vis": 60, "Wor": 105, "Acc": 70, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 50, "Sta": 75, "Str": 40 },
  "AMC - AM - Su": {"Cor": 20, "Cro": 35, "Dri": 65, "Fin": 30, "Fir": 50, "Fre": 5, "Hea": 15, "Lon": 55, "L Th": 5, "Mar": 15, "Pas": 50, "Pen": 5, "Tck": 25, "Tec": 55, "Agg": 30, "Ant": 55, "Bra": 30, "Cmp": 55, "Cnt": 50, "Dec": 75, "Det": 10, "Fla": 10, "Ldr": 20, "OtB": 55, "Pos": 35, "Tea": 50, "Vis": 65, "Wor": 65, "Acc": 65, "Agi": 35, "Bal": 25, "Jum": 10, "Nat": 0, "Pac": 50, "Sta": 75, "Str": 40 },
  "AMC - AP - At": {"Cor": 30, "Cro": 35, "Dri": 60, "Fin": 40, "Fir": 65, "Fre": 15, "Hea": 15, "Lon": 30, "L Th": 5, "Mar": 15, "Pas": 60, "Pen": 15, "Tck": 20, "Tec": 65, "Agg": 25, "Ant": 70, "Bra": 20, "Cmp": 55, "Cnt": 30, "Dec": 50, "Det": 15, "Fla": 65, "Ldr": 15, "OtB": 60, "Pos": 35, "Tea": 45, "Vis": 60, "Wor": 25, "Acc": 70, "Agi": 65, "Bal": 45, "Jum": 15, "Nat": 0, "Pac": 50, "Sta": 50, "Str": 20 },
  "AMC - Generic": {"Cor": 10, "Cro": 10, "Dri": 30, "Fin": 30, "Fir": 50, "Fre": 10, "Hea": 10, "Lon": 30, "L Th": 10, "Mar": 10, "Pas": 40, "Pen": 10, "Tck": 20, "Tec": 50, "Agg": 0, "Ant": 30, "Bra": 10, "Cmp": 30, "Cnt": 20, "Dec": 60, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 30, "Pos": 20, "Tea": 20, "Vis": 60, "Wor": 30, "Acc": 90, "Agi": 60, "Bal": 20, "Jum": 10, "Nat": 0, "Pac": 70, "Sta": 60, "Str": 30 },
  "AMC - SS - At": {"Cor": 5, "Cro": 5, "Dri": 65, "Fin": 65, "Fir": 40, "Fre": 30, "Hea": 10, "Lon": 20, "L Th": 1, "Mar": 5, "Pas": 50, "Pen": 15, "Tck": 15, "Tec": 65, "Agg": 50, "Ant": 70, "Bra": 20, "Cmp": 35, "Cnt": 25, "Dec": 40, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 35, "Pos": 10, "Tea": 10, "Vis": 30, "Wor": 80, "Acc": 100, "Agi": 30, "Bal": 50, "Jum": 10, "Nat": 10, "Pac": 80, "Sta": 80, "Str": 30 },
  "AMC - T - At": {"Cor": 5, "Cro": 20, "Dri": 75, "Fin": 55, "Fir": 70, "Fre": 5, "Hea": 35, "Lon": 35, "L Th": 0, "Mar": 0, "Pas": 60, "Pen": 20, "Tck": 5, "Tec": 65, "Agg": 50, "Ant": 55, "Bra": 20, "Cmp": 45, "Cnt": 5, "Dec": 55, "Det": 20, "Fla": 35, "Ldr": 10, "OtB": 55, "Pos": 5, "Tea": 10, "Vis": 65, "Wor": 30, "Acc": 75, "Agi": 30, "Bal": 45, "Jum": 20, "Nat": 10, "Pac": 60, "Sta": 45, "Str": 25 },
  "AMR - WG - At": {"Cor": 30, "Cro": 65, "Dri": 55, "Fin": 15, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 0, "Mar": 35, "Pas": 50, "Pen": 15, "Tck": 35, "Tec": 50, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 35, "Wor": 40, "Acc": 100, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "AMR/L - Generic": {"Cor": 10, "Cro": 50, "Dri": 50, "Fin": 20, "Fir": 50, "Fre": 10, "Hea": 10, "Lon": 20, "L Th": 10, "Mar": 10, "Pas": 20, "Pen": 10, "Tck": 20, "Tec": 40, "Agg": 0, "Ant": 30, "Bra": 10, "Cmp": 30, "Cnt": 20, "Dec": 50, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 20, "Pos": 10, "Tea": 20, "Vis": 30, "Wor": 30, "Acc": 100, "Agi": 60, "Bal": 20, "Jum": 10, "Nat": 0, "Pac": 100, "Sta": 70, "Str": 30 },
  "AMR/L - IF - At": {"Cor": 30, "Cro": 25, "Dri": 60, "Fin": 50, "Fir": 45, "Fre": 10, "Hea": 10, "Lon": 40, "L Th": 0, "Mar": 20, "Pas": 50, "Pen": 15, "Tck": 10, "Tec": 55, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 50, "Wor": 55, "Acc": 85, "Agi": 55, "Bal": 20, "Jum": 10, "Nat": 10, "Pac": 75, "Sta": 70, "Str": 35 },
  "AMR/L - IF - Su": {"Cor": 30, "Cro": 25, "Dri": 50, "Fin": 45, "Fir": 55, "Fre": 10, "Hea": 10, "Lon": 40, "L Th": 0, "Mar": 20, "Pas": 60, "Pen": 15, "Tck": 10, "Tec": 55, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 60, "Wor": 55, "Acc": 90, "Agi": 55, "Bal": 20, "Jum": 10, "Nat": 10, "Pac": 75, "Sta": 70, "Str": 35 },
  "AMR/L - IW - At": {"Cor": 30, "Cro": 55, "Dri": 65, "Fin": 35, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 30, "Mar": 35, "Pas": 50, "Pen": 15, "Tck": 35, "Tec": 50, "Agg": 35, "Ant": 35, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 10, "Vis": 30, "Wor": 60, "Acc": 90, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "AMR/L - IW - Su": {"Cor": 30, "Cro": 65, "Dri": 55, "Fin": 15, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 30, "Mar": 35, "Pas": 50, "Pen": 15, "Tck": 35, "Tec": 50, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 10, "Vis": 35, "Wor": 75, "Acc": 100, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "AMR/L - WG - At": {"Cor": 30, "Cro": 65, "Dri": 55, "Fin": 30, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 25, "L Th": 0, "Mar": 25, "Pas": 55, "Pen": 15, "Tck": 25, "Tec": 50, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 40, "Wor": 50, "Acc": 100, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "AMR/L - WG - Su": {"Cor": 30, "Cro": 70, "Dri": 45, "Fin": 15, "Fir": 30, "Fre": 10, "Hea": 10, "Lon": 10, "L Th": 0, "Mar": 40, "Pas": 60, "Pen": 15, "Tck": 40, "Tec": 55, "Agg": 35, "Ant": 45, "Bra": 15, "Cmp": 30, "Cnt": 35, "Dec": 35, "Det": 20, "Fla": 20, "Ldr": 10, "OtB": 40, "Pos": 35, "Tea": 75, "Vis": 55, "Wor": 40, "Acc": 100, "Agi": 50, "Bal": 15, "Jum": 10, "Nat": 10, "Pac": 100, "Sta": 75, "Str": 30 },
  "ST - AF - At": {"Cor": 5, "Cro": 5, "Dri": 75, "Fin": 80, "Fir": 50, "Fre": 5, "Hea": 25, "Lon": 25, "L Th": 1, "Mar": 1, "Pas": 40, "Pen": 20, "Tck": 5, "Tec": 65, "Agg": 50, "Ant": 50, "Bra": 20, "Cmp": 35, "Cnt": 5, "Dec": 45, "Det": 20, "Fla": 25, "Ldr": 10, "OtB": 45, "Pos": 5, "Tea": 10, "Vis": 20, "Wor": 60, "Acc": 100, "Agi": 30, "Bal": 50, "Jum": 20, "Nat": 10, "Pac": 70, "Sta": 65, "Str": 25 },
  "ST - CF - At": {"Cor": 5, "Cro": 20, "Dri": 65, "Fin": 70, "Fir": 65, "Fre": 5, "Hea": 55, "Lon": 50, "L Th": 0, "Mar": 0, "Pas": 40, "Pen": 20, "Tck": 5, "Tec": 45, "Agg": 50, "Ant": 60, "Bra": 35, "Cmp": 55, "Cnt": 25, "Dec": 55, "Det": 20, "Fla": 35, "Ldr": 10, "OtB": 60, "Pos": 5, "Tea": 50, "Vis": 35, "Wor": 50, "Acc": 80, "Agi": 50, "Bal": 45, "Jum": 40, "Nat": 30, "Pac": 70, "Sta": 55, "Str": 50 },
  "ST - CF - Su": {"Cor": 5, "Cro": 20, "Dri": 75, "Fin": 60, "Fir": 75, "Fre": 5, "Hea": 50, "Lon": 50, "L Th": 0, "Mar": 0, "Pas": 50, "Pen": 20, "Tck": 5, "Tec": 45, "Agg": 50, "Ant": 60, "Bra": 35, "Cmp": 55, "Cnt": 25, "Dec": 55, "Det": 20, "Fla": 35, "Ldr": 10, "OtB": 60, "Pos": 5, "Tea": 50, "Vis": 40, "Wor": 50, "Acc": 75, "Agi": 50, "Bal": 45, "Jum": 40, "Nat": 30, "Pac": 55, "Sta": 50, "Str": 45 },
  "ST - DLF - At": {"Cor": 5, "Cro": 20, "Dri": 45, "Fin": 75, "Fir": 55, "Fre": 5, "Hea": 45, "Lon": 40, "L Th": 0, "Mar": 0, "Pas": 40, "Pen": 20, "Tck": 5, "Tec": 50, "Agg": 35, "Ant": 65, "Bra": 35, "Cmp": 70, "Cnt": 25, "Dec": 60, "Det": 20, "Fla": 45, "Ldr": 10, "OtB": 65, "Pos": 5, "Tea": 55, "Vis": 35, "Wor": 45, "Acc": 45, "Agi": 50, "Bal": 35, "Jum": 30, "Nat": 30, "Pac": 55, "Sta": 55, "Str": 65 },
  "ST - DLF - Su": {"Cor": 5, "Cro": 20, "Dri": 55, "Fin": 65, "Fir": 55, "Fre": 5, "Hea": 45, "Lon": 40, "L Th": 0, "Mar": 0, "Pas": 50, "Pen": 20, "Tck": 5, "Tec": 45, "Agg": 35, "Ant": 65, "Bra": 35, "Cmp": 70, "Cnt": 25, "Dec": 60, "Det": 20, "Fla": 45, "Ldr": 10, "OtB": 65, "Pos": 5, "Tea": 55, "Vis": 60, "Wor": 45, "Acc": 45, "Agi": 50, "Bal": 35, "Jum": 30, "Nat": 30, "Pac": 55, "Sta": 55, "Str": 60 },
  "ST - F9 - Su": {"Cor": 30, "Cro": 35, "Dri": 55, "Fin": 35, "Fir": 65, "Fre": 15, "Hea": 15, "Lon": 30, "L Th": 5, "Mar": 15, "Pas": 55, "Pen": 15, "Tck": 20, "Tec": 65, "Agg": 25, "Ant": 70, "Bra": 20, "Cmp": 55, "Cnt": 30, "Dec": 50, "Det": 15, "Fla": 65, "Ldr": 15, "OtB": 60, "Pos": 35, "Tea": 45, "Vis": 65, "Wor": 25, "Acc": 75, "Agi": 75, "Bal": 45, "Jum": 15, "Nat": 0, "Pac": 50, "Sta": 50, "Str": 20 },
  "ST - Generic": {"Cor": 10, "Cro": 20, "Dri": 50, "Fin": 80, "Fir": 60, "Fre": 10, "Hea": 60, "Lon": 20, "L Th": 10, "Mar": 10, "Pas": 20, "Pen": 10, "Tck": 10, "Tec": 40, "Agg": 0, "Ant": 50, "Bra": 10, "Cmp": 60, "Cnt": 20, "Dec": 50, "Det": 0, "Fla": 0, "Ldr": 10, "OtB": 60, "Pos": 20, "Tea": 10, "Vis": 20, "Wor": 20, "Acc": 100, "Agi": 60, "Bal": 20, "Jum": 50, "Nat": 0, "Pac": 70, "Sta": 60, "Str": 60 },
  "ST - PF - At": {"Cor": 10, "Cro": 30, "Dri": 45, "Fin": 65, "Fir": 70, "Fre": 10, "Hea": 15, "Lon": 35, "L Th": 10, "Mar": 25, "Pas": 40, "Pen": 10, "Tck": 25, "Tec": 55, "Agg": 35, "Ant": 55, "Bra": 50, "Cmp": 60, "Cnt": 35, "Dec": 60, "Det": 0, "Fla": 5, "Ldr": 5, "OtB": 65, "Pos": 30, "Tea": 60, "Vis": 30, "Wor": 70, "Acc": 75, "Agi": 50, "Bal": 35, "Jum": 10, "Nat": 0, "Pac": 55, "Sta": 55, "Str": 40 },
  "ST - PF - De": {"Cor": 10, "Cro": 30, "Dri": 35, "Fin": 50, "Fir": 65, "Fre": 10, "Hea": 15, "Lon": 35, "L Th": 10, "Mar": 25, "Pas": 50, "Pen": 10, "Tck": 25, "Tec": 55, "Agg": 35, "Ant": 55, "Bra": 50, "Cmp": 60, "Cnt": 35, "Dec": 60, "Det": 0, "Fla": 5, "Ldr": 5, "OtB": 45, "Pos": 30, "Tea": 60, "Vis": 35, "Wor": 70, "Acc": 65, "Agi": 50, "Bal": 35, "Jum": 10, "Nat": 0, "Pac": 55, "Sta": 55, "Str": 40 },
  "ST - PF - Su": {"Cor": 10, "Cro": 30, "Dri": 40, "Fin": 55, "Fir": 60, "Fre": 10, "Hea": 15, "Lon": 35, "L Th": 10, "Mar": 25, "Pas": 60, "Pen": 10, "Tck": 25, "Tec": 55, "Agg": 35, "Ant": 55, "Bra": 50, "Cmp": 60, "Cnt": 35, "Dec": 60, "Det": 0, "Fla": 5, "Ldr": 5, "OtB": 65, "Pos": 30, "Tea": 60, "Vis": 35, "Wor": 70, "Acc": 65, "Agi": 50, "Bal": 35, "Jum": 10, "Nat": 0, "Pac": 55, "Sta": 55, "Str": 40 },
  "ST - POA - At": {"Cor": 5, "Cro": 5, "Dri": 55, "Fin": 75, "Fir": 70, "Fre": 5, "Hea": 55, "Lon": 25, "L Th": 0, "Mar": 0, "Pas": 25, "Pen": 20, "Tck": 5, "Tec": 50, "Agg": 50, "Ant": 70, "Bra": 20, "Cmp": 35, "Cnt": 5, "Dec": 55, "Det": 20, "Fla": 25, "Ldr": 10, "OtB": 60, "Pos": 5, "Tea": 10, "Vis": 20, "Wor": 40, "Acc": 100, "Agi": 30, "Bal": 50, "Jum": 35, "Nat": 10, "Pac": 55, "Sta": 55, "Str": 35 },
  "ST - T - At": {"Cor": 5, "Cro": 20, "Dri": 75, "Fin": 60, "Fir": 70, "Fre": 5, "Hea": 35, "Lon": 35, "L Th": 0, "Mar": 0, "Pas": 55, "Pen": 20, "Tck": 5, "Tec": 65, "Agg": 50, "Ant": 55, "Bra": 20, "Cmp": 45, "Cnt": 5, "Dec": 55, "Det": 20, "Fla": 35, "Ldr": 10, "OtB": 55, "Pos": 5, "Tea": 10, "Vis": 60, "Wor": 30, "Acc": 85, "Agi": 30, "Bal": 45, "Jum": 20, "Nat": 10, "Pac": 60, "Sta": 45, "Str": 25 },
  "ST - TF - At": {"Cor": 10, "Cro": 30, "Dri": 35, "Fin": 80, "Fir": 65, "Fre": 10, "Hea": 60, "Lon": 30, "L Th": 10, "Mar": 25, "Pas": 35, "Pen": 10, "Tck": 15, "Tec": 35, "Agg": 60, "Ant": 65, "Bra": 55, "Cmp": 65, "Cnt": 35, "Dec": 65, "Det": 0, "Fla": 5, "Ldr": 5, "OtB": 70, "Pos": 30, "Tea": 50, "Vis": 25, "Wor": 50, "Acc": 40, "Agi": 35, "Bal": 50, "Jum": 60, "Nat": 0, "Pac": 55, "Sta": 40, "Str": 70 },
  "ST - TF - Su": {"Cor": 10, "Cro": 30, "Dri": 35, "Fin": 75, "Fir": 60, "Fre": 10, "Hea": 60, "Lon": 30, "L Th": 10, "Mar": 25, "Pas": 45, "Pen": 10, "Tck": 15, "Tec": 35, "Agg": 60, "Ant": 65, "Bra": 55, "Cmp": 65, "Cnt": 35, "Dec": 65, "Det": 0, "Fla": 5, "Ldr": 5, "OtB": 70, "Pos": 30, "Tea": 50, "Vis": 35, "Wor": 50, "Acc": 40, "Agi": 35, "Bal": 50, "Jum": 60, "Nat": 0, "Pac": 55, "Sta": 40, "Str": 65 }
};
// END: Default weights

// Mapping from parsed positions to *base keys* for role_specific_overall_weights.json
// This helps find all relevant roles (e.g., "DC - BPD", "DC - CD", "DC - Generic")
const parsedPositionToBaseRoleKey = {
    "Goalkeeper": null, 
    "Sweeper": "DC", 
    "Right Back": "DR/L",
    "Left Back": "DR/L",
    "Centre Back": "DC",
    "Right Wing-Back": "WBR/L",
    "Left Wing-Back": "WBR/L",
    "Centre Wing-Back": "WBR/L", 
    "Right Defensive Midfielder": "DM",
    "Left Defensive Midfielder": "DM",
    "Centre Defensive Midfielder": "DM",
    "Right Midfielder": "MR/L", // Using MR/L as the base for Right Midfielder
    "Left Midfielder": "MR/L",  // Using MR/L as the base for Left Midfielder
    "Centre Midfielder": "MC",
    "Right Attacking Midfielder": "AMR/L",
    "Left Attacking Midfielder": "AMR/L",
    "Centre Attacking Midfielder": "AMC",
    "Striker": "ST",
    "Right Forward": "AMR/L", 
    "Left Forward": "AMR/L",
    "Centre Forward": "ST"
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

        const attributeWeights = ref({}); 
        const attributeWeightsLoaded = ref(false);
        const attributeWeightsError = ref("");
        const loadingAttributeWeights = ref(false);

        const roleSpecificOverallWeights = ref({});
        const roleSpecificOverallWeightsLoaded = ref(false);
        const roleSpecificOverallWeightsError = ref("");
        const loadingRoleSpecificOverallWeights = ref(false);

        const sortState = reactive({ key: null, direction: "asc", isAttribute: false, displayField: null });
        const filters = reactive({ name: "", club: "", transferValue: "", position: null, nationality: "" });

        const hasActiveFilters = computed(() => filters.name !== "" || filters.club !== "" || filters.transferValue !== "" || filters.position !== null || filters.nationality !== "");
        const uniqueClubsCount = computed(() => new Set(allPlayers.value.map((p) => p.club).filter(Boolean)).size);
        const uniqueParsedPositionsCount = computed(() => { const p = new Set(); allPlayers.value.forEach((player) => player.parsedPositions?.forEach((pos) => p.add(pos))); return p.size; });
        const uniqueNationalitiesCount = computed(() => new Set(allPlayers.value.map((p) => p.nationality).filter(Boolean)).size);

        const parseMonetaryValue = (valueStr) => {
            if (typeof valueStr !== "string" || !valueStr) return 0;
            const c = valueStr.split(" p/w")[0];
            let m = 1;
            const l = c.toLowerCase();
            if (l.includes("m")) m = 1000000;
            else if (l.includes("k")) m = 1000;
            let n = c.replace(/[^0-9,.]/g, "");
            n = n.replace(/,/g, ""); 
            const v = parseFloat(n);
            return Math.round(isNaN(v) ? 0 : v * m);
        };
        
        const calculateFifaStat = (playerAttributes, categoryName) => {
            const currentCategoryWeights = attributeWeightsLoaded.value && Object.keys(attributeWeights.value).length > 0 ? attributeWeights.value : defaultAttributeWeights;
            const categoryAttributeWeights = currentCategoryWeights[categoryName];
            if (!categoryAttributeWeights) return 0;
            let weightedSum = 0; let totalWeightOfPresentAttributes = 0;
            for (const attrName in categoryAttributeWeights) {
                if (playerAttributes.hasOwnProperty(attrName)) {
                    const attrValue = parseInt(playerAttributes[attrName], 10);
                    const attrWeight = categoryAttributeWeights[attrName];
                    if (!isNaN(attrValue) && attrValue >= 0 && attrValue <= 20) {
                        weightedSum += attrValue * attrWeight;
                        totalWeightOfPresentAttributes += attrWeight;
                    }
                }
            }
            if (totalWeightOfPresentAttributes === 0) return 0;
            const weightedAverage = weightedSum / totalWeightOfPresentAttributes;
            return Math.round(weightedAverage * 5);
        };

        const calculateOverallForRole = (playerNumericAttributes, roleSpecificAttributeWeights) => {
            if (!roleSpecificAttributeWeights || Object.keys(roleSpecificAttributeWeights).length === 0) {
                return 0;
            }
            let weightedAttributeSum = 0;
            let totalApplicableWeightsSum = 0;

            for (const attrKey in roleSpecificAttributeWeights) {
                if (playerNumericAttributes.hasOwnProperty(attrKey)) {
                    const attributeValue = playerNumericAttributes[attrKey]; 
                    const weightForAttribute = roleSpecificAttributeWeights[attrKey];

                    if (typeof attributeValue === 'number' && !isNaN(attributeValue) &&
                        typeof weightForAttribute === 'number' && !isNaN(weightForAttribute)) {
                        const validAttributeValue = Math.max(0, Math.min(20, attributeValue));
                        weightedAttributeSum += validAttributeValue * weightForAttribute;
                        totalApplicableWeightsSum += weightForAttribute;
                    }
                }
            }

            if (totalApplicableWeightsSum === 0) {
                return 0;
            }
            const rawPositionalOverall = weightedAttributeSum / totalApplicableWeightsSum;
            const OVERALL_SCALING_FACTOR = 5.85; 
            return Math.min(99, Math.round(rawPositionalOverall * OVERALL_SCALING_FACTOR));
        };

        const loadJsonWeights = async (filePath, targetRef, loadedFlagRef, errorRef, loadingFlagRef, defaultWeights) => {
            loadingFlagRef.value = true;
            errorRef.value = "";
            try {
                const response = await fetch(filePath);
                if (!response.ok) throw new Error(`HTTP error! status: ${response.status} for ${filePath}`);
                const data = await response.json();
                targetRef.value = data;
                loadedFlagRef.value = true;
            } catch (e) {
                console.error(`Failed to load ${filePath}:`, e);
                errorRef.value = e.message || `Unknown error loading ${filePath}.`;
                targetRef.value = { ...defaultWeights }; 
                loadedFlagRef.value = true; 
            } finally {
                loadingFlagRef.value = false;
            }
        };

        const loadAttributeWeights = () => loadJsonWeights('/attribute_weights.json', attributeWeights, attributeWeightsLoaded, attributeWeightsError, loadingAttributeWeights, defaultAttributeWeights);
        const loadRoleSpecificOverallWeights = () => loadJsonWeights('/role_specific_overall_weights.json', roleSpecificOverallWeights, roleSpecificOverallWeightsLoaded, roleSpecificOverallWeightsError, loadingRoleSpecificOverallWeights, defaultRoleSpecificOverallWeights);
        
        onMounted(() => {
            loadAttributeWeights();
            loadRoleSpecificOverallWeights();
        });

        const processPlayerData = (players) => {
            return players.map((player) => {
                const transferValue = parseMonetaryValue(player.transfer_value);
                const wageValue = parseMonetaryValue(player.wage);
                const numericAttributes = {};
                if (player.attributes) {
                    Object.keys(player.attributes).forEach((key) => {
                        const value = player.attributes[key];
                        numericAttributes[key] = value && !isNaN(parseInt(value, 10)) ? parseInt(value, 10) : 0;
                    });
                }
                const parsedPlayerPositions = parsePlayerPositions(player.position);
                const playerPosGroups = getPlayerPositionGroups(parsedPlayerPositions);

                const phy = calculateFifaStat(numericAttributes, "PHY");
                const sho = calculateFifaStat(numericAttributes, "SHO");
                const pas = calculateFifaStat(numericAttributes, "PAS");
                const dri = calculateFifaStat(numericAttributes, "DRI");
                const def = calculateFifaStat(numericAttributes, "DEF");
                const men = calculateFifaStat(numericAttributes, "MEN");

                let maxOverall = 0;
                const calculatedRoleOveralls = []; // To store { roleName, score }

                const currentRoleWeights = roleSpecificOverallWeightsLoaded.value && Object.keys(roleSpecificOverallWeights.value).length > 0
                                               ? roleSpecificOverallWeights.value
                                               : defaultRoleSpecificOverallWeights;

                if (parsedPlayerPositions && parsedPlayerPositions.length > 0) {
                    const uniqueBaseRoleKeysConsidered = new Set(); // Track base roles to avoid redundant generic calculations

                    parsedPlayerPositions.forEach(parsedPos => {
                        const baseRoleKey = parsedPositionToBaseRoleKey[parsedPos]; // e.g., "DC", "DR/L"
                        
                        if (baseRoleKey) {
                            // Iterate through all role-specific keys in the loaded JSON
                            for (const roleKeyInJson in currentRoleWeights) {
                                // Check if the JSON key starts with the baseRoleKey (e.g., "DC - BPD" starts with "DC - ")
                                // Or if it's the generic version for that base role (e.g., "DC - Generic")
                                if (roleKeyInJson.startsWith(baseRoleKey + " - ") || roleKeyInJson === baseRoleKey + " - Generic") {
                                    const roleSpecificAttributeWeights = currentRoleWeights[roleKeyInJson];
                                    const overallForThisRole = calculateOverallForRole(numericAttributes, roleSpecificAttributeWeights);
                                    
                                    calculatedRoleOveralls.push({ roleName: roleKeyInJson, score: overallForThisRole });
                                    if (overallForThisRole > maxOverall) {
                                        maxOverall = overallForThisRole;
                                    }
                                }
                            }
                             // Add generic calculation for the base role if not already done by a specific role starting with "baseRoleKey - Generic"
                            const genericRoleKey = baseRoleKey + " - Generic";
                            if (currentRoleWeights[genericRoleKey] && !calculatedRoleOveralls.find(r => r.roleName === genericRoleKey)) {
                                 if (!uniqueBaseRoleKeysConsidered.has(genericRoleKey)) {
                                    const roleSpecificAttributeWeights = currentRoleWeights[genericRoleKey];
                                    const overallForThisRole = calculateOverallForRole(numericAttributes, roleSpecificAttributeWeights);
                                    calculatedRoleOveralls.push({ roleName: genericRoleKey, score: overallForThisRole });
                                    if (overallForThisRole > maxOverall) {
                                        maxOverall = overallForThisRole;
                                    }
                                    uniqueBaseRoleKeysConsidered.add(genericRoleKey);
                                 }
                            }
                        }
                    });
                }
                const overall = maxOverall;
                
                return {
                    ...player,
                    age: parseInt(player.age, 10) || 0,
                    transferValueAmount: transferValue, wageAmount: wageValue,
                    attributes: numericAttributes,
                    PHY: phy, SHO: sho, PAS: pas, DRI: dri, DEF: def, MEN: men, 
                    Overall: overall, 
                    parsedPositions: parsedPlayerPositions, positionGroups: playerPosGroups,
                    roleSpecificOveralls: calculatedRoleOveralls, // Store all calculated role overalls
                };
            });
        };

        const positionFilterOptions = computed(() => { const o = []; Object.keys(positionGroups).forEach((g) => o.push({ label: `${g} (Group)`, value: g })); const s = new Set(); allPlayers.value.forEach((p) => p.parsedPositions?.forEach((pos) => s.add(pos))); Array.from(s).sort().forEach((pos) => { if (!positionGroups[pos]) o.push({ label: pos, value: pos }); }); return o; });
        const filteredPlayers = computed(() => { if (!allPlayers.value.length) return []; let tempPlayers = [...allPlayers.value]; if (filters.name) tempPlayers = tempPlayers.filter((p) => p.name && p.name.toLowerCase().includes(filters.name.toLowerCase())); if (filters.club) tempPlayers = tempPlayers.filter((p) => p.club && p.club.toLowerCase().includes(filters.club.toLowerCase())); if (filters.transferValue) { let o = "includes"; let c = 0; let f = filters.transferValue; if (f.startsWith(">")) { o = ">"; f = f.substring(1); } else if (f.startsWith("<")) { o = "<"; f = f.substring(1); } if (o !== "includes") c = parseMonetaryValue(f); tempPlayers = tempPlayers.filter((p) => { const v = p.transferValueAmount || 0; if (o === ">") return v > c; if (o === "<") return v < c; return String(p.transfer_value || "").toLowerCase().includes(filters.transferValue.toLowerCase()); }); } if (filters.position) { const s = filters.position; if (positionGroups[s]) tempPlayers = tempPlayers.filter((p) => p.positionGroups && p.positionGroups.includes(s)); else tempPlayers = tempPlayers.filter((p) => p.parsedPositions && p.parsedPositions.includes(s)); } if (filters.nationality) tempPlayers = tempPlayers.filter((p) => p.nationality && p.nationality.toLowerCase().includes(filters.nationality.toLowerCase())); if (sortState.key) return sortPlayersLogic([...tempPlayers]); return tempPlayers; });
        const sortPlayersLogic = (playersToSort) => { if (!sortState.key) return playersToSort; const k = sortState.isAttribute ? sortState.key : allPlayers.value.length > 0 && Object.keys(allPlayers.value[0]).includes(sortState.key + "Amount") ? sortState.key + "Amount" : sortState.key; return [...playersToSort].sort((a, b) => { let vA, vB; if (sortState.isAttribute) { vA = a.attributes ? a.attributes[k] : null; vB = b.attributes ? b.attributes[k] : null; } else { vA = a[k]; vB = b[k]; } if (vA == null && vB == null) return 0; if (vA == null) return sortState.direction === "asc" ? 1 : -1; if (vB == null) return sortState.direction === "asc" ? -1 : 1; if (typeof vA === "number" && typeof vB === "number") return sortState.direction === "asc" ? vA - vB : vB - vA; vA = String(vA).toLowerCase(); vB = String(vB).toLowerCase(); if (vA < vB) return sortState.direction === "asc" ? -1 : 1; if (vA > vB) return sortState.direction === "asc" ? 1 : -1; return 0; }); };
        
        const uploadAndParse = async () => {
            if (!playerFile.value) { error.value = "Please select an HTML file first."; return; }
            if ((!attributeWeightsLoaded.value && attributeWeightsError.value) || (!roleSpecificOverallWeightsLoaded.value && roleSpecificOverallWeightsError.value)) {
                // console.warn("Parsing with default weights for some categories due to loading errors.");
            } else if (!attributeWeightsLoaded.value || !roleSpecificOverallWeightsLoaded.value) {
                error.value = "Attribute weights are not yet fully loaded. Please wait or check console for errors.";
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
        const handleSort = (sortParams) => { sortState.key = sortParams.key; sortState.direction = sortParams.direction; sortState.isAttribute = sortParams.isAttribute; sortState.displayField = sortParams.displayField; };
        const clearAllFilters = () => { filters.name = ""; filters.club = ""; filters.transferValue = ""; filters.position = null; filters.nationality = ""; };
        const handleSearch = () => { /* Reactive filtering */ };
        const handlePlayerSelected = (player) => { selectedPlayer.value = player; showPlayerDetailDialog.value = true; };

        return {
            playerFile, loading, error, allPlayers, filteredPlayers, uniqueClubsCount, uniqueParsedPositionsCount, uniqueNationalitiesCount,
            filters, hasActiveFilters, positionFilterOptions, uploadAndParse, handleSort, handleSearch, clearAllFilters,
            selectedPlayer, showPlayerDetailDialog, handlePlayerSelected,
            attributeWeightsLoaded, attributeWeightsError, loadingAttributeWeights, loadAttributeWeights,
            roleSpecificOverallWeightsLoaded, roleSpecificOverallWeightsError, loadingRoleSpecificOverallWeights, loadRoleSpecificOverallWeights
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