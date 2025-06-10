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

  // --- NEW: A set of common, low-information words to de-emphasize in scoring ---
  const stopWords = new Set(['town', 'city', 'united', 'fc', 'cf', 'ac', 'sc', 'f.c.', 'al', 'real']);

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
      .replace(/^(fc|cf)\s+/i, '')
      .replace(/^(ac|sc)\s+/i, '')
      .replace(/\s+(fc|cf|f\.c\.?)$/i, '')
      .replace(/\s+(ac|sc)$/i, '')
      .replace(/\s+f\.c\.?$/i, '')
      .replace(/^f\.c\.?\s+/i, '')
      .replace(/\s+/g, ' ')
      .trim()
    
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
    
    for (const [id, name] of Object.entries(teamsData)) {
      exactLookup.set(name, id)
      
      const normalized = normalizeTeamName(name)
      if (!normalizedLookup.has(normalized)) {
        normalizedLookup.set(normalized, [])
      }
      normalizedLookup.get(normalized).push({ id, name, normalized })
      
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
   * Enhanced similarity calculation with stop-word weighting.
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @returns {number} - Similarity score (0-1)
   */
  const calculateSimilarity = (str1, str2) => {
    if (!str1 || !str2) return 0
    
    const norm1 = normalizeTeamName(str1)
    const norm2 = normalizeTeamName(str2)
    
    if (norm1 === norm2) return 1.0
    
    if (isAbbreviationMatch(norm1, norm2) || isAbbreviationMatch(norm2, norm1)) {
      return 0.95
    }
    
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
    
    const words1 = norm1.split(' ').filter(w => w.length > 1)
    const words2 = norm2.split(' ').filter(w => w.length > 1)
    
    if (words1.length === 0 || words2.length === 0) return 0

    // --- MODIFIED: Weighted word-based matching ---
    let scoreNumerator = 0;
    let scoreDenominator = 0;

    const uniqueWords1 = [...new Set(words1)]; // Use unique words from source for scoring basis

    for (const word1 of uniqueWords1) {
        // Assign a weight to the word. Stop words are worth much less.
        const weight = stopWords.has(word1) ? 0.1 : 1.0; 
        scoreDenominator += weight;

        let bestMatch = 0;
        if (words2.includes(word1)) {
            bestMatch = 1;
        }
        
        if (bestMatch > 0) {
            scoreNumerator += weight * bestMatch; // Add the weighted score
        }
    }

    if (scoreDenominator === 0) return 0;
    
    // The final score is the sum of weighted matches divided by the total possible weight.
    const wordBasedScore = scoreNumerator / scoreDenominator;
    
    return wordBasedScore;
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
    
    if (shortWords.length === 2 && longWords.length === 2) {
      const [short1, short2] = shortWords
      const [long1, long2] = longWords
      
      if ((long1.startsWith(short1) || short1.startsWith(long1.substring(0, 3))) &&
          (long2.startsWith(short2) || short2.startsWith(long2.substring(0, 3)))) {
        return true
      }
    }
    
    if (shortWords.length === 1 && longWords.length >= 2) {
      const shortWord = shortWords[0]
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
    if (score === 1.0) return true
    if (score < similarityThreshold) return false
    if (score >= 0.9) return true
    
    const searchWords = normalizeTeamName(searchName).split(' ').filter(w => w.length > 1)
    const foundWords = normalizeTeamName(foundName).split(' ').filter(w => w.length > 1)
    
    let exactMatches = 0
    for (const searchWord of searchWords) {
      if (foundWords.includes(searchWord)) {
        exactMatches++
      }
    }
    
    if (searchWords.length >= 2) {
      if (exactMatches === 0 && score < 0.85) return false
      
      const searchNorm = normalizeTeamName(searchName)
      
      if (searchNorm.includes('manchester') && searchNorm.includes('united')) {
        if (!foundWords.includes('manchester') && score < 0.9) {
          return false
        }
      }
      
      if (searchNorm.includes('tottenham') && searchNorm.includes('hotspur')) {
        if (!foundWords.includes('tottenham') && !foundWords.includes('hotspur') && score < 0.9) {
          return false
        }
      }
      
      if (searchNorm.includes('nottingham') && searchNorm.includes('forest')) {
        if (!foundWords.includes('nottingham') && !foundWords.some(w => w.includes('forest') || w.includes('nottm')) && score < 0.85) {
          return false
        }
      }
    }
    
    if (strictMode) {
      const hasOnlyVeryCommon = searchWords.every(word => 
        stopWords.has(word) || word.length <= 2
      )
      
      if (hasOnlyVeryCommon && score < 0.9) {
        return false
      }
    }
    
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
    
    const exactMatches = teamIndex.normalizedLookup.get(normalized)
    if (exactMatches) {
      exactMatches.forEach(team => candidates.add(team))
    }
    
    if (candidates.size === 0 && enableFuzzyMatching) {
      const words = normalized.split(' ').filter(w => w.length > 2)
      
      for (const word of words) {
        const wordMatches = teamIndex.wordIndex.get(word)
        if (wordMatches) {
          wordMatches.forEach(team => candidates.add(team))
        }
        
        if (word.length > 4) {
          for (const [indexWord, teams] of teamIndex.wordIndex) {
            if (indexWord.includes(word) || word.includes(indexWord)) {
              teams.forEach(team => candidates.add(team))
            }
          }
        }
      }
      
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
    
    const exactId = teamIndex.exactLookup.get(teamName)
    if (exactId) return exactId
    
    if (!enableFuzzyMatching) return null
    
    const candidates = getCandidates(teamName)
    
    let bestMatch = null
    let bestScore = 0
    
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
    
    if (logoCache.has(teamName)) {
      return logoCache.get(teamName)
    }
    
    const exactId = teamIndex.exactLookup.get(teamName)
    if (exactId) {
      logoCache.set(teamName, exactId)
      return exactId
    }
    
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
      for (let i = 0; i < uniqueTeams.length; i += batchSize) {
        const batch = uniqueTeams.slice(i, i + batchSize)
        
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
      if (similarity >= Math.max(0.4, similarityThreshold - 0.3)) {
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
    calculateSimilarity
  }
}
