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
                <div class="row q-col-gutter-x-lg q-col-gutter-y-md q-mb-lg">
                    <div class="col-12 col-md-6">
                        <q-list bordered separator class="rounded-borders">
                            <q-item>
                                <q-item-section avatar>
                                    <q-icon color="grey-7" name="badge" />
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label caption>Name</q-item-label>
                                    <q-item-label class="text-weight-medium">{{
                                        player.name || "-"
                                    }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item>
                                <q-item-section avatar>
                                    <img
                                        v-if="player.nationality_iso"
                                        :src="`https://flagcdn.com/w40/${player.nationality_iso.toLowerCase()}.png`"
                                        :alt="player.nationality || 'Flag'"
                                        width="30"
                                        class="q-mr-sm player-flag"
                                        @error="onFlagError"
                                        :title="player.nationality"
                                    />
                                    <q-icon v-else color="grey-7" name="flag" />
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label caption
                                        >Nationality</q-item-label
                                    >
                                    <q-item-label>{{
                                        player.nationality || "-"
                                    }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item>
                                <q-item-section avatar>
                                    <q-icon
                                        color="grey-7"
                                        name="sports_soccer"
                                    />
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label caption>Club</q-item-label>
                                    <q-item-label>{{
                                        player.club || "-"
                                    }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item>
                                <q-item-section avatar>
                                    <q-icon color="grey-7" name="engineering" />
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label caption
                                        >Position(s)</q-item-label
                                    >
                                    <q-item-label>{{
                                        player.parsedPositions?.join(", ") ||
                                        player.position ||
                                        "-"
                                    }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item>
                                <q-item-section avatar>
                                    <q-icon color="grey-7" name="euro_symbol" />
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label caption
                                        >Transfer Value</q-item-label
                                    >
                                    <q-item-label>{{
                                        player.transfer_value || "-"
                                    }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item>
                                <q-item-section avatar>
                                    <q-icon color="grey-7" name="payments" />
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label caption>Salary</q-item-label>
                                    <q-item-label>{{
                                        player.wage || "-"
                                    }}</q-item-label>
                                </q-item-section>
                            </q-item>
                            <q-item>
                                <q-item-section avatar>
                                    <q-icon color="grey-7" name="star" />
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label caption
                                        >Overall Rating (Best
                                        Role)</q-item-label
                                    >
                                    <q-item-label
                                        class="text-weight-bold text-h6"
                                        :class="
                                            getFifaStatClass(player.Overall)
                                        "
                                    >
                                        {{ player.Overall || "N/A" }}
                                    </q-item-label>
                                </q-item-section>
                            </q-item>
                        </q-list>
                    </div>
                    <div class="col-12 col-md-6">
                        <div
                            class="text-subtitle1 q-mb-sm text-center text-weight-medium"
                        >
                            FIFA-Style Ratings
                        </div>
                        <div class="row q-col-gutter-sm text-center">
                            <div
                                v-for="stat in fifaStatsOrder"
                                :key="stat.name"
                                class="col-4"
                            >
                                <q-card
                                    flat
                                    bordered
                                    class="q-pa-sm rounded-borders full-height"
                                >
                                    <div class="text-caption text-grey-8">
                                        {{ stat.label }}
                                    </div>
                                    <div
                                        :class="
                                            getFifaStatClass(player[stat.name])
                                        "
                                        class="attribute-value fifa-stat-value text-h6"
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
                    <div class="col-12 col-md-4 column">
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

                    <div class="col-12 col-md-4 column">
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

                    <div class="col-12 col-md-4 column q-gutter-y-md">
                        <q-card flat bordered class="rounded-borders">
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

// Attribute mappings and ordered keys
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

export default defineComponent({
    name: "PlayerDetailDialog",
    props: {
        player: { type: Object, default: () => null },
        show: { type: Boolean, default: false },
    },
    emits: ["close"],
    setup(props) {
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
        }));

        const fifaStatsOrder = [
            { name: "PHY", label: "PHY" },
            { name: "SHO", label: "SHO" },
            { name: "PAS", label: "PAS" },
            { name: "DRI", label: "DRI" },
            { name: "DEF", label: "DEF" },
            { name: "MEN", label: "MEN" },
        ];

        const getAttributeClass = (value) => {
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
            event.target.style.display = "none";
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
            fifaStatsOrder,
            onFlagError,
            sortedRoleSpecificOveralls,
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
    padding: 8px 16px;
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
}

/* Attribute Columns Layout */
.attribute-columns-row > .column {
    /* Each of the three main columns */
    display: flex;
    flex-direction: column;
}

/* For Technical and Mental attribute cards and their lists */
.full-height-card {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    min-height: 0; /* Important for flex item children to shrink properly */
}

.full-height-card .scroll-list {
    /* q-list inside Technical/Mental cards */
    flex-grow: 1;
    overflow-y: auto;
    min-height: 0; /* Allows the list to shrink and scroll */
}

/* Styling for attribute card headers */
.q-card .q-card__section.bg-grey-2 {
    padding: 10px 14px;
}
.q-card .text-subtitle2 {
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #555;
}

/* Constrained height and scrolling for lists in the Roles card */
.constrained-scroll-list {
    /* General class for lists that might scroll if needed */
    overflow-y: auto;
}

.role-specific-ratings-list {
    /* Specifically for the roles list */
    max-height: 200px; /* Ensures this list scrolls within a defined height */
}

/* Attribute Value Styling (0-20 scale for FM-style attributes) */
.attribute-value {
    display: inline-block;
    min-width: 32px;
    text-align: center;
    font-weight: 600;
    padding: 3px 6px;
    border-radius: 4px;
    font-size: 0.8rem;
    line-height: 1.4;
}
.attribute-excellent-fm {
    background-color: #1976d2;
    color: white;
}
.attribute-very-good-fm {
    background-color: #26a69a;
    color: white;
}
.attribute-good-fm {
    background-color: #66bb6a;
    color: white;
}
.attribute-average-fm {
    background-color: #ffee58;
    color: #333;
}
.attribute-poor-fm {
    background-color: #ffa726;
    color: white;
}
.attribute-very-poor-fm {
    background-color: #ef5350;
    color: white;
}

/* FIFA Stat Value Styling (0-100 scale) */
.fifa-stat-value {
    font-size: 1.05em;
    padding: 4px 8px;
}
.attribute-elite {
    background-color: #7b1fa2;
    color: white;
}
.attribute-excellent {
    background-color: #1976d2;
    color: white;
}
.attribute-very-good {
    background-color: #26a69a;
    color: white;
}
.attribute-good {
    background-color: #66bb6a;
    color: white;
}
.attribute-average {
    background-color: #ffee58;
    color: #333;
}
.attribute-below-average {
    background-color: #ffca28;
    color: #333;
}
.attribute-poor {
    background-color: #ffa726;
    color: white;
}
.attribute-very-poor {
    background-color: #ef5350;
    color: white;
}

.attribute-na {
    background-color: #bdbdbd;
    color: #fff;
}

/* Highlight for best role */
.best-role-highlight {
    background-color: #e3f2fd !important;
    border-left: 3px solid var(--q-primary);
}
.best-role-highlight .q-item__label {
    font-weight: 500;
}

/* Dense list item padding */
.q-list--dense .q-item,
.constrained-scroll-list .q-item {
    /* Apply to new scrollable lists too */
    padding: 6px 12px;
}
.q-item__section--side {
    padding-right: 0;
}

/* Ensure FIFA stat cards in the grid take full height of their row */
.row.text-center > .col-4 > .q-card.full-height {
    display: flex;
    flex-direction: column;
    justify-content: center;
}
</style>
