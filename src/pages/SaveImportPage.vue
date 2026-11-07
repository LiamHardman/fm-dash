<template>
    <q-page class="save-import-page">
        <div class="page-container">
            <PageHeader
                title="FM27 Save Import"
                subtitle="Read players straight from a Football Manager 27 .fm save. The save is parsed in your browser and never uploaded; only the decoded player table is sent, exactly like a CSV export."
                icon="science"
            >
                <template #actions>
                    <q-chip dense color="warning" text-color="dark" icon="warning" label="Experimental" />
                </template>
            </PageHeader>

            <SectionCard title="Choose a save" icon="upload_file">
                <div class="upload-row">
                    <q-file
                        v-model="saveFile"
                        label="Choose an FM27 .fm save"
                        accept=".fm"
                        filled
                        class="save-file-input"
                        :disable="busy"
                        @update:model-value="reset"
                    >
                        <template #prepend>
                            <q-icon name="attach_file" />
                        </template>
                    </q-file>
                    <q-btn
                        color="primary"
                        label="Read save"
                        icon="manage_search"
                        :disable="!saveFile || busy"
                        :loading="parsing"
                        @click="parseSelectedSave"
                    />
                </div>

                <div v-if="parsing" class="progress-block" role="status">
                    <div class="progress-label">{{ stageLabel }}</div>
                    <q-linear-progress
                        :value="stageProgress"
                        :indeterminate="stageProgress === null"
                        color="primary"
                        rounded
                    />
                </div>

                <q-banner v-if="errorMessage" class="bg-negative text-white q-mt-md" rounded>
                    {{ errorMessage }}
                </q-banner>
            </SectionCard>

            <SectionCard v-if="summary" title="Save contents" icon="list_alt" class="q-mt-md">
                <dl class="summary-grid">
                    <div><dt>Save</dt><dd>{{ summary.saveName || saveFile?.name }}</dd></div>
                    <div><dt>In-game date</dt><dd>{{ summary.gameDate || 'unknown' }}</dd></div>
                    <div><dt>Players</dt><dd>{{ formatCount(summary.players) }}</dd></div>
                    <div><dt>At a club</dt><dd>{{ percent(summary.withClub) }}</dd></div>
                    <div><dt>With a wage</dt><dd>{{ percent(summary.withWage) }}</dd></div>
                    <div><dt>With a transfer value</dt><dd>{{ percent(summary.withValue) }}</dd></div>
                    <div><dt>With a division</dt><dd>{{ percent(summary.withDivision) }}</dd></div>
                    <div v-if="summary.withStats !== undefined"><dt>With season stats</dt><dd>{{ percent(summary.withStats) }}</dd></div>
                    <div v-if="summary.managedClub"><dt>Managed club</dt><dd>{{ summary.managedClub.club }} · {{ summary.managedClub.division }} (detected)</dd></div>
                    <div><dt>Read in</dt><dd>{{ (summary.elapsedMs / 1000).toFixed(1) }} s</dd></div>
                </dl>

                <ul class="caveats">
                    <li>Attributes are exact for every player in the database: no scouting fog.</li>
                    <li>Personality and media handling are derived from the hidden personality attributes; current-season statistics come from the save. Advanced statistics FM didn't record for a player are left blank rather than zero.</li>
                    <li>Division is derived from each club's league matches; clubs that haven't played a league match recently have no division.</li>
                    <li v-if="summary.managedClub">The managed club is set for the new dataset unless one is already stored; you can change it on the dataset page.</li>
                    <li>Game build {{ summary.build }}. Saves from other FM27 builds may decode partially.</li>
                </ul>

                <q-banner v-if="tooLarge" class="bg-warning text-dark q-mb-md" rounded>
                    The decoded table is {{ formatSize(csvBlob.size) }}, over this server's
                    {{ formatSize(maxUploadBytes) }} upload limit.
                </q-banner>

                <div class="import-row">
                    <q-btn
                        color="primary"
                        icon="cloud_upload"
                        :label="`Import ${formatCount(summary.players)} players`"
                        :disable="busy || tooLarge"
                        :loading="uploading"
                        @click="importPlayers"
                    />
                    <span class="size-note">{{ formatSize(csvBlob.size) }} to upload</span>
                </div>
            </SectionCard>
        </div>
    </q-page>
</template>

<script setup>
import { Notify } from 'quasar'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
// biome-ignore lint/correctness/noUnusedImports: used in template
import PageHeader from '../components/layout/PageHeader.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import SectionCard from '../components/layout/SectionCard.vue'
import playerService from '../services/playerService.js'
import { usePlayerStore } from '../stores/playerStore'
import { setManagedClubIfAbsent } from '../utils/managedClubImport.js'

const router = useRouter()
const playerStore = usePlayerStore()

const saveFile = ref(null)
const parsing = ref(false)
const uploading = ref(false)
const errorMessage = ref('')
const progress = ref({ stage: 'opening' })
const summary = ref(null)
const csvBlob = ref(null)
const maxUploadBytes = ref(15 * 1024 * 1024)
let worker = null

// biome-ignore lint/correctness/noUnusedVariables: used in template
const busy = computed(() => parsing.value || uploading.value)
// biome-ignore lint/correctness/noUnusedVariables: used in template
const tooLarge = computed(() => csvBlob.value && csvBlob.value.size > maxUploadBytes.value)

const STAGE_LABELS = {
  opening: 'Opening the save…',
  decompressing: 'Decompressing the game database…',
  names: 'Reading the name tables…',
  scanning: 'Finding player records…',
  clubs: 'Reading clubs and leagues…',
  decoding: 'Decoding players…',
  done: 'Decoding players…',
  exporting: 'Building the player table…',
}
// biome-ignore lint/correctness/noUnusedVariables: used in template
const stageLabel = computed(() => {
  const { stage, current, total } = progress.value
  const label = STAGE_LABELS[stage] ?? 'Working…'
  return stage === 'decoding' && total
    ? `${label} ${formatCount(current)} / ${formatCount(total)}`
    : label
})
// biome-ignore lint/correctness/noUnusedVariables: used in template
const stageProgress = computed(() => {
  const { stage, current, total } = progress.value
  return stage === 'decoding' && total ? current / total : null
})

onMounted(async () => {
  try {
    const config = await playerService.getConfig()
    if (config?.maxUploadSizeBytes) maxUploadBytes.value = config.maxUploadSizeBytes
  } catch (_error) {
    // keep the default limit
  }
})

onBeforeUnmount(() => worker?.terminate())

const formatCount = (value) => (value ?? 0).toLocaleString()
// biome-ignore lint/correctness/noUnusedVariables: used in template
const formatSize = (bytes) => `${(bytes / (1024 * 1024)).toFixed(1)} MB`
// biome-ignore lint/correctness/noUnusedVariables: used in template
const percent = (value) =>
  summary.value?.players ? `${Math.round((100 * value) / summary.value.players)}%` : '–'

function reset() {
  worker?.terminate()
  worker = null
  parsing.value = false
  errorMessage.value = ''
  summary.value = null
  csvBlob.value = null
}

const ERROR_MESSAGES = {
  not_fm_save: 'This file is not a Football Manager save.',
  unsupported: 'Only Football Manager 27 saves can be read.',
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
function parseSelectedSave() {
  if (!saveFile.value) return
  reset()
  parsing.value = true
  progress.value = { stage: 'opening' }
  worker = new Worker(new URL('../workers/fmSaveWorker.js', import.meta.url), { type: 'module' })
  worker.onmessage = ({ data }) => {
    if (data.type === 'progress') {
      progress.value = data
      return
    }
    parsing.value = false
    worker.terminate()
    worker = null
    if (data.type === 'done') {
      summary.value = data.summary
      csvBlob.value = data.csv
    } else {
      errorMessage.value = `${ERROR_MESSAGES[data.kind] ?? 'The save could not be read.'} (${data.message})`
    }
  }
  worker.onerror = (event) => {
    parsing.value = false
    errorMessage.value = `The save could not be read: ${event.message || 'the browser ran out of memory or stopped the reader'}`
    worker?.terminate()
    worker = null
  }
  worker.postMessage({ file: saveFile.value })
}

// biome-ignore lint/correctness/noUnusedVariables: used in template
async function importPlayers() {
  if (!csvBlob.value) return
  uploading.value = true
  const baseName = (summary.value.saveName || saveFile.value.name).replace(/\.fm$/i, '')
  const formData = new FormData()
  // The worker gzips the CSV; /api/upload inflates `.csv.gz` (see src/api/upload_gzip.go).
  formData.append(
    'playerFile',
    new File([csvBlob.value], `${baseName}.csv.gz`, { type: 'application/gzip' })
  )
  try {
    await playerStore.uploadPlayerFile(formData, maxUploadBytes.value)
    if (playerStore.error) throw new Error(playerStore.error)
    await setManagedClubIfAbsent(playerStore.currentDatasetId, summary.value.managedClub)
    Notify.create({
      type: 'positive',
      message: 'Save imported. Opening the dataset…',
      position: 'top',
    })
    router.push(`/dataset/${playerStore.currentDatasetId}`)
  } catch (error) {
    errorMessage.value = error.message || 'The import failed.'
  } finally {
    uploading.value = false
  }
}
</script>

<style lang="scss" scoped>
.page-container {
    max-width: 900px;
    margin: 0 auto;
    padding: var(--page-gutter);
}

.upload-row,
.import-row {
    display: flex;
    gap: 1rem;
    align-items: center;
    flex-wrap: wrap;

    .save-file-input {
        flex: 1;
        min-width: 240px;
    }
}

.progress-block {
    margin-top: 1rem;
    display: grid;
    gap: 0.5rem;
}

.progress-label,
.size-note {
    color: var(--text-secondary);
    font-size: 0.9rem;
}

.summary-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 0.75rem 1.5rem;
    margin: 0 0 1rem;

    dt {
        color: var(--text-secondary);
        font-size: 0.8rem;
    }

    dd {
        margin: 0;
        font-weight: 600;
    }
}

.caveats {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin: 0 0 1rem;
    padding-left: 1.25rem;
}

@media (max-width: 768px) {
    .page-container {
        padding: var(--page-gutter-sm);
    }
}
</style>
