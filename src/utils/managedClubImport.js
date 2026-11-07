import { apiUrl } from './apiEndpoint'

/**
 * After a save import, store the Managed club the save names, unless the Dataset already
 * has one (a stored value is never overwritten). Failures are logged and swallowed: the
 * Dataset page's "Set managed club" banner still covers a missing value.
 * @param {string} datasetId
 * @param {{club: string, division: string}|undefined} managedClub from the worker summary
 * @returns {Promise<boolean>} true when it was stored
 */
export async function setManagedClubIfAbsent(datasetId, managedClub, fetchFn = fetch) {
  if (!datasetId || !managedClub?.club) return false
  const url = apiUrl(`/api/managed-team/${encodeURIComponent(datasetId)}`)
  try {
    const existing = await fetchFn(url)
    if (existing.status !== 404) return false
    const res = await fetchFn(url, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ club: managedClub.club, division: managedClub.division }),
    })
    if (!res.ok) throw new Error(`managed team PUT returned ${res.status}`)
    return true
  } catch (error) {
    console.warn('[SaveImport] could not set the detected managed club', error)
    return false
  }
}
