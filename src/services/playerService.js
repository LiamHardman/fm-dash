import { useApi } from '../composables/useApi'

const API_BASE_URL = ''

export default {
  async uploadPlayerFile(formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) {
    try {
      const file = formData.get('playerFile')
      if (!file) {
        throw new Error('No file found in form data')
      }
      
      const { uploadFile } = useApi('')
      const response = await uploadFile('/upload', file, onProgress)
      return response
    } catch (error) {
      console.error('Upload error in playerService:', error)
      
      if (error.message?.includes('413') || error.message?.includes('too large')) {
        const maxSizeMB = Math.round(maxSizeBytes / (1024 * 1024))
        const newError = new Error(`File too large. Maximum size allowed is ${maxSizeMB}MB.`)
        newError.status = 413
        throw newError
      }
      
    }
  },

  async getPlayersByDatasetId(
    datasetId,
    position = null,
    role = null,
    ageRange = null,
    transferValueRange = null,
    maxSalary = null,
    divisionFilter = 'all',
    targetDivision = null,
    positionCompare = 'all'
  ) {
    if (!datasetId) {
      return Promise.reject(new Error('Dataset ID is required.'))
    }
    try {
      let url = `${API_BASE_URL}/api/players/${datasetId}`
      const params = new URLSearchParams()
      if (position) {
        params.append('position', position)
      }
      if (role) {
        params.append('role', role)
      }
      if (ageRange) {
        if (ageRange.min !== null && ageRange.min !== undefined) {
          params.append('minAge', ageRange.min.toString())
        }
        if (ageRange.max !== null && ageRange.max !== undefined) {
          params.append('maxAge', ageRange.max.toString())
        }
      }
      if (transferValueRange) {
        if (transferValueRange.min !== null && transferValueRange.min !== undefined) {
          params.append('minTransferValue', transferValueRange.min.toString())
        }
        if (transferValueRange.max !== null && transferValueRange.max !== undefined) {
          params.append('maxTransferValue', transferValueRange.max.toString())
        }
      }
      if (maxSalary !== null && maxSalary !== undefined) {
        params.append('maxSalary', maxSalary.toString())
      }
      if (divisionFilter && divisionFilter !== 'all') {
        params.append('divisionFilter', divisionFilter)
      }
      if (targetDivision) {
        params.append('targetDivision', targetDivision)
      }
      if (positionCompare && positionCompare !== 'all') {
        params.append('positionCompare', positionCompare)
      }

      const queryString = params.toString()
      if (queryString) {
        url += `?${queryString}`
      }

      const response = await fetch(url)
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(`Player data not found for ID: ${datasetId}.`)
        }
        const errorText = await response.text()
        throw new Error(`API Error: ${response.status} - ${errorText || response.statusText}`)
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching players by dataset ID in playerService:', error)
      throw error
    }
  },

  async getAvailableRoles() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/roles`)
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(
          `API Error fetching roles: ${response.status} - ${errorText || response.statusText}`
        )
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching available roles in playerService:', error)
      throw error
    }
  },

  async getConfig() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/config`)
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(
          `API Error fetching config: ${response.status} - ${errorText || response.statusText}`
        )
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching config in playerService:', error)
      return {
        maxUploadSizeMB: 15,
        maxUploadSizeBytes: 15 * 1024 * 1024,
        useScaledRatings: true // Default to scaled ratings
      }
    }
  },

  async updateConfig(configUpdate) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/config`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(configUpdate)
      })
      
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(
          `API Error updating config: ${response.status} - ${errorText || response.statusText}`
        )
      }
      
      return await response.json()
    } catch (error) {
      console.error('Error updating config in playerService:', error)
      throw error
    }
  }
}
