import { reactive } from 'vue'
import { apiUrl } from '../utils/apiEndpoint'

const clone = (value) => JSON.parse(JSON.stringify(value))
let tabId
function recoveryKey(key) {
  if (!tabId) {
    try {
      tabId = sessionStorage.getItem('fm_dash_profile_tab') || crypto.randomUUID()
      sessionStorage.setItem('fm_dash_profile_tab', tabId)
    } catch {
      tabId = crypto.randomUUID()
    }
  }
  return `${key}:${tabId}`
}

// One client owns a document's revision and serialized writes. It never replaces
// the caller's live document with a response that predates more recent edits.
export function createProfileDocumentClient(path, storageKey, defaultDocument) {
  // Different tabs must not overwrite each other's unsaved recovery copies.
  const localKey = recoveryKey(storageKey)
  const state = reactive({ status: 'idle', error: '', conflict: null, localDurable: true })
  let revision = null
  let pending = null
  let running = null
  let loadPromise = null
  let generation = 0
  let draft = null

  function readCache() {
    try {
      const value = JSON.parse(
        localStorage.getItem(localKey) || localStorage.getItem(storageKey) || 'null'
      )
      if (!value) return null
      // Earlier versions stored an unversioned recovery copy. Keep it as dirty.
      return value.document ? value : { document: value, dirty: true, revision: null }
    } catch {
      return null
    }
  }
  function cache(document, dirty) {
    draft = clone(document)
    try {
      localStorage.setItem(localKey, JSON.stringify({ document: draft, dirty, revision }))
      state.localDurable = true
    } catch {
      state.localDurable = false
    }
  }
  async function readRemote() {
    const response = await fetch(apiUrl(path), {
      headers: { Accept: 'application/json' },
      cache: 'no-store',
    })
    if (!response.ok) throw new Error(`Unable to load saved data (${response.status})`)
    const document = await response.json()
    const etag = response.headers.get('ETag')
    if (!etag)
      throw new Error(
        'The server did not provide a document revision. Refresh after updating the server.'
      )
    return { document, revision: etag }
  }
  async function load() {
    if (loadPromise) return loadPromise
    loadPromise = (async () => {
      const local = readCache()
      revision = local?.revision ?? null
      try {
        const remote = await readRemote()
        if (local?.dirty) {
          draft = clone(local.document)
          state.status = local.revision === remote.revision ? 'pending' : 'conflict'
          state.conflict = state.status === 'conflict' ? remote : null
          return clone(draft)
        }
        revision = remote.revision
        cache(remote.document, false)
        state.status = 'saved'
        state.error = ''
        return remote.document
      } catch (error) {
        state.status = 'failed'
        state.error = error.message
        if (local) {
          draft = clone(local.document)
          return clone(draft)
        }
        throw error
      }
    })().finally(() => {
      loadPromise = null
    })
    return loadPromise
  }
  async function drain() {
    while (pending) {
      const entry = pending
      pending = null
      try {
        if (!revision || state.conflict)
          throw new Error('Review the server version before saving this local draft.')
        state.status = 'saving'
        const response = await fetch(apiUrl(path), {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json', 'If-Match': revision },
          body: JSON.stringify(entry.document),
        })
        if (response.status === 412) {
          state.conflict = await readRemote()
          state.status = 'conflict'
          throw new Error('Changes were saved elsewhere. Review both versions before continuing.')
        }
        if (!response.ok) throw new Error(`Unable to save changes (${response.status})`)
        revision = response.headers.get('ETag')
        if (!revision) throw new Error('The server did not confirm a document revision.')
        // The latest local edits, not this request's snapshot, are the recovery copy.
        cache(draft, entry.generation !== generation)
        state.status = entry.generation === generation ? 'saved' : 'pending'
        state.error = ''
      } catch (error) {
        state.error = error.message
        state.status = state.conflict ? 'conflict' : 'failed'
        cache(draft, true)
        pending = null
        throw error
      }
    }
  }
  function save(document) {
    generation++
    cache(document, true)
    pending = { document: clone(document), generation }
    state.status = 'pending'
    if (!running)
      running = drain().finally(() => {
        running = null
      })
    return running.then(() => clone(draft))
  }
  async function retry() {
    const remote = await readRemote()
    if (remote.revision !== revision) {
      state.conflict = remote
      state.status = 'conflict'
      throw new Error('Review the server changes before retrying.')
    }
    return save(draft)
  }
  async function resolveConflict(choice) {
    const remote = await readRemote()
    if (choice === 'local' && state.conflict && remote.revision !== state.conflict.revision) {
      state.conflict = remote
      state.status = 'conflict'
      throw new Error('The server changed again. Review the updated version before replacing it.')
    }
    revision = remote.revision
    state.conflict = null
    if (choice === 'local') return save(draft)
    pending = null
    cache(remote.document, false)
    state.status = 'saved'
    state.error = ''
    return remote.document
  }
  return {
    defaultDocument,
    state,
    load,
    list: load,
    save,
    retry,
    resolveConflict,
    recoveryKey: localKey,
  }
}
