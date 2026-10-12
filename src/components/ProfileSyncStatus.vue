<template>
  <div class="q-my-sm" role="status" aria-live="polite">
    <span v-if="store.sync.status === 'saved'" class="text-caption">All changes saved</span>
    <span v-else-if="store.sync.status === 'saving' || store.sync.status === 'pending'">Changes pending…</span>
    <q-banner v-else-if="store.sync.status === 'failed' || store.sync.status === 'conflict'" rounded class="bg-orange-1 text-orange-10">
      {{ store.sync.error || 'Your browser has unsaved changes.' }}
      <div>{{ store.sync.localDurable ? 'Your local draft is retained in this browser.' : 'Browser storage is unavailable. Keep this page open to avoid losing edits.' }}</div>
      <q-btn v-if="store.sync.status === 'failed'" flat label="Retry" @click="retry" />
      <q-btn v-if="store.sync.status === 'conflict'" flat label="Review versions" @click="showConflict = true" />
    </q-banner>
    <q-btn v-if="store.sync.status === 'pending'" flat dense label="Save pending changes" @click="retry" />
    <q-dialog v-model="showConflict">
      <q-card style="width: 800px; max-width: 95vw">
        <q-card-section><div class="text-h6">Choose which version to keep</div><p>Keeping this browser's draft replaces the server document. Download either version before choosing if you want to merge edits manually.</p></q-card-section>
        <q-card-section class="row q-col-gutter-md">
          <div class="col-12 col-sm-6"><strong>This browser</strong><pre>{{ JSON.stringify(localDocument, null, 2) }}</pre><q-btn flat label="Download local version" @click="download(localDocument, 'local')" /></div>
          <div class="col-12 col-sm-6"><strong>Server</strong><pre>{{ JSON.stringify(store.sync.conflict?.document, null, 2) }}</pre><q-btn flat label="Download server version" @click="download(store.sync.conflict.document, 'server')" /></div>
        </q-card-section>
        <q-card-actions align="right"><q-btn flat label="Cancel" v-close-popup /><q-btn flat label="Use server version" @click="resolve('server')" /><q-btn color="primary" label="Keep local draft" @click="resolve('local')" /></q-card-actions>
      </q-card>
    </q-dialog>
  </div>
</template>
<script setup>
import { useQuasar } from 'quasar'
import { computed, ref } from 'vue'

const props = defineProps({ store: { type: Object, required: true } })
const $q = useQuasar()
const showConflict = ref(false)
const localDocument = computed(
  () => props.store.document || { version: 1, projects: props.store.projects }
)
async function retry() {
  try {
    if (!props.store.loaded) await props.store.load()
    else await props.store.retry()
  } catch (error) {
    $q.notify({ type: 'negative', message: error.message })
  }
}
async function resolve(choice) {
  try {
    await props.store.resolveConflict(choice)
    showConflict.value = false
  } catch (error) {
    $q.notify({ type: 'negative', message: error.message })
  }
}
function download(value, name) {
  const url = URL.createObjectURL(
    new Blob([JSON.stringify(value, null, 2)], { type: 'application/json' })
  )
  const a = document.createElement('a')
  a.href = url
  a.download = `${name}-profile-document.json`
  a.click()
  URL.revokeObjectURL(url)
}
</script>
<style scoped>pre { max-height: 300px; overflow: auto; white-space: pre-wrap; overflow-wrap: anywhere; }</style>
