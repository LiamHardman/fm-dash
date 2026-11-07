// FM's "Inf" (information) flags for a save player, as FM's own CSV export prints them.
// Decoded by FM-Dash from the save (see .scratch/fm27-save-extras/issues/05 and 18):
//   Inj  the player is in injury_manager's current windows with a return on/after the game date
//   Loa  listed for loan: in-effect contract tail +39, bit 1
//   Wnt  tail +39 bit 0 AND tail +40 == 5. Strict reading, matches the one in-game sample;
//        its meaning is unconfirmed.
// Several flags are joined with ", " in this order (FM's multi-flag format is unverified).

const LOAN_LISTED_BIT = 0b10
const WANTED_BIT = 0b01
const WANTED_CODE = 5

export function infFlags(player) {
  const flags = []
  if (player.injured) flags.push('Inj')
  const bits = player.listingFlags ?? 0
  if (bits & LOAN_LISTED_BIT) flags.push('Loa')
  if (bits & WANTED_BIT && player.listingCode === WANTED_CODE) flags.push('Wnt')
  return flags
}

/** The Inf cell: the flags joined with ", ", or blank. */
export function infCell(player) {
  return infFlags(player).join(', ')
}
