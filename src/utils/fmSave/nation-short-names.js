// Editor database id -> the short nation name FM27 itself prints in its exports (Based In,
// Nationality), where it differs from the bundled long name in nation-names.js.
// Source: the Based In column of an FM-Dash export of an FM27 CSV (155 other names already
// match the bundle). Nations not listed keep their long editor name.
export const NATION_SHORT_NAMES = Object.freeze({
  17: 'Eq. Guinea',
  20: 'The Gambia',
  53: 'DR Congo', // the bundle calls both 49 (Congo-Brazzaville) and 53 "Congo"
  143: 'U.A.E.',
  387: 'Suriname',
  390: 'U.S.A.',
  759: 'Bosnia & Herzegovina',
  763: 'Czechia',
  785: 'N.Ireland',
  799: 'Türkiye',
  917508: 'St Martin',
})
