/**
 * Migration Example: From Frontend Team Matching to Backend API
 *
 * This file demonstrates how to migrate from the heavy frontend team matching
 * (which loads 1.7MB teams_data.json) to the lightweight backend API approach.
 */

// OLD APPROACH (Frontend-heavy)
// ===============================
/*
import { useTeamLogos } from '../composables/useTeamLogos.js'
import teamsData from '../utils/teams_data.json' with { type: 'json' } // 1.7MB load!

// This loads the entire 53,000+ team database into memory
const { getTeamId, getTeamLogoUrl } = useTeamLogos()

// Usage
const teamId = getTeamId('Valencia') // Searches through 53k teams locally
const logoUrl = getTeamLogoUrl('Valencia')
*/

// NEW APPROACH (Backend-powered)
// ===============================
import { useTeamLogosBackend } from '../composables/useTeamLogosBackend.js'

// This only loads a small composable, no team data!
const { getTeamId, getTeamLogoUrl, getTeamMatches } = useTeamLogosBackend()

// Usage examples
async function examples() {
  // 1. Basic team ID lookup
  const _teamId = await getTeamId('Valencia')

  // 2. Get team logo URL
  const _logoUrl = await getTeamLogoUrl('Valencia')

  // 3. Get all matches for fuzzy search
  const _matches = await getTeamMatches('Valenc') // Partial name
  // Example output:
  // [
  //   { id: "433", name: "Valencia", score: 0.95 },
  //   { id: "1234", name: "Valencia CF", score: 0.92 },
  //   { id: "5678", name: "Valencia Mestalla", score: 0.85 }
  // ]

  // 4. Batch processing for multiple teams
  const { batchGetTeamLogos } = useTeamLogosBackend()
  const teamNames = ['Valencia', 'Barcelona', 'Real Madrid']
  const _logoMap = await batchGetTeamLogos(teamNames, _progress => {})
}

// MIGRATION STEPS
// ===============

/**
 * Step 1: Replace imports
 *
 * OLD:
 * import { useTeamLogos } from '../composables/useTeamLogos.js'
 *
 * NEW:
 * import { useTeamLogosBackend } from '../composables/useTeamLogosBackend.js'
 */

/**
 * Step 2: Update function calls to async/await
 *
 * OLD:
 * const teamId = getTeamId('Valencia')
 *
 * NEW:
 * const teamId = await getTeamId('Valencia')
 */

/**
 * Step 3: Handle loading states (optional but recommended)
 *
 * const { getTeamId, isLoading, lastError } = useTeamLogosBackend()
 *
 * if (isLoading.value) {
 *   console.log('Loading team data...')
 * }
 *
 * if (lastError.value) {
 *   console.error('Error:', lastError.value)
 * }
 */

/**
 * Step 4: Update Vue components
 *
 * OLD:
 * <template>
 *   <img :src="getTeamLogoUrl(team.name)" :alt="team.name" />
 * </template>
 *
 * <script setup>
 * import { useTeamLogos } from '@/composables/useTeamLogos'
 * const { getTeamLogoUrl } = useTeamLogos()
 * </script>
 *
 * NEW:
 * <template>
 *   <img :src="logoUrl" :alt="team.name" v-if="logoUrl" />
 *   <div v-else-if="isLoadingLogo">Loading...</div>
 * </template>
 *
 * <script setup>
 * import { ref, watchEffect } from 'vue'
 * import { useTeamLogosBackend } from '@/composables/useTeamLogosBackend'
 *
 * const props = defineProps(['team'])
 * const { getTeamLogoUrl } = useTeamLogosBackend()
 *
 * const logoUrl = ref(null)
 * const isLoadingLogo = ref(false)
 *
 * watchEffect(async () => {
 *   if (props.team?.name) {
 *     isLoadingLogo.value = true
 *     logoUrl.value = await getTeamLogoUrl(props.team.name)
 *     isLoadingLogo.value = false
 *   }
 * })
 * </script>
 */

// PERFORMANCE BENEFITS
// ====================

/**
 * Old approach:
 * - Initial page load: +1.7MB JavaScript bundle
 * - Memory usage: ~53k team objects in memory
 * - Search time: O(n) where n = 53,000 teams
 * - Network: One large JSON download
 *
 * New approach:
 * - Initial page load: ~5KB additional JavaScript
 * - Memory usage: Small cache of recently searched teams
 * - Search time: Fast backend search with indexing
 * - Network: Small API requests only when needed
 * - Caching: Intelligent caching prevents duplicate requests
 */

// API ENDPOINT USAGE
// ==================

/**
 * The backend API endpoint: GET /api/team-match?name={teamName}
 *
 * Example request:
 * GET /api/team-match?name=Valencia
 *
 * Example response:
 * [
 *   {
 *     "id": "433",
 *     "name": "Valencia",
 *     "score": 1.0
 *   },
 *   {
 *     "id": "1234",
 *     "name": "Valencia CF",
 *     "score": 0.95
 *   }
 * ]
 *
 * The response is sorted by score (best matches first) and limited to top 10 results.
 */

export { examples }
