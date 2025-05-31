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
            class="bargain-hunter-dialog"
            :class="qInstance.dark.isActive ? 'bg-dark' : 'bg-grey-1'"
        >
            <q-card-section
                class="row items-center q-pb-none"
                :class="
                    qInstance.dark.isActive
                        ? 'bg-grey-10 text-white'
                        : 'bg-primary text-white'
                "
            >
                <q-icon name="local_offer" size="md" class="q-mr-sm" />
                <div class="text-h6">
                    Bargain Hunter (Values in {{ currencySymbol }})
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
                    <div class="col-12 col-md-2">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Max Transfer Budget:
                            {{ maxBudget ? `${currencySymbol}${(maxBudget / 1000000).toFixed(1)}M` : 'No limit' }}
                        </div>
                        <q-slider
                            v-model="maxBudget"
                            :min="0"
                            :max="100000000"
                            :step="1000000"
                            color="primary"
                            class="q-px-sm"
                            @update:model-value="onFiltersChanged"
                        />
                    </div>
                    <div class="col-12 col-md-2">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Max Salary:
                            {{ maxSalary ? `${currencySymbol}${(maxSalary / 1000).toFixed(0)}k` : 'No limit' }}
                        </div>
                        <q-slider
                            v-model="maxSalary"
                            :min="0"
                            :max="500000"
                            :step="5000"
                            color="primary"
                            class="q-px-sm"
                            @update:model-value="onFiltersChanged"
                        />
                    </div>
                    <div class="col-12 col-md-2">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Min Age:
                            {{ minAge ? `${minAge} years` : 'Any age' }}
                        </div>
                        <q-slider
                            v-model="minAge"
                            :min="14"
                            :max="50"
                            :step="1"
                            color="primary"
                            class="q-px-sm"
                            @update:model-value="onFiltersChanged"
                        />
                    </div>
                    <div class="col-12 col-md-2">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Max Age:
                            {{ maxAge ? `${maxAge} years` : 'Any age' }}
                        </div>
                        <q-slider
                            v-model="maxAge"
                            :min="14"
                            :max="50"
                            :step="1"
                            color="primary"
                            class="q-px-sm"
                            @update:model-value="onFiltersChanged"
                        />
                    </div>
                    <div class="col-12 col-md-2">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Min Overall:
                            {{ minOverall ? `${minOverall} OVR` : 'Any rating' }}
                        </div>
                        <q-slider
                            v-model="minOverall"
                            :min="1"
                            :max="99"
                            :step="1"
                            color="primary"
                            class="q-px-sm"
                            @update:model-value="onFiltersChanged"
                        />
                    </div>
                </div>

                <!-- Info Section -->
                <q-card
                    class="q-mb-md info-card"
                    :class="qInstance.dark.isActive ? 'bg-grey-8' : 'bg-blue-1'"
                    flat
                >
                    <q-card-section>
                        <div class="row items-center">
                            <q-icon name="info" color="primary" size="md" class="q-mr-sm" />
                            <div class="text-subtitle1">
                                <strong>Tiered Value Score Formula:</strong>
                            </div>
                        </div>
                        <div class="text-caption q-mt-sm text-grey-6">
                            <div><strong>Premium (60+ rating):</strong> (Rating - 50) × (Rating ÷ Value in millions)</div>
                            <div><strong>Decent (55-59 rating):</strong> Rating ÷ Value in millions</div>
                            <div><strong>Budget (<55 rating):</strong> (Rating ÷ Value in millions) × 0.5</div>
                            <div class="q-mt-xs"><em>Free transfers are excluded from results</em></div>
                            <div class="q-mt-xs"><em>Age and overall filters are optional</em></div>
                            <div class="q-mt-xs"><em>Scores are normalized (highest = 100, lowest = 0) and limited to top 500 players</em></div>
                        </div>
                    </q-card-section>
                </q-card>

                <!-- Chart Section -->
                <q-card
                    v-if="bargainResults.length > 0 && !loading"
                    class="q-mb-md chart-card"
                    :class="qInstance.dark.isActive ? 'bg-grey-8' : 'bg-white'"
                >
                    <q-card-section>
                        <div class="text-subtitle1 q-mb-md">
                            <q-icon name="scatter_plot" class="q-mr-sm" />
                            Value Score vs Overall Rating
                            <q-btn 
                                v-if="bargainResults.length > 0"
                                size="sm" 
                                flat 
                                icon="refresh" 
                                @click="createChart"
                                class="q-ml-sm"
                                title="Refresh Chart"
                            />
                        </div>
                        <div class="chart-container">
                            <canvas ref="chartCanvas" width="400" height="200"></canvas>
                            <div v-if="!chartInstance" class="chart-loading text-center q-pa-md">
                                <q-spinner color="primary" size="2em" />
                                <div class="text-caption q-mt-sm">Loading chart...</div>
                            </div>
                        </div>
                        <div class="text-caption q-mt-sm text-grey-6">
                            Top 500 players shown. Value scores are normalized (0-100). Ideal bargains appear in the top-right (high value score, high overall).
                        </div>
                    </q-card-section>
                </q-card>

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
                        Finding bargains...
                    </div>
                </div>

                <!-- Results -->
                <div v-else-if="bargainResults.length > 0">
                    <div class="text-subtitle1 q-mb-md">
                        {{ bargainResults.length }} players found within budget
                        <span v-if="maxBudget || maxSalary" class="text-caption text-grey-6">
                            (filtered by budget constraints)
                        </span>
                    </div>
                    
                    <q-card
                        class="bargain-table-container"
                        :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                    >
                        <q-card-section>
                            <q-table
                                :rows="bargainResults"
                                :columns="tableColumns"
                                row-key="player.name"
                                :pagination="tablePagination"
                                :loading="loading"
                                dense
                                :dark="qInstance.dark.isActive"
                                @row-click="handlePlayerSelected"
                            >
                                <template v-slot:body-cell-valueScore="props">
                                    <q-td :props="props">
                                        <q-chip 
                                            :color="getValueScoreColor(props.value)"
                                            text-color="white"
                                            size="sm"
                                        >
                                            {{ formatValueScore(props.value) }}
                                        </q-chip>
                                    </q-td>
                                </template>
                                
                                <template v-slot:body-cell-overall="props">
                                    <q-td :props="props">
                                        <span :class="getOverallClass(props.value)">
                                            {{ props.value }}
                                        </span>
                                    </q-td>
                                </template>
                                
                                <template v-slot:body-cell-transferValue="props">
                                    <q-td :props="props">
                                        {{ formatCurrency(props.value, currencySymbol) }}
                                    </q-td>
                                </template>
                                
                                <template v-slot:body-cell-wage="props">
                                    <q-td :props="props">
                                        {{ formatCurrency(props.value, currencySymbol) }}
                                    </q-td>
                                </template>
                            </q-table>
                        </q-card-section>
                    </q-card>
                </div>

                <!-- Empty State -->
                <div v-else class="text-center q-my-xl">
                    <q-icon name="search_off" size="4em" color="grey-5" />
                    <div class="text-h6 q-mt-md text-grey-6">
                        No players found
                    </div>
                    <div class="text-body2 text-grey-5">
                        Try adjusting your budget constraints or check if data is loaded
                    </div>
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
import {
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  PointElement,
  ScatterController,
  Title,
  Tooltip
} from 'chart.js'
import { useQuasar } from 'quasar'
import { computed, defineComponent, nextTick, onMounted, ref, watch } from 'vue'
import { formatCurrency } from '../utils/currencyUtils'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  Title,
  Tooltip,
  Legend,
  ScatterController
)

export default defineComponent({
  name: 'BargainHunterDialog',
  components: {
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
    }
  },
  emits: ['close'],
  setup(props) {
    const qInstance = useQuasar()

    // State
    const maxBudget = ref(null)
    const maxSalary = ref(null)
    const minAge = ref(null)
    const maxAge = ref(null)
    const minOverall = ref(null)
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const bargainResults = ref([])
    const chartCanvas = ref(null)
    const chartInstance = ref(null)

    // Table configuration
    const tableColumns = [
      {
        name: 'valueScore',
        label: 'Value Score',
        field: 'valueScore',
        sortable: true,
        sort: (a, b) => b - a, // Highest first
        align: 'center'
      },
      {
        name: 'name',
        label: 'Name',
        field: row => row.player.name,
        sortable: true,
        align: 'left'
      },
      {
        name: 'overall',
        label: 'Overall',
        field: row => row.player.Overall,
        sortable: true,
        align: 'center'
      },
      {
        name: 'age',
        label: 'Age',
        field: row => row.player.age,
        sortable: true,
        align: 'center'
      },
      {
        name: 'position',
        label: 'Position',
        field: row => row.player.shortPositions?.join('/') || '-',
        sortable: true,
        align: 'center'
      },
      {
        name: 'transferValue',
        label: 'Transfer Value',
        field: row => row.player.transferValueAmount,
        sortable: true,
        align: 'right'
      },
      {
        name: 'wage',
        label: 'Wage',
        field: row => row.player.wageAmount,
        sortable: true,
        align: 'right'
      },
      {
        name: 'club',
        label: 'Club',
        field: row => row.player.club,
        sortable: true,
        align: 'left'
      }
    ]

    const tablePagination = ref({
      page: 1,
      rowsPerPage: 25,
      sortBy: 'valueScore',
      descending: true
    })

    // Methods
    const findBargains = async () => {
      if (!props.datasetId) {
        qInstance.notify({
          color: 'negative',
          message: 'Dataset ID is required',
          icon: 'error'
        })
        return
      }

      loading.value = true
      try {
        const response = await fetch(`/api/bargain-hunter/${props.datasetId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            maxBudget: maxBudget.value || 0,
            maxSalary: maxSalary.value || 0,
            minAge: minAge.value || 0,
            maxAge: maxAge.value || 0,
            minOverall: minOverall.value || 0
          })
        })

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }

        const data = await response.json()
        bargainResults.value = data || []
      } catch (error) {
        console.error('Error finding bargains:', error)
        qInstance.notify({
          color: 'negative',
          message: 'Failed to find bargains',
          icon: 'error'
        })
        bargainResults.value = []
      } finally {
        loading.value = false
      }
    }

    const onFiltersChanged = () => {
      // Debounce the filter changes
      clearTimeout(onFiltersChanged.timeout)
      onFiltersChanged.timeout = setTimeout(() => {
        if (props.show) {
          findBargains()
        }
      }, 300)
    }

    const handlePlayerSelected = (_evt, row) => {
      selectedPlayer.value = row.player
      showPlayerDetail.value = true
    }

    const formatValueScore = score => {
      if (score >= 1000) {
        return `${(score / 1000).toFixed(1)}k`
      }
      return score.toFixed(1)
    }

    const getValueScoreColor = score => {
      if (score >= 100) return 'green'
      if (score >= 50) return 'positive'
      if (score >= 25) return 'orange'
      if (score >= 10) return 'warning'
      return 'grey'
    }

    const getOverallClass = overall => {
      if (overall >= 85) return 'text-green text-weight-bold'
      if (overall >= 75) return 'text-positive text-weight-bold'
      if (overall >= 65) return 'text-orange text-weight-bold'
      if (overall >= 55) return 'text-warning'
      return 'text-grey'
    }

    const createChart = () => {
      if (!chartCanvas.value) {
        console.warn('Canvas element not available for chart creation')
        return
      }

      if (!bargainResults.value.length) {
        console.warn('No bargain results available for chart')
        return
      }

      // Destroy existing chart
      if (chartInstance.value) {
        chartInstance.value.destroy()
        chartInstance.value = null
      }

      try {
        const ctx = chartCanvas.value.getContext('2d')

        if (!ctx) {
          console.error('Could not get 2D context from canvas')
          return
        }

        // Set canvas size explicitly
        chartCanvas.value.width = 400
        chartCanvas.value.height = 200

        // Prepare data for scatter plot
        const scatterData = bargainResults.value
          .slice(0, 500) // Take top 500 players for chart (increased from 100)
          .map(result => ({
            x: result.player.Overall, // Overall rating (0-100)
            y: result.valueScore, // Normalized value score (0-100)
            player: result.player,
            originalValueScore: result.valueScore
          }))

        const config = {
          type: 'scatter',
          data: {
            datasets: [
              {
                label: 'Players',
                data: scatterData,
                backgroundColor: 'rgba(26, 35, 126, 0.6)',
                borderColor: 'rgba(26, 35, 126, 0.8)',
                pointRadius: 6,
                pointHoverRadius: 8
              }
            ]
          },
          options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
              legend: {
                display: false
              },
              tooltip: {
                callbacks: {
                  title: context => {
                    const point = context[0]
                    return point.raw.player.name
                  },
                  label: context => {
                    const point = context.raw
                    return [
                      `Overall: ${point.x}`,
                      `Value Score: ${formatValueScore(point.originalValueScore)}`,
                      `Club: ${point.player.club}`,
                      `Age: ${point.player.age}`,
                      `Transfer Value: ${formatCurrency(point.player.transferValueAmount, props.currencySymbol)}`
                    ]
                  }
                }
              }
            },
            scales: {
              x: {
                type: 'linear',
                position: 'bottom',
                title: {
                  display: true,
                  text: 'Overall Rating'
                },
                min: 0,
                max: 100,
                grid: {
                  color: qInstance.dark.isActive ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)'
                },
                ticks: {
                  color: qInstance.dark.isActive ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.7)'
                }
              },
              y: {
                title: {
                  display: true,
                  text: 'Value Score (0-100)'
                },
                min: 0,
                max: 100,
                grid: {
                  color: qInstance.dark.isActive ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)'
                },
                ticks: {
                  color: qInstance.dark.isActive ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.7)'
                }
              }
            },
            onClick: (_event, elements) => {
              if (elements.length > 0) {
                const element = elements[0]
                const dataIndex = element.index
                const player = scatterData[dataIndex].player
                handlePlayerSelected(null, { player })
              }
            }
          }
        }
        chartInstance.value = new ChartJS(ctx, config)
      } catch (error) {
        console.error('Error creating chart:', error)
      }
    }

    const updateChart = async () => {
      await nextTick()
      // Wait a bit more to ensure DOM is fully rendered
      setTimeout(() => {
        createChart()
      }, 200) // Increased timeout to ensure DOM is ready
    }

    // Watchers
    watch(
      () => props.show,
      async newShow => {
        if (newShow && props.datasetId) {
          // Auto-search when dialog opens
          await findBargains()
        } else if (!newShow) {
          // Cleanup chart when dialog closes
          if (chartInstance.value) {
            chartInstance.value.destroy()
            chartInstance.value = null
          }
        }
      }
    )

    // Watch for results changes to update chart
    watch(
      () => bargainResults.value,
      async newResults => {
        if (newResults.length > 0 && props.show) {
          await updateChart()
        }
      },
      { deep: true }
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
      minAge,
      maxAge,
      minOverall,
      loading,
      selectedPlayer,
      showPlayerDetail,
      bargainResults,
      chartCanvas,
      chartInstance,
      tableColumns,
      tablePagination,
      onFiltersChanged,
      handlePlayerSelected,
      formatValueScore,
      getValueScoreColor,
      getOverallClass,
      formatCurrency,
      createChart,
      updateChart
    }
  }
})
</script>

<style lang="scss" scoped>
.bargain-hunter-dialog {
    .bargain-table-container {
        border: none;
        box-shadow: none;
        
        :deep(.q-table) {
            height: auto !important;
            max-height: calc(100vh - 400px);
            
            .q-table__middle {
                max-height: calc(100vh - 400px);
            }
        }
    }
    
    .info-card {
        border-left: 4px solid var(--q-primary);
    }
    
    .chart-card {
        border-left: 4px solid var(--q-positive);
        
        .chart-container {
            height: 300px;
            position: relative;
            
            canvas {
                max-height: 100%;
                width: 100% !important;
                height: 100% !important;
            }
            
            .chart-loading {
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                z-index: 10;
            }
        }
    }
}

// Ensure proper spacing
.q-card-section {
    &:last-child {
        padding-bottom: 24px;
    }
}
</style> 