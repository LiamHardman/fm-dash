import { zstdCompressSync } from 'node:zlib'
import { describe, expect, it } from 'vitest'
import { bufferSource, listFrames, SaveFile } from './container.js'
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

describe('unlisted frames', () => {
  const le32 = (v) => [v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, v >>> 24]
  const blockHeader = (size, type, last) => {
    const v = (size << 3) | (type << 1) | (last ? 1 : 0)
    return [v & 0xff, (v >> 8) & 0xff, (v >> 16) & 0xff]
  }
  // Single-segment frame with a 1-byte content size; blocks are [type, payload, size?].
  const frame = (blocks, { checksum = false } = {}) => {
    const out = [...le32(0xfd2fb528), 0x20 | (checksum ? 0x04 : 0), 0]
    blocks.forEach(([type, payload, size], i) => {
      out.push(...blockHeader(size ?? payload.length, type, i === blocks.length - 1), ...payload)
    })
    if (checksum) out.push(1, 2, 3, 4)
    return new Uint8Array(out)
  }
  const skippable = (payload) =>
    new Uint8Array([...le32(0x184d2a53), ...le32(payload.length), ...payload])

  it('lists frames in the gap after an entry by their headers', async () => {
    const raw = frame([[0, [1, 2, 3, 4, 5]]])
    const rle = frame([[1, [7], 300]]) // RLE: one stored byte, regenerated 300 times
    const multi = frame(
      [
        [0, [1, 2]],
        [1, [9], 40],
        [0, [3]],
      ],
      { checksum: true }
    )
    const real = new Uint8Array(zstdCompressSync(new Uint8Array(4000).map((_, i) => i % 7)))
    const file = buildSave([
      ['meta', bytes(1)],
      ['before', bytes(2, 2), [raw, skippable([0, 0, 0]), rle, multi, real]],
      ['after', bytes(3)],
    ])
    const save = await SaveFile.open(bufferSource(file))
    const [start, end] = save.regionAfter('before')
    const frames = await save.unlistedFrames('before')
    expect(frames.map((f) => f.compressedSize)).toEqual([
      raw.length,
      rle.length,
      multi.length,
      real.length,
    ])
    expect(frames[0].offset).toBe(start)
    expect(frames.at(-1).offset + frames.at(-1).compressedSize).toBe(end)
    expect(await save.readFrame(frames[3])).toEqual(new Uint8Array(4000).map((_, i) => i % 7))
  })

  it('reaches the trailer when the entry is the last one, and is empty without a gap', async () => {
    const real = new Uint8Array(zstdCompressSync(bytes(5, 5, 5)))
    const save = await SaveFile.open(
      bufferSource(
        buildSave([
          ['meta', bytes(1)],
          ['last', bytes(2), [real]],
        ])
      )
    )
    expect(await save.unlistedFrames('meta')).toEqual([])
    expect((await save.unlistedFrames('last')).map((f) => f.compressedSize)).toEqual([real.length])
  })

  it('rejects a gap that does not hold zstd frames', async () => {
    const save = await SaveFile.open(
      bufferSource(
        buildSave([
          ['meta', bytes(1), [bytes(1, 2, 3, 4, 5, 6, 7, 8)]],
          ['x', bytes(1)],
        ])
      )
    )
    await expect(save.unlistedFrames('meta')).rejects.toThrow(/no zstd frame/)
    await expect(save.unlistedFrames('nope')).rejects.toThrow(/no "nope" entry/)
  })

  it('works through the windowed reader on a plain buffer region', async () => {
    const frames = [frame([[0, new Array(70_000).fill(1)]]), frame([[1, [2], 5]])]
    const all = new Uint8Array(frames.reduce((n, f) => n + f.length, 0))
    let at = 0
    for (const f of frames) {
      all.set(f, at)
      at += f.length
    }
    expect(await listFrames(bufferSource(all), 0, all.length)).toEqual([
      { offset: 0, compressedSize: frames[0].length },
      { offset: frames[0].length, compressedSize: frames[1].length },
    ])
  })
})
