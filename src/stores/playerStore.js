// src/stores/playerStore.js
import { defineStore } from "pinia";
import { ref, computed, shallowRef } from "vue"; // Import shallowRef
import playerService from "../services/playerService";

export const usePlayerStore = defineStore("player", () => {
  // State
  const allPlayers = shallowRef([]); // Use shallowRef for the large player list
  const currentDatasetId = ref(null);
  const detectedCurrencySymbol = ref("$");
  const loading = ref(false);
  const error = ref("");

  console.log(`playerStore: Initializing.`);

  // Getters
  // Computed properties will still work with shallowRef,
  // they will re-evaluate when allPlayers.value itself is replaced.
  const uniqueClubs = computed(() => {
    console.time("playerStore_uniqueClubs_computed");
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      console.timeEnd("playerStore_uniqueClubs_computed");
      return [];
    }
    const clubs = new Set();
    allPlayers.value.forEach((p) => {
      if (p.club) clubs.add(p.club);
    });
    const result = Array.from(clubs).sort();
    console.timeEnd("playerStore_uniqueClubs_computed");
    return result;
  });

  const uniqueNationalities = computed(() => {
    console.time("playerStore_uniqueNationalities_computed");
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      console.timeEnd("playerStore_uniqueNationalities_computed");
      return [];
    }
    const nationalities = new Set();
    allPlayers.value.forEach((p) => {
      if (p.nationality) nationalities.add(p.nationality);
    });
    const result = Array.from(nationalities).sort();
    console.timeEnd("playerStore_uniqueNationalities_computed");
    return result;
  });

  const uniqueMediaHandlings = computed(() => {
    console.time("playerStore_uniqueMediaHandlings_computed");
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      console.timeEnd("playerStore_uniqueMediaHandlings_computed");
      return [];
    }
    const mediaHandlingsIndividual = new Set();
    allPlayers.value.forEach((p) => {
      if (p.media_handling) {
        p.media_handling.split(",").forEach((style) => {
          const trimmedStyle = style.trim();
          if (trimmedStyle) mediaHandlingsIndividual.add(trimmedStyle);
        });
      }
    });
    const result = Array.from(mediaHandlingsIndividual).sort();
    console.timeEnd("playerStore_uniqueMediaHandlings_computed");
    return result;
  });

  const uniquePersonalities = computed(() => {
    console.time("playerStore_uniquePersonalities_computed");
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      console.timeEnd("playerStore_uniquePersonalities_computed");
      return [];
    }
    const personalities = new Set();
    allPlayers.value.forEach((p) => {
      if (p.personality) personalities.add(p.personality);
    });
    const result = Array.from(personalities).sort();
    console.timeEnd("playerStore_uniquePersonalities_computed");
    return result;
  });

  const uniquePositionsCount = computed(() => {
    console.time("playerStore_uniquePositionsCount_computed");
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      console.timeEnd("playerStore_uniquePositionsCount_computed");
      return 0;
    }
    const s = new Set();
    allPlayers.value.forEach((player) =>
      player.parsedPositions?.forEach((pos) => s.add(pos)),
    );
    const result = s.size;
    console.timeEnd("playerStore_uniquePositionsCount_computed");
    return result;
  });

  const transferValueRange = computed(() => {
    console.time("playerStore_transferValueRange_computed");
    if (!Array.isArray(allPlayers.value) || allPlayers.value.length === 0) {
      console.timeEnd("playerStore_transferValueRange_computed");
      return { min: 0, max: 100000000 };
    }
    const transferValuesNumeric = allPlayers.value
      .filter((p) => typeof p.transferValueAmount === "number")
      .map((p) => p.transferValueAmount);

    if (transferValuesNumeric.length === 0) {
      console.timeEnd("playerStore_transferValueRange_computed");
      return { min: 0, max: 100000000 };
    }

    const min = Math.min(0, ...transferValuesNumeric);
    let max = Math.max(...transferValuesNumeric);

    if (min >= max) {
      max = min + 50000;
    }
    if (min === 0 && max === 0 && transferValuesNumeric.some((v) => v === 0)) {
      max = 50000;
    }
    console.timeEnd("playerStore_transferValueRange_computed");
    return { min, max };
  });

  // Actions
  async function uploadPlayerFile(formData) {
    console.time("playerStore_uploadPlayerFile_action_total");
    loading.value = true;
    error.value = "";
    try {
      console.log(`playerStore: Calling playerService.uploadPlayerFile`);
      console.time("playerStore_playerService_uploadPlayerFile_call");
      const response = await playerService.uploadPlayerFile(formData);
      console.timeEnd("playerStore_playerService_uploadPlayerFile_call");
      console.log(
        `playerStore: playerService.uploadPlayerFile response:`,
        JSON.parse(JSON.stringify(response)),
      );

      currentDatasetId.value = response.datasetId;
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || "$";

      sessionStorage.setItem("currentDatasetId", currentDatasetId.value);
      sessionStorage.setItem(
        "detectedCurrencySymbol",
        detectedCurrencySymbol.value,
      );

      console.log(
        `playerStore: Fetching players for new datasetId: ${currentDatasetId.value}`,
      );
      await fetchPlayersByDatasetId(currentDatasetId.value);
      console.timeEnd("playerStore_uploadPlayerFile_action_total");
      return response;
    } catch (e) {
      console.error(`playerStore: Error in uploadPlayerFile:`, e);
      error.value = `Failed to process file: ${e.message || "Unknown error"}`;
      resetState();
      console.timeEnd("playerStore_uploadPlayerFile_action_total");
      throw e;
    } finally {
      loading.value = false;
      console.log(`playerStore: uploadPlayerFile loading set to false.`);
    }
  }

  async function fetchPlayersByDatasetId(datasetId) {
    if (!datasetId) {
      console.warn(
        `playerStore: fetchPlayersByDatasetId called with no datasetId.`,
      );
      return;
    }
    console.time(
      `playerStore_fetchPlayersByDatasetId_action_dataset_${datasetId}`,
    );
    loading.value = true;
    error.value = "";
    try {
      console.log(
        `playerStore: Calling playerService.getPlayersByDatasetId for ${datasetId}`,
      );
      console.time(
        `playerStore_playerService_getPlayersByDatasetId_call_dataset_${datasetId}`,
      );
      const response = await playerService.getPlayersByDatasetId(datasetId);
      console.timeEnd(
        `playerStore_playerService_getPlayersByDatasetId_call_dataset_${datasetId}`,
      );
      console.log(
        `playerStore: playerService.getPlayersByDatasetId response for ${datasetId}:`,
        {
          playersLength: response.players?.length,
          currencySymbol: response.currencySymbol,
        },
      );

      console.time(`playerStore_processPlayersFromAPI_dataset_${datasetId}`);
      // When using shallowRef, assigning a new array to .value is what triggers reactivity.
      allPlayers.value = processPlayersFromAPI(response.players);
      console.timeEnd(`playerStore_processPlayersFromAPI_dataset_${datasetId}`);
      console.log(
        `playerStore: allPlayers.value updated (shallowRef). Length: ${allPlayers.value.length}`,
      );

      detectedCurrencySymbol.value = response.currencySymbol || "$";
      sessionStorage.setItem(
        "detectedCurrencySymbol",
        detectedCurrencySymbol.value,
      );
      console.timeEnd(
        `playerStore_fetchPlayersByDatasetId_action_dataset_${datasetId}`,
      );
      return response;
    } catch (e) {
      console.error(
        `playerStore: Error in fetchPlayersByDatasetId for ${datasetId}:`,
        e,
      );
      error.value = `Failed to fetch player data for dataset ${datasetId}: ${e.message || "Unknown error"}`;
      resetState();
      console.timeEnd(
        `playerStore_fetchPlayersByDatasetId_action_dataset_${datasetId}`,
      );
      throw e;
    } finally {
      loading.value = false;
      console.log(
        `playerStore: fetchPlayersByDatasetId loading set to false for ${datasetId}.`,
      );
    }
  }

  function processPlayersFromAPI(playersData) {
    return playersData.map((p) => ({
      ...p,
      age: parseInt(p.age, 10) || 0,
    }));
  }

  function resetState() {
    console.log(`playerStore: resetState called.`);
    allPlayers.value = [];
    currentDatasetId.value = null;
    detectedCurrencySymbol.value = "$";
    sessionStorage.removeItem("currentDatasetId");
    sessionStorage.removeItem("detectedCurrencySymbol");
  }

  function loadFromSessionStorage() {
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
      console.timeEnd("playerStore_loadFromSessionStorage_action");
      return fetchPlayersByDatasetId(storedDatasetId);
    }
    console.timeEnd("playerStore_loadFromSessionStorage_action");
    return Promise.resolve();
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
    uploadPlayerFile,
    fetchPlayersByDatasetId,
    resetState,
    loadFromSessionStorage,
  };
});
