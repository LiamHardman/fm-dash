<template>
    <q-card flat bordered class="scatter-plot-card full-height">
      <q-card-section class="q-pa-md">
        <div class="card-title">{{ title }}</div>
        <div v-if="chartData.datasets[0].data.length < 5" class="no-data-overlay">
          <q-icon name="warning" class="q-mr-sm" />
          Not enough player data to build this chart.
        </div>
        <div :class="{ 'chart-container-hidden': chartData.datasets[0].data.length < 5 }">
          <Scatter :data="chartData" :options="chartOptions" />
        </div>
      </q-card-section>
    </q-card>
  </template>
  
  <script setup>
  import { computed } from 'vue';
  import { Scatter } from 'vue-chartjs';
  import {
    Chart as ChartJS,
    Title,
    Tooltip,
    Legend,
    PointElement,
    LinearScale,
  } from 'chart.js';
  import annotationPlugin from 'chartjs-plugin-annotation';
  
  // Register Chart.js components and the annotation plugin
  ChartJS.register(Title, Tooltip, Legend, PointElement, LinearScale, annotationPlugin);
  
  // Define component properties
  const props = defineProps({
    title: { type: String, required: true },
    allPlayersData: { type: Array, required: true },
    xAxisKey: { type: String, required: true },
    yAxisKey: { type: String, required: true },
    xAxisLabel: { type: String, required: true },
    yAxisLabel: { type: String, required: true },
    quadrantLabels: { type: Object, required: true }
  });
  
  /**
   * Safely parses a statistic value into a number.
   * Handles various non-numeric formats.
   * @param {string | number} val - The value to parse.
   * @returns {number | null} - The parsed number or null if invalid.
   */
  const getNumericValue = (val) => {
    if (val === undefined || val === null || val === '-' || val === '') return null;
    const cleaned = String(val).replace(/,/g, '').replace(/%/g, '');
    const num = parseFloat(cleaned);
    return isNaN(num) ? null : num;
  };
  
  // Process the raw player data into a format suitable for Chart.js
  const processedData = computed(() => {
    return props.allPlayersData
      .map(player => {
        const x = getNumericValue(player.attributes?.[props.xAxisKey]);
        const y = getNumericValue(player.attributes?.[props.yAxisKey]);
  
        if (x !== null && y !== null) {
          return {
            x,
            y,
            player: {
              name: player.name || player.Name || player.Player || 'Unknown',
              club: player.club || player.Club || 'Unknown',
            },
          };
        }
        return null;
      })
      .filter(p => p !== null); // Filter out players with incomplete data for this chart
  });
  
  // Calculate the average X value to position the vertical quadrant line
  const avgX = computed(() => {
    if (processedData.value.length === 0) return 0;
    const sum = processedData.value.reduce((acc, p) => acc + p.x, 0);
    return sum / processedData.value.length;
  });
  
  // Calculate the average Y value to position the horizontal quadrant line
  const avgY = computed(() => {
    if (processedData.value.length === 0) return 0;
    const sum = processedData.value.reduce((acc, p) => acc + p.y, 0);
    return sum / processedData.value.length;
  });
  
  // Computed property for the chart's dataset
  const chartData = computed(() => ({
    datasets: [
      {
        label: 'Players',
        data: processedData.value,
        backgroundColor: 'rgba(25, 118, 210, 0.7)', // --q-color-primary with opacity
        pointRadius: 5,
        pointHoverRadius: 8,
      },
    ],
  }));
  
  // Computed property for the chart's options, including scales, tooltips, and annotations
  const chartOptions = computed(() => ({
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      x: {
        title: {
          display: true,
          text: props.xAxisLabel,
        },
        grid: {
          display: false,
        },
      },
      y: {
        title: {
          display: true,
          text: props.yAxisLabel,
        },
        grid: {
          display: false,
        },
      },
    },
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
        callbacks: {
          label: (context) => {
            const dataPoint = context.raw;
            return [
              `${dataPoint.player.name} (${dataPoint.player.club})`,
              `${props.xAxisLabel}: ${dataPoint.x.toFixed(2)}`,
              `${props.yAxisLabel}: ${dataPoint.y.toFixed(2)}`,
            ];
          },
        },
      },
      annotation: {
        annotations: {
          // Vertical line at average X
          lineY: {
            type: 'line',
            xMin: avgX.value,
            xMax: avgX.value,
            borderColor: 'rgba(0, 0, 0, 0.2)',
            borderWidth: 1,
            borderDash: [6, 6],
          },
          // Horizontal line at average Y
          lineX: {
            type: 'line',
            yMin: avgY.value,
            yMax: avgY.value,
            borderColor: 'rgba(0, 0, 0, 0.2)',
            borderWidth: 1,
            borderDash: [6, 6],
          },
          // Quadrant text labels
          topRight: {
            type: 'label',
            xValue: avgX.value,
            yValue: (ctx) => ctx.chart.scales.y.max,
            xAdjust: 5,
            yAdjust: -15,
            content: props.quadrantLabels.topRight,
            color: 'rgba(0, 100, 0, 0.7)',
            position: 'start',
            font: { weight: 'bold' }
          },
          topLeft: {
            type: 'label',
            xValue: avgX.value,
            yValue: (ctx) => ctx.chart.scales.y.max,
            xAdjust: -5,
            yAdjust: -15,
            content: props.quadrantLabels.topLeft,
            color: 'rgba(0, 0, 0, 0.5)',
            position: 'end',
            font: { weight: 'normal' }
          },
          bottomRight: {
            type: 'label',
            xValue: avgX.value,
            yValue: (ctx) => ctx.chart.scales.y.min,
            xAdjust: 5,
            yAdjust: 15,
            content: props.quadrantLabels.bottomRight,
            color: 'rgba(0, 0, 0, 0.5)',
            position: 'start',
            font: { weight: 'normal' }
          },
          bottomLeft: {
            type: 'label',
            xValue: avgX.value,
            yValue: (ctx) => ctx.chart.scales.y.min,
            xAdjust: -5,
            yAdjust: 15,
            content: props.quadrantLabels.bottomLeft,
            color: 'rgba(183, 28, 28, 0.7)',
            position: 'end',
            font: { weight: 'bold' }
          }
        }
      }
    },
  }));
  </script>
  
  <style scoped>
  .scatter-plot-card {
    height: 450px; /* Give the card a fixed height for consistent layout */
    display: flex;
    flex-direction: column;
    position: relative; /* Needed for absolute positioning of the overlay */
  }
  
  .card-title {
    font-size: 1.1rem;
    font-weight: 600;
    text-align: center;
    margin-bottom: 16px;
    color: var(--q-color-on-surface);
  }
  
  .no-data-overlay {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(255, 255, 255, 0.8);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 10;
      color: var(--q-color-on-surface-variant);
      font-style: italic;
      border-radius: inherit; /* Match card's border-radius */
  }
  
  .chart-container-hidden {
      visibility: hidden; /* Hide chart if there's not enough data, but keep space */
  }
  </style>
  