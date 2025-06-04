import { computed } from 'vue'
import teamsData from '../utils/teams_data.json'

/**
 * Composable for handling team logos
 * Maps team names to their numerical IDs and provides logo URLs
 */
export function useTeamLogos() {
  /**
   * Get the numerical team ID for a given team name
   * @param {string} teamName - The display name of the team
   * @returns {string|null} - The numerical team ID or null if not found
   */
  const getTeamId = (teamName) => {
    if (!teamName) return null
    
    // Find the numerical ID by searching for the team name in teams_data.json
    // The JSON structure is { "numericalId": "teamName", ... }
    for (const [numericalId, name] of Object.entries(teamsData)) {
      if (name === teamName) {
        return numericalId
      }
    }
    
    return null
  }

  /**
   * Get the logo URL for a team
   * @param {string} teamName - The display name of the team
   * @returns {string|null} - The logo URL or null if no team ID found
   */
  const getTeamLogoUrl = (teamName) => {
    const teamId = getTeamId(teamName)
    if (!teamId) return null
    
    return `/api/logos?teamId=${encodeURIComponent(teamId)}`
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
   * Check if a team has a logo available
   * @param {string} teamName - The display name of the team
   * @returns {boolean} - True if team ID exists (logo may exist)
   */
  const hasTeamLogo = (teamName) => {
    return getTeamId(teamName) !== null
  }

  return {
    getTeamId,
    getTeamLogoUrl,
    createTeamLogoUrl,
    hasTeamLogo
  }
} 