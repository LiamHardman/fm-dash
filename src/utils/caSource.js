// Where a Dataset's Current Ability comes from. Save imports carry the game's true CA
// (`saveCa`, which the backend copies into `ca`); CSV and HTML imports only have the
// fm21-cas estimate. Derived, never stored: any player with saveCa > 0 makes it 'save'.

/** @returns {'save' | 'estimated'} */
export function caSourceOf(players) {
  return Array.isArray(players) && players.some((player) => player?.saveCa > 0)
    ? 'save'
    : 'estimated'
}

/** The label for CA columns, axes and filters: "CA" for true CA, "Est. CA" otherwise. */
export function caLabel(source) {
  return source === 'save' ? 'CA' : 'Est. CA'
}
