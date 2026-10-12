import { defineStore } from 'pinia'
import { Notify } from 'quasar'
import { ref, toRaw } from 'vue'
import savedSearchService from '../services/savedSearchService'

const makeID = () =>
  globalThis.crypto?.randomUUID?.() || `search-${Date.now()}-${Math.random().toString(36).slice(2)}`
const cloneFilters = (filters) => JSON.parse(JSON.stringify(toRaw(filters)))

export const useSavedSearchStore = defineStore('savedSearch', () => {
  const document = ref(savedSearchService.defaultDocument())
  const loading = ref(false)
  const loaded = ref(false)
  let loadPromise = null

  const normalize = (value) =>
    value?.version === 1 && Array.isArray(value.searches)
      ? value
      : savedSearchService.defaultDocument()
  const load = async () => {
    if (loaded.value) return
    if (loadPromise) return loadPromise
    loading.value = true
    loadPromise = (async () => {
      try {
        document.value = normalize(await savedSearchService.load())
        loaded.value = true
      } finally {
        loading.value = false
        loadPromise = null
      }
    })()
    return loadPromise
  }
  const saveDocument = async () => {
    try {
      await savedSearchService.save(document.value)
      return true
    } catch (error) {
      console.error('Saved-search save failed:', error)
      Notify.create({
        type: 'negative',
        message: savedSearchService.state.localDurable
          ? 'Could not save searches. A local draft is retained.'
          : 'Could not save searches. Keep this page open to retain your edits.',
      })
      return false
    }
  }
  const create = async (name, filters) => {
    if (!loaded.value) await load()
    const search = { id: makeID(), name: name.trim(), filters: cloneFilters(filters) }
    document.value = { ...document.value, searches: [...document.value.searches, search] }
    await saveDocument()
    return search
  }
  const update = async (search, filters) => {
    document.value = {
      ...document.value,
      searches: document.value.searches.map((candidate) =>
        candidate === search ? { ...candidate, filters: cloneFilters(filters) } : candidate
      ),
    }
    await saveDocument()
  }
  const rename = async (search, name) => {
    document.value = {
      ...document.value,
      searches: document.value.searches.map((candidate) =>
        candidate === search ? { ...candidate, name: name.trim() } : candidate
      ),
    }
    await saveDocument()
  }
  const remove = async (search) => {
    document.value.searches = document.value.searches.filter((candidate) => candidate !== search)
    await saveDocument()
  }
  const getById = (id) => document.value.searches.find((search) => search.id === id)
  return {
    document,
    loading,
    loaded,
    load,
    create,
    update,
    rename,
    remove,
    getById,
    sync: savedSearchService.state,
    retry: () => savedSearchService.retry(),
    resolveConflict: async (choice) => {
      document.value = normalize(await savedSearchService.resolveConflict(choice))
    },
    replaceDocument: async (value) => {
      if (!loaded.value) await load()
      document.value = normalize(value)
      return saveDocument()
    },
  }
})
