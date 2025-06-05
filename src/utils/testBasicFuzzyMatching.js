/**
 * Enhanced test for improved fuzzy team matching
 * Tests the core examples mentioned by the user including problematic cases
 */

import { useTeamLogos } from '../composables/useTeamLogos.js'

// Your specific examples that should now work
const testCases = [
  // Original examples
  { input: 'Valencia', shouldMatch: 'Valencia C.F' },
  { input: 'Valencia C.F', shouldMatch: 'Valencia C.F' },
  { input: 'FC Nantes', shouldMatch: 'Nantes' },
  { input: 'Nantes', shouldMatch: 'Nantes' },
  
  // User's problematic examples
  { input: 'Al-Ittihad', shouldMatch: 'Al-Ittihad', note: 'Multiple teams possible' },
  { input: 'Tottenham Hotspur', shouldMatch: 'Tottenham Hotspur', note: 'Should not match just "Tottenham"' },
  { input: 'Basaksehir FK', shouldMatch: 'Basaksehir FK', note: 'Turkish team' },
  { input: 'AS Monaco', shouldMatch: 'AS Monaco', note: 'Should not match wrong Monaco' },
  { input: 'Nottingham Forest', shouldMatch: 'Nottingham Forest', note: 'Should not match just "Nottingham"' },
  { input: 'Angers SCO', shouldMatch: 'Angers SCO', note: 'Not matching at all previously' },
  
  // Additional test cases
  { input: 'Real Madrid', shouldMatch: 'Real Madrid' },
  { input: 'Barcelona', shouldMatch: 'Barcelona' },
  { input: 'Arsenal F.C.', shouldMatch: 'Arsenal' },
  { input: 'Liverpool FC', shouldMatch: 'Liverpool' }
]

/**
 * Test the enhanced fuzzy matching
 */
export function testEnhancedFuzzyMatching() {
  console.log('🧪 Testing Enhanced Fuzzy Team Matching\n')
  
  const { getTeamId, getTeamMatchDetails, normalizeTeamName } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.85,
    strictMode: false
  })
  
  let successCount = 0
  let totalTests = testCases.length
  
  testCases.forEach(({ input, shouldMatch, note }, index) => {
    console.log(`Test ${index + 1}: "${input}"${note ? ` (${note})` : ''}`)
    
    // Show normalization
    const normalized = normalizeTeamName(input)
    console.log(`  Normalized: "${normalized}"`)
    
    // Get match details
    const matchDetails = getTeamMatchDetails(input)
    const teamId = getTeamId(input)
    
    if (matchDetails) {
      const success = matchDetails.score >= 0.85
      const status = success ? '✅' : '❌'
      
      console.log(`  ${status} Found: "${matchDetails.name}" (Score: ${matchDetails.score.toFixed(3)})`)
      console.log(`  Team ID: ${teamId}`)
      
      // Show alternatives for debugging
      if (matchDetails.alternatives && matchDetails.alternatives.length > 1) {
        console.log(`  Alternatives:`)
        matchDetails.alternatives.slice(0, 3).forEach((alt, i) => {
          const recommended = alt.isRecommended ? '⭐' : '  '
          console.log(`    ${recommended} "${alt.name}" (${alt.score.toFixed(3)})`)
        })
      }
      
      if (success) successCount++
    } else {
      console.log(`  ❌ No match found`)
    }
    
    console.log() // Empty line
  })
  
  console.log(`📊 Results: ${successCount}/${totalTests} tests passed (${((successCount/totalTests) * 100).toFixed(1)}%)`)
  
  return {
    passed: successCount,
    total: totalTests,
    successRate: (successCount / totalTests) * 100
  }
}

/**
 * Test with strict mode enabled
 */
export function testStrictMode() {
  console.log('🔒 Testing Strict Mode (Higher Precision)\n')
  
  const { getTeamId, getTeamMatchDetails } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.9,
    strictMode: true
  })
  
  const problematicCases = [
    'Al-Ittihad',
    'AS Monaco', 
    'Tottenham Hotspur',
    'Nottingham Forest'
  ]
  
  problematicCases.forEach(teamName => {
    console.log(`Testing: "${teamName}"`)
    const matchDetails = getTeamMatchDetails(teamName)
    
    if (matchDetails) {
      console.log(`  ✅ Match: "${matchDetails.name}" (Score: ${matchDetails.score.toFixed(3)})`)
      console.log(`  Alternatives:`)
      matchDetails.alternatives.slice(0, 3).forEach(alt => {
        const status = alt.isRecommended ? '⭐' : '  '
        console.log(`    ${status} "${alt.name}" (${alt.score.toFixed(3)})`)
      })
    } else {
      console.log(`  ❌ No match found`)
    }
    console.log()
  })
}

/**
 * Debug specific team matching
 */
export function debugSpecificTeams(teamNames) {
  console.log('🐛 Debug Specific Team Matching\n')
  
  const { getTeamMatchDetails, normalizeTeamName, calculateSimilarity } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.85,
    strictMode: false
  })
  
  teamNames.forEach(name => {
    console.log(`Team: "${name}"`)
    console.log(`  Normalized: "${normalizeTeamName(name)}"`)
    
    const match = getTeamMatchDetails(name)
    if (match) {
      console.log(`  ✅ Best Match: "${match.name}" (ID: ${match.id}, Score: ${match.score.toFixed(3)})`)
      
      if (match.alternatives && match.alternatives.length > 0) {
        console.log(`  All Candidates:`)
        match.alternatives.forEach((alt, i) => {
          const status = alt.isRecommended ? '⭐' : '  '
          console.log(`    ${i + 1}. ${status} "${alt.name}" (${alt.score.toFixed(3)}) ID: ${alt.id}`)
        })
      }
    } else {
      console.log(`  ❌ No match found`)
    }
    console.log()
  })
}

/**
 * Test normalization with problematic cases
 */
export function testProblematicNormalization() {
  console.log('🔧 Testing Normalization for Problematic Cases\n')
  
  const { normalizeTeamName } = useTeamLogos()
  
  const examples = [
    'Al-Ittihad',
    'AS Monaco',
    'Basaksehir FK',
    'Angers SCO',
    'Tottenham Hotspur',
    'Nottingham Forest',
    'Valencia C.F',
    'FC Nantes'
  ]
  
  examples.forEach(name => {
    const normalized = normalizeTeamName(name)
    console.log(`"${name}" → "${normalized}"`)
  })
}

// Export convenience function to test the user's specific examples
export function testUserExamples() {
  const userProblematicTeams = [
    'Al-Ittihad',
    'Tottenham Hotspur',
    'Basaksehir FK',
    'AS Monaco',
    'Nottingham Forest',
    'Angers SCO'
  ]
  
  console.log('🎯 Testing User\'s Problematic Examples\n')
  debugSpecificTeams(userProblematicTeams)
}

// Run tests if this file is executed directly
if (typeof window === 'undefined') {
  testProblematicNormalization()
  console.log('\n')
  testEnhancedFuzzyMatching()
  console.log('\n')
  testStrictMode()
} 