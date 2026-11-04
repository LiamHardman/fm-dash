import { describe, expect, it } from 'vitest'
import { selectInEffect } from './contracts.js'

const record = (fields) => ({ hasTerms: true, wage: 1000, ...fields })

describe('selectInEffect', () => {
  const clock = '2036-07-01'

  it("prefers the running contract at the player's own club", () => {
    const own = record({ clubUid: 1, start: '2034-07-01', end: '2038-06-30', wage: 50 })
    const parent = record({ clubUid: 2, start: '2035-07-01', end: '2039-06-30', wage: 90 })
    expect(selectInEffect([parent, own], 1, clock)).toBe(own)
  })

  it('falls back to the latest started contract', () => {
    const old = record({ clubUid: 3, start: '2030-07-01', end: '2033-06-30' })
    const recent = record({ clubUid: 4, start: '2035-01-01', end: '2037-06-30' })
    expect(selectInEffect([old, recent], undefined, clock)).toBe(recent)
  })

  it('ignores contracts that start after the clock or have no terms', () => {
    const future = record({ clubUid: 1, start: '2037-01-01', end: '2040-06-30' })
    const noTerms = record({ clubUid: 1, start: '2035-01-01', hasTerms: false })
    expect(selectInEffect([future, noTerms], 1, clock)).toBeUndefined()
  })
})
