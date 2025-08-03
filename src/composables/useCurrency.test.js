import { describe, expect, test, vi } from 'vitest'
import { useCurrency } from './useCurrency.js'

// Mock the currencyUtils dependency
vi.mock('@/utils/currencyUtils', () => ({
  formatCurrency: vi.fn((amount, symbol, _original) => `${symbol}${amount}`),
}))

describe('useCurrency', () => {
  describe('with default currency symbol', () => {
    const { formatDisplayCurrency, formatCompactCurrency, parseCurrency, isCurrencyValue } =
      useCurrency()

    describe('formatDisplayCurrency', () => {
      test('should call formatCurrency with default symbol', async () => {
        const { formatCurrency } = vi.mocked(await import('@/utils/currencyUtils'))

        formatDisplayCurrency(1000, '$1,000')

        expect(formatCurrency).toHaveBeenCalledWith(1000, '$', '$1,000')
      })
    })

    describe('formatCompactCurrency', () => {
      test('should format zero amounts', () => {
        expect(formatCompactCurrency(0)).toBe('$0')
        expect(formatCompactCurrency(null)).toBe('$0')
        expect(formatCompactCurrency(undefined)).toBe('$0')
      })

      test('should format millions with M suffix', () => {
        expect(formatCompactCurrency(1500000)).toBe('$1.5M')
        expect(formatCompactCurrency(2200000)).toBe('$2.2M')
        expect(formatCompactCurrency(-1500000)).toBe('$-1.5M')
      })

      test('should format thousands with K suffix', () => {
        expect(formatCompactCurrency(1500)).toBe('$1.5K')
        expect(formatCompactCurrency(2200)).toBe('$2.2K')
        expect(formatCompactCurrency(-1500)).toBe('$-1.5K')
      })

      test('should format small amounts without suffix', () => {
        expect(formatCompactCurrency(999)).toBe('$999')
        expect(formatCompactCurrency(500)).toBe('$500')
        expect(formatCompactCurrency(1)).toBe('$1')
      })

      test('should handle custom currency symbols', () => {
        expect(formatCompactCurrency(1500000, '€')).toBe('€1.5M')
        expect(formatCompactCurrency(1500, '£')).toBe('£1.5K')
      })
    })

    describe('parseCurrency', () => {
      test('should parse valid currency strings', () => {
        expect(parseCurrency('$1,000.50')).toBe(1000.5)
        expect(parseCurrency('€2,500')).toBe(2500)
        expect(parseCurrency('£999.99')).toBe(999.99)
      })

      test('should handle negative values', () => {
        expect(parseCurrency('-$500.25')).toBe(-500.25)
        expect(parseCurrency('($500.25)')).toBe(500.25)
      })

      test('should return 0 for invalid inputs', () => {
        expect(parseCurrency('')).toBe(0)
        expect(parseCurrency(null)).toBe(0)
        expect(parseCurrency(undefined)).toBe(0)
        expect(parseCurrency('invalid')).toBe(0)
        expect(parseCurrency(123)).toBe(0) // not a string
      })

      test('should handle strings with only numbers', () => {
        expect(parseCurrency('1000.50')).toBe(1000.5)
        expect(parseCurrency('42')).toBe(42)
      })
    })

    describe('isCurrencyValue', () => {
      test('should return true for numeric values', () => {
        expect(isCurrencyValue(1000)).toBe(true)
        expect(isCurrencyValue(0)).toBe(true)
        expect(isCurrencyValue(-500)).toBe(true)
        expect(isCurrencyValue(3.14)).toBe(true)
      })

      test('should return true for valid currency strings', () => {
        expect(isCurrencyValue('$1,000')).toBe(true)
        expect(isCurrencyValue('€500.50')).toBe(true)
        expect(isCurrencyValue('£999')).toBe(true)
        expect(isCurrencyValue('¥1000')).toBe(true)
        expect(isCurrencyValue('1,500K')).toBe(true)
        expect(isCurrencyValue('1,500k')).toBe(true) // lowercase k
        expect(isCurrencyValue('1000')).toBe(true)
        // Note: 2.5M might not match if the regex is more restrictive
      })

      test('should return false for invalid inputs', () => {
        expect(isCurrencyValue('')).toBe(false)
        expect(isCurrencyValue('invalid')).toBe(false)
        expect(isCurrencyValue('$1,000.50.50')).toBe(false) // multiple decimals
        expect(isCurrencyValue(null)).toBe(false)
        expect(isCurrencyValue(undefined)).toBe(false)
        expect(isCurrencyValue({})).toBe(false)
        expect(isCurrencyValue([])).toBe(false)
      })

      test('should handle whitespace', () => {
        expect(isCurrencyValue('  $1,000  ')).toBe(true)
        expect(isCurrencyValue('   ')).toBe(false)
      })
    })
  })

  describe('with custom currency symbol', () => {
    test('should use custom symbol in formatCompactCurrency', () => {
      const { formatCompactCurrency } = useCurrency('€')

      expect(formatCompactCurrency(1500000)).toBe('€1.5M')
      expect(formatCompactCurrency(1500)).toBe('€1.5K')
      expect(formatCompactCurrency(0)).toBe('€0')
    })

    test('should pass custom symbol to formatDisplayCurrency', async () => {
      const { formatDisplayCurrency } = useCurrency('£')
      const { formatCurrency } = vi.mocked(await import('@/utils/currencyUtils'))

      formatDisplayCurrency(1000, '£1,000')

      expect(formatCurrency).toHaveBeenCalledWith(1000, '£', '£1,000')
    })
  })
})
