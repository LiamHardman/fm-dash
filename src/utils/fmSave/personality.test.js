import { describe, expect, it } from 'vitest'
import { deriveMediaHandling, derivePersonality, personalityLabels } from './personality.js'

// A neutral player: every rule just misses, so the label is Balanced / Level-headed.
const NEUTRAL = {
  ambition: 10,
  controversy: 10,
  loyalty: 12,
  pressure: 10,
  professionalism: 12,
  sportsmanship: 12,
  temperament: 12,
}
const inputs = ({ determination = 10, leadership = 10, age = 25, male = true, ...bytes } = {}) => ({
  personality: { ...NEUTRAL, ...bytes },
  determination,
  leadership,
  age,
  male,
})
const personality = (overrides) => derivePersonality(inputs(overrides))
const media = (overrides) => deriveMediaHandling(inputs(overrides))

describe('derivePersonality', () => {
  it('falls back to Balanced', () => {
    expect(personality()).toBe('Balanced')
  })

  it.each([
    // [label, at the threshold, just past it]
    [
      'Model Citizen',
      {
        sportsmanship: 15,
        loyalty: 15,
        professionalism: 15,
        determination: 14,
        ambition: 12,
        temperament: 15,
        pressure: 14,
      },
      { pressure: 13 },
    ],
    [
      'Perfectionist',
      { determination: 14, ambition: 14, professionalism: 14, temperament: 9 },
      { temperament: 10 },
    ],
    ['Temperamental', { professionalism: 10, temperament: 4 }, { temperament: 5 }],
    ['Leader', { leadership: 19, age: 23 }, { leadership: 18 }],
    ['Professional', { professionalism: 18, temperament: 10 }, { professionalism: 17 }],
    ['Determined', { determination: 18, ambition: 11 }, { determination: 17 }],
    ['Driven', { determination: 18, ambition: 12 }, { determination: 17 }],
    ['Casual', { professionalism: 4, determination: 9 }, { professionalism: 5 }],
    ['Slack', { professionalism: 1, determination: 9 }, { determination: 10 }],
    ['Fickle', { loyalty: 6, ambition: 16 }, { loyalty: 7 }],
    ['Mercenary', { loyalty: 3, ambition: 16 }, { ambition: 15 }],
    ['Ambitious', { ambition: 16, loyalty: 9 }, { loyalty: 10 }],
    ['Very Ambitious', { ambition: 20, loyalty: 9 }, { loyalty: 10 }],
    ['Unambitious', { ambition: 5, loyalty: 11 }, { ambition: 6 }],
    ['Sporting', { determination: 9, sportsmanship: 18 }, { determination: 10 }],
    ['Honest', { determination: 9, sportsmanship: 20 }, { sportsmanship: 17 }],
    ['Realist', { sportsmanship: 4, determination: 11 }, { sportsmanship: 5 }],
    ['Unsporting', { sportsmanship: 1, determination: 11 }, { determination: 10 }],
    ['Low Determination', { determination: 5, ambition: 9 }, { ambition: 10 }],
    ['Easily Discouraged', { determination: 1, ambition: 9 }, { determination: 6 }],
    ['Loyal', { ambition: 7, loyalty: 18 }, { ambition: 8 }],
    ['Very Loyal', { ambition: 7, loyalty: 20 }, { loyalty: 17 }],
    ['Resilient', { determination: 15, pressure: 17 }, { pressure: 16 }],
    ['Iron Willed', { determination: 15, pressure: 20 }, { determination: 14 }],
    ['Low Self Belief', { determination: 9, pressure: 3 }, { pressure: 4 }],
    ['Spineless', { determination: 9, pressure: 1 }, { determination: 10 }],
    ['Light-Hearted', { temperament: 10, pressure: 15, sportsmanship: 15 }, { pressure: 14 }],
    ['Spirited', { temperament: 10, pressure: 15, professionalism: 11 }, { temperament: 9 }],
    ['Jovial', { temperament: 10, pressure: 15, professionalism: 10 }, { pressure: 14 }],
    ['Resolute', { determination: 15, professionalism: 15 }, { determination: 14 }],
    ['Fairly Determined', { determination: 15 }, { determination: 14 }],
    ['Fairly Professional', { professionalism: 15 }, { professionalism: 14 }],
    ['Fairly Ambitious', { ambition: 15 }, { ambition: 14 }],
    ['Fairly Loyal', { loyalty: 15 }, { loyalty: 14 }],
    ['Fairly Sporting', { sportsmanship: 15 }, { sportsmanship: 14 }],
  ])('%s at its threshold, and not past it', (label, at, past) => {
    expect(personality(at)).toBe(label)
    expect(personality({ ...at, ...past })).not.toBe(label)
  })

  it('emits the leader variants for adults only', () => {
    expect(personality({ leadership: 20, determination: 20, age: 23 })).toBe('Born Leader')
    expect(personality({ leadership: 19, sportsmanship: 18, temperament: 18, age: 30 })).toBe(
      'Natural Leader'
    )
    expect(personality({ leadership: 20, age: 22 })).toBe('Balanced')
  })

  it('applies precedence: Model Citizen, Perfectionist over Leader, Temperamental before Professional', () => {
    const citizen = {
      sportsmanship: 15,
      loyalty: 15,
      professionalism: 15,
      determination: 14,
      ambition: 14,
      temperament: 15,
      pressure: 14,
    }
    expect(personality({ ...citizen, leadership: 20, age: 30 })).toBe('Model Citizen')
    expect(
      personality({
        determination: 20,
        ambition: 14,
        professionalism: 14,
        temperament: 9,
        leadership: 20,
      })
    ).toBe('Perfectionist')
    expect(personality({ professionalism: 10, temperament: 4, leadership: 20 })).toBe(
      'Temperamental'
    )
    expect(personality({ professionalism: 18, temperament: 10, leadership: 19 })).toBe('Leader')
  })

  it('gates Model Professional at 23', () => {
    expect(personality({ professionalism: 20, temperament: 10, age: 23 })).toBe(
      'Model Professional'
    )
    expect(personality({ professionalism: 20, temperament: 10, age: 22 })).toBe('Professional')
  })

  it('never gives women the negative labels', () => {
    expect(personality({ professionalism: 10, temperament: 4, male: false })).not.toBe(
      'Temperamental'
    )
    expect(personality({ ambition: 5, loyalty: 11, male: false })).toBe('Balanced')
    expect(personality({ determination: 9, pressure: 1, male: false })).toBe('Balanced')
  })

  it('never emits Devoted', () => {
    const seen = new Set()
    for (let v = 1; v <= 20; v++) {
      for (const key of Object.keys(NEUTRAL)) seen.add(personality({ [key]: v }))
      seen.add(personality({ determination: v }))
    }
    expect(seen.has('Devoted')).toBe(false)
  })
})

describe('deriveMediaHandling', () => {
  it('falls back to Level-headed alone', () => {
    expect(media()).toBe('Level-headed')
  })

  it('fills the temperament slot', () => {
    expect(media({ temperament: 2 })).toContain('Short-tempered')
    expect(media({ temperament: 3 })).toContain('Volatile')
    expect(media({ temperament: 6 })).toContain('Volatile')
    expect(media({ temperament: 7 })).toBe('Level-headed')
  })

  it('fills the attitude slot, first match wins', () => {
    expect(media({ professionalism: 15, temperament: 15, pressure: 15, sportsmanship: 13 })).toBe(
      'Evasive, Unflappable'
    )
    // Evasive needs sportsmanship > 12 or controversy < 15; with neither, Outspoken wins.
    expect(media({ professionalism: 15, pressure: 15, sportsmanship: 12, controversy: 15 })).toBe(
      'Outspoken'
    )
    expect(media({ controversy: 15 })).toBe('Outspoken')
    expect(media({ loyalty: 10 })).toBe('Media-friendly')
    expect(media({ sportsmanship: 11, professionalism: 12 })).toBe('Media-friendly')
    expect(media({ sportsmanship: 11, professionalism: 13 })).toBe('Level-headed')
  })

  it('fills the composure slot, first match wins', () => {
    expect(media({ sportsmanship: 7, temperament: 7 })).toContain('Confrontational')
    expect(media({ temperament: 15, pressure: 15 })).toBe('Unflappable')
    expect(media({ temperament: 15, pressure: 14 })).toBe('Level-headed')
    expect(media({ controversy: 5, professionalism: 15 })).toBe('Reserved')
    // Reserved is blocked by a fiery temperament.
    expect(media({ controversy: 5, professionalism: 15, temperament: 6 })).toBe('Volatile')
  })

  it('joins up to three labels in FM order', () => {
    expect(media({ temperament: 5, loyalty: 8, sportsmanship: 6 })).toBe(
      'Volatile, Media-friendly, Confrontational'
    )
    expect(media({ temperament: 1, controversy: 16, sportsmanship: 3 })).toBe(
      'Outspoken, Short-tempered, Confrontational'
    )
  })

  it('shows women as Outspoken instead of the combative labels', () => {
    expect(media({ temperament: 5, loyalty: 8, sportsmanship: 6, male: false })).toBe(
      'Outspoken, Media-friendly'
    )
    expect(media({ temperament: 2, male: false })).toBe('Outspoken')
    expect(media({ temperament: 15, pressure: 15, male: false })).toBe('Unflappable')
  })
})

describe('personalityLabels', () => {
  const player = (overrides = {}) => ({
    personality: { ...NEUTRAL, adaptability: 10 },
    attributes: { determination: 18, leadership: 10 },
    age: 25,
    gender: 0,
    ...overrides,
  })

  it('reads the save player fields', () => {
    expect(personalityLabels(player())).toEqual(['Determined', 'Level-headed'])
  })

  it("treats players at women's clubs (gender 1) as female", () => {
    const bytes = { ...NEUTRAL, adaptability: 10, professionalism: 10, temperament: 4 }
    expect(personalityLabels(player({ personality: bytes }))[0]).toBe('Temperamental')
    expect(personalityLabels(player({ personality: bytes, gender: 1 }))[0]).not.toBe(
      'Temperamental'
    )
  })

  it('is blank without personality bytes', () => {
    expect(personalityLabels(player({ personality: undefined }))).toEqual(['', ''])
  })
})
