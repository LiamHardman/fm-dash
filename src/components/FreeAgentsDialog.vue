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
            class="free-agents-dialog"
        >
            <q-card-section
                class="row items-center q-pb-none card-header"
            >
                <q-icon name="person_off" size="md" class="q-mr-sm" />
                <div class="text-h6">
                    Free Agents (Values in {{ currencySymbol }})
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
                <!-- Loading State -->
                <div v-if="loading" class="text-center q-my-xl">
                    <q-spinner-dots color="primary" size="3em" />
                    <div class="text-h6 q-mt-md">
                        Finding free agents...
                    </div>
                    <div class="text-caption q-mt-sm text-grey-6">
                        Analyzing players without clubs...
                    </div>
                </div>

                <!-- Results Table -->
                <div v-if="freeAgents.length > 0 && !loading">
                    <q-card 
                        class="free-agents-table-container" 
                        flat
                        bordered
                    >
                        <q-card-section>
                            <div class="row items-center q-mb-md">
                                <div class="text-subtitle1">
                                    <q-icon name="list" class="q-mr-sm" />
                                    Available Free Agents
                                </div>
                                <q-space />
                                <q-chip 
                                    color="primary" 
                                    text-color="white"
                                    :label="`${freeAgents.length} players`"
                                />
                            </div>
                            
                            <PlayerDataTable
                                :players="freeAgents"
                                :loading="loading"
                                @player-selected="handlePlayerSelected"
                                @team-selected="handleTeamSelected"
                                :currency-symbol="currencySymbol"
                                :dataset-id="datasetId"
                                :default-sort-field="'Overall'"
                                :default-sort-direction="'desc'"
                            />
                        </q-card-section>
                    </q-card>
                </div>

                <!-- Empty State -->
                <div v-else-if="!loading" class="text-center q-my-xl">
                    <q-icon name="search_off" size="4em" color="grey-5" />
                    <div class="text-h6 q-mt-md text-grey-6">
                        No free agents found
                    </div>
                    <div class="text-body2 text-grey-5 q-mt-sm">
                        All players appear to be under contract with clubs
                    </div>
                </div>

                <!-- Info Section (Expandable) -->
                <q-expansion-item
                    v-if="!loading"
                    icon="info_outline"
                    label="About Free Agents"
                    class="q-mt-md"
                    :class="qInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
                >
                    <q-card 
                        flat 
                        :class="qInstance.dark.isActive ? 'bg-grey-8' : 'bg-blue-1'"
                    >
                        <q-card-section>
                            <div class="text-body2">
                                <div class="q-mb-sm"><strong>Free Agents:</strong></div>
                                <ul class="q-pl-md">
                                    <li><strong>Definition:</strong> Players without a current club contract</li>
                                    <li><strong>Opportunity:</strong> Can be signed without transfer fees</li>
                                    <li><strong>Sorting:</strong> Displayed by highest overall rating first</li>
                                    <li><strong>Salary:</strong> Only need to cover wage costs</li>
                                </ul>
                                <div class="q-mt-sm text-caption text-grey-6">
                                    <em>Perfect for budget-conscious managers looking for immediate squad improvements without transfer costs</em>
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
import PlayerDetailDialog from './PlayerDetailDialog.vue'
import PlayerDataTable from './PlayerDataTable.vue'

export default defineComponent({
  name: 'FreeAgentsDialog',
  components: {
    PlayerDetailDialog,
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
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const freeAgents = ref([])

    // Methods
    const findFreeAgents = async () => {
      loading.value = true
      
      try {
        // Filter players without clubs and sort by overall rating
        const freePlayersList = props.players
          .filter(player => {
            // Check if player has no club (various possible formats)
            const club = player.club
            return !club || 
                   club === '' || 
                   club === '-' || 
                   club === 'Free Agent' ||
                   club === 'Unattached' ||
                   club.toLowerCase().includes('free') ||
                   club.toLowerCase().includes('unattached')
          })
          .sort((a, b) => {
            // Sort by Overall rating (highest first)
            const overallA = Number(a.Overall) || 0
            const overallB = Number(b.Overall) || 0
            return overallB - overallA
          })
        
        freeAgents.value = freePlayersList
      } catch (error) {
        console.error('Error finding free agents:', error)
        qInstance.notify({
          message: 'Error finding free agents. Please try again.',
          color: 'negative',
          icon: 'error'
        })
      } finally {
        loading.value = false
      }
    }

    const handlePlayerSelected = (player) => {
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const handleTeamSelected = (teamName) => {
      // For free agents, we don't need team selection functionality
      // but we need to provide the handler for PlayerDataTable compatibility
    }

    // Watchers
    watch(
      () => props.show,
      async newShow => {
        if (newShow && props.players.length > 0) {
          // Auto-search when dialog opens
          await findFreeAgents()
        } else if (!newShow) {
          // Reset values when dialog closes
          freeAgents.value = []
        }
      }
    )

    watch(
      () => props.players,
      newPlayers => {
        if (props.show && newPlayers.length > 0) {
          findFreeAgents()
        }
      }
    )

    // Initialize when component mounts
    onMounted(() => {
      if (props.show && props.players.length > 0) {
        findFreeAgents()
      }
    })

    return {
      qInstance,
      loading,
      selectedPlayer,
      showPlayerDetail,
      freeAgents,
      findFreeAgents,
      handlePlayerSelected,
      handleTeamSelected
    }
  }
})
</script>

<style lang="scss" scoped>
.free-agents-dialog {
    border-radius: $border-radius;
    box-shadow: $card-shadow;
    border: 1px solid rgba(0, 0, 0, 0.04);
    
    .body--dark & {
        background-color: #1e293b !important;
        border: 1px solid rgba(255, 255, 255, 0.1);
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    }

    .card-header {
        background: linear-gradient(135deg, #7c3aed 0%, #8b5cf6 100%);
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

    // Button styling
    .q-btn {
        border-radius: 8px;
        font-weight: 500;
        text-transform: none;
        
        &.q-btn--unelevated {
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            
            &:hover {
                box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
                transform: translateY(-1px);
            }
            
            .body--dark & {
                box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
                
                &:hover {
                    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.4);
                }
            }
        }
    }

    // Enhanced table styling
    :deep(.q-table) {
        border-radius: 8px;
        overflow: hidden;
        
        .q-table__top {
            padding: 1rem;
            background: linear-gradient(135deg, rgba(124, 58, 237, 0.03) 0%, rgba(124, 58, 237, 0.01) 100%);
            
            .body--dark & {
                background: rgba(255, 255, 255, 0.02);
            }
        }
        
        .q-table__container {
            border-radius: 0 0 8px 8px;
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
                    background-color: rgba(124, 58, 237, 0.04);
                }
                
                .body--dark & {
                    border-bottom-color: rgba(255, 255, 255, 0.05);
                    
                    &:hover {
                        background-color: rgba(255, 255, 255, 0.03);
                    }
                }
            }
        }
    }

    // Card improvements
    .q-card {
        border-radius: $border-radius;
        box-shadow: $card-shadow;
        border: 1px solid rgba(0, 0, 0, 0.04);
        
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.02);
            border-color: rgba(255, 255, 255, 0.08);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
        }
    }

    // Free agents table container specific styling
    .free-agents-table-container {
        background: white;
        
        .body--dark & {
            background: #1e293b !important;
            border-color: rgba(255, 255, 255, 0.1) !important;
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
    }

    @media (max-width: 480px) {
        .q-card-section {
            padding: 12px;
        }
    }
}
</style> 