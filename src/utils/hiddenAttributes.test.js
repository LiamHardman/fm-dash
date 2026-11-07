import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { usePlayerTableColumns } from '../composables/usePlayerTableColumns'
import { useUiStore } from '../stores/uiStore'
import {
  activeHiddenFilters,
  defaultHiddenFilters,
  HIDDEN_ATTRIBUTES,
  hasHiddenAttributes,
  hiddenAttributeColorValue,
  hiddenFilterKey,
  passesHiddenFilters,
} from './hiddenAttributes.js'
import { comparePlayers, projectSortRows } from './playerSorting.js'

describe('uiStore.showHiddenAttributes', () => {
  // test-setup.js replaces localStorage with vi.fn() stubs.
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(localStorage.getItem).mockReset()
    vi.mocked(localStorage.setItem).mockReset()
  })

  it('defaults off, persists, and restores', () => {
    const store = useUiStore()
    expect(store.showHiddenAttributes).toBe(false)
    store.toggleHiddenAttributes()
    expect(store.showHiddenAttributes).toBe(true)
    expect(localStorage.setItem).toHaveBeenCalledWith('showHiddenAttributes', 'true')

    setActivePinia(createPinia())
    vi.mocked(localStorage.getItem).mockImplementation((key) =>
      key === 'showHiddenAttributes' ? 'true' : null
    )
    const fresh = useUiStore()
    fresh.initHiddenAttributesDisplay()
    expect(fresh.showHiddenAttributes).toBe(true)
  })

  it('survives a throwing localStorage', () => {
    vi.mocked(localStorage.setItem).mockImplementation(() => {
      throw new Error('blocked')
    })
    vi.mocked(localStorage.getItem).mockImplementation(() => {
      throw new Error('blocked')
    })
    const store = useUiStore()
    expect(() => store.setHiddenAttributesDisplay(true)).not.toThrow()
    expect(store.showHiddenAttributes).toBe(true)
    expect(() => store.initHiddenAttributesDisplay()).not.toThrow()
  })
})

describe('hasHiddenAttributes', () => {
  it('is true only when some player carries values', () => {
    expect(hasHiddenAttributes([{}, { hiddenAttributes: { Consistency: 12 } }])).toBe(true)
    expect(hasHiddenAttributes([{}, { hiddenAttributes: {} }])).toBe(false)
    expect(hasHiddenAttributes([])).toBe(false)
    expect(hasHiddenAttributes(undefined)).toBe(false)
  })
})

describe('hidden filters', () => {
  const player = (hiddenAttributes) => ({ hiddenAttributes })

  it('uses a maximum for the three lower-is-better attributes', () => {
    expect(hiddenFilterKey('Consistency')).toBe('minHiddenConsistency')
    expect(hiddenFilterKey('Injury Proneness')).toBe('maxHiddenInjuryProneness')
    expect(hiddenFilterKey('Dirtiness')).toBe('maxHiddenDirtiness')
    expect(hiddenFilterKey('Controversy')).toBe('maxHiddenControversy')
    const defaults = defaultHiddenFilters()
    expect(defaults.minHiddenImportantMatches).toBe(0)
    expect(defaults.maxHiddenInjuryProneness).toBe(20)
    expect(activeHiddenFilters(defaults)).toEqual([])
  })

  it('applies minimums and maximums', () => {
    const active = activeHiddenFilters({
      ...defaultHiddenFilters(),
      minHiddenConsistency: 14,
      maxHiddenInjuryProneness: 8,
    })
    expect(active).toHaveLength(2)
    expect(passesHiddenFilters(player({ Consistency: 15, 'Injury Proneness': 5 }), active)).toBe(
      true
    )
    expect(passesHiddenFilters(player({ Consistency: 13, 'Injury Proneness': 5 }), active)).toBe(
      false
    )
    expect(passesHiddenFilters(player({ Consistency: 15, 'Injury Proneness': 9 }), active)).toBe(
      false
    )
  })

  it('fails a player missing a value for an active filter', () => {
    const active = activeHiddenFilters({ ...defaultHiddenFilters(), minHiddenAmbition: 10 })
    expect(passesHiddenFilters(player({ Consistency: 15 }), active)).toBe(false)
    expect(passesHiddenFilters({}, active)).toBe(false)
    expect(passesHiddenFilters({}, [])).toBe(true)
  })
})

describe('hiddenAttributeColorValue', () => {
  it('inverts the scale where lower is better', () => {
    expect(hiddenAttributeColorValue('Injury Proneness', 2)).toBe(19)
    expect(hiddenAttributeColorValue('Dirtiness', 20)).toBe(1)
    expect(hiddenAttributeColorValue('Controversy', 5)).toBe(16)
    expect(hiddenAttributeColorValue('Consistency', 5)).toBe(5)
    expect(hiddenAttributeColorValue('Consistency', undefined)).toBeUndefined()
  })
})

describe('hidden attribute table columns', () => {
  it('exist only when available, all optional', () => {
    const available = ref(false)
    const { currentColumns } = usePlayerTableColumns(false, false, false, 'save', available)
    const hidden = () => currentColumns.value.filter((c) => c.name.startsWith('hidden:'))
    expect(hidden()).toHaveLength(0)
    available.value = true
    expect(hidden().map((c) => c.label)).toEqual(HIDDEN_ATTRIBUTES)
    expect(hidden().every((c) => c.optional && c.sortable)).toBe(true)
  })

  it('sort players without a value last in both directions', () => {
    const rows = projectSortRows([
      { name: 'a', hiddenAttributes: { Consistency: 12 } },
      { name: 'b' },
      { name: 'c', hiddenAttributes: { Consistency: 18 } },
    ])
    const order = (direction) =>
      [0, 1, 2]
        .sort((x, y) =>
          comparePlayers(rows[x], rows[y], { field: 'hidden:Consistency', direction })
        )
        .map((i) => rows[i].name)
    expect(order('desc')).toEqual(['c', 'a', 'b'])
    expect(order('asc')).toEqual(['a', 'c', 'b'])
  })
})
