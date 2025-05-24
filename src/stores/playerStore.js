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

  // Getters (uniqueClubs, uniqueNationalities, etc. remain the same)
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
    if (min >= max) max = min + 50000;
    if (min === 0 && max === 0 && transferValuesNumeric.some((v) => v === 0))
      max = 50000;
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
      await fetchPlayersByDatasetId(currentDatasetId.value);
      await fetchAllAvailableRoles();
      return response;
    } catch (e) {
      error.value = `Failed to process file: ${e.message || "Unknown error"}`;
      resetState();
      throw e;
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
      resetState();
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
      detectedCurrencySymbol.value = response.currencySymbol || "$";
      sessionStorage.setItem(
        "detectedCurrencySymbol",
        detectedCurrencySymbol.value,
      );
      return response;
    } catch (e) {
      error.value = `Failed to fetch player data: ${e.message || "Unknown error"}`;
      resetState();
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function fetchAllAvailableRoles(force = false) {
    if (allAvailableRoles.value.length > 0 && !force) return;
    try {
      const roles = await playerService.getAvailableRoles();
      allAvailableRoles.value = roles.sort();
    } catch (e) {
      console.error("playerStore: Error fetching available roles:", e);
      allAvailableRoles.value = [];
    }
  }

  function processPlayersFromAPI(playersData) {
    if (!Array.isArray(playersData)) return [];
    return playersData.map((p) => ({
      ...p,
      age: parseInt(p.age, 10) || 0,
    }));
  }

  function resetState() {
    allPlayers.value = [];
    currentDatasetId.value = null;
    detectedCurrencySymbol.value = "$";
    allAvailableRoles.value = [];
    sessionStorage.removeItem("currentDatasetId");
    sessionStorage.removeItem("detectedCurrencySymbol");
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
        await fetchPlayersByDatasetId(storedDatasetId); // Fetch with no filters initially
        await fetchAllAvailableRoles();
      } catch (e) {
        // Error handled in fetch
      }
    } else {
      resetState();
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
  };
});
