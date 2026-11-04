// Parses an FM27 .fm save off the main thread and turns it into an FM-style CSV that
// /api/upload ingests. The save itself never leaves the browser.
//
// in:  { file: File }
// out: { type: 'progress', stage, current?, total? }
//      { type: 'done', csv: Blob (gzipped CSV), summary }
//      { type: 'error', message, kind? }

import { blobSource } from '../utils/fmSave/container.js'
import { playersToCsv } from '../utils/fmSave/exportCsv.js'
import { parseSave } from '../utils/fmSave/parse.js'

self.onmessage = async ({ data }) => {
  try {
    const parsed = await parseSave(blobSource(data.file), (progress) =>
      self.postMessage({ type: 'progress', ...progress })
    )
    self.postMessage({ type: 'progress', stage: 'exporting' })
    const { csv, rows, skipped } = playersToCsv(parsed.players)
    const players = parsed.players
    const count = (predicate) => players.filter(predicate).length
    // A whole-database export (~25 MB) exceeds the default upload limit; gzip brings it
    // to roughly a third, and /api/upload inflates `.csv.gz` uploads.
    const gzipped = await new Response(
      new Blob([csv]).stream().pipeThrough(new CompressionStream('gzip'))
    ).blob()
    self.postMessage({
      type: 'done',
      csv: gzipped,
      summary: {
        saveName: parsed.metadata.saveName,
        build: parsed.metadata.build,
        gameDate: parsed.metadata.gameDate,
        players: rows,
        skipped,
        withClub: count((p) => p.clubName && p.name),
        withDivision: count((p) => p.leagueName && p.name),
        withWage: count((p) => p.wage !== undefined && p.name),
        withValue: count((p) => p.transferValue !== undefined && p.name),
        elapsedMs: parsed.report.elapsedMs,
        warnings: parsed.report.warnings,
      },
    })
  } catch (error) {
    self.postMessage({ type: 'error', message: error?.message ?? String(error), kind: error?.kind })
  }
}
