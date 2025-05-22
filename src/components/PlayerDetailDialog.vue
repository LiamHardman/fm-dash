<template>
    <q-dialog :model-value="show" @hide="$emit('close')">
        <q-card
            class="player-detail-dialog-card"
            :class="
                qInstance.dark.isActive
                    ? 'bg-dark text-white'
                    : 'bg-white text-dark'
            "
            style="max-width: 1000px; width: 95vw; max-height: 90vh"
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
                <div class="row q-col-gutter-x-md q-col-gutter-y-xs q-mb-xs">
                    <div class="col-12 col-md-7">
                        <q-card
                            flat
                            bordered
                            :class="
                                qInstance.dark.isActive
                                    ? 'bg-grey-9'
                                    : 'bg-grey-1'
                            "
                        >
                            <q-card-section class="q-pa-xs">
                                <div
                                    class="text-h6 q-mb-none flex items-center player-name-header"
                                >
                                    <img
                                        v-if="player.nationality_iso"
                                        :src="`https://flagcdn.com/w40/${player.nationality_iso.toLowerCase()}.png`"
                                        :alt="player.nationality || 'Flag'"
                                        width="26"
                                        height="auto"
                                        class="player-flag q-mr-xs"
                                        @error="onFlagError"
                                        :title="player.nationality"
                                    />
                                    <q-icon
                                        v-else
                                        :color="
                                            qInstance.dark.isActive
                                                ? 'grey-5'
                                                : 'grey-7'
                                        "
                                        name="flag"
                                        size="1.3em"
                                        class="q-mr-xs"
                                    />
                                    {{ player.name || "-" }}
                                    <q-badge
                                        outline
                                        :color="
                                            qInstance.dark.isActive
                                                ? 'blue-4'
                                                : 'primary'
                                        "
                                        :label="`${player.age || '-'} yrs`"
                                        class="q-ml-sm player-age-badge"
                                    />
                                </div>

                                <q-list
                                    dense
                                    class="rounded-borders player-info-list q-pt-xs"
                                >
                                    <q-item
                                        class="q-py-none q-px-xs min-height-auto"
                                    >
                                        <q-item-section
                                            avatar
                                            style="
                                                min-width: 30px;
                                                padding-right: 8px;
                                            "
                                        >
                                            <q-icon
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="sports_soccer"
                                                size="1.1em"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label caption
                                                >Club</q-item-label
                                            >
                                            <q-item-label class="text-body1">{{
                                                player.club || "-"
                                            }}</q-item-label>
                                        </q-item-section>
                                    </q-item>

                                    <q-item
                                        class="q-py-none q-px-xs min-height-auto"
                                    >
                                        <q-item-section
                                            avatar
                                            style="
                                                min-width: 30px;
                                                padding-right: 8px;
                                            "
                                        >
                                            <q-icon
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="engineering"
                                                size="1.1em"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label caption
                                                >Position(s)</q-item-label
                                            >
                                            <q-item-label
                                                class="text-body1 ellipsis"
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
                                                    player.position || // Fallback to raw position string
                                                    "-"
                                                }}
                                            </q-item-label>
                                        </q-item-section>
                                    </q-item>

                                    <q-item
                                        class="q-py-none q-px-xs min-height-auto"
                                    >
                                        <q-item-section
                                            avatar
                                            style="
                                                min-width: 30px;
                                                padding-right: 8px;
                                            "
                                        >
                                            <q-icon
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="comment"
                                                size="1.1em"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label caption
                                                >Media Handling</q-item-label
                                            >
                                            <q-item-label class="text-body1">{{
                                                player.media_handling || "-"
                                            }}</q-item-label>
                                        </q-item-section>
                                    </q-item>

                                    <q-item
                                        class="q-py-none q-px-xs min-height-auto"
                                    >
                                        <q-item-section
                                            avatar
                                            style="
                                                min-width: 30px;
                                                padding-right: 8px;
                                            "
                                        >
                                            <q-icon
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="psychology"
                                                size="1.1em"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label caption
                                                >Personality</q-item-label
                                            >
                                            <q-item-label class="text-body1">{{
                                                player.personality || "-"
                                            }}</q-item-label>
                                        </q-item-section>
                                    </q-item>

                                    <q-item
                                        class="q-py-none q-px-xs min-height-auto"
                                    >
                                        <q-item-section
                                            avatar
                                            style="
                                                min-width: 30px;
                                                padding-right: 8px;
                                            "
                                        >
                                            <q-icon
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                :name="currencyIcon"
                                                size="1.1em"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label caption
                                                >Value</q-item-label
                                            >
                                            <q-item-label class="text-body1">{{
                                                formattedTransferValue
                                            }}</q-item-label>
                                        </q-item-section>
                                        <q-item-section side>
                                            <q-icon
                                                :color="
                                                    qInstance.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="payments"
                                                class="q-mr-xs"
                                                size="1.1em"
                                            />
                                            <div>
                                                <q-item-label caption
                                                    >Salary</q-item-label
                                                >
                                                <q-item-label
                                                    class="text-body1"
                                                    >{{
                                                        formattedWage
                                                    }}</q-item-label
                                                >
                                            </div>
                                        </q-item-section>
                                    </q-item>
                                </q-list>
                            </q-card-section>
                        </q-card>
                    </div>

                    <div class="col-12 col-md-5">
                        <q-card
                            flat
                            bordered
                            :class="
                                qInstance.dark.isActive
                                    ? 'bg-grey-9'
                                    : 'bg-grey-1'
                            "
                            class="full-height"
                        >
                            <q-card-section class="text-center q-pa-xs">
                                <div class="text-h6 q-mb-none overall-title">
                                    Overall (Best Role)
                                </div>
                                <div
                                    class="text-h3 text-weight-bold q-mb-none attribute-value main-overall-value"
                                    :class="
                                        getUnifiedRatingClass(
                                            player.Overall,
                                            100,
                                        )
                                    "
                                >
                                    {{
                                        player.Overall > 0
                                            ? player.Overall
                                            : "N/A"
                                    }}
                                </div>
                                <div
                                    class="text-subtitle1 q-my-xs fifa-ratings-title"
                                >
                                    FIFA-Style Ratings
                                </div>
                                <div class="row q-col-gutter-xs text-center">
                                    <div
                                        v-for="stat in fifaStatsToDisplay"
                                        :key="stat.name"
                                        class="col-6 col-sm-4"
                                    >
                                        <q-card
                                            flat
                                            bordered
                                            :class="
                                                qInstance.dark.isActive
                                                    ? 'bg-grey-8'
                                                    : 'bg-white'
                                            "
                                            class="q-pa-xs rounded-borders full-height fifa-stat-card"
                                        >
                                            <div
                                                class="text-caption text-grey-6 fifa-stat-label"
                                            >
                                                {{ stat.label }}
                                            </div>
                                            <div
                                                :class="
                                                    getUnifiedRatingClass(
                                                        player[stat.name],
                                                        100,
                                                    )
                                                "
                                                class="attribute-value fifa-stat-value text-h6"
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
                            </q-card-section>
                        </q-card>
                    </div>
                </div>

                <div
                    class="text-h5 q-mb-sm text-center attributes-section-title"
                >
                    Player Attributes (1-20 Scale)
                </div>
                <div class="row q-col-gutter-md attribute-columns-container">
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
                                                attributeFullNameMap[attrKey] ||
                                                attrKey
                                            }}</q-item-label
                                        >
                                    </q-item-section>
                                    <q-item-section side>
                                        <span
                                            :class="
                                                getUnifiedRatingClass(
                                                    player.attributes[attrKey],
                                                    20,
                                                )
                                            "
                                            class="attribute-value attribute-score-value"
                                        >
                                            {{
                                                player.attributes[attrKey] !==
                                                undefined
                                                    ? player.attributes[attrKey]
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
                                                attributeFullNameMap[attrKey] ||
                                                attrKey
                                            }}</q-item-label
                                        >
                                    </q-item-section>
                                    <q-item-section side>
                                        <span
                                            :class="
                                                getUnifiedRatingClass(
                                                    player.attributes[attrKey],
                                                    20,
                                                )
                                            "
                                            class="attribute-value attribute-score-value"
                                        >
                                            {{
                                                player.attributes[attrKey] !==
                                                undefined
                                                    ? player.attributes[attrKey]
                                                    : "-"
                                            }}
                                        </span>
                                    </q-item-section>
                                </q-item>
                                <q-item
                                    v-if="!attributeCategories.mental.length"
                                >
                                    <q-item-section
                                        class="text-grey-6 text-center q-py-md"
                                        >No mental attributes.</q-item-section
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
                                                attributeFullNameMap[attrKey] ||
                                                attrKey
                                            }}</q-item-label
                                        >
                                    </q-item-section>
                                    <q-item-section side>
                                        <span
                                            :class="
                                                getUnifiedRatingClass(
                                                    player.attributes[attrKey],
                                                    20,
                                                )
                                            "
                                            class="attribute-value attribute-score-value"
                                        >
                                            {{
                                                player.attributes[attrKey] !==
                                                undefined
                                                    ? player.attributes[attrKey]
                                                    : "-"
                                            }}
                                        </span>
                                    </q-item-section>
                                </q-item>
                                <q-item
                                    v-if="!attributeCategories.physical.length"
                                >
                                    <q-item-section
                                        class="text-grey-6 text-center q-py-md"
                                        >No physical attributes.</q-item-section
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
                                        roleOverall.score === player.Overall
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
                                            :title="roleOverall.roleName"
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

                <div class="q-mt-lg">
                    <div class="row items-center q-mb-sm">
                        <div class="text-h5 attributes-section-title col">
                            Performance Statistics
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
                            class="col-auto q-ml-md"
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
                    </div>
                    <q-card
                        flat
                        bordered
                        :class="[
                            qInstance.dark.isActive ? 'bg-grey-9' : 'bg-grey-1',
                            'rounded-borders',
                        ]"
                        v-if="currentPerformanceStatsToDisplay.length > 0"
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
                                Per 90 & Other Metrics (Percentiles vs.
                                {{ selectedComparisonGroupLabel }})
                            </div>
                        </q-card-section>
                        <q-list separator dense class="attribute-list">
                            <q-item
                                v-for="stat in currentPerformanceStatsToDisplay"
                                :key="stat.key"
                                class="attribute-list-item performance-stat-item"
                            >
                                <q-item-section class="stat-name-section">
                                    <q-item-label
                                        lines="1"
                                        class="attribute-name-label"
                                        :title="stat.name"
                                    >
                                        {{ stat.name }}
                                    </q-item-label>
                                </q-item-section>
                                <q-item-section class="stat-bar-section">
                                    <div class="stat-bar-container">
                                        <div class="stat-bar-track">
                                            <div
                                                class="stat-bar-fill"
                                                :style="
                                                    getBarFillStyle(
                                                        stat.percentile,
                                                    )
                                                "
                                            ></div>
                                        </div>
                                        <span
                                            v-if="
                                                stat.percentile !== null &&
                                                stat.percentile >= 0
                                            "
                                            class="stat-percentile-text"
                                        >
                                            {{ Math.round(stat.percentile) }}
                                        </span>
                                        <span
                                            v-else
                                            class="stat-percentile-text text-caption text-grey-6"
                                            >N/A</span
                                        >
                                    </div>
                                </q-item-section>
                                <q-item-section side class="stat-value-section">
                                    <span
                                        class="attribute-value performance-stat-actual-value"
                                    >
                                        {{
                                            stat.value !== "-"
                                                ? stat.value
                                                : "N/A"
                                        }}
                                    </span>
                                </q-item-section>
                            </q-item>
                        </q-list>
                    </q-card>
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
                </div>
            </q-card-section>

            <q-card-section v-else class="text-center q-pa-xl">
                <q-spinner color="primary" size="3em" />
                <div class="q-mt-md text-grey-7">Loading player data...</div>
            </q-card-section>

            <q-card-actions
                align="right"
                :class="qInstance.dark.isActive ? 'bg-grey-10' : 'bg-grey-2'"
                class="q-pa-md"
            >
                <q-btn
                    label="Close"
                    :color="qInstance.dark.isActive ? 'blue-4' : 'primary'"
                    flat
                    @click="$emit('close')"
                />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { defineComponent, computed, ref, watch, onMounted } from "vue";
import { useQuasar } from "quasar";
import { formatCurrency } from "../utils/currencyUtils";

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

// Performance statistics mapping
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
    "Tck R": "Tackle Ratio",
    "Pas %": "Pass Completion %",
    "Cr C/A": "Cross Completion %",
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

        onMounted(() => {
            if (props.player) {
                // console.log("PlayerDetailDialog Mounted. Player data:", JSON.parse(JSON.stringify(props.player)));
                // console.log("Player PositionGroups:", props.player.positionGroups);
                // console.log("Player PerformancePercentiles Keys:", Object.keys(props.player.performancePercentiles || {}));
            }
        });

        watch(
            () => props.player,
            (newPlayer) => {
                if (newPlayer) {
                    // console.log("PlayerDetailDialog Player Prop Changed. Player data:", JSON.parse(JSON.stringify(newPlayer)));
                    // console.log("Player PositionGroups:", newPlayer.positionGroups);
                    // console.log("Player PerformancePercentiles Keys:", Object.keys(newPlayer.performancePercentiles || {}));

                    const availableGroups = newPlayer.performancePercentiles
                        ? Object.keys(newPlayer.performancePercentiles)
                        : [];
                    if (
                        !availableGroups.includes(selectedComparisonGroup.value)
                    ) {
                        selectedComparisonGroup.value = "Global";
                    } else if (
                        availableGroups.length > 0 &&
                        selectedComparisonGroup.value === null
                    ) {
                        selectedComparisonGroup.value = "Global";
                    }
                } else {
                    selectedComparisonGroup.value = "Global";
                }
            },
            { immediate: true, deep: true },
        );

        const isGoalkeeper = computed(() => {
            if (!props.player) return false;
            // MODIFIED: Check shortPositions for GK as well, as it's more direct
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
                if (props.player.performancePercentiles["Global"]) {
                    options.push({ label: "Overall Dataset", value: "Global" });
                }
                if (props.player.positionGroups) {
                    props.player.positionGroups.forEach((group) => {
                        if (
                            props.player.performancePercentiles[group] &&
                            group !== "Global"
                        ) {
                            const existingOption = options.find(
                                (opt) => opt.value === group,
                            );
                            if (!existingOption) {
                                options.push({
                                    label: `vs. ${group}`,
                                    value: group,
                                });
                            }
                        }
                    });
                }
            }
            if (options.length === 0 && props.player) {
                options.push({ label: "Overall Dataset", value: "Global" });
            }
            // console.log("DEBUG: performanceComparisonOptions generated:", options);
            return options;
        });

        const selectedComparisonGroupLabel = computed(() => {
            const selectedOpt = performanceComparisonOptions.value.find(
                (opt) => opt.value === selectedComparisonGroup.value,
            );
            return selectedOpt ? selectedOpt.label : "Selected Group";
        });

        const currentPerformanceStatsToDisplay = computed(() => {
            if (
                !props.player ||
                !props.player.attributes ||
                !props.player.performancePercentiles
            ) {
                return [];
            }
            const groupKey = selectedComparisonGroup.value;
            const percentilesForGroup =
                props.player.performancePercentiles[groupKey];

            if (!percentilesForGroup) {
                return [];
            }

            return Object.keys(percentilesForGroup)
                .filter(
                    (key) =>
                        performanceStatMap[key] &&
                        Object.prototype.hasOwnProperty.call(
                            props.player.attributes,
                            key,
                        ) &&
                        props.player.attributes[key] !== "-" &&
                        props.player.attributes[key] !== "",
                )
                .map((key) => ({
                    key: key,
                    name: performanceStatMap[key],
                    value: props.player.attributes[key],
                    percentile:
                        percentilesForGroup[key] >= 0
                            ? percentilesForGroup[key]
                            : null,
                }))
                .sort((a, b) => a.name.localeCompare(b.name));
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
            else if (p <= 55) backgroundColor = "#9e9e9e";
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

        const onFlagError = (event) => {
            if (event.target) event.target.style.display = "none";
            const placeholder = event.target.nextElementSibling;
            if (placeholder && placeholder.classList.contains("q-icon")) {
                placeholder.style.display = "inline-flex";
            }
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
            currentPerformanceStatsToDisplay,
            onFlagError,
            sortedRoleSpecificOveralls,
            isGoalkeeper,
            formattedTransferValue,
            formattedWage,
            currencyIcon,
            selectedComparisonGroup,
            performanceComparisonOptions,
            selectedComparisonGroupLabel,
        };
    },
});
</script>

<style lang="scss" scoped>
.player-detail-dialog-card {
    display: flex;
    flex-direction: column;
    border-radius: 8px;
}

.main-content-section {
    flex-grow: 1;
    padding: 6px;
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
    font-size: clamp(0.95rem, 1.6vw, 1.2rem);
}
.player-age-badge {
    font-size: clamp(0.6rem, 0.8vw, 0.7rem);
    padding: 1px 2px;
}

.player-info-list .q-item__label {
    font-size: clamp(0.7rem, 1.1vw, 0.85rem);
}
.player-info-list .q-item__label--caption {
    font-size: clamp(0.55rem, 0.8vw, 0.7rem);
}
.player-info-list .text-body1 {
    font-size: clamp(0.75rem, 1.2vw, 0.9rem);
}

.overall-title {
    font-size: clamp(0.95rem, 1.6vw, 1.2rem);
}
.main-overall-value {
    font-size: clamp(1.5rem, 2.5vw, 2rem);
    padding: 2px 5px;
}
.fifa-ratings-title {
    font-size: clamp(0.85rem, 1.4vw, 1rem);
}
.fifa-stat-label {
    font-size: clamp(0.6rem, 0.8vw, 0.7rem);
}
.fifa-stat-card .attribute-value.text-h6 {
    font-size: clamp(0.85rem, 1.4vw, 1.05rem);
}

.attributes-section-title {
    font-size: clamp(1.05rem, 1.8vw, 1.3rem);
}
.attribute-category-header .text-subtitle1 {
    font-size: clamp(0.85rem, 1.4vw, 1rem);
}

.attribute-list-item .attribute-name-label {
    font-size: clamp(0.7rem, 1.1vw, 0.85rem);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 180px; /* Adjust as needed */
}
.attribute-list-item .attribute-score-value {
    font-size: clamp(0.75rem, 1.2vw, 0.9rem);
}
.performance-stat-actual-value {
    font-size: clamp(0.75rem, 1.1vw, 0.85rem);
    min-width: 40px;
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
    max-height: 15vh; /* Adjust as needed */
}

.fifa-stat-card {
    min-height: 50px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 1px !important;
}

.best-role-highlight {
    border-left: 4px solid $positive;
    .body--dark & {
        border-left: 4px solid lighten($positive, 15%);
        background-color: rgba(lighten($positive, 15%), 0.1) !important;
    }
    .body--light & {
        background-color: rgba($positive, 0.1) !important;
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
.player-info-list .q-item {
    padding: 2px 8px;
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
        flex-basis: 40%;
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
        flex-basis: 15%;
        flex-grow: 0;
        flex-shrink: 0;
        text-align: right;
        padding-left: 8px;
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
    margin-right: 8px;
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
    font-size: 0.7rem;
    min-width: 25px;
    text-align: right;
    .body--dark & {
        color: $grey-5;
    }
    .body--light & {
        color: $grey-7;
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
    .player-info-list .q-item__label {
        font-size: 0.7rem;
    }
    .player-info-list .q-item__label--caption {
        font-size: 0.55rem;
    }
    .player-info-list .text-body1 {
        font-size: 0.75rem;
    }
    .overall-title {
        font-size: 0.9rem;
    }
    .main-overall-value {
        font-size: 1.4rem;
    }
    .fifa-ratings-title {
        font-size: 0.8rem;
    }
    .fifa-stat-label {
        font-size: 0.55rem;
    }
    .fifa-stat-card .attribute-value.text-h6 {
        font-size: 0.8rem;
    }
    .attributes-section-title {
        font-size: 1rem;
    }
    .attribute-category-header .text-subtitle1 {
        font-size: 0.8rem;
    }
    .attribute-list-item .attribute-name-label {
        font-size: 0.7rem;
        max-width: 120px; /* Adjust for smaller screens */
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
        overflow-y: auto; /* Allow scroll on mobile if content overflows */
        max-height: 30vh; /* Example max height */
    }
    .role-ratings-card .role-specific-ratings-list {
        max-height: 12vh; /* Adjust for mobile */
    }

    .performance-stat-item {
        .stat-name-section {
            flex-basis: 35%;
            font-size: 0.65rem;
        }
        .stat-value-section {
            flex-basis: 20%;
            font-size: 0.65rem;
        }
    }
}
</style>
