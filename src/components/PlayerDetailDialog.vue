<template>
    <q-dialog :model-value="show" @hide="$emit('close')" full-width persistent>
        <q-card
            class="player-detail-dialog-card"
            style="max-width: 950px; width: 100%"
        >
            <q-bar class="bg-primary text-white">
                <q-icon name="person" class="q-mr-sm" />
                <div>{{ player.name }} - Detailed View</div>
                <q-space />
                <q-btn dense flat icon="close" @click="$emit('close')">
                    <q-tooltip class="bg-white text-primary">Close</q-tooltip>
                </q-btn>
            </q-bar>

            <q-card-section v-if="player" class="q-card__section--main-content">
                <div class="row q-col-gutter-x-md q-col-gutter-y-sm q-mb-md">
                    <div class="col-12 col-md-7">
                        <div
                            class="q-pa-sm rounded-borders"
                            style="border: 1px solid #e0e0e0; height: 100%"
                        >
                            <div class="row items-start q-mb-xs">
                                <div class="col-auto q-mr-sm q-pt-xs">
                                    <q-icon
                                        name="badge"
                                        color="grey-7"
                                        size="1.1em"
                                    />
                                </div>
                                <div class="col">
                                    <q-item-label caption class="q-mb-none"
                                        >Name</q-item-label
                                    >
                                    <q-item-label
                                        class="text-weight-medium text-body2"
                                        >{{ player.name || "-" }}</q-item-label
                                    >
                                </div>
                                <div class="col-auto q-ml-md">
                                    <q-item-label caption class="q-mb-none"
                                        >Age</q-item-label
                                    >
                                    <q-item-label class="text-body2">{{
                                        player.age || "-"
                                    }}</q-item-label>
                                </div>
                                <div
                                    class="col-auto q-ml-md row items-center no-wrap"
                                >
                                    <img
                                        v-if="player.nationality_iso"
                                        :src="`https://flagcdn.com/w40/${player.nationality_iso.toLowerCase()}.png`"
                                        :alt="player.nationality || 'Flag'"
                                        width="26"
                                        class="player-flag q-mr-xs"
                                        @error="onFlagError"
                                        :title="player.nationality"
                                    />
                                    <q-icon
                                        v-else
                                        color="grey-7"
                                        name="flag"
                                        size="1.1em"
                                        class="q-mr-xs"
                                    />
                                    <q-item-label class="text-body2">{{
                                        player.nationality || "-"
                                    }}</q-item-label>
                                </div>
                            </div>
                            <q-separator spaced="xs" />
                            <div class="row items-start q-mb-xs">
                                <div class="col-auto q-mr-sm q-pt-xs">
                                    <q-icon
                                        name="sports_soccer"
                                        color="grey-7"
                                        size="1.1em"
                                    />
                                </div>
                                <div class="col">
                                    <q-item-label caption class="q-mb-none"
                                        >Club</q-item-label
                                    >
                                    <q-item-label class="text-body2">{{
                                        player.club || "-"
                                    }}</q-item-label>
                                </div>
                                <div class="col-auto q-ml-md">
                                    <q-item-label caption class="q-mb-none"
                                        >Position(s)</q-item-label
                                    >
                                    <q-item-label
                                        class="text-body2 ellipsis"
                                        style="max-width: 150px"
                                        :title="
                                            player.parsedPositions?.join(
                                                ', ',
                                            ) ||
                                            player.position ||
                                            '-'
                                        "
                                    >
                                        {{
                                            player.parsedPositions?.join(
                                                ", ",
                                            ) ||
                                            player.position ||
                                            "-"
                                        }}
                                    </q-item-label>
                                </div>
                            </div>
                            <q-separator spaced="xs" />
                            <div class="row items-start q-mb-xs">
                                <div class="col-auto q-mr-sm q-pt-xs">
                                    <q-icon
                                        name="comment"
                                        color="grey-7"
                                        size="1.1em"
                                    />
                                </div>
                                <div class="col">
                                    <q-item-label caption class="q-mb-none"
                                        >Media Handling</q-item-label
                                    >
                                    <q-item-label class="text-body2">{{
                                        player.media_handling || "-"
                                    }}</q-item-label>
                                </div>
                                <div class="col-auto q-mr-sm q-pt-xs q-ml-md">
                                    <q-icon
                                        name="psychology"
                                        color="grey-7"
                                        size="1.1em"
                                    />
                                </div>
                                <div class="col">
                                    <q-item-label caption class="q-mb-none"
                                        >Personality</q-item-label
                                    >
                                    <q-item-label class="text-body2">{{
                                        player.personality || "-"
                                    }}</q-item-label>
                                </div>
                            </div>
                            <q-separator spaced="xs" />
                            <div class="row items-start q-mb-xs">
                                <div class="col-auto q-mr-sm q-pt-xs">
                                    <q-icon
                                        name="euro_symbol"
                                        color="grey-7"
                                        size="1.1em"
                                    />
                                </div>
                                <div class="col">
                                    <q-item-label caption class="q-mb-none"
                                        >Value</q-item-label
                                    >
                                    <q-item-label class="text-body2">{{
                                        player.transfer_value || "-"
                                    }}</q-item-label>
                                </div>
                                <div
                                    class="col-auto q-ml-md row items-center no-wrap"
                                >
                                    <q-icon
                                        name="payments"
                                        color="grey-7"
                                        size="1.1em"
                                        class="q-mr-xs"
                                    />
                                    <div class="col">
                                        <q-item-label caption class="q-mb-none"
                                            >Salary</q-item-label
                                        >
                                        <q-item-label class="text-body2">{{
                                            player.wage || "-"
                                        }}</q-item-label>
                                    </div>
                                </div>
                            </div>
                            <q-separator spaced="xs" />
                            <div class="row items-start">
                                <div class="col-auto q-mr-sm q-pt-xs">
                                    <q-icon
                                        name="star"
                                        color="grey-7"
                                        size="1.1em"
                                    />
                                </div>
                                <div class="col">
                                    <q-item-label caption class="q-mb-none"
                                        >Overall (Best Role)</q-item-label
                                    >
                                    <q-item-label
                                        class="text-weight-bold text-subtitle1"
                                        :class="
                                            getFifaStatClass(player.Overall)
                                        "
                                    >
                                        {{ player.Overall || "N/A" }}
                                    </q-item-label>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="col-12 col-md-5">
                        <div
                            class="text-subtitle1 q-mb-xs text-center text-weight-medium"
                        >
                            FIFA-Style Ratings
                        </div>
                        <div class="row q-col-gutter-xs text-center">
                            <div
                                v-for="stat in fifaStatsToDisplay"
                                :key="stat.name"
                                class="col-4"
                            >
                                <q-card
                                    flat
                                    bordered
                                    class="q-pa-xs rounded-borders full-height"
                                >
                                    <div class="text-caption text-grey-8">
                                        {{ stat.label }}
                                    </div>
                                    <div
                                        :class="
                                            getFifaStatClass(player[stat.name])
                                        "
                                        class="attribute-value fifa-stat-value text-subtitle1"
                                    >
                                        {{
                                            player[stat.name] !== undefined
                                                ? player[stat.name]
                                                : "-"
                                        }}
                                    </div>
                                </q-card>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="text-h6 q-mb-md text-center">
                    Player Attributes (0-20 Scale)
                </div>
                <div class="row q-col-gutter-md attribute-columns-row">
                    <div
                        :class="
                            isGoalkeeper ? 'col-12 col-md-3' : 'col-12 col-md-4'
                        "
                    >
                        <q-card
                            flat
                            bordered
                            class="full-height-card rounded-borders"
                        >
                            <q-card-section class="bg-grey-2 q-pa-sm">
                                <div
                                    class="text-subtitle2 text-weight-medium text-center"
                                >
                                    Technical
                                </div>
                            </q-card-section>
                            <q-list separator dense class="col scroll-list">
                                <q-item
                                    v-for="attrKey in attributeCategories.technical"
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
                                                getAttributeClass(
                                                    player.attributes[attrKey],
                                                )
                                            "
                                            class="attribute-value"
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
                                    v-if="!attributeCategories.technical.length"
                                >
                                    <q-item-section
                                        class="text-grey-6 text-center q-py-md"
                                        >No technical
                                        attributes.</q-item-section
                                    >
                                </q-item>
                            </q-list>
                        </q-card>
                    </div>
                    <div
                        :class="
                            isGoalkeeper ? 'col-12 col-md-3' : 'col-12 col-md-4'
                        "
                    >
                        <q-card
                            flat
                            bordered
                            class="full-height-card rounded-borders"
                        >
                            <q-card-section class="bg-grey-2 q-pa-sm">
                                <div
                                    class="text-subtitle2 text-weight-medium text-center"
                                >
                                    Mental
                                </div>
                            </q-card-section>
                            <q-list separator dense class="col scroll-list">
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
                                                getAttributeClass(
                                                    player.attributes[attrKey],
                                                )
                                            "
                                            class="attribute-value"
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
                    <div
                        :class="
                            isGoalkeeper ? 'col-12 col-md-3' : 'col-12 col-md-4'
                        "
                    >
                        <q-card flat bordered class="rounded-borders q-mb-md">
                            <q-card-section class="bg-grey-2 q-pa-sm">
                                <div
                                    class="text-subtitle2 text-weight-medium text-center"
                                >
                                    Physical
                                </div>
                            </q-card-section>
                            <q-list separator dense>
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
                                                getAttributeClass(
                                                    player.attributes[attrKey],
                                                )
                                            "
                                            class="attribute-value"
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
                            v-if="isGoalkeeper"
                            flat
                            bordered
                            class="rounded-borders q-mb-md"
                        >
                            <q-card-section class="bg-grey-2 q-pa-sm">
                                <div
                                    class="text-subtitle2 text-weight-medium text-center"
                                >
                                    Goalkeeping
                                </div>
                            </q-card-section>
                            <q-list separator dense>
                                <q-item
                                    v-for="attrKey in attributeCategories.goalkeeping"
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
                                                getAttributeClass(
                                                    player.attributes[attrKey],
                                                )
                                            "
                                            class="attribute-value"
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
                                        !attributeCategories.goalkeeping ||
                                        !attributeCategories.goalkeeping.length
                                    "
                                >
                                    <q-item-section
                                        class="text-grey-6 text-center q-py-md"
                                        >No goalkeeping
                                        attributes.</q-item-section
                                    >
                                </q-item>
                            </q-list>
                        </q-card>
                    </div>
                    <div
                        :class="
                            isGoalkeeper ? 'col-12 col-md-3' : 'col-12 col-md-4'
                        "
                    >
                        <q-card
                            flat
                            bordered
                            class="rounded-borders"
                            v-if="
                                player.roleSpecificOveralls &&
                                player.roleSpecificOveralls.length > 0
                            "
                        >
                            <q-card-section class="bg-grey-2 q-pa-sm">
                                <div
                                    class="text-subtitle2 text-weight-medium text-center"
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
                                        'bg-light-blue-1 best-role-highlight':
                                            roleOverall.score ===
                                            player.Overall,
                                    }"
                                    style="min-height: 40px"
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
                                                getFifaStatClass(
                                                    roleOverall.score,
                                                )
                                            "
                                            class="attribute-value fifa-stat-value"
                                            style="font-size: 0.9em"
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

            <q-card-actions align="right" class="bg-grey-1 q-pa-md">
                <q-btn
                    label="Close"
                    color="primary"
                    flat
                    @click="$emit('close')"
                />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { defineComponent, computed } from "vue";

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
    // New Goalkeeper Attributes
    Aer: "Aerial Reach",
    Cmd: "Command of Area",
    Com: "Communication",
    Ecc: "Eccentricity",
    Han: "Handling",
    Kic: "Kicking",
    "1v1": "One on Ones",
    Ref: "Reflexes",
    TRO: "Tendency To Rush Out",
    Thr: "Throwing",
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
    // NEW
    "Aer",
    "Cmd",
    "Com",
    "Ecc",
    "Han",
    "Kic",
    "1v1",
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
                : [], // NEW
        }));

        const fifaStatsOrderBase = [
            { name: "PHY", label: "PHY" },
            { name: "SHO", label: "SHO" },
            { name: "PAS", label: "PAS" },
            { name: "DRI", label: "DRI" },
            { name: "DEF", label: "DEF" },
            { name: "MEN", label: "MEN" },
        ];

        const fifaStatsToDisplay = computed(() => {
            if (isGoalkeeper.value && props.player?.GK) {
                // For GKs, might want to show GK instead of SHO, or add it.
                // Let's add GK and keep others for now, or replace one.
                // Replacing SHO with GK for this example.
                const gkSpecificStats = [...fifaStatsOrderBase];
                const shoIndex = gkSpecificStats.findIndex(
                    (s) => s.name === "SHO",
                );
                if (shoIndex !== -1) {
                    gkSpecificStats.splice(shoIndex, 1, {
                        name: "GK",
                        label: "GK",
                    });
                } else {
                    gkSpecificStats.push({ name: "GK", label: "GK" });
                }
                return gkSpecificStats;
            }
            return fifaStatsOrderBase;
        });

        const getAttributeClass = (value) => {
            // ... (existing logic)
            if (value === null || value === undefined || value === "-")
                return "attribute-na";
            const numValue =
                typeof value === "number" ? value : parseInt(value, 10);
            if (isNaN(numValue)) return "attribute-na";
            if (numValue >= 18) return "attribute-excellent-fm";
            if (numValue >= 15) return "attribute-very-good-fm";
            if (numValue >= 12) return "attribute-good-fm";
            if (numValue >= 9) return "attribute-average-fm";
            if (numValue >= 6) return "attribute-poor-fm";
            return "attribute-very-poor-fm";
        };

        const getFifaStatClass = (value) => {
            // ... (existing logic)
            if (value === null || value === undefined || value === "-")
                return "attribute-na";
            const numValue =
                typeof value === "number" ? value : parseInt(value, 10);
            if (isNaN(numValue)) return "attribute-na";
            if (numValue >= 90) return "attribute-elite";
            if (numValue >= 80) return "attribute-excellent";
            if (numValue >= 70) return "attribute-very-good";
            if (numValue >= 60) return "attribute-good";
            if (numValue >= 50) return "attribute-average";
            if (numValue >= 40) return "attribute-below-average";
            if (numValue >= 30) return "attribute-poor";
            return "attribute-very-poor";
        };

        const onFlagError = (event) => {
            if (event.target) event.target.style.display = "none";
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
            attributeCategories,
            attributeFullNameMap,
            getAttributeClass,
            getFifaStatClass,
            fifaStatsToDisplay, // Use this instead of fifaStatsOrder
            onFlagError,
            sortedRoleSpecificOveralls,
            isGoalkeeper, // Expose to template
        };
    },
});
</script>

<style scoped>
/* General Dialog and Card Styling */
.player-detail-dialog-card {
    max-height: calc(100vh - 48px);
    display: flex;
    flex-direction: column;
    border-radius: 10px;
}

.player-detail-dialog-card > .q-bar {
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
}

.player-detail-dialog-card > .q-card__section--main-content {
    flex-grow: 1;
    overflow-y: auto;
    padding: 20px;
}

.player-detail-dialog-card > .q-card__actions {
    border-top: 1px solid rgba(0, 0, 0, 0.12);
    flex-shrink: 0;
}

.rounded-borders {
    border-radius: 6px;
}

.player-flag {
    border: 1px solid #ddd;
    border-radius: 3px;
    object-fit: cover;
    vertical-align: middle;
}

.q-item-label.text-body2 {
    font-size: 0.875rem;
    line-height: 1.25;
}
.q-item-label.q-mb-none {
    margin-bottom: 0 !important;
}
.q-pt-xs {
    padding-top: 2px;
}

.attribute-columns-row > .column {
    display: flex;
    flex-direction: column;
}
.attribute-columns-row > .col-12.col-md-3 {
    /* For 4 columns when GK */
    padding-left: 8px;
    padding-right: 8px;
}

.full-height-card {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
}

.full-height-card .scroll-list {
    flex-grow: 1;
    overflow-y: auto;
    min-height: 0;
}

.full-height-card .q-card__section.bg-grey-2 {
    background-color: #f8f9fa !important;
    padding: 10px 12px;
}
.full-height-card .text-subtitle2 {
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #495057;
}

.constrained-scroll-list {
    overflow-y: auto;
}

.role-specific-ratings-list {
    max-height: 180px;
}

.attribute-value {
    display: inline-block;
    min-width: 30px;
    text-align: center;
    font-weight: 600;
    padding: 1px 3px;
    border-radius: 4px;
    font-size: 0.85rem;
    line-height: 1.3;
}

.attribute-excellent-fm {
    color: #1565c0;
}
.attribute-very-good-fm {
    color: #00897b;
}
.attribute-good-fm {
    color: #388e3c;
}
.attribute-average-fm {
    color: #b28e00;
}
.attribute-poor-fm {
    color: #d84315;
}
.attribute-very-poor-fm {
    color: #c62828;
}

.fifa-stat-value {
    font-size: 0.9em;
    padding: 2px 4px;
}
.attribute-elite {
    color: #9c27b0;
}
.attribute-excellent {
    color: #1e88e5;
}
.attribute-very-good {
    color: #00acc1;
}
.attribute-good {
    color: #43a047;
}
.attribute-average {
    color: #b28e00;
}
.attribute-below-average {
    color: #fb8c00;
}
.attribute-poor {
    color: #e53935;
}
.attribute-very-poor {
    color: #d32f2f;
}
.attribute-na {
    color: #757575;
}

.best-role-highlight {
    background-color: #e8f5e9 !important;
    border-left: 3px solid var(--q-positive);
}
.best-role-highlight .q-item__label {
    font-weight: 600;
    color: var(--q-positive);
}

.q-list--dense .q-item,
.constrained-scroll-list .q-item {
    padding: 6px 12px;
    min-height: auto;
}
.q-list--separator > .q-item:not(:first-child):before {
    border-top: 1px solid #eff2f5;
}
.q-list--dense .q-item__section--avatar {
    min-width: 38px;
    padding-right: 10px;
}

.q-item__section--side {
    padding-right: 0;
}

.row.text-center > .col-4 > .q-card.full-height {
    display: flex;
    flex-direction: column;
    justify-content: center;
    min-height: 60px;
}

.q-list--dense .q-item__label--caption {
    font-size: 0.7rem;
    line-height: 1.2;
}
.q-list--dense .q-item__label:not(.q-item__label--caption) {
    font-size: 0.85rem;
    line-height: 1.3;
}

/* Ensure the FIFA stats cards have consistent height */
.row.q-col-gutter-xs.text-center .q-card {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-around; /* Or center, depending on desired alignment */
}
</style>
