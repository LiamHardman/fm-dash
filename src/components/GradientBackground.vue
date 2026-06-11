<template>
  <div 
    class="gradient-background" 
    :style="gradientStyle"
    :class="{ 'gradient-background--loading': isLoading }"
  >
    <slot />
  </div>
</template>

<script>
import { computed, defineComponent, ref, watch } from 'vue'

export default defineComponent({
  name: 'GradientBackground',
  props: {
    imageUrl: {
      type: String,
      default: null,
    },
    fallbackGradient: {
      type: String,
      default: 'linear-gradient(135deg, #1a237e 0%, #283593 50%, #3949ab 100%)',
    },
    intensity: {
      type: Number,
      default: 0.3,
      validator: (value) => value >= 0 && value <= 1,
    },
  },
  setup(props) {
    const isLoading = ref(false)
    const dominantColors = ref([])

    const gradientStyle = computed(() => {
      if (dominantColors.value.length > 0) {
        const colors = dominantColors.value
        const stops = []

        // Create gradient stops with the extracted colors
        colors.forEach((color, index) => {
          const percentage = (index / (colors.length - 1)) * 100
          stops.push(`${color} ${percentage}%`)
        })

        return {
          background: `linear-gradient(135deg, ${stops.join(', ')})`,
          position: 'relative',
          overflow: 'hidden',
        }
      }

      return {
        background: props.fallbackGradient,
        position: 'relative',
        overflow: 'hidden',
      }
    })

    const extractColorsFromImage = async (imageUrl) => {
      if (!imageUrl) return

      isLoading.value = true

      try {
        const img = new Image()
        img.crossOrigin = 'anonymous'

        await new Promise((resolve, reject) => {
          img.onload = resolve
          img.onerror = () => {
            console.warn('Failed to load image for gradient background:', imageUrl)
            reject(new Error('Image load failed'))
          }
          img.src = imageUrl
        })

        // Create canvas to analyze image
        const canvas = document.createElement('canvas')
        const ctx = canvas.getContext('2d')

        // Set canvas size (smaller for performance)
        canvas.width = 100
        canvas.height = 100

        // Draw image to canvas
        ctx.drawImage(img, 0, 0, canvas.width, canvas.height)

        // Get image data
        const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
        const data = imageData.data

        // Sample colors from the image
        const colors = []
        const sampleStep = Math.max(1, Math.floor(data.length / 4 / 20)) // Sample 20 colors

        for (let i = 0; i < data.length; i += sampleStep * 4) {
          const r = data[i]
          const g = data[i + 1]
          const b = data[i + 2]
          const a = data[i + 3]

          // Skip transparent pixels and very light/dark pixels
          if (a > 128 && !(r > 240 && g > 240 && b > 240) && !(r < 20 && g < 20 && b < 20)) {
            colors.push({ r, g, b })
          }
        }

        // If we don't have enough colors, try sampling more
        if (colors.length < 5) {
          for (let i = 0; i < data.length; i += 4) {
            const r = data[i]
            const g = data[i + 1]
            const b = data[i + 2]
            const a = data[i + 3]

            if (a > 128) {
              colors.push({ r, g, b })
            }
          }
        }

        // Simple color extraction - take first 3 distinct colors
        const distinctColors = []
        const seen = new Set()

        for (const color of colors) {
          const key = `${color.r},${color.g},${color.b}`
          if (!seen.has(key) && distinctColors.length < 3) {
            seen.add(key)
            distinctColors.push(color)
          }
        }

        dominantColors.value = distinctColors.map(
          (color) => `rgba(${color.r}, ${color.g}, ${color.b}, ${props.intensity})`
        )
      } catch (error) {
        console.warn('Failed to extract colors from image:', error)
        dominantColors.value = []
      } finally {
        isLoading.value = false
      }
    }

    // Watch for image URL changes
    watch(
      () => props.imageUrl,
      (newUrl) => {
        if (newUrl) {
          extractColorsFromImage(newUrl)
        } else {
          dominantColors.value = []
        }
      },
      { immediate: true }
    )

    return {
      isLoading,
      gradientStyle,
    }
  },
})
</script>

<style scoped>
.gradient-background {
  position: relative;
  transition: all 0.3s ease;
}

.gradient-background--loading {
  opacity: 0.8;
}

.gradient-background::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, transparent 100%);
  pointer-events: none;
  z-index: 1;
}

.gradient-background > * {
  position: relative;
  z-index: 2;
}
</style> 