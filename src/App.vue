<template>
    <q-layout view="hHh lpR fFf">
        <q-header
            flat
            bordered
            :class="
                $q.dark.isActive
                    ? 'bg-dark text-white'
                    : 'bg-primary text-white'
            "
            class="app-header"
        >
            <q-toolbar>
                <q-avatar class="q-mr-sm">
                    <img
                        src="https://cdn.quasar.dev/logo-v2/svg/logo-mono-white.svg"
                        alt="App Logo"
                    />
                </q-avatar>
                <q-toolbar-title class="header-title">
                    FM Player Parser
                </q-toolbar-title>
                <q-space />
                <q-toggle
                    v-model="isDarkModeActive"
                    checked-icon="dark_mode"
                    unchecked-icon="light_mode"
                    :color="$q.dark.isActive ? 'yellow' : 'grey-4'"
                    size="lg"
                    @update:model-value="toggleDarkMode"
                    class="dark-mode-toggle"
                >
                    <q-tooltip
                        anchor="center left"
                        self="center right"
                        :offset="[10, 10]"
                        :class="
                            $q.dark.isActive
                                ? 'bg-grey-7 text-white'
                                : 'bg-grey-3 text-dark'
                        "
                    >
                        Toggle {{ $q.dark.isActive ? "Light" : "Dark" }} Mode
                    </q-tooltip>
                </q-toggle>
            </q-toolbar>
        </q-header>

        <q-page-container>
            <router-view />
        </q-page-container>

        <q-footer
            elevated
            :class="
                $q.dark.isActive
                    ? 'bg-grey-9 text-grey-5'
                    : 'bg-grey-2 text-grey-7'
            "
            class="app-footer"
        >
            <q-toolbar>
                <q-toolbar-title class="text-caption text-center">
                    &copy; {{ new Date().getFullYear() }} FM Player Parser. All
                    rights reserved.
                </q-toolbar-title>
            </q-toolbar>
        </q-footer>
    </q-layout>
</template>

<script>
import { defineComponent, ref, watch, onMounted } from "vue";
import { useQuasar } from "quasar";

export default defineComponent({
    name: "App",
    setup() {
        const $q = useQuasar();
        const isDarkModeActive = ref(false); // Local state for the toggle

        // Function to toggle dark mode
        const toggleDarkMode = (value) => {
            $q.dark.set(value);
            try {
                localStorage.setItem("darkMode", value ? "true" : "false");
            } catch (e) {
                console.warn(
                    "Could not save dark mode preference to localStorage:",
                    e,
                );
            }
        };

        // On component mount, check localStorage for saved preference
        onMounted(() => {
            let darkModePreference = false; // Default to light mode
            try {
                const storedPreference = localStorage.getItem("darkMode");
                if (storedPreference !== null) {
                    darkModePreference = storedPreference === "true";
                } else {
                    // If no preference, check system preference
                    if (
                        window.matchMedia &&
                        window.matchMedia("(prefers-color-scheme: dark)")
                            .matches
                    ) {
                        darkModePreference = true;
                    }
                }
            } catch (e) {
                console.warn(
                    "Could not read dark mode preference from localStorage:",
                    e,
                );
                // Fallback to system preference if localStorage fails
                if (
                    window.matchMedia &&
                    window.matchMedia("(prefers-color-scheme: dark)").matches
                ) {
                    darkModePreference = true;
                }
            }
            isDarkModeActive.value = darkModePreference;
            $q.dark.set(darkModePreference); // Apply it
        });

        // Watch for changes in Quasar's dark mode state (e.g., if changed by other means)
        // and update the toggle's local state.
        watch(
            () => $q.dark.isActive,
            (newValue) => {
                if (isDarkModeActive.value !== newValue) {
                    isDarkModeActive.value = newValue;
                }
            },
        );

        return {
            isDarkModeActive,
            toggleDarkMode,
        };
    },
});
</script>

<style lang="scss" scoped>
.app-header {
    padding: 0 8px; // Add some padding to the header
}

.header-title {
    font-weight: 600; // Make title a bit bolder
    font-size: 1.25rem; // Slightly larger title
}

.dark-mode-toggle {
    // Styles for the toggle if needed, e.g., margins
    margin-right: 8px;
}

.app-footer {
    border-top: 1px solid rgba(0, 0, 0, 0.12);
    .body--dark & {
        border-top: 1px solid rgba(255, 255, 255, 0.12);
    }
}

// Responsive adjustments for header
@media (max-width: $breakpoint-xs-max) {
    .header-title {
        font-size: 1rem; // Smaller title on mobile
    }
    .q-toolbar__title {
        // Allow title to shrink if needed when toggle is present
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    .dark-mode-toggle {
        margin-right: 0;
    }
}
</style>
