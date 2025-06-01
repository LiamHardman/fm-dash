
/**
 * Formats a numeric monetary value into a string with a currency symbol and K/M suffix.
 * Example: 1500000 with symbol '€' becomes "€1.5M".
 * Example: 50000 with symbol '$' becomes "$50K".
 * Example: 1200 with symbol '£' becomes "£1.2K" (or "£1200" if prefer no K for <10k).
 *
 * @param {number|string} amount - The numeric amount or a string that can be parsed to a number.
 * @param {string} symbol - The currency symbol (e.g., "€", "$", "£").
 * @param {string} originalDisplayValue - The original string value from the HTML (e.g., "€1.5M", "£55K p/w").
 * This is used as a fallback if the amount is not a valid number
 * or if the formatting logic doesn't produce a better result.
 * @returns {string} The formatted currency string or the original display value if formatting fails.
 */
export function formatCurrency(amount, symbol, originalDisplayValue) {
  const defaultSymbol = symbol || '$' // Default to '$' if no symbol is provided

  // If the originalDisplayValue already seems well-formatted (e.g., "€1.5M", "Not for Sale"), return it.
  // This is a simple heuristic. More complex logic might be needed for edge cases.
  if (
    typeof originalDisplayValue === 'string' &&
    originalDisplayValue.trim() !== '-' &&
    originalDisplayValue.trim() !== ''
  ) {
    if (
      originalDisplayValue.includes(defaultSymbol) &&
      (originalDisplayValue.toUpperCase().includes('M') ||
        originalDisplayValue.toUpperCase().includes('K'))
    ) {
      return originalDisplayValue
    }
    // Handle cases like "Not for Sale" or other non-numeric strings
    if (Number.isNaN(Number.parseFloat(originalDisplayValue.replace(/[^0-9.-]+/g, '')))) {
      return originalDisplayValue
    }
  }

  const numAmount = Number(amount)

  if (Number.isNaN(numAmount)) {
    // If the amount is not a number, return the original display value or the symbol if no original value
    return originalDisplayValue || defaultSymbol
  }

  if (numAmount === 0 && (originalDisplayValue === '-' || !originalDisplayValue)) {
    return `${defaultSymbol}0`
  }
  if (
    numAmount === 0 &&
    originalDisplayValue &&
    (originalDisplayValue.includes(defaultSymbol) || /[€$£]/.test(originalDisplayValue))
  ) {
    return originalDisplayValue // e.g. €0, $0 - return original if it has any currency symbol
  }
  if (numAmount === 0 && originalDisplayValue) {
    // e.g. if original was just "0" or something non-standard
    return originalDisplayValue // Return the original value instead of forcing defaultSymbol
  }

  let valueString
  if (Math.abs(numAmount) >= 1000000) {
    valueString = `${(numAmount / 1000000).toFixed(1).replace(/\.0$/, '')}M`
  } else if (Math.abs(numAmount) >= 1000) {
    valueString = `${(numAmount / 1000).toFixed(1).replace(/\.0$/, '')}K`
  } else {
    valueString = numAmount.toString()
  }

  return `${defaultSymbol}${valueString}`
}

/**
 * Parses a monetary string (like "€1.5M", "£500K", "$1200") into a numeric value.
 * This is primarily for client-side filtering or input parsing if needed.
 * The backend already does this with `parseMonetaryValueGo`.
 *
 * @param {string} valueString - The string to parse.
 * @returns {number|null} The numeric value, or null if parsing fails.
 */
export function parseCurrencyString(valueString) {
  if (typeof valueString !== 'string' || !valueString.trim()) {
    return null
  }

  const cleanedStr = valueString
    .toUpperCase()
    .replace(/[€$£]/g, '') // Remove common currency symbols
    .replace(/P\/W|\/W/g, '') // Remove per week indicators
    .trim()

  let multiplier = 1
  let numericPart = cleanedStr

  if (cleanedStr.endsWith('M')) {
    multiplier = 1000000
    numericPart = cleanedStr.substring(0, cleanedStr.length - 1)
  } else if (cleanedStr.endsWith('K')) {
    multiplier = 1000
    numericPart = cleanedStr.substring(0, cleanedStr.length - 1)
  }

  numericPart = numericPart.replace(/,/g, '') // Remove commas

  const value = Number.parseFloat(numericPart)

  if (Number.isNaN(value)) {
    return null
  }

  return Math.round(value * multiplier)
}
