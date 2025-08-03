import { mount } from '@vue/test-utils'
import { describe, expect, test, vi } from 'vitest'
import StatCard from './StatCard.vue'

// Mock Quasar components
const mockQuasarComponents = {
  QCard: { template: '<div class="q-card"><slot /></div>' },
  QCardSection: { template: '<div class="q-card-section"><slot /></div>' },
  QList: { template: '<div class="q-list"><slot /></div>' },
  QItem: {
    template: '<div class="q-item" @click="$emit(\'click\')"><slot /></div>',
    emits: ['click'],
  },
  QItemSection: { template: '<div class="q-item-section"><slot /></div>' },
  QItemLabel: { template: '<div class="q-item-label"><slot /></div>' },
}

describe('StatCard', () => {
  const mockStat = {
    name: 'Goals',
    key: 'goals',
  }

  const mockPlayers = [
    {
      id: 1,
      name: 'Lionel Messi',
      club: 'PSG',
      attributes: { goals: 25 },
    },
    {
      id: 2,
      name: 'Cristiano Ronaldo',
      club: 'Al Nassr',
      attributes: { goals: 20 },
    },
    {
      id: 3,
      name: 'Kylian Mbappé',
      club: 'PSG',
      attributes: { goals: 18 },
    },
  ]

  const createWrapper = (props = {}) => {
    return mount(StatCard, {
      props: {
        stat: mockStat,
        players: mockPlayers,
        ...props,
      },
      global: {
        components: mockQuasarComponents,
      },
    })
  }

  describe('rendering', () => {
    test('should render stat name', () => {
      const wrapper = createWrapper()
      expect(wrapper.text()).toContain('Goals')
    })

    test('should render players list when players exist', () => {
      const wrapper = createWrapper()
      expect(wrapper.text()).toContain('Lionel Messi')
      expect(wrapper.text()).toContain('Cristiano Ronaldo')
      expect(wrapper.text()).toContain('Kylian Mbappé')
    })

    test('should render clubs for players', () => {
      const wrapper = createWrapper()
      expect(wrapper.text()).toContain('PSG')
      expect(wrapper.text()).toContain('Al Nassr')
    })

    test('should render stat values', () => {
      const wrapper = createWrapper()
      expect(wrapper.text()).toContain('25')
      expect(wrapper.text()).toContain('20')
      expect(wrapper.text()).toContain('18')
    })

    test('should render rank badges', () => {
      const wrapper = createWrapper()
      expect(wrapper.text()).toContain('1')
      expect(wrapper.text()).toContain('2')
      expect(wrapper.text()).toContain('3')
    })

    test('should show no data message when no players', () => {
      const wrapper = createWrapper({ players: [] })
      expect(wrapper.text()).toContain('No players match filters')
    })

    test('should show no data message when players is null', () => {
      const wrapper = createWrapper({ players: null })
      expect(wrapper.text()).toContain('No players match filters')
    })
  })

  describe('player interactions', () => {
    test('should emit player-click when player is clicked', async () => {
      const wrapper = createWrapper()
      const playerItems = wrapper.findAll('.q-item')

      await playerItems[0].trigger('click')

      expect(wrapper.emitted('player-click')).toBeTruthy()
      expect(wrapper.emitted('player-click')[0][0]).toEqual(mockPlayers[0])
    })

    test('should emit correct player data for each player', async () => {
      const wrapper = createWrapper()
      const playerItems = wrapper.findAll('.q-item')

      await playerItems[1].trigger('click')

      expect(wrapper.emitted('player-click')[0][0]).toEqual(mockPlayers[1])
    })
  })

  describe('getPlayerName method', () => {
    test('should return name field', () => {
      const wrapper = createWrapper()
      expect(wrapper.vm.getPlayerName({ name: 'John Doe' })).toBe('John Doe')
    })

    test('should return Name field if name not available', () => {
      const wrapper = createWrapper()
      expect(wrapper.vm.getPlayerName({ Name: 'Jane Doe' })).toBe('Jane Doe')
    })

    test('should return Player field if others not available', () => {
      const wrapper = createWrapper()
      expect(wrapper.vm.getPlayerName({ Player: 'Test Player' })).toBe('Test Player')
    })

    test('should return Unknown Player if no name fields', () => {
      const wrapper = createWrapper()
      expect(wrapper.vm.getPlayerName({})).toBe('Unknown Player')
    })

    test('should prefer name over Name and Player', () => {
      const wrapper = createWrapper()
      const player = { name: 'First', Name: 'Second', Player: 'Third' }
      expect(wrapper.vm.getPlayerName(player)).toBe('First')
    })
  })

  describe('getPlayerClub method', () => {
    test('should return club field', () => {
      const wrapper = createWrapper()
      expect(wrapper.vm.getPlayerClub({ club: 'Barcelona' })).toBe('Barcelona')
    })

    test('should return Club field if club not available', () => {
      const wrapper = createWrapper()
      expect(wrapper.vm.getPlayerClub({ Club: 'Real Madrid' })).toBe('Real Madrid')
    })

    test('should return Unknown Club if no club fields', () => {
      const wrapper = createWrapper()
      expect(wrapper.vm.getPlayerClub({})).toBe('Unknown Club')
    })

    test('should prefer club over Club', () => {
      const wrapper = createWrapper()
      const player = { club: 'First Club', Club: 'Second Club' }
      expect(wrapper.vm.getPlayerClub(player)).toBe('First Club')
    })
  })

  describe('formatStatValue method', () => {
    const wrapper = createWrapper()

    test('should return N/A for null/undefined/empty values', () => {
      expect(wrapper.vm.formatStatValue(null, 'goals')).toBe('N/A')
      expect(wrapper.vm.formatStatValue(undefined, 'goals')).toBe('N/A')
      expect(wrapper.vm.formatStatValue('-', 'goals')).toBe('N/A')
      expect(wrapper.vm.formatStatValue('', 'goals')).toBe('N/A')
    })

    test('should format percentage values', () => {
      expect(wrapper.vm.formatStatValue('85%', 'accuracy')).toBe('85.0%')
      expect(wrapper.vm.formatStatValue(85, 'Sv %')).toBe('85.0%')
      expect(wrapper.vm.formatStatValue(90, 'Conv %')).toBe('90.0%')
      expect(wrapper.vm.formatStatValue(75, 'Pas %')).toBe('75.0%')
      expect(wrapper.vm.formatStatValue(80, 'Tck R')).toBe('80.0%')
    })

    test('should format large integers with thousands separator', () => {
      expect(wrapper.vm.formatStatValue(15000, 'value')).toBe('15,000')
      expect(wrapper.vm.formatStatValue(1500000, 'transfer_value')).toBe('1,500,000')
    })

    test('should format integers without decimal places', () => {
      expect(wrapper.vm.formatStatValue(25, 'goals')).toBe('25')
      expect(wrapper.vm.formatStatValue(100, 'games')).toBe('100')
    })

    test('should format decimals with 2 decimal places', () => {
      expect(wrapper.vm.formatStatValue(2.567, 'rating')).toBe('2.57')
      expect(wrapper.vm.formatStatValue(1.234, 'ratio')).toBe('1.23')
    })

    test('should handle string numbers', () => {
      expect(wrapper.vm.formatStatValue('25', 'goals')).toBe('25')
      expect(wrapper.vm.formatStatValue('2.5', 'rating')).toBe('2.50')
    })

    test('should handle values with commas', () => {
      expect(wrapper.vm.formatStatValue('1,500', 'value')).toBe('1500')
      expect(wrapper.vm.formatStatValue('10,000', 'value')).toBe('10,000')
    })

    test('should return original string for non-numeric values', () => {
      expect(wrapper.vm.formatStatValue('N/A', 'position')).toBe('N/A')
      expect(wrapper.vm.formatStatValue('Unknown', 'status')).toBe('Unknown')
    })

    test('should handle edge cases', () => {
      expect(wrapper.vm.formatStatValue(0, 'goals')).toBe('0')
      expect(wrapper.vm.formatStatValue('0', 'goals')).toBe('0')
      expect(wrapper.vm.formatStatValue(0.0, 'rating')).toBe('0')
    })
  })

  describe('component props', () => {
    test('should handle different stat configurations', () => {
      const customStat = { name: 'Assists', key: 'assists' }
      const wrapper = createWrapper({ stat: customStat })
      expect(wrapper.text()).toContain('Assists')
    })

    test('should handle players with missing attributes', () => {
      const playersWithMissingData = [
        { id: 1, name: 'Player 1', club: 'Club 1', attributes: {} },
        { id: 2, name: 'Player 2', club: 'Club 2', attributes: { goals: null } },
      ]
      const wrapper = createWrapper({ players: playersWithMissingData })
      expect(wrapper.text()).toContain('N/A')
    })

    test('should handle players without id', () => {
      const playersWithoutId = [
        { name: 'Player 1', club: 'Club 1', attributes: { goals: 5 } },
        { name: 'Player 2', club: 'Club 2', attributes: { goals: 3 } },
      ]
      const wrapper = createWrapper({ players: playersWithoutId })
      expect(wrapper.text()).toContain('Player 1')
      expect(wrapper.text()).toContain('Player 2')
    })
  })
})
