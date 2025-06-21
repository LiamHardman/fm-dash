<template>
  <div v-if="showLogos" class="team-logo-container" :class="containerClass">
    <img
      v-if="logoUrl && !logoLoadError"
      :src="logoUrl"
      :alt="`${teamName} logo`"
      :width="size"
      :height="size"
      class="team-logo"
      :class="logoClass"
      @error="handleLogoError"
      @load="handleLogoLoad"
    />
    <div
      v-else
      class="team-logo-placeholder"
      :class="placeholderClass"
      :style="placeholderStyle"
    >
      <q-icon 
        :name="fallbackIcon" 
        :size="iconSize" 
        :color="iconColor"
      />
    </div>
  </div>
</template>

<script>
import { defineComponent, ref, computed, watch } from 'vue'
import { useQuasar } from 'quasar'
import { useUiStore } from '../stores/uiStore'

// Simple global cache - maps team name directly to logo URL
const logoCache = new Map()

// Single global composable instance to avoid recreation
let globalTeamLogos = null

export default defineComponent({
  name: 'TeamLogo',
  props: {
    teamName: {
      type: String,
      required: true
    },
    size: {
      type: [String, Number],
      default: 32
    },
    rounded: {
      type: Boolean,
      default: false
    },
    fallbackIcon: {
      type: String,
      default: 'shield'
    },
    logoClass: {
      type: String,
      default: ''
    },
    placeholderClass: {
      type: String,
      default: ''
    },
    containerClass: {
      type: String,
      default: ''
    }
  },
  setup(props) {
    const $q = useQuasar()
    const uiStore = useUiStore()
    
    const logoLoadError = ref(false)
    const logoUrl = ref(null)

    // Load logo when team name changes
    const loadLogo = async (teamName) => {
      if (!teamName || teamName === '-') {
        logoUrl.value = null
        return
      }

      // Check cache first - instant if cached
      if (logoCache.has(teamName)) {
        logoUrl.value = logoCache.get(teamName)
        return
      }

      // Initialize composable if needed
      if (!globalTeamLogos) {
        const module = await import('../composables/useTeamLogos')
        globalTeamLogos = module.useTeamLogos()
      }

      // Load and cache
      try {
        const url = await globalTeamLogos.getTeamLogoUrlAsync(teamName)
        logoCache.set(teamName, url)
        logoUrl.value = url
      } catch (error) {
        logoCache.set(teamName, null)
        logoUrl.value = null
      }
    }

    const iconSize = computed(() => {
      const sizeNum = typeof props.size === 'string' ? parseInt(props.size) : props.size
      return `${Math.max(12, sizeNum * 0.6)}px`
    })

    const iconColor = computed(() => {
      return $q.dark.isActive ? 'grey-5' : 'grey-7'
    })

    const placeholderStyle = computed(() => {
      const sizeStr = typeof props.size === 'string' ? props.size : `${props.size}px`
      return {
        width: sizeStr,
        height: sizeStr,
        borderRadius: props.rounded ? '50%' : '4px'
      }
    })

    const handleLogoError = () => {
      logoLoadError.value = true
    }

    const handleLogoLoad = () => {
      logoLoadError.value = false
    }

    // Watch for team name changes
    watch(
      () => props.teamName,
      (newTeamName) => {
        logoLoadError.value = false
        loadLogo(newTeamName)
      },
      { immediate: true }
    )

    return {
      logoUrl,
      logoLoadError,
      iconSize,
      iconColor,
      placeholderStyle,
      handleLogoError,
      handleLogoLoad,
      showLogos: computed(() => uiStore.showLogos)
    }
  }
})
</script>

<style scoped>
.team-logo-container {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.team-logo {
  object-fit: contain;
  transition: all 0.2s ease;
}

.team-logo-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(128, 128, 128, 0.1);
  border: 1px solid rgba(128, 128, 128, 0.2);
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.body--dark .team-logo-placeholder {
  background-color: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

/* Optional hover effects */
.team-logo:hover,
.team-logo-placeholder:hover {
  transform: scale(1.05);
}
</style> 