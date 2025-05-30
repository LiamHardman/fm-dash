import { computed, reactive, ref } from 'vue'
import { usePlayerRatings } from './usePlayerRatings'

export function useAttributeFilters() {
  const { attributeFullNameMap, attributeGroups } = usePlayerRatings()

  // Attribute filter state
  const selectedAttributes = ref([])
  const attributeThresholds = reactive({})
  const showAttributeFilters = ref(false)

  // Available attributes for filtering
  const allAttributes = computed(() => {
    const attrs = []
    for (const group of Object.values(attributeGroups.value)) {
      for (const attr of group.attrs) {
        attrs.push({
          value: attr,
          label: attributeFullNameMap[attr] || attr,
          group: group.name
        })
      }
    }
    return attrs.sort((a, b) => a.label.localeCompare(b.label))
  })

  // Grouped attributes for better UI organization
  const groupedAttributes = computed(() => {
    const grouped = {}
    for (const [key, group] of Object.entries(attributeGroups.value)) {
      grouped[key] = {
        name: group.name,
        attributes: group.attrs.map(attr => ({
          value: attr,
          label: attributeFullNameMap[attr] || attr
        }))
      }
    }
    return grouped
  })

  // Add attribute filter
  const addAttributeFilter = attribute => {
    if (!selectedAttributes.value.includes(attribute)) {
      selectedAttributes.value.push(attribute)
      if (!attributeThresholds[attribute]) {
        attributeThresholds[attribute] = { min: 1, max: 20 }
      }
    }
  }

  // Remove attribute filter
  const removeAttributeFilter = attribute => {
    const index = selectedAttributes.value.indexOf(attribute)
    if (index > -1) {
      selectedAttributes.value.splice(index, 1)
      delete attributeThresholds[attribute]
    }
  }

  // Clear all attribute filters
  const clearAttributeFilters = () => {
    selectedAttributes.value = []
    for (const key of Object.keys(attributeThresholds)) {
      delete attributeThresholds[key]
    }
  }

  // Get attribute threshold
  const getAttributeThreshold = attribute => {
    return attributeThresholds[attribute] || { min: 1, max: 20 }
  }

  // Set attribute threshold
  const setAttributeThreshold = (attribute, threshold) => {
    attributeThresholds[attribute] = threshold
  }

  // Check if player matches attribute filters
  const playerMatchesAttributeFilters = player => {
    if (selectedAttributes.value.length === 0) return true

    return selectedAttributes.value.every(attr => {
      const threshold = attributeThresholds[attr]
      if (!threshold) return true

      const playerValue = player[attr]
      if (playerValue === undefined || playerValue === null) return false

      return playerValue >= threshold.min && playerValue <= threshold.max
    })
  }

  return {
    selectedAttributes,
    attributeThresholds,
    showAttributeFilters,
    allAttributes,
    groupedAttributes,
    addAttributeFilter,
    removeAttributeFilter,
    clearAttributeFilters,
    getAttributeThreshold,
    setAttributeThreshold,
    playerMatchesAttributeFilters
  }
}
