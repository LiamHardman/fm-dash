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
            <!-- Dialog chrome: header (icon/title/close), the same convention used by
                 PlayerDetailDialog/BargainHunterDialog — an icon, a title, then a close
                 button, all in normal flow. -->
            <div class="dialog-chrome">
                <div class="dialog-chrome__header">
                    <q-icon name="settings" class="dialog-chrome__icon" />
                    <div class="dialog-chrome__title">Settings</div>
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

            <q-card-section class="settings-content">
                <div class="settings-sections">
                    <!-- Rating Calculation Section -->
                    <!-- Appearance Section -->
                    <q-expansion-item
                        expand-separator
                        icon="palette"
                        label="Appearance"
                        caption="Accent color and density"
                        header-class="settings-expansion-header"
                        class="settings-expansion"
                        :default-opened="false"
                    >
                        <q-card flat class="expansion-content">
                            <q-card-section>
                                <div class="section-description">
                                    Personalize the app's accent color and how dense the layout feels.
                                </div>

                                <div class="appearance-block">
                                    <div class="appearance-block__label">Accent color</div>
                                    <div class="accent-swatches">
                                        <button
                                            v-for="swatch in accentSwatches"
                                            :key="swatch"
                                            type="button"
                                            class="accent-swatch"
                                            :class="{ 'accent-swatch--selected': isAccentSelected(swatch) }"
                                            :style="{ backgroundColor: swatch }"
                                            :aria-label="`Set accent color to ${swatch}`"
                                            @click="accentColor = swatch"
                                        >
                                            <q-icon v-if="isAccentSelected(swatch)" name="check" size="1rem" color="white" />
                                        </button>
                                        <q-btn
                                            round
                                            flat
                                            dense
                                            icon="colorize"
                                            class="accent-swatch accent-swatch--custom"
                                            aria-label="Pick a custom accent color"
                                        >
                                            <q-popup-proxy cover transition-show="scale" transition-hide="scale">
                                                <q-color v-model="accentColor" no-header no-footer />
                                            </q-popup-proxy>
                                        </q-btn>
                                        <q-btn
                                            flat
                                            dense
                                            no-caps
                                            label="Reset"
                                            class="accent-reset-btn"
                                            @click="accentColor = ''"
                                        />
                                    </div>
                                </div>

                                <div class="appearance-block">
                                    <div class="appearance-block__label">Density</div>
                                    <q-btn-toggle
                                        v-model="density"
                                        no-caps
                                        unelevated
                                        toggle-color="primary"
                                        color="grey-3"
                                        text-color="grey-8"
                                        :options="[
                                            { label: 'Comfortable', value: 'comfortable' },
                                            { label: 'Compact', value: 'compact' },
                                        ]"
                                    />
                                </div>
                            </q-card-section>
                        </q-card>
                    </q-expansion-item>

                    <q-expansion-item
                        expand-separator
                        icon="assessment"
                        label="Rating Calculation Method"
                        caption="Configure how player ratings are calculated"
                        header-class="settings-expansion-header"
                        class="settings-expansion"
                        :default-opened="false"
                    >
                        <q-card flat class="expansion-content">
                            <q-card-section>
                                <div class="section-description">
                                    Choose how player ratings are calculated throughout the application.
                                </div>

                                <div class="rating-method-options">
                                    <q-card 
                                        :class="{
                                            'method-card': true,
                                            'method-card--selected': useScaledRatings,
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

                                <div class="rating-info">
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
                            </q-card-section>
                        </q-card>
                    </q-expansion-item>

                    <!-- Overall Calculation Source Section -->
                    <q-expansion-item
                        expand-separator
                        icon="insights"
                        label="Overall Rating Source"
                        caption="Choose how overall is computed when no specific role is selected"
                        header-class="settings-expansion-header"
                        class="settings-expansion"
                        :default-opened="false"
                    >
                        <q-card flat class="expansion-content">
                            <q-card-section>
                                <div class="section-description">
                                    When enabled, overall uses FIFA category stats with position-based weights unless a role filter is applied. Otherwise, role-based overall from FM attributes is used.
                                </div>

                                <q-card flat bordered class="option-card q-mt-md">
                                    <q-card-section class="option-content">
                                        <div class="option-header">
                                            <div class="option-info">
                                                <q-icon name="tune" size="1.5rem" class="option-icon" />
                                                <div class="option-text">
                                                    <div class="option-title">Use stat summary ratings to calculate overall rating</div>
                                                    <div class="option-description">Off by default. Applies FIFA category weights per position when no specific role is selected.</div>
                                                </div>
                                            </div>
                                            <q-toggle
                                                v-model="useStatSummaryForOverall"
                                                color="primary"
                                                size="lg"
                                            />
                                        </div>
                                    </q-card-section>
                                </q-card>
                            </q-card-section>
                        </q-card>
                    </q-expansion-item>

                    <!-- Display Preferences Section -->
                    <q-expansion-item
                        expand-separator
                        icon="view_module"
                        label="Display Preferences"
                        caption="Customize player faces, team logos, and layout options"
                        header-class="settings-expansion-header"
                        class="settings-expansion"
                        :default-opened="false"
                    >
                        <q-card flat class="expansion-content">
                            <q-card-section>
                                <div class="section-description">
                                    Configure what visual elements are displayed throughout the application.
                                </div>

                                <div class="display-options">
                                    <!-- Faces Toggle -->
                                    <q-card flat bordered class="option-card">
                                        <q-card-section class="option-content">
                                            <div class="option-header">
                                                <div class="option-info">
                                                    <q-icon name="face" size="1.5rem" class="option-icon" />
                                                    <div class="option-text">
                                                        <div class="option-title">Player Faces</div>
                                                        <div class="option-description">Show or hide player face images in player cards and tables</div>
                                                    </div>
                                                </div>
                                                <q-toggle
                                                    v-model="showFaces"
                                                    color="primary"
                                                    size="lg"
                                                />
                                            </div>
                                        </q-card-section>
                                    </q-card>

                                    <!-- Logos Toggle -->
                                    <q-card flat bordered class="option-card">
                                        <q-card-section class="option-content">
                                            <div class="option-header">
                                                <div class="option-info">
                                                    <q-icon name="shield" size="1.5rem" class="option-icon" />
                                                    <div class="option-text">
                                                        <div class="option-title">Team Logos</div>
                                                        <div class="option-description">Show or hide team logo images in team displays and player cards</div>
                                                        <div class="option-disclaimer">Note: Saves with no real name fixes may have incorect logos!</div>
                                                    </div>
                                                </div>
                                                <q-toggle
                                                    v-model="showLogos"
                                                    color="primary"
                                                    size="lg"
                                                />
                                            </div>
                                        </q-card-section>
                                    </q-card>

                                    <!-- Logo Corrections Toggle -->
                                    <q-card flat bordered class="option-card">
                                        <q-card-section class="option-content">
                                            <div class="option-header">
                                                <div class="option-info">
                                                    <q-icon name="rate_review" size="1.5rem" class="option-icon" />
                                                    <div class="option-text">
                                                        <div class="option-title">Logo Correction Buttons</div>
                                                        <div class="option-description">Show tick/cross buttons next to club logos to confirm or reject automatic matches. Saved corrections are used for all future lookups.</div>
                                                    </div>
                                                </div>
                                                <q-toggle
                                                    v-model="showLogoCorrections"
                                                    color="primary"
                                                    size="lg"
                                                />
                                            </div>
                                        </q-card-section>
                                    </q-card>

                                    <!-- Attribute Masks Toggle -->
                                    <q-card flat bordered class="option-card">
                                        <q-card-section class="option-content">
                                            <div class="option-header">
                                                <div class="option-info">
                                                    <q-icon name="visibility_off" size="1.5rem" class="option-icon" />
                                                    <div class="option-text">
                                                        <div class="option-title">View Attribute Masks</div>
                                                        <div class="option-description">Show attribute ranges (e.g., 12-18) instead of the calculated midpoint</div>
                                                    </div>
                                                </div>
                                                <q-toggle
                                                    v-model="showAttributeMasks"
                                                    color="primary"
                                                    size="lg"
                                                />
                                            </div>
                                        </q-card-section>
                                    </q-card>

                                    <!-- Current Ability (CA) Toggle -->
                                    <q-card flat bordered class="option-card">
                                        <q-card-section class="option-content">
                                            <div class="option-header">
                                                <div class="option-info">
                                                    <q-icon name="psychology" size="1.5rem" class="option-icon" />
                                                    <div class="option-text">
                                                        <div class="option-title">Current Ability (CA) Column</div>
                                                        <div class="option-description">Show the estimated Current Ability score (0–200) in the player table and detail view. Based on the fm21-cas algorithm.</div>
                                                    </div>
                                                </div>
                                                <q-toggle
                                                    v-model="showCA"
                                                    color="teal"
                                                    size="lg"
                                                />
                                            </div>
                                        </q-card-section>
                                    </q-card>
                                </div>

                                <div class="display-info">
                                    <q-card flat bordered class="info-card">
                                        <q-card-section>
                                            <div class="info-text">
                                                <p><strong>Note:</strong> These display preferences will take effect immediately and affect:</p>
                                                <ul>
                                                    <li>Player tables and search results</li>
                                                    <li>Individual player detail views</li>
                                                    <li>Team displays and comparisons</li>
                                                    <li>Dashboard and overview screens</li>
                                                </ul>
                                                <p>Settings are saved automatically and will persist across browser sessions.</p>
                                            </div>
                                        </q-card-section>
                                    </q-card>
                                </div>
                            </q-card-section>
                        </q-card>
                    </q-expansion-item>



                    <q-expansion-item
                        expand-separator
                        icon="person_search"
                        label="AI Features"
                        caption="Bring your own OpenAI API key for the Chatbot, Who to Sign, and AI Scout Report"
                        header-class="settings-expansion-header"
                        class="settings-expansion"
                        :default-opened="false"
                    >
                        <q-card flat class="expansion-content">
                            <q-card-section>
                                <div class="section-description">
                                    The Chatbot, "Who to Sign", and AI Scout Report features call OpenAI's API
                                    directly using your own key — it is never sent anywhere except OpenAI (or a
                                    custom endpoint you configure below), and this app does not provide a shared
                                    key. Your key is stored only in this browser.
                                </div>
                                <q-input
                                    filled
                                    v-model="openaiApiKey"
                                    type="password"
                                    label="OpenAI API key"
                                    placeholder="sk-..."
                                    class="q-mt-md"
                                    autocomplete="off"
                                >
                                    <template v-slot:prepend>
                                        <q-icon name="vpn_key" />
                                    </template>
                                </q-input>
                                <q-input
                                    filled
                                    v-model="openaiBaseUrl"
                                    label="Base URL (optional)"
                                    placeholder="https://api.openai.com"
                                    class="q-mt-md"
                                    autocomplete="off"
                                >
                                    <template v-slot:prepend>
                                        <q-icon name="link" />
                                    </template>
                                </q-input>
                                <p class="text-caption q-mt-sm text-warning">
                                    Your API key is sent to whatever URL you configure here — only point this at
                                    an endpoint you trust.
                                </p>
                                <q-input
                                    filled
                                    v-model="openaiModel"
                                    label="Model (optional)"
                                    placeholder="gpt-5.6-luna"
                                    class="q-mt-md"
                                    autocomplete="off"
                                >
                                    <template v-slot:prepend>
                                        <q-icon name="smart_toy" />
                                    </template>
                                </q-input>
                                <p class="text-caption q-mt-sm">
                                    Leave Base URL or Model blank to use this app's defaults — each is independent,
                                    so you can override just one if you want. Settings are saved automatically and
                                    persist across browser sessions.
                                </p>
                            </q-card-section>
                        </q-card>
                    </q-expansion-item>

                    <q-expansion-item
                        expand-separator
                        icon="filter_alt"
                        label="Default Filters"
                        caption="Set default age ranges, positions, and other filter preferences"
                        header-class="settings-expansion-header"
                        class="settings-expansion"
                        :default-opened="false"
                    >
                        <q-card flat class="expansion-content">
                            <q-card-section>
                                <div class="coming-soon">
                                    <q-icon name="construction" size="2rem" class="q-mb-md" />
                                    <p>Default filters coming soon...</p>
                                    <p class="text-caption">Configure default age ranges, positions, and filter preferences</p>
                                </div>
                            </q-card-section>
                        </q-card>
                    </q-expansion-item>

                    <q-expansion-item
                        expand-separator
                        icon="school"
                        label="Help & Tutorial"
                        caption="Get help and view the first time setup guide"
                        header-class="settings-expansion-header"
                        class="settings-expansion"
                    >
                        <q-card flat class="expansion-content">
                            <q-card-section>
                                <div class="help-section">
                                    <div class="help-description">
                                        <p>Need help getting started? View our comprehensive first time setup guide.</p>
                                    </div>
                                    <div class="help-actions">
                                        <q-btn
                                            unelevated
                                            color="primary"
                                            icon="school"
                                            label="Open First Time Guide"
                                            @click="showTutorial"
                                            class="tutorial-btn"
                                        />
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>
                    </q-expansion-item>
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
import { useQuasar } from 'quasar'
import { computed, defineComponent, onMounted, ref } from 'vue'
import { usePlayerStore } from '@/stores/playerStore'
import { useUiStore } from '@/stores/uiStore'
import playerService from '../services/playerService'

export default defineComponent({
  name: 'SettingsModal',
  components: {},
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

    const uiStore = useUiStore()
    const playerStore = usePlayerStore()
    const $q = useQuasar()

    const useScaledRatings = computed({
      get: () => uiStore.useScaledRatings,
      set: (value) => uiStore.setRatingCalculation(value),
    })

    const showFaces = computed({
      get: () => uiStore.showFaces,
      set: (value) => uiStore.setFacesDisplay(value),
    })

    const showLogos = computed({
      get: () => uiStore.showLogos,
      set: (value) => uiStore.setLogosDisplay(value),
    })

    const showLogoCorrections = computed({
      get: () => uiStore.showLogoCorrections,
      set: (value) => uiStore.setLogoCorrections(value),
    })

    const openaiApiKey = computed({
      get: () => uiStore.openaiApiKey,
      set: (value) => uiStore.setOpenaiApiKey(value),
    })

    const openaiBaseUrl = computed({
      get: () => uiStore.openaiBaseUrl,
      set: (value) => uiStore.setOpenaiBaseUrl(value),
    })

    const openaiModel = computed({
      get: () => uiStore.openaiModel,
      set: (value) => uiStore.setOpenaiModel(value),
    })

    const showAttributeMasks = computed({
      get: () => uiStore.showAttributeMasks,
      set: (_value) => uiStore.toggleAttributeMasks(),
    })

    const showCA = computed({
      get: () => uiStore.showCA,
      set: (value) => uiStore.setCADisplay(value),
    })

    const useStatSummaryForOverall = computed({
      get: () => uiStore.useStatSummaryForOverall,
      set: (value) => uiStore.setStatSummaryOverall(value),
    })

    const accentColor = computed({
      get: () => uiStore.accentColor,
      set: (value) => uiStore.setAccentColor(value),
    })

    const accentSwatches = [
      '#1a237e', // brand default (indigo)
      '#2563eb', // blue
      '#0d9488', // teal
      '#16a34a', // emerald
      '#d97706', // amber
      '#e11d48', // rose
      '#7c3aed', // purple
      '#475569', // slate
    ]

    const isAccentSelected = (swatch) => {
      const current = (uiStore.accentColor || accentSwatches[0]).toLowerCase()
      return current === swatch.toLowerCase()
    }

    const density = computed({
      get: () => uiStore.density,
      set: (value) => uiStore.setDensity(value),
    })

    const isLoading = ref(false)
    const activeTab = ref('general')

    // Load backend configuration on component mount
    onMounted(async () => {
      try {
        const config = await playerService.getConfig()
        if (config.useScaledRatings !== undefined) {
          uiStore.setRatingCalculation(config.useScaledRatings)
        }
      } catch (_error) {}
    })

    const setRatingMethod = async (useScaled) => {
      if (isLoading.value) return

      isLoading.value = true

      try {
        // Update backend first
        await playerService.updateConfig({
          useScaledRatings: useScaled,
        })

        // Update local store
        uiStore.setRatingCalculation(useScaled)

        // Trigger data refresh if we have a current dataset
        if (playerStore.currentDatasetId) {
          // Rating calculation method changed - data will refresh automatically via store reactivity
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
              handler: () => {},
            },
          ],
        })
      } catch (_error) {
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
              handler: () => {},
            },
          ],
        })
      } finally {
        isLoading.value = false
      }
    }

    const closeModal = () => {
      emit('update:modelValue', false)
    }

    const showTutorial = () => {
      closeModal()
      uiStore.showTutorial()
    }

    return {
      showDialog,
      closeModal,
      useScaledRatings,
      useStatSummaryForOverall,
      setRatingMethod,
      showFaces,
      showLogos,
      showLogoCorrections,
      showAttributeMasks,
      showCA,
      openaiApiKey,
      openaiBaseUrl,
      openaiModel,
      accentColor,
      accentSwatches,
      isAccentSelected,
      density,
      isLoading,
      activeTab,
      showTutorial,
    }
  },
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
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
}

// Dialog chrome: unified header convention shared with PlayerDetailDialog /
// BargainHunterDialog — icon, title, actions, close, all in normal flow.
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

.settings-content {
    padding: 2rem;
}

.settings-sections {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.settings-expansion {
    .q-expansion-item__content {
        padding: 0;
    }
}

.settings-expansion-header {
    padding: 1rem 1.5rem;
    background: var(--surface-raised);
}

.expansion-content {
    padding: 1rem;
}

.section-description {
    margin-bottom: 1.5rem;
    color: var(--text-secondary);
    font-size: 0.95rem;
}

.rating-method-options {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 2rem;
}

.method-card {
    border: 2px solid var(--surface-border-strong);
    cursor: pointer;
    transition: all 0.3s ease;
    background: var(--surface-card);

    &:hover {
        border-color: var(--accent);
        box-shadow: var(--shadow-2);
    }

    &--selected {
        border-color: var(--accent);
        background: var(--accent-soft);
    }

    &--disabled {
        opacity: 0.6;
        cursor: not-allowed;

        &:hover {
            border-color: var(--surface-border-strong);
            box-shadow: none;
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
    color: var(--text-secondary);

    ul {
        margin: 0.5rem 0;
        padding-left: 1.5rem;

        li {
            margin-bottom: 0.25rem;
        }
    }
}

.rating-preview {
    background: var(--surface-raised);
    border-radius: var(--radius-sm);
    padding: 1.5rem;
    border: 1px solid var(--surface-border);
}

.preview-header {
    display: flex;
    align-items: center;
    margin-bottom: 1rem;
    font-weight: 600;
    color: var(--text-primary);
}

.preview-content {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
}

.preview-example {
    flex: 1;
    min-width: 200px;
    background: var(--surface-card);
    border-radius: var(--radius-sm);
    padding: 1rem;
    border: 1px solid var(--surface-border);
}

.example-header {
    font-weight: 600;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: var(--text-primary);
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
    color: var(--text-secondary);
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

.rating-info {
    margin-top: 1rem;
}

.info-card {
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
}

.info-text {
    color: var(--text-secondary);
    font-size: 0.9rem;

    ul {
        margin: 0.5rem 0;
        padding-left: 1.5rem;

        li {
            margin-bottom: 0.25rem;
        }
    }
}

.coming-soon {
    text-align: center;
    padding: 2rem;
    color: var(--text-secondary);

    p {
        margin: 0.5rem 0;
    }

    .q-icon {
        color: var(--text-muted);
    }
}

.help-section {
    padding: 1rem 0;

    .help-description {
        margin-bottom: 1.5rem;

        p {
            color: var(--text-secondary);
            line-height: 1.6;
            margin: 0;
        }
    }

    .help-actions {
        text-align: center;

        .tutorial-btn {
            padding: 0.8rem 1.5rem;
            font-weight: 600;
        }
    }
}

.display-options {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 1.5rem;
}

.appearance-block {
    margin-bottom: 1.5rem;

    &:last-child {
        margin-bottom: 0;
    }

    &__label {
        font-size: 0.85rem;
        font-weight: 600;
        color: var(--text-secondary);
        margin-bottom: 0.65rem;
    }
}

.accent-swatches {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex-wrap: wrap;
}

.accent-swatch {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    border: 2px solid transparent;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 0;
    transition: transform 0.15s ease, border-color 0.15s ease;

    &:hover {
        transform: scale(1.08);
    }

    &--selected {
        border-color: var(--text-primary);
    }

    &--custom {
        background: conic-gradient(from 180deg, #e11d48, #d97706, #16a34a, #0d9488, #2563eb, #7c3aed, #e11d48);
        color: white;
    }
}

.accent-reset-btn {
    color: var(--text-secondary);
    font-size: 0.8rem;
}

.option-card {
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
    transition: all 0.2s ease;

    &:hover {
        border-color: var(--surface-border-strong);
        box-shadow: var(--shadow-1);
    }
}

.option-content {
    padding: 1.25rem;
}

.option-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
}

.option-info {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex: 1;
}

.option-icon {
    color: var(--accent);
    flex-shrink: 0;
}

.option-text {
    flex: 1;
}

.option-title {
    font-weight: 600;
    font-size: 1rem;
    color: var(--text-primary);
    margin-bottom: 0.25rem;
}

.option-description {
    font-size: 0.875rem;
    color: var(--text-secondary);
    line-height: 1.4;
}

.option-disclaimer {
    // Warning-tier color, kept semantic/hardcoded per established precedent.
    font-size: 0.8rem;
    color: #f57c00;
    margin-top: 0.5rem;
    font-style: italic;

    .body--dark & {
        color: #ffb74d;
    }
}

.display-info {
    margin-top: 1rem;
}

.settings-actions {
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