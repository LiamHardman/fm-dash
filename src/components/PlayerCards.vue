<template>
    <div class="fifa-card" @click="handleCardClick">
        <div class="card-bg"></div>
        <div class="card-content">
            <div class="card-header">
                <div class="player-rating">{{ player.overall }}</div>
                <div class="player-position" :style="positionStyle">{{ formattedPosition }}</div>
                <div class="player-name">{{ formattedPlayerName }}</div>
                <div class="player-vitals-container">
                    <div class="player-vitals">{{ playerVitals }}</div>
                </div>
            </div>

            <div class="card-middle">
                <div class="nation-flag" v-if="effectiveNationFlagUrl">
                    <img :src="effectiveNationFlagUrl" alt="Nation Flag" />
                </div>
                <div class="club-logo" v-if="player.club && player.club !== '-'">
                    <TeamLogo :team-name="player.club" :size="40" class="player-club-logo" />
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
                        <q-icon name="person" size="150px" color="grey-8" />
                    </div>
                </div>
            </div>

            <div class="card-footer">
                <div class="stats-grid">
                    <div class="stats-column">
                        <div class="stat-item"><span>{{ player.pac }}</span> PAC</div>
                        <div class="stat-item"><span>{{ player.sho }}</span> SHO</div>
                        <div class="stat-item"><span>{{ player.pas }}</span> PAS</div>
                    </div>
                    <div class="stats-column">
                        <div class="stat-item"><span>{{ player.dri }}</span> DRI</div>
                        <div class="stat-item"><span>{{ player.def }}</span> DEF</div>
                        <div class="stat-item"><span>{{ player.phy }}</span> PHY</div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref } from 'vue'
import { fetchFullPlayerStats } from '../services/playerService'
import TeamLogo from './TeamLogo.vue'

// Helper function to format currency
const formatCurrency = (value, symbol) => {
  if (!value || value === 'N/A') return 'N/A'

  // Handle transfer value ranges like "£127M - £381M" - extract upper bound
  if (typeof value === 'string' && value.includes(' - ')) {
    const parts = value.split(' - ')
    if (parts.length > 1) {
      value = parts[parts.length - 1].trim() // Take the upper bound
    }
  }

  // Check if the value already has proper formatting (like "£381M" or "£150K")
  if (typeof value === 'string') {
    const hasCurrencySymbol = /[£$€¥]/.test(value)
    const hasSuffix = /[MK]$/i.test(value)
    if (hasCurrencySymbol && hasSuffix) {
      // Already properly formatted, return as is
      return value
    }
  }

  // Remove any existing currency symbols and "p/w" or similar suffixes
  let cleanValue = value
  if (typeof cleanValue === 'string') {
    cleanValue = cleanValue
      .replace(/[£$€¥]/g, '')
      .replace(/\s*p\/w\s*$/i, '')
      .replace(/\s*per week\s*$/i, '')
      .replace(/,/g, '')
  }

  // Convert to number
  const numValue = parseFloat(cleanValue)
  if (isNaN(numValue)) return 'N/A'

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
      required: true
    },
    currencySymbol: {
      type: String,
      default: '£'
    },
    clubImageUrl: {
      type: String,
      default: null
    },
    nationFlagUrl: {
      type: String,
      default: null
    },
    playerFaceUrl: {
      type: String,
      default: null
    },
    datasetId: {
      type: String,
      default: null
    }
  },
  emits: ['click'],
  setup(props, { emit }) {
    const qInstance = useQuasar()
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
              score
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
        const positions = props.player.position.split(',').map(pos => pos.trim())
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
        STC: 'ST'
      }

      // Added CF as a direct mapping for striker roles
      if (bestPosition.includes('ST')) return 'ST'
      if (bestPosition.includes('AM')) {
        if (bestPosition.includes('C')) return 'CF' // Treat central AM as CF
      }

      return fmToFifaPositionMap[bestPosition] || bestPosition
    })

    // Helper function to format player name
    const formatPlayerName = fullName => {
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

    // Use detailed data if available, otherwise use props
    const playerData = computed(() => detailedPlayerData.value || props.player)

    // Generate image URLs based on available data
    const effectiveNationFlagUrl = computed(() => {
      if (props.nationFlagUrl) return props.nationFlagUrl
      if (playerData.value.nationality_iso) {
        return `https://flagcdn.com/w80/${playerData.value.nationality_iso.toLowerCase()}.png`
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

    // Fetch detailed data logic (unchanged)
    const needsDetailedData = computed(
      () => !props.player.nationality_iso && props.datasetId && props.player.uid
    )
    const fetchDetailedData = async () => {
      if (!needsDetailedData.value) return
      isLoadingDetailedData.value = true
      try {
        const result = await fetchFullPlayerStats(props.datasetId, props.player.uid)
        if (result.data && result.data.player) {
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
        let totalPixels = data.length / 4
        
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
      handleImageLoad
    }
  }
})
</script>

<style lang="scss" scoped>
// Define variables
$card-bg: #1e1e1e;
$gold-accent: #c89b3c;
$text-light: #e0e0e0;
$border-color: #444;

.fifa-card {
    width: 280px;
    height: 420px;
    background: $card-bg;
    border-radius: 12px;
    position: relative;
    overflow: hidden;
    color: $text-light;
    font-family: 'Roboto', 'Helvetica Neue', 'Arial', sans-serif;
    border: 2px solid $gold-accent;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
    transition: transform 0.3s ease, box-shadow 0.3s ease;
    cursor: pointer;

    &:hover {
        transform: translateY(-10px) scale(1.03);
        box-shadow: 0 20px 40px rgba($gold-accent, 0.3);
    }

    // This pseudo-element creates the grey diagonal background
    &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 185px; // Height of the top section
        background: #2a2a2a;
        transform: skewY(-5deg);
        transform-origin: top left;
        z-index: 1;
    }

    // This pseudo-element creates the gold diagonal line
    &::after {
        content: '';
        position: absolute;
        top: 280px; // Positioned between player face and stats
        left: -5%;
        width: 110%;
        height: 3px;
        background: linear-gradient(90deg, #d4af37, #c89b3c);
        z-index: 3;
        box-shadow: 0 0 10px rgba($gold-accent, 0.7);
    }
}

.card-content {
    position: relative;
    z-index: 2;
    height: 100%;
    display: flex;
    flex-direction: column;
}

// Header section
.card-header {
    position: relative;
    height: 170px;
    padding: 1rem;
    z-index: 4;

    .player-rating {
        position: absolute;
        top: 15px;
        left: 20px;
        font-size: 3rem;
        font-weight: 900;
        color: $text-light;
        line-height: 1;
    }

    .player-position {
        position: absolute;
        top: 60px;
        left: 20px;
        font-size: 1.4rem;
        font-weight: 600;
        color: $text-light;
        opacity: 0.9;
        min-width: 35px; // Reduced to account for 2-letter positions
        text-align: left;
    }

    .player-name {
        position: absolute;
        top: 25px;
        left: 80px; // Start after the rating area
        right: 20px; // Leave some space on the right
        font-size: 1.5rem;
        font-weight: 700;
        color: $text-light;
        white-space: nowrap;
        text-align: center;
        overflow: hidden;
        text-overflow: ellipsis; // Handle very long names
    }

    .player-vitals-container {
        position: absolute;
        top: 60px;
        right: 20px;
        background: $gold-accent;
        padding: 4px 8px;
        border-radius: 4px;
    }

    .player-vitals {
        font-size: 0.8rem;
        font-weight: 600;
        color: #1e1e1e;
        white-space: nowrap;
    }
}


// Middle section
.card-middle {
    flex-grow: 1;
    position: relative;
    z-index: 2;

    .nation-flag {
        position: absolute;
        top: -65px;
        left: 20px;
        width: 45px;
        z-index: 5;
        
        img {
            width: 100%;
            height: 30px;
            object-fit: cover;
            border-radius: 3px;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
    }

    .club-logo {
        position: absolute;
        top: -25px;
        left: 20px;
        width: 45px;
        height: 45px;
        z-index: 5;

        :deep(.team-logo),
        :deep(img) {
            width: 100%;
            height: 100%;
            object-fit: contain;
        }
    }

    .player-photo {
        position: absolute;
        top: -112px;
        right: 0px;
        width: 180px;
        height: 200px;
        
        img {
            width: 100%;
            height: 100%;
            object-fit: contain;
            object-position: bottom center;
        }
        
        .fallback-icon {
            width: 100%;
            height: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
    }
}

// Footer section
.card-footer {
    padding: 0 1.5rem 1rem 1.5rem;
    position: relative;
    z-index: 2;
    margin-top: auto; // Pushes footer to the bottom

    .stats-grid {
        display: flex;
        justify-content: space-between;
        gap: 1.5rem;
    }

    .stats-column {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .stat-item {
        display: flex;
        align-items: baseline;
        font-size: 1.1rem;
        font-weight: 500;
        color: $text-light;

        span {
            color: $gold-accent;
            font-weight: 700;
            font-size: 1.2rem;
            margin-right: 0.5rem;
            width: 35px; // Aligns the stat labels
            text-align: left;
        }
    }
}
</style>