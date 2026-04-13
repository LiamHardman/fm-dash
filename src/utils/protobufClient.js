/**
 * ProtobufClient - Core utility for handling protobuf API responses
 */

// Remove all dynamic import attempts and fallback paths
// Only use the standard import
import protobuf from 'protobufjs'
import logger from './logger.js'
import protoDescriptor from './proto-descriptor.js'

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
          // logger.info('Attempting to load protobufjs library...')
          // The protobuf library is already imported at the top
          // No additional import needed here
          // logger.info('Successfully loaded protobufjs library', {
          //   hasRoot: !!protobuf.Root,
          //   hasFromJSON: !!protobuf.Root?.fromJSON,
          //   protobufType: typeof protobuf,
          //   protobufKeys: Object.keys(protobuf || {}),
          //   rootType: typeof protobuf.Root
          // })
        }

        const definitionsLoaded = await this.loadProtobufDefinitions()
        if (!definitionsLoaded) {
          logger.error('Failed to load protobuf definitions, protobuf will not work')
          this.protobufSupported = false
        } else {
          // logger.info('Protobuf definitions loaded successfully')
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
      // Access the Root constructor
      const Root = protobuf.Root

      if (!Root) {
        logger.error('Protobuf library not available', {
          hasProtobuf: !!protobuf,
          hasRoot: !!protobuf?.Root,
          protobufType: typeof protobuf,
          protobufKeys: Object.keys(protobuf || {}),
        })
        this.protoDefinitions = null
        return false
      }

      // logger.info('Loading protobuf definitions...')

      this.protoDefinitions = Root.fromJSON(protoDescriptor)

      // logger.info('Successfully loaded protobuf definitions')
      return true
    } catch (error) {
      logger.error('Failed to load protobuf definitions:', error)
      logger.error('Protobuf library state:', {
        protobufAvailable: !!protobuf,
        protobufRoot: !!protobuf?.Root,
        protobufFromJSON: !!protobuf?.Root?.fromJSON,
      })
      this.protoDefinitions = null
      return false
    }
  }

  /**
   * Fetch data with protobuf support
   */
  async fetchWithProtobuf(url, options, messageType) {
    await this.initialize()

    if (!this.protobufSupported || !this.protobufEnabled) {
      return this.fallbackToJSON(url, options)
    }

    try {
      const headers = options.headers || {}
      headers.Accept = 'application/x-protobuf'

      const response = await fetch(url, {
        ...options,
        headers,
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
          payloadSize: buffer.byteLength,
        },
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
      logger.error('Protobuf definitions not loaded. Status:', {
        protobufSupported: this.protobufSupported,
        protobufEnabled: this.protobufEnabled,
        initialized: this.initialized,
        protoDefinitions: !!this.protoDefinitions,
      })
      throw new Error('Protobuf definitions not loaded')
    }

    try {
      const MessageType = this.protoDefinitions.lookupType(messageType)
      if (!MessageType) {
        throw new Error(`Message type '${messageType}' not found in protobuf definitions`)
      }

      const message = MessageType.decode(new Uint8Array(buffer))

      const decodedData = MessageType.toObject(message, {
        longs: String,
        enums: String,
        bytes: String,
        defaults: true,
        arrays: true,
        objects: true,
        oneofs: true,
      })

      return decodedData
    } catch (error) {
      logger.error('Protobuf decoding error:', {
        messageType,
        bufferSize: buffer?.byteLength,
        error: error.message,
        protoDefinitionsAvailable: !!this.protoDefinitions,
      })

      const decodingError = new Error(`Failed to decode protobuf response: ${error.message}`)
      decodingError.name = 'ProtobufDecodingError'
      decodingError.originalError = error
      decodingError.bufferSize = buffer?.byteLength
      throw decodingError
    }
  }

  /**
   * Handle a non-protobuf response
   */
  async handleNonProtobufResponse(response) {
    const contentType = response.headers.get('Content-Type')

    if (contentType?.includes('application/json')) {
      const jsonData = await response.json()
      return {
        ...jsonData,
        _protobuf: {
          format: 'json',
          fallbackReason: 'server_returned_json',
        },
      }
    }

    const textData = await response.text()
    return {
      data: textData,
      _protobuf: {
        format: 'text',
        fallbackReason: 'unknown_content_type',
      },
    }
  }

  /**
   * Fall back to JSON request
   */
  async fallbackToJSON(url, options = {}) {
    try {
      const headers = options.headers || {}
      headers.Accept = 'application/json'

      const response = await fetch(url, {
        ...options,
        headers,
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const jsonData = await response.json()

      return {
        ...jsonData,
        _protobuf: {
          format: 'json',
          fallbackReason: this.getFallbackReason(),
        },
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
   * Force re-initialization of the protobuf client
   */
  async reinitialize() {
    logger.info('Forcing protobuf client re-initialization...')
    this.initialized = false
    this.initializationPromise = null
    this.protoDefinitions = null
    this.protobufSupported = false

    return await this.initialize()
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
      definitionsLoaded: !!this.protoDefinitions,
      protobufLibraryAvailable: !!protobuf,
      protobufRootAvailable: !!protobuf?.Root,
      protobufFromJSONAvailable: !!protobuf?.Root?.fromJSON,
    }
  }
}

// Export a singleton instance
const protobufClient = new ProtobufClient()
export default protobufClient
