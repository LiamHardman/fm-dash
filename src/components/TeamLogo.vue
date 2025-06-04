<template>
  <div class="team-logo-container" :class="containerClass">
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
import { useTeamLogos } from '../composables/useTeamLogos'

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
    const { getTeamLogoUrl } = useTeamLogos()
    
    const logoLoadError = ref(false)

    const logoUrl = computed(() => {
      return getTeamLogoUrl(props.teamName)
    })

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

    // Reset error state when team changes
    watch(
      () => props.teamName,
      () => {
        logoLoadError.value = false
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
      handleLogoLoad
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