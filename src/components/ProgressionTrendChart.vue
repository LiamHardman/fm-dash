<template>
  <div :style="{ width: `${width}px`, height: `${height}px` }">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from 'chart.js'
import { computed } from 'vue'
import { Line } from 'vue-chartjs'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip)

const props = defineProps({
  labels: { type: Array, required: true },
  values: { type: Array, required: true },
  color: { type: String, default: '#5b8def' },
  width: { type: Number, default: 120 },
  height: { type: Number, default: 32 },
  showAxes: { type: Boolean, default: false },
})

const chartData = computed(() => ({
  labels: props.labels,
  datasets: [
    {
      data: props.values,
      borderColor: props.color,
      backgroundColor: `${props.color}33`,
      fill: props.showAxes,
      tension: 0.3,
      pointRadius: props.showAxes ? 3 : 0,
      pointHoverRadius: 4,
      borderWidth: 2,
    },
  ],
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  plugins: {
    tooltip: { enabled: true, mode: 'index', intersect: false },
    legend: { display: false },
  },
  scales: props.showAxes
    ? {
        x: { display: true, ticks: { font: { size: 9 } } },
        y: { display: true, ticks: { font: { size: 9 } } },
      }
    : { x: { display: false }, y: { display: false } },
}))
</script>
