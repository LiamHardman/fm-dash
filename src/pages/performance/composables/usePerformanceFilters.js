import { debounce } from 'quasar'
import { computed, ref, watch } from 'vue'
import { getNumericValue, getPlayerDivision } from '../../../utils/playerUtils'

export function usePerformanceFilters(allPlayersData) {
  const sliderValue = ref(0)
  const selectedMinutes = ref(0)
  const overallSliderValue = ref(0)
  const selectedOverall = ref(0)
  const selectedDivisions = ref([])
  const divisionOptions = ref([])
  const selectedPositions = ref([])
  const positionOptions = ref([])

  const maxMinutes = computed(() =>
    Math.max(2000, ...allPlayersData.value.map((p) => getNumericValue(p.attributes?.Mins) || 0))
  )
  const maxOverall = computed(() =>
    Math.max(100, ...allPlayersData.value.map((p) => p.Overall || 0))
  )

  const availableDivisions = computed(() => {
    return [
      ...new Set(allPlayersData.value.map((p) => getPlayerDivision(p)).filter(Boolean)),
    ].sort()
  })

  const availablePositions = computed(() => {
    return [
      { label: '--- Position Groups ---', value: null, disable: true },
      { label: 'Goalkeeper', value: 'Goalkeeper', group: true },
      { label: 'Defender', value: 'Defender', group: true },
      { label: 'Midfielder', value: 'Midfielder', group: true },
      { label: 'Forward', value: 'Forward', group: true },
      { label: '--- Specific Positions ---', value: null, disable: true },
      { label: 'GK - Goalkeeper', value: 'GK' },
      { label: 'DC - Centre Back', value: 'DC' },
      { label: 'DR - Right Back', value: 'DR' },
      { label: 'DL - Left Back', value: 'DL' },
      { label: 'WBR - Right Wing-Back', value: 'WBR' },
      { label: 'WBL - Left Wing-Back', value: 'WBL' },
      { label: 'DM - Defensive Midfielder', value: 'DM' },
      { label: 'MC - Centre Midfielder', value: 'MC' },
      { label: 'MR - Right Midfielder', value: 'MR' },
      { label: 'ML - Left Midfielder', value: 'ML' },
      { label: 'AMC - Attacking Mid (Centre)', value: 'AMC' },
      { label: 'AMR - Attacking Mid (Right)', value: 'AMR' },
      { label: 'AML - Attacking Mid (Left)', value: 'AML' },
      { label: 'ST - Striker', value: 'ST' },
    ]
  })

  const selectedDivisionsDisplayText = computed(() => {
    const count = selectedDivisions.value.length
    const allAvailableDivisions = [
      ...new Set(allPlayersData.value.map((p) => getPlayerDivision(p)).filter(Boolean)),
    ]

    if (count === 0 || count === allAvailableDivisions.length) return 'All divisions'
    if (count <= 5) return selectedDivisions.value.join(', ')
    return `${count} divisions selected`
  })

  const selectedPositionsDisplayText = computed(() => {
    const count = selectedPositions.value.length
    if (count === 0) return ''
    if (count <= 3) {
      const labels = selectedPositions.value.map((value) => {
        const option = availablePositions.value.find((opt) => opt.value === value)
        return option ? option.label : value
      })
      return labels.join(', ')
    }
    return `${count} positions selected`
  })

  // Filter functions
  const filterDivisionsFn = (val, update) => {
    update(() => {
      const needle = String(val || '').toLowerCase()
      divisionOptions.value = availableDivisions.value.filter((v) =>
        v.toLowerCase().includes(needle)
      )
    })
  }

  const filterPositionsFn = (val, update) => {
    update(() => {
      const needle = String(val || '').toLowerCase()
      positionOptions.value = availablePositions.value.filter((opt) =>
        (opt.label || '').toLowerCase().includes(needle)
      )
    })
  }

  const selectAllDivisions = () => {
    selectedDivisions.value = [...availableDivisions.value]
  }
  const clearAllDivisions = () => {
    selectedDivisions.value = []
  }
  const selectAllPositions = () => {
    selectedPositions.value = availablePositions.value
      .filter((option) => option.value !== null && !option.disable)
      .map((option) => option.value)
  }
  const clearAllPositions = () => {
    selectedPositions.value = []
  }

  // Debounced syncing of sliders to selected thresholds
  watch(
    sliderValue,
    debounce((newVal) => {
      selectedMinutes.value = newVal
    }, 300)
  )
  watch(
    overallSliderValue,
    debounce((newVal) => {
      selectedOverall.value = newVal
    }, 300)
  )

  // Keep selected divisions within available
  watch(
    availableDivisions,
    (divs) => {
      selectedDivisions.value = selectedDivisions.value.filter((d) => divs.includes(d))
      divisionOptions.value = divs
    },
    { immediate: true }
  )

  watch(
    availablePositions,
    (positions) => {
      positionOptions.value = positions
    },
    { immediate: true }
  )

  return {
    sliderValue,
    selectedMinutes,
    overallSliderValue,
    selectedOverall,
    selectedDivisions,
    divisionOptions,
    selectedPositions,
    positionOptions,
    availableDivisions,
    availablePositions,
    selectedDivisionsDisplayText,
    selectedPositionsDisplayText,
    maxMinutes,
    maxOverall,
    filterDivisionsFn,
    filterPositionsFn,
    selectAllDivisions,
    clearAllDivisions,
    selectAllPositions,
    clearAllPositions,
  }
}
