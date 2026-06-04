<template>
  <div style="position:relative;display:inline-flex;align-items:center;justify-content:center;width:32px;height:32px;">
    <svg width="32" height="32" style="position:absolute;top:0;left:0;">
      <circle cx="16" cy="16" r="11" :fill="color" />
      <circle v-if="pa" cx="16" cy="16" r="13.5" fill="none"
        stroke="rgba(217,180,6,0.65)" stroke-width="2.5"
        :stroke-dasharray="paDash"
        stroke-linecap="round"
        transform="rotate(-90 16 16)" />
    </svg>
    <span style="position:relative;z-index:1;color:#fff;font-size:11px;font-weight:800;line-height:1;text-shadow:0 1px 2px rgba(0,0,0,0.3)">{{ value }}</span>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'OverallCircle',
  props: {
    value: { type: Number, default: 0 },
    pa: { type: Number, default: null },
  },
  setup(props) {
    const color = computed(() => {
      if (props.value >= 86) return '#d97706'
      if (props.value >= 76) return '#16a34a'
      if (props.value >= 60) return '#ea580c'
      return '#dc2626'
    })
    const circ = 2 * Math.PI * 13.5 // ~84.82
    const paDash = computed(() => {
      if (!props.pa) return ''
      const filled = (props.pa / 99) * circ
      return `${filled} ${circ - filled}`
    })
    return { color, paDash }
  },
})
</script>
