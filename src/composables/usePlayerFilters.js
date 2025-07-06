import { computed, reactive, ref, watch } from 'vue'

export function usePlayerFilters(props, emit) {
  // Filter state
  const filters = reactive({
    overall: { min: 1, max: 99 },
    potential: { min: 1, max: 99 },
    transferValue: { min: 0, max: 100000000 },
    salary: { min: 0, max: 1000000 },
    age: { min: 15, max: 50 },
    clubs: [],
    nationalities: [],
    mediaHandlings: [],
    personalities: [],
    positions: [],
    selectedAttributes: [],
    attributeThresholds: {}
  })

  // UI state
  const showAdvancedFilters = ref(false)
  const selectedFilters = ref(new Set())

  // Computed ranges based on props
  const transferValueRange = computed(() => props.transferValueRange)
  const salaryRange = computed(() => props.salaryRange)

  // Initialize filter ranges from props
  const initializeRanges = () => {
    if (props.transferValueRange) {
      filters.transferValue.min = props.transferValueRange.min
      filters.transferValue.max = props.transferValueRange.max
    }
    if (props.salaryRange) {
      filters.salary.min = props.salaryRange.min
      filters.salary.max = props.salaryRange.max
    }
    if (props.initialDatasetRange) {
      filters.overall.min = props.initialDatasetRange.min || 1
      filters.overall.max = props.initialDatasetRange.max || 99
    }
  }

  // Clear all filters
  const clearAllFilters = () => {
    filters.overall = { min: 1, max: 99 }
    filters.potential = { min: 1, max: 99 }
    filters.age = { min: 15, max: 50 }
    filters.clubs = []
    filters.nationalities = []
    filters.mediaHandlings = []
    filters.personalities = []
    filters.positions = []
    filters.selectedAttributes = []
    filters.attributeThresholds = {}

    initializeRanges()
    selectedFilters.value.clear()
    emitFilterChanged()
  }

  // Emit filter changes
  const emitFilterChanged = () => {
    emit('filter-changed', { ...filters })
  }

  // Watch for filter changes and emit
  watch(filters, emitFilterChanged, { deep: true })

  // Toggle filter selection
  const toggleFilter = filterName => {
    if (selectedFilters.value.has(filterName)) {
      selectedFilters.value.delete(filterName)
    } else {
      selectedFilters.value.add(filterName)
    }
  }

  // Check if filter is active
  const isFilterActive = filterName => {
    switch (filterName) {
      case 'overall':
        return filters.overall.min > 1 || filters.overall.max < 99
      case 'potential':
        return filters.potential.min > 1 || filters.potential.max < 99
      case 'age':
        return filters.age.min > 15 || filters.age.max < 50
      case 'transferValue':
        return (
          filters.transferValue.min > props.transferValueRange?.min ||
          filters.transferValue.max < props.transferValueRange?.max
        )
      case 'salary':
        return (
          filters.salary.min > props.salaryRange?.min || filters.salary.max < props.salaryRange?.max
        )
      case 'clubs':
        return filters.clubs.length > 0
      case 'nationalities':
        return filters.nationalities.length > 0
      case 'positions':
        return filters.positions.length > 0
      case 'attributes':
        return filters.selectedAttributes.length > 0
      default:
        return false
    }
  }

  // Get active filter count
  const activeFilterCount = computed(() => {
    const filterNames = [
      'overall',
      'potential',
      'age',
      'transferValue',
      'salary',
      'clubs',
      'nationalities',
      'positions',
      'attributes'
    ]
    return filterNames.filter(isFilterActive).length
  })

  return {
    filters,
    showAdvancedFilters,
    selectedFilters,
    transferValueRange,
    salaryRange,
    activeFilterCount,
    initializeRanges,
    clearAllFilters,
    toggleFilter,
    isFilterActive,
    emitFilterChanged
  }
}
