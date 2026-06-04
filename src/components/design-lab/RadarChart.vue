<template>
  <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`">
    <polygon v-for="scale in [0.33, 0.66, 1]" :key="scale"
      :points="gridPoints(scale)"
      fill="none" stroke="rgba(0,0,0,0.12)" stroke-width="0.6" />
    <line v-for="(ax, i) in axisLines" :key="i"
      :x1="cx" :y1="cy" :x2="ax.x2" :y2="ax.y2"
      stroke="rgba(0,0,0,0.1)" stroke-width="0.6" />
    <polygon :points="dataPoints" fill="rgba(59,130,246,0.22)" stroke="#3b82f6" stroke-width="1.5" />
    <circle v-for="(pt, i) in dataPointCoords" :key="'d'+i"
      :cx="pt.x" :cy="pt.y" r="1.8" fill="#3b82f6" />
  </svg>
</template>

<script>
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'RadarChart',
  props: {
    player: { type: Object, required: true },
    size: { type: Number, default: 72 },
  },
  setup(props) {
    const n = 6
    const cx = computed(() => props.size / 2)
    const cy = computed(() => props.size / 2)
    const r = computed(() => props.size / 2 - 5)

    const statValues = computed(() => {
      const p = props.player
      if (p.shortPositions?.includes('GK'))
        return [p.div || 0, p.han || 0, p.ref || 0, p.kic || 0, p.spd || 0, p.pos || 0]
      return [p.pac || 0, p.sho || 0, p.pas || 0, p.dri || 0, p.def || 0, p.phy || 0]
    })

    const angleOf = (i) => (Math.PI * 2 * i) / n - Math.PI / 2

    const coordAt = (i, ratio) => ({
      x: cx.value + r.value * ratio * Math.cos(angleOf(i)),
      y: cy.value + r.value * ratio * Math.sin(angleOf(i)),
    })

    const gridPoints = (scale) =>
      Array.from({ length: n }, (_, i) => {
        const c = coordAt(i, scale)
        return `${c.x},${c.y}`
      }).join(' ')

    const axisLines = computed(() =>
      Array.from({ length: n }, (_, i) => {
        const c = coordAt(i, 1)
        return { x2: c.x, y2: c.y }
      })
    )

    const dataPointCoords = computed(() => statValues.value.map((v, i) => coordAt(i, v / 99)))

    const dataPoints = computed(() => dataPointCoords.value.map((c) => `${c.x},${c.y}`).join(' '))

    return { cx, cy, r, gridPoints, axisLines, dataPoints, dataPointCoords }
  },
})
</script>
