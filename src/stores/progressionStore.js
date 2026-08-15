import { defineStore } from 'pinia'
import { ref } from 'vue'
import progressionService from '../services/progressionService'

// State machine: idle -> uploading -> analyzed | ambiguous-order -> analyzed (after confirm)
export const useProgressionStore = defineStore('progression', () => {
  const slots = ref([]) // [{ id, filename, datasetId, status: 'uploading'|'parsed'|'error' }]
  const status = ref('idle') // idle | analyzing | analyzed | ambiguous-order | error
  const order = ref([])
  const ambiguousDatasetIds = ref([])
  const players = ref([])
  const emptyIntersection = ref(false)
  const currencySymbol = ref('')
  const errorMessage = ref('')
  // Sort state lives here (not page-local) so every analyze/confirmOrder call automatically
  // applies the currently-selected sort instead of silently falling back to the backend's
  // name-order default until the user manually re-triggers a sort change.
  const sortField = ref('Overall')
  const sortDir = ref('desc')

  function addSlot(slot) {
    slots.value.push(slot)
  }

  function updateSlot(id, patch) {
    const slot = slots.value.find((s) => s.id === id)
    if (slot) Object.assign(slot, patch)
  }

  function removeSlot(id) {
    slots.value = slots.value.filter((s) => s.id !== id)
  }

  function reset() {
    slots.value = []
    status.value = 'idle'
    order.value = []
    ambiguousDatasetIds.value = []
    players.value = []
    emptyIntersection.value = false
    currencySymbol.value = ''
    errorMessage.value = ''
    sortField.value = 'Overall'
    sortDir.value = 'desc'
  }

  const parsedDatasetIds = () =>
    slots.value.filter((s) => s.status === 'parsed' && s.datasetId).map((s) => s.datasetId)

  async function analyze({ order: explicitOrder } = {}) {
    const datasetIds = parsedDatasetIds()
    if (datasetIds.length < 2) return

    status.value = 'analyzing'
    errorMessage.value = ''
    try {
      const result = await progressionService.analyze(datasetIds, {
        order: explicitOrder,
        sortField: sortField.value,
        sortDir: sortDir.value,
      })

      if (result.orderAmbiguous) {
        ambiguousDatasetIds.value = result.ambiguousDatasetIds || []
        status.value = 'ambiguous-order'
        return
      }

      order.value = result.order || []
      players.value = result.players || []
      emptyIntersection.value = !!result.emptyIntersection
      currencySymbol.value = result.currencySymbol || ''
      status.value = 'analyzed'
    } catch (error) {
      errorMessage.value = error.message || 'Failed to analyze snapshots'
      status.value = 'error'
    }
  }

  async function confirmOrder(confirmedOrder) {
    await analyze({ order: confirmedOrder })
  }

  // Change the active sort and, if a result is already showing, re-run analyze with the
  // already-resolved order pinned so re-sorting never re-triggers order detection/ambiguity.
  async function setSort(field, dir) {
    sortField.value = field
    sortDir.value = dir
    if (status.value === 'analyzed') {
      await analyze({ order: order.value.length ? order.value : undefined })
    }
  }

  return {
    slots,
    status,
    order,
    ambiguousDatasetIds,
    players,
    emptyIntersection,
    currencySymbol,
    errorMessage,
    sortField,
    sortDir,
    addSlot,
    updateSlot,
    removeSlot,
    reset,
    analyze,
    confirmOrder,
    setSort,
  }
})
