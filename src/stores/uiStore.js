// src/stores/uiStore.js
import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import { useQuasar } from 'quasar';

export const useUiStore = defineStore('ui', () => {
  const $q = useQuasar();
  const isDarkModeActive = ref(false);

  // Function to toggle dark mode
  function toggleDarkMode(value) {
    isDarkModeActive.value = value;
    $q.dark.set(value);
    try {
      localStorage.setItem('darkMode', value ? 'true' : 'false');
    } catch (e) {
      console.warn('Could not save dark mode preference to localStorage:', e);
    }
  }

  // Initialize dark mode from localStorage or system preference
  function initDarkMode() {
    let darkModePreference = false; // Default to light mode
    try {
      const storedPreference = localStorage.getItem('darkMode');
      if (storedPreference !== null) {
        darkModePreference = storedPreference === 'true';
      } else {
        // If no preference, check system preference
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
          darkModePreference = true;
        }
      }
    } catch (e) {
      console.warn('Could not read dark mode preference from localStorage:', e);
      // Fallback to system preference if localStorage fails
      if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        darkModePreference = true;
      }
    }
    isDarkModeActive.value = darkModePreference;
    $q.dark.set(darkModePreference); // Apply it
  }

  // Watch for changes in Quasar's dark mode state and update the store
  if ($q && $q.dark) {
    watch(
      () => $q.dark.isActive,
      (newValue) => {
        if (isDarkModeActive.value !== newValue) {
          isDarkModeActive.value = newValue;
        }
      }
    );
  }

  return {
    isDarkModeActive,
    toggleDarkMode,
    initDarkMode
  };
});