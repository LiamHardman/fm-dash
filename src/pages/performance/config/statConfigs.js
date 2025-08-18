// Centralized stat category configuration for Performance page

export const statCategories = {
  offensive: [
    { key: 'Gls/90', name: 'Goals per 90' },
    { key: 'xG/90', name: 'xG per 90' },
    { key: 'Shot/90', name: 'Shots per 90' },
    { key: 'Conv %', name: 'Conversion %' },
  ],
  passing: [
    { key: 'Asts/90', name: 'Assists per 90' },
    { key: 'xA/90', name: 'xA per 90' },
    { key: 'K Ps/90', name: 'Key Passes per 90' },
    { key: 'Pas %', name: 'Pass Completion %' },
  ],
  defensive: [
    { key: 'Tck/90', name: 'Tackles per 90' },
    { key: 'Int/90', name: 'Interceptions per 90' },
    { key: 'Hdrs W/90', name: 'Headers Won per 90' },
    { key: 'Pres C/90', name: 'Pressures Completed p90' },
  ],
  goalkeeping: [
    { key: 'Con/90', name: 'Goals Conceded p90' },
    { key: 'xGP/90', name: 'xG Prevented p90' },
    { key: 'Sv %', name: 'Save Percentage' },
    { key: 'Clean Sheets', name: 'Clean Sheets' },
  ],
}
