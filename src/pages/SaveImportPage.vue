<template>
    <q-page class="save-import-page">
        <PageHero
            badge="Experimental"
            icon="science"
            title="FM Save"
            highlight="Import"
            subtitle="Upload a native Football Manager .fm save file to see whatever basic information can be extracted from it. This format is undocumented and proprietary, so results are raw and best-effort -- not a curated report."
        />

        <div class="save-import-container">
            <q-card class="save-import-card" flat bordered>
                <q-card-section>
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
                </q-card-section>
            </q-card>

            <q-card v-if="result" class="save-import-card q-mt-md" flat bordered>
                <q-card-section>
                    <div class="results-header">
                        <h3>Extracted fields</h3>
                        <span class="results-count">{{ result.fields?.length || 0 }} found</span>
                    </div>

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
                </q-card-section>
            </q-card>
        </div>
    </q-page>
</template>

<script setup>
import { computed, ref } from 'vue'
import PageHero from '../components/PageHero.vue'

const saveFile = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const result = ref(null)

const columns = [
    { name: 'offset', label: 'Offset', field: 'offset', align: 'left', sortable: true },
    { name: 'value', label: 'Value', field: 'value', align: 'left', sortable: true },
]

const tableRows = computed(() => result.value?.fields || [])

const versionField = computed(() =>
    tableRows.value.find((f) => /^\d+\.\d+\.\d+/.test(f.value))
)

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
.save-import-container {
    max-width: 900px;
    margin: 0 auto;
    padding: 1.5rem;
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
    color: var(--text-secondary, #666);
}

.results-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;

    h3 {
        margin: 0;
    }

    .results-count {
        color: var(--text-secondary, #666);
        font-size: 0.9rem;
    }
}

.results-note {
    color: var(--text-secondary, #666);
    font-size: 0.9rem;
    margin: 0.75rem 0 1rem;
}
</style>
