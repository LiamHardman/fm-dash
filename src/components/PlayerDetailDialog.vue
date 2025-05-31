<template>
    <q-dialog :model-value="show" @hide="$emit('close')">
        <q-card
            class="player-detail-dialog-card modern-dialog-card"
            :class="
                qInstance.dark.isActive
                    ? 'bg-dark text-white'
                    : 'bg-white text-dark'
            "
            style="max-width: 1300px; width: 95vw; max-height: 90vh"
        >
            <q-bar
                class="modern-dialog-header"
                :class="
                    qInstance.dark.isActive
                        ? 'bg-grey-10'
                        : 'bg-primary text-white'
                "
            >
                <div class="dialog-header-content">
                    <q-icon name="person" class="q-mr-sm header-icon" />
                    <div class="header-text">
                        <div class="text-subtitle1 dialog-title">
                            {{ player?.name || "Player" }}
                        </div>
                        <div class="dialog-subtitle">Detailed Analysis</div>
                    </div>
                </div>
                <q-space />
                <q-btn dense flat icon="close" @click="$emit('close')" class="close-btn">
                    <q-tooltip
                        :class="
                            qInstance.dark.isActive
                                ? 'bg-grey-7'
                                : 'bg-white text-primary'
                        "
                        >Close</q-tooltip
                    >
                </q-btn>
            </q-bar>

            <q-card-section v-if="player" class="scroll main-content-section">
                <div class="row q-col-gutter-md">
                    <div class="col-12 col-md-4">
                        <div class="row q-col-gutter-xs q-mb-md">
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
                            :class="[
                                qInstance.dark.isActive
                                    ? 'bg-grey-9'
                                    : 'bg-grey-1',
                                'rounded-borders performance-percentiles-card-left modern-stats-card',
                            ]"
                        >
                            <div
                                v-if="hasAnyPerformanceData"
                                class="percentile-content-area"
                            >
                                <div v-if="averageRatingData" class="q-mb-sm">
                                    <q-item
                                        class="attribute-list-item performance-stat-item"
                                    >
                                        <q-item-section
                                            class="stat-name-section"
                                        >
                                            <q-item-label
                                                lines="1"
                                                class="attribute-name-label"
                                                :title="averageRatingData.name"
                                            >
                                                {{ averageRatingData.name }}
                                            </q-item-label>
                                        </q-item-section>
                                        <q-item-section
                                            class="stat-bar-section"
                                        >
                                            <div class="stat-bar-container">
                                                <div class="stat-bar-track">
                                                    <div
                                                        class="stat-bar-fill"
                                                        :style="
                                                            getBarFillStyle(
                                                                averageRatingData.percentile,
                                                            )
                                                        "
                                                    ></div>
                                                </div>
                                                <span
                                                    v-if="
                                                        averageRatingData.percentile !==
                                                            null &&
                                                        averageRatingData.percentile >=
                                                            0
                                                    "
                                                    class="stat-percentile-text"
                                                >
                                                    {{
                                                        Math.round(
                                                            averageRatingData.percentile,
                                                        )
                                                    }}
                                                </span>
                                                <span
                                                    v-else
                                                    class="stat-percentile-text text-caption text-grey-6"
                                                    >N/A</span
                                                >
                                            </div>
                                        </q-item-section>
                                        <q-item-section
                                            side
                                            class="stat-value-section"
                                        >
                                            <span
                                                class="attribute-value performance-stat-actual-value"
                                            >
                                                {{
                                                    averageRatingData.value !==
                                                    "-"
                                                        ? averageRatingData.value
                                                        : "N/A"
                                                }}
                                            </span>
                                        </q-item-section>
                                    </q-item>
                                </div>

                                <hr
                                    v-if="
                                        averageRatingData &&
                                        Object.keys(categorizedPerformanceStats)
                                            .length > 0
                                    "
                                    class="q-my-sm"
                                />

                                <div
                                    v-for="(
                                        stats, category, index
                                    ) in categorizedPerformanceStats"
                                    :key="category"
                                >
                                    <hr
                                        v-if="
                                            index > 0 ||
                                            (averageRatingData && index === 0)
                                        "
                                        class="q-my-sm"
                                    />
                                    <q-list
                                        separator
                                        dense
                                        class="attribute-list performance-stats-list q-pt-xs"
                                    >
                                        <q-item
                                            v-for="statItem in stats"
                                            :key="statItem.key"
                                            class="attribute-list-item performance-stat-item"
                                        >
                                            <q-item-section
                                                class="stat-name-section"
                                            >
                                                <q-item-label
                                                    lines="1"
                                                    class="attribute-name-label"
                                                    :title="statItem.name"
                                                >
                                                    {{ statItem.name }}
                                                </q-item-label>
                                            </q-item-section>
                                            <q-item-section
                                                class="stat-bar-section"
                                            >
                                                <div class="stat-bar-container">
                                                    <div class="stat-bar-track">
                                                        <div
                                                            class="stat-bar-fill"
                                                            :style="
                                                                getBarFillStyle(
                                                                    statItem.percentile,
                                                                )
                                                            "
                                                        ></div>
                                                    </div>
                                                    <span
                                                        v-if="
                                                            statItem.percentile !==
                                                                null &&
                                                            statItem.percentile >=
                                                                0
                                                        "
                                                        class="stat-percentile-text"
                                                    >
                                                        {{
                                                            Math.round(
                                                                statItem.percentile,
                                                            )
                                                        }}
                                                    </span>
                                                    <span
                                                        v-else
                                                        class="stat-percentile-text text-caption text-grey-6"
                                                        >N/A</span
                                                    >
                                                </div>
                                            </q-item-section>
                                            <q-item-section
                                                side
                                                class="stat-value-section"
                                            >
                                                <span
                                                    class="attribute-value performance-stat-actual-value"
                                                >
                                                    {{
                                                        statItem.value !== "-"
                                                            ? statItem.value
                                                            : "N/A"
                                                    }}
                                                </span>
                                            </q-item-section>
                                        </q-item>
                                    </q-list>
                                </div>
                            </div>
                            <q-banner
                                v-else
                                class="q-mt-md text-center"
                                :class="
                                    qInstance.dark.isActive
                                        ? 'bg-grey-8 text-grey-5'
                                        : 'bg-grey-2 text-grey-7'
                                "
                            >
                                No performance data available for the selected
                                comparison group or this player.
                            </q-banner>
                        </q-card>
                    </div>

                    <div class="col-12 col-md-8">
                        <q-card
                            flat
                            bordered
                            :class="
                                qInstance.dark.isActive
                                    ? 'bg-grey-9'
                                    : 'bg-grey-1'
                            "
                            class="q-mb-md player-profile-card-redesigned modern-profile-card"
                        >
                            <q-card-section
                                class="q-pa-md player-profile-content"
                            >
                                <div class="profile-header-section">
                                    <div class="player-identity-extended">
                                        <div class="row items-center q-mb-xs">
                                            <!-- Player Face Image -->
                                            <div class="col-auto q-mr-sm player-face-container">
                                                <img
                                                    v-if="playerFaceImageUrl && !faceImageLoadError"
                                                    :src="playerFaceImageUrl"
                                                    :alt="`${player.name || 'Player'} face`"
                                                    width="60"
                                                    height="60"
                                                    class="player-face-image"
                                                    @error="handleFaceImageError"
                                                    @load="handleFaceImageLoad"
                                                />
                                                <q-avatar
                                                    v-else
                                                    size="60px"
                                                    :color="qInstance.dark.isActive ? 'grey-7' : 'grey-4'"
                                                    :text-color="qInstance.dark.isActive ? 'grey-4' : 'grey-7'"
                                                    class="player-face-placeholder"
                                                >
                                                    <q-icon name="person" size="24px" />
                                                </q-avatar>
                                            </div>
                                            <div
                                                class="col-auto q-mr-sm player-flag-container-redesigned"
                                            >
                                                <img
                                                    v-if="
                                                        player.nationality_iso &&
                                                        !flagLoadError
                                                    "
                                                    :src="`https://flagcdn.com/w40/${player.nationality_iso.toLowerCase()}.png`"
                                                    :alt="
                                                        player.nationality ||
                                                        'Flag'
                                                    "
                                                    width="36"
                                                    height="24"
                                                    class="player-flag-redesigned"
                                                    @error="handleFlagError"
                                                    :title="player.nationality"
                                                />
                                                <q-icon
                                                    v-if="
                                                        !player.nationality_iso ||
                                                        flagLoadError
                                                    "
                                                    :color="
                                                        qInstance.dark.isActive
                                                            ? 'grey-5'
                                                            : 'grey-7'
                                                    "
                                                    name="flag"
                                                    size="2em"
                                                    class="player-flag-placeholder-redesigned"
                                                />
                                            </div>
                                            <div
                                                class="col player-name-age-positions-redesigned"
                                            >
                                                <div
                                                    class="player-name-and-age"
                                                >
                                                    <div
                                                        class="text-h6 player-name-text-redesigned"
                                                        :title="
                                                            player.name || '-'
                                                        "
                                                    >
                                                        {{ player.name || "-" }}
                                                    </div>
                                                    <q-badge
                                                        outline
                                                        :color="
                                                            qInstance.dark
                                                                .isActive
                                                                ? 'blue-4'
                                                                : 'primary'
                                                        "
                                                        :label="`${player.age || '-'} yrs`"
                                                        class="player-age-badge-redesigned q-ml-sm"
                                                    />
                                                </div>
                                                <div
                                                    class="player-positions-inline q-mt-xs"
                                                    v-if="
                                                        player.shortPositions
                                                            ?.length ||
                                                        player.position
                                                    "
                                                >
                                                    <q-badge
                                                        v-for="pos in player.shortPositions || [
                                                            player.position,
                                                        ]"
                                                        :key="pos"
                                                        outline
                                                        :color="
                                                            qInstance.dark
                                                                .isActive
                                                                ? 'blue-3'
                                                                : 'indigo-5'
                                                        "
                                                        :label="pos"
                                                        class="q-mr-xs q-mb-xs player-position-badge"
                                                    />
                                                </div>
                                            </div>
                                        </div>
                                        <div
                                            class="player-additional-details q-mt-sm"
                                        >
                                            <div
                                                class="row q-col-gutter-x-sm q-col-gutter-y-xs items-center"
                                            >
                                                <div
                                                    class="col-12 col-sm-auto additional-detail-item"
                                                >
                                                    <q-icon
                                                        name="sports_soccer"
                                                        size="1.2em"
                                                        class="q-mr-xs additional-detail-icon"
                                                        :color="
                                                            qInstance.dark
                                                                .isActive
                                                                ? 'blue-4'
                                                                : 'primary'
                                                        "
                                                    />
                                                    <div
                                                        class="additional-detail-text-block"
                                                    >
                                                        <q-item-label
                                                            caption
                                                            class="additional-detail-caption"
                                                            >Club</q-item-label
                                                        >
                                                        <q-item-label
                                                            class="additional-detail-label ellipsis"
                                                            :title="
                                                                player.club ||
                                                                '-'
                                                            "
                                                            >{{
                                                                player.club ||
                                                                "-"
                                                            }}</q-item-label
                                                        >
                                                    </div>
                                                </div>
                                                <div
                                                    class="col-12 col-sm-auto additional-detail-item-separator gt-xs"
                                                >
                                                    &bull;
                                                </div>
                                                <div
                                                    class="col-12 col-sm-auto additional-detail-item"
                                                >
                                                    <q-icon
                                                        name="psychology"
                                                        size="1.2em"
                                                        class="q-mr-xs additional-detail-icon"
                                                        :color="
                                                            qInstance.dark
                                                                .isActive
                                                                ? 'blue-4'
                                                                : 'primary'
                                                        "
                                                    />
                                                    <div
                                                        class="additional-detail-text-block"
                                                    >
                                                        <q-item-label
                                                            caption
                                                            class="additional-detail-caption"
                                                            >Personality</q-item-label
                                                        >
                                                        <q-item-label
                                                            class="additional-detail-label ellipsis"
                                                            :title="
                                                                player.personality ||
                                                                '-'
                                                            "
                                                            >{{
                                                                player.personality ||
                                                                "-"
                                                            }}</q-item-label
                                                        >
                                                    </div>
                                                </div>
                                                <div
                                                    class="col-12 col-sm-auto additional-detail-item-separator gt-xs"
                                                >
                                                    &bull;
                                                </div>
                                                <div
                                                    class="col-12 col-sm-auto additional-detail-item"
                                                >
                                                    <q-icon
                                                        name="comment"
                                                        size="1.2em"
                                                        class="q-mr-xs additional-detail-icon"
                                                        :color="
                                                            qInstance.dark
                                                                .isActive
                                                                ? 'blue-4'
                                                                : 'primary'
                                                        "
                                                    />
                                                    <div
                                                        class="additional-detail-text-block"
                                                    >
                                                        <q-item-label
                                                            caption
                                                            class="additional-detail-caption"
                                                            >Media
                                                            Handling</q-item-label
                                                        >
                                                        <q-item-label
                                                            class="additional-detail-label ellipsis"
                                                            :title="
                                                                player.media_handling ||
                                                                '-'
                                                            "
                                                            >{{
                                                                player.media_handling ||
                                                                "-"
                                                            }}</q-item-label
                                                        >
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <div class="financial-details-top-right">
                                        <div
                                            class="financial-item-large"
                                            :title="formattedTransferValue"
                                        >
                                            <q-icon
                                                name="trending_up"
                                                class="q-mr-xs"
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'green-4'
                                                        : 'green-7'
                                                "
                                            />
                                            <span
                                                :class="
                                                    qInstance.dark.isActive
                                                        ? 'text-green-4'
                                                        : 'text-green-7'
                                                "
                                                >{{
                                                    formattedTransferValue
                                                }}</span
                                            >
                                        </div>
                                        <div
                                            class="financial-item-small"
                                            :title="formattedWage"
                                        >
                                            <q-icon
                                                name="payments"
                                                class="q-mr-xs"
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'light-blue-4'
                                                        : 'light-blue-7'
                                                "
                                            />
                                            <span
                                                :class="
                                                    qInstance.dark.isActive
                                                        ? 'text-light-blue-4'
                                                        : 'text-light-blue-7'
                                                "
                                                >{{ formattedWage }}</span
                                            >
                                        </div>
                                    </div>
                                </div>

                                <q-separator spaced="sm" class="q-my-md" />

                                <div class="q-mt-md fifa-stats-section">
                                    <div
                                        class="row q-col-gutter-xs justify-center fifa-stats-grid-redesigned"
                                    >
                                        <div
                                            v-for="stat in fifaStatsToDisplay"
                                            :key="stat.name"
                                            class="col-fifa-stat"
                                        >
                                            <q-card
                                                flat
                                                bordered
                                                :class="[
                                                    'q-pa-xs rounded-borders fifa-stat-item-redesigned text-center',
                                                    getUnifiedRatingClass(
                                                        player[stat.name],
                                                        100,
                                                    ),
                                                ]"
                                            >
                                                <div
                                                    class="text-caption fifa-label-redesigned"
                                                >
                                                    {{ stat.label }}
                                                </div>
                                                <div
                                                    class="text-subtitle1 text-weight-medium fifa-value-redesigned"
                                                >
                                                    {{
                                                        player[stat.name] !==
                                                        undefined
                                                            ? player[stat.name]
                                                            : "-"
                                                    }}
                                                </div>
                                                <q-tooltip
                                                    :class="
                                                        qInstance.dark.isActive
                                                            ? 'bg-grey-7 text-white'
                                                            : 'bg-white text-dark'
                                                    "
                                                    :delay="500"
                                                    max-width="350px"
                                                >
                                                    <div class="text-weight-medium q-mb-xs">
                                                        {{ stat.label }}
                                                    </div>
                                                    <div class="text-caption">
                                                        {{ fifaToFmAttributeMapping[stat.name]?.description || 'No FM attribute mapping available' }}
                                                    </div>
                                                </q-tooltip>
                                            </q-card>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>
                        <div
                            class="row q-col-gutter-md attribute-columns-container"
                        >
                            <div class="col-12 col-md-4 column">
                                <q-card
                                    flat
                                    bordered
                                    :class="[
                                        qInstance.dark.isActive
                                            ? 'bg-grey-9'
                                            : 'bg-grey-1',
                                        'full-height-card',
                                        'rounded-borders',
                                        'modern-attribute-card',
                                    ]"
                                >
                                    <q-card-section
                                        :class="
                                            qInstance.dark.isActive
                                                ? 'bg-grey-8'
                                                : 'bg-grey-3'
                                        "
                                        class="q-pa-sm attribute-category-header"
                                    >
                                        <div
                                            class="text-subtitle1 text-weight-medium text-center"
                                        >
                                            {{
                                                isGoalkeeper
                                                    ? "Goalkeeping"
                                                    : "Technical"
                                            }}
                                        </div>
                                    </q-card-section>
                                    <q-list
                                        separator
                                        dense
                                        class="attribute-list no-scroll"
                                    >
                                        <q-item
                                            v-for="attrKey in isGoalkeeper
                                                ? attributeCategories.goalkeeping
                                                : attributeCategories.technical"
                                            :key="attrKey"
                                            class="attribute-list-item"
                                        >
                                            <q-item-section>
                                                <q-item-label
                                                    lines="1"
                                                    class="attribute-name-label"
                                                    >{{
                                                        attributeFullNameMap[
                                                            attrKey
                                                        ] || attrKey
                                                    }}</q-item-label
                                                >
                                                <q-tooltip
                                                    :class="
                                                        qInstance.dark.isActive
                                                            ? 'bg-grey-7 text-white'
                                                            : 'bg-white text-dark'
                                                    "
                                                    :delay="500"
                                                    max-width="300px"
                                                >
                                                    <div class="text-weight-medium q-mb-xs">
                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                    </div>
                                                    <div class="text-caption">
                                                        {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                    </div>
                                                </q-tooltip>
                                            </q-item-section>
                                            <q-item-section side>
                                                <span
                                                    :class="
                                                        getUnifiedRatingClass(
                                                            player.attributes[
                                                                attrKey
                                                            ],
                                                            20,
                                                        )
                                                    "
                                                    class="attribute-value attribute-score-value"
                                                >
                                                    {{
                                                        player.attributes[
                                                            attrKey
                                                        ] !== undefined
                                                            ? player.attributes[
                                                                  attrKey
                                                              ]
                                                            : "-"
                                                    }}
                                                </span>
                                            </q-item-section>
                                        </q-item>
                                        <q-item
                                            v-if="
                                                !(
                                                    isGoalkeeper
                                                        ? attributeCategories.goalkeeping
                                                        : attributeCategories.technical
                                                ).length
                                            "
                                        >
                                            <q-item-section
                                                class="text-grey-6 text-center q-py-md"
                                                >No
                                                {{
                                                    isGoalkeeper
                                                        ? "goalkeeping"
                                                        : "technical"
                                                }}
                                                attributes.</q-item-section
                                            >
                                        </q-item>
                                    </q-list>
                                </q-card>
                            </div>
                            <div class="col-12 col-md-4 column">
                                <q-card
                                    flat
                                    bordered
                                    :class="[
                                        qInstance.dark.isActive
                                            ? 'bg-grey-9'
                                            : 'bg-grey-1',
                                        'full-height-card',
                                        'rounded-borders',
                                        'modern-attribute-card',
                                    ]"
                                >
                                    <q-card-section
                                        :class="
                                            qInstance.dark.isActive
                                                ? 'bg-grey-8'
                                                : 'bg-grey-3'
                                        "
                                        class="q-pa-sm attribute-category-header"
                                    >
                                        <div
                                            class="text-subtitle1 text-weight-medium text-center"
                                        >
                                            Mental
                                        </div>
                                    </q-card-section>
                                    <q-list
                                        separator
                                        dense
                                        class="attribute-list no-scroll"
                                    >
                                        <q-item
                                            v-for="attrKey in attributeCategories.mental"
                                            :key="attrKey"
                                            class="attribute-list-item"
                                        >
                                            <q-item-section>
                                                <q-item-label
                                                    lines="1"
                                                    class="attribute-name-label"
                                                    >{{
                                                        attributeFullNameMap[
                                                            attrKey
                                                        ] || attrKey
                                                    }}</q-item-label
                                                >
                                                <q-tooltip
                                                    :class="
                                                        qInstance.dark.isActive
                                                            ? 'bg-grey-7 text-white'
                                                            : 'bg-white text-dark'
                                                    "
                                                    :delay="500"
                                                    max-width="300px"
                                                >
                                                    <div class="text-weight-medium q-mb-xs">
                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                    </div>
                                                    <div class="text-caption">
                                                        {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                    </div>
                                                </q-tooltip>
                                            </q-item-section>
                                            <q-item-section side>
                                                <span
                                                    :class="
                                                        getUnifiedRatingClass(
                                                            player.attributes[
                                                                attrKey
                                                            ],
                                                            20,
                                                        )
                                                    "
                                                    class="attribute-value attribute-score-value"
                                                >
                                                    {{
                                                        player.attributes[
                                                            attrKey
                                                        ] !== undefined
                                                            ? player.attributes[
                                                                  attrKey
                                                              ]
                                                            : "-"
                                                    }}
                                                </span>
                                            </q-item-section>
                                        </q-item>
                                        <q-item
                                            v-if="
                                                !attributeCategories.mental
                                                    .length
                                            "
                                        >
                                            <q-item-section
                                                class="text-grey-6 text-center q-py-md"
                                                >No mental
                                                attributes.</q-item-section
                                            >
                                        </q-item>
                                    </q-list>
                                </q-card>
                            </div>
                            <div class="col-12 col-md-4 column q-gutter-y-md">
                                <q-card
                                    flat
                                    bordered
                                    :class="[
                                        qInstance.dark.isActive
                                            ? 'bg-grey-9'
                                            : 'bg-grey-1',
                                        'rounded-borders',
                                        'physical-attributes-card',
                                    ]"
                                >
                                    <q-card-section
                                        :class="
                                            qInstance.dark.isActive
                                                ? 'bg-grey-8'
                                                : 'bg-grey-3'
                                        "
                                        class="q-pa-sm attribute-category-header"
                                    >
                                        <div
                                            class="text-subtitle1 text-weight-medium text-center"
                                        >
                                            Physical
                                        </div>
                                    </q-card-section>
                                    <q-list
                                        separator
                                        dense
                                        class="attribute-list physical-list no-scroll"
                                    >
                                        <q-item
                                            v-for="attrKey in attributeCategories.physical"
                                            :key="attrKey"
                                            class="attribute-list-item"
                                        >
                                            <q-item-section>
                                                <q-item-label
                                                    lines="1"
                                                    class="attribute-name-label"
                                                    >{{
                                                        attributeFullNameMap[
                                                            attrKey
                                                        ] || attrKey
                                                    }}</q-item-label
                                                >
                                                <q-tooltip
                                                    :class="
                                                        qInstance.dark.isActive
                                                            ? 'bg-grey-7 text-white'
                                                            : 'bg-white text-dark'
                                                    "
                                                    :delay="500"
                                                    max-width="300px"
                                                >
                                                    <div class="text-weight-medium q-mb-xs">
                                                        {{ attributeFullNameMap[attrKey] || attrKey }}
                                                    </div>
                                                    <div class="text-caption">
                                                        {{ attributeDescriptions[attrKey] || 'No description available' }}
                                                    </div>
                                                </q-tooltip>
                                            </q-item-section>
                                            <q-item-section side>
                                                <span
                                                    :class="
                                                        getUnifiedRatingClass(
                                                            player.attributes[
                                                                attrKey
                                                            ],
                                                            20,
                                                        )
                                                    "
                                                    class="attribute-value attribute-score-value"
                                                >
                                                    {{
                                                        player.attributes[
                                                            attrKey
                                                        ] !== undefined
                                                            ? player.attributes[
                                                                  attrKey
                                                              ]
                                                            : "-"
                                                    }}
                                                </span>
                                            </q-item-section>
                                        </q-item>
                                        <q-item
                                            v-if="
                                                !attributeCategories.physical
                                                    .length
                                            "
                                        >
                                            <q-item-section
                                                class="text-grey-6 text-center q-py-md"
                                                >No physical
                                                attributes.</q-item-section
                                            >
                                        </q-item>
                                    </q-list>
                                </q-card>
                                <q-card
                                    flat
                                    bordered
                                    :class="[
                                        qInstance.dark.isActive
                                            ? 'bg-grey-9'
                                            : 'bg-grey-1',
                                        'rounded-borders',
                                        'role-ratings-card',
                                    ]"
                                    v-if="
                                        player.roleSpecificOveralls &&
                                        player.roleSpecificOveralls.length > 0
                                    "
                                >
                                    <q-card-section
                                        :class="
                                            qInstance.dark.isActive
                                                ? 'bg-grey-8'
                                                : 'bg-grey-3'
                                        "
                                        class="q-pa-sm attribute-category-header"
                                    >
                                        <div
                                            class="text-subtitle1 text-weight-medium text-center"
                                        >
                                            Best Roles
                                        </div>
                                    </q-card-section>
                                    <q-list
                                        separator
                                        dense
                                        class="constrained-scroll-list role-specific-ratings-list"
                                    >
                                        <q-item
                                            v-for="roleOverall in sortedRoleSpecificOveralls"
                                            :key="roleOverall.roleName"
                                            :class="{
                                                'best-role-highlight':
                                                    roleOverall.score ===
                                                    player.Overall,
                                            }"
                                            :style="
                                                roleOverall.score ===
                                                player.Overall
                                                    ? qInstance.dark.isActive
                                                        ? 'background-color: #2a5270 !important;'
                                                        : 'background-color: #e3f2fd !important;'
                                                    : ''
                                            "
                                            class="attribute-list-item"
                                        >
                                            <q-item-section>
                                                <q-item-label
                                                    lines="1"
                                                    class="attribute-name-label"
                                                    :title="
                                                        roleOverall.roleName
                                                    "
                                                    >{{
                                                        roleOverall.roleName
                                                    }}</q-item-label
                                                >
                                            </q-item-section>
                                            <q-item-section side>
                                                <span
                                                    :class="
                                                        getUnifiedRatingClass(
                                                            roleOverall.score,
                                                            100,
                                                        )
                                                    "
                                                    class="attribute-value fifa-stat-value attribute-score-value"
                                                >
                                                    {{ roleOverall.score }}
                                                </span>
                                            </q-item-section>
                                        </q-item>
                                    </q-list>
                                </q-card>
                            </div>
                        </div>
                    </div>
                </div>
            </q-card-section>

            <q-card-section v-else class="text-center q-pa-xl">
                <q-spinner color="primary" size="3em" />
                <div class="q-mt-md text-grey-7">Loading player data...</div>
            </q-card-section>
        </q-card>
    </q-dialog>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref, watch } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import { formatCurrency } from '../utils/currencyUtils'

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
    description: 'Based on Finishing, Long Shots, Penalties, Heading, Composure, Technique, Anticipation, Decisions, and Flair'
  },
  PAS: {
    primary: ['Pas', 'Vis'],
    secondary: ['Cro', 'Tec', 'Fre', 'Tea', 'Dec', 'Fir', 'Cor', 'OtB'],
    description: 'Based on Passing, Vision, Crossing, Technique, Free Kicks, Teamwork, Decisions, First Touch, Corners, and Off the Ball'
  },
  DRI: {
    primary: ['Dri', 'Fir', 'Tec'],
    secondary: ['Fla', 'Cmp', 'OtB'],
    description: 'Based on Dribbling, First Touch, Technique, Flair, Composure, and Off the Ball'
  },
  DEF: {
    primary: ['Mar', 'Tck', 'Ant', 'Pos'],
    secondary: ['Hea', 'Cnt', 'Dec', 'Cmp', 'Bra', 'Agg', 'Wor'],
    description: 'Based on Marking, Tackling, Anticipation, Positioning, Heading, Concentration, Decisions, Composure, Bravery, Aggression, and Work Rate'
  },
  PHY: {
    primary: ['Str', 'Sta', 'Nat'],
    secondary: ['Jum', 'Agg', 'Bra', 'Wor', 'Bal'],
    description: 'Based on Strength, Stamina, Natural Fitness, Jumping Reach, Aggression, Bravery, Work Rate, and Balance'
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
    description: 'Based on Positioning, Command of Area, Anticipation, Decisions, Rushing Out Tendency, Concentration, and Communication'
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
  'Cr C/A': 'Cross Completion %'
}

const performanceStatCategories = {
  Offensive: ['Gls/90', 'xG/90', 'Shot/90', 'ShT/90', 'Conv %', 'Drb/90'],
  Passing: [
    'Asts/90',
    'xA/90',
    'Ch C/90',
    'K Ps/90',
    'Ps C/90',
    'Pas %',
    'Pr passes/90',
    'Cr C/90',
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
    'Poss Won/90'
  ]
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
  props: {
    player: { type: Object, default: () => null },
    show: { type: Boolean, default: false },
    currencySymbol: { type: String, default: '$' },
    datasetId: { type: String, default: null }
  },
  emits: ['close'],
  setup(props) {
    const qInstance = useQuasar()
    const _playerStore = usePlayerStore()
    const selectedComparisonGroup = ref('Global')
    const flagLoadError = ref(false)
    const divisionFilter = ref('all')

    // Face image handling
    const faceImageLoadError = ref(false)

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
      () => {
        faceImageLoadError.value = false
      },
      { immediate: true }
    )

    onMounted(() => {
      /* Initialization logic if needed */
    })

    const performanceComparisonOptions = computed(() => {
      const options = []
      if (
        !props.player ||
        !props.player.performancePercentiles ||
        !props.player.shortPositions ||
        !props.player.positionGroups
      ) {
        // If essential player data is missing, return empty or just Global if it exists
        if (props.player?.performancePercentiles?.Global) {
          return [{ label: 'Overall Dataset', value: 'Global' }]
        }
        return options
      }

      const playerPercentiles = props.player.performancePercentiles
      const playerShortPositions = props.player.shortPositions // e.g., ["DC", "DM"]
      const playerBroadGroups = props.player.positionGroups // e.g., ["Defenders", "Midfielders"]

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

      for (const groupKey of preferredOrder) {
        if (Object.prototype.hasOwnProperty.call(playerPercentiles, groupKey)) {
          let includeGroup = false
          if (groupKey === 'Global') {
            includeGroup = true
          } else if (['Goalkeepers', 'Defenders', 'Midfielders', 'Attackers'].includes(groupKey)) {
            // Broad groups: check if player belongs to this broad group
            if (playerBroadGroups.includes(groupKey)) {
              includeGroup = true
            }
          } else if (detailedGroupToShortPositionsMap[groupKey]) {
            // Detailed groups: check if player's short positions match any in the detailed group
            const requiredShortPos = detailedGroupToShortPositionsMap[groupKey]
            if (playerShortPositions.some(psp => requiredShortPos.includes(psp))) {
              includeGroup = true
            }
          }

          if (includeGroup && !options.some(opt => opt.value === groupKey)) {
            options.push({
              label: groupKey === 'Global' ? 'Overall Dataset' : `vs. ${groupKey}`,
              value: groupKey
            })
          }
        }
      }

      // Add any other percentile groups the player might have that are not in preferredOrder
      // (This ensures if backend adds new groups not yet in preferredOrder, they still show up if relevant)
      for (const groupKey of Object.keys(playerPercentiles)) {
        if (!options.some(opt => opt.value === groupKey)) {
          let includeGroup = false
          if (['Goalkeepers', 'Defenders', 'Midfielders', 'Attackers'].includes(groupKey)) {
            if (playerBroadGroups.includes(groupKey)) includeGroup = true
          } else if (detailedGroupToShortPositionsMap[groupKey]) {
            const requiredShortPos = detailedGroupToShortPositionsMap[groupKey]
            if (playerShortPositions.some(psp => requiredShortPos.includes(psp)))
              includeGroup = true
          } else if (groupKey === 'Global') {
            // Should be caught by preferredOrder but good fallback
            includeGroup = true
          }
          // For any other arbitrary group key not covered, we might include it if it exists,
          // or decide to only show explicitly defined/matched ones.
          // For now, let's assume if it's in playerPercentiles and not yet added, and it's not a known type we already checked,
          // it might be a custom group or a new one we should display.
          // However, to be safer and more aligned with the request, only add if it's a known type.
          // The current logic with preferredOrder and specific checks should cover all defined groups.
          // This secondary loop might not be strictly necessary if preferredOrder is exhaustive for relevant groups.
          // Let's refine to ensure it doesn't add groups the player doesn't belong to.
          if (includeGroup) {
            // Re-check belonging for keys not in preferredOrder
            options.push({
              label: `vs. ${groupKey}`,
              value: groupKey
            })
          }
        }
      }

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

    watch(
      () => props.player,
      newPlayer => {
        flagLoadError.value = false
        const newOptions = performanceComparisonOptions.value

        if (newPlayer?.performancePercentiles) {
          // Always try to set to the position group for the highest role first
          const highestRoleGroup = getPositionGroupForHighestRole()

          if (highestRoleGroup && newOptions.some(opt => opt.value === highestRoleGroup)) {
            selectedComparisonGroup.value = highestRoleGroup
          } else if (!newOptions.some(opt => opt.value === selectedComparisonGroup.value)) {
            // If highest role group not available and current selection is invalid, fallback
            if (newOptions.some(opt => opt.value === 'Global')) {
              selectedComparisonGroup.value = 'Global'
            } else if (newOptions.length > 0) {
              selectedComparisonGroup.value = newOptions[0].value
            } else {
              selectedComparisonGroup.value = 'Global'
            }
          }
        } else {
          selectedComparisonGroup.value = 'Global'
        }
      },
      { immediate: true, deep: true }
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
        // Instead of refetching all data, we need to fetch just updated percentiles for this player
        // For now, let's create a dedicated API call for this
        const response = await fetch(`/api/percentiles/${props.datasetId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            playerName: props.player.name,
            divisionFilter: divisionFilter.value,
            targetDivision: targetDivision
          })
        })

        if (response.ok) {
          const updatedPercentiles = await response.json()
          // Update the player's percentiles without affecting the main dataset
          if (props.player.performancePercentiles) {
            Object.assign(props.player.performancePercentiles, updatedPercentiles)
          }
        }
      } catch (error) {
        console.error('Error updating percentiles with division filter:', error)
      }
    }

    const isGoalkeeper = computed(() => {
      if (!props.player) return false
      return (
        props.player.shortPositions?.includes('GK') ||
        props.player.positionGroups?.includes('Goalkeepers') ||
        props.player.parsedPositions?.includes('Goalkeeper')
      )
    })

    const getPlayerAttributesInOrder = categoryOrderedKeys => {
      if (!props.player || !props.player.attributes) return []
      return categoryOrderedKeys.filter(key =>
        Object.prototype.hasOwnProperty.call(props.player.attributes, key)
      )
    }

    const attributeCategories = computed(() => ({
      technical: getPlayerAttributesInOrder(technicalAttrsOrdered),
      mental: getPlayerAttributesInOrder(mentalAttrsOrdered),
      physical: getPlayerAttributesInOrder(physicalAttrsOrdered),
      goalkeeping: isGoalkeeper.value ? getPlayerAttributesInOrder(goalkeepingAttrsOrdered) : []
    }))

    const fifaStatsToDisplay = computed(() => {
      let orderedStats = []
      if (isGoalkeeper.value) {
        orderedStats = [
          { name: 'DIV', label: 'DIV' },
          { name: 'HAN', label: 'HAN' },
          { name: 'REF', label: 'REF' },
          { name: 'KIC', label: 'KIC' },
          { name: 'SPD', label: 'SPD' },
          { name: 'POS', label: 'POS' }
        ]
      } else {
        orderedStats = [
          { name: 'PAC', label: 'PAC' },
          { name: 'SHO', label: 'SHO' },
          { name: 'PAS', label: 'PAS' },
          { name: 'DRI', label: 'DRI' },
          { name: 'DEF', label: 'DEF' },
          { name: 'PHY', label: 'PHY' }
        ]
      }
      return orderedStats.filter(stat => props.player && props.player[stat.name] !== undefined)
    })

    const averageRatingData = computed(() => {
      if (!props.player || !props.player.attributes || !props.player.performancePercentiles)
        return null
      const groupKey = selectedComparisonGroup.value
      const percentilesForGroup = props.player.performancePercentiles[groupKey]
      if (
        !percentilesForGroup ||
        !Object.prototype.hasOwnProperty.call(props.player.attributes, 'Av Rat') ||
        props.player.attributes['Av Rat'] === '-' ||
        props.player.attributes['Av Rat'] === '' ||
        !Object.prototype.hasOwnProperty.call(percentilesForGroup, 'Av Rat')
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

    // Cache for performance stats to avoid rebuilding on every change
    const performanceStatsCache = new Map()

    const categorizedPerformanceStats = computed(() => {
      if (!props.player || !props.player.attributes || !props.player.performancePercentiles)
        return {}

      const groupKey = selectedComparisonGroup.value
      const percentilesForGroup = props.player.performancePercentiles[groupKey]
      if (!percentilesForGroup) return {}

      // Create cache key based on player UID and group
      let playerUID = props.player.UID || props.player.uid

      // If no UID available or UID is empty, create a composite unique key
      if (!playerUID || playerUID === '') {
        playerUID = `${props.player.name || 'unknown'}-${props.player.club || 'unknown'}-${props.player.age || 'unknown'}-${props.player.position || 'unknown'}`
      }

      const cacheKey = `${playerUID}-${groupKey}`

      // Return cached result if available
      if (performanceStatsCache.has(cacheKey)) {
        return performanceStatsCache.get(cacheKey)
      }

      const result = {}
      const categoryOrder = ['Offensive', 'Passing', 'Defensive']

      for (const categoryName of categoryOrder) {
        if (performanceStatCategories[categoryName]) {
          const statsInCategory = []
          for (const statKey of performanceStatCategories[categoryName]) {
            const rawAttributeValue = props.player.attributes[statKey]
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

          if (statsInCategory.length > 0) {
            result[categoryName] = statsInCategory.sort((a, b) => a.name.localeCompare(b.name))
          }
        }
      }

      // Cache result (limit cache size)
      if (performanceStatsCache.size > 20) {
        performanceStatsCache.clear()
      }
      performanceStatsCache.set(cacheKey, result)

      return result
    })

    const hasAnyPerformanceData = computed(
      () => averageRatingData.value || Object.keys(categorizedPerformanceStats.value).length > 0
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

    const sortedRoleSpecificOveralls = computed(() => {
      if (!props.player?.roleSpecificOveralls) {
        return []
      }

      // Only sort if the array has changed, avoiding unnecessary operations
      const roleOveralls = props.player.roleSpecificOveralls
      if (roleOveralls.length <= 1) {
        return roleOveralls
      }

      return [...roleOveralls].sort((a, b) => b.score - a.score)
    })

    const formattedTransferValue = computed(() => {
      if (!props.player) return '-'
      return formatCurrency(
        props.player.transferValueAmount,
        props.currencySymbol,
        props.player.transfer_value
      )
    })

    const formattedWage = computed(() => {
      if (!props.player) return '-'
      return formatCurrency(props.player.wageAmount, props.currencySymbol, props.player.wage)
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
      averageRatingData,
      hasAnyPerformanceData,
      flagLoadError,
      handleFlagError,
      divisionFilter,
      divisionFilterOptions,
      onDivisionFilterChange,
      faceImageLoadError,
      handleFaceImageError,
      handleFaceImageLoad,
      playerFaceImageUrl
    }
  }
})
</script>

<style lang="scss" scoped>
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

.player-detail-dialog-card {
    display: flex;
    flex-direction: column;
    border-radius: 8px;
}
.main-content-section {
    flex-grow: 1;
    padding: 12px;
}
.dialog-title {
    font-size: clamp(0.9rem, 1.4vw, 1.1rem);
}

.profile-header-section {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 0;
}
.player-identity-extended {
    flex-grow: 1;
    padding-right: 16px;
    display: flex;
    flex-direction: column;
}

.player-flag-container-redesigned {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    padding-top: 4px;
}
.player-flag-redesigned {
    border: 1px solid rgba(128, 128, 128, 0.5);
    border-radius: 3px;
    object-fit: cover;
    vertical-align: middle;
}

.player-name-age-positions-redesigned {
    display: flex;
    flex-direction: column;
    justify-content: center;
    overflow: hidden;
    text-align: left;
}
.player-name-and-age {
    display: flex;
    align-items: baseline;
    flex-wrap: nowrap;
}
.player-name-text-redesigned {
    font-size: clamp(1.1rem, 2vw, 1.4rem);
    line-height: 1.2;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-right: 8px;
}
.player-age-badge-redesigned {
    font-size: clamp(0.7rem, 1.1vw, 0.8rem);
    font-weight: 600;
    padding: 2px 6px;
    align-self: baseline;
    white-space: nowrap;
}
.player-positions-inline {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    margin-top: 2px;
}
.player-position-badge {
    font-size: clamp(0.65rem, 0.9vw, 0.75rem);
    font-weight: 500;
    padding: 1px 4px;
    margin-right: 4px;
    margin-bottom: 2px;
}

.player-additional-details {
    margin-top: 6px;
    .row {
        align-items: flex-start;
    }
}
.additional-detail-item {
    display: flex;
    align-items: center;
    padding: 2px 0;
    line-height: 1.2;
}
.additional-detail-icon {
    align-self: center;
}
.additional-detail-text-block {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    text-align: left;
}
.additional-detail-caption {
    font-size: clamp(0.6rem, 0.8vw, 0.65rem);
    line-height: 1.1;
    color: $grey-6;
    .body--dark & {
        color: $grey-5;
    }
    white-space: nowrap;
    margin-bottom: -2px;
}
.additional-detail-label {
    font-size: clamp(0.75rem, 1vw, 0.85rem);
    line-height: 1.2;
    font-weight: 500;
    white-space: nowrap;
    &.ellipsis {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}
.additional-detail-item-separator {
    padding: 0 4px;
    color: $grey-5;
    align-self: center;
    font-size: 0.8em;
    line-height: 1.2;
}

.financial-details-top-right {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    text-align: right;
    min-width: 120px;
}
.financial-item-large {
    font-size: clamp(1.1rem, 2.2vw, 1.5rem);
    font-weight: 700;
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: flex;
    align-items: center;
}
.financial-item-small {
    font-size: clamp(0.8rem, 1.5vw, 0.95rem);
    font-weight: 500;
    line-height: 1.3;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-top: 2px;
    display: flex;
    align-items: center;
    color: $grey-7;
    .body--dark & {
        color: $grey-4;
    }
}

.q-separator.q-my-md {
    margin-top: 12px !important;
    margin-bottom: 12px !important;
}

.fifa-stats-section {
}
.fifa-title-redesigned {
    font-size: clamp(0.8rem, 1.3vw, 0.95rem);
    font-weight: 500;
}
.fifa-stats-grid-redesigned {
}
.col-fifa-stat {
    padding: 1px;
    flex-basis: calc(100% / 8);
    max-width: calc(100% / 8);
    @media (max-width: $breakpoint-sm-max) {
        flex-basis: calc(100% / 6);
        max-width: calc(100% / 6);
    }
    @media (max-width: $breakpoint-xs-max) {
        flex-basis: calc(100% / 4);
        max-width: calc(100% / 4);
    }
}
.fifa-stat-item-redesigned {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 4px 6px !important;
    border-width: 1px;
    overflow: hidden;
    line-height: 1.1;
    min-width: 40px;

    &.rating-na {
        background-color: #757575;
        color: white;
    }
    &.rating-tier-1 {
        background-color: #d32f2f;
        color: white;
    }
    &.rating-tier-2 {
        background-color: #ef6c00;
        color: white;
    }
    &.rating-tier-3 {
        background-color: #fdd835;
        color: #333;
    }
    &.rating-tier-4 {
        background-color: #aed581;
        color: #333;
    }
    &.rating-tier-5 {
        background-color: #66bb6a;
        color: white;
    }
    &.rating-tier-6 {
        background-color: #388e3c;
        color: white;
    }
}
.fifa-label-redesigned {
    font-size: clamp(0.7rem, 1.6vw, 0.85rem);
    font-weight: 500;
    margin-bottom: 1px;
    display: block;
    line-height: 1.1;
}
.fifa-value-redesigned {
    font-size: clamp(1rem, 2.4vw, 1.5rem);
    font-weight: 700;
    display: block;
    line-height: 1.1;
}

.attributes-section-title {
    font-size: clamp(1.05rem, 1.8vw, 1.3rem);
}
.attribute-category-header .text-subtitle1 {
    font-size: clamp(0.85rem, 1.4vw, 1rem);
}
.attribute-list-item .attribute-name-label {
    font-size: clamp(0.7rem, 1vw, 0.8rem);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 150px;
}
.attribute-list-item .attribute-score-value {
    font-size: clamp(0.75rem, 1.1vw, 0.85rem);
}
.performance-stat-actual-value {
    font-size: clamp(0.7rem, 1vw, 0.8rem);
    min-width: 35px;
    text-align: right;
}
.attribute-columns-container > .column {
    display: flex;
    flex-direction: column;
}
.full-height-card {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
}
.percentile-content-area {
    padding: 8px;
}
.attribute-list.no-scroll {
    flex-grow: 0;
    flex-shrink: 0;
    overflow-y: visible;
}
.physical-attributes-card .attribute-list.physical-list.no-scroll {
    flex-grow: 0;
    flex-shrink: 0;
    overflow-y: visible;
}
.role-ratings-card .role-specific-ratings-list {
    overflow-y: auto;
    flex-shrink: 1;
    max-height: 18vh;
    min-height: 70px;
}
.best-role-highlight {
    border-left: 4px solid $positive;
    .body--dark & {
        border-left: 4px solid lighten($positive, 20%);
        background-color: rgba(lighten($positive, 25%), 0.15) !important;
    }
    .body--light & {
        background-color: rgba($positive, 0.08) !important;
    }
}
.best-role-highlight .q-item__label {
    font-weight: 600;
}
.q-list--dense .q-item,
.attribute-list-item,
.constrained-scroll-list .q-item.attribute-list-item {
    padding: 1px 6px;
    min-height: auto;
}
.min-height-auto {
    min-height: auto !important;
}
.q-list--separator > .q-item:not(:first-child):before {
    background: rgba(128, 128, 128, 0.2);
}
.q-card__section.bg-grey-3 {
    background-color: $grey-3 !important;
}
.q-card__section.bg-grey-8 {
    background-color: $grey-8 !important;
}
.q-card[flat][bordered] {
    border: 1px solid rgba(128, 128, 128, 0.3);
    .body--dark & {
        border: 1px solid rgba(128, 128, 128, 0.4);
    }
}
.performance-stat-item {
    .stat-name-section {
        flex-basis: 45%;
        flex-grow: 0;
        flex-shrink: 0;
        padding-right: 6px;
    }
    .stat-bar-section {
        flex-grow: 1;
        display: flex;
        align-items: center;
    }
    .stat-value-section {
        flex-basis: 15%;
        flex-grow: 0;
        flex-shrink: 0;
        text-align: right;
        padding-left: 6px;
    }
}
.stat-bar-container {
    display: flex;
    align-items: center;
    width: 100%;
}
.stat-bar-track {
    flex-grow: 1;
    height: 12px;
    background-color: $grey-2;
    border-radius: 3px;
    margin-right: 6px;
    overflow: hidden;
    .body--dark & {
        background-color: $grey-7;
    }
}
.stat-bar-fill {
    height: 100%;
    border-radius: 3px;
}
.stat-percentile-text {
    font-size: 0.65rem;
    min-width: 22px;
    text-align: right;
    .body--dark & {
        color: $grey-5;
    }
    .body--light & {
        color: $grey-7;
    }
}
hr.q-my-sm {
    border: none;
    height: 1px;
    background-color: rgba(128, 128, 128, 0.2);
    margin-top: 8px;
    margin-bottom: 8px;
}

@media (max-width: $breakpoint-sm-max) {
    .main-content-section {
        padding: 8px;
    }
    .role-ratings-card .role-specific-ratings-list {
        max-height: 18vh;
    }
    .player-name-text-redesigned {
        font-size: clamp(0.9rem, 1.6vw, 1.15rem);
    }
    .player-age-badge-redesigned {
        font-size: clamp(0.6rem, 0.9vw, 0.7rem);
    }
    .player-position-badge {
        font-size: clamp(0.6rem, 0.8vw, 0.7rem);
    }
    .additional-detail-caption {
        font-size: clamp(0.55rem, 0.7vw, 0.6rem);
    }
    .additional-detail-label {
        font-size: clamp(0.7rem, 0.9vw, 0.8rem);
    }
    .financial-item-large {
        font-size: clamp(1rem, 2vw, 1.3rem);
    }
    .financial-item-small {
        font-size: clamp(0.75rem, 1.3vw, 0.9rem);
    }
    .fifa-label-redesigned {
        font-size: clamp(0.65rem, 1.3vw, 0.8rem);
    }
    .fifa-value-redesigned {
        font-size: clamp(0.9rem, 2vw, 1.3rem);
    }
}

@media (max-width: $breakpoint-xs-max) {
    .main-content-section {
        padding: 4px;
    }
    .dialog-title {
        font-size: 0.85rem;
    }
    .profile-header-section {
        flex-direction: column;
        align-items: stretch;
    }
    .player-identity-extended {
        padding-right: 0;
        margin-bottom: 8px;
        width: 100%;
    }
    .player-name-age-positions-redesigned {
        flex-direction: column;
        align-items: flex-start;
    }
    .player-name-and-age {
        flex-wrap: wrap;
    }
    .player-additional-details .row {
        justify-content: flex-start;
    }
    .additional-detail-item {
        flex-basis: 100% !important;
        justify-content: flex-start;
        margin-bottom: 2px;
    }
    .additional-detail-item-separator {
        display: none;
    }
    .financial-details-top-right {
        align-items: flex-start;
        text-align: left;
        width: 100%;
        margin-bottom: 8px;
    }
    .player-name-text-redesigned {
        font-size: clamp(0.9rem, 2.8vw, 1.1rem);
    }
    .player-age-badge-redesigned {
        font-size: clamp(0.65rem, 2vw, 0.75rem);
        margin-left: 0;
        margin-top: 2px;
    }
    .player-positions-inline {
        margin-top: 4px;
    }
    .player-position-badge {
        font-size: clamp(0.6rem, 1.8vw, 0.7rem);
    }
    .additional-detail-caption {
        font-size: clamp(0.6rem, 1.5vw, 0.65rem);
    }
    .additional-detail-label {
        font-size: clamp(0.75rem, 2vw, 0.85rem);
    }
    .fifa-title-redesigned {
        font-size: clamp(0.75rem, 2vw, 0.85rem);
    }
    .fifa-label-redesigned {
        font-size: clamp(0.6rem, 1.4vw, 0.7rem);
    }
    .fifa-value-redesigned {
        font-size: clamp(0.8rem, 2.2vw, 1.1rem);
    }
    .attributes-section-title {
        font-size: 1rem;
    }
    .attribute-category-header .text-subtitle1 {
        font-size: 0.8rem;
    }
    .attribute-list-item .attribute-name-label {
        font-size: 0.7rem;
        max-width: 90px;
    }
    .attribute-list-item .attribute-score-value {
        font-size: 0.75rem;
    }
    .performance-stat-actual-value {
        font-size: 0.7rem;
        min-width: 30px;
    }
    .stat-percentile-text {
        font-size: 0.6rem;
        min-width: 20px;
    }
    .attribute-list.no-scroll {
        overflow-y: auto;
        max-height: 22vh;
    }
    .role-ratings-card .role-specific-ratings-list {
        max-height: 14vh;
    }
    .performance-stat-item {
        .stat-name-section {
            flex-basis: 40%;
            font-size: 0.6rem;
        }
        .stat-value-section {
            flex-basis: 20%;
            font-size: 0.6rem;
        }
    }
}

// Modern Dialog Enhancements
.modern-dialog-card {
    border-radius: 16px !important;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
    overflow: hidden;
    
    .body--dark & {
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    }
}

.modern-dialog-header {
    background: linear-gradient(135deg, #1976d2 0%, #1565c0 100%);
    padding: 16px 20px;
    min-height: 60px;
    
    .dialog-header-content {
        display: flex;
        align-items: center;
        gap: 12px;
        
        .header-icon {
            font-size: 1.5rem;
            opacity: 0.9;
        }
        
        .header-text {
            .dialog-title {
                font-weight: 600;
                margin: 0;
                font-size: 1.2rem;
            }
            
            .dialog-subtitle {
                font-size: 0.85rem;
                opacity: 0.8;
                margin-top: 2px;
            }
        }
    }
    
    .close-btn {
        border-radius: 8px;
        transition: all 0.2s ease;
        
        &:hover {
            background: rgba(255, 255, 255, 0.1);
            transform: scale(1.05);
        }
    }
}

.modern-stats-card {
    border-radius: 12px !important;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    border: 1px solid rgba(0, 0, 0, 0.05);
    overflow: hidden;
    
    .body--dark & {
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }
}

.modern-profile-card {
    border-radius: 12px !important;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    border: 1px solid rgba(0, 0, 0, 0.05);
    
    .body--dark & {
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }
    
    .player-profile-content {
        padding: 20px !important;
    }
}

.modern-attribute-card {
    border-radius: 12px !important;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
    border: 1px solid rgba(0, 0, 0, 0.05);
    transition: all 0.3s ease;
    
    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
    }
    
    .body--dark & {
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
        border: 1px solid rgba(255, 255, 255, 0.1);
        
        &:hover {
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
        }
    }
    
    .attribute-category-header {
        background: linear-gradient(135deg, rgba(25, 118, 210, 0.1) 0%, rgba(25, 118, 210, 0.05) 100%);
        border-radius: 12px 12px 0 0;
        
        .body--dark & {
            background: linear-gradient(135deg, rgba(144, 202, 249, 0.1) 0%, rgba(144, 202, 249, 0.05) 100%);
        }
        
        .text-subtitle1 {
            color: #1976d2;
            font-weight: 600;
            
            .body--dark & {
                color: #90caf9;
            }
        }
    }
}

// Enhanced attribute list items
.attribute-list-item {
    transition: background-color 0.2s ease;
    border-radius: 4px;
    margin: 2px 4px;
    
    &:hover {
        background: rgba(25, 118, 210, 0.05);
        
        .body--dark & {
            background: rgba(144, 202, 249, 0.05);
        }
    }
}

// Enhanced FIFA stats
.fifa-stat-item-redesigned {
    transition: transform 0.2s ease;
    border-radius: 8px !important;
    
    &:hover {
        transform: scale(1.05);
    }
}

// Enhanced performance stats
.performance-stat-item {
    transition: background-color 0.2s ease;
    border-radius: 4px;
    margin: 2px 0;
    
    &:hover {
        background: rgba(25, 118, 210, 0.05);
        
        .body--dark & {
            background: rgba(144, 202, 249, 0.05);
        }
    }
}

.stat-bar-fill {
    transition: width 0.5s ease, background-color 0.3s ease;
}

// Enhanced badges and positions
.player-position-badge, .player-age-badge-redesigned {
    transition: all 0.2s ease;
    
    &:hover {
        transform: scale(1.05);
    }
}

// Player face image styles
.player-face-container {
    display: flex;
    align-items: center;
    justify-content: center;
}

.player-face-image {
    border-radius: 50%;
    border: 2px solid rgba(25, 118, 210, 0.2);
    object-fit: cover;
    transition: all 0.3s ease;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    
    &:hover {
        transform: scale(1.05);
        border-color: rgba(25, 118, 210, 0.4);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }
    
    .body--dark & {
        border-color: rgba(144, 202, 249, 0.2);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
        
        &:hover {
            border-color: rgba(144, 202, 249, 0.4);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
        }
    }
}

.player-face-placeholder {
    border: 2px solid rgba(25, 118, 210, 0.2);
    transition: all 0.3s ease;
    
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
}
</style>
