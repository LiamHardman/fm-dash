import { beforeEach, describe, expect, it } from 'vitest'
import { watchEffect } from 'vue'
import {
  activeColourBands,
  applyColourVariables,
  colourText,
  defaultBands,
  getUnifiedRatingClass,
  validateBands,
} from './colourSets'

beforeEach(() => {
  activeColourBands.fm = defaultBands('fm')
  activeColourBands.fifa = defaultBands('fifa')
})

describe('attribute colour bands', () => {
  it.each(['fm', 'fifa'])('supports every boundary with 20 bands for %s', (scale) => {
    const min = scale === 'fm' ? 1 : 0
    const step = scale === 'fm' ? 1 : 5
    const max = scale === 'fm' ? 20 : 100
    activeColourBands[scale] = Array.from({ length: 20 }, (_, i) => ({
      min: min + i * step,
      color: '#abcdef',
    }))
    expect(validateBands(activeColourBands[scale], scale)).toBe('')
    for (let i = 0; i < 20; i++) {
      expect(getUnifiedRatingClass(min + i * step, max)).toBe(
        `attribute-colour attribute-colour-${scale}-${i}`
      )
    }
    expect(getUnifiedRatingClass(max, max)).toBe(`attribute-colour attribute-colour-${scale}-19`)
  })

  it('updates repeated values reactively and independently when thresholds change', () => {
    let actual
    const stop = watchEffect(
      () => {
        actual = getUnifiedRatingClass(14, 20)
      },
      { flush: 'sync' }
    )
    const fifa = getUnifiedRatingClass(70, 100)
    activeColourBands.fm = [
      { min: 1, color: '#ffffff' },
      { min: 15, color: '#000000' },
    ]
    expect(actual).toBe('attribute-colour attribute-colour-fm-0')
    expect(getUnifiedRatingClass('15', 20)).toBe('attribute-colour attribute-colour-fm-1')
    expect(getUnifiedRatingClass(14.9, 20)).toBe('attribute-colour attribute-colour-fm-0')
    expect(getUnifiedRatingClass(70, 100)).toBe(fifa)
    stop()
  })

  it.each([null, undefined, '', ' ', '-', '12-15', Number.NaN, Number.POSITIVE_INFINITY])(
    'keeps missing or masked value %s neutral',
    (value) => {
      expect(getUnifiedRatingClass(value, 20)).toBe('rating-na')
    }
  )

  it('covers zero FIFA stats but keeps zero FM attributes neutral', () => {
    expect(getUnifiedRatingClass(0, 100)).toBe('attribute-colour attribute-colour-fifa-0')
    expect(getUnifiedRatingClass(0, 20)).toBe('rating-na')
  })

  it('rejects invalid bands and selects readable text colours', () => {
    expect(validateBands([{ min: 1, color: 'red' }], 'fm')).toBeTruthy()
    expect(
      validateBands(
        [
          { min: 1, color: '#ffffff' },
          { min: 1, color: '#ffffff' },
        ],
        'fm'
      )
    ).toBeTruthy()
    expect(validateBands([{ min: 0, color: '#ffffff' }], 'fm')).toBeTruthy()
    expect(
      validateBands(
        [
          { min: 1, color: '#ffffff' },
          { min: 21, color: '#ffffff' },
        ],
        'fm'
      )
    ).toBeTruthy()
    expect(colourText('#ffffff')).toBe('#000000')
    expect(colourText('#000000')).toBe('#ffffff')
    applyColourVariables(document.documentElement.style)
    expect(document.documentElement.style.getPropertyValue('--attribute-fm-0')).toBe('#ef5350')
  })
})
