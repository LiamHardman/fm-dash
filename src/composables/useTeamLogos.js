import { computed } from 'vue'
import teamsData from '../utils/teams_data.json'

/**
 * Composable for handling team logos
 * Maps team names to their numerical IDs and provides logo URLs
 */
export function useTeamLogos(options = {}) {
  const { 
    similarityThreshold = 0.7,
    enableFuzzyMatching = true 
  } = options

  /**
   * Normalize a team name for better matching
   * @param {string} name - The team name to normalize
   * @returns {string} - Normalized team name
   */
  const normalizeTeamName = (name) => {
    if (!name) return ''
    
    return name
      .toLowerCase()
      .trim()
      // Remove common prefixes
      .replace(/^(fc|cf|ac|sc|cd|ud|real|club|athletic|atletico|athletico|deportivo|sporting)\s+/i, '')
      // Remove common suffixes
      .replace(/\s+(fc|cf|ac|sc|cd|ud|club|united|city|town|rovers|wanderers|albion|villa|county|athletic|atletico|athletico|deportivo|sporting|utd|f\.?c\.?|c\.?f\.?|s\.?c\.?)$/i, '')
      // Remove dots and common abbreviations
      .replace(/\./g, '')
      .replace(/\s+/g, ' ')
      .trim()
  }

  /**
   * Calculate similarity between two normalized strings
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @returns {number} - Similarity score (0-1)
   */
  const calculateSimilarity = (str1, str2) => {
    const norm1 = normalizeTeamName(str1)
    const norm2 = normalizeTeamName(str2)
    
    // Exact match after normalization
    if (norm1 === norm2) return 1.0
    
    // Check if one is contained in the other
    if (norm1.includes(norm2) || norm2.includes(norm1)) {
      return 0.8
    }
    
    // Simple word-based similarity
    const words1 = norm1.split(' ').filter(w => w.length > 1)
    const words2 = norm2.split(' ').filter(w => w.length > 1)
    
    if (words1.length === 0 || words2.length === 0) return 0
    
    const matchingWords = words1.filter(word1 => 
      words2.some(word2 => word1 === word2 || word1.includes(word2) || word2.includes(word1))
    )
    
    return matchingWords.length / Math.max(words1.length, words2.length)
  }

  /**
   * Get the numerical team ID for a given team name
   * @param {string} teamName - The display name of the team
   * @returns {string|null} - The numerical team ID or null if not found
   */
  const getTeamId = (teamName) => {
    if (!teamName) return null
    
    // First try exact match
    for (const [numericalId, name] of Object.entries(teamsData)) {
      if (name === teamName) {
        return numericalId
      }
    }
    
    // If fuzzy matching is disabled, return null
    if (!enableFuzzyMatching) return null
    
    // Then try fuzzy matching
    let bestMatch = null
    let bestScore = 0
    
    for (const [numericalId, name] of Object.entries(teamsData)) {
      const similarity = calculateSimilarity(teamName, name)
      if (similarity >= similarityThreshold && similarity > bestScore) {
        bestScore = similarity
        bestMatch = numericalId
      }
    }
    
    return bestMatch
  }

  /**
   * Get detailed match information for debugging
   * @param {string} teamName - The display name of the team
   * @returns {Object|null} - Match details { id, name, score } or null
   */
  const getTeamMatchDetails = (teamName) => {
    if (!teamName) return null
    
    // First try exact match
    for (const [numericalId, name] of Object.entries(teamsData)) {
      if (name === teamName) {
        return { id: numericalId, name, score: 1.0 }
      }
    }
    
    // If fuzzy matching is disabled, return null
    if (!enableFuzzyMatching) return null
    
    // Then try fuzzy matching
    let bestMatch = null
    let bestScore = 0
    
    for (const [numericalId, name] of Object.entries(teamsData)) {
      const similarity = calculateSimilarity(teamName, name)
      if (similarity >= similarityThreshold && similarity > bestScore) {
        bestScore = similarity
        bestMatch = { id: numericalId, name, score: similarity }
      }
    }
    
    return bestMatch
  }

  /**
   * Get the logo URL for a team
   * @param {string} teamName - The display name of the team
   * @returns {string|null} - The logo URL or null if no team ID found
   */
  const getTeamLogoUrl = (teamName) => {
    const teamId = getTeamId(teamName)
    if (!teamId) return null
    
    return `/api/logos?teamId=${encodeURIComponent(teamId)}`
  }

  /**
   * Computed property factory for team logo URLs
   * @param {import('vue').Ref<string>} teamNameRef - Reactive team name
   * @returns {import('vue').ComputedRef<string|null>} - Reactive logo URL
   */
  const createTeamLogoUrl = (teamNameRef) => {
    return computed(() => getTeamLogoUrl(teamNameRef.value))
  }

  /**
   * Check if a team has a logo available
   * @param {string} teamName - The display name of the team
   * @returns {boolean} - True if team ID exists (logo may exist)
   */
  const hasTeamLogo = (teamName) => {
    return getTeamId(teamName) !== null
  }

  return {
    getTeamId,
    getTeamLogoUrl,
    createTeamLogoUrl,
    hasTeamLogo,
    getTeamMatchDetails,
    // Utility methods
    normalizeTeamName,
    calculateSimilarity
  }
} 