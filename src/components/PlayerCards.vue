<template>
    <div class="fifa-card" @click="handleCardClick">
        <div class="card-content">
            <!-- Header: Rating, Name -->
            <div class="card-top">
                <div class="player-details">
                    <div class="player-rating">{{ player.overall }}</div>
                    <div class="player-name">{{ player.name }}</div>
                </div>
                <!-- Position moved below rating -->
                <div class="player-position" :class="{ 'position-2': formattedPosition.length === 2, 'position-3': formattedPosition.length === 3 }">{{ formattedPosition }}</div>
            </div>

            <!-- Middle: Photo, Nation, Club -->
            <div class="card-middle">
                <div class="nation-flag" v-if="effectiveNationFlagUrl">
                    <img :src="effectiveNationFlagUrl" alt="Nation Flag" />
                </div>
                <div class="club-logo" v-if="player.club && player.club !== '-'">
                    <TeamLogo 
                        :team-name="player.club"
                        :size="32"
                        class="player-club-logo"
                    />
                </div>
                <div class="player-photo">
                    <q-avatar
                        size="150px"
                        :color="'transparent'"
                        :text-color="$q.dark.isActive ? 'grey-5' : 'grey-8'"
                    >
                        <img v-if="effectivePlayerFaceUrl" :src="effectivePlayerFaceUrl" alt="Player Face" />
                        <q-icon v-else name="person" size="90px" />
                    </q-avatar>
                </div>
            </div>

            <!-- Footer: Stats -->
            <div class="card-bottom">
                <div class="stats-grid">
                    <div class="stat-item"><span>{{ player.pac }}</span> PAC</div>
                    <div class="stat-item"><span>{{ player.dri }}</span> DRI</div>
                    <div class="stat-item"><span>{{ player.sho }}</span> SHO</div>
                    <div class="stat-item"><span>{{ player.def }}</span> DEF</div>
                    <div class="stat-item"><span>{{ player.pas }}</span> PAS</div>
                    <div class="stat-item"><span>{{ player.phy }}</span> PHY</div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { defineComponent, ref, computed, onMounted } from 'vue'
import { fetchFullPlayerStats } from '../services/playerService'
import TeamLogo from './TeamLogo.vue'

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
      default: '$'
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

    // Format position to show only the first position
    const formattedPosition = computed(() => {
      if (!props.player.position) return ''
      
      // Find the position with the highest role-specific overall
      let bestPosition = null
      let bestScore = 0
      
      console.log('PlayerCards formattedPosition - Player:', props.player.name, 'RoleSpecificOveralls:', props.player.roleSpecificOveralls)
      console.log('PlayerCards - Full player object keys:', Object.keys(props.player))
      console.log('PlayerCards - Player object:', props.player)
      
      if (props.player.roleSpecificOveralls) {
        if (Array.isArray(props.player.roleSpecificOveralls)) {
          // Handle array format
          console.log('PlayerCards - Processing array format roleSpecificOveralls')
          console.log('PlayerCards - Array length:', props.player.roleSpecificOveralls.length)
          for (const rso of props.player.roleSpecificOveralls) {
            console.log('PlayerCards - Checking role:', rso.roleName, 'score:', rso.score)
            if (rso.score > bestScore) {
              bestScore = rso.score
              // Extract position from role name (e.g., "DC - Central Defender - Defend" -> "DC")
              const positionMatch = rso.roleName.match(/^([A-Z]+)\s*-/)
              if (positionMatch) {
                bestPosition = positionMatch[1]
                console.log('PlayerCards - Found better position:', bestPosition, 'from role:', rso.roleName)
              }
            }
          }
        } else {
          // Handle object format
          console.log('PlayerCards - Processing object format roleSpecificOveralls')
          for (const [roleName, score] of Object.entries(props.player.roleSpecificOveralls)) {
            console.log('PlayerCards - Checking role:', roleName, 'score:', score)
            if (score > bestScore) {
              bestScore = score
              // Extract position from role name (e.g., "DC - Central Defender - Defend" -> "DC")
              const positionMatch = roleName.match(/^([A-Z]+)\s*-/)
              if (positionMatch) {
                bestPosition = positionMatch[1]
                console.log('PlayerCards - Found better position:', bestPosition, 'from role:', roleName)
              }
            }
          }
        }
      }
      
      // If no role-specific overalls found, try using best_role_overall
      if (!bestPosition && props.player.best_role_overall) {
        console.log('PlayerCards - Using best_role_overall:', props.player.best_role_overall)
        const positionMatch = props.player.best_role_overall.match(/^([A-Z]+)\s*-/)
        if (positionMatch) {
          bestPosition = positionMatch[1]
          console.log('PlayerCards - Found position from best_role_overall:', bestPosition)
        }
      }
      
      console.log('PlayerCards - Best position from role-specific overalls:', bestPosition, 'with score:', bestScore)
      
      // If no role-specific overalls found, fall back to the original position logic
      if (!bestPosition) {
        console.log('PlayerCards - No role-specific overalls found, falling back to position string')
        // Split by comma and take the first position
        const positions = props.player.position.split(',').map(pos => pos.trim())
        const firstPosition = positions[0]
        
        // Extract the position type (e.g., "AM", "M", "D") and side (e.g., "R", "C", "L")
        const match = firstPosition.match(/^([A-Z]+)\s*\(([A-Z]+)\)$/)
        if (match) {
          const positionType = match[1]
          const side = match[2]
          bestPosition = positionType + side
          console.log('PlayerCards - Extracted position from parentheses:', bestPosition)
        } else {
          bestPosition = firstPosition
          console.log('PlayerCards - Using first position as-is:', bestPosition)
        }
      }
      
      // Translate FM positions to FIFA positions
      const fmToFifaPositionMap = {
        'GK': 'GK',
        'SW': 'SW',
        'DC': 'CB', // Centre Back
        'DR': 'RB', // Right Back
        'DL': 'LB', // Left Back
        'WBR': 'RWB', // Right Wing Back
        'WBL': 'LWB', // Left Wing Back
        'DM': 'CDM', // Centre Defensive Midfielder
        'MC': 'CM', // Centre Midfielder
        'MR': 'RM', // Right Midfielder
        'ML': 'LM', // Left Midfielder
        'AMC': 'CAM', // Centre Attacking Midfielder
        'AMR': 'RW', // Right Winger
        'AML': 'LW', // Left Winger
        'ST': 'ST' // Striker
      }
      
      const fifaPosition = fmToFifaPositionMap[bestPosition] || bestPosition
      console.log('PlayerCards - Final FIFA position:', fifaPosition, 'from FM position:', bestPosition)
      
      return fifaPosition
    })

    // Check if we need to fetch detailed data
    const needsDetailedData = computed(() => {
      const needsData = !props.player.nationality_iso && props.datasetId && props.player.uid
      console.log('PlayerCards needsDetailedData check:', {
        playerName: props.player.name,
        hasNationalityIso: !!props.player.nationality_iso,
        hasDatasetId: !!props.datasetId,
        hasUid: !!props.player.uid,
        needsData
      })
      return needsData
    })

    // Fetch detailed player data if nationality_iso is missing
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

    // Use detailed data if available, otherwise use props
    const playerData = computed(() => {
      return detailedPlayerData.value || props.player
    })

    // Generate image URLs based on available data
    const effectiveNationFlagUrl = computed(() => {
      if (props.nationFlagUrl) {
        console.log('PlayerCards using props.nationFlagUrl:', props.nationFlagUrl)
        return props.nationFlagUrl
      }
      if (playerData.value.nationality_iso) {
        const flagUrl = `https://flagcdn.com/w80/${playerData.value.nationality_iso.toLowerCase()}.png`
        console.log('PlayerCards generated flag URL:', flagUrl)
        return flagUrl
      }
      console.log('PlayerCards no flag URL available')
      return null
    })

    const effectiveClubImageUrl = computed(() => {
      if (props.clubImageUrl) {
        console.log('PlayerCards using props.clubImageUrl:', props.clubImageUrl, 'for player:', props.player.name)
        return props.clubImageUrl
      }
      // Generate club logo URL using the same approach as TeamLogo component
      if (props.player.club && props.player.club !== '-') {
        // The TeamLogo component handles this via useTeamLogosBackend
        // For now, we'll return null and let the TeamLogo component handle it
        console.log('PlayerCards: club logo will be handled by TeamLogo component for player:', props.player.name, 'club:', props.player.club)
        return null
      }
      console.log('PlayerCards no club logo URL available for player:', props.player.name, 'club:', props.player.club)
      return null
    })

    const effectivePlayerFaceUrl = computed(() => {
      if (props.playerFaceUrl) {
        console.log('PlayerCards using props.playerFaceUrl:', props.playerFaceUrl, 'for player:', props.player.name)
        return props.playerFaceUrl
      }
      // Generate player face URL using the same approach as PlayerDetailDialog
      const playerUID = props.player.UID || props.player.uid
      if (playerUID) {
        const faceUrl = `/api/faces?uid=${encodeURIComponent(playerUID)}`
        console.log('PlayerCards generated face URL:', faceUrl, 'for player:', props.player.name, 'UID:', playerUID)
        return faceUrl
      }
      console.log('PlayerCards no player face URL available for player:', props.player.name, 'UID not found')
      return null
    })

    const handleCardClick = () => {
      emit('click')
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
      effectiveClubImageUrl,
      effectivePlayerFaceUrl,
      isLoadingDetailedData,
      formattedPosition
    }
  }
})
</script>

<style lang="scss" scoped>
// Define variables for easy theme changes
$card-bg: #2a2a2a;
$gold-accent: #c89b3c;
$gold-gradient: linear-gradient(45deg, #d4af37, #c89b3c);
$text-light: #e0e0e0;
$text-dark: #333;
$border-color: #444;

// New FIFA Card Design
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

    // This pseudo-element creates the large diagonal background shape
    &::before {
        content: '';
        position: absolute;
        top: -50px;
        left: 0;
        width: 100%;
        height: 220px;
        background: #1c1c1c;
        transform: skewY(-8deg);
        z-index: 1;
    }

    // This pseudo-element creates the gold diagonal line
    &::after {
        content: '';
        position: absolute;
        top: 140px;
        left: -10%;
        width: 120%;
        height: 4px;
        background: $gold-gradient;
        transform: skewY(-8deg);
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
    padding: 1rem;
}

// Top section of the card
.card-top {
    position: relative;
    z-index: 4;
    padding: 0.5rem 1rem;
    
    .player-details {
        display: flex;
        align-items: baseline;
        gap: 0.75rem;
    }

    .player-rating {
        font-size: 2.5rem;
        font-weight: 900;
        color: $gold-accent;
        line-height: 1;
    }

    .player-name {
        font-size: 1.25rem;
        font-weight: 500;
        margin-left: auto; // Pushes the name to the right
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .player-position {
        font-size: 1.25rem;
        font-weight: 700;
        line-height: 1;
        margin-top: 0.25rem; // Reduced space for tighter alignment
        color: $text-light;
        text-align: center; // Default center alignment
        width: 3ch; // Fixed width for consistent spacing
        display: inline-block;
        
        // 2-letter positions: left align with small offset
        &.position-2 {
            text-align: left;
            margin-left: 0.25rem;
        }
        
        // 3-letter positions: center align
        &.position-3 {
            text-align: center;
            margin-left: 0;
        }
    }
}

// Middle section with placeholders and photo
.card-middle {
    position: relative;
    flex-grow: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2; // Below the gold line but above the bg

    .player-photo {
        position: absolute;
        right: 10px;
        top: 35px;
        
        .q-avatar {
            border: none;
            background: transparent;
            
            img {
                width: 100%;
                height: 100%;
                object-fit: cover;
            }
        }
    }

    .nation-flag, .club-logo {
        position: absolute;
        border: none;
        box-shadow: none;
    }

    .nation-flag {
        top: 25px; // Moved up to align with rating
        left: 10px;
        width: 50px;
        height: 35px;
        border-radius: 4px;
        overflow: hidden;
        
        img {
            width: 100%;
            height: 100%;
            object-fit: cover;
        }
    }

    .club-logo {
        top: 70px; // Moved up to align with position
        left: 10px;
        width: 50px;
        height: 50px;
        border-radius: 0;
        overflow: hidden;
        background: transparent;
        
        :deep(.team-logo) {
            width: 100%;
            height: 100%;
            object-fit: contain;
        }
        
        :deep(.team-logo-placeholder) {
            display: none;
        }
    }
}

// Bottom section with stats
.card-bottom {
    padding: 1rem;
    position: relative;
    z-index: 2;

    .stats-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        grid-template-rows: repeat(3, 1fr);
        gap: 0.5rem 1.5rem;
    }

    .stat-item {
        display: flex;
        align-items: baseline;
        font-size: 1.1rem;
        font-weight: 500;

        span {
            color: $gold-accent;
            font-weight: 700;
            font-size: 1.2rem;
            margin-right: 0.5rem;
            width: 30px; // Aligns the stat labels
        }
    }
}

// Responsive adjustments
@media (max-width: 600px) {
    .fifa-card {
        width: 260px;
        height: 400px;
    }
}
</style>
