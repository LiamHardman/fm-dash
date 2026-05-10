<template>
  <div class="stat-gauge-wrap">
    <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`">
      <!-- Track ring -->
      <circle
        :cx="center"
        :cy="center"
        :r="radius"
        fill="none"
        :stroke="isDark ? 'rgba(255,255,255,0.12)' : 'rgba(0,0,0,0.1)'"
        :stroke-width="strokeWidth"
      />
      <!-- Progress arc -->
      <circle
        v-if="hasValue"
        :cx="center"
        :cy="center"
        :r="radius"
        fill="none"
        :stroke="gaugeColor"
        :stroke-width="strokeWidth"
        stroke-linecap="round"
        :stroke-dasharray="circumference"
        :stroke-dashoffset="dashOffset"
        :transform="`rotate(-90 ${center} ${center})`"
      />
      <!-- Value label -->
      <text
        :x="center"
        :y="center"
        text-anchor="middle"
        dominant-baseline="central"
        :font-size="textSize"
        font-weight="700"
        font-family="inherit"
        :fill="isDark ? 'rgba(255,255,255,0.9)' : '#1e293b'"
      >{{ label }}</text>
    </svg>
  </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed } from 'vue'

export default {
  name: 'StatGauge',
  props: {
    value: { type: Number, default: null },
    max: { type: Number, default: 100 },
    size: { type: Number, default: 46 },
    strokeWidth: { type: Number, default: 3 },
  },
  setup(props) {
    const $q = useQuasar()
    const isDark = computed(() => $q.dark.isActive)

    const center = computed(() => props.size / 2)
    const radius = computed(() => (props.size - props.strokeWidth) / 2)
    const circumference = computed(() => 2 * Math.PI * radius.value)

    const hasValue = computed(
      () => props.value !== null && props.value !== undefined && !isNaN(props.value)
    )

    const fraction = computed(() => {
      if (!hasValue.value) return 0
      return Math.min(Math.max(props.value / props.max, 0), 1)
    })

    const dashOffset = computed(() => circumference.value * (1 - fraction.value))

    const label = computed(() => {
      if (!hasValue.value) return '-'
      return String(Math.round(props.value))
    })

    const textSize = computed(() => {
      const digits = label.value.length
      if (digits >= 3) return props.size * 0.3
      return props.size * 0.37
    })

    const gaugeColor = computed(() => {
      if (!hasValue.value) return '#9ca3af'
      const pct = fraction.value
      if (pct >= 0.9) return '#9333ea'
      if (pct >= 0.8) return '#14b8a6'
      if (pct >= 0.65) return '#22c55e'
      if (pct >= 0.5) return '#3b82f6'
      if (pct >= 0.3) return '#f97316'
      return '#ef4444'
    })

    return {
      isDark,
      center,
      radius,
      circumference,
      hasValue,
      fraction,
      dashOffset,
      label,
      textSize,
      gaugeColor,
    }
  },
}
</script>

<style scoped>
.stat-gauge-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
