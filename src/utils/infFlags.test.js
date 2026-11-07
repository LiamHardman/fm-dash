import { expect, it } from 'vitest'
import { ref } from 'vue'
import { usePlayerTableColumns } from '../composables/usePlayerTableColumns'
import { describeInfCode, hasInfFlags, passesInfFilters, splitInfCodes } from './infFlags.js'

it('splits codes on commas and whitespace', () => {
  expect(splitInfCodes('Inj, Loa')).toEqual(['Inj', 'Loa'])
  expect(splitInfCodes('Inj Wnt')).toEqual(['Inj', 'Wnt'])
  expect(splitInfCodes('')).toEqual([])
  expect(splitInfCodes(undefined)).toEqual([])
})

it('describes known codes and falls back for unknown ones', () => {
  expect(describeInfCode('Inj')).toBe('Injured')
  expect(describeInfCode('Loa')).toBe('Listed for loan')
  expect(describeInfCode('FrA')).toBe('Free agent')
  expect(describeInfCode('Xyz')).toBe('FM info flag')
})

it('adds the Inf column only when some player has flags', () => {
  const available = ref(false)
  const { currentColumns } = usePlayerTableColumns(
    false,
    false,
    false,
    'estimated',
    false,
    available
  )
  expect(currentColumns.value.find((c) => c.name === 'inf')).toBeUndefined()
  available.value = true
  const column = currentColumns.value.find((c) => c.name === 'inf')
  expect(column).toMatchObject({ label: 'Inf', field: 'inf', sortable: true, isInfFlags: true })
  expect(column.optional).toBeFalsy() // shown by default
  expect(hasInfFlags([{ inf: '' }, { inf: 'Loa' }])).toBe(true)
  expect(hasInfFlags([{ inf: '' }, {}])).toBe(false)
})

it('filters out injured players and keeps only loan-listed ones', () => {
  const injured = { inf: 'Inj' }
  const listed = { inf: 'Loa' }
  const both = { inf: 'Inj, Loa' }
  const none = {}
  const keep = (filters) =>
    [injured, listed, both, none]
      .filter((p) => passesInfFilters(p, filters))
      .map((p) => p.inf ?? '-')
  expect(keep({})).toEqual(['Inj', 'Loa', 'Inj, Loa', '-'])
  expect(keep({ hideInjured: true })).toEqual(['Loa', '-'])
  expect(keep({ loanListedOnly: true })).toEqual(['Loa', 'Inj, Loa'])
  expect(keep({ hideInjured: true, loanListedOnly: true })).toEqual(['Loa'])
})
