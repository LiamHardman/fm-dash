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
            class="upgrade-finder-dialog"
        >
            <q-card-section
                class="row items-center q-pb-none card-header"
            >
                <q-icon name="manage_search" size="md" class="q-mr-sm" />
                <div class="text-h6">
                    Upgrade Finder (Values in {{ currencySymbol }})
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
                <div class="row q-col-gutter-x-md q-col-gutter-y-sm q-mb-md">
                    <div class="col-12 col-md-6 col-lg-3">
                        <q-select
                            v-model="teamName"
                            :options="teamOptions"
                            label="Team Name"
                            outlined
                            dense
                            use-input
                            hide-selected
                            fill-input
                            input-debounce="300"
                            @filter="filterTeams"
                            :rules="[(val) => !!val || 'Team name is required']"
                            clearable
                            @clear="
                                teamName = null;
                                selectedTeamPlayer = null;
                                selectedRole = null; // Clear role when team clears
                                teamPlayersForSelection = [];
                            "
                            :label-color="
                                qInstance.dark.isActive ? 'grey-4' : ''
                            "
                            :input-class="
                                qInstance.dark.isActive ? 'text-grey-3' : ''
                            "
                            :popup-content-class="
                                qInstance.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        />
                    </div>

                    <div class="col-12 col-md-6 col-lg-3">
                        <q-select
                            v-model="selectedPosition"
                            :options="positionFilterOptions"
                            label="Position / Group"
                            dense
                            outlined
                            emit-value
                            map-options
                            :rules="[(val) => !!val || 'Position is required']"
                            clearable
                            @clear="
                                selectedPosition = null;
                                selectedTeamPlayer = null;
                                selectedRole = null; // Clear role when position clears
                                teamPlayersForSelection = [];
                            "
                            @update:model-value="onPositionOrTeamChange"
                            :label-color="
                                qInstance.dark.isActive ? 'grey-4' : ''
                            "
                            :popup-content-class="
                                qInstance.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        />
                    </div>
                    <div class="col-12 col-md-6 col-lg-3">
                        <q-select
                            v-model="selectedRole"
                            :options="roleOptionsForSelectedPosition"
                            label="Role"
                            dense
                            outlined
                            emit-value
                            map-options
                            clearable
                            @clear="selectedRole = null"
                            :disable="
                                !selectedPosition ||
                                roleOptionsForSelectedPosition.length <= 1
                            "
                            :hint="
                                !selectedPosition
                                    ? 'Select position first'
                                    : roleOptionsForSelectedPosition.length <= 1
                                      ? 'No specific roles for this position'
                                      : ''
                            "
                            :label-color="
                                qInstance.dark.isActive ? 'grey-4' : ''
                            "
                            :popup-content-class="
                                qInstance.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        >
                            <template v-slot:no-option>
                                <q-item>
                                    <q-item-section class="text-grey">
                                        {{
                                            !selectedPosition
                                                ? "Select position first"
                                                : "No roles available"
                                        }}
                                    </q-item-section>
                                </q-item>
                            </template>
                        </q-select>
                    </div>

                    <div class="col-12 col-md-6 col-lg-3">
                        <q-select
                            v-model="selectedTeamPlayer"
                            :options="teamPlayersForSelection"
                            label="Select Player for Upgrade Base"
                            option-label="name"
                            option-value="name"
                            map-options
                            emit-value
                            dense
                            outlined
                            clearable
                            :disable="
                                !teamName ||
                                !selectedPosition ||
                                teamPlayersForSelection.length === 0
                            "
                            :hint="
                                selectedTeamPlayer
                                    ? `Base Overall (${selectedRole ? getRoleShortName(selectedRole) : getPositionShortName(selectedPosition)}): ${getBaseOverallFromSelectedPlayer()}`
                                    : 'Select a player to set base overall'
                            "
                            :label-color="
                                qInstance.dark.isActive ? 'grey-4' : ''
                            "
                            :popup-content-class="
                                qInstance.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        >
                            <template v-slot:option="scope">
                                <q-item
                                    v-bind="scope.itemProps"
                                    :dark="qInstance.dark.isActive"
                                >
                                    <q-item-section>
                                        <q-item-label>{{
                                            scope.opt.name
                                        }}</q-item-label>
                                        <q-item-label caption
                                            >Overall ({{
                                                selectedRole
                                                    ? getRoleShortName(selectedRole)
                                                    : getPositionShortName(
                                                          selectedPosition,
                                                      )
                                            }}):
                                            {{
                                                getPlayerOverallForRoleOrPosition(
                                                    scope.opt,
                                                    selectedRole,
                                                    selectedPosition,
                                                )
                                            }}</q-item-label
                                        >
                                    </q-item-section>
                                </q-item>
                            </template>
                            <template v-slot:no-option>
                                <q-item :dark="qInstance.dark.isActive">
                                    <q-item-section class="text-grey">
                                        {{
                                            teamName && selectedPosition
                                                ? "No players in this team/position"
                                                : "Select team and position first"
                                        }}
                                    </q-item-section>
                                </q-item>
                            </template>
                        </q-select>
                    </div>

                    <div class="col-12 col-md-6 col-lg-3">
                        <div>
                            <div
                                class="text-caption q-mb-xs slider-label"
                                :class="
                                    qInstance.dark.isActive
                                        ? 'text-grey-4'
                                        : 'text-grey-7'
                                "
                            >
                                Upgrade By: {{ upgradeByValue }}
                            </div>
                            <q-slider
                                v-model="upgradeByValue"
                                :min="-10"
                                :max="10"
                                :step="1"
                                label
                                label-always
                                color="primary"
                                :dark="qInstance.dark.isActive"
                                :disable="!selectedTeamPlayer"
                                class="q-px-sm"
                            />
                        </div>
                    </div>
                    <div class="col-12 col-md-6 col-lg-3 filter-item-container">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Maximum Age:
                            {{
                                maxAgeFilter === ageSliderMax
                                    ? "Any"
                                    : maxAgeFilter
                            }}
                            <q-btn
                                flat
                                dense
                                icon="clear"
                                size="sm"
                                @click="maxAgeFilter = ageSliderMax"
                                v-if="maxAgeFilter < ageSliderMax"
                                class="q-ml-xs"
                                round
                                :text-color="
                                    qInstance.dark.isActive
                                        ? 'grey-5'
                                        : 'grey-7'
                                "
                            >
                                <q-tooltip>Clear age filter (Any)</q-tooltip>
                            </q-btn>
                        </div>
                        <q-slider
                            v-model="maxAgeFilter"
                            :min="ageSliderMin"
                            :max="ageSliderMax"
                            :step="1"
                            label
                            label-always
                            :label-value="
                                maxAgeFilter +
                                (maxAgeFilter === ageSliderMax ? '+' : '') +
                                ' yrs'
                            "
                            color="primary"
                            :dark="qInstance.dark.isActive"
                            class="q-px-sm"
                        />
                    </div>

                    <div class="col-12 col-md-6 col-lg-3 filter-item-container">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Max Transfer Value ({{ currencySymbol }}):
                            <q-btn
                                flat
                                dense
                                icon="clear"
                                size="sm"
                                @click="
                                    maxTransferValueFilter =
                                        computedMaxSliderTransferValue
                                "
                                v-if="
                                    maxTransferValueFilter <
                                        computedMaxSliderTransferValue &&
                                    props.players &&
                                    props.players.length > 0
                                "
                                class="q-ml-xs"
                                round
                                :text-color="
                                    qInstance.dark.isActive
                                        ? 'grey-5'
                                        : 'grey-7'
                                "
                            >
                                <q-tooltip>Clear value filter (Any)</q-tooltip>
                            </q-btn>
                        </div>
                        <q-slider
                            v-model="maxTransferValueFilter"
                            :min="computedMinSliderTransferValue"
                            :max="computedMaxSliderTransferValue"
                            :step="computedStepSliderTransferValue"
                            label
                            label-always
                            :label-value="formattedMaxTransferValueLabel"
                            color="primary"
                            :dark="qInstance.dark.isActive"
                            :disable="
                                !props.players || props.players.length === 0
                            "
                            class="q-px-sm"
                        />
                    </div>

                    <div class="col-12 col-md-6 col-lg-3 filter-item-container">
                        <div
                            class="text-caption q-mb-xs slider-label"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-grey-7'
                            "
                        >
                            Max Salary ({{ currencySymbol }}):
                            <q-btn
                                flat
                                dense
                                icon="clear"
                                size="sm"
                                @click="
                                    maxSalaryFilter = computedMaxSliderSalary
                                "
                                v-if="
                                    maxSalaryFilter < computedMaxSliderSalary &&
                                    props.players &&
                                    props.players.length > 0
                                "
                                class="q-ml-xs"
                                round
                                :text-color="
                                    qInstance.dark.isActive
                                        ? 'grey-5'
                                        : 'grey-7'
                                "
                            >
                                <q-tooltip>Clear salary filter (Any)</q-tooltip>
                            </q-btn>
                        </div>
                        <q-slider
                            v-model="maxSalaryFilter"
                            :min="computedMinSliderSalary"
                            :max="computedMaxSliderSalary"
                            :step="computedStepSliderSalary"
                            label
                            label-always
                            :label-value="formattedMaxSalaryLabel"
                            color="primary"
                            :dark="qInstance.dark.isActive"
                            :disable="
                                !props.players || props.players.length === 0
                            "
                            class="q-px-sm"
                        />
                    </div>
                </div>

                <div class="row q-col-gutter-md">
                    <div class="col-12">
                        <q-btn
                            color="primary"
                            icon="search"
                            label="Find Upgrades"
                            class="full-width q-py-sm"
                            @click="findUpgrades"
                            :loading="loading"
                            :disable="
                                !teamName ||
                                !selectedPosition ||
                                !selectedTeamPlayer ||
                                loading
                            "
                        />
                    </div>
                </div>
            </q-card-section>

            <q-card-section v-if="showResults" class="q-mt-md results-section">
                <q-separator :dark="qInstance.dark.isActive" class="q-mb-md" />
                <div
                    class="text-h6 q-mb-md"
                    :class="
                        qInstance.dark.isActive ? 'text-grey-2' : 'text-grey-9'
                    "
                >
                    Results
                </div>

                <div v-if="selectedTeamPlayerObject" class="q-mb-lg">
                    <div
                        class="text-subtitle1 q-mb-sm"
                        :class="
                            qInstance.dark.isActive
                                ? 'text-grey-3'
                                : 'text-grey-8'
                        "
                    >
                        Baseline Player:
                    </div>
                    <q-card
                        flat
                        bordered
                        :class="
                            qInstance.dark.isActive
                                ? 'bg-grey-8 text-grey-3'
                                : 'bg-blue-grey-1 text-blue-grey-10'
                        "
                    >
                        <q-item>
                            <q-item-section avatar>
                                <q-avatar>
                                    <img
                                        v-if="
                                            selectedTeamPlayerObject.nationality_iso
                                        "
                                        :src="`https://flagcdn.com/w40/${selectedTeamPlayerObject.nationality_iso.toLowerCase()}.png`"
                                        :alt="
                                            selectedTeamPlayerObject.nationality ||
                                            'Flag'
                                        "
                                        class="nationality-flag-dialog"
                                    />
                                    <q-icon v-else name="person" />
                                </q-avatar>
                            </q-item-section>
                            <q-item-section>
                                <q-item-label class="text-weight-bold">{{
                                    selectedTeamPlayerObject.name
                                }}</q-item-label>
                                <q-item-label
                                    caption
                                    :class="
                                        qInstance.dark.isActive
                                            ? 'text-grey-5'
                                            : 'text-blue-grey-7'
                                    "
                                >
                                    {{ selectedTeamPlayerObject.position }} |
                                    Age: {{ selectedTeamPlayerObject.age }} |
                                    Club: {{ selectedTeamPlayerObject.club }}
                                </q-item-label>
                            </q-item-section>
                            <q-item-section side top>
                                <q-item-label
                                    caption
                                    :class="
                                        qInstance.dark.isActive
                                            ? 'text-grey-5'
                                            : 'text-blue-grey-7'
                                    "
                                    >Overall ({{
                                        selectedRole
                                            ? getRoleShortName(selectedRole)
                                            : getPositionShortName(
                                                  selectedPosition,
                                              )
                                    }})</q-item-label
                                >
                                <div
                                    class="attribute-value fifa-stat-value text-h6"
                                    :class="
                                        getUnifiedRatingClass(
                                            getBaseOverallFromSelectedPlayer(),
                                            100,
                                        )
                                    "
                                >
                                    {{ getBaseOverallFromSelectedPlayer() }}
                                </div>
                            </q-item-section>
                        </q-item>
                        <q-card-section
                            class="q-pt-none"
                            :class="
                                qInstance.dark.isActive
                                    ? 'text-grey-4'
                                    : 'text-blue-grey-8'
                            "
                        >
                            Target Overall for Upgrades:
                            <span
                                class="text-weight-bold attribute-value"
                                :class="
                                    getUnifiedRatingClass(
                                        targetOverallForSearch,
                                        100,
                                    )
                                "
                            >
                                {{ targetOverallForSearch }}
                            </span>
                            (Base {{ getBaseOverallFromSelectedPlayer() }} +
                            Upgrade By {{ upgradeByValue }})
                        </q-card-section>
                    </q-card>
                </div>

                <div
                    class="text-subtitle1 q-mb-sm"
                    v-if="upgradePlayers.length > 0"
                    :class="
                        qInstance.dark.isActive ? 'text-grey-3' : 'text-grey-8'
                    "
                >
                    Potential upgrades ({{ upgradePlayers.length }} players
                    found):
                </div>

                <PlayerDataTable
                    v-if="upgradePlayers.length > 0"
                    :players="processedUpgradePlayers"
                    :loading="loading"
                    @player-selected="handlePlayerSelectedForDetailView"
                    :is-goalkeeper-view="upgradeFinderIsGoalkeeperView"
                    :currency-symbol="currencySymbol"
                    :dataset-id="datasetId"
                />

                <q-banner
                    v-else-if="showResults && !loading && !initialLoad"
                    class="q-mt-md"
                    :class="
                        qInstance.dark.isActive
                            ? 'bg-blue-grey-8 text-blue-grey-2'
                            : 'bg-info text-white'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="info" />
                    </template>
                    No upgrades found matching all criteria. Try adjusting
                    filters.
                </q-banner>
                <q-banner
                    v-else-if="
                        showResults &&
                        !loading &&
                        initialLoad &&
                        !selectedTeamPlayer
                    "
                    class="q-mt-md"
                    :class="
                        qInstance.dark.isActive
                            ? 'bg-orange-9 text-white'
                            : 'bg-amber text-dark'
                    "
                >
                    <template v-slot:avatar>
                        <q-icon name="warning" />
                    </template>
                    Please select a team, position, and a player from that team
                    to serve as the upgrade baseline.
                </q-banner>
            </q-card-section>
            <q-inner-loading :showing="loading">
                <q-spinner-gears size="50px" color="primary" />
            </q-inner-loading>
        </q-card>
    </q-dialog>

    <PlayerDetailDialog
        :player="playerForDetailView"
        :show="showPlayerDetailDialog"
        @close="showPlayerDetailDialog = false"
        :currency-symbol="currencySymbol"
    />
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch } from 'vue'
import { usePlayerStore } from '@/stores/playerStore' // Corrected Import Path
import { formatCurrency } from '@/utils/currencyUtils'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

// From PlayerFilters.vue for consistency
const AGE_SLIDER_MIN = 15
const AGE_SLIDER_MAX = 50
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
  'ST'
]

export default {
  name: 'UpgradeFinderDialog',
  components: { PlayerDataTable, PlayerDetailDialog },
  props: {
    show: { type: Boolean, default: false },
    players: { type: Array, required: true },
    currencySymbol: { type: String, default: '$' },
    datasetId: { type: String, default: null }
  },
  emits: ['close'],
  setup(props) {
    const $q = useQuasar()
    const playerStore = usePlayerStore()
    const teamName = ref(null)
    const teamOptions = ref([])
    const allTeamNamesCache = ref([])

    const selectedPosition = ref(null)
    const selectedRole = ref(null)
    const selectedTeamPlayer = ref(null)
    const teamPlayersForSelection = ref([])

    const upgradeByValue = ref(1)

    const ageSliderMin = AGE_SLIDER_MIN
    const ageSliderMax = AGE_SLIDER_MAX
    const maxAgeFilter = ref(ageSliderMax)

    const maxTransferValueFilter = ref(null)
    const dynamicMinTransferValue = ref(0)
    const dynamicMaxTransferValue = ref(100000000)

    const maxSalaryFilter = ref(null)
    const dynamicMinSalary = ref(0)
    const dynamicMaxSalary = ref(1000000)

    const loading = ref(false)
    const showResults = ref(false)
    const initialLoad = ref(true)

    const upgradePlayers = ref([])
    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)

    const populateAllTeamNames = () => {
      if (!props.players) {
        allTeamNamesCache.value = []
        teamOptions.value = []
        return
      }
      const uniqueTeams = new Set()
      for (const player of props.players) {
        if (player.club && player.club.trim() !== '') {
          uniqueTeams.add(player.club)
        }
      }
      allTeamNamesCache.value = Array.from(uniqueTeams).sort()
      teamOptions.value = allTeamNamesCache.value
    }

    const updateTransferValueSliderBounds = () => {
      if (!props.players || props.players.length === 0) {
        dynamicMinTransferValue.value = 0
        dynamicMaxTransferValue.value = 100000000
        maxTransferValueFilter.value = dynamicMaxTransferValue.value
        return
      }
      let minVal = Number.POSITIVE_INFINITY
      let maxVal = 0
      for (const p of props.players) {
        if (typeof p.transferValueAmount === 'number') {
          minVal = Math.min(minVal, p.transferValueAmount)
          maxVal = Math.max(maxVal, p.transferValueAmount)
        }
      }
      dynamicMinTransferValue.value = minVal === Number.POSITIVE_INFINITY ? 0 : Math.max(0, minVal)
      dynamicMaxTransferValue.value =
        maxVal === 0 && minVal === Number.POSITIVE_INFINITY ? 100000000 : maxVal
      if (
        maxTransferValueFilter.value === null ||
        maxTransferValueFilter.value > dynamicMaxTransferValue.value ||
        maxTransferValueFilter.value < dynamicMinTransferValue.value
      ) {
        maxTransferValueFilter.value = dynamicMaxTransferValue.value
      }
    }

    const updateSalarySliderBounds = () => {
      if (!props.players || props.players.length === 0) {
        dynamicMinSalary.value = 0
        dynamicMaxSalary.value = 1000000
        maxSalaryFilter.value = dynamicMaxSalary.value
        return
      }
      let minVal = Number.POSITIVE_INFINITY
      let maxVal = 0
      for (const p of props.players) {
        if (typeof p.wageAmount === 'number') {
          minVal = Math.min(minVal, p.wageAmount)
          maxVal = Math.max(maxVal, p.wageAmount)
        }
      }
      dynamicMinSalary.value = minVal === Number.POSITIVE_INFINITY ? 0 : Math.max(0, minVal)
      dynamicMaxSalary.value =
        maxVal === 0 && minVal === Number.POSITIVE_INFINITY ? 1000000 : maxVal
      if (
        maxSalaryFilter.value === null ||
        maxSalaryFilter.value > dynamicMaxSalary.value ||
        maxSalaryFilter.value < dynamicMinSalary.value
      ) {
        maxSalaryFilter.value = dynamicMaxSalary.value
      }
    }

    onMounted(async () => {
      if (playerStore.allAvailableRoles.length === 0 && playerStore.currentDatasetId) {
        await playerStore.fetchAllAvailableRoles()
      }
      populateAllTeamNames()
      updateTransferValueSliderBounds()
      updateSalarySliderBounds()
      maxAgeFilter.value = ageSliderMax
    })

    watch(
      () => props.players,
      newPlayers => {
        populateAllTeamNames()
        updateTransferValueSliderBounds()
        updateSalarySliderBounds()
        if (newPlayers && newPlayers.length > 0) {
          if (
            maxTransferValueFilter.value > dynamicMaxTransferValue.value ||
            maxTransferValueFilter.value < dynamicMinTransferValue.value
          ) {
            maxTransferValueFilter.value = dynamicMaxTransferValue.value
          }
        } else {
          allTeamNamesCache.value = []
          teamOptions.value = []
          dynamicMinTransferValue.value = 0
          dynamicMaxTransferValue.value = 100000000
          maxTransferValueFilter.value = dynamicMaxTransferValue.value
        }
      },
      { immediate: true, deep: true }
    )

    const positionFilterOptions = computed(() => {
      const options = [{ label: 'Any Position Group', value: null }]
      for (const pos of orderedShortPositions) {
        options.push({ label: pos, value: pos })
      }
      return options
    })

    const roleOptionsForSelectedPosition = computed(() => {
      if (
        !selectedPosition.value ||
        !playerStore.allAvailableRoles ||
        playerStore.allAvailableRoles.length === 0
      ) {
        return [{ label: 'Any Role', value: null }]
      }
      const roles = playerStore.allAvailableRoles
        .filter(roleFullName => roleFullName.startsWith(`${selectedPosition.value} - `))
        .map(roleFullName => ({
          label: roleFullName,
          value: roleFullName
        }))
        .sort((a, b) => a.label.localeCompare(b.label))
      return [{ label: 'Any Role', value: null }, ...roles]
    })

    const getRoleShortName = fullRoleName => {
      if (!fullRoleName) return ''
      const parts = fullRoleName.split(' - ')
      return parts.length > 1 ? parts[1] : fullRoleName
    }
    const getPositionShortName = shortPos => {
      return shortPos || ''
    }

    const filterTeams = (val, update) => {
      if (val === '') {
        update(() => {
          teamOptions.value = allTeamNamesCache.value
        })
        return
      }
      update(() => {
        const needle = val.toLowerCase()
        teamOptions.value = allTeamNamesCache.value.filter(
          team => team.toLowerCase().indexOf(needle) > -1
        )
      })
    }

    const onPositionOrTeamChange = () => {
      selectedTeamPlayer.value = null
      selectedRole.value = null
      updateTeamPlayersForSelection()
    }

    const updateTeamPlayersForSelection = () => {
      if (teamName.value && selectedPosition.value && props.players) {
        teamPlayersForSelection.value = props.players
          .filter(player => {
            if (player.club !== teamName.value) return false
            return player.shortPositions?.includes(selectedPosition.value)
          })
          .sort((a, b) => {
            const overallA = getPlayerOverallForRoleOrPosition(
              a,
              selectedRole.value,
              selectedPosition.value
            )
            const overallB = getPlayerOverallForRoleOrPosition(
              b,
              selectedRole.value,
              selectedPosition.value
            )
            return (overallB || 0) - (overallA || 0)
          })
      } else {
        teamPlayersForSelection.value = []
      }
    }

    watch([teamName, selectedPosition, selectedRole], updateTeamPlayersForSelection)

    const getPlayerOverallForRoleOrPosition = (player, role, position) => {
      if (!player) return 0
      if (role) {
        const roleData = player.roleSpecificOveralls?.find(r => r.roleName === role)
        return roleData ? roleData.score : 0
      }
      if (position) {
        let maxOverallForPosition = 0
        if (player.roleSpecificOveralls) {
          for (const rso of player.roleSpecificOveralls) {
            if (rso.roleName.startsWith(`${position} - `)) {
              if (rso.score > maxOverallForPosition) {
                maxOverallForPosition = rso.score
              }
            }
          }
        }
        return maxOverallForPosition > 0 ? maxOverallForPosition : player.Overall || 0
      }
      return player.Overall || 0
    }

    const getBaseOverallFromSelectedPlayer = () => {
      if (!selectedTeamPlayer.value) return null
      const player = teamPlayersForSelection.value.find(p => p.name === selectedTeamPlayer.value)
      if (!player) return null
      return getPlayerOverallForRoleOrPosition(player, selectedRole.value, selectedPosition.value)
    }

    const selectedTeamPlayerObject = computed(() => {
      if (!selectedTeamPlayer.value) return null
      return teamPlayersForSelection.value.find(p => p.name === selectedTeamPlayer.value) || null
    })

    const targetOverallForSearch = computed(() => {
      const base = getBaseOverallFromSelectedPlayer()
      if (base === null) return null
      return base + upgradeByValue.value
    })

    const computedMinSliderTransferValue = computed(() => dynamicMinTransferValue.value)
    const computedMaxSliderTransferValue = computed(() => dynamicMaxTransferValue.value)

    const computedStepSliderTransferValue = computed(() => {
      const range = computedMaxSliderTransferValue.value - computedMinSliderTransferValue.value
      if (range <= 0) return 10000
      if (range < 100000) return 5000
      if (range < 1000000) return 25000
      if (range < 10000000) return 100000
      if (range < 50000000) return 250000
      return 500000
    })

    const formattedMaxTransferValueLabel = computed(() => {
      if (maxTransferValueFilter.value === computedMaxSliderTransferValue.value) return 'Any'
      return formatCurrency(maxTransferValueFilter.value, props.currencySymbol)
    })

    const computedMinSliderSalary = computed(() => dynamicMinSalary.value)
    const computedMaxSliderSalary = computed(() => dynamicMaxSalary.value)

    const computedStepSliderSalary = computed(() => {
      const range = computedMaxSliderSalary.value - computedMinSliderSalary.value
      if (range <= 0) return 1000
      if (range < 50000) return 500
      if (range < 250000) return 2500
      if (range < 1000000) return 5000
      if (range < 10000000) return 25000
      return 50000
    })

    const formattedMaxSalaryLabel = computed(() => {
      if (maxSalaryFilter.value === computedMaxSliderSalary.value) return 'Any'
      return formatCurrency(maxSalaryFilter.value, props.currencySymbol)
    })

    const findUpgrades = async () => {
      if (!selectedTeamPlayer.value) {
        upgradePlayers.value = []
        showResults.value = true
        initialLoad.value = false
        return
      }
      if (!props.players) {
        loading.value = false
        return
      }

      loading.value = true
      showResults.value = true
      initialLoad.value = false
      const baseOverall = getBaseOverallFromSelectedPlayer()
      if (baseOverall === null) {
        loading.value = false
        upgradePlayers.value = []
        return
      }

      const targetOverall = baseOverall + upgradeByValue.value
      const currentMaxTransferValue = maxTransferValueFilter.value
      const currentMaxAge = maxAgeFilter.value
      const currentMaxSalary = maxSalaryFilter.value

      await new Promise(resolve => setTimeout(resolve, 300))

      try {
        upgradePlayers.value = props.players
          .filter(player => {
            if (player.club === teamName.value) return false
            if (player.transfer_value && player.transfer_value.toLowerCase() === 'not for sale')
              return false
            if (!player.shortPositions || !player.shortPositions.includes(selectedPosition.value))
              return false

            const playerOverallForContext = getPlayerOverallForRoleOrPosition(
              player,
              selectedRole.value,
              selectedPosition.value
            )
            if ((playerOverallForContext || 0) < targetOverall) return false

            if (
              currentMaxAge < ageSliderMax &&
              (Number.parseInt(player.age, 10) || 0) > currentMaxAge
            )
              return false
            if (
              currentMaxTransferValue < computedMaxSliderTransferValue.value &&
              (player.transferValueAmount || 0) > currentMaxTransferValue
            )
              return false
            if (
              currentMaxSalary < computedMaxSliderSalary.value &&
              (player.wageAmount || 0) > currentMaxSalary
            )
              return false

            return true
          })
          .sort((a, b) => {
            const overallA = getPlayerOverallForRoleOrPosition(
              a,
              selectedRole.value,
              selectedPosition.value
            )
            const overallB = getPlayerOverallForRoleOrPosition(
              b,
              selectedRole.value,
              selectedPosition.value
            )
            return (overallB || 0) - (overallA || 0)
          })
      } catch (_error) {
      } finally {
        loading.value = false
      }
    }

    const processedUpgradePlayers = computed(() => {
      return upgradePlayers.value.map(player => {
        const displayOverall = getPlayerOverallForRoleOrPosition(
          player,
          selectedRole.value,
          selectedPosition.value
        )
        return {
          ...player,
          Overall: displayOverall // This 'Overall' will be used by PlayerDataTable
        }
      })
    })

    const handlePlayerSelectedForDetailView = player => {
      // Ensure we pass the original player object, not the one with potentially modified 'Overall'
      const originalPlayer = props.players.find(
        p => p.name === player.name && p.club === player.club
      )
      playerForDetailView.value = originalPlayer || player
      showPlayerDetailDialog.value = true
    }

    const getUnifiedRatingClass = (value, maxScale) => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue) || value === null || value === undefined || value === '-')
        return 'rating-na'
      const percentage = (numValue / maxScale) * 100
      if (percentage >= 90) return 'rating-tier-6'
      if (percentage >= 80) return 'rating-tier-5'
      if (percentage >= 70) return 'rating-tier-4'
      if (percentage >= 55) return 'rating-tier-3'
      if (percentage >= 40) return 'rating-tier-2'
      return 'rating-tier-1'
    }

    const upgradeFinderIsGoalkeeperView = computed(() => selectedPosition.value === 'GK')

    watch(
      () => props.show,
      newValue => {
        if (!newValue) {
          teamName.value = null
          selectedPosition.value = null
          selectedRole.value = null
          selectedTeamPlayer.value = null
          teamPlayersForSelection.value = []
          upgradeByValue.value = 1
          maxAgeFilter.value = ageSliderMax
          if (props.players && props.players.length > 0) {
            maxTransferValueFilter.value = computedMaxSliderTransferValue.value
          } else {
            maxTransferValueFilter.value = 100000000
          }
          if (props.players && props.players.length > 0) {
            maxSalaryFilter.value = computedMaxSliderSalary.value
          } else {
            maxSalaryFilter.value = 1000000
          }
          showResults.value = false
          upgradePlayers.value = []
          loading.value = false
          initialLoad.value = true
        } else {
          if (playerStore.allAvailableRoles.length === 0 && playerStore.currentDatasetId) {
            playerStore.fetchAllAvailableRoles()
          }
          populateAllTeamNames()
          updateTransferValueSliderBounds()
          updateSalarySliderBounds()
          maxAgeFilter.value = ageSliderMax
          maxTransferValueFilter.value = computedMaxSliderTransferValue.value
          maxSalaryFilter.value = computedMaxSliderSalary.value
        }
      }
    )

    return {
      qInstance: $q,
      teamName,
      teamOptions,
      filterTeams,
      selectedPosition,
      positionFilterOptions,
      selectedRole,
      roleOptionsForSelectedPosition,
      getRoleShortName,
      getPositionShortName,
      selectedTeamPlayer,
      teamPlayersForSelection,
      getBaseOverallFromSelectedPlayer,
      selectedTeamPlayerObject,
      targetOverallForSearch,
      upgradeByValue,
      maxAgeFilter,
      ageSliderMin,
      ageSliderMax,
      maxTransferValueFilter,
      computedMinSliderTransferValue,
      computedMaxSliderTransferValue,
      computedStepSliderTransferValue,
      formattedMaxTransferValueLabel,
      maxSalaryFilter,
      computedMinSliderSalary,
      computedMaxSliderSalary,
      computedStepSliderSalary,
      formattedMaxSalaryLabel,
      loading,
      showResults,
      initialLoad,
      upgradePlayers,
      processedUpgradePlayers, // Use processed list for table
      findUpgrades,
      getUnifiedRatingClass,
      playerForDetailView,
      showPlayerDetailDialog,
      handlePlayerSelectedForDetailView,
      props,
      upgradeFinderIsGoalkeeperView,
      onPositionOrTeamChange,
      getPlayerOverallForRoleOrPosition
    }
  }
}
</script>

<style lang="scss" scoped>
.upgrade-finder-dialog {
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

    // Slider styling
    .slider-label {
        font-weight: 500;
        color: #374151;
        margin-bottom: 0.5rem;
        
        .body--dark & {
            color: #d1d5db;
        }
    }

    :deep(.q-slider) {
        .q-slider__track-container {
            .q-slider__track {
                background: rgba(46, 116, 181, 0.2);
            }
            
            .q-slider__selection {
                background: #2e74b5;
            }
        }
        
        .q-slider__thumb {
            background: #2e74b5;
            border: 2px solid white;
            box-shadow: 0 2px 8px rgba(46, 116, 181, 0.4);
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

    // Select dropdown styling
    :deep(.q-menu) {
        border-radius: 8px;
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            background-color: #374151 !important;
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
            
            .q-item {
                color: #d1d5db;
                
                &:hover,
                &.q-item--active {
                    background-color: rgba(46, 116, 181, 0.2) !important;
                    color: white;
                }
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
        
        &.q-btn--outline {
            border-width: 2px;
            
            &:hover {
                background-color: rgba(46, 116, 181, 0.1);
            }
        }
    }

    // Enhanced table styling
    :deep(.q-table) {
        border-radius: 8px;
        overflow: hidden;
        
        .q-table__top {
            padding: 1rem;
            background: linear-gradient(135deg, rgba(46, 116, 181, 0.03) 0%, rgba(46, 116, 181, 0.01) 100%);
            
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

    // Banner styling
    .q-banner {
        border-radius: $border-radius;
        margin-bottom: 1rem;
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
        }
    }

    @media (max-width: 480px) {
        .card-header {
            padding: 0.75rem;
        }
        
        .q-card-section {
            padding: 0.75rem;
        }
    }
}

// Filter item container styling
.filter-item-container {
    .slider-label {
        font-weight: 500;
        color: #374151;
        margin-bottom: 0.5rem;
        
        .body--dark & {
            color: #d1d5db;
        }
    }
}

// Results section styling
.results-section {
    .q-card {
        border-radius: $border-radius;
        box-shadow: $card-shadow;
        
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.02);
        }
    }
}
</style>
