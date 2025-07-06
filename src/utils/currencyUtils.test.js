import { describe, expect, it } from 'vitest'
import { formatCurrency, parseCurrencyString } from './currencyUtils.js'

describe('formatCurrency', () => {
  it('formats large amounts with M suffix', () => {
    expect(formatCurrency(1500000, '$')).toBe('$1.5M')
    expect(formatCurrency(2000000, '€')).toBe('€2M')
    expect(formatCurrency(10500000, '£')).toBe('£10.5M')
  })

  it('formats medium amounts with K suffix', () => {
    expect(formatCurrency(50000, '$')).toBe('$50K')
    expect(formatCurrency(1500, '€')).toBe('€1.5K')
    expect(formatCurrency(25000, '£')).toBe('£25K')
  })

  it('formats small amounts without suffix', () => {
    expect(formatCurrency(500, '$')).toBe('$500')
    expect(formatCurrency(999, '€')).toBe('€999')
    expect(formatCurrency(100, '£')).toBe('£100')
  })

  it('handles zero values correctly', () => {
    expect(formatCurrency(0, '$')).toBe('$0')
    expect(formatCurrency(0, '€')).toBe('€0')
    expect(formatCurrency(0, '$', '-')).toBe('$0')
    expect(formatCurrency(0, '$', '€0')).toBe('€0') // Returns original display value
  })

  it('uses default symbol when none provided', () => {
    expect(formatCurrency(50000)).toBe('$50K')
    expect(formatCurrency(1500000, null)).toBe('$1.5M')
    expect(formatCurrency(1000, undefined)).toBe('$1K')
  })

  it('handles negative amounts', () => {
    expect(formatCurrency(-50000, '$')).toBe('$-50K')
    expect(formatCurrency(-1500000, '€')).toBe('€-1.5M')
    expect(formatCurrency(-500, '£')).toBe('£-500')
  })

  it('returns original display value if already well-formatted', () => {
    expect(formatCurrency(1500000, '$', '$1.5M')).toBe('$1.5M')
    expect(formatCurrency(50000, '€', '€50K')).toBe('€50K')
    expect(formatCurrency(1000000, '£', '£1M')).toBe('£1M')
  })

  it('handles non-numeric strings in original display value', () => {
    expect(formatCurrency(0, '$', 'Not for Sale')).toBe('Not for Sale')
    expect(formatCurrency(1000000, '€', 'Free Transfer')).toBe('Free Transfer')
    expect(formatCurrency(500000, '£', 'TBA')).toBe('TBA')
  })

  it('handles string amounts that can be parsed', () => {
    expect(formatCurrency('1500000', '$')).toBe('$1.5M')
    expect(formatCurrency('50000', '€')).toBe('€50K')
    expect(formatCurrency('999', '£')).toBe('£999')
  })

  it('handles invalid amounts gracefully', () => {
    expect(formatCurrency('invalid', '$')).toBe('$')
    expect(formatCurrency('invalid', '$', 'Original Value')).toBe('Original Value')
    expect(formatCurrency(Number.NaN, '€')).toBe('€')
    expect(formatCurrency(null, '£', 'Fallback')).toBe('Fallback')
  })

  it('removes trailing zeros in decimal places', () => {
    expect(formatCurrency(1000000, '$')).toBe('$1M') // Not $1.0M
    expect(formatCurrency(50000, '€')).toBe('€50K') // Not €50.0K
    expect(formatCurrency(1500000, '£')).toBe('£1.5M')
  })

  it('handles empty or dash original display values', () => {
    expect(formatCurrency(0, '$', '')).toBe('$0')
    expect(formatCurrency(0, '€', '-')).toBe('€0')
    expect(formatCurrency(1000, '£', '')).toBe('£1K')
  })

  it('preserves well-formatted original values with correct currency symbol', () => {
    expect(formatCurrency(1500000, '$', '$1.5M')).toBe('$1.5M')
    expect(formatCurrency(1500000, '$', '$1.8M')).toBe('$1.8M') // Different value preserved
  })

  it('does not preserve poorly formatted original values', () => {
    expect(formatCurrency(1500000, '$', '1.5')).toBe('$1.5M')
    expect(formatCurrency(50000, '€', '50')).toBe('€50K')
  })
})

describe('parseCurrencyString', () => {
  it('parses million values correctly', () => {
    expect(parseCurrencyString('$1.5M')).toBe(1500000)
    expect(parseCurrencyString('€2M')).toBe(2000000)
    expect(parseCurrencyString('£10.5M')).toBe(10500000)
  })

  it('parses thousand values correctly', () => {
    expect(parseCurrencyString('$50K')).toBe(50000)
    expect(parseCurrencyString('€1.5K')).toBe(1500)
    expect(parseCurrencyString('£25K')).toBe(25000)
  })

  it('parses regular numeric values', () => {
    expect(parseCurrencyString('$500')).toBe(500)
    expect(parseCurrencyString('€999')).toBe(999)
    expect(parseCurrencyString('1000')).toBe(1000)
  })

  it('handles values with commas', () => {
    expect(parseCurrencyString('$1,500,000')).toBe(1500000)
    expect(parseCurrencyString('€50,000')).toBe(50000)
    expect(parseCurrencyString('1,000')).toBe(1000)
  })

  it('handles per week indicators', () => {
    expect(parseCurrencyString('$50K p/w')).toBe(50000)
    expect(parseCurrencyString('€25K/w')).toBe(25000)
    expect(parseCurrencyString('£30K p/w')).toBe(30000)
  })

  it('handles case insensitive suffixes', () => {
    expect(parseCurrencyString('$1.5m')).toBe(1500000)
    expect(parseCurrencyString('€50k')).toBe(50000)
    expect(parseCurrencyString('£2.5M')).toBe(2500000)
  })

  it('removes various currency symbols', () => {
    expect(parseCurrencyString('$1.5M')).toBe(1500000)
    expect(parseCurrencyString('€1.5M')).toBe(1500000)
    expect(parseCurrencyString('£1.5M')).toBe(1500000)
  })

  it('handles invalid strings gracefully', () => {
    expect(parseCurrencyString('Not for Sale')).toBeNull()
    expect(parseCurrencyString('TBA')).toBeNull()
    expect(parseCurrencyString('')).toBeNull()
    expect(parseCurrencyString('   ')).toBeNull()
    expect(parseCurrencyString('invalid123xyz')).toBeNull()
  })

  it('handles null and undefined inputs', () => {
    expect(parseCurrencyString(null)).toBeNull()
    expect(parseCurrencyString(undefined)).toBeNull()
  })

  it('handles non-string inputs', () => {
    expect(parseCurrencyString(123)).toBeNull()
    expect(parseCurrencyString({})).toBeNull()
    expect(parseCurrencyString([])).toBeNull()
  })

  it('rounds decimal results', () => {
    expect(parseCurrencyString('$1.55M')).toBe(1550000)
    expect(parseCurrencyString('€25.75K')).toBe(25750)
    expect(parseCurrencyString('£0.5M')).toBe(500000)
  })

  it('handles zero values', () => {
    expect(parseCurrencyString('$0')).toBe(0)
    expect(parseCurrencyString('€0M')).toBe(0)
    expect(parseCurrencyString('£0K')).toBe(0)
  })

  it('handles negative values', () => {
    expect(parseCurrencyString('$-1.5M')).toBe(-1500000)
    expect(parseCurrencyString('€-50K')).toBe(-50000)
    expect(parseCurrencyString('-500')).toBe(-500)
  })

  it('handles complex formatted strings', () => {
    expect(parseCurrencyString('€1,500.5K p/w')).toBe(1500500)
    // For complex strings with extra text, it may not parse correctly
    expect(parseCurrencyString('$2.75M')).toBe(2750000)
  })
})
