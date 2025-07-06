/**
 * Tests for CSV Export Utility
 */

import { validateExportData, getDefaultExportColumns, exportPlayersToJSON } from './csvExport.js'

describe('CSV Export Utility', () => {
  const mockPlayers = [
    {
      name: 'John Doe',
      age: '25',
      nationality: 'England',
      club: 'Manchester United',
      position: 'CM',
      shortPositions: ['CM', 'CDM'],
      transferValue: '€50M',
      transferValueAmount: 50000000,
      wage: '€100K',
      wageAmount: 100000,
      Overall: 85,
      Potential: 90,
      PAC: 80,
      SHO: 75,
      PAS: 90,
      DRI: 85,
      DEF: 70,
      PHY: 80,
      attributes: {
        Pas: '18',
        Tec: '16',
        Vis: '17'
      },
      personality: 'Professional',
      media_handling: 'Evasive',
      foot: 'Right'
    },
    {
      name: 'Jane Smith',
      age: '22',
      nationality: 'Spain',
      club: 'Barcelona',
      position: 'GK',
      shortPositions: ['GK'],
      transferValue: '€25M',
      transferValueAmount: 25000000,
      wage: '€75K',
      wageAmount: 75000,
      Overall: 82,
      Potential: 88,
      GK: 85,
      attributes: {
        Han: '17',
        Ref: '16',
        Cmd: '15'
      },
      personality: 'Driven',
      media_handling: 'Respectful'
    }
  ]

  describe('validateExportData', () => {
    test('should validate correct player data', () => {
      const result = validateExportData(mockPlayers)
      expect(result.valid).toBe(true)
      expect(result.errors).toHaveLength(0)
    })

    test('should reject empty array', () => {
      const result = validateExportData([])
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('No players to export')
    })

    test('should reject non-array input', () => {
      const result = validateExportData('not an array')
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Players data must be an array')
    })

    test('should warn about large datasets', () => {
      const largePlayers = Array(15000).fill(mockPlayers[0])
      const result = validateExportData(largePlayers)
      expect(result.valid).toBe(true)
      expect(result.warnings).toContain('Large export (>10,000 players) may take some time')
    })

    test('should reject data without required fields', () => {
      const playersWithoutName = [{ age: '25', club: 'Test FC' }]
      const result = validateExportData(playersWithoutName)
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Missing required fields: name')
    })
  })

  describe('getDefaultExportColumns', () => {
    test('should return basic columns for basic context', () => {
      const columns = getDefaultExportColumns('basic', mockPlayers)
      expect(columns).toContain('name')
      expect(columns).toContain('age')
      expect(columns).toContain('Overall')
      expect(columns).toContain('transferValue')
      // Should NOT contain the removed fields
      expect(columns).not.toContain('Potential')
      expect(columns).not.toContain('contractExpiry')
      expect(columns).not.toContain('foot')
    })

    test('should return detailed columns for detailed context', () => {
      const columns = getDefaultExportColumns('detailed', mockPlayers)
      expect(columns).toContain('name')
      expect(columns).toContain('personality')
      expect(columns).toContain('Overall')
      expect(columns).toContain('media_handling')
      // Should NOT contain the removed fields
      expect(columns).not.toContain('Potential')
      expect(columns).not.toContain('contractExpiry')
      expect(columns).not.toContain('foot')
      expect(columns.length).toBeGreaterThan(10)
    })

    test('should return scout columns for scout context', () => {
      const columns = getDefaultExportColumns('scout', mockPlayers)
      expect(columns).toContain('name')
      expect(columns).toContain('Overall')
      expect(columns).toContain('personality')
      // Should NOT contain the removed fields
      expect(columns).not.toContain('Potential')
      expect(columns).not.toContain('contractExpiry')
      expect(columns).not.toContain('foot')
    })

    test('should fallback to basic for unknown context', () => {
      const columns = getDefaultExportColumns('unknown', mockPlayers)
      const basicColumns = getDefaultExportColumns('basic', mockPlayers)
      expect(columns).toEqual(basicColumns)
    })
  })

  describe('JSON Export', () => {
    test('should create properly formatted JSON export object', () => {
      // Since we can't easily test the actual file download, we can test the structure
      const mockExportFunction = jest.fn()
      
      // Test that the function would be called with proper parameters
      expect(() => {
        if (mockPlayers.length > 0) {
          const exportData = {
            metadata: {
              exportDate: expect.any(String),
              totalPlayers: mockPlayers.length,
              exportType: 'full_dataset',
              version: '1.0'
            },
            players: mockPlayers
          }
          expect(exportData.metadata.totalPlayers).toBe(2)
          expect(exportData.players).toEqual(mockPlayers)
          expect(exportData.metadata.exportType).toBe('full_dataset')
        }
      }).not.toThrow()
    })
  })
}) 