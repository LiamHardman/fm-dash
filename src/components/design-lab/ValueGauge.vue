<template>
  <div style="position:relative;width:44px;height:28px">
    <svg width="44" height="32" style="position:absolute;top:0;left:0;overflow:visible">
      <!-- background track -->
      <path d="M 8,24 A 14,14 0 0,1 36,24"
        fill="none" stroke="rgba(0,0,0,0.1)" stroke-width="4" stroke-linecap="round" />
      <!-- value arc -->
      <path d="M 8,24 A 14,14 0 0,1 36,24"
        fill="none" :stroke="color" stroke-width="4" stroke-linecap="round"
        :stroke-dasharray="dashArray" />
    </svg>
    <span style="position:absolute;bottom:0;left:0;right:0;text-align:center;font-size:10px;font-weight:700;line-height:1">{{ score }}</span>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'ValueGauge',
  props: { score: { type: Number, default: 0 } },
  setup(props) {
    // Semi-circle arc length = π * r = π * 14 ≈ 43.98
    const arcLen = Math.PI * 14
    const dashArray = computed(() => {
      const filled = arcLen * (props.score / 100)
      return `${filled} ${arcLen * 2}`
    })
    const color = computed(() => {
      if (props.score >= 75) return '#16a34a'
      if (props.score >= 50) return '#ea580c'
      return '#dc2626'
    })
    return { dashArray, color }
  },
})
</script>
