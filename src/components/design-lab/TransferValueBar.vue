<template>
  <div>
    <div style="font-size:12px;font-weight:500;white-space:nowrap">{{ label }}</div>
    <div style="height:3px;background:rgba(0,0,0,0.08);border-radius:2px;margin-top:3px;width:80px">
      <div :style="{ width: pct + '%', background: '#3b82f6', height: '100%', borderRadius: '2px' }"></div>
    </div>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'TransferValueBar',
  props: {
    amount: { type: Number, default: 0 },
    max: { type: Number, default: 1 },
    label: { type: String, default: '' },
  },
  setup(props) {
    const pct = computed(() => Math.max(4, Math.round((props.amount / props.max) * 100)))
    return { pct }
  },
})
</script>
