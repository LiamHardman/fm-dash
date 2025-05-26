<template>
    <q-page class="upload-page">
        <div class="upload-container">
            <div class="page-header"></div>

            <div class="help-section">
                <q-btn flat class="docs-btn" @click="$router.push('/docs')">
                    <q-icon name="help_outline" class="q-mr-sm" />
                    First time? Check out the getting started guide
                </q-btn>
            </div>

            <div class="upload-section">
                <h2 class="section-title">Select File</h2>
                <div class="upload-area">
                    <q-file
                        v-model="playerFile"
                        label="Choose HTML file"
                        accept=".html"
                        outlined
                        class="file-input"
                    >
                        <template v-slot:prepend>
                            <q-icon name="upload_file" />
                        </template>
                    </q-file>
                    <div class="file-hint">Maximum file size: 15MB</div>

                    <q-btn
                        class="upload-btn"
                        :loading="loading"
                        :disable="
                            !playerFile ||
                            !attributeWeightsLoadedForFeedback ||
                            !roleSpecificOverallWeightsLoadedForFeedback ||
                            loading
                        "
                        @click="uploadAndParse"
                        flat
                    >
                        {{ loading ? "Processing..." : "Upload and Parse" }}
                    </q-btn>
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
                    Please ensure your HTML export contains 10,000 players or
                    less. (Max file size: 15MB)
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
    </q-page>
</template>

<script>
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useQuasar, Notify } from "quasar";
import { usePlayerStore } from "../stores/playerStore";

const MAX_FILE_SIZE_BYTES = 15 * 1024 * 1024;

export default {
    name: "PlayerUploadPage",
    setup() {
        const router = useRouter();
        const playerStore = usePlayerStore();
        const $q = useQuasar();
        const playerFile = ref(null);
        const showFileSizeLimitModal = ref(false);

        const loading = computed(() => playerStore.loading);
        const error = computed({
            get: () => playerStore.error,
            set: (value) => {
                playerStore.error = value;
            },
        });

        const attributeWeightsLoadedForFeedback = ref(false);
        const roleSpecificOverallWeightsLoadedForFeedback = ref(false);

        const loadJsonForFeedback = async (filePath, loadedFlagRef) => {
            try {
                const response = await fetch(filePath);
                if (!response.ok)
                    throw new Error(`HTTP error! status: ${response.status}`);
                await response.json();
                loadedFlagRef.value = true;
            } catch (e) {
                console.warn(
                    `Client-side check: Failed to load ${filePath}:`,
                    e,
                );
                loadedFlagRef.value = true;
            }
        };

        onMounted(async () => {
            await loadJsonForFeedback(
                "/public/attribute_weights.json",
                attributeWeightsLoadedForFeedback,
            );
            await loadJsonForFeedback(
                "/public/role_specific_overall_weights.json",
                roleSpecificOverallWeightsLoadedForFeedback,
            );
        });

        const uploadAndParse = async () => {
            if (!playerFile.value) {
                playerStore.error = "Please select an HTML file first.";
                return;
            }
            if (playerFile.value.size > MAX_FILE_SIZE_BYTES) {
                showFileSizeLimitModal.value = true;
                return;
            }
            try {
                const formData = new FormData();
                formData.append("playerFile", playerFile.value);
                await playerStore.uploadPlayerFile(formData);
                if (!playerStore.error) {
                    Notify.create({
                        type: "positive",
                        message:
                            "File uploaded and parsed successfully! Redirecting to dataset view...",
                        position: "top",
                        timeout: 2000,
                    });
                    // Redirect to the dataset page
                    setTimeout(() => {
                        if (playerStore.currentDatasetId) {
                            router.push(
                                `/dataset/${playerStore.currentDatasetId}`,
                            );
                        }
                    }, 1000);
                }
            } catch (e) {
                console.error("Upload and Parse error in page:", e);
                if (playerStore.error) {
                    Notify.create({
                        type: "negative",
                        message: playerStore.error,
                        position: "top",
                        timeout: 5000,
                        actions: [{ label: "Dismiss", color: "white" }],
                    });
                } else {
                    Notify.create({
                        type: "negative",
                        message: `Upload failed: ${e.message}`,
                        position: "top",
                        timeout: 5000,
                        actions: [{ label: "Dismiss", color: "white" }],
                    });
                }
            }
        };

        return {
            playerFile,
            showFileSizeLimitModal,
            loading,
            error,
            attributeWeightsLoadedForFeedback,
            roleSpecificOverallWeightsLoadedForFeedback,
            uploadAndParse,
            playerStore,
        };
    },
};
</script>

<style lang="scss" scoped>
.upload-page {
    min-height: calc(100vh - 120px);
    background: linear-gradient(135deg, #f8f9fc 0%, #ffffff 100%);

    .body--dark & {
        background: linear-gradient(135deg, #1e1e1e 0%, #2a2a2a 100%);
    }
}

.upload-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 3rem 2rem;
}

.page-header {
    text-align: center;
    margin-bottom: 3rem;
}

.page-title {
    font-size: 2.5rem;
    font-weight: 300;
    color: #1a237e;
    margin: 0 0 1rem 0;
    letter-spacing: 1px;

    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.page-subtitle {
    font-size: 1.1rem;
    color: #666;
    margin: 0;
    font-weight: 300;
    line-height: 1.6;

    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.section-title {
    font-size: 1.3rem;
    font-weight: 400;
    color: #1a237e;
    margin: 0 0 1.5rem 0;
    letter-spacing: 0.5px;

    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.help-section {
    text-align: center;
    margin-bottom: 2rem;
}

.docs-btn {
    color: #666;
    font-size: 0.9rem;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    border: 1px solid rgba(26, 35, 126, 0.2);
    background: rgba(26, 35, 126, 0.02);

    &:hover {
        color: #1a237e;
        background: rgba(26, 35, 126, 0.05);
        border-color: rgba(26, 35, 126, 0.3);
    }

    .q-icon {
        font-size: 1.1rem;
    }

    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
        border-color: rgba(255, 255, 255, 0.2);
        background: rgba(255, 255, 255, 0.02);

        &:hover {
            color: rgba(255, 255, 255, 0.9);
            background: rgba(255, 255, 255, 0.05);
            border-color: rgba(255, 255, 255, 0.3);
        }
    }
}

.upload-section {
    margin-bottom: 2rem;
}

.upload-area {
    background: rgba(255, 255, 255, 0.9);
    border-radius: 12px;
    padding: 2rem;
    border: 2px dashed rgba(26, 35, 126, 0.2);
    text-align: center;

    .body--dark & {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
    }
}

.file-input {
    margin-bottom: 1rem;

    .q-field__control {
        border-color: rgba(26, 35, 126, 0.3);
        border-radius: 8px;

        .body--dark & {
            border-color: rgba(255, 255, 255, 0.3);
        }
    }

    .q-field__label {
        color: #666;

        .body--dark & {
            color: rgba(255, 255, 255, 0.7);
        }
    }
}

.file-hint {
    color: #666;
    font-size: 0.85rem;
    margin-bottom: 1.5rem;

    .body--dark & {
        color: rgba(255, 255, 255, 0.6);
    }
}

.upload-btn {
    background: #1a237e;
    color: white;
    padding: 12px 32px;
    font-weight: 500;
    letter-spacing: 0.5px;
    border-radius: 8px;
    min-width: 160px;

    &:hover {
        background: #283593;
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(26, 35, 126, 0.3);
    }

    &:disabled {
        background: #ccc;
        color: #999;
        transform: none;
        box-shadow: none;
    }

    transition: all 0.2s ease;

    .body--dark & {
        background: rgba(255, 255, 255, 0.9);
        color: #1a237e;

        &:hover {
            background: white;
        }

        &:disabled {
            background: rgba(255, 255, 255, 0.3);
            color: rgba(255, 255, 255, 0.5);
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

@media (max-width: 768px) {
    .upload-container {
        padding: 2rem 1rem;
    }

    .page-title {
        font-size: 2rem;
    }

    .page-subtitle {
        font-size: 1rem;
    }

    .instruction-item {
        padding: 1rem;
        gap: 0.75rem;
    }

    .step-number {
        width: 28px;
        height: 28px;
        font-size: 0.8rem;
    }

    .upload-area {
        padding: 1.5rem;
    }
}
</style>
