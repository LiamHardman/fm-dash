// src/stores/playerStore.js
import { defineStore } from "pinia";
import { ref, computed, shallowRef } from "vue";
import playerService from "../services/playerService";

export const usePlayerStore = defineStore("player", () => {
  // State
  const allPlayers = shallowRef([]); // Use shallowRef for the large player list
  const currentDatasetId = ref(
    sessionStorage.getItem("currentDatasetId") || null,
  );
  const detectedCurrencySymbol = ref(
    sessionStorage.getItem("detectedCurrencySymbol") || "$",
  );
  const loading = ref(false);
  const error = ref("");
  const allAvailableRoles = ref([]); // New state for roles

  console.log(
    `playerStore: Initializing. DatasetID from session: ${currentDatasetId.value}`,
  );

  // Getters
  const uniqueClubs = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0)
      return [];
    const clubs = new Set();
    allPlayers.value.forEach((p) => {
      if (p.club) clubs.add(p.club);
    });
    return Array.from(clubs).sort();
  });

  const uniqueNationalities = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0)
      return [];
    const nationalities = new Set();
    allPlayers.value.forEach((p) => {
      if (p.nationality) nationalities.add(p.nationality);
    });
    return Array.from(nationalities).sort();
  });

  const uniqueMediaHandlings = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0)
      return [];
    const mediaHandlingsIndividual = new Set();
    allPlayers.value.forEach((p) => {
      if (p.media_handling) {
        p.media_handling.split(",").forEach((style) => {
          const trimmedStyle = style.trim();
          if (trimmedStyle) mediaHandlingsIndividual.add(trimmedStyle);
        });
      }
    });
    return Array.from(mediaHandlingsIndividual).sort();
  });

  const uniquePersonalities = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0)
      return [];
    const personalities = new Set();
    allPlayers.value.forEach((p) => {
      if (p.personality) personalities.add(p.personality);
    });
    return Array.from(personalities).sort();
  });

  const uniquePositionsCount = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0)
      return 0;
    const s = new Set();
    allPlayers.value.forEach((player) =>
      player.parsedPositions?.forEach((pos) => s.add(pos)),
    );
    return s.size;
  });

  const transferValueRange = computed(() => {
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      return { min: 0, max: 100000000 }; // Default range if no players
    }
    const transferValuesNumeric = allPlayers.value
      .filter((p) => typeof p.transferValueAmount === "number")
      .map((p) => p.transferValueAmount);

    if (transferValuesNumeric.length === 0) return { min: 0, max: 100000000 }; // Default if no numeric values

    const min = Math.min(0, ...transferValuesNumeric); // Ensure min is at least 0
    let max = Math.max(...transferValuesNumeric);

    // Ensure max is greater than min, and provide a sensible default if all values are 0
    if (min >= max) {
      max = min + 50000; // Add a small amount if min and max are same
    }
    if (min === 0 && max === 0 && transferValuesNumeric.some((v) => v === 0)) {
      // If all values are 0
      max = 50000; // Default max if all values are 0
    }
    return { min, max };
  });

  // Actions
  async function uploadPlayerFile(formData) {
    loading.value = true;
    error.value = "";
    try {
      const response = await playerService.uploadPlayerFile(formData);
      currentDatasetId.value = response.datasetId;
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || "$";
      sessionStorage.setItem("currentDatasetId", currentDatasetId.value);
      sessionStorage.setItem(
        "detectedCurrencySymbol",
        detectedCurrencySymbol.value,
      );

      // Fetch initial player data (no filters) and roles
      await fetchPlayersByDatasetId(currentDatasetId.value);
      await fetchAllAvailableRoles();
      return response;
    } catch (e) {
      error.value = `Failed to process file: ${e.message || "Unknown error"}`;
      resetState(); // Clear data on error
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function fetchPlayersByDatasetId(
    datasetId,
    positionFilter = null,
    roleFilter = null,
  ) {
    if (!datasetId) {
      console.warn(
        "playerStore: fetchPlayersByDatasetId called with no datasetId.",
      );
      resetState();
      return;
    }
    loading.value = true;
    error.value = "";
    try {
      console.log(
        `playerStore: Fetching players for datasetId: ${datasetId}, Position: ${positionFilter}, Role: ${roleFilter}`,
      );
      const response = await playerService.getPlayersByDatasetId(
        datasetId,
        positionFilter,
        roleFilter,
      );

      // Update players list. shallowRef requires a new array assignment to trigger updates.
      allPlayers.value = processPlayersFromAPI(response.players);
      detectedCurrencySymbol.value = response.currencySymbol || "$";
      // Update currency in session storage as it might change per dataset (though unlikely in this app's current design)
      sessionStorage.setItem(
        "detectedCurrencySymbol",
        detectedCurrencySymbol.value,
      );

      console.log(
        `playerStore: Fetched and processed ${allPlayers.value.length} players.`,
      );
      return response;
    } catch (e) {
      error.value = `Failed to fetch player data for dataset ${datasetId}: ${e.message || "Unknown error"}`;
      resetState();
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function fetchAllAvailableRoles(force = false) {
    // Only fetch if roles are not already populated or if forced
    if (allAvailableRoles.value.length > 0 && !force) {
      console.log("playerStore: Roles already loaded, skipping fetch.");
      return;
    }
    try {
      console.log(`playerStore: Fetching all available roles.`);
      const roles = await playerService.getAvailableRoles();
      allAvailableRoles.value = roles.sort(); // Sort roles alphabetically
      console.log(
        `playerStore: Fetched ${allAvailableRoles.value.length} roles.`,
      );
    } catch (e) {
      console.error("playerStore: Error fetching available roles:", e);
      // Optionally set an error state for roles or just leave it empty
      allAvailableRoles.value = []; // Set to empty on error
      // Do not set global error.value here, as it might overwrite a more critical player data error
    }
  }

  function processPlayersFromAPI(playersData) {
    if (!Array.isArray(playersData)) return []; // Ensure it's an array
    return playersData.map((p) => ({
      ...p,
      age: parseInt(p.age, 10) || 0, // Ensure age is a number
      // Any other client-side processing can happen here
    }));
  }

  function resetState() {
    console.log(`playerStore: resetState called.`);
    allPlayers.value = [];
    currentDatasetId.value = null;
    detectedCurrencySymbol.value = "$";
    allAvailableRoles.value = []; // Also reset roles
    sessionStorage.removeItem("currentDatasetId");
    sessionStorage.removeItem("detectedCurrencySymbol");
    // No session storage for allAvailableRoles as it's fetched on demand
  }

  async function loadFromSessionStorage() {
    console.time("playerStore_loadFromSessionStorage_action");
    const storedDatasetId = sessionStorage.getItem("currentDatasetId");
    const storedCurrencySymbol = sessionStorage.getItem(
      "detectedCurrencySymbol",
    );
    console.log(
      `playerStore: loadFromSessionStorage. Stored datasetId: ${storedDatasetId}, Stored currency: ${storedCurrencySymbol}`,
    );

    if (storedDatasetId) {
      currentDatasetId.value = storedDatasetId;
      if (storedCurrencySymbol) {
        detectedCurrencySymbol.value = storedCurrencySymbol;
      }
      // Fetch initial player data (no filters) and roles
      try {
        // Fetch players first
        await fetchPlayersByDatasetId(storedDatasetId);
        // Then fetch roles, as roles might depend on a valid dataset context (though currently they don't)
        await fetchAllAvailableRoles();
      } catch (e) {
        console.error("Error loading from session storage in playerStore:", e);
        // Error is handled within fetchPlayersByDatasetId, which calls resetState
      }
    } else {
      resetState(); // Ensure clean state if no datasetId in session
    }
    console.timeEnd("playerStore_loadFromSessionStorage_action");
  }

  return {
    allPlayers,
    currentDatasetId,
    detectedCurrencySymbol,
    loading,
    error,
    uniqueClubs,
    uniqueNationalities,
    uniqueMediaHandlings,
    uniquePersonalities,
    uniquePositionsCount,
    transferValueRange,
    allAvailableRoles, // Expose roles
    uploadPlayerFile,
    fetchPlayersByDatasetId,
    fetchAllAvailableRoles, // Expose action to fetch roles
    resetState,
    loadFromSessionStorage,
  };
});
