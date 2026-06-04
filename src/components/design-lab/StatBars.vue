<template>
  <div style="min-width:130px">
    <div v-for="s in stats" :key="s.label" style="display:flex;align-items:center;gap:4px;margin-bottom:2px">
      <span style="font-size:9px;color:#9ca3af;width:22px;text-align:right;flex-shrink:0">{{ s.label }}</span>
      <div style="flex:1;height:5px;background:rgba(0,0,0,0.08);border-radius:3px;overflow:hidden">
        <div :style="{ width: (s.v || 0) + '%', background: barColor(s.v), height: '100%', borderRadius: '3px' }"></div>
      </div>
      <span style="font-size:9px;font-weight:700;width:18px;color:#374151">{{ s.v ?? '-' }}</span>
    </div>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'StatBars',
  props: { player: { type: Object, required: true } },
  setup(props) {
    const isGK = computed(() => props.player.shortPositions?.includes('GK'))
    const stats = computed(() => {
      if (isGK.value) {
        return [
          { label: 'DIV', v: props.player.div },
          { label: 'HAN', v: props.player.han },
          { label: 'REF', v: props.player.ref },
          { label: 'KIC', v: props.player.kic },
          { label: 'SPD', v: props.player.spd },
          { label: 'POS', v: props.player.pos },
        ]
      }
      return [
        { label: 'PAC', v: props.player.pac },
        { label: 'SHO', v: props.player.sho },
        { label: 'PAS', v: props.player.pas },
        { label: 'DRI', v: props.player.dri },
        { label: 'DEF', v: props.player.def },
        { label: 'PHY', v: props.player.phy },
      ]
    })
    const barColor = (v) => {
      if (!v) return '#e5e7eb'
      if (v >= 86) return '#d97706'
      if (v >= 76) return '#16a34a'
      if (v >= 60) return '#ea580c'
      return '#dc2626'
    }
    return { stats, barColor }
  },
})
</script>
