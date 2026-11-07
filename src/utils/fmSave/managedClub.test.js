import { describe, expect, it } from 'vitest'
import { resolveManagedClub } from './managedClub.js'
import { readSummaryTail } from './metadata.js'

const le32 = (v) => [v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, v >>> 24]
const str = (s) => {
  const bytes = [...new TextEncoder().encode(s)]
  return [...le32(bytes.length), ...bytes]
}
const BUILD = '27.0.0+2493426'
// Same shape as the real summary: 8 bytes, save name, build, divisions, 8 bytes, manager,
// club short name, u32 club uid, then the clock (day-of-year word, year).
const summary = (uid) =>
  new Uint8Array([
    ...new Array(8).fill(0),
    ...str('My Save'),
    ...str(BUILD),
    ...le32(1),
    ...str('Vtora Liga'),
    ...new Array(8).fill(1),
    ...str('El Aych'),
    ...str('PFK Montana'),
    ...le32(uid),
    220,
    0,
    0xf7,
    0x07, // day 220 of 2039
  ])

describe('readSummaryTail', () => {
  it('reads the managed club uid and the clock', () => {
    expect(readSummaryTail(summary(353), BUILD)).toEqual({
      managedClubUid: 353,
      clock: '2039-08-08',
    })
  })

  it('gives no uid for 0 or the missing marker, and nothing for another build', () => {
    expect(readSummaryTail(summary(0), BUILD).managedClubUid).toBeUndefined()
    expect(readSummaryTail(summary(0xffffffff), BUILD).managedClubUid).toBeUndefined()
    expect(readSummaryTail(summary(353), '27.0.0+1')).toBeUndefined()
  })
})

describe('resolveManagedClub', () => {
  const montana = {
    uid: 353,
    name: 'Profesionalen Futbolen Klub Montana',
    shortName: 'PFK Montana',
    leagueName: 'Vtora Liga',
  }
  const clubs = {
    clubByUid: new Map([
      [353, montana],
      [9, { uid: 9, name: 'Nowhere FC' }],
    ]),
  }
  const players = [{ name: 'A Player', clubName: 'PFK Montana' }]

  it('resolves to the exported club name and its Division', () => {
    expect(resolveManagedClub(353, clubs, players)).toEqual({
      club: 'PFK Montana',
      division: 'Vtora Liga',
    })
  })

  it('skips a missing uid, an unknown club, a club without players or without a Division', () => {
    expect(resolveManagedClub(undefined, clubs, players)).toBeUndefined()
    expect(resolveManagedClub(353, undefined, players)).toBeUndefined()
    expect(resolveManagedClub(77, clubs, players)).toBeUndefined()
    expect(resolveManagedClub(353, clubs, [{ name: 'X', clubName: 'Elsewhere' }])).toBeUndefined()
    expect(resolveManagedClub(353, clubs, [{ clubName: 'PFK Montana' }])).toBeUndefined() // nameless
    expect(resolveManagedClub(9, clubs, [{ name: 'Y', clubName: 'Nowhere FC' }])).toBeUndefined()
  })
})
