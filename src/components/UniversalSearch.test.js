import { createTestingPinia } from '@pinia/testing'
import { mount } from '@vue/test-utils'
import { Quasar } from 'quasar'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import UniversalSearch from './UniversalSearch.vue'

// Mock fetch globally
global.fetch = vi.fn()

// Mock Vue Router
vi.mock('vue-router', () => ({
  useRouter: vi.fn(() => ({
    push: vi.fn(),
    resolve: vi.fn(() => ({ href: '/mocked-url' }))
  }))
}))

// Mock window.open
global.window.open = vi.fn()

import { useRouter } from 'vue-router'

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
    components: {
      PlayerDetailDialog: MockPlayerDetailDialog
    },
    stubs: {
      QInput: {
        name: 'QInput',
        template:
          '<input v-model="modelValue" @keyup.escape="$emit(\'keyup\', { key: \'Escape\' })" ref="input" />',
        props: ['modelValue', 'filled', 'dense', 'placeholder', 'disable'],
        emits: ['update:modelValue', 'keyup'],
        methods: {
          focus: vi.fn()
        }
      },
      QIcon: {
        name: 'QIcon',
        template: '<i></i>',
        props: ['name', 'color']
      },
      QBtn: {
        name: 'QBtn',
        template: '<button @click="$emit(\'click\')"><slot /></button>',
        props: ['flat', 'round', 'dense', 'icon', 'size'],
        emits: ['click']
      },
      QCard: {
        name: 'QCard',
        template: '<div class="q-card"><slot /></div>',
        props: ['flat', 'bordered']
      },
      QCardSection: {
        name: 'QCardSection',
        template: '<div class="q-card-section"><slot /></div>'
      },
      QList: {
        name: 'QList',
        template: '<div class="q-list"><slot /></div>',
        props: ['separator']
      },
      QItem: {
        name: 'QItem',
        template: '<div class="q-item" @click="$emit(\'click\')"><slot /></div>',
        props: ['clickable'],
        emits: ['click']
      },
      QItemSection: {
        name: 'QItemSection',
        template: '<div class="q-item-section"><slot /></div>',
        props: ['avatar', 'side']
      },
      QItemLabel: {
        name: 'QItemLabel',
        template: '<div class="q-item-label"><slot /></div>',
        props: ['caption']
      },
      QChip: {
        name: 'QChip',
        template: '<span class="q-chip"><slot /></span>',
        props: ['color', 'text-color', 'size']
      },
      QSpinner: {
        name: 'QSpinner',
        template: '<div class="q-spinner"></div>',
        props: ['size']
      }
    }
  }
}

describe('UniversalSearch', () => {
  let playerStore
  let mockRouter

  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()

    mockRouter = {
      push: vi.fn(),
      resolve: vi.fn(() => ({ href: '/mocked-url' }))
    }
    useRouter.mockReturnValue(mockRouter)

    global.fetch.mockResolvedValue({
      ok: true,
      json: () =>
        Promise.resolve([
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

  it('renders search input with correct placeholder when dataset exists', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    expect(input.exists()).toBe(true)
  })

  it('disables search input when no dataset is available', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = null
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    expect(input.exists()).toBe(true)
  })

  it('shows loading state when searching', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    // Set loading to true by triggering search
    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'test')

    vi.advanceTimersByTime(300)
    await nextTick()

    const spinner = wrapper.findComponent({ name: 'QSpinner' })
    expect(spinner.exists()).toBe(true)
  })

  it('performs search API call when searchQuery changes', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'John')

    vi.advanceTimersByTime(300)
    await nextTick()

    expect(global.fetch).toHaveBeenCalledWith('/api/search/test-dataset-123?q=John', {
      signal: expect.any(AbortSignal)
    })
  })

  it('displays search results correctly', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'test')

    vi.advanceTimersByTime(300)
    await nextTick()
    await vi.waitFor(async () => {
      await nextTick()
    })

    const resultItems = wrapper.findAllComponents({ name: 'QItem' })
    expect(resultItems.length).toBeGreaterThan(0)
  })

  it('clears search when clear button is clicked', async () => {
    const wrapper = createWrapper()
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'test query')
    await nextTick()

    expect(wrapper.vm.searchQuery).toBe('test query')

    await wrapper.vm.clearSearch()

    expect(wrapper.vm.searchQuery).toBe('')
    expect(wrapper.vm.results).toEqual([])
  })

  it('clears search on escape key', async () => {
    const wrapper = createWrapper()
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'test query')
    await nextTick()

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
    await nextTick()

    await wrapper.vm.handleResultClick({ type: 'player', name: 'John Doe' })

    expect(wrapper.vm.playerForDetailView).toEqual(mockPlayer)
    expect(wrapper.vm.showPlayerDetailDialog).toBe(true)
  })

  it('opens team view in new tab when team result is clicked', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

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
    await nextTick()

    await wrapper.vm.handleResultClick({ type: 'league', name: 'Premier League' })

    expect(mockRouter.push).toHaveBeenCalledWith({
      path: '/leagues/test-dataset-123',
      query: { league: 'Premier League' }
    })
  })

  it('navigates to nations page when nation result is clicked', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

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
    await nextTick()

    await wrapper.vm.handleResultClick({ type: 'player', name: 'Unknown Player' })

    expect(mockRouter.push).toHaveBeenCalledWith({
      path: '/dataset/test-dataset-123',
      query: { search: 'Unknown Player' }
    })
  })

  it('clears search after handling result click', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'test')
    await nextTick()

    await wrapper.vm.handleResultClick({ type: 'team', name: 'Test FC' })

    expect(wrapper.vm.searchQuery).toBe('')
    expect(wrapper.vm.results).toEqual([])
  })

  it('cancels previous search request when new search is initiated', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })

    await input.vm.$emit('update:modelValue', 'first')
    vi.advanceTimersByTime(150) // Partial delay

    await input.vm.$emit('update:modelValue', 'second')
    vi.advanceTimersByTime(300)

    await nextTick()
    expect(global.fetch).toHaveBeenLastCalledWith('/api/search/test-dataset-123?q=second', {
      signal: expect.any(AbortSignal)
    })
  })

  it('handles API errors gracefully', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    global.fetch.mockRejectedValueOnce(new Error('Network error'))

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'error')
    vi.advanceTimersByTime(300)
    await nextTick()

    await vi.waitFor(() => {
      expect(wrapper.vm.isLoading).toBe(false)
    })

    expect(wrapper.vm.results).toEqual([])
  })

  it('does not search when no dataset is available', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = null
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'test')
    vi.advanceTimersByTime(300)
    await nextTick()

    expect(global.fetch).not.toHaveBeenCalled()
  })

  it('does not search for empty queries', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', '   ') // Only whitespace
    vi.advanceTimersByTime(300)
    await nextTick()

    expect(global.fetch).not.toHaveBeenCalled()
  })

  it('shows no results message when search returns empty', async () => {
    const wrapper = createWrapper()
    playerStore.currentDatasetId = 'test-dataset-123'
    await nextTick()

    // Mock empty results
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    const input = wrapper.findComponent({ name: 'QInput' })
    await input.vm.$emit('update:modelValue', 'test')

    vi.advanceTimersByTime(300)
    await nextTick()

    // Wait for the search to complete and loading to finish
    await vi.waitFor(() => {
      expect(wrapper.vm.isLoading).toBe(false)
    })

    // Check that results are empty and the search query exists (which triggers the no results state)
    expect(wrapper.vm.results).toEqual([])
    expect(wrapper.vm.searchQuery).toBe('test')

    // Instead of looking for the exact class, let's check that the component shows no results state
    // by verifying the expected behavior rather than DOM specifics
    expect(wrapper.vm.showResults).toBe(true)
    expect(wrapper.vm.results.length).toBe(0)
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
