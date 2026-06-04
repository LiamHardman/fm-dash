<template>
  <div class="expand-detail q-pa-sm">
    <div class="row q-gutter-md items-start">
      <!-- Stat bars -->
      <div class="col-auto">
        <div class="text-caption text-grey-6 q-mb-xs">Attributes</div>
        <StatBars :player="player" />
      </div>

      <!-- Numbers / info -->
      <div class="col">
        <div class="row q-gutter-sm q-mb-sm">
          <div class="col-auto">
            <div class="text-caption text-grey-6">OVR</div>
            <OverallCircle :value="player.Overall" :pa="player.pa" />
            <div v-if="paGap" style="font-size:10px;color:#d97706;margin-top:2px">+{{ paGap }} PA</div>
          </div>
          <div class="col-auto">
            <div class="text-caption text-grey-6">Transfer value</div>
            <div style="font-size:13px;font-weight:600">{{ player.transfer_value }}</div>
            <ValueGauge :score="player.valueScore" />
          </div>
          <div class="col-auto">
            <div class="text-caption text-grey-6">Wage</div>
            <div style="font-size:13px;font-weight:600">{{ player.wage }}</div>
            <div style="font-size:10px;color:#9ca3af">£{{ wagePerPoint }}/pt</div>
          </div>
          <div class="col-auto">
            <div class="text-caption text-grey-6">Personality</div>
            <div style="font-size:12px;font-weight:500">{{ player.personality || '—' }}</div>
            <div v-if="player.media_handling" style="font-size:11px;color:#9ca3af">{{ player.media_handling }}</div>
          </div>
          <div class="col-auto">
            <div class="text-caption text-grey-6">Club / League</div>
            <div style="font-size:12px;font-weight:500">{{ player.club }}</div>
            <div style="font-size:11px;color:#9ca3af">{{ player.division }}</div>
          </div>
        </div>

        <!-- Role overalls -->
        <div v-if="player.roleSpecificOveralls?.length">
          <div class="text-caption text-grey-6 q-mb-xs">Best roles</div>
          <div class="flex" style="gap:6px;flex-wrap:wrap">
            <q-chip v-for="role in player.roleSpecificOveralls" :key="role.roleName"
              dense size="sm" color="blue-1" text-color="blue-9">
              {{ role.roleName }} · {{ role.score }}
            </q-chip>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'
import OverallCircle from './OverallCircle.vue'
import StatBars from './StatBars.vue'
import ValueGauge from './ValueGauge.vue'

export default defineComponent({
  name: 'PlayerExpandDetail',
  components: { StatBars, OverallCircle, ValueGauge },
  props: { player: { type: Object, required: true } },
  setup(props) {
    const paGap = computed(() => (props.player.pa ? props.player.pa - props.player.ca : null))
    const wagePerPoint = computed(() => {
      if (!props.player.wageAmount || !props.player.Overall) return '—'
      return `${(props.player.wageAmount / 1000 / props.player.Overall).toFixed(1)}K`
    })
    return { paGap, wagePerPoint }
  },
})
</script>

<style scoped>
.expand-detail {
  background: rgba(59, 130, 246, 0.03);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}
</style>
