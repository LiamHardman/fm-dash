/**
 * Usage examples for the enhanced useTeamLogos with basic fuzzy matching
 * Shows how to integrate this into your existing components
 */

import { useTeamLogos } from '../composables/useTeamLogos.js'

// Example: Enhanced team logo component
export function createTeamLogoComponent() {
  const { getTeamLogoUrl, getTeamMatchDetails } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.7
  })
  
  return {
    /**
     * Get team logo with fuzzy matching support
     * @param {string} teamName - Team name (can have variations)
     * @returns {Object} - { logoUrl, matchInfo }
     */
    getTeamLogo(teamName) {
      const logoUrl = getTeamLogoUrl(teamName)
      const matchInfo = getTeamMatchDetails(teamName)
      
      return {
        logoUrl,
        matchInfo,
        hasLogo: !!logoUrl,
        confidence: matchInfo?.score || 0
      }
    }
  }
}

// Example: Batch processing team names from a dataset
export function processTeamDataset(players) {
  console.log('🔄 Processing team dataset with fuzzy matching...\n')
  
  const { getTeamId, getTeamMatchDetails } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.7
  })
  
  const results = {
    processed: 0,
    matched: 0,
    unmatched: 0,
    teams: new Map()
  }
  
  players.forEach(player => {
    if (player.club) {
      results.processed++
      
      // Try to get team ID
      const teamId = getTeamId(player.club)
      const matchDetails = getTeamMatchDetails(player.club)
      
      if (teamId) {
        results.matched++
        
        // Store team info
        if (!results.teams.has(player.club)) {
          results.teams.set(player.club, {
            originalName: player.club,
            matchedName: matchDetails?.name,
            teamId: teamId,
            confidence: matchDetails?.score || 1.0,
            playerCount: 0
          })
        }
        
        results.teams.get(player.club).playerCount++
      } else {
        results.unmatched++
        console.log(`⚠️  No match found for: "${player.club}"`)
      }
    }
  })
  
  console.log(`📊 Team Processing Results:`)
  console.log(`   Processed: ${results.processed} teams`)
  console.log(`   Matched: ${results.matched} (${((results.matched/results.processed)*100).toFixed(1)}%)`)
  console.log(`   Unmatched: ${results.unmatched}`)
  
  return results
}

// Example: Vue component integration
export const TeamLogoComponent = {
  template: `
    <div class="team-logo-container">
      <img 
        v-if="logoInfo.hasLogo" 
        :src="logoInfo.logoUrl" 
        :alt="teamName"
        :title="logoTooltip"
        class="team-logo"
      />
      <div v-else class="team-logo-placeholder">
        {{ teamNameShort }}
      </div>
    </div>
  `,
  
  props: {
    teamName: {
      type: String,
      required: true
    }
  },
  
  setup(props) {
    const { getTeamLogoUrl, getTeamMatchDetails } = useTeamLogos({
      enableFuzzyMatching: true,
      similarityThreshold: 0.7
    })
    
    const logoInfo = computed(() => {
      const logoUrl = getTeamLogoUrl(props.teamName)
      const matchDetails = getTeamMatchDetails(props.teamName)
      
      return {
        logoUrl,
        hasLogo: !!logoUrl,
        matchDetails,
        confidence: matchDetails?.score || 0
      }
    })
    
    const logoTooltip = computed(() => {
      const info = logoInfo.value
      if (!info.hasLogo) return props.teamName
      
      if (info.confidence < 1.0) {
        return `${props.teamName} (matched: ${info.matchDetails.name}, confidence: ${(info.confidence * 100).toFixed(0)}%)`
      }
      
      return props.teamName
    })
    
    const teamNameShort = computed(() => {
      return props.teamName.split(' ').map(word => word[0]).join('').toUpperCase().slice(0, 3)
    })
    
    return {
      logoInfo,
      logoTooltip,
      teamNameShort
    }
  }
}

// Example: Debugging helpers
export function debugTeamMatching(teamNames) {
  console.log('🐛 Debug Team Matching\n')
  
  const { getTeamMatchDetails, normalizeTeamName, calculateSimilarity } = useTeamLogos({
    enableFuzzyMatching: true,
    similarityThreshold: 0.7
  })
  
  teamNames.forEach(name => {
    console.log(`Team: "${name}"`)
    console.log(`  Normalized: "${normalizeTeamName(name)}"`)
    
    const match = getTeamMatchDetails(name)
    if (match) {
      console.log(`  ✅ Match: "${match.name}" (ID: ${match.id}, Score: ${match.score.toFixed(3)})`)
    } else {
      console.log(`  ❌ No match found`)
    }
    console.log()
  })
}

// Example usage patterns
export const usageExamples = {
  // Basic usage (same as before)
  basic() {
    const { getTeamId, getTeamLogoUrl } = useTeamLogos()
    const teamId = getTeamId('Valencia')  // Now works with "Valencia C.F" in data
    const logoUrl = getTeamLogoUrl('FC Nantes')  // Now works with "Nantes" in data
    return { teamId, logoUrl }
  },
  
  // Disable fuzzy matching if needed
  exactOnly() {
    const { getTeamId } = useTeamLogos({ enableFuzzyMatching: false })
    return getTeamId('Valencia')  // Only exact matches
  },
  
  // Stricter matching
  strict() {
    const { getTeamId } = useTeamLogos({ similarityThreshold: 0.9 })
    return getTeamId('Valencia')  // Requires 90% similarity
  },
  
  // Get detailed information
  detailed() {
    const { getTeamMatchDetails } = useTeamLogos()
    const match = getTeamMatchDetails('Valencia')
    return {
      found: !!match,
      teamId: match?.id,
      actualName: match?.name,
      confidence: match?.score
    }
  }
} 