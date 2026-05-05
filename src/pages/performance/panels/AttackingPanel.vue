<template>
  <div class="tab-content-layout">
    <q-tabs v-model="plotTab" dense class="text-grey" active-color="primary" indicator-color="primary" align="justify" narrow-indicator>
      <q-tab name="shooting" icon="sports_soccer" label="Shooting" />
    </q-tabs>
    <q-tab-panels v-model="plotTab" animated>
      <q-tab-panel name="shooting">
        <div class="charts-grid">
          <DynamicScatterPlotCard
            v-for="config in attackingShootingCharts"
            :key="config.title"
            v-bind="config"
            :is-dark-mode="isDarkMode"
            :all-players-data="filteredPlayers"
            @player-click="$emit('player-click', $event)"
          />
        </div>
        <div class="stats-grid">
          <StatCard
            v-for="stat in attackingStats"
            :key="stat.key"
            :stat="stat"
            :players="topPlayersByStat[stat.key]"
            @player-click="$emit('player-click', $event)"
          />
        </div>
      </q-tab-panel>
    </q-tab-panels>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import StatCard from '../../../components/StatCard.vue'
import { useDynamicComponents } from '../../../composables/useDynamicComponents.js'

// biome-ignore lint/correctness/noUnusedVariables: used in template
const { DynamicScatterPlotCard } = useDynamicComponents()

const props = defineProps({
  filteredPlayers: { type: Array, required: true },
  isDarkMode: { type: Boolean, default: false },
  topPlayersByStat: { type: Object, required: true },
  attackingStats: { type: Array, required: true },
  attackingCharts: { type: Array, required: true },
})

defineEmits(['player-click'])

// biome-ignore lint/correctness/noUnusedVariables: used in template
const plotTab = ref('shooting')
// biome-ignore lint/correctness/noUnusedVariables: used in template
const attackingShootingCharts = computed(() =>
  props.attackingCharts.filter((c) => c.group === 'shooting')
)
</script>

<style scoped>
.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 2rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

@media (max-width: 1200px) {
  .charts-grid { grid-template-columns: 1fr; }
  .stats-grid { grid-template-columns: 1fr; }
}
</style>

