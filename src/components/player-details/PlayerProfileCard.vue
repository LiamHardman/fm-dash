<template>
  <q-card
    flat
    bordered
    class="q-mb-sm player-profile-card modern-profile-card"
  >
    <q-card-section class="player-profile-content">
      <div class="profile-header-section">
        <div class="player-identity-section">
          <div class="row items-center q-mb-sm">
            <!-- Player Face Image -->
            <div v-if="showFaces" class="col-auto q-mr-md player-face-container">
              <img
                v-if="playerFaceImageUrl && !faceImageLoadError"
                :src="playerFaceImageUrl"
                :alt="`${player.name || 'Player'} face`"
                width="80"
                height="80"
                class="player-face-image"
                @error="handleFaceImageError"
                @load="handleFaceImageLoad"
              />
              <q-avatar
                v-else
                size="80px"
                :color="$q.dark.isActive ? 'grey-7' : 'grey-4'"
                :text-color="$q.dark.isActive ? 'grey-4' : 'grey-7'"
                class="player-face-placeholder"
              >
                <q-icon name="person" size="32px" />
              </q-avatar>
            </div>
            
            <div class="col-auto q-mr-md player-flag-container">
              <img
                v-if="player.nationality_iso && !flagLoadError"
                :src="`https://flagcdn.com/w80/${player.nationality_iso.toLowerCase()}.png`"
                :alt="player.nationality || 'Flag'"
                width="48"
                height="32"
                class="player-flag"
                @error="handleFlagError"
                :title="player.nationality"
              />
              <q-icon
                v-if="!player.nationality_iso || flagLoadError"
                :color="$q.dark.isActive ? 'grey-5' : 'grey-7'"
                name="flag"
                size="2.5em"
                class="player-flag-placeholder"
              />
              
              <!-- Club logo below nationality flag -->
              <div class="q-mt-sm club-logo-container" v-if="player.club && player.club !== '-'">
                <Suspense v-if="shouldShowTeamLogo">
                  <template #default>
                    <TeamLogo 
                      :team-name="player.club"
                      :size="32"
                      class="player-club-logo"
                    />
                  </template>
                  <template #fallback>
                    <div class="club-logo-placeholder">
                      <q-skeleton 
                        type="circle" 
                        size="32px"
                        class="club-logo-skeleton"
                      />
                    </div>
                  </template>
                </Suspense>
                <div v-else class="club-logo-placeholder">
                  <q-skeleton 
                    type="circle" 
                    size="32px"
                    class="club-logo-skeleton"
                  />
                </div>
              </div>
            </div>
            
            <div class="col player-name-section">
              <div class="player-name-container">
                <div class="player-name-and-status">
                  <h5
                    class="text-h5 player-name no-margin"
                    :class="$q.dark.isActive ? 'text-white' : 'text-dark'"
                    :title="player.name"
                  >
                    {{ player.name }}
                    <q-icon
                      v-if="player.attributeMasked"
                      name="warning"
                      color="warning"
                      size="sm"
                      class="q-ml-sm scouting-warning-icon"
                    >
                      <q-tooltip
                        :class="$q.dark.isActive ? 'bg-grey-7 text-white' : 'bg-white text-dark'"
                        :delay="300"
                        max-width="300px"
                        class="modern-tooltip"
                      >
                        <div class="tooltip-header">⚠️ Scouting Required</div>
                        <div class="tooltip-description">
                          Some of this player's attributes are masked. Scout this player before attempting to sign them to see their full attributes.
                        </div>
                      </q-tooltip>
                    </q-icon>
                  </h5>
                  <div class="player-status-badges q-mt-xs">
                    <q-badge
                      v-if="player.isNew"
                      outline
                      color="primary"
                      label="New"
                      class="player-status-badge q-mr-sm"
                    />
                    <q-badge
                      v-if="player.isLoaned"
                      outline
                      color="secondary"
                      label="Loaned"
                      class="player-status-badge q-mr-sm"
                    />
                    <q-badge
                      v-if="player.isOnLoan"
                      outline
                      color="teal"
                      label="On Loan"
                      class="player-status-badge q-mr-sm"
                    />
                    <q-badge
                      v-if="player.isFree"
                      outline
                      color="purple"
                      label="Free"
                      class="player-status-badge q-mr-sm"
                    />
                  </div>
                </div>
                <div class="player-badges-row q-mt-xs">
                  <q-badge
                    outline
                    color="primary"
                    :label="`${player.age || '-'} years`"
                    class="player-age-badge q-mr-sm"
                  />
                  <q-badge
                    outline
                    color="secondary"
                    :label="player.nationality || 'Unknown'"
                    class="player-nationality-badge q-mr-sm"
                  />
                  <q-badge
                    v-if="player.club"
                    outline
                    color="teal"
                    :label="player.club"
                    class="player-club-badge q-mr-sm"
                  />
                  <q-badge
                    v-if="player.personality"
                    outline
                    color="purple"
                    :label="player.personality"
                    class="player-personality-badge q-mr-sm"
                  />
                  <q-badge
                    v-if="player.media_handling"
                    outline
                    color="orange"
                    :label="player.media_handling"
                    class="player-media-badge"
                  />
                </div>
              </div>
              
              <div class="player-positions-section q-mt-sm" v-if="player.shortPositions?.length || player.position">
                <q-badge
                  v-for="pos in player.shortPositions || [player.position]"
                  :key="pos"
                  outline
                  color="indigo-6"
                  :label="pos"
                  class="position-badge q-mr-xs q-mb-xs"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="financial-details-section">
          <div class="financial-combined-item" :title="`${formattedTransferValue} / ${formattedWage}`">
            <div class="financial-content">
              <div class="financial-row">
                <q-icon name="trending_up" class="financial-icon q-mr-sm" />
                <div class="financial-item-content">
                  <div class="financial-label">Transfer Value</div>
                  <div class="financial-value transfer-value">{{ formattedTransferValue }}</div>
                </div>
              </div>
              <div class="financial-row q-mt-sm">
                <q-icon name="payments" class="financial-icon q-mr-sm" />
                <div class="financial-item-content">
                  <div class="financial-label">Weekly Salary</div>
                  <div class="financial-value wage-value">{{ formattedWage }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import { defineComponent, computed, ref } from 'vue'
import { useUiStore } from '@/stores/uiStore'
import TeamLogo from '@/components/TeamLogo.vue'

export default defineComponent({
  name: 'PlayerProfileCard',
  components: {
    TeamLogo
  },
  props: {
    player: {
      type: Object,
      required: true
    },
    currencySymbol: {
      type: String,
      default: '$'
    }
  },
  setup(props) {
    const uiStore = useUiStore()
    
    // Image loading states
    const faceImageLoadError = ref(false)
    const flagLoadError = ref(false)
    
    // UI settings
    const showFaces = computed(() => uiStore.showFaces)
    const shouldShowTeamLogo = computed(() => uiStore.shouldShowTeamLogo)
    
    // Player face image URL
    const playerFaceImageUrl = computed(() => {
      if (!props.player?.uid || !showFaces.value) return null
      return `/faces/${props.player.uid}.png`
    })
    
    // Formatted financial values
    const formattedTransferValue = computed(() => {
      if (!props.player?.transferValue) return 'N/A'
      return props.player.transferValue
    })
    
    const formattedWage = computed(() => {
      if (!props.player?.wage) return 'N/A'
      return props.player.wage
    })
    
    // Image error handlers
    const handleFaceImageError = () => {
      faceImageLoadError.value = true
    }
    
    const handleFaceImageLoad = () => {
      faceImageLoadError.value = false
    }
    
    const handleFlagError = () => {
      flagLoadError.value = true
    }
    
    return {
      faceImageLoadError,
      flagLoadError,
      showFaces,
      shouldShowTeamLogo,
      playerFaceImageUrl,
      formattedTransferValue,
      formattedWage,
      handleFaceImageError,
      handleFaceImageLoad,
      handleFlagError
    }
  }
})
</script>

<style lang="scss" scoped>
.player-profile-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  
  .body--dark & {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
}

.player-face-image {
  border-radius: 8px;
  object-fit: cover;
}

.player-face-placeholder {
  border-radius: 8px;
}

.player-flag {
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.player-name {
  font-weight: 600;
  letter-spacing: 0.5px;
}

.scouting-warning-icon {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.player-status-badge,
.player-age-badge,
.player-nationality-badge,
.player-club-badge,
.player-personality-badge,
.player-media-badge {
  font-size: 0.75rem;
  padding: 4px 8px;
}

.position-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 4px 8px;
}

.financial-details-section {
  margin-top: 1rem;
  padding: 1rem;
  background: rgba(25, 118, 210, 0.05);
  border-radius: 8px;
  
  .body--dark & {
    background: rgba(255, 255, 255, 0.03);
  }
}

.financial-row {
  display: flex;
  align-items: center;
}

.financial-icon {
  color: #1976d2;
  font-size: 1.2rem;
  
  .body--dark & {
    color: #64b5f6;
  }
}

.financial-label {
  font-size: 0.75rem;
  opacity: 0.8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.financial-value {
  font-size: 1rem;
  font-weight: 600;
  
  &.transfer-value {
    color: #1976d2;
    
    .body--dark & {
      color: #64b5f6;
    }
  }
  
  &.wage-value {
    color: #388e3c;
    
    .body--dark & {
      color: #81c784;
    }
  }
}

.club-logo-container {
  display: flex;
  justify-content: center;
}

.club-logo-skeleton {
  background: rgba(0, 0, 0, 0.1);
  
  .body--dark & {
    background: rgba(255, 255, 255, 0.1);
  }
}
</style> 