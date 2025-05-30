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

                <!-- Age Tabs -->
                <q-tabs
                    v-model="selectedAge"
                    dense
                    class="text-grey"
                    active-color="primary"
                    indicator-color="primary"
                    align="left"
                    :class="qInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
                >
                    <q-tab
                        v-for="age in ages"
                        :key="age"
                        :name="age"
                        :label="`Age ${age}`"
                        class="q-px-md"
                    />
                </q-tabs>

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
                        Top {{ currentWonderkids.length }} players aged {{ selectedAge }}
                        <span v-if="maxTransferValue || maxSalary" class="text-caption text-grey-6">
                            (filtered)
                        </span>
                    </div>
                    
                    <q-card
                        class="wonderkids-table-container"
                        :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                    >
                        <q-card-section>
                            <PlayerDataTable
                                :key="`wonderkids-${selectedAge}`"
                                :players="currentWonderkids"
                                :loading="loading"
                                @player-selected="handlePlayerSelected"
                                @team-selected="handleTeamSelected"
                                :currency-symbol="currencySymbol"
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
    const selectedAge = ref(21)
    const maxTransferValue = ref(null)
    const maxSalary = ref(null)
    const loading = ref(false)
    const selectedPlayer = ref(null)
    const showPlayerDetail = ref(false)
    const wonderkidsByAge = ref({})

    // Constants
    const ages = [15, 16, 17, 18, 19, 20, 21]

    // Computed
    const currentWonderkids = computed(() => {
      const result = wonderkidsByAge.value[selectedAge.value] || []
      return result
    })

    // Methods
    const getWonderkidsForAge = (age, allPlayers) => {
      const playersOfAge = allPlayers
        .filter(player => {
          const playerAge = Number(player.age)
          const matches = playerAge === age
          if (matches && age === 20) {
          }
          return matches
        })
        .filter(player => {
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
        .slice(0, 25)
      return playersOfAge
    }

    const loadWonderkids = async () => {
      loading.value = true
      try {
        const allPlayers = props.players
        const newWonderkidsByAge = {}

        for (const age of ages) {
          newWonderkidsByAge[age] = getWonderkidsForAge(age, allPlayers)
        }

        wonderkidsByAge.value = newWonderkidsByAge
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

    // Simplified age watcher - just log when it changes
    watch(selectedAge, (_newAge, _oldAge) => {})

    // Initialize when component mounts
    onMounted(() => {
      if (props.show && props.players.length > 0) {
        loadWonderkids()
      }
    })

    return {
      qInstance,
      selectedAge,
      maxTransferValue,
      maxSalary,
      loading,
      selectedPlayer,
      showPlayerDetail,
      ages,
      currentWonderkids,
      onFiltersChanged,
      handlePlayerSelected,
      handleTeamSelected
    }
  }
})
</script>

<style lang="scss" scoped>
.wonderkids-dialog {
    .wonderkids-table-container {
        border: none;
        box-shadow: none;
        
        :deep(.q-table) {
            // Override the hardcoded height in PlayerDataTable to allow full height
            height: auto !important;
            max-height: calc(100vh - 400px); // Allow for dialog header, filters, and tabs
            
            // Ensure virtual scroll works properly with dynamic height
            .q-table__middle {
                max-height: calc(100vh - 400px);
            }
        }
    }
}

.q-tab {
    min-width: 80px;
}

.text-subtitle1 {
    font-weight: 500;
}

// Ensure the dialog content area has proper spacing
.q-card-section {
    &:last-child {
        padding-bottom: 24px;
    }
}
</style> 