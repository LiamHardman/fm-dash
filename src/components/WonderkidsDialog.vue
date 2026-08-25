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
            <!-- Dialog chrome: header (icon/title/close), the same convention used by
                 PlayerDetailDialog/UpgradeFinderDialog — an icon, a title, then a close
                 button, all in normal flow. -->
            <div class="dialog-chrome">
                <div class="dialog-chrome__header">
                    <q-icon name="stars" class="dialog-chrome__icon" />
                    <div class="dialog-chrome__title">
                        Find Wonderkids (Values in {{ currencySymbol }})
                    </div>
                    <q-space />
                    <div class="dialog-chrome__actions">
                        <q-btn
                            icon="close"
                            flat
                            round
                            dense
                            class="dialog-chrome__close"
                            v-close-popup
                            @click="$emit('close')"
                        />
                    </div>
                </div>
            </div>

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

                    <!-- Age / Ability Scatter Plot -->
                    <SectionCard class="wonderkids-chart-container q-mb-md">
                        <WonderkidsScatterPlot
                            :players="allWonderkids"
                            :is-dark-mode="qInstance.dark.isActive"
                            @player-click="handlePlayerSelected"
                        />
                    </SectionCard>

                    <SectionCard class="wonderkids-table-container">
                        <PlayerDataTable
                            :key="'wonderkids-all-ages'"
                            :players="allWonderkids"
                            :loading="loading"
                            @player-selected="handlePlayerSelected"
                            @team-selected="handleTeamSelected"
                            :currency-symbol="currencySymbol"
                            :dataset-id="datasetId"
                        />
                    </SectionCard>
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
import SectionCard from './layout/SectionCard.vue'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'
import WonderkidsScatterPlot from './WonderkidsScatterPlot.vue'

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
  'ST',
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
    PlayerDetailDialog,
    WonderkidsScatterPlot,
    SectionCard,
  },
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    players: {
      type: Array,
      default: () => [],
    },
    currencySymbol: {
      type: String,
      default: '$',
    },
    datasetId: {
      type: String,
      default: null,
    },
    transferValueRange: {
      type: Object,
      default: () => ({ min: 0, max: 100000000 }),
    },
    salaryRange: {
      type: Object,
      default: () => ({ min: 0, max: 1000000 }),
    },
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
    const getAllWonderkids = (allPlayers) => {
      // Get top 10 players from each age (15-21) that meet the filter criteria
      const allWonderkidsFromAllAges = []

      for (const age of ages) {
        const playersOfAge = allPlayers
          .filter((player) => {
            const playerAge = Number(player.age)
            // Must match this specific age
            if (playerAge !== age) {
              return false
            }

            // Apply position filter
            if (selectedPosition.value) {
              const playerPositions = player.short_positions || player.shortPositions || []
              if (!playerPositions.includes(selectedPosition.value)) {
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
          icon: 'error',
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

    const handlePlayerSelected = (player) => {
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const handleTeamSelected = (_teamName) => {}

    // Initialize sliders when props change
    watch(
      () => props.transferValueRange,
      (newRange) => {
        if (newRange && maxTransferValue.value === 100000000) {
          maxTransferValue.value = newRange.max
        }
      },
      { immediate: true }
    )

    watch(
      () => props.salaryRange,
      (newRange) => {
        if (newRange && maxSalary.value === 1000000) {
          maxSalary.value = newRange.max
        }
      },
      { immediate: true }
    )

    // Watchers
    watch(
      () => props.show,
      (newShow) => {
        if (newShow && props.players.length > 0) {
          loadWonderkids()
        }
      }
    )

    watch(
      () => props.players,
      (newPlayers) => {
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
      formatCurrency,
    }
  },
})
</script>

<style lang="scss" scoped>
.wonderkids-dialog {
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-3);
    background: var(--surface-card);
}

// Dialog chrome: unified header convention shared with PlayerDetailDialog /
// UpgradeFinderDialog - icon, title, actions, close, all in normal flow.
.dialog-chrome {
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    background: var(--surface-raised);
    border-bottom: 1px solid var(--surface-border);
}

.dialog-chrome__header {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 12px var(--density-card-padding, 16px);
}

.dialog-chrome__icon {
    font-size: 1.3rem;
    color: var(--accent);
    flex-shrink: 0;
}

.dialog-chrome__title {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.dialog-chrome__actions {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
}

.dialog-chrome__close {
    transition: transform 0.15s ease;

    &:hover {
        transform: scale(1.08);
    }
}

.wonderkids-dialog {
    // Slider styling
    .slider-label {
        color: var(--text-primary);
        font-weight: 600;
        font-size: 0.9rem;
        margin-bottom: 0.5rem;
    }

    :deep(.q-slider) {
        .q-slider__track {
            background: var(--accent-soft-strong);
        }

        .q-slider__track-active {
            background: var(--accent);
        }

        .q-slider__thumb {
            background: var(--accent);
            border: 2px solid var(--surface-card);
            box-shadow: var(--shadow-1);
        }
    }

    // Select field styling
    :deep(.q-select) {
        .q-field__control {
            border-radius: 8px;
            background: var(--accent-soft);
            border: 1px solid var(--surface-border);
            transition: all 0.2s ease;

            &:hover {
                border-color: var(--surface-border-strong);
                background: var(--accent-soft-strong);
            }
        }

        .q-field__native,
        .q-field__input {
            color: var(--text-primary);
            font-weight: 500;
        }

        .q-field__label {
            color: var(--text-secondary);
            font-weight: 600;
        }
    }

    // Separator styling
    .q-separator {
        background-color: var(--surface-border);
    }

    // Loading state styling
    .loading-state {
        .q-spinner-dots {
            color: var(--accent);
        }

        .text-caption {
            color: var(--text-secondary);
        }
    }

    // Table container styling
    .wonderkids-table-container {
        :deep(.q-table) {
            border-radius: var(--radius-md);
            overflow: hidden;

            .q-table__top {
                padding: 1rem;
                background: var(--accent-soft);
            }

            .q-table__container {
                border-radius: 0 0 var(--radius-md) var(--radius-md);
            }

            thead {
                th {
                    background: var(--surface-raised);
                    color: var(--text-secondary);
                    font-weight: 600;
                    border-bottom: 2px solid var(--surface-border-strong);
                }
            }

            tbody {
                tr {
                    border-bottom: 1px solid var(--surface-border);

                    &:hover {
                        background-color: var(--accent-soft);
                    }
                }
            }

            // Auto height management — reduced to account for the scatter chart above
            height: auto !important;
            max-height: max(200px, calc(100vh - 660px));

            .q-table__middle {
                max-height: max(200px, calc(100vh - 660px));
            }
        }
    }

    // Subtitle styling
    .text-subtitle1 {
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 1rem;

        .text-caption {
            font-weight: 400;
            color: var(--text-secondary);
        }
    }

    // Responsive design
    @media (max-width: 768px) {
        .dialog-chrome__header {
            padding: 10px 12px;
        }

        .dialog-chrome__title {
            font-size: 1rem;
        }

        .q-card-section {
            padding: 1rem;

            &:last-child {
                padding-bottom: 1rem;
            }
        }

        .wonderkids-table-container {
            :deep(.q-table) {
                max-height: max(150px, calc(100vh - 620px));

                .q-table__middle {
                    max-height: max(150px, calc(100vh - 620px));
                }
            }
        }
    }

    @media (max-width: 480px) {
        .dialog-chrome__header {
            padding: 8px 8px;
            gap: 0.4rem;
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
                max-height: max(150px, calc(100vh - 580px));

                .q-table__middle {
                    max-height: max(150px, calc(100vh - 580px));
                }
            }
        }
    }
}
</style> 