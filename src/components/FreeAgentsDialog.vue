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
            <!-- Dialog chrome: header (icon/title/close), the same convention used by
                 PlayerDetailDialog/UpgradeFinderDialog — an icon, a title, then a close
                 button, all in normal flow. -->
            <div class="dialog-chrome">
                <div class="dialog-chrome__header">
                    <q-icon name="person_off" class="dialog-chrome__icon" />
                    <div class="dialog-chrome__title">
                        Free Agents (Values in {{ currencySymbol }})
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
                    <SectionCard title="Available Free Agents" icon="list" class="free-agents-table-container">
                        <template #actions>
                            <q-chip
                                color="primary"
                                text-color="white"
                                :label="`${freeAgents.length} players`"
                            />
                        </template>

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
                    </SectionCard>
                </div>

                <!-- Empty State -->
                <EmptyState
                    v-else-if="!loading"
                    icon="search_off"
                    title="No free agents found"
                    description="All players appear to be under contract with clubs"
                />

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
import { defineComponent, onMounted, ref, watch } from 'vue'
import EmptyState from './layout/EmptyState.vue'
import SectionCard from './layout/SectionCard.vue'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

export default defineComponent({
  name: 'FreeAgentsDialog',
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
          .filter((player) => {
            // Check if player has no club (various possible formats)
            const club = player.club
            return (
              !club ||
              club === '' ||
              club === '-' ||
              club === 'Free Agent' ||
              club === 'Unattached' ||
              club.toLowerCase().includes('free') ||
              club.toLowerCase().includes('unattached')
            )
          })
          .sort((a, b) => {
            // Sort by Overall rating (highest first)
            const overallA = Number(a.Overall) || 0
            const overallB = Number(b.Overall) || 0
            return overallB - overallA
          })

        freeAgents.value = freePlayersList
      } catch (_error) {
        qInstance.notify({
          message: 'Error finding free agents. Please try again.',
          color: 'negative',
          icon: 'error',
        })
      } finally {
        loading.value = false
      }
    }

    const handlePlayerSelected = (player) => {
      selectedPlayer.value = player
      showPlayerDetail.value = true
    }

    const handleTeamSelected = (_teamName) => {
      // For free agents, we don't need team selection functionality
      // but we need to provide the handler for PlayerDataTable compatibility
    }

    // Watchers
    watch(
      () => props.show,
      async (newShow) => {
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
      (newPlayers) => {
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
      handleTeamSelected,
    }
  },
})
</script>

<style lang="scss" scoped>
.free-agents-dialog {
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

.free-agents-dialog {
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
    }

    @media (max-width: 480px) {
        .q-card-section {
            padding: 12px;
        }
    }
}
</style>