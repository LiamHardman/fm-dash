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
                <!-- Team Selection Row -->
                <div class="row q-col-gutter-x-md q-col-gutter-y-sm q-mb-md">
                    <div class="col-12 col-md-6 col-lg-4">
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
                                selectedRole = null;
                                teamPlayersForSelection = [];
                                selectedFormationKey = null;
                                squadComposition = {};
                            "
                            :label-color="
                                $q.dark.isActive ? 'grey-4' : ''
                            "
                            :input-class="
                                $q.dark.isActive ? 'text-grey-3' : ''
                            "
                            :popup-content-class="
                                $q.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        />
                    </div>

                    <div class="col-12 col-md-6 col-lg-4">
                        <q-select
                            v-model="selectedFormationKey"
                            :options="formationOptions"
                            label="Formation"
                            outlined
                            dense
                            emit-value
                            map-options
                            clearable
                            @clear="selectedFormationKey = null"
                            :disable="!teamName"
                            :label-color="
                                $q.dark.isActive ? 'grey-4' : ''
                            "
                            :popup-content-class="
                                $q.dark.isActive
                                    ? 'bg-grey-8 text-white'
                                    : 'bg-white text-dark'
                            "
                        />
                    </div>

                    <div class="col-12 col-md-6 col-lg-4">
                        <div>
                            <div
                                class="text-caption q-mb-xs slider-label"
                                :class="
                                    $q.dark.isActive
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
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>
                    </div>
                </div>

                <!-- Main Content Layout -->
                <div class="row q-col-gutter-md">
                    <!-- Left Side - Filters -->
                    <div class="col-12 col-lg-6">
                        <q-card class="filters-card">
                            <q-card-section>
                                <div class="card-header">
                                    <h3 class="card-title">
                                        <q-icon name="filter_list" class="card-icon" />
                                        Upgrade Filters
                                    </h3>
                                    <p class="card-subtitle">Configure your search criteria</p>
                                </div>

                                <div class="row q-col-gutter-y-md">
                                    <div class="col-12 col-md-6">
                                        <q-select
                                            v-model="selectedPosition"
                                            :options="positionFilterOptions"
                                            label="Position / Group"
                                            dense
                                            outlined
                                            emit-value
                                            map-options
                                            clearable
                                            @clear="
                                                selectedPosition = null;
                                                selectedTeamPlayer = null;
                                                selectedRole = null;
                                                teamPlayersForSelection = [];
                                            "
                                            @update:model-value="onPositionOrTeamChange"
                                            :label-color="
                                                $q.dark.isActive ? 'grey-4' : ''
                                            "
                                            :popup-content-class="
                                                $q.dark.isActive
                                                    ? 'bg-grey-8 text-white'
                                                    : 'bg-white text-dark'
                                            "
                                        />
                                    </div>

                                    <div class="col-12 col-md-6">
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
                                                $q.dark.isActive ? 'grey-4' : ''
                                            "
                                            :popup-content-class="
                                                $q.dark.isActive
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

                                    <div class="col-12 col-md-6">
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
                                                $q.dark.isActive ? 'grey-4' : ''
                                            "
                                            :popup-content-class="
                                                $q.dark.isActive
                                                    ? 'bg-grey-8 text-white'
                                                    : 'bg-white text-dark'
                                            "
                                        >
                                            <template v-slot:option="scope">
                                                <q-item
                                                    v-bind="scope.itemProps"
                                                    :dark="$q.dark.isActive"
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
                                                <q-item :dark="$q.dark.isActive">
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

                                    <div class="col-12 col-md-6">
                                        <div>
                                            <div
                                                class="text-caption q-mb-xs slider-label"
                                                :class="
                                                    $q.dark.isActive
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
                                                        $q.dark.isActive
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
                                                :dark="$q.dark.isActive"
                                                class="q-px-sm"
                                            />
                                        </div>
                                    </div>

                                    <div class="col-12 col-md-6">
                                        <div>
                                            <div
                                                class="text-caption q-mb-xs slider-label"
                                                :class="
                                                    $q.dark.isActive
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
                                                        $q.dark.isActive
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
                                                :dark="$q.dark.isActive"
                                                :disable="
                                                    !props.players || props.players.length === 0
                                                "
                                                class="q-px-sm"
                                            />
                                        </div>
                                    </div>

                                    <div class="col-12 col-md-6">
                                        <div>
                                            <div
                                                class="text-caption q-mb-xs slider-label"
                                                :class="
                                                    $q.dark.isActive
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
                                                        $q.dark.isActive
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
                                                :dark="$q.dark.isActive"
                                                :disable="
                                                    !props.players || props.players.length === 0
                                                "
                                                class="q-px-sm"
                                            />
                                        </div>
                                    </div>
                                </div>

                                <div class="row q-col-gutter-md q-mt-md">
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
                        </q-card>
                        
                        <!-- Results Section -->
                        <div v-if="showResults" class="q-mt-md">
                            <q-card class="results-card">
                                <q-card-section>
                                    <div
                                        class="text-h6 q-mb-md"
                                        :class="
                                            $q.dark.isActive ? 'text-grey-2' : 'text-grey-9'
                                        "
                                    >
                                        Potential upgrades ({{ upgradePlayers.length }} players
                                        found):
                                    </div>

                                    <!-- Player Cards Grid -->
                                    <div v-if="upgradePlayers.length > 0" class="upgrade-players-grid">
                                        <div 
                                            v-for="player in processedUpgradePlayers" 
                                            :key="player.id"
                                            class="upgrade-player-card-container"
                                        >
                                            <q-card 
                                                class="upgrade-player-card"
                                                flat 
                                                bordered
                                                clickable
                                                @click="handlePlayerSelectedForDetailView(player)"
                                            >
                                                <q-card-section class="player-card-header">
                                                    <div class="player-basic-info">
                                                        <!-- Player Face -->
                                                        <div class="player-face-section">
                                                            <q-avatar
                                                                size="60px"
                                                                :color="$q.dark.isActive ? 'grey-7' : 'grey-4'"
                                                                :text-color="$q.dark.isActive ? 'grey-4' : 'grey-7'"
                                                            >
                                                                <q-icon name="person" size="32px" />
                                                            </q-avatar>
                                                        </div>
                                                        
                                                        <!-- Player Details -->
                                                        <div class="player-details">
                                                            <div class="player-name">{{ player.name }}</div>
                                                            <div class="player-club">{{ player.club || 'Free Agent' }}</div>
                                                            <div class="player-position">{{ player.position }}</div>
                                                        </div>
                                                        
                                                        <!-- Overall Rating -->
                                                        <div class="player-overall">
                                                            <div class="overall-label">Overall</div>
                                                            <div class="overall-value" :class="getUnifiedRatingClass(player.overall, 100)">{{ player.overall }}</div>
                                                        </div>
                                                    </div>
                                                </q-card-section>
                                                
                                                <q-card-section class="player-attributes-section">
                                                    <div class="attributes-grid">
                                                        <!-- Key Attributes -->
                                                        <div class="key-attributes">
                                                            <div 
                                                                v-for="attr in getKeyAttributes(player)" 
                                                                :key="attr.key"
                                                                class="attribute-item"
                                                            >
                                                                <span class="attribute-name">{{ attr.name }}</span>
                                                                                                                        <span 
                                                            class="attribute-value"
                                                            :class="getUnifiedRatingClass(attr.value, 100)"
                                                        >
                                                            {{ attr.value }}
                                                        </span>
                                                            </div>
                                                        </div>
                                                        
                                                        <!-- Player Stats -->
                                                        <div class="player-stats">
                                                            <div class="stat-item">
                                                                <span class="stat-label">Age:</span>
                                                                <span class="stat-value">{{ player.age }}</span>
                                                            </div>
                                                            <div class="stat-item">
                                                                <span class="stat-label">Value:</span>
                                                                <span class="stat-value">{{ formatValue(player.transfer_value) }}</span>
                                                            </div>
                                                            <div class="stat-item">
                                                                <span class="stat-label">Wage:</span>
                                                                <span class="stat-value">{{ formatWage(player.wage) }}</span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </q-card-section>
                                            </q-card>
                                        </div>
                                    </div>

                                    <q-banner
                                        v-else-if="showResults && !loading && !initialLoad"
                                        class="q-mt-md"
                                        :class="
                                            $q.dark.isActive
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
                                            $q.dark.isActive
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
                            </q-card>
                        </div>
                    </div>

                    <!-- Right Side - Team & Formation Display -->
                    <div class="col-12 col-lg-6">
                        <div v-if="teamName && selectedFormationKey" class="team-formation-section">
                            <!-- Team Header -->
                            <q-card class="team-header-card q-mb-md">
                                <q-card-section>
                                    <div class="team-header-content">
                                        <div class="team-info">
                                            <TeamLogo 
                                                :team-name="teamName"
                                                :size="32"
                                                class="team-logo"
                                            />
                                            <div class="team-details">
                                                <h3 class="team-name">{{ teamName }}</h3>
                                                <div v-if="bestTeamAverageOverall !== null" class="team-rating">
                                                    <span class="rating-value">{{ bestTeamAverageOverall }}</span>
                                                    <span class="rating-label">Average Overall</span>
                                                </div>
                                            </div>
                                        </div>
                                        <div v-if="calculationMessage" class="calculation-status">
                                            <q-banner
                                                :class="calculationMessageClass"
                                                class="calculation-banner"
                                            >
                                                {{ calculationMessage }}
                                            </q-banner>
                                        </div>
                                    </div>
                                </q-card-section>
                            </q-card>

                            <!-- Formation Display -->
                            <q-card class="formation-card">
                                <q-card-section>
                                    <div class="card-header">
                                        <h3 class="card-title">
                                            <q-icon name="stadium" class="card-icon" />
                                            Formation View
                                        </h3>
                                        <p class="card-subtitle">Click on any position to find upgrades</p>
                                    </div>
                                    
                                    <div class="pitch-container">
                                        <PitchDisplay
                                            :formation="currentFormationLayout"
                                            :players="bestTeamPlayersForPitch"
                                            :disable-player-clicks="true"
                                            @position-click="handlePositionClick"
                                        />
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>

                        <!-- No Team/Formation State -->
                        <div v-else-if="!teamName" class="empty-state">
                            <q-card class="empty-state-card">
                                <q-card-section class="empty-state-content">
                                    <div class="empty-state-icon">
                                        <q-icon name="groups" size="4rem" />
                                    </div>
                                    <h3 class="empty-state-title">Select a Team</h3>
                                    <p class="empty-state-description">
                                        Choose a team to view their formation and find potential upgrades for each position.
                                    </p>
                                </q-card-section>
                            </q-card>
                        </div>

                        <!-- No Formation State -->
                        <div v-else-if="!selectedFormationKey" class="empty-state">
                            <q-card class="empty-state-card">
                                <q-card-section class="empty-state-content">
                                    <div class="empty-state-icon">
                                        <q-icon name="diagram" size="4rem" />
                                    </div>
                                    <h3 class="empty-state-title">Select a Formation</h3>
                                    <p class="empty-state-description">
                                        Choose a formation to see the team layout and identify positions for upgrades.
                                    </p>
                                </q-card-section>
                            </q-card>
                        </div>
                    </div>
                </div>
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
        :dataset-id="datasetId"
    />
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch, nextTick } from 'vue'
import { usePlayerStore } from '@/stores/playerStore' // Corrected Import Path
import { formatCurrency } from '@/utils/currencyUtils'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'
import PitchDisplay from './PitchDisplay.vue'
import TeamLogo from './TeamLogo.vue'
import { formations, getFormationLayout } from '@/utils/formations'
import { formationCache } from '@/utils/formationCache'

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

// Formation calculation constants (from TeamViewPage)
const MIN_SUITABILITY_THRESHOLD = 10

const positionSideMap = {
  'D (R)': ['DR'],
  'D (L)': ['DL'],
  'D (C)': ['DC'],
  'WB (R)': ['WBR'],
  'WB (L)': ['WBL'],
  'DM (C)': ['DM'],
  'M (R)': ['MR'],
  'M (L)': ['ML'],
  'M (C)': ['MC'],
  'AM (R)': ['AMR'],
  'AM (L)': ['AML'],
  'AM (C)': ['AMC'],
  'ST (C)': ['ST'],
  GK: ['GK']
}

const fallbackPositionMap = {
  'D (R)': ['DR', 'WBR', 'MR'],
  'D (L)': ['DL', 'WBL', 'ML'],
  'D (C)': ['DC', 'DM'],
  'WB (R)': ['WBR', 'DR', 'MR'],
  'WB (L)': ['WBL', 'DL', 'ML'],
  'DM (C)': ['DM', 'DC', 'MC'],
  'M (R)': ['MR', 'WBR', 'AMR'],
  'M (L)': ['ML', 'WBL', 'AML'],
  'M (C)': ['MC', 'DM'],
  'AM (R)': ['AMR', 'MR'],
  'AM (L)': ['AML', 'ML'],
  'AM (C)': ['AMC', 'MC'],
  'ST (C)': ['ST', 'AMC'],
  GK: ['GK']
}

// Additional constants from TeamViewPage
const fmSlotRoleMatcher = {
  GK: ['Goalkeeper'],
  'D (R)': ['Defender (Right)', 'Right Back'],
  'D (L)': ['Defender (Left)', 'Left Back'],
  'D (C)': ['Defender (Centre)', 'Centre Back'],
  'WB (R)': ['Wing-Back (Right)', 'Right Wing-Back'],
  'WB (L)': ['Wing-Back (Left)', 'Left Wing-Back'],
  'DM (C)': ['Defensive Midfielder (Centre)', 'Centre Defensive Midfielder'],
  'M (R)': ['Midfielder (Right)', 'Right Midfielder'],
  'M (L)': ['Midfielder (Left)', 'Left Midfielder'],
  'M (C)': ['Midfielder (Centre)', 'Centre Midfielder'],
  'AM (R)': ['Attacking Midfielder (Right)', 'Right Attacking Midfielder', 'Winger (Right)'],
  'AM (L)': ['Attacking Midfielder (Left)', 'Left Attacking Midfielder', 'Winger (Left)'],
  'AM (C)': ['Attacking Midfielder (Centre)', 'Centre Attacking Midfielder'],
  'ST (C)': ['Striker (Centre)', 'Striker']
}

const fmMatcherToRoleKeyPrefix = {
  GOALKEEPER: 'GK',
  SWEEPER: 'DC',
  'DEFENDER (RIGHT)': 'DR',
  'RIGHT BACK': 'DR',
  'DEFENDER (LEFT)': 'DL',
  'LEFT BACK': 'DL',
  'DEFENDER (CENTRE)': 'DC',
  'CENTRE BACK': 'DC',
  'WING-BACK (RIGHT)': 'WBR',
  'RIGHT WING-BACK': 'WBR',
  'WING-BACK (LEFT)': 'WBL',
  'LEFT WING-BACK': 'WBL',
  'DEFENSIVE MIDFIELDER (CENTRE)': 'DM',
  'CENTRE DEFENSIVE MIDFIELDER': 'DM',
  'MIDFIELDER (RIGHT)': 'MR',
  'RIGHT MIDFIELDER': 'MR',
  'MIDFIELDER (LEFT)': 'ML',
  'LEFT MIDFIELDER': 'ML',
  'MIDFIELDER (CENTRE)': 'MC',
  'CENTRE MIDFIELDER': 'MC',
  'ATTACKING MIDFIELDER (RIGHT)': 'AMR',
  'RIGHT ATTACKING MIDFIELDER': 'AMR',
  'WINGER (RIGHT)': 'AMR',
  'ATTACKING MIDFIELDER (LEFT)': 'AML',
  'LEFT ATTACKING MIDFIELDER': 'AML',
  'WINGER (LEFT)': 'AML',
  'ATTACKING MIDFIELDER (CENTRE)': 'AMC',
  'CENTRE ATTACKING MIDFIELDER': 'AMC',
  'STRIKER (CENTRE)': 'ST',
  STRIKER: 'ST'
}

export default {
  name: 'UpgradeFinderDialog',
  components: { PlayerDataTable, PlayerDetailDialog, PitchDisplay, TeamLogo },
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

    // Formation-related variables
    const selectedFormationKey = ref(null)
    const squadComposition = ref({})
    const bestTeamAverageOverall = ref(null)
    const calculationMessage = ref('')
    const calculationMessageClass = ref('')

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
        const transferValue = parseTransferValue(p.transfer_value)
        if (transferValue > 0) {
          minVal = Math.min(minVal, transferValue)
          maxVal = Math.max(maxVal, transferValue)
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
      if (props.players && props.players.length > 0) {
        // Check if players have the required data for formation calculation
        const samplePlayer = props.players[0]
      }
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

    const formationOptions = computed(() => {
      return Object.keys(formations).map(key => ({
        label: formations[key].name,
        value: key
      }))
    })

    const currentFormationLayout = computed(() => {
      if (!selectedFormationKey.value) {
        return []
      }
      return getFormationLayout(selectedFormationKey.value) || []
    })

    const bestTeamPlayersForPitch = computed(() => {
      const starters = {}
      if (!squadComposition.value || Object.keys(squadComposition.value).length === 0) {
        return starters
      }
      for (const slotId in squadComposition.value) {
        if (squadComposition.value[slotId] && squadComposition.value[slotId].length > 0) {
          const starterEntry = squadComposition.value[slotId][0]
          starters[slotId] = {
            ...starterEntry.player,
            Overall: starterEntry.overallInRole,
            exactPositionMatch: starterEntry.exactMatch
          }
        } else {
          starters[slotId] = null
        }
      }
      return starters
    })

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
            if (!player.club || player.club.toLowerCase() !== teamName.value.toLowerCase()) return false
            
            // Try both field names for compatibility
            const positions = player.shortPositions || player.short_positions || []
            const hasPosition = positions.includes(selectedPosition.value)
            
            return hasPosition
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

    // Also update team players when team changes (for formation calculation)
    const updateTeamPlayersForFormation = () => {
      if (teamName.value && props.players) {
        const teamPlayers = props.players.filter(player => 
          player.club && player.club.toLowerCase() === teamName.value.toLowerCase()
        )
        if (teamPlayers.length > 0 && selectedFormationKey.value) {
          calculateBestTeamAndDepth(teamPlayers)
        }
      }
    }

    watch([teamName, selectedPosition, selectedRole], updateTeamPlayersForSelection)

    // Formation watchers
    watch(selectedFormationKey, newKey => {
      if (newKey && teamName.value && props.players) {
        const teamPlayers = props.players.filter(player => 
          player.club && player.club.toLowerCase() === teamName.value.toLowerCase()
        )
        calculateBestTeamAndDepth(teamPlayers)
      } else {
        squadComposition.value = {}
        bestTeamAverageOverall.value = null
        calculationMessage.value = 'Select a team and formation.'
        calculationMessageClass.value = $q.dark.isActive ? 'text-grey-5' : 'text-grey-7'
      }
    })

    // Auto-select best formation when team changes
    watch(teamName, async (newTeamName) => {
      if (newTeamName && props.players) {
        // Get all players for the selected team (case-insensitive matching)
        const teamPlayers = props.players.filter(player => 
          player.club && player.club.toLowerCase() === newTeamName.toLowerCase()
        )
        if (teamPlayers.length > 0) {
          const bestFormation = calculateBestFormationForTeam(teamPlayers)
          if (bestFormation) {
            selectedFormationKey.value = bestFormation
            calculationMessage.value = `Auto-selected best formation: ${formations[bestFormation].name}. Calculating Best XI...`
            calculationMessageClass.value = $q.dark.isActive
              ? 'bg-info text-white'
              : 'bg-blue-2 text-primary'
          }
        }
      }
    })

    // Update formation when team changes
    watch(teamName, updateTeamPlayersForFormation)

    const getPlayerOverallForRoleOrPosition = (player, role, position) => {
      if (!player) return 0
      

      
      // Handle both array and object formats for roleSpecificOveralls
      const hasRoleOveralls = Array.isArray(player.roleSpecificOveralls)
        ? player.roleSpecificOveralls.length > 0
        : Object.keys(player.roleSpecificOveralls || {}).length > 0
      
      let result = 0
      
      if (role) {
        if (Array.isArray(player.roleSpecificOveralls)) {
          const roleData = player.roleSpecificOveralls.find(r => r.roleName === role)
          result = roleData ? roleData.score : (player.Overall || player.overall || 0)
        } else if (player.roleSpecificOveralls) {
          result = player.roleSpecificOveralls[role] || (player.Overall || player.overall || 0)
        } else {
          result = player.Overall || player.overall || 0
        }
      } else if (position) {
        let maxOverallForPosition = 0
        
        if (hasRoleOveralls) {
          if (Array.isArray(player.roleSpecificOveralls)) {
            for (const rso of player.roleSpecificOveralls) {
              if (rso.roleName.startsWith(`${position} - `)) {
                if (rso.score > maxOverallForPosition) {
                  maxOverallForPosition = rso.score
                }
              }
            }
          } else {
            for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
              if (roleName.startsWith(`${position} - `)) {
                if (score > maxOverallForPosition) {
                  maxOverallForPosition = score
                }
              }
            }
          }
        }
        
        result = maxOverallForPosition > 0 ? maxOverallForPosition : (player.Overall || player.overall || 0)
      } else {
        result = player.Overall || player.overall || 0
      }
      

      
      return result
    }

    // Formation calculation functions (from TeamViewPage)
    const getPlayerOverallForRole = (player, slotFormationRole) => {
      if (!player || !slotFormationRole) return 0

      let bestScoreForRole = 0

      if (!player.roleSpecificOveralls) {
        return player.Overall || 0
      }

      const hasRoleOveralls = Array.isArray(player.roleSpecificOveralls)
        ? player.roleSpecificOveralls.length > 0
        : Object.keys(player.roleSpecificOveralls).length > 0

      if (!hasRoleOveralls) {
        return player.Overall || 0
      }

      const upperSlotRoleOriginal = slotFormationRole.toUpperCase()
      const requiredPositions = positionSideMap[upperSlotRoleOriginal] || []

      if (player.shortPositions && player.shortPositions.length > 0) {
        const exactPositionMatches = player.shortPositions.filter(pos =>
          requiredPositions.includes(pos)
        )

        if (exactPositionMatches.length > 0) {
          if (Array.isArray(player.roleSpecificOveralls)) {
            for (const rso of player.roleSpecificOveralls) {
              const rsoBasePosition = rso.roleName.split(' - ')[0].trim()

              if (exactPositionMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, rso.score)
              }
            }
          } else {
            for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
              const rsoBasePosition = roleName.split(' - ')[0].trim()

              if (exactPositionMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, score)
              }
            }
          }

          if (bestScoreForRole === 0) {
            bestScoreForRole = Math.max(MIN_SUITABILITY_THRESHOLD, player.Overall || 0)
          }
        }
      }

      if (bestScoreForRole > 0) {
        return bestScoreForRole
      }

      const fallbackPositions = fallbackPositionMap[upperSlotRoleOriginal] || []

      if (player.shortPositions && player.shortPositions.length > 0) {
        const fallbackMatches = player.shortPositions.filter(pos => fallbackPositions.includes(pos))

        if (fallbackMatches.length > 0) {
          if (Array.isArray(player.roleSpecificOveralls)) {
            for (const rso of player.roleSpecificOveralls) {
              const rsoBasePosition = rso.roleName.split(' - ')[0].trim()

              if (fallbackMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, rso.score)
              }
            }
          } else {
            for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
              const rsoBasePosition = roleName.split(' - ')[0].trim()

              if (fallbackMatches.includes(rsoBasePosition)) {
                bestScoreForRole = Math.max(bestScoreForRole, score)
              }
            }
          }

          if (bestScoreForRole === 0) {
            bestScoreForRole = Math.max(MIN_SUITABILITY_THRESHOLD - 10, (player.Overall || 0) - 5)
          }
        }
      }

      if (bestScoreForRole === 0) {
        const upperSlotRole = slotFormationRole.toUpperCase()
        const fmPositionMatchers = fmSlotRoleMatcher[upperSlotRole] || [upperSlotRole]

        const targetRoleKeyPrefixes = fmPositionMatchers
          .map(matcher => fmMatcherToRoleKeyPrefix[matcher.toUpperCase()])
          .filter(prefix => !!prefix)
          .reduce((acc, val) => {
            if (!acc.includes(val)) {
              acc.push(val)
            }
            return acc
          }, [])

        if (Array.isArray(player.roleSpecificOveralls)) {
          for (const rso of player.roleSpecificOveralls) {
            const rsoBasePosition = rso.roleName.split(' - ')[0].trim()

            if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
              bestScoreForRole = Math.max(bestScoreForRole, rso.score)
            }
          }
        } else if (player.roleSpecificOveralls) {
          for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
            const rsoBasePosition = roleName.split(' - ')[0].trim()

            if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
              bestScoreForRole = Math.max(bestScoreForRole, score)
            }
          }
        }

        if (bestScoreForRole === 0) {
          bestScoreForRole = Math.max(0, (player.Overall || 0) - 10)
        }
      }

      return bestScoreForRole
    }

    const calculateBestFormationForTeam = (teamPlayers) => {
      if (!teamPlayers || teamPlayers.length === 0) {
        return null
      }

      // Check cache first
      const cacheKey = formationCache.generateKey(teamPlayers, 'team-best')
      const cachedResult = formationCache.get(cacheKey)
      if (cachedResult) {
        return cachedResult.bestFormationKey
      }

      let bestFormationKey = null
      let bestAverageOverall = 0

      // Test each formation to find the one with highest average overall
      for (const formationKey of Object.keys(formations)) {
        const formationLayoutForCalc = getFormationLayout(formationKey)
        if (!formationLayoutForCalc) continue

        const formationSlots = formationLayoutForCalc.flatMap(row => row.positions)
        const tempSquadComposition = {}

        // Initialize slots
        for (const slot of formationSlots) {
          tempSquadComposition[slot.id] = []
        }

        // Calculate player scores for each position in this formation
        const allPotentialPlayerAssignments = []
        for (const slot of formationSlots) {
          for (const player of teamPlayers) {
            const overallInRole = getPlayerOverallForRole(player, slot.role)

            if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
              const slotPositions = positionSideMap[slot.role.toUpperCase()] || []
              const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || []
              const playerPositions = player.shortPositions || []
              
              const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos))
              const isFallbackMatch = playerPositions.some(pos => fallbackPositions.includes(pos))

              if (isExactMatch || isFallbackMatch) {
                const assignment = {
                  player,
                  slotId: slot.id,
                  slotRole: slot.role,
                  overallInRole: overallInRole,
                  sortScore: overallInRole,
                  exactMatch: isExactMatch
                }

                if (isExactMatch) {
                  assignment.sortScore += 10000
                } else {
                  assignment.sortScore -= 5000
                }

                allPotentialPlayerAssignments.push(assignment)
              }
            }
          }
        }

        // Sort assignments by sort score
        allPotentialPlayerAssignments.sort((a, b) => b.sortScore - a.sortScore)

        const assignedPlayersToSlots = new Set()

        // Fill starting XI for this formation
        for (const slot of formationSlots) {
          for (const assignment of allPotentialPlayerAssignments) {
            if (
              assignment.slotId === slot.id &&
              !assignedPlayersToSlots.has(assignment.player.name)
            ) {
              tempSquadComposition[slot.id].push({
                player: assignment.player,
                overallInRole: assignment.overallInRole,
                exactMatch: assignment.exactMatch
              })
              assignedPlayersToSlots.add(assignment.player.name)
              break
            }
          }
        }

        // Calculate average overall for this formation
        let sumOfStartersOverall = 0
        let startersCount = 0
        let filledPositions = 0
        for (const slotPlayers of Object.values(tempSquadComposition)) {
          if (slotPlayers && slotPlayers.length > 0) {
            sumOfStartersOverall += slotPlayers[0].overallInRole
            startersCount++
            filledPositions++
          }
        }

        const hasEnoughPlayers = filledPositions >= 5
        
        if (startersCount > 0 && hasEnoughPlayers) {
          const averageOverall = sumOfStartersOverall / startersCount
          if (averageOverall > bestAverageOverall) {
            bestAverageOverall = averageOverall
            bestFormationKey = formationKey
          }
        }
      }

      // Cache the result
      if (bestFormationKey) {
        formationCache.set(cacheKey, {
          bestFormationKey,
          bestAverageOverall,
          teamName: teamName.value
        })
      }

      return bestFormationKey
    }

    const calculateBestTeamAndDepth = (teamPlayers) => {
      if (!selectedFormationKey.value || !teamPlayers || teamPlayers.length === 0) {
        squadComposition.value = {}
        bestTeamAverageOverall.value = null
        calculationMessage.value = selectedFormationKey.value
          ? 'No players in the selected team.'
          : 'Select a formation.'
        calculationMessageClass.value = 'bg-warning text-dark'
        return
      }

      // Check cache first for squad composition
      const cacheKey = formationCache.generateKey(
        teamPlayers,
        `team-depth-${selectedFormationKey.value}`
      )
      const cachedResult = formationCache.get(cacheKey)
      if (cachedResult) {
        squadComposition.value = cachedResult.squadComposition
        bestTeamAverageOverall.value = cachedResult.bestTeamAverageOverall
        calculationMessage.value = `Best XI & Depth calculated (cached). Average Overall: ${cachedResult.bestTeamAverageOverall}.`
        calculationMessageClass.value = $q.dark.isActive
          ? 'bg-positive text-white'
          : 'bg-green-2 text-positive'
        return
      }

      calculationMessage.value = 'Calculating best team and depth...'
      calculationMessageClass.value = $q.dark.isActive
        ? 'bg-info text-white'
        : 'bg-blue-2 text-primary'

      const tempSquadComposition = {}
      const formationLayoutForCalc = getFormationLayout(selectedFormationKey.value)
      if (!formationLayoutForCalc) {
        calculationMessage.value = 'Invalid formation selected.'
        calculationMessageClass.value = 'bg-negative text-white'
        return
      }

      const formationSlots = formationLayoutForCalc.flatMap(row => row.positions)

      // Initialize slots
      for (const slot of formationSlots) {
        tempSquadComposition[slot.id] = []
      }

      // Calculate player scores for each position
      const allPotentialPlayerAssignments = []
      for (const slot of formationSlots) {
        for (const player of teamPlayers) {
          const overallInRole = getPlayerOverallForRole(player, slot.role)

          if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
            const slotPositions = positionSideMap[slot.role.toUpperCase()] || []
            const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || []
            const playerPositions = player.shortPositions || []
            
            const isExactMatch = playerPositions.some(pos => slotPositions.includes(pos))
            const canPlayInPosition = isExactMatch || playerPositions.some(pos => fallbackPositions.includes(pos))

            if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
              const assignment = {
                player,
                slotId: slot.id,
                slotRole: slot.role,
                overallInRole: overallInRole,
                sortScore: overallInRole,
                exactMatch: isExactMatch
              }

              if (isExactMatch) {
                assignment.sortScore += 10000
              } else {
                assignment.sortScore -= 5000
              }

              allPotentialPlayerAssignments.push(assignment)
            }
          }
        }
      }

      // Sort assignments by the sort score
      allPotentialPlayerAssignments.sort((a, b) => b.sortScore - a.sortScore)

      const assignedPlayersToSlots = new Set()

      for (let depthIndex = 0; depthIndex < 3; depthIndex++) {
        // First pass: fill positions with exact matches
        for (const slot of formationSlots) {
          if (tempSquadComposition[slot.id].length === depthIndex) {
            for (const assignment of allPotentialPlayerAssignments) {
              if (
                assignment.slotId === slot.id &&
                assignment.exactMatch &&
                !assignedPlayersToSlots.has(assignment.player.name)
              ) {
                let alreadyStarterElsewhere = false
                if (depthIndex > 0) {
                  for (const sId in tempSquadComposition) {
                    if (
                      tempSquadComposition[sId].length > 0 &&
                      tempSquadComposition[sId][0].player.name === assignment.player.name
                    ) {
                      alreadyStarterElsewhere = true
                      break
                    }
                  }
                }

                if (!alreadyStarterElsewhere) {
                  tempSquadComposition[slot.id].push({
                    player: assignment.player,
                    overallInRole: assignment.overallInRole,
                    exactMatch: assignment.exactMatch
                  })
                  assignedPlayersToSlots.add(assignment.player.name)
                  break
                }
              }
            }
          }
        }

        // Second pass: fill remaining positions with fallback matches
        for (const slot of formationSlots) {
          if (tempSquadComposition[slot.id].length === depthIndex) {
            for (const assignment of allPotentialPlayerAssignments) {
              if (
                assignment.slotId === slot.id &&
                !assignedPlayersToSlots.has(assignment.player.name)
              ) {
                let alreadyStarterElsewhere = false
                if (depthIndex > 0) {
                  for (const sId in tempSquadComposition) {
                    if (
                      tempSquadComposition[sId].length > 0 &&
                      tempSquadComposition[sId][0].player.name === assignment.player.name
                    ) {
                      alreadyStarterElsewhere = true
                      break
                    }
                  }
                }

                if (!alreadyStarterElsewhere) {
                  tempSquadComposition[slot.id].push({
                    player: assignment.player,
                    overallInRole: assignment.overallInRole,
                    exactMatch: assignment.exactMatch
                  })
                  assignedPlayersToSlots.add(assignment.player.name)
                  break
                }
              }
            }
          }
        }
      }

      // Ensure each slot in tempSquadComposition is sorted by overallInRole descending
      for (const slotId in tempSquadComposition) {
        tempSquadComposition[slotId].sort((a, b) => b.overallInRole - a.overallInRole)
      }

      squadComposition.value = tempSquadComposition

      let sumOfStartersOverall = 0
      let startersCount = 0
      for (const slotPlayers of Object.values(squadComposition.value)) {
        if (slotPlayers && slotPlayers.length > 0) {
          sumOfStartersOverall += slotPlayers[0].overallInRole
          startersCount++
        }
      }

      if (startersCount > 0) {
        bestTeamAverageOverall.value = Math.round(sumOfStartersOverall / startersCount)
        calculationMessage.value = `Best XI & Depth calculated. Average Overall: ${bestTeamAverageOverall.value}.`
        calculationMessageClass.value = $q.dark.isActive
          ? 'bg-positive text-white'
          : 'bg-green-2 text-positive'
      } else {
        bestTeamAverageOverall.value = 0
        calculationMessage.value = 'Could not assign any suitable players to form a Best XI.'
        calculationMessageClass.value = $q.dark.isActive
          ? 'bg-negative text-white'
          : 'bg-red-2 text-negative'
      }

      // Cache the result
      if (bestTeamAverageOverall.value > 0) {
        formationCache.set(cacheKey, {
          squadComposition: squadComposition.value,
          bestTeamAverageOverall: bestTeamAverageOverall.value,
          teamName: teamName.value,
          formation: selectedFormationKey.value
        })
      }
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
              parseTransferValue(player.transfer_value) > currentMaxTransferValue
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

    const handlePlayerSelectedFromTeam = player => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    const handlePositionClick = (positionData) => {
      // Extract position information from the clicked position
      const { slotId, slotRole } = positionData
      
      // Translate formation position to searchable position
      const basePosition = translateFormationPositionToSearchable(slotRole)
      
      // Set the position filter to this position
      selectedPosition.value = basePosition
      
      // Find the player in this position from squad composition
      const positionPlayers = squadComposition.value[slotId]
      if (positionPlayers && positionPlayers.length > 0) {
        const playerInPosition = positionPlayers[0].player
        
        // Set the team player filter to this player
        selectedTeamPlayer.value = playerInPosition.name
        
        // Find the best role for this player in this position
        const bestRole = findBestRoleForPlayerInPosition(playerInPosition, basePosition)
        selectedRole.value = bestRole
        
        // If no specific role found, try to find any role for this position
        if (!bestRole && playerInPosition.roleSpecificOveralls) {
          if (Array.isArray(playerInPosition.roleSpecificOveralls)) {
            const anyRoleForPosition = playerInPosition.roleSpecificOveralls.find(rso => 
              rso.roleName.includes(basePosition)
            )
            if (anyRoleForPosition) {
              selectedRole.value = anyRoleForPosition.roleName
            }
          } else {
            const anyRoleForPosition = Object.keys(playerInPosition.roleSpecificOveralls).find(roleName => 
              roleName.includes(basePosition)
            )
            if (anyRoleForPosition) {
              selectedRole.value = anyRoleForPosition
            }
          }
        }
        
        // Automatically trigger upgrade search for this player after reactive updates
        nextTick(() => {
          findUpgrades()
        })
      } else {
        // No player in this position, just set position and clear others
        selectedRole.value = null
        selectedTeamPlayer.value = null
      }
      
      // Update the team players for selection with the new position
      updateTeamPlayersForSelection()
    }

    const translateFormationPositionToSearchable = (formationPosition) => {
      // Map formation positions to searchable positions
      const positionMap = {
        'GK': 'GK',
        'SW': 'SW',
        'D (L)': 'DL',
        'D (C)': 'DC',
        'D (R)': 'DR',
        'D (RLC)': 'DR',
        'M (L)': 'ML',
        'M (C)': 'MC',
        'M (R)': 'MR',
        'AM (L)': 'AML',
        'AM (C)': 'AMC',
        'AM (R)': 'AMR',
        'ST (L)': 'STL',
        'ST (C)': 'ST',
        'ST (R)': 'STR',
        'W (L)': 'WL',
        'W (R)': 'WR',
        'DM (L)': 'DML',
        'DM (C)': 'DMC',
        'DM (R)': 'DMR',
        'WB (L)': 'WBL',
        'WB (C)': 'WBC',
        'WB (R)': 'WBR',
        'WB (RL)': 'WBR'
      }
      
      const searchablePosition = positionMap[formationPosition]
      return searchablePosition || formationPosition
    }

    // Helper methods for player cards
    const getKeyAttributes = (player) => {
      const keyAttrs = []
      
      // Add key attributes based on position
      if (player.position === 'GK') {
        keyAttrs.push(
          { key: 'han', name: 'HAN', value: player.han || 0 },
          { key: 'ref', name: 'REF', value: player.ref || 0 },
          { key: 'kic', name: 'KIC', value: player.kic || 0 },
          { key: 'pos', name: 'POS', value: player.pos || 0 }
        )
      } else {
        // Use the specified order: PAC DRI, SHO DEF, PAS PHY
        keyAttrs.push(
          { key: 'pac', name: 'PAC', value: player.pac || 0 },
          { key: 'dri', name: 'DRI', value: player.dri || 0 },
          { key: 'sho', name: 'SHO', value: player.sho || 0 },
          { key: 'def', name: 'DEF', value: player.def || 0 },
          { key: 'pas', name: 'PAS', value: player.pas || 0 },
          { key: 'phy', name: 'PHY', value: player.phy || 0 }
        )
      }
      
      return keyAttrs
    }



    const formatWage = (wageString) => {
      if (!wageString) return 'N/A'
      
      // Extract the number from strings like "£13,750 p/w"
      const match = wageString.match(/£([\d,]+)/)
      if (!match) return wageString
      
      const number = parseInt(match[1].replace(/,/g, ''))
      if (number >= 1000) {
        return `£${(number / 1000).toFixed(2)}K`
      }
      return wageString
    }

    const parseTransferValue = (valueString) => {
      if (!valueString) return 0
      
      // Extract the upper bound from strings like "£28M - £34M"
      const match = valueString.match(/£([\d.]+)M/)
      if (match) {
        return parseFloat(match[1]) * 1000000
      }
      
      // Handle other formats
      const upperMatch = valueString.match(/£([\d,]+)/)
      if (upperMatch) {
        const number = parseInt(upperMatch[1].replace(/,/g, ''))
        if (number >= 1000000) {
          return (number / 1000000) * 1000000
        } else if (number >= 1000) {
          return (number / 1000) * 1000
        }
        return number
      }
      
      return 0
    }

    const formatValue = (valueString) => {
      if (!valueString) return 'N/A'
      
      // Extract the upper bound from strings like "£28M - £34M"
      const match = valueString.match(/£([\d.]+)M/)
      if (match) {
        return `£${match[1]}M`
      }
      
      // Handle other formats
      const upperMatch = valueString.match(/£([\d,]+)/)
      if (upperMatch) {
        const number = parseInt(upperMatch[1].replace(/,/g, ''))
        if (number >= 1000000) {
          return `£${(number / 1000000).toFixed(1)}M`
        } else if (number >= 1000) {
          return `£${(number / 1000).toFixed(0)}k`
        }
      }
      
      return valueString
    }

    const findBestRoleForPlayerInPosition = (player, position) => {
      if (!player || !position) return null
      
      let bestRole = null
      let bestScore = 0
      
      // Check if player has role-specific overalls
      if (player.roleSpecificOveralls) {
        if (Array.isArray(player.roleSpecificOveralls)) {
          // Handle array format
          for (const rso of player.roleSpecificOveralls) {
            if (rso.roleName.startsWith(`${position} - `)) {
              if (rso.score > bestScore) {
                bestScore = rso.score
                bestRole = rso.roleName
              }
            }
          }
        } else {
          // Handle object format
          for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
            if (roleName.startsWith(`${position} - `)) {
              if (score > bestScore) {
                bestScore = score
                bestRole = roleName
              }
            }
          }
        }
      }
      
      return bestRole
    }

    const getUnifiedRatingClass = (value, maxScale) => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue) || value === null || value === undefined || value === '-')
        return 'rating-na'
      
      // For FIFA stats (1-100 scale), calculate percentage correctly
      if (maxScale === 20) {
        const percentage = ((numValue - 1) / (maxScale - 1)) * 100
        if (percentage >= 90) return 'rating-tier-6'
        if (percentage >= 80) return 'rating-tier-5'
        if (percentage >= 70) return 'rating-tier-4'
        if (percentage >= 55) return 'rating-tier-3'
        if (percentage >= 40) return 'rating-tier-2'
        return 'rating-tier-1'
      } else {
        // For FIFA stats (1-100 scale) and overall ratings (0-100 scale)
        const percentage = (numValue / maxScale) * 100
        if (percentage >= 90) return 'rating-tier-6'
        if (percentage >= 80) return 'rating-tier-5'
        if (percentage >= 70) return 'rating-tier-4'
        if (percentage >= 55) return 'rating-tier-3'
        if (percentage >= 40) return 'rating-tier-2'
        return 'rating-tier-1'
      }
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
      $q,
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
      processedUpgradePlayers,
      findUpgrades,
      getUnifiedRatingClass,
      playerForDetailView,
      showPlayerDetailDialog,
      handlePlayerSelectedForDetailView,
      props,
      upgradeFinderIsGoalkeeperView,
      onPositionOrTeamChange,
      getPlayerOverallForRoleOrPosition,
      // Formation-related returns
      selectedFormationKey,
      formationOptions,
      currentFormationLayout,
      bestTeamPlayersForPitch,
      squadComposition,
      bestTeamAverageOverall,
      calculationMessage,
      calculationMessageClass,
      handlePlayerSelectedFromTeam,
              handlePositionClick,
        findBestRoleForPlayerInPosition,
        translateFormationPositionToSearchable,
        getKeyAttributes,
        getUnifiedRatingClass,
        formatWage,
        formatValue,
      parseTransferValue,
        formatCurrency,
      updateTeamPlayersForFormation
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

// Team formation section styling
.team-formation-section {
    .team-header-card {
        .team-header-content {
            display: flex;
            align-items: center;
            justify-content: space-between;
            
            .team-info {
                display: flex;
                align-items: center;
                gap: 1rem;
                
                .team-logo {
                    flex-shrink: 0;
                }
                
                .team-details {
                    .team-name {
                        margin: 0 0 0.5rem 0;
                        font-size: 1.5rem;
                        font-weight: 600;
                        color: #374151;
                        
                        .body--dark & {
                            color: #d1d5db;
                        }
                    }
                    
                    .team-rating {
                        display: flex;
                        align-items: center;
                        gap: 0.5rem;
                        
                        .rating-value {
                            font-size: 1.25rem;
                            font-weight: 600;
                            color: #2e74b5;
                            
                            .body--dark & {
                                color: #60a5fa;
                            }
                        }
                        
                        .rating-label {
                            font-size: 0.875rem;
                            color: #6b7280;
                            
                            .body--dark & {
                                color: #9ca3af;
                            }
                        }
                    }
                }
            }
            
            .calculation-status {
                flex-shrink: 0;
                
                .calculation-banner {
                    padding: 0.5rem 1rem;
                    border-radius: 6px;
                    font-size: 0.875rem;
                }
            }
        }
    }
    
    .formation-card {
        .pitch-container {
            max-width: 100%;
            margin: 1rem auto;
        }
    }
}

// Empty state styling
.empty-state {
    .empty-state-card {
        text-align: center;
        padding: 2rem;
        
        .empty-state-content {
            .empty-state-icon {
                margin-bottom: 1rem;
                color: #9ca3af;
                
                .body--dark & {
                    color: #6b7280;
                }
            }
            
            .empty-state-title {
                margin: 0 0 0.5rem 0;
                font-size: 1.25rem;
                font-weight: 600;
                color: #374151;
                
                .body--dark & {
                    color: #d1d5db;
                }
            }
            
            .empty-state-description {
                margin: 0;
                color: #6b7280;
                font-size: 0.875rem;
                
                .body--dark & {
                    color: #9ca3af;
                }
            }
        }
    }
}

// Filters card styling
.filters-card {
    .card-header {
        margin-bottom: 1.5rem;
        
        .card-title {
            margin: 0 0 0.25rem 0;
            font-size: 1.125rem;
            font-weight: 600;
            color: #374151;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            
            .body--dark & {
                color: #d1d5db;
            }
            
            .card-icon {
                color: #2e74b5;
                
                .body--dark & {
                    color: #60a5fa;
                }
            }
        }
        
        .card-subtitle {
            margin: 0;
            color: #6b7280;
            font-size: 0.875rem;
            
            .body--dark & {
                color: #9ca3af;
            }
        }
    }
}

// Player cards styling
.upgrade-players-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
    margin-top: 1rem;
}

.upgrade-player-card-container {
    .upgrade-player-card {
        transition: all 0.2s ease;
        border-radius: 12px;
        overflow: hidden;
        
        &:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
            
            .body--dark & {
                box-shadow: 0 8px 25px rgba(0, 0, 0, 0.4);
            }
        }
        
        .player-card-header {
            padding: 1rem;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            
            .body--dark & {
                background: linear-gradient(135deg, #4c51bf 0%, #553c9a 100%);
            }
            
            .player-basic-info {
                display: flex;
                align-items: center;
                gap: 1rem;
                
                .player-face-section {
                    flex-shrink: 0;
                }
                
                .player-details {
                    flex: 1;
                    
                    .player-name {
                        font-size: 1.125rem;
                        font-weight: 600;
                        margin: 0 0 0.25rem 0;
                        line-height: 1.2;
                    }
                    
                    .player-club {
                        font-size: 0.875rem;
                        opacity: 0.9;
                        margin: 0 0 0.125rem 0;
                    }
                    
                    .player-position {
                        font-size: 0.75rem;
                        opacity: 0.8;
                        margin: 0;
                        text-transform: uppercase;
                        letter-spacing: 0.5px;
                    }
                }
                
                .player-overall {
                    text-align: center;
                    flex-shrink: 0;
                    
                    .overall-label {
                        font-size: 0.75rem;
                        opacity: 0.8;
                        margin: 0 0 0.25rem 0;
                        text-transform: uppercase;
                        letter-spacing: 0.5px;
                    }
                    
                    .overall-value {
                        font-size: 1.5rem;
                        font-weight: 700;
                        margin: 0;
                        line-height: 1;
                        padding: 0.25rem 0.5rem;
                        border-radius: 6px;
                        min-width: 3rem;
                        text-align: center;
                        
                        &.rating-tier-6 {
                            background-color: #7e57c2;
                            color: white !important;
                            font-weight: 700;
                            border: 1px solid #5e35b1;
                            
                            .body--dark & {
                                background-color: #9575cd;
                                color: white !important;
                                border-color: #7e57c2;
                            }
                        }
                        
                        &.rating-tier-5 {
                            background-color: #26a69a;
                            color: white !important;
                            
                            .body--dark & {
                                background-color: #00897b;
                                color: white !important;
                            }
                        }
                        
                        &.rating-tier-4 {
                            background-color: #66bb6a;
                            color: white !important;
                            
                            .body--dark & {
                                background-color: #4caf50;
                                color: white !important;
                            }
                        }
                        
                        &.rating-tier-3 {
                            background-color: #42a5f5;
                            color: white !important;
                            
                            .body--dark & {
                                background-color: #2196f3;
                                color: white !important;
                            }
                        }
                        
                        &.rating-tier-2 {
                            background-color: #ffa726;
                            color: #333333 !important;
                            
                            .body--dark & {
                                background-color: #fb8c00;
                                color: white !important;
                            }
                        }
                        
                        &.rating-tier-1 {
                            background-color: #ef5350;
                            color: white !important;
                            
                            .body--dark & {
                                background-color: #e53935;
                                color: white !important;
                            }
                        }
                        
                        &.rating-na {
                            background-color: #bdbdbd;
                            color: #424242 !important;
                            
                            .body--dark & {
                                background-color: #424242;
                                color: #bdbdbd !important;
                            }
                        }
                    }
                }
            }
        }
        
        .player-attributes-section {
            padding: 1rem;
            
            .attributes-grid {
                display: grid;
                grid-template-columns: 1fr auto;
                gap: 1rem;
                align-items: start;
                
                .key-attributes {
                    display: grid;
                    grid-template-columns: repeat(2, 1fr);
                    gap: 0.5rem;
                    
                    .attribute-item {
                        display: flex;
                        justify-content: space-between;
                        align-items: center;
                        padding: 0.25rem 0;
                        
                        .attribute-name {
                            font-size: 0.75rem;
                            color: #6b7280;
                            font-weight: 500;
                            
                            .body--dark & {
                                color: #9ca3af;
                            }
                        }
                        
                        .attribute-value {
                            font-size: 0.875rem;
                            font-weight: 600;
                            padding: 0.125rem 0.375rem;
                            border-radius: 4px;
                            min-width: 2rem;
                            text-align: center;
                            
                            &.rating-tier-6 {
                                background-color: #7e57c2;
                                color: white !important;
                                font-weight: 700;
                                border: 1px solid #5e35b1;
                                
                                .body--dark & {
                                    background-color: #9575cd;
                                    color: white !important;
                                    border-color: #7e57c2;
                                }
                            }
                            
                            &.rating-tier-5 {
                                background-color: #26a69a;
                                color: white !important;
                                
                                .body--dark & {
                                    background-color: #00897b;
                                    color: white !important;
                                }
                            }
                            
                            &.rating-tier-4 {
                                background-color: #66bb6a;
                                color: white !important;
                                
                                .body--dark & {
                                    background-color: #4caf50;
                                    color: white !important;
                                }
                            }
                            
                            &.rating-tier-3 {
                                background-color: #42a5f5;
                                color: white !important;
                                
                                .body--dark & {
                                    background-color: #2196f3;
                                    color: white !important;
                                }
                            }
                            
                            &.rating-tier-2 {
                                background-color: #ffa726;
                                color: #333333 !important;
                                
                                .body--dark & {
                                    background-color: #fb8c00;
                                    color: white !important;
                                }
                            }
                            
                            &.rating-tier-1 {
                                background-color: #ef5350;
                                color: white !important;
                                
                                .body--dark & {
                                    background-color: #e53935;
                                    color: white !important;
                                }
                            }
                            
                            &.rating-na {
                                background-color: #bdbdbd;
                                color: #424242 !important;
                                
                                .body--dark & {
                                    background-color: #424242;
                                    color: #bdbdbd !important;
                                }
                            }
                        }
                    }
                }
                
                .player-stats {
                    display: flex;
                    flex-direction: column;
                    gap: 0.5rem;
                    padding-left: 1rem;
                    border-left: 1px solid #e5e7eb;
                    
                    .body--dark & {
                        border-left-color: #374151;
                    }
                    
                    .stat-item {
                        display: flex;
                        justify-content: space-between;
                        align-items: center;
                        gap: 0.5rem;
                        
                        .stat-label {
                            font-size: 0.75rem;
                            color: #6b7280;
                            
                            .body--dark & {
                                color: #9ca3af;
                            }
                        }
                        
                        .stat-value {
                            font-size: 0.875rem;
                            font-weight: 600;
                            color: #374151;
                            
                            .body--dark & {
                                color: #d1d5db;
                            }
                        }
                    }
                }
            }
        }
    }
}

// Responsive design for player cards
@media (max-width: 768px) {
    .upgrade-players-grid {
        grid-template-columns: 1fr;
    }
    
    .upgrade-player-card-container {
        .upgrade-player-card {
            .player-card-header {
                .player-basic-info {
                    .player-details {
                        .player-name {
                            font-size: 1rem;
                        }
                    }
                    
                    .player-overall {
                        .overall-value {
                            font-size: 1.25rem;
                        }
                    }
                }
            }
            
            .player-attributes-section {
                .attributes-grid {
                    grid-template-columns: 1fr;
                    gap: 0.75rem;
                    
                    .player-stats {
                        padding-left: 0;
                        border-left: none;
                        border-top: 1px solid #e5e7eb;
                        padding-top: 0.75rem;
                        
                        .body--dark & {
                            border-top-color: #374151;
                        }
                    }
                }
            }
        }
    }
}
</style>


