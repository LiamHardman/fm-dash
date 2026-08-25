<template>
    <q-dialog
        v-model="showDialog"
        persistent
        maximized
        :class="{
            'tutorial-modal': true,
            'tutorial-modal--dark': $q.dark.isActive
        }"
    >
        <q-card class="tutorial-card">
            <!-- Dialog chrome: header (icon/title/close), the same convention used by
                 PlayerDetailDialog/SettingsModal — an icon, a title, then a close
                 button, all in normal flow. -->
            <div class="dialog-chrome">
                <div class="dialog-chrome__header">
                    <q-icon name="school" class="dialog-chrome__icon" />
                    <div class="dialog-chrome__title">First Time Setup Guide</div>
                    <q-space />
                    <div class="dialog-chrome__actions">
                        <q-btn
                            icon="close"
                            flat
                            round
                            dense
                            class="dialog-chrome__close"
                            @click="closeModal"
                        />
                    </div>
                </div>
            </div>

            <q-card-section class="tutorial-content">
                <!-- Step Indicator -->
                <div class="step-indicator">
                    <div 
                        v-for="(step, index) in steps" 
                        :key="index"
                        :class="{
                            'step-dot': true,
                            'step-dot--active': currentStep === index,
                            'step-dot--completed': currentStep > index
                        }"
                    >
                        <q-icon 
                            v-if="currentStep > index" 
                            name="check" 
                            size="1rem" 
                            color="white" 
                        />
                        <span v-else>{{ index + 1 }}</span>
                    </div>
                </div>

                <!-- Step Content -->
                <div class="step-content">
                    <div class="step-header">
                        <h3 class="step-title">{{ steps[currentStep].title }}</h3>
                        <p class="step-subtitle">{{ steps[currentStep].subtitle }}</p>
                    </div>

                    <div class="step-body">
                        <!-- Step 1: Download Scouting View -->
                        <div v-if="currentStep === 0" class="step-1">
                            <div class="instruction-card">
                                <div class="instruction-header">
                                    <q-icon name="download" size="2rem" color="primary" />
                                    <h4>Download the FM Dash Search View</h4>
                                </div>
                                <div class="instruction-content">
                                    <p>First, you need to download a custom search view from the Steam Workshop that contains all the player attributes FM-Dash needs for analysis:</p>
                                    <ol>
                                        <li>Make sure you're logged into Steam</li>
                                        <li>Click the link below to go to the Steam Workshop</li>
                                        <li>Subscribe to the "FM Dash Search" workshop item</li>
                                        <li>Wait for Steam to download the workshop item</li>
                                    </ol>
                                    <div class="note-box">
                                        <q-icon name="info" size="1rem" color="info" />
                                        <span>Make sure you're logged into Steam and subscribed to the workshop item.</span>
                                    </div>
                                    <div class="workshop-link">
                                        <q-btn
                                            unelevated
                                            color="primary"
                                            icon="link"
                                            label="Download FM Dash Search View"
                                            @click="openWorkshopLink"
                                            class="workshop-btn"
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Step 2: Import the View -->
                        <div v-if="currentStep === 1" class="step-2">
                            <div class="instruction-card">
                                <div class="instruction-header">
                                    <q-icon name="import_export" size="2rem" color="secondary" />
                                    <h4>Import the View in FM24</h4>
                                </div>
                                <div class="instruction-content">
                                    <p>After downloading, you need to import the view into your game:</p>
                                    <ol>
                                        <li>Open Football Manager 24</li>
                                        <li>Navigate to <strong>Scouting</strong></li>
                                        <li>Click <strong>"Overview"</strong> (next to the "X Players Filtered" text)</li>
                                        <li>Select <strong>"Custom"</strong> → <strong>"Import View"</strong></li>
                                        <li>Choose <strong>"FM Dash Search"</strong></li>
                                    </ol>
                                    <div class="note-box">
                                        <q-icon name="info" size="1rem" color="info" />
                                        <span>If you don't see the view, restart FM24 and make sure Steam has downloaded the workshop item.</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Step 3: Export Data -->
                        <div v-if="currentStep === 2" class="step-3">
                            <div class="instruction-card">
                                <div class="instruction-header">
                                    <q-icon name="print" size="2rem" color="accent" />
                                    <h4>Export Your Data</h4>
                                </div>
                                <div class="instruction-content">
                                    <p>Now you need to export your scouting data:</p>
                                    <ol>
                                        <li>Use FM24's filtering options to narrow down your player selection</li>
                                        <li>Consider filtering by league, position, age, or other criteria</li>
                                        <li>Press <kbd>Ctrl+A</kbd> (or <kbd>Cmd+A</kbd> on Mac) to select all players</li>
                                        <li>Press <kbd>Ctrl+P</kbd> (or <kbd>Cmd+P</kbd> on Mac) to open print dialog</li>
                                        <li>Select <strong>"Web Page"</strong> as the format</li>
                                        <li>Save the file somewhere you'll remember</li>
                                    </ol>
                                    <div class="note-box">
                                        <q-icon name="lightbulb" size="1rem" color="orange" />
                                        <span>Start with under 5,000 players for your first export to test the process quickly.</span>
                                    </div>
                                    <div class="warning-box">
                                        <q-icon name="warning" size="1rem" color="warning" />
                                        <span>This process can be slow for large datasets (10,000+ players). Expect 10+ seconds and don't interact with the screen during export.</span>
                                    </div>
                                    <div class="keyboard-shortcuts">
                                        <div class="shortcut">
                                            <kbd>Ctrl+A</kbd>
                                            <span>Select All</span>
                                        </div>
                                        <div class="shortcut">
                                            <kbd>Ctrl+P</kbd>
                                            <span>Print</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Step 4: Upload to FM-Dash -->
                        <div v-if="currentStep === 3" class="step-4">
                            <div class="instruction-card">
                                <div class="instruction-header">
                                    <q-icon name="upload" size="2rem" color="positive" />
                                    <h4>Upload to FM-Dash</h4>
                                </div>
                                <div class="instruction-content">
                                    <p>Finally, upload your exported data to FM-Dash:</p>
                                    <ol>
                                        <li>Go to the <strong>Upload</strong> page</li>
                                        <li>Click <strong>"Choose File"</strong> or drag and drop</li>
                                        <li>Select your saved HTML file</li>
                                        <li>Wait for the upload to complete</li>
                                        <li>Start analyzing your players!</li>
                                    </ol>
                                    <div class="upload-preview">
                                        <q-icon name="upload_file" size="3rem" color="primary" />
                                        <p>Your data will be processed and ready for analysis</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </q-card-section>

            <q-separator />

            <q-card-actions align="right" class="tutorial-actions">
                <q-btn
                    v-if="currentStep > 0"
                    flat
                    label="Back"
                    icon="arrow_back"
                    @click="previousStep"
                    class="back-btn"
                />
                <q-space />
                <q-btn
                    v-if="currentStep < steps.length - 1"
                    unelevated
                    color="primary"
                    :label="currentStep === steps.length - 2 ? 'Finish' : 'Next'"
                    :icon="currentStep === steps.length - 2 ? 'check' : 'arrow_forward'"
                    @click="nextStep"
                    class="next-btn"
                />
                <q-btn
                    v-else
                    unelevated
                    color="positive"
                    label="Get Started"
                    icon="play_arrow"
                    @click="finishTutorial"
                    class="finish-btn"
                />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, ref } from 'vue'
import { useRouter } from 'vue-router'

export default defineComponent({
  name: 'FirstTimeTutorialModal',
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const showDialog = computed({
      get: () => props.modelValue,
      set: (value) => emit('update:modelValue', value),
    })

    const _$q = useQuasar()
    const router = useRouter()
    const currentStep = ref(0)

    const steps = [
      {
        title: 'Download FM Dash Search View',
        subtitle: 'Get the custom search view from Steam Workshop',
      },
      {
        title: 'Import the View in FM24',
        subtitle: 'Add the view to your Football Manager game',
      },
      {
        title: 'Export Your Data',
        subtitle: 'Export your scouting data as HTML',
      },
      {
        title: 'Upload to FM-Dash',
        subtitle: 'Upload your data and start analyzing',
      },
    ]

    const openWorkshopLink = () => {
      // Open Steam Workshop link in new tab
      window.open('https://steamcommunity.com/sharedfiles/filedetails/?id=3498467200', '_blank')
    }

    const nextStep = () => {
      if (currentStep.value < steps.length - 1) {
        currentStep.value++
      }
    }

    const previousStep = () => {
      if (currentStep.value > 0) {
        currentStep.value--
      }
    }

    const finishTutorial = () => {
      closeModal()
      // Navigate to upload page
      router.push('/upload')
    }

    const closeModal = () => {
      emit('update:modelValue', false)
    }

    return {
      showDialog,
      currentStep,
      steps,
      openWorkshopLink,
      nextStep,
      previousStep,
      finishTutorial,
      closeModal,
    }
  },
})
</script>

<style lang="scss" scoped>
.tutorial-modal {
    .q-dialog__inner {
        padding: 0;
    }
}

.tutorial-card {
    width: 100%;
    max-width: 800px;
    margin: 2rem auto;
    max-height: 90vh;
    overflow-y: auto;
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
}

// Dialog chrome: unified header convention shared with PlayerDetailDialog /
// SettingsModal — icon, title, actions, close, all in normal flow.
.dialog-chrome {
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    background: var(--surface-raised);
    border-bottom: 1px solid var(--surface-border);
}

.dialog-chrome__header {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 12px var(--density-card-padding, 16px);
}

.dialog-chrome__icon {
    font-size: 1.3rem;
    color: var(--accent);
    flex-shrink: 0;
}

.dialog-chrome__title {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.dialog-chrome__actions {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
}

.dialog-chrome__close {
    transition: transform 0.15s ease;

    &:hover {
        transform: scale(1.08);
    }
}

.tutorial-content {
    padding: 2rem;
}

.step-indicator {
    display: flex;
    justify-content: center;
    gap: 1rem;
    margin-bottom: 2rem;
}

.step-dot {
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 50%;
    background: var(--surface-raised);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    color: var(--text-secondary);
    transition: all 0.3s ease;
    border: 2px solid transparent;

    &--active {
        background: var(--accent);
        color: var(--text-on-brand);
        border-color: var(--accent);
    }

    &--completed {
        // Success-tier color, kept semantic/hardcoded per established precedent.
        background: #4caf50;
        color: white;
        border-color: #4caf50;
    }
}

.step-content {
    max-width: 600px;
    margin: 0 auto;
}

.step-header {
    text-align: center;
    margin-bottom: 2rem;
}

.step-title {
    font-size: 1.8rem;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0 0 0.5rem 0;
}

.step-subtitle {
    font-size: 1.1rem;
    color: var(--text-secondary);
    margin: 0;
}

.step-body {
    .instruction-card {
        background: var(--surface-raised);
        border-radius: var(--radius-md);
        padding: 2rem;
        border: 1px solid var(--surface-border);
    }
}

.instruction-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1.5rem;

    h4 {
        font-size: 1.3rem;
        font-weight: 600;
        color: var(--text-primary);
        margin: 0;
    }
}

.instruction-content {
    p {
        font-size: 1rem;
        line-height: 1.6;
        color: var(--text-primary);
        margin-bottom: 1rem;
    }

    ol {
        margin: 1rem 0;
        padding-left: 1.5rem;

        li {
            margin-bottom: 0.5rem;
            line-height: 1.6;
            color: var(--text-primary);

            strong {
                color: var(--accent);
                font-weight: 600;
            }
        }
    }
}

.workshop-link {
    margin-top: 1.5rem;
    text-align: center;
}

.workshop-btn {
    padding: 0.8rem 1.5rem;
    font-weight: 600;
}

.note-box {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(33, 150, 243, 0.1);
    padding: 1rem;
    border-radius: 8px;
    margin-top: 1rem;
    border-left: 4px solid #2196f3;
    
    .body--dark & {
        background: rgba(33, 150, 243, 0.1);
        border-left-color: #2196f3;
    }
    
    span {
        color: #1976d2;
        font-size: 0.9rem;
        
        .body--dark & {
            color: #90caf9;
        }
    }
}

.warning-box {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 152, 0, 0.1);
    padding: 1rem;
    border-radius: 8px;
    margin-top: 1rem;
    border-left: 4px solid #ff9800;
    
    .body--dark & {
        background: rgba(255, 152, 0, 0.1);
        border-left-color: #ff9800;
    }
    
    span {
        color: #e65100;
        font-size: 0.9rem;
        
        .body--dark & {
            color: #ffb74d;
        }
    }
}

.keyboard-shortcuts {
    display: flex;
    gap: 1rem;
    margin-top: 1rem;
    justify-content: center;
}

.shortcut {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;

    kbd {
        background: var(--surface-raised);
        border: 1px solid var(--surface-border-strong);
        border-radius: 4px;
        padding: 0.5rem 0.8rem;
        font-family: monospace;
        font-size: 0.9rem;
        font-weight: 600;
        color: var(--text-primary);
    }

    span {
        font-size: 0.8rem;
        color: var(--text-secondary);
    }
}

.upload-preview {
    text-align: center;
    margin-top: 1.5rem;
    padding: 1.5rem;
    background: rgba(76, 175, 80, 0.1);
    border-radius: 8px;
    border: 1px solid rgba(76, 175, 80, 0.2);
    
    .body--dark & {
        background: rgba(76, 175, 80, 0.1);
        border-color: rgba(76, 175, 80, 0.3);
    }
    
    p {
        margin: 1rem 0 0 0;
        color: #2e7d32;
        font-weight: 500;
        
        .body--dark & {
            color: #81c784;
        }
    }
}

.tutorial-actions {
    padding: 1rem 2rem;
    background: var(--surface-raised);
}

.back-btn {
    color: var(--text-secondary);

    &:hover {
        background: var(--accent-soft);
        color: var(--accent);
    }
}

.next-btn,
.finish-btn {
    padding: 0.8rem 1.5rem;
    font-weight: 600;
}

@media (max-width: 600px) {
    .tutorial-card {
        margin: 1rem;
        max-height: 95vh;
    }
    
    .tutorial-header,
    .tutorial-content,
    .tutorial-actions {
        padding: 1rem;
    }
    
    .step-title {
        font-size: 1.5rem;
    }
    
    .step-subtitle {
        font-size: 1rem;
    }
    
    .instruction-card {
        padding: 1.5rem;
    }
    
    .keyboard-shortcuts {
        flex-direction: column;
        gap: 0.5rem;
    }
}
</style> 