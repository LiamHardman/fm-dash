<template>
    <q-page class="upload-page">
        <div class="upload-container">
            <!-- Hero Section -->
            <div class="hero-section">
            </div>

            <!-- Upload Section -->
            <div class="upload-section">
                <q-card class="upload-card" flat bordered>
                    <q-card-section>
                        <!-- Help Section -->
                        <div class="help-section">
                            <q-btn
                                flat
                                color="info"
                                size="md"
                                label="First Time?"
                                icon="help_outline"
                                class="help-btn"
                            />
                        </div>
                        
                        <div class="upload-header">
                            <q-icon
                                name="cloud_upload"
                                size="3rem"
                                color="primary"
                                class="upload-icon"
                            />
                            <h3 class="upload-title">
                                Select Your FM24 Export File
                            </h3>
                            <p class="upload-description">
                                Choose your exported HTML file from Football
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
                                        Drop your HTML file here or click to
                                        browse
                                    </div>
                                    <div class="dropzone-secondary">
                                        Supports .html files up to {{ maxFileSizeMB }}MB (≈{{ formatNumber(maxPlayersSupported) }} players)
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
                                accept=".html"
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
                                <span>HTML format only</span>
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
                                    loading
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
                    </q-card-section>
                </q-card>
            </div>



            <!-- Notification Preferences - Subtle -->
            <div class="notification-preferences" v-if="notificationSupported">
                <div class="subtle-preference">
                    <q-toggle
                        v-model="uiStore.notificationsEnabled"
                        @update:model-value="uiStore.toggleNotifications"
                        color="primary"
                        size="sm"
                    />
                    <div class="preference-text">
                        <span class="preference-label"
                            >Desktop notifications</span
                        >
                        <span class="preference-hint"
                            >Get notified when large uploads finish</span
                        >
                    </div>
                </div>
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
import { useWebNotification } from '@vueuse/core'
import { Notify, useQuasar } from 'quasar'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import InteractiveUploadLoader from '../components/InteractiveUploadLoader.vue'
import playerService from '../services/playerService'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'

export default {
  name: 'PlayerUploadPage',
  components: {
    InteractiveUploadLoader
  },
  setup() {
    const router = useRouter()
    const playerStore = usePlayerStore()
    const uiStore = useUiStore()
    const _$q = useQuasar()
    const playerFile = ref(null)
    const showFileSizeLimitModal = ref(false)
    const uploadProgress = ref(0)

    const maxFileSizeBytes = ref(15 * 1024 * 1024)
    const maxFileSizeMB = ref(15)
    const largeFileSizeBytes = ref(20 * 1024 * 1024)
    const datasetRetentionDays = ref(30)

    const {
      isSupported: notificationSupported,
      permissionGranted,
      show: showNotification
    } = useWebNotification({
      title: 'FM24 Data Processing Complete',
      body: 'Your large file has been processed and is ready to view!',
      icon: '/favicon.ico',
      tag: 'upload-complete'
    })

    const loading = computed(() => playerStore.loading)
    const error = computed({
      get: () => playerStore.error,
      set: value => {
        playerStore.error = value
      }
    })

    // Computed values for the interactive loader
    const loaderProps = computed(() => ({
      visible: loading.value,
      filename: playerFile.value?.name || '',
      fileSize: playerFile.value ? formatFileSize(playerFile.value.size) : '',
      playersFound: 0, // Could be enhanced to show real-time player count
      progress: uploadProgress.value
    }))

    onMounted(async () => {
      // Fetch config first
      try {
        const config = await playerService.getConfig()
        maxFileSizeBytes.value = config.maxUploadSizeBytes
        maxFileSizeMB.value = config.maxUploadSizeMB
        datasetRetentionDays.value = config.datasetRetentionDays || 30
      } catch (error) {
        console.error('Error loading config in upload page:', error)
      }
      // Initialize UI preferences
      uiStore.initNotifications()
    })

    const formatFileSize = bytes => {
      if (bytes === 0) return '0 Bytes'
      const k = 1024
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return `${Number.parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`
    }

    const onFileSelected = file => {
      if (file && file.size > maxFileSizeBytes.value) {
        showFileSizeLimitModal.value = true
        playerFile.value = null
      }
    }

    const handleUploadCancel = () => {
      // Reset the upload state
      playerStore.loading = false
      uploadProgress.value = 0
      Notify.create({
        type: 'info',
        message: 'Upload cancelled',
        position: 'top',
        timeout: 2000
      })
    }

    const updateProgress = progress => {
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
        playerStore.error = 'Please select an HTML file first.'
        return
      }
      if (playerFile.value.size > maxFileSizeBytes.value) {
        showFileSizeLimitModal.value = true
        return
      }

      const isLargeFile = playerFile.value.size > largeFileSizeBytes.value

      // Reset progress
      uploadProgress.value = 0

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
            timeout: isDuplicate ? 3000 : 2000
          })

          // Show web notification for large files if enabled and supported (but not for duplicates)
          if (
            !isDuplicate &&
            isLargeFile &&
            uiStore.notificationsEnabled &&
            notificationSupported.value &&
            permissionGranted.value
          ) {
            showNotification()
          }

          // Redirect to the dataset page
          setTimeout(() => {
            if (playerStore.currentDatasetId) {
              router.push(`/dataset/${playerStore.currentDatasetId}`)
            }
          }, 1000)
        }
      } catch (e) {
        uploadProgress.value = 0
        if (playerStore.error) {
          Notify.create({
            type: 'negative',
            message: playerStore.error,
            position: 'top',
            timeout: 5000,
            actions: [{ label: 'Dismiss', color: 'white' }]
          })
        } else {
          Notify.create({
            type: 'negative',
            message: `Upload failed: ${e.message}`,
            position: 'top',
            timeout: 5000,
            actions: [{ label: 'Dismiss', color: 'white' }]
          })
        }
      } finally {
        setTimeout(() => {
          uploadProgress.value = 0
        }, 2000)
      }
    }

    const formatNumber = value => {
      return Number(value).toLocaleString()
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
      notificationSupported,
      maxFileSizeMB,
      loaderProps,
      handleUploadCancel,
      formatNumber,
      maxPlayersSupported,
      datasetRetentionDays
    }
  }
}
</script>

<style lang="scss" scoped>
.upload-page {
    min-height: calc(100vh - 120px);
    background: linear-gradient(135deg, #f8f9fc 0%, #ffffff 100%);

    .body--dark & {
        background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
    }
}

.upload-container {
    max-width: 900px;
    margin: 0 auto;
    padding: 2rem;
}

// Hero Section
.hero-section {
    text-align: center;
    margin-bottom: 3rem;

    .hero-content {
        margin-bottom: 2rem;
    }

    .hero-title {
        font-size: 3rem;
        font-weight: 300;
        color: #1a237e;
        margin: 0 0 1rem 0;
        letter-spacing: 1px;

        .body--dark & {
            color: rgba(255, 255, 255, 0.9);
        }
    }

    .hero-subtitle {
        font-size: 1.2rem;
        color: #666;
        margin: 0;
        font-weight: 300;
        line-height: 1.6;
        max-width: 600px;
        margin: 0 auto;

        .body--dark & {
            color: rgba(255, 255, 255, 0.7);
        }
    }

    .hero-stats {
        display: flex;
        justify-content: center;
        gap: 3rem;
        margin-top: 2rem;

        .stat-item {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 0.5rem;

            .stat-text {
                font-size: 0.9rem;
                color: #666;
                font-weight: 500;

                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
        }
    }
}

// Help Section
.help-section {
    text-align: center;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: rgba(26, 35, 126, 0.02);
    border-radius: 8px;
    border: 1px solid rgba(26, 35, 126, 0.05);

    .body--dark & {
        background: rgba(33, 150, 243, 0.05);
        border-color: rgba(33, 150, 243, 0.1);
    }

    .help-btn {
        font-weight: 500;
        transition: all 0.3s ease;
        border-radius: 6px;
        padding: 0.5rem 1.5rem;
        min-width: 140px;
        
        &:hover {
            background: rgba(33, 150, 243, 0.1);
            transform: translateY(-1px);
        }
    }
}

// Help Section
.help-section {
    margin-bottom: 2rem;

    .help-card {
        border-radius: 12px;
        border: 1px solid rgba(26, 35, 126, 0.1);

        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .help-text {
            font-size: 1rem;
            color: #666;
            margin-bottom: 0.5rem;

            .body--dark & {
                color: rgba(255, 255, 255, 0.7);
            }
        }
    }
}

// Notification Preferences - Subtle
.notification-preferences {
    margin-bottom: 2rem;

    .subtle-preference {
        display: flex;
        align-items: flex-start;
        gap: 0.75rem;
        padding: 1rem;
        background: rgba(26, 35, 126, 0.02);
        border-radius: 8px;
        border: 1px solid rgba(26, 35, 126, 0.05);

        .body--dark & {
            background: rgba(255, 255, 255, 0.02);
            border-color: rgba(255, 255, 255, 0.05);
        }

        .preference-text {
            display: flex;
            flex-direction: column;
            gap: 0.25rem;

            .preference-label {
                font-size: 0.9rem;
                color: #333;
                font-weight: 500;

                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }

            .preference-hint {
                font-size: 0.8rem;
                color: #666;

                .body--dark & {
                    color: rgba(255, 255, 255, 0.6);
                }
            }
        }
    }
}

// Upload Section
.upload-section {
    margin-bottom: 2rem;

    .upload-card {
        position: relative;
        border-radius: 16px;
        border: 1px solid rgba(26, 35, 126, 0.1);
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);

        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
        }

        .upload-header {
            text-align: center;
            margin-bottom: 2rem;

            .upload-icon {
                margin-bottom: 1rem;
            }

            .upload-title {
                font-size: 1.5rem;
                font-weight: 500;
                color: #333;
                margin: 0 0 0.5rem 0;

                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }

            .upload-description {
                color: #666;
                font-size: 1rem;
                margin: 0;

                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
        }

        .upload-dropzone {
            border: 2px dashed rgba(26, 35, 126, 0.3);
            border-radius: 12px;
            padding: 3rem 2rem;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-bottom: 2rem;
            position: relative;

            &:hover {
                border-color: rgba(26, 35, 126, 0.5);
                background: rgba(26, 35, 126, 0.02);
            }

            &.file-selected {
                border-color: #4caf50;
                background: rgba(76, 175, 80, 0.05);
            }

            .body--dark & {
                border-color: rgba(255, 255, 255, 0.3);

                &:hover {
                    border-color: rgba(255, 255, 255, 0.5);
                    background: rgba(255, 255, 255, 0.02);
                }

                &.file-selected {
                    border-color: #4caf50;
                    background: rgba(76, 175, 80, 0.1);
                }
            }

            .dropzone-content {
                .dropzone-text {
                    .dropzone-primary {
                        font-size: 1.1rem;
                        color: #333;
                        font-weight: 500;
                        margin-bottom: 0.5rem;

                        .body--dark & {
                            color: rgba(255, 255, 255, 0.9);
                        }
                    }

                    .dropzone-secondary {
                        color: #666;
                        font-size: 0.9rem;

                        .body--dark & {
                            color: rgba(255, 255, 255, 0.6);
                        }
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
                    color: #666;
                    font-size: 0.9rem;
                    margin-bottom: 1rem;

                    .body--dark & {
                        color: rgba(255, 255, 255, 0.6);
                    }
                }

                .remove-file-btn {
                    position: absolute;
                    top: 1rem;
                    right: 1rem;
                }
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
                color: #666;

                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
        }

        .retention-disclaimer {
            display: flex;
            align-items: flex-start;
            gap: 0.75rem;
            background: rgba(26, 35, 126, 0.02);
            border: 1px solid rgba(26, 35, 126, 0.1);
            border-radius: 8px;
            padding: 1rem;
            margin-bottom: 2rem;

            .body--dark & {
                background: rgba(33, 150, 243, 0.05);
                border-color: rgba(33, 150, 243, 0.2);
            }

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
                    color: #1a237e;

                    .body--dark & {
                        color: rgba(33, 150, 243, 0.9);
                    }
                }

                .disclaimer-content {
                    font-size: 0.85rem;
                    color: #666;
                    line-height: 1.4;

                    .body--dark & {
                        color: rgba(255, 255, 255, 0.7);
                    }
                }
            }
        }

        .upload-actions {
            text-align: center;

            .upload-btn-modern {
                padding: 1rem 2rem;
                font-weight: 500;
                letter-spacing: 0.5px;
                border-radius: 12px;
                min-width: 200px;
                transition: all 0.3s ease;

                &:hover {
                    transform: translateY(-2px);
                    box-shadow: 0 8px 25px rgba(26, 35, 126, 0.3);
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
    background: #fee;
    border: 1px solid #f5c6cb;
    border-radius: 8px;
    padding: 1rem;
    margin-top: 1rem;

    .body--dark & {
        background: rgba(220, 53, 69, 0.1);
        border-color: rgba(220, 53, 69, 0.3);
    }
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
    color: #721c24;
    margin-bottom: 0.5rem;

    .body--dark & {
        color: #f8d7da;
    }
}

.error-dismiss {
    color: #dc3545;
    font-size: 0.85rem;
    padding: 0.25rem 0.5rem;

    &:hover {
        background: rgba(220, 53, 69, 0.1);
    }

    .body--dark & {
        color: #f8d7da;
    }
}

// Responsive Design
@media (max-width: 768px) {
    .upload-container {
        padding: 1rem;
    }

    .hero-section {
        .hero-title {
            font-size: 2.2rem;
        }

        .hero-subtitle {
            font-size: 1rem;
        }

        .hero-stats {
            flex-direction: column;
            gap: 1.5rem;

            .stat-item {
                flex-direction: row;
                justify-content: center;
                gap: 1rem;
            }
        }
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
    .hero-section {
        .hero-title {
            font-size: 1.8rem;
        }

        .hero-stats {
            .stat-item {
                .stat-text {
                    font-size: 0.8rem;
                }
            }
        }
    }

    .upload-section {
        .upload-card {
            .upload-header {
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
