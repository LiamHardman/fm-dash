import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import playerService from './playerService'

// Mock the composables
vi.mock('../composables/useProtobufApi.js', () => ({
  useProtobufApi: vi.fn(() => ({
    getPlayerData: vi.fn(),
    getRoles: vi.fn(),
    getConfig: vi.fn(),
    post: vi.fn(),
    uploadPlayerFile: vi.fn(),
    withRetry: vi.fn((fn) => fn()),
    getClientStatus: vi.fn(),
    setProtobufEnabled: vi.fn()
  }))
}))

vi.mock('../composables/useErrorHandling.js', () => ({
  useErrorHandling: vi.fn(() => ({
    handleFetchError: vi.fn(),
    withRetry: vi.fn((fn) => fn()),
    safeAsync: vi.fn((fn) => fn())
  }))
}))

vi.mock('../utils/logger.js', () => ({
  default: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn()
  }
}))

describe('playerService', () => {
  let mockProtobufApi
  
  beforeEach(() => {
    // Reset mocks
    vi.resetAllMocks()
    
    // Setup mock implementation
    mockProtobufApi = {
      getPlayerData: vi.fn().mockResolvedValue({ players: [] }),
      getRoles: vi.fn().mockResolvedValue({ roles: [] }),
      getConfig: vi.fn().mockResolvedValue({ maxUploadSizeMB: 15 }),
      post: vi.fn().mockResolvedValue({ success: true }),
      uploadPlayerFile: vi.fn().mockResolvedValue({ datasetId: '123' }),
      withRetry: vi.fn(fn => fn()),
      getClientStatus: vi.fn().mockReturnValue({ protobufSupported: true }),
      setProtobufEnabled: vi.fn()
    }
    
    // Update the mock implementation
    vi.mocked(require('../composables/useProtobufApi.js').useProtobufApi).mockReturnValue(mockProtobufApi)
  })
  
  afterEach(() => {
    vi.clearAllMocks()
  })
  
  describe('uploadPlayerFile', () => {
    it('should call uploadPlayerFile from useProtobufApi', async () => {
      const formData = new FormData()
      const file = new File(['test'], 'test.html')
      formData.append('playerFile', file)
      
      await playerService.uploadPlayerFile(formData)
      
      expect(mockProtobufApi.uploadPlayerFile).toHaveBeenCalledWith(formData, 15 * 1024 * 1024, null)
    })
    
    it('should handle file size errors', async () => {
      const formData = new FormData()
      const file = new File(['test'], 'test.html')
      formData.append('playerFile', file)
      
      const error = new Error('413 Request Entity Too Large')
      mockProtobufApi.uploadPlayerFile.mockRejectedValue(error)
      
      await expect(playerService.uploadPlayerFile(formData)).rejects.toThrow('File too large')
    })
    
    it('should throw error when no file is provided', async () => {
      const formData = new FormData()
      
      await expect(playerService.uploadPlayerFile(formData)).rejects.toThrow('No file found')
    })
  })
  
  describe('getPlayersByDatasetId', () => {
    it('should call getPlayerData with correct parameters', async () => {
      await playerService.getPlayersByDatasetId('123', 'ST', 'Advanced Forward')
      
      expect(mockProtobufApi.getPlayerData).toHaveBeenCalledWith('123', {
        position: 'ST',
        role: 'Advanced Forward'
      })
    })
    
    it('should handle all filter parameters', async () => {
      await playerService.getPlayersByDatasetId(
        '123',
        'ST',
        'Advanced Forward',
        { min: 18, max: 25 },
        { min: 1000000, max: 5000000 },
        10000,
        'top',
        'Premier Division',
        'similar'
      )
      
      expect(mockProtobufApi.getPlayerData).toHaveBeenCalledWith('123', {
        position: 'ST',
        role: 'Advanced Forward',
        minAge: '18',
        maxAge: '25',
        minTransferValue: '1000000',
        maxTransferValue: '5000000',
        maxSalary: '10000',
        divisionFilter: 'top',
        targetDivision: 'Premier Division',
        positionCompare: 'similar'
      })
    })
    
    it('should throw error when no datasetId is provided', async () => {
      await expect(playerService.getPlayersByDatasetId()).rejects.toThrow('Dataset ID is required')
    })
    
    it('should handle 404 errors with specific message', async () => {
      mockProtobufApi.getPlayerData.mockRejectedValue(new Error('404 Not Found'))
      
      await expect(playerService.getPlayersByDatasetId('123')).rejects.toThrow('Player data not found for ID: 123')
    })
  })
  
  describe('getAvailableRoles', () => {
    it('should call getRoles from useProtobufApi', async () => {
      await playerService.getAvailableRoles()
      
      expect(mockProtobufApi.getRoles).toHaveBeenCalled()
    })
  })
  
  describe('getConfig', () => {
    it('should call getConfig from useProtobufApi', async () => {
      await playerService.getConfig()
      
      expect(mockProtobufApi.getConfig).toHaveBeenCalled()
    })
    
    it('should return default config on error', async () => {
      mockProtobufApi.getConfig.mockRejectedValue(new Error('Failed to fetch config'))
      
      const result = await playerService.getConfig()
      
      expect(result).toEqual({
        maxUploadSizeMB: 15,
        maxUploadSizeBytes: 15 * 1024 * 1024,
        useScaledRatings: true,
        datasetRetentionDays: 30
      })
    })
  })
  
  describe('updateConfig', () => {
    it('should call post with correct parameters', async () => {
      const configUpdate = { useScaledRatings: false }
      
      await playerService.updateConfig(configUpdate)
      
      expect(mockProtobufApi.post).toHaveBeenCalledWith('/api/config', configUpdate, {}, 'api.GenericResponse')
    })
  })
  
  describe('protobuf utilities', () => {
    it('should get client status', () => {
      playerService.getClientStatus()
      
      expect(mockProtobufApi.getClientStatus).toHaveBeenCalled()
    })
    
    it('should set protobuf enabled state', () => {
      playerService.setProtobufEnabled(false)
      
      expect(mockProtobufApi.setProtobufEnabled).toHaveBeenCalledWith(false)
    })
  })
})