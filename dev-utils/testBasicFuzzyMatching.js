/**
 * Comprehensive test for balanced fuzzy team matching
 * Tests both original functionality and new improvements
 */

import { useTeamLogos } from '../composables/useTeamLogos.js'

// Test both original working cases and problematic cases
const testCases = [
  // Original examples that should keep working
  { input: 'Valencia', expectMatch: true, note: 'Should work like before' },
  { input: 'Valencia C.F', expectMatch: true, note: 'Exact match' },
  { input: 'FC Nantes', expectMatch: true, note: 'Should work like before' },
  { input: 'Nantes', expectMatch: true, note: 'Should work like before' },
  { input: 'Real Madrid', expectMatch: true, note: 'Basic case' },
  { input: 'Barcelona', expectMatch: true, note: 'Basic case' },
  { input: 'Arsenal F.C.', expectMatch: true, note: 'Should work like before' },
  { input: 'Liverpool FC', expectMatch: true, note: 'Should work like before' },
  { input: 'Manchester United', expectMatch: true, note: 'Should work like before' },
  { input: 'Chelsea', expectMatch: true, note: 'Should work like before' },
  
  // Previously problematic cases
  { input: 'Al-Ittihad', expectMatch: true, note: 'Problematic - multiple possible' },
  { input: 'Tottenham Hotspur', expectMatch: true, note: 'Problematic - false matches' },
  { input: 'Basaksehir FK', expectMatch: true, note: 'Problematic - Turkish team' },
  { input: 'AS Monaco', expectMatch: true, note: 'Problematic - wrong Monaco matches' },
  { input: 'Nottingham Forest', expectMatch: true, note: 'Problematic - partial matches' },
  { input: 'Angers SCO', expectMatch: true, note: 'Problematic - not matching at all' },
]

/**
 * Test the balanced approach
 */
function testBalancedMatching() {
  console.log('⚖️ Testing Balanced Fuzzy Team Matching\n')
  
  const { getTeamId, getTeamMatchDetails, normalizeTeamName } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.7,
    strictMode: false
  })
  
  let successCount = 0
  let originallyWorkingCount = 0
  let originallyWorkingSuccess = 0
  let problematicCount = 0
  let problematicSuccess = 0
  
  testCases.forEach(({ input, expectMatch, note }, index) => {
    const isProblematic = note.includes('Problematic')
    
    console.log(`Test ${index + 1}: "${input}" (${note})`)
    
    // Show normalization
    const normalized = normalizeTeamName(input)
    console.log(`  Normalized: "${normalized}"`)
    
    // Get match details
    const matchDetails = getTeamMatchDetails(input)
    const teamId = getTeamId(input)
    
    if (matchDetails && teamId) {
      const status = '✅'
      console.log(`  ${status} Found: "${matchDetails.name}" (Score: ${matchDetails.score.toFixed(3)})`)
      console.log(`  Team ID: ${teamId}`)
      
      successCount++
      if (isProblematic) {
        problematicSuccess++
      } else {
        originallyWorkingSuccess++
      }
    } else {
      console.log(`  ❌ No match found`)
    }
    
    if (isProblematic) {
      problematicCount++
    } else {
      originallyWorkingCount++
    }
    
    console.log() // Empty line
  })
  
  console.log(`📊 Overall Results: ${successCount}/${testCases.length} tests passed (${((successCount/testCases.length) * 100).toFixed(1)}%)`)
  console.log(`📊 Originally Working: ${originallyWorkingSuccess}/${originallyWorkingCount} (${((originallyWorkingSuccess/originallyWorkingCount) * 100).toFixed(1)}%)`)
  console.log(`📊 Previously Problematic: ${problematicSuccess}/${problematicCount} (${((problematicSuccess/problematicCount) * 100).toFixed(1)}%)`)
  
  return {
    overall: { passed: successCount, total: testCases.length },
    original: { passed: originallyWorkingSuccess, total: originallyWorkingCount },
    problematic: { passed: problematicSuccess, total: problematicCount }
  }
}

/**
 * Quick test for common teams to ensure we didn't break basic functionality
 */
function testCommonTeams() {
  console.log('🏃‍♂️ Quick Test: Common Teams\n')
  
  const { getTeamId } = useTeamLogos()
  
  const commonTeams = [
    'Real Madrid', 'Barcelona', 'Manchester United', 'Liverpool',
    'Chelsea', 'Arsenal', 'Bayern Munich', 'PSG', 'Juventus', 'Inter Milan'
  ]
  
  let found = 0
  
  commonTeams.forEach(team => {
    const id = getTeamId(team)
    const status = id ? '✅' : '❌'
    console.log(`${status} ${team}: ${id || 'No match'}`)
    if (id) found++
  })
  
  console.log(`\n📊 ${found}/${commonTeams.length} common teams found`)
  
  return found >= commonTeams.length * 0.8 // Should find at least 80%
}

/**
 * Test with strict mode to see the difference
 */
function compareStrictMode() {
  console.log('🔒 Comparing Normal vs Strict Mode\n')
  
  const normalMatcher = useTeamLogos({
    enableFuzzyMatching: true,
    strictMode: false
  })
  
  const strictMatcher = useTeamLogos({
    enableFuzzyMatching: true,
    strictMode: true
  })
  
  const testTeams = ['Al-Ittihad', 'AS Monaco', 'Real Madrid', 'Valencia']
  
  testTeams.forEach(team => {
    console.log(`Team: "${team}"`)
    
    const normalMatch = normalMatcher.getTeamMatchDetails(team)
    const strictMatch = strictMatcher.getTeamMatchDetails(team)
    
    console.log(`  Normal: ${normalMatch ? `"${normalMatch.name}" (${normalMatch.score.toFixed(3)})` : 'No match'}`)
    console.log(`  Strict: ${strictMatch ? `"${strictMatch.name}" (${strictMatch.score.toFixed(3)})` : 'No match'}`)
    console.log()
  })
}

/**
 * Test fuzzy vs exact matching to ensure we have good fallback
 */
function testFuzzyVsExact() {
  console.log('🎯 Testing Fuzzy vs Exact Matching\n')
  
  const exactOnly = useTeamLogos({ enableFuzzyMatching: false })
  const withFuzzy = useTeamLogos({ enableFuzzyMatching: true })
  
  const testCases = [
    'Valencia C.F', // Should work in both
    'Valencia',     // Should only work with fuzzy
    'FC Nantes',    // Should only work with fuzzy
    'Real Madrid'   // Should work in both
  ]
  
  testCases.forEach(team => {
    const exactId = exactOnly.getTeamId(team)
    const fuzzyId = withFuzzy.getTeamId(team)
    
    console.log(`"${team}":`)
    console.log(`  Exact only: ${exactId || 'No match'}`)
    console.log(`  With fuzzy: ${fuzzyId || 'No match'}`)
    console.log()
  })
}

// Main test function
function runAllTests() {
  console.log('🧪 Running Comprehensive Team Matching Tests\n')
  console.log('='.repeat(50))
  
  const quickTest = testCommonTeams()
  if (!quickTest) {
    console.log('❌ Basic functionality appears broken! Stopping here.')
    return
  }
  
  console.log('\n' + '='.repeat(50))
  const balancedResults = testBalancedMatching()
  
  console.log('\n' + '='.repeat(50))
  compareStrictMode()
  
  console.log('\n' + '='.repeat(50))
  testFuzzyVsExact()
  
  console.log('\n' + '='.repeat(50))
  console.log('🏁 Test Summary:')
  console.log(`Common teams working: ${quickTest ? '✅' : '❌'}`)
  console.log(`Overall success rate: ${((balancedResults.overall.passed/balancedResults.overall.total) * 100).toFixed(1)}%`)
  console.log(`Original functionality preserved: ${((balancedResults.original.passed/balancedResults.original.total) * 100).toFixed(1)}%`)
  console.log(`Problematic cases improved: ${((balancedResults.problematic.passed/balancedResults.problematic.total) * 100).toFixed(1)}%`)
}

// Run tests if this file is executed directly
if (typeof window === 'undefined') {
  runAllTests()
} 

// Export all functions
export {
  testBalancedMatching,
  testCommonTeams,
  compareStrictMode,
  testFuzzyVsExact,
  runAllTests
} 