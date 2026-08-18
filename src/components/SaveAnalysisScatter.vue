<template>
    <div class="save-analysis-scatter">
        <div
            v-if="hasNoData"
            class="flex flex-center text-caption text-grey-5 no-data-placeholder"
        >
            No players match the current filters.
        </div>
        <div v-else class="scatter-canvas-wrapper">
            <Scatter
                :data="chartData"
                :options="chartOptions"
            />
        </div>
    </div>
</template>

<script setup>
import { Chart as ChartJS, Legend, LinearScale, PointElement, Title, Tooltip } from 'chart.js'
import { computed } from 'vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import { Scatter } from 'vue-chartjs'

ChartJS.register(Title, Tooltip, Legend, PointElement, LinearScale)

const props = defineProps({
  players: { type: Array, required: true },
  isDarkMode: { type: Boolean, default: false },
})

const emit = defineEmits(['player-click'])

// Matches the exact strings GetPlayerPositionGroupsGo (src/api/positions.go) assigns to
// player.positionGroups.
const GROUP_ORDER = ['Goalkeepers', 'Defenders', 'Wing-Backs', 'Midfielders', 'Attackers']

const GROUP_STYLES = {
  Goalkeepers: { label: 'GK', light: 'rgba(245,158,11,0.75)', dark: 'rgba(252,211,77,0.8)' },
  Defenders: { label: 'DEF', light: 'rgba(59,130,246,0.75)', dark: 'rgba(147,197,253,0.8)' },
  'Wing-Backs': { label: 'WB', light: 'rgba(20,184,166,0.75)', dark: 'rgba(94,234,212,0.8)' },
  Midfielders: { label: 'MID', light: 'rgba(34,197,94,0.75)', dark: 'rgba(134,239,172,0.8)' },
  Attackers: { label: 'ATT', light: 'rgba(239,68,68,0.75)', dark: 'rgba(252,165,165,0.8)' },
}

const getPrimaryGroup = (player) => {
  const groups = player.positionGroups || player.PositionGroups || []
  return GROUP_ORDER.find((g) => groups.includes(g)) || 'Midfielders'
}

const getCA = (player) => Number(player.ca ?? player.CA)
const getOverall = (player) => Number(player.overall ?? player.Overall)
const getBestRole = (player) => player.bestRoleOverall || player.BestRoleOverall || 'Unknown'

const chartData = computed(() => {
  const buckets = Object.fromEntries(GROUP_ORDER.map((g) => [g, []]))

  for (const player of props.players) {
    const x = getCA(player)
    const y = getOverall(player)
    if (!x || !y) continue
    buckets[getPrimaryGroup(player)].push({ x, y, player })
  }

  const colorKey = props.isDarkMode ? 'dark' : 'light'
  return {
    datasets: GROUP_ORDER.filter((g) => buckets[g].length > 0).map((g) => ({
      label: GROUP_STYLES[g].label,
      data: buckets[g],
      backgroundColor: GROUP_STYLES[g][colorKey],
      borderColor: GROUP_STYLES[g][colorKey].replace('0.75', '1').replace('0.8', '1'),
      borderWidth: 1,
      pointRadius: 4,
      pointHoverRadius: 7,
    })),
  }
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const hasNoData = computed(() => chartData.value.datasets.every((ds) => ds.data.length === 0))

const theme = computed(() =>
  props.isDarkMode
    ? {
        text: 'rgba(255,255,255,0.85)',
        muted: 'rgba(255,255,255,0.5)',
        grid: 'rgba(255,255,255,0.07)',
      }
    : { text: '#374151', muted: '#9ca3af', grid: 'rgba(0,0,0,0.07)' }
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  layout: { padding: { top: 4, right: 12, bottom: 4, left: 4 } },
  scales: {
    x: {
      title: {
        display: true,
        text: 'CA (Current Ability)',
        color: theme.value.text,
        font: { size: 12, weight: 'bold' },
      },
      min: 0,
      max: 200,
      ticks: { color: theme.value.muted, font: { size: 10 } },
      grid: { color: theme.value.grid },
      border: { color: theme.value.muted },
    },
    y: {
      title: {
        display: true,
        text: 'Overall',
        color: theme.value.text,
        font: { size: 12, weight: 'bold' },
      },
      min: 0,
      max: 99,
      ticks: { color: theme.value.muted, font: { size: 10 } },
      grid: { color: theme.value.grid },
      border: { color: theme.value.muted },
    },
  },
  plugins: {
    legend: {
      display: true,
      position: 'top',
      align: 'end',
      labels: {
        color: theme.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        font: { size: 11 },
        padding: 10,
        boxWidth: 10,
      },
    },
    tooltip: {
      callbacks: {
        title: () => '',
        label: (ctx) => {
          const p = ctx.raw.player
          return [
            p.name,
            `${p.club || '—'}  •  Best role: ${getBestRole(p)}`,
            `CA: ${ctx.raw.x}  •  Overall: ${ctx.raw.y}`,
          ]
        },
      },
    },
  },
  onClick: (_event, elements) => {
    if (elements?.length > 0) {
      const { datasetIndex, index } = elements[0]
      const player = chartData.value.datasets[datasetIndex].data[index].player
      emit('player-click', player)
    }
  },
}))
</script>

<style scoped>
.save-analysis-scatter {
    padding: 0.25rem 0;
}
.scatter-canvas-wrapper {
    position: relative;
    height: 520px;
}
.no-data-placeholder {
    height: 200px;
}
</style>
