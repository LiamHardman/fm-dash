const getApiEndpoint = () => {
  if (
    typeof window !== 'undefined' &&
    window.APP_CONFIG &&
    typeof window.APP_CONFIG.API_ENDPOINT !== 'undefined'
  ) {
    return window.APP_CONFIG.API_ENDPOINT
  }
  if (typeof import.meta.env.VITE_API_ENDPOINT !== 'undefined') {
    return import.meta.env.VITE_API_ENDPOINT
  }
  return ''
}

const API_ENDPOINT = getApiEndpoint()

export default {
  /**
   * Analyze a set of uploaded snapshots.
   * @param {string[]} datasetIds
   * @param {object} [options]
   * @param {string[]} [options.order] - explicit chronological order override (dataset IDs)
   * @param {string} [options.sortField] - "known interesting field" to sort delta by
   * @param {string} [options.sortDir] - 'asc' | 'desc'
   */
  async analyze(datasetIds, { order, sortField, sortDir } = {}) {
    const params = new URLSearchParams()
    if (sortField) params.set('sortField', sortField)
    if (sortDir) params.set('sortDir', sortDir)
    const query = params.toString() ? `?${params.toString()}` : ''

    const response = await fetch(`${API_ENDPOINT}/api/progression/analyze${query}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ datasetIds, ...(order ? { order } : {}) }),
    })

    const data = await response.json()

    if (!response.ok) {
      const message = data?.message || `API Error: ${response.status}`
      const error = new Error(message)
      error.status = response.status
      error.body = data
      throw error
    }

    return data
  },
}
