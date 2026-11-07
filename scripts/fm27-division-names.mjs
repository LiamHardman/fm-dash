#!/usr/bin/env node
// Propose Division short names for src/utils/fmSave/competition-short-names.js.
//
//   node --max-old-space-size=8192 scripts/fm27-division-names.mjs <save.fm> <export.csv> [--all]
//
// <export.csv> is an in-game FM27 CSV export from the same save, with Unique ID, Club and
// Division columns. Rows are joined to the save's players by Unique ID and grouped by the
// competition the save derives for their club; for each competition not yet mapped, the
// export's Division names are printed with their vote counts. --all also lists mapped ones.

import { closeSync, fstatSync, openSync, readFileSync, readSync } from 'node:fs'
import { COMPETITION_NAMES } from '../src/utils/fmSave/competition-names.js'
import { COMPETITION_SHORT_NAMES } from '../src/utils/fmSave/competition-short-names.js'
import { parseSave } from '../src/utils/fmSave/parse.js'
import { readCsv } from '../src/utils/fmSave/testing/csv.js'

const [savePath, csvPath, flag] = process.argv.slice(2)
if (!savePath || !csvPath) {
  console.error('usage: fm27-division-names.mjs <save.fm> <export.csv> [--all]')
  process.exit(2)
}

const fd = openSync(savePath, 'r')
let players
try {
  const source = {
    size: fstatSync(fd).size,
    read: async (offset, length) => {
      const buffer = new Uint8Array(length)
      readSync(fd, buffer, 0, length, offset)
      return buffer
    },
  }
  ;({ players } = await parseSave(source))
} finally {
  closeSync(fd)
}

const byUniqueId = new Map(players.filter((p) => p.uniqueId).map((p) => [String(p.uniqueId), p]))
const votes = new Map() // database id -> Map(division -> count)
let joined = 0
let withoutDivision = 0
for (const row of readCsv(readFileSync(csvPath, 'utf8'))) {
  const player = byUniqueId.get(row['Unique ID'])
  if (!player) continue
  joined++
  if (player.leagueDatabaseId === undefined || !row.Division) {
    withoutDivision++
    continue
  }
  if (!votes.has(player.leagueDatabaseId)) votes.set(player.leagueDatabaseId, new Map())
  const tally = votes.get(player.leagueDatabaseId)
  tally.set(row.Division, (tally.get(row.Division) ?? 0) + 1)
}

console.log(`// joined ${joined} rows by Unique ID; ${withoutDivision} without a derived or exported Division`)
for (const [databaseId, tally] of [...votes].sort((a, b) => a[0] - b[0])) {
  const mapped = COMPETITION_SHORT_NAMES[databaseId]
  if (mapped && flag !== '--all') continue
  const ranked = [...tally].sort((a, b) => b[1] - a[1])
  const total = ranked.reduce((sum, [, n]) => sum + n, 0)
  const [best, count] = ranked[0]
  const long = COMPETITION_NAMES[databaseId] ?? '(no bundled name)'
  const note = mapped ? ` (mapped: ${mapped})` : best === long ? ' (same as bundled)' : ''
  const others = ranked.slice(1).map(([name, n]) => `${name} ${n}`).join(', ')
  console.log(
    `  ${databaseId}: ${JSON.stringify(best)}, // ${long}; ${count}/${total} votes${others ? `; also ${others}` : ''}${note}`
  )
}
