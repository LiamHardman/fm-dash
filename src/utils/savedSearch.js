const labels = {
  name: 'name',
  club: 'club',
  role: 'role',
  nationality: 'nationality',
  division: 'division',
  mediaHandling: 'media handling',
  personality: 'personality',
  minOverall: 'overall',
  minSalary: 'minimum wage',
}

const display = (value) => (Array.isArray(value) ? value.join(', ') : String(value))

export function describeSavedSearch(filters) {
  const result = []
  if (filters.name) result.push(`Name contains “${filters.name}”`)
  if (filters.club) result.push(`Club is ${filters.club}`)
  if (filters.position?.length) result.push(`Position: ${filters.position.join(', ')}`)
  if (filters.role) result.push(`Role: ${filters.role}`)
  if (filters.nationality?.length) result.push(`Nationality: ${filters.nationality.join(', ')}`)
  if (filters.division?.length) result.push(`Division: ${filters.division.join(', ')}`)
  if (filters.ageRange && (filters.ageRange.min > 15 || filters.ageRange.max < 100))
    result.push(`Age ${filters.ageRange.min}–${filters.ageRange.max}`)
  if (
    filters.transferValueRangeLocal &&
    (filters.transferValueRangeLocal.min > 0 || filters.transferValueRangeLocal.max < 100000000)
  )
    result.push('Transfer value is within the selected range')
  if (filters.minSalary > 0 || (filters.maxSalary > 0 && filters.maxSalary < 100000000))
    result.push('Wage is within the selected range')
  for (const [key, value] of Object.entries(filters)) {
    if (key.startsWith('min') && Number(value) > 0 && !['minSalary', 'minOverall'].includes(key))
      result.push(`${labels[key] || key.slice(3)} ≥ ${value}`)
  }
  if (Number(filters.minOverall) > 0) result.push(`Overall ≥ ${filters.minOverall}`)
  if (filters.mediaHandling?.length)
    result.push(`Media handling: ${display(filters.mediaHandling)}`)
  if (filters.personality?.length) result.push(`Personality: ${display(filters.personality)}`)
  return result
}

export function explainPlayerMatch(filters, player) {
  const reasons = describeSavedSearch(filters)
  if (!reasons.length)
    return ['No restrictive criteria; this player is included in the broad result set.']
  const age = Number(player.age ?? player.Age)
  if (filters.ageRange && Number.isFinite(age)) reasons.push(`Current age is ${age}`)
  const overall = Number(player.Overall ?? player.overall)
  if (Number(filters.minOverall) > 0 && Number.isFinite(overall))
    reasons.push(`Current overall is ${overall}`)
  return reasons.slice(0, 6)
}

export function encodeSearchRecipe(filters) {
  const bytes = new TextEncoder().encode(JSON.stringify({ version: 1, filters }))
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary).replaceAll('+', '-').replaceAll('/', '_').replaceAll('=', '')
}

export function decodeSearchRecipe(recipe) {
  try {
    const binary = atob(
      recipe.replaceAll('-', '+').replaceAll('_', '/') + '='.repeat((4 - (recipe.length % 4)) % 4)
    )
    const bytes = Uint8Array.from(binary, (character) => character.charCodeAt(0))
    const value = JSON.parse(new TextDecoder().decode(bytes))
    return value.version === 1 && value.filters && typeof value.filters === 'object'
      ? value.filters
      : null
  } catch (_error) {
    return null
  }
}
