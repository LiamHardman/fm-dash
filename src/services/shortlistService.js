const getApiEndpoint = () => {
  if (typeof window !== 'undefined' && window.APP_CONFIG?.API_ENDPOINT !== undefined) {
    return window.APP_CONFIG.API_ENDPOINT
  }
  return import.meta.env.VITE_API_ENDPOINT || ''
}

const API_ENDPOINT = getApiEndpoint()
const LOCAL_STORAGE_KEY = 'fm_dash_shortlists_v1'

const defaultDocument = () => ({
  version: 1,
  lists: [{ id: 'default', name: 'My shortlist', items: [] }],
})

export default {
  defaultDocument,

  async load() {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/shortlists`)
      if (!response.ok) throw new Error(`API Error: ${response.status}`)
      return await response.json()
    } catch (_error) {
      return this.loadFromLocalStorage()
    }
  },

  async save(document) {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/shortlists`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(document),
      })
      if (!response.ok) throw new Error(`API Error: ${response.status}`)
      const saved = await response.json()
      this.saveToLocalStorage(saved)
      return saved
    } catch (error) {
      this.saveToLocalStorage(document)
      throw error
    }
  },

  saveToLocalStorage(document) {
    try {
      localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(document))
    } catch (_error) {}
  },

  loadFromLocalStorage() {
    try {
      const stored = localStorage.getItem(LOCAL_STORAGE_KEY)
      return stored ? JSON.parse(stored) : defaultDocument()
    } catch (_error) {
      return defaultDocument()
    }
  },
}
