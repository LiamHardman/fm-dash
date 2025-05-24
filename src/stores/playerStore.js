// src/stores/playerStore.js
import { defineStore } from "pinia";
import { ref, computed, shallowRef } from "vue";
import playerService from "../services/playerService";

export const usePlayerStore = defineStore("player", () => {
  const allPlayers = shallowRef([]);
  const currentDatasetId = ref(
    sessionStorage.getItem("currentDatasetId") || null,
  );
  const detectedCurrencySymbol = ref(
    sessionStorage.getItem("detectedCurrencySymbol") || "$",
  );
  const loading = ref(false);
  const error = ref("");
  const allAvailableRoles = ref([]);

  // Default age slider values, can be accessed by components
  const AGE_SLIDER_MIN_DEFAULT = 15;
  const AGE_SLIDER_MAX_DEFAULT = 50;

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
      return { min: 0, max: 100000000 };
    }
    const transferValuesNumeric = allPlayers.value
      .filter((p) => typeof p.transferValueAmount === "number")
      .map((p) => p.transferValueAmount);
    if (transferValuesNumeric.length === 0) return { min: 0, max: 100000000 };
    const min = Math.min(0, ...transferValuesNumeric);
    let max = Math.max(...transferValuesNumeric);
    if (min >= max) max = min + 50000; // Ensure max is always greater than min for range slider
    if (min === 0 && max === 0 && transferValuesNumeric.some((v) => v === 0))
      max = 50000; // Handle case where all values are 0
    return { min, max };
  });

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
      await fetchPlayersByDatasetId(currentDatasetId.value); // Fetch all players for the new dataset
      await fetchAllAvailableRoles(); // Fetch roles for the new dataset
      return response;
    } catch (e) {
      // Use the error message directly from the service if it's a 413 or other specific error
      if (e.status === 413 && e.message) {
        error.value = e.message; // Use the specific message from the backend
      } else {
        error.value = `Failed to process file: ${e.message || "Unknown error"}`;
      }
      resetState(); // Reset state on error
      throw e; // Re-throw for component to potentially handle further
    } finally {
      loading.value = false;
    }
  }

  async function fetchPlayersByDatasetId(
    datasetId,
    positionFilter = null,
    roleFilter = null,
    ageRangeFilter = null,
    transferValueRangeFilter = null,
  ) {
    if (!datasetId) {
      resetState(); // Clear data if no datasetId
      return;
    }
    loading.value = true;
    error.value = "";
    try {
      console.log(
        `playerStore: Fetching players for datasetId: ${datasetId}, Pos: ${positionFilter}, Role: ${roleFilter}, Age: ${JSON.stringify(ageRangeFilter)}, Val: ${JSON.stringify(transferValueRangeFilter)}`,
      );
      const response = await playerService.getPlayersByDatasetId(
        datasetId,
        positionFilter,
        roleFilter,
        ageRangeFilter,
        transferValueRangeFilter,
      );
      allPlayers.value = processPlayersFromAPI(response.players);
      detectedCurrencySymbol.value = response.currencySymbol || "$"; // Update currency symbol from response
      sessionStorage.setItem(
        "detectedCurrencySymbol",
        detectedCurrencySymbol.value,
      ); // Persist it
      return response; // Return the full response if needed by caller
    } catch (e) {
      error.value = `Failed to fetch player data: ${e.message || "Unknown error"}`;
      resetState(); // Clear data on error
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function fetchAllAvailableRoles(force = false) {
    // Fetches all unique role names available in the current dataset (from backend)
    if (allAvailableRoles.value.length > 0 && !force) return; // Avoid refetch if already populated unless forced
    try {
      const roles = await playerService.getAvailableRoles();
      allAvailableRoles.value = roles.sort(); // Sort for consistent display
    } catch (e) {
      console.error("playerStore: Error fetching available roles:", e);
      allAvailableRoles.value = []; // Reset or handle error appropriately
    }
  }

  function processPlayersFromAPI(playersData) {
    if (!Array.isArray(playersData)) return [];
    // Ensure age is an integer, default to 0 if not parsable
    return playersData.map((p) => ({
      ...p,
      age: parseInt(p.age, 10) || 0,
      // Other per-player processing can go here if needed
    }));
  }

  function resetState() {
    allPlayers.value = [];
    currentDatasetId.value = null;
    detectedCurrencySymbol.value = "$"; // Reset to default
    allAvailableRoles.value = [];
    sessionStorage.removeItem("currentDatasetId");
    sessionStorage.removeItem("detectedCurrencySymbol");
    // Do not clear error.value here, let components decide
  }

  async function loadFromSessionStorage() {
    const storedDatasetId = sessionStorage.getItem("currentDatasetId");
    const storedCurrencySymbol = sessionStorage.getItem(
      "detectedCurrencySymbol",
    );
    if (storedDatasetId) {
      currentDatasetId.value = storedDatasetId;
      if (storedCurrencySymbol) {
        detectedCurrencySymbol.value = storedCurrencySymbol;
      }
      try {
        // Fetch players and roles when loading from session
        await fetchPlayersByDatasetId(storedDatasetId);
        await fetchAllAvailableRoles();
      } catch (e) {
        // Error is handled within fetchPlayersByDatasetId and fetchAllAvailableRoles
        // If loading from session fails, the state will be reset by those functions.
      }
    } else {
      resetState(); // If no dataset ID in session, ensure clean state
    }
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
    allAvailableRoles,
    uploadPlayerFile,
    fetchPlayersByDatasetId,
    fetchAllAvailableRoles,
    resetState,
    loadFromSessionStorage,
    AGE_SLIDER_MIN_DEFAULT, // Expose defaults
    AGE_SLIDER_MAX_DEFAULT,
  };
});
