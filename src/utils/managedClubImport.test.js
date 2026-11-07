import { describe, expect, it, vi } from 'vitest'
import { setManagedClubIfAbsent } from './managedClubImport.js'

const response = (status) => ({ status, ok: status >= 200 && status < 300 })
const detected = { club: 'PFK Montana', division: 'Vtora Liga' }

describe('setManagedClubIfAbsent', () => {
  it('PUTs the detected club when none is stored (404)', async () => {
    const fetchFn = vi
      .fn()
      .mockResolvedValueOnce(response(404))
      .mockResolvedValueOnce(response(200))
    expect(await setManagedClubIfAbsent('ds1', detected, fetchFn)).toBe(true)
    expect(fetchFn).toHaveBeenCalledTimes(2)
    const [url, init] = fetchFn.mock.calls[1]
    expect(url).toContain('/api/managed-team/ds1')
    expect(init.method).toBe('PUT')
    expect(JSON.parse(init.body)).toEqual(detected)
  })

  it('never overwrites a stored club (200)', async () => {
    const fetchFn = vi.fn().mockResolvedValueOnce(response(200))
    expect(await setManagedClubIfAbsent('ds1', detected, fetchFn)).toBe(false)
    expect(fetchFn).toHaveBeenCalledTimes(1)
  })

  it('swallows a failing PUT or a network error', async () => {
    const failing = vi
      .fn()
      .mockResolvedValueOnce(response(404))
      .mockResolvedValueOnce(response(500))
    await expect(setManagedClubIfAbsent('ds1', detected, failing)).resolves.toBe(false)
    const offline = vi.fn().mockRejectedValue(new Error('offline'))
    await expect(setManagedClubIfAbsent('ds1', detected, offline)).resolves.toBe(false)
  })

  it('does nothing without a detected club', async () => {
    const fetchFn = vi.fn()
    expect(await setManagedClubIfAbsent('ds1', undefined, fetchFn)).toBe(false)
    expect(fetchFn).not.toHaveBeenCalled()
  })
})
