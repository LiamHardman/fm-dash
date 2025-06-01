import { formatCurrency } from '@/utils/currencyUtils'
import { computed } from 'vue'

export function useCurrency(currencySymbol = '$') {
  // Format currency with the provided symbol
  const formatDisplayCurrency = (numericAmount, originalDisplayValue) => {
    return formatCurrency(numericAmount, currencySymbol, originalDisplayValue)
  }

  // Format large numbers in a compact way (e.g., 1.5M, 2.3K)
  const formatCompactCurrency = (amount, symbol = currencySymbol) => {
    if (!amount || amount === 0) return `${symbol}0`

    const absAmount = Math.abs(amount)

    if (absAmount >= 1000000) {
      return `${symbol}${(amount / 1000000).toFixed(1)}M`
    }
    if (absAmount >= 1000) {
      return `${symbol}${(amount / 1000).toFixed(1)}K`
    }
    return `${symbol}${amount.toLocaleString()}`
  }

  // Parse currency string back to number
  const parseCurrency = currencyString => {
    if (!currencyString || typeof currencyString !== 'string') return 0

    const cleaned = currencyString.replace(/[^\d.-]/g, '')
    const parsed = Number.parseFloat(cleaned)
    return Number.isNaN(parsed) ? 0 : parsed
  }

  // Check if a value represents currency
  const isCurrencyValue = value => {
    if (typeof value === 'number') return true
    if (typeof value === 'string') {
      return /^[\$£€¥]?[\d,]+(\.\d{2})?[KkMm]?$/.test(value.trim())
    }
    return false
  }

  return {
    formatDisplayCurrency,
    formatCompactCurrency,
    parseCurrency,
    isCurrencyValue
  }
}
