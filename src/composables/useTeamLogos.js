import { computed, ref, reactive, readonly } from 'vue'
import teamsData from '../utils/teams_data.json' with { type: 'json' }

/**
 * Composable for handling team logos with async progressive loading
 * Maps team names to their numerical IDs and provides logo URLs
 */
export function useTeamLogos(options = {}) {
  const { 
    similarityThreshold = 0.7,
    enableFuzzyMatching = true,
    strictMode = false,
    batchSize = 10, // Process logos in batches
    batchDelay = 5   // ms delay between batches
  } = options

  // Reactive state for async processing
  const logoCache = reactive(new Map())
  const processingQueue = ref([])
  const isProcessing = ref(false)
  const processedCount = ref(0)
  const totalCount = ref(0)

  /**
   * Improved normalization with better abbreviation handling
   * @param {string} name - The team name to normalize
   * @returns {string} - Normalized team name
   */
  const normalizeTeamName = (name) => {
    if (!name) return ''
    
    let normalized = name
      .toLowerCase()
      .trim()
      // Remove common prefixes (but more selectively)
      .replace(/^(fc|cf)\s+/i, '')
      .replace(/^(ac|sc)\s+/i, '') // Add these back but carefully
      // Remove common suffixes
      .replace(/\s+(fc|cf|f\.c\.?)$/i, '')
      .replace(/\s+(ac|sc)$/i, '')
      // Handle some abbreviations
      .replace(/\s+f\.c\.?$/i, '')
      .replace(/^f\.c\.?\s+/i, '')
      // Normalize spacing
      .replace(/\s+/g, ' ')
      .trim()
    
    // Handle common abbreviations and expansions
    if (normalized === 'psg') {
      normalized = 'paris saint-germain'
    } else if (normalized === 'paris saint-germain (psg)') {
      normalized = 'paris saint-germain'
    } else if (normalized === 'man utd') {
      normalized = 'manchester united'
    } else if (normalized === 'man united') {
      normalized = 'manchester united'
    }
    
    return normalized
  }

  // Pre-process and index team data for performance
  const teamIndex = (() => {
    const exactLookup = new Map()
    const normalizedLookup = new Map()
    const wordIndex = new Map()
    
    // Pre-normalize all team names once
    for (const [id, name] of Object.entries(teamsData)) {
      // Exact lookup
      exactLookup.set(name, id)
      
      // Normalized lookup
      const normalized = normalizeTeamName(name)
      if (!normalizedLookup.has(normalized)) {
        normalizedLookup.set(normalized, [])
      }
      normalizedLookup.get(normalized).push({ id, name, normalized })
      
      // Word index for faster fuzzy search
      const words = normalized.split(' ').filter(w => w.length > 2)
      for (const word of words) {
        if (!wordIndex.has(word)) {
          wordIndex.set(word, [])
        }
        wordIndex.get(word).push({ id, name, normalized })
      }
    }
    
    return { exactLookup, normalizedLookup, wordIndex }
  })()

  /**
   * Enhanced similarity calculation with abbreviation awareness
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
    
    // Handle common abbreviation patterns
    if (isAbbreviationMatch(norm1, norm2) || isAbbreviationMatch(norm2, norm1)) {
      return 0.95
    }
    
    // Split into words
    const words1 = norm1.split(' ').filter(w => w.length > 1)
    const words2 = norm2.split(' ').filter(w => w.length > 1)
    
    if (words1.length === 0 || words2.length === 0) return 0
    
    // Check for exact substring matches with validation
    if (norm1.length > 3 && norm2.length > 3) {
      if (norm1.includes(norm2)) {
        const ratio = norm2.length / norm1.length
        if (ratio > 0.5) return 0.95
        if (ratio > 0.3) return 0.85
      }
      if (norm2.includes(norm1)) {
        const ratio = norm1.length / norm2.length
        if (ratio > 0.5) return 0.95
        if (ratio > 0.3) return 0.85
      }
    }
    
    // Word-based matching
    let exactWordMatches = 0
    let partialMatches = 0
    
    for (const word1 of words1) {
      let bestMatch = 0
      for (const word2 of words2) {
        if (word1 === word2) {
          exactWordMatches++
          bestMatch = 1
          break
        } else if (word1.length > 3 && word2.length > 3) {
          // Allow partial matching for variations
          if (word1.includes(word2) && word2.length >= 3) {
            bestMatch = Math.max(bestMatch, 0.8)
          } else if (word2.includes(word1) && word1.length >= 3) {
            bestMatch = Math.max(bestMatch, 0.8)
          }
        }
      }
      if (bestMatch > 0 && bestMatch < 1) {
        partialMatches += bestMatch
      }
    }
    
    const maxWords = Math.max(words1.length, words2.length)
    const exactScore = exactWordMatches / maxWords
    const partialScore = partialMatches / maxWords
    
    // Combine scores with reasonable weighting
    return Math.min(1.0, exactScore + partialScore * 0.5)
  }

  /**
   * Check if two strings are abbreviation matches
   * @param {string} short - Potentially shorter string
   * @param {string} long - Potentially longer string
   * @returns {boolean} - Whether short could be an abbreviation of long
   */
  const isAbbreviationMatch = (short, long) => {
    if (short.length >= long.length) return false
    
    const shortWords = short.split(' ').filter(w => w.length > 1)
    const longWords = long.split(' ').filter(w => w.length > 1)
    
    // Handle cases like "man utd" vs "manchester united"
    if (shortWords.length === 2 && longWords.length === 2) {
      const [short1, short2] = shortWords
      const [long1, long2] = longWords
      
      // Check if each short word is a prefix or abbreviation of long word
      if ((long1.startsWith(short1) || short1.startsWith(long1.substring(0, 3))) &&
          (long2.startsWith(short2) || short2.startsWith(long2.substring(0, 3)))) {
        return true
      }
    }
    
    // Handle cases like "psg" vs "paris saint-germain"
    if (shortWords.length === 1 && longWords.length >= 2) {
      const shortWord = shortWords[0]
      // Check if short word is made of first letters of long words
      if (shortWord.length === longWords.length && 
          shortWord === longWords.map(w => w[0]).join('')) {
        return true
      }
    }
    
    return false
  }

  /**
   * Balanced validation that prevents false positives without breaking good matches
   * @param {string} searchName - Original search name
   * @param {string} foundName - Found team name
   * @param {number} score - Similarity score
   * @returns {boolean} - Whether this is likely a good match
   */
  const isLikelyCorrectMatch = (searchName, foundName, score) => {
    // If exact match, always good
    if (score === 1.0) return true
    
    // If score is too low, reject
    if (score < similarityThreshold) return false
    
    // High scores are generally good
    if (score >= 0.9) return true
    
    const searchWords = normalizeTeamName(searchName).split(' ').filter(w => w.length > 1)
    const foundWords = normalizeTeamName(foundName).split(' ').filter(w => w.length > 1)
    
    // Count exact word matches
    let exactMatches = 0
    for (const searchWord of searchWords) {
      for (const foundWord of foundWords) {
        if (searchWord === foundWord) {
          exactMatches++
          break
        }
      }
    }
    
    // For multi-word search terms, be more careful
    if (searchWords.length >= 2) {
      // If no exact word matches, be suspicious
      if (exactMatches === 0) {
        // But allow if it's a very high similarity score
        if (score < 0.85) return false
      }
      
      // For specific problematic cases
      const searchNorm = normalizeTeamName(searchName)
      
      // "Manchester United" should not match random teams with just "united"
      if (searchNorm.includes('manchester') && searchNorm.includes('united')) {
        if (!foundWords.includes('manchester') && score < 0.9) {
          return false
        }
      }
      
      // "Tottenham Hotspur" should match something with "tottenham" or "hotspur"
      if (searchNorm.includes('tottenham') && searchNorm.includes('hotspur')) {
        if (!foundWords.includes('tottenham') && !foundWords.includes('hotspur') && score < 0.9) {
          return false
        }
      }
      
      // "Nottingham Forest" should match something with "nottingham" or reasonable forest match
      if (searchNorm.includes('nottingham') && searchNorm.includes('forest')) {
        if (!foundWords.includes('nottingham') && !foundWords.some(w => w.includes('forest') || w.includes('nottm')) && score < 0.85) {
          return false
        }
      }
    }
    
    // In strict mode, apply additional validation
    if (strictMode) {
      const veryCommonWords = ['al', 'real', 'united', 'city', 'fc', 'cf']
      const hasOnlyVeryCommon = searchWords.every(word => 
        veryCommonWords.includes(word) || word.length <= 2
      )
      
      if (hasOnlyVeryCommon && score < 0.9) {
        return false
      }
    }
    
    // Check for reasonable length similarity (but be lenient)
    const lengthRatio = Math.min(searchName.length, foundName.length) / Math.max(searchName.length, foundName.length)
    if (lengthRatio < 0.25 && score < 0.85) {
      return false
    }
    
    return true
  }

  /**
   * OPTIMIZED: Get candidates efficiently using indexes
   * @param {string} teamName - The display name of the team
   * @returns {Array} - Array of candidate teams to check
   */
  const getCandidates = (teamName) => {
    const normalized = normalizeTeamName(teamName)
    const candidates = new Set()
    
    // 1. Check normalized exact matches (very fast)
    const exactMatches = teamIndex.normalizedLookup.get(normalized)
    if (exactMatches) {
      exactMatches.forEach(team => candidates.add(team))
    }
    
    // 2. If no exact matches, check word index (much faster than full scan)
    if (candidates.size === 0 && enableFuzzyMatching) {
      const words = normalized.split(' ').filter(w => w.length > 2)
      
      for (const word of words) {
        const wordMatches = teamIndex.wordIndex.get(word)
        if (wordMatches) {
          wordMatches.forEach(team => candidates.add(team))
        }
        
        // Also check partial word matches for the longest words
        if (word.length > 4) {
          for (const [indexWord, teams] of teamIndex.wordIndex) {
            if (indexWord.includes(word) || word.includes(indexWord)) {
              teams.forEach(team => candidates.add(team))
            }
          }
        }
      }
      
      // If still no candidates, fallback to full search (rare)
      if (candidates.size === 0) {
        for (const teams of teamIndex.normalizedLookup.values()) {
          teams.forEach(team => candidates.add(team))
        }
      }
    }
    
    return Array.from(candidates)
  }

  /**
   * SYNC: Get the numerical team ID for a given team name (blocking)
   * @param {string} teamName - The display name of the team
   * @returns {string|null} - The numerical team ID or null if not found
   */
  const getTeamId = (teamName) => {
    if (!teamName) return null
    
    // First try exact match (very fast lookup)
    const exactId = teamIndex.exactLookup.get(teamName)
    if (exactId) return exactId
    
    // If fuzzy matching is disabled, return null
    if (!enableFuzzyMatching) return null
    
    // Get candidates efficiently (much smaller set than all 53k teams)
    const candidates = getCandidates(teamName)
    
    let bestMatch = null
    let bestScore = 0
    
    // Only process the candidate teams (much faster!)
    for (const candidate of candidates) {
      const similarity = calculateSimilarity(teamName, candidate.name)
      if (similarity >= similarityThreshold && 
          similarity > bestScore && 
          isLikelyCorrectMatch(teamName, candidate.name, similarity)) {
        bestScore = similarity
        bestMatch = candidate.id
      }
    }
    
    return bestMatch
  }

  /**
   * ASYNC: Get the numerical team ID for a given team name (non-blocking)
   * @param {string} teamName - The display name of the team
   * @returns {Promise<string|null>} - Promise resolving to team ID or null
   */
  const getTeamIdAsync = async (teamName) => {
    if (!teamName) return null
    
    // Check cache first
    if (logoCache.has(teamName)) {
      return logoCache.get(teamName)
    }
    
    // For exact matches, process immediately
    const exactId = teamIndex.exactLookup.get(teamName)
    if (exactId) {
      logoCache.set(teamName, exactId)
      return exactId
    }
    
    // For fuzzy matches, yield to event loop occasionally
    return new Promise(resolve => {
      setTimeout(() => {
        const result = getTeamId(teamName)
        logoCache.set(teamName, result)
        resolve(result)
      }, 0)
    })
  }

  /**
   * ASYNC: Process multiple team logos in batches
   * @param {string[]} teamNames - Array of team names to process
   * @param {function} onProgress - Callback for progress updates
   * @returns {Promise<Map>} - Promise resolving to map of teamName -> teamId
   */
  const processTeamLogos = async (teamNames, onProgress = null) => {
    const uniqueTeams = [...new Set(teamNames.filter(Boolean))]
    const results = new Map()
    
    totalCount.value = uniqueTeams.length
    processedCount.value = 0
    isProcessing.value = true
    
    try {
      // Process in batches to avoid blocking the UI
      for (let i = 0; i < uniqueTeams.length; i += batchSize) {
        const batch = uniqueTeams.slice(i, i + batchSize)
        
        // Process batch
        const batchPromises = batch.map(async teamName => {
          const teamId = await getTeamIdAsync(teamName)
          results.set(teamName, teamId)
          processedCount.value++
          
          if (onProgress) {
            onProgress({
              teamName,
              teamId,
              processed: processedCount.value,
              total: totalCount.value,
              progress: processedCount.value / totalCount.value
            })
          }
          
          return { teamName, teamId }
        })
        
        await Promise.all(batchPromises)
        
        // Small delay between batches to keep UI responsive
        if (i + batchSize < uniqueTeams.length) {
          await new Promise(resolve => setTimeout(resolve, batchDelay))
        }
      }
    } finally {
      isProcessing.value = false
    }
    
    return results
  }

  /**
   * Get the logo URL for a team (sync)
   * @param {string} teamName - The display name of the team
   * @returns {string|null} - The logo URL or null if no team ID found
   */
  const getTeamLogoUrl = (teamName) => {
    const teamId = getTeamId(teamName)
    if (!teamId) return null
    
    return `/api/logos?teamId=${encodeURIComponent(teamId)}`
  }

  /**
   * Get the logo URL for a team (async)
   * @param {string} teamName - The display name of the team
   * @returns {Promise<string|null>} - Promise resolving to logo URL or null
   */
  const getTeamLogoUrlAsync = async (teamName) => {
    const teamId = await getTeamIdAsync(teamName)
    if (!teamId) return null
    
    return `/api/logos?teamId=${encodeURIComponent(teamId)}`
  }

  /**
   * Reactive logo URL getter - returns cached value and updates when available
   * @param {string} teamName - The display name of the team
   * @returns {import('vue').ComputedRef<string|null>} - Reactive logo URL
   */
  const getReactiveLogoUrl = (teamName) => {
    return computed(() => {
      const teamId = logoCache.get(teamName)
      if (!teamId) return null
      return `/api/logos?teamId=${encodeURIComponent(teamId)}`
    })
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
   * Check if a team has a logo available (sync)
   * @param {string} teamName - The display name of the team
   * @returns {boolean} - True if team ID exists (logo may exist)
   */
  const hasTeamLogo = (teamName) => {
    return getTeamId(teamName) !== null
  }

  /**
   * Get detailed match information for debugging
   * @param {string} teamName - The display name of the team
   * @returns {Object|null} - Match details or null
   */
  const getTeamMatchDetails = (teamName) => {
    if (!teamName) return null
    
    // First try exact match
    const exactId = teamIndex.exactLookup.get(teamName)
    if (exactId) {
      return { id: exactId, name: teamName, score: 1.0, alternatives: [] }
    }
    
    // If fuzzy matching is disabled, return null
    if (!enableFuzzyMatching) return null
    
    // Get candidates efficiently
    const candidates = getCandidates(teamName)
    const scoredCandidates = []
    
    // Score only the candidates
    for (const candidate of candidates) {
      const similarity = calculateSimilarity(teamName, candidate.name)
      if (similarity >= Math.max(0.4, similarityThreshold - 0.3)) {
        scoredCandidates.push({
          id: candidate.id,
          name: candidate.name,
          score: similarity,
          isValid: isLikelyCorrectMatch(teamName, candidate.name, similarity)
        })
      }
    }
    
    // Sort by score
    scoredCandidates.sort((a, b) => b.score - a.score)
    
    if (scoredCandidates.length === 0) return null
    
    // Find the best valid match
    const bestValid = scoredCandidates.find(c => c.isValid)
    const alternatives = scoredCandidates.slice(0, 5).map(c => ({
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

  return {
    // Sync methods (for backward compatibility)
    getTeamId,
    getTeamLogoUrl,
    createTeamLogoUrl,
    hasTeamLogo,
    getTeamMatchDetails,
    
    // Async methods (for progressive loading)
    getTeamIdAsync,
    getTeamLogoUrlAsync,
    processTeamLogos,
    getReactiveLogoUrl,
    
    // Reactive state for async processing
    logoCache: readonly(logoCache),
    isProcessing: readonly(isProcessing),
    processedCount: readonly(processedCount),
    totalCount: readonly(totalCount),
    
    // Utility methods
    normalizeTeamName,
    calculateSimilarity
  }
} 