import { onScopeDispose, ref, toValue, watch } from 'vue'
import { apiUrl } from '../utils/apiEndpoint'

export const playerReferenceKey = (ref) => `${ref.datasetId}:${ref.playerUid}`

export function usePlayerReferences(references) {
  const resolved = ref(new Map())
  let controller
  async function refresh() {
    controller?.abort()
    const current = new AbortController()
    controller = current
    const unique = [
      ...new Map(toValue(references).map((ref) => [playerReferenceKey(ref), ref])).values(),
    ]
    resolved.value = new Map(
      unique.map((ref) => [playerReferenceKey(ref), { state: 'loading', player: null }])
    )
    if (!unique.length) return
    try {
      const response = await fetch(apiUrl('/api/profile/player-references'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ references: unique }),
        signal: current.signal,
      })
      if (!response.ok) throw new Error('Could not resolve Player references')
      const data = await response.json()
      if (current.signal.aborted) return
      resolved.value = new Map(data.items.map((item) => [playerReferenceKey(item.playerRef), item]))
    } catch (error) {
      if (current.signal.aborted) return
      resolved.value = new Map(
        unique.map((ref) => [playerReferenceKey(ref), { state: 'failed', player: null }])
      )
    }
  }
  watch(() => JSON.stringify(toValue(references)), refresh, { immediate: true })
  onScopeDispose(() => controller?.abort())
  return { resolved, refresh }
}
