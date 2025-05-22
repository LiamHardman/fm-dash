// src/components/PitchDisplay.vue
<template>
    <div
        class="pitch-container"
        :class="{ 'dark-mode': quasar.dark.isActive }"
        @dragover.prevent="handleDragOver"
        @drop.prevent="handleDropOnPitch"
    >
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
                :title="getPlayerSlotTitle(players[pos.id], pos.role)"
                :data-slot-id="pos.id"
                :data-slot-role="pos.role"
                @dragover.prevent
                @drop.prevent="handleDropOnSlot($event, pos.id, pos.role)"
            >
                <div
                    class="player-representation"
                    :class="[
                        { 'has-player': !!players[pos.id] },
                        getPlayerOverallClass(players[pos.id]?.Overall, 100),
                    ]"
                    :draggable="!!players[pos.id]"
                    @dragstart="
                        handleDragStart($event, players[pos.id], pos.id)
                    "
                    @dragend="handleDragEnd"
                >
                    <template v-if="players[pos.id]">
                        <span class="player-overall-display">
                            {{ players[pos.id].Overall || "N/A" }}
                        </span>
                    </template>
                    <q-icon
                        v-else
                        name="add_circle_outline"
                        size="28px"
                        class="empty-slot-icon"
                    />
                </div>
                <div
                    class="position-label"
                    :class="{
                        'dark-text': !quasar.dark.isActive && !players[pos.id],
                    }"
                >
                    {{ pos.role }}
                </div>
                <div
                    v-if="players[pos.id]"
                    class="player-name-label"
                    :class="{ 'dark-text': !quasar.dark.isActive }"
                    :title="players[pos.id].name"
                >
                    {{ players[pos.id].name }}
                </div>
                <div
                    v-if="players[pos.id]"
                    class="player-best-role-label"
                    :class="{ 'dark-text': !quasar.dark.isActive }"
                    :title="
                        getBestRoleForPlayerInSlot(players[pos.id], pos.role)
                    "
                >
                    ({{
                        getBestRoleForPlayerInSlot(players[pos.id], pos.role)
                    }})
                </div>
            </div>
        </div>
        <div v-if="formation.length === 0" class="no-formation-message">
            Select a formation to view the pitch layout.
        </div>

        <div v-if="isDragging" class="drag-overlay">
            <div
                v-for="(row, rowIndex) in formation"
                :key="`overlay-row-${rowIndex}`"
                class="formation-row overlay-row"
            >
                <div
                    v-for="pos in row.positions"
                    :key="`overlay-slot-${pos.id}`"
                    class="drop-zone"
                    :style="getPlayerSlotStyle(row.positions.length)"
                    :data-slot-id="pos.id"
                    :data-slot-role="pos.role"
                    @dragenter.prevent="handleDragEnterSlot"
                    @dragleave.prevent="handleDragLeaveSlot"
                    @drop.prevent="handleDropOnSlot($event, pos.id, pos.role)"
                    @dragover.prevent
                >
                    <span class="drop-zone-role">{{ pos.role }}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import { ref } from "vue";
import { useQuasar } from "quasar";

// Helper map to convert general formation slot roles to base position key prefixes
// used in player.roleSpecificOveralls (e.g., "ST (C)" -> "ST").
const fmSlotRoleToKeyPrefixMap = {
    GK: "GK",
    "D (R)": "DR/L", // Assuming D (R) maps to a general DR/L prefix in roleSpecificOveralls
    "D (L)": "DR/L", // Assuming D (L) maps to a general DR/L prefix
    "D (C)": "DC",
    "WB (R)": "WBR/L", // Assuming WB (R) maps to a general WBR/L prefix
    "WB (L)": "WBR/L", // Assuming WB (L) maps to a general WBR/L prefix
    "DM (C)": "DM",
    "M (R)": "MR/L", // Assuming M (R) maps to a general MR/L prefix
    "M (L)": "MR/L", // Assuming M (L) maps to a general MR/L prefix
    "M (C)": "MC",
    "AM (R)": "AMR/L", // Assuming AM (R) maps to a general AMR/L prefix
    "AM (L)": "AMR/L", // Assuming AM (L) maps to a general AMR/L prefix
    "AM (C)": "AMC",
    "ST (C)": "ST",
    // Add other generic slot roles if they appear in your formations.js and need mapping
    // For example, if a formation slot is just "ST", you might add: "ST": "ST"
};

export default {
    name: "PitchDisplay",
    props: {
        formation: {
            type: Array,
            default: () => [],
        },
        players: {
            // This is the bestTeamPlayers object from TeamViewPage
            // OR bestTeamPlayersForPitch from TeamViewPage (if using squad depth)
            type: Object,
            default: () => ({}),
        },
    },
    emits: ["player-click", "player-moved"],
    setup(props, { emit }) {
        const quasar = useQuasar();
        const isDragging = ref(false);
        const draggedPlayerInfo = ref(null);

        const getPlayerSlotStyle = (numPlayersInRow) => {
            const percentageWidth = 100 / Math.max(1, numPlayersInRow);
            return {
                flex: `1 1 ${percentageWidth}%`,
                maxWidth: `${percentageWidth}%`,
            };
        };

        const getPlayerOverallClass = (overall, maxScale = 100) => {
            const numValue = parseInt(overall, 10);
            if (isNaN(numValue) || overall === null || overall === undefined)
                return "rating-na";
            const percentage = (numValue / maxScale) * 100;
            if (percentage >= 90) return "rating-tier-6";
            if (percentage >= 80) return "rating-tier-5";
            if (percentage >= 70) return "rating-tier-4";
            if (percentage >= 55) return "rating-tier-3";
            if (percentage >= 40) return "rating-tier-2";
            return "rating-tier-1";
        };

        // Revised function to get the best role name, excluding "Generic" roles if a non-Generic alternative exists.
        const getBestRoleForPlayerInSlot = (player, slotRole) => {
            if (
                !player ||
                !player.roleSpecificOveralls ||
                player.roleSpecificOveralls.length === 0 ||
                !slotRole
            ) {
                return slotRole;
            }

            const upperSlotRole = slotRole.toUpperCase();
            const expectedRoleKeyPrefix =
                fmSlotRoleToKeyPrefixMap[upperSlotRole] ||
                upperSlotRole.split(" ")[0];

            const matchingRoles = player.roleSpecificOveralls
                .filter((rso) => {
                    const rsoBasePosition = rso.roleName
                        .split(" - ")[0]
                        .trim()
                        .toUpperCase();
                    return rsoBasePosition === expectedRoleKeyPrefix;
                })
                .sort((a, b) => b.score - a.score); // Sort by score descending

            if (matchingRoles.length === 0) {
                return slotRole; // No specific roles match the slot's general position
            }

            // Try to find the best non-Generic role
            const bestNonGenericRole = matchingRoles.find(
                (rso) => !rso.roleName.toUpperCase().includes("GENERIC"),
            );

            if (bestNonGenericRole) {
                return bestNonGenericRole.roleName;
            }

            return slotRole; // Fallback if no suitable non-Generic role is found.
        };

        const getPlayerSlotTitle = (player, slotRole) => {
            if (player) {
                const bestRoleName = getBestRoleForPlayerInSlot(
                    player,
                    slotRole,
                );
                return `${player.name} (${player.Overall || "N/A"}) - ${bestRoleName}`;
            }
            return `Empty - ${slotRole}`;
        };

        const handleDragStart = (event, player, fromSlotId) => {
            isDragging.value = true;
            draggedPlayerInfo.value = {
                player: props.players[fromSlotId],
                fromSlotId,
            };
            event.dataTransfer.effectAllowed = "move";
            if (player && player.name) {
                event.dataTransfer.setData("text/plain", player.name);
            } else {
                event.dataTransfer.setData("text/plain", "unknown_player");
            }
            document.body.classList.add("grabbing-cursor");
        };

        const handleDragEnd = () => {
            isDragging.value = false;
            draggedPlayerInfo.value = null;
            document.body.classList.remove("grabbing-cursor");
        };

        const handleDragOver = (event) => {
            event.preventDefault();
        };

        const handleDragEnterSlot = (event) => {
            if (event.target.classList.contains("drop-zone")) {
                event.target.classList.add("drop-zone-hover");
            }
        };

        const handleDragLeaveSlot = (event) => {
            if (event.target.classList.contains("drop-zone")) {
                event.target.classList.remove("drop-zone-hover");
            }
        };

        const handleDropOnSlot = (event, toSlotId, toSlotRole) => {
            event.preventDefault();
            let targetElement = event.target;
            if (!targetElement.classList.contains("drop-zone")) {
                targetElement = targetElement.closest(".drop-zone");
            }
            if (targetElement) {
                targetElement.classList.remove("drop-zone-hover");
            }

            if (draggedPlayerInfo.value && draggedPlayerInfo.value.player) {
                const { player, fromSlotId } = draggedPlayerInfo.value;
                if (fromSlotId !== toSlotId) {
                    emit("player-moved", {
                        player,
                        fromSlotId,
                        toSlotId,
                        toSlotRole,
                    });
                }
            }
            handleDragEnd();
        };

        const handleDropOnPitch = (event) => {
            handleDragEnd();
        };

        return {
            quasar,
            isDragging,
            getPlayerSlotStyle,
            getPlayerOverallClass,
            getBestRoleForPlayerInSlot,
            getPlayerSlotTitle,
            handleDragStart,
            handleDragEnd,
            handleDragOver,
            handleDropOnSlot,
            handleDropOnPitch,
            handleDragEnterSlot,
            handleDragLeaveSlot,
        };
    },
};
</script>

<style lang="scss" scoped>
.pitch-container {
    width: 100%;
    max-width: 600px;
    aspect-ratio: 7 / 10;
    background-color: #4caf50;
    border: 2px solid white;
    margin: 20px auto;
    position: relative;
    display: flex;
    flex-direction: column;
    justify-content: space-around;
    padding: 15px 10px;
    box-sizing: border-box;
    border-radius: 8px;
    overflow: hidden;

    &.dark-mode {
        background-color: #388e3c;
        border-color: #bdbdbd;
    }
}

.pitch-background {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 0;

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
        width: 18%;
        aspect-ratio: 1/1;
        border: 2px solid rgba(255, 255, 255, 0.6);
        border-radius: 50%;
        transform: translate(-50%, -50%);
    }
    .penalty-area {
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
        width: 66%;
        height: 22%;
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
        width: 33%;
        height: 8%;
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
    }
    .penalty-spot-bottom {
        bottom: 14%;
    }

    .penalty-arc {
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
        width: 25%;
        aspect-ratio: 2/1;
        border: 2px solid rgba(255, 255, 255, 0.6);
        border-radius: 50% / 100%;
        box-sizing: border-box;
    }
    .penalty-arc-top {
        top: 22%;
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
        width: 12%;
        height: 3%;
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
    position: relative;
    z-index: 1;
    margin: 2px 0;
}

.player-slot {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 1px;
    min-height: 85px; // Increased min-height to accommodate larger text
    cursor: pointer;
    transition: transform 0.2s ease-in-out;
    position: relative;

    &:hover .player-representation.has-player {
        transform: scale(1.05);
    }
}

.player-representation {
    width: 42px;
    height: 42px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background-color: rgba(255, 255, 255, 0.1);
    transition:
        background-color 0.3s,
        transform 0.2s;
    margin-bottom: 1px;
    color: white;
    font-weight: bold;
    font-size: 1rem;
    border: 1px solid rgba(0, 0, 0, 0.2);

    &.has-player {
        // Background color and text color will be set by getPlayerOverallClass (e.g., .rating-tier-X)
    }
    &.dragging-feedback {
        outline: 2px dashed #fff;
        background-color: rgba(255, 255, 255, 0.3);
    }
}

.player-overall-display {
    line-height: 1;
}

.empty-slot-icon {
    color: rgba(255, 255, 255, 0.4);
    .dark-mode & {
        color: rgba(0, 0, 0, 0.3);
    }
}

.position-label {
    font-size: 0.6rem; // Kept this one small as it's just the generic role
    font-weight: bold;
    color: rgba(255, 255, 255, 0.85);
    margin-top: 0px;
    line-height: 1;

    &.dark-text {
        color: rgba(0, 0, 0, 0.65);
    }
    .dark-mode & {
        color: rgba(255, 255, 255, 0.85);
    }
}

.player-name-label {
    font-size: 0.65rem; // Increased font size
    color: rgba(255, 255, 255, 0.85); // Slightly more opaque
    margin-top: 2px; // Adjusted margin
    line-height: 1.1; // Allow slightly more line height for wrapping
    // Removed max-width and ellipsis properties to allow wrapping
    // white-space: nowrap; // Removed
    // overflow: hidden; // Removed
    // text-overflow: ellipsis; // Removed
    width: 90%; // Allow it to take more width of the slot for better wrapping
    word-wrap: break-word; // Ensure long names break and wrap

    &.dark-text {
        color: rgba(0, 0, 0, 0.7);
    }
    .dark-mode & {
        color: rgba(255, 255, 255, 0.85);
    }
}
.player-best-role-label {
    font-size: 0.6rem; // Increased font size
    color: rgba(255, 255, 255, 0.75); // Slightly more opaque
    margin-top: 1px;
    line-height: 1.1; // Allow slightly more line height for wrapping
    font-style: italic;
    // Removed max-width and ellipsis properties to allow wrapping
    // white-space: nowrap; // Removed
    // overflow: hidden; // Removed
    // text-overflow: ellipsis; // Removed
    width: 90%; // Allow it to take more width of the slot for better wrapping
    word-wrap: break-word; // Ensure long role names break and wrap

    &.dark-text {
        color: rgba(0, 0, 0, 0.6);
    }
    .dark-mode & {
        color: rgba(255, 255, 255, 0.75);
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

// .ellipsis class is removed from the template for these labels, so it's no longer needed here for them.
// If you need it elsewhere, it can remain.
// .ellipsis {
// white-space: nowrap;
// overflow: hidden;
// text-overflow: ellipsis;
// }

.drag-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.3);
    z-index: 10;
    display: flex;
    flex-direction: column;
    justify-content: space-around;
    padding: 15px 10px;
    box-sizing: border-box;
}

.overlay-row {
    z-index: 11;
}

.drop-zone {
    border: 2px dashed rgba(255, 255, 255, 0.5);
    border-radius: 8px;
    min-height: 75px; // Keep drop zone size consistent
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s ease;
    box-sizing: border-box;
    padding: 2px;
}
.drop-zone-role {
    color: rgba(255, 255, 255, 0.7);
    font-size: 0.8em;
    font-weight: bold;
    text-shadow: 1px 1px 1px rgba(0, 0, 0, 0.5);
}

.drop-zone-hover {
    background-color: rgba(255, 255, 255, 0.2);
    border-style: solid;
}

/* Global rating tier classes are defined in app.scss */
</style>
