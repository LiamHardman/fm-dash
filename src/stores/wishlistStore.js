import { defineStore } from "pinia";
import { ref, computed } from "vue";

export const useWishlistStore = defineStore("wishlist", () => {
  // Wishlist data structure: { [datasetId]: [player objects] }
  const wishlistsByDataset = ref({});

  // Get wishlist for current dataset
  const getWishlistForDataset = (datasetId) => {
    if (!datasetId) return [];
    return wishlistsByDataset.value[datasetId] || [];
  };

  // Add player to wishlist for specific dataset
  const addToWishlist = (datasetId, player) => {
    if (!datasetId || !player) return false;
    
    if (!wishlistsByDataset.value[datasetId]) {
      wishlistsByDataset.value[datasetId] = [];
    }
    
    // Check if player is already in wishlist
    const existingIndex = wishlistsByDataset.value[datasetId].findIndex(
      p => p.name === player.name && p.club === player.club
    );
    
    if (existingIndex === -1) {
      wishlistsByDataset.value[datasetId].push(player);
      saveToLocalStorage();
      return true;
    }
    
    return false; // Player already in wishlist
  };

  // Remove player from wishlist for specific dataset
  const removeFromWishlist = (datasetId, player) => {
    if (!datasetId || !player || !wishlistsByDataset.value[datasetId]) return false;
    
    const index = wishlistsByDataset.value[datasetId].findIndex(
      p => p.name === player.name && p.club === player.club
    );
    
    if (index !== -1) {
      wishlistsByDataset.value[datasetId].splice(index, 1);
      saveToLocalStorage();
      return true;
    }
    
    return false;
  };

  // Check if player is in wishlist for specific dataset
  const isInWishlist = (datasetId, player) => {
    if (!datasetId || !player || !wishlistsByDataset.value[datasetId]) return false;
    
    return wishlistsByDataset.value[datasetId].some(
      p => p.name === player.name && p.club === player.club
    );
  };

  // Clear wishlist for specific dataset
  const clearWishlistForDataset = (datasetId) => {
    if (datasetId && wishlistsByDataset.value[datasetId]) {
      delete wishlistsByDataset.value[datasetId];
      saveToLocalStorage();
    }
  };

  // Get wishlist count for dataset
  const getWishlistCount = (datasetId) => {
    if (!datasetId) return 0;
    return wishlistsByDataset.value[datasetId]?.length || 0;
  };

  // Save to localStorage
  const saveToLocalStorage = () => {
    try {
      localStorage.setItem("fmdb_wishlists", JSON.stringify(wishlistsByDataset.value));
    } catch (error) {
      console.error("Failed to save wishlists to localStorage:", error);
    }
  };

  // Load from localStorage
  const loadFromLocalStorage = () => {
    try {
      const stored = localStorage.getItem("fmdb_wishlists");
      if (stored) {
        wishlistsByDataset.value = JSON.parse(stored);
      }
    } catch (error) {
      console.error("Failed to load wishlists from localStorage:", error);
      wishlistsByDataset.value = {};
    }
  };

  // Initialize from localStorage
  loadFromLocalStorage();

  return {
    wishlistsByDataset,
    getWishlistForDataset,
    addToWishlist,
    removeFromWishlist,
    isInWishlist,
    clearWishlistForDataset,
    getWishlistCount,
    loadFromLocalStorage,
  };
}); 