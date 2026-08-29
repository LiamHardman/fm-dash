<template>
  <div v-if="entries.length" class="dataset-switcher">
    <q-select dense outlined options-dense emit-value map-options label="Active Dataset" :options="options" :model-value="playerStore.currentDatasetId" @update:model-value="switchDataset" />
    <q-btn flat round dense icon="more_vert" aria-label="Dataset actions"><q-menu><q-list dense>
      <q-item clickable v-close-popup @click="rename"><q-item-section avatar><q-icon name="edit" /></q-item-section><q-item-section>Rename</q-item-section></q-item>
      <q-item clickable v-close-popup @click="remove"><q-item-section avatar><q-icon name="delete" /></q-item-section><q-item-section>Remove from history</q-item-section></q-item>
    </q-list></q-menu></q-btn>
  </div>
</template>
<script setup>
import { useQuasar } from 'quasar'
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/playerStore'
import { useRecentDatasetsStore } from '../stores/recentDatasetsStore'

const $q = useQuasar()
const router = useRouter()
const playerStore = usePlayerStore()
const recent = useRecentDatasetsStore()
const entries = computed(() => recent.entries)
const options = computed(() =>
  entries.value.map((e) => ({
    value: e.datasetId,
    label: e.label,
    meta: `${e.playerCount || 0} players · ${new Date(e.uploadedAt).toLocaleDateString()}`,
  }))
)
function switchDataset(id) {
  playerStore.setCurrentDatasetId(id)
  router.push(`/dataset/${id}`)
}
function rename() {
  const current = entries.value.find((e) => e.datasetId === playerStore.currentDatasetId)
  if (!current) return
  $q.dialog({
    title: 'Rename Dataset',
    prompt: { model: current.label, type: 'text' },
    cancel: true,
  }).onOk((value) => recent.renameDataset(current.datasetId, value))
}
function remove() {
  const id = playerStore.currentDatasetId
  if (!id) return
  recent.removeDataset(id)
  playerStore.setCurrentDatasetId(null)
  router.push('/upload')
}
onMounted(() => recent.load())
</script>
<style scoped>
.dataset-switcher { display: flex; align-items: center; gap: 4px; min-width: 260px; max-width: 360px; }
@media (max-width: 768px) { .dataset-switcher { min-width: 180px; max-width: 220px; } }
</style>
