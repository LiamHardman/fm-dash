// src/stores/uiStore.js
import { defineStore } from "pinia";
import { ref, watch } from "vue";
import { useQuasar } from "quasar";

export const useUiStore = defineStore("ui", () => {
  const $q = useQuasar();
  // Initialize with a default value, initDarkMode will override this.
  // Defaulting to true here to align with the user's intent for dark mode by default.
  const isDarkModeActive = ref(true);

  // Function to toggle dark mode
  function toggleDarkMode() {
    // Directly toggle the current state
    isDarkModeActive.value = !isDarkModeActive.value;
    $q.dark.set(isDarkModeActive.value);
    try {
      localStorage.setItem(
        "darkMode",
        isDarkModeActive.value ? "true" : "false",
      );
    } catch (e) {
      console.warn("Could not save dark mode preference to localStorage:", e);
    }
  }

  // Initialize dark mode from localStorage or system preference
  function initDarkMode() {
    let darkModePreference = true; // Default to dark mode as intended by the user
    try {
      const storedPreference = localStorage.getItem("darkMode");
      if (storedPreference !== null) {
        darkModePreference = storedPreference === "true";
      } else {
        // If no preference in localStorage, check system preference
        // but still prioritize the app's default if system preference isn't explicitly dark.
        // The user wants the app to default to dark, so we only switch if system prefers light AND no local storage.
        // However, the initial ref(true) and this function's default to true handles the "default dark" requirement.
        // If localStorage has a value, that takes precedence.
        // If no localStorage, it defaults to true (dark).
        // We can still check system preference for the very first load if desired,
        // but the user's explicit "defaults to dark mode all the time (intended)" means we should honor that.
      }
    } catch (e) {
      console.warn("Could not read dark mode preference from localStorage:", e);
      // Fallback to true (dark mode) if localStorage fails
      darkModePreference = true;
    }
    // Set the ref and Quasar's dark mode
    isDarkModeActive.value = darkModePreference;
    $q.dark.set(darkModePreference);
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
  };
});
