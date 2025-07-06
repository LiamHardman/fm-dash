/**
 * Examples and tests for team name fuzzy matching
 * This file demonstrates how the fuzzy matching handles various team name variations
 */

import { useTeamLogos } from '../composables/useTeamLogos.js'
import { findBestTeamMatch, normalizeTeamName, calculateTeamSimilarity } from './fuzzyTeamMatcher.js'

// Example test cases that your system should now handle
const testCases = [
  // Your mentioned examples
  { input: 'Valencia', expected: 'Valencia C.F' },
  { input: 'Valencia C.F', expected: 'Valencia C.F' },
  { input: 'FC Nantes', expected: 'Nantes' },
  { input: 'Nantes', expected: 'Nantes' },
  
  // Common variations
  { input: 'Real Madrid', expected: 'Real Madrid CF' },
  { input: 'Real Madrid CF', expected: 'Real Madrid CF' },
  { input: 'Barcelona', expected: 'FC Barcelona' },
  { input: 'FC Barcelona', expected: 'FC Barcelona' },
  { input: 'Man United', expected: 'Manchester United' },
  { input: 'Manchester Utd', expected: 'Manchester United' },
  { input: 'AC Milan', expected: 'Milan' },
  { input: 'Milan', expected: 'Milan' },
  
  // Edge cases
  { input: 'Liverpool F.C.', expected: 'Liverpool' },
  { input: 'Chelsea FC', expected: 'Chelsea' },
  { input: 'Arsenal F.C', expected: 'Arsenal' },
  { input: 'Tottenham', expected: 'Tottenham Hotspur' },
  { input: 'Spurs', expected: 'Tottenham Hotspur' }, // This might not work without additional aliases
  
  // International examples
  { input: 'Bayern Munich', expected: 'Bayern München' },
  { input: 'Bayern München', expected: 'Bayern München' },
  { input: 'Borussia Dortmund', expected: 'Borussia Dortmund' },
  { input: 'BVB', expected: 'Borussia Dortmund' }, // Might need aliases
]

/**
 * Test the fuzzy matching system
 * @param {Object} teamsData - The teams data from JSON
 */
export function testTeamMatching(teamsData) {
  const { getTeamId, getTeamMatchDetails } = useTeamLogos({ 
    useAdvancedMatching: true,
    similarityThreshold: 0.7 
  })
  
  console.log('=== Team Fuzzy Matching Test Results ===\n')
  
  const results = {
    total: testCases.length,
    successful: 0,
    failed: 0,
    details: []
  }
  
  testCases.forEach(({ input, expected }, index) => {
    const matchDetails = getTeamMatchDetails(input)
    const teamId = getTeamId(input)
    
    const result = {
      test: index + 1,
      input,
      expected,
      found: matchDetails?.name || 'No match',
      teamId: teamId || 'None',
      score: matchDetails?.score || 0,
      success: false
    }
    
    // Check if we found a reasonable match
    if (matchDetails && matchDetails.score >= 0.7) {
      result.success = true
      results.successful++
    } else {
      results.failed++
    }
    
    results.details.push(result)
    
    // Log individual result
    const status = result.success ? '✅' : '❌'
    console.log(`${status} Test ${result.test}: "${input}"`)
    console.log(`   Expected: ${expected}`)
    console.log(`   Found: ${result.found} (Score: ${result.score.toFixed(3)})`)
    console.log(`   Team ID: ${result.teamId}\n`)
  })
  
  console.log('=== Summary ===')
  console.log(`Total tests: ${results.total}`)
  console.log(`Successful: ${results.successful}`)
  console.log(`Failed: ${results.failed}`)
  console.log(`Success rate: ${((results.successful / results.total) * 100).toFixed(1)}%`)
  
  return results
}

/**
 * Demonstrate normalization process
 */
export function demonstrateNormalization() {
  const examples = [
    'FC Barcelona',
    'Real Madrid C.F.',
    'Manchester United F.C.',
    'AC Milan',
    'Chelsea FC',
    'Liverpool F.C.',
    'Arsenal F.C.',
    'Tottenham Hotspur F.C.'
  ]
  
  console.log('=== Team Name Normalization Examples ===\n')
  
  examples.forEach(name => {
    const normalized = normalizeTeamName(name)
    console.log(`"${name}" → "${normalized}"`)
  })
}

/**
 * Show similarity calculations between team names
 */
export function demonstrateSimilarity() {
  const pairs = [
    ['Valencia', 'Valencia C.F'],
    ['FC Nantes', 'Nantes'],
    ['Real Madrid', 'Real Madrid CF'],
    ['Barcelona', 'FC Barcelona'],
    ['Man United', 'Manchester United'],
    ['Arsenal', 'Arsenal F.C.'],
    ['Liverpool', 'Liverpool F.C.'],
    ['Chelsea', 'Chelsea FC']
  ]
  
  console.log('=== Similarity Calculations ===\n')
  
  pairs.forEach(([name1, name2]) => {
    const similarity = calculateTeamSimilarity(name1, name2)
    console.log(`"${name1}" vs "${name2}": ${similarity.toFixed(3)}`)
  })
}

/**
 * Run all demonstrations
 * @param {Object} teamsData - Teams data from JSON
 */
export function runAllDemos(teamsData) {
  demonstrateNormalization()
  console.log('\n')
  demonstrateSimilarity()
  console.log('\n')
  return testTeamMatching(teamsData)
} 