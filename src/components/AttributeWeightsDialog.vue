<template>
    <q-dialog
        v-model="showDialog"
        persistent
        maximized
        :class="{
            'weights-modal': true,
            'weights-modal--dark': $q.dark.isActive
        }"
    >
        <q-card class="weights-card">
            <div class="dialog-chrome">
                <div class="dialog-chrome__header">
                    <q-icon name="tune" class="dialog-chrome__icon" />
                    <div class="dialog-chrome__title">Attribute Weights</div>
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

            <q-card-section class="weights-content">
                <q-card flat bordered class="info-card q-mb-lg">
                    <q-card-section>
                        <div class="info-text">
                            <p>
                                <strong>How this works:</strong> FIFA-style category stats
                                (PAC, SHO, PAS, DRI, DEF, PHY, and the goalkeeper categories)
                                are each a weighted average of specific FM attributes on their
                                1–20 scale, then scaled up to a 0–99 rating. A higher weight
                                means that attribute pulls the category score more; weights
                                are relative to each other within a category, so they don't
                                need to add up to any particular total.
                            </p>
                            <p>
                                Pick a built-in preset for a different scouting philosophy, or
                                build your own weighting by editing the values below, then
                                <strong>Save As</strong> to name and keep it. Use the dropdown
                                to switch between presets and your saved weight sets —
                                whichever one is selected is applied immediately across the
                                whole app for the currently loaded dataset. Custom weight sets
                                are saved in this browser only.
                            </p>
                        </div>
                    </q-card-section>
                </q-card>

                <div class="profile-bar">
                    <q-select
                        filled
                        dense
                        :model-value="selectedOption"
                        :options="profileOptions"
                        emit-value
                        map-options
                        label="Active weight set"
                        class="profile-select"
                        :disable="applying"
                        @update:model-value="onSelectProfile"
                    >
                        <template v-slot:prepend>
                            <q-icon name="balance" />
                        </template>
                    </q-select>

                    <q-badge v-if="selectedPreset" color="primary" outline class="dirty-badge">
                        Built-in preset — edit and "Save As" to customize
                    </q-badge>

                    <q-badge v-if="isDirty" color="warning" outline class="dirty-badge">
                        Unsaved changes
                    </q-badge>

                    <q-space />

                    <q-btn
                        unelevated
                        color="primary"
                        icon="save"
                        label="Save"
                        :disable="!canSave || applying"
                        :loading="applying"
                        @click="saveCurrentProfile"
                    />
                    <q-btn
                        outline
                        color="primary"
                        icon="save_as"
                        label="Save As..."
                        :disable="applying"
                        @click="promptSaveAs"
                    />
                    <q-btn
                        flat
                        icon="restart_alt"
                        label="Reset to Default"
                        :disable="applying"
                        @click="resetToDefault"
                    />
                    <q-btn
                        v-if="selectedProfile"
                        flat
                        color="negative"
                        icon="delete"
                        label="Delete"
                        :disable="applying"
                        @click="confirmDelete"
                    />
                </div>

                <div v-if="selectedPreset" class="preset-description">
                    {{ selectedPreset.description }}
                </div>

                <q-tabs
                    v-model="activeGroup"
                    class="weights-tabs"
                    active-color="primary"
                    indicator-color="primary"
                    align="left"
                    dense
                >
                    <q-tab
                        v-for="group in categoryGroups"
                        :key="group.id"
                        :name="group.id"
                        :label="group.label"
                    />
                </q-tabs>

                <q-separator class="q-mb-md" />

                <q-tab-panels v-model="activeGroup" animated class="weights-panels">
                    <q-tab-panel
                        v-for="group in categoryGroups"
                        :key="group.id"
                        :name="group.id"
                        class="weights-panel"
                    >
                        <q-expansion-item
                            v-for="category in group.categories"
                            :key="category.id"
                            expand-separator
                            :label="category.label"
                            :caption="category.note"
                            header-class="category-header"
                            class="category-expansion"
                        >
                            <div class="attribute-grid">
                                <div
                                    v-for="attrKey in categoryAttributeKeys(category.id)"
                                    :key="attrKey"
                                    class="attribute-row"
                                >
                                    <div class="attribute-label">
                                        <span class="attribute-abbr">{{ attrKey }}</span>
                                        <span class="attribute-name">{{ attributeLabel(attrKey) }}</span>
                                    </div>
                                    <q-input
                                        dense
                                        filled
                                        type="number"
                                        min="0"
                                        max="999"
                                        step="1"
                                        class="attribute-input"
                                        :model-value="editorWeights[category.id]?.[attrKey] ?? 0"
                                        @update:model-value="(val) => setWeight(category.id, attrKey, val)"
                                    />
                                </div>
                            </div>
                        </q-expansion-item>
                    </q-tab-panel>
                </q-tab-panels>
            </q-card-section>

            <q-separator />

            <q-card-actions align="right" class="weights-actions">
                <q-btn flat label="Close" @click="closeModal" class="close-action-btn" />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref, watch } from 'vue'
import { usePlayerStore } from '@/stores/playerStore'
import {
  ATTRIBUTE_LABELS,
  PRESET_WEIGHT_PROFILES,
  useWeightsStore,
  WEIGHT_CATEGORY_GROUPS,
} from '@/stores/weightsStore'
import playerService from '../services/playerService'

const DEFAULT_OPTION_VALUE = 'default'

function cloneWeights(weights) {
  return JSON.parse(JSON.stringify(weights || {}))
}

export default defineComponent({
  name: 'AttributeWeightsDialog',
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const $q = useQuasar()
    const weightsStore = useWeightsStore()
    const playerStore = usePlayerStore()

    const showDialog = computed({
      get: () => props.modelValue,
      set: (value) => emit('update:modelValue', value),
    })

    const applying = ref(false)
    const activeGroup = ref(WEIGHT_CATEGORY_GROUPS[0].id)
    const selectedOption = ref(DEFAULT_OPTION_VALUE)
    const editorWeights = ref(cloneWeights(weightsStore.defaultWeights))

    const categoryGroups = WEIGHT_CATEGORY_GROUPS

    const profileOptions = computed(() => [
      { label: 'Default (App Defaults)', value: DEFAULT_OPTION_VALUE },
      ...PRESET_WEIGHT_PROFILES.map((p) => ({ label: p.name, value: p.id })),
      ...weightsStore.profiles.map((p) => ({ label: p.name, value: p.id })),
    ])

    // Presets ship built-in (like Default) and can't be overwritten or deleted --
    // editing one and hitting "Save As" spins off a new custom profile instead.
    const selectedPreset = computed(
      () => PRESET_WEIGHT_PROFILES.find((p) => p.id === selectedOption.value) || null
    )

    const selectedCustomProfile = computed(
      () => weightsStore.profiles.find((p) => p.id === selectedOption.value) || null
    )

    // Kept as an alias for template/legacy readability: "the saved profile behind
    // the current selection, if any" -- used for the Delete button.
    const selectedProfile = selectedCustomProfile

    const baselineWeights = computed(
      () =>
        selectedCustomProfile.value?.weights ||
        selectedPreset.value?.weights ||
        weightsStore.defaultWeights
    )

    const isDirty = computed(
      () => JSON.stringify(editorWeights.value) !== JSON.stringify(baselineWeights.value)
    )

    const canSave = computed(() => !!selectedCustomProfile.value && isDirty.value)

    function categoryAttributeKeys(categoryId) {
      const fromEditor = Object.keys(editorWeights.value[categoryId] || {})
      if (fromEditor.length > 0) return fromEditor
      return Object.keys(weightsStore.defaultWeights[categoryId] || {})
    }

    function attributeLabel(attrKey) {
      return ATTRIBUTE_LABELS[attrKey] || attrKey
    }

    function setWeight(categoryId, attrKey, rawValue) {
      const numeric = Math.max(0, Math.min(999, Math.round(Number(rawValue) || 0)))
      const next = cloneWeights(editorWeights.value)
      if (!next[categoryId]) next[categoryId] = {}
      next[categoryId][attrKey] = numeric
      editorWeights.value = next
    }

    async function activateWeights(weights) {
      applying.value = true
      try {
        await playerService.updateConfig({ attributeWeights: weights })

        if (playerStore.currentDatasetId) {
          await playerStore.fetchPlayersByDatasetId(playerStore.currentDatasetId)
        }

        $q.notify({
          message: 'Attribute weights applied',
          caption: 'Player ratings have been recalculated using the new weights',
          color: 'positive',
          position: 'top',
          timeout: 3000,
          icon: 'tune',
        })
      } catch (_error) {
        $q.notify({
          message: 'Failed to apply attribute weights',
          caption: 'Please try again or check your connection',
          color: 'negative',
          position: 'top',
          timeout: 5000,
          icon: 'error',
        })
      } finally {
        applying.value = false
      }
    }

    function loadEditorFromSelection() {
      editorWeights.value = cloneWeights(baselineWeights.value)
    }

    function onSelectProfile(value) {
      selectedOption.value = value
      loadEditorFromSelection()
      weightsStore.setActiveProfileId(value === DEFAULT_OPTION_VALUE ? null : value)
      activateWeights(baselineWeights.value)
    }

    function saveCurrentProfile() {
      if (!selectedCustomProfile.value) return
      weightsStore.updateProfileWeights(selectedCustomProfile.value.id, editorWeights.value)
      activateWeights(editorWeights.value)
    }

    function promptSaveAs() {
      const currentName = selectedCustomProfile.value?.name || selectedPreset.value?.name
      $q.dialog({
        title: 'Save Weight Set As',
        message: 'Give this weight set a name so you can pick it from the dropdown later.',
        prompt: {
          model: currentName ? `${currentName} copy` : '',
          type: 'text',
        },
        cancel: true,
        persistent: true,
      }).onOk((name) => {
        const trimmed = (name || '').trim()
        if (!trimmed) return
        const profile = weightsStore.createProfile(trimmed, editorWeights.value)
        if (!profile) return
        selectedOption.value = profile.id
        weightsStore.setActiveProfileId(profile.id)
        activateWeights(profile.weights)
      })
    }

    function resetToDefault() {
      editorWeights.value = cloneWeights(weightsStore.defaultWeights)
    }

    function confirmDelete() {
      if (!selectedProfile.value) return
      const profile = selectedProfile.value
      $q.dialog({
        title: 'Delete Weight Set',
        message: `Delete "${profile.name}"? This can't be undone.`,
        cancel: true,
        persistent: true,
        ok: { label: 'Delete', color: 'negative', flat: true },
      }).onOk(() => {
        weightsStore.deleteProfile(profile.id)
        selectedOption.value = DEFAULT_OPTION_VALUE
        loadEditorFromSelection()
        activateWeights(weightsStore.defaultWeights)
      })
    }

    onMounted(async () => {
      weightsStore.load()

      try {
        const config = await playerService.getConfig()
        if (config?.attributeWeights) {
          weightsStore.setDefaultWeights(config.attributeWeights)
        }
      } catch (_error) {}

      selectedOption.value = weightsStore.activeProfileId || DEFAULT_OPTION_VALUE
      loadEditorFromSelection()
    })

    watch(showDialog, (isOpen) => {
      if (isOpen) {
        selectedOption.value = weightsStore.activeProfileId || DEFAULT_OPTION_VALUE
        loadEditorFromSelection()
      }
    })

    function closeModal() {
      emit('update:modelValue', false)
    }

    return {
      showDialog,
      closeModal,
      applying,
      activeGroup,
      categoryGroups,
      profileOptions,
      selectedOption,
      selectedProfile,
      selectedPreset,
      editorWeights,
      isDirty,
      canSave,
      categoryAttributeKeys,
      attributeLabel,
      setWeight,
      onSelectProfile,
      saveCurrentProfile,
      promptSaveAs,
      resetToDefault,
      confirmDelete,
    }
  },
})
</script>

<style lang="scss" scoped>
.weights-modal {
    .q-dialog__inner {
        padding: 0;
    }
}

.weights-card {
    width: 100%;
    max-width: 900px;
    margin: 2rem auto;
    max-height: 90vh;
    overflow-y: auto;
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
}

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

.weights-content {
    padding: 2rem;
}

.info-card {
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
}

.info-text {
    color: var(--text-secondary);
    font-size: 0.9rem;
    line-height: 1.5;

    p {
        margin: 0 0 0.75rem 0;

        &:last-child {
            margin-bottom: 0;
        }
    }
}

.profile-bar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
    margin-bottom: 1.5rem;
}

.profile-select {
    min-width: 280px;
}

.dirty-badge {
    font-weight: 600;
}

.preset-description {
    color: var(--text-secondary);
    font-size: 0.85rem;
    line-height: 1.4;
    margin: -0.75rem 0 1rem 0;
}

.weights-tabs {
    border-bottom: none;
}

.weights-panel {
    padding: 1rem 0 0 0;
}

.category-expansion {
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-sm);
    margin-bottom: 0.75rem;
    overflow: hidden;
}

.category-header {
    background: var(--surface-raised);
}

.attribute-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.75rem;
    padding: 1rem;
}

.attribute-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
}

.attribute-label {
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.attribute-abbr {
    font-weight: 600;
    font-size: 0.85rem;
    color: var(--text-primary);
}

.attribute-name {
    font-size: 0.75rem;
    color: var(--text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.attribute-input {
    width: 5rem;
    flex-shrink: 0;
}

.weights-actions {
    padding: 1rem 2rem;
    background: var(--surface-raised);
}

.close-action-btn {
    color: var(--text-secondary);

    &:hover {
        background: var(--accent-soft);
        color: var(--accent);
    }
}

@media (max-width: 600px) {
    .weights-card {
        margin: 1rem;
        max-height: 95vh;
    }

    .weights-content {
        padding: 1rem;
    }

    .profile-select {
        min-width: 100%;
    }
}
</style>
