import { expect, it } from 'vitest'
import cliFormations from '../api/public/cli_formations.json'
import cliWeights from '../api/public/cli_weight_presets.json'
import squadFixture from '../api/testdata/cli/squad.json'
import { computeSquadComposition } from '../composables/useBestXI.js'
import { PRESET_WEIGHT_PROFILES } from '../stores/weightsStore.js'
import { formations } from './formations.js'

it('keeps CLI formation and weight assets aligned with the GUI', () => {
  expect(cliFormations).toEqual(formations)
  expect(cliWeights).toEqual(PRESET_WEIGHT_PROFILES)
})

it('keeps the Go squad parity fixture aligned with the GUI algorithm', () => {
  expect(
    computeSquadComposition(squadFixture.players, [{ positions: squadFixture.slots }])
  ).toEqual(squadFixture.expected)
})
