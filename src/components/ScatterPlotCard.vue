<template>
    <q-card flat bordered :class="isDarkMode ? 'theme-dark' : 'theme-light'" class="scatter-plot-card full-height">
      <q-card-section class="q-pa-md card-header">
        <div class="title-section">
          <div class="card-title">{{ title }}</div>
          <div class="card-subtitle">{{ subtitle }}</div>
        </div>
        <img v-if="logoUrl" :src="logoUrl" alt="Logo" class="chart-logo" />
      </q-card-section>
      <q-card-section class="chart-area">
        <div v-if="chartData.datasets[0].data.length < 5" class="no-data-overlay">
          <q-icon name="warning" class="q-mr-sm" />
          Not enough player data to build this chart.
        </div>
        <div :class="{ 'chart-container-hidden': chartData.datasets[0].data.length < 5 }">
          <Scatter 
            :data="chartData" 
            :options="chartOptions"
          />
        </div>
      </q-card-section>
    </q-card>
  </template>
  
  <script setup>
import { Chart as ChartJS, Legend, LinearScale, PointElement, Title, Tooltip } from 'chart.js'
import annotationPlugin from 'chartjs-plugin-annotation'
import { computed } from 'vue'
import { Scatter } from 'vue-chartjs'

// Register Chart.js components and the annotation plugin
ChartJS.register(Title, Tooltip, Legend, PointElement, LinearScale, annotationPlugin)

// --- Component Properties ---
const props = defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  logoUrl: { type: String, default: '' },
  allPlayersData: { type: Array, required: true },
  highlightedPlayers: { type: Array, default: () => [] },
  xAxisKey: { type: String, required: true },
  yAxisKey: { type: String, required: true },
  xAxisLabel: { type: String, required: true },
  yAxisLabel: { type: String, required: true },
  quadrantLabels: { type: Object, required: true },
  isDarkMode: { type: Boolean, default: false }
})

const emit = defineEmits(['player-click'])

// --- Theme Colors ---
const lightThemeColors = {
  background: '#FFFFFF',
  text: '#333333',
  axis: '#666666',
  gridLine: 'rgba(0, 0, 0, 0.1)',
  point: 'rgba(25, 118, 210, 0.7)',
  good: 'rgba(0, 100, 0, 0.7)',
  bad: 'rgba(183, 28, 28, 0.7)',
  highlightBg: 'rgba(255, 255, 255, 0.85)',
  highlightFaceBorder: '#00695c'
}

const darkThemeColors = {
  background: '#313742',
  text: 'rgba(255, 255, 255, 0.9)',
  axis: 'rgba(255, 255, 255, 0.7)',
  gridLine: 'rgba(255, 255, 255, 0.5)',
  point: 'rgba(211, 211, 211, 0.7)',
  good: '#A2C592',
  bad: '#E6827A',
  highlightBg: 'rgba(40, 44, 52, 0.85)',
  highlightFaceBorder: '#E6827A'
}

const themeColors = computed(() => (props.isDarkMode ? darkThemeColors : lightThemeColors))

// --- Data Processing ---
const getNumericValue = val => {
  if (val === undefined || val === null || val === '-' || val === '') return null
  const cleaned = String(val).replace(/,/g, '').replace(/%/g, '')
  const num = parseFloat(cleaned)
  return Number.isNaN(num) ? null : num
}

const processedData = computed(() => {
  return props.allPlayersData
    .map(player => {
      const xValue = getNumericValue(
        player.performanceStatsNumeric?.[props.xAxisKey] || player.attributes?.[props.xAxisKey]
      )
      const yValue = getNumericValue(
        player.performanceStatsNumeric?.[props.yAxisKey] || player.attributes?.[props.yAxisKey]
      )

      if (xValue === null || yValue === null || xValue === 0 || yValue === 0) return null

      return {
        x: xValue,
        y: yValue,
        player: player // Store the full player object
      }
    })
    .filter(item => item !== null)
})

const avgX = computed(() => {
  if (processedData.value.length === 0) return 0
  const sum = processedData.value.reduce((acc, p) => acc + p.x, 0)
  return sum / processedData.value.length
})

const avgY = computed(() => {
  if (processedData.value.length === 0) return 0
  const sum = processedData.value.reduce((acc, p) => acc + p.y, 0)
  return sum / processedData.value.length
})

// --- Chart Configuration ---
const chartData = computed(() => ({
  datasets: [
    {
      data: processedData.value,
      backgroundColor: themeColors.value.point,
      pointRadius: 4,
      pointHoverRadius: 7
    }
  ]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: true,
  aspectRatio: 1.5,
  layout: {
    padding: 20
  },
  scales: {
    x: {
      title: {
        display: true,
        text: props.xAxisLabel,
        color: themeColors.value.text,
        font: { size: 10, weight: 'bold' }
      },
      grid: { display: false },
      ticks: { color: themeColors.value.axis, font: { size: 12 } },
      border: { color: themeColors.value.axis }
    },
    y: {
      title: {
        display: true,
        text: props.yAxisLabel,
        color: themeColors.value.text,
        font: { size: 10, weight: 'bold' }
      },
      grid: { display: false },
      ticks: { color: themeColors.value.axis, font: { size: 12 } },
      border: { color: themeColors.value.axis }
    }
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: ctx => [
          `${ctx.raw.player.name}`,
          `${props.xAxisLabel}: ${ctx.raw.x.toFixed(2)}`,
          `${props.yAxisLabel}: ${ctx.raw.y.toFixed(2)}`
        ]
      }
    },
    annotation: {
      annotations: {
        // Quadrant Lines
        avgLineX: {
          type: 'line',
          yMin: avgY.value,
          yMax: avgY.value,
          borderColor: themeColors.value.gridLine,
          borderWidth: 1.5
        },
        avgLineY: {
          type: 'line',
          xMin: avgX.value,
          xMax: avgX.value,
          borderColor: themeColors.value.gridLine,
          borderWidth: 1.5
        },

        // Quadrant Labels
        ...generateQuadrantLabels(),

        // Highlighted Player Labels & Images
        ...generatePlayerHighlightAnnotations()
      }
    }
  },
  onClick: (_event, elements) => {
    if (elements && elements.length > 0) {
      const index = elements[0].index
      const player = processedData.value[index].player
      emit('player-click', player)
    }
  }
}))

// --- Annotation Generators ---
const generateQuadrantLabels = () => {
  const { topRight, topLeft, bottomRight, bottomLeft } = props.quadrantLabels
  const font = { size: 11 }
  const createLabel = (key, content, color, position, xAdjust, yAdjust, textAlign) => ({
    [key]: {
      type: 'label',
      xValue:
        position.x === 'start' ? ctx => ctx.chart.scales.x.min : ctx => ctx.chart.scales.x.max,
      yValue:
        position.y === 'start' ? ctx => ctx.chart.scales.y.max : ctx => ctx.chart.scales.y.min,
      content: content,
      color: [color, themeColors.value.text],
      font: [{ ...font, weight: 'bold' }, { ...font }],
      position: { x: position.x, y: 'start' }, // Keep text aligned at its 'start' edge
      xAdjust,
      yAdjust,
      textAlign
    }
  })

  return {
    ...createLabel(
      'topRightLabel',
      topRight,
      themeColors.value.good,
      { x: 'end', y: 'start' },
      -15,
      15,
      'right'
    ),
    ...createLabel(
      'topLeftLabel',
      topLeft,
      themeColors.value.good,
      { x: 'start', y: 'start' },
      15,
      15,
      'left'
    ),
    ...createLabel(
      'bottomRightLabel',
      bottomRight,
      themeColors.value.bad,
      { x: 'end', y: 'end' },
      -15,
      -15,
      'right'
    ),
    ...createLabel(
      'bottomLeftLabel',
      bottomLeft,
      themeColors.value.bad,
      { x: 'start', y: 'end' },
      15,
      -15,
      'left'
    )
  }
}

const generatePlayerHighlightAnnotations = () => {
  const annotations = {}

  props.highlightedPlayers.forEach((player, index) => {
    // Find the full player data object to get the faceUrl
    const fullPlayerData = processedData.value.find(p => p.player.name === player.name)

    // Player Name Label
    annotations[`playerLabel${index}`] = {
      type: 'label',
      xValue: player.x,
      yValue: player.y,
      content: player.name,
      backgroundColor: themeColors.value.highlightBg,
      color: themeColors.value.text,
      font: { size: 12, weight: 'bold' },
      borderRadius: 4,
      padding: 6,
      yAdjust: -35 // Position above the player point
    }

    // Player Face Image
    if (fullPlayerData?.player?.faceUrl) {
      const img = new Image()
      img.src = fullPlayerData.player.faceUrl
      annotations[`playerImage${index}`] = {
        type: 'point', // Using a point to apply border styling easily
        xValue: player.x,
        yValue: player.y,
        radius: 12,
        backgroundColor: ctx => {
          const pattern = ctx.chart.ctx.createPattern(img, 'repeat')
          return pattern
        },
        borderWidth: 2,
        borderColor: themeColors.value.highlightFaceBorder,
        drawTime: 'afterDatasetsDraw'
      }
    }
  })

  return annotations
}
</script>
  
  <style scoped>
  .scatter-plot-card {
    display: flex;
    flex-direction: column;
    position: relative;
    height: 600px; /* Reduced from 1000px */
    transition: all 0.3s ease;
    border-radius: 16px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
    border: 1px solid rgba(0, 0, 0, 0.05);
    
    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
    }
  }
  
  /* Light Theme */
  .theme-light {
    background-color: #ffffff;
    color: #333333;
    border: 1px solid rgba(0, 0, 0, 0.05);
  }
  .theme-light .card-subtitle { color: #666; }
  .theme-light .no-data-overlay { 
    background: rgba(255, 255, 255, 0.9); 
    color: #666;
    backdrop-filter: blur(10px);
  }
  
  /* Dark Theme */
  .theme-dark {
    background-color: #1e1e1e;
    color: rgba(255, 255, 255, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
  .theme-dark .card-subtitle { color: rgba(255, 255, 255, 0.7); }
  .theme-dark .no-data-overlay { 
    background: rgba(30, 30, 30, 0.9); 
    color: rgba(255, 255, 255, 0.7);
    backdrop-filter: blur(10px);
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 1.5rem;
    border-bottom: 1px solid rgba(0, 0, 0, 0.05);
    
    .theme-dark & {
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    }
  }
  
  .title-section {
    text-align: left;
  }
  
  .card-title {
    font-size: 1.2rem;
    font-weight: 600;
    line-height: 1.2;
    margin: 0;
    color: #2d3436;
    
    .theme-dark & {
        color: rgba(255, 255, 255, 0.9);
    }
  }
  
  .card-subtitle {
    font-size: 0.9rem;
    font-weight: 500;
    margin-top: 0.25rem;
  }
  
  .chart-logo {
    height: 35px;
    width: auto;
    opacity: 0.8;
    transition: opacity 0.3s ease;
    
    &:hover {
        opacity: 1;
    }
  }
  
  .chart-area {
    flex: 1;
    position: relative;
    height: 500px; /* Reduced from 900px */
    padding: 1rem;
  }
  
  .no-data-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10;
    font-style: italic;
    border-radius: inherit;
    gap: 0.5rem;
    
    .q-icon {
        font-size: 1.5rem;
        opacity: 0.7;
    }
  }
  
  .chart-container-hidden {
    visibility: hidden;
  }
  
  .chart-container {
    width: 100%;
    height: 100%;
    transition: background-color 0.3s;
    border-radius: 8px;
    overflow: hidden;
  }
  </style>