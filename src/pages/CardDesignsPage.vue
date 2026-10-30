<script setup>
import { computed, ref } from 'vue'
import { FAMILY_PRESENTATIONS, resolveEdition } from '../cards/editions.mjs'
import CardGeneratorDialog from '../components/CardGeneratorDialog.vue'
import PageHeader from '../components/layout/PageHeader.vue'
import SectionCard from '../components/layout/SectionCard.vue'
import PlayerCards from '../components/PlayerCards.vue'
import { CARD_STUDIO_PLAYER } from '../utils/cardStudioPlayer'

const goalkeeper = ref(false)
const showSupportingInfo = ref(true)
const width = ref(240)
const showGenerator = ref(false)
const exportError = ref('')
const exporting = ref('')
const cards = new Map()
const player = computed(() => ({
  ...CARD_STUDIO_PLAYER,
  ...(goalkeeper.value
    ? { name: 'Andr\u00e9 ter Stegen', position: 'GK', playstylePlus: 'footwork' }
    : {}),
}))
const presentations = [
  ...FAMILY_PRESENTATIONS,
  { edition: 'squad-player', baseEdition: 'silver', rare: true },
]
const keyFor = (presentation) => `${presentation.edition}-${presentation.rare}`
async function download(presentation) {
  const key = keyFor(presentation)
  exporting.value = key
  exportError.value = ''
  try {
    const blob = await cards.get(key).exportPng(2)
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = `${key}.png`
    anchor.click()
    setTimeout(() => URL.revokeObjectURL(url), 10000)
  } catch (error) {
    exportError.value = error.message
  } finally {
    exporting.value = ''
  }
}
</script>

<template>
  <q-page class="card-designs-page page-container">
    <PageHeader title="Card Design Studio" subtitle="Explore the complete player card family and create your own card." icon="style" />
    <SectionCard>
      <div class="studio-controls">
        <q-toggle v-model="goalkeeper" label="Goalkeeper" />
        <q-toggle v-model="showSupportingInfo" label="Age, value and weekly wage" />
        <label class="width-control">Card width: {{ width }}px <input v-model.number="width" type="range" min="190" max="340" /></label>
        <q-btn color="primary" label="Open Card Generator" @click="showGenerator = true" />
      </div>
    </SectionCard>
    <p v-if="exportError" role="alert">{{ exportError }}</p>
    <div class="card-gallery" :style="{ '--preview-width': `${width}px` }">
      <article v-for="presentation in presentations" :key="keyFor(presentation)" class="design-panel">
        <h2>{{ resolveEdition(presentation).label }}</h2>
        <PlayerCards :ref="card => card ? cards.set(keyFor(presentation), card) : cards.delete(keyFor(presentation))"
          :player="player" v-bind="presentation" :show-supporting-info="showSupportingInfo" />
        <q-btn outline icon="download" label="Download PNG" :loading="exporting === keyFor(presentation)" @click="download(presentation)" />
      </article>
    </div>
    <CardGeneratorDialog v-if="showGenerator" :show="showGenerator" :player="player" @close="showGenerator = false" />
  </q-page>
</template>

<style scoped>
.card-designs-page { padding-bottom: 32px; }
.studio-controls { display: flex; flex-wrap: wrap; align-items: center; gap: 16px; }
.width-control { display: flex; flex-direction: column; }
.card-gallery { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, var(--preview-width)), 1fr)); gap: 24px; margin-top: 24px; }
.design-panel { display: flex; flex-direction: column; align-items: center; gap: 12px; min-width: 0; }
.design-panel h2 { font-size: 16px; line-height: 1.4; margin: 0; text-align: center; }
.design-panel :deep(.fm-player-card) { width: var(--preview-width); }
</style>
