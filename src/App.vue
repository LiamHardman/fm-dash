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
import { defineComponent, onMounted } from "vue";
import { useUiStore } from "./stores/uiStore";

export default defineComponent({
    name: "App",
    setup() {
        const uiStore = useUiStore();
        
        onMounted(() => {
            uiStore.initDarkMode();
        });

        return {
            isDarkModeActive: uiStore.isDarkModeActive,
            toggleDarkMode: uiStore.toggleDarkMode,
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
