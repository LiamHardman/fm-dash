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
                    <div class="col-12 col-md-6">
                        <q-input
                            v-model.number="maxTransferValue"
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
                        <span v-if="maxTransferValue || maxSalary" class="text-caption text-grey-6">
                            (filtered by transfer value and/or salary)
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
import { computed, defineComponent, nextTick, onMounted, ref, watch } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

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
    }
  },
  emits: ['close'],
  setup(props) {
    const qInstance = useQuasar()
    const _playerStore = usePlayerStore()

    // State
    const maxTransferValue = ref(null)
    const maxSalary = ref(null)
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const wonderkidsData = ref([])

    // Constants
    const ages = [15, 16, 17, 18, 19, 20, 21]

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
          .filter(player => {
            const playerAge = Number(player.age)
            // Must match this specific age
            if (playerAge !== age) {
              return false
            }
            
            // Apply transfer value filter
            if (maxTransferValue.value && player.transferValueAmount > maxTransferValue.value) {
              return false
            }
            
            // Apply salary filter
            if (maxSalary.value && player.wageAmount > maxSalary.value) {
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
      } catch (error) {
        console.error('Error loading wonderkids:', error)
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
      // Debounce the filter changes
      clearTimeout(onFiltersChanged.timeout)
      onFiltersChanged.timeout = setTimeout(() => {
        loadWonderkids()
      }, 300)
    }

    const handlePlayerSelected = player => {
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const handleTeamSelected = _teamName => {}

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
      maxTransferValue,
      maxSalary,
      loading,
      selectedPlayer,
      showPlayerDetail,
      allWonderkids,
      onFiltersChanged,
      handlePlayerSelected,
      handleTeamSelected
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

    // Input field styling
    :deep(.q-field) {
        .q-field__control {
            border-radius: 8px;
            
            &:before {
                border-color: rgba(0, 0, 0, 0.12);
            }
            
            &:hover:before {
                border-color: #2e74b5;
            }
        }
        
        &.q-field--outlined {
            .q-field__control {
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
                transition: all 0.2s ease;
                
                &:hover {
                    box-shadow: 0 2px 6px rgba(46, 116, 181, 0.1);
                }
            }
        }
        
        &.q-field--focused {
            .q-field__control {
                box-shadow: 0 0 0 2px rgba(46, 116, 181, 0.2);
            }
        }
        
        .body--dark & {
            .q-field__control {
                background-color: rgba(255, 255, 255, 0.05);
                border-color: rgba(255, 255, 255, 0.12);
                
                &:hover {
                    border-color: #2e74b5;
                    background-color: rgba(255, 255, 255, 0.08);
                }
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