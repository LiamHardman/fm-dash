import { expect, it } from 'vitest'
import { computed, ref } from 'vue'
import { usePlayerTableCells } from './usePlayerTableCells'
import { usePlayerTableColumns } from './usePlayerTableColumns'

it('updates columns and cell values when goalkeeper mode changes without remounting', () => {
  const goalkeeper = ref(false)
  const valueScore = ref(false)
  const columns = usePlayerTableColumns(() => goalkeeper.value, valueScore, ref(false))
  const cells = usePlayerTableCells(goalkeeper, ref('£'), ref(0))
  const names = computed(() => columns.currentColumns.value.map((c) => c.name))
  expect(names.value).toContain('PAC')
  expect(cells.getPlayerValue({ pac: 50, div: 80 }, 'pac')).toBe(50)
  goalkeeper.value = true
  valueScore.value = true
  expect(names.value).toContain('DIV')
  expect(names.value).not.toContain('PAC')
  expect(names.value).toContain('valueScore')
  expect(cells.getPlayerValue({ pac: 50, div: 80 }, 'pac')).toBe(80)
})
