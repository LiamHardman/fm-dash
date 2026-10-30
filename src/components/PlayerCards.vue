<script setup>
import { computed, nextTick, ref, useId, watch } from 'vue'
import PlayerCard from '../cards/PlayerCard.vue'
import { fetchFullPlayerStats } from '../services/playerService'
import { useUiStore } from '../stores/uiStore'
import { mapPlayerCard, selectPlayerCardEdition } from '../utils/playerCardMapper'

const props = defineProps({
  player: { type: Object, required: true },
  currencySymbol: { type: String, default: '\u00a3' },
  clubImageUrl: { type: String, default: null },
  nationFlagUrl: { type: String, default: null },
  playerFaceUrl: { type: String, default: null },
  datasetId: { type: String, default: null },
  selectedRole: { type: String, default: null },
  positionOverride: { type: String, default: null },
  edition: { type: String, default: null },
  rare: { type: Boolean, default: undefined },
  baseEdition: { type: String, default: null },
  playstylePlus: { type: String, default: undefined },
  portraitMode: { type: String, default: 'cutout' },
  showSupportingInfo: { type: Boolean, default: true },
  interactive: { type: Boolean, default: false },
})
const emit = defineEmits(['click', 'loaded', 'error'])
const uiStore = useUiStore()
const card = ref(null)
const idPrefix = `fm-card-${useId().replace(/[^a-z0-9_-]/gi, '-')}`
const detail = ref(null)
const clubLogo = ref('')
const pendingData = new Set()
watch(
  () => [props.datasetId, props.player.uid || props.player.UID, props.selectedRole],
  async (_, __, onCleanup) => {
    let active = true
    onCleanup(() => {
      active = false
    })
    detail.value = null
    const uid = props.player.uid || props.player.UID
    if (
      !props.datasetId ||
      !uid ||
      (props.player.nationalityIso && (!props.selectedRole || props.player.roleSpecificOveralls))
    )
      return
    const request = fetchFullPlayerStats(props.datasetId, uid)
    pendingData.add(request)
    try {
      const result = await request
      if (active) detail.value = result.data?.player || null
    } catch {
      /* The supplied player remains usable if detail loading fails. */
    } finally {
      pendingData.delete(request)
    }
  },
  { immediate: true }
)
watch(
  () => [props.player.club, props.clubImageUrl, uiStore.showLogos],
  async (_, __, onCleanup) => {
    let active = true
    onCleanup(() => {
      active = false
    })
    clubLogo.value = ''
    if (
      props.clubImageUrl ||
      props.player.clubImageUrl ||
      !uiStore.showLogos ||
      !props.player.club ||
      props.player.club === '-'
    )
      return
    const club = props.player.club
    const request = import('../composables/useTeamLogosBackend').then(({ useTeamLogosBackend }) =>
      useTeamLogosBackend().getTeamLogoUrl(club)
    )
    pendingData.add(request)
    try {
      const url = await request
      if (active) clubLogo.value = url || ''
    } catch {
      /* A missing badge uses the club code tile. */
    } finally {
      pendingData.delete(request)
    }
  },
  { immediate: true }
)
const mappedPlayer = computed(() =>
  mapPlayerCard(
    {
      ...detail.value,
      ...Object.fromEntries(Object.entries(props.player).filter(([, value]) => value != null)),
      roleSpecificOveralls: props.player.roleSpecificOveralls || detail.value?.roleSpecificOveralls,
    },
    {
      ...props,
      clubImageUrl: props.clubImageUrl || props.player.clubImageUrl || clubLogo.value,
      useStatSummaryForOverall: uiStore.useStatSummaryForOverall,
    }
  )
)
const appearance = computed(() => selectPlayerCardEdition(props.player, mappedPlayer.value, props))
const portrait = computed(() => {
  if (props.playerFaceUrl) return props.playerFaceUrl
  if (!uiStore.showFaces) return ''
  const uid = props.player.UID || props.player.uid
  return uid
    ? `/api/faces?uid=${encodeURIComponent(uid)}&nationality=${encodeURIComponent(props.player.nationality_fifa_code || '')}`
    : ''
})
async function exportPng(scale = 2) {
  await nextTick()
  while (pendingData.size) {
    await Promise.allSettled([...pendingData])
    await nextTick()
  }
  if (!card.value) throw new Error('The card is not ready to export.')
  return card.value.exportPng(scale)
}
defineExpose({ exportPng })
</script>

<template>
  <PlayerCard ref="card" class="fm-player-card" :player="mappedPlayer" v-bind="appearance"
    :id-prefix="idPrefix" :portrait="portrait" :portrait-mode="portraitMode"
    :show-supporting-info="showSupportingInfo" :interactive="interactive"
    @activate="emit('click', props.player)" @loaded="emit('loaded', $event)" @error="emit('error', $event)" />
</template>

<style scoped>
.fm-player-card { width: 280px; max-width: 100%; flex-shrink: 0; }
</style>
