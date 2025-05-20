<template>
    <q-dialog :model-value="show" @hide="$emit('close')" full-width persistent>
        <q-card style="max-width: 900px; width: 100%">
            <q-bar class="bg-primary text-white">
                <q-icon name="person" />
                <div>{{ player.name }} - Detailed View</div>
                <q-space />
                <q-btn dense flat icon="close" @click="$emit('close')">
                    <q-tooltip class="bg-white text-primary">Close</q-tooltip>
                </q-btn>
            </q-bar>

            <q-card-section v-if="player">
                <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-md-6">
                        <q-list bordered separator>
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
                                <q-card flat bordered class="q-pa-sm">
                                    <div class="text-caption text-grey-7">
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

                <q-separator class="q-my-lg" />

                <div class="text-h6 q-mb-md text-center">
                    Individual Attributes (0-20)
                </div>
                <div class="row q-col-gutter-md">
                    <div class="col-12 col-md-4">
                        <q-card flat bordered>
                            <q-card-section class="bg-grey-2">
                                <div class="text-subtitle1 text-weight-medium">
                                    Technical
                                </div>
                            </q-card-section>
                            <q-list separator dense>
                                <q-item
                                    v-for="attrKey in attributeCategories.technical"
                                    :key="attrKey"
                                >
                                    <q-item-section>{{
                                        attributeFullNameMap[attrKey] || attrKey
                                    }}</q-item-section>
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
                                    ><q-item-section class="text-grey-6"
                                        >No technical attributes
                                        found.</q-item-section
                                    ></q-item
                                >
                            </q-list>
                        </q-card>
                    </div>

                    <div class="col-12 col-md-4">
                        <q-card flat bordered>
                            <q-card-section class="bg-grey-2">
                                <div class="text-subtitle1 text-weight-medium">
                                    Mental
                                </div>
                            </q-card-section>
                            <q-list separator dense>
                                <q-item
                                    v-for="attrKey in attributeCategories.mental"
                                    :key="attrKey"
                                >
                                    <q-item-section>{{
                                        attributeFullNameMap[attrKey] || attrKey
                                    }}</q-item-section>
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
                                    ><q-item-section class="text-grey-6"
                                        >No mental attributes
                                        found.</q-item-section
                                    ></q-item
                                >
                            </q-list>
                        </q-card>
                    </div>

                    <div class="col-12 col-md-4">
                        <q-card flat bordered>
                            <q-card-section class="bg-grey-2">
                                <div class="text-subtitle1 text-weight-medium">
                                    Physical
                                </div>
                            </q-card-section>
                            <q-list separator dense>
                                <q-item
                                    v-for="attrKey in attributeCategories.physical"
                                    :key="attrKey"
                                >
                                    <q-item-section>{{
                                        attributeFullNameMap[attrKey] || attrKey
                                    }}</q-item-section>
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
                                    ><q-item-section class="text-grey-6"
                                        >No physical attributes
                                        found.</q-item-section
                                    ></q-item
                                >
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

// --- START: Attribute Definitions ---
// Map for short codes to full attribute names
const attributeFullNameMap = {
    // Technical
    Cor: "Corners",
    Cro: "Crossing",
    Dri: "Dribbling",
    Fin: "Finishing",
    Fir: "First Touch",
    Fre: "Free Kick Taking",
    Hea: "Heading",
    Lon: "Long Shots",
    "L Th": "Long Throws", // Key with space, correctly quoted
    Mar: "Marking",
    Pas: "Passing",
    Pen: "Penalty Taking",
    Tck: "Tackling",
    Tec: "Technique",

    // Mental
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

    // Physical
    Acc: "Acceleration",
    Agi: "Agility",
    Bal: "Balance",
    Jum: "Jumping Reach",
    Nat: "Natural Fitness",
    Pac: "Pace",
    Sta: "Stamina",
    Str: "Strength",
};

// Define the desired order of attributes within each category
const technicalAttrsOrdered = [
    "Cor",
    "Cro",
    "Dri",
    "Fin",
    "Fir",
    "Fre",
    "Hea",
    "Lon",
    "L Th", // "L Th" is correctly placed here
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
// --- END: Attribute Definitions ---

export default defineComponent({
    name: "PlayerDetailDialog",
    props: {
        player: {
            type: Object,
            default: () => null,
        },
        show: {
            type: Boolean,
            default: false,
        },
    },
    emits: ["close"],

    setup(props) {
        // This function filters the predefined ordered list of attribute keys
        // to include only those that actually exist in the current player's attributes.
        const getPlayerAttributesInOrder = (categoryOrderedKeys) => {
            if (!props.player || !props.player.attributes) {
                // console.warn("Player data or attributes missing for getPlayerAttributesInOrder");
                return [];
            }
            // For debugging: Log the keys available in player.attributes
            // if (categoryOrderedKeys === technicalAttrsOrdered) { // Log only for technical to reduce spam
            //   console.log("Available player attributes keys:", Object.keys(props.player.attributes));
            //   console.log("Checking for 'L Th' in player attributes:", props.player.attributes.hasOwnProperty('L Th'));
            // }
            return categoryOrderedKeys.filter((key) =>
                props.player.attributes.hasOwnProperty(key),
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

        return {
            attributeCategories,
            attributeFullNameMap, // Expose the map to the template
            getAttributeClass,
            getFifaStatClass,
            fifaStatsOrder,
        };
    },
});
</script>

<style scoped>
.q-dialog .q-card {
    border-radius: 8px;
}
.q-bar {
    border-top-left-radius: 8px;
    border-top-right-radius: 8px;
}

.attribute-value {
    display: inline-block;
    min-width: 30px;
    text-align: center;
    font-weight: 600;
    padding: 2px 5px;
    border-radius: 3px;
    font-size: 0.85em;
}

/* FM Style Attributes (0-20) */
.attribute-excellent-fm {
    background-color: #20c997;
    color: white;
}
.attribute-very-good-fm {
    background-color: #4dabf7;
    color: white;
}
.attribute-good-fm {
    background-color: #82c91e;
    color: #212529;
}
.attribute-average-fm {
    background-color: #fab005;
    color: #212529;
}
.attribute-poor-fm {
    background-color: #ff922b;
    color: #212529;
}
.attribute-very-poor-fm {
    background-color: #fa5252;
    color: white;
}

/* FIFA Style Attributes (0-100) */
.fifa-stat-value {
    font-size: 1.1em;
    padding: 4px 8px;
}
.attribute-elite {
    background-color: #9c27b0;
    color: white;
}
.attribute-excellent {
    background-color: #20c997;
    color: white;
}
.attribute-very-good {
    background-color: #4dabf7;
    color: white;
}
.attribute-good {
    background-color: #82c91e;
    color: #212529;
}
.attribute-average {
    background-color: #ffc107;
    color: #212529;
}
.attribute-below-average {
    background-color: #fab005;
    color: #212529;
}
.attribute-poor {
    background-color: #ff922b;
    color: #212529;
}
.attribute-very-poor {
    background-color: #fa5252;
    color: white;
}

.attribute-na {
    background-color: #e9ecef;
    color: #868e96;
}

.q-list--dense .q-item {
    padding: 6px 12px;
}
.q-item__section--side {
    padding-right: 0;
}
</style>
