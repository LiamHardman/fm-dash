<template>
    <q-dialog
        :model-value="show"
        @update:model-value="$emit('close')"
        persistent
        maximized
        transition-show="slide-up"
        transition-hide="slide-down"
    >
        <q-card class="recommend-signing-dialog">
            <q-card-section class="row items-center q-pb-none card-header">
                <q-icon name="person_add" size="md" class="q-mr-sm" color="warning" />
                <div class="text-h6">Recommend Signing — {{ slotRole }}</div>
                <q-space />
                <q-btn
                    icon="close"
                    flat
                    round
                    dense
                    v-close-popup
                    @click="$emit('close')"
                />
            </q-card-section>

            <q-card-section class="q-pt-md">
                <!-- Filter Row -->
                <div class="row q-col-gutter-md q-mb-lg">
                    <div class="col-12 col-md-4">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
                        >
                            Overall Range:
                            <span
                                class="overall-badge q-ml-xs"
                                :class="getOverallClass(overallRange.min)"
                            >{{ overallRange.min }}</span>
                            –
                            <span
                                class="overall-badge"
                                :class="getOverallClass(overallRange.max)"
                            >{{ overallRange.max }}</span>
                        </div>
                        <q-range
                            v-model="overallRange"
                            :min="0"
                            :max="100"
                            :step="1"
                            label-always
                            :left-label-value="overallRange.min"
                            :right-label-value="overallRange.max"
                            color="primary"
                            class="q-px-sm"
                            :disable="loading"
                        />
                    </div>

                    <div class="col-12 col-md-4">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
                        >
                            Max Transfer Value:
                            {{ maxTransferValue >= sliderMaxValue ? 'Any' : formatCurrency(maxTransferValue, currencySymbol) }}
                        </div>
                        <q-slider
                            v-model="maxTransferValue"
                            :min="0"
                            :max="sliderMaxValue"
                            :step="sliderStep"
                            label-always
                            :label-value="maxTransferValue >= sliderMaxValue ? 'Any' : formatCurrency(maxTransferValue, currencySymbol)"
                            color="primary"
                            class="q-px-sm"
                            :disable="loading"
                        />
                    </div>

                    <div class="col-12 col-md-4">
                        <div class="filter-info-box" :class="$q.dark.isActive ? 'filter-info-dark' : ''">
                            <div class="info-row">
                                <q-icon name="sports_soccer" size="xs" class="q-mr-xs" />
                                <span class="text-caption">Position: {{ slotRole }}</span>
                            </div>
                            <div class="info-row">
                                <q-icon name="group_remove" size="xs" class="q-mr-xs" />
                                <span class="text-caption">Excluding {{ teamPlayers.length }} squad members</span>
                            </div>
                            <div class="info-row">
                                <q-icon name="sort" size="xs" class="q-mr-xs" />
                                <span class="text-caption">Sorted by Moneyball Rating</span>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Loading State -->
                <div v-if="loading" class="text-center q-my-xl">
                    <q-spinner-dots color="primary" size="3em" />
                    <div class="text-subtitle1 q-mt-md">Loading player database...</div>
                    <div class="text-caption q-mt-xs text-grey-6">
                        Finding players who can play {{ slotRole }}
                    </div>
                </div>

                <!-- No Results -->
                <div
                    v-else-if="!loading && filteredPlayers.length === 0"
                    class="text-center q-my-xl"
                >
                    <q-icon name="manage_search" size="3em" color="grey-5" />
                    <div class="text-subtitle1 q-mt-md">No matching players found</div>
                    <div class="text-caption q-mt-xs text-grey-6">
                        Try lowering the minimum overall or increasing the max transfer value
                    </div>
                </div>

                <!-- Results Table -->
                <template v-else>
                    <div class="row items-center q-mb-sm">
                        <div class="text-subtitle2">
                            <q-icon name="person_search" class="q-mr-xs" />
                            {{ filteredPlayers.length }} potential signings
                        </div>
                        <q-space />
                        <q-chip
                            color="primary"
                            text-color="white"
                            icon="trending_up"
                            size="sm"
                            class="q-mr-xs"
                        >
                            Sorted by MBR
                        </q-chip>
                    </div>

                    <PlayerDataTable
                        :players="filteredPlayers"
                        :loading="false"
                        @player-selected="handlePlayerSelected"
                        @team-selected="handleTeamSelected"
                        :currency-symbol="currencySymbol"
                        :dataset-id="datasetId"
                        default-sort-field="MBR"
                        default-sort-direction="desc"
                    />
                </template>
            </q-card-section>
        </q-card>

        <PlayerDetailDialog
            :player="selectedPlayer"
            :show="showPlayerDetail"
            @close="showPlayerDetail = false"
            :currency-symbol="currencySymbol"
            :dataset-id="datasetId"
        />
    </q-dialog>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, ref, watch } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import { formatCurrency } from '../utils/currencyUtils'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

export default defineComponent({
  name: 'RecommendSigningDialog',
  components: { PlayerDataTable, PlayerDetailDialog },
  props: {
    show: { type: Boolean, default: false },
    slotRole: { type: String, default: '' },
    slotPositions: { type: Array, default: () => [] },
    teamPlayers: { type: Array, default: () => [] },
    datasetId: { type: String, default: null },
    currencySymbol: { type: String, default: '£' },
  },
  emits: ['close'],
  setup(props) {
    const $q = useQuasar()
    const playerStore = usePlayerStore()

    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)

    const overallRange = ref({ min: 40, max: 100 })
    const maxTransferValue = ref(0)

    const teamAvgOverall = computed(() => {
      if (!props.teamPlayers.length) return 60
      const overalls = props.teamPlayers
        .map((p) => Number(p.Overall || p.overall || 0))
        .filter((v) => v > 0)
      if (!overalls.length) return 60
      return Math.round(overalls.reduce((a, b) => a + b, 0) / overalls.length)
    })

    const teamMaxTransferValue = computed(() => {
      if (!props.teamPlayers.length) return 50000000
      const values = props.teamPlayers
        .map((p) => Number(p.transferValueAmount || 0))
        .filter((v) => v > 0)
      if (!values.length) return 50000000
      return Math.max(...values)
    })

    const sliderMaxValue = computed(() => {
      const base = teamMaxTransferValue.value
      const raw = Math.max(base * 5, 100000000)
      return Math.ceil(raw / 1000000) * 1000000
    })

    const sliderStep = computed(() => {
      const max = sliderMaxValue.value
      if (max >= 500000000) return 5000000
      if (max >= 100000000) return 1000000
      if (max >= 10000000) return 500000
      return 100000
    })

    const initFilters = () => {
      // Find players in the team who already play this specific position
      const positionPlayers = props.teamPlayers.filter((p) => {
        const positions = p.shortPositions || p.short_positions || []
        return positions.some((pos) => props.slotPositions.includes(pos))
      })

      if (positionPlayers.length > 0) {
        // Anchor on the lowest-rated player already in this slot, minus a small buffer,
        // rounded down to the nearest 5 for a clean threshold.
        const lowestOverall = Math.min(
          ...positionPlayers.map((p) => Number(p.Overall || p.overall || 0)).filter((v) => v > 0)
        )
        const buffered = Math.max(0, lowestOverall - 3)
        overallRange.value = { min: Math.floor(buffered / 5) * 5, max: 100 }
      } else {
        // No players in this position yet — fall back to team average minus a margin
        overallRange.value = { min: Math.max(0, teamAvgOverall.value - 10), max: 100 }
      }

      // Default max transfer value to the team's actual maximum (not a multiple of it)
      maxTransferValue.value = Math.min(teamMaxTransferValue.value, sliderMaxValue.value)
    }

    const loadAllPlayers = async () => {
      if (!props.datasetId) return
      if (playerStore.allPlayers.length > 0) return

      loading.value = true
      try {
        await playerStore.fetchPlayersByDatasetId(props.datasetId)
      } catch (err) {
        console.error('Failed to load players for signing recommendations:', err)
      } finally {
        loading.value = false
      }
    }

    const teamMemberNames = computed(
      () => new Set(props.teamPlayers.map((p) => p.name).filter(Boolean))
    )

    const filteredPlayers = computed(() => {
      if (loading.value || !playerStore.allPlayers.length) return []
      const positions = props.slotPositions
      if (!positions.length) return []

      const isAny = maxTransferValue.value >= sliderMaxValue.value
      const maxTV = isAny ? Infinity : maxTransferValue.value
      const minOvr = overallRange.value.min
      const maxOvr = overallRange.value.max
      const teamNames = teamMemberNames.value

      return playerStore.allPlayers.filter((player) => {
        const playerPositions = player.shortPositions || player.short_positions || []
        if (!playerPositions.some((pos) => positions.includes(pos))) return false
        if (teamNames.has(player.name)) return false
        const overall = Number(player.Overall || player.overall || 0)
        if (overall < minOvr || overall > maxOvr) return false
        if (maxTV !== Infinity) {
          const tv = Number(player.transferValueAmount || 0)
          if (tv > maxTV) return false
        }
        return true
      })
    })

    watch(
      () => props.show,
      async (newVal) => {
        if (newVal) {
          initFilters()
          await loadAllPlayers()
        }
      }
    )

    const getOverallClass = (overall) => {
      if (!overall) return ''
      if (overall >= 90) return 'rating-tier-6'
      if (overall >= 80) return 'rating-tier-5'
      if (overall >= 70) return 'rating-tier-4'
      if (overall >= 55) return 'rating-tier-3'
      if (overall >= 40) return 'rating-tier-2'
      return 'rating-tier-1'
    }

    const handlePlayerSelected = (player) => {
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const handleTeamSelected = () => {}

    return {
      $q,
      loading,
      overallRange,
      maxTransferValue,
      sliderMaxValue,
      sliderStep,
      filteredPlayers,
      selectedPlayer,
      showPlayerDetail,
      handlePlayerSelected,
      handleTeamSelected,
      getOverallClass,
      formatCurrency,
    }
  },
})
</script>

<style lang="scss" scoped>
.recommend-signing-dialog {
    display: flex;
    flex-direction: column;
    height: 100%;

    .card-header {
        padding: 1rem 1.5rem;
        border-bottom: 1px solid rgba(0, 0, 0, 0.08);

        .body--dark & {
            border-bottom-color: rgba(255, 255, 255, 0.1);
        }
    }

    .q-card__section:last-child {
        flex: 1;
        overflow-y: auto;
    }
}

.slider-label {
    font-weight: 500;
}

.overall-badge {
    display: inline-block;
    padding: 0 6px;
    border-radius: 4px;
    font-weight: 700;
    font-size: 0.75rem;
    color: white;

    &.rating-tier-6 { background: #10b981; }
    &.rating-tier-5 { background: #34d399; color: #1e293b; }
    &.rating-tier-4 { background: #fbbf24; color: #1e293b; }
    &.rating-tier-3 { background: #f59e0b; color: #1e293b; }
    &.rating-tier-2 { background: #ef4444; }
    &.rating-tier-1 { background: #991b1b; }
}

.filter-info-box {
    background: rgba(0, 0, 0, 0.04);
    border-radius: 8px;
    padding: 0.75rem 1rem;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 0.4rem;

    &.filter-info-dark {
        background: rgba(255, 255, 255, 0.06);
    }

    .info-row {
        display: flex;
        align-items: center;
        color: #64748b;

        .body--dark & {
            color: #94a3b8;
        }
    }
}
</style>
