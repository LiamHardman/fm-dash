// Manual-gated check against a real FM27 save, which can't be committed (≈210 MB).
// Run with:  FM27_SAVE=fm27save.fm npx vitest run src/utils/fmSave/realSave
// Thresholds are set from the user's fm27save.fm (build 27.0.0+2493426, clock 2036-07-01),
// loosened so another FM27 save of a similar shape also passes.

import { closeSync, fstatSync, openSync, readSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { parseSave } from './parse.js'

const savePath = process.env.FM27_SAVE

describe.skipIf(!savePath)('real FM27 save', () => {
  it('decodes players with names, attributes, clubs and wages', async () => {
    const fd = openSync(savePath, 'r')
    try {
      const source = {
        size: fstatSync(fd).size,
        read: async (offset, length) => {
          const buffer = new Uint8Array(length)
          readSync(fd, buffer, 0, length, offset)
          return buffer
        },
      }
      const { metadata, players, report } = await parseSave(source)
      const share = (predicate) => players.filter(predicate).length / players.length

      expect(metadata.build).toMatch(/^27\./)
      expect(report.players).toBeGreaterThan(10_000)
      expect(report.warnings).toEqual([])
      expect(share((p) => p.name)).toBeGreaterThan(0.98)
      expect(share((p) => p.age >= 14 && p.age <= 50)).toBeGreaterThan(0.98)
      expect(share((p) => p.nationName)).toBeGreaterThan(0.95)
      expect(share((p) => p.uniqueId)).toBeGreaterThan(0.85)
      expect(share((p) => p.clubName)).toBeGreaterThan(0.8)
      expect(share((p) => p.wage !== undefined)).toBeGreaterThan(0.5)
      for (const player of players.slice(0, 2000)) {
        for (const value of Object.values(player.attributes)) {
          expect(value).toBeGreaterThanOrEqual(1)
          expect(value).toBeLessThanOrEqual(20)
        }
      }
    } finally {
      closeSync(fd)
    }
  }, 120_000)
})
