const API_URL = 'http://localhost:8091'

export default {
  /**
   * Upload a player file to the API for parsing
   * @param {FormData} formData - The form data containing the file
   * @returns {Promise<Array>} - A promise that resolves to the parsed player data
   */
  async uploadPlayerFile(formData) {
    try {
      const response = await fetch(`${API_URL}/upload`, {
        method: 'POST',
        body: formData
      })
      
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`API Error: ${response.status} - ${errorText || response.statusText}`)
      }
      
      return await response.json()
    } catch (error) {
      console.error('Upload error:', error)
      throw error
    }
  }
}