// Client-side filter predicate for Progression's latest-snapshot player view.
// Mirrors DatasetPage.vue's filteredPlayers predicate (name/club/position/nationality/
// division/age/value/salary/min-stat fields) since that page's filtering is itself entirely
// client-side. Deliberately narrower than PlayerFilters.vue's full field set: the Role filter
// and the per-raw-FM-attribute minimum grid are dropped — Role depends on the single-dataset
// playerStore's global role list (not meaningful across N progression snapshots), and the raw
// attribute grid reads Player.Attributes, which the backend never serializes (`json:"-"`) so
// it's already a no-op filter on DatasetPage.vue too.
import { deriveShortPositionsFromPositionString } from '../utils/playerUtils'

export function filterLatestSnapshotPlayers(players, filters) {
  if (!players || players.length === 0) return []

  return players.filter((player) => {
    if (filters.name && !player.name.toLowerCase().includes(filters.name.toLowerCase())) {
      return false
    }

    if (filters.club && player.club !== filters.club) {
      return false
    }

    if (filters.position && filters.position.length > 0) {
      const playerPositions = player.shortPositions || []
      const playerPosition = player.position || ''
      const parsedPlayerPositions = deriveShortPositionsFromPositionString(playerPosition)
      const hasMatchingPosition = filters.position.some(
        (selectedPos) =>
          playerPositions.includes(selectedPos) ||
          parsedPlayerPositions.includes(selectedPos) ||
          playerPosition.includes(selectedPos)
      )
      if (!hasMatchingPosition) return false
    }

    if (
      filters.nationality &&
      filters.nationality.length > 0 &&
      !filters.nationality.includes(player.nationality)
    ) {
      return false
    }

    if (
      filters.division &&
      filters.division.length > 0 &&
      !filters.division.includes(player.division)
    ) {
      return false
    }

    if (player.age !== undefined && player.age !== null) {
      const playerAge = Number.parseInt(player.age, 10)
      const filterMinAge = filters.ageRange?.min || 0
      const filterMaxAge = filters.ageRange?.max || 0
      if (
        !Number.isNaN(playerAge) &&
        filterMaxAge > 0 &&
        (playerAge < filterMinAge || playerAge > filterMaxAge)
      ) {
        return false
      }
    }

    const filterMinValue = filters.transferValueRange?.min || 0
    const filterMaxValue = filters.transferValueRange?.max || 0
    if (filterMinValue > 0 && (player.transferValueAmount || 0) < filterMinValue) {
      return false
    }
    if (filterMaxValue > 0 && (player.transferValueAmount || 0) > filterMaxValue) {
      return false
    }

    const filterMaxSalary = filters.maxSalary || 0
    if (filterMaxSalary > 0 && (player.wageAmount || 0) > filterMaxSalary) {
      return false
    }

    if (filters.minOverall > 0 && (player.overall || 0) < filters.minOverall) {
      return false
    }
    if (filters.minCA > 0 && (player.ca || 0) < filters.minCA) {
      return false
    }
    if (filters.minMBR > 0 && (player.mbr || 0) < filters.minMBR) {
      return false
    }

    return true
  })
}

export function uniqueValues(players, field) {
  return Array.from(new Set(players.map((p) => p[field]).filter(Boolean))).sort()
}
