<template>
  <svg width="22" height="36" viewBox="0 0 22 36" style="flex-shrink:0;display:block">
    <!-- pitch -->
    <rect x="0.5" y="0.5" width="21" height="35" rx="1" ry="1"
      fill="rgba(34,197,94,0.18)" stroke="#16a34a" stroke-width="0.6" />
    <!-- halfway line -->
    <line x1="0.5" y1="18" x2="21.5" y2="18" stroke="#16a34a" stroke-width="0.4" opacity="0.5" />
    <!-- centre circle -->
    <circle cx="11" cy="18" r="3" fill="none" stroke="#16a34a" stroke-width="0.4" opacity="0.4" />
    <!-- penalty areas -->
    <rect x="4" y="0.5" width="14" height="6" fill="none" stroke="#16a34a" stroke-width="0.4" opacity="0.4" />
    <rect x="4" y="29.5" width="14" height="6" fill="none" stroke="#16a34a" stroke-width="0.4" opacity="0.4" />
    <!-- player dot -->
    <circle :cx="dotX" :cy="dotY" r="2.8" fill="#2563eb" />
  </svg>
</template>

<script>
import { computed, defineComponent } from 'vue'

const POSITION_COORDS = {
  GK: { x: 50, y: 88 },
  DC: { x: 50, y: 74 },
  DL: { x: 18, y: 74 },
  DR: { x: 82, y: 74 },
  WBL: { x: 12, y: 62 },
  WBR: { x: 88, y: 62 },
  DM: { x: 50, y: 56 },
  MC: { x: 50, y: 44 },
  ML: { x: 18, y: 44 },
  MR: { x: 82, y: 44 },
  AMC: { x: 50, y: 30 },
  AML: { x: 18, y: 30 },
  AMR: { x: 82, y: 30 },
  ST: { x: 50, y: 14 },
}

export default defineComponent({
  name: 'PositionPitchDot',
  props: { position: { type: String, default: 'MC' } },
  setup(props) {
    const coords = computed(() => POSITION_COORDS[props.position] || { x: 50, y: 50 })
    const dotX = computed(() => (coords.value.x / 100) * 22)
    const dotY = computed(() => (coords.value.y / 100) * 36)
    return { dotX, dotY }
  },
})
</script>
