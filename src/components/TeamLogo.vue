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
      v-else-if="isLoadingLogo"
      class="team-logo-placeholder loading"
      :class="placeholderClass"
      :style="placeholderStyle"
    >
      <q-spinner 
        :size="iconSize" 
        :color="iconColor"
      />
    </div>
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
import { useQuasar } from 'quasar'
import { computed, defineComponent, ref, watch } from 'vue'
import { useUiStore } from '../stores/uiStore'

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
    const isLoadingLogo = ref(false)

    // Load logo when team name changes
    const loadLogo = async teamName => {
      if (!teamName || teamName === '-') {
        logoUrl.value = null
        isLoadingLogo.value = false
        return
      }

      // Initialize backend composable if needed
      if (!globalTeamLogos) {
        const module = await import('../composables/useTeamLogosBackend')
        globalTeamLogos = module.useTeamLogosBackend({
          cacheTimeout: 3600000, // 1 hour cache
          similarityThreshold: 0.7
        })
      }

      // Load with loading state
      isLoadingLogo.value = true
      logoLoadError.value = false

      try {
        const url = await globalTeamLogos.getTeamLogoUrl(teamName)
        logoUrl.value = url
        logoLoadError.value = !url // Set error if no URL found
      } catch (_error) {
        logoUrl.value = null
        logoLoadError.value = true
      } finally {
        isLoadingLogo.value = false
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
      newTeamName => {
        logoLoadError.value = false
        loadLogo(newTeamName)
      },
      { immediate: true }
    )

    return {
      logoUrl,
      logoLoadError,
      isLoadingLogo,
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

.team-logo-placeholder.loading {
  background-color: rgba(128, 128, 128, 0.05);
}

.body--dark .team-logo-placeholder {
  background-color: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

.body--dark .team-logo-placeholder.loading {
  background-color: rgba(255, 255, 255, 0.02);
}

/* Optional hover effects */
.team-logo:hover,
.team-logo-placeholder:hover:not(.loading) {
  transform: scale(1.05);
}
</style> 