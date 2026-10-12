import { expect, it } from 'vitest'
import { mergeProfileBackup } from './profileBackup'
import { evaluateScenario } from './recruitmentScenario'

it('calculates net scenario budgets and position coverage from manual plans', () => {
  const signing = { datasetId: 'a', playerUid: 1 },
    sale = { datasetId: 'b', playerUid: 2 }
  const result = evaluateScenario(
    {
      feeBudget: 100,
      wageBudget: 10,
      signings: [signing],
      sales: [{ playerRef: sale, fee: 30, wage: 5 }],
    },
    [{ playerRef: signing, targetFee: 80, targetWage: 12 }],
    [{ uid: 2, position: 'GK' }],
    () => ({ position: 'ST' })
  )
  expect(result.feeRemaining).toBe(50)
  expect(result.wageRemaining).toBe(3)
  expect(result.coverage).toEqual([
    { position: 'GK', count: 0 },
    { position: 'ST', count: 1 },
  ])
})

it('flags deleted signing targets rather than presenting complete totals', () => {
  const result = evaluateScenario(
    { feeBudget: 1, wageBudget: 1, signings: [{ datasetId: 'a', playerUid: 1 }], sales: [] },
    [],
    [],
    () => null
  )
  expect(result.missingTargets).toBe(1)
})

it('previews ID conflicts and preserves unrelated profile records', () => {
  const backup = (name) => ({
    version: 1,
    shortlists: { version: 1, lists: [{ id: 'one', name }] },
    savedSearches: { version: 1, searches: [] },
    progressionProjects: { version: 1, projects: [] },
  })
  const current = backup('Existing'),
    incoming = backup('Incoming')
  current.shortlists.lists.push({ id: 'keep', name: 'Retained' })
  expect(mergeProfileBackup(current, incoming).document.shortlists.lists[0].name).toBe('Existing')
  const merged = mergeProfileBackup(current, incoming, true)
  expect(merged.document.shortlists.lists.map((l) => l.name)).toEqual(['Incoming', 'Retained'])
  expect(merged.changes[0].action).toBe('Replace existing')
})
