<template>
  <q-tooltip v-if="info.label" anchor="top middle" self="bottom middle">{{ info.label }}</q-tooltip>
  <div class="flex items-center no-wrap" style="gap:5px">
    <q-icon :name="info.icon" :style="{ color: info.color }" size="16px" />
    <span style="font-size:11px;font-weight:500">{{ info.label || '—' }}</span>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

const MAP = {
  Professional: { icon: 'workspace_premium', color: '#16a34a' },
  Ambitious: { icon: 'trending_up', color: '#3b82f6' },
  Temperamental: { icon: 'bolt', color: '#dc2626' },
  Unflappable: { icon: 'self_improvement', color: '#8b5cf6' },
  Calm: { icon: 'spa', color: '#0ea5e9' },
  Outgoing: { icon: 'celebration', color: '#f59e0b' },
  'Media-friendly': { icon: 'mic', color: '#6366f1' },
}

export default defineComponent({
  name: 'PersonalityIcon',
  props: { personality: { type: String, default: '' } },
  setup(props) {
    const info = computed(
      () =>
        MAP[props.personality] || {
          icon: 'help_outline',
          color: '#9ca3af',
          label: props.personality || '—',
        }
    )
    const infoWithLabel = computed(() => ({ ...info.value, label: props.personality || '—' }))
    return { info: infoWithLabel }
  },
})
</script>
