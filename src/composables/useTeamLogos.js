import { computed, ref, reactive, readonly } from 'vue'
import teamsData from '../utils/teams_data.json' with { type: 'json' }

/**
 * Enhanced composable for handling team logos with optimized matching
 * Maps team names to their numerical IDs and provides logo URLs
 * 
 * @param {Object} options - Configuration options
 * @param {number} options.similarityThreshold - Minimum similarity score to consider a match (default: 0.7)
 * @param {boolean} options.enableFuzzyMatching - Whether to enable fuzzy matching (default: true)
 * @param {boolean} options.strictMode - Whether to use strict matching mode (default: false)
 * @param {number} options.batchSize - Number of teams to process in each batch (default: 50)
 * @param {number} options.batchDelay - Delay between batches in milliseconds (default: 1)
 * @param {number} options.lowerIdPreferenceThreshold - How close similarity scores need to be to prefer lower IDs (default: 0.05)
 */
export function useTeamLogos(options = {}) {
  const { 
    similarityThreshold = 0.7,
    enableFuzzyMatching = true,
    strictMode = false,
    batchSize = 50, // Increased batch size for better performance
    batchDelay = 1,   // Reduced delay for faster processing
    lowerIdPreferenceThreshold = 0.05 // New option: how close scores need to be to prefer lower IDs
  } = options

  // Enhanced stop words for better matching
  const stopWords = new Set([
    'town', 'city', 'united', 'fc', 'cf', 'ac', 'sc', 'f.c.', 'al', 'real', 
    'club', 'de', 'del', 'da', 'di', 'le', 'la', 'los', 'las', 'el', 'il',
    'sporting', 'athletic', 'atletico', 'football', 'soccer', 'sport', 'sports'
  ])

  // Reactive state for async processing
  const logoCache = reactive(new Map())
  const processingQueue = ref([])
  const isProcessing = ref(false)
  const processedCount = ref(0)
  const totalCount = ref(0)

  /**
   * Enhanced normalization with comprehensive abbreviation handling
   * @param {string} name - The team name to normalize
   * @returns {string} - Normalized team name
   */
  const normalizeTeamName = (name) => {
    if (!name) return ''
    
    let normalized = name
      .toLowerCase()
      .trim()
      // Remove common prefixes and suffixes
      .replace(/^(fc|cf|ac|sc|as|ca|cs|rc|rs|cd|ud|rcd|rsd|rfc|afc|cfc|sfc)\s+/i, '')
      .replace(/\s+(fc|cf|ac|sc|as|ca|cs|rc|rs|cd|ud|rcd|rsd|rfc|afc|cfc|sfc|f\.c\.?|a\.c\.?|s\.c\.?)$/i, '')
      .replace(/\s+f\.c\.?$/i, '')
      .replace(/^f\.c\.?\s+/i, '')
      // Handle punctuation and special characters
      .replace(/[^\w\s]/g, ' ')
      .replace(/\s+/g, ' ')
      .trim()
    
    // Handle common abbreviations and variations
    const abbreviations = {
      'psg': 'paris saint germain',
      'paris saint-germain': 'paris saint germain',
      'man utd': 'manchester united',
      'man city': 'manchester city',
      'tottenham': 'tottenham hotspur',
      'spurs': 'tottenham hotspur',
      'brighton': 'brighton hove albion',
      'west ham': 'west ham united',
      'newcastle': 'newcastle united',
      'sheffield utd': 'sheffield united',
      'sheffield wed': 'sheffield wednesday',
      'nottm forest': 'nottingham forest',
      'notts forest': 'nottingham forest',
      'wolves': 'wolverhampton wanderers',
      'crystal palace': 'crystal palace',
      'qpr': 'queens park rangers',
      'barca': 'barcelona',
      'real madrid': 'real madrid',
      'atletico madrid': 'atletico madrid',
      'sevilla': 'sevilla',
      'valencia': 'valencia',
      'villarreal': 'villarreal',
      'real sociedad': 'real sociedad',
      'athletic bilbao': 'athletic bilbao',
      'real betis': 'real betis',
      'celta vigo': 'celta vigo',
      'espanyol': 'espanyol',
      'getafe': 'getafe',
      'levante': 'levante',
      'granada': 'granada',
      'cadiz': 'cadiz',
      'osasuna': 'osasuna',
      'elche': 'elche',
      'mallorca': 'mallorca',
      'alaves': 'alaves',
      'rayo vallecano': 'rayo vallecano'
    }
    
    return abbreviations[normalized] || normalized
  }

  // Enhanced pre-processing with multiple index types for optimal performance
  const teamIndex = (() => {
    const exactLookup = new Map()
    const normalizedLookup = new Map()
    const wordIndex = new Map()
    const prefixIndex = new Map()
    const suffixIndex = new Map()
    const nGramIndex = new Map()
    
    // Helper function to generate n-grams
    const generateNGrams = (text, n = 3) => {
      const ngrams = []
      for (let i = 0; i <= text.length - n; i++) {
        ngrams.push(text.substr(i, n))
      }
      return ngrams
    }
    
    for (const [id, name] of Object.entries(teamsData)) {
      // Exact lookup for perfect matches
      exactLookup.set(name, id)
      
      const normalized = normalizeTeamName(name)
      
      // Normalized lookup for direct matches
      if (!normalizedLookup.has(normalized)) {
        normalizedLookup.set(normalized, [])
      }
      normalizedLookup.get(normalized).push({ id, name, normalized })
      
      // Word index for partial matches
      const words = normalized.split(' ').filter(w => w.length > 2 && !stopWords.has(w))
      for (const word of words) {
        if (!wordIndex.has(word)) {
          wordIndex.set(word, [])
        }
        wordIndex.get(word).push({ id, name, normalized, word })
        
        // Prefix index for partial word matches
        for (let i = 3; i <= Math.min(word.length, 6); i++) {
          const prefix = word.substr(0, i)
          if (!prefixIndex.has(prefix)) {
            prefixIndex.set(prefix, [])
          }
          prefixIndex.get(prefix).push({ id, name, normalized, word })
        }
        
        // Suffix index for partial word matches
        for (let i = 3; i <= Math.min(word.length, 6); i++) {
          const suffix = word.substr(-i)
          if (!suffixIndex.has(suffix)) {
            suffixIndex.set(suffix, [])
          }
          suffixIndex.get(suffix).push({ id, name, normalized, word })
        }
      }
      
      // N-gram index for fuzzy matching
      if (normalized.length > 6) {
        const ngrams = generateNGrams(normalized.replace(/\s/g, ''), 3)
        for (const ngram of ngrams) {
          if (!nGramIndex.has(ngram)) {
            nGramIndex.set(ngram, [])
          }
          nGramIndex.get(ngram).push({ id, name, normalized })
        }
      }
    }
    
    return { 
      exactLookup, 
      normalizedLookup, 
      wordIndex, 
      prefixIndex, 
      suffixIndex, 
      nGramIndex,
      generateNGrams
    }
  })()

  /**
   * Optimized similarity calculation focused on speed and accuracy
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @returns {number} - Similarity score (0-1)
   */
  const calculateSimilarity = (str1, str2) => {
    if (!str1 || !str2) return 0
    
    const norm1 = normalizeTeamName(str1)
    const norm2 = normalizeTeamName(str2)
    
    if (norm1 === norm2) return 1.0
    
    // Quick substring check for performance
    if (norm1.length > 3 && norm2.length > 3) {
      if (norm1.includes(norm2)) {
        return Math.min(0.95, 0.8 + (norm2.length / norm1.length * 0.15))
      }
      if (norm2.includes(norm1)) {
        return Math.min(0.95, 0.8 + (norm1.length / norm2.length * 0.15))
      }
    }
    
    // Optimized word-based similarity
    const words1 = norm1.split(' ').filter(w => w.length > 1)
    const words2 = norm2.split(' ').filter(w => w.length > 1)
    
    if (words1.length === 0 || words2.length === 0) return 0
    
    // Fast word matching with early exit
    let exactMatches = 0
    let partialMatches = 0
    let totalWeight = 0
    
    for (const word1 of words1) {
      const weight = stopWords.has(word1) ? 0.2 : 1.0
      totalWeight += weight
      
      let found = false
      for (const word2 of words2) {
        if (word1 === word2) {
          exactMatches += weight
          found = true
          break
        }
        // Quick partial match check for longer words
        if (!found && word1.length > 3 && word2.length > 3) {
          if (word1.startsWith(word2) || word2.startsWith(word1)) {
            partialMatches += weight * 0.7
            found = true
          }
        }
      }
    }
    
    if (totalWeight === 0) return 0
    
    // Calculate final score with early return for perfect matches
    const score = (exactMatches + partialMatches) / totalWeight
    
    // Only use Levenshtein for close matches to avoid expensive calculation
    if (score >= 0.6) {
      const maxLen = Math.max(norm1.length, norm2.length)
      if (maxLen <= 50) { // Only for reasonable length strings
        const distance = levenshteinDistance(norm1, norm2)
        const levenshteinScore = 1 - (distance / maxLen)
        return Math.min(1.0, score * 0.8 + levenshteinScore * 0.2)
      }
    }
    
    return Math.min(1.0, score)
  }

  /**
   * Optimized Levenshtein distance calculation
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @returns {number} - Edit distance
   */
  const levenshteinDistance = (str1, str2) => {
    const matrix = []
    
    for (let i = 0; i <= str2.length; i++) {
      matrix[i] = [i]
    }
    
    for (let j = 0; j <= str1.length; j++) {
      matrix[0][j] = j
    }
    
    for (let i = 1; i <= str2.length; i++) {
      for (let j = 1; j <= str1.length; j++) {
        if (str2.charAt(i - 1) === str1.charAt(j - 1)) {
          matrix[i][j] = matrix[i - 1][j - 1]
        } else {
          matrix[i][j] = Math.min(
            matrix[i - 1][j - 1] + 1,
            matrix[i][j - 1] + 1,
            matrix[i - 1][j] + 1
          )
        }
      }
    }
    
    return matrix[str2.length][str1.length]
  }

  /**
   * Enhanced validation with more sophisticated matching logic
   * @param {string} searchName - Original search name
   * @param {string} foundName - Found team name
   * @param {number} score - Similarity score
   * @returns {boolean} - Whether this is likely a good match
   */
  const isLikelyCorrectMatch = (searchName, foundName, score) => {
    if (score === 1.0) return true
    if (score < similarityThreshold) return false
    if (score >= 0.95) return true
    
    const searchWords = normalizeTeamName(searchName).split(' ').filter(w => w.length > 1)
    const foundWords = normalizeTeamName(foundName).split(' ').filter(w => w.length > 1)
    
    // Count exact word matches
    let exactMatches = 0
    let significantMatches = 0
    
    for (const searchWord of searchWords) {
      if (foundWords.includes(searchWord)) {
        exactMatches++
        if (!stopWords.has(searchWord) && searchWord.length > 3) {
          significantMatches++
        }
      }
    }
    
    // Require at least one significant match for multi-word searches
    if (searchWords.length >= 2) {
      if (significantMatches === 0 && score < 0.85) return false
      if (exactMatches === 0 && score < 0.8) return false
    }
    
    // Enhanced specific team validations
    const searchNorm = normalizeTeamName(searchName)
    const foundNorm = normalizeTeamName(foundName)
    
    // Prevent mismatches for specific teams
    const specificValidations = [
      { search: ['manchester', 'united'], found: ['manchester'], minScore: 0.9 },
      { search: ['manchester', 'city'], found: ['manchester'], minScore: 0.9 },
      { search: ['tottenham', 'hotspur'], found: ['tottenham', 'hotspur'], minScore: 0.85 },
      { search: ['nottingham', 'forest'], found: ['nottingham', 'forest', 'nottm'], minScore: 0.85 },
      { search: ['crystal', 'palace'], found: ['crystal', 'palace'], minScore: 0.9 },
      { search: ['west', 'ham'], found: ['west', 'ham'], minScore: 0.9 },
      { search: ['brighton', 'hove'], found: ['brighton', 'hove'], minScore: 0.85 }
    ]
    
    for (const validation of specificValidations) {
      const hasSearchTerms = validation.search.every(term => searchNorm.includes(term))
      if (hasSearchTerms) {
        const hasFoundTerms = validation.found.some(term => foundNorm.includes(term))
        if (!hasFoundTerms && score < validation.minScore) {
          return false
        }
      }
    }
    
    // Length ratio check for very different lengths
    const lengthRatio = Math.min(searchName.length, foundName.length) / Math.max(searchName.length, foundName.length)
    if (lengthRatio < 0.3 && score < 0.85) {
      return false
    }
    
    // Strict mode additional checks
    if (strictMode) {
      const hasOnlyCommonWords = searchWords.every(word => 
        stopWords.has(word) || word.length <= 2
      )
      
      if (hasOnlyCommonWords && score < 0.9) {
        return false
      }
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
    const candidates = new Map() // Use Map to avoid duplicates and store priorities
    
    // Priority 1: Exact normalized matches (highest priority)
    const exactMatches = teamIndex.normalizedLookup.get(normalized)
    if (exactMatches) {
      exactMatches.forEach(team => candidates.set(team.id, { ...team, priority: 1 }))
      // If we found exact matches, sort by ID (lower IDs preferred) and return early for performance
      if (exactMatches.length > 0) {
        const sortedExactMatches = Array.from(candidates.values()).sort((a, b) => {
          const idA = parseInt(a.id, 10)
          const idB = parseInt(b.id, 10)
          return idA - idB // Lower ID wins
        })
        return sortedExactMatches
      }
    }
    
    if (!enableFuzzyMatching) {
      return []
    }
    
    // Priority 2: Significant word matches
    const words = normalized.split(' ').filter(w => w.length > 2 && !stopWords.has(w))
    const wordScores = new Map() // Track how many words match
    
    for (const word of words) {
      const wordMatches = teamIndex.wordIndex.get(word)
      if (wordMatches) {
        wordMatches.forEach(team => {
          if (!candidates.has(team.id)) {
            wordScores.set(team.id, (wordScores.get(team.id) || 0) + 1)
            candidates.set(team.id, { ...team, priority: 2 })
          }
        })
      }
    }
    
    // Limit candidates for performance - prioritize teams with more word matches, then by lower ID
    const sortedCandidates = Array.from(candidates.values())
      .sort((a, b) => {
        const scoreA = wordScores.get(a.id) || 0
        const scoreB = wordScores.get(b.id) || 0
        if (scoreA !== scoreB) return scoreB - scoreA // Higher word match count first
        if (a.priority !== b.priority) return a.priority - b.priority // Then by priority
        
        // If same score and priority, favor lower ID
        const idA = parseInt(a.id, 10)
        const idB = parseInt(b.id, 10)
        return idA - idB // Lower ID wins
      })
      .slice(0, 50) // Limit to top 50 candidates for performance
    
    return sortedCandidates
  }

  /**
   * Optimized synchronous team ID retrieval
   * @param {string} teamName - The display name of the team
   * @returns {string|null} - The numerical team ID or null if not found
   */
  const getTeamId = (teamName) => {
    if (!teamName) return null
    
    // Check cache first
    if (logoCache.has(teamName)) {
      return logoCache.get(teamName)
    }
    
    // Fast exact lookup
    const exactId = teamIndex.exactLookup.get(teamName)
    if (exactId) {
      logoCache.set(teamName, exactId)
      return exactId
    }

    // Quick normalized lookup before expensive fuzzy matching
    const normalized = normalizeTeamName(teamName)
    const normalizedCandidates = teamIndex.normalizedLookup.get(normalized)
    if (normalizedCandidates && normalizedCandidates.length === 1) {
      // Single exact normalized match - very fast path
      const result = normalizedCandidates[0].id
      logoCache.set(teamName, result)
      return result
    }
    
    if (!enableFuzzyMatching) {
      logoCache.set(teamName, null)
      return null
    }
    
    // Fuzzy matching with optimized candidate selection
    const candidates = getCandidates(teamName)
    
    let bestMatch = null
    let bestScore = 0
    const validCandidates = []
    
    // First pass: collect all candidates with their scores
    for (const candidate of candidates) {
      const similarity = calculateSimilarity(teamName, candidate.name)
      
      if (similarity >= similarityThreshold && 
          isLikelyCorrectMatch(teamName, candidate.name, similarity)) {
        validCandidates.push({
          id: candidate.id,
          score: similarity,
          candidate: candidate
        })
      }
    }
    
    if (validCandidates.length === 0) {
      logoCache.set(teamName, null)
      return null
    }
    
    // Sort candidates by score first, then by ID (lower IDs preferred)
    // This implements a "lower ID preference" feature where teams with numerically
    // lower IDs are favored when similarity scores are close (within the threshold).
    // This helps ensure consistent results and can prioritize more "canonical" teams
    // that might have been assigned lower IDs in the original data source.
    validCandidates.sort((a, b) => {
      const scoreDiff = b.score - a.score
      
      // If scores are very close (within configured threshold), favor the team with lower ID
      if (Math.abs(scoreDiff) <= lowerIdPreferenceThreshold) {
        const idA = parseInt(a.id, 10)
        const idB = parseInt(b.id, 10)
        return idA - idB // Lower ID wins
      }
      
      // Otherwise, higher score wins (original behavior)
      return scoreDiff
    })
    
    // Select the best candidate after sorting
    bestMatch = validCandidates[0].id
    bestScore = validCandidates[0].score
    
    // Cache the result
    logoCache.set(teamName, bestMatch)
    return bestMatch
  }

  /**
   * Async team ID retrieval with batching
   * @param {string} teamName - The display name of the team
   * @returns {Promise<string|null>} - Promise resolving to team ID or null
   */
  const getTeamIdAsync = async (teamName) => {
    if (!teamName) return null
    
    // Check cache first
    if (logoCache.has(teamName)) {
      return logoCache.get(teamName)
    }
    
    // Use sync method for now, but wrapped in Promise for consistency
    return new Promise(resolve => {
      setTimeout(() => {
        const result = getTeamId(teamName)
        resolve(result)
      }, 0)
    })
  }

  /**
   * Optimized batch processing for multiple team logos
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
      // Process in larger batches for better performance
      for (let i = 0; i < uniqueTeams.length; i += batchSize) {
        const batch = uniqueTeams.slice(i, i + batchSize)
        
        // Process entire batch synchronously for speed
        const batchResults = batch.map(teamName => {
          const teamId = getTeamId(teamName)
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
        
        // Small delay between batches to prevent UI blocking
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
    
    return `/api/logos?teamId=${encodeURIComponent(teamId)}&size=256`
  }

  /**
   * Get the logo URL for a team (async)
   * @param {string} teamName - The display name of the team
   * @returns {Promise<string|null>} - Promise resolving to logo URL or null
   */
  const getTeamLogoUrlAsync = async (teamName) => {
    const teamId = await getTeamIdAsync(teamName)
    if (!teamId) return null
    
    return `/api/logos?teamId=${encodeURIComponent(teamId)}&size=256`
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
      return `/api/logos?teamId=${encodeURIComponent(teamId)}&size=256`
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
    
    const exactId = teamIndex.exactLookup.get(teamName)
    if (exactId) {
      return { id: exactId, name: teamName, score: 1.0, alternatives: [] }
    }
    
    if (!enableFuzzyMatching) return null
    
    const candidates = getCandidates(teamName)
    const scoredCandidates = []
    
    for (const candidate of candidates) {
      const similarity = calculateSimilarity(teamName, candidate.name)
      if (similarity >= Math.max(0.3, similarityThreshold - 0.4)) {
        scoredCandidates.push({
          id: candidate.id,
          name: candidate.name,
          score: similarity,
          isValid: isLikelyCorrectMatch(teamName, candidate.name, similarity)
        })
      }
    }
    
    scoredCandidates.sort((a, b) => b.score - a.score)
    
    if (scoredCandidates.length === 0) return null
    
    const bestValid = scoredCandidates.find(c => c.isValid)
    const alternatives = scoredCandidates.slice(0, 10).map(c => ({
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
   * Clear the logo cache
   */
  const clearCache = () => {
    logoCache.clear()
  }

  /**
   * Get cache statistics
   * @returns {Object} - Cache statistics
   */
  const getCacheStats = () => {
    return {
      size: logoCache.size,
      totalTeams: Object.keys(teamsData).length,
      hitRate: logoCache.size > 0 ? (logoCache.size / Object.keys(teamsData).length) : 0
    }
  }

  return {
    // Sync methods
    getTeamId,
    getTeamLogoUrl,
    createTeamLogoUrl,
    hasTeamLogo,
    getTeamMatchDetails,
    
    // Async methods
    getTeamIdAsync,
    getTeamLogoUrlAsync,
    processTeamLogos,
    getReactiveLogoUrl,
    
    // Reactive state
    logoCache: readonly(logoCache),
    isProcessing: readonly(isProcessing),
    processedCount: readonly(processedCount),
    totalCount: readonly(totalCount),
    
    // Utility methods
    normalizeTeamName,
    calculateSimilarity,
    clearCache,
    getCacheStats,
    
    // Performance monitoring
    teamIndex: readonly(teamIndex)
  }
}
