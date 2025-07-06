// Runtime API Endpoint Configuration
const getApiEndpoint = () => {
  // 1. Check for runtime config (injected by config-injector.sh)
  if (
    typeof window !== 'undefined' &&
    window.APP_CONFIG &&
    typeof window.APP_CONFIG.API_ENDPOINT !== 'undefined'
  ) {
    return window.APP_CONFIG.API_ENDPOINT
  }

  // 2. Fallback to build-time config (for local development)
  if (typeof import.meta.env.VITE_API_ENDPOINT !== 'undefined') {
    return import.meta.env.VITE_API_ENDPOINT
  }

  // 3. Default to an empty string for relative paths
  return ''
}

const API_ENDPOINT = getApiEndpoint()

export default {
  async saveWishlist(datasetId, wishlistData) {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/wishlists/${datasetId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(wishlistData)
      })

      if (!response.ok) {
        throw new Error(`API Error: ${response.status} - ${response.statusText}`)
      }
      return await response.json()
    } catch (error) {
      // Fallback to localStorage
      this.saveToLocalStorage(datasetId, wishlistData)
      throw error // Re-throw to let caller know MinIO failed
    }
  },

  async loadWishlist(datasetId) {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/wishlists/${datasetId}`)

      if (response.status === 404) {
        return this.loadFromLocalStorage(datasetId)
      }

      if (!response.ok) {
        throw new Error(`API Error: ${response.status} - ${response.statusText}`)
      }

      const data = await response.json()

      // Sync to localStorage as backup
      this.saveToLocalStorage(datasetId, data)

      return data
    } catch (_error) {
      // Fallback to localStorage
      return this.loadFromLocalStorage(datasetId)
    }
  },

  async deleteWishlist(datasetId) {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/wishlists/${datasetId}`, {
        method: 'DELETE'
      })

      if (!response.ok && response.status !== 404) {
        throw new Error(`API Error: ${response.status} - ${response.statusText}`)
      }
    } catch (_error) {
    } finally {
      // Always delete from localStorage regardless of MinIO result
      this.deleteFromLocalStorage(datasetId)
    }
  },

  // localStorage fallback methods
  saveToLocalStorage(datasetId, wishlistData) {
    try {
      const key = `fm_dash_wishlist_${datasetId}`
      localStorage.setItem(key, JSON.stringify(wishlistData))
    } catch (_error) {}
  },

  loadFromLocalStorage(datasetId) {
    try {
      const key = `fm_dash_wishlist_${datasetId}`
      const stored = localStorage.getItem(key)
      if (stored) {
        return JSON.parse(stored)
      }
      return []
    } catch (_error) {
      return []
    }
  },

  deleteFromLocalStorage(datasetId) {
    try {
      const key = `fm_dash_wishlist_${datasetId}`
      localStorage.removeItem(key)
    } catch (_error) {}
  },

  // Migration method to sync all localStorage wishlists to MinIO
  async migrateLocalStorageToMinIO() {
    try {
      const allKeys = Object.keys(localStorage)
      const wishlistKeys = allKeys.filter(key => key.startsWith('fm_dash_wishlist_'))

      for (const key of wishlistKeys) {
        const datasetId = key.replace('fm_dash_wishlist_', '')
        const wishlistData = this.loadFromLocalStorage(datasetId)

        if (wishlistData && wishlistData.length > 0) {
          try {
            await this.saveWishlist(datasetId, wishlistData)
          } catch (_error) {}
        }
      }
    } catch (_error) {}
  }
}
