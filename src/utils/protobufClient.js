/**
 * ProtobufClient - Core utility for handling protobuf API responses
 */

import logger from './logger.js'

// Protobuf library reference
let protobuf = null

class ProtobufClient {
  constructor() {
    this.protobufSupported = false
    this.protoDefinitions = null
    this.initialized = false
    this.initializationPromise = null
    this.serverSupportsProtobuf = null
    this.protobufEnabled = this.getConfiguredProtobufEnabled()
  }

  /**
   * Get protobuf enabled status from configuration
   */
  getConfiguredProtobufEnabled() {
    try {
      if (window.APP_CONFIG?.PROTOBUF_ENABLED !== undefined) {
        return !!window.APP_CONFIG.PROTOBUF_ENABLED
      }
      
      const userPreference = localStorage.getItem('protobuf_enabled')
      if (userPreference !== null) {
        return userPreference === 'true'
      }
      
      return true
    } catch (error) {
      logger.warn('Error reading protobuf configuration, defaulting to enabled:', error)
      return true
    }
  }

  /**
   * Initialize the protobuf client
   */
  async initialize() {
    if (this.initialized || this.initializationPromise) {
      return this.initializationPromise
    }

    this.initializationPromise = (async () => {
      try {
        this.protobufSupported = await this.detectProtobufSupport()
        
        if (this.protobufSupported) {
          try {
            const protobufModule = await import('protobufjs/minimal')
            protobuf = protobufModule.default
            logger.info('Successfully loaded protobufjs library')
          } catch (error) {
            logger.warn('Failed to load protobufjs, using mock implementation:', error)
            protobuf = {
              Root: {
                fromJSON: (json) => ({
                  lookupType: (type) => ({
                    decode: (buffer) => ({ decoded: true, type }),
                    toObject: (decoded) => ({ ...decoded, converted: true })
                  })
                })
              }
            }
          }
          
          const definitionsLoaded = await this.loadProtobufDefinitions()
          if (!definitionsLoaded) {
            logger.error('Failed to load protobuf definitions, protobuf will not work')
            this.protobufSupported = false
          }
        }
        
        this.initialized = true
        return this.protobufSupported
      } catch (error) {
        logger.error('Failed to initialize protobuf client:', error)
        this.protobufSupported = false
        this.initialized = true
        return false
      }
    })()

    return this.initializationPromise
  }

  /**
   * Detect if protobuf is supported in the current environment
   */
  async detectProtobufSupport() {
    try {
      const hasArrayBuffer = typeof ArrayBuffer !== 'undefined'
      const hasTextEncoder = typeof TextEncoder !== 'undefined'
      const hasTextDecoder = typeof TextDecoder !== 'undefined'
      
      return hasArrayBuffer && hasTextEncoder && hasTextDecoder
    } catch (error) {
      logger.warn('Error detecting protobuf support:', error)
      return false
    }
  }

  /**
   * Load protobuf message definitions
   */
  async loadProtobufDefinitions() {
    try {
      this.protoDefinitions = protobuf.Root.fromJSON({
        nested: {
          api: {
            nested: {
              ResponseMetadata: {
                fields: {
                  timestamp: { type: "int64", id: 1 },
                  apiVersion: { type: "string", id: 2 },
                  fromCache: { type: "bool", id: 3 },
                  requestId: { type: "string", id: 4 },
                  totalCount: { type: "int32", id: 5 }
                }
              },
              PlayerDataResponse: {
                fields: {
                  players: { rule: "repeated", type: "player.Player", id: 1 },
                  currencySymbol: { type: "string", id: 2 },
                  metadata: { type: "ResponseMetadata", id: 3 }
                }
              },
              RolesResponse: {
                fields: {
                  roles: { rule: "repeated", type: "string", id: 1 },
                  metadata: { type: "ResponseMetadata", id: 2 }
                }
              },
              GenericResponse: {
                fields: {
                  data: { type: "string", id: 1 },
                  metadata: { type: "ResponseMetadata", id: 2 }
                }
              }
            }
          },
          player: {
            nested: {
              Player: {
                fields: {
                  uid: { type: "int64", id: 1 },
                  name: { type: "string", id: 2 },
                  position: { type: "string", id: 3 },
                  age: { type: "string", id: 4 },
                  club: { type: "string", id: 5 },
                  overall: { type: "int32", id: 35 }
                }
              }
            }
          }
        }
      })
      
      return true
    } catch (error) {
      logger.error('Failed to load protobuf definitions:', error)
      this.protoDefinitions = null
      return false
    }
  }

  /**
   * Fetch data with protobuf support
   */
  async fetchWithProtobuf(url, options = {}, messageType) {
    await this.initialize()
    
    if (!this.protobufSupported || !this.protobufEnabled) {
      return this.fallbackToJSON(url, options)
    }
    
    try {
      const headers = options.headers || {}
      headers['Accept'] = 'application/x-protobuf'
      
      const response = await fetch(url, {
        ...options,
        headers
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      
      const contentType = response.headers.get('Content-Type')
      if (!contentType || !contentType.includes('application/x-protobuf')) {
        return this.handleNonProtobufResponse(response)
      }
      
      const buffer = await response.arrayBuffer()
      const decodedData = await this.decodeProtobufResponse(buffer, messageType)
      
      return {
        ...decodedData,
        _protobuf: {
          format: 'protobuf',
          payloadSize: buffer.byteLength
        }
      }
    } catch (error) {
      logger.error('Protobuf request failed, falling back to JSON:', error)
      return this.fallbackToJSON(url, options)
    }
  }

  /**
   * Decode a protobuf response
   */
  async decodeProtobufResponse(buffer, messageType) {
    if (!this.protoDefinitions) {
      throw new Error('Protobuf definitions not loaded')
    }
    
    try {
      const MessageType = this.protoDefinitions.lookupType(messageType)
      const message = MessageType.decode(new Uint8Array(buffer))
      
      return MessageType.toObject(message, {
        longs: String,
        enums: String,
        bytes: String,
        defaults: true,
        arrays: true,
        objects: true,
        oneofs: true
      })
    } catch (error) {
      const decodingError = new Error(`Failed to decode protobuf response: ${error.message}`)
      decodingError.name = 'ProtobufDecodingError'
      throw decodingError
    }
  }

  /**
   * Handle a non-protobuf response
   */
  async handleNonProtobufResponse(response) {
    const contentType = response.headers.get('Content-Type')
    
    if (contentType && contentType.includes('application/json')) {
      const jsonData = await response.json()
      return {
        ...jsonData,
        _protobuf: {
          format: 'json',
          fallbackReason: 'server_returned_json'
        }
      }
    }
    
    const textData = await response.text()
    return {
      data: textData,
      _protobuf: {
        format: 'text',
        fallbackReason: 'unknown_content_type'
      }
    }
  }

  /**
   * Fall back to JSON request
   */
  async fallbackToJSON(url, options = {}) {
    try {
      const headers = options.headers || {}
      headers['Accept'] = 'application/json'
      
      const response = await fetch(url, {
        ...options,
        headers
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      
      const jsonData = await response.json()
      
      return {
        ...jsonData,
        _protobuf: {
          format: 'json',
          fallbackReason: this.getFallbackReason()
        }
      }
    } catch (error) {
      logger.error('Error in JSON fallback:', error)
      throw error
    }
  }

  /**
   * Get the reason for falling back to JSON
   */
  getFallbackReason() {
    if (!this.protobufSupported) {
      return 'client_unsupported'
    }
    if (!this.protobufEnabled) {
      return 'client_disabled'
    }
    return 'unknown'
  }

  /**
   * Enable or disable protobuf support
   */
  setProtobufEnabled(enabled) {
    this.protobufEnabled = !!enabled
  }

  /**
   * Get client capabilities and status
   */
  getStatus() {
    return {
      protobufSupported: this.protobufSupported,
      protobufEnabled: this.protobufEnabled,
      initialized: this.initialized,
      serverSupportsProtobuf: this.serverSupportsProtobuf,
      definitionsLoaded: !!this.protoDefinitions
    }
  }
}

// Export a singleton instance
const protobufClient = new ProtobufClient()
export default protobufClient