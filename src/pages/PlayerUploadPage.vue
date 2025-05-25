<template>
    <q-page padding>
        <div class="q-pa-md">
            <h1
                class="text-h4 text-center q-mb-lg page-title"
                :class="$q.dark.isActive ? 'text-grey-2' : 'text-grey-9'"
            >
                Football Manager HTML Player Parser
            </h1>

            <q-card
                class="q-mb-md instructions-card"
                :class="
                    $q.dark.isActive
                        ? 'bg-grey-9 text-grey-3'
                        : 'bg-blue-grey-1 text-blue-grey-10'
                "
            >
                <q-card-section>
                    <div class="text-subtitle1 text-weight-bold">
                        Instructions:
                    </div>
                    <ol class="q-ml-md">
                        <li>Ensure the Go API (port 8091) is running.</li>
                        <li>
                            API loads <code>attribute_weights.json</code> and
                            <code>role_specific_overall_weights.json</code> from
                            its <code>public</code> folder.
                        </li>
                        <li>
                            Select an HTML file exported from Football Manager.
                            <br />
                            <span
                                class="text-caption"
                                :class="
                                    $q.dark.isActive
                                        ? 'text-grey-5'
                                        : 'text-grey-7'
                                "
                            >
                                Maximum file size: 15MB (approx. 10,000
                                players).
                            </span>
                        </li>
                        <li>
                            Click "Upload and Parse". Currency symbol will be
                            auto-detected.
                        </li>
                        <li>
                            You'll be redirected to your dataset page where you can view players, analyze teams, and share your data.
                        </li>
                    </ol>
                </q-card-section>
            </q-card>

            <q-card
                class="q-mb-md upload-card"
                :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'"
            >
                <q-card-section>
                    <div class="text-subtitle1 q-mb-sm">Upload HTML File:</div>
                    <q-file
                        v-model="playerFile"
                        label="Select HTML file"
                        accept=".html"
                        outlined
                        counter
                        :label-color="$q.dark.isActive ? 'grey-4' : ''"
                        :input-class="$q.dark.isActive ? 'text-grey-3' : ''"
                    >
                        <template v-slot:prepend
                            ><q-icon name="attach_file"
                        /></template>
                        <template v-slot:hint> Max file size: 15MB </template>
                    </q-file>
                    <q-btn
                        class="q-mt-md full-width"
                        color="primary"
                        label="Upload and Parse"
                        :loading="loading"
                        :disable="
                            !playerFile ||
                            !attributeWeightsLoadedForFeedback ||
                            !roleSpecificOverallWeightsLoadedForFeedback ||
                            loading
                        "
                        @click="uploadAndParse"
                    >
                        <q-tooltip
                            v-if="
                                !attributeWeightsLoadedForFeedback ||
                                !roleSpecificOverallWeightsLoadedForFeedback
                            "
                        >
                            Client-side check for weight files pending...
                        </q-tooltip>
                    </q-btn>
                </q-card-section>
            </q-card>

            <q-banner
                v-if="error"
                class="bg-negative text-white q-mb-md"
                rounded
            >
                {{ error }}
                <template v-slot:action
                    ><q-btn
                        flat
                        color="white"
                        label="Dismiss"
                        @click="playerStore.error = ''"
                /></template>
            </q-banner>
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
                        message: "File uploaded and parsed successfully! Redirecting to dataset view...",
                        position: "top",
                        timeout: 2000,
                    });
                    // Redirect to the dataset page
                    setTimeout(() => {
                        if (playerStore.currentDatasetId) {
                            router.push(`/dataset/${playerStore.currentDatasetId}`);
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
.page-title {
    font-weight: 300;
}

.instructions-card,
.upload-card {
    border-radius: $generic-border-radius;
}

.no-data-card {
    border-radius: $generic-border-radius;
}
</style>