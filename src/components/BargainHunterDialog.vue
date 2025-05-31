<template>
    <q-dialog
        :model-value="show"
        @hide="$emit('close')"
        persistent
        maximized
    >
        <q-card
            class="bargain-hunter-dialog"
            :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-grey-1'"
        >
            <q-card-section
                class="row items-center q-pb-none"
                :class="
                    qInstance.dark.isActive
                        ? 'bg-grey-10 text-white'
                        : 'bg-primary text-white'
                "
            >
                <q-icon name="search" class="q-mr-sm" />
                <div class="text-h6">
                    Bargain Hunter
                </div>
                <q-space />
                <q-btn
                    icon="close"
                    flat
                    round
                    dense
                    @click="$emit('close')"
                />
            </q-card-section>

            <q-card-section>
                <!-- Filters Section -->
                <q-card 
                    class="q-mb-md filters-card" 
                    :class="qInstance.dark.isActive ? 'bg-grey-8' : 'bg-white'"
                    flat
                >
                    <q-card-section>
                        <div class="text-subtitle1 q-mb-md">
                            <q-icon name="tune" class="q-mr-sm" />
                            Search Filters
                        </div>
                        
                        <div class="row q-gutter-md">
                            <div class="col-12 col-md-2">
                                <q-input
                                    v-model.number="maxBudget"
                                    type="number"
                                    label="Max Transfer Value (M)"
                                    outlined
                                    dense
                                    @update:model-value="onFiltersChanged"
                                />
                            </div>
                            
                            <div class="col-12 col-md-2">
                                <q-input
                                    v-model.number="maxSalary"
                                    type="number"
                                    label="Max Wage (K/week)"
                                    outlined
                                    dense
                                    @update:model-value="onFiltersChanged"
                                />
                            </div>
                            
                            <div class="col-12 col-md-2">
                                <q-input
                                    v-model.number="minAge"
                                    type="number"
                                    label="Min Age"
                                    outlined
                                    dense
                                    @update:model-value="onFiltersChanged"
                                />
                            </div>
                            
                            <div class="col-12 col-md-2">
                                <q-input
                                    v-model.number="maxAge"
                                    type="number"
                                    label="Max Age"
                                    outlined
                                    dense
                                    @update:model-value="onFiltersChanged"
                                />
                            </div>
                            
                            <div class="col-12 col-md-2">
                                <q-input
                                    v-model.number="minOverall"
                                    type="number"
                                    label="Min Overall"
                                    outlined
                                    dense
                                    @update:model-value="onFiltersChanged"
                                />
                            </div>
                            
                            <div class="col-12 col-md-1">
                                <q-btn 
                                    color="primary" 
                                    icon="search" 
                                    label="Search" 
                                    @click="findBargains"
                                    :loading="loading"
                                    class="full-width"
                                />
                            </div>
                        </div>

                        <!-- Value Score Tier Filters -->
                        <q-separator class="q-my-md" />
                        <div class="text-subtitle2 q-mb-sm">
                            <q-icon name="filter_list" class="q-mr-sm" />
                            Show Value Score Tiers
                        </div>
                        <div class="row q-gutter-md">
                            <div class="col-auto">
                                <q-checkbox
                                    v-model="showExcellentValue"
                                    label="Excellent (80-100)"
                                    color="positive"
                                    @update:model-value="onValueTierChanged"
                                />
                                <q-icon 
                                    name="circle" 
                                    color="positive" 
                                    size="xs" 
                                    class="q-ml-xs"
                                    style="color: #00C851 !important;"
                                />
                            </div>
                            <div class="col-auto">
                                <q-checkbox
                                    v-model="showGreatValue"
                                    label="Great (60-79)"
                                    color="positive"
                                    @update:model-value="onValueTierChanged"
                                />
                                <q-icon 
                                    name="circle" 
                                    color="positive" 
                                    size="xs" 
                                    class="q-ml-xs"
                                    style="color: #2E7D32 !important;"
                                />
                            </div>
                            <div class="col-auto">
                                <q-checkbox
                                    v-model="showGoodValue"
                                    label="Good (40-59)"
                                    color="orange"
                                    @update:model-value="onValueTierChanged"
                                />
                                <q-icon 
                                    name="circle" 
                                    color="orange" 
                                    size="xs" 
                                    class="q-ml-xs"
                                    style="color: #FF6F00 !important;"
                                />
                            </div>
                            <div class="col-auto">
                                <q-checkbox
                                    v-model="showMediocreValue"
                                    label="Mediocre (20-39)"
                                    color="negative"
                                    @update:model-value="onValueTierChanged"
                                />
                                <q-icon 
                                    name="circle" 
                                    color="negative" 
                                    size="xs" 
                                    class="q-ml-xs"
                                    style="color: #FF5722 !important;"
                                />
                            </div>
                            <div class="col-auto">
                                <q-checkbox
                                    v-model="showPoorValue"
                                    label="Poor (0-19)"
                                    color="grey"
                                    @update:model-value="onValueTierChanged"
                                />
                                <q-icon 
                                    name="circle" 
                                    color="grey" 
                                    size="xs" 
                                    class="q-ml-xs"
                                    style="color: #757575 !important;"
                                />
                            </div>
                        </div>
                    </q-card-section>
                </q-card>

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
                            <div><strong>Elite (80+ rating):</strong> High efficiency with premium pricing tolerance</div>
                            <div><strong>Quality (70-79 rating):</strong> Balanced efficiency and value expectation</div>
                            <div><strong>Decent (60-69 rating):</strong> Good value for money expected</div>
                            <div><strong>Budget (55-59 rating):</strong> Lower cost expected</div>
                            <div><strong>Youth (<55 rating):</strong> Development potential focus</div>
                            <div class="q-mt-xs"><em>Uses logarithmic scaling to fairly compare expensive vs cheap players</em></div>
                            <div class="q-mt-xs"><em>Bonuses applied for exceptional value (30%+ below expected pricing)</em></div>
                            <div class="q-mt-xs"><em>Free transfers are excluded from results</em></div>
                            <div class="q-mt-xs"><em>Scores are normalized (highest = 100, lowest = 0) and limited to top 500 players</em></div>
                        </div>
                    </q-card-section>
                </q-card>

                <!-- Enhanced Chart Section with ECharts -->
                <q-card
                    v-if="filteredBargainResults.length > 0 && !loading"
                    class="q-mb-md chart-card"
                    :class="qInstance.dark.isActive ? 'bg-grey-8' : 'bg-white'"
                    flat
                >
                    <q-card-section>
                        <div class="text-subtitle1 q-mb-md">
                            <q-icon name="scatter_plot" class="q-mr-sm" />
                            Value Score vs Overall Rating
                            <q-btn 
                                v-if="filteredBargainResults.length > 0"
                                size="sm" 
                                flat 
                                icon="refresh" 
                                @click="refreshChart"
                                class="q-ml-sm"
                                title="Refresh Chart"
                            />
                        </div>
                        <div class="chart-container">
                            <v-chart
                                ref="chartRef"
                                :option="chartOption"
                                :theme="qInstance.dark.isActive ? 'dark' : 'light'"
                                autoresize
                                @click="handleChartClick"
                                class="chart"
                            />
                        </div>
                        <div class="text-caption q-mt-sm text-grey-6">
                            {{ filteredBargainResults.length }} players shown (filtered from {{ bargainResults.length }} total results). 
                            Value scores are normalized (0-100). Ideal bargains appear in the top-right (high value score, high overall).
                            <strong>Tip:</strong> Click points to view player details, use mouse wheel to zoom, drag to pan.
                        </div>
                    </q-card-section>
                </q-card>

                <!-- Loading State -->
                <div v-if="loading" class="text-center q-my-xl">
                    <q-spinner color="primary" size="3em" />
                    <div class="text-h6 q-mt-md">
                        Finding the best bargains...
                    </div>
                </div>

                <!-- Results Table -->
                <div v-if="filteredBargainResults.length > 0 && !loading">
                    <q-card class="bargain-table-container" flat>
                        <q-card-section>
                            <div class="text-subtitle1 q-mb-md">
                                <q-icon name="list" class="q-mr-sm" />
                                Best Value Players ({{ filteredBargainResults.length }} shown{{ bargainResults.length !== filteredBargainResults.length ? ` of ${bargainResults.length} total` : '' }})
                            </div>
                            
                            <q-table
                                :rows="filteredBargainResults"
                                :columns="tableColumns"
                                :pagination="tablePagination"
                                row-key="player.name"
                                flat
                                class="bargain-table"
                                :class="qInstance.dark.isActive ? 'bg-grey-8' : 'bg-white'"
                            >
                                <template v-slot:body-cell-player="props">
                                    <q-td :props="props">
                                        <div 
                                            class="player-cell cursor-pointer" 
                                            @click="handlePlayerSelected(null, props.row)"
                                        >
                                            <div class="text-weight-bold text-primary">{{ props.value.name }}</div>
                                            <div class="text-caption text-grey-6">{{ props.value.club }}</div>
                                        </div>
                                    </q-td>
                                </template>
                                
                                <template v-slot:body-cell-valueScore="props">
                                    <q-td :props="props">
                                        <q-badge 
                                            :color="getValueScoreColor(props.value)" 
                                            :label="formatValueScore(props.value)"
                                        />
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
                <div v-else-if="!loading" class="text-center q-my-xl">
                    <q-icon name="search_off" size="4em" color="grey-5" />
                    <div class="text-h6 q-mt-md text-grey-6">
                        <span v-if="bargainResults.length === 0">No players found</span>
                        <span v-else>No players match current value score filters</span>
                    </div>
                    <div class="text-body2 text-grey-5">
                        <span v-if="bargainResults.length === 0">Try adjusting your budget constraints or check if data is loaded</span>
                        <span v-else>Try enabling more value score tiers above or adjusting your search filters</span>
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
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { ScatterChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DataZoomComponent,
  BrushComponent,
  ToolboxComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { useQuasar } from 'quasar'
import { computed, defineComponent, nextTick, onMounted, ref, watch } from 'vue'
import { formatCurrency } from '../utils/currencyUtils'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

// Register ECharts components
use([
  CanvasRenderer,
  ScatterChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DataZoomComponent,
  BrushComponent,
  ToolboxComponent
])

export default defineComponent({
  name: 'BargainHunterDialog',
  components: {
    PlayerDetailDialog,
    VChart
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
    const maxAge = ref(27)
    const minOverall = ref(75)
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const bargainResults = ref([])
    const chartRef = ref(null)
    const showExcellentValue = ref(true)
    const showGreatValue = ref(true)
    const showGoodValue = ref(true)
    const showMediocreValue = ref(false)
    const showPoorValue = ref(false)

    // Table configuration
    const tableColumns = [
      {
        name: 'player',
        required: true,
        label: 'Player',
        align: 'left',
        field: row => row.player,
        format: val => val.name,
        sortable: true
      },
      {
        name: 'valueScore',
        align: 'center',
        label: 'Value Score',
        field: 'valueScore',
        sortable: true,
        sort: (a, b) => parseFloat(a) - parseFloat(b)
      },
      {
        name: 'overall',
        align: 'center',
        label: 'Overall',
        field: row => row.player.Overall,
        sortable: true
      },
      {
        name: 'age',
        align: 'center',
        label: 'Age',
        field: row => row.player.age,
        sortable: true
      },
      {
        name: 'transferValue',
        align: 'right',
        label: 'Transfer Value',
        field: row => row.player.transferValueAmount,
        sortable: true
      },
      {
        name: 'wage',
        align: 'right',
        label: 'Wage',
        field: row => row.player.wageAmount,
        sortable: true
      }
    ]

    const tablePagination = ref({
      rowsPerPage: 25
    })

    // Enhanced ECharts configuration
    const chartOption = computed(() => {
      if (!bargainResults.value.length) return {}

      // Filter data based on selected value score tiers
      const filteredResults = bargainResults.value.filter(result => {
        const score = result.valueScore
        if (score >= 80 && showExcellentValue.value) return true
        if (score >= 60 && score < 80 && showGreatValue.value) return true
        if (score >= 40 && score < 60 && showGoodValue.value) return true
        if (score >= 20 && score < 40 && showMediocreValue.value) return true
        if (score < 20 && showPoorValue.value) return true
        return false
      })

      const scatterData = filteredResults
        .slice(0, 500)
        .map((result, index) => [
          result.player.Overall, // x: Overall rating
          result.valueScore,     // y: Value score
          result.player.name,    // name for tooltip
          result.player.club,    // club for tooltip
          result.player.age,     // age for tooltip
          result.player.transferValueAmount, // transfer value for tooltip
          bargainResults.value.indexOf(result) // original index for click handling
        ])

      return {
        animation: true,
        animationDuration: 1000,
        animationEasing: 'elasticOut',
        grid: {
          left: '8%',
          right: '8%',
          top: '15%',
          bottom: '20%',
          containLabel: true
        },
        toolbox: {
          feature: {
            dataZoom: {
              yAxisIndex: false
            },
            brush: {
              type: ['rect', 'polygon', 'clear']
            },
            saveAsImage: {
              title: 'Save Chart'
            }
          },
          right: '2%',
          top: '2%'
        },
        brush: {
          toolbox: ['rect', 'polygon', 'clear'],
          xAxisIndex: 0,
          yAxisIndex: 0
        },
        tooltip: {
          trigger: 'item',
          backgroundColor: qInstance.dark.isActive ? 'rgba(50, 50, 50, 0.95)' : 'rgba(255, 255, 255, 0.95)',
          borderColor: qInstance.dark.isActive ? '#555' : '#ddd',
          textStyle: {
            color: qInstance.dark.isActive ? '#fff' : '#333'
          },
          formatter: (params) => {
            const [overall, valueScore, name, club, age, transferValue] = params.data
            return `
              <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px;">${name}</div>
              <div style="margin-bottom: 4px;"><strong>Club:</strong> ${club}</div>
              <div style="margin-bottom: 4px;"><strong>Overall:</strong> ${overall}</div>
              <div style="margin-bottom: 4px;"><strong>Value Score:</strong> ${formatValueScore(valueScore)}</div>
              <div style="margin-bottom: 4px;"><strong>Age:</strong> ${age}</div>
              <div><strong>Transfer Value:</strong> ${formatCurrency(transferValue, props.currencySymbol)}</div>
            `
          }
        },
        xAxis: {
          type: 'value',
          name: 'Overall Rating',
          nameLocation: 'middle',
          nameTextStyle: {
            padding: [20, 0, 0, 0],
            fontSize: 14,
            fontWeight: 'bold'
          },
          min: 45,
          max: 100,
          splitLine: {
            show: true,
            lineStyle: {
              opacity: 0.3
            }
          },
          axisLine: {
            lineStyle: {
              color: qInstance.dark.isActive ? '#555' : '#999'
            }
          }
        },
        yAxis: {
          type: 'value',
          name: 'Value Score',
          nameLocation: 'middle',
          nameTextStyle: {
            padding: [0, 0, 40, 0],
            fontSize: 14,
            fontWeight: 'bold'
          },
          min: 0,
          max: 100,
          splitLine: {
            show: true,
            lineStyle: {
              opacity: 0.3
            }
          },
          axisLine: {
            lineStyle: {
              color: qInstance.dark.isActive ? '#555' : '#999'
            }
          }
        },
        dataZoom: [
          {
            type: 'inside',
            xAxisIndex: 0,
            filterMode: 'none'
          },
          {
            type: 'inside',
            yAxisIndex: 0,
            filterMode: 'none'
          },
          {
            type: 'slider',
            xAxisIndex: 0,
            height: 20,
            bottom: 40
          }
        ],
        series: [
          {
            type: 'scatter',
            data: scatterData,
            symbolSize: (data) => {
              // Make high-value players slightly larger
              const valueScore = data[1]
              return Math.max(6, Math.min(12, valueScore / 8))
            },
            itemStyle: {
              color: (params) => {
                const valueScore = params.data[1]
                const overall = params.data[0]
                
                // Color based on value score with overall rating influence
                if (valueScore >= 80) return '#00C851' // Green for excellent value
                if (valueScore >= 60) return '#2E7D32' // Dark green for great value
                if (valueScore >= 40) return '#FF6F00' // Orange for good value
                if (valueScore >= 20) return '#FF5722' // Red-orange for fair value
                return '#757575' // Grey for poor value
              },
              borderColor: '#fff',
              borderWidth: 1,
              opacity: 0.8
            },
            emphasis: {
              itemStyle: {
                opacity: 1,
                borderWidth: 2,
                shadowBlur: 10,
                shadowColor: 'rgba(0, 0, 0, 0.3)'
              }
            },
            markLine: {
              silent: true,
              lineStyle: {
                color: qInstance.dark.isActive ? '#666' : '#ccc',
                type: 'dashed',
                opacity: 0.6
              },
              data: [
                { xAxis: 75, label: { formatter: 'Good Overall (75)', position: 'end' } },
                { yAxis: 50, label: { formatter: 'Decent Value (50)', position: 'end' } }
              ]
            }
          }
        ]
      }
    })

    // Filtered results for table display
    const filteredBargainResults = computed(() => {
      return bargainResults.value.filter(result => {
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
        console.warn('No dataset ID available')
        return
      }

      loading.value = true

      try {
        // Prepare request payload
        const requestBody = {
          maxBudget: maxBudget.value ? maxBudget.value * 1000000 : 0, // Convert to actual amount
          maxSalary: maxSalary.value ? maxSalary.value * 1000 : 0, // Convert to actual amount
          minAge: minAge.value || 0,
          maxAge: maxAge.value || 0,
          minOverall: minOverall.value || 0
        }

        // Call backend API
        const response = await fetch(`/api/bargain-hunter/${props.datasetId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(requestBody)
        })

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }

        const bargainData = await response.json()
        bargainResults.value = bargainData || []

        // Update chart after a short delay
        await nextTick()
        setTimeout(() => {
          if (chartRef.value && bargainResults.value.length > 0) {
            refreshChart()
          }
        }, 100)

      } catch (error) {
        console.error('Error finding bargains:', error)
        qInstance.notify({
          message: 'Error finding bargains. Please try again.',
          color: 'negative',
          icon: 'error'
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

    const onValueTierChanged = async () => {
      // Debounce the search to avoid too many calls
      clearTimeout(onValueTierChanged.timeoutId)
      onValueTierChanged.timeoutId = setTimeout(() => {
        if (props.show && props.datasetId) {
          findBargains()
        }
      }, 300)
    }

    const handlePlayerSelected = (evt, rowData) => {
      const player = rowData.player || rowData
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const formatValueScore = (score) => {
      if (typeof score !== 'number') return '0'
      return Math.round(score).toString()
    }

    const getValueScoreColor = (score) => {
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

    const refreshChart = () => {
      if (chartRef.value) {
        // Force chart to refresh with current data
        chartRef.value.setOption(chartOption.value, true)
      }
    }

    const handleChartClick = (event) => {
      if (event.data && event.data.length > 6) {
        const originalIndex = event.data[6] // Get the original index we stored
        if (bargainResults.value[originalIndex]) {
          const player = bargainResults.value[originalIndex].player
          handlePlayerSelected(null, { player: player })
        }
      }
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
          if (chartRef.value) {
            chartRef.value.dispose()
            chartRef.value = null
          }
        }
      }
    )

    // Watch for results changes to update chart
    watch(
      () => bargainResults.value,
      async newResults => {
        if (newResults.length > 0 && props.show) {
          await nextTick()
          // Wait a bit more to ensure DOM is fully rendered
          setTimeout(() => {
            if (chartRef.value) {
              refreshChart()
            }
          }, 200)
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
      chartRef,
      chartOption,
      tableColumns,
      tablePagination,
      onFiltersChanged,
      onValueTierChanged,
      handlePlayerSelected,
      formatValueScore,
      getValueScoreColor,
      getOverallClass,
      formatCurrency,
      findBargains,
      refreshChart,
      handleChartClick,
      showExcellentValue,
      showGreatValue,
      showGoodValue,
      showMediocreValue,
      showPoorValue,
      filteredBargainResults
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
    
    .filters-card {
        border-left: 4px solid var(--q-primary);
    }
    
    .info-card {
        border-left: 4px solid var(--q-primary);
    }
    
    .chart-card {
        border-left: 4px solid var(--q-positive);
        
        .chart-container {
            height: 500px;
            position: relative;
            
            .chart {
                width: 100% !important;
                height: 100% !important;
            }
        }
    }
    
    .player-cell {
        &:hover {
            background-color: rgba(25, 118, 210, 0.04);
            border-radius: 4px;
        }
        
        .text-primary {
            &:hover {
                text-decoration: underline;
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

// Enhanced responsive design
@media (max-width: 768px) {
    .bargain-hunter-dialog {
        .chart-card .chart-container {
            height: 400px;
        }
    }
}

@media (max-width: 480px) {
    .bargain-hunter-dialog {
        .chart-card .chart-container {
            height: 300px;
        }
    }
}
</style> 