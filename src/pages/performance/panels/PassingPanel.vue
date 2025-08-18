<template>
  <div class="tab-content-layout">
    <q-tabs v-model="plotTab" dense class="text-grey" active-color="primary" indicator-color="primary" align="justify" narrow-indicator>
      <q-tab name="creative" icon="lightbulb" label="Creative" />
      <q-tab name="progression" icon="trending_up" label="Progression" />
      <q-tab name="crossing" icon="swap_horiz" label="Crossing" />
    </q-tabs>
    <q-tab-panels v-model="plotTab" animated>
      <q-tab-panel name="creative">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in creativeCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in passingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
        </div>
      </q-tab-panel>
      <q-tab-panel name="progression">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in progressionCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in passingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
        </div>
      </q-tab-panel>
      <q-tab-panel name="crossing">
        <div class="charts-grid">
          <DynamicScatterPlotCard v-for="config in crossingCharts" :key="config.title" v-bind="config" :is-dark-mode="isDarkMode" :all-players-data="filteredPlayers" @player-click="$emit('player-click', $event)" />
        </div>
        <div class="stats-grid">
          <StatCard v-for="stat in passingStats" :key="stat.key" :stat="stat" :players="topPlayersByStat[stat.key]" @player-click="$emit('player-click', $event)" />
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
  passingStats: { type: Array, required: true },
  passingCharts: { type: Array, required: true },
})

defineEmits(['player-click'])

const plotTab = ref('creative')
const creativeCharts = computed(() => props.passingCharts.filter((c) => c.group === 'creative'))
const progressionCharts = computed(() =>
  props.passingCharts.filter((c) => c.group === 'progression')
)
const crossingCharts = computed(() => props.passingCharts.filter((c) => c.group === 'crossing'))
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

