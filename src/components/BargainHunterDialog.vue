<template>
    <q-dialog
        :model-value="show"
        @hide="$emit('close')"
        persistent
        maximized
        transition-show="slide-up"
        transition-hide="slide-down"
    >
        <q-card
            class="bargain-hunter-dialog"
        >
            <!-- Dialog chrome: header (icon/title/close), the same convention used by
                 PlayerDetailDialog/UpgradeFinderDialog — an icon, a title, then a close
                 button, all in normal flow. -->
            <div class="dialog-chrome">
                <div class="dialog-chrome__header">
                    <q-icon name="shopping_cart" class="dialog-chrome__icon" />
                    <div class="dialog-chrome__title">
                        Bargain Hunter (Values in {{ currencySymbol }})
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
                <!-- Filter Section with Sliders -->
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-3">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Max Transfer Value:
                            {{
                                maxBudget === maxBudgetSliderMax
                                    ? "Any"
                                    : formatCurrency(
                                          maxBudget * 1000000,
                                          currencySymbol,
                                      )
                            }}
                        </div>
                        <q-slider
                            v-model="maxBudget"
                            :min="maxBudgetSliderMin"
                            :max="maxBudgetSliderMax"
                            :step="maxBudgetSliderStep"
                            label-always
                            :label-value="
                                maxBudget === maxBudgetSliderMax
                                    ? 'Any'
                                    : formatCurrency(
                                          maxBudget * 1000000,
                                          currencySymbol,
                                      )
                            "
                            @update:model-value="onFiltersChanged"
                            color="primary"
                            class="q-px-sm"
                            :disable="loading"
                        />
                    </div>

                    <div class="col-12 col-md-3">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Max Weekly Wage:
                            {{
                                maxSalary === maxSalarySliderMax
                                    ? "Any"
                                    : formatCurrency(
                                          maxSalary * 1000,
                                          currencySymbol,
                                      )
                            }}
                        </div>
                        <q-slider
                            v-model="maxSalary"
                            :min="maxSalarySliderMin"
                            :max="maxSalarySliderMax"
                            :step="maxSalarySliderStep"
                            label-always
                            :label-value="
                                maxSalary === maxSalarySliderMax
                                    ? 'Any'
                                    : formatCurrency(
                                          maxSalary * 1000,
                                          currencySymbol,
                                      )
                            "
                            @update:model-value="onFiltersChanged"
                            color="primary"
                            class="q-px-sm"
                            :disable="loading"
                        />
                    </div>

                    <div class="col-12 col-md-3">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Age Range: {{ ageRange.min }} - 
                            {{
                                ageRange.max === ageSliderMax
                                    ? ageSliderMax + "+"
                                    : ageRange.max
                            }}
                        </div>
                        <q-range
                            v-model="ageRange"
                            :min="ageSliderMin"
                            :max="ageSliderMax"
                            :step="1"
                            label-always
                            :left-label-value="ageRange.min + ' yrs'"
                            :right-label-value="
                                ageRange.max +
                                (ageRange.max === ageSliderMax ? '+' : '') +
                                ' yrs'
                            "
                            @update:model-value="onFiltersChanged"
                            color="primary"
                            class="q-px-sm"
                            :disable="loading"
                        />
                    </div>

                    <div class="col-12 col-md-3">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Min Overall:
                            <span
                                class="stat-value-badge q-ml-xs"
                                :class="getUnifiedRatingClass(minOverall, 100)"
                            >
                                {{ minOverall || 0 }}
                            </span>
                        </div>
                        <q-slider
                            v-model="minOverall"
                            :min="minOverallSliderMin"
                            :max="minOverallSliderMax"
                            :step="1"
                            color="primary"
                            class="q-px-sm"
                            @update:model-value="onFiltersChanged"
                            :disable="loading"
                        />
                    </div>
                </div>

                <!-- Search Button and Value Tier Selection -->
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-3 search-button">
                        <q-btn 
                            color="primary" 
                            icon="search" 
                            label="Find Bargains" 
                            @click="findBargains"
                            :loading="loading"
                            class="full-width"
                            size="md"
                        />
                    </div>
                    
                    <div class="col-12 col-md-9">
                        <div class="text-subtitle2 q-mb-sm">
                            <q-icon name="filter_alt" class="q-mr-sm" />
                            Show Value Score Ranges
                        </div>
                        <div class="row q-col-gutter-sm value-tier-buttons">
                            <div class="col-auto">
                                <q-btn
                                    :color="showExcellentValue ? 'positive' : 'grey-5'"
                                    :outline="!showExcellentValue"
                                    size="sm"
                                    label="Excellent (80-100)"
                                    @click="toggleValueTier('excellent')"
                                    class="q-px-sm"
                                />
                            </div>
                            <div class="col-auto">
                                <q-btn
                                    :color="showGreatValue ? 'positive' : 'grey-5'"
                                    :outline="!showGreatValue"
                                    size="sm"
                                    label="Great (60-79)"
                                    @click="toggleValueTier('great')"
                                    class="q-px-sm"
                                />
                            </div>
                            <div class="col-auto">
                                <q-btn
                                    :color="showGoodValue ? 'orange' : 'grey-5'"
                                    :outline="!showGoodValue"
                                    size="sm"
                                    label="Good (40-59)"
                                    @click="toggleValueTier('good')"
                                    class="q-px-sm"
                                />
                            </div>
                            <div class="col-auto">
                                <q-btn
                                    :color="showMediocreValue ? 'warning' : 'grey-5'"
                                    :outline="!showMediocreValue"
                                    size="sm"
                                    label="Fair (20-39)"
                                    @click="toggleValueTier('mediocre')"
                                    class="q-px-sm"
                                />
                            </div>
                            <div class="col-auto">
                                <q-btn
                                    :color="showPoorValue ? 'grey-7' : 'grey-5'"
                                    :outline="!showPoorValue"
                                    size="sm"
                                    label="Poor (0-19)"
                                    @click="toggleValueTier('poor')"
                                    class="q-px-sm"
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <q-separator class="q-my-md" />

                <!-- Loading State -->
                <div v-if="loading" class="text-center q-my-xl">
                    <q-spinner-dots color="primary" size="3em" />
                    <div class="text-h6 q-mt-md">
                        Finding the best bargains...
                    </div>
                    <div class="text-caption q-mt-sm text-grey-6">
                        Analyzing player values and calculating bargain scores...
                    </div>
                </div>

                <!-- Results Table -->
                <div v-if="filteredBargainResults.length > 0 && !loading">
                    <SectionCard title="Best Value Players" icon="list" class="bargain-table-container">
                        <template #actions>
                            <q-chip
                                color="primary"
                                text-color="white"
                                :label="`${filteredBargainResults.length} players`"
                            />
                        </template>

                        <PlayerDataTable
                            :players="playersForTable"
                            :loading="loading"
                            @player-selected="handlePlayerSelected"
                            @team-selected="handleTeamSelected"
                            :currency-symbol="currencySymbol"
                            :dataset-id="datasetId"
                            :show-value-score="true"
                            :default-sort-field="'valueScore'"
                            :default-sort-direction="'desc'"
                        />
                    </SectionCard>
                </div>

                <!-- Empty State -->
                <EmptyState
                    v-else-if="!loading"
                    icon="search_off"
                    :title="bargainResults.length === 0 ? 'No bargains found' : 'No players match current filters'"
                    :description="bargainResults.length === 0 ? 'Try adjusting your budget or age criteria' : 'Try enabling more value score ranges or adjusting your filters'"
                />

                <!-- Help Section (Expandable) -->
                <q-expansion-item
                    v-if="!loading"
                    icon="help_outline"
                    label="How Value Scores Work"
                    class="q-mt-md"
                    :class="qInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
                >
                    <q-card 
                        flat 
                        :class="qInstance.dark.isActive ? 'bg-grey-8' : 'bg-blue-1'"
                    >
                        <q-card-section>
                            <div class="text-body2">
                                <div class="q-mb-sm"><strong>Value Score Formula:</strong></div>
                                <ul class="q-pl-md">
                                    <li><strong>Elite (80+ rating):</strong> High efficiency with premium pricing tolerance</li>
                                    <li><strong>Quality (70-79 rating):</strong> Balanced efficiency and value expectation</li>
                                    <li><strong>Decent (60-69 rating):</strong> Good value for money expected</li>
                                    <li><strong>Budget (55-59 rating):</strong> Lower cost expected</li>
                                    <li><strong>Youth (<55 rating):</strong> Development potential focus</li>
                                </ul>
                                <div class="q-mt-sm text-caption text-grey-6">
                                    <em>Uses logarithmic scaling • Bonuses for exceptional value (30%+ below expected) • 
                                    Free transfers excluded • Scores normalized (0-100) • Limited to top 500 players</em>
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>
                </q-expansion-item>
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
import { formatCurrency } from '../utils/currencyUtils'
import EmptyState from './layout/EmptyState.vue'
import SectionCard from './layout/SectionCard.vue'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

export default defineComponent({
  name: 'BargainHunterDialog',
  components: {
    PlayerDetailDialog,
    PlayerDataTable,
    SectionCard,
    EmptyState,
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
  },
  emits: ['close'],
  setup(props) {
    const qInstance = useQuasar()

    // Constants matching PlayerFilters
    const AGE_SLIDER_MIN = 15
    const AGE_SLIDER_MAX = 50
    const MAX_BUDGET_SLIDER_MIN = 0
    const MAX_BUDGET_SLIDER_MAX = 500 // 500M
    const MAX_SALARY_SLIDER_MIN = 0
    const MAX_SALARY_SLIDER_MAX = 1000 // 1000K per week
    const MIN_OVERALL_SLIDER_MIN = 40
    const MIN_OVERALL_SLIDER_MAX = 99

    // State - converted to match slider values
    const maxBudget = ref(MAX_BUDGET_SLIDER_MAX) // In millions
    const maxSalary = ref(MAX_SALARY_SLIDER_MAX) // In thousands per week
    const ageRange = ref({ min: AGE_SLIDER_MIN, max: 27 }) // Age range object
    const minOverall = ref(75)
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const bargainResults = ref([])
    const showExcellentValue = ref(true)
    const showGreatValue = ref(true)
    const showGoodValue = ref(true)
    const showMediocreValue = ref(false)
    const showPoorValue = ref(false)

    // Slider configuration computed properties
    const ageSliderMin = AGE_SLIDER_MIN
    const ageSliderMax = AGE_SLIDER_MAX
    const maxBudgetSliderMin = MAX_BUDGET_SLIDER_MIN
    const maxBudgetSliderMax = MAX_BUDGET_SLIDER_MAX
    const maxSalarySliderMin = MAX_SALARY_SLIDER_MIN
    const maxSalarySliderMax = MAX_SALARY_SLIDER_MAX
    const minOverallSliderMin = MIN_OVERALL_SLIDER_MIN
    const minOverallSliderMax = MIN_OVERALL_SLIDER_MAX

    // Step calculations for sliders
    const maxBudgetSliderStep = computed(() => {
      const range = maxBudgetSliderMax - maxBudgetSliderMin
      if (range <= 0) return 1
      if (range < 50) return 0.5
      if (range < 250) return 1
      return 5
    })

    const maxSalarySliderStep = computed(() => {
      const range = maxSalarySliderMax - maxSalarySliderMin
      if (range <= 0) return 1
      if (range < 50) return 0.5
      if (range < 250) return 1
      if (range < 1000) return 5
      return 10
    })

    // Utility function to get unified rating class (matching PlayerFilters)
    const getUnifiedRatingClass = (value, maxScale = 100) => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue) || value === null || value === undefined || value === '-')
        return 'rating-na'

      const percentage = (numValue / maxScale) * 100

      if (maxScale === 20) {
        if (numValue >= 18) return 'rating-tier-6'
        if (numValue >= 15) return 'rating-tier-5'
        if (numValue >= 13) return 'rating-tier-4'
        if (numValue >= 10) return 'rating-tier-3'
        if (numValue >= 7) return 'rating-tier-2'
        if (numValue >= 1) return 'rating-tier-1'
      } else {
        if (percentage >= 90) return 'rating-tier-6'
        if (percentage >= 80) return 'rating-tier-5'
        if (percentage >= 70) return 'rating-tier-4'
        if (percentage >= 55) return 'rating-tier-3'
        if (percentage >= 40) return 'rating-tier-2'
        if (percentage > 0) return 'rating-tier-1'
      }
      return 'rating-na'
    }

    // Computed property to transform bargain results for PlayerDataTable
    const playersForTable = computed(() => {
      return filteredBargainResults.value.map((result) => ({
        ...result.player,
        valueScore: result.valueScore,
      }))
    })

    // Filtered results for table display
    const filteredBargainResults = computed(() => {
      return bargainResults.value.filter((result) => {
        const score = result.valueScore
        if (score >= 80 && showExcellentValue.value) return true
        if (score >= 60 && score < 80 && showGreatValue.value) return true
        if (score >= 40 && score < 60 && showGoodValue.value) return true
        if (score >= 20 && score < 40 && showMediocreValue.value) return true
        if (score < 20 && showPoorValue.value) return true
        return false
      })
    })

    // Methods
    const findBargains = async () => {
      if (!props.datasetId) {
        return
      }

      loading.value = true

      try {
        // Prepare request payload
        const requestBody = {
          maxBudget: maxBudget.value ? maxBudget.value * 1000000 : 0, // Convert to actual amount
          maxSalary: maxSalary.value ? maxSalary.value * 1000 : 0, // Convert to actual amount
          minAge: ageRange.value.min || 0,
          maxAge: ageRange.value.max || 0,
          minOverall: minOverall.value || 0,
        }

        // Call backend API
        const response = await fetch(`/api/bargain-hunter/${props.datasetId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestBody),
        })

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }

        const bargainData = await response.json()
        bargainResults.value = bargainData || []
      } catch (_error) {
        qInstance.notify({
          message: 'Error finding bargains. Please try again.',
          color: 'negative',
          icon: 'error',
        })
      } finally {
        loading.value = false
      }
    }

    const onFiltersChanged = async () => {
      // Debounce the search to avoid too many calls
      clearTimeout(onFiltersChanged.timeoutId)
      onFiltersChanged.timeoutId = setTimeout(() => {
        if (props.show && props.datasetId) {
          findBargains()
        }
      }, 300)
    }

    const handlePlayerSelected = (player) => {
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const handleTeamSelected = (_teamName) => {
      // For bargain hunter, we don't need team selection functionality
      // but we need to provide the handler for PlayerDataTable compatibility
    }

    const _formatValueScore = (score) => {
      if (typeof score !== 'number') return '0'
      return Math.round(score).toString()
    }

    const toggleValueTier = (tier) => {
      if (tier === 'excellent') {
        showExcellentValue.value = !showExcellentValue.value
      } else if (tier === 'great') {
        showGreatValue.value = !showGreatValue.value
      } else if (tier === 'good') {
        showGoodValue.value = !showGoodValue.value
      } else if (tier === 'mediocre') {
        showMediocreValue.value = !showMediocreValue.value
      } else if (tier === 'poor') {
        showPoorValue.value = !showPoorValue.value
      }

      // No need to trigger a new backend search for value tier changes
      // The filteredBargainResults computed property will handle the filtering
    }

    // Watchers
    watch(
      () => props.show,
      async (newShow) => {
        if (newShow && props.datasetId) {
          // Auto-search when dialog opens
          await findBargains()
        } else if (!newShow) {
          // Reset values when dialog closes
          maxBudget.value = maxBudgetSliderMax
          maxSalary.value = maxSalarySliderMax
          ageRange.value = { min: ageSliderMin, max: 27 }
          minOverall.value = 75
          bargainResults.value = []
          showExcellentValue.value = true
          showGreatValue.value = true
          showGoodValue.value = true
          showMediocreValue.value = false
          showPoorValue.value = false
        }
      }
    )

    // Initialize when component mounts
    onMounted(() => {
      if (props.show && props.datasetId) {
        findBargains()
      }
    })

    return {
      qInstance,
      maxBudget,
      maxSalary,
      ageRange,
      minOverall,
      loading,
      selectedPlayer,
      showPlayerDetail,
      bargainResults,
      onFiltersChanged,
      handlePlayerSelected,
      handleTeamSelected,
      formatCurrency,
      findBargains,
      showExcellentValue,
      showGreatValue,
      showGoodValue,
      showMediocreValue,
      showPoorValue,
      filteredBargainResults,
      toggleValueTier,
      playersForTable,
      ageSliderMin,
      ageSliderMax,
      maxBudgetSliderMin,
      maxBudgetSliderMax,
      maxSalarySliderMin,
      maxSalarySliderMax,
      minOverallSliderMin,
      minOverallSliderMax,
      maxBudgetSliderStep,
      maxSalarySliderStep,
      getUnifiedRatingClass,
    }
  },
})
</script>

<style lang="scss" scoped>
.bargain-hunter-dialog {
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-3);
    background: var(--surface-card);
}

// Dialog chrome: unified header convention shared with PlayerDetailDialog /
// UpgradeFinderDialog — icon, title, actions, close, all in normal flow.
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

.bargain-hunter-dialog {
    // Slider styling
    .slider-label {
        font-weight: 500;
        color: var(--text-secondary);
        margin-bottom: 0.5rem;
    }

    :deep(.q-slider) {
        .q-slider__track-container {
            .q-slider__track {
                background: var(--accent-soft-strong);
            }

            .q-slider__selection {
                background: var(--accent);
            }
        }

        .q-slider__thumb {
            background: var(--accent);
            border: 2px solid var(--surface-card);
            box-shadow: 0 2px 8px color-mix(in srgb, var(--accent) 40%, transparent);
        }
    }

    :deep(.q-range) {
        .q-slider__track-container {
            .q-slider__track {
                background: var(--accent-soft-strong);
            }

            .q-slider__selection {
                background: var(--accent);
            }
        }

        .q-slider__thumb {
            background: var(--accent);
            border: 2px solid var(--surface-card);
            box-shadow: 0 2px 8px color-mix(in srgb, var(--accent) 40%, transparent);
        }
    }

    // Button styling
    .q-btn {
        border-radius: 8px;
        font-weight: 500;
        text-transform: none;

        &.q-btn--unelevated {
            box-shadow: var(--shadow-1);

            &:hover {
                box-shadow: var(--shadow-2);
                transform: translateY(-1px);
            }
        }

        &.q-btn--outline {
            border-width: 2px;

            &:hover {
                background-color: var(--accent-soft);
            }
        }
    }

    // Input field styling
    :deep(.q-field) {
        .q-field__control {
            border-radius: 8px;

            &:before {
                border-color: var(--surface-border-strong);
            }

            &:hover:before {
                border-color: var(--accent);
            }
        }

        &.q-field--outlined {
            .q-field__control {
                box-shadow: var(--shadow-1);
                transition: all 0.2s ease;

                &:hover {
                    box-shadow: 0 2px 6px color-mix(in srgb, var(--accent) 10%, transparent);
                }
            }
        }

        &.q-field--focused {
            .q-field__control {
                box-shadow: 0 0 0 2px var(--accent-soft-strong);
            }
        }
    }

    // Badge styling
    .stat-value-badge {
        padding: 2px 8px;
        border-radius: 4px;
        font-size: 0.75rem;
        font-weight: 600;
        
        &.rating-tier-6 {
            background-color: #10b981;
            color: white;
        }
        
        &.rating-tier-5 {
            background-color: #059669;
            color: white;
        }
        
        &.rating-tier-4 {
            background-color: #f59e0b;
            color: white;
        }
        
        &.rating-tier-3 {
            background-color: #ef4444;
            color: white;
        }
        
        &.rating-tier-2 {
            background-color: #dc2626;
            color: white;
        }
        
        &.rating-tier-1 {
            background-color: #b91c1c;
            color: white;
        }
        
        &.rating-na {
            background-color: #6b7280;
            color: white;
        }
    }

    // Value tier buttons
    .value-tier-buttons {
        .q-btn {
            font-weight: 500;
            text-transform: none;
            border-radius: 6px;

            &:not(.q-btn--outline) {
                box-shadow: var(--shadow-1);
            }

            &.q-btn--outline {
                border-width: 2px;
                opacity: 0.7;

                &:hover {
                    opacity: 1;
                }
            }
        }
    }

    // Search button
    .search-button {
        .q-btn {
            border-radius: 8px;
            font-weight: 600;
            text-transform: none;
            padding: 12px 24px;
        }
    }

    // Enhanced table styling
    :deep(.q-table) {
        border-radius: 8px;
        overflow: hidden;

        .q-table__top {
            padding: 1rem;
            background: var(--accent-soft);
        }

        .q-table__container {
            border-radius: 0 0 8px 8px;
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
    }

    // Responsive design
    @media (max-width: 768px) {
        .dialog-chrome__header {
            padding: 10px 12px;
        }

        .dialog-chrome__title {
            font-size: 1rem;
        }

        .value-tier-buttons {
            .q-btn {
                font-size: 0.75rem;
                padding: 4px 8px;
            }
        }
    }

    @media (max-width: 480px) {
        .q-card-section {
            padding: 12px;
        }
        
        .value-tier-buttons {
            .col-auto {
                width: 100%;
                
                .q-btn {
                    width: 100%;
                    margin-bottom: 4px;
                }
            }
        }
    }
}
</style> 