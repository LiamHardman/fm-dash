<template>
  <div class="tab-content-layout">
    <q-tabs v-model="plotTab" dense class="text-grey" active-color="primary" indicator-color="primary" align="justify" narrow-indicator>
      <q-tab name="duels" icon="sports_martial_arts" label="Duels" />
      <q-tab name="pressing" icon="speed" label="Pressing" />
      <q-tab name="aerial" icon="vertical_align_top" label="Aerial" />
      <q-tab name="workrate" icon="directions_run" label="Work Rate" />
    </q-tabs>
    <q-tab-panels v-model="plotTab" animated>
      <q-tab-panel name="duels">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in duelsCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
        </div>
      </q-tab-panel>
      <q-tab-panel name="pressing">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in pressingCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
        </div>
      </q-tab-panel>
      <q-tab-panel name="aerial">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in aerialCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
        </div>
      </q-tab-panel>
      <q-tab-panel name="workrate">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in workrateCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in defendingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
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
  defendingStats: { type: Array, required: true },
  defendingCharts: { type: Array, required: true },
})

defineEmits(['player-click'])

const plotTab = ref('duels')
const duelsCharts = computed(() => props.defendingCharts.filter((c) => c.group === 'duels'))
const pressingCharts = computed(() => props.defendingCharts.filter((c) => c.group === 'pressing'))
const aerialCharts = computed(() => props.defendingCharts.filter((c) => c.group === 'aerial'))
const workrateCharts = computed(() => props.defendingCharts.filter((c) => c.group === 'workrate'))
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

