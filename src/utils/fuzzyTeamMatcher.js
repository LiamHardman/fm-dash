/**
 * Advanced fuzzy team name matching utility
 * Provides multiple algorithms for matching team names with slight variations
 */

/**
 * Calculate Levenshtein distance between two strings
 * @param {string} str1 - First string
 * @param {string} str2 - Second string
 * @returns {number} - Edit distance
 */
function levenshteinDistance(str1, str2) {
  const matrix = []
  
  // Initialize matrix
  for (let i = 0; i <= str2.length; i++) {
    matrix[i] = [i]
  }
  for (let j = 0; j <= str1.length; j++) {
    matrix[0][j] = j
  }
  
  // Fill matrix
  for (let i = 1; i <= str2.length; i++) {
    for (let j = 1; j <= str1.length; j++) {
      if (str2.charAt(i - 1) === str1.charAt(j - 1)) {
        matrix[i][j] = matrix[i - 1][j - 1]
      } else {
        matrix[i][j] = Math.min(
          matrix[i - 1][j - 1] + 1, // substitution
          matrix[i][j - 1] + 1,     // insertion
          matrix[i - 1][j] + 1      // deletion
        )
      }
    }
  }
  
  return matrix[str2.length][str1.length]
}

/**
 * Normalize team name by removing common prefixes, suffixes, and formatting
 * @param {string} name - Team name to normalize
 * @returns {string} - Normalized name
 */
export function normalizeTeamName(name) {
  if (!name) return ''
  
  return name
    .toLowerCase()
    .trim()
    // Remove common prefixes (more comprehensive list)
    .replace(/^(fc|cf|ac|sc|cd|ud|real|club|athletic|atletico|athletico|deportivo|sporting|association|society|union|united)\s+/i, '')
    // Remove common suffixes
    .replace(/\s+(fc|cf|ac|sc|cd|ud|club|united|city|town|rovers|wanderers|albion|villa|county|athletic|atletico|athletico|deportivo|sporting|utd|f\.?c\.?|c\.?f\.?|s\.?c\.?|association|society|union)$/i, '')
    // Remove special characters and standardize spacing
    .replace(/[^\w\s]/g, '')
    .replace(/\s+/g, ' ')
    .trim()
}

/**
 * Calculate similarity score between two team names
 * @param {string} name1 - First team name
 * @param {string} name2 - Second team name
 * @returns {number} - Similarity score (0-1)
 */
export function calculateTeamSimilarity(name1, name2) {
  const norm1 = normalizeTeamName(name1)
  const norm2 = normalizeTeamName(name2)
  
  // Exact match after normalization
  if (norm1 === norm2) return 1.0
  
  // Check for substring matches
  if (norm1.includes(norm2) || norm2.includes(norm1)) {
    const longer = norm1.length > norm2.length ? norm1 : norm2
    const shorter = norm1.length > norm2.length ? norm2 : norm1
    return shorter.length / longer.length * 0.9 // Slight penalty for partial match
  }
  
  // Levenshtein-based similarity
  const maxLength = Math.max(norm1.length, norm2.length)
  if (maxLength === 0) return 1.0
  
  const distance = levenshteinDistance(norm1, norm2)
  const similarity = (maxLength - distance) / maxLength
  
  // Word-based matching for additional confidence
  const words1 = norm1.split(' ').filter(w => w.length > 1)
  const words2 = norm2.split(' ').filter(w => w.length > 1)
  
  if (words1.length > 0 && words2.length > 0) {
    const wordMatches = words1.filter(word1 => 
      words2.some(word2 => {
        if (word1 === word2) return true
        if (word1.length > 3 && word2.length > 3) {
          return levenshteinDistance(word1, word2) <= 1
        }
        return false
      })
    )
    
    const wordSimilarity = wordMatches.length / Math.max(words1.length, words2.length)
    
    // Combine Levenshtein and word-based similarities
    return Math.max(similarity, wordSimilarity * 0.8)
  }
  
  return similarity
}

/**
 * Find the best matching team from a teams dataset
 * @param {string} searchName - Team name to search for
 * @param {Object} teamsData - Teams dataset { id: name }
 * @param {number} threshold - Minimum similarity threshold (default: 0.7)
 * @returns {Object|null} - { id, name, score } or null
 */
export function findBestTeamMatch(searchName, teamsData, threshold = 0.7) {
  if (!searchName) return null
  
  let bestMatch = null
  let bestScore = 0
  
  // First try exact match
  for (const [id, name] of Object.entries(teamsData)) {
    if (name === searchName) {
      return { id, name, score: 1.0 }
    }
  }
  
  // Then try fuzzy matching
  for (const [id, name] of Object.entries(teamsData)) {
    const score = calculateTeamSimilarity(searchName, name)
    if (score >= threshold && score > bestScore) {
      bestScore = score
      bestMatch = { id, name, score }
    }
  }
  
  return bestMatch
}

/**
 * Create a mapping cache for frequently searched teams
 * This can help improve performance for repeated lookups
 */
export class TeamNameCache {
  constructor(teamsData, threshold = 0.7) {
    this.teamsData = teamsData
    this.threshold = threshold
    this.cache = new Map()
  }
  
  getTeamId(searchName) {
    if (!searchName) return null
    
    // Check cache first
    if (this.cache.has(searchName)) {
      return this.cache.get(searchName)
    }
    
    // Find best match
    const match = findBestTeamMatch(searchName, this.teamsData, this.threshold)
    const result = match ? match.id : null
    
    // Cache the result
    this.cache.set(searchName, result)
    
    return result
  }
  
  clearCache() {
    this.cache.clear()
  }
  
  getCacheStats() {
    return {
      size: this.cache.size,
      entries: Array.from(this.cache.entries())
    }
  }
} 