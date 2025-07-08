/**
 * JavaScript Security Utilities
 * Provides validation and sanitization functions to prevent common security vulnerabilities
 */

/**
 * Safe property names that are allowed for dynamic access
 * This prevents prototype pollution and remote property injection
 */
const SAFE_PLAYER_PROPERTIES = new Set([
  'name', 'age', 'nationality', 'club', 'position', 'shortPositions',
  'Overall', 'Potential', 'PAC', 'SHO', 'PAS', 'DRI', 'DEF', 'PHY', 'GK',
  'transferValue', 'transferValueAmount', 'wage', 'wageAmount', 'contractExpiry',
  'personality', 'media_handling', 'foot', 'height', 'weight',
  'attributes', 'roleSpecificOveralls', 'performancePercentiles'
])

/**
 * Dangerous prototype properties that should never be set
 */
const DANGEROUS_PROPS = new Set([
  '__proto__', 'constructor', 'prototype', 'toString', 'valueOf',
  'hasOwnProperty', 'isPrototypeOf', 'propertyIsEnumerable'
])

/**
 * Validates if a property name is safe for dynamic access
 * @param {string} propertyName - The property name to validate
 * @returns {boolean} - True if safe, false otherwise
 */
export function isValidPlayerProperty(propertyName) {
  if (typeof propertyName !== 'string') return false
  if (propertyName.includes('__')) return false
  if (DANGEROUS_PROPS.has(propertyName)) return false
  return SAFE_PLAYER_PROPERTIES.has(propertyName)
}

/**
 * Safely access a property on a player object
 * @param {Object} player - The player object
 * @param {string} fieldKey - The field key to access
 * @returns {*} - The property value or undefined if invalid
 */
export function safeGetPlayerProperty(player, fieldKey) {
  if (!player || typeof player !== 'object') return undefined
  if (!isValidPlayerProperty(fieldKey)) return undefined
  
  return player[fieldKey]
}

/**
 * Validates if a URL belongs to a trusted domain
 * @param {string} url - The URL to validate
 * @param {string[]} trustedDomains - Array of trusted domain patterns
 * @returns {boolean} - True if URL is from a trusted domain
 */
export function isValidImageUrl(url, trustedDomains = ['flagcdn.com']) {
  if (typeof url !== 'string') return false
  
  try {
    const urlObj = new URL(url)
    
    // Check if the hostname ends with any of the trusted domains
    return trustedDomains.some(domain => {
      // Ensure the domain is at the end of the hostname with proper boundary
      const hostname = urlObj.hostname.toLowerCase()
      const trustedDomain = domain.toLowerCase()
      
      // Exact match or subdomain of trusted domain
      return hostname === trustedDomain || 
             hostname.endsWith('.' + trustedDomain)
    })
  } catch (error) {
    return false
  }
}

/**
 * Safely merge objects without prototype pollution
 * @param {Object} target - Target object to merge into
 * @param {Object} source - Source object to merge from
 * @returns {Object} - Merged object
 */
export function safeMerge(target, source) {
  if (!target || typeof target !== 'object') target = {}
  if (!source || typeof source !== 'object') return target
  
  const result = { ...target }
  
  for (const key in source) {
    // Skip dangerous properties
    if (DANGEROUS_PROPS.has(key)) continue
    
    // Only copy own properties
    if (!Object.prototype.hasOwnProperty.call(source, key)) continue
    
    // Skip properties starting with double underscore
    if (key.startsWith('__')) continue
    
    result[key] = source[key]
  }
  
  return result
}

/**
 * Validates PostMessage origin against allowed origins
 * @param {string} origin - The origin to validate
 * @param {string[]} allowedOrigins - Array of allowed origins
 * @returns {boolean} - True if origin is allowed
 */
export function isValidOrigin(origin, allowedOrigins = []) {
  if (typeof origin !== 'string') return false
  
  // Always allow same origin
  if (origin === window.location.origin) return true
  
  // Check against explicitly allowed origins
  return allowedOrigins.includes(origin)
}

/**
 * Sanitizes a string for safe use in contexts where injection is possible
 * @param {string} input - The input string to sanitize
 * @param {number} maxLength - Maximum allowed length
 * @returns {string} - Sanitized string
 */
export function sanitizeString(input, maxLength = 1000) {
  if (typeof input !== 'string') return ''
  
  // Truncate if too long
  if (input.length > maxLength) {
    input = input.substring(0, maxLength)
  }
  
  // Remove potentially dangerous characters
  return input
    .replace(/[<>'"]/g, '') // Remove HTML/JS injection chars
    .replace(/\x00/g, '') // Remove null bytes
    .trim()
}

/**
 * Creates a secure property getter with validation
 * @param {Object} obj - The object to get properties from
 * @param {string} property - The property name
 * @param {Set} allowedProperties - Set of allowed property names
 * @returns {*} - Property value or undefined if invalid
 */
export function securePropertyAccess(obj, property, allowedProperties) {
  if (!obj || typeof obj !== 'object') return undefined
  if (typeof property !== 'string') return undefined
  if (!allowedProperties.has(property)) return undefined
  if (DANGEROUS_PROPS.has(property)) return undefined
  
  return obj[property]
}

/**
 * Validates and sanitizes filter input
 * @param {Object} filters - Filter object to validate
 * @returns {Object} - Sanitized filter object
 */
export function sanitizeFilters(filters) {
  if (!filters || typeof filters !== 'object') return {}
  
  const sanitized = {}
  
  for (const [key, value] of Object.entries(filters)) {
    // Skip dangerous properties
    if (DANGEROUS_PROPS.has(key) || key.startsWith('__')) continue
    
    // Sanitize string values
    if (typeof value === 'string') {
      sanitized[key] = sanitizeString(value, 200)
    }
    // Validate number ranges
    else if (typeof value === 'number') {
      if (Number.isFinite(value) && value >= 0) {
        sanitized[key] = value
      }
    }
    // Handle arrays (like positions, clubs)
    else if (Array.isArray(value)) {
      sanitized[key] = value
        .filter(item => typeof item === 'string')
        .map(item => sanitizeString(item, 100))
        .slice(0, 50) // Limit array size
    }
    // Handle objects (like ranges)
    else if (value && typeof value === 'object' && !Array.isArray(value)) {
      sanitized[key] = safeMerge({}, value)
    }
  }
  
  return sanitized
} 