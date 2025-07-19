import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

/**
 * Composable for handling percentile loading states and retry logic
 * @param {Object} player - The player object
 * @param {String} datasetId - The dataset ID
 * @param {String} selectedComparisonGroup - The selected comparison group
 * @returns {Object} - Contains loading states and retry functionality
 */
export function usePercentileRetry(player, datasetId, selectedComparisonGroup, divisionFilter) {
  // Loading states
  const isLoadingPercentiles = ref(false)
  const percentilesRetryCount = ref(0)
  const maxRetries = 5
  const retryDelays = [1000, 2000, 3000, 5000, 8000] // Progressive backoff

  // Retry timeout reference
  let retryTimeout = null

  // Stop retry process - define this before resetState
  const stopPercentileRetry = () => {
    if (retryTimeout) {
      clearTimeout(retryTimeout)
      retryTimeout = null
    }
  }

  // Reset state when player changes
  const resetState = () => {
    isLoadingPercentiles.value = false
    percentilesRetryCount.value = 0
    stopPercentileRetry()
  }

  // Watch for player changes and reset state
  watch(
    () => player?.value,
    (newPlayer, oldPlayer) => {
      // Reset state when player changes
      if (newPlayer !== oldPlayer) {
        resetState()
      }
    },
    { immediate: true }
  )

  // Watch for player UID changes specifically to ensure proper reset
  watch(
    () => player?.value?.uid || player?.value?.UID,
    (newUID, oldUID) => {
      // Reset state when player UID changes
      if (newUID !== oldUID) {
        resetState()
      }
    },
    { immediate: true }
  )

  // Check if percentiles are available and valid
  const hasValidPercentiles = computed(() => {
    if (!player?.value?.performancePercentiles) return false

    const percentiles =
      player.value.performancePercentiles[selectedComparisonGroup?.value || 'Global']
    if (!percentiles) return false

    // Check if there are any non-negative percentile values (actual data)
    return Object.values(percentiles).some(
      value => value !== null && value !== undefined && value >= 0
    )
  })

  // Check if percentiles appear to be empty/not calculated yet
  const percentilesNeedRetry = computed(() => {
    if (!player?.value?.performancePercentiles) return true

    const percentiles =
      player.value.performancePercentiles[selectedComparisonGroup?.value || 'Global']
    if (!percentiles) return true

    // If all percentiles are -1, 0, null, or undefined, they likely aren't calculated yet
    const values = Object.values(percentiles)
    if (values.length === 0) return true

    const validValues = values.filter(value => value !== null && value !== undefined && value >= 0)

    // If less than 30% of percentiles are valid, consider retry needed
    return validValues.length < values.length * 0.3
  })

  // Show loading state if percentiles need retry or are currently being retried
  const showLoadingState = computed(() => {
    return (
      isLoadingPercentiles.value ||
      (percentilesNeedRetry.value && percentilesRetryCount.value < maxRetries)
    )
  })

  // Retry percentile calculation
  const retryPercentiles = async () => {
    if (!datasetId?.value || !player?.value || percentilesRetryCount.value >= maxRetries) {
      return false
    }

    isLoadingPercentiles.value = true

    try {
      // Handle the 'same' division filter by converting it to the player's actual division
      let effectiveDivision = divisionFilter?.value || 'all'
      if (effectiveDivision === 'same') {
        const targetDivision = player.value?.division
        if (targetDivision) {
          effectiveDivision = targetDivision
        } else {
          // If no target division is available, fall back to 'all'
          effectiveDivision = 'all'
        }
      }

      const requestPayload = {
        playerUID: player.value.uid?.toString() || player.value.UID?.toString(),
        compareDivision: effectiveDivision,
        comparePosition: selectedComparisonGroup?.value || 'Global'
      }

      const response = await fetch(`/api/player-percentiles/${datasetId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestPayload)
      })

      if (response.ok) {
        const updatedPercentiles = await response.json()

        // Update the player's percentiles
        if (player.value.performancePercentiles) {
          Object.assign(player.value.performancePercentiles, updatedPercentiles)
        } else {
          player.value.performancePercentiles = updatedPercentiles
        }

        return true
      } else {
        return false
      }
    } catch (error) {
      return false
    } finally {
      isLoadingPercentiles.value = false
      percentilesRetryCount.value++
    }
  }

  // Start automatic retry process if percentiles are needed
  const startPercentileRetry = () => {
    if (percentilesNeedRetry.value && percentilesRetryCount.value < maxRetries) {
      const delay = retryDelays[Math.min(percentilesRetryCount.value, retryDelays.length - 1)]

      retryTimeout = setTimeout(async () => {
        const success = await retryPercentiles()

        // If retry failed and we haven't reached max retries, schedule another retry
        if (!success && percentilesRetryCount.value < maxRetries) {
          startPercentileRetry()
        }
      }, delay)
    }
  }

  // Manual retry function for user-triggered retries
  const manualRetry = async () => {
    stopPercentileRetry() // Stop any automatic retries
    percentilesRetryCount.value = 0 // Reset retry count
    return await retryPercentiles()
  }

  // Start retry process on mount if needed
  onMounted(() => {
    // Small delay to allow component to initialize
    setTimeout(() => {
      if (percentilesNeedRetry.value) {
        startPercentileRetry()
      }
    }, 100)
  })

  // Cleanup on unmount
  onUnmounted(() => {
    stopPercentileRetry()
  })

  return {
    // States
    isLoadingPercentiles,
    hasValidPercentiles,
    percentilesNeedRetry,
    showLoadingState,
    percentilesRetryCount,
    maxRetries,

    // Methods
    retryPercentiles,
    manualRetry,
    startPercentileRetry,
    stopPercentileRetry,
    resetState
  }
}
