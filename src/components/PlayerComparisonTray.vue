<template>
  <transition name="tray-slide">
    <div v-if="players.length > 0" class="comparison-tray">
      <div class="tray-inner">
        <div class="tray-left">
          <q-icon name="compare_arrows" size="1.1rem" class="q-mr-xs tray-icon" />
          <span class="tray-label">Compare ({{ players.length }}/{{ MAX }})</span>
          <div class="tray-chips">
            <div
              v-for="(player, idx) in players"
              :key="player.name + player.club"
              class="tray-chip"
              :style="{ borderColor: PLAYER_COLORS[idx], color: PLAYER_COLORS[idx] }"
            >
              <span class="tray-chip-name">{{ player.name }}</span>
              <q-btn
                flat dense round
                icon="close"
                size="xs"
                :style="{ color: PLAYER_COLORS[idx] }"
                @click="$emit('remove-player', player)"
              />
            </div>
          </div>
        </div>

        <div class="tray-right">
          <q-btn
            flat dense
            icon="delete_outline"
            label="Clear all"
            size="sm"
            color="grey-6"
            @click="$emit('clear')"
          />
          <q-btn
            unelevated
            icon="compare_arrows"
            label="Compare"
            color="primary"
            size="sm"
            :disable="players.length < 2"
            @click="$emit('open')"
          >
            <q-tooltip v-if="players.length < 2">Add at least 2 players</q-tooltip>
          </q-btn>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
defineProps({
  players: { type: Array, default: () => [] },
})

defineEmits(['remove-player', 'clear', 'open'])

// biome-ignore lint/correctness/noUnusedVariables: used in template
const MAX = 4
// biome-ignore lint/correctness/noUnusedVariables: used in template
const PLAYER_COLORS = ['#3b82f6', '#ef4444', '#22c55e', '#f97316']
</script>

<style lang="scss" scoped>
.comparison-tray {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 3000;
  padding: 10px 20px;
  background: var(--surface-card);
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.1);

  .body--dark & {
    border-top-color: rgba(255, 255, 255, 0.1);
    box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.3);
  }
}

.tray-inner {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 12px;
}

.tray-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.tray-icon { color: var(--text-secondary); }

.tray-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  white-space: nowrap;
}

.tray-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.tray-chip {
  display: flex;
  align-items: center;
  gap: 2px;
  border: 1.5px solid;
  border-radius: 20px;
  padding: 2px 4px 2px 10px;
  font-size: 12px;
  font-weight: 600;
}

.tray-chip-name {
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tray-right {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-shrink: 0;
}

.tray-slide-enter-active,
.tray-slide-leave-active { transition: transform 0.25s ease, opacity 0.25s ease; }
.tray-slide-enter-from,
.tray-slide-leave-to { transform: translateY(100%); opacity: 0; }
</style>
