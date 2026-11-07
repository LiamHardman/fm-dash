// The Managed club detected from a save: the club uid in save_game_summary, resolved
// through the save's club index to the names the export prints.

import { clubDisplayName } from './clubs.js'

/**
 * `{club, division}` for the import summary, or undefined unless the uid resolves to a
 * club that at least one exported (named) player plays for and that has a derived Division.
 * @param {number|undefined} uid metadata.managedClubUid
 * @param {import('./clubs.js').ClubIndex|undefined} clubs
 * @param {Array<{name?: string, clubName?: string}>} players
 */
export function resolveManagedClub(uid, clubs, players) {
  if (!uid || !clubs) return undefined
  const record = clubs.clubByUid.get(uid)
  const club = clubDisplayName(record)
  const division = record?.leagueName
  if (!club || !division) return undefined
  if (!players.some((player) => player.name && player.clubName === club)) return undefined
  return { club, division }
}
