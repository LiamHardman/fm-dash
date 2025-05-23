// src/stores/playerStore.js
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import playerService from '../services/playerService';

export const usePlayerStore = defineStore('player', () => {
  // State
  const allPlayers = ref([]);
  const currentDatasetId = ref(null);
  const detectedCurrencySymbol = ref('$');
  const loading = ref(false);
  const error = ref('');

  // Getters
  const uniqueClubs = computed(() => {
    const clubs = new Set();
    allPlayers.value.forEach(p => {
      if (p.club) clubs.add(p.club);
    });
    return Array.from(clubs).sort();
  });

  const uniqueNationalities = computed(() => {
    const nationalities = new Set();
    allPlayers.value.forEach(p => {
      if (p.nationality) nationalities.add(p.nationality);
    });
    return Array.from(nationalities).sort();
  });

  const uniqueMediaHandlings = computed(() => {
    const mediaHandlingsIndividual = new Set();
    allPlayers.value.forEach(p => {
      if (p.media_handling) {
        p.media_handling.split(',').forEach(style => {
          const trimmedStyle = style.trim();
          if (trimmedStyle) mediaHandlingsIndividual.add(trimmedStyle);
        });
      }
    });
    return Array.from(mediaHandlingsIndividual).sort();
  });

  const uniquePersonalities = computed(() => {
    const personalities = new Set();
    allPlayers.value.forEach(p => {
      if (p.personality) personalities.add(p.personality);
    });
    return Array.from(personalities).sort();
  });

  const uniquePositionsCount = computed(() => {
    const s = new Set();
    allPlayers.value.forEach(player => 
      player.parsedPositions?.forEach(pos => s.add(pos))
    );
    return s.size;
  });

  const transferValueRange = computed(() => {
    const transferValuesNumeric = allPlayers.value
      .filter(p => typeof p.transferValueAmount === 'number')
      .map(p => p.transferValueAmount);
    
    if (transferValuesNumeric.length === 0) {
      return { min: 0, max: 100000000 };
    }
    
    const min = Math.min(0, ...transferValuesNumeric);
    let max = Math.max(...transferValuesNumeric);
    
    // Handle edge cases
    if (min >= max) {
      max = min + 50000;
    }
    if (min === 0 && max === 0 && transferValuesNumeric.some(v => v === 0)) {
      max = 50000;
    }
    
    return { min, max };
  });

  // Actions
  async function uploadPlayerFile(formData) {
    loading.value = true;
    error.value = '';
    try {
      const response = await playerService.uploadPlayerFile(formData);
      currentDatasetId.value = response.datasetId;
      detectedCurrencySymbol.value = response.detectedCurrencySymbol || '$';

      sessionStorage.setItem('currentDatasetId', currentDatasetId.value);
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value);

      await fetchPlayersByDatasetId(currentDatasetId.value);
      return response;
    } catch (e) {
      error.value = `Failed to process file: ${e.message || 'Unknown error'}`;
      resetState();
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function fetchPlayersByDatasetId(datasetId) {
    if (!datasetId) return;
    loading.value = true;
    error.value = '';
    try {
      const response = await playerService.getPlayersByDatasetId(datasetId);
      allPlayers.value = processPlayersFromAPI(response.players);
      detectedCurrencySymbol.value = response.currencySymbol || '$';
      sessionStorage.setItem('detectedCurrencySymbol', detectedCurrencySymbol.value);
      return response;
    } catch (e) {
      error.value = `Failed to fetch player data for dataset ${datasetId}: ${e.message || 'Unknown error'}`;
      resetState();
      throw e;
    } finally {
      loading.value = false;
    }
  }

  function processPlayersFromAPI(playersData) {
    return playersData.map(p => ({
      ...p,
      age: parseInt(p.age, 10) || 0,
    }));
  }

  function resetState() {
    allPlayers.value = [];
    currentDatasetId.value = null;
    detectedCurrencySymbol.value = '$';
    sessionStorage.removeItem('currentDatasetId');
    sessionStorage.removeItem('detectedCurrencySymbol');
  }

  function loadFromSessionStorage() {
    const storedDatasetId = sessionStorage.getItem('currentDatasetId');
    const storedCurrencySymbol = sessionStorage.getItem('detectedCurrencySymbol');
    if (storedDatasetId) {
      currentDatasetId.value = storedDatasetId;
      if (storedCurrencySymbol) {
        detectedCurrencySymbol.value = storedCurrencySymbol;
      }
      return fetchPlayersByDatasetId(storedDatasetId);
    }
    return Promise.resolve();
  }

  return {
    // State
    allPlayers,
    currentDatasetId,
    detectedCurrencySymbol,
    loading,
    error,
    
    // Getters
    uniqueClubs,
    uniqueNationalities,
    uniqueMediaHandlings,
    uniquePersonalities,
    uniquePositionsCount,
    transferValueRange,
    
    // Actions
    uploadPlayerFile,
    fetchPlayersByDatasetId,
    resetState,
    loadFromSessionStorage
  };
});