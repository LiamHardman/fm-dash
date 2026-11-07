import { expect, it } from 'vitest'
import { ref } from 'vue'
import { usePlayerTableColumns } from '../composables/usePlayerTableColumns'
import { caLabel, caSourceOf } from './caSource.js'

it('derives the CA source from any player with a true CA', () => {
  expect(caSourceOf([{ ca: 120 }, { ca: 140, saveCa: 140 }])).toBe('save')
  expect(caSourceOf([{ ca: 120 }, { ca: 90, saveCa: 0 }])).toBe('estimated')
  expect(caSourceOf([])).toBe('estimated')
  expect(caSourceOf(undefined)).toBe('estimated')
})

it('labels true CA "CA" and the estimate "Est. CA"', () => {
  expect(caLabel('save')).toBe('CA')
  expect(caLabel('estimated')).toBe('Est. CA')
})

it('adds the PA column only for save Datasets with CA shown', () => {
  const showCA = ref(true)
  const source = ref('save')
  const { currentColumns } = usePlayerTableColumns(false, false, showCA, source)
  const column = (name) => currentColumns.value.find((c) => c.name === name)

  expect(column('CA').label).toBe('CA')
  expect(column('PA')).toBeDefined()
  expect(currentColumns.value.findIndex((c) => c.name === 'PA')).toBe(
    currentColumns.value.findIndex((c) => c.name === 'CA') + 1
  )

  source.value = 'estimated'
  expect(column('CA').label).toBe('Est. CA')
  expect(column('PA')).toBeUndefined()

  source.value = 'save'
  showCA.value = false
  expect(column('CA')).toBeUndefined()
  expect(column('PA')).toBeUndefined()
})
