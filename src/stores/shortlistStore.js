import { defineStore } from 'pinia'
import { Notify } from 'quasar'
import { computed, ref } from 'vue'
import shortlistService from '../services/shortlistService'

const makeID = () => {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID()
  return `shortlist-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

export const useShortlistStore = defineStore('shortlist', () => {
  const document = ref(shortlistService.defaultDocument())
  const activeListId = ref('default')
  const loading = ref(false)
  const loaded = ref(false)
  let loadPromise = null
  const activeList = computed(
    () =>
      document.value.lists.find((list) => list.id === activeListId.value) || document.value.lists[0]
  )

  const normalize = (value) => {
    if (!value || value.version !== 1 || !Array.isArray(value.lists) || !value.lists.length) {
      return shortlistService.defaultDocument()
    }
    return value
  }

  const load = async () => {
    if (loaded.value) return
    if (loadPromise) return loadPromise
    loading.value = true
    loadPromise = (async () => {
      try {
        document.value = normalize(await shortlistService.load())
        if (!document.value.lists.some((list) => list.id === activeListId.value)) {
          activeListId.value = document.value.lists[0].id
        }
        loaded.value = true
      } finally {
        loading.value = false
        loadPromise = null
      }
    })()
    return loadPromise
  }

  const save = async () => {
    try {
      document.value = normalize(await shortlistService.save(document.value))
      return true
    } catch (error) {
      console.error('Shortlist save failed:', error)
      Notify.create({
        type: 'negative',
        message: 'Could not save Shortlists. Changes are kept in this browser.',
      })
      return false
    }
  }

  const createList = async (name) => {
    const list = { id: makeID(), name: name.trim(), items: [] }
    document.value.lists.push(list)
    activeListId.value = list.id
    await save()
  }

  const renameActiveList = async (name) => {
    if (!activeList.value || !name.trim()) return
    activeList.value.name = name.trim()
    await save()
  }

  const deleteActiveList = async () => {
    if (!activeList.value || document.value.lists.length === 1) return false
    document.value.lists = document.value.lists.filter((list) => list.id !== activeListId.value)
    activeListId.value = document.value.lists[0].id
    await save()
    return true
  }

  const addPlayer = async (datasetId, player) => {
    if (!loaded.value) await load()
    const playerUid = Number(player?.uid ?? player?.UID)
    if (!activeList.value || !datasetId || !Number.isFinite(playerUid) || playerUid <= 0)
      return false
    const alreadyAdded = activeList.value.items.some(
      (item) =>
        item.playerRef.datasetId === datasetId && Number(item.playerRef.playerUid) === playerUid
    )
    if (alreadyAdded) return false
    activeList.value.items.push({
      playerRef: { datasetId, playerUid },
      status: 'watching',
      priority: 'medium',
      tags: [],
      notes: '',
      targetFee: 0,
      targetWage: 0,
    })
    await save()
    return true
  }

  const removeItem = async (item) => {
    if (!activeList.value) return
    activeList.value.items = activeList.value.items.filter((candidate) => candidate !== item)
    await save()
  }

  const updateItem = async (item, patch) => {
    Object.assign(item, patch)
    await save()
  }

  return {
    document,
    activeListId,
    activeList,
    loading,
    loaded,
    load,
    createList,
    renameActiveList,
    deleteActiveList,
    addPlayer,
    removeItem,
    updateItem,
  }
})
