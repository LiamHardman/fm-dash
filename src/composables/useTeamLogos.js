import { computed } from 'vue'
import teamsData from '../utils/teams_data.json'

/**
 * Composable for handling team logos
 * Maps team names to their numerical IDs and provides logo URLs
 */
export function useTeamLogos(options = {}) {
  const { 
    similarityThreshold = 0.85, // Increased for better precision
    enableFuzzyMatching = true,
    strictMode = false // If true, requires higher similarity and more exact matches
  } = options

  /**
   * More conservative team name normalization
   * @param {string} name - The team name to normalize
   * @returns {string} - Normalized team name
   */
  const normalizeTeamName = (name) => {
    if (!name) return ''
    
    return name
      .toLowerCase()
      .trim()
      // Only remove very common and unambiguous prefixes/suffixes
      .replace(/^(fc|cf)\s+/i, '') // Only FC/CF, not others that might be important
      .replace(/\s+(fc|cf)$/i, '') // Only FC/CF suffixes
      // Handle dots but preserve important abbreviations
      .replace(/\s+f\.c\.?$/i, ' fc')
      .replace(/^f\.c\.?\s+/i, 'fc ')
      // Normalize spacing but preserve structure
      .replace(/\s+/g, ' ')
      .trim()
  }

  /**
   * Enhanced similarity calculation with multiple strategies
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @returns {number} - Similarity score (0-1)
   */
  const calculateSimilarity = (str1, str2) => {
    if (!str1 || !str2) return 0
    
    const norm1 = normalizeTeamName(str1)
    const norm2 = normalizeTeamName(str2)
    
    // Exact match after normalization
    if (norm1 === norm2) return 1.0
    
    // Check for exact substring matches (but be more careful)
    if (norm1.length > 3 && norm2.length > 3) {
      if (norm1.includes(norm2) && norm2.length / norm1.length > 0.8) {
        return 0.95
      }
      if (norm2.includes(norm1) && norm1.length / norm2.length > 0.8) {
        return 0.95
      }
    }
    
    // Word-based matching with better scoring
    const words1 = norm1.split(' ').filter(w => w.length > 1)
    const words2 = norm2.split(' ').filter(w => w.length > 1)
    
    if (words1.length === 0 || words2.length === 0) return 0
    
    // Count exact word matches
    let exactMatches = 0
    let partialMatches = 0
    
    for (const word1 of words1) {
      let bestWordMatch = 0
      for (const word2 of words2) {
        if (word1 === word2) {
          exactMatches++
          bestWordMatch = 1
          break
        } else if (word1.length > 3 && word2.length > 3) {
          // Check if words are very similar (for typos/variations)
          if (word1.includes(word2) || word2.includes(word1)) {
            const similarity = Math.min(word1.length, word2.length) / Math.max(word1.length, word2.length)
            bestWordMatch = Math.max(bestWordMatch, similarity * 0.8)
          }
        }
      }
      if (bestWordMatch > 0 && bestWordMatch < 1) {
        partialMatches += bestWordMatch
      }
    }
    
    // Calculate final score
    const totalWords = Math.max(words1.length, words2.length)
    const exactScore = exactMatches / totalWords
    const partialScore = partialMatches / totalWords
    
    // Require at least one exact word match for high scores
    if (exactMatches === 0 && totalWords > 1) {
      return Math.min(0.7, exactScore + partialScore)
    }
    
    return exactScore + partialScore * 0.5
  }

  /**
   * Check if a match is likely to be correct based on additional heuristics
   * @param {string} searchName - Original search name
   * @param {string} foundName - Found team name
   * @param {number} score - Similarity score
   * @returns {boolean} - Whether this is likely a good match
   */
  const isLikelyCorrectMatch = (searchName, foundName, score) => {
    // If exact match, always good
    if (score === 1.0) return true
    
    // In strict mode, require higher scores
    if (strictMode && score < 0.9) return false
    
    const searchWords = normalizeTeamName(searchName).split(' ').filter(w => w.length > 1)
    const foundWords = normalizeTeamName(foundName).split(' ').filter(w => w.length > 1)
    
    // For teams with unique words, require higher precision
    const hasUniqueWords = searchWords.some(word => 
      word.length > 6 || // Long words are usually unique
      ['united', 'city', 'town', 'rovers', 'wanderers'].includes(word)
    )
    
    if (hasUniqueWords && score < 0.9) {
      return false
    }
    
    // Check for common false positive patterns
    const commonNames = ['al', 'real', 'atletico', 'sporting', 'club', 'union']
    const hasOnlyCommonWords = searchWords.every(word => 
      commonNames.includes(word) || word.length <= 2
    )
    
    if (hasOnlyCommonWords && score < 0.95) {
      return false
    }
    
    return score >= similarityThreshold
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
    
    // Then try fuzzy matching with better validation
    let bestMatch = null
    let bestScore = 0
    const candidates = []
    
    // Collect all potential matches above threshold
    for (const [numericalId, name] of Object.entries(teamsData)) {
      const similarity = calculateSimilarity(teamName, name)
      if (similarity >= similarityThreshold) {
        candidates.push({
          id: numericalId,
          name,
          score: similarity
        })
      }
    }
    
    // Sort by score and apply additional validation
    candidates.sort((a, b) => b.score - a.score)
    
    for (const candidate of candidates) {
      if (isLikelyCorrectMatch(teamName, candidate.name, candidate.score)) {
        bestMatch = candidate.id
        bestScore = candidate.score
        break
      }
    }
    
    return bestMatch
  }

  /**
   * Get detailed match information for debugging
   * @param {string} teamName - The display name of the team
   * @returns {Object|null} - Match details { id, name, score, alternatives } or null
   */
  const getTeamMatchDetails = (teamName) => {
    if (!teamName) return null
    
    // First try exact match
    for (const [numericalId, name] of Object.entries(teamsData)) {
      if (name === teamName) {
        return { id: numericalId, name, score: 1.0, alternatives: [] }
      }
    }
    
    // If fuzzy matching is disabled, return null
    if (!enableFuzzyMatching) return null
    
    // Collect all potential matches
    const candidates = []
    
    for (const [numericalId, name] of Object.entries(teamsData)) {
      const similarity = calculateSimilarity(teamName, name)
      if (similarity >= Math.max(0.5, similarityThreshold - 0.2)) { // Lower threshold for debugging
        candidates.push({
          id: numericalId,
          name,
          score: similarity,
          isValid: isLikelyCorrectMatch(teamName, name, similarity)
        })
      }
    }
    
    // Sort by score
    candidates.sort((a, b) => b.score - a.score)
    
    if (candidates.length === 0) return null
    
    // Find the best valid match
    const bestValid = candidates.find(c => c.isValid)
    const alternatives = candidates.slice(0, 5).map(c => ({
      id: c.id,
      name: c.name,
      score: c.score,
      isRecommended: c.isValid
    }))
    
    if (bestValid) {
      return {
        id: bestValid.id,
        name: bestValid.name,
        score: bestValid.score,
        alternatives
      }
    }
    
    return null
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