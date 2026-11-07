// The whole save -> CSV pipeline the import worker runs, kept free of worker and DOM APIs
// so the manual real-save test exercises exactly what the browser does.

import { playersToCsv } from './exportCsv.js'
import { resolveManagedClub } from './managedClub.js'
import { parseSave } from './parse.js'

/**
 * Parse a save and build its FM-style CSV text plus the import summary.
 * @param {{size:number, read(offset:number, length:number):Promise<Uint8Array>}} source
 * @param {(progress: {stage: string, current?: number, total?: number}) => void} [onProgress]
 */
export async function exportSave(source, onProgress = () => {}) {
  const parsed = await parseSave(source, onProgress)
  onProgress({ stage: 'exporting' })
  const { csv, rows, skipped } = playersToCsv(parsed.players)
  const players = parsed.players
  const count = (predicate) => players.filter(predicate).length
  const managedClub = resolveManagedClub(parsed.metadata.managedClubUid, parsed.clubs, players)
  return {
    csv,
    parsed,
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
      withStats: count((p) => p.seasonStats && p.name),
      elapsedMs: parsed.report.elapsedMs,
      warnings: parsed.report.warnings,
      ...(managedClub ? { managedClub } : {}),
    },
  }
}
