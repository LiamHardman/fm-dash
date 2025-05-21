<template>
    <q-dialog :model-value="show" @hide="$emit('close')">
        <q-card
            class="player-detail-dialog-card"
            :class="
                $q.dark.isActive ? 'bg-dark text-white' : 'bg-white text-dark'
            "
            style="max-width: 1000px; width: 95vw; max-height: 90vh"
        >
            <q-bar
                :class="
                    $q.dark.isActive ? 'bg-grey-10' : 'bg-primary text-white'
                "
            >
                <q-icon name="person" class="q-mr-sm" />
                <div class="text-subtitle1">
                    {{ player?.name || "Player" }} - Detailed View
                </div>
                <q-space />
                <q-btn dense flat icon="close" @click="$emit('close')">
                    <q-tooltip
                        :class="
                            $q.dark.isActive
                                ? 'bg-grey-7'
                                : 'bg-white text-primary'
                        "
                        >Close</q-tooltip
                    >
                </q-btn>
            </q-bar>

            <q-card-section v-if="player" class="scroll main-content-section">
                <div class="row q-col-gutter-x-lg q-col-gutter-y-md q-mb-lg">
                    <div class="col-12 col-md-7">
                        <q-card
                            flat
                            bordered
                            :class="
                                $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1'
                            "
                        >
                            <q-card-section>
                                <div class="text-h6 q-mb-sm flex items-center">
                                    <img
                                        v-if="player.nationality_iso"
                                        :src="`https://flagcdn.com/w40/${player.nationality_iso.toLowerCase()}.png`"
                                        :alt="player.nationality || 'Flag'"
                                        width="30"
                                        class="player-flag q-mr-sm"
                                        @error="onFlagError"
                                        :title="player.nationality"
                                    />
                                    <q-icon
                                        v-else
                                        :color="
                                            $q.dark.isActive
                                                ? 'grey-5'
                                                : 'grey-7'
                                        "
                                        name="flag"
                                        size="1.5em"
                                        class="q-mr-sm"
                                    />
                                    {{ player.name || "-" }}
                                    <q-badge
                                        outline
                                        :color="
                                            $q.dark.isActive
                                                ? 'blue-4'
                                                : 'primary'
                                        "
                                        :label="`${player.age || '-'} yrs`"
                                        class="q-ml-md"
                                    />
                                </div>

                                <q-list dense padding class="rounded-borders">
                                    <q-item>
                                        <q-item-section avatar>
                                            <q-icon
                                                :color="
                                                    $q.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="sports_soccer"
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

                                    <q-item>
                                        <q-item-section avatar>
                                            <q-icon
                                                :color="
                                                    $q.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="engineering"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label caption
                                                >Position(s)</q-item-label
                                            >
                                            <q-item-label
                                                class="text-body1 ellipsis"
                                                :title="
                                                    player.parsedPositions?.join(
                                                        ', ',
                                                    ) ||
                                                    player.position ||
                                                    '-'
                                                "
                                            >
                                                {{
                                                    player.position ||
                                                    player.parsedPositions?.join(
                                                        ", ",
                                                    ) ||
                                                    "-"
                                                }}
                                            </q-item-label>
                                        </q-item-section>
                                    </q-item>

                                    <q-item>
                                        <q-item-section avatar>
                                            <q-icon
                                                :color="
                                                    $q.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="comment"
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

                                    <q-item>
                                        <q-item-section avatar>
                                            <q-icon
                                                :color="
                                                    $q.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="psychology"
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

                                    <q-item>
                                        <q-item-section avatar>
                                            <q-icon
                                                :color="
                                                    $q.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="euro_symbol"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label caption
                                                >Value</q-item-label
                                            >
                                            <q-item-label class="text-body1">{{
                                                player.transfer_value || "-"
                                            }}</q-item-label>
                                        </q-item-section>
                                        <q-item-section side>
                                            <q-icon
                                                :color="
                                                    $q.dark.isActive
                                                        ? 'blue-3'
                                                        : 'primary'
                                                "
                                                name="payments"
                                                class="q-mr-xs"
                                            />
                                            <div>
                                                <q-item-label caption
                                                    >Salary</q-item-label
                                                >
                                                <q-item-label
                                                    class="text-body1"
                                                    >{{
                                                        player.wage || "-"
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
                                $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1'
                            "
                            class="full-height"
                        >
                            <q-card-section class="text-center">
                                <div class="text-h6 q-mb-md">
                                    Overall (Best Role)
                                </div>
                                <div
                                    class="text-h3 text-weight-bold q-mb-md attribute-value"
                                    :class="
                                        getUnifiedRatingClass(
                                            player.Overall,
                                            100,
                                        )
                                    "
                                >
                                    {{ player.Overall || "N/A" }}
                                </div>
                                <div class="text-subtitle1 q-mb-sm">
                                    FIFA-Style Ratings
                                </div>
                                <div class="row q-col-gutter-sm text-center">
                                    <div
                                        v-for="stat in fifaStatsToDisplay"
                                        :key="stat.name"
                                        class="col-6 col-sm-4"
                                    >
                                        <q-card
                                            flat
                                            bordered
                                            :class="
                                                $q.dark.isActive
                                                    ? 'bg-grey-8'
                                                    : 'bg-white'
                                            "
                                            class="q-pa-sm rounded-borders full-height fifa-stat-card"
                                        >
                                            <div
                                                class="text-caption text-grey-6"
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

                <div class="text-h5 q-mb-md text-center">
                    Player Attributes (1-20 Scale)
                </div>
                <div class="row q-col-gutter-md attribute-columns-container">
                    <div class="col-12 col-md-4 column">
                        <q-card
                            flat
                            bordered
                            :class="[
                                $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1',
                                'full-height-card',
                                'rounded-borders',
                            ]"
                        >
                            <q-card-section
                                :class="
                                    $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-3'
                                "
                                class="q-pa-sm"
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
                                class="col scroll-list attribute-list"
                            >
                                <q-item
                                    v-for="attrKey in isGoalkeeper
                                        ? attributeCategories.goalkeeping
                                        : attributeCategories.technical"
                                    :key="attrKey"
                                >
                                    <q-item-section>
                                        <q-item-label lines="1">{{
                                            attributeFullNameMap[attrKey] ||
                                            attrKey
                                        }}</q-item-label>
                                    </q-item-section>
                                    <q-item-section side>
                                        <span
                                            :class="
                                                getUnifiedRatingClass(
                                                    player.attributes[attrKey],
                                                    20,
                                                )
                                            "
                                            class="attribute-value text-body1"
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
                                $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1',
                                'full-height-card',
                                'rounded-borders',
                            ]"
                        >
                            <q-card-section
                                :class="
                                    $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-3'
                                "
                                class="q-pa-sm"
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
                                class="col scroll-list attribute-list"
                            >
                                <q-item
                                    v-for="attrKey in attributeCategories.mental"
                                    :key="attrKey"
                                >
                                    <q-item-section>
                                        <q-item-label lines="1">{{
                                            attributeFullNameMap[attrKey] ||
                                            attrKey
                                        }}</q-item-label>
                                    </q-item-section>
                                    <q-item-section side>
                                        <span
                                            :class="
                                                getUnifiedRatingClass(
                                                    player.attributes[attrKey],
                                                    20,
                                                )
                                            "
                                            class="attribute-value text-body1"
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
                                $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1',
                                'rounded-borders',
                                'physical-attributes-card',
                            ]"
                        >
                            <q-card-section
                                :class="
                                    $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-3'
                                "
                                class="q-pa-sm"
                            >
                                <div
                                    class="text-subtitle1 text-weight-medium text-center"
                                >
                                    Physical
                                </div>
                            </q-card-section>
                            <q-list separator dense class="attribute-list">
                                <q-item
                                    v-for="attrKey in attributeCategories.physical"
                                    :key="attrKey"
                                >
                                    <q-item-section>
                                        <q-item-label lines="1">{{
                                            attributeFullNameMap[attrKey] ||
                                            attrKey
                                        }}</q-item-label>
                                    </q-item-section>
                                    <q-item-section side>
                                        <span
                                            :class="
                                                getUnifiedRatingClass(
                                                    player.attributes[attrKey],
                                                    20,
                                                )
                                            "
                                            class="attribute-value text-body1"
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
                                $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1',
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
                                    $q.dark.isActive ? 'bg-grey-8' : 'bg-grey-3'
                                "
                                class="q-pa-sm"
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
                                            ? $q.dark.isActive
                                                ? 'background-color: #2a5270 !important;'
                                                : 'background-color: #e3f2fd !important;'
                                            : ''
                                    "
                                >
                                    <q-item-section>
                                        <q-item-label
                                            lines="1"
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
                                            class="attribute-value fifa-stat-value text-body1"
                                        >
                                            {{ roleOverall.score }}
                                        </span>
                                    </q-item-section>
                                </q-item>
                            </q-list>
                        </q-card>
                    </div>
                </div>
            </q-card-section>

            <q-card-section v-else class="text-center q-pa-xl">
                <q-spinner color="primary" size="3em" />
                <div class="q-mt-md text-grey-7">Loading player data...</div>
            </q-card-section>

            <q-card-actions
                align="right"
                :class="$q.dark.isActive ? 'bg-grey-10' : 'bg-grey-2'"
                class="q-pa-md"
            >
                <q-btn
                    label="Close"
                    :color="$q.dark.isActive ? 'blue-4' : 'primary'"
                    flat
                    @click="$emit('close')"
                />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { defineComponent, computed } from "vue";
import { useQuasar } from "quasar";

// Attribute mappings and ordered keys (unchanged from original)
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
    Ref: "Reflexes",
    TRO: "Rushing Out (Tendency)",
    Thr: "Throwing",
    Pun: "Punching (Tendency)",
};
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

export default defineComponent({
    name: "PlayerDetailDialog",
    props: {
        player: { type: Object, default: () => null },
        show: { type: Boolean, default: false },
    },
    emits: ["close"],
    setup(props) {
        const $q = useQuasar();

        const isGoalkeeper = computed(() => {
            if (!props.player) return false;
            return (
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
                (stat) => props.player?.[stat.name] !== undefined,
            );
        });

        // Unified rating class function
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

        return {
            $q,
            attributeCategories,
            attributeFullNameMap,
            getUnifiedRatingClass,
            fifaStatsToDisplay,
            onFlagError,
            sortedRoleSpecificOveralls,
            isGoalkeeper,
        };
    },
});
</script>

<style lang="scss" scoped>
/* Rating tier styles are now global in app.scss */
/* .attribute-value and .fifa-stat-value are also global or defined in PlayerDataTable */

.player-detail-dialog-card {
    display: flex;
    flex-direction: column;
    border-radius: 8px;
}

.main-content-section {
    flex-grow: 1;
    padding: 12px 16px;
}

.player-flag {
    border: 1px solid rgba(128, 128, 128, 0.5);
    border-radius: 3px;
    object-fit: cover;
    vertical-align: middle;
}

.q-item__label {
    font-size: 1rem;
}
.q-item__label--caption {
    font-size: 0.8rem;
    opacity: 0.8;
}
.text-body1 {
    font-size: 1rem;
}
.text-subtitle1 {
    font-size: 1.1rem;
}
.text-h6 {
    font-size: 1.3rem;
}
.text-h5 {
    font-size: 1.5rem;
}
.text-h3.attribute-value {
    font-size: 2.5rem;
    padding: 5px 10px;
    line-height: 1.2;
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

.full-height-card .attribute-list {
    flex-grow: 1;
    overflow-y: auto;
    min-height: 0;
}

.physical-attributes-card {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    .q-list {
        flex-grow: 1;
        overflow-y: auto;
    }
}

.role-ratings-card .role-specific-ratings-list {
    max-height: 180px;
    flex-shrink: 0;
}

.fifa-stat-card {
    min-height: 70px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    .attribute-value.text-h6 {
        font-size: 1.2rem;
        padding: 3px 5px;
    }
}

.constrained-scroll-list {
    overflow-y: auto;
}

.role-specific-ratings-list {
    max-height: 200px;
}

.best-role-highlight {
    border-left: 4px solid $positive;
    .body--dark & {
        border-left: 4px solid lighten($positive, 15%);
    }
}
.best-role-highlight .q-item__label {
    font-weight: 600;
}

.q-list--dense .q-item,
.constrained-scroll-list .q-item {
    padding: 8px 12px;
    min-height: 44px;
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

@media (max-width: $breakpoint-xs-max) {
    .player-detail-dialog-card {
        width: 98vw;
        max-height: 95vh;
    }
    .main-content-section {
        padding: 8px;
    }
    .text-h3.attribute-value {
        font-size: 2rem;
    }
    .text-h5 {
        font-size: 1.3rem;
    }
    .text-h6 {
        font-size: 1.15rem;
    }
    .q-item__label {
        font-size: 0.9rem;
    }
    .q-item__label--caption {
        font-size: 0.75rem;
    }

    .attribute-columns-container {
        .col-12.col-md-4 {
            flex-basis: 100%;
            max-width: 100%;
            &.column {
                display: flex;
                flex-direction: column;
            }
        }
    }
    .fifa-stat-card .attribute-value.text-h6 {
        font-size: 1.1rem;
    }
    .full-height-card .attribute-list {
        max-height: 250px;
    }
    .role-specific-ratings-list {
        max-height: 150px;
    }
}
</style>
