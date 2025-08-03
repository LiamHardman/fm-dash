/**
 * CSV Export Utility for Player Data
 * Handles exporting filtered player datasets to CSV format
 */

import { safeMerge } from './security.js'

/**
 * Escape CSV value to prevent injection
 * @param {*} value - The value to escape
 * @returns {string} CSV-safe string
 */
function _escapeCSVValue(value) {
  if (value === null || value === undefined) return ''

  // Convert to string and escape quotes
  const stringValue = String(value)
  if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
    return `"${stringValue.replace(/"/g, '""')}"`
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

  const columns = {
    basic: [
      { key: 'name', label: 'Name' },
      { key: 'age', label: 'Age' },
      { key: 'nationality', label: 'Nationality' },
      { key: 'nationalityIso', label: 'Nationality ISO' },
      { key: 'nationality_fifa_code', label: 'Nationality FIFA Code' },
      { key: 'club', label: 'Club' },
      { key: 'division', label: 'Division' },
      { key: 'position', label: 'Position' },
      { key: 'transferValue', label: 'Transfer Value' },
      { key: 'wage', label: 'Wage' },
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
      { key: 'GK', label: 'Goalkeeping' },
    ],
    attributes: samplePlayer?.attributes
      ? Object.keys(samplePlayer.attributes).map((key) => ({
          key: `attributes.${key}`,
          label: key,
        }))
      : [],
    personal: [
      { key: 'personality', label: 'Personality' },
      { key: 'media_handling', label: 'Media Handling' },
      { key: 'attributeMasked', label: 'Attributes Masked' },
    ],
  }

  // Add role ratings if available
  if (samplePlayer?.roleSpecificOveralls && Array.isArray(samplePlayer.roleSpecificOveralls)) {
    const allRoleNames = new Set()

    // Collect all unique role names from all players
    for (const player of players) {
      if (player.roleSpecificOveralls && Array.isArray(player.roleSpecificOveralls)) {
        for (const role of player.roleSpecificOveralls) {
          if (role.roleName) {
            allRoleNames.add(role.roleName)
          }
        }
      }
    }

    const roleRatings = Array.from(allRoleNames).map((roleName) => ({
      key: `roleRating.${roleName}`,
      label: `${roleName} Rating`,
    }))

    if (roleRatings.length > 0) {
      columns.roleRatings = roleRatings
    }
  }

  return columns
}

/**
 * Get nested value from object using dot notation
 * @param {Object} obj - The object to get value from
 * @param {string} path - The path to the value (e.g., 'attributes.technical.passing')
 * @returns {*} The value at the path
 */
function _getNestedValue(obj, path) {
  return path.split('.').reduce((current, key) => {
    return current && current[key] !== undefined ? current[key] : ''
  }, obj)
}

/**
 * Export players to CSV format using the backend export API
 * @param {string} datasetId - Dataset ID to export
 * @param {string} format - Export format ('csv' or 'json')
 * @param {string} filename - Optional filename (defaults to auto-generated)
 * @returns {Promise<void>} Promise that resolves when download starts
 */
export async function exportPlayersToCSV(datasetId, format = 'csv', filename = null) {
  if (!datasetId) {
    throw new Error('Dataset ID is required for export')
  }

  console.log(`Exporting dataset ${datasetId} in ${format} format`)

  try {
    // Call the backend export API
    const response = await fetch(`/api/export/${datasetId}?format=${format}`, {
      method: 'GET',
      headers: {
        Accept: format === 'csv' ? 'text/csv' : 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`Export failed: ${response.status} ${response.statusText}`)
    }

    // Get the filename from the response headers or generate one
    let finalFilename = filename
    if (!finalFilename) {
      const contentDisposition = response.headers.get('Content-Disposition')
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="(.+)"/)
        if (match) {
          finalFilename = match[1]
        }
      }

      if (!finalFilename) {
        const timestamp = new Date().toISOString().split('T')[0]
        finalFilename = `fm_players_export_${timestamp}.${format}`
      }
    }

    // Download the file
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = finalFilename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    console.log(`Export completed: ${finalFilename}`)
  } catch (error) {
    console.error('Export failed:', error)
    throw error
  }
}

/**
 * Download CSV content as file
 * @param {string} csvContent - The CSV content string
 * @param {string} filename - The filename for download
 */
function _downloadCSV(csvContent, filename) {
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
      return ['name', 'age', 'nationality', 'club', 'position', 'transferValue', 'wage', 'Overall']

    case 'detailed':
      return [
        'name',
        'age',
        'nationality',
        'club',
        'position',
        'transferValue',
        'wage',
        'Overall',
        ...availableColumns.ratings
          .filter((col) => col.key !== 'Potential' && col.key !== 'Overall')
          .map((col) => col.key),
        'personality',
        'media_handling',
      ]

    case 'scout':
      return [
        'name',
        'age',
        'nationality',
        'club',
        'position',
        'Overall',
        'transferValue',
        'wage',
        'personality',
      ]

    case 'analysis': {
      const analysisColumns = [
        'name',
        'club',
        'position',
        'Overall',
        ...availableColumns.ratings
          .filter((col) => col.key !== 'Potential' && col.key !== 'Overall')
          .map((col) => col.key),
        ...availableColumns.attributes.slice(0, 10).map((col) => col.key), // First 10 attributes
      ]
      // Add top 5 role ratings if available
      if (availableColumns.roleRatings) {
        analysisColumns.push(...availableColumns.roleRatings.slice(0, 5).map((col) => col.key))
      }
      return analysisColumns
    }

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
  const missingFields = requiredFields.filter((field) => !Object.hasOwn(players[0], field))

  if (missingFields.length > 0) {
    errors.push(`Missing required fields: ${missingFields.join(', ')}`)
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
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
  let finalFilename = filename
  if (!finalFilename) {
    const timestamp = new Date().toISOString().split('T')[0]
    finalFilename = `fm_players_full_dataset_${timestamp}.json`
  }

  // Create the JSON export object with metadata
  const exportData = {
    metadata: {
      exportDate: new Date().toISOString(),
      totalPlayers: players.length,
      exportType: 'full_dataset',
      version: '1.0',
    },
    players: players,
  }

  // Convert to JSON string with proper formatting
  const jsonContent = JSON.stringify(exportData, null, 2)

  // Create and download file
  downloadJSON(jsonContent, finalFilename)
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

/**
 * Get columns based on export options selected in the modal
 * @param {Object} exportOptions - Export options from the modal
 * @param {Array} players - Array of players to determine available columns
 * @returns {Array} Array of column keys to include
 */
export function getColumnsFromExportOptions(exportOptions, players = []) {
  if (!exportOptions || !exportOptions.options) {
    return getDefaultExportColumns('basic', players)
  }

  const availableColumns = getAvailableColumns(players)
  const selectedColumns = []
  const { options } = exportOptions

  // Basic Info
  if (options.basicInfo) {
    selectedColumns.push(...availableColumns.basic.map((col) => col.key))
  }

  // FIFA Stats (Overall ratings)
  if (options.fifahStats) {
    selectedColumns.push(...availableColumns.ratings.map((col) => col.key))
  }

  // FM Attributes
  if (options.fmStats) {
    selectedColumns.push(...availableColumns.attributes.map((col) => col.key))
  }

  // Role Ratings
  if (options.roleRatings && availableColumns.roleRatings) {
    selectedColumns.push(...availableColumns.roleRatings.map((col) => col.key))
  }

  // Performance Percentiles (add performance data if available)
  if (options.performancePercentiles && players.length > 0) {
    const samplePlayer = players[0]
    if (samplePlayer?.performancePercentiles) {
      // Add performance percentile columns for different position groups
      for (const group of Object.keys(samplePlayer.performancePercentiles)) {
        const groupData = samplePlayer.performancePercentiles[group]
        if (groupData && typeof groupData === 'object') {
          // Add individual percentile stats for this group
          for (const statKey of Object.keys(groupData)) {
            selectedColumns.push(`performancePercentiles.${group}.${statKey}`)
          }
        }
      }
    }
  }

  // Contract Info
  if (options.contractInfo) {
    selectedColumns.push('transferValue', 'wage')
    // Add contract expiry if available
    if (players.length > 0 && players[0].contractExpiry) {
      selectedColumns.push('contractExpiry')
    }
  }

  // Personal Info
  if (options.personalInfo) {
    selectedColumns.push(...availableColumns.personal.map((col) => col.key))
    // Add additional personal info if available
    if (players.length > 0) {
      const samplePlayer = players[0]
      if (samplePlayer.foot) selectedColumns.push('foot')
      if (samplePlayer.height) selectedColumns.push('height')
      if (samplePlayer.weight) selectedColumns.push('weight')
    }
  }

  // Remove duplicates and return
  return [...new Set(selectedColumns)]
}

/**
 * Export players with custom options from the modal
 * @param {Array} players - Array of player objects to export
 * @param {Object} exportOptions - Export options from the modal
 * @param {string} filename - Optional filename (defaults to auto-generated)
 * @returns {Promise<void>} Promise that resolves when download starts
 */
export async function exportPlayersWithOptions(players, exportOptions, filename = null) {
  if (!players || players.length === 0) {
    throw new Error('No players to export')
  }

  const selectedColumns = getColumnsFromExportOptions(exportOptions, players)

  if (exportOptions.format === 'csv') {
    await exportPlayersToCSV(players, selectedColumns, filename)
  } else if (exportOptions.format === 'json') {
    await exportPlayersToJSONWithOptions(players, exportOptions, filename)
  } else {
    throw new Error('Invalid export format')
  }
}

/**
 * Export players to JSON with custom options
 * @param {Array} players - Array of player objects to export
 * @param {Object} exportOptions - Export options from the modal
 * @param {string} filename - Optional filename (defaults to auto-generated)
 * @returns {Promise<void>} Promise that resolves when download starts
 */
export async function exportPlayersToJSONWithOptions(players, exportOptions, filename = null) {
  if (!players || players.length === 0) {
    throw new Error('No players to export')
  }

  // Generate filename if not provided
  let finalFilename = filename
  if (!finalFilename) {
    const timestamp = new Date().toISOString().split('T')[0]
    const presetSuffix = exportOptions.preset ? `_${exportOptions.preset}` : '_custom'
    finalFilename = `fm_players${presetSuffix}_${timestamp}.json`
  }

  // Debug: Log the first player to see what data is available
  if (players.length > 0) {
    const samplePlayer = players[0]
    console.log('JSON Export sample player data:', {
      name: samplePlayer.name,
      club: samplePlayer.club,
      transferValue: samplePlayer.transfer_value,
      transferValueAmount: samplePlayer.transferValueAmount,
      hasAttributes: !!samplePlayer.attributes,
      attributesCount: Object.keys(samplePlayer.attributes || {}).length,
      hasPerformancePercentiles: !!samplePlayer.performancePercentiles,
      percentileGroups: Object.keys(samplePlayer.performancePercentiles || {}),
      hasRoleSpecificOveralls: !!samplePlayer.roleSpecificOveralls,
      roleCount: samplePlayer.roleSpecificOveralls?.length || 0,
      overall: samplePlayer.Overall,
      fifaStats: {
        PAC: samplePlayer.PAC,
        SHO: samplePlayer.SHO,
        PAS: samplePlayer.PAS,
        DRI: samplePlayer.DRI,
        DEF: samplePlayer.DEF,
        PHY: samplePlayer.PHY,
        GK: samplePlayer.GK,
      },
    })
  }

  // Filter players data based on export options
  const filteredPlayers = players.map((player) => {
    const filteredPlayer = {}
    const { options } = exportOptions

    if (options.basicInfo) {
      const basicInfo = {
        name: player.name,
        age: player.age,
        nationality: player.nationality,
        nationalityIso: player.nationalityIso,
        nationality_fifa_code: player.nationality_fifa_code,
        club: player.club,
        division: player.division,
        position: player.position,
        shortPositions: player.shortPositions,
        parsedPositions: player.parsedPositions,
        positionGroups: player.positionGroups,
      }
      Object.assign(filteredPlayer, safeMerge(filteredPlayer, basicInfo))
    }

    if (options.fifahStats) {
      const fifaStats = {
        Overall: player.Overall,
        Potential: player.Potential,
        PAC: player.PAC || player.pac,
        SHO: player.SHO || player.sho,
        PAS: player.PAS || player.pas,
        DRI: player.DRI || player.dri,
        DEF: player.DEF || player.def,
        PHY: player.PHY || player.phy,
        GK: player.GK || player.gk,
      }
      Object.assign(filteredPlayer, safeMerge(filteredPlayer, fifaStats))
    }

    if (options.fmStats && player.attributes) {
      filteredPlayer.attributes = player.attributes
    }

    if (options.roleRatings && player.roleSpecificOveralls) {
      filteredPlayer.roleSpecificOveralls = player.roleSpecificOveralls
    }

    if (options.performancePercentiles && player.performancePercentiles) {
      filteredPlayer.performancePercentiles = player.performancePercentiles
    }

    if (options.contractInfo) {
      const contractInfo = {
        transferValue: player.transfer_value || player.transferValue,
        transferValueAmount: player.transferValueAmount,
        wage: player.wage,
        wageAmount: player.wageAmount,
        contractExpiry: player.contractExpiry,
      }
      Object.assign(filteredPlayer, safeMerge(filteredPlayer, contractInfo))
    }

    if (options.personalInfo) {
      const personalInfo = {
        personality: player.personality,
        media_handling: player.media_handling,
        attributeMasked: player.attributeMasked,
        foot: player.foot,
        height: player.height,
        weight: player.weight,
      }
      Object.assign(filteredPlayer, safeMerge(filteredPlayer, personalInfo))
    }

    return filteredPlayer
  })

  // Create the JSON export object with metadata
  const exportData = {
    metadata: {
      exportDate: new Date().toISOString(),
      totalPlayers: players.length,
      exportType: exportOptions.preset || 'custom',
      exportOptions: exportOptions.options,
      version: '1.0',
    },
    players: filteredPlayers,
  }

  // Convert to JSON string with proper formatting
  const jsonContent = JSON.stringify(exportData, null, 2)

  // Create and download file
  downloadJSON(jsonContent, finalFilename)
}
