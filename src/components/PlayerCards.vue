<template>
    <div class="fifa-card" :class="[cardTypeClass, rarityClass, cardDesignClass]" @click="handleCardClick">
        <div class="card-bg"></div>
        <div class="card-content">
            <div class="card-header">
                <div class="rating-block">
                    <div class="player-rating">{{ playerOverall }}</div>
                    <div class="player-position" :style="positionStyle">{{ formattedPosition }}</div>
                </div>
                <div class="player-identity">
                    <div class="player-name">{{ formattedPlayerName }}</div>
                    <div class="player-vitals-container">
                        <div class="player-vitals">{{ playerVitals }}</div>
                    </div>
                </div>
            </div>

            <div class="card-middle">
                <div class="card-marks">
                    <div class="nation-flag" v-if="effectiveNationFlagUrl">
                        <img :src="effectiveNationFlagUrl" alt="Nation Flag" />
                    </div>
                    <div class="club-logo" v-if="player.club && player.club !== '-'">
                        <TeamLogo :team-name="player.club" :size="40" class="player-club-logo" />
                    </div>
                </div>
                <div class="player-photo">
                    <img 
                        v-if="effectivePlayerFaceUrl && !imageLoadError" 
                        :src="effectivePlayerFaceUrl" 
                        alt="Player Face" 
                        @error="handleImageError"
                        @load="handleImageLoad"
                        ref="playerImage"
                    />
                    <div v-else class="fallback-icon">
                        <q-icon name="person" size="124px" />
                    </div>
                </div>
            </div>

            <div class="card-footer">
                <div class="stats-grid">
                    <div
                        v-for="stat in playerStats"
                        :key="stat.label"
                        class="stat-item"
                    >
                        <span>{{ stat.value }}</span> {{ stat.label }}
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref, watch } from 'vue'
import { useUiStore } from '@/stores/uiStore'
import { getFifaWeightedOverallByPosition } from '@/utils/playerUtils'
import { fetchFullPlayerStats } from '../services/playerService'
import TeamLogo from './TeamLogo.vue'

// Helper function to format currency
const formatCurrency = (value, symbol) => {
  if (!value || value === 'N/A') return 'N/A'

  // Handle transfer value ranges like "£127M - £381M" - extract upper bound
  let processedValue = value
  if (typeof processedValue === 'string' && processedValue.includes(' - ')) {
    const parts = processedValue.split(' - ')
    if (parts.length > 1) {
      processedValue = parts[parts.length - 1].trim() // Take the upper bound
    }
  }

  // Check if the value already has proper formatting (like "£381M" or "£150K")
  if (typeof processedValue === 'string') {
    const hasCurrencySymbol = /[£$€¥]/.test(processedValue)
    const hasSuffix = /[MK]$/i.test(processedValue)
    if (hasCurrencySymbol && hasSuffix) {
      // Already properly formatted, return as is
      return processedValue
    }
  }

  // Remove any existing currency symbols and "p/w" or similar suffixes
  let cleanValue = processedValue
  if (typeof cleanValue === 'string') {
    cleanValue = cleanValue
      .replace(/[£$€¥]/g, '')
      .replace(/\s*p\/w\s*$/i, '')
      .replace(/\s*per week\s*$/i, '')
      .replace(/,/g, '')
  }

  // Convert to number
  const numValue = Number.parseFloat(cleanValue)
  if (Number.isNaN(numValue)) return 'N/A'

  if (numValue >= 1000000) {
    return `${symbol}${(numValue / 1000000).toFixed(1).replace('.0', '')}M`
  }
  if (numValue >= 1000) {
    return `${symbol}${(numValue / 1000).toFixed(0)}K`
  }
  return `${symbol}${numValue}`
}

export default defineComponent({
  name: 'PlayerCards',
  components: { TeamLogo },
  props: {
    player: {
      type: Object,
      required: true,
    },
    currencySymbol: {
      type: String,
      default: '£',
    },
    clubImageUrl: {
      type: String,
      default: null,
    },
    nationFlagUrl: {
      type: String,
      default: null,
    },
    playerFaceUrl: {
      type: String,
      default: null,
    },
    datasetId: {
      type: String,
      default: null,
    },
    selectedRole: {
      type: String,
      default: null,
    },
  },
  emits: ['click'],
  setup(props, { emit }) {
    const qInstance = useQuasar()
    const uiStore = useUiStore()
    const detailedPlayerData = ref(null)
    const isLoadingDetailedData = ref(false)
    const imageLoadError = ref(false)

    // Format position to show only the first position
    const formattedPosition = computed(() => {
      if (!props.player.position) return ''
      let bestPosition = null
      let bestScore = 0

      if (props.player.roleSpecificOveralls) {
        const roles = Array.isArray(props.player.roleSpecificOveralls)
          ? props.player.roleSpecificOveralls
          : Object.entries(props.player.roleSpecificOveralls).map(([roleName, score]) => ({
              roleName,
              score,
            }))

        for (const rso of roles) {
          const score = typeof rso.score === 'number' ? rso.score : rso[1]
          const roleName = rso.roleName || rso[0]
          if (score > bestScore) {
            bestScore = score
            const positionMatch = roleName.match(/^([A-Z]+)\s*-/)
            if (positionMatch) {
              bestPosition = positionMatch[1]
            }
          }
        }
      }

      if (!bestPosition && props.player.best_role_overall) {
        const positionMatch = props.player.best_role_overall.match(/^([A-Z]+)\s*-/)
        if (positionMatch) {
          bestPosition = positionMatch[1]
        }
      }

      if (!bestPosition) {
        const positions = props.player.position.split(',').map((pos) => pos.trim())
        const firstPosition = positions[0]
        const match = firstPosition.match(/^([A-Z]+)\s*\(([A-Z]+)\)$/)
        bestPosition = match ? match[1] + match[2] : firstPosition
      }

      const fmToFifaPositionMap = {
        GK: 'GK',
        SW: 'SW',
        DC: 'CB',
        DR: 'RB',
        DL: 'LB',
        WBR: 'RWB',
        WBL: 'LWB',
        DM: 'CDM',
        MC: 'CM',
        MR: 'RM',
        ML: 'LM',
        AMC: 'CAM',
        AMR: 'RW',
        AML: 'LW',
        ST: 'ST',
        STC: 'ST',
      }

      // Added CF as a direct mapping for striker roles
      if (bestPosition.includes('ST')) return 'ST'
      if (bestPosition.includes('AM')) {
        if (bestPosition.includes('C')) return 'CAM'
      }

      return fmToFifaPositionMap[bestPosition] || bestPosition
    })

    // Helper function to format player name
    const formatPlayerName = (fullName) => {
      if (!fullName) return ''

      const nameParts = fullName.trim().split(' ')

      // If only one name, return as is (e.g., "Ronaldinho")
      if (nameParts.length === 1) {
        return fullName
      }

      // For multiple names, return the last name
      // Handle cases like "Frenkie de Jong" -> "De Jong"
      const lastName = nameParts[nameParts.length - 1]

      // Check if the last name starts with "de", "van", "von", etc. and include the previous part
      const lowercaseLastName = lastName.toLowerCase()
      if (
        nameParts.length >= 2 &&
        ['de', 'van', 'von', 'del', 'da', 'di', 'du', 'le', 'la'].includes(lowercaseLastName)
      ) {
        const secondToLast = nameParts[nameParts.length - 2]
        return `${secondToLast} ${lastName}`
      }

      return lastName
    }

    // New computed property for player vitals
    const playerVitals = computed(() => {
      const age = props.player.age || 'N/A'

      // Check multiple possible property names for transfer value - prioritize numeric amount
      const transferValue =
        props.player.transferValueAmount ||
        props.player.transfer_value ||
        props.player.value ||
        props.player.market_value
      const formattedValue = transferValue
        ? formatCurrency(transferValue, props.currencySymbol)
        : 'N/A'

      // Check multiple possible property names for salary/wage - prioritize numeric amount
      const salary =
        props.player.wageAmount ||
        props.player.wage ||
        props.player.salary ||
        props.player.weekly_wage
      const formattedSalary = salary ? formatCurrency(salary, props.currencySymbol) : 'N/A'

      return `${age} | ${formattedValue} | ${formattedSalary}`
    })

    // Computed property for formatted player name
    const formattedPlayerName = computed(() => {
      return formatPlayerName(props.player.name)
    })

    // Computed property for position styling
    const positionStyle = computed(() => {
      const position = formattedPosition.value
      if (position && position.length === 2) {
        return { marginLeft: '8px' } // Move 2-letter positions slightly right
      }
      return {}
    })

    // Computed property for player overall - shows role-specific rating when role is selected
    const playerOverall = computed(() => {
      // Use detailed player data if available, otherwise use props
      const playerData = detailedPlayerData.value || props.player

      // If no role is selected, optionally use stat summary overall if setting is enabled
      if (!props.selectedRole) {
        if (uiStore?.useStatSummaryForOverall) {
          // Prefer the already-calibrated Overall from the row mapping when available
          const mapped = playerData.Overall ?? playerData.overall
          if (typeof mapped === 'number' && mapped > 0) return mapped
          // Fallback to raw FIFA summary if mapping isn't present (e.g., detail view)
          try {
            return getFifaWeightedOverallByPosition(playerData)
          } catch (_e) {}
        }
        return playerData.overall || playerData.Overall || 0
      }

      // If a role is selected, try to get the role-specific overall
      if (playerData.roleSpecificOveralls) {
        let roleOverall = null

        if (Array.isArray(playerData.roleSpecificOveralls)) {
          // Try exact match first
          let roleData = playerData.roleSpecificOveralls.find(
            (r) => r.roleName === props.selectedRole
          )

          // If no exact match, try partial match (e.g., "WBL" matches "WBL - Complete Wing Back - Support")
          if (!roleData) {
            const selectedRolePrefix = props.selectedRole.split(' - ')[0] // Get "WBL" from "WBL - Complete Wing Back - Support"
            roleData = playerData.roleSpecificOveralls.find((r) => {
              const roleName = r.roleName || r[0]
              return roleName?.startsWith(selectedRolePrefix)
            })
          }

          roleOverall = roleData ? roleData.score : null
        } else if (typeof playerData.roleSpecificOveralls === 'object') {
          // Try exact match first
          roleOverall = playerData.roleSpecificOveralls[props.selectedRole] || null

          // If no exact match, try partial match
          if (roleOverall === null) {
            const selectedRolePrefix = props.selectedRole.split(' - ')[0]
            const matchingKey = Object.keys(playerData.roleSpecificOveralls).find((key) =>
              key.startsWith(selectedRolePrefix)
            )
            roleOverall = matchingKey ? playerData.roleSpecificOveralls[matchingKey] : null
          }
        }

        // If we found a role-specific rating, use it; otherwise fall back to main overall
        if (roleOverall !== null && roleOverall > 0) {
          return roleOverall
        }
      }

      // Fall back to main overall if no role-specific rating found
      if (uiStore?.useStatSummaryForOverall && !props.selectedRole) {
        try {
          return getFifaWeightedOverallByPosition(playerData)
        } catch (_e) {}
      }
      return playerData.overall || playerData.Overall || 0
    })

    // Use detailed data if available, otherwise use props
    const playerData = computed(() => detailedPlayerData.value || props.player)

    // Card type and rarity logic
    const cardType = computed(() => {
      if (isIconCard.value) return 'icon'

      const overall = playerOverall.value || 0
      if (overall >= 75) return 'gold'
      if (overall >= 65) return 'silver'
      return 'bronze'
    })

    const isIconCard = computed(() => {
      const player = props.player || {}
      const cardFields = [
        player.cardType,
        player.card_type,
        player.card_design,
        player.cardDesign,
        player.rarity,
        player.playerType,
        player.player_type,
      ]

      return (
        player.isIcon === true ||
        cardFields.some((value) => String(value || '').toLowerCase() === 'icon')
      )
    })

    const isRare = computed(() => {
      const overall = playerOverall.value || 0

      // If player is 85 rated or higher, they should be rare no matter what
      if (overall >= 85) return true

      // Get appropriate stats based on player type
      let stats = []
      if (isGoalkeeper.value) {
        // Use goalkeeper stats
        stats = [
          props.player.div || 0,
          props.player.han || 0,
          props.player.kic || 0,
          props.player.ref || 0,
          props.player.spd || 0,
          props.player.pos || 0,
        ]
      } else {
        // Use outfield player stats
        stats = [
          props.player.pac || 0,
          props.player.sho || 0,
          props.player.pas || 0,
          props.player.dri || 0,
          props.player.def || 0,
          props.player.phy || 0,
        ]
      }

      // If a player has at least 2 stats that are within 1 point of the player's overall or higher, they should be rare
      const statsCloseToOverall = stats.filter((stat) => stat >= overall - 1).length
      if (statsCloseToOverall >= 2) return true

      // If 4 of their stats are within 4 points of the overall rating, they should be rare
      const statsWithinRange = stats.filter((stat) => stat >= overall - 4).length
      if (statsWithinRange >= 4) return true

      return false
    })

    const cardTypeClass = computed(() => `card-${cardType.value}`)
    const rarityClass = computed(() => (isRare.value ? 'rare' : 'non-rare'))
    const cardDesignClass = computed(() => {
      if (cardType.value === 'icon') return 'card-design--ivory-museum'

      const designByTypeAndRarity = {
        bronze: {
          'non-rare': 'card-design--aged-plate',
          rare: 'card-design--amber-facets',
        },
        silver: {
          'non-rare': 'card-design--pale-geometric',
          rare: 'card-design--polished-silver',
        },
        gold: {
          'non-rare': 'card-design--antique-foil',
          rare: 'card-design--confetti-foil',
        },
      }

      return designByTypeAndRarity[cardType.value]?.[rarityClass.value] || 'card-design--aged-plate'
    })

    // Check if player is a goalkeeper
    const isGoalkeeper = computed(() => {
      if (!props.player.position) return false

      // Check various ways the position might indicate goalkeeper
      const position = props.player.position.toLowerCase()
      const shortPositions = props.player.shortPositions || props.player.short_positions || []
      const positionGroups = props.player.positionGroups || []
      const parsedPositions = props.player.parsedPositions || []

      return (
        position.includes('gk') ||
        position.includes('goalkeeper') ||
        shortPositions.some((pos) => pos === 'GK') ||
        positionGroups.some((group) => group === 'Goalkeepers') ||
        parsedPositions.some((pos) => pos === 'Goalkeeper')
      )
    })

    // Generate image URLs based on available data
    const effectiveNationFlagUrl = computed(() => {
      if (props.nationFlagUrl) return props.nationFlagUrl
      if (playerData.value.nationalityIso) {
        return `https://flagcdn.com/w80/${playerData.value.nationalityIso.toLowerCase()}.png`
      }
      return null
    })

    const effectivePlayerFaceUrl = computed(() => {
      if (props.playerFaceUrl) return props.playerFaceUrl
      const playerUID = props.player.UID || props.player.uid
      if (playerUID) {
        return `/api/faces?uid=${encodeURIComponent(playerUID)}`
      }
      return null
    })

    const playerStats = computed(() => {
      if (isGoalkeeper.value) {
        return [
          { label: 'DIV', value: props.player.div },
          { label: 'HAN', value: props.player.han },
          { label: 'KIC', value: props.player.kic },
          { label: 'REF', value: props.player.ref },
          { label: 'SPD', value: props.player.spd },
          { label: 'POS', value: props.player.pos },
        ]
      }

      return [
        { label: 'PAC', value: props.player.pac },
        { label: 'DRI', value: props.player.dri },
        { label: 'SHO', value: props.player.sho },
        { label: 'DEF', value: props.player.def },
        { label: 'PAS', value: props.player.pas },
        { label: 'PHY', value: props.player.phy },
      ]
    })

    // Fetch detailed data logic (unchanged)
    const needsDetailedData = computed(() => {
      // Fetch if missing nationalityIso
      if (!props.player.nationalityIso && props.datasetId && props.player.uid) {
        return true
      }

      // Fetch if a role is selected but player doesn't have role-specific overalls
      if (props.selectedRole && props.datasetId && props.player.uid) {
        const hasRoleSpecificOveralls =
          props.player.roleSpecificOveralls &&
          (Array.isArray(props.player.roleSpecificOveralls)
            ? props.player.roleSpecificOveralls.length > 0
            : Object.keys(props.player.roleSpecificOveralls).length > 0)

        if (!hasRoleSpecificOveralls) {
          return true
        }
      }

      return false
    })
    const fetchDetailedData = async () => {
      if (!needsDetailedData.value) return
      isLoadingDetailedData.value = true
      try {
        const result = await fetchFullPlayerStats(props.datasetId, props.player.uid)
        if (result.data?.player) {
          detailedPlayerData.value = result.data.player
        }
      } catch (error) {
        console.error('Failed to fetch detailed player data:', error)
      } finally {
        isLoadingDetailedData.value = false
      }
    }

    const handleCardClick = () => emit('click')

    const handleImageError = () => {
      console.log('Image failed to load')
      imageLoadError.value = true
    }

    const handleImageLoad = (event) => {
      console.log('Image loaded:', event.target.src)

      // Simple check: if the image dimensions are very small, it's likely a placeholder
      if (event.target.naturalWidth < 50 || event.target.naturalHeight < 50) {
        console.log('Image too small, treating as error')
        imageLoadError.value = true
        return
      }

      // Try canvas analysis, but catch any errors
      try {
        const canvas = document.createElement('canvas')
        const ctx = canvas.getContext('2d')
        canvas.width = event.target.naturalWidth
        canvas.height = event.target.naturalHeight
        ctx.drawImage(event.target, 0, 0)

        const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
        const data = imageData.data

        // Check if the image is mostly white/transparent
        let whitePixels = 0
        const totalPixels = data.length / 4

        for (let i = 0; i < data.length; i += 4) {
          const r = data[i]
          const g = data[i + 1]
          const b = data[i + 2]
          const a = data[i + 3]

          // Consider pixel white if RGB values are high and alpha is low, or if alpha is very low
          if ((r > 240 && g > 240 && b > 240 && a < 50) || a < 10) {
            whitePixels++
          }
        }

        const whitePercentage = whitePixels / totalPixels
        console.log('White pixel percentage:', whitePercentage)

        // If more than 90% of pixels are white/transparent, treat as error
        if (whitePercentage > 0.9) {
          console.log('Image mostly white, treating as error')
          imageLoadError.value = true
        }
      } catch (error) {
        console.log('Canvas analysis failed:', error)
        // If canvas analysis fails, we'll keep the image as is
      }

      // Additional check: if the image appears to be a 1x1 pixel or very small, it's likely a placeholder
      if (event.target.naturalWidth === 1 && event.target.naturalHeight === 1) {
        console.log('Image is 1x1 pixel, treating as error')
        imageLoadError.value = true
      }
    }

    onMounted(() => {
      if (needsDetailedData.value) {
        fetchDetailedData()
      }
    })

    // Watch for changes in selectedRole and refetch detailed data if needed
    watch(
      () => props.selectedRole,
      () => {
        if (needsDetailedData.value) {
          fetchDetailedData()
        }
      }
    )

    return {
      qInstance,
      handleCardClick,
      effectiveNationFlagUrl,
      effectivePlayerFaceUrl,
      formattedPosition,
      playerVitals,
      formattedPlayerName,
      positionStyle,
      imageLoadError,
      handleImageError,
      handleImageLoad,
      cardTypeClass,
      rarityClass,
      cardDesignClass,
      isGoalkeeper,
      playerOverall,
      playerStats,
    }
  },
})
</script>

<style lang="scss" scoped>
@use "../styles/card-designs";
</style>
