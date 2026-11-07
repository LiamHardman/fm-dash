// Competition editor database id -> the Division name FM27 prints in its exports.
// Only names verified against an in-game FM27 export (fm27.csv) are listed; grow it with
// scripts/fm27-division-names.mjs. Leagues not listed keep their bundled long name from
// competition-names.js, so they won't equal FM27 CSV imports for those leagues.
export const COMPETITION_SHORT_NAMES = Object.freeze({
  11: 'Premier League',
  12: 'EFL Championship', // bundled: Sky Bet Championship
  16: "Ligue 1 McDonald's",
  60: 'Liga Portugal', // bundled: Portuguese Premier League
  67: 'LaLiga', // bundled: Spanish First Division
  102423: 'Série A', // bundled: Brazilian National First Division
  107191: 'Série B', // bundled: Brazilian National Second Division
  7880000: 'ADNOC Pro League',
  7920263: 'Saudi Pro League', // bundled: ROSHN Saudi League
})
