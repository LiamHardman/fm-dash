<template>
    <q-page class="save-import-page">
        <div class="page-container">
            <PageHeader
                title="FM Save Import"
                subtitle="Upload a native Football Manager .fm save file to see whatever basic information can be extracted from it. This format is undocumented and proprietary, so results are raw and best-effort -- not a curated report."
                icon="science"
            >
                <template #actions>
                    <q-chip dense color="warning" text-color="dark" icon="warning" label="Experimental" />
                </template>
            </PageHeader>

            <SectionCard title="Import a save" icon="upload_file">
                <div class="upload-row">
                    <q-file
                        v-model="saveFile"
                        label="Choose .fm save file"
                        accept=".fm"
                        filled
                        class="save-file-input"
                        :disable="loading"
                    >
                        <template #prepend>
                            <q-icon name="attach_file" />
                        </template>
                    </q-file>
                    <q-btn
                        color="primary"
                        label="Import"
                        icon="cloud_upload"
                        :disable="!saveFile || loading"
                        :loading="loading"
                        @click="importSave"
                    />
                </div>

                <q-banner v-if="errorMessage" class="bg-negative text-white q-mt-md" rounded>
                    {{ errorMessage }}
                </q-banner>

                <div v-if="loading" class="loading-note">
                    <q-spinner color="primary" size="1.5rem" />
                    <span>Decompressing and scanning the save -- this can take a little while for large files.</span>
                </div>
            </SectionCard>

            <SectionCard v-if="result" title="Extracted fields" icon="list_alt" class="q-mt-md">
                <template #actions>
                    <span class="results-count">{{ result.fields?.length || 0 }} found</span>
                </template>

                <q-banner v-if="versionField" class="bg-positive text-white q-mb-md" rounded>
                    Save/engine version: <strong>{{ versionField.value }}</strong>
                </q-banner>

                <p class="results-note">
                    These are raw strings recovered by scanning near the start of the
                    decompressed save for the one confirmed encoding primitive (a
                    length-prefixed ASCII string). There is no known record schema for this
                    format, so most entries are internal engine bookkeeping (type names,
                    patch/migration identifiers) rather than curated save info -- see the
                    version field above for the one field proven meaningful so far.
                </p>

                <q-table
                    :rows="tableRows"
                    :columns="columns"
                    row-key="offset"
                    flat
                    dense
                    :pagination="{ rowsPerPage: 25 }"
                />
            </SectionCard>
        </div>
    </q-page>
</template>

<script setup>
import { computed, ref } from 'vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import PageHeader from '../components/layout/PageHeader.vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import SectionCard from '../components/layout/SectionCard.vue'

const saveFile = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const result = ref(null)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const columns = [
  { name: 'offset', label: 'Offset', field: 'offset', align: 'left', sortable: true },
  { name: 'value', label: 'Value', field: 'value', align: 'left', sortable: true },
]

const tableRows = computed(() => result.value?.fields || [])

// biome-ignore lint/correctness/noUnusedVariables: used in template
const versionField = computed(() => tableRows.value.find((f) => /^\d+\.\d+\.\d+/.test(f.value)))

// biome-ignore lint/correctness/noUnusedVariables: used in template
async function importSave() {
  if (!saveFile.value) return
  loading.value = true
  errorMessage.value = ''
  result.value = null

  try {
    const formData = new FormData()
    formData.append('saveFile', saveFile.value)

    const response = await fetch('/api/fm-save-import', {
      method: 'POST',
      body: formData,
    })

    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || `Request failed with status ${response.status}`)
    }

    result.value = await response.json()
  } catch (err) {
    errorMessage.value = err.message || 'Failed to import save file'
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.page-container {
    max-width: 900px;
    margin: 0 auto;
    padding: var(--page-gutter);
}

.upload-row {
    display: flex;
    gap: 1rem;
    align-items: flex-start;

    .save-file-input {
        flex: 1;
    }
}

.loading-note {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 1rem;
    color: var(--text-secondary);
}

.results-count {
    color: var(--text-secondary);
    font-size: 0.9rem;
}

.results-note {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin: 0 0 1rem;
}

@media (max-width: 768px) {
    .page-container {
        padding: var(--page-gutter-sm);
    }
}
</style>
