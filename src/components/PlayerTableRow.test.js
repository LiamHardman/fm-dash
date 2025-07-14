import { mount } from '@vue/test-utils'
import { Quasar } from 'quasar'
import { describe, expect, it, vi } from 'vitest'
import PlayerTableRow from './PlayerTableRow.vue'

// Mock Quasar
const mockQuasar = {
  dark: {
    isActive: false
  }
}

// Global test config
const globalConfig = {
  global: {
    plugins: [Quasar],
    mocks: {
      $q: mockQuasar
    }
  }
}

describe('PlayerTableRow', () => {
  const mockPlayer = {
    name: 'John Doe',
    age: 25,
    club: 'Test FC',
    position: 'CAM',
    rating: 85,
    transferValue: 5000000,
    wage: 50000
  }

  const mockColumns = [
    { name: 'name', label: 'Name', type: 'text' },
    { name: 'age', label: 'Age', type: 'number' },
    { name: 'club', label: 'Club', type: 'text' },
    { name: 'rating', label: 'Rating', type: 'rating' },
    { name: 'transferValue', label: 'Value', type: 'currency' }
  ]

  const mockGetDisplayValue = vi.fn((player, col) => {
    return player[col.name]
  })

  const mockFormatCurrency = vi.fn(value => {
    return `$${value?.toLocaleString() || '0'}`
  })

  const mockGetRatingClass = vi.fn(rating => {
    if (rating >= 80) return 'rating-excellent'
    if (rating >= 70) return 'rating-good'
    return 'rating-average'
  })

  const defaultProps = {
    player: mockPlayer,
    columns: mockColumns,
    currencySymbol: '$',
    isGoalkeeperView: false,
    getDisplayValue: mockGetDisplayValue,
    formatCurrency: mockFormatCurrency,
    getRatingClass: mockGetRatingClass
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders all table cells for columns', () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const cells = wrapper.findAll('td')
    expect(cells).toHaveLength(5)
  })

  it('displays correct values in cells', () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const cells = wrapper.findAll('td')

    // The displayValue function should be called when rendering
    expect(cells[0].text()).toContain('John Doe') // Name
    expect(cells[1].text()).toContain('25') // Age
    expect(cells[2].text()).toContain('Test FC') // Club
  })

  it('renders club as clickable link', () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const clubCell = wrapper.findAll('td')[2] // Club column
    const clubLink = clubCell.find('.club-link a')

    expect(clubLink.exists()).toBe(true)
    expect(clubLink.text()).toBe('Test FC')
  })

  it('emits team-selected when club link is clicked', async () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const clubLink = wrapper.find('.club-link a')
    await clubLink.trigger('click')

    expect(wrapper.emitted('team-selected')).toBeTruthy()
    expect(wrapper.emitted('team-selected')[0]).toEqual(['Test FC'])
  })

  it('formats currency values correctly', () => {
    mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    // The formatCurrency function should be called
    expect(mockFormatCurrency).toHaveBeenCalled()
  })

  it('applies rating class to rating columns', () => {
    mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    // The getRatingClass function should be called
    expect(mockGetRatingClass).toHaveBeenCalled()
  })

  it('emits player-selected when row is clicked', async () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const row = wrapper.find('tr')
    await row.trigger('click')

    expect(wrapper.emitted('player-selected')).toBeTruthy()
    expect(wrapper.emitted('player-selected')[0]).toEqual([mockPlayer])
  })

  it('emits context-menu when right-clicked', async () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const row = wrapper.find('tr')
    await row.trigger('contextmenu')

    expect(wrapper.emitted('context-menu')).toBeTruthy()
    expect(wrapper.emitted('context-menu')[0][0]).toEqual(mockPlayer)
  })

  it('applies goalkeeper-row class for goalkeepers', () => {
    const gkPlayer = {
      ...mockPlayer,
      position: 'GK'
    }

    const wrapper = mount(PlayerTableRow, {
      props: {
        ...defaultProps,
        player: gkPlayer
      },
      ...globalConfig
    })

    const row = wrapper.find('tr')
    expect(row.classes()).toContain('goalkeeper-row')
  })

  it('applies text-right class to number and rating columns', () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const cells = wrapper.findAll('td')

    // Name column (text type) should not have text-right
    expect(cells[0].classes()).not.toContain('text-right')

    // Age column (number type) should have text-right
    expect(cells[1].classes()).toContain('text-right')

    // Rating column (rating type) should have text-right
    expect(cells[3].classes()).toContain('text-right')
  })

  it('stops propagation when club link is clicked', async () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const clubLink = wrapper.find('.club-link a')
    await clubLink.trigger('click')

    // Should emit team-selected but not player-selected due to stop propagation
    expect(wrapper.emitted('team-selected')).toBeTruthy()
    expect(wrapper.emitted('player-selected')).toBeFalsy()
  })

  it('renders regular text for non-special column types', () => {
    const wrapper = mount(PlayerTableRow, {
      props: defaultProps,
      ...globalConfig
    })

    const nameCell = wrapper.findAll('td')[0] // Name column
    expect(nameCell.text()).toBe('John Doe')
    expect(nameCell.find('.club-link').exists()).toBe(false)
  })

  it('handles missing player data gracefully', () => {
    const emptyPlayer = {}

    const wrapper = mount(PlayerTableRow, {
      props: {
        ...defaultProps,
        player: emptyPlayer
      },
      ...globalConfig
    })

    expect(wrapper.findAll('td')).toHaveLength(5)
    // Should not throw errors and still render cells
  })

  it('uses correct currency symbol', () => {
    const wrapper = mount(PlayerTableRow, {
      props: {
        ...defaultProps,
        currencySymbol: '€'
      },
      ...globalConfig
    })

    expect(wrapper.props('currencySymbol')).toBe('€')
  })

  it('handles isGoalkeeperView prop', () => {
    const wrapper = mount(PlayerTableRow, {
      props: {
        ...defaultProps,
        isGoalkeeperView: true
      },
      ...globalConfig
    })

    expect(wrapper.props('isGoalkeeperView')).toBe(true)
  })
})
