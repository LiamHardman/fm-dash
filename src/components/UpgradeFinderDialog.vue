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
                    <!-- Left Side - Filters and Results -->
                    <div class="col-12 col-lg-6">
                        <!-- Filters Card -->
                        <q-card class="filters-card q-mb-md">
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
                                                    ? `Base Overall (${selectedRole ? getRoleShortName(selectedRole) : getPositionShortName(selectedPosition)}): ${baseOverallForSelectedPlayer || 'Loading...'}`
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

                                <!-- Stat Filters Button -->
                                <div class="row q-col-gutter-md q-mt-md">
                                    <div class="col-12">
                                        <q-btn
                                            color="secondary"
                                            icon="tune"
                                            label="Stat Filters"
                                            class="full-width q-py-sm"
                                            @click="showStatFiltersModal = true"
                                            outline
                                        />
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
                        
                        <!-- Player Cards Section -->
                        <div v-if="showResults && upgradePlayers.length > 0" class="q-mt-md">
                            <q-card class="results-card">
                                <q-card-section>
                                    <div
                                        class="text-h6 q-mb-md"
                                        :class="
                                            $q.dark.isActive ? 'text-grey-2' : 'text-grey-9'
                                        "
                                    >
                                        <q-icon name="sports_soccer" class="q-mr-sm" />
                                        Potential Upgrades ({{ upgradePlayers.length }} players found)
                                    </div>
                                    
                                    <!-- Player Cards Grid -->
                                    <div class="player-cards-grid">
                                        <div 
                                            v-for="player in paginatedPlayers" 
                                            :key="`${player.name}-${player.club}`"
                                            class="player-card-container"
                                        >
                                            <PlayerCards 
                                                :player="player"
                                                :currency-symbol="currencySymbol"
                                                :nation-flag-url="getFlagUrl(player.nationality_iso || player.nationality_fifa_code)"
                                                :club-image-url="getTeamLogoUrl(player.club)"
                                                :player-face-url="getPlayerFaceUrl(player.name, player.club)"
                                                :dataset-id="props.datasetId"
                                                :selected-role="selectedRole"
                                                @click="handleCardClick(player)"
                                                class="clickable-card"
                                            />
                                        </div>
                                    </div>
                                    
                                    <!-- Pagination Controls if there are more than 6 players -->
                                    <div v-if="upgradePlayers.length > 6" class="text-center q-mt-md">
                                        <div class="row justify-center items-center q-gutter-sm">
                                            <q-btn
                                                color="primary"
                                                icon="chevron_left"
                                                :disable="currentPage === 1"
                                                @click="previousPage"
                                                outline
                                                size="sm"
                                            />
                                            <span class="text-caption q-px-md">
                                                Page {{ currentPage }} of {{ totalPages }}
                                            </span>
                                            <q-btn
                                                color="primary"
                                                icon="chevron_right"
                                                :disable="currentPage === totalPages"
                                                @click="nextPage"
                                                outline
                                                size="sm"
                                            />
                                        </div>
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>
                        
                        <!-- No Results Banners -->
                        <div v-if="showResults && upgradePlayers.length === 0" class="q-mt-md">
                            <q-banner
                                v-if="!loading && !initialLoad"
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
                                v-else-if="!loading && initialLoad && !selectedTeamPlayer"
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
                
                <!-- Player Data Table Below Both Columns -->
                <div v-if="showResults && upgradePlayers.length > 0" class="q-mt-lg">
                    <q-card class="table-card">
                        <q-card-section>
                            <div class="text-h6 q-mb-md">
                                <q-icon name="table_chart" class="q-mr-sm" />
                                Detailed Player Comparison
                            </div>
                            <PlayerDataTable
                                :players="processedUpgradePlayers"
                                :loading="loading"
                                @player-selected="handlePlayerSelectedForDetailView"
                                @team-selected="handleTeamSelected"
                                :currency-symbol="currencySymbol"
                                :dataset-id="datasetId"
                            />
                        </q-card-section>
                    </q-card>
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

    <!-- Stat Filters Modal -->
    <q-dialog v-model="showStatFiltersModal" persistent>
        <q-card class="stat-filters-modal" style="min-width: 600px; max-width: 800px;">
            <q-card-section class="row items-center q-pb-none">
                <div class="text-h6">Stat Filters</div>
                <q-space />
                <q-btn icon="close" flat round dense v-close-popup />
            </q-card-section>

            <q-card-section class="q-pt-none">
                <div class="text-caption q-mb-md">
                    Set minimum values for player attributes. Only players meeting all criteria will be shown.
                </div>

                <div class="row q-col-gutter-md">
                    <!-- Outfield Player Attributes -->
                    <div class="col-12 col-md-6">
                        <div class="text-subtitle2 q-mb-sm">Outfield Player Attributes</div>
                        
                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">PAC (Pace): {{ minPACFilter }}</div>
                            <q-slider
                                v-model="minPACFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="orange"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">DRI (Dribbling): {{ minDRIFilter }}</div>
                            <q-slider
                                v-model="minDRIFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="blue"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">SHO (Shooting): {{ minSHOFilter }}</div>
                            <q-slider
                                v-model="minSHOFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="red"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">PAS (Passing): {{ minPASFilter }}</div>
                            <q-slider
                                v-model="minPASFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="green"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">DEF (Defending): {{ minDEFFilter }}</div>
                            <q-slider
                                v-model="minDEFFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="purple"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">PHY (Physical): {{ minPHYFilter }}</div>
                            <q-slider
                                v-model="minPHYFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="brown"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>
                    </div>

                    <!-- Goalkeeper Attributes -->
                    <div class="col-12 col-md-6">
                        <div class="text-subtitle2 q-mb-sm">Goalkeeper Attributes</div>
                        
                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">GK (Goalkeeping): {{ minGKFilter }}</div>
                            <q-slider
                                v-model="minGKFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="teal"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">DIV (Diving): {{ minDIVFilter }}</div>
                            <q-slider
                                v-model="minDIVFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="cyan"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">HAN (Handling): {{ minHANFilter }}</div>
                            <q-slider
                                v-model="minHANFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="indigo"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">REF (Reflexes): {{ minREFFilter }}</div>
                            <q-slider
                                v-model="minREFFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="deep-orange"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">KIC (Kicking): {{ minKICFilter }}</div>
                            <q-slider
                                v-model="minKICFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="lime"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">SPD (Speed): {{ minSPDFilter }}</div>
                            <q-slider
                                v-model="minSPDFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="amber"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>

                        <div class="q-mb-md">
                            <div class="text-caption q-mb-xs">POS (Positioning): {{ minPOSFilter }}</div>
                            <q-slider
                                v-model="minPOSFilter"
                                :min="0"
                                :max="99"
                                label
                                label-always
                                color="pink"
                                :dark="$q.dark.isActive"
                                class="q-px-sm"
                            />
                        </div>
                    </div>
                </div>

                <!-- Action Buttons -->
                <div class="row q-col-gutter-md q-mt-lg">
                    <div class="col-6">
                        <q-btn
                            color="grey"
                            label="Reset All"
                            class="full-width"
                            @click="resetStatFilters"
                            outline
                        />
                    </div>
                    <div class="col-6">
                        <q-btn
                            color="primary"
                            label="Apply Filters"
                            class="full-width"
                            @click="showStatFiltersModal = false"
                        />
                    </div>
                </div>
            </q-card-section>
        </q-card>
    </q-dialog>

</template>

<script>
import { useQuasar } from 'quasar'
import { computed, onMounted, ref, watch, nextTick } from 'vue'
import { usePlayerStore } from '@/stores/playerStore' // Corrected Import Path
import { formatCurrency } from '@/utils/currencyUtils'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'
import PlayerCards from './PlayerCards.vue'
import PitchDisplay from './PitchDisplay.vue'
import TeamLogo from './TeamLogo.vue'
import { formations, getFormationLayout } from '@/utils/formations'
import { formationCache } from '@/utils/formationCache'
import { getFlagUrl, getTeamLogoUrl, getPlayerFaceUrl } from '@/utils/imageOptimization'
import { fetchFullPlayerStats, findPlayerUpgrades } from '../services/playerService'

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
  components: { PlayerDataTable, PlayerDetailDialog, PlayerCards, PitchDisplay, TeamLogo },
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
    const baseOverallForSelectedPlayer = ref(null)
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

    // Minimum attribute filters
    const minPACFilter = ref(0)
    const minDRIFilter = ref(0)
    const minSHOFilter = ref(0)
    const minPASFilter = ref(0)
    const minDEFFilter = ref(0)
    const minPHYFilter = ref(0)
    const minGKFilter = ref(0)
    const minDIVFilter = ref(0)
    const minHANFilter = ref(0)
    const minREFFilter = ref(0)
    const minKICFilter = ref(0)
    const minSPDFilter = ref(0)
    const minPOSFilter = ref(0)

    // Modal state
    const showStatFiltersModal = ref(false)

    const loading = ref(false)
    const showResults = ref(false)
    const initialLoad = ref(true)

    const upgradePlayers = ref([])
    const playerForDetailView = ref(null)
    const showPlayerDetailDialog = ref(false)
    const showAllCards = ref(false)
    
    // Pagination variables
    const currentPage = ref(1)
    const playersPerPage = 6

    // Computed property for safe access to currentDatasetId
    const currentDatasetId = computed(() => {
      return playerStore?.currentDatasetId || null
    })

    const populateAllTeamNames = () => {
      const playersToUse = hasCompletePlayerData(playerStore.allPlayers) ? playerStore.allPlayers : props.players
      if (!playersToUse) {
        allTeamNamesCache.value = []
        teamOptions.value = []
        return
      }
      const uniqueTeams = new Set()
      for (const player of playersToUse) {
        if (player.club && player.club.trim() !== '') {
          uniqueTeams.add(player.club)
        }
      }
      allTeamNamesCache.value = Array.from(uniqueTeams).sort()
      teamOptions.value = allTeamNamesCache.value
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

    const parseTransferValueRange = (valueString) => {
      if (!valueString) return { min: 0, max: 0 }
      
      // Handle range format like "£4.8M - £14.5M"
      const rangeMatch = valueString.match(/£([\d.]+)M\s*-\s*£([\d.]+)M/)
      if (rangeMatch) {
        const minValue = parseFloat(rangeMatch[1]) * 1000000
        const maxValue = parseFloat(rangeMatch[2]) * 1000000
        return { min: minValue, max: maxValue }
      }
      
      // Handle single value format like "£28M"
      const singleMatch = valueString.match(/£([\d.]+)M/)
      if (singleMatch) {
        const value = parseFloat(singleMatch[1]) * 1000000
        return { min: value, max: value }
      }
      
      // Handle k format like "£500k"
      const kMatch = valueString.match(/£([\d.]+)k/i)
      if (kMatch) {
        const value = parseFloat(kMatch[1]) * 1000
        return { min: value, max: value }
      }
      
      // Handle K format like "£500K"
      const kMatchUpper = valueString.match(/£([\d.]+)K/)
      if (kMatchUpper) {
        const value = parseFloat(kMatchUpper[1]) * 1000
        return { min: value, max: value }
      }
      
      // Handle other formats with single values
      const singleValueMatch = valueString.match(/£([\d,]+)/)
      if (singleValueMatch) {
        const number = parseInt(singleValueMatch[1].replace(/,/g, ''))
        let value = number
        if (number >= 1000000) {
          value = (number / 1000000) * 1000000
        } else if (number >= 1000) {
          value = (number / 1000) * 1000
        }
        return { min: value, max: value }
      }
      
      return { min: 0, max: 0 }
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

    const formatValue = (valueString) => {
      if (!valueString) return 'N/A'
      
      // Handle range format like "£4.8M - £14.5M"
      const rangeMatch = valueString.match(/£([\d.]+)M\s*-\s*£([\d.]+)M/)
      if (rangeMatch) {
        return `£${rangeMatch[1]}M - £${rangeMatch[2]}M`
      }
      
      // Handle single value format like "£28M"
      const match = valueString.match(/£([\d.]+)M/)
      if (match) {
        return `£${match[1]}M`
      }
      
      // Handle k/K format like "£500k" or "£500K"
      const kMatch = valueString.match(/£([\d.]+)[kK]/)
      if (kMatch) {
        return `£${kMatch[1]}k`
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

    const updateTransferValueSliderBounds = () => {
      const playersToUse = hasCompletePlayerData(playerStore.allPlayers) ? playerStore.allPlayers : props.players
      if (!playersToUse || playersToUse.length === 0) {
        dynamicMinTransferValue.value = 0
        dynamicMaxTransferValue.value = 100000000
        maxTransferValueFilter.value = dynamicMaxTransferValue.value
        return
      }
      let minVal = Number.POSITIVE_INFINITY
      let maxVal = 0
      for (const p of playersToUse) {
        const transferValueRange = parseTransferValueRange(p.transfer_value)
        if (transferValueRange.max > 0) {
          minVal = Math.min(minVal, transferValueRange.min)
          maxVal = Math.max(maxVal, transferValueRange.max)
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
      const playersToUse = hasCompletePlayerData(playerStore.allPlayers) ? playerStore.allPlayers : props.players
      if (!playersToUse || playersToUse.length === 0) {
        dynamicMinSalary.value = 0
        dynamicMaxSalary.value = 1000000
        maxSalaryFilter.value = dynamicMaxSalary.value
        return
      }
      let minVal = Number.POSITIVE_INFINITY
      let maxVal = 0
      for (const p of playersToUse) {
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
      // Ensure we have complete player data with positions and role-specific overalls
      if (props.datasetId && (!props.players || props.players.length === 0 || !hasCompletePlayerData(props.players))) {
        console.log('UpgradeFinderDialog: Fetching complete player data...')
        try {
          await playerStore.fetchPlayersByDatasetId(props.datasetId)
        } catch (error) {
          console.error('Failed to fetch complete player data:', error)
        }
      }
      
      if (playerStore.allAvailableRoles.length === 0 && playerStore.currentDatasetId) {
        await playerStore.fetchAllAvailableRoles()
      }
      populateAllTeamNames()
      updateTransferValueSliderBounds()
      updateSalarySliderBounds()
      maxAgeFilter.value = ageSliderMax
    })
    
    // Helper function to check if player data is complete
    const hasCompletePlayerData = (players) => {
      if (!players || players.length === 0) return false
      
      // Check first few players for complete data
      const sampleSize = Math.min(5, players.length)
      for (let i = 0; i < sampleSize; i++) {
        const player = players[i]
        if (!player.shortPositions || player.shortPositions.length === 0) {
          // console.log(`Player ${player.name} missing shortPositions`)
          continue
        }
        if (!player.roleSpecificOveralls || 
            (Array.isArray(player.roleSpecificOveralls) && player.roleSpecificOveralls.length === 0) ||
            (typeof player.roleSpecificOveralls === 'object' && Object.keys(player.roleSpecificOveralls).length === 0)) {
          // console.log(`Player ${player.name} missing roleSpecificOveralls`)
          continue
        }
      }
      return true
    }

    watch(
      () => [props.players, playerStore.allPlayers],
      async ([newPlayers, storePlayers]) => {
        // Use store players if they have complete data, otherwise use props players
        const playersToUse = hasCompletePlayerData(storePlayers) ? storePlayers : newPlayers
        
        populateAllTeamNames()
        updateTransferValueSliderBounds()
        updateSalarySliderBounds()
        
        // If we have a team selected and new player data, recalculate formation
        if (playersToUse && playersToUse.length > 0 && teamName.value) {
          await nextTick()
          const teamPlayers = playersToUse.filter(player => 
            player.club && player.club.toLowerCase() === teamName.value.toLowerCase()
          )
          if (teamPlayers.length > 0) {
            const bestFormation = calculateBestFormationForTeam(teamPlayers)
            if (bestFormation && !selectedFormationKey.value) {
              selectedFormationKey.value = bestFormation
              calculationMessage.value = `Auto-selected best formation: ${formations[bestFormation].name}. Calculating Best XI...`
              calculationMessageClass.value = $q.dark.isActive
                ? 'bg-info text-white'
                : 'bg-blue-2 text-primary'
            }
          }
        }
        
        if (playersToUse && playersToUse.length > 0) {
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
      const playersToUse = hasCompletePlayerData(playerStore.allPlayers) ? playerStore.allPlayers : props.players
      if (teamName.value && selectedPosition.value && playersToUse) {
        teamPlayersForSelection.value = playersToUse
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
      const playersToUse = hasCompletePlayerData(playerStore.allPlayers) ? playerStore.allPlayers : props.players
      if (teamName.value && playersToUse) {
        const teamPlayers = playersToUse.filter(player => 
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
      const playersToUse = hasCompletePlayerData(playerStore.allPlayers) ? playerStore.allPlayers : props.players
      if (newKey && teamName.value && playersToUse) {
        const teamPlayers = playersToUse.filter(player => 
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
      const playersToUse = hasCompletePlayerData(playerStore.allPlayers) ? playerStore.allPlayers : props.players
      if (newTeamName && playersToUse && playersToUse.length > 0) {
        // Wait a bit to ensure player data is fully loaded
        await nextTick()
        
        // Add a small delay to ensure player data is fully processed
        await new Promise(resolve => setTimeout(resolve, 100))
        
        // Get all players for the selected team (case-insensitive matching)
        const teamPlayers = playersToUse.filter(player => 
          player.club && player.club.toLowerCase() === newTeamName.toLowerCase()
        )
        
        if (teamPlayers.length > 0) {
          
          // Wait for player data to be fully loaded by checking if shortPositions are populated
          let attempts = 0
          const maxAttempts = 10
          
          const waitForPlayerData = () => {
            // Check for players with at least shortPositions (roleSpecificOveralls might not be available for all players)
            const playersWithPositions = teamPlayers.filter(p => 
              p.shortPositions && p.shortPositions.length > 0
            )
            
            // Also check if we have any players with roleSpecificOveralls
            const playersWithRoleOveralls = teamPlayers.filter(p => 
              p.roleSpecificOveralls && (
                Array.isArray(p.roleSpecificOveralls) ? p.roleSpecificOveralls.length > 0 : Object.keys(p.roleSpecificOveralls).length > 0
              )
            )
            
            console.log(`Attempt ${attempts + 1} - Players with positions: ${playersWithPositions.length}, Players with role overalls: ${playersWithRoleOveralls.length}`)
            
            // Proceed if we have players with positions, even if roleSpecificOveralls are missing
            if (playersWithPositions.length > 0 || attempts >= maxAttempts) {
              if (playersWithPositions.length > 0) {
                const bestFormation = calculateBestFormationForTeam(teamPlayers)
                if (bestFormation) {
                  selectedFormationKey.value = bestFormation
                  calculationMessage.value = `Auto-selected best formation: ${formations[bestFormation].name}. Calculating Best XI...`
                  calculationMessageClass.value = $q.dark.isActive
                    ? 'bg-info text-white'
                    : 'bg-blue-2 text-primary'
                } else {
                  calculationMessage.value = 'Could not determine best formation for this team.'
                  calculationMessageClass.value = $q.dark.isActive
                    ? 'bg-warning text-white'
                    : 'bg-orange-2 text-dark'
                }
              } else {
                calculationMessage.value = 'Player data not fully loaded yet. Please try again in a moment.'
                calculationMessageClass.value = $q.dark.isActive
                  ? 'bg-warning text-white'
                  : 'bg-orange-2 text-dark'
              }
            } else {
              attempts++
              setTimeout(waitForPlayerData, 200)
            }
          }
          
          waitForPlayerData()
        } else {
          calculationMessage.value = 'No players found for the selected team.'
          calculationMessageClass.value = $q.dark.isActive
            ? 'bg-warning text-white'
            : 'bg-orange-2 text-dark'
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
      
      // Debug logging for role-specific overalls
      if (role && !hasRoleOveralls) {
        console.log(`Warning: Player ${player.name} has no role-specific overalls for role ${role}, using main overall: ${result}`)
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

      // Debug logging for first few calls (commented out for production)
      // if (Math.random() < 0.1) { // Only log 10% of calls to avoid spam
      //   console.log('getPlayerOverallForRole called for:', player.name, 'role:', slotFormationRole, 'has roleSpecificOveralls:', !!player.roleSpecificOveralls, 'has shortPositions:', !!player.shortPositions)
      // }

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
        console.log('No team players provided for formation calculation')
        return null
      }

              // Check if players have required data
        const playersWithPositions = teamPlayers.filter(p => p.shortPositions && p.shortPositions.length > 0)
        
        if (playersWithPositions.length === 0) {
          console.log('No players with position data found')
          return null
        }
        
        console.log(`Formation calculation: ${playersWithPositions.length} players with positions out of ${teamPlayers.length} total players`)
        
        // Debug: Check role-specific overalls for first few players
        for (let i = 0; i < Math.min(3, teamPlayers.length); i++) {
          const player = teamPlayers[i]
          console.log(`Player ${i + 1}: ${player.name}`, {
            shortPositions: player.shortPositions,
            roleSpecificOveralls: player.roleSpecificOveralls ? 
              (Array.isArray(player.roleSpecificOveralls) ? 
                `${player.roleSpecificOveralls.length} roles` : 
                `${Object.keys(player.roleSpecificOveralls).length} roles`) : 
              'missing',
            Overall: player.Overall
          })
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

    const getBaseOverallFromSelectedPlayer = async () => {
      if (!selectedTeamPlayer.value) return null
      const player = teamPlayersForSelection.value.find(p => p.name === selectedTeamPlayer.value)
      if (!player) return null
      
      // If we have role-specific overalls and a selected role, try to get the role-specific rating
      if (selectedRole.value && player.roleSpecificOveralls) {
        let roleOverall = null
        if (Array.isArray(player.roleSpecificOveralls)) {
          const roleData = player.roleSpecificOveralls.find(r => r.roleName === selectedRole.value)
          roleOverall = roleData ? roleData.score : null
        } else if (typeof player.roleSpecificOveralls === 'object') {
          roleOverall = player.roleSpecificOveralls[selectedRole.value] || null
        }
        
        // If we found a role-specific rating, use it; otherwise fall back to main overall
        if (roleOverall !== null && roleOverall > 0) {
          return roleOverall
        }
      }
      
      // If we don't have role-specific overalls but have a selected role, try to fetch detailed data
      if (selectedRole.value && (!player.roleSpecificOveralls || 
          (Array.isArray(player.roleSpecificOveralls) && player.roleSpecificOveralls.length === 0) ||
          (typeof player.roleSpecificOveralls === 'object' && Object.keys(player.roleSpecificOveralls).length === 0))) {
        
        try {
          if (player.uid && props.datasetId) {
            const result = await fetchFullPlayerStats(props.datasetId, player.uid)
            if (result.data && result.data.player) {
              const detailedPlayer = result.data.player
              
              // Try to get role-specific overall from detailed data
              if (detailedPlayer.roleSpecificOveralls) {
                let roleOverall = null
                if (Array.isArray(detailedPlayer.roleSpecificOveralls)) {
                  const roleData = detailedPlayer.roleSpecificOveralls.find(r => r.roleName === selectedRole.value)
                  roleOverall = roleData ? roleData.score : null
                } else if (typeof detailedPlayer.roleSpecificOveralls === 'object') {
                  roleOverall = detailedPlayer.roleSpecificOveralls[selectedRole.value] || null
                }
                
                if (roleOverall !== null && roleOverall > 0) {
                  return roleOverall
                }
              }
            }
          }
        } catch (error) {
          console.error(`Failed to fetch detailed data for selected player ${player.name}:`, error)
        }
      }
      
      // For upgrade comparison, use the player's main overall rating as baseline
      // This ensures we're comparing against their actual ability, not their position-specific rating
      // which might be null if they can't play that position well
      const mainOverall = player.Overall || player.overall || 0
      return mainOverall
    }

    const selectedTeamPlayerObject = computed(() => {
      if (!selectedTeamPlayer.value) return null
      return teamPlayersForSelection.value.find(p => p.name === selectedTeamPlayer.value) || null
    })

    const updateBaseOverallForSelectedPlayer = async () => {
      if (!selectedTeamPlayer.value) {
        baseOverallForSelectedPlayer.value = null
        return
      }

      try {
        // Get the player object from the team players
        const player = teamPlayersForSelection.value.find(p => p.uid === selectedTeamPlayer.value)
        if (!player) {
          baseOverallForSelectedPlayer.value = null
          return
        }

        // Check if we need to fetch detailed data for this player
        if (!player.roleSpecificOveralls || player.roleSpecificOveralls.length === 0) {
          console.log('UpgradeFinderDialog: Fetching detailed data for selected player:', player.name)
          const detailedData = await fetchFullPlayerStats(props.datasetId, player.uid)
          if (detailedData && detailedData.data) {
            Object.assign(player, detailedData.data)
          }
        }

        // Get the overall for the selected role
        const overall = getPlayerOverallForRoleOrPosition(player, selectedRole.value, selectedPosition.value)
        baseOverallForSelectedPlayer.value = overall
      } catch (error) {
        console.error('Error updating base overall for selected player:', error)
        baseOverallForSelectedPlayer.value = null
      }
    }

    const targetOverallForSearch = async () => {
      const base = await getBaseOverallFromSelectedPlayer()
      if (base === null) return null
      return base + upgradeByValue.value
    }

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

      loading.value = true
      showResults.value = true
      currentPage.value = 1 // Reset to first page when new results are found
      initialLoad.value = false
      
      const baseOverall = await getBaseOverallFromSelectedPlayer()
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
        // Use the new API endpoint for finding upgrades
        const request = {
          datasetId: props.datasetId,
          team: teamName.value,
          position: selectedPosition.value,
          role: selectedRole.value,
          minOverall: targetOverall,
          maxAge: currentMaxAge < ageSliderMax ? currentMaxAge : 0,
          maxTransferValue: currentMaxTransferValue < computedMaxSliderTransferValue.value ? currentMaxTransferValue : 0,
          maxSalary: currentMaxSalary < computedMaxSliderSalary.value ? currentMaxSalary : 0,
          // Minimum attribute filters
          minPAC: minPACFilter.value,
          minDRI: minDRIFilter.value,
          minSHO: minSHOFilter.value,
          minPAS: minPASFilter.value,
          minDEF: minDEFFilter.value,
          minPHY: minPHYFilter.value,
          minGK: minGKFilter.value,
          minDIV: minDIVFilter.value,
          minHAN: minHANFilter.value,
          minREF: minREFFilter.value,
          minKIC: minKICFilter.value,
          minSPD: minSPDFilter.value,
          minPOS: minPOSFilter.value
        }

        console.log('UpgradeFinderDialog: Sending request to API:', request)

        const response = await findPlayerUpgrades(request)
        
        if (response.data && response.data.players) {
          console.log(`Found ${response.data.players.length} upgrades via API`)
          upgradePlayers.value = response.data.players
        } else {
          console.log('No upgrades found or invalid response')
          upgradePlayers.value = []
        }
      } catch (error) {
        console.error('Error finding upgrades:', error)
        upgradePlayers.value = []
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

    const processedUpgradePlayersForCards = computed(() => {
      return upgradePlayers.value.map(player => {
        console.log('UpgradeFinderDialog - Original player roleSpecificOveralls:', player.roleSpecificOveralls)
        console.log('UpgradeFinderDialog - Original player name:', player.name)
        
        const displayOverall = getPlayerOverallForRoleOrPosition(
          player,
          selectedRole.value,
          selectedPosition.value
        )
        
        const processedPlayer = {
          ...player, // Keep all original player data including nationality_iso
          name: player.name,
          nationality: player.nationality || 'Unknown',
          division: player.division || 'Unknown',
          club: player.club || 'Free Agent',
          upperTransferValue: player.transfer_value ? parseTransferValue(player.transfer_value) : 0,
          age: player.age,
          overall: displayOverall || player.overall,
          pac: player.pac || 0,
          sho: player.sho || 0,
          pas: player.pas || 0,
          dri: player.dri || 0,
          def: player.def || 0,
          phy: player.phy || 0,
          position: player.position || player.shortPositions?.[0] || 'Unknown',
          // Explicitly preserve roleSpecificOveralls
          roleSpecificOveralls: player.roleSpecificOveralls || []
        }
        
        console.log('UpgradeFinderDialog - Processed player roleSpecificOveralls:', processedPlayer.roleSpecificOveralls)
        console.log('UpgradeFinderDialog - Processed player name:', processedPlayer.name)
        
        return processedPlayer
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

    const handleTeamSelected = _teamName => {
      // For upgrade finder, we don't need team selection functionality
      // but we need to provide the handler for PlayerDataTable compatibility
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
        'DM (C)': 'DM',
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

    // Get card rarity class based on overall rating
    const getCardRarityClass = (overall) => {
      if (overall >= 90) return 'card-rare-gold'
      if (overall >= 85) return 'card-gold'
      if (overall >= 80) return 'card-silver'
      if (overall >= 75) return 'card-bronze'
      return 'card-common'
    }

    const upgradeFinderIsGoalkeeperView = computed(() => selectedPosition.value === 'GK')
    
    // Pagination computed properties
    const totalPages = computed(() => {
      return Math.ceil(processedUpgradePlayersForCards.value.length / playersPerPage)
    })
    
    const paginatedPlayers = computed(() => {
      const startIndex = (currentPage.value - 1) * playersPerPage
      const endIndex = startIndex + playersPerPage
      return processedUpgradePlayersForCards.value.slice(startIndex, endIndex)
    })

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
    
    // Pagination methods
    const nextPage = () => {
      if (currentPage.value < totalPages.value) {
        currentPage.value++
      }
    }
    
    const previousPage = () => {
      if (currentPage.value > 1) {
        currentPage.value--
      }
    }
    
    // Card click handler
    const handleCardClick = (player) => {
      playerForDetailView.value = player
      showPlayerDetailDialog.value = true
    }

    // Watch for changes in selectedRole to re-trigger search
    watch(selectedRole, async () => {
      await updateBaseOverallForSelectedPlayer()
      if (selectedTeamPlayer.value && showResults.value) {
        console.log('Role changed to:', selectedRole.value, '- re-searching for upgrades')
        findUpgrades()
      }
    })

    // Watch for selected player changes to update base overall
    watch(selectedTeamPlayer, async () => {
      await updateBaseOverallForSelectedPlayer()
    })

    const resetStatFilters = () => {
      minPACFilter.value = 0
      minDRIFilter.value = 0
      minSHOFilter.value = 0
      minPASFilter.value = 0
      minDEFFilter.value = 0
      minPHYFilter.value = 0
      minGKFilter.value = 0
      minDIVFilter.value = 0
      minHANFilter.value = 0
      minREFFilter.value = 0
      minKICFilter.value = 0
      minSPDFilter.value = 0
      minPOSFilter.value = 0
    }

    return {
      quasar: $q,
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
      updateBaseOverallForSelectedPlayer,
      baseOverallForSelectedPlayer,
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
      // Minimum attribute filters
      minPACFilter,
      minDRIFilter,
      minSHOFilter,
      minPASFilter,
      minDEFFilter,
      minPHYFilter,
      minGKFilter,
      minDIVFilter,
      minHANFilter,
      minREFFilter,
      minKICFilter,
      minSPDFilter,
      minPOSFilter,
      loading,
      showResults,
      initialLoad,
      upgradePlayers,
      processedUpgradePlayers,
      processedUpgradePlayersForCards,
      // Pagination
      currentPage,
      totalPages,
      paginatedPlayers,
      nextPage,
      previousPage,
      handleCardClick,
      showAllCards,
      findUpgrades,
      getUnifiedRatingClass,
      getCardRarityClass,
      playerForDetailView,
      showPlayerDetailDialog,
      handlePlayerSelectedForDetailView,
      handleTeamSelected,
      props,
      upgradeFinderIsGoalkeeperView,
      onPositionOrTeamChange,
      getPlayerOverallForRoleOrPosition,
      // Image utility functions
      getFlagUrl,
      getTeamLogoUrl,
      getPlayerFaceUrl,
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
      parseTransferValueRange,
        formatCurrency,
      updateTeamPlayersForFormation,
      resetStatFilters,
      showStatFiltersModal
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

// Player cards section styling
.player-cards-section {
    .text-subtitle1 {
        color: #374151;
        font-weight: 600;
        
        .body--dark & {
            color: #d1d5db;
        }
    }
}

// Embedded Player Cards Grid
.player-cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 1.5rem;
    margin-top: 1rem;
}

.player-card-container {
    display: flex;
    justify-content: center;
}

// FIFA Ultimate Team Card Design (Authentic Version)
.fut-card {
    width: 220px;
    height: 320px;
    position: relative;
    perspective: 1000px;
    
    .card-background {
        width: 100%;
        height: 100%;
        border-radius: 12px;
        position: relative;
        overflow: hidden;
        box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
        transition: all 0.3s ease;
        
        &:hover {
            transform: translateY(-5px) rotateY(5deg);
            box-shadow: 0 15px 35px rgba(0, 0, 0, 0.25);
        }
        
        // Authentic FIFA card backgrounds
        &.card-rare-gold {
            background: linear-gradient(135deg, #ffd700 0%, #ffed4e 50%, #ffd700 100%);
            border: 3px solid #b8860b;
            
            &::before {
                content: '';
                position: absolute;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: linear-gradient(45deg, transparent 30%, rgba(255, 255, 255, 0.4) 50%, transparent 70%);
                animation: shimmer 2s infinite;
            }
        }
        
        &.card-gold {
            background: linear-gradient(135deg, #ffd700 0%, #ffed4e 50%, #ffd700 100%);
            border: 3px solid #b8860b;
        }
        
        &.card-silver {
            background: linear-gradient(135deg, #c0c0c0 0%, #e5e5e5 50%, #c0c0c0 100%);
            border: 3px solid #808080;
        }
        
        &.card-bronze {
            background: linear-gradient(135deg, #cd7f32 0%, #daa520 50%, #cd7f32 100%);
            border: 3px solid #8b4513;
        }
        
        &.card-common {
            background: linear-gradient(135deg, #f5f5f5 0%, #ffffff 50%, #f5f5f5 100%);
            border: 3px solid #d3d3d3;
        }
    }
    
    // Card Header - FIFA style
    .card-header {
        padding: 0.75rem;
        text-align: center;
        background: rgba(0, 0, 0, 0.15);
        border-bottom: 1px solid rgba(0, 0, 0, 0.1);
        
        .player-name {
            font-size: 1rem;
            font-weight: 700;
            color: #1a1a1a;
            margin-bottom: 0.25rem;
            text-shadow: 1px 1px 2px rgba(255, 255, 255, 0.9);
            line-height: 1.2;
            letter-spacing: 0.5px;
        }
        
        .player-nationality {
            font-size: 0.7rem;
            color: #4a4a4a;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
    }
    
    // Player Photo Area - FIFA style
    .player-photo-area {
        padding: 0.75rem;
        text-align: center;
        background: rgba(255, 255, 255, 0.1);
        
        .player-photo {
            margin-bottom: 0.5rem;
        }
        
        .player-position {
            font-size: 0.8rem;
            font-weight: 700;
            color: #1a1a1a;
            text-transform: uppercase;
            letter-spacing: 1px;
            background: rgba(0, 0, 0, 0.1);
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            display: inline-block;
        }
    }
    
    // Club Info - FIFA style
    .club-info {
        padding: 0.5rem 0.75rem;
        text-align: center;
        margin-bottom: 0.75rem;
        background: rgba(255, 255, 255, 0.05);
        
        .club-name {
            font-size: 0.8rem;
            font-weight: 700;
            color: #1a1a1a;
            margin-bottom: 0.25rem;
            line-height: 1.2;
            text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.8);
        }
        
        .division {
            font-size: 0.65rem;
            color: #4a4a4a;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
    }
    
    // Overall Rating - FIFA style
    .overall-rating {
        text-align: center;
        margin-bottom: 0.75rem;
        
        .rating-number {
            display: inline-block;
            width: 45px;
            height: 45px;
            border-radius: 50%;
            font-size: 1.25rem;
            font-weight: 800;
            line-height: 45px;
            text-align: center;
            border: 3px solid #1a1a1a;
            box-shadow: 0 3px 10px rgba(0, 0, 0, 0.3);
            text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
            
            &.rating-tier-6 {
                background: linear-gradient(135deg, #7e57c2 0%, #9575cd 100%);
                color: white;
                border-color: #5e35b1;
            }
            
            &.rating-tier-5 {
                background: linear-gradient(135deg, #26a69a 0%, #4db6ac 100%);
                color: white;
                border-color: #00897b;
            }
            
            &.rating-tier-4 {
                background: linear-gradient(135deg, #66bb6a 0%, #81c784 100%);
                color: white;
                border-color: #4caf50;
            }
            
            &.rating-tier-3 {
                background: linear-gradient(135deg, #42a5f5 0%, #64b5f6 100%);
                color: white;
                border-color: #2196f3;
            }
            
            &.rating-tier-2 {
                background: linear-gradient(135deg, #ffa726 0%, #ffb74d 100%);
                color: #333333;
                border-color: #f57c00;
            }
            
            &.rating-tier-1 {
                background: linear-gradient(135deg, #ef5350 0%, #e57373 100%);
                color: white;
                border-color: #d32f2f;
            }
            
            &.rating-na {
                background: linear-gradient(135deg, #bdbdbd 0%, #e0e0e0 100%);
                color: #424242;
                border-color: #757575;
            }
        }
    }
    
    // Player Stats - FIFA style with proper spacing
    .player-stats {
        padding: 0.5rem 0.75rem;
        margin-bottom: 0.75rem;
        background: rgba(255, 255, 255, 0.05);
        border-radius: 6px;
        
        .stat-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 0.5rem;
            
            &:last-child {
                margin-bottom: 0;
            }
            
            .stat-item {
                display: flex;
                flex-direction: column;
                align-items: center;
                flex: 1;
                margin: 0 0.25rem;
                
                &:first-child {
                    margin-left: 0;
                }
                
                &:last-child {
                    margin-right: 0;
                }
                
                .stat-label {
                    font-size: 0.6rem;
                    font-weight: 700;
                    color: #1a1a1a;
                    margin-bottom: 0.25rem;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                    text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.8);
                }
                
                .stat-value {
                    font-size: 0.85rem;
                    font-weight: 800;
                    padding: 0.25rem 0.5rem;
                    border-radius: 4px;
                    min-width: 2rem;
                    text-align: center;
                    border: 2px solid rgba(0, 0, 0, 0.2);
                    text-shadow: 1px 1px 1px rgba(0, 0, 0, 0.3);
                    
                    &.rating-tier-6 {
                        background: linear-gradient(135deg, #7e57c2 0%, #9575cd 100%);
                        color: white;
                        border-color: #5e35b1;
                    }
                    
                    &.rating-tier-5 {
                        background: linear-gradient(135deg, #26a69a 0%, #4db6ac 100%);
                        color: white;
                        border-color: #00897b;
                    }
                    
                    &.rating-tier-4 {
                        background: linear-gradient(135deg, #66bb6a 0%, #81c784 100%);
                        color: white;
                        border-color: #4caf50;
                    }
                    
                    &.rating-tier-3 {
                        background: linear-gradient(135deg, #42a5f5 0%, #64b5f6 100%);
                        color: white;
                        border-color: #2196f3;
                    }
                    
                    &.rating-tier-2 {
                        background: linear-gradient(135deg, #ffa726 0%, #ffb74d 100%);
                        color: #333333;
                        border-color: #f57c00;
                    }
                    
                    &.rating-tier-1 {
                        background: linear-gradient(135deg, #ef5350 0%, #e57373 100%);
                        color: white;
                        border-color: #d32f2f;
                    }
                    
                    &.rating-na {
                        background: linear-gradient(135deg, #bdbdbd 0%, #e0e0e0 100%);
                        color: #424242;
                        border-color: #757575;
                    }
                }
            }
        }
    }
    
    // Additional Info - FIFA style
    .additional-info {
        padding: 0.5rem 0.75rem;
        background: rgba(0, 0, 0, 0.1);
        border-radius: 6px;
        margin: 0 0.75rem;
        
        .info-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.25rem;
            
            &:last-child {
                margin-bottom: 0;
            }
            
            .info-label {
                font-size: 0.65rem;
                color: #4a4a4a;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.25px;
            }
            
            .info-value {
                font-size: 0.65rem;
                color: #1a1a1a;
                font-weight: 700;
                text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.8);
            }
        }
    }
}

// Shimmer animation for rare gold cards
@keyframes shimmer {
    0% {
        transform: translateX(-100%);
    }
    100% {
        transform: translateX(100%);
    }
}

// Dark mode adjustments for embedded cards
.body--dark {
    .fut-card {
        .card-background {
            &.card-rare-gold {
                background: linear-gradient(135deg, #ffd700 0%, #ffed4e 50%, #ffd700 100%);
            }
            
            &.card-gold {
                background: linear-gradient(135deg, #ffd700 0%, #ffed4e 50%, #ffd700 100%);
            }
            
            &.card-silver {
                background: linear-gradient(135deg, #c0c0c0 0%, #e5e5e5 50%, #c0c0c0 100%);
            }
            
            &.card-bronze {
                background: linear-gradient(135deg, #cd7f32 0%, #daa520 50%, #cd7f32 100%);
            }
            
            &.card-common {
                background: linear-gradient(135deg, #f5f5f5 0%, #ffffff 50%, #f5f5f5 100%);
            }
        }
        
        .card-header {
            .player-name {
                color: #1a1a1a;
            }
            
            .player-nationality {
                color: #4a4a4a;
            }
        }
        
        .player-photo-area {
            .player-position {
                color: #1a1a1a;
            }
        }
        
        .club-info {
            .club-name {
                color: #1a1a1a;
            }
            
            .division {
                color: #4a4a4a;
            }
        }
        
        .player-stats {
            .stat-item {
                .stat-label {
                    color: #1a1a1a;
                }
            }
        }
        
        .additional-info {
            .info-row {
                .info-label {
                    color: #4a4a4a;
                }
                
                .info-value {
                    color: #1a1a1a;
                }
            }
        }
    }
}

// Player Cards Grid Styling
.player-cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
    padding: 1rem;
}

.player-card-container {
    display: flex;
    justify-content: center;
    align-items: center;
}

// Responsive design for embedded cards
@media (max-width: 768px) {
    .player-cards-grid {
        grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
        gap: 1rem;
    }
    
    .fut-card {
        width: 180px;
        height: 260px;
        
        .card-header {
            .player-name {
                font-size: 0.9rem;
            }
        }
        
        .overall-rating {
            .rating-number {
                width: 40px;
                height: 40px;
                font-size: 1.1rem;
                line-height: 40px;
            }
        }
        
        .player-stats {
            .stat-row {
                .stat-item {
                    .stat-value {
                        font-size: 0.8rem;
                    }
                }
            }
        }
    }
}
</style>


