const API_BASE_URL = '' // Use relative paths if proxy is set up for all API routes

export default {
  async saveWishlist(datasetId, wishlistData) {
    try {
      console.log(`Attempting to save wishlist for dataset ${datasetId} to MinIO`)

      const response = await fetch(`${API_BASE_URL}/api/wishlists/${datasetId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(wishlistData)
      })

      if (!response.ok) {
        throw new Error(`API Error: ${response.status} - ${response.statusText}`)
      }

      console.log(`Successfully saved wishlist for dataset ${datasetId} to MinIO`)
      return await response.json()
    } catch (error) {
      console.warn('Failed to save wishlist to MinIO, falling back to localStorage:', error)
      // Fallback to localStorage
      this.saveToLocalStorage(datasetId, wishlistData)
      throw error // Re-throw to let caller know MinIO failed
    }
  },

  async loadWishlist(datasetId) {
    try {
      console.log(`Attempting to load wishlist for dataset ${datasetId} from MinIO`)

      const response = await fetch(`${API_BASE_URL}/api/wishlists/${datasetId}`)

      if (response.status === 404) {
        // Wishlist doesn't exist in MinIO, try localStorage
        console.log(`No wishlist found in MinIO for dataset ${datasetId}, checking localStorage`)
        return this.loadFromLocalStorage(datasetId)
      }

      if (!response.ok) {
        throw new Error(`API Error: ${response.status} - ${response.statusText}`)
      }

      const data = await response.json()
      console.log(`Successfully loaded wishlist for dataset ${datasetId} from MinIO`)

      // Sync to localStorage as backup
      this.saveToLocalStorage(datasetId, data)

      return data
    } catch (error) {
      console.warn('Failed to load wishlist from MinIO, falling back to localStorage:', error)
      // Fallback to localStorage
      return this.loadFromLocalStorage(datasetId)
    }
  },

  async deleteWishlist(datasetId) {
    try {
      console.log(`Attempting to delete wishlist for dataset ${datasetId} from MinIO`)

      const response = await fetch(`${API_BASE_URL}/api/wishlists/${datasetId}`, {
        method: 'DELETE'
      })

      if (!response.ok && response.status !== 404) {
        throw new Error(`API Error: ${response.status} - ${response.statusText}`)
      }

      console.log(`Successfully deleted wishlist for dataset ${datasetId} from MinIO`)
    } catch (error) {
      console.warn('Failed to delete wishlist from MinIO:', error)
    } finally {
      // Always delete from localStorage regardless of MinIO result
      this.deleteFromLocalStorage(datasetId)
    }
  },

  // localStorage fallback methods
  saveToLocalStorage(datasetId, wishlistData) {
    try {
      const key = `fmdb_wishlist_${datasetId}`
      localStorage.setItem(key, JSON.stringify(wishlistData))
      console.log(`Saved wishlist for dataset ${datasetId} to localStorage`)
    } catch (error) {
      console.error('Failed to save to localStorage:', error)
    }
  },

  loadFromLocalStorage(datasetId) {
    try {
      const key = `fmdb_wishlist_${datasetId}`
      const stored = localStorage.getItem(key)
      if (stored) {
        console.log(`Loaded wishlist for dataset ${datasetId} from localStorage`)
        return JSON.parse(stored)
      }
      return []
    } catch (error) {
      console.error('Failed to load from localStorage:', error)
      return []
    }
  },

  deleteFromLocalStorage(datasetId) {
    try {
      const key = `fmdb_wishlist_${datasetId}`
      localStorage.removeItem(key)
      console.log(`Deleted wishlist for dataset ${datasetId} from localStorage`)
    } catch (error) {
      console.error('Failed to delete from localStorage:', error)
    }
  },

  // Migration method to sync all localStorage wishlists to MinIO
  async migrateLocalStorageToMinIO() {
    try {
      const allKeys = Object.keys(localStorage)
      const wishlistKeys = allKeys.filter(key => key.startsWith('fmdb_wishlist_'))

      for (const key of wishlistKeys) {
        const datasetId = key.replace('fmdb_wishlist_', '')
        const wishlistData = this.loadFromLocalStorage(datasetId)

        if (wishlistData && wishlistData.length > 0) {
          try {
            await this.saveWishlist(datasetId, wishlistData)
            console.log(`Migrated wishlist for dataset ${datasetId} to MinIO`)
          } catch (error) {
            console.warn(`Failed to migrate wishlist for dataset ${datasetId}:`, error)
          }
        }
      }
    } catch (error) {
      console.error('Error during wishlist migration:', error)
    }
  }
}
