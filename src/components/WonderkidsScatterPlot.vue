<template>
    <div class="wonderkids-scatter">
        <div class="row items-center q-mb-xs">
            <span
                class="text-caption text-weight-semibold q-mr-sm"
                :class="isDarkMode ? 'text-grey-4' : 'text-grey-8'"
            >
                Age / Ability
            </span>
        </div>
        <div
            v-if="hasNoData"
            class="flex flex-center text-caption text-grey-5 no-data-placeholder"
        >
            No data to display
        </div>
        <div v-else class="scatter-canvas-wrapper">
            <Bubble
                :data="chartData"
                :options="chartOptions"
            />
        </div>
    </div>
</template>

<script setup>
import {
    BubbleController,
    Chart as ChartJS,
    Legend,
    LinearScale,
    PointElement,
    Title,
    Tooltip,
} from 'chart.js'
import { computed } from 'vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import { Bubble } from 'vue-chartjs'

ChartJS.register(Title, Tooltip, Legend, PointElement, LinearScale, BubbleController)

const props = defineProps({
    players: { type: Array, required: true },
    isDarkMode: { type: Boolean, default: false },
})

const emit = defineEmits(['player-click'])

const DEF_POSITIONS = ['DR', 'DC', 'DL', 'WBR', 'WBL']
const MID_POSITIONS = ['DM', 'MR', 'MC', 'ML', 'AMR', 'AMC', 'AML']

const POSITION_GROUPS = {
    GK: { label: 'GK', light: 'rgba(245,158,11,0.75)', dark: 'rgba(252,211,77,0.8)' },
    DEF: { label: 'DEF', light: 'rgba(59,130,246,0.75)', dark: 'rgba(147,197,253,0.8)' },
    MID: { label: 'MID', light: 'rgba(34,197,94,0.75)', dark: 'rgba(134,239,172,0.8)' },
    ATT: { label: 'ATT', light: 'rgba(239,68,68,0.75)', dark: 'rgba(252,165,165,0.8)' },
}

const getPositionGroup = (player) => {
    const positions = player.short_positions || player.shortPositions || []
    if (positions.includes('GK')) return 'GK'
    if (positions.some((p) => DEF_POSITIONS.includes(p))) return 'DEF'
    if (positions.includes('ST')) return 'ATT'
    if (positions.some((p) => MID_POSITIONS.includes(p))) return 'MID'
    return 'MID'
}

// Deterministic jitter so same-age players spread into a cluster instead of stacking
function ageJitter(player) {
    let hash = 0
    const key = (player.name || '') + (player.id || '')
    for (let i = 0; i < key.length; i++) hash = (hash * 31 + key.charCodeAt(i)) | 0
    return ((Math.abs(hash) % 1000) / 1000 - 0.5) * 0.4
}

const chartData = computed(() => {
    const buckets = Object.fromEntries(Object.keys(POSITION_GROUPS).map((k) => [k, []]))

    // Compute mean overall per age so bubble size encodes above-average-for-age performance
    const meanByAge = {}
    for (const player of props.players) {
        const age = Number(player.age)
        const overall = Number(player.Overall || player.overall) || 0
        if (!age || !overall) continue
        if (!meanByAge[age]) meanByAge[age] = { sum: 0, count: 0 }
        meanByAge[age].sum += overall
        meanByAge[age].count++
    }
    for (const age of Object.keys(meanByAge)) {
        meanByAge[age] = meanByAge[age].sum / meanByAge[age].count
    }

    // Collect all deltas first to normalise radius across the dataset
    const deltas = []
    for (const player of props.players) {
        const age = Number(player.age)
        const overall = Number(player.Overall || player.overall) || 0
        if (!age || !overall) continue
        deltas.push(overall - (meanByAge[age] || 0))
    }
    const maxDelta = Math.max(...deltas)
    const minDelta = Math.min(...deltas)
    const deltaRange = maxDelta - minDelta

    for (const player of props.players) {
        const x = Number(player.age)
        const y = Number(player.Overall || player.overall || 0)
        if (!x || y <= 0) continue

        const delta = y - (meanByAge[x] || 0)
        const r = deltaRange > 0
            ? Math.round(3 + ((delta - minDelta) / deltaRange) * 9)
            : 5

        buckets[getPositionGroup(player)].push({ x: x + ageJitter(player), y, r, delta, player })
    }

    const colorKey = props.isDarkMode ? 'dark' : 'light'
    return {
        datasets: Object.entries(POSITION_GROUPS).map(([key, g]) => ({
            label: g.label,
            data: buckets[key],
            backgroundColor: g[colorKey],
            borderColor: g[colorKey].replace('0.75', '1').replace('0.8', '1'),
            borderWidth: 1,
        })),
    }
})

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
                text: 'Age',
                color: theme.value.text,
                font: { size: 11, weight: 'bold' },
            },
            min: 14,
            max: 22,
            ticks: { stepSize: 1, color: theme.value.muted, font: { size: 10 } },
            grid: { color: theme.value.grid },
            border: { color: theme.value.muted },
        },
        y: {
            title: {
                display: true,
                text: 'Overall',
                color: theme.value.text,
                font: { size: 11, weight: 'bold' },
            },
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
                    const delta = ctx.raw.delta
                    const sign = delta >= 0 ? '+' : ''
                    return [p.name, `${p.club || '—'}  •  Age ${p.age}  •  Overall: ${ctx.raw.y}  •  vs peers: ${sign}${delta.toFixed(1)}`]
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
.wonderkids-scatter {
    padding: 0.25rem 0;
}
.scatter-canvas-wrapper {
    position: relative;
    height: 260px;
    cursor: pointer;
}
.no-data-placeholder {
    height: 100px;
}
</style>
