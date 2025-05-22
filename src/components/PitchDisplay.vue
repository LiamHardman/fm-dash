// src/components/PitchDisplay.vue
<template>
    <div class="pitch-container" :class="{ 'dark-mode': $q.dark.isActive }">
        <div class="pitch-background">
            <div class="center-circle"></div>
            <div class="center-line"></div>
            <div class="penalty-area penalty-area-top">
                <div class="goal-area goal-area-top"></div>
                <div class="penalty-spot penalty-spot-top"></div>
                <div class="penalty-arc penalty-arc-top"></div>
            </div>
            <div class="penalty-area penalty-area-bottom">
                <div class="goal-area goal-area-bottom"></div>
                <div class="penalty-spot penalty-spot-bottom"></div>
                <div class="penalty-arc penalty-arc-bottom"></div>
            </div>
            <div class="goal goal-top"></div>
            <div class="goal goal-bottom"></div>
        </div>

        <div
            v-for="(row, rowIndex) in formation"
            :key="`row-${rowIndex}`"
            class="formation-row"
        >
            <div
                v-for="pos in row.positions"
                :key="pos.id"
                class="player-slot"
                :style="getPlayerSlotStyle(row.positions.length)"
                @click="
                    players[pos.id] && $emit('player-click', players[pos.id])
                "
                :title="
                    players[pos.id]
                        ? `${players[pos.id].name} (${players[pos.id].Overall || 'N/A'}) - ${pos.role}`
                        : `Empty - ${pos.role}`
                "
            >
                <div
                    class="player-representation"
                    :class="{ 'has-player': !!players[pos.id] }"
                >
                    <q-avatar
                        v-if="players[pos.id]"
                        size="48px"
                        font-size="18px"
                        :color="getPlayerColor(players[pos.id])"
                        text-color="white"
                        class="player-avatar"
                    >
                        {{ getPlayerInitials(players[pos.id].name) }}
                        <q-badge
                            floating
                            color="black"
                            text-color="white"
                            transparent
                            class="player-overall-badge"
                        >
                            {{ players[pos.id].Overall || "N/A" }}
                        </q-badge>
                    </q-avatar>
                    <q-icon
                        v-else
                        name="person_outline"
                        size="28px"
                        class="empty-slot-icon"
                    />
                </div>
                <div
                    class="position-label"
                    :class="{
                        'dark-text': !$q.dark.isActive && !players[pos.id],
                    }"
                >
                    {{ pos.role }}
                </div>
                <div
                    v-if="players[pos.id]"
                    class="player-name-label ellipsis"
                    :class="{ 'dark-text': !$q.dark.isActive }"
                >
                    {{ players[pos.id].name }}
                </div>
            </div>
        </div>
        <div v-if="formation.length === 0" class="no-formation-message">
            Select a formation to view the pitch layout.
        </div>
    </div>
</template>

<script>
import { useQuasar } from "quasar";

export default {
    name: "PitchDisplay",
    props: {
        formation: {
            // Array of rows, where each row has { count, positions: [{id, role}] }
            type: Array,
            default: () => [],
        },
        players: {
            // Object mapping position ID (e.g., 'GK', 'LCB') to player object
            type: Object,
            default: () => ({}),
        },
    },
    emits: ["player-click"],
    setup() {
        const $q = useQuasar();

        const getPlayerSlotStyle = (numPlayersInRow) => {
            // Adjust width based on number of players in the row to fill space
            // This is a simple approach; more complex layouts might need absolute positioning
            const percentageWidth = 100 / Math.max(1, numPlayersInRow);
            return {
                flex: `1 1 ${percentageWidth}%`, // Allow shrinking but prefer defined width
                maxWidth: `${percentageWidth}%`,
            };
        };

        const getPlayerInitials = (name) => {
            if (!name) return "";
            const parts = name.split(" ");
            if (parts.length > 1) {
                return (
                    parts[0][0].toUpperCase() +
                    parts[parts.length - 1][0].toUpperCase()
                );
            }
            return name.substring(0, 2).toUpperCase();
        };

        const getPlayerColor = (player) => {
            // Basic color coding based on overall or position group
            if (!player || player.Overall === undefined) return "grey-7";
            if (player.Overall >= 85) return "positive";
            if (player.Overall >= 75) return "primary";
            if (player.Overall >= 65) return "accent";
            if (player.positionGroups?.includes("Goalkeepers"))
                return "deep-orange-7";
            if (player.positionGroups?.includes("Defenders")) return "blue-7";
            if (player.positionGroups?.includes("Midfielders")) return "teal-7";
            if (player.positionGroups?.includes("Attackers")) return "red-7";
            return "grey-6";
        };

        return {
            $q,
            getPlayerSlotStyle,
            getPlayerInitials,
            getPlayerColor,
        };
    },
};
</script>

<style lang="scss" scoped>
.pitch-container {
    width: 100%;
    max-width: 600px; // Adjust as needed
    aspect-ratio: 7 / 10; // Typical pitch aspect ratio (height / width)
    background-color: #4caf50; // Pitch green
    border: 2px solid white;
    margin: 20px auto;
    position: relative;
    display: flex;
    flex-direction: column;
    justify-content: space-around; // Distribute rows vertically
    padding: 15px 10px; // Padding for rows from edge
    box-sizing: border-box;
    border-radius: 8px;
    overflow: hidden;

    &.dark-mode {
        background-color: #388e3c; // Slightly darker green for dark mode
        border-color: #bdbdbd; // Lighter border for dark mode
    }
}

.pitch-background {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 0;

    // Pitch Markings
    .center-line {
        position: absolute;
        top: 50%;
        left: 0;
        width: 100%;
        height: 2px;
        background-color: rgba(255, 255, 255, 0.6);
        transform: translateY(-50%);
    }
    .center-circle {
        position: absolute;
        top: 50%;
        left: 50%;
        width: 18%; // Relative to pitch width
        aspect-ratio: 1/1;
        border: 2px solid rgba(255, 255, 255, 0.6);
        border-radius: 50%;
        transform: translate(-50%, -50%);
    }
    .penalty-area {
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
        width: 66%; // Approx 40 yards on a 60 yard wide pitch
        height: 22%; // Approx 18 yards on a 100 yard long pitch
        border: 2px solid rgba(255, 255, 255, 0.6);
        box-sizing: border-box;
    }
    .penalty-area-top {
        top: 0;
        border-top: none;
        border-bottom-left-radius: 6px;
        border-bottom-right-radius: 6px;
    }
    .penalty-area-bottom {
        bottom: 0;
        border-bottom: none;
        border-top-left-radius: 6px;
        border-top-right-radius: 6px;
    }
    .goal-area {
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
        width: 33%; // Approx 20 yards
        height: 8%; // Approx 6 yards
        border: 1px solid rgba(255, 255, 255, 0.5);
        box-sizing: border-box;
    }
    .goal-area-top {
        top: 0;
        border-top: none;
    }
    .goal-area-bottom {
        bottom: 0;
        border-bottom: none;
    }
    .penalty-spot {
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
        width: 1.5%;
        aspect-ratio: 1/1;
        background-color: rgba(255, 255, 255, 0.6);
        border-radius: 50%;
    }
    .penalty-spot-top {
        top: 14%;
    } // Approx 12 yards from goal line
    .penalty-spot-bottom {
        bottom: 14%;
    }

    .penalty-arc {
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
        width: 25%; // Approx 10 yard radius circle part
        aspect-ratio: 2/1; // To make it an arc
        border: 2px solid rgba(255, 255, 255, 0.6);
        border-radius: 50% / 100%;
        box-sizing: border-box;
    }
    .penalty-arc-top {
        top: 22%; // Edge of penalty area
        border-top-color: transparent;
        border-left-color: transparent;
        border-right-color: transparent;
        transform: translateX(-50%) rotate(180deg);
    }
    .penalty-arc-bottom {
        bottom: 22%;
        border-bottom-color: transparent;
        border-left-color: transparent;
        border-right-color: transparent;
    }
    .goal {
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
        width: 12%; // Approx 8 yards
        height: 3%; // Goal depth
        background-color: rgba(255, 255, 255, 0.2);
        border: 2px solid white;
        box-sizing: border-box;
    }
    .goal-top {
        top: -2px;
        border-top: none;
    }
    .goal-bottom {
        bottom: -2px;
        border-bottom: none;
    }
}

.formation-row {
    display: flex;
    justify-content: space-around;
    align-items: center;
    width: 100%;
    position: relative; // For z-index stacking above background
    z-index: 1;
    margin: 5px 0; // Small margin between rows
}

.player-slot {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 2px;
    min-height: 70px; // Ensure slots have some height
    cursor: pointer;
    transition: transform 0.2s ease-in-out;

    &:hover .player-representation.has-player {
        transform: scale(1.1);
    }
}

.player-representation {
    width: 50px; // Fixed size for avatar/icon container
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background-color: rgba(
        255,
        255,
        255,
        0.15
    ); // Slight background for empty slots
    transition: background-color 0.3s;
    margin-bottom: 2px;

    &.has-player {
        background-color: transparent; // No background if player avatar is present
    }
}
.player-avatar {
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}
.player-overall-badge {
    font-size: 0.6rem;
    padding: 1px 3px;
    min-height: 12px;
    line-height: 1;
}

.empty-slot-icon {
    color: rgba(255, 255, 255, 0.5);
    .dark-mode & {
        color: rgba(0, 0, 0, 0.4);
    }
}

.position-label {
    font-size: 0.65rem;
    font-weight: bold;
    color: rgba(255, 255, 255, 0.9);
    margin-top: 1px;
    line-height: 1;

    &.dark-text {
        // For light mode pitch, if player is not present
        color: rgba(0, 0, 0, 0.7);
    }
    .dark-mode & {
        // Ensure always light text on dark pitch
        color: rgba(255, 255, 255, 0.9);
    }
}

.player-name-label {
    font-size: 0.6rem;
    color: rgba(255, 255, 255, 0.8);
    margin-top: 1px;
    line-height: 1.1;
    max-width: 60px; // Prevent long names from breaking layout

    &.dark-text {
        color: rgba(0, 0, 0, 0.6);
    }
    .dark-mode & {
        color: rgba(255, 255, 255, 0.8);
    }
}
.no-formation-message {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: white;
    font-size: 0.9rem;
    text-align: center;
    padding: 10px;
    background-color: rgba(0, 0, 0, 0.3);
    border-radius: 4px;
    z-index: 2;
    .dark-mode & {
        color: #e0e0e0;
    }
}

.ellipsis {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
