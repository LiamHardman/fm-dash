/**
 * CSV Export Utility for Player Data
 * Handles exporting filtered player datasets to CSV format
 */

/**
 * Convert a value to CSV-safe string
 * @param {*} value - The value to convert
 * @returns {string} CSV-safe string
 */
function escapeCSVValue(value) {
  if (value === null || value === undefined) return ''
  
  let stringValue = String(value)
  
  // If the value contains comma, newline, or double quote, wrap it in quotes
  if (stringValue.includes(',') || stringValue.includes('\n') || stringValue.includes('"')) {
    // Escape any double quotes by doubling them
    stringValue = stringValue.replace(/"/g, '""')
    return `"${stringValue}"`
  }
  
  return stringValue
}

/**
 * Get all available columns for player data
 * @param {Array} players - Array of player objects
 * @returns {Object} Object with column categories and their fields
 */
function getAvailableColumns(players) {
  if (!players || players.length === 0) return {}
  
  const samplePlayer = players[0]
  
  return {
    basic: [
      { key: 'name', label: 'Name' },
      { key: 'age', label: 'Age' },
      { key: 'nationality', label: 'Nationality' },
      { key: 'club', label: 'Club' },
      { key: 'position', label: 'Position' },
      { key: 'transferValue', label: 'Transfer Value' },
      { key: 'wage', label: 'Wage' }
    ],
    ratings: [
      { key: 'Overall', label: 'Overall' },
      { key: 'Potential', label: 'Potential' },
      { key: 'PAC', label: 'Pace' },
      { key: 'SHO', label: 'Shooting' },
      { key: 'PAS', label: 'Passing' },
      { key: 'DRI', label: 'Dribbling' },
      { key: 'DEF', label: 'Defending' },
      { key: 'PHY', label: 'Physical' },
      { key: 'GK', label: 'Goalkeeping' }
    ],
    attributes: samplePlayer?.attributes ? Object.keys(samplePlayer.attributes).map(key => ({
      key: `attributes.${key}`,
      label: key
    })) : [],
    personal: [
      { key: 'personality', label: 'Personality' },
      { key: 'media_handling', label: 'Media Handling' }
    ]
  }
}

/**
 * Get value from nested object path
 * @param {Object} obj - The object to get value from
 * @param {string} path - The path (e.g., 'attributes.Pac')
 * @returns {*} The value at the path
 */
function getNestedValue(obj, path) {
  return path.split('.').reduce((current, key) => {
    return current && current[key] !== undefined ? current[key] : ''
  }, obj)
}

/**
 * Export players to CSV format
 * @param {Array} players - Array of player objects to export
 * @param {Array} selectedColumns - Array of column keys to include
 * @param {string} filename - Optional filename (defaults to auto-generated)
 * @returns {Promise<void>} Promise that resolves when download starts
 */
export async function exportPlayersToCSV(players, selectedColumns = null, filename = null) {
  if (!players || players.length === 0) {
    throw new Error('No players to export')
  }
  
  // Get available columns
  const availableColumns = getAvailableColumns(players)
  
  // If no columns specified, use basic + ratings by default
  if (!selectedColumns) {
    selectedColumns = [
      ...availableColumns.basic.map(col => col.key),
      ...availableColumns.ratings.map(col => col.key)
    ]
  }
  
  // Create column mapping
  const columnMap = {}
  Object.values(availableColumns).forEach(category => {
    category.forEach(col => {
      columnMap[col.key] = col.label
    })
  })
  
  // Generate header row
  const headers = selectedColumns.map(key => columnMap[key] || key)
  
  // Generate data rows
  const rows = players.map(player => {
    return selectedColumns.map(key => {
      let value = getNestedValue(player, key)
      
      // Special handling for certain fields
      if (key === 'transferValue') {
        value = player.transferValue || `${player.transferValueAmount || 0}`
      } else if (key === 'wage') {
        value = player.wage || `${player.wageAmount || 0}`
      } else if (key === 'position') {
        value = Array.isArray(player.shortPositions) 
          ? player.shortPositions.join(', ') 
          : (player.position || '')
      }
      
      return escapeCSVValue(value)
    })
  })
  
  // Combine headers and rows
  const csvContent = [headers.join(','), ...rows.map(row => row.join(','))].join('\n')
  
  // Generate filename if not provided
  if (!filename) {
    const timestamp = new Date().toISOString().split('T')[0]
    filename = `fm_players_export_${timestamp}.csv`
  }
  
  // Create and download file
  downloadCSV(csvContent, filename)
}

/**
 * Download CSV content as file
 * @param {string} csvContent - The CSV content string
 * @param {string} filename - The filename for download
 */
function downloadCSV(csvContent, filename) {
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  
  if (link.download !== undefined) {
    // Use HTML5 download attribute
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', filename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } else {
    // Fallback for older browsers
    if (navigator.msSaveBlob) {
      navigator.msSaveBlob(blob, filename)
    }
  }
}

/**
 * Get default export columns for different contexts
 * @param {string} context - The export context ('basic', 'detailed', 'scout', 'analysis')
 * @param {Array} players - Array of players to determine available columns
 * @returns {Array} Array of column keys
 */
export function getDefaultExportColumns(context = 'basic', players = []) {
  const availableColumns = getAvailableColumns(players)
  
  switch (context) {
    case 'basic':
      return [
        'name', 'age', 'nationality', 'club', 'position', 
        'transferValue', 'wage', 'Overall'
      ]
    
    case 'detailed':
      return [
        'name', 'age', 'nationality', 'club', 'position',
        'transferValue', 'wage', 'Overall',
        ...availableColumns.ratings.filter(col => col.key !== 'Potential' && col.key !== 'Overall').map(col => col.key),
        'personality', 'media_handling'
      ]
    
    case 'scout':
      return [
        'name', 'age', 'nationality', 'club', 'position',
        'Overall', 'transferValue', 'wage',
        'personality'
      ]
    
    case 'analysis':
      return [
        'name', 'club', 'position', 'Overall',
        ...availableColumns.ratings.filter(col => col.key !== 'Potential' && col.key !== 'Overall').map(col => col.key),
        ...availableColumns.attributes.slice(0, 10).map(col => col.key) // First 10 attributes
      ]
    
    default:
      return getDefaultExportColumns('basic', players)
  }
}

/**
 * Validate export data before processing
 * @param {Array} players - Array of players to validate
 * @returns {Object} Validation result
 */
export function validateExportData(players) {
  const errors = []
  const warnings = []
  
  if (!Array.isArray(players)) {
    errors.push('Players data must be an array')
    return { valid: false, errors, warnings }
  }
  
  if (players.length === 0) {
    errors.push('No players to export')
    return { valid: false, errors, warnings }
  }
  
  if (players.length > 10000) {
    warnings.push('Large export (>10,000 players) may take some time')
  }
  
  // Check for required fields
  const requiredFields = ['name']
  const missingFields = requiredFields.filter(field => 
    !players[0].hasOwnProperty(field)
  )
  
  if (missingFields.length > 0) {
    errors.push(`Missing required fields: ${missingFields.join(', ')}`)
  }
  
  return {
    valid: errors.length === 0,
    errors,
    warnings
  }
}

/**
 * Export full dataset to JSON format
 * @param {Array} players - Array of all player objects to export
 * @param {string} filename - Optional filename (defaults to auto-generated)
 * @returns {Promise<void>} Promise that resolves when download starts
 */
export async function exportPlayersToJSON(players, filename = null) {
  if (!players || players.length === 0) {
    throw new Error('No players to export')
  }
  
  // Generate filename if not provided
  if (!filename) {
    const timestamp = new Date().toISOString().split('T')[0]
    filename = `fm_players_full_dataset_${timestamp}.json`
  }
  
  // Create the JSON export object with metadata
  const exportData = {
    metadata: {
      exportDate: new Date().toISOString(),
      totalPlayers: players.length,
      exportType: 'full_dataset',
      version: '1.0'
    },
    players: players
  }
  
  // Convert to JSON string with proper formatting
  const jsonContent = JSON.stringify(exportData, null, 2)
  
  // Create and download file
  downloadJSON(jsonContent, filename)
}

/**
 * Download JSON content as file
 * @param {string} jsonContent - The JSON content string
 * @param {string} filename - The filename for download
 */
function downloadJSON(jsonContent, filename) {
  const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' })
  const link = document.createElement('a')
  
  if (link.download !== undefined) {
    // Use HTML5 download attribute
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', filename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } else {
    // Fallback for older browsers
    if (navigator.msSaveBlob) {
      navigator.msSaveBlob(blob, filename)
    }
  }
} 