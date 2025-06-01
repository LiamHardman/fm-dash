import { computed, ref } from 'vue'

export function useTableUtils(items, options = {}) {
  const { defaultSortField = 'id', defaultSortDirection = 'asc', defaultPageSize = 50 } = options

  // Table state
  const sortField = ref(defaultSortField)
  const sortDirection = ref(defaultSortDirection)
  const currentPage = ref(1)
  const pageSize = ref(defaultPageSize)
  const searchQuery = ref('')

  // Sorting utilities
  const setSortField = field => {
    if (sortField.value === field) {
      // Toggle direction if same field
      sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
    } else {
      // New field, default to ascending
      sortField.value = field
      sortDirection.value = 'asc'
    }
    currentPage.value = 1 // Reset to first page when sorting changes
  }

  // Generic sort function
  const sortItems = (itemsToSort, field, direction, getValue = null) => {
    return [...itemsToSort].sort((a, b) => {
      let vA = getValue ? getValue(a, field) : a[field]
      let vB = getValue ? getValue(b, field) : b[field]

      const aIsNull = vA === null || vA === undefined
      const bIsNull = vB === null || vB === undefined

      if (aIsNull && bIsNull) return 0
      if (aIsNull) return direction === 'asc' ? 1 : -1
      if (bIsNull) return direction === 'asc' ? -1 : 1

      if (typeof vA === 'number' && typeof vB === 'number') {
        return direction === 'asc' ? vA - vB : vB - vA
      }
      if (typeof vA === 'string' && typeof vB === 'string') {
        vA = vA.toLowerCase()
        vB = vB.toLowerCase()
        if (vA < vB) return direction === 'asc' ? -1 : 1
        if (vA > vB) return direction === 'asc' ? 1 : -1
        return 0
      }
      return 0
    })
  }

  // Filter items based on search query
  const filterItems = (itemsToFilter, query, searchFields = []) => {
    if (!query.trim()) return itemsToFilter

    const lowercaseQuery = query.toLowerCase().trim()

    return itemsToFilter.filter(item => {
      // If searchFields provided, only search those fields
      if (searchFields.length > 0) {
        return searchFields.some(field => {
          const value = item[field]
          return value?.toString().toLowerCase().includes(lowercaseQuery)
        })
      }

      // Otherwise search all string fields
      return Object.values(item).some(value => {
        return value?.toString().toLowerCase().includes(lowercaseQuery)
      })
    })
  }

  // Paginate items
  const paginateItems = (itemsToPaginate, page, size) => {
    const startIndex = (page - 1) * size
    const endIndex = startIndex + size
    return itemsToPaginate.slice(startIndex, endIndex)
  }

  // Calculate pagination info
  const paginationInfo = computed(() => {
    if (!items.value) return { total: 0, pages: 0, start: 0, end: 0 }

    const total = items.value.length
    const pages = Math.ceil(total / pageSize.value)
    const start = total === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1
    const end = Math.min(currentPage.value * pageSize.value, total)

    return { total, pages, start, end }
  })

  // Navigate pagination
  const goToPage = page => {
    const maxPage = paginationInfo.value.pages
    currentPage.value = Math.max(1, Math.min(page, maxPage))
  }

  const nextPage = () => goToPage(currentPage.value + 1)
  const prevPage = () => goToPage(currentPage.value - 1)
  const firstPage = () => goToPage(1)
  const lastPage = () => goToPage(paginationInfo.value.pages)

  // Reset all filters and pagination
  const resetTable = () => {
    sortField.value = defaultSortField
    sortDirection.value = defaultSortDirection
    currentPage.value = 1
    searchQuery.value = ''
  }

  return {
    sortField,
    sortDirection,
    currentPage,
    pageSize,
    searchQuery,

    paginationInfo,

    setSortField,
    sortItems,
    filterItems,
    paginateItems,
    goToPage,
    nextPage,
    prevPage,
    firstPage,
    lastPage,
    resetTable
  }
}
