import { describe, expect, it } from 'vitest'
import { bufferSource, SaveFile } from './container.js'
import { buildSave } from './testing/buildSave.js'

const bytes = (...values) => new Uint8Array(values)

describe('SaveFile', () => {
  it('reads the directory, groups and a named section', async () => {
    const gameDb = new Uint8Array(5000).map((_, i) => i % 251)
    const file = buildSave(
      [
        ['meta', bytes(1, 2, 3)],
        ['game_db', gameDb],
      ],
      [['comp_1', bytes(9, 9)]],
      'My Save'
    )
    const save = await SaveFile.open(bufferSource(file))

    expect(save.saveName).toBe('My Save')
    expect(save.sectionNames()).toEqual(['game_db', 'meta'])
    expect(save.directory.groups).toHaveLength(1)
    expect(save.directory.groups[0].name).toBe('rgman')
    expect(save.directory.groups[0].entries[0].name).toBe('comp_1')
    expect(await save.readSection('game_db')).toEqual(gameDb)
    expect(await save.readSection('meta')).toEqual(bytes(1, 2, 3))
  })

  it('rejects a file without the fmf magic', async () => {
    await expect(SaveFile.open(bufferSource(new Uint8Array(64)))).rejects.toMatchObject({
      kind: 'not_fm_save',
    })
  })

  it('rejects a truncated file', async () => {
    await expect(SaveFile.open(bufferSource(bytes(2, 1)))).rejects.toMatchObject({
      kind: 'corrupt',
    })
  })

  it('reports a missing section', async () => {
    const save = await SaveFile.open(bufferSource(buildSave([['meta', bytes(1)]])))
    await expect(save.readSection('game_db')).rejects.toThrow(/no "game_db" section/)
  })

  it('rejects a directory entry that points past the trailer', async () => {
    const file = buildSave([['meta', bytes(1, 2, 3, 4, 5, 6, 7, 8)]])
    // Corrupt the trailer pointer so the frame area ends before the frame does.
    const view = new DataView(file.buffer)
    view.setBigUint64(9, 20n, true)
    await expect(SaveFile.open(bufferSource(file))).rejects.toMatchObject({ kind: 'corrupt' })
  })
})
