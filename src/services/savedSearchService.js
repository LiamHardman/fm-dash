const API_ENDPOINT =
  typeof window !== 'undefined' && window.APP_CONFIG?.API_ENDPOINT !== undefined
    ? window.APP_CONFIG.API_ENDPOINT
    : import.meta.env.VITE_API_ENDPOINT || ''
const LOCAL_STORAGE_KEY = 'fm_dash_saved_searches_v1'

const defaultDocument = () => ({ version: 1, searches: [] })

export default {
  defaultDocument,
  async load() {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/saved-searches`)
      if (!response.ok) throw new Error(`API Error: ${response.status}`)
      return await response.json()
    } catch (_error) {
      try {
        return JSON.parse(localStorage.getItem(LOCAL_STORAGE_KEY) || '')
      } catch (_localError) {
        return defaultDocument()
      }
    }
  },
  async save(document) {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/saved-searches`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(document),
      })
      if (!response.ok) throw new Error(`API Error: ${response.status}`)
      const saved = await response.json()
      localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(saved))
      return saved
    } catch (error) {
      try {
        localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(document))
      } catch (_localError) {}
      throw error
    }
  },
}
