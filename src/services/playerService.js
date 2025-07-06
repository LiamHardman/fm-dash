import { useApi } from '../composables/useApi.js'
import logger from '../utils/logger.js'

const API_ENDPOINT = import.meta.env.VITE_API_ENDPOINT || ''

const getApiEndpoint = () => {
  if (typeof window !== 'undefined' && window.location) {
    // For production, use current protocol and host
    return `${window.location.protocol}//${window.location.host}`
  }
  // Fallback for SSR or testing
  return API_ENDPOINT || 'http://localhost:8080'
}

// If no API_ENDPOINT is set, use the current location
const resolvedApiEndpoint = API_ENDPOINT || getApiEndpoint()

export default {
  async uploadPlayerFile(formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) {
    const file = formData.get('playerFile')
    if (!file) {
      throw new Error('No file found in form data')
    }

    try {
      const { uploadFile } = useApi('')
      
      const response = await uploadFile('/upload', file, onProgress)
      return response
    } catch (error) {
      logger.error('Upload error in playerService:', error)
      
      if (error.message?.includes('413') || error.message?.includes('too large')) {
        const maxSizeMB = Math.round(maxSizeBytes / (1024 * 1024))
        const newError = new Error(`File too large. Maximum size allowed is ${maxSizeMB}MB.`)
        newError.status = 413
        throw newError
      }
      
      throw error
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
    positionCompare = 'all',
    retryCount = 0,
    maxRetries = 3
  ) {
    if (!datasetId) {
      return Promise.reject(new Error('Dataset ID is required.'))
    }
    
    const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms))
    
    try {
      let url = `${API_ENDPOINT}/api/players/${datasetId}`
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
          // Handle potential race condition in multi-replica deployments
          if (retryCount < maxRetries) {
            logger.warn(`Dataset not found (attempt ${retryCount + 1}/${maxRetries + 1}), retrying in ${200 * Math.pow(2, retryCount)}ms...`)
            await delay(200 * Math.pow(2, retryCount)) // Exponential backoff: 200ms, 400ms, 800ms
            return this.getPlayersByDatasetId(
              datasetId, position, role, ageRange, transferValueRange, 
              maxSalary, divisionFilter, targetDivision, positionCompare, 
              retryCount + 1, maxRetries
            )
          }
          throw new Error(`Player data not found for ID: ${datasetId}.`)
        }
        const errorText = await response.text()
        throw new Error(`API Error: ${response.status} - ${errorText || response.statusText}`)
      }
      return await response.json()
    } catch (error) {
      logger.error('Error fetching players by dataset ID in playerService:', error)
      throw error
    }
  },

  async getAvailableRoles() {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/roles`)
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(
          `API Error fetching roles: ${response.status} - ${errorText || response.statusText}`
        )
      }
      return await response.json()
    } catch (error) {
      logger.error('Error fetching available roles in playerService:', error)
      throw error
    }
  },

  async getConfig() {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/config`)
      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(
          `API Error fetching config: ${response.status} - ${errorText || response.statusText}`
        )
      }
      return await response.json()
    } catch (error) {
      logger.error('Error fetching config in playerService:', error)
      return {
        maxUploadSizeMB: 15,
        maxUploadSizeBytes: 15 * 1024 * 1024,
        useScaledRatings: true, // Default to scaled ratings
        datasetRetentionDays: 30 // Default retention period
      }
    }
  },

  async updateConfig(configUpdate) {
    try {
      const response = await fetch(`${API_ENDPOINT}/api/config`, {
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
      logger.error('Error updating config in playerService:', error)
      throw error
    }
  }
}
