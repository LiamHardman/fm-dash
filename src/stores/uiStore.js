import { defineStore } from 'pinia'
import { useQuasar } from 'quasar'
import { ref, watch } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const $q = useQuasar()
  // Initialize with a default value, initDarkMode will override this.
  // Defaulting to true here to align with the user's intent for dark mode by default.
  const isDarkModeActive = ref(true)

  // Notification preferences
  const notificationsEnabled = ref(true)

  // Rating calculation method preference
  const useScaledRatings = ref(true) // Default to new scaled ratings

  // Function to toggle dark mode
  function toggleDarkMode() {
    // Directly toggle the current state
    isDarkModeActive.value = !isDarkModeActive.value
    $q.dark.set(isDarkModeActive.value)
    try {
      localStorage.setItem('darkMode', isDarkModeActive.value ? 'true' : 'false')
    } catch (e) {
      console.warn('Could not save dark mode preference to localStorage:', e)
    }
  }

  // Function to toggle notifications
  function toggleNotifications() {
    notificationsEnabled.value = !notificationsEnabled.value
    try {
      localStorage.setItem('notificationsEnabled', notificationsEnabled.value ? 'true' : 'false')
    } catch (e) {
      console.warn('Could not save notification preference to localStorage:', e)
    }
  }

  // Function to toggle rating calculation method
  function toggleRatingCalculation() {
    useScaledRatings.value = !useScaledRatings.value
    try {
      localStorage.setItem('useScaledRatings', useScaledRatings.value ? 'true' : 'false')
    } catch (e) {
      console.warn('Could not save rating calculation preference to localStorage:', e)
    }
  }

  // Function to set rating calculation method directly
  function setRatingCalculation(useScaled) {
    useScaledRatings.value = useScaled
    try {
      localStorage.setItem('useScaledRatings', useScaled ? 'true' : 'false')
    } catch (e) {
      console.warn('Could not save rating calculation preference to localStorage:', e)
    }
  }

  // Initialize dark mode from localStorage or system preference
  function initDarkMode() {
    let darkModePreference = true
    try {
      const storedPreference = localStorage.getItem('darkMode')
      if (storedPreference !== null) {
        darkModePreference = storedPreference === 'true'
      }
    } catch (e) {
      console.warn('Could not read dark mode preference from localStorage:', e)
      darkModePreference = true
    }
    // Set the ref and Quasar's dark mode
    isDarkModeActive.value = darkModePreference
    $q.dark.set(darkModePreference)
  }

  // Initialize notification preferences
  function initNotifications() {
    try {
      const storedPreference = localStorage.getItem('notificationsEnabled')
      if (storedPreference !== null) {
        notificationsEnabled.value = storedPreference === 'true'
      }
    } catch (e) {
      console.warn('Could not read notification preference from localStorage:', e)
    }
  }

  // Initialize rating calculation preferences
  function initRatingCalculation() {
    try {
      const storedPreference = localStorage.getItem('useScaledRatings')
      if (storedPreference !== null) {
        useScaledRatings.value = storedPreference === 'true'
      }
    } catch (e) {
      console.warn('Could not read rating calculation preference from localStorage:', e)
    }
  }

  // Initialize all settings
  function initSettings() {
    initDarkMode()
    initNotifications()
    initRatingCalculation()
  }

  // Watch for changes in Quasar's dark mode state and update the store
  // This is generally not needed if the store is the single source of truth
  // and all changes go through toggleDarkMode.
  // However, if Quasar's dark mode could be changed externally, this would be useful.
  // For this specific fix, we'll assume the store is the master.
  // if ($q && $q.dark) {
  //   watch(
  //     () => $q.dark.isActive,
  //     (newValue) => {
  //       if (isDarkModeActive.value !== newValue) {
  //         isDarkModeActive.value = newValue;
  //         // Optionally update localStorage here too if Quasar's state can change independently
  //         // localStorage.setItem('darkMode', newValue ? 'true' : 'false');
  //       }
  //     }
  //   );
  // }

  return {
    isDarkModeActive,
    toggleDarkMode,
    initDarkMode,
    notificationsEnabled,
    toggleNotifications,
    initNotifications,
    useScaledRatings,
    toggleRatingCalculation,
    setRatingCalculation,
    initRatingCalculation,
    initSettings
  }
})
