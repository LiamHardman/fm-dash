<template>
    <q-page padding>
        <div class="q-pa-md">
            <h1
                class="text-h4 text-center q-mb-lg page-title"
                :class="$q.dark.isActive ? 'text-grey-2' : 'text-grey-9'"
            >
                Football Manager HTML Player Parser
            </h1>

            <q-card
                class="q-mb-md instructions-card"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-grey-3'
                        : 'bg-blue-grey-1 text-blue-grey-10'
                "
            >
                <q-card-section>
                    <div class="text-subtitle1 text-weight-bold">
                        Instructions:
                    </div>
                    <ol class="q-ml-md">
                        <li>Ensure the Go API (port 8091) is running.</li>
                        <li>
                            API loads <code>attribute_weights.json</code> and
                            <code>role_specific_overall_weights.json</code> from
                            its <code>public</code> folder.
                        </li>
                        <li>
                            Select an HTML file exported from Football Manager.
                            <br />
                            <span
                                class="text-caption"
                                :class="
                                    $q.dark.isActive
                                        ? 'text-grey-5'
                                        : 'text-grey-7'
                                "
                            >
                                Maximum file size: 15MB (approx. 10,000
                                players).
                            </span>
                        </li>
                        <li>
                            Click "Upload and Parse". Currency symbol will be
                            auto-detected.
                        </li>
                        <li>
                            Table shows players with FIFA-style stats, Overall
                            (best role), Age, etc.
                        </li>
                        <li>
                            Use filters: Name, Club, Position (short codes),
                            <strong>Role (specific to position)</strong>,
                            Nationality, Value ({{ detectedCurrencySymbol }}),
                            Media Handling, Personality, Age Range.
                        </li>
                        <li>
                            Click player row for detailed view (all role
                            overalls).
                        </li>
                        <li>Use "View Team Page" for team analysis.</li>
                    </ol>
                </q-card-section>
            </q-card>

            <q-card
                class="q-mb-md upload-card"
                :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'"
            >
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Upload HTML File:</div>
                    <q-file
                        v-model="playerFile"
                        label="Select HTML file"
                        accept=".html"
                        outlined
                        counter
                        :label-color="$q.dark.isActive ? 'grey-4' : ''"
                        :input-class="$q.dark.isActive ? 'text-grey-3' : ''"
                    >
                        <template v-slot:prepend
                            ><q-icon name="attach_file"
                        /></template>
                        <template v-slot:hint> Max file size: 15MB </template>
                    </q-file>
                    <q-btn
                        class="q-mt-md full-width"
                        color="primary"
                        label="Upload and Parse"
                        :loading="loading"
                        :disable="
                            !playerFile ||
                            !attributeWeightsLoadedForFeedback ||
                            !roleSpecificOverallWeightsLoadedForFeedback ||
                            loading
                        "
                        @click="uploadAndParse"
                    >
                        <q-tooltip
                            v-if="
                                !attributeWeightsLoadedForFeedback ||
                                !roleSpecificOverallWeightsLoadedForFeedback
                            "
                        >
                            Client-side check for weight files pending...
                        </q-tooltip>
                    </q-btn>
                </q-card-section>
            </q-card>

            <PlayerFilters
                v-if="playerStore.currentDatasetId"
                :players="allPlayers"
                :currency-symbol="detectedCurrencySymbol"
                :transfer-value-range="playerStore.transferValueRange"
                :unique-clubs="playerStore.uniqueClubs"
                :unique-nationalities="playerStore.uniqueNationalities"
                :unique-media-handlings="playerStore.uniqueMediaHandlings"
                :unique-personalities="playerStore.uniquePersonalities"
                @filter-changed="handleFilterChanged"
                :is-loading="loading"
            />

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
                        @click="playerStore.error = ''"
                /></template>
            </q-banner>

            <template v-if="allPlayers.length > 0">
                <div class="row justify-between items-center q-mb-md q-mt-md">
                    <q-btn
                        color="info"
                        icon="groups"
                        label="View Team Page"
                        @click="goToTeamView"
                        :disable="
                            allPlayers.length === 0 ||
                            !currentDatasetId ||
                            loading
                        "
                        class="q-px-lg"
                    />
                    <q-btn
                        color="secondary"
                        icon="find_replace"
                        label="Find Upgrades"
                        @click="showUpgradeFinder = true"
                        :disable="allPlayers.length === 0 || loading"
                        class="q-px-lg"
                    />
                </div>

                <PlayerDataTable
                    :players="filteredPlayers"
                    :loading="loading"
                    @update:sort="handleSort"
                    @player-selected="handlePlayerSelected"
                    :is-goalkeeper-view="isGoalkeeperView"
                    :currency-symbol="detectedCurrencySymbol"
                    :filtered-player-count="filteredPlayers.length"
                />
            </template>

            <q-card
                v-else-if="!loading && !playerStore.currentDatasetId"
                class="q-pa-lg text-center no-data-card"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-grey-5'
                        : 'bg-grey-1 text-grey-7'
                "
                flat
                bordered
            >
                <q-icon name="upload_file" size="4rem" />
                <div class="text-h6 q-mt-md">No Player Data Yet</div>
                <div>Upload a file to see player data</div>
            </q-card>
            <q-card
                v-else-if="
                    !loading &&
                    playerStore.currentDatasetId &&
                    allPlayers.length === 0
                "
                class="q-pa-lg text-center no-data-card"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-grey-5'
                        : 'bg-grey-1 text-grey-7'
                "
                flat
                bordered
            >
                <q-icon name="sentiment_dissatisfied" size="4rem" />
                <div class="text-h6 q-mt-md">No Players Found</div>
                <div>
                    The uploaded file might not contain player data or an error
                    occurred during parsing.
                </div>
            </q-card>
        </div>

        <PlayerDetailDialog
            :player="selectedPlayer"
            :show="showPlayerDetailDialog"
            @close="showPlayerDetailDialog = false"
            :currency-symbol="detectedCurrencySymbol"
        />
        <UpgradeFinderDialog
            :show="showUpgradeFinder"
            :players="allPlayers"
            @close="showUpgradeFinder = false"
            :currency-symbol="detectedCurrencySymbol"
        />

        <q-dialog v-model="showFileSizeLimitModal" persistent>
            <q-card
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-white'
                        : 'bg-white text-dark'
                "
            >
                <q-card-section class="row items-center">
                    <q-avatar
                        icon="warning"
                        color="negative"
                        text-color="white"
                    />
                    <span class="q-ml-sm text-subtitle1">File Too Large</span>
                </q-card-section>

                <q-card-section class="q-pt-none">
                    Please ensure your HTML export contains 10,000 players or
                    less. (Max file size: 15MB)
                </q-card-section>

                <q-card-actions align="right">
                    <q-btn
                        flat
                        label="OK"
                        color="primary"
                        v-close-popup
                        @click="showFileSizeLimitModal = false"
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>
    </q-page>
</template>

<script>
import { ref, computed, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { usePlayerStore } from "../stores/playerStore";
import PlayerDataTable from "../components/PlayerDataTable.vue";
import PlayerDetailDialog from "../components/PlayerDetailDialog.vue";
import UpgradeFinderDialog from "../components/UpgradeFinderDialog.vue";
import PlayerFilters from "../components/filters/PlayerFilters.vue";

// Define the maximum file size in bytes (15MB)
const MAX_FILE_SIZE_BYTES = 15 * 1024 * 1024;

export default {
    name: "PlayerUploadPage",
    components: {
        PlayerDataTable,
        PlayerDetailDialog,
        UpgradeFinderDialog,
        PlayerFilters,
    },
    setup() {
        const router = useRouter();
        const playerStore = usePlayerStore();
        const $q = useQuasar();
        const playerFile = ref(null);
        const filteredPlayers = ref([]);
        const selectedPlayer = ref(null);
        const showPlayerDetailDialog = ref(false);
        const showUpgradeFinder = ref(false);
        const showFileSizeLimitModal = ref(false); // New ref for controlling the modal

        const allPlayers = computed(() =>
            Array.isArray(playerStore.allPlayers) ? playerStore.allPlayers : [],
        );
        const currentDatasetId = computed(() => playerStore.currentDatasetId);
        const detectedCurrencySymbol = computed(
            () => playerStore.detectedCurrencySymbol,
        );
        const loading = computed(() => playerStore.loading);
        const error = computed({
            get: () => playerStore.error,
            set: (value) => {
                playerStore.error = value;
            },
        });

        const attributeWeightsLoadedForFeedback = ref(false);
        const roleSpecificOverallWeightsLoadedForFeedback = ref(false);

        const activeFilters = ref({});

        const isGoalkeeperView = computed(
            () => activeFilters.value.position === "GK",
        );

        const loadJsonForFeedback = async (filePath, loadedFlagRef) => {
            try {
                const response = await fetch(filePath);
                if (!response.ok)
                    throw new Error(`HTTP error! status: ${response.status}`);
                await response.json();
                loadedFlagRef.value = true;
            } catch (e) {
                console.warn(
                    `Client-side check: Failed to load ${filePath}:`,
                    e,
                );
                loadedFlagRef.value = true; // Still set to true to enable upload button, backend will handle if files are truly missing
            }
        };

        onMounted(async () => {
            await loadJsonForFeedback(
                "/public/attribute_weights.json",
                attributeWeightsLoadedForFeedback,
            );
            await loadJsonForFeedback(
                "/public/role_specific_overall_weights.json",
                roleSpecificOverallWeightsLoadedForFeedback,
            );

            if (!playerStore.currentDatasetId) {
                await playerStore.loadFromSessionStorage();
            } else if (
                playerStore.allPlayers.length === 0 &&
                playerStore.currentDatasetId
            ) {
                // If there's a dataset ID but no players, attempt to fetch them
                await playerStore.fetchPlayersByDatasetId(
                    playerStore.currentDatasetId,
                    activeFilters.value.position, // Pass current filters if any
                    activeFilters.value.role,
                    activeFilters.value.ageRange,
                    activeFilters.value.transferValueRangeLocal,
                );
                if (playerStore.allAvailableRoles.length === 0) {
                    await playerStore.fetchAllAvailableRoles();
                }
            } else if (
                playerStore.currentDatasetId &&
                playerStore.allAvailableRoles.length === 0
            ) {
                // If dataset ID and players exist, but no roles, fetch roles
                await playerStore.fetchAllAvailableRoles();
            }
            // Apply filters if players are loaded
            if (allPlayers.value.length > 0) {
                applyClientSideFilters(allPlayers.value, activeFilters.value);
            }
        });

        const applyClientSideFilters = (playersToFilter, currentFilters) => {
            // Ensure playersToFilter is an array
            if (!Array.isArray(playersToFilter)) {
                filteredPlayers.value = [];
                return;
            }

            let tempPlayers = [...playersToFilter];

            // Name filter
            if (currentFilters.name) {
                tempPlayers = tempPlayers.filter(
                    (p) =>
                        p.name &&
                        p.name
                            .toLowerCase()
                            .includes(currentFilters.name.toLowerCase()),
                );
            }
            // Club filter
            if (currentFilters.club) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.club === currentFilters.club,
                );
            }
            // Nationality filter
            if (currentFilters.nationality) {
                tempPlayers = tempPlayers.filter(
                    (p) => p.nationality === currentFilters.nationality,
                );
            }
            // Media Handling filter
            if (
                currentFilters.mediaHandling &&
                currentFilters.mediaHandling.length > 0
            ) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.media_handling) return false;
                    const playerStyles = p.media_handling
                        .split(",")
                        .map((s) => s.trim().toLowerCase());
                    const filterStylesLower = currentFilters.mediaHandling.map(
                        (s) => s.toLowerCase(),
                    );
                    return playerStyles.some((style) =>
                        filterStylesLower.includes(style),
                    );
                });
            }
            // Personality filter
            if (
                currentFilters.personality &&
                currentFilters.personality.length > 0
            ) {
                tempPlayers = tempPlayers.filter((p) => {
                    if (!p.personality) return false;
                    return currentFilters.personality.includes(p.personality);
                });
            }

            // Age range filter
            if (
                currentFilters.ageRange &&
                typeof currentFilters.ageRange.min === "number" &&
                typeof currentFilters.ageRange.max === "number"
            ) {
                // Only filter if not default min
                if (
                    currentFilters.ageRange.min >
                    playerStore.AGE_SLIDER_MIN_DEFAULT
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) => p.age >= currentFilters.ageRange.min,
                    );
                }
                // Only filter if not default max
                if (
                    currentFilters.ageRange.max <
                    playerStore.AGE_SLIDER_MAX_DEFAULT
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) => p.age <= currentFilters.ageRange.max,
                    );
                }
            }

            // Transfer value range filter
            if (
                currentFilters.transferValueRangeLocal &&
                playerStore.transferValueRange && // Ensure store range is available
                typeof currentFilters.transferValueRangeLocal.min ===
                    "number" &&
                typeof currentFilters.transferValueRangeLocal.max === "number"
            ) {
                // Only filter if not default min
                if (
                    currentFilters.transferValueRangeLocal.min >
                    playerStore.transferValueRange.min
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            (p.transferValueAmount || 0) >=
                            currentFilters.transferValueRangeLocal.min,
                    );
                }
                // Only filter if not default max
                if (
                    currentFilters.transferValueRangeLocal.max <
                    playerStore.transferValueRange.max
                ) {
                    tempPlayers = tempPlayers.filter(
                        (p) =>
                            (p.transferValueAmount || 0) <=
                            currentFilters.transferValueRangeLocal.max,
                    );
                }
            }
            filteredPlayers.value = tempPlayers;
        };

        const uploadAndParse = async () => {
            if (!playerFile.value) {
                playerStore.error = "Please select an HTML file first.";
                return;
            }

            // Client-side file size check
            if (playerFile.value.size > MAX_FILE_SIZE_BYTES) {
                showFileSizeLimitModal.value = true;
                return; // Stop the upload process
            }

            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                await playerStore.uploadPlayerFile(formData);
                activeFilters.value = {}; // Reset filters on new upload
                // Clear any previous specific error notifications if successful
                if (!playerStore.error) {
                    $q.notify({
                        type: "positive",
                        message: "File uploaded and parsed successfully!",
                        position: "top",
                        timeout: 3000,
                    });
                }
            } catch (e) {
                console.error("Upload and Parse error in page:", e);
                // Error is already set in the store by uploadPlayerFile action
                // The q-banner will display it.
                // If the error was due to file size (caught by backend as a fallback),
                // the store should already have the correct message.
                if (playerStore.error) {
                    $q.notify({
                        type: "negative",
                        message: playerStore.error,
                        position: "top",
                        timeout: 5000,
                        actions: [
                            {
                                label: "Dismiss",
                                color: "white",
                                handler: () => {
                                    /* ... */
                                },
                            },
                        ],
                    });
                }
            }
        };

        const handleSort = (sortParams) => {
            // The PlayerDataTable handles its own internal sorting display.
            // This event is here if we need to react to sort changes at the page level.
            console.log(
                "PlayerUploadPage: Sort requested by PlayerDataTable:",
                sortParams,
            );
            // If server-side sorting were implemented, this is where you'd trigger a fetch.
        };

        const handlePlayerSelected = (player) => {
            selectedPlayer.value = player;
            showPlayerDetailDialog.value = true;
        };

        const handleFilterChanged = async (newFilters) => {
            activeFilters.value = newFilters;
            // When filters change, if we have a dataset ID, we should refetch from the backend
            // as the backend might do some pre-filtering or adjustments based on position/role.
            // If no dataset ID, it means we are working with locally loaded (but not yet uploaded/processed) data,
            // so we apply filters client-side.
            if (playerStore.currentDatasetId) {
                await playerStore.fetchPlayersByDatasetId(
                    playerStore.currentDatasetId,
                    newFilters.position,
                    newFilters.role,
                    newFilters.ageRange,
                    newFilters.transferValueRangeLocal,
                );
                // After fetching, allPlayers in store is updated, watcher will apply client-side filters.
            } else {
                // This case should be rare if uploads always lead to a datasetId.
                // But as a fallback, apply client-side if no datasetId.
                applyClientSideFilters(allPlayers.value, newFilters);
            }
        };

        // Watch for changes in allPlayers (e.g., after fetch) and re-apply client-side filters
        watch(
            allPlayers,
            (newVal) => {
                applyClientSideFilters(newVal, activeFilters.value);
            },
            { immediate: true }, // Apply on initial load too
        );

        const goToTeamView = () => {
            if (playerStore.currentDatasetId) {
                router.push({
                    name: "team-view",
                    query: { datasetId: playerStore.currentDatasetId },
                });
            } else {
                // This should ideally not happen if button is disabled correctly
                playerStore.error =
                    "No data uploaded yet. Please upload a file first.";
            }
        };

        return {
            $q,
            playerFile,
            playerStore, // Expose store for template access if needed
            loading,
            error,
            allPlayers,
            filteredPlayers,
            uploadAndParse,
            handleSort,
            selectedPlayer,
            showPlayerDetailDialog,
            handlePlayerSelected,
            attributeWeightsLoadedForFeedback, // For disabling upload button
            roleSpecificOverallWeightsLoadedForFeedback, // For disabling upload button
            showUpgradeFinder,
            isGoalkeeperView, // For PlayerDataTable prop
            goToTeamView,
            currentDatasetId, // For disabling team view button
            detectedCurrencySymbol, // For PlayerDataTable prop
            handleFilterChanged, // For PlayerFilters component
            showFileSizeLimitModal, // For the new modal
        };
    },
};
</script>

<style lang="scss" scoped>
.q-page {
    max-width: 1600px; /* Or your preferred max width */
    margin: 0 auto;
    padding-top: 24px; /* Add some padding at the top */
    padding-bottom: 24px; /* Add some padding at the bottom */
}

.page-title {
    // Styling for the main page title if needed
}

.instructions-card {
    ol {
        padding-left: 20px; /* Standard padding for ordered lists */
        li {
            margin-bottom: 0.5em; /* Space between list items */
        }
    }
}

.upload-card,
.no-data-card {
    border-radius: 8px; /* Consistent rounded corners */
}

// Ensure the q-file hint text is visible in dark mode
.body--dark .q-field__bottom {
    color: rgba(255, 255, 255, 0.5) !important;
}
</style>
```go // src/api/handlers.go package main import ( "encoding/json" "log"
"net/http" "runtime" // Standard Go runtime package "sort" "strconv" "strings"
"sync" "time" "[github.com/google/uuid](https://github.com/google/uuid)" ) const
( // MaxUploadSize defines the maximum allowed file size for uploads (15MB) //
This is an approximation for about 10,000 players. MaxUploadSize = 15 * 1024 *
1024 // User-facing error message for file size limit. FileSizeLimitErrorMessage
= "Only 10,000 players or less can be in a given dataset. (Max file size: 15MB)"
) // uploadHandler handles POST requests for uploading HTML player files. // It
parses the file, processes player data concurrently, and stores the results.
func uploadHandler(w http.ResponseWriter, r *http.Request) { if r.Method !=
http.MethodPost { http.Error(w, "Only POST method is allowed",
http.StatusMethodNotAllowed) return } startTime := time.Now() // Check
Content-Length header first for a quick check, though it can be spoofed. //
r.ContentLength is an int64 if r.ContentLength > MaxUploadSize {
log.Printf("Upload rejected: Content-Length (%d bytes) exceeds limit (%d
bytes)", r.ContentLength, MaxUploadSize) http.Error(w,
FileSizeLimitErrorMessage, http.StatusRequestEntityTooLarge) return } //
ParseMultipartForm will also respect the maxMemory argument for in-memory parts,
// but the total request size is what we're primarily concerned with for the
file part. // We'll check the actual file handler size after getting the file.
if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB for other form
data, not the file itself immediately http.Error(w, "Error parsing multipart
form: "+err.Error(), http.StatusBadRequest) return } file, handler, err :=
r.FormFile("playerFile") if err != nil { http.Error(w, "Error retrieving the
file: "+err.Error(), http.StatusBadRequest) return } defer file.Close() fileSize
:= handler.Size log.Printf("Uploaded File: %s (Size: %d bytes)",
handler.Filename, fileSize) // Enforce the 15MB limit on the actual file size if
fileSize > MaxUploadSize { log.Printf("Upload rejected: Actual file size (%d
bytes) exceeds limit (%d bytes)", fileSize, MaxUploadSize) http.Error(w,
FileSizeLimitErrorMessage, http.StatusRequestEntityTooLarge) return }
parseStartTime := time.Now() playersList := make([]Player, 0,
defaultPlayerCapacity) // Assumes defaultPlayerCapacity is defined in config.go
var processingError error numWorkers := runtime.NumCPU() if numWorkers == 0 {
numWorkers = 1 } const rowBufferMultiplier = 10 rowCellsChan := make(chan
[]string, numWorkers*rowBufferMultiplier) resultsChan := make(chan
PlayerParseResult, numWorkers*rowBufferMultiplier) var wg sync.WaitGroup var
headersSnapshot []string doneConsumingResults := make(chan struct{}) go func() {
defer close(doneConsumingResults) for result := range resultsChan { if
result.Err == nil { playersList = append(playersList, result.Player) } else {
log.Printf("Skipping row due to error from worker: %v", result.Err) } }
log.Println("Finished collecting results from resultsChan.") }() processingError
= ParseHTMLPlayerTable(file, &headersSnapshot, rowCellsChan, numWorkers,
resultsChan, &wg) // Assumes ParseHTMLPlayerTable is in parsing.go
close(rowCellsChan) log.Println("Row cells channel closed (HTML parsing attempt
finished).") if processingError != nil { log.Printf("Error during HTML parsing
or worker setup: %v", processingError) if len(headersSnapshot) > 0 {
log.Println("Waiting for any potentially started workers after parsing
error...") wg.Wait() } close(resultsChan) <-doneConsumingResults http.Error(w,
processingError.Error(), http.StatusInternalServerError) return } if
len(headersSnapshot) == 0 { log.Println("Critical: No headers were parsed from
the HTML file.") close(resultsChan) <-doneConsumingResults http.Error(w, "Could
not parse table headers, no data processed.", http.StatusInternalServerError)
return } log.Println("Waiting for all player data parser workers to finish...")
wg.Wait() log.Println("All workers have completed (wg.Wait() returned).")
close(resultsChan) log.Println("ResultsChan closed after all workers finished.")
<-doneConsumingResults log.Println("Results consumer goroutine finished
processing all items.") finalDatasetCurrencySymbol := "$" // Default if
len(playersList) > 0 { var foundSymbol bool for _, p := range playersList { _,
_, tvSymbol := ParseMonetaryValueGo(p.TransferValue) // Assumes
ParseMonetaryValueGo is in parsing.go if tvSymbol != "" {
finalDatasetCurrencySymbol = tvSymbol foundSymbol = true break } _, _, wSymbol
:= ParseMonetaryValueGo(p.Wage) if wSymbol != "" { finalDatasetCurrencySymbol =
wSymbol foundSymbol = true break } } if !foundSymbol { log.Println("No currency
symbol detected from parsed player monetary values, using default '$'.") } } if
len(playersList) > 0 { log.Println("Calculating player performance
percentiles...") CalculatePlayerPerformancePercentiles(playersList) // Assumes
CalculatePlayerPerformancePercentiles is in performance_stats.go
log.Println("Finished calculating percentiles.") } parseDuration :=
time.Since(parseStartTime) datasetID := uuid.New().String() // Assumes
playerDataStore and storeMutex are defined in store.go storeMutex.Lock()
playerDataStore[datasetID] = struct { Players []Player CurrencySymbol string
}{Players: playersList, CurrencySymbol: finalDatasetCurrencySymbol}
storeMutex.Unlock() log.Printf("Stored %d players with DatasetID: %s. Detected
Currency: %s", len(playersList), datasetID, finalDatasetCurrencySymbol) if
len(playersList) > 0 { log.Printf("DEBUG: Sample Player 1 after all processing:
Name='%s', Overall=%d, ParsedPositions=%v, ShortPositions=%v,
PositionGroups=%v", playersList[0].Name, playersList[0].Overall,
playersList[0].ParsedPositions, playersList[0].ShortPositions,
playersList[0].PositionGroups) } else { log.Println("No players were
successfully parsed or list is empty after processing.") } response :=
UploadResponse{DatasetID: datasetID, Message: "File uploaded and parsed
successfully.", DetectedCurrencySymbol: finalDatasetCurrencySymbol}
w.Header().Set("Content-Type", "application/json")
w.Header().Set("Access-Control-Allow-Origin", "*") // Ensure CORS is set if
frontend is on a different domain/port if err :=
json.NewEncoder(w).Encode(response); err != nil { log.Printf("Error encoding
JSON response for upload: %v", err) http.Error(w, "Error encoding JSON response:
"+err.Error(), http.StatusInternalServerError) } var memStats runtime.MemStats
runtime.ReadMemStats(&memStats) rowsPerSecond := 0.0 if parseDuration.Seconds()
> 0 { rowsPerSecond = float64(len(playersList)) / parseDuration.Seconds() }
totalDuration := time.Since(startTime) // Assumes BToMb is defined in utils.go
log.Printf("--- Perf Metrics --- File: %s, Size: %d KB, Total Time: %v, Parse
Time: %v, Parsed Players: %d, Rows/Sec: %.2f, MemAlloc: %.2f MiB, Workers: %d,
Goroutines: %d ---", handler.Filename, fileSize/1024, totalDuration,
parseDuration, len(playersList), rowsPerSecond, BToMb(memStats.Alloc),
numWorkers, runtime.NumGoroutine()) } // playerDataHandler handles GET requests
to retrieve processed player data by dataset ID. func playerDataHandler(w
http.ResponseWriter, r *http.Request) { if r.Method != http.MethodGet {
http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed) return
} pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"),
"/") if len(pathParts) == 0 || pathParts[0] == "" { http.Error(w, "Dataset ID is
missing in the request path", http.StatusBadRequest) return } datasetID :=
pathParts[0] queryValues := r.URL.Query() filterPosition :=
queryValues.Get("position") filterRole := queryValues.Get("role") minAgeStr :=
queryValues.Get("minAge") maxAgeStr := queryValues.Get("maxAge")
minTransferValueStr := queryValues.Get("minTransferValue") maxTransferValueStr
:= queryValues.Get("maxTransferValue") log.Printf("playerDataHandler:
DatasetID=%s, PositionFilter=%s, RoleFilter=%s, MinAge=%s, MaxAge=%s, MinVal=%s,
MaxVal=%s", datasetID, filterPosition, filterRole, minAgeStr, maxAgeStr,
minTransferValueStr, maxTransferValueStr) storeMutex.RLock() // Assumes
storeMutex is defined in store.go data, found := playerDataStore[datasetID] //
Assumes playerDataStore is defined in store.go storeMutex.RUnlock() if !found {
log.Printf("Player data not found for DatasetID: %s", datasetID) http.Error(w,
"Player data not found for the given ID.", http.StatusNotFound) return }
processedPlayers := make([]Player, 0, len(data.Players)) var minAge, maxAge int
= -1, -1 var minTransferValue, maxTransferValue int64 = -1, -1 if val, err :=
strconv.Atoi(minAgeStr); err == nil { minAge = val } if val, err :=
strconv.Atoi(maxAgeStr); err == nil { maxAge = val } if val, err :=
strconv.ParseInt(minTransferValueStr, 10, 64); err == nil { minTransferValue =
val } if val, err := strconv.ParseInt(maxTransferValueStr, 10, 64); err == nil {
maxTransferValue = val } for _, p := range data.Players { playerCopy := p if
filterPosition != "" { canPlayPosition := false for _, shortPos := range
playerCopy.ShortPositions { if shortPos == filterPosition { canPlayPosition =
true break } } if !canPlayPosition { continue } } playerAgeVal, ageErr :=
strconv.Atoi(playerCopy.Age) if ageErr == nil { if minAge != -1 && playerAgeVal
< minAge { continue } if maxAge != -1 && playerAgeVal > maxAge { continue } }
else if minAge != -1 || maxAge != -1 { log.Printf("Skipping player %s due to
unparsable age '%s' while age filters are active.", playerCopy.Name,
playerCopy.Age) continue } if minTransferValue != -1 &&
playerCopy.TransferValueAmount < minTransferValue { continue } if
maxTransferValue != -1 && playerCopy.TransferValueAmount > maxTransferValue {
continue } if filterRole != "" { roleMatched := false for _, roleOverall :=
range playerCopy.RoleSpecificOveralls { if roleOverall.RoleName == filterRole {
playerCopy.Overall = roleOverall.Score // Update player's main overall to the
role-specific one for display roleMatched = true break } } if !roleMatched { //
If the role filter is active but the player doesn't have that specific role
calculated, // you might want to either exclude them or set their overall to 0
or a special value. // For now, let's assume if a role filter is active, we only
want players matching that role's overall. // So, if not matched, their
'Overall' (which might be their best general overall) might not be relevant. //
Depending on requirements, you might set playerCopy.Overall = 0 or skip. // For
this example, if a role is filtered, we are primarily interested in the players'
score for THAT role. // If they don't have that role, they are effectively 0 for
that role. // However, the primary filtering for display happens on the frontend
based on the RoleSpecificOveralls array. // The backend here is just adjusting
the main 'Overall' field if a role is specified. // If no match, the original
player.Overall (best general) remains. } } processedPlayers =
append(processedPlayers, playerCopy) } log.Printf("playerDataHandler: Returning
%d players after processing for DatasetID=%s", len(processedPlayers), datasetID)
response := PlayerDataWithCurrency{Players: processedPlayers, CurrencySymbol:
data.CurrencySymbol} w.Header().Set("Content-Type", "application/json")
w.Header().Set("Access-Control-Allow-Origin", "*") // Ensure CORS if err :=
json.NewEncoder(w).Encode(response); err != nil { log.Printf("Error encoding
JSON response for playerData (DatasetID: %s): %v", datasetID, err) } } //
rolesHandler returns a list of all available role names. func rolesHandler(w
http.ResponseWriter, r *http.Request) { if r.Method != http.MethodGet {
http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed) return
} // Assumes muRoleSpecificOverallWeights and roleSpecificOverallWeights are
defined in config.go muRoleSpecificOverallWeights.RLock() roleNames :=
make([]string, 0, len(roleSpecificOverallWeights)) for roleName := range
roleSpecificOverallWeights { roleNames = append(roleNames, roleName) }
muRoleSpecificOverallWeights.RUnlock() sort.Strings(roleNames) // Sort for
consistent frontend display w.Header().Set("Content-Type", "application/json")
w.Header().Set("Access-Control-Allow-Origin", "*") // Ensure CORS if err :=
json.NewEncoder(w).Encode(roleNames); err != nil { log.Printf("Error encoding
JSON response for roles: %v", err) http.Error(w, "Error encoding JSON response",
http.StatusInternalServerError) } } ```js // src/stores/playerStore.js import {
defineStore } from "pinia"; import { ref, computed, shallowRef } from "vue";
import playerService from "../services/playerService"; export const
usePlayerStore = defineStore("player", () => { const allPlayers =
shallowRef([]); const currentDatasetId = ref(
sessionStorage.getItem("currentDatasetId") || null, ); const
detectedCurrencySymbol = ref( sessionStorage.getItem("detectedCurrencySymbol")
|| "$", ); const loading = ref(false); const error = ref(""); const
allAvailableRoles = ref([]); // Default age slider values, can be accessed by
components const AGE_SLIDER_MIN_DEFAULT = 15; const AGE_SLIDER_MAX_DEFAULT = 50;
const uniqueClubs = computed(() => { if (!Array.isArray(allPlayers.value) ||
allPlayers.value.length === 0) return []; const clubs = new Set();
allPlayers.value.forEach((p) => { if (p.club) clubs.add(p.club); }); return
Array.from(clubs).sort(); }); const uniqueNationalities = computed(() => { if
(!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return [];
const nationalities = new Set(); allPlayers.value.forEach((p) => { if
(p.nationality) nationalities.add(p.nationality); }); return
Array.from(nationalities).sort(); }); const uniqueMediaHandlings = computed(()
=> { if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0)
return []; const mediaHandlingsIndividual = new Set();
allPlayers.value.forEach((p) => { if (p.media_handling) {
p.media_handling.split(",").forEach((style) => { const trimmedStyle =
style.trim(); if (trimmedStyle) mediaHandlingsIndividual.add(trimmedStyle); });
} }); return Array.from(mediaHandlingsIndividual).sort(); }); const
uniquePersonalities = computed(() => { if (!Array.isArray(allPlayers.value) ||
allPlayers.value.length === 0) return []; const personalities = new Set();
allPlayers.value.forEach((p) => { if (p.personality)
personalities.add(p.personality); }); return Array.from(personalities).sort();
}); const uniquePositionsCount = computed(() => { if
(!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) return 0;
const s = new Set(); allPlayers.value.forEach((player) =>
player.parsedPositions?.forEach((pos) => s.add(pos)), ); return s.size; });
const transferValueRange = computed(() => { if (!Array.isArray(allPlayers.value)
|| allPlayers.value.length === 0) { return { min: 0, max: 100000000 }; } const
transferValuesNumeric = allPlayers.value .filter((p) => typeof
p.transferValueAmount === "number") .map((p) => p.transferValueAmount); if
(transferValuesNumeric.length === 0) return { min: 0, max: 100000000 }; const
min = Math.min(0, ...transferValuesNumeric); let max =
Math.max(...transferValuesNumeric); if (min >= max) max = min + 50000; // Ensure
max is always greater than min for range slider if (min === 0 && max === 0 &&
transferValuesNumeric.some((v) => v === 0)) max = 50000; // Handle case where
all values are 0 return { min, max }; }); async function
uploadPlayerFile(formData) { loading.value = true; error.value = ""; try { const
response = await playerService.uploadPlayerFile(formData);
currentDatasetId.value = response.datasetId; detectedCurrencySymbol.value =
response.detectedCurrencySymbol || "$";
sessionStorage.setItem("currentDatasetId", currentDatasetId.value);
sessionStorage.setItem( "detectedCurrencySymbol", detectedCurrencySymbol.value,
); await fetchPlayersByDatasetId(currentDatasetId.value); // Fetch all players
for the new dataset await fetchAllAvailableRoles(); // Fetch roles for the new
dataset return response; } catch (e) { // Use the error message directly from
the service if it's a 413 or other specific error if (e.status === 413 &&
e.message) { error.value = e.message; // Use the specific message from the
backend } else { error.value = `Failed to process file: ${e.message || "Unknown
error"}`; } resetState(); // Reset state on error throw e; // Re-throw for
component to potentially handle further } finally { loading.value = false; } }
async function fetchPlayersByDatasetId( datasetId, positionFilter = null,
roleFilter = null, ageRangeFilter = null, transferValueRangeFilter = null, ) {
if (!datasetId) { resetState(); // Clear data if no datasetId return; }
loading.value = true; error.value = ""; try { console.log( `playerStore:
Fetching players for datasetId: ${datasetId}, Pos: ${positionFilter}, Role:
${roleFilter}, Age: ${JSON.stringify(ageRangeFilter)}, Val:
${JSON.stringify(transferValueRangeFilter)}`, ); const response = await
playerService.getPlayersByDatasetId( datasetId, positionFilter, roleFilter,
ageRangeFilter, transferValueRangeFilter, ); allPlayers.value =
processPlayersFromAPI(response.players); detectedCurrencySymbol.value =
response.currencySymbol || "$"; // Update currency symbol from response
sessionStorage.setItem( "detectedCurrencySymbol", detectedCurrencySymbol.value,
); // Persist it return response; // Return the full response if needed by
caller } catch (e) { error.value = `Failed to fetch player data: ${e.message ||
"Unknown error"}`; resetState(); // Clear data on error throw e; } finally {
loading.value = false; } } async function fetchAllAvailableRoles(force = false)
{ // Fetches all unique role names available in the current dataset (from
backend) if (allAvailableRoles.value.length > 0 && !force) return; // Avoid
refetch if already populated unless forced try { const roles = await
playerService.getAvailableRoles(); allAvailableRoles.value = roles.sort(); //
Sort for consistent display } catch (e) { console.error("playerStore: Error
fetching available roles:", e); allAvailableRoles.value = []; // Reset or handle
error appropriately } } function processPlayersFromAPI(playersData) { if
(!Array.isArray(playersData)) return []; // Ensure age is an integer, default to
0 if not parsable return playersData.map((p) => ({ ...p, age: parseInt(p.age,
10) || 0, // Other per-player processing can go here if needed })); } function
resetState() { allPlayers.value = []; currentDatasetId.value = null;
detectedCurrencySymbol.value = "$"; // Reset to default allAvailableRoles.value
= []; sessionStorage.removeItem("currentDatasetId");
sessionStorage.removeItem("detectedCurrencySymbol"); // Do not clear error.value
here, let components decide } async function loadFromSessionStorage() { const
storedDatasetId = sessionStorage.getItem("currentDatasetId"); const
storedCurrencySymbol = sessionStorage.getItem( "detectedCurrencySymbol", ); if
(storedDatasetId) { currentDatasetId.value = storedDatasetId; if
(storedCurrencySymbol) { detectedCurrencySymbol.value = storedCurrencySymbol; }
try { // Fetch players and roles when loading from session await
fetchPlayersByDatasetId(storedDatasetId); await fetchAllAvailableRoles(); }
catch (e) { // Error is handled within fetchPlayersByDatasetId and
fetchAllAvailableRoles // If loading from session fails, the state will be reset
by those functions. } } else { resetState(); // If no dataset ID in session,
ensure clean state } } return { allPlayers, currentDatasetId,
detectedCurrencySymbol, loading, error, uniqueClubs, uniqueNationalities,
uniqueMediaHandlings, uniquePersonalities, uniquePositionsCount,
transferValueRange, allAvailableRoles, uploadPlayerFile,
fetchPlayersByDatasetId, fetchAllAvailableRoles, resetState,
loadFromSessionStorage, AGE_SLIDER_MIN_DEFAULT, // Expose defaults
AGE_SLIDER_MAX_DEFAULT, }; });
