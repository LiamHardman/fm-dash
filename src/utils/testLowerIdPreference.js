/**
 * Test script to demonstrate the lower ID preference functionality
 * in the team name to ID mapping algorithm.
 * 
 * This script shows how teams with lower IDs are preferred when
 * similarity scores are close (within the configured threshold).
 */

import { useTeamLogos } from '../composables/useTeamLogos.js'

/**
 * Test the lower ID preference with different threshold settings
 */
function testLowerIdPreference() {
  console.log('Testing Lower ID Preference Functionality')
  console.log('=' .repeat(50))
  
  // Test with default threshold (0.05)
  console.log('\n1. Testing with default threshold (0.05):')
  const defaultConfig = useTeamLogos()
  
  // Test with more aggressive threshold (0.1)
  console.log('\n2. Testing with aggressive threshold (0.1):')
  const aggressiveConfig = useTeamLogos({ 
    lowerIdPreferenceThreshold: 0.1 
  })
  
  // Test with disabled preference (0.0)
  console.log('\n3. Testing with disabled preference (0.0):')
  const disabledConfig = useTeamLogos({ 
    lowerIdPreferenceThreshold: 0.0 
  })
  
  // Test cases that might have multiple similar matches
  const testCases = [
    'Real Madrid',
    'Barcelona',
    'Manchester United',
    'Arsenal',
    'Valencia',
    'Atletico Madrid'
  ]
  
  testCases.forEach(teamName => {
    console.log(`\nTesting: "${teamName}"`)
    
    const defaultResult = defaultConfig.getTeamId(teamName)
    const aggressiveResult = aggressiveConfig.getTeamId(teamName)
    const disabledResult = disabledConfig.getTeamId(teamName)
    
    console.log(`  Default (0.05):    ID ${defaultResult}`)
    console.log(`  Aggressive (0.1):  ID ${aggressiveResult}`)
    console.log(`  Disabled (0.0):    ID ${disabledResult}`)
    
    // Show match details for the default configuration
    const matchDetails = defaultConfig.getTeamMatchDetails(teamName)
    if (matchDetails) {
      console.log(`  Match: "${matchDetails.name}" (confidence: ${(matchDetails.confidence * 100).toFixed(1)}%)`)
    }
  })
}

/**
 * Demonstrate the impact of different threshold values
 */
function demonstrateThresholdImpact() {
  console.log('\n\nDemonstrating Threshold Impact')
  console.log('=' .repeat(50))
  
  const testTeam = 'Valencia' // Likely to have multiple similar matches
  const thresholds = [0.0, 0.02, 0.05, 0.1, 0.15, 0.2]
  
  console.log(`\nTesting with team: "${testTeam}"`)
  console.log('Threshold | Selected ID | Notes')
  console.log('-'.repeat(40))
  
  thresholds.forEach(threshold => {
    const config = useTeamLogos({ lowerIdPreferenceThreshold: threshold })
    const teamId = config.getTeamId(testTeam)
    const notes = threshold === 0.0 ? 'No lower ID preference' : 
                  threshold >= 0.1 ? 'Aggressive preference' : 'Normal preference'
    
    console.log(`${threshold.toFixed(2).padEnd(9)} | ${(teamId || 'None').padEnd(11)} | ${notes}`)
  })
}

/**
 * Show statistics about team IDs to understand the distribution
 */
function showTeamIdStatistics() {
  console.log('\n\nTeam ID Statistics')
  console.log('=' .repeat(50))
  
  // This would require access to the teams data
  // For now, just show what information would be useful
  console.log('Useful statistics to analyze:')
  console.log('- Total number of teams in database')
  console.log('- Range of team IDs (min to max)')
  console.log('- Distribution of ID lengths')
  console.log('- Teams with similar names and different IDs')
  console.log('- Impact of lower ID preference on match accuracy')
}

/**
 * Example usage of the new configuration option
 */
function usageExamples() {
  console.log('\n\nUsage Examples')
  console.log('=' .repeat(50))
  
  console.log(`
// Default configuration (moderate lower ID preference)
const teamLogos = useTeamLogos()

// Aggressive lower ID preference (larger threshold)
const aggressiveConfig = useTeamLogos({
  lowerIdPreferenceThreshold: 0.15  // Prefer lower IDs when scores are within 0.15
})

// Minimal lower ID preference (smaller threshold)
const minimalConfig = useTeamLogos({
  lowerIdPreferenceThreshold: 0.02  // Only prefer lower IDs when scores are very close
})

// Disabled lower ID preference (original behavior)
const originalBehavior = useTeamLogos({
  lowerIdPreferenceThreshold: 0.0   // Always pick highest scoring match regardless of ID
})

// Combined with other options
const customConfig = useTeamLogos({
  similarityThreshold: 0.8,           // Higher similarity requirement
  lowerIdPreferenceThreshold: 0.1,   // Moderate lower ID preference
  enableFuzzyMatching: true           // Keep fuzzy matching enabled
})
`)
}

// Run the tests if this file is executed directly
if (typeof window === 'undefined') {
  // Node.js environment
  testLowerIdPreference()
  demonstrateThresholdImpact()
  showTeamIdStatistics()
  usageExamples()
} else {
  // Browser environment - export functions for manual testing
  window.testLowerIdPreference = {
    testLowerIdPreference,
    demonstrateThresholdImpact,
    showTeamIdStatistics,
    usageExamples
  }
  
  console.log('Lower ID preference test functions available in window.testLowerIdPreference')
}

export {
  testLowerIdPreference,
  demonstrateThresholdImpact,
  showTeamIdStatistics,
  usageExamples
} 