<template>
    <q-dialog
        :model-value="show"
        @update:model-value="$emit('close')"
        persistent
        maximized
        transition-show="slide-up"
        transition-hide="slide-down"
    >
        <q-card
            class="wonderkids-dialog"
        >
            <q-card-section
                class="row items-center q-pb-none card-header"
            >
                <q-icon name="stars" size="md" class="q-mr-sm" />
                <div class="text-h6">
                    Find Wonderkids (Values in {{ currencySymbol }})
                </div>
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
                <!-- Filters Section -->
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-4">
                        <q-select
                            v-model="selectedPosition"
                            :options="positionOptions"
                            label="Position"
                            dense
                            filled
                            clearable
                            emit-value
                            map-options
                            @update:model-value="onFiltersChanged"
                            behavior="menu"
                            :disable="loading"
                        />
                    </div>
                    <div class="col-12 col-md-4">
                        <div class="text-caption q-mb-xs slider-label">
                            Max Transfer Value:
                            {{
                                maxTransferValue === transferValueSliderMax
                                    ? "Any"
                                    : formatCurrency(maxTransferValue, currencySymbol)
                            }}
                        </div>
                        <q-slider
                            v-model="maxTransferValue"
                            :min="transferValueSliderMin"
                            :max="transferValueSliderMax"
                            :step="transferValueSliderStep"
                            label-always
                            :label-value="
                                maxTransferValue === transferValueSliderMax
                                    ? 'Any'
                                    : formatCurrency(maxTransferValue, currencySymbol)
                            "
                            @update:model-value="debouncedFiltersChanged"
                            color="primary"
                            class="q-px-sm"
                            :disable="loading || !isDataAvailable"
                        />
                    </div>
                    <div class="col-12 col-md-4">
                        <div class="text-caption q-mb-xs slider-label">
                            Max Salary:
                            {{
                                maxSalary === salarySliderMax
                                    ? "Any"
                                    : formatCurrency(maxSalary, currencySymbol)
                            }}
                        </div>
                        <q-slider
                            v-model="maxSalary"
                            :min="salarySliderMin"
                            :max="salarySliderMax"
                            :step="salarySliderStep"
                            label-always
                            :label-value="
                                maxSalary === salarySliderMax
                                    ? 'Any'
                                    : formatCurrency(maxSalary, currencySymbol)
                            "
                            @update:model-value="debouncedFiltersChanged"
                            color="primary"
                            class="q-px-sm"
                            :disable="loading || !isDataAvailable"
                        />
                    </div>
                </div>

                <q-separator class="q-my-md" />

                <!-- Loading State -->
                <div v-if="loading" class="text-center q-my-xl">
                    <q-spinner-dots color="primary" size="3em" />
                    <div
                        class="q-mt-md text-caption"
                        :class="
                            qInstance.dark.isActive
                                ? 'text-grey-5'
                                : 'text-grey-7'
                        "
                    >
                        Finding wonderkids...
                    </div>
                </div>

                <!-- Player Data Table -->
                <div v-else>
                    <div class="text-subtitle1 q-mb-md">
                        Top {{ allWonderkids.length }} wonderkids aged 15-21
                        <span v-if="hasActiveFilters" class="text-caption text-grey-6">
                            (filtered by {{ filterSummary }})
                        </span>
                    </div>
                    
                    <q-card
                        class="wonderkids-table-container"
                    >
                        <q-card-section>
                            <PlayerDataTable
                                :key="'wonderkids-all-ages'"
                                :players="allWonderkids"
                                :loading="loading"
                                @player-selected="handlePlayerSelected"
                                @team-selected="handleTeamSelected"
                                :currency-symbol="currencySymbol"
                                :dataset-id="datasetId"
                            />
                        </q-card-section>
                    </q-card>
                </div>
            </q-card-section>
        </q-card>

        <!-- Player Detail Dialog -->
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
import { computed, defineComponent, onMounted, ref, watch } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import { formatCurrency } from '../utils/currencyUtils'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

// Position options matching PlayerFilters
const orderedShortPositions = [
  'GK',
  'DR',
  'DC',
  'DL',
  'WBR',
  'WBL',
  'DM',
  'MR',
  'MC',
  'ML',
  'AMR',
  'AMC',
  'AML',
  'ST'
]

// Debounce utility
function debounce(fn, delay) {
  let timeoutID = null
  return function (...args) {
    clearTimeout(timeoutID)
    timeoutID = setTimeout(() => {
      fn.apply(this, args)
    }, delay)
  }
}

export default defineComponent({
  name: 'WonderkidsDialog',
  components: {
    PlayerDataTable,
    PlayerDetailDialog
  },
  props: {
    show: {
      type: Boolean,
      default: false
    },
    players: {
      type: Array,
      default: () => []
    },
    currencySymbol: {
      type: String,
      default: '$'
    },
    datasetId: {
      type: String,
      default: null
    },
    transferValueRange: {
      type: Object,
      default: () => ({ min: 0, max: 100000000 })
    },
    salaryRange: {
      type: Object,
      default: () => ({ min: 0, max: 1000000 })
    }
  },
  emits: ['close'],
  setup(props) {
    const qInstance = useQuasar()
    const _playerStore = usePlayerStore()

    // State
    const selectedPosition = ref(null)
    const maxTransferValue = ref(100000000)
    const maxSalary = ref(1000000)
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const wonderkidsData = ref([])

    // Constants
    const ages = [15, 16, 17, 18, 19, 20, 21]

    // Computed values for sliders
    const transferValueSliderMin = computed(() => props.transferValueRange?.min || 0)
    const transferValueSliderMax = computed(() => props.transferValueRange?.max || 100000000)
    const salarySliderMin = computed(() => props.salaryRange?.min || 0)
    const salarySliderMax = computed(() => props.salaryRange?.max || 1000000)

    const isDataAvailable = computed(() => props.players && props.players.length > 0)

    const transferValueSliderStep = computed(() => {
      const range = transferValueSliderMax.value - transferValueSliderMin.value
      if (range <= 0) return 10000
      if (range < 50000) return 1000
      if (range < 250000) return 5000
      if (range < 1000000) return 10000
      if (range < 10000000) return 50000
      if (range < 50000000) return 100000
      return 250000
    })

    const salarySliderStep = computed(() => {
      const range = salarySliderMax.value - salarySliderMin.value
      if (range <= 0) return 1000
      if (range < 50000) return 500
      if (range < 250000) return 2500
      if (range < 1000000) return 5000
      if (range < 10000000) return 25000
      return 50000
    })

    const positionOptions = computed(() => {
      const options = [{ label: 'Any Position', value: null }]
      for (const shortPos of orderedShortPositions) {
        options.push({ label: shortPos, value: shortPos })
      }
      return options
    })

    const hasActiveFilters = computed(() => {
      return (
        selectedPosition.value !== null ||
        maxTransferValue.value < transferValueSliderMax.value ||
        maxSalary.value < salarySliderMax.value
      )
    })

    const filterSummary = computed(() => {
      const filters = []
      if (selectedPosition.value) filters.push(`position: ${selectedPosition.value}`)
      if (maxTransferValue.value < transferValueSliderMax.value) filters.push('transfer value')
      if (maxSalary.value < salarySliderMax.value) filters.push('salary')
      return filters.join(', ')
    })

    // Computed - top 10 across all ages
    const allWonderkids = computed(() => {
      return wonderkidsData.value
    })

    // Methods
    const getAllWonderkids = allPlayers => {
      // Get top 10 players from each age (15-21) that meet the filter criteria
      const allWonderkidsFromAllAges = []

      for (const age of ages) {
        const playersOfAge = allPlayers
          .filter(player => {
            const playerAge = Number(player.age)
            // Must match this specific age
            if (playerAge !== age) {
              return false
            }

            // Apply position filter
            if (selectedPosition.value) {
              if (!player.shortPositions?.includes(selectedPosition.value)) {
                return false
              }
            }

            // Apply transfer value filter
            if (
              maxTransferValue.value < transferValueSliderMax.value &&
              player.transferValueAmount > maxTransferValue.value
            ) {
              return false
            }

            // Apply salary filter
            if (maxSalary.value < salarySliderMax.value && player.wageAmount > maxSalary.value) {
              return false
            }

            return true
          })
          .sort((a, b) => (b.Overall || 0) - (a.Overall || 0))
          .slice(0, 10) // Take top 10 for this age

        // Add all top 10 players from this age to the combined array
        allWonderkidsFromAllAges.push(...playersOfAge)
      }

      // Sort the final combined array by overall rating (best first)
      return allWonderkidsFromAllAges.sort((a, b) => (b.Overall || 0) - (a.Overall || 0))
    }

    const loadWonderkids = async () => {
      loading.value = true
      try {
        const allPlayers = props.players
        wonderkidsData.value = getAllWonderkids(allPlayers)
      } catch (_error) {
        qInstance.notify({
          color: 'negative',
          message: 'Failed to load wonderkids',
          icon: 'error'
        })
      } finally {
        loading.value = false
      }
    }

    const onFiltersChanged = () => {
      // Immediate update for position changes
      loadWonderkids()
    }

    const debouncedFiltersChanged = debounce(() => {
      loadWonderkids()
    }, 300)

    const handlePlayerSelected = player => {
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const handleTeamSelected = _teamName => {}

    // Initialize sliders when props change
    watch(
      () => props.transferValueRange,
      newRange => {
        if (newRange && maxTransferValue.value === 100000000) {
          maxTransferValue.value = newRange.max
        }
      },
      { immediate: true }
    )

    watch(
      () => props.salaryRange,
      newRange => {
        if (newRange && maxSalary.value === 1000000) {
          maxSalary.value = newRange.max
        }
      },
      { immediate: true }
    )

    // Watchers
    watch(
      () => props.show,
      newShow => {
        if (newShow && props.players.length > 0) {
          loadWonderkids()
        }
      }
    )

    watch(
      () => props.players,
      newPlayers => {
        if (props.show && newPlayers.length > 0) {
          loadWonderkids()
        }
      }
    )

    // Initialize when component mounts
    onMounted(() => {
      if (props.show && props.players.length > 0) {
        loadWonderkids()
      }
    })

    return {
      qInstance,
      selectedPosition,
      maxTransferValue,
      maxSalary,
      loading,
      selectedPlayer,
      showPlayerDetail,
      allWonderkids,
      positionOptions,
      transferValueSliderMin,
      transferValueSliderMax,
      transferValueSliderStep,
      salarySliderMin,
      salarySliderMax,
      salarySliderStep,
      isDataAvailable,
      hasActiveFilters,
      filterSummary,
      onFiltersChanged,
      debouncedFiltersChanged,
      handlePlayerSelected,
      handleTeamSelected,
      formatCurrency
    }
  }
})
</script>

<style lang="scss" scoped>
.wonderkids-dialog {
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    border: 1px solid rgba(0, 0, 0, 0.04);
    
    .body--dark & {
        background-color: #1e293b !important;
        border: 1px solid rgba(255, 255, 255, 0.1);
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    }

    .card-header {
        background: linear-gradient(135deg, #2e74b5 0%, #3b82c7 100%);
        color: white;
        padding: 1.5rem;
        border-radius: $border-radius $border-radius 0 0;
        
        .q-icon {
            color: rgba(255, 255, 255, 0.9);
        }
        
        .text-h6 {
            font-weight: 600;
            font-size: 1.25rem;
        }
        
        .q-btn {
            color: rgba(255, 255, 255, 0.8);
            
            &:hover {
                background-color: rgba(255, 255, 255, 0.1);
                color: white;
            }
        }
    }

    .q-card-section {
        &:not(.card-header) {
            background: transparent;
            
            .body--dark & {
                background: transparent;
            }
        }
    }

    // Slider styling
    .slider-label {
        color: #334155;
        font-weight: 600;
        font-size: 0.9rem;
        margin-bottom: 0.5rem;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.9);
        }
    }

    :deep(.q-slider) {
        .q-slider__track {
            background: rgba(46, 116, 181, 0.1);
            
            .body--dark & {
                background: rgba(255, 255, 255, 0.1);
            }
        }
        
        .q-slider__track-active {
            background: #2e74b5;
            
            .body--dark & {
                background: #60a5fa;
            }
        }
        
        .q-slider__thumb {
            background: #2e74b5;
            border: 2px solid white;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            
            .body--dark & {
                background: #60a5fa;
                border-color: rgba(255, 255, 255, 0.1);
            }
        }
    }

    // Select field styling
    :deep(.q-select) {
        .q-field__control {
            border-radius: 8px;
            background: rgba(46, 116, 181, 0.02);
            border: 1px solid rgba(46, 116, 181, 0.1);
            transition: all 0.2s ease;
            
            .body--dark & {
                background: rgba(96, 165, 250, 0.05);
                border: 1px solid rgba(255, 255, 255, 0.1);
            }
            
            &:hover {
                border-color: rgba(46, 116, 181, 0.2);
                background: rgba(46, 116, 181, 0.04);
                
                .body--dark & {
                    border-color: rgba(255, 255, 255, 0.2);
                    background: rgba(96, 165, 250, 0.08);
                }
            }
        }
        
        .q-field__native,
        .q-field__input {
            color: #334155;
            font-weight: 500;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.9);
            }
        }
        
        .q-field__label {
            color: #64748b;
            font-weight: 600;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.7);
            }
        }
    }

    // Separator styling
    .q-separator {
        background-color: rgba(0, 0, 0, 0.08);
        
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.1);
        }
    }

    // Loading state styling
    .loading-state {
        .q-spinner-dots {
            color: #2e74b5;
        }
        
        .text-caption {
            color: #6b7280;
            
            .body--dark & {
                color: #9ca3af;
            }
        }
    }

    // Table container styling
    .wonderkids-table-container {
        border-radius: $border-radius;
        box-shadow: $card-shadow;
        border: 1px solid rgba(0, 0, 0, 0.04);
        
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.02);
            border-color: rgba(255, 255, 255, 0.08);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
        }
        
        :deep(.q-table) {
            border-radius: $border-radius;
            overflow: hidden;
            
            .q-table__top {
                padding: 1rem;
                background: linear-gradient(135deg, rgba(46, 116, 181, 0.03) 0%, rgba(46, 116, 181, 0.01) 100%);
                
                .body--dark & {
                    background: rgba(255, 255, 255, 0.02);
                }
            }
            
            .q-table__container {
                border-radius: 0 0 $border-radius $border-radius;
            }
            
            thead {
                th {
                    background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
                    color: #374151;
                    font-weight: 600;
                    border-bottom: 2px solid #e5e7eb;
                    
                    .body--dark & {
                        background: linear-gradient(135deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.05) 100%);
                        color: #d1d5db;
                        border-bottom-color: rgba(255, 255, 255, 0.1);
                    }
                }
            }
            
            tbody {
                tr {
                    border-bottom: 1px solid #f3f4f6;
                    
                    &:hover {
                        background-color: rgba(46, 116, 181, 0.04);
                    }
                    
                    .body--dark & {
                        border-bottom-color: rgba(255, 255, 255, 0.05);
                        
                        &:hover {
                            background-color: rgba(255, 255, 255, 0.03);
                        }
                    }
                }
            }
            
            // Auto height management
            height: auto !important;
            max-height: calc(100vh - 350px);
            
            .q-table__middle {
                max-height: calc(100vh - 350px);
            }
        }
    }

    // Subtitle styling
    .text-subtitle1 {
        font-weight: 600;
        color: #374151;
        margin-bottom: 1rem;
        
        .body--dark & {
            color: #d1d5db;
        }
        
        .text-caption {
            font-weight: 400;
            color: #6b7280;
            
            .body--dark & {
                color: #9ca3af;
            }
        }
    }

    // Responsive design
    @media (max-width: 768px) {
        .card-header {
            padding: 1rem;
            
            .text-h6 {
                font-size: 1.1rem;
            }
        }
        
        .q-card-section {
            padding: 1rem;
            
            &:last-child {
                padding-bottom: 1rem;
            }
        }
        
        .wonderkids-table-container {
            :deep(.q-table) {
                max-height: calc(100vh - 300px);
                
                .q-table__middle {
                    max-height: calc(100vh - 300px);
                }
            }
        }
    }

    @media (max-width: 480px) {
        .card-header {
            padding: 0.75rem;
        }
        
        .q-card-section {
            padding: 0.75rem;
            
            &:last-child {
                padding-bottom: 0.75rem;
            }
        }
        
        .text-subtitle1 {
            font-size: 1rem;
        }
        
        .wonderkids-table-container {
            :deep(.q-table) {
                max-height: calc(100vh - 250px);
                
                .q-table__middle {
                    max-height: calc(100vh - 250px);
                }
            }
        }
    }
}
</style> 