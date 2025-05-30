import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { Quasar } from 'quasar'
import { createTestingPinia } from '@pinia/testing'
import UniversalSearch from './UniversalSearch.vue'
import { usePlayerStore } from '../stores/playerStore'

// Mock fetch globally
global.fetch = vi.fn()

// Mock Vue Router
const mockRouter = {
  push: vi.fn(),
  resolve: vi.fn(() => ({ href: '/mocked-url' }))
}

// Mock window.open
global.window.open = vi.fn()

// Mock PlayerDetailDialog component
const MockPlayerDetailDialog = {
  name: 'PlayerDetailDialog',
  template: '<div class="mock-player-detail-dialog"></div>',
  props: ['player', 'show', 'currency-symbol', 'dataset-id'],
  emits: ['close']
}

const globalConfig = {
  global: {
    plugins: [
      Quasar,
      createTestingPinia({
        createSpy: vi.fn
      })
    ],
    mocks: {
      $router: mockRouter
    },
    stubs: {
      QInput: true,
      QIcon: true,
      QBtn: true,
      QCard: true,
      QCardSection: true,
      QList: true,
      QItem: true,
      QItemSection: true,
      QItemLabel: true,
      QChip: true,
      QSpinner: true,
      PlayerDetailDialog: MockPlayerDetailDialog
    }
  }
}

describe('UniversalSearch', () => {
  let playerStore

  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    
    // Setup mock fetch responses
    global.fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([
        { type: 'player', id: 1, name: 'John Doe', description: 'Forward at Test FC' },
        { type: 'team', id: 2, name: 'Test FC', description: 'Football Club' },
        { type: 'league', id: 3, name: 'Test League', description: 'Premier League' },
        { type: 'nation', id: 4, name: 'England', description: 'National Team' }
      ])
    })
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  const createWrapper = (options = {}) => {
    const wrapper = mount(UniversalSearch, {
      ...globalConfig,
      ...options
    })
    
    playerStore = usePlayerStore()
    return wrapper
  }

  it('renders search input with correct placeholder when dataset exists', () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    const input = wrapper.findComponent({ name: 'QInput' })
    expect(input.exists()).toBe(true)
  })

  it('disables search input when no dataset is available', () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = null
    
    const input = wrapper.findComponent({ name: 'QInput' })
    expect(input.exists()).toBe(true)
  })

  it('shows loading state when searching', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    // Set loading to true
    await wrapper.setData({ isLoading: true, searchQuery: 'test' })
    
    const spinner = wrapper.findComponent({ name: 'QSpinner' })
    expect(spinner.exists()).toBe(true)
  })

  it('performs search API call when searchQuery changes', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    await wrapper.setData({ searchQuery: 'John' })
    
    // Fast-forward timers to trigger debounced search
    vi.advanceTimersByTime(300)
    await nextTick()
    
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/search/test-dataset-123?q=John',
      { signal: expect.any(AbortSignal) }
    )
  })

  it('displays search results correctly', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    await wrapper.setData({ 
      searchQuery: 'test',
      results: [
        { type: 'player', id: 1, name: 'John Doe', description: 'Forward' },
        { type: 'team', id: 2, name: 'Test FC', description: 'Football Club' }
      ]
    })
    
    const resultItems = wrapper.findAllComponents({ name: 'QItem' })
    expect(resultItems.length).toBe(2)
  })

  it('clears search when clear button is clicked', async () => {
    const wrapper = createWrapper()
    
    await wrapper.setData({ 
      searchQuery: 'test query',
      results: [{ type: 'player', id: 1, name: 'Test' }]
    })
    
    const clearButton = wrapper.findComponent({ name: 'QBtn' })
    await clearButton.vm.$emit('click')
    
    expect(wrapper.vm.searchQuery).toBe('')
    expect(wrapper.vm.results).toEqual([])
  })

  it('clears search on escape key', async () => {
    const wrapper = createWrapper()
    
    await wrapper.setData({ searchQuery: 'test query' })
    
    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('keyup', { key: 'Escape' })
    
    expect(wrapper.vm.searchQuery).toBe('')
  })

  it('returns correct icon for different result types', () => {
    const wrapper = createWrapper()
    
    expect(wrapper.vm.getResultIcon('player')).toBe('person')
    expect(wrapper.vm.getResultIcon('team')).toBe('groups')
    expect(wrapper.vm.getResultIcon('league')).toBe('emoji_events')
    expect(wrapper.vm.getResultIcon('nation')).toBe('flag')
    expect(wrapper.vm.getResultIcon('unknown')).toBe('search')
  })

  it('returns correct color for different result types', () => {
    const wrapper = createWrapper()
    
    expect(wrapper.vm.getResultColor('player')).toBe('blue')
    expect(wrapper.vm.getResultColor('team')).toBe('green')
    expect(wrapper.vm.getResultColor('league')).toBe('orange')
    expect(wrapper.vm.getResultColor('nation')).toBe('red')
    expect(wrapper.vm.getResultColor('unknown')).toBe('grey')
  })

  it('opens player detail dialog when player result is clicked', async () => {
    const wrapper = createWrapper()
    const mockPlayer = { name: 'John Doe', age: 25 }
    playerStore.allPlayers = [mockPlayer]
    
    await wrapper.vm.handleResultClick({ type: 'player', name: 'John Doe' })
    
    expect(wrapper.vm.playerForDetailView).toEqual(mockPlayer)
    expect(wrapper.vm.showPlayerDetailDialog).toBe(true)
  })

  it('opens team view in new tab when team result is clicked', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    await wrapper.vm.handleResultClick({ type: 'team', name: 'Test FC' })
    
    expect(mockRouter.resolve).toHaveBeenCalledWith({
      path: '/team-view',
      query: {
        datasetId: 'test-dataset-123',
        team: 'Test FC'
      }
    })
    expect(global.window.open).toHaveBeenCalledWith('/mocked-url', '_blank')
  })

  it('navigates to leagues page when league result is clicked', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    await wrapper.vm.handleResultClick({ type: 'league', name: 'Premier League' })
    
    expect(mockRouter.push).toHaveBeenCalledWith({
      path: '/leagues/test-dataset-123',
      query: { league: 'Premier League' }
    })
  })

  it('navigates to nations page when nation result is clicked', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    await wrapper.vm.handleResultClick({ type: 'nation', name: 'England' })
    
    expect(mockRouter.push).toHaveBeenCalledWith({
      path: '/nations/test-dataset-123',
      query: { nation: 'England' }
    })
  })

  it('falls back to dataset page for unknown player', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    playerStore.allPlayers = [] // No players found
    
    await wrapper.vm.handleResultClick({ type: 'player', name: 'Unknown Player' })
    
    expect(mockRouter.push).toHaveBeenCalledWith({
      path: '/dataset/test-dataset-123',
      query: { search: 'Unknown Player' }
    })
  })

  it('clears search after handling result click', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    await wrapper.setData({ searchQuery: 'test', results: [] })
    await wrapper.vm.handleResultClick({ type: 'team', name: 'Test FC' })
    
    expect(wrapper.vm.searchQuery).toBe('')
    expect(wrapper.vm.results).toEqual([])
  })

  it('cancels previous search request when new search is initiated', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    // Start first search
    await wrapper.setData({ searchQuery: 'first' })
    vi.advanceTimersByTime(150) // Partial delay
    
    // Start second search before first completes
    await wrapper.setData({ searchQuery: 'second' })
    vi.advanceTimersByTime(300)
    
    // Should only call API for the latest search
    await nextTick()
    expect(global.fetch).toHaveBeenLastCalledWith(
      '/api/search/test-dataset-123?q=second',
      { signal: expect.any(AbortSignal) }
    )
  })

  it('handles API errors gracefully', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    // Mock fetch to reject
    global.fetch.mockRejectedValueOnce(new Error('Network error'))
    
    await wrapper.setData({ searchQuery: 'error' })
    vi.advanceTimersByTime(300)
    await nextTick()
    
    expect(wrapper.vm.results).toEqual([])
    expect(wrapper.vm.isLoading).toBe(false)
  })

  it('does not search when no dataset is available', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = null
    
    await wrapper.setData({ searchQuery: 'test' })
    vi.advanceTimersByTime(300)
    await nextTick()
    
    expect(global.fetch).not.toHaveBeenCalled()
  })

  it('does not search for empty queries', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    
    await wrapper.setData({ searchQuery: '   ' }) // Only whitespace
    vi.advanceTimersByTime(300)
    await nextTick()
    
    expect(global.fetch).not.toHaveBeenCalled()
  })

  it('shows no results message when search returns empty', async () => {
    const wrapper = createWrapper()
    
    await wrapper.setData({ 
      searchQuery: 'test',
      results: [],
      isLoading: false
    })
    
    const noResultsMessage = wrapper.find('.text-grey-6')
    expect(noResultsMessage.exists()).toBe(true)
  })

  it('handles currency symbol from store', () => {
    const wrapper = createWrapper()
    playerStore.detectedCurrencySymbol = '€'
    
    expect(wrapper.vm.detectedCurrencySymbol).toBe('€')
  })

  it('uses default currency symbol when none detected', () => {
    const wrapper = createWrapper()
    playerStore.detectedCurrencySymbol = null
    
    expect(wrapper.vm.detectedCurrencySymbol).toBe('$')
  })
}) 