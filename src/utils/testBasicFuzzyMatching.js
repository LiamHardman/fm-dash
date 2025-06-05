/**
 * Simple test for basic fuzzy team matching (Option 1)
 * Tests the core examples mentioned by the user
 */

import { useTeamLogos } from '../composables/useTeamLogos.js'

// Your specific examples that should now work
const testCases = [
  { input: 'Valencia', shouldMatch: 'Valencia C.F' },
  { input: 'Valencia C.F', shouldMatch: 'Valencia C.F' },
  { input: 'FC Nantes', shouldMatch: 'Nantes' },
  { input: 'Nantes', shouldMatch: 'Nantes' },
  
  // Additional test cases
  { input: 'Real Madrid', shouldMatch: 'Real Madrid' },
  { input: 'Barcelona', shouldMatch: 'Barcelona' },
  { input: 'Arsenal F.C.', shouldMatch: 'Arsenal' },
  { input: 'Liverpool FC', shouldMatch: 'Liverpool' }
]

/**
 * Test the basic fuzzy matching
 */
export function testBasicFuzzyMatching() {
  console.log('🧪 Testing Basic Fuzzy Team Matching (Option 1)\n')
  
  const { getTeamId, getTeamMatchDetails, normalizeTeamName } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.7
  })
  
  let successCount = 0
  let totalTests = testCases.length
  
  testCases.forEach(({ input, shouldMatch }, index) => {
    console.log(`Test ${index + 1}: "${input}"`)
    
    // Show normalization
    const normalized = normalizeTeamName(input)
    console.log(`  Normalized: "${normalized}"`)
    
    // Get match details
    const matchDetails = getTeamMatchDetails(input)
    const teamId = getTeamId(input)
    
    if (matchDetails) {
      const success = matchDetails.score >= 0.7
      const status = success ? '✅' : '❌'
      
      console.log(`  ${status} Found: "${matchDetails.name}" (Score: ${matchDetails.score.toFixed(3)})`)
      console.log(`  Team ID: ${teamId}`)
      
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
 * Test specific normalization examples
 */
export function testNormalization() {
  console.log('🔧 Testing Team Name Normalization\n')
  
  const { normalizeTeamName } = useTeamLogos()
  
  const examples = [
    'Valencia C.F',
    'FC Nantes', 
    'Real Madrid C.F.',
    'Manchester United F.C.',
    'AC Milan',
    'Chelsea FC',
    'Liverpool F.C.',
    'Arsenal F.C.'
  ]
  
  examples.forEach(name => {
    const normalized = normalizeTeamName(name)
    console.log(`"${name}" → "${normalized}"`)
  })
}

/**
 * Quick similarity test
 */
export function testSimilarity() {
  console.log('\n🎯 Testing Similarity Calculations\n')
  
  const { calculateSimilarity } = useTeamLogos()
  
  const pairs = [
    ['Valencia', 'Valencia C.F'],
    ['FC Nantes', 'Nantes'],
    ['Real Madrid', 'Real Madrid C.F.'],
    ['Arsenal', 'Arsenal F.C.'],
    ['Liverpool', 'Liverpool FC']
  ]
  
  pairs.forEach(([name1, name2]) => {
    const similarity = calculateSimilarity(name1, name2)
    const status = similarity >= 0.7 ? '✅' : '❌'
    console.log(`${status} "${name1}" vs "${name2}": ${similarity.toFixed(3)}`)
  })
}

// Run all tests if this file is executed directly
if (typeof window === 'undefined') {
  testNormalization()
  testSimilarity()
  testBasicFuzzyMatching()
} 