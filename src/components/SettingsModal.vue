<template>
    <q-dialog
        v-model="showDialog"
        persistent
        maximized
        :class="{
            'settings-modal': true,
            'settings-modal--dark': $q.dark.isActive
        }"
    >
        <q-card class="settings-card">
            <q-card-section class="settings-header">
                <div class="settings-title">
                    <q-icon name="settings" size="2rem" class="q-mr-md" />
                    <span class="text-h5">Settings</span>
                </div>
                <q-btn
                    flat
                    round
                    icon="close"
                    @click="closeModal"
                    class="close-btn"
                />
            </q-card-section>

            <q-separator />

            <q-card-section class="settings-content">
                <div class="settings-section">
                    <div class="section-header">
                        <q-icon name="assessment" size="1.5rem" class="q-mr-sm" />
                        <span class="section-title">Rating Calculation Method</span>
                    </div>
                    
                    <div class="section-description">
                        Choose how player ratings are calculated throughout the application.
                    </div>

                    <div class="rating-method-options">
                        <q-card 
                            :class="{
                                'method-card': true,
                                'method-card--selected': useScaledRatings,
                                'method-card--dark': $q.dark.isActive,
                                'method-card--disabled': isLoading
                            }"
                            @click="!isLoading && setRatingMethod(true)"
                        >
                            <q-card-section class="method-content">
                                <div class="method-header">
                                    <q-radio 
                                        v-model="useScaledRatings" 
                                        :val="true" 
                                        color="primary"
                                        :disable="isLoading"
                                        @click.stop="!isLoading && setRatingMethod(true)"
                                    />
                                    <span class="method-name">Scaled Ratings (Recommended)</span>
                                    <q-badge color="positive" label="NEW" class="q-ml-sm" />
                                    <q-spinner 
                                        v-if="isLoading && useScaledRatings" 
                                        color="primary" 
                                        size="1.2rem" 
                                        class="q-ml-md" 
                                    />
                                </div>
                                <div class="method-description">
                                    <p>Uses an enhanced rating system that:</p>
                                    <ul>
                                        <li>Keeps elite players (75+) at their current ratings</li>
                                        <li>Progressively lowers average players (50-75)</li>
                                        <li>Significantly reduces poor players (below 50)</li>
                                        <li>Creates better differentiation between skill levels</li>
                                    </ul>
                                </div>
                            </q-card-section>
                        </q-card>

                        <q-card 
                            :class="{
                                'method-card': true,
                                'method-card--selected': !useScaledRatings,
                                'method-card--dark': $q.dark.isActive,
                                'method-card--disabled': isLoading
                            }"
                            @click="!isLoading && setRatingMethod(false)"
                        >
                            <q-card-section class="method-content">
                                <div class="method-header">
                                    <q-radio 
                                        v-model="useScaledRatings" 
                                        :val="false" 
                                        color="primary"
                                        :disable="isLoading"
                                        @click.stop="!isLoading && setRatingMethod(false)"
                                    />
                                    <span class="method-name">Linear Ratings</span>
                                    <q-badge color="grey-6" label="LEGACY" class="q-ml-sm" />
                                    <q-spinner 
                                        v-if="isLoading && !useScaledRatings" 
                                        color="primary" 
                                        size="1.2rem" 
                                        class="q-ml-md" 
                                    />
                                </div>
                                <div class="method-description">
                                    <p>Uses the original linear scaling system:</p>
                                    <ul>
                                        <li>Direct linear conversion from attributes to ratings</li>
                                        <li>Equal distribution across all rating ranges</li>
                                        <li>Traditional FIFA-style calculation method</li>
                                        <li>Consistent with previous versions</li>
                                    </ul>
                                </div>
                            </q-card-section>
                        </q-card>
                    </div>

                    <div class="rating-preview">
                        <div class="preview-header">
                            <q-icon name="preview" class="q-mr-sm" />
                            <span>Rating Comparison Preview</span>
                        </div>
                        <div class="preview-content">
                            <div class="preview-example">
                                <div class="example-header">High-level Player (18/20 avg attributes)</div>
                                <div class="example-ratings">
                                    <div class="rating-comparison">
                                        <span class="rating-label">Linear:</span>
                                        <span class="rating-value rating-high">95</span>
                                    </div>
                                    <div class="rating-comparison">
                                        <span class="rating-label">Scaled:</span>
                                        <span class="rating-value rating-high">94</span>
                                    </div>
                                </div>
                            </div>
                            <div class="preview-example">
                                <div class="example-header">Average Player (12/20 avg attributes)</div>
                                <div class="example-ratings">
                                    <div class="rating-comparison">
                                        <span class="rating-label">Linear:</span>
                                        <span class="rating-value rating-medium">64</span>
                                    </div>
                                    <div class="rating-comparison">
                                        <span class="rating-label">Scaled:</span>
                                        <span class="rating-value rating-medium">56</span>
                                    </div>
                                </div>
                            </div>
                            <div class="preview-example">
                                <div class="example-header">Poor Player (8/20 avg attributes)</div>
                                <div class="example-ratings">
                                    <div class="rating-comparison">
                                        <span class="rating-label">Linear:</span>
                                        <span class="rating-value rating-low">42</span>
                                    </div>
                                    <div class="rating-comparison">
                                        <span class="rating-label">Scaled:</span>
                                        <span class="rating-value rating-low">27</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="settings-section">
                    <div class="section-header">
                        <q-icon name="info_outline" size="1.5rem" class="q-mr-sm" />
                        <span class="section-title">About Rating Changes</span>
                    </div>
                    <div class="info-content">
                        <q-card flat bordered class="info-card">
                            <q-card-section>
                                <div class="info-text">
                                    <p><strong>Note:</strong> Changing the rating calculation method will affect how all player ratings are displayed throughout the application. This includes:</p>
                                    <ul>
                                        <li>Individual player FIFA stats (PAC, SHO, PAS, DRI, DEF, PHY)</li>
                                        <li>Role-specific overall ratings</li>
                                        <li>Player comparisons and rankings</li>
                                        <li>Team ratings and analysis</li>
                                    </ul>
                                    <p>The setting is saved automatically and will persist across browser sessions.</p>
                                </div>
                            </q-card-section>
                        </q-card>
                    </div>
                </div>
            </q-card-section>

            <q-separator />

            <q-card-actions align="right" class="settings-actions">
                <q-btn
                    flat
                    label="Close"
                    @click="closeModal"
                    class="close-action-btn"
                />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { defineComponent, ref, computed, onMounted } from 'vue'
import { useUiStore } from '../stores/uiStore'
import { usePlayerStore } from '../stores/playerStore'
import { useQuasar } from 'quasar'
import playerService from '../services/playerService'

export default defineComponent({
    name: 'SettingsModal',
    props: {
        modelValue: {
            type: Boolean,
            default: false
        }
    },
    emits: ['update:modelValue'],
    setup(props, { emit }) {
        const uiStore = useUiStore()
        const playerStore = usePlayerStore()
        const $q = useQuasar()

        const showDialog = computed({
            get: () => props.modelValue,
            set: (value) => emit('update:modelValue', value)
        })

        const useScaledRatings = computed({
            get: () => uiStore.useScaledRatings,
            set: (value) => uiStore.setRatingCalculation(value)
        })

        const isLoading = ref(false)

        // Load backend configuration on component mount
        onMounted(async () => {
            try {
                const config = await playerService.getConfig()
                if (config.useScaledRatings !== undefined) {
                    uiStore.setRatingCalculation(config.useScaledRatings)
                }
            } catch (error) {
                console.warn('Failed to load backend configuration:', error)
            }
        })

        const setRatingMethod = async (useScaled) => {
            if (isLoading.value) return
            
            isLoading.value = true
            
            try {
                // Update backend first
                await playerService.updateConfig({
                    useScaledRatings: useScaled
                })
                
                // Update local store
                uiStore.setRatingCalculation(useScaled)
                
                // Trigger data refresh if we have a current dataset
                if (playerStore.currentDatasetId) {
                    console.log('🔄 Refreshing player data with new rating calculation method:', useScaled ? 'scaled' : 'linear')
                    await playerStore.fetchPlayersByDatasetId(playerStore.currentDatasetId)
                }
                
                // Show success notification
                $q.notify({
                    message: useScaled ? 'Switched to Scaled Ratings' : 'Switched to Linear Ratings',
                    caption: 'Ratings have been recalculated using the new method',
                    color: 'positive',
                    position: 'top',
                    timeout: 3000,
                    icon: 'assessment',
                    actions: [
                        {
                            icon: 'close',
                            color: 'white',
                            round: true,
                            handler: () => {}
                        }
                    ]
                })
                
            } catch (error) {
                console.error('Failed to update rating calculation method:', error)
                
                // Show error notification
                $q.notify({
                    message: 'Failed to update rating calculation method',
                    caption: 'Please try again or check your connection',
                    color: 'negative',
                    position: 'top',
                    timeout: 5000,
                    icon: 'error',
                    actions: [
                        {
                            icon: 'close',
                            color: 'white',
                            round: true,
                            handler: () => {}
                        }
                    ]
                })
            } finally {
                isLoading.value = false
            }
        }

        const closeModal = () => {
            showDialog.value = false
        }

        return {
            showDialog,
            useScaledRatings,
            setRatingMethod,
            closeModal,
            isLoading
        }
    }
})
</script>

<style lang="scss" scoped>
.settings-modal {
    .q-dialog__inner {
        padding: 0;
    }
}

.settings-card {
    width: 100%;
    max-width: 800px;
    margin: 2rem auto;
    max-height: 90vh;
    overflow-y: auto;
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }
}

.settings-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.5rem 2rem;
    background: rgba(26, 35, 126, 0.05);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.05);
    }
}

.settings-title {
    display: flex;
    align-items: center;
    color: #1a237e;
    font-weight: 600;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.close-btn {
    color: #666;
    
    &:hover {
        background: rgba(26, 35, 126, 0.1);
        color: #1a237e;
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
        
        &:hover {
            background: rgba(255, 255, 255, 0.1);
            color: rgba(255, 255, 255, 0.9);
        }
    }
}

.settings-content {
    padding: 2rem;
}

.settings-section {
    margin-bottom: 2rem;
    
    &:last-child {
        margin-bottom: 0;
    }
}

.section-header {
    display: flex;
    align-items: center;
    margin-bottom: 1rem;
    color: #1a237e;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.section-title {
    font-size: 1.2rem;
    font-weight: 600;
}

.section-description {
    margin-bottom: 1.5rem;
    color: #666;
    font-size: 0.95rem;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.rating-method-options {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 2rem;
}

.method-card {
    border: 2px solid rgba(26, 35, 126, 0.1);
    cursor: pointer;
    transition: all 0.3s ease;
    
    &:hover {
        border-color: rgba(26, 35, 126, 0.3);
        box-shadow: 0 4px 12px rgba(26, 35, 126, 0.1);
    }
    
    &--selected {
        border-color: #1a237e;
        background: rgba(26, 35, 126, 0.05);
    }
    
    &--disabled {
        opacity: 0.6;
        cursor: not-allowed;
        
        &:hover {
            border-color: rgba(26, 35, 126, 0.1);
            box-shadow: none;
        }
    }
    
    &--dark {
        border-color: rgba(255, 255, 255, 0.2);
        background: rgba(255, 255, 255, 0.02);
        
        &:hover {
            border-color: rgba(255, 255, 255, 0.4);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
        }
        
        &.method-card--selected {
            border-color: rgba(255, 255, 255, 0.9);
            background: rgba(255, 255, 255, 0.05);
        }
        
        &.method-card--disabled {
            opacity: 0.6;
            cursor: not-allowed;
            
            &:hover {
                border-color: rgba(255, 255, 255, 0.2);
                box-shadow: none;
            }
        }
    }
}

.method-content {
    padding: 1.5rem;
}

.method-header {
    display: flex;
    align-items: center;
    margin-bottom: 1rem;
}

.method-name {
    font-weight: 600;
    font-size: 1.1rem;
    margin-left: 0.5rem;
}

.method-description {
    margin-left: 2rem;
    color: #666;
    
    ul {
        margin: 0.5rem 0;
        padding-left: 1.5rem;
        
        li {
            margin-bottom: 0.25rem;
        }
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.rating-preview {
    background: rgba(26, 35, 126, 0.03);
    border-radius: 8px;
    padding: 1.5rem;
    border: 1px solid rgba(26, 35, 126, 0.1);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.02);
        border-color: rgba(255, 255, 255, 0.1);
    }
}

.preview-header {
    display: flex;
    align-items: center;
    margin-bottom: 1rem;
    font-weight: 600;
    color: #1a237e;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.preview-content {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
}

.preview-example {
    flex: 1;
    min-width: 200px;
    background: white;
    border-radius: 6px;
    padding: 1rem;
    border: 1px solid rgba(26, 35, 126, 0.1);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.1);
    }
}

.example-header {
    font-weight: 600;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: #1a237e;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.9);
    }
}

.example-ratings {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.rating-comparison {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.rating-label {
    font-size: 0.85rem;
    color: #666;
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.rating-value {
    font-weight: 700;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.9rem;
    
    &.rating-high {
        background: #4caf50;
        color: white;
    }
    
    &.rating-medium {
        background: #ff9800;
        color: white;
    }
    
    &.rating-low {
        background: #f44336;
        color: white;
    }
}

.info-content {
    margin-top: 1rem;
}

.info-card {
    border: 1px solid rgba(26, 35, 126, 0.1);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.02);
        border-color: rgba(255, 255, 255, 0.1);
    }
}

.info-text {
    color: #666;
    font-size: 0.9rem;
    
    ul {
        margin: 0.5rem 0;
        padding-left: 1.5rem;
        
        li {
            margin-bottom: 0.25rem;
        }
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.settings-actions {
    padding: 1rem 2rem;
    background: rgba(26, 35, 126, 0.02);
    
    .body--dark & {
        background: rgba(255, 255, 255, 0.02);
    }
}

.close-action-btn {
    color: #1a237e;
    
    &:hover {
        background: rgba(26, 35, 126, 0.1);
    }
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
        
        &:hover {
            background: rgba(255, 255, 255, 0.1);
            color: rgba(255, 255, 255, 0.9);
        }
    }
}

@media (max-width: 600px) {
    .settings-card {
        margin: 1rem;
        max-height: 95vh;
    }
    
    .settings-header,
    .settings-content,
    .settings-actions {
        padding: 1rem;
    }
    
    .preview-content {
        flex-direction: column;
    }
    
    .preview-example {
        min-width: auto;
    }
}
</style> 