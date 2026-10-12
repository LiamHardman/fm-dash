<template>
  <q-btn flat icon="backup" label="Profile backup" @click="open" />
  <q-dialog v-model="visible">
    <q-card style="width: 720px; max-width: 95vw">
      <q-card-section><div class="text-h6">Installation-profile backup</div><p>Back up Shortlists, recruitment scenarios, Saved searches, and Progression projects. Uploaded Dataset player data is excluded; expired Player references remain unavailable after restore.</p></q-card-section>
      <q-card-section>
        <q-banner v-if="error" class="bg-red-1 text-negative q-mb-md" role="alert">{{ error }}</q-banner>
        <q-btn label="Download backup" icon="download" :disable="busy || !ready" @click="exportBackup" />
        <q-file class="q-mt-md" outlined label="Choose backup to restore" accept="application/json,.json" :max-file-size="6291456" :disable="busy || !ready" @update:model-value="inspect" />
        <template v-if="incoming">
          <q-toggle v-model="replaceConflicts" label="Replace records with matching IDs (otherwise keep existing)" />
          <p>Preview: {{ preview.changes.length }} incoming records. Existing records absent from the backup are retained.</p>
          <q-list bordered style="max-height: 300px; overflow: auto"><q-item v-for="change in preview.changes" :key="`${change.section}:${change.id}`"><q-item-section>{{ change.name }}<q-item-label caption>{{ change.section }} · {{ change.action }}</q-item-label></q-item-section></q-item></q-list>
          <p class="text-caption">Each collection is saved separately. If a save fails, successful collections remain restored and unsaved drafts stay available for recovery.</p>
          <q-btn class="q-mt-md" color="primary" label="Restore previewed records" :loading="busy" @click="restore" />
        </template>
        <p v-if="message" role="status">{{ message }}</p>
      </q-card-section>
      <q-card-actions align="right"><q-btn flat label="Close" :disable="busy" v-close-popup /></q-card-actions>
    </q-card>
  </q-dialog>
</template>
<script setup>
import { computed, ref } from 'vue'
import { useProgressionProjectStore } from '../stores/progressionProjectStore'
import { useSavedSearchStore } from '../stores/savedSearchStore'
import { useShortlistStore } from '../stores/shortlistStore'
import { apiUrl } from '../utils/apiEndpoint'
import { mergeProfileBackup } from '../utils/profileBackup'

const lists = useShortlistStore(),
  searches = useSavedSearchStore(),
  projects = useProgressionProjectStore()
const visible = ref(false),
  busy = ref(false),
  error = ref(''),
  message = ref(''),
  incoming = ref(null),
  replaceConflicts = ref(false)
const ready = computed(() => lists.loaded && searches.loaded && projects.loaded)
const current = () => ({
  version: 1,
  shortlists: lists.document,
  savedSearches: searches.document,
  progressionProjects: { version: 1, projects: projects.projects },
})
const preview = computed(() =>
  incoming.value ? mergeProfileBackup(current(), incoming.value, replaceConflicts.value) : null
)
async function open() {
  visible.value = true
  error.value = ''
  busy.value = true
  try {
    await Promise.all([lists.load(), searches.load(), projects.load()])
    if (!ready.value)
      throw new Error('Load all profile collections before backing up or restoring.')
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
async function validate(value) {
  const response = await fetch(apiUrl('/api/profile/validate-backup'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(value),
  })
  const result = await response.json()
  if (!response.ok) throw new Error(result.message || 'Invalid backup')
  return result
}
async function inspect(file) {
  incoming.value = null
  error.value = ''
  message.value = ''
  if (!file) return
  busy.value = true
  try {
    if (file.size > 6 * 1024 * 1024) throw new Error('Backup must be smaller than 6 MB')
    incoming.value = await validate(JSON.parse(await file.text()))
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
async function exportBackup() {
  busy.value = true
  error.value = ''
  try {
    const data = await validate(current())
    const url = URL.createObjectURL(
      new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    )
    const link = document.createElement('a')
    link.href = url
    link.download = 'fm-dash-profile.json'
    link.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
async function restore() {
  busy.value = true
  error.value = ''
  message.value = ''
  const completed = []
  try {
    const data = await validate(preview.value.document)
    for (const [name, store, document] of [
      ['Shortlists', lists, data.shortlists],
      ['Saved searches', searches, data.savedSearches],
      ['Progression projects', projects, data.progressionProjects],
    ]) {
      if (!(await store.replaceDocument(document)))
        throw new Error(`${name} could not be saved. Review its save status before retrying.`)
      completed.push(name)
    }
    incoming.value = null
    message.value = 'Profile restored.'
  } catch (e) {
    error.value = `${e.message} ${completed.length ? `Already restored: ${completed.join(', ')}.` : 'No collections confirmed restored.'}`
  } finally {
    busy.value = false
  }
}
</script>
