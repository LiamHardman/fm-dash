<template>
    <q-dialog 
        :model-value="show" 
        @hide="$emit('close')"
        :class="qInstance.dark.isActive ? 'dark-dialog' : 'light-dialog'"
        backdrop-filter="blur(3px)"
        :backdrop-color="qInstance.dark.isActive ? 'rgba(0, 0, 0, 0.8)' : 'rgba(0, 0, 0, 0.5)'"
        transition-show="scale"
        transition-hide="scale"
    >
        <q-card
            class="player-detail-dialog-card modern-dialog-card"
            :class="
                qInstance.dark.isActive
                    ? 'text-white'
                    : 'bg-white text-dark'
            "
            style="max-width: 1400px; width: 95vw; max-height: 90vh; position: relative;"
        >
            <!-- Floating Close Button -->
            <q-btn 
                dense 
                flat 
                icon="close" 
                @click="$emit('close')" 
                class="floating-close-btn"
                :class="qInstance.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
            >
                <q-tooltip
                    :class="
                        qInstance.dark.isActive
                            ? 'bg-grey-7'
                            : 'bg-white text-primary'
                    "
                    >Close</q-tooltip
                >
            </q-btn>

            <q-card-section v-if="player" class="scroll main-content-section no-header-section">
                <div class="row q-col-gutter-lg">
                    <div class="col-12 col-md-4">
                        <div class="row q-col-gutter-sm q-mb-md">
                            <div class="col-6">
                                <q-select
                                    v-if="performanceComparisonOptions.length > 0"
                                    :disable="performanceComparisonOptions.length <= 1"
                                    v-model="selectedComparisonGroup"
                                    :options="performanceComparisonOptions"
                                    label="Compare Position"
                                    dense
                                    outlined
                                    emit-value
                                    map-options
                                    class="modern-select"
                                    :label-color="
                                        qInstance.dark.isActive ? 'grey-4' : ''
                                    "
                                    :popup-content-class="
                                        qInstance.dark.isActive
                                            ? 'bg-grey-8 text-white'
                                            : 'bg-white text-dark'
                                    "
                                />
                                <q-tooltip
                                    v-if="
                                        performanceComparisonOptions.length <= 1 &&
                                        performanceComparisonOptions.length > 0
                                    "
                                >
                                    Only global comparison available for this player.
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
                                        qInstance.dark.isActive ? 'grey-4' : ''
                                    "
                                    :popup-content-class="
                                        qInstance.dark.isActive
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
                                <div v-else-if="hasAnyPerformanceData" class="percentile-content-area">
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

                    <div class="col-12 col-md-8">
                        <q-card
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
                                                    :color="qInstance.dark.isActive ? 'grey-7' : 'grey-4'"
                                                    :text-color="qInstance.dark.isActive ? 'grey-4' : 'grey-7'"
                                                    class="player-face-placeholder"
                                                >
                                                    <q-icon name="person" size="32px" />
                                                </q-avatar>
                                            </div>
                                            
                                            <div class="col-auto q-mr-md player-flag-container">
                                                <img
                                                    v-if="player.nationality_iso && !flagLoadError"
                                                    :src="`https://flagcdn.com/w80/${player.nationality_iso.toLowerCase()}.png`"
                                                    :alt="player.nationality || 'Flag'"
                                                    width="48"
                                                    height="32"
                                                    class="player-flag"
                                                    @error="handleFlagError"
                                                    :title="player.nationality"
                                                />
                                                <q-icon
                                                    v-if="!player.nationality_iso || flagLoadError"
                                                    :color="qInstance.dark.isActive ? 'grey-5' : 'grey-7'"
                                                    name="flag"
                                                    size="2.5em"
                                                    class="player-flag-placeholder"
                                                />
                                                
                                                <!-- Club logo below nationality flag -->
                                                <div class="q-mt-sm club-logo-container" v-if="player.club && player.club !== '-'">
                                                    <Suspense v-if="shouldShowTeamLogo">
                                                        <template #default>
                                                            <TeamLogo 
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
                                                        <h5
                                                            class="text-h5 player-name no-margin"
                                                            :class="
                                                                qInstance.dark.isActive ? 'text-white' : 'text-dark'
                                                            "
                                                            :title="player.name"
                                                        >
                                                            {{ player.name }}
                                                            <q-icon
                                                                v-if="player.attributeMasked"
                                                                name="warning"
                                                                color="warning"
                                                                size="sm"
                                                                class="q-ml-sm scouting-warning-icon"
                                                            >
                                                                <q-tooltip
                                                                    :class="
                                                                        qInstance.dark.isActive
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
                                                        </h5>
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
                                                            :label="`${player.age || '-'} years`"
                                                            class="player-age-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            outline
                                                            color="secondary"
                                                            :label="player.nationality || 'Unknown'"
                                                            class="player-nationality-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            v-if="player.club"
                                                            outline
                                                            color="teal"
                                                            :label="player.club"
                                                            class="player-club-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            v-if="player.personality"
                                                            outline
                                                            color="purple"
                                                            :label="player.personality"
                                                            class="player-personality-badge q-mr-sm"
                                                        />
                                                        <q-badge
                                                            v-if="player.media_handling"
                                                            outline
                                                            color="orange"
                                                            :label="player.media_handling"
                                                            class="player-media-badge"
                                                        />
                                                    </div>
                                                </div>
                                                
                                                <div class="player-positions-section q-mt-sm" v-if="player.shortPositions?.length || player.position">
                                                    <q-badge
                                                        v-for="pos in player.shortPositions || [player.position]"
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

                                <div class="fifa-stats-section">
                                    
                                    <div class="fifa-stats-grid">
                                        <div
                                            v-for="stat in fifaStatsToDisplay"
                                            :key="`fifa-${stat.name}-${player?.UID || player?.uid}`"
                                            class="fifa-stat-card"
                                        >
                                            <q-card
                                                flat
                                                bordered
                                                :class="[
                                                    'fifa-stat-item text-center',
                                                    getUnifiedRatingClass(player[stat.name], 100),
                                                ]"
                                            >
                                                <div class="fifa-stat-label">{{ stat.label }}</div>
                                                <div class="fifa-stat-value">
                                                    {{ player[stat.name] !== undefined ? player[stat.name] : "-" }}
                                                </div>
                                                <q-tooltip
                                                    :class="
                                                        qInstance.dark.isActive
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
                                            >
                                                <q-item-section>
                                                    <q-item-label lines="1" class="attribute-name">
                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                    </q-item-label>
                                                    <q-tooltip
                                                        :class="
                                                            qInstance.dark.isActive
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
                                                            getUnifiedRatingClass(player.attributes[attrKey], 20)
                                                        ]"
                                                    >
                                                        {{ getDisplayAttribute(attrKey) }}
                                                    </span>
                                                </q-item-section>
                                            </q-item>
                                            
                                            <q-item
                                                v-if="!(isGoalkeeper ? attributeCategories.goalkeeping : attributeCategories.technical).length"
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
                                            >
                                                <q-item-section>
                                                    <q-item-label lines="1" class="attribute-name">
                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                    </q-item-label>
                                                    <q-tooltip
                                                        :class="
                                                            qInstance.dark.isActive
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
                                                            getUnifiedRatingClass(player.attributes[attrKey], 20)
                                                        ]"
                                                    >
                                                        {{ getDisplayAttribute(attrKey) }}
                                                    </span>
                                                </q-item-section>
                                            </q-item>
                                            
                                            <q-item v-if="!attributeCategories.mental.length" class="no-attributes-item">
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
                                                    >
                                                        <q-item-section>
                                                            <q-item-label lines="1" class="attribute-name">
                                                                {{ attributeFullNameMap[attrKey] || attrKey }}
                                                            </q-item-label>
                                                            <q-tooltip
                                                                :class="
                                                                    qInstance.dark.isActive
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
                                                                    getUnifiedRatingClass(player.attributes[attrKey], 20)
                                                                ]"
                                                            >
                                                                {{ getDisplayAttribute(attrKey) }}
                                                            </span>
                                                        </q-item-section>
                                                    </q-item>
                                                    
                                                    <q-item v-if="!attributeCategories.physical.length" class="no-attributes-item">
                                                        <q-item-section class="text-center q-py-md">
                                                            <q-icon name="info_outline" size="sm" class="q-mr-sm" />
                                                            No physical attributes.
                                                        </q-item-section>
                                                    </q-item>
                                                </q-list>
                                            </q-card-section>
                                        </q-card>
                                    </div>
                                    
                                    <div class="col-12" v-if="player.roleSpecificOveralls && player.roleSpecificOveralls.length > 0">
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
                                                            'best-role-highlight': roleOverall.score === player.Overall,
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
            </q-card-section>

            <q-card-section v-else class="loading-section">
                <div class="loading-content">
                    <q-spinner color="primary" size="3em" />
                    <div class="loading-text">Loading player data...</div>
                </div>
            </q-card-section>
        </q-card>
    </q-dialog>
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
  watch
} from 'vue'
import { usePercentileRetry } from '../composables/usePercentileRetry'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
import { formatCurrency } from '../utils/currencyUtils'

// Lazy load TeamLogo component to prevent blocking dialog opening
const TeamLogo = defineAsyncComponent(() => import('../components/TeamLogo.vue'))

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
  Thr: 'Throwing'
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
  Thr: 'Accuracy and effectiveness when throwing the ball to teammates'
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
  POS: 'Positioning and decision-making when coming off the goal line'
}

const fifaToFmAttributeMapping = {
  // Outfield Player FIFA Stats mapped to FM attributes (based on weights)
  PAC: {
    primary: ['Acc', 'Pac'],
    secondary: ['Agi'],
    description: 'Based on Acceleration, Pace, and Agility'
  },
  SHO: {
    primary: ['Fin', 'Lon'],
    secondary: ['Pen', 'Hea', 'Cmp', 'Tec', 'Ant', 'Dec', 'Fla'],
    description:
      'Based on Finishing, Long Shots, Penalties, Heading, Composure, Technique, Anticipation, Decisions, and Flair'
  },
  PAS: {
    primary: ['Pas', 'Vis'],
    secondary: ['Cro', 'Tec', 'Fre', 'Tea', 'Dec', 'Fir', 'Cor', 'OtB'],
    description:
      'Based on Passing, Vision, Crossing, Technique, Free Kicks, Teamwork, Decisions, First Touch, Corners, and Off the Ball'
  },
  DRI: {
    primary: ['Dri', 'Fir', 'Tec'],
    secondary: ['Fla', 'Cmp', 'OtB'],
    description: 'Based on Dribbling, First Touch, Technique, Flair, Composure, and Off the Ball'
  },
  DEF: {
    primary: ['Mar', 'Tck', 'Ant', 'Pos'],
    secondary: ['Hea', 'Cnt', 'Dec', 'Cmp', 'Bra', 'Agg', 'Wor'],
    description:
      'Based on Marking, Tackling, Anticipation, Positioning, Heading, Concentration, Decisions, Composure, Bravery, Aggression, and Work Rate'
  },
  PHY: {
    primary: ['Str', 'Sta', 'Nat'],
    secondary: ['Jum', 'Agg', 'Bra', 'Wor', 'Bal'],
    description:
      'Based on Strength, Stamina, Natural Fitness, Jumping Reach, Aggression, Bravery, Work Rate, and Balance'
  },

  // Goalkeeper FIFA Stats mapped to FM attributes
  DIV: {
    primary: ['Aer', 'Ref', '1v1'],
    secondary: ['Agi', 'Han'],
    description: 'Based on Aerial Reach, Reflexes, One on Ones, Agility, and Handling'
  },
  HAN: {
    primary: ['Han', 'Cmd'],
    secondary: ['Cmp', 'Cnt'],
    description: 'Based on Handling, Command of Area, Composure, and Concentration'
  },
  REF: {
    primary: ['Ref', 'Ant', '1v1'],
    secondary: ['Cnt'],
    description: 'Based on Reflexes, Anticipation, One on Ones, and Concentration'
  },
  KIC: {
    primary: ['Kic', 'Thr'],
    secondary: ['Tec', 'Vis', 'Pas'],
    description: 'Based on Kicking, Throwing, Technique, Vision, and Passing'
  },
  SPD: {
    primary: ['Acc', 'Pac', 'TRO'],
    secondary: [],
    description: 'Based on Acceleration, Pace, and Rushing Out Tendency'
  },
  POS: {
    primary: ['Pos', 'Cmd', 'Ant', 'Dec'],
    secondary: ['TRO', 'Cnt', 'Com'],
    description:
      'Based on Positioning, Command of Area, Anticipation, Decisions, Rushing Out Tendency, Concentration, and Communication'
  }
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
  'Tec'
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
  'Wor'
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
  'Thr'
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
  'Sv %': 'Save Percentage'
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
    'Poss Lost/90'
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
    'FA'
  ],
  Goalkeeping: ['Con/90', 'Cln/90', 'xGP/90', 'Sv %']
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
  Strikers: ['ST']
}

export default defineComponent({
  name: 'PlayerDetailDialog',
  components: {
    TeamLogo
  },
  props: {
    player: { type: Object, default: () => null },
    show: { type: Boolean, default: false },
    currencySymbol: { type: String, default: '$' },
    datasetId: { type: String, default: null }
  },
  emits: ['close'],
  setup(props) {
    const qInstance = useQuasar()
    const uiStore = useUiStore()
    const _playerStore = usePlayerStore()
    const selectedComparisonGroup = ref('Global')
    const flagLoadError = ref(false)
    const divisionFilter = ref('all')

    // Convert props to refs for the percentile retry composable
    const playerRef = toRef(props, 'player')
    const datasetIdRef = toRef(props, 'datasetId')

    // Use the percentile retry composable
    const {
      isLoadingPercentiles,
      hasValidPercentiles,
      percentilesNeedRetry,
      showLoadingState,
      percentilesRetryCount,
      maxRetries,
      manualRetry
    } = usePercentileRetry(playerRef, datasetIdRef, selectedComparisonGroup)

    // Face image handling
    const faceImageLoadError = ref(false)
    const shouldShowTeamLogo = ref(false)

    const handleFlagError = () => {
      flagLoadError.value = true
    }

    const handleFaceImageError = () => {
      faceImageLoadError.value = true
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
      return `/api/faces?uid=${encodeURIComponent(playerUID)}`
    })

    // Reset face image error when player changes
    watch(
      () => props.player,
      _newPlayer => {
        faceImageLoadError.value = false
      },
      { immediate: true }
    )

    // Cleanup all caches
    const clearAllCaches = () => {
      performanceStatsCache.clear()
      performanceComparisonOptionsCache.clear()
      currencyCache.clear()
      attributeDisplayCache.clear()
    }

    onMounted(() => {
      /* Initialization logic if needed */
    })

    onUnmounted(() => {
      clearAllCaches()
    })

    // Memoized performance comparison options with better caching
    const performanceComparisonOptionsCache = new Map()

    const performanceComparisonOptions = computed(() => {
      if (!props.player?.performancePercentiles) {
        return props.player?.performancePercentiles?.Global
          ? [{ label: 'Overall Dataset', value: 'Global' }]
          : []
      }

      const player = props.player
      const cacheKey = `${getCacheKey(player, 'options')}-${JSON.stringify(player.shortPositions)}-${JSON.stringify(player.positionGroups)}`

      if (performanceComparisonOptionsCache.has(cacheKey)) {
        return performanceComparisonOptionsCache.get(cacheKey)
      }

      const playerPercentiles = player.performancePercentiles
      const playerShortPositions = player.shortPositions || []
      const playerBroadGroups = player.positionGroups || []
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
        'Strikers'
      ]

      const shouldIncludeGroup = groupKey => {
        if (groupKey === 'Global') return true

        // Check broad groups
        if (broadGroupsSet.has(groupKey)) return true

        // Check detailed groups
        const requiredPositions = detailedGroupToShortPositionsMap[groupKey]
        if (requiredPositions) {
          return requiredPositions.some(pos => shortPositionsSet.has(pos))
        }

        return false
      }

      // Process preferred order groups
      for (let i = 0; i < preferredOrder.length; i++) {
        const groupKey = preferredOrder[i]
        if (
          playerPercentiles[groupKey] &&
          shouldIncludeGroup(groupKey) &&
          !addedValues.has(groupKey)
        ) {
          options.push({
            label: groupKey === 'Global' ? 'Overall Dataset' : `vs. ${groupKey}`,
            value: groupKey
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
      if (!props.player?.roleSpecificOveralls?.length) {
        return null
      }

      // Find the role with the highest overall rating
      const highestRole = props.player.roleSpecificOveralls.reduce((max, role) =>
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
        ST: 'Attackers'
      }

      return positionToGroupMap[shortPosition] || null
    }

    // Optimized player watcher with cache cleanup
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

        if (!newPlayer?.performancePercentiles) {
          selectedComparisonGroup.value = 'Global'
          return
        }

        const newOptions = performanceComparisonOptions.value
        const highestRoleGroup = getPositionGroupForHighestRole()

        // Priority-based selection logic
        if (highestRoleGroup && newOptions.some(opt => opt.value === highestRoleGroup)) {
          selectedComparisonGroup.value = highestRoleGroup
        } else if (newOptions.some(opt => opt.value === 'Global')) {
          selectedComparisonGroup.value = 'Global'
        } else if (newOptions.length > 0) {
          selectedComparisonGroup.value = newOptions[0].value
        } else {
          selectedComparisonGroup.value = 'Global'
        }
      },
      { immediate: true }
    )

    // Watch dialog visibility to delay team logo loading
    watch(
      () => props.show,
      isShowing => {
        if (isShowing) {
          // Delay team logo rendering until dialog is fully opened
          setTimeout(() => {
            shouldShowTeamLogo.value = true
          }, 50) // Small delay to ensure smooth dialog opening
        } else {
          shouldShowTeamLogo.value = false
        }
      },
      { immediate: true }
    )

    const divisionFilterOptions = computed(() => [
      { label: 'All', value: 'all' },
      { label: 'Same', value: 'same' },
      { label: 'Top 5', value: 'top5' }
    ])

    const getTargetDivision = () => {
      if (!props.player?.division) return null
      return props.player.division
    }

    const onDivisionFilterChange = async () => {
      if (!props.datasetId || !props.player) return

      try {
        const targetDivision = getTargetDivision()
        const requestPayload = {
          playerName: props.player.name,
          divisionFilter: divisionFilter.value,
          targetDivision: targetDivision
        }

        // Instead of refetching all data, we need to fetch just updated percentiles for this player
        // For now, let's create a dedicated API call for this
        const url = `/api/percentiles/${props.datasetId}`

        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(requestPayload)
        })

        if (response.ok) {
          const updatedPercentiles = await response.json()

          // Count non-empty percentile groups for debugging
          let _nonEmptyGroups = 0
          let _totalStats = 0
          for (const [_groupKey, groupPercentiles] of Object.entries(updatedPercentiles)) {
            const statsInGroup = Object.keys(groupPercentiles).length
            const nonNegativeStats = Object.values(groupPercentiles).filter(val => val >= 0).length
            if (nonNegativeStats > 0) {
              _nonEmptyGroups++
            }
            _totalStats += statsInGroup
          }

          // Log whether this is a cache hit or miss
          const _cacheStatus = response.headers.get('X-Cache-Status')

          // Update the player's percentiles without affecting the main dataset
          if (props.player.performancePercentiles) {
            Object.assign(props.player.performancePercentiles, updatedPercentiles)

            // Clear relevant caches to force recomputation
            performanceStatsCache.clear()
            performanceComparisonOptionsCache.clear()
          }
        } else {
          const _errorText = await response.text()
        }
      } catch (_error) {}
    }

    const isGoalkeeper = computed(() => {
      if (!props.player) return false
      return (
        props.player.shortPositions?.includes('GK') ||
        props.player.positionGroups?.includes('Goalkeepers') ||
        props.player.parsedPositions?.includes('Goalkeeper')
      )
    })

    // Memoized attribute filtering for better performance
    const attributeCategories = computed(() => {
      if (!props.player?.attributes) {
        return {
          technical: [],
          mental: [],
          physical: [],
          goalkeeping: []
        }
      }

      const playerAttributes = props.player.attributes
      const hasAttribute = key => Object.hasOwn(playerAttributes, key)

      return {
        technical: technicalAttrsOrdered.filter(hasAttribute),
        mental: mentalAttrsOrdered.filter(hasAttribute),
        physical: physicalAttrsOrdered.filter(hasAttribute),
        goalkeeping: isGoalkeeper.value ? goalkeepingAttrsOrdered.filter(hasAttribute) : []
      }
    })

    // Pre-defined stat configurations for better performance
    const goalkeepingStats = [
      { name: 'div', label: 'DIV' },
      { name: 'han', label: 'HAN' },
      { name: 'ref', label: 'REF' },
      { name: 'kic', label: 'KIC' },
      { name: 'spd', label: 'SPD' },
      { name: 'pos', label: 'POS' }
    ]

    const outfieldStats = [
      { name: 'pac', label: 'PAC' },
      { name: 'sho', label: 'SHO' },
      { name: 'pas', label: 'PAS' },
      { name: 'dri', label: 'DRI' },
      { name: 'def', label: 'DEF' },
      { name: 'phy', label: 'PHY' }
    ]

    const fifaStatsToDisplay = computed(() => {
      if (!props.player) return []

      const statsTemplate = isGoalkeeper.value ? goalkeepingStats : outfieldStats
      return statsTemplate.filter(stat => props.player[stat.name] !== undefined)
    })

    const _averageRatingData = computed(() => {
      if (!props.player || !props.player.attributes || !props.player.performancePercentiles)
        return null
      const groupKey = selectedComparisonGroup.value
      const percentilesForGroup = props.player.performancePercentiles[groupKey]
      if (
        !percentilesForGroup ||
        !Object.hasOwn(props.player.attributes, 'Av Rat') ||
        props.player.attributes['Av Rat'] === '-' ||
        props.player.attributes['Av Rat'] === '' ||
        !Object.hasOwn(percentilesForGroup, 'Av Rat')
      ) {
        return null
      }
      return {
        key: 'Av Rat',
        name: performanceStatMap['Av Rat'] || 'Average Rating',
        value: props.player.attributes['Av Rat'],
        percentile: percentilesForGroup['Av Rat'] >= 0 ? percentilesForGroup['Av Rat'] : null
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
        const percentileValue = percentilesForGroup[statKey]

        if (
          performanceStatMap[statKey] &&
          rawAttributeValue !== undefined &&
          rawAttributeValue !== '-' &&
          rawAttributeValue !== '' &&
          percentileValue !== undefined
        ) {
          statsInCategory.push({
            key: statKey,
            name: performanceStatMap[statKey],
            value: rawAttributeValue,
            percentile: percentileValue >= 0 ? percentileValue : null
          })
        }
      }

      if (statsInCategory.length === 0) return []

      // Optimize sorting for General category
      if (categoryName === 'General') {
        const avgRatingIndex = statsInCategory.findIndex(stat => stat.key === 'Av Rat')
        if (avgRatingIndex > -1) {
          const avgRatingStat = statsInCategory.splice(avgRatingIndex, 1)[0]
          statsInCategory.sort((a, b) => a.name.localeCompare(b.name))
          return [avgRatingStat, ...statsInCategory]
        }
      }

      return statsInCategory.sort((a, b) => a.name.localeCompare(b.name))
    }

    const categorizedPerformanceStats = computed(() => {
      if (!props.player?.attributes || !props.player?.performancePercentiles) {
        return {}
      }

      const groupKey = selectedComparisonGroup.value
      const percentilesForGroup = props.player.performancePercentiles[groupKey]
      if (!percentilesForGroup) return {}

      const cacheKey = getCacheKey(props.player, groupKey)

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
          percentilesForGroup,
          props.player.attributes
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

    const hasAnyPerformanceData = computed(
      () => Object.keys(categorizedPerformanceStats.value).length > 0
    )

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

    const getBarFillStyle = percentile => {
      if (percentile === null || percentile === undefined || percentile < 0) {
        return {
          width: '0%',
          backgroundColor: '#9e9e9e',
          height: '12px',
          borderRadius: '3px'
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
        transition: 'width 0.3s ease, background-color 0.3s ease'
      }
    }

    // Memoized role sorting with shallow comparison optimization
    let lastRoleOveralls = null
    let lastSortedRoles = []

    const sortedRoleSpecificOveralls = computed(() => {
      const roleOveralls = props.player?.roleSpecificOveralls
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

    // Optimized attribute display with memoization
    const attributeDisplayCache = new Map()

    const getDisplayAttribute = attrKey => {
      if (!props.player) return '-'

      const cacheKey = `${attrKey}-${props.player.UID || props.player.uid}-${showAttributeMasks.value}`
      if (attributeDisplayCache.has(cacheKey)) {
        return attributeDisplayCache.get(cacheKey)
      }

      const rawValue = props.player.attributes?.[attrKey]
      if (rawValue === undefined) return '-'

      let displayValue
      if (rawValue === '-') {
        displayValue = '?'
      } else if (showAttributeMasks.value && String(rawValue).includes('-')) {
        displayValue = rawValue
      } else {
        const numericValue = props.player.numericAttributes?.[attrKey]
        displayValue = numericValue !== undefined ? numericValue : rawValue
      }

      // Cache management
      if (attributeDisplayCache.size > 200) {
        attributeDisplayCache.clear()
      }

      attributeDisplayCache.set(cacheKey, displayValue)
      return displayValue
    }

    return {
      qInstance,
      attributeCategories,
      attributeFullNameMap,
      attributeDescriptions,
      fifaAttributeDescriptions,
      fifaToFmAttributeMapping,
      getUnifiedRatingClass,
      getBarFillStyle,
      fifaStatsToDisplay,
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

      // Percentile retry functionality
      isLoadingPercentiles,
      hasValidPercentiles,
      percentilesNeedRetry,
      showLoadingState,
      percentilesRetryCount,
      maxRetries,
      manualRetry
    }
  }
})
</script>

<style lang="scss" scoped>
@use "sass:color";

// Import Quasar SCSS variables
$grey-1: #f5f5f5 !default;
$grey-2: #eeeeee !default;
$grey-3: #e0e0e0 !default;
$grey-4: #bdbdbd !default;
$grey-5: #9e9e9e !default;
$grey-6: #757575 !default;
$grey-7: #616161 !default;
$grey-8: #424242 !default;
$grey-9: #303030 !default;
$grey-10: #212121 !default;
$positive: #21ba45 !default;
$primary: #1976d2 !default;
$indigo-5: #3f51b5 !default;
$breakpoint-sm-max: 1023px !default;
$breakpoint-xs-max: 599px !default;

// Modern Dialog Card
.player-detail-dialog-card {
    display: flex;
    flex-direction: column;
    border-radius: 16px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
    overflow: hidden;
    background: white;
    
    .body--dark & {
        background: #1e293b;
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    }
}

// Dark mode dialog backdrop fix
:deep(.dark-dialog .q-dialog__backdrop) {
    background: rgba(0, 0, 0, 0.6) !important;
}

:deep(.light-dialog .q-dialog__backdrop) {
    background: rgba(0, 0, 0, 0.4) !important;
}

// Additional dark mode dialog fixes
.body--dark {
    :deep(.q-dialog__backdrop) {
        background: rgba(0, 0, 0, 0.6) !important;
    }
}

.body--light {
    :deep(.q-dialog__backdrop) {
        background: rgba(0, 0, 0, 0.4) !important;
    }
}

.main-content-section {
    flex-grow: 1;
    padding: 20px;
    background: transparent;
    
    .body--dark & {
        background: transparent;
    }
    
    &.no-header-section {
        padding-top: 20px; // Minimal room for floating close button
    }
}

// Modern Select Styling
.modern-select {
    :deep(.q-field__control) {
        border-radius: 8px;
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
        }
        
        .body--light & {
            background: rgba(0, 0, 0, 0.02);
        }
    }
}

// Performance Card
.performance-percentiles-card {
    border-radius: 12px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    border: 1px solid rgba(0, 0, 0, 0.05);
    background: white;
    
    .body--dark & {
        background: #1e293b;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.25);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }
}

.performance-card-header {
    background: linear-gradient(135deg, rgba(25, 118, 210, 0.1) 0%, rgba(25, 118, 210, 0.05) 100%);
    border-radius: 12px 12px 0 0;
    padding: 16px 20px;
    
    .body--dark & {
        background: linear-gradient(135deg, rgba(144, 202, 249, 0.1) 0%, rgba(144, 202, 249, 0.05) 100%);
    }
}

.performance-header-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: #1976d2;
    display: flex;
    align-items: center;
    
    .body--dark & {
        color: #90caf9;
    }
}

.performance-category-header {
    margin-bottom: 8px;
}

.performance-category-title {
    font-size: 0.9rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.performance-separator {
    background: rgba(25, 118, 210, 0.15);
    height: 1px;
    border: none;
    
    .body--dark & {
        background: rgba(144, 202, 249, 0.15);
    }
}

.modern-stat-item {
    transition: background-color 0.2s ease;
    border-radius: 6px;
    margin: 2px 0;
    padding: 8px 12px;
    
    &:hover {
        background: rgba(25, 118, 210, 0.05);
        
        .body--dark & {
            background: rgba(144, 202, 249, 0.05);
        }
    }
    
    &.average-rating-item {
        background: rgba(25, 118, 210, 0.08);
        border-left: 4px solid #1976d2;
        
        .body--dark & {
            background: rgba(144, 202, 249, 0.08);
            border-left-color: #90caf9;
        }
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
    color: #334155;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.85);
    }
}

.stat-bar-container {
    display: flex;
    align-items: center;
    width: 100%;
}

.stat-bar-track {
    flex-grow: 1;
    height: 10px;
    background-color: #e5e7eb;
    border-radius: 5px;
    margin-right: 8px;
    overflow: hidden;
    
    .body--dark & {
        background-color: #374151;
    }
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
    color: #64748b;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.6);
    }
}

.performance-stat-value {
    font-size: 0.8rem;
    font-weight: 600;
    color: #334155;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.85);
    }
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
    color: #64748b;
    font-size: 0.9rem;
    gap: 0.5rem;
    text-align: center;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.6);
    }
}

// Profile Card
.player-profile-card {
    border-radius: 12px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    border: 1px solid rgba(0, 0, 0, 0.05);
    background: white;
    
    .body--dark & {
        background: #1e293b;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.25);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }
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
    border: 3px solid rgba(25, 118, 210, 0.2);
    object-fit: cover;
    transition: all 0.3s ease;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    
    &:hover {
        transform: scale(1.05);
        border-color: rgba(25, 118, 210, 0.4);
        box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
    }
    
    .body--dark & {
        border-color: rgba(144, 202, 249, 0.2);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
        
        &:hover {
            border-color: rgba(144, 202, 249, 0.4);
            box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
        }
    }
}

.player-face-placeholder {
    border: 3px solid rgba(25, 118, 210, 0.2);
    transition: all 0.3s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    
    &:hover {
        transform: scale(1.05);
        border-color: rgba(25, 118, 210, 0.4);
    }
    
    .body--dark & {
        border-color: rgba(144, 202, 249, 0.2);
        
        &:hover {
            border-color: rgba(144, 202, 249, 0.4);
        }
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
    border: 2px solid rgba(128, 128, 128, 0.3);
    border-radius: 4px;
    object-fit: cover;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
    transition: all 0.2s ease;
    
    &:hover {
        transform: scale(1.05);
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
    }
    
    .body--dark & {
        border-color: rgba(255, 255, 255, 0.2);
        box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
    }
}

.player-flag-placeholder {
    color: #9ca3af;
    
    .body--dark & {
        color: #6b7280;
    }
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
    color: #1e293b;
    margin-bottom: 8px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.95);
    }
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
    background: rgba(25, 118, 210, 0.03);
    border-radius: 8px;
    border: 1px solid rgba(25, 118, 210, 0.1);
    transition: all 0.2s ease;
    flex: 1;
    min-width: 0;
    
    &:hover {
        background: rgba(25, 118, 210, 0.06);
        border-color: rgba(25, 118, 210, 0.2);
        transform: translateY(-1px);
    }
    
    .body--dark & {
        background: rgba(144, 202, 249, 0.05);
        border-color: rgba(144, 202, 249, 0.1);
        
        &:hover {
            background: rgba(144, 202, 249, 0.08);
            border-color: rgba(144, 202, 249, 0.2);
        }
    }
}

.detail-icon {
    color: #1976d2;
    
    .body--dark & {
        color: #90caf9;
    }
}

.detail-content {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    overflow: hidden;
}

.detail-label {
    font-size: 0.7rem;
    color: #64748b;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 2px;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.6);
    }
}

.detail-value {
    font-size: 0.9rem;
    font-weight: 600;
    color: #334155;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.85);
    }
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
    color: #64748b;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 2px;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.6);
    }
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
    background: linear-gradient(90deg, transparent 0%, rgba(25, 118, 210, 0.3) 50%, transparent 100%);
    height: 2px;
    border: none;
    
    .body--dark & {
        background: linear-gradient(90deg, transparent 0%, rgba(144, 202, 249, 0.3) 50%, transparent 100%);
    }
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
    color: #1976d2;
    
    .body--dark & {
        color: #90caf9;
    }
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
    border-radius: 12px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
    border: 1px solid rgba(0, 0, 0, 0.05);
    background: white;
    transition: all 0.3s ease;
    
    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
    }
    
    .body--dark & {
        background: #1e293b;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
        border: 1px solid rgba(255, 255, 255, 0.1);
        
        &:hover {
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
        }
    }
}

.full-height-card {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.attribute-card-header {
    background: linear-gradient(135deg, rgba(25, 118, 210, 0.1) 0%, rgba(25, 118, 210, 0.05) 100%);
    border-radius: 12px 12px 0 0;
    padding: 16px 20px;
    
    .body--dark & {
        background: linear-gradient(135deg, rgba(144, 202, 249, 0.1) 0%, rgba(144, 202, 249, 0.05) 100%);
    }
}

.attribute-section-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: #1976d2;
    display: flex;
    align-items: center;
    
    .body--dark & {
        color: #90caf9;
    }
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
        background: rgba(25, 118, 210, 0.05);
        
        .body--dark & {
            background: rgba(144, 202, 249, 0.05);
        }
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
    color: #334155;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.85);
    }
    
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
    color: #64748b;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.6);
    }
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
            background: rgba(25, 118, 210, 0.3);
            border-radius: 2px;
            
            &:hover {
                background: rgba(25, 118, 210, 0.5);
            }
        }
    }
}

// Modern Tooltips
.modern-tooltip {
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    
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
    color: #64748b;
    font-size: 1rem;
    font-weight: 500;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

// Responsive Design
@media (max-width: 768px) {
    .main-content-section {
        padding: 16px;
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

// Floating Close Button
.floating-close-btn {
    position: absolute;
    top: 16px;
    right: 16px;
    z-index: 10;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    background: rgba(255, 255, 255, 0.9);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    transition: all 0.2s ease;
    
    &:hover {
        background: rgba(255, 255, 255, 1);
        transform: scale(1.1);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    }
    
    .body--dark & {
        background: rgba(30, 41, 59, 0.9);
        color: rgba(255, 255, 255, 0.8);
        
        &:hover {
            background: rgba(30, 41, 59, 1);
            color: rgba(255, 255, 255, 1);
        }
    }
}

.club-logo-container {
    display: flex;
    align-items: center;
    justify-content: center;
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
</style>

<style lang="scss">
/* Global styles for dialog backdrop - unscoped to override Quasar defaults */

/* Ensure body background is preserved in dark mode */
html.body--dark,
.body--dark,
body.body--dark {
    background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%) !important;
}

/* Dialog backdrop styles */
html.body--dark .q-dialog__backdrop,
.body--dark .q-dialog__backdrop,
body.body--dark .q-dialog__backdrop {
    background-color: rgba(0, 0, 0, 0.8) !important;
    backdrop-filter: blur(3px) !important;
}

html.body--light .q-dialog__backdrop,
.body--light .q-dialog__backdrop,
body.body--light .q-dialog__backdrop {
    background-color: rgba(0, 0, 0, 0.5) !important;
    backdrop-filter: blur(3px) !important;
}

/* Target the dialog inner container that might be causing white background */
.body--dark .q-dialog__inner,
.dark .q-dialog__inner,
body.body--dark .q-dialog__inner {
    background: transparent !important;
}

/* Target potential white background sources */
.body--dark .q-dialog,
.dark .q-dialog,
body.body--dark .q-dialog {
    background: transparent !important;
}

/* Additional specificity for the backdrop element */
.body--dark .q-dialog .q-dialog__backdrop,
.dark .q-dialog .q-dialog__backdrop,
body.body--dark .q-dialog .q-dialog__backdrop {
    background: rgba(0, 0, 0, 0.8) !important;
    backdrop-filter: blur(3px) !important;
}

/* Prevent body scroll changes when dialog is open */
.body--dark.q-body--dialog {
    background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%) !important;
}

/* Target any potential page container that might be getting a white background */
.body--dark .q-page-container,
.body--dark .q-page,
body.body--dark .q-page-container,
body.body--dark .q-page {
    background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%) !important;
}
</style>

