import { describe, expect, test } from 'vitest'
import {
  getNumericValue,
  getPlayerAge,
  getPlayerDivision,
  getPlayerOverall,
  getPlayerPositions,
  getPlayerTransferValue,
  getPlayerWage,
  playerMatchesPosition,
} from './playerUtils.js'

describe('playerUtils', () => {
  describe('getPlayerDivision', () => {
    test('should return division from lowercase field', () => {
      const player = { division: 'Premier League' }
      expect(getPlayerDivision(player)).toBe('Premier League')
    })

    test('should return division from capitalized field', () => {
      const player = { Division: 'Championship' }
      expect(getPlayerDivision(player)).toBe('Championship')
    })

    test('should prefer lowercase over capitalized field', () => {
      const player = { division: 'Premier League', Division: 'Championship' }
      expect(getPlayerDivision(player)).toBe('Premier League')
    })

    test('should return N/A when no division found', () => {
      const player = {}
      expect(getPlayerDivision(player)).toBe('N/A')
    })

    test('should return N/A for null division', () => {
      const player = { division: null }
      expect(getPlayerDivision(player)).toBe('N/A')
    })
  })

  describe('getNumericValue', () => {
    test('should convert valid numbers', () => {
      expect(getNumericValue(42)).toBe(42)
      expect(getNumericValue(3.14)).toBe(3.14)
      expect(getNumericValue('100')).toBe(100)
      expect(getNumericValue('99.5')).toBe(99.5)
    })

    test('should handle values with commas', () => {
      expect(getNumericValue('1,000')).toBe(1000)
      expect(getNumericValue('1,234,567')).toBe(1234567)
    })

    test('should handle percentage values', () => {
      expect(getNumericValue('75%')).toBe(75)
      expect(getNumericValue('100%')).toBe(100)
    })

    test('should return null for invalid values', () => {
      expect(getNumericValue(undefined)).toBe(null)
      expect(getNumericValue(null)).toBe(null)
      expect(getNumericValue('-')).toBe(null)
      expect(getNumericValue('')).toBe(null)
      expect(getNumericValue('invalid')).toBe(null)
    })

    test('should handle negative numbers', () => {
      expect(getNumericValue('-42')).toBe(-42)
      expect(getNumericValue(-100)).toBe(-100)
    })

    test('should handle zero values', () => {
      expect(getNumericValue(0)).toBe(0)
      expect(getNumericValue('0')).toBe(0)
    })
  })

  describe('getPlayerPositions', () => {
    test('should return shortPositions array', () => {
      const player = { shortPositions: ['ST', 'LW', 'RW'] }
      expect(getPlayerPositions(player)).toEqual(['ST', 'LW', 'RW'])
    })

    test('should return short_positions array', () => {
      const player = { short_positions: ['CB', 'RB'] }
      expect(getPlayerPositions(player)).toEqual(['CB', 'RB'])
    })

    test('should prefer shortPositions over short_positions', () => {
      const player = {
        shortPositions: ['ST', 'LW'],
        short_positions: ['CB', 'RB'],
      }
      expect(getPlayerPositions(player)).toEqual(['ST', 'LW'])
    })

    test('should return empty array when no positions found', () => {
      const player = {}
      expect(getPlayerPositions(player)).toEqual([])
    })

    test('should return empty array for null positions', () => {
      const player = { shortPositions: null }
      expect(getPlayerPositions(player)).toEqual([])
    })
  })

  describe('playerMatchesPosition', () => {
    test('should return true when player has the position', () => {
      const player = { shortPositions: ['ST', 'LW', 'RW'] }
      expect(playerMatchesPosition(player, 'ST')).toBe(true)
      expect(playerMatchesPosition(player, 'LW')).toBe(true)
    })

    test('should return false when player does not have the position', () => {
      const player = { shortPositions: ['ST', 'LW', 'RW'] }
      expect(playerMatchesPosition(player, 'CB')).toBe(false)
      expect(playerMatchesPosition(player, 'GK')).toBe(false)
    })

    test('should return false for empty positions', () => {
      const player = { shortPositions: [] }
      expect(playerMatchesPosition(player, 'ST')).toBe(false)
    })

    test('should work with short_positions field', () => {
      const player = { short_positions: ['CB', 'RB'] }
      expect(playerMatchesPosition(player, 'CB')).toBe(true)
      expect(playerMatchesPosition(player, 'ST')).toBe(false)
    })
  })

  describe('getPlayerOverall', () => {
    test('should return Overall field value', () => {
      const player = { Overall: 85 }
      expect(getPlayerOverall(player)).toBe(85)
    })

    test('should return overall field value', () => {
      const player = { overall: 90 }
      expect(getPlayerOverall(player)).toBe(90)
    })

    test('should prefer Overall over overall', () => {
      const player = { Overall: 85, overall: 90 }
      expect(getPlayerOverall(player)).toBe(85)
    })

    test('should return 0 when no overall found', () => {
      const player = {}
      expect(getPlayerOverall(player)).toBe(0)
    })

    test('should return 0 for null overall', () => {
      const player = { Overall: null }
      expect(getPlayerOverall(player)).toBe(0)
    })
  })

  describe('getPlayerAge', () => {
    test('should return age as number', () => {
      const player = { age: 25 }
      expect(getPlayerAge(player)).toBe(25)
    })

    test('should return Age field value', () => {
      const player = { Age: 30 }
      expect(getPlayerAge(player)).toBe(30)
    })

    test('should parse string ages', () => {
      const player = { age: '28' }
      expect(getPlayerAge(player)).toBe(28)
    })

    test('should prefer age over Age field', () => {
      const player = { age: 25, Age: 30 }
      expect(getPlayerAge(player)).toBe(25)
    })

    test('should return 0 when no age found', () => {
      const player = {}
      expect(getPlayerAge(player)).toBe(0)
    })

    test('should handle invalid age values', () => {
      const player = { age: 'invalid' }
      expect(Number.isNaN(getPlayerAge(player))).toBe(true)
    })
  })

  describe('getPlayerTransferValue', () => {
    test('should return transferValueAmount as number', () => {
      const player = { transferValueAmount: 50000000 }
      expect(getPlayerTransferValue(player)).toBe(50000000)
    })

    test('should return TransferValueAmount field value', () => {
      const player = { TransferValueAmount: 25000000 }
      expect(getPlayerTransferValue(player)).toBe(25000000)
    })

    test('should parse string values', () => {
      const player = { transferValueAmount: '30000000' }
      expect(getPlayerTransferValue(player)).toBe(30000000)
    })

    test('should prefer transferValueAmount over TransferValueAmount', () => {
      const player = { transferValueAmount: 50000000, TransferValueAmount: 25000000 }
      expect(getPlayerTransferValue(player)).toBe(50000000)
    })

    test('should return 0 when no transfer value found', () => {
      const player = {}
      expect(getPlayerTransferValue(player)).toBe(0)
    })

    test('should handle invalid transfer values', () => {
      const player = { transferValueAmount: 'invalid' }
      expect(Number.isNaN(getPlayerTransferValue(player))).toBe(true)
    })
  })

  describe('getPlayerWage', () => {
    test('should return wageAmount as number', () => {
      const player = { wageAmount: 100000 }
      expect(getPlayerWage(player)).toBe(100000)
    })

    test('should return WageAmount field value', () => {
      const player = { WageAmount: 75000 }
      expect(getPlayerWage(player)).toBe(75000)
    })

    test('should parse string values', () => {
      const player = { wageAmount: '85000' }
      expect(getPlayerWage(player)).toBe(85000)
    })

    test('should prefer wageAmount over WageAmount', () => {
      const player = { wageAmount: 100000, WageAmount: 75000 }
      expect(getPlayerWage(player)).toBe(100000)
    })

    test('should return 0 when no wage found', () => {
      const player = {}
      expect(getPlayerWage(player)).toBe(0)
    })

    test('should handle invalid wage values', () => {
      const player = { wageAmount: 'invalid' }
      expect(Number.isNaN(getPlayerWage(player))).toBe(true)
    })
  })
})
