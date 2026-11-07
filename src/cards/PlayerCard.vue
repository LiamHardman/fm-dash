<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { resolveEdition } from './editions.mjs'
import { cardPngBlob } from './export-png.mjs'
import { renderPlayerCard } from './forged-crest.mjs'
import { prepareCardImages } from './images.mjs'

const props = defineProps({
  player: { type: Object, required: true },
  edition: { type: String, default: 'squad-player' },
  rare: { type: Boolean, default: false },
  baseEdition: { type: String, default: 'gold' },
  artwork: { type: String, default: '' },
  portrait: { type: String, default: '' },
  portraitMode: { type: String, default: 'cutout' },
  showSupportingInfo: { type: Boolean, default: false },
  interactive: { type: Boolean, default: false },
  idPrefix: { type: String, required: true },
})
const root = ref(null)
const emit = defineEmits(['loaded', 'error', 'activate'])
const images = ref(null)
const failure = ref('')
let generation = 0,
  readyPromise = Promise.resolve()
watch(
  () => [
    props.edition,
    props.rare,
    props.baseEdition,
    props.artwork,
    props.portrait,
    props.player.nation?.src,
    props.player.club?.src,
  ],
  () => {
    const current = ++generation
    failure.value = ''
    images.value = null
    readyPromise = (async () => {
      try {
        const appearance = resolveEdition(props)
        const loaded = await prepareCardImages({
          artwork: props.artwork || appearance.artwork,
          portrait: props.portrait,
          nation: props.player.nation?.src,
          club: props.player.club?.src,
        })
        if (generation !== current) return
        images.value = loaded
        await nextTick()
        if (generation === current)
          emit('loaded', { edition: appearance.edition, presentation: appearance.id })
      } catch (error) {
        if (generation !== current) return
        failure.value = error.message
        emit('error', error)
      }
    })()
  },
  { immediate: true }
)
const svg = computed(() => {
  if (!images.value) return ''
  return renderPlayerCard({
    ...props,
    artwork: images.value.artwork,
    portrait: images.value.portrait,
    player: {
      ...props.player,
      nation: props.player.nation && { ...props.player.nation, src: images.value.nation },
      club: props.player.club && { ...props.player.club, src: images.value.club },
    },
  })
})
async function exportPng(scale = 2) {
  let pending
  do {
    pending = readyPromise
    await pending
  } while (pending !== readyPromise)
  await nextTick()
  if (failure.value) throw new Error(failure.value)
  const scene = root.value?.querySelector('svg')
  if (!scene) throw new Error('The card is not ready to export.')
  return cardPngBlob(scene, scale)
}
function activate() {
  if (props.interactive) emit('activate', props.player)
}
function onKey(event) {
  if (props.interactive && (event.key === 'Enter' || event.key === ' ')) {
    event.preventDefault()
    activate()
  }
}
defineExpose({ exportPng })
</script>

<template>
  <div ref="root" class="player-card" :style="{ aspectRatio: `1100 / ${showSupportingInfo ? 1870 : 1536}` }"
    :role="interactive ? 'button' : undefined" :tabindex="interactive ? 0 : undefined"
    :aria-busy="!images && !failure" @click="activate" @keydown="onKey">
    <p v-if="failure" role="alert">{{ failure }}</p>
    <div v-else-if="svg" class="scene" v-html="svg" />
    <span v-else class="loading" role="status">Loading player card…</span>
  </div>
</template>

<style scoped>
.player-card { width: 100%; position: relative; }
.player-card[role="button"] { cursor: pointer; }
.player-card:focus-visible { outline: 3px solid currentColor; outline-offset: 4px; }
.scene > :deep(svg) { display: block; width: 100%; height: auto; }
.loading { position: absolute; inset: 0; display: grid; place-items: center; font: 14px sans-serif; }
</style>
