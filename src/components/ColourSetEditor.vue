<template>
  <section class="colour-editor">
    <h3 class="text-subtitle1 q-my-sm">{{ range.label }} ({{ range.min }}–{{ range.max }})</h3>
    <q-banner v-if="store.errors[scale]" class="bg-negative text-white q-mb-md" rounded role="alert">
      {{ store.errors[scale] }}
      <template v-if="!store.loaded[scale]" #action>
        <q-btn flat label="Retry" :loading="store.busy[scale]" @click="store.load(scale)" />
      </template>
    </q-banner>
    <q-linear-progress v-if="store.busy[scale]" indeterminate class="q-mb-md" />
    <fieldset :disabled="!store.loaded[scale] || store.busy[scale]" class="colour-editor__fields">
      <div class="row q-col-gutter-sm items-center q-mb-md">
        <q-select
          v-model="selectedId" class="col" outlined dense emit-value map-options
          :options="options" label="Colour set" :disable="unavailable"
          @update:model-value="resetDraft"
        />
        <q-btn
          class="col-auto" outline no-caps label="Use set" :disable="unavailable || selectedId === document.activeSetId"
          @click="activate"
        />
      </div>
      <p class="text-caption">Active set: {{ activeName }}. Select a set and choose “Use set” to apply it.</p>
      <q-input v-model="name" outlined dense maxlength="100" label="Set name" :disable="unavailable" class="q-mb-md" />
      <p class="text-caption">Each band includes its starting value and runs up to the next band. The last band includes {{ range.max }}.</p>
      <div v-for="(band, index) in bands" :key="index" class="colour-editor__band">
        <q-input
          v-model.number="band.min" type="number" outlined dense label="From"
          :min="index ? bands[index - 1].min + 1 : range.min"
          :max="index + 1 < bands.length ? bands[index + 1].min - 1 : range.max"
          step="1" :disable="unavailable || index === 0" :aria-label="`Band ${index + 1} starts at`"
        />
        <span class="text-caption">{{ index + 1 < bands.length ? `to < ${bands[index + 1].min}` : `to ${range.max}` }}</span>
        <label class="colour-editor__picker">
          <span class="sr-only">Band {{ index + 1 }} colour</span>
          <input v-model="band.color" type="color" :disabled="unavailable" :aria-label="`Band ${index + 1} colour`" />
        </label>
        <q-input v-model="band.color" outlined dense maxlength="7" label="Hex colour" :disable="unavailable" />
        <q-btn
          flat round dense icon="delete_outline" :aria-label="`Remove band ${index + 1}`"
          :disable="unavailable || index === 0" @click="bands.splice(index, 1)"
        />
      </div>
      <q-btn flat no-caps icon="add" :label="`Add band (${bands.length}/20)`" :disable="unavailable || bands.length >= 20 || !!bandError" @click="addBand" />
      <p v-if="bandError" class="text-negative" role="alert">{{ bandError }}</p>
      <div v-else class="colour-editor__preview q-my-md" aria-label="Colour band preview">
        <span
          v-for="value in previewValues" :key="value"
          :style="previewStyle(value)"
        >{{ value }}</span>
      </div>
      <div class="row q-gutter-sm q-mt-sm">
        <q-btn color="primary" no-caps label="Save and apply" :disable="unavailable || !!bandError || !name.trim()" @click="saveSet(false)" />
        <q-btn outline no-caps label="Save as new set" :disable="unavailable || !!bandError || !name.trim()" @click="saveSet(true)" />
        <q-btn flat no-caps label="Reset edits" :disable="unavailable" @click="resetDraft" />
        <q-btn v-if="selectedId" flat no-caps color="negative" label="Delete set" :disable="unavailable" @click="confirmDelete = true" />
      </div>
      <div v-if="confirmDelete" class="q-mt-md" role="alert">
        Delete “{{ selectedSet?.name }}”? Deleting the active set restores default colours.
        <q-btn flat no-caps color="negative" label="Confirm delete" :disable="unavailable" @click="removeSet" />
        <q-btn flat no-caps label="Cancel" @click="confirmDelete = false" />
      </div>
      <p v-if="savedMessage" class="text-positive q-mt-sm" role="status">{{ savedMessage }}</p>
    </fieldset>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useColourSetStore } from '../stores/colourSetStore'
import {
  bandIndex,
  COLOUR_SCALES,
  colourText,
  defaultBands,
  validateBands,
} from '../utils/colourSets'

const props = defineProps({ scale: { type: String, required: true } })
const store = useColourSetStore()
const range = COLOUR_SCALES[props.scale]
const document = computed(() => store.documents[props.scale])
const unavailable = computed(() => !store.loaded[props.scale] || store.busy[props.scale])
const selectedId = ref('')
const name = ref('My colour set')
const bands = ref(defaultBands(props.scale))
const confirmDelete = ref(false)
const savedMessage = ref('')
const selectedSet = computed(() => document.value.sets.find((set) => set.id === selectedId.value))
const activeName = computed(
  () => document.value.sets.find((set) => set.id === document.value.activeSetId)?.name || 'Default'
)
const options = computed(() => [
  { label: 'Default', value: '' },
  ...document.value.sets.map((set) => ({ label: set.name, value: set.id })),
])
const bandError = computed(() => validateBands(bands.value, props.scale))
const previewValues = Array.from({ length: range.max - range.min + 1 }, (_, i) => i + range.min)

function resetDraft() {
  name.value = selectedSet.value?.name || 'My colour set'
  bands.value = (selectedSet.value?.bands || defaultBands(props.scale)).map((band) => ({ ...band }))
  confirmDelete.value = false
  savedMessage.value = ''
}

watch(
  document,
  () => {
    selectedId.value = document.value.activeSetId
    resetDraft()
  },
  { immediate: true }
)
watch(
  [name, bands],
  () => {
    savedMessage.value = ''
  },
  { deep: true }
)
onMounted(() => store.load(props.scale))

function addBand() {
  let splitIndex = 0
  let largestWidth = 0
  bands.value.forEach((band, index) => {
    const end = bands.value[index + 1]?.min ?? range.max + 1
    if (end - band.min > largestWidth) {
      largestWidth = end - band.min
      splitIndex = index
    }
  })
  if (largestWidth < 2) return
  const band = bands.value[splitIndex]
  bands.value.splice(splitIndex + 1, 0, {
    min: band.min + Math.floor(largestWidth / 2),
    color: band.color,
  })
}

function previewStyle(value) {
  const color = bands.value[bandIndex(value, bands.value)].color
  return { backgroundColor: color, color: colourText(color) }
}

async function persist(next, message) {
  if (await store.save(props.scale, next)) savedMessage.value = message
}

async function activate() {
  await persist(
    { ...document.value, activeSetId: selectedId.value },
    'Colour set applied and saved.'
  )
}

async function saveSet(asNew) {
  if (bandError.value || !name.value.trim()) return
  const id =
    asNew || !selectedId.value
      ? globalThis.crypto?.randomUUID?.() ||
        `colour-${Date.now()}-${Math.random().toString(36).slice(2)}`
      : selectedId.value
  const set = { id, name: name.value.trim(), bands: bands.value.map((band) => ({ ...band })) }
  const sets = document.value.sets.filter((candidate) => candidate.id !== id)
  await persist(
    { version: 1, activeSetId: id, sets: [...sets, set] },
    'Colour set saved and applied.'
  )
}

async function removeSet() {
  const sets = document.value.sets.filter((set) => set.id !== selectedId.value)
  const activeSetId =
    document.value.activeSetId === selectedId.value ? '' : document.value.activeSetId
  await persist({ ...document.value, activeSetId, sets }, 'Colour set deleted.')
}
</script>

<style scoped lang="scss">
.colour-editor + .colour-editor { margin-top: 2rem; border-top: 1px solid var(--surface-border); padding-top: 1rem; }
.colour-editor__fields { border: 0; padding: 0; margin: 0; min-width: 0; }
.colour-editor__band { display: grid; grid-template-columns: 80px 65px 40px minmax(90px, 1fr) 32px; gap: 8px; align-items: center; margin-bottom: 8px; }
.colour-editor__picker input { width: 36px; height: 36px; padding: 0; cursor: pointer; border: 0; background: transparent; }
.colour-editor__preview { display: flex; flex-wrap: wrap; gap: 4px; }
.colour-editor__preview span { min-width: 28px; padding: 4px; border-radius: 4px; text-align: center; font-weight: 600; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
@media (max-width: 450px) {
  .colour-editor__band { grid-template-columns: 70px 1fr 32px 28px; gap: 4px; }
  .colour-editor__band > :nth-child(4) { grid-column: 1 / -1; grid-row: 2; }
  .colour-editor__band > :nth-child(5) { grid-column: 4; grid-row: 1; }
  .colour-editor__picker input { width: 30px; }
}
</style>
