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
                            Ensure the Go API (running on port 8091) is started.
                        </li>
                        <li>
                            The Go API will attempt to load
                            <code>attribute_weights.json</code> and
                            <code>role_specific_overall_weights.json</code> from
                            its <code>public</code> folder. If not found, it
                            uses internal defaults.
                        </li>
                        <li>
                            Select an HTML file exported from Football Manager.
                        </li>
                        <li>Click "Upload and Parse".</li>
                        <li>
                            The table will display players with pre-calculated
                            FIFA-style stats (PHY, SHO, etc.), parsed positions,
                            Overall ratings (based on their best role), Age,
                            Media Handling, and Personality.
                        </li>
                        <li>
                            Use filters for Name, Club (searchable dropdown),
                            Position, Nationality (searchable dropdown),
                            Transfer Value, Media Handling (multi-select),
                            Personality (multi-select), and Age range. Input
                            fields are debounced for performance.
                        </li>
                        <li>
                            Click on any player row for a detailed view, which
                            will show all calculated role-specific overalls
                            provided by the API.
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
                        <template v-slot:prepend
                            ><q-icon name="attach_file"
                        /></template>
                    </q-file>
                    <q-btn
                        class="q-mt-md full-width"
                        color="primary"
                        label="Upload and Parse"
                        :loading="loading"
                        :disable="
                            !playerFile ||
                            !attributeWeightsLoadedForFeedback ||
                            !roleSpecificOverallWeightsLoadedForFeedback
                        "
                        @click="uploadAndParse"
                    >
                        <q-tooltip
                            v-if="
                                !attributeWeightsLoadedForFeedback ||
                                !roleSpecificOverallWeightsLoadedForFeedback
                            "
                        >
                            Client-side check for weight files in public folder
                            pending... (Note: Go API has its own loading logic)
                        </q-tooltip>
                    </q-btn>
                </q-card-section>
            </q-card>

            <q-card class="q-mb-md" v-if="allPlayers.length > 0">
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Search Players</div>
                    <div class="row q-col-gutter-md">
                        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                            <q-input
                                v-model="filters.name"
                                label="Player Name"
                                dense
                                outlined
                                clearable
                                @update:model-value="debouncedApplyFilters"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                            <q-select
                                v-model="filters.club"
                                :options="clubOptions"
                                label="Club"
                                dense
                                outlined
                                clearable
                                use-input
                                hide-selected
                                fill-input
                                input-debounce="300"
                                @filter="filterClubOptions"
                                @update:model-value="applyFiltersAndSort"
                            >
                                <template v-slot:no-option>
                                    <q-item>
                                        <q-item-section class="text-grey">
                                            No results
                                        </q-item-section>
                                    </q-item>
                                </template>
                            </q-select>
                        </div>
                        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                            <q-select
                                v-model="filters.nationality"
                                :options="nationalityOptions"
                                label="Nationality"
                                dense
                                outlined
                                clearable
                                use-input
                                hide-selected
                                fill-input
                                input-debounce="300"
                                @filter="filterNationalityOptions"
                                @update:model-value="applyFiltersAndSort"
                            >
                                <template v-slot:no-option>
                                    <q-item>
                                        <q-item-section class="text-grey">
                                            No results
                                        </q-item-section>
                                    </q-item>
                                </template>
                            </q-select>
                        </div>
                        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                            <q-select
                                v-model="filters.position"
                                :options="positionFilterOptions"
                                label="Position / Group"
                                dense
                                outlined
                                clearable
                                emit-value
                                map-options
                                @update:model-value="applyFiltersAndSort"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                            <q-select
                                v-model="filters.mediaHandling"
                                :options="mediaHandlingOptions"
                                label="Media Handling"
                                dense
                                outlined
                                multiple
                                use-chips
                                clearable
                                emit-value
                                map-options
                                @update:model-value="applyFiltersAndSort"
                            />
                        </div>
                        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                            <q-select
                                v-model="filters.personality"
                                :options="personalityOptions"
                                label="Personality"
                                dense
                                outlined
                                multiple
                                use-chips
                                clearable
                                emit-value
                                map-options
                                @update:model-value="applyFiltersAndSort"
                            />
                        </div>
                        <div class="col-12 col-sm-4 col-md-2 col-lg-1">
                            <q-input
                                v-model.number="filters.minAge"
                                type="number"
                                label="Min Age"
                                dense
                                outlined
                                clearable
                                :min="0"
                                @update:model-value="debouncedApplyFilters"
                            />
                        </div>
                        <div class="col-12 col-sm-4 col-md-2 col-lg-1">
                            <q-input
                                v-model.number="filters.maxAge"
                                type="number"
                                label="Max Age"
                                dense
                                outlined
                                clearable
                                :min="0"
                                @update:model-value="debouncedApplyFilters"
                            />
                        </div>
                        <div class="col-12 col-sm-4 col-md-4 col-lg-2">
                            <q-input
                                v-model="filters.transferValue"
                                label="Transfer Value"
                                dense
                                outlined
                                clearable
                                placeholder="e.g., €1.5M, >1M, <500K"
                                @update:model-value="debouncedApplyFilters"
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
            <q-banner
                v-if="attributeWeightsErrorForFeedback"
                class="bg-warning text-dark q-mb-md"
                rounded
            >
                Client feedback: Could not load
                <code>public/attribute_weights.json</code>. The Go API will use
                its internal defaults if it also fails to load this file. Error:
                {{ attributeWeightsErrorForFeedback }}
            </q-banner>
            <q-banner
                v-if="roleSpecificOverallWeightsErrorForFeedback"
                class="bg-warning text-dark q-mb-md"
                rounded
            >
                Client feedback: Could not load
                <code>public/role_specific_overall_weights.json</code>. The Go
                API will use its internal defaults if it also fails to load this
                file. Error: {{ roleSpecificOverallWeightsErrorForFeedback }}
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

                <div class="row justify-end q-mb-md">
                    <q-btn
                        color="secondary"
                        icon="find_replace"
                        label="Find Upgrades"
                        @click="showUpgradeFinder = true"
                        :disable="allPlayers.length === 0"
                    />
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

        <UpgradeFinderDialog
            :show="showUpgradeFinder"
            :players="allPlayers"
            @close="showUpgradeFinder = false"
        />
    </q-page>
</template>

<script>
import { ref, computed, reactive, onMounted, watch } from "vue";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import UpgradeFinderDialog from "../components/UpgradeFinderDialog.vue"; // Import the new component
import playerService from "../services/playerService";

// Position groups for filtering
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

// Debounce function to delay filter application
function debounce(fn, delay) {
    let timeoutID = null;
    return function (...args) {
        clearTimeout(timeoutID);
        timeoutID = setTimeout(() => {
            fn.apply(this, args);
        }, delay);
    };
}

export default {
    name: "PlayerUploadPage",
    components: { PlayerDataTable, PlayerDetailDialog, UpgradeFinderDialog }, // Add UpgradeFinderDialog
    setup() {
        const playerFile = ref(null);
        const loading = ref(false);
        const error = ref("");
        const allPlayers = ref([]);
        const filteredPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false); // For the new dialog

        // Feedback refs for JSON loading
        const attributeWeightsLoadedForFeedback = ref(false);
        const attributeWeightsErrorForFeedback = ref("");
        const roleSpecificOverallWeightsLoadedForFeedback = ref(false);
        const roleSpecificOverallWeightsErrorForFeedback = ref("");

        // Reactive state for sorting
        const sortState = reactive({
            key: null,
            direction: "asc",
            isAttribute: false, // Note: This might need review if attributes are sorted directly
            displayField: null,
        });

        // Reactive state for filters
        const filters = reactive({
            name: "",
            club: null,
            transferValue: "",
            position: null,
            nationality: null,
            mediaHandling: [], // Array for multi-select
            personality: [], // Array for multi-select
            minAge: null,
            maxAge: null,
        });

        // Options for q-select dropdowns
        const clubOptions = ref([]);
        const nationalityOptions = ref([]);
        const mediaHandlingOptions = ref([]);
        const personalityOptions = ref([]);

        // Store all unique values for filtering q-select options more efficiently
        let allUniqueClubs = [];
        let allUniqueNationalities = [];
        let allUniqueMediaHandlings = []; // Will store individual styles
        let allUniquePersonalities = [];

        // Computed property to check if any filter is active
        const hasActiveFilters = computed(
            () =>
                filters.name !== "" ||
                filters.club !== null ||
                filters.transferValue !== "" ||
                filters.position !== null ||
                filters.nationality !== null ||
                (Array.isArray(filters.mediaHandling) &&
                    filters.mediaHandling.length > 0) ||
                (Array.isArray(filters.personality) &&
                    filters.personality.length > 0) ||
                filters.minAge !== null ||
                filters.maxAge !== null,
        );

        // Computed properties for summary counts
        const uniqueClubsCount = computed(() => allUniqueClubs.length);
        const uniqueParsedPositionsCount = computed(() => {
            const s = new Set();
            allPlayers.value.forEach((player) =>
                player.parsedPositions?.forEach((pos) => s.add(pos)),
            );
            return s.size;
        });
        const uniqueNationalitiesCount = computed(
            () => allUniqueNationalities.length,
        );

        // Helper to parse monetary values from strings like "€1.5M" or "£50K p/w"
        const parseMonetaryValueForFilter = (valueStr) => {
            if (typeof valueStr !== "string" || !valueStr) return 0;
            const cleanedStr = valueStr.split(" p/w")[0]; // Remove " p/w" if present
            let multiplier = 1;
            const lowerStr = cleanedStr.toLowerCase();
            if (lowerStr.includes("m")) multiplier = 1000000;
            else if (lowerStr.includes("k")) multiplier = 1000;
            let numStr = cleanedStr.replace(/[^0-9,.]/g, ""); // Keep only numbers, comma, and dot
            numStr = numStr.replace(/,/g, ""); // Remove commas for float parsing
            const value = parseFloat(numStr);
            return Math.round(isNaN(value) ? 0 : value * multiplier);
        };

        // Function to load JSON files for client-side feedback (Go API handles its own loading)
        const loadJsonForFeedback = async (
            filePath,
            loadedFlagRef,
            errorRef,
        ) => {
            errorRef.value = ""; // Clear previous error
            try {
                const response = await fetch(filePath);
                if (!response.ok)
                    throw new Error(
                        `HTTP error! status: ${response.status} for ${filePath}`,
                    );
                await response.json(); // We just need to know if it's parsable
                loadedFlagRef.value = true;
            } catch (e) {
                console.warn(
                    `Client-side check: Failed to load ${filePath}:`,
                    e,
                );
                errorRef.value =
                    e.message || `Unknown error loading ${filePath}.`;
                loadedFlagRef.value = true; // Still set to true to enable upload button
            }
        };

        // Load JSONs for feedback on component mount
        onMounted(() => {
            loadJsonForFeedback(
                "/attribute_weights.json", // Assuming it's in the public folder
                attributeWeightsLoadedForFeedback,
                attributeWeightsErrorForFeedback,
            );
            loadJsonForFeedback(
                "/role_specific_overall_weights.json", // Assuming it's in the public folder
                roleSpecificOverallWeightsLoadedForFeedback,
                roleSpecificOverallWeightsErrorForFeedback,
            );
        });

        // Process players data received from the API
        const processPlayersFromAPI = (playersData) => {
            return playersData.map((p) => ({
                ...p,
                age: parseInt(p.age, 10) || 0, // Ensure age is a number
                // transferValueAmount and wageAmount are assumed to be provided by Go API
            }));
        };

        // Computed property for position filter options
        const positionFilterOptions = computed(() => {
            const options = [{ label: "Any Position", value: null }]; // Add "Any" option
            Object.keys(positionGroups).forEach((group) => {
                options.push({ label: `${group} (Group)`, value: group });
            });
            const uniquePositions = new Set();
            allPlayers.value.forEach((player) => {
                player.parsedPositions?.forEach((pos) =>
                    uniquePositions.add(pos),
                );
            });
            Array.from(uniquePositions)
                .sort()
                .forEach((pos) => {
                    // Avoid adding individual positions if they are already covered by a group value
                    if (!positionGroups[pos]) {
                        // This check might be redundant if groups are named differently
                        options.push({ label: pos, value: pos });
                    }
                });
            return options;
        });

        // Function to update all dropdown options based on current player data
        const updateDropdownOptions = () => {
            const clubs = new Set();
            const nationalities = new Set();
            const mediaHandlingsIndividual = new Set(); // For individual styles
            const personalities = new Set();

            allPlayers.value.forEach((p) => {
                if (p.club) clubs.add(p.club);
                if (p.nationality) nationalities.add(p.nationality);
                if (p.media_handling) {
                    // Split comma-separated media handling styles
                    p.media_handling.split(",").forEach((style) => {
                        const trimmedStyle = style.trim();
                        if (trimmedStyle)
                            mediaHandlingsIndividual.add(trimmedStyle);
                    });
                }
                if (p.personality) personalities.add(p.personality);
            });

            allUniqueClubs = Array.from(clubs).sort();
            allUniqueNationalities = Array.from(nationalities).sort();
            allUniqueMediaHandlings = Array.from(
                mediaHandlingsIndividual,
            ).sort(); // Now contains individual styles
            allUniquePersonalities = Array.from(personalities).sort();

            clubOptions.value = allUniqueClubs;
            nationalityOptions.value = allUniqueNationalities;
            // Map individual media handling styles to options
            mediaHandlingOptions.value = allUniqueMediaHandlings.map((mh) => ({
                label: mh,
                value: mh,
            }));
            personalityOptions.value = allUniquePersonalities.map((p) => ({
                label: p,
                value: p,
            }));
        };

        // Main function to apply all filters and sorting
        const applyFiltersAndSort = () => {
            if (!allPlayers.value.length) {
                filteredPlayers.value = [];
                return;
            }
            let tempPlayers = [...allPlayers.value];

            // Apply text-based filters
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
                    (p) => p.club === filters.club,
                );
            }
            if (filters.nationality) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.nationality === filters.nationality,
                );
            }

            // Apply Media Handling filter (NEW LOGIC)
            if (filters.mediaHandling && filters.mediaHandling.length > 0) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.media_handling) return false;
                    const playerStyles = p.media_handling
                        .split(",")
                        .map((s) => s.trim().toLowerCase());
                    const filterStyles = filters.mediaHandling.map((s) =>
                        s.toLowerCase(),
                    );
                    return playerStyles.some((style) =>
                        filterStyles.includes(style),
                    );
                });
            }

            // Apply Personality filter
            if (filters.personality && filters.personality.length > 0) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.personality) return false;
                    // Assuming personality is a single string, not comma-separated that needs splitting.
                    // If personality can also be comma-separated, the logic should mirror mediaHandling.
                    return filters.personality.includes(p.personality);
                });
            }

            // Apply age filters
            if (filters.minAge !== null && filters.minAge >= 0) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.age >= filters.minAge,
                );
            }
            if (filters.maxAge !== null && filters.maxAge >= 0) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.age <= filters.maxAge,
                );
            }

            // Apply position filter
            if (filters.position) {
                const selectedPosFilter = filters.position;
                if (positionGroups[selectedPosFilter]) {
                    // Check if it's a group
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.positionGroups &&
                            p.positionGroups.includes(selectedPosFilter),
                    );
                } else {
                    // It's an individual position
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            p.parsedPositions &&
                            p.parsedPositions.includes(selectedPosFilter),
                    );
                }
            }

            // Apply transfer value filter
            if (filters.transferValue) {
                let operator = "includes"; // Default for simple text match
                let compareValueNum = 0;
                let filterStr = filters.transferValue;

                if (filterStr.startsWith(">")) {
                    operator = ">";
                    filterStr = filterStr.substring(1);
                } else if (filterStr.startsWith("<")) {
                    operator = "<";
                    filterStr = filterStr.substring(1);
                }

                if (operator === ">" || operator === "<") {
                    compareValueNum = parseMonetaryValueForFilter(filterStr);
                }

                tempPlayers = tempPlayers.filter((p) => {
                    const playerValueNum = p.transferValueAmount || 0; // Use pre-parsed numeric value
                    const playerValueStr = String(
                        p.transfer_value || "",
                    ).toLowerCase(); // Original string for "includes"

                    if (operator === ">")
                        return playerValueNum > compareValueNum;
                    if (operator === "<")
                        return playerValueNum < compareValueNum;
                    // Fallback to "includes" for non-numeric or non-operator inputs
                    return playerValueStr.includes(
                        filters.transferValue.toLowerCase(),
                    );
                });
            }

            // Apply sorting
            if (sortState.key) {
                const sortKey = sortState.key;
                tempPlayers.sort((a, b) => {
                    let valA = a[sortKey];
                    let valB = b[sortKey];

                    // Handle null or undefined values to sort them consistently
                    if (valA == null && valB == null) return 0;
                    if (valA == null)
                        return sortState.direction === "asc" ? 1 : -1; // nulls last for asc, first for desc
                    if (valB == null)
                        return sortState.direction === "asc" ? -1 : 1; // nulls first for asc, last for desc

                    if (typeof valA === "number" && typeof valB === "number") {
                        return sortState.direction === "asc"
                            ? valA - valB
                            : valB - valA;
                    }
                    // Fallback to string comparison
                    valA = String(valA).toLowerCase();
                    valB = String(valB).toLowerCase();
                    if (valA < valB)
                        return sortState.direction === "asc" ? -1 : 1;
                    if (valA > valB)
                        return sortState.direction === "asc" ? 1 : -1;
                    return 0;
                });
            }
            filteredPlayers.value = tempPlayers;
        };

        // Debounced version of filter application for text inputs
        const debouncedApplyFilters = debounce(applyFiltersAndSort, 300);

        // Function to handle file upload and parsing
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
                const playersDataFromApi =
                    await playerService.uploadPlayerFile(formData);
                allPlayers.value = processPlayersFromAPI(playersDataFromApi);
                updateDropdownOptions(); // Update dropdowns after new data
                sortState.key = null; // Reset sort
                applyFiltersAndSort(); // Apply initial filters (if any) and sort
            } catch (e) {
                error.value = `Failed to parse player data: ${e.message || "Unknown error"}`;
                allPlayers.value = [];
                filteredPlayers.value = [];
            } finally {
                loading.value = false;
            }
        };

        // Handler for sort updates from PlayerDataTable
        const handleSort = (sortParams) => {
            sortState.key = sortParams.key;
            sortState.direction = sortParams.direction;
            // sortState.isAttribute = sortParams.isAttribute; // If needed
            // sortState.displayField = sortParams.displayField; // If needed
            applyFiltersAndSort();
        };

        // Function to clear all active filters
        const clearAllFilters = () => {
            filters.name = "";
            filters.club = null;
            filters.transferValue = "";
            filters.position = null;
            filters.nationality = null;
            filters.mediaHandling = [];
            filters.personality = [];
            filters.minAge = null;
            filters.maxAge = null;
            applyFiltersAndSort(); // Re-apply to show all players
        };

        // Handler for when a player row is selected (for detail view)
        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        // Filter functions for q-select with use-input (for Club and Nationality)
        const filterClubOptions = (val, update) => {
            if (val === "") {
                update(() => {
                    clubOptions.value = allUniqueClubs;
                });
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                clubOptions.value = allUniqueClubs.filter(
                    (v) => v.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        const filterNationalityOptions = (val, update) => {
            if (val === "") {
                update(() => {
                    nationalityOptions.value = allUniqueNationalities;
                });
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                nationalityOptions.value = allUniqueNationalities.filter(
                    (v) => v.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        // Watch for changes in allPlayers to update dropdowns and re-filter
        watch(
            () => allPlayers.value,
            () => {
                updateDropdownOptions();
                applyFiltersAndSort();
            },
            { deep: true, immediate: true },
        );

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
            clubOptions,
            nationalityOptions,
            mediaHandlingOptions,
            personalityOptions,
            filterClubOptions,
            filterNationalityOptions,
            uploadAndParse,
            handleSort,
            debouncedApplyFilters,
            applyFiltersAndSort, // Expose for direct calls if needed
            clearAllFilters,
            selectedPlayer,
            showPlayerDetailDialog,
            handlePlayerSelected,
            attributeWeightsLoadedForFeedback,
            attributeWeightsErrorForFeedback,
            roleSpecificOverallWeightsLoadedForFeedback,
            roleSpecificOverallWeightsErrorForFeedback,
            showUpgradeFinder, // For the new dialog
        };
    },
};
</script>

<style>
.q-page {
    max-width: 1600px; /* Or your preferred max width */
    margin: 0 auto;
}
/* Add any additional global styles or component-specific styles here if needed */
</style>
