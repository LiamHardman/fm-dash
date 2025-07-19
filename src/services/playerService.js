import { useApi } from '../composables/useApi.js'
import { useProtobufApi } from '../composables/useProtobufApi.js'
import { useErrorHandling } from '../composables/useErrorHandling.js'
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
const _resolvedApiEndpoint = API_ENDPOINT || getApiEndpoint()

export default {
  async uploadPlayerFile(formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) {
    const file = formData.get('playerFile')
    if (!file) {
      throw new Error('No file found in form data')
    }

    try {
      // Use protobuf-aware API for uploads (though uploads always use JSON)
      const { uploadPlayerFile } = useProtobufApi('')
      const response = await uploadPlayerFile(formData, maxSizeBytes, onProgress)
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

    const delay = ms => new Promise(resolve => setTimeout(resolve, ms))

    try {
      // Build query parameters
      const params = {}
      if (position) {
        params.position = position
      }
      if (role) {
        params.role = role
      }
      if (ageRange) {
        if (ageRange.min !== null && ageRange.min !== undefined) {
          params.minAge = ageRange.min.toString()
        }
        if (ageRange.max !== null && ageRange.max !== undefined) {
          params.maxAge = ageRange.max.toString()
        }
      }
      if (transferValueRange) {
        if (transferValueRange.min !== null && transferValueRange.min !== undefined) {
          params.minTransferValue = transferValueRange.min.toString()
        }
        if (transferValueRange.max !== null && transferValueRange.max !== undefined) {
          params.maxTransferValue = transferValueRange.max.toString()
        }
      }
      if (maxSalary !== null && maxSalary !== undefined) {
        params.maxSalary = maxSalary.toString()
      }
      if (divisionFilter && divisionFilter !== 'all') {
        params.divisionFilter = divisionFilter
      }
      if (targetDivision) {
        params.targetDivision = targetDivision
      }
      if (positionCompare && positionCompare !== 'all') {
        params.positionCompare = positionCompare
      }

      // Use protobuf-aware API for player data
      const { getPlayerData } = useProtobufApi('')
      const { withRetry } = useErrorHandling()
      
      // Use withRetry to handle potential race conditions with exponential backoff
      return await withRetry(
        () => getPlayerData(datasetId, params),
        {
          retries: maxRetries,
          initialDelay: 200,
          onRetry: (attempt) => {
            logger.warn(
              `Dataset not found (attempt ${attempt + 1}/${maxRetries + 1}), retrying...`
            )
          },
          shouldRetry: (error) => error.message?.includes('404')
        }
      )
    } catch (error) {
      logger.error('Error fetching players by dataset ID in playerService:', error)
      
      // Provide more specific error messages
      if (error.message?.includes('404')) {
        throw new Error(`Player data not found for ID: ${datasetId}.`)
      }
      
      throw error
    }
  },

  async getAvailableRoles() {
    try {
      // Use protobuf-aware API for roles
      const { getRoles } = useProtobufApi('')
      return await getRoles()
    } catch (error) {
      logger.error('Error fetching available roles in playerService:', error)
      throw error
    }
  },

  async getConfig() {
    try {
      // Use protobuf-aware API for config
      const { getConfig } = useProtobufApi('')
      return await getConfig()
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
      // Use protobuf-aware API for config updates
      const { post } = useProtobufApi('')
      return await post('/api/config', configUpdate, {}, 'api.GenericResponse')
    } catch (error) {
      logger.error('Error updating config in playerService:', error)
      throw error
    }
  },
  
  /**
   * Get client status information including protobuf capabilities
   */
  getClientStatus() {
    const { getClientStatus } = useProtobufApi('')
    return getClientStatus()
  },
  
  /**
   * Enable or disable protobuf support
   * @param {boolean} enabled - Whether protobuf should be enabled
   */
  setProtobufEnabled(enabled) {
    const { setProtobufEnabled } = useProtobufApi('')
    setProtobufEnabled(enabled)
  }
}

/**
 * Fetch detailed player stats for a single player
 * @param {string} datasetID - The dataset ID
 * @param {number} playerUID - The player UID (numeric, not UUID)
 * @returns {Promise<Object>} - Detailed player data
 */
export async function fetchFullPlayerStats(datasetID, playerUID) {
  try {
    // Use protobuf-aware API for detailed player stats
    const { get } = useProtobufApi('')
    const url = `/api/fullplayerstats/${datasetID}/${playerUID}`
    
    const response = await get(url, {}, 'api.GenericResponse')
    
    // Handle protobuf response structure where data is in the data field
    if (response.data) {
      try {
        const parsedData = JSON.parse(response.data)
        return { data: parsedData, format: 'json' }
      } catch (parseError) {
        logger.error('Error parsing detailed player data from protobuf response:', parseError)
        throw new Error('Invalid detailed player data format')
      }
    }
    
    // Fallback for JSON responses or direct data objects
    return { data: response, format: 'json' }
  } catch (error) {
    logger.error('Error fetching full player stats:', error)
    throw error
  }
}