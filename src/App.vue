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
                    <router-link to="/" class="app-title-link">FMDB</router-link>
                </q-toolbar-title>
                
                <div class="nav-links">
                    <q-btn
                        flat
                        label="Upload"
                        @click="$router.push('/upload')"
                        class="nav-btn"
                    />
                    <q-btn
                        flat
                        label="Team View"
                        @click="$router.push('/team-view')"
                        class="nav-btn"
                    />
                    <q-btn
                        flat
                        label="Docs"
                        @click="$router.push('/docs')"
                        class="nav-btn"
                    />
                </div>
                
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
                    &copy; {{ new Date().getFullYear() }} FMDB. All
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

.app-title-link {
    text-decoration: none;
    color: inherit;
    font-weight: 700;
    font-size: 1.5rem;
    
    &:hover {
        opacity: 0.8;
    }
}

.nav-links {
    display: flex;
    gap: 0.5rem;
}

.nav-btn {
    color: white;
    font-weight: 500;
    
    &:hover {
        background: rgba(255, 255, 255, 0.1);
    }
}

// Responsive adjustments for header
@media (max-width: $breakpoint-xs-max) {
    .header-title {
        font-size: 1rem;
    }
    
    .app-title-link {
        font-size: 1.2rem;
    }
    
    .nav-links {
        display: none; // Hide nav links on mobile to save space
    }
    
    .q-toolbar__title {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    .dark-mode-toggle {
        margin-right: 0;
    }
}
</style>
