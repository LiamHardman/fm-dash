<template>
    <q-dialog 
        :model-value="show" 
        @hide="$emit('close')"
        :class="isDarkMode ? 'dark-dialog' : 'light-dialog'"
        backdrop-filter="blur(3px)"
        :backdrop-color="isDarkMode ? 'rgba(0, 0, 0, 0.8)' : 'rgba(0, 0, 0, 0.5)'"
        transition-show="scale"
        transition-hide="scale"
    >
        <q-card
            class="player-detail-dialog-card modern-dialog-card"
            :class="
                isDarkMode
                    ? 'bg-dark text-white'
                    : 'bg-white text-dark'
            "
            :style="{
                'max-width': '1400px',
                'width': '95vw',
                'max-height': '90vh',
                'position': 'relative',
                'pointer-events': showCardGenerator ? 'none' : 'auto'
            }"
            :data-dark-mode="isDarkMode"
        >
            <!-- Dialog chrome: header (icon/title/actions/close) + tab strip(s), in
                 normal flow (not overlaid). This is the shared chrome convention also
                 used by UpgradeFinderDialog: a header row, then any tab navigation
                 directly beneath it. -->
            <div class="dialog-chrome">
                <div class="dialog-chrome__header">
                    <q-icon name="person" class="dialog-chrome__icon" />
                    <div class="dialog-chrome__title">
                        {{ displayPlayer?.name || 'Player Details' }}
                    </div>
                    <q-space />
                    <div class="dialog-chrome__actions">
                        <q-btn
                            v-if="player"
                            dense
                            flat
                            round
                            :icon="isCurrentPlayerInComparison ? 'compare_arrows' : 'add_to_queue'"
                            :color="isCurrentPlayerInComparison ? 'warning' : 'primary'"
                            @click="toggleComparison"
                            class="dialog-chrome__action-btn"
                        >
                            <q-tooltip :class="isDarkMode ? 'bg-grey-7' : 'bg-white text-primary'">
                                {{ isCurrentPlayerInComparison ? 'Remove from Comparison' : 'Add to Comparison' }}
                            </q-tooltip>
                        </q-btn>
                        <q-btn
                            dense
                            flat
                            round
                            icon="close"
                            @click="$emit('close')"
                            class="dialog-chrome__close"
                        >
                            <q-tooltip
                                :class="
                                    isDarkMode
                                        ? 'bg-grey-7'
                                        : 'bg-white text-primary'
                                "
                                >Close</q-tooltip
                            >
                        </q-btn>
                    </div>
                </div>

                <!-- Snapshot Navigation (only present when the caller passes 2+ snapshots, e.g. Progression) -->
                <div v-if="snapshotTabs.length > 1" class="dialog-chrome__tabs dialog-chrome__tabs--snapshot">
                    <q-tabs
                        :model-value="activeSnapshotIndex"
                        @update:model-value="(v) => $emit('update:activeSnapshotIndex', v)"
                        dense
                        class="snapshot-tabs"
                        active-color="primary"
                        indicator-color="primary"
                        align="left"
                    >
                        <q-tab
                            v-for="(tab, i) in snapshotTabs"
                            :key="i"
                            :name="i"
                            :label="tab.label"
                        />
                    </q-tabs>
                </div>

                <!-- Tab Navigation -->
                <div class="dialog-chrome__tabs">
                    <q-tabs
                        v-model="activeTab"
                        dense
                        class="view-tabs"
                        active-color="primary"
                        :indicator-color="'transparent'"
                        align="left"
                    >
                        <q-tab name="simple" label="Simple" icon="style" />
                        <q-tab name="advanced" label="Advanced" icon="analytics" />
                        <q-tab name="scoutReport" label="AI Scout Report" icon="travel_explore" />
                    </q-tabs>
                </div>
            </div>

            <!-- Player data not available -->
            <q-card-section v-if="player && (!displayPlayer || !displayPlayer.name)" class="scroll main-content-section">
                <EmptyState
                    icon="person_off"
                    title="Player data not available"
                    description="Unable to load player information."
                />
            </q-card-section>

            <q-card-section v-else-if="player && displayPlayer && displayPlayer.name" class="scroll main-content-section">
                <!-- Progressive loading indicator -->
                <div v-if="isLoadingDetailedData" class="loading-indicator q-mb-md">
                    <div class="text-center">
                        <q-spinner-dots color="primary" size="1em" />
                        <div class="text-caption q-mt-xs">Loading detailed attributes...</div>
                    </div>
                </div>

                <!-- Error message for detailed data -->
                <div v-if="detailedDataError" class="error-message q-mb-md">
                    <q-banner class="bg-red-1 text-red-8">
                        <template v-slot:avatar>
                            <q-icon name="error" color="red" />
                        </template>
                        Failed to load detailed attributes: {{ detailedDataError }}
                        <template v-slot:action>
                            <q-btn flat color="red" label="Retry" @click="fetchDetailedPlayerData" />
                        </template>
                    </q-banner>
                </div>

                <div class="row q-col-gutter-lg" v-if="activeTab !== 'scoutReport'">
                    <div class="col-12 col-md-4">
                        <!-- Simple Tab Content - Player Cards -->
                        <div v-if="activeTab === 'simple'" class="simple-view">
                            <div class="player-card-container">
                                <PlayerCards
                                    :key="`card-${logoKey}`"
                                    :player="cardDisplayPlayer"
                                    :currency-symbol="currencySymbol"
                                    :dataset-id="datasetId"
                                    :card-design-override="cardDisplayPlayer?.totsCardDesign"
                                    :position-override="cardDisplayPlayer?.totsDisplayPosition"
                                    class="player-detail-card"
                                />

                                <!-- Placed in normal document flow here (not up in the dialog
                                     header's action cluster) — a floating overlay button in this
                                     spot proved unreliable to click reliably during development. -->
                                <q-btn
                                    outline
                                    dense
                                    icon="badge"
                                    label="Card Generator"
                                    color="secondary"
                                    class="card-generator-trigger"
                                    @click="showCardGenerator = true"
                                />

                                <!-- Logo correction UI shown below the card -->
                                <div
                                    v-if="showLogoCorrections && shouldShowTeamLogo && player.club && player.club !== '-'"
                                    class="logo-correction-area"
                                >
                                    <!-- Confirmed state -->
                                    <template v-if="logoOverrideState === 'confirmed'">
                                        <div class="logo-correction-row">
                                            <q-icon name="check_circle" color="positive" size="16px" />
                                            <span class="logo-correction-label">Logo confirmed for {{ player.club }}</span>
                                        </div>
                                    </template>

                                    <!-- Alternatives picker (shown after rejection) -->
                                    <template v-else-if="showingAlternatives">
                                        <div class="logo-correction-row">
                                            <q-icon name="cancel" color="negative" size="14px" />
                                            <span class="logo-correction-label">Logo rejected.</span>
                                        </div>
                                        <!-- Automatic alternatives -->
                                        <div v-if="logoAlternatives.length > 0" class="logo-alternatives">
                                            <span class="logo-correction-label">Is it one of these?</span>
                                            <div
                                                v-for="alt in logoAlternatives"
                                                :key="alt.id"
                                                class="logo-alternative-row"
                                            >
                                                <img
                                                    v-if="alt.logoUrl"
                                                    :src="alt.logoUrl"
                                                    :alt="alt.name"
                                                    width="24"
                                                    height="24"
                                                    class="logo-alt-img"
                                                />
                                                <span class="logo-correction-label">{{ alt.name }}</span>
                                                <q-btn
                                                    flat dense round size="xs"
                                                    icon="check" color="positive"
                                                    :disable="isSubmittingLogoOverride"
                                                    @click.stop="confirmLogo(alt.id)"
                                                >
                                                    <q-tooltip>Use {{ alt.name }} for {{ player.club }}</q-tooltip>
                                                </q-btn>
                                            </div>
                                        </div>

                                        <!-- Manual search -->
                                        <div class="logo-search-area">
                                            <span class="logo-correction-label">Search by team name:</span>
                                            <div class="logo-search-row">
                                                <q-input
                                                    v-model="logoSearchQuery"
                                                    dense outlined
                                                    placeholder="e.g. Lille"
                                                    class="logo-search-input"
                                                    @keyup.enter="searchLogoAlternatives"
                                                />
                                                <q-btn
                                                    flat dense round size="sm"
                                                    icon="search"
                                                    :loading="isSearchingLogo"
                                                    @click.stop="searchLogoAlternatives"
                                                />
                                            </div>
                                            <div v-if="logoSearchResults.length > 0" class="logo-alternatives">
                                                <div
                                                    v-for="result in logoSearchResults"
                                                    :key="result.id"
                                                    class="logo-alternative-row"
                                                >
                                                    <img
                                                        :src="result.logoUrl"
                                                        :alt="result.name"
                                                        width="24"
                                                        height="24"
                                                        class="logo-alt-img"
                                                        @error="$event.target.style.display='none'"
                                                    />
                                                    <span class="logo-correction-label">{{ result.name }}</span>
                                                    <q-btn
                                                        flat dense round size="xs"
                                                        icon="check" color="positive"
                                                        :disable="isSubmittingLogoOverride"
                                                        @click.stop="confirmLogo(result.id)"
                                                    >
                                                        <q-tooltip>Use {{ result.name }} for {{ player.club }}</q-tooltip>
                                                    </q-btn>
                                                </div>
                                            </div>
                                        </div>

                                        <div class="logo-correction-row">
                                            <q-btn flat dense size="xs" label="Skip" color="grey" @click.stop="showingAlternatives = false" />
                                        </div>
                                    </template>

                                    <!-- Default: ask user to confirm or reject current match -->
                                    <template v-else-if="logoResolution?.teamId">
                                        <div class="logo-correction-row">
                                            <span class="logo-correction-label">Correct logo for {{ player.club }}?</span>
                                            <q-btn
                                                flat dense round size="sm"
                                                icon="check" color="positive"
                                                :disable="isSubmittingLogoOverride"
                                                @click.stop="confirmLogo()"
                                            >
                                                <q-tooltip>Yes, this logo is correct</q-tooltip>
                                            </q-btn>
                                            <q-btn
                                                flat dense round size="sm"
                                                icon="close" color="negative"
                                                :disable="isSubmittingLogoOverride"
                                                @click.stop="rejectLogo"
                                            >
                                                <q-tooltip>No, this logo is wrong</q-tooltip>
                                            </q-btn>
                                        </div>
                                    </template>
                                </div>
                            </div>
                            
                            <!-- Pros/Cons Component -->
                            <ProsCons 
                                :player="displayPlayer"
                                :selected-comparison-group="selectedComparisonGroup"
                                v-if="displayPlayer && displayPlayer.performancePercentiles"
                            />
                        </div>

                        <!-- Advanced Tab Content - Performance Percentiles -->
                        <div v-if="activeTab === 'advanced'" class="advanced-view">
                        <div class="row q-col-gutter-sm q-mb-md">
                            <div class="col-6">
                                <q-select
                                    v-if="performanceComparisonOptions.length > 0"
                                    :disable="false"
                                    v-model="selectedComparisonGroup"
                                    :options="performanceComparisonOptions"
                                    label="Compare Position"
                                    dense
                                    outlined
                                    emit-value
                                    map-options
                                    class="modern-select"
                                    :label-color="
                                        isDarkMode ? 'grey-4' : ''
                                    "
                                    :popup-content-class="
                                        isDarkMode
                                            ? 'bg-grey-8 text-white'
                                            : 'bg-white text-dark'
                                    "
                                />
                                <q-tooltip
                                    v-if="performanceComparisonOptions.length === 1"
                                >
                                    Only one comparison option available for this player and division.
                                </q-tooltip>
                            </div>
                            <div class="col-6">
                                <q-select
                                    v-model="divisionFilter"
                                    :options="divisionFilterOptions"
                                    label="Compare Division"
                                    dense
                                    outlined
                                    emit-value
                                    map-options
                                    class="modern-select"
                                    :label-color="
                                        isDarkMode ? 'grey-4' : ''
                                    "
                                    :popup-content-class="
                                        isDarkMode
                                            ? 'bg-grey-8 text-white'
                                            : 'bg-white text-dark'
                                    "
                                    @update:model-value="onDivisionFilterChange"
                                />
                            </div>
                        </div>

                        <q-card
                            flat
                            bordered
                            class="performance-percentiles-card modern-stats-card"
                        >
                            <q-card-section
                                class="performance-card-header"
                            >
                                <div class="performance-header-title">
                                    <q-icon name="analytics" class="q-mr-sm" />
                                    Performance Analysis
                                </div>
                            </q-card-section>
                            
                            <q-card-section class="q-pa-md">
                                <!-- Loading State for Percentiles -->
                                <div v-if="showLoadingState" class="percentile-loading-area">
                                    <div class="loading-content">
                                        <q-spinner-dots color="primary" size="2em" />
                                        <div class="loading-text">
                                            <div class="text-subtitle2">Calculating Performance Percentiles...</div>
                                            <div class="text-caption text-grey-6">
                                                {{ isLoadingPercentiles ? 'Fetching data...' : `Retry ${percentilesRetryCount + 1}/${maxRetries}` }}
                                            </div>
                                        </div>
                                        <q-btn 
                                            v-if="percentilesRetryCount > 0" 
                                            flat 
                                            size="sm" 
                                            color="primary" 
                                            label="Retry Now" 
                                            @click="manualRetry"
                                            class="q-mt-sm"
                                        />
                                    </div>
                                </div>

                                <!-- Percentile Content -->
                                <div v-else-if="hasAnyPerformanceData" class="percentile-content-area" :key="`${displayPlayer?.uid || displayPlayer?.UID || 'unknown'}-${displayPlayer?.name || 'unknown'}-${divisionFilter}`">
                                    <!-- Debug info -->
                                    <div v-if="false" class="debug-info">
                                        forceRecompute: {{ forceRecompute }}, 
                                        hasAnyPerformanceData: {{ hasAnyPerformanceData }}, 
                                        categorizedPerformanceStats keys: {{ Object.keys(categorizedPerformanceStats) }}
                                    </div>
                                    <div
                                        v-for="(stats, category, index) in categorizedPerformanceStats"
                                        :key="`perf-${category}-${selectedComparisonGroup}`"
                                        class="performance-category"
                                    >
                                        <div class="performance-category-header q-mb-sm">
                                            <span class="performance-category-title">{{ category }}</span>
                                        </div>
                                        
                                        <q-list separator dense class="performance-stats-list">
                                            <q-item
                                                v-for="statItem in stats"
                                                :key="`${statItem.key}-${selectedComparisonGroup}-${divisionFilter}`"
                                                class="performance-stat-item modern-stat-item"
                                            >
                                                <q-item-section class="stat-name-section">
                                                    <q-item-label
                                                        lines="1"
                                                        class="stat-name-label"
                                                        :title="statItem.name"
                                                    >
                                                        {{ statItem.name }}
                                                    </q-item-label>
                                                </q-item-section>
                                                <q-item-section class="stat-bar-section">
                                                    <div class="stat-bar-container">
                                                        <div class="stat-bar-track">
                                                            <div
                                                                class="stat-bar-fill"
                                                                :style="getBarFillStyle(statItem.percentile)"
                                                            ></div>
                                                        </div>
                                                        <span
                                                            v-if="
                                                                statItem.percentile !== null &&
                                                                statItem.percentile >= 0
                                                            "
                                                            class="stat-percentile-text"
                                                        >
                                                            {{ Math.round(statItem.percentile) }}
                                                        </span>
                                                        <span
                                                            v-else
                                                            class="stat-percentile-text text-caption text-grey-6"
                                                            >N/A</span
                                                        >
                                                    </div>
                                                </q-item-section>
                                                <q-item-section side class="stat-value-section">
                                                    <span class="performance-stat-value">
                                                        {{ statItem.value !== "-" ? statItem.value : "N/A" }}
                                                    </span>
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                        
                                        <q-separator
                                            v-if="index < Object.keys(categorizedPerformanceStats).length - 1"
                                            class="q-my-md performance-separator"
                                        />
                                    </div>
                                </div>
                                
                                <!-- No Data State -->
                                <div v-else class="no-performance-data">
                                    <q-icon name="analytics" size="3em" class="text-grey-4 q-mb-md" />
                                    <div class="text-subtitle1 text-grey-6">Performance data unavailable</div>
                                    <div class="text-caption text-grey-6 q-mb-md">
                                        {{ percentilesRetryCount >= maxRetries 
                                            ? 'Could not load performance percentiles after multiple attempts.' 
                                            : 'Performance percentiles are not available for this player.' }}
                                    </div>
                                    <q-btn 
                                        v-if="percentilesRetryCount >= maxRetries" 
                                        flat 
                                        color="primary" 
                                        label="Try Again" 
                                        @click="manualRetry"
                                    />
                                </div>
                            </q-card-section>
                        </q-card>
                        </div>
                    </div>

                    <div class="col-12 col-md-8">
                        <q-card
                            v-if="activeTab === 'advanced'"
                            flat
                            bordered
                            class="q-mb-sm player-profile-card modern-profile-card"
                        >
                            <q-card-section class="player-profile-content">
                                <div class="profile-header-section">
                                    <div class="player-identity-section">
                                        <div class="row items-center q-mb-sm">
                                            <!-- Player Face Image -->
                                            <div v-if="showFaces" class="col-auto q-mr-md player-face-container">
                                                <img
                                                    v-if="playerFaceImageUrl && !faceImageLoadError"
                                                    :src="playerFaceImageUrl"
                                                    :alt="`${player.name || 'Player'} face`"
                                                    width="80"
                                                    height="80"
                                                    class="player-face-image"
                                                    @error="handleFaceImageError"
                                                    @load="handleFaceImageLoad"
                                                />
                                                <q-avatar
                                                    v-else
                                                    size="80px"
                                                    :color="isDarkMode ? 'grey-7' : 'grey-4'"
                                                    :text-color="isDarkMode ? 'grey-4' : 'grey-7'"
                                                    class="player-face-placeholder"
                                                >
                                                    <q-icon name="person" size="32px" />
                                                </q-avatar>
                                            </div>
                                            
                                            <div class="col-auto q-mr-md player-flag-container">
                                                <img
                                                                v-if="player.nationalityIso && !flagLoadError"
            :src="`https://flagcdn.com/w80/${player.nationalityIso.toLowerCase()}.png`"
                                                    :alt="player.nationality || 'Flag'"
                                                    width="48"
                                                    height="32"
                                                    class="player-flag"
                                                    @error="handleFlagError"
                                                    :title="player.nationality"
                                                />
                                                <q-icon
                                                    v-if="!player.nationalityIso || flagLoadError"
                                                    :color="isDarkMode ? 'grey-5' : 'grey-7'"
                                                    name="flag"
                                                    size="2.5em"
                                                    class="player-flag-placeholder"
                                                />
                                                
                                                <!-- Club logo below nationality flag -->
                                                <div class="q-mt-sm club-logo-container" v-if="player.club && player.club !== '-'">
                                                    <Suspense v-if="shouldShowTeamLogo">
                                                        <template #default>
                                                            <TeamLogo
                                                                :key="`sidebar-logo-${logoKey}`"
                                                                :team-name="player.club"
                                                                :size="32"
                                                                class="player-club-logo"
                                                            />
                                                        </template>
                                                        <template #fallback>
                                                            <div class="club-logo-placeholder">
                                                                <q-skeleton
                                                                    type="circle"
                                                                    size="32px"
                                                                    class="club-logo-skeleton"
                                                                />
                                                            </div>
                                                        </template>
                                                    </Suspense>
                                                    <div v-else class="club-logo-placeholder">
                                                        <q-skeleton
                                                            type="circle"
                                                            size="32px"
                                                            class="club-logo-skeleton"
                                                        />
                                                    </div>

                                                </div>
                                            </div>
                                            
                                            <div class="col player-name-section">
                                                <div class="player-name-container">
                                                    <div class="player-name-and-status">
                                                        <div class="player-name-with-copy">
                                                            <h5
                                                                class="text-h5 player-name no-margin"
                                                                :class="
                                                                    isDarkMode ? 'text-white' : 'text-dark'
                                                                "
                                                                :title="displayPlayer?.name || 'Unknown Player'"
                                                            >
                                                                {{ displayPlayer?.name || 'Unknown Player' }}
                                                            </h5>
                                                            <q-btn
                                                                flat
                                                                dense
                                                                round
                                                                size="sm"
                                                                icon="content_copy"
                                                                @click="copyPlayerName"
                                                                class="copy-name-btn q-ml-xs"
                                                                :class="isDarkMode ? 'text-grey-4' : 'text-grey-7'"
                                                            >
                                                                <q-tooltip
                                                                    :class="
                                                                        isDarkMode
                                                                            ? 'bg-grey-7 text-white'
                                                                            : 'bg-white text-dark'
                                                                    "
                                                                    :delay="300"
                                                                    class="modern-tooltip"
                                                                >
                                                                    Copy player name
                                                                </q-tooltip>
                                                            </q-btn>
                                                            <q-icon
                                                                v-if="player.attributeMasked"
                                                                name="warning"
                                                                color="warning"
                                                                size="sm"
                                                                class="q-ml-sm scouting-warning-icon"
                                                            >
                                                                <q-tooltip
                                                                    :class="
                                                                        isDarkMode
                                                                            ? 'bg-grey-7 text-white'
                                                                            : 'bg-white text-dark'
                                                                    "
                                                                    :delay="300"
                                                                    max-width="300px"
                                                                    class="modern-tooltip"
                                                                >
                                                                    <div class="tooltip-header">⚠️ Scouting Required</div>
                                                                    <div class="tooltip-description">
                                                                        Some of this player's attributes are masked. Scout this player before attempting to sign them to see their full attributes.
                                                                    </div>
                                                                </q-tooltip>
                                                            </q-icon>
                                                        </div>
                                                        <div v-if="bestPlaystyleTagline" class="player-tagline q-mt-xs">
                                                            <q-badge
                                                              outline
                                                              color="grey"
                                                              :label="bestPlaystyleTagline.playstyle"
                                                              class="player-tagline-badge"
                                                            />
                                                            <q-tooltip
                                                              :class="isDarkMode ? 'bg-grey-7 text-white' : 'bg-white text-dark'"
                                                              :delay="300"
                                                              max-width="340px"
                                                              class="modern-tooltip"
                                                            >
                                                              {{ bestPlaystyleTagline.significance }}
                                                            </q-tooltip>
                                                        </div>
                                                        <div class="player-status-badges q-mt-xs">
                                                            <q-badge
                                                                v-if="player.isNew"
                                                                outline
                                                                color="primary"
                                                                label="New"
                                                                class="player-status-badge q-mr-sm"
                                                            />
                                                            <q-badge
                                                                v-if="player.isLoaned"
                                                                outline
                                                                color="secondary"
                                                                label="Loaned"
                                                                class="player-status-badge q-mr-sm"
                                                            />
                                                            <q-badge
                                                                v-if="player.isOnLoan"
                                                                outline
                                                                color="teal"
                                                                label="On Loan"
                                                                class="player-status-badge q-mr-sm"
                                                            />
                                                            <q-badge
                                                                v-if="player.isFree"
                                                                outline
                                                                color="purple"
                                                                label="Free"
                                                                class="player-status-badge q-mr-sm"
                                                            />
                                                        </div>
                                                    </div>
                                                    <div class="player-badges-row q-mt-xs">
                                                        <q-badge
                                                            outline
                                                            color="primary"
                                                            :label="`${displayPlayer?.age || '-'} years`"
                                                            class="player-age-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            outline
                                                            color="secondary"
                                                            :label="displayPlayer?.nationality || 'Unknown'"
                                                            class="player-nationality-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            v-if="displayPlayer?.club"
                                                            outline
                                                            color="teal"
                                                            :label="displayPlayer?.club"
                                                            class="player-club-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            v-if="displayPlayer?.personality"
                                                            outline
                                                            color="purple"
                                                            :label="displayPlayer?.personality"
                                                            class="player-personality-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            v-if="displayPlayer?.media_handling"
                                                            outline
                                                            color="orange"
                                                            :label="displayPlayer?.media_handling"
                                                            class="player-media-badge"
                                                        />
                                                    </div>
                                                </div>
                                                
                                                <div class="player-positions-section q-mt-sm" v-if="playerPositions.length > 0">
                                                    <q-badge
                                                        v-for="pos in playerPositions"
                                                        :key="pos"
                                                        outline
                                                        color="indigo-6"
                                                        :label="pos"
                                                        class="position-badge q-mr-xs q-mb-xs"
                                                    />
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <div class="financial-details-section">
                                        <div class="financial-combined-item" :title="`${formattedTransferValue} / ${formattedWage}`">
                                            <div class="financial-content">
                                                <div class="financial-row">
                                                    <q-icon name="trending_up" class="financial-icon q-mr-sm" />
                                                    <div class="financial-item-content">
                                                        <div class="financial-label">Transfer Value</div>
                                                        <div class="financial-value transfer-value">{{ formattedTransferValue }}</div>
                                                    </div>
                                                </div>
                                                <div class="financial-row q-mt-sm">
                                                    <q-icon name="payments" class="financial-icon q-mr-sm" />
                                                    <div class="financial-item-content">
                                                        <div class="financial-label">Weekly Salary</div>
                                                        <div class="financial-value wage-value">{{ formattedWage }}</div>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <q-separator spaced="md" class="profile-separator" />

                                <div v-if="activeTab === 'advanced'" class="fifa-stats-section">
                                    <div class="fifa-stats-grid">
                                        <div
                                            v-for="stat in fifaStatsToDisplay"
                                            :key="`fifa-${stat.name}-${displayPlayer?.UID || displayPlayer?.uid || 'unknown'}`"
                                            class="fifa-stat-card"
                                        >
                                            <q-card
                                                flat
                                                bordered
                                                :class="[
                                                    'fifa-stat-item text-center',
                                                    getUnifiedRatingClass(getFifaStatValue(stat.name), 100),
                                                ]"
                                            >
                                                <div class="fifa-stat-label">{{ stat.label }}</div>
                                                <div class="fifa-stat-value">
                                                    {{ getFifaStatValue(stat.name) }}
                                                </div>
                                                <q-tooltip
                                                    :class="
                                                        isDarkMode
                                                            ? 'bg-grey-7 text-white'
                                                            : 'bg-white text-dark'
                                                    "
                                                    :delay="500"
                                                    max-width="350px"
                                                    class="modern-tooltip"
                                                >
                                                    <div class="tooltip-header">{{ stat.label }}</div>
                                                    <div class="tooltip-description">
                                                        {{ fifaToFmAttributeMapping[stat.name]?.description || 'No FM attribute mapping available' }}
                                                    </div>
                                                </q-tooltip>
                                            </q-card>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>
                        
                        <div class="row q-col-gutter-lg attribute-columns-container">
                            <div class="col-12 col-md-4">
                                <q-card flat bordered class="attribute-card modern-attribute-card full-height-card">
                                    <q-card-section class="attribute-card-header">
                                        <div class="attribute-section-title">
                                            <q-icon :name="isGoalkeeper ? 'sports_soccer' : 'build'" class="q-mr-sm" />
                                            {{ isGoalkeeper ? "Goalkeeping" : "Technical" }}
                                        </div>
                                    </q-card-section>
                                    
                                    <q-card-section class="q-pa-md">
                                        <q-list separator dense class="attribute-list">
                                            <q-item
                                                v-for="attrKey in isGoalkeeper
                                                    ? attributeCategories.goalkeeping
                                                    : attributeCategories.technical"
                                                :key="attrKey"
                                                class="attribute-list-item modern-attribute-item"
                                                v-show="!isLoadingDetailedData"
                                            >
                                                <q-item-section>
                                                    <q-item-label lines="1" class="attribute-name">
                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                    </q-item-label>
                                                    <q-tooltip
                                                        :class="
                                                            isDarkMode
                                                                ? 'bg-grey-7 text-white'
                                                                : 'bg-white text-dark'
                                                        "
                                                        :delay="500"
                                                        max-width="300px"
                                                        class="modern-tooltip"
                                                    >
                                                        <div class="tooltip-header">
                                                            {{ attributeFullNameMap[attrKey] || attrKey }}
                                                        </div>
                                                        <div class="tooltip-description">
                                                            {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                        </div>
                                                    </q-tooltip>
                                                </q-item-section>
                                                <q-item-section side>
                                                    <span
                                                        :class="[
                                                            'attribute-value modern-attribute-value',
                                                            getUnifiedRatingClass(displayPlayer?.attributes?.[attrKey], 20),
                                                            { 'loading-attribute': isLoadingDetailedData }
                                                        ]"
                                                    >
                                                        {{ getDisplayAttribute(attrKey) }}
                                                    </span>
                                                </q-item-section>
                                            </q-item>
                                            
                                            <!-- Loading skeleton placeholders -->
                                            <template v-if="isLoadingDetailedData">
                                                <q-item 
                                                    v-for="attrKey in (isGoalkeeper ? goalkeepingAttrsOrdered : technicalAttrsOrdered)" 
                                                    :key="`loading-tech-${attrKey}`" 
                                                    class="attribute-list-item modern-attribute-item"
                                                >
                                                    <q-item-section>
                                                        <q-item-label lines="1" class="attribute-name">
                                                            {{ attributeFullNameMap[attrKey] || attrKey }}
                                                        </q-item-label>
                                                        <q-tooltip
                                                            :class="
                                                                isDarkMode
                                                                    ? 'bg-grey-7 text-white'
                                                                    : 'bg-white text-dark'
                                                            "
                                                            :delay="500"
                                                            max-width="300px"
                                                            class="modern-tooltip"
                                                        >
                                                            <div class="tooltip-header">
                                                                {{ attributeFullNameMap[attrKey] || attrKey }}
                                                            </div>
                                                            <div class="tooltip-description">
                                                                {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                            </div>
                                                        </q-tooltip>
                                                    </q-item-section>
                                                    <q-item-section side>
                                                        <q-skeleton type="text" width="30px" class="loading-attribute-skeleton" />
                                                    </q-item-section>
                                                </q-item>
                                            </template>
                                            
                                            <!-- No attributes message -->
                                            <q-item
                                                v-if="!isLoadingDetailedData && !(isGoalkeeper ? attributeCategories.goalkeeping : attributeCategories.technical).length"
                                                class="no-attributes-item"
                                            >
                                                <q-item-section class="text-center q-py-md">
                                                    <q-icon name="info_outline" size="sm" class="q-mr-sm" />
                                                    No {{ isGoalkeeper ? "goalkeeping" : "technical" }} attributes.
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                    </q-card-section>
                                </q-card>
                            </div>
                            
                            <div class="col-12 col-md-4">
                                <q-card flat bordered class="attribute-card modern-attribute-card full-height-card">
                                    <q-card-section class="attribute-card-header">
                                        <div class="attribute-section-title">
                                            <q-icon name="psychology" class="q-mr-sm" />
                                            Mental
                                        </div>
                                    </q-card-section>
                                    
                                    <q-card-section class="q-pa-md">
                                        <q-list separator dense class="attribute-list">
                                            <q-item
                                                v-for="attrKey in attributeCategories.mental"
                                                :key="attrKey"
                                                class="attribute-list-item modern-attribute-item"
                                                v-show="!isLoadingDetailedData"
                                            >
                                                <q-item-section>
                                                    <q-item-label lines="1" class="attribute-name">
                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                    </q-item-label>
                                                    <q-tooltip
                                                        :class="
                                                            isDarkMode
                                                                ? 'bg-grey-7 text-white'
                                                                : 'bg-white text-dark'
                                                        "
                                                        :delay="500"
                                                        max-width="300px"
                                                        class="modern-tooltip"
                                                    >
                                                        <div class="tooltip-header">
                                                            {{ attributeFullNameMap[attrKey] || attrKey }}
                                                        </div>
                                                        <div class="tooltip-description">
                                                            {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                        </div>
                                                    </q-tooltip>
                                                </q-item-section>
                                                <q-item-section side>
                                                    <span
                                                        :class="[
                                                            'attribute-value modern-attribute-value',
                                                            getUnifiedRatingClass(displayPlayer?.attributes?.[attrKey], 20),
                                                            { 'loading-attribute': isLoadingDetailedData }
                                                        ]"
                                                    >
                                                        {{ getDisplayAttribute(attrKey) }}
                                                    </span>
                                                </q-item-section>
                                            </q-item>
                                            
                                            <!-- Loading skeleton placeholders -->
                                            <template v-if="isLoadingDetailedData">
                                                <q-item 
                                                    v-for="attrKey in mentalAttrsOrdered" 
                                                    :key="`loading-mental-${attrKey}`" 
                                                    class="attribute-list-item modern-attribute-item"
                                                >
                                                    <q-item-section>
                                                        <q-item-label lines="1" class="attribute-name">
                                                            {{ attributeFullNameMap[attrKey] || attrKey }}
                                                        </q-item-label>
                                                        <q-tooltip
                                                            :class="
                                                                isDarkMode
                                                                    ? 'bg-grey-7 text-white'
                                                                    : 'bg-white text-dark'
                                                            "
                                                            :delay="500"
                                                            max-width="300px"
                                                            class="modern-tooltip"
                                                        >
                                                            <div class="tooltip-header">
                                                                {{ attributeFullNameMap[attrKey] || attrKey }}
                                                            </div>
                                                            <div class="tooltip-description">
                                                                {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                            </div>
                                                        </q-tooltip>
                                                    </q-item-section>
                                                    <q-item-section side>
                                                        <q-skeleton type="text" width="30px" class="loading-attribute-skeleton" />
                                                    </q-item-section>
                                                </q-item>
                                            </template>
                                            
                                            <!-- No attributes message -->
                                            <q-item v-if="!isLoadingDetailedData && !attributeCategories.mental.length" class="no-attributes-item">
                                                <q-item-section class="text-center q-py-md">
                                                    <q-icon name="info_outline" size="sm" class="q-mr-sm" />
                                                    No mental attributes.
                                                </q-item-section>
                                            </q-item>
                                        </q-list>
                                    </q-card-section>
                                </q-card>
                            </div>
                            
                            <div class="col-12 col-md-4">
                                <div class="row q-col-gutter-md">
                                    <div class="col-12">
                                        <q-card flat bordered class="attribute-card modern-attribute-card">
                                            <q-card-section class="attribute-card-header">
                                                <div class="attribute-section-title">
                                                    <q-icon name="fitness_center" class="q-mr-sm" />
                                                    Physical
                                                </div>
                                            </q-card-section>
                                            
                                            <q-card-section class="q-pa-md">
                                                <q-list separator dense class="attribute-list">
                                                    <q-item
                                                        v-for="attrKey in attributeCategories.physical"
                                                        :key="attrKey"
                                                        class="attribute-list-item modern-attribute-item"
                                                        v-show="!isLoadingDetailedData"
                                                    >
                                                        <q-item-section>
                                                            <q-item-label lines="1" class="attribute-name">
                                                                {{ attributeFullNameMap[attrKey] || attrKey }}
                                                            </q-item-label>
                                                            <q-tooltip
                                                                :class="
                                                                    isDarkMode
                                                                        ? 'bg-grey-7 text-white'
                                                                        : 'bg-white text-dark'
                                                                "
                                                                :delay="500"
                                                                max-width="300px"
                                                                class="modern-tooltip"
                                                            >
                                                                <div class="tooltip-header">
                                                                    {{ attributeFullNameMap[attrKey] || attrKey }}
                                                                </div>
                                                                <div class="tooltip-description">
                                                                    {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                                </div>
                                                            </q-tooltip>
                                                        </q-item-section>
                                                        <q-item-section side>
                                                            <span
                                                                :class="[
                                                                    'attribute-value modern-attribute-value',
                                                                    getUnifiedRatingClass(displayPlayer?.attributes?.[attrKey], 20),
                                                                    { 'loading-attribute': isLoadingDetailedData }
                                                                ]"
                                                            >
                                                                {{ getDisplayAttribute(attrKey) }}
                                                            </span>
                                                        </q-item-section>
                                                    </q-item>
                                                    
                                                    <!-- Loading skeleton placeholders -->
                                                    <template v-if="isLoadingDetailedData">
                                                        <q-item 
                                                            v-for="attrKey in physicalAttrsOrdered" 
                                                            :key="`loading-physical-${attrKey}`" 
                                                            class="attribute-list-item modern-attribute-item"
                                                        >
                                                            <q-item-section>
                                                                <q-item-label lines="1" class="attribute-name">
                                                                    {{ attributeFullNameMap[attrKey] || attrKey }}
                                                                </q-item-label>
                                                                <q-tooltip
                                                                    :class="
                                                                        isDarkMode
                                                                            ? 'bg-grey-7 text-white'
                                                                            : 'bg-white text-dark'
                                                                    "
                                                                    :delay="500"
                                                                    max-width="300px"
                                                                    class="modern-tooltip"
                                                                >
                                                                    <div class="tooltip-header">
                                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                                    </div>
                                                                    <div class="tooltip-description">
                                                                        {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                                    </div>
                                                                </q-tooltip>
                                                            </q-item-section>
                                                            <q-item-section side>
                                                                <q-skeleton type="text" width="30px" class="loading-attribute-skeleton" />
                                                            </q-item-section>
                                                        </q-item>
                                                    </template>
                                                    
                                                    <!-- No attributes message -->
                                                    <q-item v-if="!isLoadingDetailedData && !attributeCategories.physical.length" class="no-attributes-item">
                                                        <q-item-section class="text-center q-py-md">
                                                            <q-icon name="info_outline" size="sm" class="q-mr-sm" />
                                                            No physical attributes.
                                                        </q-item-section>
                                                    </q-item>
                                                </q-list>
                                            </q-card-section>
                                        </q-card>
                                    </div>
                                    
                                    <div class="col-12" v-if="displayPlayer && displayPlayer.roleSpecificOveralls && displayPlayer.roleSpecificOveralls.length > 0">
                                        <q-card flat bordered class="attribute-card modern-attribute-card role-ratings-card">
                                            <q-card-section class="attribute-card-header">
                                                <div class="attribute-section-title">
                                                    <q-icon name="star" class="q-mr-sm" />
                                                    Best Roles
                                                </div>
                                            </q-card-section>
                                            
                                            <q-card-section class="q-pa-md">
                                                <q-list separator dense class="role-specific-ratings-list">
                                                    <q-item
                                                        v-for="roleOverall in sortedRoleSpecificOveralls"
                                                        :key="`role-${roleOverall.roleName}-${roleOverall.score}`"
                                                        :class="{
                                                            'best-role-highlight': roleOverall.score === displayPlayer.Overall,
                                                        }"
                                                        class="attribute-list-item modern-attribute-item role-item"
                                                    >
                                                        <q-item-section>
                                                            <q-item-label lines="1" class="attribute-name role-name" :title="roleOverall.roleName">
                                                                {{ roleOverall.roleName }}
                                                            </q-item-label>
                                                        </q-item-section>
                                                        <q-item-section side>
                                                            <span
                                                                :class="[
                                                                    'attribute-value modern-attribute-value',
                                                                    getUnifiedRatingClass(roleOverall.score, 100)
                                                                ]"
                                                            >
                                                                {{ roleOverall.score }}
                                                            </span>
                                                        </q-item-section>
                                                    </q-item>
                                                </q-list>
                                            </q-card-section>
                                        </q-card>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- AI Scout Report Tab Content — its own sibling block, not nested in the row
                     above. That row's right column (attribute-columns-container) renders
                     unconditionally regardless of activeTab, so nesting inside it would leave
                     leftover FM-attribute panels bleeding in above this tab's content. -->
                <div v-if="activeTab === 'scoutReport'">
                    <ScoutReportTab
                        v-if="player"
                        :player="displayPlayer"
                        :dataset-id="datasetId"
                        :currency-symbol="currencySymbol"
                    />
                </div>
            </q-card-section>

            <q-card-section v-else class="loading-section">
                <div class="loading-content">
                    <q-spinner color="primary" size="3em" />
                    <div class="loading-text">Loading player data...</div>
                </div>
            </q-card-section>
        </q-card>
    </q-dialog>

    <!-- Mounted only once opened (not just toggled via :show) so nothing
         sits in the DOM — and potentially intercepts clicks — while the
         Card Generator hasn't been opened. -->
    <CardGeneratorDialog
        v-if="player && showCardGenerator"
        :show="showCardGenerator"
        :player="player"
        @close="showCardGenerator = false"
    />
</template>

<script>
import { storeToRefs } from 'pinia'
import { useQuasar } from 'quasar'
import {
  computed,
  defineAsyncComponent,
  defineComponent,
  onMounted,
  onUnmounted,
  ref,
  toRef,
  watch,
} from 'vue'
import { usePercentileRetry } from '../composables/usePercentileRetry'
import { fetchFullPlayerStats } from '../services/playerService.js'
import { useComparisonStore } from '../stores/comparisonStore'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
import { formatCurrency } from '../utils/currencyUtils'
import logger from '../utils/logger.js'
import { getCachedPlayerData, setCachedPlayerData } from '../utils/playerDetailOptimizer.js'
import { deriveShortPositionsFromPositionString } from '../utils/playerUtils'
import { ATTRIBUTE_NAME_TO_KEY, PLAYSTYLE_TAGLINES } from '../utils/playstyleTaglines.js'
import EmptyState from './layout/EmptyState.vue'

// Lazy load TeamLogo component to prevent blocking dialog opening
const TeamLogo = defineAsyncComponent(() => import('../components/TeamLogo.vue'))

// Lazy load PlayerCards component for the simple view
const PlayerCards = defineAsyncComponent(() => import('../components/PlayerCards.vue'))

// Lazy load ProsCons component for the simple view
const ProsCons = defineAsyncComponent(() => import('../components/ProsCons.vue'))

// Lazy load ScoutReportTab component for the AI Scout Report view
const ScoutReportTab = defineAsyncComponent(() => import('../components/ScoutReportTab.vue'))

// Lazy load CardGeneratorDialog component for the Card Generator feature
const CardGeneratorDialog = defineAsyncComponent(() => import('./CardGeneratorDialog.vue'))

const attributeFullNameMap = {
  Cor: 'Corners',
  Cro: 'Crossing',
  Dri: 'Dribbling',
  Fin: 'Finishing',
  Fir: 'First Touch',
  Fre: 'Free Kick Taking',
  Hea: 'Heading',
  Lon: 'Long Shots',
  'L Th': 'Long Throws',
  Mar: 'Marking',
  Pas: 'Passing',
  Pen: 'Penalty Taking',
  Tck: 'Tackling',
  Tec: 'Technique',
  Agg: 'Aggression',
  Ant: 'Anticipation',
  Bra: 'Bravery',
  Cmp: 'Composure',
  Cnt: 'Concentration',
  Dec: 'Decisions',
  Det: 'Determination',
  Fla: 'Flair',
  Ldr: 'Leadership',
  OtB: 'Off the Ball',
  Pos: 'Positioning',
  Tea: 'Teamwork',
  Vis: 'Vision',
  Wor: 'Work Rate',
  Acc: 'Acceleration',
  Agi: 'Agility',
  Bal: 'Balance',
  Jum: 'Jumping Reach',
  Nat: 'Natural Fitness',
  Pac: 'Pace',
  Sta: 'Stamina',
  Str: 'Strength',
  Aer: 'Aerial Reach',
  Cmd: 'Command of Area',
  Com: 'Communication',
  Ecc: 'Eccentricity',
  Han: 'Handling',
  Kic: 'Kicking',
  '1v1': 'One on Ones',
  Pun: 'Punching (Tendency)',
  Ref: 'Reflexes',
  TRO: 'Rushing Out (Tendency)',
  Thr: 'Throwing',
}

const attributeDescriptions = {
  // Technical Attributes
  Cor: 'Ability to deliver effective corner kicks with accuracy and technique',
  Cro: 'Quality of crosses from wide positions, affecting accuracy and delivery timing',
  Dri: 'Skill in close ball control while running, beating opponents in 1v1 situations',
  Fin: 'Composure and ability to score when presented with goal-scoring opportunities',
  Fir: 'Quality of the first touch when receiving the ball under pressure',
  Fre: 'Ability to score or create chances from free kick situations',
  Hea: 'Effectiveness when using the head to win aerial duels and score goals',
  Lon: 'Ability to score or create chances from shots taken outside the penalty area',
  'L Th': 'Capability to throw the ball long distances accurately from throw-in situations',
  Mar: 'Defensive positioning and ability to track and stay close to opposing players',
  Pas: 'Accuracy and effectiveness of short and medium-range passing',
  Pen: 'Composure and technique when taking penalty kicks',
  Tck: 'Timing and success rate of defensive challenges and tackles',
  Tec: 'Overall technical ability with the ball, including touch and ball manipulation',

  // Mental Attributes
  Agg: 'Willingness to compete physically and commit to challenges',
  Ant: 'Ability to read the game and anticipate what will happen next',
  Bra: 'Courage when facing physical challenges or difficult situations',
  Cmp: 'Ability to remain calm under pressure and in high-stakes situations',
  Cnt: 'Mental focus and ability to maintain concentration throughout the match',
  Dec: 'Quality of decision-making in various game situations',
  Det: 'Drive and willingness to work hard and overcome obstacles',
  Fla: 'Creativity and unpredictability in play, ability to produce unexpected moments',
  Ldr: 'Ability to motivate teammates and take responsibility in crucial moments',
  OtB: 'Intelligence in finding space and making runs without the ball',
  Pos: 'Understanding of where to be positioned tactically and defensively',
  Tea: 'Willingness to work for the team and follow tactical instructions',
  Vis: 'Ability to spot and execute passes that others might not see',
  Wor: 'Stamina and willingness to maintain effort levels throughout the match',

  // Physical Attributes
  Acc: 'Speed of reaching maximum velocity from a standing start',
  Agi: 'Ability to change direction quickly and maintain balance during movement',
  Bal: 'Ability to maintain equilibrium and stay on feet when challenged',
  Jum: 'Height and timing achieved when jumping for aerial challenges',
  Nat: 'Inherent fitness level and resistance to fatigue and injury',
  Pac: 'Maximum running speed when in full sprint',
  Sta: 'Ability to maintain physical performance throughout the entire match',
  Str: 'Physical power for winning challenges and holding off opponents',

  // Goalkeeping Attributes
  Aer: 'Ability to deal with high balls and crosses in the penalty area',
  Cmd: 'Presence and ability to organize the defense and claim crosses',
  Com: 'Effectiveness in communicating with defenders and organizing the backline',
  Ecc: 'Unpredictability in decision-making and unconventional actions',
  Han: 'Security when catching and holding onto the ball',
  Kic: 'Power and accuracy when distributing the ball with kicks',
  '1v1': 'Ability to save in one-on-one situations with attackers',
  Pun: 'Tendency to punch the ball away rather than catch it',
  Ref: 'Speed of reaction when making saves',
  TRO: 'Tendency to rush out of goal to close down attackers',
  Thr: 'Accuracy and effectiveness when throwing the ball to teammates',
}

const fifaAttributeDescriptions = {
  // Outfield Player FIFA Stats
  PAC: 'Pace combines Acceleration and Sprint Speed to determine how fast a player can move',
  SHO: 'Shooting represents finishing ability, shot power, long shots, volleys, and penalties',
  PAS: 'Passing includes short passing, long passing, vision, crossing, and free kick accuracy',
  DRI: 'Dribbling covers ball control, dribbling skill, agility, balance, and reactions',
  DEF: 'Defending encompasses marking, standing tackle, sliding tackle, and heading accuracy',
  PHY: 'Physical attributes include strength, stamina, aggression, jumping, and balance',

  // Goalkeeper FIFA Stats
  DIV: 'Diving ability to reach shots in different areas of the goal',
  HAN: 'Handling security when catching or parrying shots and crosses',
  KIC: 'Kicking power and accuracy for goal kicks and distribution',
  REF: 'Reflexes and reaction speed when making saves',
  SPD: 'Speed when rushing out or moving around the penalty area',
  POS: 'Positioning and decision-making when coming off the goal line',
}

const fifaToFmAttributeMapping = {
  // Outfield Player FIFA Stats mapped to FM attributes (based on weights)
  PAC: {
    primary: ['Acc', 'Pac'],
    secondary: ['Agi'],
    description: 'Based on Acceleration, Pace, and Agility',
  },
  SHO: {
    primary: ['Fin', 'Lon'],
    secondary: ['Pen', 'Hea', 'Cmp', 'Tec', 'Ant', 'Dec', 'Fla'],
    description:
      'Based on Finishing, Long Shots, Penalties, Heading, Composure, Technique, Anticipation, Decisions, and Flair',
  },
  PAS: {
    primary: ['Pas', 'Vis'],
    secondary: ['Cro', 'Tec', 'Fre', 'Tea', 'Dec', 'Fir', 'Cor', 'OtB'],
    description:
      'Based on Passing, Vision, Crossing, Technique, Free Kicks, Teamwork, Decisions, First Touch, Corners, and Off the Ball',
  },
  DRI: {
    primary: ['Dri', 'Fir', 'Tec'],
    secondary: ['Fla', 'Cmp', 'OtB'],
    description: 'Based on Dribbling, First Touch, Technique, Flair, Composure, and Off the Ball',
  },
  DEF: {
    primary: ['Mar', 'Tck', 'Ant', 'Pos'],
    secondary: ['Hea', 'Cnt', 'Dec', 'Cmp', 'Bra', 'Agg', 'Wor'],
    description:
      'Based on Marking, Tackling, Anticipation, Positioning, Heading, Concentration, Decisions, Composure, Bravery, Aggression, and Work Rate',
  },
  PHY: {
    primary: ['Str', 'Sta', 'Nat'],
    secondary: ['Jum', 'Agg', 'Bra', 'Wor', 'Bal'],
    description:
      'Based on Strength, Stamina, Natural Fitness, Jumping Reach, Aggression, Bravery, Work Rate, and Balance',
  },

  // Goalkeeper FIFA Stats mapped to FM attributes
  DIV: {
    primary: ['Aer', 'Ref', '1v1'],
    secondary: ['Agi', 'Han'],
    description: 'Based on Aerial Reach, Reflexes, One on Ones, Agility, and Handling',
  },
  HAN: {
    primary: ['Han', 'Cmd'],
    secondary: ['Cmp', 'Cnt'],
    description: 'Based on Handling, Command of Area, Composure, and Concentration',
  },
  REF: {
    primary: ['Ref', 'Ant', '1v1'],
    secondary: ['Cnt'],
    description: 'Based on Reflexes, Anticipation, One on Ones, and Concentration',
  },
  KIC: {
    primary: ['Kic', 'Thr'],
    secondary: ['Tec', 'Vis', 'Pas'],
    description: 'Based on Kicking, Throwing, Technique, Vision, and Passing',
  },
  SPD: {
    primary: ['Acc', 'Pac', 'TRO'],
    secondary: [],
    description: 'Based on Acceleration, Pace, and Rushing Out Tendency',
  },
  POS: {
    primary: ['Pos', 'Cmd', 'Ant', 'Dec'],
    secondary: ['TRO', 'Cnt', 'Com'],
    description:
      'Based on Positioning, Command of Area, Anticipation, Decisions, Rushing Out Tendency, Concentration, and Communication',
  },
}

const technicalAttrsOrdered = [
  'Cor',
  'Cro',
  'Dri',
  'Fin',
  'Fir',
  'Fre',
  'Hea',
  'Lon',
  'L Th',
  'Mar',
  'Pas',
  'Pen',
  'Tck',
  'Tec',
]
const mentalAttrsOrdered = [
  'Agg',
  'Ant',
  'Bra',
  'Cmp',
  'Cnt',
  'Dec',
  'Det',
  'Fla',
  'Ldr',
  'OtB',
  'Pos',
  'Tea',
  'Vis',
  'Wor',
]
const physicalAttrsOrdered = ['Acc', 'Agi', 'Bal', 'Jum', 'Nat', 'Pac', 'Sta', 'Str']
const goalkeepingAttrsOrdered = [
  'Aer',
  'Cmd',
  'Com',
  'Ecc',
  'Fir',
  'Han',
  'Kic',
  '1v1',
  'Pas',
  'Pun',
  'Ref',
  'TRO',
  'Thr',
]

const performanceStatMap = {
  'Asts/90': 'Assists per 90',
  'Av Rat': 'Average Rating',
  'Blk/90': 'Blocks per 90',
  'Ch C/90': 'Chances Created per 90',
  'Clr/90': 'Clearances per 90',
  'Cr C/90': 'Crosses Completed per 90',
  'Drb/90': 'Dribbles per 90',
  'xA/90': 'Expected Assists per 90',
  'xG/90': 'Expected Goals per 90',
  'Gls/90': 'Goals per 90',
  'Hdrs W/90': 'Headers Won per 90',
  'Int/90': 'Interceptions per 90',
  'K Ps/90': 'Key Passes per 90',
  'Ps C/90': 'Passes Completed per 90',
  'Shot/90': 'Shots per 90',
  'Tck/90': 'Tackles per 90',
  'Poss Won/90': 'Possession Won per 90',
  'ShT/90': 'Shots on Target per 90',
  'Pres C/90': 'Pressures Completed per 90',
  'Poss Lost/90': 'Possession Lost per 90',
  'Pr passes/90': 'Progressive Passes per 90',
  'Conv %': 'Conversion %',
  'Tck R': 'Tackle Ratio %',
  'Pas %': 'Pass Completion %',
  'Cr C/A': 'Cross Completion %',
  // New performance stats
  Fls: 'Fouls',
  Apps: 'Appearances',
  'NP-xG/90': 'Non-Penalty xG per 90',
  'Ps A/90': 'Pass Attempts per 90',
  Mins: 'Minutes Played',
  'Clean Sheets': 'Clean Sheets',
  FA: 'Fouls Against',
  'CRS A/90': 'Crosses Attempted per 90',
  // Goalkeeper-specific stats
  'Con/90': 'Goals Conceded per 90',
  'Cln/90': 'Clean Sheets per 90',
  'xGP/90': 'Expected Goals Prevented per 90',
  'Sv %': 'Save Percentage',
}

const performanceStatCategories = {
  General: ['Av Rat', 'Apps', 'Mins', 'Clean Sheets'],
  Offensive: ['Gls/90', 'xG/90', 'NP-xG/90', 'Shot/90', 'ShT/90', 'Conv %', 'Drb/90'],
  Passing: [
    'Asts/90',
    'xA/90',
    'Ch C/90',
    'K Ps/90',
    'Ps C/90',
    'Ps A/90',
    'Pas %',
    'Pr passes/90',
    'Cr C/90',
    'CRS A/90',
    'Cr C/A',
    'Poss Lost/90',
  ],
  Defensive: [
    'Tck/90',
    'Tck R',
    'Int/90',
    'Clr/90',
    'Blk/90',
    'Hdrs W/90',
    'Pres C/90',
    'Poss Won/90',
    'Fls',
    'FA',
  ],
  Goalkeeping: ['Con/90', 'Cln/90', 'xGP/90', 'Sv %'],
}

// Mapping from detailed group name (key in performancePercentiles) to the ShortPositions that define them
const detailedGroupToShortPositionsMap = {
  'Full-backs': ['DR', 'DL'],
  'Centre-backs': ['DC'],
  'Wing-backs': ['WBR', 'WBL'],
  'Defensive Midfielders': ['DM'],
  'Central Midfielders': ['MC'],
  'Wide Midfielders': ['MR', 'ML'],
  'Attacking Midfielders (Central)': ['AMC'],
  Wingers: ['AMR', 'AML'],
  Strikers: ['ST'],
}

export default defineComponent({
  name: 'PlayerDetailDialog',
  components: {
    TeamLogo,
    PlayerCards,
    ProsCons,
    ScoutReportTab,
    CardGeneratorDialog,
    EmptyState,
  },
  props: {
    player: { type: Object, default: () => null },
    show: { type: Boolean, default: false },
    currencySymbol: { type: String, default: '$' },
    datasetId: { type: String, default: null },
    // Optional: when the caller has more than one snapshot of the same player (e.g. the
    // Progression page comparing saves over time), pass one entry per snapshot here and the
    // dialog renders a tab strip to switch between them. The caller owns which snapshot is
    // active and is responsible for swapping `player`/`datasetId` in response to
    // `update:activeSnapshotIndex` — the dialog itself has no notion of "snapshots", it just
    // reacts to whichever player/datasetId it's given, same as it always has.
    snapshotTabs: { type: Array, default: () => [] }, // [{ label }]
    activeSnapshotIndex: { type: Number, default: 0 },
  },
  emits: ['close', 'update:activeSnapshotIndex'],
  setup(props) {
    const qInstance = useQuasar()
    const uiStore = useUiStore()
    const playerStore = usePlayerStore()
    const comparisonStore = useComparisonStore()

    const showCardGenerator = ref(false)

    // Create a reactive ref for dark mode state
    const darkModeState = ref(false)

    // Function to update dark mode state
    const updateDarkModeState = () => {
      const quasarDark = qInstance.dark.isActive
      const bodyDark = document.body.classList.contains('body--dark')
      const newState = quasarDark || bodyDark

      darkModeState.value = newState
    }

    // Ensure dark mode detection is reactive
    const isDarkMode = computed(() => {
      return darkModeState.value
    })

    // Watch for dark mode changes to trigger reactivity
    watch(
      () => qInstance.dark.isActive,
      () => {
        updateDarkModeState()
      }
    )

    // Also watch for body class changes
    const observer = new MutationObserver(() => {
      updateDarkModeState()
    })

    onMounted(() => {
      updateDarkModeState() // Initial state
      observer.observe(document.body, { attributes: true, attributeFilter: ['class'] })
    })

    onUnmounted(() => {
      observer.disconnect()
    })

    const selectedComparisonGroup = ref('Global')
    const flagLoadError = ref(false)
    const activeTab = ref('simple') // Default to simple view

    const isTop5League = (division, basedIn) => {
      if (!division) return false
      switch (basedIn) {
        case 'England':
          return division === 'Premier League'
        case 'Spain':
          return (
            division === 'La Liga' ||
            division === 'Primera División' ||
            division === 'Primera Division'
          )
        case 'Germany':
          return division === 'Bundesliga' || division === '1. Bundesliga'
        case 'France':
          return division.startsWith('Ligue 1')
        case 'Italy':
          return division === 'Serie A'
        default:
          return (
            division === 'Premier League' ||
            division === 'La Liga' ||
            division === 'Primera División' ||
            division === 'Primera Division' ||
            division === 'Bundesliga' ||
            division === '1. Bundesliga' ||
            division.startsWith('Ligue 1') ||
            division === 'Serie A'
          )
      }
    }

    const getDefaultDivisionFilter = (player) => {
      if (!player) return 'all'
      return isTop5League(player.division, player.basedIn) ? 'top5' : 'all'
    }

    const divisionFilter = ref(getDefaultDivisionFilter(props.player))

    // Convert props to refs for the percentile retry composable
    const _playerRef = toRef(props, 'player')
    const datasetIdRef = toRef(props, 'datasetId')

    // Add reactive data for detailed player stats
    const detailedPlayerData = ref(null)
    const isLoadingDetailedData = ref(false)
    const detailedDataError = ref(null)
    const percentileUpdateCounter = ref(0) // Force reactivity when percentiles change
    const percentileDataTrigger = ref(0) // Additional trigger for percentile data changes
    const forceRecompute = ref(0) // Force template re-render

    // Computed property to get the player data to display (detailed or basic)
    function deriveAttributes(player) {
      if (player.attributes && Object.keys(player.attributes).length > 0) return player.attributes
      const derived = {}
      const numAttrs = player.numericAttributes || {}
      const perfStats = player.performanceStatsNumeric || {}
      for (const k in numAttrs) derived[k] = String(numAttrs[k])
      for (const k in perfStats) derived[k] = String(perfStats[k])
      return derived
    }

    const displayPlayer = computed(() => {
      const result = detailedPlayerData.value?.name
        ? detailedPlayerData.value
        : props.player?.name
          ? props.player
          : null
      if (!result) return null
      const attrs = deriveAttributes(result)
      if (attrs === result.attributes) return result
      return { ...result, attributes: attrs }
    })

    const cardDisplayPlayer = computed(() => {
      if (!props.player?.totsDisplayPosition) return displayPlayer.value
      if (!displayPlayer.value) return props.player

      return {
        ...displayPlayer.value,
        ...props.player,
        attributes: displayPlayer.value.attributes || props.player.attributes,
        numericAttributes: displayPlayer.value.numericAttributes || props.player.numericAttributes,
        performancePercentiles:
          displayPlayer.value.performancePercentiles || props.player.performancePercentiles,
      }
    })

    // Use the percentile retry composable with displayPlayer instead of props.player
    const {
      isLoadingPercentiles,
      hasValidPercentiles,
      percentilesNeedRetry,
      showLoadingState,
      percentilesRetryCount,
      maxRetries,
      manualRetry,
    } = usePercentileRetry(
      displayPlayer,
      datasetIdRef,
      selectedComparisonGroup,
      divisionFilter,
      (updatedPercentiles) => {
        if (detailedPlayerData.value) {
          // Replace reference (not Object.assign) so watchers fire and counters increment
          detailedPlayerData.value.performancePercentiles = {
            ...(detailedPlayerData.value.performancePercentiles || {}),
            ...updatedPercentiles,
          }
          percentileUpdateCounter.value++
          percentileDataTrigger.value++
          forceRecompute.value++
          performanceStatsCache.clear()
        }
      }
    )

    // Face image handling
    const faceImageLoadError = ref(false)
    const shouldShowTeamLogo = ref(false)

    // Logo correction state
    const logoResolution = ref(null)
    const logoOverrideState = ref(null) // null | 'confirmed' | 'rejected'
    const isSubmittingLogoOverride = ref(false)
    const showLogoCorrections = computed(() => uiStore.showLogoCorrections)

    let _logoComposable = null
    const getLogoComposable = async () => {
      if (!_logoComposable) {
        const { useTeamLogosBackend } = await import('../composables/useTeamLogosBackend')
        _logoComposable = useTeamLogosBackend({ cacheTimeout: 3600000 })
      }
      return _logoComposable
    }

    // After rejection, show alternatives from the resolution so the user can pick the right one
    const showingAlternatives = ref(false)

    // Manual search for when automatic alternatives are all wrong
    const logoSearchQuery = ref('')
    const logoSearchResults = ref([])

    watch(
      () => [props.show, props.player?.club],
      async ([isShowing, club]) => {
        if (!isShowing || !club || club === '-') {
          logoResolution.value = null
          logoOverrideState.value = null
          showingAlternatives.value = false
          logoSearchQuery.value = ''
          logoSearchResults.value = []
          return
        }
        try {
          const composable = await getLogoComposable()
          const resolution = await composable.getClubLogoResolution(club)
          logoResolution.value = resolution
          if (resolution?.reason === 'user override') {
            logoOverrideState.value = 'confirmed'
          } else if (resolution?.reason === 'user-rejected') {
            logoOverrideState.value = 'rejected'
          } else {
            logoOverrideState.value = null
          }
        } catch (_e) {
          logoResolution.value = null
        }
      },
      { immediate: true }
    )

    // Incremented after a confirmation to force TeamLogo / PlayerCards to remount and re-fetch
    const logoKey = ref(0)

    const confirmLogo = async (teamId) => {
      const club = props.player?.club
      const resolvedTeamId = teamId ?? logoResolution.value?.teamId
      if (!club || !resolvedTeamId || isSubmittingLogoOverride.value) return
      isSubmittingLogoOverride.value = true
      try {
        const composable = await getLogoComposable()
        await composable.submitLogoOverride(club, resolvedTeamId)
        logoOverrideState.value = 'confirmed'
        showingAlternatives.value = false
        logoKey.value++ // trigger remount so the card picks up the new logo immediately
      } finally {
        isSubmittingLogoOverride.value = false
      }
    }

    const rejectLogo = async () => {
      const club = props.player?.club
      if (!club || isSubmittingLogoOverride.value) return
      isSubmittingLogoOverride.value = true
      try {
        const composable = await getLogoComposable()
        await composable.submitLogoOverride(club, '')
        logoOverrideState.value = 'rejected'
        // Switch to alternatives view so the user can pick the correct logo
        showingAlternatives.value = true
      } finally {
        isSubmittingLogoOverride.value = false
      }
    }

    // Alternatives are the resolution candidates excluding the one just rejected
    const logoAlternatives = computed(() => {
      const alts = logoResolution.value?.alternatives ?? []
      const rejectedId = logoResolution.value?.teamId
      return alts.filter((a) => a.id !== rejectedId && a.logoAvailable)
    })

    const isSearchingLogo = ref(false)

    const searchLogoAlternatives = async () => {
      if (!logoSearchQuery.value.trim()) return
      isSearchingLogo.value = true
      logoSearchResults.value = []
      try {
        const composable = await getLogoComposable()
        const matches = await composable.getTeamMatches(logoSearchQuery.value)
        logoSearchResults.value = matches.slice(0, 6).map((m) => ({
          id: m.id,
          name: m.name,
          logoUrl: `/api/logos?size=32&teamId=${encodeURIComponent(m.id)}`,
        }))
      } catch (_e) {
        logoSearchResults.value = []
      } finally {
        isSearchingLogo.value = false
      }
    }

    const handleFlagError = () => {
      flagLoadError.value = true
    }

    const handleFaceImageError = (event) => {
      // Set error state and hide the image if it fails to load (404 or other error)
      faceImageLoadError.value = true
      if (event?.target) {
        event.target.style.display = 'none'
      }
    }

    const handleFaceImageLoad = () => {
      faceImageLoadError.value = false
    }

    // Computed property for player face image URL
    const playerFaceImageUrl = computed(() => {
      if (!props.player) return ''

      const playerUID = props.player.UID || props.player.uid
      if (!playerUID) {
        return ''
      }

      // Construct the face API URL
      const nat = props.player.nationality_fifa_code || ''
      return `/api/faces?uid=${encodeURIComponent(playerUID)}&nationality=${encodeURIComponent(nat)}`
    })

    // Reset face image error when player changes
    watch(
      () => props.player,
      (_newPlayer) => {
        faceImageLoadError.value = false
      },
      { immediate: true }
    )

    // Cleanup all caches
    const clearAllCaches = () => {
      performanceStatsCache.clear()
      performanceComparisonOptionsCache.clear()
      currencyCache.clear()
    }

    // Optimize cache size management using more efficient data structures
    const manageCacheSize = () => {
      if (performanceStatsCache.size > maxCacheSize) {
        // Use more efficient LRU-style cleanup
        const entriesToRemove = Math.floor(maxCacheSize * 0.3) // Remove 30% instead of 20%
        const entries = Array.from(performanceStatsCache.entries())

        // Sort by access time if available, otherwise remove oldest
        entries.sort((a, b) => (a[1]._lastAccess || 0) - (b[1]._lastAccess || 0))

        for (let i = 0; i < entriesToRemove; i++) {
          performanceStatsCache.delete(entries[i][0])
        }

        console.log(
          `🧹 Cleaned ${entriesToRemove} cache entries, size: ${performanceStatsCache.size}`
        )
      }
    }

    onMounted(() => {
      /* Initialization logic if needed */
    })

    onUnmounted(() => {
      clearAllCaches()
    })

    // Watch dialog visibility to delay team logo loading
    watch(
      () => props.show,
      (isShowing) => {
        if (isShowing) {
          // Delay team logo rendering until dialog is fully opened and data is loaded
          setTimeout(() => {
            shouldShowTeamLogo.value = true
          }, 100) // Increased delay to prioritize data loading
        } else {
          shouldShowTeamLogo.value = false
        }
      },
      { immediate: true }
    )

    const divisionFilterOptions = computed(() => [
      { label: 'All', value: 'all' },
      { label: 'Same', value: 'same' },
      { label: 'Top 5', value: 'top5' },
    ])

    const getTargetDivision = () => {
      if (!props.player?.division) return null
      return props.player.division
    }

    const deriveShortPositions = (player) => {
      const positions = new Set()
      if (player.position) {
        for (const p of deriveShortPositionsFromPositionString(player.position)) {
          positions.add(p)
        }
      }
      if (player.roleSpecificOveralls) {
        for (const role of player.roleSpecificOveralls) {
          const shortPos = role.roleName.split(' - ')[0]
          if (shortPos) positions.add(shortPos)
        }
      }
      return Array.from(positions)
    }

    const derivePositionGroups = (player) => {
      const shortPositions = deriveShortPositions(player)
      const groups = []

      // Map short positions to broad groups
      const positionToGroupMap = {
        GK: 'Goalkeepers',
        SW: 'Defenders',
        DR: 'Defenders',
        DL: 'Defenders',
        DC: 'Defenders',
        WBR: 'Defenders',
        WBL: 'Defenders',
        DM: 'Midfielders',
        MC: 'Midfielders',
        MR: 'Midfielders',
        ML: 'Midfielders',
        AMC: 'Midfielders',
        AMR: 'Midfielders',
        AML: 'Midfielders',
        ST: 'Attackers',
      }

      for (const pos of shortPositions) {
        const group = positionToGroupMap[pos]
        if (group && !groups.includes(group)) {
          groups.push(group)
        }
      }

      return groups
    }

    const isGoalkeeper = computed(() => {
      if (!displayPlayer.value || !displayPlayer.value.name) return false

      // Use derived position information
      const derivedShortPositions = deriveShortPositions(displayPlayer.value)
      const derivedPositionGroups = derivePositionGroups(displayPlayer.value)

      const isGK =
        derivedShortPositions.includes('GK') ||
        derivedPositionGroups.includes('Goalkeepers') ||
        displayPlayer.value.parsed_positions?.includes('Goalkeeper') ||
        displayPlayer.value.position?.includes('GK')
      return isGK
    })

    // Pre-defined stat configurations for better performance
    const goalkeepingStats = [
      { name: 'DIV', label: 'DIV' },
      { name: 'HAN', label: 'HAN' },
      { name: 'REF', label: 'REF' },
      { name: 'KIC', label: 'KIC' },
      { name: 'SPD', label: 'SPD' },
      { name: 'POS', label: 'POS' },
    ]

    const outfieldStats = [
      { name: 'PAC', label: 'PAC' },
      { name: 'SHO', label: 'SHO' },
      { name: 'PAS', label: 'PAS' },
      { name: 'DRI', label: 'DRI' },
      { name: 'DEF', label: 'DEF' },
      { name: 'PHY', label: 'PHY' },
    ]

    // FIFA stats that show immediately with loading states
    const fifaStatsToDisplay = computed(() => {
      const statsTemplate = isGoalkeeper.value ? goalkeepingStats : outfieldStats

      // If we have detailed player data, use it
      if (displayPlayer.value?.name) {
        const filteredStats = statsTemplate.filter((stat) => {
          const hasStat =
            displayPlayer.value[stat.name] !== undefined && displayPlayer.value[stat.name] !== null
          return hasStat
        })
        return filteredStats
      }

      // Check if we have basic player data with FIFA stats
      if (props.player?.name) {
        const filteredStats = statsTemplate.filter((stat) => {
          const lowercaseStatName = stat.name.toLowerCase()
          const hasStat =
            props.player[lowercaseStatName] !== undefined &&
            props.player[lowercaseStatName] !== null
          return hasStat
        })
        return filteredStats
      }

      // Otherwise, show all stats with loading state
      return statsTemplate
    })

    // Basic player data for template access
    const basicPlayer = computed(() => props.player || {})

    // Get FIFA stat value with loading state
    const getFifaStatValue = (statName) => {
      // First try basic player data (lowercase) - this is the immediate data
      const lowercaseStatName = statName.toLowerCase()
      if (
        props.player &&
        props.player[lowercaseStatName] !== undefined &&
        props.player[lowercaseStatName] !== null
      ) {
        return props.player[lowercaseStatName]
      }

      // Then try detailed player data (uppercase) - this is the enhanced data
      if (
        displayPlayer.value &&
        displayPlayer.value[statName] !== undefined &&
        displayPlayer.value[statName] !== null
      ) {
        return displayPlayer.value[statName]
      }

      if (isLoadingDetailedData.value) {
        return '...'
      }

      return '-'
    }

    const _averageRatingData = computed(() => {
      if (!displayPlayer.value || !displayPlayer.value.attributes || !displayPlayer.value.name)
        return null
      if (
        !Object.hasOwn(displayPlayer.value.attributes, 'Av Rat') ||
        displayPlayer.value.attributes['Av Rat'] === '-' ||
        displayPlayer.value.attributes['Av Rat'] === ''
      ) {
        return null
      }
      // Get percentile for average rating from detailed player data
      const percentilesForGroup =
        displayPlayer.value.performancePercentiles?.[selectedComparisonGroup.value]
      const avgRatingPercentile = percentilesForGroup?.['Av Rat'] ?? null

      return {
        key: 'Av Rat',
        name: performanceStatMap['Av Rat'] || 'Average Rating',
        value: displayPlayer.value.attributes['Av Rat'],
        percentile: avgRatingPercentile,
      }
    })

    // Optimized cache with LRU-like behavior
    const performanceStatsCache = new Map()
    const maxCacheSize = 50

    // Memoized helper functions
    const getCategoryOrder = computed(() => {
      return isGoalkeeper.value
        ? ['General', 'Goalkeeping', 'Passing']
        : ['General', 'Passing', 'Offensive', 'Defensive']
    })

    const getCacheKey = (player, groupKey) => {
      const playerUID = player.UID || player.uid
      if (playerUID && playerUID !== '') {
        return `${playerUID}-${groupKey}-${player.version || 0}`
      }
      return `${player.name || 'unknown'}-${player.club || 'unknown'}-${player.age || 'unknown'}-${groupKey}`
    }

    const buildStatsForCategory = (categoryName, percentilesForGroup, playerAttributes) => {
      const categoryStats = performanceStatCategories[categoryName]
      if (!categoryStats) return []

      const statsInCategory = []

      for (let i = 0; i < categoryStats.length; i++) {
        const statKey = categoryStats[i]
        const rawAttributeValue = playerAttributes[statKey]

        if (
          performanceStatMap[statKey] &&
          rawAttributeValue !== undefined &&
          rawAttributeValue !== '-' &&
          rawAttributeValue !== ''
        ) {
          // Get percentile from the detailed player data
          const percentile = percentilesForGroup?.[statKey] ?? null

          statsInCategory.push({
            key: statKey,
            name: performanceStatMap[statKey],
            value: rawAttributeValue,
            percentile: percentile,
          })
        }
      }

      if (statsInCategory.length === 0) return []

      // Optimize sorting for General category
      if (categoryName === 'General') {
        const avgRatingIndex = statsInCategory.findIndex((stat) => stat.key === 'Av Rat')
        if (avgRatingIndex > -1) {
          const avgRatingStat = statsInCategory.splice(avgRatingIndex, 1)[0]
          statsInCategory.sort((a, b) => a.name.localeCompare(b.name))
          return [avgRatingStat, ...statsInCategory]
        }
      }

      const result = statsInCategory.sort((a, b) => a.name.localeCompare(b.name))

      // Manage cache size after building stats
      manageCacheSize()

      return result
    }

    const hasAnyPerformanceData = computed(() => {
      const keys = Object.keys(categorizedPerformanceStats.value)

      return keys.length > 0
    })

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

    const getBarFillStyle = (percentile) => {
      if (percentile === null || percentile === undefined || percentile < 0) {
        return {
          width: '0%',
          backgroundColor: '#9e9e9e',
          height: '12px',
          borderRadius: '3px',
        }
      }
      const p = Math.max(0, Math.min(100, percentile))
      let backgroundColor
      if (p <= 10) backgroundColor = '#d32f2f'
      else if (p <= 30) backgroundColor = '#ef6c00'
      else if (p <= 45) backgroundColor = '#fdd835'
      else if (p <= 55) backgroundColor = '#bdbdbd'
      else if (p <= 70) backgroundColor = '#aed581'
      else if (p <= 90) backgroundColor = '#66bb6a'
      else backgroundColor = '#388e3c'
      return {
        width: `${p}%`,
        backgroundColor: backgroundColor,
        height: '12px',
        borderRadius: '3px',
        transition: 'width 0.3s ease, background-color 0.3s ease',
      }
    }

    // Memoized role sorting with shallow comparison optimization
    let lastRoleOveralls = null
    let lastSortedRoles = []

    const sortedRoleSpecificOveralls = computed(() => {
      const roleOveralls = displayPlayer.value?.roleSpecificOveralls
      if (!roleOveralls || roleOveralls.length === 0) {
        return []
      }

      if (roleOveralls.length === 1) {
        return roleOveralls
      }

      // Shallow comparison optimization - only re-sort if array reference changed
      if (roleOveralls === lastRoleOveralls) {
        return lastSortedRoles
      }

      lastRoleOveralls = roleOveralls
      lastSortedRoles = [...roleOveralls].sort((a, b) => b.score - a.score)
      return lastSortedRoles
    })

    // Determine best-base position for tagline from roleSpecificOveralls, fallback to derived short positions
    const bestBasePosition = computed(() => {
      const roleOveralls = displayPlayer.value?.roleSpecificOveralls
      if (roleOveralls && roleOveralls.length > 0) {
        const best = [...roleOveralls].sort((a, b) => b.score - a.score)[0]
        if (best?.roleName) {
          const short = String(best.roleName).split(' - ')[0]
          return short
        }
      }
      // fallback: first derived short position if available
      const derived = deriveShortPositions(displayPlayer.value || {})
      return derived.length > 0 ? derived[0] : null
    })

    // Score a playstyle against player's numeric attributes (1-20 scale where available).
    // For each attribute in fm_attributes, map to our attribute key and average available values.
    const scorePlaystyle = (player, fmAttributes) => {
      if (!player || !Array.isArray(fmAttributes) || fmAttributes.length === 0) return 0
      const attrs = player.numericAttributes || player.attributes || {}
      let sum = 0
      let count = 0
      for (const name of fmAttributes) {
        const key = ATTRIBUTE_NAME_TO_KEY[name]
        if (!key) continue
        const val = attrs[key]
        const num = typeof val === 'number' ? val : Number.parseInt(val, 10)
        if (!Number.isNaN(num)) {
          sum += num
          count += 1
        }
      }
      if (count === 0) return 0
      return Math.round(sum / count)
    }

    // Compute the best matching playstyle for the best base position
    const bestPlaystyleTagline = computed(() => {
      const pos = bestBasePosition.value
      if (!pos) return null
      const list = PLAYSTYLE_TAGLINES[pos]
      if (!list || list.length === 0) return null
      let best = null
      let bestScore = -1
      for (const item of list) {
        const score = scorePlaystyle(displayPlayer.value, item.fm_attributes)
        if (score > bestScore) {
          bestScore = score
          best = item
        }
      }
      return best
    })

    // Memoized currency formatting with caching
    const currencyCache = new Map()

    const createCurrencyFormatter = (amount, symbol, fallback) => {
      const cacheKey = `${amount}-${symbol}-${fallback}`
      if (currencyCache.has(cacheKey)) {
        return currencyCache.get(cacheKey)
      }

      const formatted = formatCurrency(amount, symbol, fallback)

      // Keep cache size reasonable
      if (currencyCache.size > 100) {
        currencyCache.clear()
      }

      currencyCache.set(cacheKey, formatted)
      return formatted
    }

    const formattedTransferValue = computed(() => {
      if (!props.player) return '-'
      return createCurrencyFormatter(
        props.player.transferValueAmount,
        props.currencySymbol,
        props.player.transfer_value
      )
    })

    const formattedWage = computed(() => {
      if (!props.player) return '-'
      return createCurrencyFormatter(
        props.player.wageAmount,
        props.currencySymbol,
        props.player.wage
      )
    })

    const currencyIcon = computed(() => {
      switch (props.currencySymbol) {
        case '€':
          return 'euro_symbol'
        case '£':
          return 'currency_pound'
        case '$':
          return 'attach_money'
        default:
          return 'payments'
      }
    })

    // Get showAttributeMasks from the uiStore
    const { showAttributeMasks } = storeToRefs(uiStore)

    const getDisplayAttribute = (attrKey) => {
      // Show loading state if detailed data is still loading
      if (isLoadingDetailedData.value) {
        return '...'
      }

      if (!displayPlayer.value) return '-'

      const rawValue = displayPlayer.value.attributes?.[attrKey]
      if (rawValue === undefined) return '-'

      if (rawValue === '-') {
        return '?'
      }
      if (showAttributeMasks.value && String(rawValue).includes('-')) {
        return rawValue
      }
      // Use numericAttributes if available, otherwise use the raw value
      const numericValue = displayPlayer.value.numericAttributes?.[attrKey]
      return numericValue !== undefined ? numericValue : rawValue
    }

    // Force recomputation when player changes
    watch(
      () => props.player,
      (_newPlayer) => {
        percentileUpdateCounter.value++
        percentileDataTrigger.value++
        forceRecompute.value++
      },
      { immediate: true }
    )

    // Division filter change handler - moved after reactive variables are declared
    const onDivisionFilterChange = async () => {
      if (!props.datasetId || !displayPlayer.value) return

      try {
        // Update percentiles for the current player with new division filter
        const playerUID = displayPlayer.value.uid || displayPlayer.value.UID
        if (playerUID) {
          // Preserve the current comparison group
          const currentComparisonGroup = selectedComparisonGroup.value

          // Handle the 'same' division filter by converting it to the player's actual division
          let effectiveDivision = divisionFilter.value
          if (divisionFilter.value === 'same') {
            const targetDivision = getTargetDivision()
            if (targetDivision) {
              effectiveDivision = targetDivision
            } else {
              // If no target division is available, fall back to 'same'
              effectiveDivision = 'same'
            }
          }

          const updatedPercentiles = await fetchPlayerPercentiles(
            playerUID,
            effectiveDivision,
            currentComparisonGroup
          )

          if (updatedPercentiles) {
            if (detailedPlayerData.value) {
              detailedPlayerData.value.performancePercentiles = updatedPercentiles
              percentileUpdateCounter.value++
              percentileDataTrigger.value++
              forceRecompute.value++
            }
            performanceStatsCache.clear()
            performanceComparisonOptionsCache.clear()
          }
        }
      } catch (error) {
        logger.error('Failed to update percentiles on division filter change', {
          error: error.message,
        })
      }
    }

    // Categorized performance stats computed property - moved after reactive variables are declared
    const categorizedPerformanceStats = computed(() => {
      // Force reactivity by accessing the detailed player data and the update counter
      const player = displayPlayer.value
      const playerName = player?.name
      const playerUID = player?.uid || player?.UID
      const playerAttributes = player?.attributes
      const performancePercentiles = player?.performancePercentiles
      const updateCounter = percentileUpdateCounter.value // Force dependency on percentile updates
      const dataTrigger = percentileDataTrigger.value // Force dependency on data changes
      const recomputeTrigger = forceRecompute.value // Force dependency on template re-render

      // Force dependency on forceRecompute to ensure re-evaluation
      const _forceRecomputeValue = forceRecompute.value

      // Force reactivity by accessing the actual percentile data structure
      const percentilesHash = performancePercentiles
        ? JSON.stringify(Object.keys(performancePercentiles).sort())
        : ''

      // Force reactivity by accessing the actual percentile values
      const _percentileValues = performancePercentiles
        ? Object.values(performancePercentiles).slice(0, 3)
        : []

      // Force dependency on performance percentiles by accessing specific values
      const percentilesKeys = performancePercentiles ? Object.keys(performancePercentiles) : []
      const _hasPercentiles = percentilesKeys.length > 0

      // Force reactivity by accessing the specific percentile data for the selected group
      const selectedGroupData = performancePercentiles?.[selectedComparisonGroup.value]
      const selectedGroupKeys = selectedGroupData ? Object.keys(selectedGroupData) : []

      if (!playerAttributes || !playerName || !playerUID) {
        return {}
      }

      // Use full attributes from detailed player data

      // Get percentiles from the detailed player data
      const percentilesForGroup = performancePercentiles?.[selectedComparisonGroup.value]

      // Force reactivity by accessing the percentile data structure
      const _availableGroups = performancePercentiles ? Object.keys(performancePercentiles) : []

      // Check if we have percentile data for the selected group
      if (!percentilesForGroup || Object.keys(percentilesForGroup).length === 0) {
        return {}
      }

      // Force reactivity by accessing the specific percentile values
      const percentileKeys = Object.keys(percentilesForGroup)
      if (percentileKeys.length === 0) {
        return {}
      }

      // Access a few percentile values to ensure Vue tracks the dependency
      const samplePercentiles = percentileKeys.slice(0, 3).map((key) => percentilesForGroup[key])
      if (samplePercentiles.some((p) => p === undefined || p === null)) {
        return {}
      }

      // Force reactivity by accessing the actual percentile values
      // This ensures Vue tracks changes to the specific percentile data
      const _sampleValues = samplePercentiles.map((p) => p?.toString() || '0')

      // Create a more specific cache key that includes player UID to ensure cache invalidation
      const cacheKey = `${playerUID}-${selectedComparisonGroup.value}-${divisionFilter.value}-${updateCounter}-${dataTrigger}-${recomputeTrigger}-${percentilesHash}-${selectedGroupKeys.length}`

      // Return cached result if available
      if (performanceStatsCache.has(cacheKey)) {
        const cached = performanceStatsCache.get(cacheKey)
        // Move to end for LRU behavior
        performanceStatsCache.delete(cacheKey)
        performanceStatsCache.set(cacheKey, cached)
        return cached
      }

      const result = {}
      const categoryOrder = getCategoryOrder.value

      for (let i = 0; i < categoryOrder.length; i++) {
        const categoryName = categoryOrder[i]
        const categoryStats = buildStatsForCategory(
          categoryName,
          percentilesForGroup, // Pass percentiles from detailed player data
          playerAttributes
        )

        if (categoryStats.length > 0) {
          result[categoryName] = categoryStats
        }
      }

      // Implement LRU cache eviction
      if (performanceStatsCache.size >= maxCacheSize) {
        const firstKey = performanceStatsCache.keys().next().value
        performanceStatsCache.delete(firstKey)
      }

      performanceStatsCache.set(cacheKey, result)
      return result
    })

    // Function to fetch percentiles for a specific player using the new API
    const fetchPlayerPercentiles = async (
      playerUID,
      compareDivision = 'same',
      comparePosition = 'Global'
    ) => {
      if (!props.datasetId || !playerUID) return null

      try {
        const requestPayload = {
          playerUID: playerUID.toString(),
          compareDivision: compareDivision,
          comparePosition: comparePosition,
        }

        const url = `/api/player-percentiles/${props.datasetId}`
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestPayload),
        })

        if (response.ok) {
          const percentiles = await response.json()

          // Validate the response
          if (!percentiles || typeof percentiles !== 'object') {
            throw new Error('Invalid percentile response format')
          }

          return percentiles
        }
        // Log the error response
        const errorText = await response.text()
        logger.error('Percentile fetch failed', {
          status: response.status,
          statusText: response.statusText,
          error_text: errorText,
        })
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      } catch (error) {
        logger.error('Failed to fetch percentiles', {
          error: error?.message || 'Unknown error',
          player_uid: playerUID,
          stack: error?.stack || 'No stack trace',
          error_type: error?.constructor?.name || 'Unknown',
        })
        return null
      }
    }

    // Function to fetch detailed player data
    const fetchDetailedPlayerData = async () => {
      console.log('fetchDetailedPlayerData called with:', {
        hasPlayer: !!props.player,
        hasDatasetId: !!props.datasetId,
        playerName: props.player?.name,
        playerUID: props.player?.uid || props.player?.UID,
      })

      if (!props.player || !props.datasetId) {
        console.log('fetchDetailedPlayerData early return - missing player or datasetId')
        return
      }

      const startTime = performance.now()
      isLoadingDetailedData.value = true
      detailedDataError.value = null

      try {
        // Check if the player already has detailed data (from performance API)
        console.log('Checking player data structure:', {
          hasAttributes: !!props.player.attributes,
          hasPerformancePercentiles: !!props.player.performancePercentiles,
          attributesKeys: props.player.attributes ? Object.keys(props.player.attributes) : [],
          performancePercentilesKeys: props.player.performancePercentiles
            ? Object.keys(props.player.performancePercentiles)
            : [],
          playerKeys: Object.keys(props.player),
        })

        if (props.player.attributes && props.player.performancePercentiles) {
          // Use the pre-loaded data from performance API
          detailedPlayerData.value = props.player
          const _loadTime = performance.now() - startTime

          // Eagerly fetch if percentiles are missing for the current comparison group
          const currentGroup = selectedComparisonGroup.value
          const existingGroup = props.player.performancePercentiles[currentGroup]
          const hasValidGroup =
            existingGroup &&
            Object.values(existingGroup).some((v) => v !== null && v !== undefined && v >= 0)
          if (!hasValidGroup) {
            const uid = props.player.uid || props.player.UID
            let effectiveDivision = divisionFilter.value
            if (effectiveDivision === 'same') {
              effectiveDivision = props.player.division || 'same'
            }
            fetchPlayerPercentiles(uid, effectiveDivision, 'Global').then((updated) => {
              if (updated && detailedPlayerData.value) {
                detailedPlayerData.value.performancePercentiles = {
                  ...detailedPlayerData.value.performancePercentiles,
                  ...updated,
                }
                percentileUpdateCounter.value++
                percentileDataTrigger.value++
                forceRecompute.value++
              }
            })
          }

          // Clear caches to force recomputation
          performanceStatsCache.clear()
          performanceComparisonOptionsCache.clear()
        } else {
          // OPTIMIZATION: Check cache first
          const playerUID = props.player.uid || props.player.UID
          const cacheStartTime = performance.now()
          const cachedData = getCachedPlayerData(props.datasetId, playerUID)
          const _cacheTime = performance.now() - cacheStartTime

          if (cachedData) {
            detailedPlayerData.value = cachedData
            const _loadTime = performance.now() - startTime

            // Eagerly fetch if cached data is missing percentiles for the current group
            const currentGroup = selectedComparisonGroup.value
            const existingGroup = cachedData.performancePercentiles?.[currentGroup]
            const hasValidGroup =
              existingGroup &&
              Object.values(existingGroup).some((v) => v !== null && v !== undefined && v >= 0)
            if (!hasValidGroup) {
              const uid = cachedData.uid || cachedData.UID
              let effectiveDivision = divisionFilter.value
              if (effectiveDivision === 'same') {
                effectiveDivision = cachedData.division || 'same'
              }
              fetchPlayerPercentiles(uid, effectiveDivision, 'Global').then((updated) => {
                if (updated && detailedPlayerData.value) {
                  detailedPlayerData.value.performancePercentiles = {
                    ...(detailedPlayerData.value.performancePercentiles || {}),
                    ...updated,
                  }
                  percentileUpdateCounter.value++
                  percentileDataTrigger.value++
                  forceRecompute.value++
                }
              })
            }
            return
          }
          // OPTIMIZATION: Fetch player data and percentiles in parallel

          // Handle the 'same' division filter by converting it to the player's actual division
          let effectiveDivision = divisionFilter.value
          if (divisionFilter.value === 'same') {
            const targetDivision = props.player?.division
            if (targetDivision) {
              effectiveDivision = targetDivision
            } else {
              // If no target division is available, fall back to 'same'
              effectiveDivision = 'same'
            }
          }

          // OPTIMIZATION: Parallel API calls for better performance
          const apiStartTime = performance.now()

          const [playerResult, percentileResult] = await Promise.allSettled([
            fetchFullPlayerStats(props.datasetId, playerUID),
            fetchPlayerPercentiles(playerUID, effectiveDivision, selectedComparisonGroup.value),
          ])

          const _apiTime = performance.now() - apiStartTime

          // Handle player data result
          if (
            playerResult.status === 'fulfilled' &&
            playerResult.value.format === 'json' &&
            playerResult.value.data.player
          ) {
            console.log('API returned player data:', {
              dataKeys: Object.keys(playerResult.value.data),
              playerKeys: playerResult.value.data.player
                ? Object.keys(playerResult.value.data.player)
                : [],
              hasAttributes: playerResult.value.data.player?.attributes
                ? Object.keys(playerResult.value.data.player.attributes).length
                : 0,
              hasPerformancePercentiles: playerResult.value.data.player?.performancePercentiles
                ? Object.keys(playerResult.value.data.player.performancePercentiles).length
                : 0,
            })

            detailedPlayerData.value = playerResult.value.data.player
            const _playerDataTime = performance.now() - apiStartTime

            // Cache the player data for future use
            setCachedPlayerData(props.datasetId, playerUID, playerResult.value.data.player)
          } else {
            console.error('Failed to fetch player data:', {
              status: playerResult.status,
              format: playerResult.value?.format,
              hasData: !!playerResult.value?.data,
              hasPlayer: !!playerResult.value?.data?.player,
            })
            throw new Error('Failed to fetch player data')
          }

          // Handle percentile result
          if (
            percentileResult.status === 'fulfilled' &&
            percentileResult.value &&
            detailedPlayerData.value
          ) {
            if (!detailedPlayerData.value.performancePercentiles) {
              detailedPlayerData.value.performancePercentiles = {}
            }
            Object.assign(detailedPlayerData.value.performancePercentiles, percentileResult.value)

            // Force reactivity by incrementing the counters
            percentileUpdateCounter.value++
            percentileDataTrigger.value++
            forceRecompute.value++

            const _percentileTime = performance.now() - apiStartTime
          } else {
            logger.warn('Failed to fetch percentiles, will retry', {
              player_name: detailedPlayerData.value?.name,
              error: percentileResult.reason,
            })
          }

          // Clear caches to force recomputation
          performanceStatsCache.clear()
          performanceComparisonOptionsCache.clear()

          const _totalTime = performance.now() - startTime
        }
      } catch (error) {
        const errorTime = performance.now() - startTime
        detailedDataError.value = error.message
        logger.error('Failed to fetch detailed player data', {
          error: error.message,
          time_ms: Math.round(errorTime),
        })
      } finally {
        isLoadingDetailedData.value = false
      }
    }

    // Watch for player changes to fetch detailed data with debouncing
    let fetchTimeout = null
    watch(
      () => props.player,
      (newPlayer) => {
        //   console.log('PlayerDetailDialog watch triggered:', {
        //     hasNewPlayer: !!newPlayer,
        //     playerName: newPlayer?.name,
        //     playerUID: newPlayer?.uid || newPlayer?.UID
        //   })

        if (newPlayer && (newPlayer.uid || newPlayer.UID)) {
          console.log('Setting up fetchDetailedPlayerData timeout')
          // Clear any existing timeout to prevent rapid successive calls
          if (fetchTimeout) {
            clearTimeout(fetchTimeout)
          }

          // Debounce the fetch to prevent rapid successive API calls
          fetchTimeout = setTimeout(() => {
            console.log('fetchDetailedPlayerData timeout triggered')
            const dialogOpenStartTime = performance.now()

            // Reset detailed data when player changes
            detailedPlayerData.value = null
            detailedDataError.value = null

            // Fetch detailed data with optimized loading
            fetchDetailedPlayerData().catch((error) => {
              const errorTime = performance.now() - dialogOpenStartTime
              console.error('PlayerDetailDialog failed to load', {
                player_name: newPlayer.name,
                player_uid: newPlayer.uid || newPlayer.UID,
                error: error.message,
                time_ms: Math.round(errorTime),
              })
            })
          }, 50) // Small debounce to prevent rapid successive calls
        }
      },
      { immediate: true }
    )

    // Force recomputation when displayPlayer changes
    watch(
      () => displayPlayer.value,
      (_newPlayer, _oldPlayer) => {
        percentileUpdateCounter.value++
        percentileDataTrigger.value++
        forceRecompute.value++
      }
    )

    // Force recomputation when forceRecompute changes
    watch(
      () => forceRecompute.value,
      (_newValue, _oldValue) => {
        // Force recomputation when forceRecompute changes
      }
    )

    // Watch for changes in categorizedPerformanceStats
    watch(
      () => categorizedPerformanceStats.value,
      (_newStats, _oldStats) => {
        // Stats changed
      },
      { deep: true }
    )

    // Watch for changes in hasAnyPerformanceData
    watch(
      () => hasAnyPerformanceData.value,
      (_newValue, _oldValue) => {
        // Performance data availability changed
      }
    )

    // Optimized player watcher with cache cleanup - moved after displayPlayer definition
    watch(
      () => props.player,
      (newPlayer, oldPlayer) => {
        // Clear caches when player changes to prevent stale data
        if (oldPlayer && newPlayer !== oldPlayer) {
          clearAllCaches()
        }

        flagLoadError.value = false
        faceImageLoadError.value = false
        shouldShowTeamLogo.value = false

        // Reset to Global initially when player changes
        selectedComparisonGroup.value = 'Global'

        // Reset division filter based on whether the new player is in a top 5 league
        divisionFilter.value = getDefaultDivisionFilter(newPlayer)
      },
      { immediate: true }
    )

    // Watch for detailed player data changes to update comparison group
    watch(
      () => detailedPlayerData.value,
      (newData, oldData) => {
        // Clear performance stats cache when detailed player data changes
        performanceStatsCache.clear()

        // Reset loading state when detailed player data changes
        if (newData !== oldData) {
          // Force recomputation to ensure loading state updates
          percentileUpdateCounter.value++
          percentileDataTrigger.value++
          forceRecompute.value++
        }

        // Update comparison group when detailed data is available
        if (newData?.performancePercentiles) {
          updateComparisonGroupForPlayer(newData)
        }
      }
    )

    // Function to update comparison group for a specific player
    const updateComparisonGroupForPlayer = (player) => {
      if (!player?.performancePercentiles) {
        selectedComparisonGroup.value = 'Global'
        return
      }

      const newOptions = performanceComparisonOptions.value
      const highestRoleGroup = getPositionGroupForHighestRole()

      // Get the player's available positions
      const _playerPositions = deriveShortPositions(player)
      const playerPositionGroups = derivePositionGroups(player)

      // Priority-based selection logic
      // 1. Try to use the highest-rated role's position group if available
      if (highestRoleGroup && newOptions.some((opt) => opt.value === highestRoleGroup)) {
        selectedComparisonGroup.value = highestRoleGroup
      }
      // 2. Try to use the first available position group that the player can play
      else {
        const availablePlayerGroup = newOptions.find((opt) =>
          playerPositionGroups.includes(opt.value)
        )
        if (availablePlayerGroup) {
          selectedComparisonGroup.value = availablePlayerGroup.value
        }
        // 3. Fall back to Global if available
        else if (newOptions.some((opt) => opt.value === 'Global')) {
          selectedComparisonGroup.value = 'Global'
        }
        // 4. Use first available option
        else if (newOptions.length > 0) {
          selectedComparisonGroup.value = newOptions[0].value
        }
        // 5. Default to Global
        else {
          selectedComparisonGroup.value = 'Global'
        }
      }
    }

    // Watch for comparison group changes to clear performance stats cache
    watch(
      () => selectedComparisonGroup.value,
      async (newGroup, oldGroup) => {
        performanceStatsCache.clear()

        if (newGroup !== oldGroup && displayPlayer.value && props.datasetId) {
          const playerUID = displayPlayer.value.uid || displayPlayer.value.UID
          if (playerUID) {
            // Skip the API call if valid percentiles already exist for this group
            const existingGroup = detailedPlayerData.value?.performancePercentiles?.[newGroup]
            const hasExistingData =
              existingGroup &&
              Object.values(existingGroup).some((v) => v !== null && v !== undefined && v >= 0)
            if (hasExistingData) {
              percentileUpdateCounter.value++
              return
            }

            let effectiveDivision = divisionFilter.value
            if (divisionFilter.value === 'same') {
              effectiveDivision = displayPlayer.value?.division || 'same'
            }

            // Always fetch 'Global' so all position groups arrive in one request
            const updatedPercentiles = await fetchPlayerPercentiles(
              playerUID,
              effectiveDivision,
              'Global'
            )

            if (updatedPercentiles && detailedPlayerData.value) {
              // Merge rather than replace so already-cached groups are preserved
              detailedPlayerData.value.performancePercentiles = {
                ...(detailedPlayerData.value.performancePercentiles || {}),
                ...updatedPercentiles,
              }
              percentileUpdateCounter.value++
              percentileDataTrigger.value++
              forceRecompute.value++
              performanceStatsCache.clear()
              performanceComparisonOptionsCache.clear()
            }
          }
        }
      }
    )

    // Watch for division filter changes to clear performance stats cache
    watch(
      () => divisionFilter.value,
      async (newFilter, oldFilter) => {
        // Clear performance stats cache when division filter changes
        performanceStatsCache.clear()

        // Call the division filter change handler
        if (newFilter !== oldFilter) {
          await onDivisionFilterChange()
        }
      }
    )

    // Watch for detailed player data changes to clear performance stats cache
    watch(
      () => detailedPlayerData.value,
      (newData, oldData) => {
        // Clear performance stats cache when detailed player data changes
        performanceStatsCache.clear()

        // Reset loading state when detailed player data changes
        if (newData !== oldData) {
          // Force recomputation to ensure loading state updates
          percentileUpdateCounter.value++
          percentileDataTrigger.value++
          forceRecompute.value++
        }
      }
    )

    // Watch for performance percentiles changes to increment counter
    watch(
      () => detailedPlayerData.value?.performancePercentiles,
      (newPercentiles) => {
        if (newPercentiles) {
          // Force reactivity by incrementing the counters
          percentileUpdateCounter.value++
          percentileDataTrigger.value++
          forceRecompute.value++
        }
      },
      { deep: true }
    )

    // Helper functions to derive position information from available data
    // Memoized performance comparison options with better caching - moved after displayPlayer definition
    const performanceComparisonOptionsCache = new Map()

    const performanceComparisonOptions = computed(() => {
      if (!displayPlayer.value) {
        return []
      }

      const player = displayPlayer.value

      // Derive position information from available data
      const derivedShortPositions = deriveShortPositions(player)
      const derivedPositionGroups = derivePositionGroups(player)

      const cacheKey = `${getCacheKey(player, 'options')}-${JSON.stringify(derivedShortPositions)}-${JSON.stringify(derivedPositionGroups)}`

      if (performanceComparisonOptionsCache.has(cacheKey)) {
        return performanceComparisonOptionsCache.get(cacheKey)
      }

      const playerShortPositions = derivedShortPositions
      const playerBroadGroups = derivedPositionGroups
      const options = []

      // Pre-create sets for faster lookups
      const broadGroupsSet = new Set(playerBroadGroups)
      const shortPositionsSet = new Set(playerShortPositions)
      const addedValues = new Set()

      const preferredOrder = [
        'Global',
        'Goalkeepers',
        'Defenders',
        'Midfielders',
        'Attackers',
        'Full-backs',
        'Centre-backs',
        'Wing-backs',
        'Defensive Midfielders',
        'Central Midfielders',
        'Wide Midfielders',
        'Attacking Midfielders (Central)',
        'Wingers',
        'Strikers',
      ]

      const shouldIncludeGroup = (groupKey) => {
        if (groupKey === 'Global') return true

        // Check broad groups
        if (broadGroupsSet.has(groupKey)) return true

        // Check detailed groups
        const requiredPositions = detailedGroupToShortPositionsMap[groupKey]
        if (requiredPositions) {
          return requiredPositions.some((pos) => shortPositionsSet.has(pos))
        }

        return false
      }

      // Process preferred order groups - don't check if they exist in current percentiles
      for (let i = 0; i < preferredOrder.length; i++) {
        const groupKey = preferredOrder[i]
        if (shouldIncludeGroup(groupKey) && !addedValues.has(groupKey)) {
          options.push({
            label: groupKey === 'Global' ? 'Overall Dataset' : `vs. ${groupKey}`,
            value: groupKey,
          })
          addedValues.add(groupKey)
        }
      }

      // Clean up cache if it gets too large
      if (performanceComparisonOptionsCache.size > 20) {
        performanceComparisonOptionsCache.clear()
      }

      performanceComparisonOptionsCache.set(cacheKey, options)
      return options
    })

    const getPositionGroupForHighestRole = () => {
      if (!displayPlayer.value?.roleSpecificOveralls?.length) {
        return null
      }

      // Find the role with the highest overall rating
      const highestRole = displayPlayer.value.roleSpecificOveralls.reduce((max, role) =>
        role.score > max.score ? role : max
      )

      // Extract the short position from the role name (e.g., "DM" from "DM - Defensive Midfielder - Support")
      const shortPosition = highestRole.roleName.split(' - ')[0]

      // Map short position to detailed group first
      const detailedGroup = Object.entries(detailedGroupToShortPositionsMap).find(
        ([_groupName, positions]) => positions.includes(shortPosition)
      )

      if (detailedGroup) {
        return detailedGroup[0] // Return the detailed group name
      }

      // Fallback to broad position group
      const positionToGroupMap = {
        GK: 'Goalkeepers',
        SW: 'Defenders',
        DR: 'Defenders',
        DL: 'Defenders',
        DC: 'Defenders',
        WBR: 'Defenders',
        WBL: 'Defenders', // Wing-backs are in defenders for broad groups
        DM: 'Midfielders',
        MC: 'Midfielders',
        MR: 'Midfielders',
        ML: 'Midfielders',
        AMC: 'Midfielders',
        AMR: 'Midfielders',
        AML: 'Midfielders',
        ST: 'Attackers',
      }

      return positionToGroupMap[shortPosition] || null
    }

    // Updated attribute categories to use displayPlayer
    const attributeCategories = computed(() => {
      if (!displayPlayer.value?.attributes || !displayPlayer.value.name) {
        return {
          technical: [],
          mental: [],
          physical: [],
          goalkeeping: [],
        }
      }

      const playerAttributes = displayPlayer.value.attributes
      const hasAttribute = (key) => Object.hasOwn(playerAttributes, key)

      return {
        technical: technicalAttrsOrdered.filter(hasAttribute),
        mental: mentalAttrsOrdered.filter(hasAttribute),
        physical: physicalAttrsOrdered.filter(hasAttribute),
        goalkeeping: isGoalkeeper.value ? goalkeepingAttrsOrdered.filter(hasAttribute) : [],
      }
    })

    // Computed property for player positions using the deriveShortPositions function
    const playerPositions = computed(() => {
      if (!displayPlayer.value) return []
      return deriveShortPositions(displayPlayer.value)
    })

    // Enhanced cache with access tracking
    const _getCachedData = (cache, key) => {
      const data = cache.get(key)
      if (data) {
        data._lastAccess = Date.now()
        return data
      }
      return null
    }

    // Enhanced cache setter with access tracking
    const _setCachedData = (cache, key, data) => {
      data._lastAccess = Date.now()
      cache.set(key, data)

      // Trigger cleanup if cache is getting large
      if (cache.size > maxCacheSize * 0.9) {
        manageCacheSize()
      }
    }

    // Copy player name to clipboard
    const copyPlayerName = async () => {
      const playerName = displayPlayer.value?.name || 'Unknown Player'
      try {
        if (navigator.clipboard && window.isSecureContext) {
          // Use modern Clipboard API
          await navigator.clipboard.writeText(playerName)
        } else {
          // Fallback for older browsers or non-secure contexts
          const textArea = document.createElement('textarea')
          textArea.value = playerName
          textArea.style.position = 'fixed'
          textArea.style.left = '-999999px'
          textArea.style.top = '-999999px'
          document.body.appendChild(textArea)
          textArea.focus()
          textArea.select()
          document.execCommand('copy')
          textArea.remove()
        }

        // Show success notification
        qInstance.notify({
          type: 'positive',
          message: `Copied "${playerName}" to clipboard`,
          position: 'top',
          timeout: 2000,
          icon: 'content_copy',
        })
      } catch (error) {
        logger.error('Failed to copy player name to clipboard', {
          error: error.message,
          player_name: playerName,
        })

        // Show error notification
        qInstance.notify({
          type: 'negative',
          message: 'Failed to copy to clipboard',
          position: 'top',
          timeout: 2000,
          icon: 'error',
        })
      }
    }

    const isCurrentPlayerInComparison = computed(() => {
      const p = props.player
      const dsId = playerStore.currentDatasetId
      if (!p || !dsId) return false
      return comparisonStore.isInComparison(dsId, p)
    })

    const toggleComparison = () => {
      const p = props.player
      const dsId = playerStore.currentDatasetId
      if (!p || !dsId) return
      if (comparisonStore.isInComparison(dsId, p)) {
        comparisonStore.removeFromComparison(dsId, p)
        qInstance.notify({
          type: 'info',
          message: `${p.name} removed from comparison`,
          timeout: 1500,
        })
      } else {
        const added = comparisonStore.addToComparison(dsId, p)
        if (added) {
          qInstance.notify({
            type: 'positive',
            message: `${p.name} added to comparison`,
            timeout: 1500,
          })
        } else {
          qInstance.notify({
            type: 'warning',
            message: 'Comparison is full (max 4 players)',
            timeout: 1500,
          })
        }
      }
    }

    return {
      qInstance,
      showCardGenerator,
      attributeCategories,
      attributeFullNameMap,
      attributeDescriptions,
      fifaAttributeDescriptions,
      fifaToFmAttributeMapping,
      getUnifiedRatingClass,
      getBarFillStyle,
      fifaStatsToDisplay,
      getFifaStatValue,
      basicPlayer,
      sortedRoleSpecificOveralls,
      isGoalkeeper,
      formattedTransferValue,
      formattedWage,
      currencyIcon,
      selectedComparisonGroup,
      performanceComparisonOptions,
      categorizedPerformanceStats,
      hasAnyPerformanceData,
      flagLoadError,
      handleFlagError,
      divisionFilter,
      divisionFilterOptions,
      onDivisionFilterChange,
      faceImageLoadError,
      handleFaceImageError,
      handleFaceImageLoad,
      playerFaceImageUrl,
      showFaces: computed(() => uiStore.showFaces),
      getDisplayAttribute,
      shouldShowTeamLogo,
      showLogoCorrections,
      logoResolution,
      logoOverrideState,
      isSubmittingLogoOverride,
      logoKey,
      showingAlternatives,
      logoAlternatives,
      logoSearchQuery,
      logoSearchResults,
      isSearchingLogo,
      searchLogoAlternatives,
      confirmLogo,
      rejectLogo,

      // Attribute order arrays for loading placeholders
      technicalAttrsOrdered,
      mentalAttrsOrdered,
      physicalAttrsOrdered,
      goalkeepingAttrsOrdered,

      // Percentile retry functionality
      isLoadingPercentiles,
      hasValidPercentiles,
      percentilesNeedRetry,
      showLoadingState,
      percentilesRetryCount,
      maxRetries,
      manualRetry,

      // Detailed player data
      detailedPlayerData,
      isLoadingDetailedData,
      detailedDataError,
      displayPlayer,
      cardDisplayPlayer,
      forceRecompute,
      updateComparisonGroupForPlayer,
      isDarkMode, // <-- add this line

      // Player positions computed property
      playerPositions,
      bestPlaystyleTagline,

      // Tab system
      activeTab,

      // Clipboard functionality
      copyPlayerName,

      comparisonStore,
      playerStore,
      isCurrentPlayerInComparison,
      toggleComparison,
    }
  },
})
</script>

<style lang="scss" scoped>

// Modern Dialog Card
.player-detail-dialog-card {
    display: flex;
    flex-direction: column;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-3);
    overflow: hidden;
    background: var(--surface-card);
}

// Dialog backdrop styling lives in the unscoped block at the end of this file.

// Dialog chrome: unified header + tab-strip convention shared with
// UpgradeFinderDialog's `.card-header` — icon, title, actions, close, all in
// normal flow above the content (not overlaid).
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

.dialog-chrome__action-btn,
.dialog-chrome__close {
    transition: transform 0.15s ease;

    &:hover {
        transform: scale(1.08);
    }
}

.dialog-chrome__tabs {
    padding: 0 12px 8px;

    &--snapshot {
        padding-top: 6px;
        padding-bottom: 4px;
        border-bottom: 1px solid var(--surface-border);
    }
}

.main-content-section {
    flex-grow: 1;
    padding: 20px;
    background: transparent;
}

// Modern Select Styling
.modern-select {
    :deep(.q-field__control) {
        border-radius: 8px;
        background: var(--surface-raised);
    }
}

// Performance Card
.performance-percentiles-card {
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-2);
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
}

.performance-card-header {
    background: linear-gradient(135deg, var(--accent-soft-strong) 0%, var(--accent-soft) 100%);
    border-radius: 12px 12px 0 0;
    padding: 16px 20px;
}

.performance-header-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--accent);
    display: flex;
    align-items: center;
}

.performance-category-header {
    margin-bottom: 8px;
}

.performance-category-title {
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.performance-separator {
    background: color-mix(in srgb, var(--accent) 15%, transparent);
    height: 1px;
    border: none;
}

.modern-stat-item {
    transition: background-color 0.2s ease;
    border-radius: 6px;
    margin: 2px 0;
    padding: 8px 12px;

    &:hover {
        background: var(--accent-soft);
    }

    &.average-rating-item {
        background: var(--accent-soft-strong);
        border-left: 4px solid var(--accent);
    }
}

.stat-name-section {
    flex-basis: 45%;
    flex-grow: 0;
    flex-shrink: 0;
    padding-right: 8px;
}

.stat-bar-section {
    flex-grow: 1;
    display: flex;
    align-items: center;
}

.stat-value-section {
    flex-basis: 18%;
    flex-grow: 0;
    flex-shrink: 0;
    text-align: right;
    padding-left: 8px;
}

.stat-name-label {
    font-size: 0.8rem;
    font-weight: 500;
    color: var(--text-primary);
}

.stat-bar-container {
    display: flex;
    align-items: center;
    width: 100%;
}

.stat-bar-track {
    flex-grow: 1;
    height: 10px;
    background-color: var(--surface-border-strong);
    border-radius: 5px;
    margin-right: 8px;
    overflow: hidden;
}

.stat-bar-fill {
    height: 100%;
    border-radius: 5px;
    transition: width 0.5s ease, background-color 0.3s ease;
}

.stat-percentile-text {
    font-size: 0.7rem;
    font-weight: 600;
    min-width: 24px;
    text-align: right;
    color: var(--text-secondary);
}

.performance-stat-value {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-primary);
}

.percentile-loading-area {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 3rem 2rem;
}

.loading-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 1rem;
}

.loading-text {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.no-performance-data {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 24px;
    color: var(--text-secondary);
    font-size: 0.9rem;
    gap: 0.5rem;
    text-align: center;
}

// Profile Card
.player-profile-card {
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-2);
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
}

.player-profile-content {
    padding: 24px;
}

.profile-header-section {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 0;
    gap: 24px;

    @media (max-width: 768px) {
        flex-direction: column;
        gap: 16px;
    }
}

.player-identity-section {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
}

.player-face-container {
    display: flex;
    align-items: center;
    justify-content: center;
}

.player-face-image {
    border-radius: 50%;
    border: 3px solid var(--accent-soft-strong);
    object-fit: cover;
    transition: all 0.3s ease;
    box-shadow: var(--shadow-2);

    &:hover {
        transform: scale(1.05);
        border-color: color-mix(in srgb, var(--accent) 40%, transparent);
        box-shadow: var(--shadow-3);
    }
}

.player-face-placeholder {
    border: 3px solid var(--accent-soft-strong);
    transition: all 0.3s ease;
    display: flex;
    align-items: center;
    justify-content: center;

    &:hover {
        transform: scale(1.05);
        border-color: color-mix(in srgb, var(--accent) 40%, transparent);
    }

    .q-icon {
        margin: 0;
        line-height: 1;
    }
}

.player-flag-container {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
}

.player-flag {
    border: 1px solid var(--surface-border-strong);
    border-radius: 4px;
    object-fit: cover;
    box-shadow: var(--shadow-1);
    transition: all 0.2s ease;

    &:hover {
        transform: scale(1.05);
        box-shadow: var(--shadow-2);
    }
}

.player-flag-placeholder {
    color: var(--text-muted);
}

.player-name-section {
    display: flex;
    flex-direction: column;
    justify-content: center;
    overflow: hidden;
}

.player-name-container {
    display: flex;
    flex-direction: column;
}

.player-name-text {
    font-size: 1.8rem;
    line-height: 1.2;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 8px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.player-badges-row {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
}

.player-age-badge,
.player-nationality-badge {
    font-size: 0.8rem;
    font-weight: 600;
    padding: 4px 8px;
    transition: all 0.2s ease;
    
    &:hover {
        transform: scale(1.05);
    }
}

.player-positions-section {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}

.position-badge {
    font-size: 0.75rem;
    font-weight: 600;
    padding: 3px 6px;
    transition: all 0.2s ease;
    
    &:hover {
        transform: scale(1.05);
    }
}

.player-details-grid {
    display: flex;
    gap: 12px;
    margin-top: 16px;
    
    @media (max-width: 768px) {
        flex-direction: column;
        gap: 12px;
    }
}

.detail-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: color-mix(in srgb, var(--accent) 4%, transparent);
    border-radius: 8px;
    border: 1px solid color-mix(in srgb, var(--accent) 10%, transparent);
    transition: all 0.2s ease;
    flex: 1;
    min-width: 0;

    &:hover {
        background: var(--accent-soft);
        border-color: var(--accent-soft-strong);
        transform: translateY(-1px);
    }
}

.detail-icon {
    color: var(--accent);
}

.detail-content {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    overflow: hidden;
}

.detail-label {
    font-size: 0.7rem;
    color: var(--text-secondary);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 2px;
}

.detail-value {
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.financial-details-section {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    min-width: 200px;
    
    @media (max-width: 768px) {
        align-items: flex-start;
        width: 100%;
    }
}

.financial-combined-item {
    display: flex;
    flex-direction: column;
    padding: 16px;
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.05) 0%, rgba(59, 130, 246, 0.05) 100%);
    border-radius: 12px;
    border: 1px solid rgba(34, 197, 94, 0.15);
    transition: all 0.2s ease;
    
    &:hover {
        background: linear-gradient(135deg, rgba(34, 197, 94, 0.08) 0%, rgba(59, 130, 246, 0.08) 100%);
        border-color: rgba(34, 197, 94, 0.25);
        transform: translateY(-1px);
    }
    
    .body--dark & {
        background: linear-gradient(135deg, rgba(34, 197, 94, 0.08) 0%, rgba(59, 130, 246, 0.08) 100%);
        border-color: rgba(34, 197, 94, 0.2);
        
        &:hover {
            background: linear-gradient(135deg, rgba(34, 197, 94, 0.12) 0%, rgba(59, 130, 246, 0.12) 100%);
            border-color: rgba(34, 197, 94, 0.3);
        }
    }
}

.financial-row {
    display: flex;
    align-items: center;
}

.financial-item-content {
    display: flex;
    flex-direction: column;
}

.financial-item-large,
.financial-item-small {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: rgba(34, 197, 94, 0.05);
    border-radius: 8px;
    border: 1px solid rgba(34, 197, 94, 0.15);
    transition: all 0.2s ease;
    
    &:hover {
        background: rgba(34, 197, 94, 0.08);
        border-color: rgba(34, 197, 94, 0.25);
        transform: translateY(-1px);
    }
    
    .body--dark & {
        background: rgba(34, 197, 94, 0.08);
        border-color: rgba(34, 197, 94, 0.2);
        
        &:hover {
            background: rgba(34, 197, 94, 0.12);
            border-color: rgba(34, 197, 94, 0.3);
        }
    }
}

.financial-item-small {
    background: rgba(59, 130, 246, 0.05);
    border-color: rgba(59, 130, 246, 0.15);
    
    &:hover {
        background: rgba(59, 130, 246, 0.08);
        border-color: rgba(59, 130, 246, 0.25);
    }
    
    .body--dark & {
        background: rgba(59, 130, 246, 0.08);
        border-color: rgba(59, 130, 246, 0.2);
        
        &:hover {
            background: rgba(59, 130, 246, 0.12);
            border-color: rgba(59, 130, 246, 0.3);
        }
    }
}

.financial-icon {
    color: #059669;
    
    .financial-item-small & {
        color: #2563eb;
    }
    
    .body--dark & {
        color: #34d399;
        
        .financial-item-small & {
            color: #60a5fa;
        }
    }
}

.financial-content {
    display: flex;
    flex-direction: column;
}

.financial-label {
    font-size: 0.7rem;
    color: var(--text-secondary);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 2px;
}

.financial-value {
    font-size: 1.1rem;
    font-weight: 700;
    white-space: nowrap;
    
    &.transfer-value {
        color: #059669;
        
        .body--dark & {
            color: #34d399;
        }
    }
    
    &.wage-value {
        color: #2563eb;
        
        .body--dark & {
            color: #60a5fa;
        }
    }
}

.profile-separator {
    background: linear-gradient(90deg, transparent 0%, color-mix(in srgb, var(--accent) 30%, transparent) 50%, transparent 100%);
    height: 2px;
    border: none;
}

// FIFA Stats Section
.fifa-stats-section {
    margin-top: 4px;
}

.section-header {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
}

.section-title {
    font-size: 1.2rem;
    font-weight: 600;
    color: var(--accent);
}

.fifa-stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(80px, 1fr));
    gap: 8px;
    
    @media (max-width: 768px) {
        grid-template-columns: repeat(auto-fit, minmax(70px, 1fr));
        gap: 6px;
    }
}

.fifa-stat-card {
    display: flex;
    flex-direction: column;
}

.fifa-stat-item {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 12px 8px;
    border-radius: 8px;
    transition: all 0.2s ease;
    min-height: 70px;
    border-width: 1px;
    
    &:hover {
        transform: scale(1.05);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }
    
    // Use the same colors as global rating tier classes from app.scss
    &.rating-na {
        background-color: #bdbdbd;
        color: #424242;
        
        .body--dark & {
            background-color: #424242;
            color: #bdbdbd;
        }
    }
    &.rating-tier-1 {
        background-color: #ef5350;
        color: white;
        
        .body--dark & {
            background-color: #e53935;
            color: white;
        }
    }
    &.rating-tier-2 {
        background-color: #ffa726;
        color: #333333;
        
        .body--dark & {
            background-color: #fb8c00;
            color: white;
        }
    }
    &.rating-tier-3 {
        background-color: #42a5f5;
        color: white;
        
        .body--dark & {
            background-color: #2196f3;
            color: white;
        }
    }
    &.rating-tier-4 {
        background-color: #66bb6a;
        color: white;
        
        .body--dark & {
            background-color: #4caf50;
            color: white;
        }
    }
    &.rating-tier-5 {
        background-color: #26a69a;
        color: white;
        
        .body--dark & {
            background-color: #00897b;
            color: white;
        }
    }
    &.rating-tier-6 {
        background-color: #7e57c2;
        color: white;
        
        .body--dark & {
            background-color: #9575cd;
            color: white;
        }
    }
}

.fifa-stat-label {
    font-size: 0.75rem;
    font-weight: 600;
    margin-bottom: 4px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.fifa-stat-value {
    font-size: 1.3rem;
    font-weight: 700;
    line-height: 1;
}

// Attribute Cards
.attribute-columns-container {
    margin-top: 4px;
}

.attribute-columns-container {
    margin-top: 0;
}

.attribute-card {
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-1);
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
    transition: all 0.3s ease;

    &:hover {
        transform: translateY(-2px);
        box-shadow: var(--shadow-2);
    }
}

.full-height-card {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.attribute-card-header {
    background: linear-gradient(135deg, var(--accent-soft-strong) 0%, var(--accent-soft) 100%);
    border-radius: 12px 12px 0 0;
    padding: 16px 20px;
}

.attribute-section-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--accent);
    display: flex;
    align-items: center;
}

.attribute-list {
    flex-grow: 1;

    .q-item {
        padding: 4px 12px;
        min-height: 32px;
    }
}

.modern-attribute-item {
    transition: background-color 0.2s ease;
    border-radius: 6px;
    margin: 1px 4px;
    padding: 4px 12px;
    min-height: 32px;

    &:hover {
        background: var(--accent-soft);
    }

    &.role-item {
        &.best-role-highlight {
            background: rgba(34, 197, 94, 0.1);
            border-left: 4px solid #22c55e;
            
            .body--dark & {
                background: rgba(34, 197, 94, 0.15);
                border-left-color: #34d399;
            }
            
            .role-name {
                font-weight: 700;
            }
        }
    }
}

.attribute-name {
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    &.role-name {
        max-width: 180px;
    }
}

.modern-attribute-value {
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 700;
    text-align: center;
    min-width: 32px;
    display: inline-block;
    transition: all 0.2s ease;
    
    &:hover {
        transform: scale(1.05);
    }
    
    // Rating color classes - matching PlayerDataTable with correct purple for elite
    &.rating-na {
        background-color: #bdbdbd;
        color: #424242;
        
        .body--dark & {
            background-color: #424242;
            color: #bdbdbd;
        }
    }
    &.rating-tier-1 {
        background-color: #ef5350;
        color: white;
        
        .body--dark & {
            background-color: #e53935;
            color: white;
        }
    }
    &.rating-tier-2 {
        background-color: #ffa726;
        color: #333333;
        
        .body--dark & {
            background-color: #fb8c00;
            color: white;
        }
    }
    &.rating-tier-3 {
        background-color: #42a5f5;
        color: white;
        
        .body--dark & {
            background-color: #2196f3;
            color: white;
        }
    }
    &.rating-tier-4 {
        background-color: #66bb6a;
        color: white;
        
        .body--dark & {
            background-color: #4caf50;
            color: white;
        }
    }
    &.rating-tier-5 {
        background-color: #26a69a;
        color: white;
        
        .body--dark & {
            background-color: #00897b;
            color: white;
        }
    }
    &.rating-tier-6 {
        background-color: #7e57c2;
        color: white;
        
        .body--dark & {
            background-color: #9575cd;
            color: white;
        }
    }
}

.no-attributes-item {
    opacity: 0.7;
    font-style: italic;
    color: var(--text-secondary);
}

.role-ratings-card {
    .role-specific-ratings-list {
        max-height: 280px;
        overflow-y: auto;

        &::-webkit-scrollbar {
            width: 4px;
        }

        &::-webkit-scrollbar-track {
            background: transparent;
        }

        &::-webkit-scrollbar-thumb {
            background: color-mix(in srgb, var(--accent) 30%, transparent);
            border-radius: 2px;

            &:hover {
                background: color-mix(in srgb, var(--accent) 50%, transparent);
            }
        }
    }
}

// Modern Tooltips
.modern-tooltip {
    border-radius: 8px;
    box-shadow: var(--shadow-2);

    .tooltip-header {
        font-weight: 600;
        margin-bottom: 6px;
        font-size: 0.9rem;
    }
    
    .tooltip-description {
        font-size: 0.8rem;
        line-height: 1.4;
        opacity: 0.9;
    }
}

// Loading Section
.loading-section {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 60px 20px;
}

.loading-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
}

.loading-text {
    color: var(--text-secondary);
    font-size: 1rem;
    font-weight: 500;
}

// Responsive Design
@media (max-width: 768px) {
    .main-content-section {
        padding: 16px;
    }

    .dialog-chrome__header {
        padding: 10px 12px;
    }

    .dialog-chrome__tabs {
        padding: 0 8px 6px;

        .view-tabs {
            gap: 3px;

            .q-tab {
                min-height: 32px;
                height: 32px;
                padding: 0 8px;
                font-size: 0.6rem;
                transition: none;
            }
        }
    }

    .player-profile-content {
        padding: 20px;
    }
    
    .player-name-text {
        font-size: 1.5rem;
    }
    
    .fifa-stats-grid {
        grid-template-columns: repeat(auto-fit, minmax(65px, 1fr));
        gap: 4px;
    }
    
    .fifa-stat-item {
        padding: 8px 6px;
        min-height: 60px;
    }
    
    .fifa-stat-label {
        font-size: 0.7rem;
    }
    
    .fifa-stat-value {
        font-size: 1.1rem;
    }
    
    .player-details-grid {
        grid-template-columns: 1fr;
        gap: 12px;
    }
    
    .financial-details-section {
        width: 100%;
    }
    
    .financial-item-large,
    .financial-item-small {
        width: 100%;
    }
}

@media (max-width: 480px) {
    .main-content-section {
        padding: 12px;
    }

    .dialog-chrome__header {
        padding: 8px 8px;
        gap: 0.4rem;
    }

    .dialog-chrome__title {
        font-size: 1rem;
    }

    .dialog-chrome__tabs {
        padding: 0 6px 4px;

        .view-tabs {
            gap: 2px;

            .q-tab {
                min-height: 28px;
                height: 28px;
                padding: 0 6px;
                font-size: 0.55rem;
                transition: none;
            }
        }
    }

    .player-profile-content {
        padding: 16px;
    }
    
    .player-name-text {
        font-size: 1.3rem;
    }
    
    
    .attribute-name {
        font-size: 0.8rem;
    }
    
    .modern-attribute-value {
        font-size: 0.75rem;
        padding: 3px 6px;
    }
}

// Snapshot Navigation - a second row beneath the Simple/Advanced pill toggle, only present
// when the caller (Progression) passes 2+ snapshots. In-flow, part of .dialog-chrome.
.snapshot-tabs {
    min-height: 32px;

    :deep(.q-tab) {
        min-height: 32px;
        font-size: 0.7rem;
        font-weight: 600;
        text-transform: none;
        color: var(--text-secondary);

        &:hover {
            color: var(--text-primary);
        }
    }
}

// View tabs (Simple/Advanced/AI Scout Report) - pill-shaped toggle style, in-flow.
.view-tabs {
    display: flex;
    gap: 4px;

    .q-tab {
        min-height: 36px;
        height: 36px;
        padding: 0 12px;
        font-size: 0.65rem;
        font-weight: 600;
        border-radius: 18px;
        transition: none;
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--surface-card);
        border: 1px solid var(--surface-border);
        color: var(--text-secondary);

        &:hover {
            background: var(--surface-raised);
            color: var(--text-primary);
        }

        &.q-tab--active {
            background: var(--accent);
            border-color: var(--accent);
            color: var(--text-on-brand);
        }
    }
}

// Simple View Styles
.simple-view {
    .player-card-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 16px;
        margin-bottom: 16px; // ensure space below the card
        
        .player-detail-card {
            --card-scale: 0.86;
            --card-hover-lift: -10px;
            transform-origin: top center;
            margin-bottom: 12px;
            
            @media (max-width: 768px) {
                --card-scale: 0.78;
            }
            
            @media (max-width: 480px) {
                --card-scale: 0.68;
            }
        }
    }
}

// (Advanced view specific styles removed - no empty ruleset)

.club-logo-container {
    display: flex;
    align-items: center;
    justify-content: center;
}

.logo-correction-area {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    margin-top: 4px;
    opacity: 0.75;
    transition: opacity 0.15s;

    &:hover {
        opacity: 1;
    }
}

.logo-correction-row {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
}

.logo-alternatives {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    margin-top: 4px;
}

.logo-alternative-row {
    display: flex;
    align-items: center;
    gap: 6px;
}

.logo-alt-img {
    border-radius: 2px;
    object-fit: contain;
}

.logo-search-area {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
}

.logo-search-row {
    display: flex;
    align-items: center;
    gap: 4px;
}

.logo-search-input {
    width: 140px;
    font-size: 0.75rem;
}

.logo-correction-label {
    font-size: 0.7rem;
    opacity: 0.85;
}

.club-logo-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
}

.club-logo-skeleton {
    opacity: 0.6;
}

.player-club-logo {
    transition: all 0.2s ease;
    
    /* Better image rendering */
    image-rendering: auto;
    image-rendering: smooth;
    image-rendering: high-quality;
    
    &:hover {
        transform: scale(1.1);
    }
}

.scouting-warning {
    font-size: 0.8rem;
    font-weight: 600;
    color: #f57c00;
    display: inline-flex;
    align-items: center;
    padding: 2px 6px;
    border-radius: 4px;
    background: rgba(245, 124, 0, 0.1);
    border: 1px solid rgba(245, 124, 0, 0.3);
    transition: all 0.2s ease;
    cursor: help;
    
    &:hover {
        background: rgba(245, 124, 0, 0.15);
        border-color: rgba(245, 124, 0, 0.4);
        transform: scale(1.02);
    }
    
    .body--dark & {
        color: #ffb74d;
        background: rgba(255, 183, 77, 0.1);
        border-color: rgba(255, 183, 77, 0.3);
        
        &:hover {
            background: rgba(255, 183, 77, 0.15);
            border-color: rgba(255, 183, 77, 0.4);
        }
    }
}

.scouting-warning-icon {
    cursor: help;
    transition: all 0.2s ease;
    vertical-align: middle;
    
    &:hover {
        transform: scale(1.1);
        filter: brightness(1.2);
    }
    
    .body--dark & {
        color: #ffb74d !important;
        
        &:hover {
            color: #fff3c4 !important;
        }
    }
}

.loading-attribute {
    opacity: 0.6;
    animation: pulse 1.5s ease-in-out infinite;
}

.loading-attribute-skeleton {
    border-radius: 4px;
    opacity: 0.7;
    animation: skeleton-pulse 2s ease-in-out infinite;
}

@keyframes pulse {
    0%, 100% {
        opacity: 0.6;
    }
    50% {
        opacity: 1;
    }
}

@keyframes skeleton-pulse {
    0%, 100% {
        opacity: 0.4;
    }
    50% {
        opacity: 0.8;
    }
}

.player-name-with-copy {
    display: flex;
    align-items: center;
    gap: 4px;
    
    .player-name {
        margin: 0;
    }
}

.copy-name-btn {
    opacity: 0.7;
    transition: all 0.2s ease;

    &:hover {
        opacity: 1;
        transform: scale(1.1);
        background: var(--accent-soft-strong);
    }

    &:active {
        transform: scale(0.95);
    }
}
</style>

<style lang="scss">
/* Unscoped: dialog backdrop + card surface overrides for this dialog.
   Body/page backgrounds are handled globally by src/css/app.scss tokens. */

.body--dark .q-dialog__backdrop {
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(3px);
}

.body--light .q-dialog__backdrop {
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(3px);
}

/* The card carries Quasar's .bg-dark utility (which has !important), so the
   override needs !important to win. */
.dark-dialog .player-detail-dialog-card,
.q-card.player-detail-dialog-card.bg-dark {
    background: var(--surface-page) !important;
    color: var(--text-primary) !important;
    border: none;
    box-shadow: var(--shadow-3);
}

.light-dialog .player-detail-dialog-card {
    background: var(--surface-card) !important;
    color: var(--text-primary) !important;
}
</style>
