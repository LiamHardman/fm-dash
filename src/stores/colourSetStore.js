import { defineStore } from 'pinia'
import { reactive } from 'vue'
import colourSetService from '../services/colourSetService'
import { activeColourBands, defaultBands, validateBands } from '../utils/colourSets'

const emptyDocument = () => ({ version: 1, activeSetId: '', sets: [] })

export const useColourSetStore = defineStore('colourSets', () => {
  const documents = reactive({ fm: emptyDocument(), fifa: emptyDocument() })
  const loaded = reactive({ fm: false, fifa: false })
  const busy = reactive({ fm: false, fifa: false })
  const errors = reactive({ fm: '', fifa: '' })
  const pending = {}

  function accept(scale, document) {
    if (
      document?.version !== 1 ||
      !Array.isArray(document.sets) ||
      document.sets.some((set) => !set.id || !set.name || validateBands(set.bands, scale)) ||
      (document.activeSetId !== '' && !document.sets.some((set) => set.id === document.activeSetId))
    )
      throw new Error('The backend returned invalid colour sets.')
    documents[scale] = document
    const active = document.sets.find((set) => set.id === document.activeSetId)
    activeColourBands[scale] = active
      ? active.bands.map((band) => ({ ...band }))
      : defaultBands(scale)
  }

  async function load(scale) {
    if (loaded[scale]) return
    if (pending[scale]) return pending[scale]
    busy[scale] = true
    errors[scale] = ''
    pending[scale] = (async () => {
      try {
        accept(scale, await colourSetService.load(scale))
        loaded[scale] = true
      } catch (error) {
        errors[scale] = error.message
      } finally {
        busy[scale] = false
        pending[scale] = null
      }
    })()
    return pending[scale]
  }

  async function save(scale, document) {
    if (!loaded[scale] || busy[scale]) return false
    busy[scale] = true
    errors[scale] = ''
    try {
      accept(scale, await colourSetService.save(scale, document))
      return true
    } catch (error) {
      errors[scale] = `${error.message} Your changes have not been saved. Please retry.`
      return false
    } finally {
      busy[scale] = false
    }
  }

  return { documents, loaded, busy, errors, load, save }
})
