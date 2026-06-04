<template>
  <q-card flat bordered class="scout-card">
    <q-badge v-if="player.isNew" color="blue-5" label="NEW" class="new-badge" />

    <!-- Header: overall + name -->
    <div class="card-header">
      <OverallCircle :value="player.Overall" :pa="player.pa" />
      <div class="card-name-block">
        <div class="player-name">{{ player.name }}</div>
        <div class="player-sub">{{ player.position }} · {{ player.age }}y · {{ player.nationality }}</div>
      </div>
    </div>

    <!-- Club -->
    <div class="card-club">{{ player.club }} <span class="card-division">· {{ player.division }}</span></div>

    <!-- Radar chart -->
    <div class="card-radar">
      <RadarChart :player="player" :size="80" />
    </div>

    <!-- Stat labels under radar -->
    <div class="card-stat-labels">
      <span v-for="s in statLabels" :key="s.label" class="stat-label-item">
        <span class="stat-label-key">{{ s.label }}</span>
        <span class="stat-label-val">{{ s.v ?? '-' }}</span>
      </span>
    </div>

    <q-separator class="q-my-xs" />

    <!-- Value / wage row -->
    <div class="card-footer">
      <div>
        <div class="card-foot-label">Value</div>
        <div class="card-foot-val">{{ player.transfer_value }}</div>
      </div>
      <div>
        <div class="card-foot-label">Score</div>
        <div class="card-foot-val">{{ player.valueScore }}</div>
      </div>
      <div>
        <div class="card-foot-label">Wage</div>
        <div class="card-foot-val">{{ player.wage }}</div>
      </div>
    </div>

    <!-- Prospect bar -->
    <div v-if="paGap" class="prospect-bar">
      ⭐ Prospect · +{{ paGap }} PA gap
    </div>
  </q-card>
</template>

<script>
import { computed, defineComponent } from 'vue'
import OverallCircle from './OverallCircle.vue'
import RadarChart from './RadarChart.vue'

export default defineComponent({
  name: 'PlayerScoutCard',
  components: { OverallCircle, RadarChart },
  props: { player: { type: Object, required: true } },
  setup(props) {
    const paGap = computed(() => (props.player.pa ? props.player.pa - props.player.ca : null))
    const isGK = computed(() => props.player.shortPositions?.includes('GK'))
    const statLabels = computed(() => {
      const p = props.player
      if (isGK.value)
        return [
          { label: 'DIV', v: p.div },
          { label: 'HAN', v: p.han },
          { label: 'REF', v: p.ref },
          { label: 'KIC', v: p.kic },
          { label: 'SPD', v: p.spd },
          { label: 'POS', v: p.pos },
        ]
      return [
        { label: 'PAC', v: p.pac },
        { label: 'SHO', v: p.sho },
        { label: 'PAS', v: p.pas },
        { label: 'DRI', v: p.dri },
        { label: 'DEF', v: p.def },
        { label: 'PHY', v: p.phy },
      ]
    })
    return { paGap, statLabels }
  },
})
</script>

<style scoped>
.scout-card {
  width: 200px;
  padding: 12px;
  position: relative;
}
.new-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  font-size: 9px;
  animation: newFlash 2s ease-out forwards;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}
.card-name-block { min-width: 0; }
.player-name { font-size: 13px; font-weight: 700; line-height: 1.2; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.player-sub  { font-size: 10px; color: #9ca3af; line-height: 1.3; }
.card-club   { font-size: 11px; color: #6b7280; margin-bottom: 8px; }
.card-division { color: #d1d5db; }
.card-radar  { display: flex; justify-content: center; margin-bottom: 4px; }
.card-stat-labels {
  display: flex;
  justify-content: space-around;
  margin-bottom: 6px;
}
.stat-label-item { display: flex; flex-direction: column; align-items: center; }
.stat-label-key  { font-size: 8px; color: #9ca3af; }
.stat-label-val  { font-size: 10px; font-weight: 700; color: #374151; }
.card-footer {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
}
.card-foot-label { font-size: 9px; color: #9ca3af; }
.card-foot-val   { font-size: 11px; font-weight: 600; }
.prospect-bar {
  font-size: 10px;
  color: #d97706;
  margin-top: 4px;
  font-weight: 600;
}

@keyframes newFlash {
  0%   { background: #60a5fa; box-shadow: 0 0 8px #60a5fa; }
  60%  { background: #60a5fa; }
  100% { background: #3b82f6; box-shadow: none; }
}
</style>
