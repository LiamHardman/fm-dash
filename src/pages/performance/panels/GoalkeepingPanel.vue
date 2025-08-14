<template>
  <div class="tab-content-layout goalkeeping">
    <q-tabs v-model="plotTab" dense class="text-grey" active-color="primary" indicator-color="primary" align="justify" narrow-indicator>
      <q-tab name="shotstopping" icon="sports_hockey" label="Shot-Stopping" />
    </q-tabs>
    <q-tab-panels v-model="plotTab" animated>
      <q-tab-panel name="shotstopping">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in shotstoppingCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in goalkeepingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
        </div>
      </q-tab-panel>
    </q-tab-panels>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import StatCard from '../../../components/StatCard.vue'
import { useDynamicComponents } from '../../../composables/useDynamicComponents.js'

const { DynamicScatterPlotCard } = useDynamicComponents()

const props = defineProps({
  filteredPlayers: { type: Array, required: true },
  isDarkMode: { type: Boolean, default: false },
  topPlayersByStat: { type: Object, required: true },
  goalkeepingStats: { type: Array, required: true },
  goalkeepingCharts: { type: Array, required: true },
})

defineEmits(['player-click'])

const plotTab = ref('shotstopping')
const shotstoppingCharts = computed(() =>
  props.goalkeepingCharts.filter((c) => c.group === 'shotstopping')
)
</script>

<style scoped>
.charts-grid {
  display: grid;
  grid-template-columns: 1fr; /* goalkeeping only one chart column */
  gap: 1rem;
  margin-bottom: 2rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

@media (max-width: 1200px) {
  .stats-grid { grid-template-columns: 1fr; }
}
</style>

