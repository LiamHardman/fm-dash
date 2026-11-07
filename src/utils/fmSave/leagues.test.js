import { describe, expect, it } from 'vitest'
import { divisionName, isLeagueName, resolveClubLeagues, scanMatchLogLeagues } from './leagues.js'

const le16 = (v) => [v & 0xff, (v >> 8) & 0xff]
const le32 = (v) => [v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, v >>> 24]
const FULL_TAIL = [0x00, 0x01, 0x03, 0x02]
const COMPACT_TAIL = [0x0c, 0x00, 0x01]

/** One match-log entry; the day word's high bits are flags and must be masked off. */
const entry = ({ day, year, team, competition, compact = false, flags = 0 }) => [
  1,
  ...le16(day | flags),
  ...le16(year),
  ...le32(team),
  ...le32(competition),
  ...(compact ? COMPACT_TAIL : FULL_TAIL),
  0xee, // separator so entries don't run together
]
const gameDbOf = (...entries) => new Uint8Array([0, 0, ...entries.flat(), 0, 0, 0, 0, 0])

// Clubs: uid -> {nationId, teams}. Team ids are uid * 10 + slot.
function clubIndex(specs) {
  const clubs = Object.entries(specs).map(([uid, spec]) => ({
    uid: Number(uid),
    nationId: spec.nationId ?? 1,
    teams: [Number(uid) * 10, Number(uid) * 10 + 1],
  }))
  const teamToClub = new Map()
  for (const club of clubs)
    club.teams.forEach((team, slot) => teamToClub.set(team, [club.uid, slot]))
  return {
    clubs,
    teamToClub,
    clubByUid: new Map(clubs.map((club) => [club.uid, club])),
    recordSpans: new Map(),
  }
}

const COMPETITIONS = [
  { id: 100, databaseId: 102423, name: 'Brazilian National First Division' },
  { id: 101, databaseId: 107191, name: 'Brazilian National Second Division' },
  { id: 102, databaseId: 9001, name: 'Swiss Super League' },
  { id: 103, databaseId: 9002, name: 'Swiss Super Cup' },
  { id: 104, databaseId: 9003, name: 'Copa do Brasil' },
  { id: 105, databaseId: 9004, name: 'Unmapped Regional League' },
  { id: 106, databaseId: 9005, name: undefined },
  { id: 107, databaseId: 11, name: 'Premier League' },
]
const byId = new Map(COMPETITIONS.map((c) => [c.id, c]))
const twice = (spec) => [entry(spec), entry({ ...spec, day: spec.day + 7 })]

describe('isLeagueName', () => {
  it('keeps "Super" leagues and drops cups, friendlies and youth competitions', () => {
    expect(isLeagueName('Swiss Super League')).toBe(true)
    expect(isLeagueName('3F Superliga')).toBe(true)
    expect(isLeagueName('Swiss Super Cup')).toBe(false)
    expect(isLeagueName('Supercopa de España')).toBe(false)
    expect(isLeagueName('UEFA Champions League')).toBe(false)
    expect(isLeagueName('Club Friendly')).toBe(false)
    expect(isLeagueName('Premier League U21')).toBe(false)
    expect(isLeagueName('Promotion Play-off')).toBe(false)
    expect(isLeagueName(undefined)).toBe(false)
  })
})

describe('scanMatchLogLeagues', () => {
  it('reads full and compact entries and masks the day flags', () => {
    const gameDb = gameDbOf(
      entry({ day: 40, year: 2039, team: 10, competition: 100, flags: 0x8000 }),
      entry({ day: 47, year: 2039, team: 10, competition: 100, compact: true })
    )
    const leagues = scanMatchLogLeagues(gameDb, clubIndex({ 1: {} }), byId, 2039)
    expect(leagues.get(1)?.id).toBe(100)
  })

  it('bounds the day to 1-366 and the year to game year -2..+1', () => {
    const club = clubIndex({ 1: {} })
    const at = (spec) => scanMatchLogLeagues(gameDbOf(...twice(spec)), club, byId, 2039).get(1)
    expect(at({ day: 0, year: 2039, team: 10, competition: 100 })).toBeUndefined()
    expect(at({ day: 360, year: 2039, team: 10, competition: 100 })).toBeUndefined() // 367 on the second
    expect(at({ day: 10, year: 2037, team: 10, competition: 100 })?.id).toBe(100)
    expect(at({ day: 10, year: 2036, team: 10, competition: 100 })).toBeUndefined()
    expect(at({ day: 10, year: 2040, team: 10, competition: 100 })?.id).toBe(100)
    expect(at({ day: 10, year: 2041, team: 10, competition: 100 })).toBeUndefined()
  })

  it('ignores reserve teams', () => {
    const gameDb = gameDbOf(...twice({ day: 10, year: 2039, team: 11, competition: 101 }))
    expect(scanMatchLogLeagues(gameDb, clubIndex({ 1: {} }), byId, 2039).size).toBe(0)
  })

  it('needs two hits for the club, or three clubs of its nation', () => {
    const single = (team) => entry({ day: 10, year: 2039, team, competition: 105 })
    const clubs = clubIndex({ 1: {}, 2: {}, 3: { nationId: 2 }, 4: {} })
    // Clubs 1 and 2 share nation 1 with one hit each: noise.
    let leagues = scanMatchLogLeagues(
      gameDbOf(single(10), single(20), single(30)),
      clubs,
      byId,
      2039
    )
    expect(leagues.size).toBe(0)
    // A third club of nation 1 makes the competition count for all three.
    leagues = scanMatchLogLeagues(
      gameDbOf(single(10), single(20), single(30), single(40)),
      clubs,
      byId,
      2039
    )
    expect([...leagues.keys()].sort()).toEqual([1, 2, 4])
  })

  it('takes the latest counted entry', () => {
    const gameDb = gameDbOf(
      ...twice({ day: 300, year: 2038, team: 10, competition: 101 }),
      ...twice({ day: 20, year: 2039, team: 10, competition: 100 }), // promoted
      entry({ day: 200, year: 2039, team: 10, competition: 105 }) // one-off noise, later
    )
    expect(scanMatchLogLeagues(gameDb, clubIndex({ 1: {} }), byId, 2039).get(1)?.id).toBe(100)
  })

  it('keeps "Super" leagues and drops cups and unnamed competitions', () => {
    const clubs = clubIndex({ 1: {}, 2: {}, 3: {} })
    const leagues = scanMatchLogLeagues(
      gameDbOf(
        ...twice({ day: 10, year: 2039, team: 10, competition: 102 }),
        ...twice({ day: 90, year: 2039, team: 10, competition: 103 }),
        ...twice({ day: 90, year: 2039, team: 20, competition: 104 }),
        ...twice({ day: 90, year: 2039, team: 30, competition: 106 })
      ),
      clubs,
      byId,
      2039
    )
    expect(leagues.get(1)?.name).toBe('Swiss Super League')
    expect(leagues.has(2)).toBe(false)
    expect(leagues.has(3)).toBe(false)
  })
})

describe('divisionName', () => {
  it('prefers the verified short name, then the bundled long name, else blank', () => {
    expect(divisionName(byId.get(100))).toBe('Série A')
    expect(divisionName(byId.get(105))).toBe('Unmapped Regional League')
    expect(divisionName(byId.get(106))).toBeUndefined()
    expect(divisionName(undefined)).toBeUndefined()
  })
})

describe('resolveClubLeagues', () => {
  const nations = { databaseIdFor: (nationId) => ({ 1: 769, 7: 765 })[nationId] }

  it('uses the match log first and falls back to the three-window top-league rule', () => {
    const log = [...twice({ day: 10, year: 2039, team: 10, competition: 101 })].flat()
    const gameDb = new Uint8Array(4000)
    gameDb.set(log, 0)
    // Club 2 (England) has no log entries; its record holds Premier League (107) in all
    // three season windows (350-500, 600-750, 1100-1250).
    const recordStart = 2000
    for (const offset of [400, 650, 1150]) gameDb.set(le32(107), recordStart + offset)
    const clubs = clubIndex({ 1: {}, 2: { nationId: 7 }, 3: {} })
    clubs.recordSpans.set(2, { recordStart, recordEnd: recordStart + 1500 })
    clubs.recordSpans.set(3, { recordStart: 100, recordEnd: 1600 })

    const report = resolveClubLeagues(gameDb, clubs, nations, '2039-08-08', COMPETITIONS)
    const [one, two, three] = clubs.clubs
    expect(one.leagueName).toBe('Série B')
    expect(one.leagueCompetition.databaseId).toBe(107191)
    expect(two.leagueName).toBe('Premier League')
    expect(three.leagueName).toBeUndefined()
    expect(report.divisionClubs).toBe(2)
  })
})
