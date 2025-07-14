import { defineStore } from 'pinia'
import { ref } from 'vue'
import { analytics } from '../services/analytics'
import wishlistService from '../services/wishlistService'

export const useWishlistStore = defineStore('wishlist', () => {
  // Wishlist data structure: { [datasetId]: [player objects] }
  const wishlistsByDataset = ref({})
  const loading = ref(false)

  // Get wishlist for current dataset
  const getWishlistForDataset = datasetId => {
    if (!datasetId) return []
    return wishlistsByDataset.value[datasetId] || []
  }

  // Load wishlist for specific dataset (from MinIO with localStorage fallback)
  const loadWishlistForDataset = async datasetId => {
    if (!datasetId) return []

    loading.value = true
    try {
      const wishlistData = await wishlistService.loadWishlist(datasetId)
      wishlistsByDataset.value[datasetId] = wishlistData || []
      return wishlistsByDataset.value[datasetId]
    } catch (_error) {
      // Fallback to empty array
      wishlistsByDataset.value[datasetId] = []
      return []
    } finally {
      loading.value = false
    }
  }

  // Save wishlist for specific dataset (to MinIO with localStorage fallback)
  const saveWishlistForDataset = async datasetId => {
    if (!datasetId || !wishlistsByDataset.value[datasetId]) return

    try {
      await wishlistService.saveWishlist(datasetId, wishlistsByDataset.value[datasetId])
    } catch (_error) {}
  }

  // Add player to wishlist for specific dataset
  const addToWishlist = async (datasetId, player) => {
    if (!datasetId || !player) return false

    // Ensure dataset wishlist exists
    if (!wishlistsByDataset.value[datasetId]) {
      wishlistsByDataset.value[datasetId] = []
    }

    // Check if player is already in wishlist
    const existingIndex = wishlistsByDataset.value[datasetId].findIndex(
      p => p.name === player.name && p.club === player.club
    )

    if (existingIndex === -1) {
      wishlistsByDataset.value[datasetId].push(player)
      await saveWishlistForDataset(datasetId)

      // Track the wishlist addition
      analytics.useWishlist('add', player.id || player.name)

      return true
    }

    return false // Player already in wishlist
  }

  // Remove player from wishlist for specific dataset
  const removeFromWishlist = async (datasetId, player) => {
    if (!datasetId || !player || !wishlistsByDataset.value[datasetId]) return false

    const index = wishlistsByDataset.value[datasetId].findIndex(
      p => p.name === player.name && p.club === player.club
    )

    if (index !== -1) {
      wishlistsByDataset.value[datasetId].splice(index, 1)
      await saveWishlistForDataset(datasetId)

      // Track the wishlist removal
      analytics.useWishlist('remove', player.id || player.name)

      return true
    }

    return false
  }

  // Check if player is in wishlist for specific dataset
  const isInWishlist = (datasetId, player) => {
    if (!datasetId || !player || !wishlistsByDataset.value[datasetId]) return false

    return wishlistsByDataset.value[datasetId].some(
      p => p.name === player.name && p.club === player.club
    )
  }

  // Clear wishlist for specific dataset
  const clearWishlistForDataset = async datasetId => {
    if (datasetId && wishlistsByDataset.value[datasetId]) {
      wishlistsByDataset.value[datasetId] = []
      await saveWishlistForDataset(datasetId)

      // Also delete from both MinIO and localStorage
      try {
        await wishlistService.deleteWishlist(datasetId)
      } catch (_error) {}
    }
  }

  // Get wishlist count for dataset
  const getWishlistCount = datasetId => {
    if (!datasetId) return 0
    return wishlistsByDataset.value[datasetId]?.length || 0
  }

  // Initialize wishlists for a dataset (call this when dataset is loaded)
  const initializeWishlistForDataset = async datasetId => {
    if (!datasetId) return

    // Only load if not already loaded
    if (!wishlistsByDataset.value[datasetId]) {
      await loadWishlistForDataset(datasetId)
    }
  }

  // Migrate from old localStorage format to new service-based approach
  const migrateFromOldLocalStorage = async () => {
    try {
      const oldData = localStorage.getItem('fm_dash_wishlists')
      if (oldData) {
        const parsedData = JSON.parse(oldData)

        // Migrate each dataset's wishlist
        for (const [datasetId, wishlistData] of Object.entries(parsedData)) {
          if (Array.isArray(wishlistData) && wishlistData.length > 0) {
            wishlistsByDataset.value[datasetId] = wishlistData
            await saveWishlistForDataset(datasetId)
          }
        }

        // Remove old localStorage entry after successful migration
        localStorage.removeItem('fm_dash_wishlists')
      }
    } catch (_error) {}
  }

  // Auto-migrate on store initialization
  migrateFromOldLocalStorage()

  return {
    wishlistsByDataset,
    loading,
    getWishlistForDataset,
    loadWishlistForDataset,
    addToWishlist,
    removeFromWishlist,
    isInWishlist,
    clearWishlistForDataset,
    getWishlistCount,
    initializeWishlistForDataset
  }
})
