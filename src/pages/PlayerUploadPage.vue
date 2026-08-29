<template>
    <q-page class="upload-page">
        <div class="page-container">
            <PageHeader
                title="Upload"
                subtitle="Choose your exported HTML or CSV file from Football Manager 2024 to start analyzing your save."
                icon="cloud_upload"
            >
                <template #actions>
                    <q-btn
                        flat
                        color="info"
                        size="md"
                        label="First Time?"
                        icon="help_outline"
                        @click="showTutorial"
                    />
                </template>
            </PageHeader>

            <section class="guided-upload" aria-label="Export and upload steps">
                <div class="guided-upload__steps">
                    <span class="guided-upload__step"><strong>1.</strong> Install view</span>
                    <q-icon name="arrow_forward" aria-hidden="true" />
                    <span class="guided-upload__step"><strong>2.</strong> Export</span>
                    <q-icon name="arrow_forward" aria-hidden="true" />
                    <span class="guided-upload__step guided-upload__step--current"><strong>3.</strong> Upload</span>
                </div>
                <q-expansion-item
                    v-model="showExportGuide"
                    dense
                    icon="help_outline"
                    label="How to export from Football Manager"
                    class="export-guide"
                >
                    <div class="export-guide__content">
                        <ol>
                            <li>Subscribe to the FM Dash Search View from the Steam Workshop.</li>
                            <li>In FM, open Scouting, choose <strong>Overview → Custom → Import View</strong>, then select the view.</li>
                            <li>Select your players, press <kbd>Ctrl/Cmd + A</kbd>, then <kbd>Ctrl/Cmd + P</kbd> and save as <strong>Web Page</strong> or CSV.</li>
                        </ol>
                        <q-btn flat color="primary" icon="open_in_new" label="Open FM Dash Search View" @click="openWorkshopLink" />
                    </div>
                </q-expansion-item>
            </section>

            <!-- Upload Section -->
            <div class="upload-section">
                <SectionCard class="upload-card">
                        <div class="upload-dropzone-header">
                            <q-icon
                                name="file_upload"
                                size="2.5rem"
                                color="primary"
                                class="upload-icon"
                            />
                            <h3 class="upload-title">
                                Select Your FM24 Export File
                            </h3>
                            <p class="upload-description">
                                Choose your exported HTML or CSV file from Football
                                Manager 2024
                            </p>
                        </div>

                        <div
                            class="upload-dropzone"
                            :class="{ 'file-selected': playerFile }"
                        >
                            <div v-if="!playerFile" class="dropzone-content">
                                <q-icon
                                    name="file_upload"
                                    size="4rem"
                                    color="grey-5"
                                    class="q-mb-md"
                                />
                                <div class="dropzone-text">
                                    <div class="dropzone-primary">
                                        Drop your HTML or CSV file here or click to
                                        browse
                                    </div>
                                    <div class="dropzone-secondary">
                                        Supports .html and .csv files up to {{ maxFileSizeMB }}MB (≈{{ formatNumber(maxPlayersSupported) }} players)
                                    </div>
                                </div>
                            </div>

                            <div v-else class="file-selected-content">
                                <q-icon
                                    name="description"
                                    size="2rem"
                                    color="positive"
                                    class="q-mb-sm"
                                />
                                <div class="selected-file-name">
                                    {{ playerFile.name }}
                                </div>
                                <div class="selected-file-size">
                                    {{ formatFileSize(playerFile.size) }}
                                </div>
                                <div v-if="preflightChecking" class="preflight-status" role="status">
                                    <q-spinner size="1.1rem" /> Checking the export columns…
                                </div>
                                <div
                                    v-else-if="preflightResult"
                                    class="preflight-status"
                                    :class="preflightResult.valid ? 'preflight-status--valid' : 'preflight-status--invalid'"
                                    role="status"
                                >
                                    <q-icon :name="preflightResult.valid ? 'check_circle' : 'error_outline'" />
                                    <div>
                                        <strong>{{ preflightResult.valid ? 'Ready to upload' : 'This export needs the FM Dash Search View' }}</strong>
                                        <span>{{ preflightResult.message }}</span>
                                        <span v-if="preflightResult.missingColumns?.length" class="preflight-status__columns">
                                            Missing: {{ preflightResult.missingColumns.join(', ') }}
                                        </span>
                                        <q-btn
                                            v-if="!preflightResult.valid"
                                            flat
                                            dense
                                            color="primary"
                                            label="Open export guide"
                                            @click="showExportGuide = true"
                                        />
                                    </div>
                                </div>
                                <q-btn
                                    flat
                                    icon="close"
                                    size="sm"
                                    @click="playerFile = null"
                                    class="remove-file-btn"
                                >
                                    <q-tooltip>Remove file</q-tooltip>
                                </q-btn>
                            </div>

                            <q-file
                                v-model="playerFile"
                                accept=".html,.csv"
                                class="hidden-file-input"
                                @update:model-value="onFileSelected"
                            />
                        </div>

                        <!-- File Requirements -->
                        <div class="file-requirements">
                            <div class="requirement-item">
                                <q-icon
                                    name="check_circle"
                                    size="1.2rem"
                                    color="positive"
                                />
                                <span>HTML or CSV format</span>
                            </div>
                            <div class="requirement-item">
                                <q-icon
                                    name="check_circle"
                                    size="1.2rem"
                                    color="positive"
                                />
                                <span>Maximum {{ maxFileSizeMB }}MB file size</span>
                            </div>
                            <div class="requirement-item">
                                <q-icon
                                    name="check_circle"
                                    size="1.2rem"
                                    color="positive"
                                />
                                <span>Up to {{ formatNumber(maxPlayersSupported) }} players supported</span>
                            </div>
                        </div>

                        <div class="upload-recovery-actions">
                            <q-btn flat color="primary" icon="download" label="Download example export" href="/upload-demo.html" download="fm-dash-example-export.html" />
                            <q-btn outline color="primary" icon="sports_soccer" label="Use demo data" :loading="demoLoading" @click="loadDemoData" />
                        </div>

                        <!-- Data Retention Disclaimer -->
                        <div class="retention-disclaimer">
                            <q-icon
                                name="schedule"
                                size="1.2rem"
                                color="info"
                                class="disclaimer-icon"
                            />
                            <div class="disclaimer-text">
                                <span class="disclaimer-title">Data Retention Policy</span>
                                <span class="disclaimer-content">
                                    Uploaded data is automatically deleted after {{ datasetRetentionDays }} days for privacy and storage optimization.
                                </span>
                            </div>
                        </div>

                        <!-- Upload Button -->
                        <div class="upload-actions">
                            <q-btn
                                unelevated
                                color="primary"
                                size="lg"
                                :loading="loading"
                                :disable="
                                    !playerFile ||
                                    loading ||
                                    preflightChecking ||
                                    !preflightResult?.valid
                                "
                                @click="uploadAndParse"
                                class="upload-btn-modern"
                            >
                                <q-icon name="cloud_upload" class="q-mr-sm" />
                                {{
                                    loading
                                        ? "Processing..."
                                        : "Upload and Process"
                                }}
                            </q-btn>
                        </div>
                </SectionCard>
            </div>

            <div v-if="error" class="error-message">
                <q-icon name="error_outline" class="error-icon" />
                <div class="error-content">
                    <div class="error-text">{{ error }}</div>
                    <q-btn
                        flat
                        class="error-dismiss"
                        @click="playerStore.error = ''"
                    >
                        Dismiss
                    </q-btn>
                </div>
            </div>
        </div>

        <q-dialog v-model="showFileSizeLimitModal" persistent>
            <q-card
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-white'
                        : 'bg-white text-dark'
                "
            >
                <q-card-section class="row items-center">
                    <q-avatar
                        icon="warning"
                        color="negative"
                        text-color="white"
                    />
                    <span class="q-ml-sm text-subtitle1">File Too Large</span>
                </q-card-section>

                <q-card-section class="q-pt-none">
                    Please ensure your HTML export contains {{ formatNumber(maxPlayersSupported) }} players or
                    less. (Max file size: {{ maxFileSizeMB }}MB)
                </q-card-section>

                <q-card-actions align="right">
                    <q-btn
                        flat
                        label="OK"
                        color="primary"
                        v-close-popup
                        @click="showFileSizeLimitModal = false"
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>

        <InteractiveUploadLoader 
          v-bind="loaderProps" 
          @cancel="handleUploadCancel" 
        />


    </q-page>
</template>

<script>
import { Notify, useQuasar } from 'quasar'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import InteractiveUploadLoader from '../components/InteractiveUploadLoader.vue'
import PageHeader from '../components/layout/PageHeader.vue'
import SectionCard from '../components/layout/SectionCard.vue'
import playerService from '../services/playerService.js'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'

export default {
  name: 'PlayerUploadPage',
  components: {
    InteractiveUploadLoader,
    PageHeader,
    SectionCard,
  },
  setup() {
    const router = useRouter()
    const route = useRoute()
    const playerStore = usePlayerStore()
    const uiStore = useUiStore()
    const _$q = useQuasar()
    const playerFile = ref(null)
    const showFileSizeLimitModal = ref(false)
    const uploadProgress = ref(0)
    const preflightResult = ref(null)
    const preflightChecking = ref(false)
    const demoLoading = ref(false)
    const showExportGuide = ref(true)

    const maxFileSizeBytes = ref(15 * 1024 * 1024)
    const maxFileSizeMB = ref(15)
    const largeFileSizeBytes = ref(20 * 1024 * 1024)
    const datasetRetentionDays = ref(30)

    const loading = computed(() => playerStore.loading)
    const error = computed({
      get: () => playerStore.error,
      set: (value) => {
        playerStore.error = value
      },
    })

    // Computed values for the interactive loader
    const loaderProps = computed(() => ({
      visible: loading.value,
      filename: playerFile.value?.name || '',
      fileSize: playerFile.value ? formatFileSize(playerFile.value.size) : '',
      playersFound: 0, // Could be enhanced to show real-time player count
      progress: uploadProgress.value,
      dataReady: dataReady.value,
    }))

    const dataReady = ref(false)

    onMounted(async () => {
      // Fetch config first with cache clearing
      try {
        const config = await playerService.getConfig(true) // Clear cache to get fresh config
        maxFileSizeBytes.value = config.maxUploadSizeBytes
        maxFileSizeMB.value = config.maxUploadSizeMB
        datasetRetentionDays.value = config.datasetRetentionDays || 30
      } catch (_error) {
        console.error('Error fetching config:', _error)
      }

      if (route.query.demo === 'true') {
        await loadDemoData()
      }
    })

    const formatFileSize = (bytes) => {
      if (bytes === 0) return '0 Bytes'
      const k = 1024
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return `${Number.parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`
    }

    const inspectExportInBrowser = async (file) => {
      const content = await file.text()
      if (file.name.toLowerCase().endsWith('.csv')) {
        const headerLine = content.split(/\r?\n/).find((line) => line.trim())
        return headerLine
          ? headerLine
              .split(';')
              .map((header) => header.trim())
              .filter(Boolean)
          : []
      }

      const document = new DOMParser().parseFromString(content, 'text/html')
      const firstRow = document.querySelector('table tr')
      return firstRow
        ? [...firstRow.querySelectorAll('th, td')]
            .map((cell) => cell.textContent.trim())
            .filter(Boolean)
        : []
    }

    const preflightFile = async (file) => {
      preflightResult.value = null
      if (!file) return

      try {
        const headers = await inspectExportInBrowser(file)
        if (headers.length === 0) {
          preflightResult.value = {
            valid: false,
            missingColumns: [],
            message: 'We could not find a player table with column headings in this export.',
          }
          return
        }

        preflightChecking.value = true
        preflightResult.value = await playerService.preflightPlayerFile(file)
      } catch (error) {
        preflightResult.value = {
          valid: false,
          missingColumns: [],
          message:
            error.message ||
            'We could not check this export. Try exporting it again from Football Manager.',
        }
      } finally {
        preflightChecking.value = false
      }
    }

    const onFileSelected = async (file) => {
      if (file && file.size > maxFileSizeBytes.value) {
        showFileSizeLimitModal.value = true
        playerFile.value = null
        preflightResult.value = null
        return
      }
      await preflightFile(file)
    }

    const loadDemoData = async () => {
      demoLoading.value = true
      try {
        const response = await fetch('/upload-demo.html')
        if (!response.ok) throw new Error('Demo data is unavailable')
        const blob = await response.blob()
        playerFile.value = new File([blob], 'fm-dash-demo.html', { type: 'text/html' })
        await preflightFile(playerFile.value)
      } catch (error) {
        Notify.create({
          type: 'negative',
          message: error.message || 'Failed to load demo data. Please try again.',
          position: 'top',
          timeout: 5000,
        })
      } finally {
        demoLoading.value = false
      }
    }

    const handleUploadCancel = () => {
      // Reset the upload state
      playerStore.loading = false
      uploadProgress.value = 0
      dataReady.value = false
      Notify.create({
        type: 'info',
        message: 'Upload cancelled',
        position: 'top',
        timeout: 2000,
      })
    }

    const updateProgress = (progress) => {
      // If progress is from file upload (0-100 range), scale to 70%
      // If progress is explicit stage progress (80, 95, 100), use directly
      if (progress <= 100 && uploadProgress.value < 70) {
        // This is file upload progress
        uploadProgress.value = Math.min(progress * 0.7, 70)
      } else {
        // This is explicit stage progress from the store
        uploadProgress.value = Math.min(progress, 100)
      }
    }

    const uploadAndParse = async () => {
      if (!playerFile.value) {
        playerStore.error = 'Please select an HTML or CSV export first.'
        return
      }

      if (!preflightResult.value?.valid) {
        playerStore.error = 'Check your export columns before uploading.'
        showExportGuide.value = true
        return
      }
      if (playerFile.value.size > maxFileSizeBytes.value) {
        showFileSizeLimitModal.value = true
        return
      }

      const _isLargeFile = playerFile.value.size > largeFileSizeBytes.value

      // Reset progress
      uploadProgress.value = 0
      dataReady.value = false

      try {
        const formData = new FormData()
        formData.append('playerFile', playerFile.value)

        // The player store will now handle all progress stages
        const response = await playerStore.uploadPlayerFile(
          formData,
          maxFileSizeBytes.value,
          updateProgress
        )

        if (!playerStore.error) {
          // Check if this was a duplicate upload by looking at the response message
          const isDuplicate = response.message?.includes('Duplicate file detected')

          const successMessage = isDuplicate
            ? 'Duplicate file detected! Redirecting to existing dataset...'
            : 'File uploaded and parsed successfully! Redirecting to dataset view...'

          Notify.create({
            type: 'positive',
            message: successMessage,
            position: 'top',
            timeout: isDuplicate ? 3000 : 2000,
          })

          // Always redirect to dataset page regardless of file size
          setTimeout(() => {
            if (playerStore.currentDatasetId) {
              router.push(`/dataset/${playerStore.currentDatasetId}`)
            }
          }, 500)
        }
      } catch (e) {
        uploadProgress.value = 0
        if (playerStore.error) {
          Notify.create({
            type: 'negative',
            message: playerStore.error,
            position: 'top',
            timeout: 5000,
            actions: [{ label: 'Dismiss', color: 'white' }],
          })
        } else {
          Notify.create({
            type: 'negative',
            message: `Upload failed: ${e.message}`,
            position: 'top',
            timeout: 5000,
            actions: [{ label: 'Dismiss', color: 'white' }],
          })
        }
      } finally {
        setTimeout(() => {
          uploadProgress.value = 0
        }, 2000)
      }
    }

    const formatNumber = (value) => {
      return Number(value).toLocaleString()
    }

    const showTutorial = () => {
      uiStore.showTutorial()
    }

    const openWorkshopLink = () => {
      window.open('https://steamcommunity.com/sharedfiles/filedetails/?id=3498467200', '_blank')
    }

    const maxPlayersSupported = computed(() => {
      // Rule: 15MB = 10,000 players
      // So players = (current max file size in MB / 15MB) * 10,000
      const exactPlayers = Math.floor((maxFileSizeMB.value / 20) * 10000)

      // Round to nearest 5000 for cleaner display
      // For example: 66,666 becomes 65,000
      return Math.floor(exactPlayers / 5000) * 5000
    })

    return {
      playerFile,
      showFileSizeLimitModal,
      loading,
      error,
      uploadAndParse,
      formatFileSize,
      onFileSelected,
      playerStore,
      uiStore,
      preflightResult,
      preflightChecking,
      demoLoading,
      showExportGuide,

      maxFileSizeMB,
      loaderProps,
      handleUploadCancel,
      formatNumber,
      maxPlayersSupported,
      datasetRetentionDays,
      showTutorial,
      loadDemoData,
      openWorkshopLink,
    }
  },
}
</script>

<style lang="scss" scoped>
.upload-page {
    min-height: calc(100vh - 120px);
    background: var(--surface-page);
}

.page-container {
    max-width: 900px;
    margin: 0 auto;
    padding: var(--page-gutter);
}

.guided-upload {
    margin-bottom: 1.5rem;
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-md);
    background: var(--surface-card);
}

.guided-upload__steps {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.7rem;
    padding: 1rem;
    color: var(--text-secondary);
}

.guided-upload__step strong {
    color: var(--accent);
}

.guided-upload__step--current {
    color: var(--text-primary);
    font-weight: 600;
}

.export-guide {
    border-top: 1px solid var(--surface-border);
    color: var(--text-primary);
}

.export-guide__content {
    padding: 0 1rem 1rem;
    color: var(--text-secondary);

    ol {
        margin: 0 0 0.75rem;
        padding-left: 1.4rem;
    }

    li + li {
        margin-top: 0.45rem;
    }

    kbd {
        border: 1px solid var(--surface-border-strong);
        border-radius: 3px;
        padding: 0.1rem 0.25rem;
        color: var(--text-primary);
    }
}

// Upload Section
.upload-section {
    margin-bottom: 2rem;

    .upload-card {
        position: relative;

        .upload-dropzone-header {
            text-align: center;
            margin-bottom: 2rem;

            .upload-icon {
                margin-bottom: 1rem;
            }

            .upload-title {
                font-size: 1.5rem;
                font-weight: 500;
                color: var(--text-primary);
                margin: 0 0 0.5rem 0;
            }

            .upload-description {
                color: var(--text-secondary);
                font-size: 1rem;
                margin: 0;
            }
        }

        .upload-dropzone {
            border: 2px dashed color-mix(in srgb, var(--accent) 30%, transparent);
            border-radius: var(--radius-md);
            padding: 3rem 2rem;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-bottom: 2rem;
            position: relative;

            &:hover {
                border-color: color-mix(in srgb, var(--accent) 50%, transparent);
                background: var(--accent-soft);
            }

            &.file-selected {
                border-color: #4caf50;
                background: rgba(76, 175, 80, 0.05);

                .body--dark & {
                    background: rgba(76, 175, 80, 0.1);
                }
            }

            .dropzone-content {
                .dropzone-text {
                    .dropzone-primary {
                        font-size: 1.1rem;
                        color: var(--text-primary);
                        font-weight: 500;
                        margin-bottom: 0.5rem;
                    }

                    .dropzone-secondary {
                        color: var(--text-secondary);
                        font-size: 0.9rem;
                    }
                }
            }

            .file-selected-content {
                .selected-file-name {
                    font-size: 1.1rem;
                    font-weight: 500;
                    color: #4caf50;
                    margin-bottom: 0.25rem;
                }

                .selected-file-size {
                    color: var(--text-secondary);
                    font-size: 0.9rem;
                    margin-bottom: 1rem;
                }

                .remove-file-btn {
                    position: absolute;
                    top: 1rem;
                    right: 1rem;
                }
            }

            .preflight-status {
                display: flex;
                align-items: flex-start;
                justify-content: center;
                gap: 0.45rem;
                margin: 0.8rem auto 0;
                max-width: 44rem;
                color: var(--text-secondary);

                > div {
                    display: flex;
                    flex-direction: column;
                    align-items: flex-start;
                    gap: 0.2rem;
                    text-align: left;
                }

                &--valid {
                    color: #2e7d32;
                }

                &--invalid {
                    color: var(--negative, #c62828);
                }
            }

            .preflight-status__columns {
                font-size: 0.85rem;
            }

            .hidden-file-input {
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                opacity: 0;
                cursor: pointer;

                :deep(.q-field__inner) {
                    display: none;
                }
            }
        }

        .file-requirements {
            display: flex;
            justify-content: center;
            gap: 2rem;
            margin-bottom: 2rem;

            .requirement-item {
                display: flex;
                align-items: center;
                gap: 0.5rem;
                font-size: 0.9rem;
                color: var(--text-secondary);
            }
        }

        .upload-recovery-actions {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            gap: 0.75rem;
            margin: 0.5rem 0 1.5rem;
        }

        .retention-disclaimer {
            display: flex;
            align-items: flex-start;
            gap: 0.75rem;
            background: var(--accent-soft);
            border: 1px solid var(--surface-border);
            border-radius: var(--radius-sm);
            padding: 1rem;
            margin-bottom: 2rem;

            .disclaimer-icon {
                margin-top: 0.1rem;
                flex-shrink: 0;
            }

            .disclaimer-text {
                display: flex;
                flex-direction: column;
                gap: 0.25rem;

                .disclaimer-title {
                    font-weight: 600;
                    font-size: 0.9rem;
                    color: var(--accent);
                }

                .disclaimer-content {
                    font-size: 0.85rem;
                    color: var(--text-secondary);
                    line-height: 1.4;
                }
            }
        }

        .upload-actions {
            text-align: center;

            .upload-btn-modern {
                padding: 1rem 2rem;
                font-weight: 500;
                letter-spacing: 0.5px;
                border-radius: var(--radius-md);
                min-width: 200px;
                transition: all 0.3s ease;

                &:hover {
                    transform: translateY(-2px);
                    box-shadow: 0 8px 25px color-mix(in srgb, var(--accent) 30%, transparent);
                }

                &:disabled {
                    transform: none;
                    box-shadow: none;
                }
            }
        }
    }
}

.error-message {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    background: color-mix(in srgb, #dc3545 8%, var(--surface-card));
    border: 1px solid color-mix(in srgb, #dc3545 25%, transparent);
    border-radius: var(--radius-sm);
    padding: 1rem;
    margin-top: 1rem;
}

.error-icon {
    color: #dc3545;
    font-size: 1.2rem;
    margin-top: 0.1rem;
}

.error-content {
    flex: 1;
}

.error-text {
    color: var(--text-primary);
    margin-bottom: 0.5rem;
}

.error-dismiss {
    color: #dc3545;
    font-size: 0.85rem;
    padding: 0.25rem 0.5rem;

    &:hover {
        background: rgba(220, 53, 69, 0.1);
    }
}

// Responsive Design
@media (max-width: 768px) {
    .guided-upload__steps {
        gap: 0.35rem;
        font-size: 0.82rem;
    }

    .page-container {
        padding: var(--page-gutter-sm);
    }

    .upload-section {
        .upload-card {
            .upload-dropzone {
                padding: 2rem 1rem;
            }

            .file-requirements {
                flex-direction: column;
                gap: 1rem;
                text-align: left;
            }

            .retention-disclaimer {
                .disclaimer-text {
                    .disclaimer-title {
                        font-size: 0.85rem;
                    }

                    .disclaimer-content {
                        font-size: 0.8rem;
                    }
                }
            }

            .upload-btn-modern {
                min-width: 100%;
            }
        }
    }
}

@media (max-width: 480px) {
    .upload-section {
        .upload-card {
            .upload-dropzone-header {
                .upload-title {
                    font-size: 1.2rem;
                }

                .upload-description {
                    font-size: 0.9rem;
                }
            }
        }
    }
}
</style>
