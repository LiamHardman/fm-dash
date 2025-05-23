<template>
    <q-dialog :model-value="show" @hide="$emit('close')">
        <q-card
            class="player-detail-dialog-card"
            :class="
                qInstance.dark.isActive
                    ? 'bg-dark text-white'
                    : 'bg-white text-dark'
            "
            style="max-width: 1300px; width: 95vw; max-height: 90vh"
        >
            <q-bar
                :class="
                    qInstance.dark.isActive
                        ? 'bg-grey-10'
                        : 'bg-primary text-white'
                "
            >
                <q-icon name="person" class="q-mr-sm" />
                <div class="text-subtitle1 dialog-title">
                    {{ player?.name || "Player" }} - Detailed View
                </div>
                <q-space />
                <q-btn dense flat icon="close" @click="$emit('close')">
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
                        <div
                            class="text-h5 q-mb-sm text-center attributes-section-title"
                        >
                            Performance Percentiles
                        </div>
                        <q-select
                            v-if="performanceComparisonOptions.length > 0"
                            :disable="performanceComparisonOptions.length <= 1"
                            v-model="selectedComparisonGroup"
                            :options="performanceComparisonOptions"
                            label="Compare Against"
                            dense
                            outlined
                            emit-value
                            map-options
                            class="q-mb-md"
                            style="min-width: 200px"
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
                            Only global comparison available. More options
                            appear if player belongs to specific position groups
                            with percentile data.
                        </q-tooltip>

                        <q-card
                            flat
                            bordered
                            :class="[
                                qInstance.dark.isActive
                                    ? 'bg-grey-9'
                                    : 'bg-grey-1',
                                'rounded-borders performance-percentiles-card-left',
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
                            class="q-mb-md player-profile-card-redesigned"
                        >
                            <q-card-section class="q-pa-sm">
                                <div
                                    class="row items-stretch q-col-gutter-x-sm q-mb-sm player-header-redesigned"
                                >
                                    <div
                                        class="col-12 col-sm player-identity-section-redesigned"
                                    >
                                        <div
                                            class="row items-center no-wrap full-height"
                                        >
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
                                                class="col player-name-age-redesigned"
                                            >
                                                <div
                                                    class="text-h6 player-name-text-redesigned"
                                                    :title="player.name || '-'"
                                                >
                                                    {{ player.name || "-" }}
                                                </div>
                                                <q-badge
                                                    outline
                                                    :color="
                                                        qInstance.dark.isActive
                                                            ? 'blue-4'
                                                            : 'primary'
                                                    "
                                                    :label="`${player.age || '-'} yrs`"
                                                    class="player-age-badge-redesigned"
                                                />
                                            </div>
                                        </div>
                                    </div>

                                    <q-separator
                                        vertical
                                        inset
                                        class="q-mx-xs gt-xs"
                                        :dark="qInstance.dark.isActive"
                                    />

                                    <div
                                        class="col-12 col-sm-auto text-center player-value-section-redesigned"
                                    >
                                        <div
                                            class="text-caption value-subtitle-redesigned"
                                            :class="
                                                qInstance.dark.isActive
                                                    ? 'text-grey-5'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Value
                                        </div>
                                        <div
                                            class="text-h5 text-weight-bold value-text-redesigned"
                                            :class="
                                                qInstance.dark.isActive
                                                    ? 'text-green-4'
                                                    : 'text-green-7'
                                            "
                                            :title="formattedTransferValue"
                                        >
                                            {{ formattedTransferValue }}
                                        </div>
                                    </div>
                                </div>

                                <q-separator
                                    :dark="qInstance.dark.isActive"
                                    class="q-my-xs"
                                />

                                <div
                                    class="row q-col-gutter-x-sm q-col-gutter-y-xs q-mt-sm q-mb-sm condensed-info-grid-redesigned"
                                >
                                    <div
                                        class="col-6 col-sm-4 info-item-redesigned"
                                    >
                                        <q-icon
                                            :color="
                                                qInstance.dark.isActive
                                                    ? 'blue-3'
                                                    : 'primary'
                                            "
                                            name="sports_soccer"
                                            size="1.1em"
                                            class="q-mr-xs info-icon-redesigned"
                                        />
                                        <div>
                                            <q-item-label
                                                caption
                                                class="info-caption-redesigned"
                                                >Club</q-item-label
                                            >
                                            <q-item-label
                                                class="info-label-redesigned ellipsis"
                                                :title="player.club || '-'"
                                                >{{
                                                    player.club || "-"
                                                }}</q-item-label
                                            >
                                        </div>
                                    </div>
                                    <div
                                        class="col-6 col-sm-4 info-item-redesigned"
                                    >
                                        <q-icon
                                            :color="
                                                qInstance.dark.isActive
                                                    ? 'blue-3'
                                                    : 'primary'
                                            "
                                            name="engineering"
                                            size="1.1em"
                                            class="q-mr-xs info-icon-redesigned"
                                        />
                                        <div>
                                            <q-item-label
                                                caption
                                                class="info-caption-redesigned"
                                                >Position(s)</q-item-label
                                            >
                                            <q-item-label
                                                class="info-label-redesigned ellipsis"
                                                :title="
                                                    player.shortPositions?.join(
                                                        ', ',
                                                    ) ||
                                                    player.position ||
                                                    '-'
                                                "
                                            >
                                                {{
                                                    player.shortPositions?.join(
                                                        ", ",
                                                    ) ||
                                                    player.position ||
                                                    "-"
                                                }}
                                            </q-item-label>
                                        </div>
                                    </div>
                                    <div
                                        class="col-6 col-sm-4 info-item-redesigned"
                                    >
                                        <q-icon
                                            :color="
                                                qInstance.dark.isActive
                                                    ? 'blue-3'
                                                    : 'primary'
                                            "
                                            name="payments"
                                            size="1.1em"
                                            class="q-mr-xs info-icon-redesigned"
                                        />
                                        <div>
                                            <q-item-label
                                                caption
                                                class="info-caption-redesigned"
                                                >Salary</q-item-label
                                            >
                                            <q-item-label
                                                class="info-label-redesigned ellipsis"
                                                :title="formattedWage"
                                                >{{
                                                    formattedWage
                                                }}</q-item-label
                                            >
                                        </div>
                                    </div>
                                    <div
                                        class="col-6 col-sm-4 info-item-redesigned"
                                    >
                                        <q-icon
                                            :color="
                                                qInstance.dark.isActive
                                                    ? 'blue-3'
                                                    : 'primary'
                                            "
                                            name="comment"
                                            size="1.1em"
                                            class="q-mr-xs info-icon-redesigned"
                                        />
                                        <div>
                                            <q-item-label
                                                caption
                                                class="info-caption-redesigned"
                                                >Media</q-item-label
                                            >
                                            <q-item-label
                                                class="info-label-redesigned ellipsis"
                                                :title="
                                                    player.media_handling || '-'
                                                "
                                                >{{
                                                    player.media_handling || "-"
                                                }}</q-item-label
                                            >
                                        </div>
                                    </div>
                                    <div
                                        class="col-6 col-sm-4 info-item-redesigned"
                                    >
                                        <q-icon
                                            :color="
                                                qInstance.dark.isActive
                                                    ? 'blue-3'
                                                    : 'primary'
                                            "
                                            name="psychology"
                                            size="1.1em"
                                            class="q-mr-xs info-icon-redesigned"
                                        />
                                        <div>
                                            <q-item-label
                                                caption
                                                class="info-caption-redesigned"
                                                >Personality</q-item-label
                                            >
                                            <q-item-label
                                                class="info-label-redesigned ellipsis"
                                                :title="
                                                    player.personality || '-'
                                                "
                                                >{{
                                                    player.personality || "-"
                                                }}</q-item-label
                                            >
                                        </div>
                                    </div>
                                </div>

                                <div class="q-mt-md">
                                    <div
                                        class="text-subtitle2 text-center q-mb-xs fifa-title-redesigned"
                                        :class="
                                            qInstance.dark.isActive
                                                ? 'text-grey-4'
                                                : 'text-grey-8'
                                        "
                                    >
                                        FIFA-Style Ratings
                                    </div>
                                    <div
                                        class="row q-col-gutter-xs justify-center fifa-stats-grid-redesigned"
                                    >
                                        <div
                                            v-for="stat in fifaStatsToDisplay"
                                            :key="stat.name"
                                            class="col-4 col-sm-2"
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
                                            </q-card>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>
                        <div
                            class="text-h5 q-mb-sm text-center attributes-section-title"
                        >
                            Player Attributes (1-20 Scale)
                        </div>
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
                                            Role-Specific Ratings (0-100)
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
import { defineComponent, computed, ref, watch, onMounted } from "vue";
import { useQuasar } from "quasar";
import { formatCurrency } from "../utils/currencyUtils"; // Assuming this path is correct

// Player attribute full names mapping
const attributeFullNameMap = {
    Cor: "Corners",
    Cro: "Crossing",
    Dri: "Dribbling",
    Fin: "Finishing",
    Fir: "First Touch",
    Fre: "Free Kick Taking",
    Hea: "Heading",
    Lon: "Long Shots",
    "L Th": "Long Throws",
    Mar: "Marking",
    Pas: "Passing",
    Pen: "Penalty Taking",
    Tck: "Tackling",
    Tec: "Technique",
    Agg: "Aggression",
    Ant: "Anticipation",
    Bra: "Bravery",
    Cmp: "Composure",
    Cnt: "Concentration",
    Dec: "Decisions",
    Det: "Determination",
    Fla: "Flair",
    Ldr: "Leadership",
    OtB: "Off the Ball",
    Pos: "Positioning",
    Tea: "Teamwork",
    Vis: "Vision",
    Wor: "Work Rate",
    Acc: "Acceleration",
    Agi: "Agility",
    Bal: "Balance",
    Jum: "Jumping Reach",
    Nat: "Natural Fitness",
    Pac: "Pace",
    Sta: "Stamina",
    Str: "Strength",
    Aer: "Aerial Reach",
    Cmd: "Command of Area",
    Com: "Communication",
    Ecc: "Eccentricity",
    Han: "Handling",
    Kic: "Kicking",
    "1v1": "One on Ones",
    Pun: "Punching (Tendency)",
    Ref: "Reflexes",
    TRO: "Rushing Out (Tendency)",
    Thr: "Throwing",
};

// Ordered attribute categories
const technicalAttrsOrdered = [
    "Cor",
    "Cro",
    "Dri",
    "Fin",
    "Fir",
    "Fre",
    "Hea",
    "Lon",
    "L Th",
    "Mar",
    "Pas",
    "Pen",
    "Tck",
    "Tec",
];
const mentalAttrsOrdered = [
    "Agg",
    "Ant",
    "Bra",
    "Cmp",
    "Cnt",
    "Dec",
    "Det",
    "Fla",
    "Ldr",
    "OtB",
    "Pos",
    "Tea",
    "Vis",
    "Wor",
];
const physicalAttrsOrdered = [
    "Acc",
    "Agi",
    "Bal",
    "Jum",
    "Nat",
    "Pac",
    "Sta",
    "Str",
];
const goalkeepingAttrsOrdered = [
    "Aer",
    "Cmd",
    "Com",
    "Ecc",
    "Fir",
    "Han",
    "Kic",
    "1v1",
    "Pas",
    "Pun",
    "Ref",
    "TRO",
    "Thr",
];

// Performance statistics mapping (key to display name)
const performanceStatMap = {
    "Asts/90": "Assists per 90",
    "Av Rat": "Average Rating",
    "Blk/90": "Blocks per 90",
    "Ch C/90": "Chances Created per 90",
    "Clr/90": "Clearances per 90",
    "Cr C/90": "Crosses Completed per 90",
    "Drb/90": "Dribbles per 90",
    "xA/90": "Expected Assists per 90",
    "xG/90": "Expected Goals per 90",
    "Gls/90": "Goals per 90",
    "Hdrs W/90": "Headers Won per 90",
    "Int/90": "Interceptions per 90",
    "K Ps/90": "Key Passes per 90",
    "Ps C/90": "Passes Completed per 90",
    "Shot/90": "Shots per 90",
    "Tck/90": "Tackles per 90",
    "Poss Won/90": "Possession Won per 90",
    "ShT/90": "Shots on Target per 90",
    "Pres C/90": "Pressures Completed per 90",
    "Poss Lost/90": "Possession Lost per 90",
    "Pr passes/90": "Progressive Passes per 90",
    "Conv %": "Conversion %",
    "Tck R": "Tackle Ratio %",
    "Pas %": "Pass Completion %",
    "Cr C/A": "Cross Completion %",
};

// UPDATED Categories for Performance Percentiles
const performanceStatCategories = {
    Offensive: ["Gls/90", "xG/90", "Shot/90", "ShT/90", "Conv %", "Drb/90"],
    Passing: [
        "Asts/90",
        "xA/90",
        "Ch C/90",
        "K Ps/90",
        "Ps C/90",
        "Pas %",
        "Pr passes/90",
        "Cr C/90",
        "Cr C/A",
        "Poss Lost/90",
    ],
    Defensive: [
        "Tck/90",
        "Tck R",
        "Int/90",
        "Clr/90",
        "Blk/90",
        "Hdrs W/90",
        "Pres C/90",
        "Poss Won/90",
    ],
};

export default defineComponent({
    name: "PlayerDetailDialog",
    props: {
        player: { type: Object, default: () => null },
        show: { type: Boolean, default: false },
        currencySymbol: { type: String, default: "$" },
    },
    emits: ["close"],
    setup(props) {
        const qInstance = useQuasar();
        const selectedComparisonGroup = ref("Global");
        const flagLoadError = ref(false);

        const handleFlagError = () => {
            flagLoadError.value = true;
        };

        onMounted(() => {
            // Initialization logic if needed
        });

        watch(
            () => props.player,
            (newPlayer) => {
                flagLoadError.value = false;
                if (newPlayer && newPlayer.performancePercentiles) {
                    const availableGroups = Object.keys(
                        newPlayer.performancePercentiles,
                    );
                    if (
                        !availableGroups.includes(selectedComparisonGroup.value)
                    ) {
                        if (availableGroups.includes("Global")) {
                            selectedComparisonGroup.value = "Global";
                        } else if (availableGroups.length > 0) {
                            selectedComparisonGroup.value = availableGroups[0];
                        } else {
                            selectedComparisonGroup.value = "Global";
                        }
                    }
                } else {
                    selectedComparisonGroup.value = "Global";
                }
            },
            { immediate: true, deep: true },
        );

        const isGoalkeeper = computed(() => {
            if (!props.player) return false;
            return (
                props.player.shortPositions?.includes("GK") ||
                props.player.positionGroups?.includes("Goalkeepers") ||
                props.player.parsedPositions?.includes("Goalkeeper")
            );
        });

        const getPlayerAttributesInOrder = (categoryOrderedKeys) => {
            if (!props.player || !props.player.attributes) return [];
            return categoryOrderedKeys.filter((key) =>
                Object.prototype.hasOwnProperty.call(
                    props.player.attributes,
                    key,
                ),
            );
        };

        const attributeCategories = computed(() => ({
            technical: getPlayerAttributesInOrder(technicalAttrsOrdered),
            mental: getPlayerAttributesInOrder(mentalAttrsOrdered),
            physical: getPlayerAttributesInOrder(physicalAttrsOrdered),
            goalkeeping: isGoalkeeper.value
                ? getPlayerAttributesInOrder(goalkeepingAttrsOrdered)
                : [],
        }));

        const fifaStatsToDisplay = computed(() => {
            let orderedStats = [];
            if (isGoalkeeper.value) {
                orderedStats = [
                    { name: "GK", label: "GK" },
                    { name: "PHY", label: "PHY" },
                    { name: "PAS", label: "PAS" },
                    { name: "MEN", label: "MEN" },
                    { name: "DRI", label: "DRI" },
                    { name: "DEF", label: "DEF" },
                ];
            } else {
                orderedStats = [
                    { name: "PHY", label: "PHY" },
                    { name: "SHO", label: "SHO" },
                    { name: "PAS", label: "PAS" },
                    { name: "DRI", label: "DRI" },
                    { name: "DEF", label: "DEF" },
                    { name: "MEN", label: "MEN" },
                ];
            }
            return orderedStats.filter(
                (stat) => props.player && props.player[stat.name] !== undefined,
            );
        });

        const performanceComparisonOptions = computed(() => {
            const options = [];
            if (props.player && props.player.performancePercentiles) {
                const availablePercentileGroups = Object.keys(
                    props.player.performancePercentiles,
                );

                if (availablePercentileGroups.includes("Global")) {
                    options.push({ label: "Overall Dataset", value: "Global" });
                }

                if (props.player.positionGroups) {
                    props.player.positionGroups.forEach((group) => {
                        if (
                            availablePercentileGroups.includes(group) &&
                            group !== "Global" &&
                            !options.some((opt) => opt.value === group)
                        ) {
                            options.push({
                                label: `vs. ${group}`,
                                value: group,
                            });
                        }
                    });
                }
                if (
                    !options.some((opt) => opt.value === "Global") &&
                    availablePercentileGroups.includes("Global")
                ) {
                    options.unshift({
                        label: "Overall Dataset",
                        value: "Global",
                    });
                }

                if (
                    options.length === 0 &&
                    availablePercentileGroups.includes("Global")
                ) {
                    options.push({ label: "Overall Dataset", value: "Global" });
                }
            }
            return options;
        });

        const averageRatingData = computed(() => {
            if (
                !props.player ||
                !props.player.attributes ||
                !props.player.performancePercentiles
            ) {
                return null;
            }
            const groupKey = selectedComparisonGroup.value;
            const percentilesForGroup =
                props.player.performancePercentiles[groupKey];

            if (
                !percentilesForGroup ||
                !Object.prototype.hasOwnProperty.call(
                    props.player.attributes,
                    "Av Rat",
                ) ||
                props.player.attributes["Av Rat"] === "-" ||
                props.player.attributes["Av Rat"] === "" ||
                !Object.prototype.hasOwnProperty.call(
                    percentilesForGroup,
                    "Av Rat",
                )
            ) {
                return null;
            }
            return {
                key: "Av Rat",
                name: performanceStatMap["Av Rat"] || "Average Rating",
                value: props.player.attributes["Av Rat"],
                percentile:
                    percentilesForGroup["Av Rat"] >= 0
                        ? percentilesForGroup["Av Rat"]
                        : null,
            };
        });

        const categorizedPerformanceStats = computed(() => {
            if (
                !props.player ||
                !props.player.attributes ||
                !props.player.performancePercentiles
            ) {
                return {};
            }
            const groupKey = selectedComparisonGroup.value;
            const percentilesForGroup =
                props.player.performancePercentiles[groupKey];

            if (!percentilesForGroup) {
                return {};
            }

            const result = {};
            const categoryOrder = ["Offensive", "Passing", "Defensive"];

            categoryOrder.forEach((categoryName) => {
                if (performanceStatCategories[categoryName]) {
                    const statsInCategory = [];
                    performanceStatCategories[categoryName].forEach(
                        (statKey) => {
                            const hasRawAttribute =
                                Object.prototype.hasOwnProperty.call(
                                    props.player.attributes,
                                    statKey,
                                );
                            const rawAttributeValue =
                                props.player.attributes[statKey];
                            const hasPercentile =
                                Object.prototype.hasOwnProperty.call(
                                    percentilesForGroup,
                                    statKey,
                                );
                            const percentileValue =
                                percentilesForGroup[statKey];

                            if (
                                performanceStatMap[statKey] &&
                                hasRawAttribute &&
                                rawAttributeValue !== "-" &&
                                rawAttributeValue !== "" &&
                                hasPercentile
                            ) {
                                statsInCategory.push({
                                    key: statKey,
                                    name: performanceStatMap[statKey],
                                    value: rawAttributeValue,
                                    percentile:
                                        percentileValue >= 0
                                            ? percentileValue
                                            : null,
                                });
                            }
                        },
                    );
                    if (statsInCategory.length > 0) {
                        result[categoryName] = statsInCategory.sort((a, b) =>
                            a.name.localeCompare(b.name),
                        );
                    }
                }
            });
            return result;
        });

        const hasAnyPerformanceData = computed(() => {
            return (
                averageRatingData.value ||
                Object.keys(categorizedPerformanceStats.value).length > 0
            );
        });

        const getUnifiedRatingClass = (value, maxScale) => {
            const numValue = parseInt(value, 10);
            if (
                isNaN(numValue) ||
                value === null ||
                value === undefined ||
                value === "-"
            )
                return "rating-na";
            const percentage = (numValue / maxScale) * 100;
            if (percentage >= 90) return "rating-tier-6";
            if (percentage >= 80) return "rating-tier-5";
            if (percentage >= 70) return "rating-tier-4";
            if (percentage >= 55) return "rating-tier-3";
            if (percentage >= 40) return "rating-tier-2";
            return "rating-tier-1";
        };

        const getBarFillStyle = (percentile) => {
            if (
                percentile === null ||
                percentile === undefined ||
                percentile < 0
            ) {
                return {
                    width: "0%",
                    backgroundColor: "#9e9e9e",
                    height: "12px",
                    borderRadius: "3px",
                };
            }
            const p = Math.max(0, Math.min(100, percentile));
            let backgroundColor;
            if (p <= 10) backgroundColor = "#d32f2f";
            else if (p <= 30) backgroundColor = "#ef6c00";
            else if (p <= 45) backgroundColor = "#fdd835";
            else if (p <= 55) backgroundColor = "#bdbdbd";
            else if (p <= 70) backgroundColor = "#aed581";
            else if (p <= 90) backgroundColor = "#66bb6a";
            else backgroundColor = "#388e3c";
            return {
                width: `${p}%`,
                backgroundColor: backgroundColor,
                height: "12px",
                borderRadius: "3px",
                transition: "width 0.3s ease, background-color 0.3s ease",
            };
        };

        const sortedRoleSpecificOveralls = computed(() => {
            if (props.player && props.player.roleSpecificOveralls) {
                return [...props.player.roleSpecificOveralls].sort(
                    (a, b) => b.score - a.score,
                );
            }
            return [];
        });

        const formattedTransferValue = computed(() => {
            if (!props.player) return "-";
            return formatCurrency(
                props.player.transferValueAmount,
                props.currencySymbol,
                props.player.transfer_value,
            );
        });

        const formattedWage = computed(() => {
            if (!props.player) return "-";
            return formatCurrency(
                props.player.wageAmount,
                props.currencySymbol,
                props.player.wage,
            );
        });

        const currencyIcon = computed(() => {
            switch (props.currencySymbol) {
                case "€":
                    return "euro_symbol";
                case "£":
                    return "currency_pound";
                case "$":
                    return "attach_money";
                default:
                    return "payments";
            }
        });

        return {
            qInstance,
            attributeCategories,
            attributeFullNameMap,
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
        };
    },
});
</script>

<style lang="scss" scoped>
.player-profile-card-redesigned {
    // General container for the redesigned section
}

.player-header-redesigned {
    min-height: 60px;
    align-items: stretch;
}

.player-identity-section-redesigned {
    display: flex;
    align-items: center;
    min-width: 160px;
    flex-grow: 1.5; // Give more space to identity
    flex-basis: 0;
}

.player-flag-container-redesigned {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
}
.player-flag-redesigned {
    border: 1px solid rgba(128, 128, 128, 0.5);
    border-radius: 3px;
    object-fit: cover;
    vertical-align: middle;
}
.player-flag-placeholder-redesigned {
    // Styles for the placeholder icon if needed
}

.player-name-age-redesigned {
    display: flex;
    flex-direction: column;
    justify-content: center;
    overflow: hidden;
    text-align: left;
}
.player-name-text-redesigned {
    font-size: clamp(1rem, 1.8vw, 1.25rem);
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 500;
}
.player-age-badge-redesigned {
    font-size: clamp(0.65rem, 1vw, 0.75rem);
    font-weight: 600;
    padding: 2px 5px;
    margin-top: 2px;
    align-self: flex-start;
}

.player-value-section-redesigned {
    padding: 0 8px; // Increased padding for better spacing
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    min-width: 120px; // Ensure enough base width for value
    flex-grow: 1;
    flex-basis: 0;
}

.value-subtitle-redesigned {
    font-size: clamp(0.6rem, 0.9vw, 0.7rem);
    line-height: 1;
    margin-bottom: 1px;
    white-space: nowrap;
}
.value-text-redesigned {
    font-size: clamp(1rem, 1.8vw, 1.3rem); // Slightly increased value text
    line-height: 1.2; // Allow for slightly more height for ranges
    white-space: normal; // Allow wrapping for ranges
    overflow: hidden; // Still hide overflow if it's too long even with wrapping
    text-overflow: ellipsis;
    width: 100%;
    font-weight: 600; // Bolder value
}

// Overall rating section has been removed from header

.condensed-info-grid-redesigned {
    // Grid for Club, Position, Value etc.
}
.info-item-redesigned {
    display: flex;
    align-items: center;
    padding: 2px 0;
    min-height: auto;
}
.info-icon-redesigned {
}
.info-caption-redesigned {
    font-size: clamp(0.6rem, 0.8vw, 0.65rem);
    line-height: 1.1;
    color: $grey-6;
    .body--dark & {
        color: $grey-5;
    }
}
.info-label-redesigned {
    font-size: clamp(0.7rem, 1vw, 0.8rem);
    line-height: 1.2;
    font-weight: 500;
}

.fifa-title-redesigned {
    font-size: clamp(0.8rem, 1.3vw, 0.95rem);
    font-weight: 500;
}
.fifa-stats-grid-redesigned {
}

.fifa-stat-item-redesigned {
    aspect-ratio: 1 / 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 2px !important; // Adjusted padding
    border-width: 1px;
    overflow: hidden;
    line-height: 1.1; // General line height for the box

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
    font-size: clamp(
        0.65rem,
        1.8vw,
        0.85rem
    ); // Significantly increased label size
    font-weight: 500;
    margin-bottom: -2px; // Pull value closer
    display: block;
}
.fifa-value-redesigned {
    font-size: clamp(
        1.1rem,
        2.8vw,
        1.7rem
    ); // Significantly increased value size
    font-weight: 700; // Bolder value
    display: block;
}

// --- Original Styles ---
.player-detail-dialog-card {
    display: flex;
    flex-direction: column;
    border-radius: 8px;
}

.main-content-section {
    flex-grow: 1;
    padding: 12px;
}

.player-flag {
    border: 1px solid rgba(128, 128, 128, 0.5);
    border-radius: 3px;
    object-fit: cover;
    vertical-align: middle;
    height: auto;
}

.dialog-title {
    font-size: clamp(0.9rem, 1.4vw, 1.1rem);
}
.player-name-header {
    font-size: clamp(0.95rem, 1.5vw, 1.15rem);
}
.player-age-badge {
    font-size: clamp(0.6rem, 0.8vw, 0.7rem);
    padding: 1px 2px;
}

.player-info-list-condensed .q-item {
    padding: 0px 0px;
    min-height: auto;
}
.player-info-list-condensed .condensed-caption {
    font-size: clamp(0.5rem, 0.75vw, 0.65rem);
    line-height: 1.2;
}
.player-info-list-condensed .condensed-label {
    font-size: clamp(0.7rem, 1vw, 0.8rem);
    line-height: 1.3;
}

.overall-title-condensed {
    font-size: clamp(0.85rem, 1.3vw, 1rem);
}
.main-overall-value {
    font-size: clamp(1.5rem, 2.5vw, 2rem);
    padding: 0px 5px;
    margin-bottom: 2px;
}
.fifa-ratings-title-condensed {
    font-size: clamp(0.75rem, 1.2vw, 0.9rem);
}

.fifa-stat-card-condensed {
    min-height: 45px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 2px !important;
}
.fifa-stat-card-condensed .fifa-stat-label {
    font-size: clamp(0.55rem, 0.7vw, 0.65rem);
}
.fifa-stat-card-condensed .attribute-value.text-h6 {
    font-size: clamp(0.8rem, 1.3vw, 1rem);
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

.performance-percentiles-card-left {
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
    background-color: #f0f0f0 !important;
}
.q-card__section.bg-grey-8 {
    background-color: #303030 !important;
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
    background-color: #e0e0e0;
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
    .value-text-redesigned {
        font-size: clamp(0.9rem, 1.7vw, 1.2rem);
    } // Adjusted for SM
    .info-label-redesigned {
        font-size: clamp(0.65rem, 0.9vw, 0.75rem);
    }
    .fifa-label-redesigned {
        font-size: clamp(0.6rem, 1.3vw, 0.7rem);
    } // Adjusted for SM
    .fifa-value-redesigned {
        font-size: clamp(1rem, 2.1vw, 1.4rem);
    } // Adjusted for SM
    .player-age-badge-redesigned {
        font-size: clamp(0.6rem, 0.9vw, 0.7rem);
    }
}

@media (max-width: $breakpoint-xs-max) {
    .main-content-section {
        padding: 4px;
    }
    .dialog-title {
        font-size: 0.85rem;
    }
    .player-name-header {
        font-size: 0.9rem;
    }
    .player-age-badge {
        font-size: 0.55rem;
    }
    .player-info-list-condensed .condensed-caption {
        font-size: 0.5rem;
    }
    .player-info-list-condensed .condensed-label {
        font-size: 0.65rem;
    }
    .overall-title-condensed {
        font-size: 0.8rem;
    }
    .main-overall-value {
        font-size: 1.4rem;
    }
    .fifa-ratings-title-condensed {
        font-size: 0.7rem;
    }
    .fifa-stat-card-condensed .fifa-stat-label {
        font-size: 0.5rem;
    }
    .fifa-stat-card-condensed .attribute-value.text-h6 {
        font-size: 0.75rem;
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

    .player-header-redesigned {
        flex-wrap: wrap;
        .player-identity-section-redesigned {
            flex-basis: 60%;
            margin-bottom: 4px;
            order: 1;
        } // Order 1
        .player-value-section-redesigned {
            flex-basis: 40%;
            order: 2;
            text-align: right;
            align-items: flex-end;
            padding-right: 8px;
        } // Order 2
    }
    .player-name-text-redesigned {
        font-size: clamp(0.85rem, 2.5vw, 1rem);
    }
    .player-age-badge-redesigned {
        font-size: clamp(0.6rem, 1.8vw, 0.7rem);
    }
    .value-text-redesigned {
        font-size: clamp(0.9rem, 2.5vw, 1.1rem);
        text-align: right;
    }
    .value-subtitle-redesigned {
        font-size: clamp(0.6rem, 1.8vw, 0.7rem);
    }

    .info-item-redesigned {
        flex-basis: 50%;
    }
    .info-caption-redesigned {
        font-size: clamp(0.55rem, 1.5vw, 0.6rem);
    }
    .info-label-redesigned {
        font-size: clamp(0.65rem, 1.8vw, 0.7rem);
    }

    .fifa-title-redesigned {
        font-size: clamp(0.75rem, 2vw, 0.85rem);
    }
    .fifa-stat-item-redesigned.col-4 {
        flex-basis: calc(100% / 3);
        max-width: calc(100% / 3);
    }
    .fifa-label-redesigned {
        font-size: clamp(0.55rem, 1.5vw, 0.65rem);
    }
    .fifa-value-redesigned {
        font-size: clamp(0.9rem, 2.5vw, 1.2rem);
    }
}
</style>
