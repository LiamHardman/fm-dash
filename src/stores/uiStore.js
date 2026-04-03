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
  const showCA = ref(false) // Default to hiding Current Ability (CA) column

  // Tutorial state
  const showFirstTimeTutorial = ref(false) // Control tutorial modal visibility

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

  // Initialize all settings
  function initSettings() {
    initDarkMode()
    initRatingCalculation()
    initStatSummaryOverall()
    initFacesDisplay()
    initLogosDisplay()
    initAttributeMasksDisplay()
    initCADisplay()
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
    showAttributeMasks,
    toggleAttributeMasks,
    initAttributeMasksDisplay,
    showCA,
    toggleCA,
    setCADisplay,
    initCADisplay,
    showFirstTimeTutorial,
    showTutorial,
    hideTutorial,
    initSettings,
  }
})
