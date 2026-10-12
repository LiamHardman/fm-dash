import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createProfileDocumentClient } from './profileDocumentClient'

const initial = () => ({ version: 1, searches: [] })
const reply = (document, revision = '"1"', status = 200) =>
  new Response(JSON.stringify(document), { status, headers: { ETag: revision } })
beforeEach(() => {
  const cache = new Map()
  vi.stubGlobal('localStorage', {
    getItem: (key) => cache.get(key) ?? null,
    setItem: (key, value) => cache.set(key, value),
    clear: () => cache.clear(),
  })
  vi.stubGlobal('fetch', vi.fn())
})

describe('profile document recovery', () => {
  it('serializes rapid edits and sends the latest draft with the updated revision', async () => {
    fetch.mockResolvedValueOnce(reply(initial()))
    const client = createProfileDocumentClient('/api/test', 'test', initial)
    await client.load()
    let finishFirst
    fetch.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finishFirst = resolve
        })
    )
    fetch.mockResolvedValueOnce(reply({ version: 1, searches: ['new'] }, '"3"'))
    const first = client.save({ version: 1, searches: ['old'] })
    const second = client.save({ version: 1, searches: ['new'] })
    expect(fetch).toHaveBeenCalledTimes(2)
    finishFirst(reply({ version: 1, searches: ['old'] }, '"2"'))
    await Promise.all([first, second])
    expect(fetch.mock.calls[2][1].headers['If-Match']).toBe('"2"')
    expect(JSON.parse(fetch.mock.calls[2][1].body).searches).toEqual(['new'])
    expect(JSON.parse(localStorage.getItem(client.recoveryKey)).dirty).toBe(false)
  })

  it('retains failed edits across reload and retries when the server is unchanged', async () => {
    fetch.mockResolvedValueOnce(reply(initial()))
    const first = createProfileDocumentClient('/api/test', 'test', initial)
    await first.load()
    fetch.mockRejectedValueOnce(new Error('offline'))
    await expect(first.save({ version: 1, searches: ['draft'] })).rejects.toThrow()
    fetch.mockResolvedValueOnce(reply(initial()))
    const reloaded = createProfileDocumentClient('/api/test', 'test', initial)
    expect((await reloaded.load()).searches).toEqual(['draft'])
    expect(reloaded.state.status).toBe('pending')
    fetch
      .mockResolvedValueOnce(reply(initial()))
      .mockResolvedValueOnce(reply({ version: 1, searches: ['draft'] }, '"2"'))
    await reloaded.retry()
    expect(reloaded.state.status).toBe('saved')
  })

  it('never substitutes defaults for an initial storage outage', async () => {
    fetch.mockRejectedValueOnce(new Error('offline'))
    const client = createProfileDocumentClient('/api/test', 'test', initial)
    await expect(client.load()).rejects.toThrow('offline')
    await expect(client.save(initial())).rejects.toThrow('Review')
    expect(fetch).toHaveBeenCalledTimes(1)
  })

  it('preserves a local draft and presents a changed server document as a conflict', async () => {
    localStorage.setItem(
      'test',
      JSON.stringify({
        document: { version: 1, searches: ['draft'] },
        dirty: true,
        revision: '"1"',
      })
    )
    fetch.mockResolvedValueOnce(reply({ version: 1, searches: ['remote'] }, '"2"'))
    const client = createProfileDocumentClient('/api/test', 'test', initial)
    expect((await client.load()).searches).toEqual(['draft'])
    expect(client.state.status).toBe('conflict')
    expect(client.state.conflict.document.searches).toEqual(['remote'])
  })
})
