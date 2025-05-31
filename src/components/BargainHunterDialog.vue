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
                <q-icon name="shopping_cart" size="md" class="q-mr-sm" />
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
                <!-- Quick Search Section -->
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-3">
                        <q-input
                            v-model.number="maxBudget"
                            type="number"
                            label="Max Transfer Value"
                            outlined
                            dense
                            :prefix="currencySymbol"
                            :min="0"
                            @update:model-value="onFiltersChanged"
                            :label-color="qInstance.dark.isActive ? 'grey-4' : ''"
                            :input-class="qInstance.dark.isActive ? 'text-grey-3' : ''"
                        />
                    </div>
                    
                    <div class="col-12 col-md-3">
                        <q-input
                            v-model.number="maxSalary"
                            type="number"
                            label="Max Wage per Week"
                            outlined
                            dense
                            :prefix="currencySymbol"
                            :min="0"
                            @update:model-value="onFiltersChanged"
                            :label-color="qInstance.dark.isActive ? 'grey-4' : ''"
                            :input-class="qInstance.dark.isActive ? 'text-grey-3' : ''"
                        />
                    </div>
                    
                    <div class="col-12 col-md-2">
                        <q-input
                            v-model.number="minAge"
                            type="number"
                            label="Min Age"
                            outlined
                            dense
                            :min="15"
                            :max="45"
                            @update:model-value="onFiltersChanged"
                            :label-color="qInstance.dark.isActive ? 'grey-4' : ''"
                            :input-class="qInstance.dark.isActive ? 'text-grey-3' : ''"
                        />
                    </div>
                    
                    <div class="col-12 col-md-2">
                        <q-input
                            v-model.number="maxAge"
                            type="number"
                            label="Max Age"
                            outlined
                            dense
                            :min="15"
                            :max="45"
                            @update:model-value="onFiltersChanged"
                            :label-color="qInstance.dark.isActive ? 'grey-4' : ''"
                            :input-class="qInstance.dark.isActive ? 'text-grey-3' : ''"
                        />
                    </div>
                    
                    <div class="col-12 col-md-2">
                        <q-input
                            v-model.number="minOverall"
                            type="number"
                            label="Min Overall"
                            outlined
                            dense
                            :min="40"
                            :max="99"
                            @update:model-value="onFiltersChanged"
                            :label-color="qInstance.dark.isActive ? 'grey-4' : ''"
                            :input-class="qInstance.dark.isActive ? 'text-grey-3' : ''"
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

                <!-- Enhanced Chart Section with ECharts -->
                <q-card
                    v-if="filteredBargainResults.length > 0 && !loading"
                    class="q-mb-md"
                    :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                    flat
                    bordered
                >
                    <q-card-section>
                        <div class="row items-center q-mb-md">
                            <div class="text-subtitle1">
                                <q-icon name="scatter_plot" class="q-mr-sm" />
                                Value Score vs Overall Rating
                            </div>
                            <q-space />
                            <q-btn 
                                size="sm" 
                                flat 
                                icon="refresh" 
                                @click="refreshChart"
                                label="Refresh"
                                class="q-ml-sm"
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
                            Showing {{ filteredBargainResults.length }} players{{ bargainResults.length !== filteredBargainResults.length ? ` (filtered from ${bargainResults.length} total)` : '' }}. 
                            <strong>Tip:</strong> Click points to view details, use mouse wheel to zoom, drag to pan.
                        </div>
                    </q-card-section>
                </q-card>

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
                    <q-card 
                        class="bargain-table-container" 
                        :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                        flat
                        bordered
                    >
                        <q-card-section>
                            <div class="row items-center q-mb-md">
                                <div class="text-subtitle1">
                                    <q-icon name="list" class="q-mr-sm" />
                                    Best Value Players
                                </div>
                                <q-space />
                                <q-chip 
                                    color="primary" 
                                    text-color="white"
                                    :label="`${filteredBargainResults.length} players`"
                                />
                            </div>
                            
                            <PlayerDataTable
                                :players="playersForTable"
                                :loading="loading"
                                @player-selected="handlePlayerSelected"
                                @team-selected="handleTeamSelected"
                                :currency-symbol="currencySymbol"
                            />
                        </q-card-section>
                    </q-card>
                </div>

                <!-- Empty State -->
                <div v-else-if="!loading" class="text-center q-my-xl">
                    <q-icon name="search_off" size="4em" color="grey-5" />
                    <div class="text-h6 q-mt-md text-grey-6">
                        <span v-if="bargainResults.length === 0">No bargains found</span>
                        <span v-else>No players match current filters</span>
                    </div>
                    <div class="text-body2 text-grey-5 q-mt-sm">
                        <span v-if="bargainResults.length === 0">Try adjusting your budget or age criteria</span>
                        <span v-else>Try enabling more value score ranges or adjusting your filters</span>
                    </div>
                </div>

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
import PlayerDataTable from './PlayerDataTable.vue'

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
    VChart,
    PlayerDataTable
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

    // Computed property to transform bargain results for PlayerDataTable
    const playersForTable = computed(() => {
      return filteredBargainResults.value.map(result => ({
        ...result.player,
        valueScore: result.valueScore
      }))
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

    const handleTeamSelected = (teamName) => {
      // For bargain hunter, we don't need team selection functionality
      // but we need to provide the handler for PlayerDataTable compatibility
    }

    const formatValueScore = (score) => {
      if (typeof score !== 'number') return '0'
      return Math.round(score).toString()
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
      
      // Trigger filtering after tier change
      onValueTierChanged()
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
      onFiltersChanged,
      onValueTierChanged,
      handlePlayerSelected,
      handleTeamSelected,
      formatCurrency,
      findBargains,
      refreshChart,
      handleChartClick,
      showExcellentValue,
      showGreatValue,
      showGoodValue,
      showMediocreValue,
      showPoorValue,
      filteredBargainResults,
      toggleValueTier,
      playersForTable
    }
  }
})
</script>

<style lang="scss" scoped>
.bargain-hunter-dialog {
    .chart-container {
        height: 500px;
        position: relative;
        
        .chart {
            width: 100% !important;
            height: 100% !important;
        }
    }
    
    // Value tier button styling
    .value-tier-buttons {
        .q-btn {
            transition: all 0.2s ease;
            font-weight: 500;
            text-transform: none;
            
            &:not(.q-btn--outline) {
                box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
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
    
    // Improved input styling
    :deep(.q-field) {
        .q-field__control {
            border-radius: 8px;
        }
        
        &.q-field--outlined .q-field__control {
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
            transition: box-shadow 0.2s ease;
            
            &:hover {
                box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
            }
        }
        
        &.q-field--focused .q-field__control {
            box-shadow: 0 0 0 2px rgba(25, 118, 210, 0.2);
        }
    }
    
    // Card styling improvements
    .q-card {
        border-radius: 12px;
        transition: box-shadow 0.2s ease;
        
        &[flat] {
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
            
            &:hover {
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            }
        }
        
        &[bordered] {
            border: 1px solid rgba(0, 0, 0, 0.05);
        }
    }
    
    // Search button styling
    .search-button {
        .q-btn {
            border-radius: 8px;
            font-weight: 600;
            text-transform: none;
            padding: 12px 24px;
            box-shadow: 0 2px 8px rgba(25, 118, 210, 0.3);
            
            &:hover {
                box-shadow: 0 4px 12px rgba(25, 118, 210, 0.4);
                transform: translateY(-1px);
            }
        }
    }
    
    // Chip improvements
    .q-chip {
        border-radius: 6px;
        font-weight: 500;
    }
    
    // Expansion item styling
    :deep(.q-expansion-item) {
        border-radius: 8px;
        
        .q-expansion-item__container {
            border-radius: 8px;
        }
        
        .q-expansion-item__toggle {
            padding: 16px;
        }
    }
}

// Loading spinner improvements
.loading-container {
    .q-spinner-dots {
        color: var(--q-primary);
    }
    
    .loading-text {
        color: var(--q-primary);
        font-weight: 500;
    }
}

// Enhanced responsive design
@media (max-width: 1024px) {
    .bargain-hunter-dialog {
        .chart-container {
            height: 450px;
        }
    }
}

@media (max-width: 768px) {
    .bargain-hunter-dialog {
        .chart-container {
            height: 400px;
        }
        
        // Stack value tier buttons on mobile
        .value-tier-buttons {
            .row {
                flex-direction: column;
                gap: 8px;
                
                .col-auto {
                    width: 100%;
                    
                    .q-btn {
                        width: 100%;
                    }
                }
            }
        }
    }
}

@media (max-width: 480px) {
    .bargain-hunter-dialog {
        .chart-container {
            height: 300px;
        }
        
        // Adjust padding for mobile
        .q-card-section {
            padding: 16px 12px;
        }
    }
}

// Dark mode specific adjustments
.body--dark {
    .bargain-hunter-dialog {
        .q-card[flat] {
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
            
            &:hover {
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
            }
        }
        
        .q-card[bordered] {
            border-color: rgba(255, 255, 255, 0.1);
        }
        
        :deep(.q-field--outlined .q-field__control) {
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
            
            &:hover {
                box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
            }
        }
    }
}
</style> 