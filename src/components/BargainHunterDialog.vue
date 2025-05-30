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
                    <div class="col-12 col-md-6">
                        <q-input
                            v-model.number="maxBudget"
                            type="number"
                            label="Max Transfer Budget"
                            outlined
                            dense
                            :prefix="currencySymbol"
                            :min="0"
                            @update:model-value="onFiltersChanged"
                            :label-color="qInstance.dark.isActive ? 'grey-4' : ''"
                            :input-class="qInstance.dark.isActive ? 'text-grey-3' : ''"
                        />
                    </div>
                    <div class="col-12 col-md-6">
                        <q-input
                            v-model.number="maxSalary"
                            type="number"
                            label="Max Salary"
                            outlined
                            dense
                            :prefix="currencySymbol"
                            :min="0"
                            @update:model-value="onFiltersChanged"
                            :label-color="qInstance.dark.isActive ? 'grey-4' : ''"
                            :input-class="qInstance.dark.isActive ? 'text-grey-3' : ''"
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
                                <strong>Value Score Formula:</strong> Overall Rating ÷ Transfer Value (in millions)
                            </div>
                        </div>
                        <div class="text-caption q-mt-sm text-grey-6">
                            Higher value scores indicate better value for money. Free transfers get a very high value score.
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
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref, watch } from 'vue'
import { formatCurrency } from '../utils/currencyUtils'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

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
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const bargainResults = ref([])

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
            maxSalary: maxSalary.value || 0
          })
        })

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }

        const data = await response.json()
        bargainResults.value = data || []

        console.log(`Found ${bargainResults.value.length} bargain players:`, bargainResults.value)

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

    const handlePlayerSelected = (evt, row) => {
      selectedPlayer.value = row.player
      showPlayerDetail.value = true
    }

    const formatValueScore = (score) => {
      if (score >= 1000) {
        return (score / 1000).toFixed(1) + 'k'
      }
      return score.toFixed(1)
    }

    const getValueScoreColor = (score) => {
      if (score >= 100) return 'green'
      if (score >= 50) return 'positive'
      if (score >= 25) return 'orange'
      if (score >= 10) return 'warning'
      return 'grey'
    }

    const getOverallClass = (overall) => {
      if (overall >= 85) return 'text-green text-weight-bold'
      if (overall >= 75) return 'text-positive text-weight-bold'
      if (overall >= 65) return 'text-orange text-weight-bold'
      if (overall >= 55) return 'text-warning'
      return 'text-grey'
    }

    // Watchers
    watch(
      () => props.show,
      (newShow) => {
        if (newShow && props.datasetId) {
          // Auto-search when dialog opens
          findBargains()
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
      loading,
      selectedPlayer,
      showPlayerDetail,
      bargainResults,
      tableColumns,
      tablePagination,
      onFiltersChanged,
      handlePlayerSelected,
      formatValueScore,
      getValueScoreColor,
      getOverallClass,
      formatCurrency
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
}

// Ensure proper spacing
.q-card-section {
    &:last-child {
        padding-bottom: 24px;
    }
}
</style> 