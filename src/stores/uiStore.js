import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  // Initialize with a default value, initDarkMode will override this.
  // Defaulting to true here to align with the user's intent for dark mode by default.
  const isDarkModeActive = ref(true)

  // Rating calculation method preference
  const useScaledRatings = ref(true) // Default to new scaled ratings

  // Overall calculation preference: use FIFA stat summary by position (off by default)
  const useStatSummaryForOverall = ref(false)

  // Display preferences
  const showFaces = ref(true) // Default to showing faces
  const showLogos = ref(true) // Default to showing logos
  const showAttributeMasks = ref(true) // Default to showing attribute masks
  const showLogoCorrections = ref(true) // Default to showing logo correction buttons
  const showCA = ref(false) // Default to hiding Current Ability (CA) column

  // Tutorial state
  const showFirstTimeTutorial = ref(false) // Control tutorial modal visibility

  // Accent color override (hex string, e.g. '#1a237e'). Empty string means
  // "use the brand default" -- see src/css/tokens.scss's --accent-user fallback.
  const accentColor = ref('')

  // Density mode governs component-internal spacing/row-height via the
  // --density-* tokens in src/css/tokens.scss (body.density-compact).
  const density = ref('comfortable') // 'comfortable' | 'compact'

  // Sidebar nav customization: hidden route names and a custom order.
  // Empty arrays mean "show everything, default order".
  const sidebarHiddenItems = ref([])
  const sidebarItemOrder = ref([])

  // Dashboard home widget customization: hidden widget ids and a custom order.
  const dashboardHiddenWidgets = ref([])
  const dashboardWidgetOrder = ref([])

  // Player table column customization: hidden column names and a custom order.
  // The 'name' column is never included here -- it's always forced visible and
  // first, so pinning it isn't a concept that needs storage. "Pinning" any other
  // column is expressed by simply ordering it to the front of this list (right
  // after 'name'); there's no separate pinned-ids array, see ticket 11's Answer.
  const playerTableHiddenColumns = ref([])
  const playerTableColumnOrder = ref([])

  // User's own OpenAI API key for the "Who to Sign" scouting feature (bring-your-own-key)
  const openaiApiKey = ref('')

  // Configurable LLM endpoint/model (map decision, see .scratch/llm-refinements/issues/
  // 01-configurable-llm-endpoint-and-model.md), shared by all three LLM features. Both
  // empty by default -- each independently falls back to the backend's own hardcoded
  // default when unset.
  const openaiBaseUrl = ref('')
  const openaiModel = ref('')

  // Function to toggle dark mode
  function toggleDarkMode() {
    // Directly toggle the current state
    isDarkModeActive.value = !isDarkModeActive.value

    // Apply dark mode to document body as fallback
    if (isDarkModeActive.value) {
      document.body.classList.add('body--dark')
    } else {
      document.body.classList.remove('body--dark')
    }

    // Try to use Quasar if available - use a different approach
    if (typeof window !== 'undefined' && window.$q) {
      window.$q.dark.set(isDarkModeActive.value)
    }

    try {
      localStorage.setItem('darkMode', isDarkModeActive.value ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to toggle rating calculation method
  function toggleRatingCalculation() {
    useScaledRatings.value = !useScaledRatings.value
    try {
      localStorage.setItem('useScaledRatings', useScaledRatings.value ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to set rating calculation method directly
  function setRatingCalculation(useScaled) {
    useScaledRatings.value = useScaled
    try {
      localStorage.setItem('useScaledRatings', useScaled ? 'true' : 'false')
    } catch (_e) {}
  }

  // Toggle and set functions for stat summary overall preference
  function toggleStatSummaryOverall() {
    useStatSummaryForOverall.value = !useStatSummaryForOverall.value
    try {
      localStorage.setItem(
        'useStatSummaryForOverall',
        useStatSummaryForOverall.value ? 'true' : 'false'
      )
    } catch (_e) {}
  }

  function setStatSummaryOverall(enabled) {
    useStatSummaryForOverall.value = !!enabled
    try {
      localStorage.setItem('useStatSummaryForOverall', enabled ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to toggle faces display
  function toggleFaces() {
    showFaces.value = !showFaces.value
    try {
      localStorage.setItem('showFaces', showFaces.value ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to set faces display directly
  function setFacesDisplay(showFacesEnabled) {
    showFaces.value = showFacesEnabled
    try {
      localStorage.setItem('showFaces', showFacesEnabled ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to toggle logos display
  function toggleLogos() {
    showLogos.value = !showLogos.value
    try {
      localStorage.setItem('showLogos', showLogos.value ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to set logos display directly
  function setLogosDisplay(showLogosEnabled) {
    showLogos.value = showLogosEnabled
    try {
      localStorage.setItem('showLogos', showLogosEnabled ? 'true' : 'false')
    } catch (_e) {}
  }

  function toggleLogoCorrections() {
    showLogoCorrections.value = !showLogoCorrections.value
    try {
      localStorage.setItem('showLogoCorrections', showLogoCorrections.value ? 'true' : 'false')
    } catch (_e) {}
  }

  function setLogoCorrections(enabled) {
    showLogoCorrections.value = enabled
    try {
      localStorage.setItem('showLogoCorrections', enabled ? 'true' : 'false')
    } catch (_e) {}
  }

  function initLogoCorrections() {
    try {
      const storedPreference = localStorage.getItem('showLogoCorrections')
      if (storedPreference !== null) {
        showLogoCorrections.value = storedPreference === 'true'
      }
    } catch (_e) {}
  }

  // Function to toggle attribute masks display
  function toggleAttributeMasks() {
    showAttributeMasks.value = !showAttributeMasks.value
    try {
      localStorage.setItem('showAttributeMasks', showAttributeMasks.value ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to toggle CA (Current Ability) column display
  function toggleCA() {
    showCA.value = !showCA.value
    try {
      localStorage.setItem('showCA', showCA.value ? 'true' : 'false')
    } catch (_e) {}
  }

  // Function to set CA display directly
  function setCADisplay(showCAEnabled) {
    showCA.value = showCAEnabled
    try {
      localStorage.setItem('showCA', showCAEnabled ? 'true' : 'false')
    } catch (_e) {}
  }

  // Apply the current accent color to the document root as a CSS custom
  // property. tokens.scss's --accent falls back to the brand default via
  // var(--accent-user, <default>), so an empty string here is a valid
  // "use default" state -- just clear the property.
  //
  // Also mirrors the value onto Quasar's own --q-primary custom property
  // (ticket 13 final-QA finding: Quasar's built CSS already themes every
  // color="primary" component -- buttons, spinners, tabs, EmptyState
  // actions, etc. -- via var(--q-primary), but nothing was ever writing to
  // it, so the accent picker silently missed every native-Quasar-colored
  // element app-wide even though app-authored CSS using var(--accent)
  // picked it up correctly). Same "unset to fall back to the compiled
  // default" behavior as --accent-user.
  function applyAccentColor() {
    if (typeof document === 'undefined') return
    if (accentColor.value) {
      document.documentElement.style.setProperty('--accent-user', accentColor.value)
      document.documentElement.style.setProperty('--q-primary', accentColor.value)
    } else {
      document.documentElement.style.removeProperty('--accent-user')
      document.documentElement.style.removeProperty('--q-primary')
    }
  }

  function setAccentColor(hex) {
    accentColor.value = hex || ''
    applyAccentColor()
    try {
      localStorage.setItem('accentColor', accentColor.value)
    } catch (_e) {}
  }

  function initAccentColor() {
    try {
      const stored = localStorage.getItem('accentColor')
      if (stored !== null) {
        accentColor.value = stored
      }
    } catch (_e) {}
    applyAccentColor()
  }

  // Apply the current density to the document body as a class, matching the
  // dark-mode pattern (body.density-compact toggles the --density-* tokens).
  function applyDensity() {
    if (typeof document === 'undefined') return
    document.body.classList.toggle('density-compact', density.value === 'compact')
  }

  function setDensity(value) {
    density.value = value === 'compact' ? 'compact' : 'comfortable'
    applyDensity()
    try {
      localStorage.setItem('density', density.value)
    } catch (_e) {}
  }

  function toggleDensity() {
    setDensity(density.value === 'compact' ? 'comfortable' : 'compact')
  }

  function initDensity() {
    try {
      const stored = localStorage.getItem('density')
      if (stored !== null) {
        density.value = stored === 'compact' ? 'compact' : 'comfortable'
      }
    } catch (_e) {}
    applyDensity()
  }

  // Generic persisted-list helpers backing sidebar and dashboard-widget
  // customization (tickets 02 and 03 of the UI redesign map). Each pair
  // stores a hidden-id array and an explicit order array as JSON.
  function _loadJsonArray(key) {
    try {
      const stored = localStorage.getItem(key)
      if (stored !== null) return JSON.parse(stored)
    } catch (_e) {}
    return []
  }

  function _saveJsonArray(key, arr) {
    try {
      localStorage.setItem(key, JSON.stringify(arr))
    } catch (_e) {}
  }

  function setSidebarHiddenItems(ids) {
    sidebarHiddenItems.value = ids || []
    _saveJsonArray('sidebarHiddenItems', sidebarHiddenItems.value)
  }

  function setSidebarItemOrder(ids) {
    sidebarItemOrder.value = ids || []
    _saveJsonArray('sidebarItemOrder', sidebarItemOrder.value)
  }

  function initSidebarCustomization() {
    sidebarHiddenItems.value = _loadJsonArray('sidebarHiddenItems')
    sidebarItemOrder.value = _loadJsonArray('sidebarItemOrder')
  }

  function setDashboardHiddenWidgets(ids) {
    dashboardHiddenWidgets.value = ids || []
    _saveJsonArray('dashboardHiddenWidgets', dashboardHiddenWidgets.value)
  }

  function setDashboardWidgetOrder(ids) {
    dashboardWidgetOrder.value = ids || []
    _saveJsonArray('dashboardWidgetOrder', dashboardWidgetOrder.value)
  }

  function initDashboardCustomization() {
    dashboardHiddenWidgets.value = _loadJsonArray('dashboardHiddenWidgets')
    dashboardWidgetOrder.value = _loadJsonArray('dashboardWidgetOrder')
  }

  function setPlayerTableHiddenColumns(ids) {
    playerTableHiddenColumns.value = ids || []
    _saveJsonArray('playerTableHiddenColumns', playerTableHiddenColumns.value)
  }

  function setPlayerTableColumnOrder(ids) {
    playerTableColumnOrder.value = ids || []
    _saveJsonArray('playerTableColumnOrder', playerTableColumnOrder.value)
  }

  function initPlayerTableColumnCustomization() {
    playerTableHiddenColumns.value = _loadJsonArray('playerTableHiddenColumns')
    playerTableColumnOrder.value = _loadJsonArray('playerTableColumnOrder')
  }

  // Initialize dark mode from localStorage or system preference
  function initDarkMode() {
    let darkModePreference = true
    try {
      const storedPreference = localStorage.getItem('darkMode')
      if (storedPreference !== null) {
        darkModePreference = storedPreference === 'true'
      }
    } catch (_e) {
      darkModePreference = true
    }

    // Set the ref
    isDarkModeActive.value = darkModePreference

    // Apply dark mode to document body as fallback
    if (darkModePreference) {
      document.body.classList.add('body--dark')
    } else {
      document.body.classList.remove('body--dark')
    }

    // Try to use Quasar if available - use a different approach
    if (typeof window !== 'undefined' && window.$q) {
      window.$q.dark.set(darkModePreference)
    }
  }

  // Initialize rating calculation preferences
  function initRatingCalculation() {
    try {
      const storedPreference = localStorage.getItem('useScaledRatings')
      if (storedPreference !== null) {
        useScaledRatings.value = storedPreference === 'true'
      }
    } catch (_e) {}
  }

  // Initialize overall calculation preference (FIFA stat summary)
  function initStatSummaryOverall() {
    try {
      const storedPreference = localStorage.getItem('useStatSummaryForOverall')
      if (storedPreference !== null) {
        useStatSummaryForOverall.value = storedPreference === 'true'
      }
    } catch (_e) {}
  }

  // Initialize faces display preferences
  function initFacesDisplay() {
    try {
      const storedPreference = localStorage.getItem('showFaces')
      if (storedPreference !== null) {
        showFaces.value = storedPreference === 'true'
      }
    } catch (_e) {}
  }

  // Initialize logos display preferences
  function initLogosDisplay() {
    try {
      const storedPreference = localStorage.getItem('showLogos')
      if (storedPreference !== null) {
        showLogos.value = storedPreference === 'true'
      }
    } catch (_e) {}
  }

  // Initialize attribute masks display preferences
  function initAttributeMasksDisplay() {
    try {
      const storedPreference = localStorage.getItem('showAttributeMasks')
      if (storedPreference !== null) {
        showAttributeMasks.value = storedPreference === 'true'
      }
    } catch (_e) {}
  }

  // Initialize CA display preferences
  function initCADisplay() {
    try {
      const storedPreference = localStorage.getItem('showCA')
      if (storedPreference !== null) {
        showCA.value = storedPreference === 'true'
      }
    } catch (_e) {}
  }

  // Function to set the user's OpenAI API key directly
  function setOpenaiApiKey(key) {
    openaiApiKey.value = key
    try {
      localStorage.setItem('openaiApiKey', key || '')
    } catch (_e) {}
  }

  // Initialize OpenAI API key from localStorage
  function initOpenaiApiKey() {
    try {
      const storedKey = localStorage.getItem('openaiApiKey')
      if (storedKey !== null) {
        openaiApiKey.value = storedKey
      }
    } catch (_e) {}
  }

  // Function to set the user's custom LLM base URL
  function setOpenaiBaseUrl(url) {
    openaiBaseUrl.value = url
    try {
      localStorage.setItem('openaiBaseUrl', url || '')
    } catch (_e) {}
  }

  // Initialize custom LLM base URL from localStorage
  function initOpenaiBaseUrl() {
    try {
      const stored = localStorage.getItem('openaiBaseUrl')
      if (stored !== null) {
        openaiBaseUrl.value = stored
      }
    } catch (_e) {}
  }

  // Function to set the user's custom LLM model name
  function setOpenaiModel(model) {
    openaiModel.value = model
    try {
      localStorage.setItem('openaiModel', model || '')
    } catch (_e) {}
  }

  // Initialize custom LLM model name from localStorage
  function initOpenaiModel() {
    try {
      const stored = localStorage.getItem('openaiModel')
      if (stored !== null) {
        openaiModel.value = stored
      }
    } catch (_e) {}
  }

  // Initialize all settings
  function initSettings() {
    initDarkMode()
    initRatingCalculation()
    initStatSummaryOverall()
    initFacesDisplay()
    initLogosDisplay()
    initLogoCorrections()
    initAttributeMasksDisplay()
    initCADisplay()
    initOpenaiApiKey()
    initOpenaiBaseUrl()
    initOpenaiModel()
    initAccentColor()
    initDensity()
    initSidebarCustomization()
    initDashboardCustomization()
    initPlayerTableColumnCustomization()
  }

  // Tutorial functions
  function showTutorial() {
    showFirstTimeTutorial.value = true
  }

  function hideTutorial() {
    showFirstTimeTutorial.value = false
  }

  return {
    isDarkModeActive,
    toggleDarkMode,
    initDarkMode,
    useScaledRatings,
    toggleRatingCalculation,
    setRatingCalculation,
    initRatingCalculation,
    useStatSummaryForOverall,
    toggleStatSummaryOverall,
    setStatSummaryOverall,
    initStatSummaryOverall,
    showFaces,
    toggleFaces,
    setFacesDisplay,
    initFacesDisplay,
    showLogos,
    toggleLogos,
    setLogosDisplay,
    initLogosDisplay,
    showLogoCorrections,
    toggleLogoCorrections,
    setLogoCorrections,
    initLogoCorrections,
    showAttributeMasks,
    toggleAttributeMasks,
    initAttributeMasksDisplay,
    showCA,
    toggleCA,
    setCADisplay,
    initCADisplay,
    openaiApiKey,
    setOpenaiApiKey,
    initOpenaiApiKey,
    openaiBaseUrl,
    setOpenaiBaseUrl,
    initOpenaiBaseUrl,
    openaiModel,
    setOpenaiModel,
    initOpenaiModel,
    showFirstTimeTutorial,
    showTutorial,
    hideTutorial,
    initSettings,
    accentColor,
    setAccentColor,
    initAccentColor,
    density,
    setDensity,
    toggleDensity,
    initDensity,
    sidebarHiddenItems,
    setSidebarHiddenItems,
    sidebarItemOrder,
    setSidebarItemOrder,
    initSidebarCustomization,
    dashboardHiddenWidgets,
    setDashboardHiddenWidgets,
    dashboardWidgetOrder,
    setDashboardWidgetOrder,
    initDashboardCustomization,
    playerTableHiddenColumns,
    setPlayerTableHiddenColumns,
    playerTableColumnOrder,
    setPlayerTableColumnOrder,
    initPlayerTableColumnCustomization,
  }
})
