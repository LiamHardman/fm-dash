<template>
    <q-dialog
        :model-value="show"
        @update:model-value="$emit('close')"
        persistent
        transition-show="scale"
        transition-hide="scale"
    >
        <q-card
            class="export-options-dialog"
            style="min-width: 500px; max-width: 600px; width: 90vw;"
        >
            <q-card-section
                class="row items-center q-pb-none card-header"
            >
                <q-icon name="download" size="md" class="q-mr-sm" />
                <div class="text-h6">
                    Export Options
                </div>
                <q-space />
                <q-btn
                    icon="close"
                    flat
                    round
                    dense
                    v-close-popup
                    @click="$emit('close')"
                />
            </q-card-section>

            <q-card-section class="q-pt-md">
                <div class="text-subtitle2 q-mb-md text-grey-7">
                    Select what information to include in your export
                </div>

                <!-- Export Format Selection -->
                <div class="export-format-section q-mb-lg">
                    <div class="text-h6 q-mb-md">Export Format</div>
                    <q-option-group
                        v-model="selectedFormat"
                        :options="formatOptions"
                        color="primary"
                        inline
                        class="format-options"
                    />
                </div>

                <!-- Data Selection -->
                <div class="data-selection-section q-mb-lg">
                    <div class="text-h6 q-mb-md">Data to Include</div>
                    
                    <!-- Quick Presets -->
                    <div class="preset-section q-mb-md">
                        <div class="text-subtitle2 q-mb-sm">Quick Presets:</div>
                        <div class="preset-buttons">
                            <q-btn
                                v-for="preset in presetOptions"
                                :key="preset.value"
                                :label="preset.label"
                                :color="selectedPreset === preset.value ? 'primary' : 'grey-5'"
                                :unelevated="selectedPreset === preset.value"
                                :outline="selectedPreset !== preset.value"
                                @click="selectPreset(preset.value)"
                                size="sm"
                                class="q-mr-sm q-mb-sm"
                            />
                        </div>
                    </div>

                    <!-- Custom Selection -->
                    <div class="custom-selection">
                        <div class="text-subtitle2 q-mb-sm">Or customize your selection:</div>
                        <div class="selection-grid">
                            <q-checkbox
                                v-model="selectedOptions.basicInfo"
                                label="Basic Info"
                                color="primary"
                                class="option-checkbox"
                            />
                            <q-checkbox
                                v-model="selectedOptions.fifahStats"
                                label="FIFA Stats"
                                color="primary"
                                class="option-checkbox"
                            />
                            <q-checkbox
                                v-model="selectedOptions.fmStats"
                                label="FM Attributes"
                                color="primary"
                                class="option-checkbox"
                            />
                            <q-checkbox
                                v-model="selectedOptions.roleRatings"
                                label="Role Ratings"
                                color="primary"
                                class="option-checkbox"
                            />
                            <q-checkbox
                                v-model="selectedOptions.performancePercentiles"
                                label="Performance Percentiles"
                                color="primary"
                                class="option-checkbox"
                            />
                            <q-checkbox
                                v-model="selectedOptions.contractInfo"
                                label="Contract Info"
                                color="primary"
                                class="option-checkbox"
                            />
                            <q-checkbox
                                v-model="selectedOptions.personalInfo"
                                label="Personal Info"
                                color="primary"
                                class="option-checkbox"
                            />
                        </div>
                    </div>
                </div>

                <!-- Export Summary -->
                <div class="export-summary q-mt-lg">
                    <q-card flat bordered class="summary-card">
                        <q-card-section>
                            <div class="text-subtitle2 q-mb-sm">Export Summary</div>
                            <div class="summary-details">
                                <div class="summary-item">
                                    <q-icon name="description" size="sm" class="q-mr-xs" />
                                    <span>Format: {{ selectedFormat.toUpperCase() }}</span>
                                </div>
                                <div class="summary-item">
                                    <q-icon name="group" size="sm" class="q-mr-xs" />
                                    <span>Players: {{ playerCount.toLocaleString() }}</span>
                                </div>
                                <div class="summary-item">
                                    <q-icon name="view_column" size="sm" class="q-mr-xs" />
                                    <span>Data Categories: {{ selectedCategoriesCount }}</span>
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>
                </div>
            </q-card-section>

            <q-card-actions align="right" class="q-pa-md">
                <q-btn
                    flat
                    label="Cancel"
                    color="grey"
                    @click="$emit('close')"
                />
                <q-btn
                    unelevated
                    :label="`Export ${selectedFormat.toUpperCase()}`"
                    color="primary"
                    @click="handleExport"
                    :disable="selectedCategoriesCount === 0"
                    :loading="exporting"
                />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, ref, watch } from 'vue'

export default defineComponent({
  name: 'ExportOptionsDialog',
  props: {
    show: {
      type: Boolean,
      default: false
    },
    playerCount: {
      type: Number,
      default: 0
    },
    exportType: {
      type: String, // 'csv' or 'json'
      default: 'csv'
    }
  },
  emits: ['close', 'export'],
  setup(props, { emit }) {
    const $q = useQuasar()
    const exporting = ref(false)

    const selectedFormat = ref(props.exportType)
    const selectedPreset = ref('detailed')

    const selectedOptions = ref({
      basicInfo: true,
      fifahStats: true,
      fmStats: true,
      roleRatings: false,
      performancePercentiles: false,
      contractInfo: true,
      personalInfo: false
    })

    const formatOptions = [
      { label: 'CSV', value: 'csv' },
      { label: 'JSON', value: 'json' }
    ]

    const presetOptions = [
      { label: 'Basic', value: 'basic' },
      { label: 'Detailed', value: 'detailed' },
      { label: 'Scout Report', value: 'scout' },
      { label: 'Analysis', value: 'analysis' },
      { label: 'All Data', value: 'all' }
    ]

    const selectedCategoriesCount = computed(() => {
      return Object.values(selectedOptions.value).filter(Boolean).length
    })

    const selectPreset = preset => {
      selectedPreset.value = preset

      switch (preset) {
        case 'basic':
          selectedOptions.value = {
            basicInfo: true,
            fifahStats: true,
            fmStats: false,
            roleRatings: false,
            performancePercentiles: false,
            contractInfo: true,
            personalInfo: false
          }
          break
        case 'detailed':
          selectedOptions.value = {
            basicInfo: true,
            fifahStats: true,
            fmStats: true,
            roleRatings: false,
            performancePercentiles: false,
            contractInfo: true,
            personalInfo: true
          }
          break
        case 'scout':
          selectedOptions.value = {
            basicInfo: true,
            fifahStats: true,
            fmStats: false,
            roleRatings: true,
            performancePercentiles: true,
            contractInfo: true,
            personalInfo: true
          }
          break
        case 'analysis':
          selectedOptions.value = {
            basicInfo: true,
            fifahStats: true,
            fmStats: true,
            roleRatings: true,
            performancePercentiles: true,
            contractInfo: false,
            personalInfo: false
          }
          break
        case 'all':
          selectedOptions.value = {
            basicInfo: true,
            fifahStats: true,
            fmStats: true,
            roleRatings: false,
            performancePercentiles: true,
            contractInfo: true,
            personalInfo: true
          }
          break
      }
    }

    // Watch for changes in selectedOptions to reset preset selection
    watch(
      () => selectedOptions.value,
      () => {
        // Check if current selection matches any preset
        const currentSelection = selectedOptions.value
        let matchingPreset = null

        for (const preset of presetOptions) {
          let tempOptions = {}

          // Simulate selecting each preset to compare
          switch (preset.value) {
            case 'basic':
              tempOptions = {
                basicInfo: true,
                fifahStats: true,
                fmStats: false,
                roleRatings: false,
                performancePercentiles: false,
                contractInfo: true,
                personalInfo: false
              }
              break
            case 'detailed':
              tempOptions = {
                basicInfo: true,
                fifahStats: true,
                fmStats: true,
                roleRatings: false,
                performancePercentiles: false,
                contractInfo: true,
                personalInfo: true
              }
              break
            case 'scout':
              tempOptions = {
                basicInfo: true,
                fifahStats: true,
                fmStats: false,
                roleRatings: true,
                performancePercentiles: true,
                contractInfo: true,
                personalInfo: true
              }
              break
            case 'analysis':
              tempOptions = {
                basicInfo: true,
                fifahStats: true,
                fmStats: true,
                roleRatings: true,
                performancePercentiles: true,
                contractInfo: false,
                personalInfo: false
              }
              break
            case 'all':
              tempOptions = {
                basicInfo: true,
                fifahStats: true,
                fmStats: true,
                roleRatings: false,
                performancePercentiles: true,
                contractInfo: true,
                personalInfo: true
              }
              break
          }

          // Check if all options match
          const matches = Object.keys(tempOptions).every(
            key => tempOptions[key] === currentSelection[key]
          )

          if (matches) {
            matchingPreset = preset.value
            break
          }
        }

        if (matchingPreset) {
          selectedPreset.value = matchingPreset
        } else {
          selectedPreset.value = null // Custom selection
        }
      },
      { deep: true }
    )

    const handleExport = async () => {
      if (selectedCategoriesCount.value === 0) {
        $q.notify({
          type: 'warning',
          message: 'Please select at least one data category to export',
          position: 'top'
        })
        return
      }

      exporting.value = true

      try {
        const exportOptions = {
          format: selectedFormat.value,
          options: selectedOptions.value,
          preset: selectedPreset.value
        }

        emit('export', exportOptions)
      } catch (_error) {
        $q.notify({
          type: 'negative',
          message: 'Export failed. Please try again.',
          position: 'top'
        })
      } finally {
        exporting.value = false
      }
    }

    // Initialize format based on prop
    watch(
      () => props.exportType,
      newType => {
        selectedFormat.value = newType
      },
      { immediate: true }
    )

    return {
      selectedFormat,
      selectedPreset,
      selectedOptions,
      formatOptions,
      presetOptions,
      selectedCategoriesCount,
      exporting,
      selectPreset,
      handleExport
    }
  }
})
</script>

<style lang="scss" scoped>
.export-options-dialog {
    .card-header {
        background: rgba(25, 118, 210, 0.05);
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
        }
    }

    .export-format-section {
        .format-options {
            .q-radio {
                margin-right: 1rem;
            }
        }
    }

    .preset-section {
        .preset-buttons {
            display: flex;
            flex-wrap: wrap;
            gap: 0.5rem;
        }
    }

    .selection-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 1rem;
        
        .option-checkbox {
            .q-checkbox__label {
                font-weight: 500;
            }
        }
    }

    .summary-card {
        background: rgba(25, 118, 210, 0.05);
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
        }

        .summary-details {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        .summary-item {
            display: flex;
            align-items: center;
            font-size: 0.9rem;
        }
    }
}

@media (max-width: 600px) {
    .export-options-dialog {
        .selection-grid {
            grid-template-columns: 1fr;
        }
        
        .preset-section .preset-buttons {
            flex-direction: column;
            align-items: stretch;
            
            .q-btn {
                margin-right: 0;
            }
        }
    }
}
</style> 