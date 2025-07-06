import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { debounce, throttle } from './debounce.js'

describe('debounce', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should delay function execution', () => {
    const mockFn = vi.fn()
    const debouncedFn = debounce(mockFn, 100)

    debouncedFn()
    expect(mockFn).not.toHaveBeenCalled()

    vi.advanceTimersByTime(50)
    expect(mockFn).not.toHaveBeenCalled()

    vi.advanceTimersByTime(50)
    expect(mockFn).toHaveBeenCalledTimes(1)
  })

  it('should reset delay on subsequent calls', () => {
    const mockFn = vi.fn()
    const debouncedFn = debounce(mockFn, 100)

    debouncedFn()
    vi.advanceTimersByTime(50)
    debouncedFn() // Reset the timer

    vi.advanceTimersByTime(50)
    expect(mockFn).not.toHaveBeenCalled()

    vi.advanceTimersByTime(50)
    expect(mockFn).toHaveBeenCalledTimes(1)
  })

  it('should pass arguments correctly', () => {
    const mockFn = vi.fn()
    const debouncedFn = debounce(mockFn, 100)

    debouncedFn('arg1', 'arg2', 123)
    vi.advanceTimersByTime(100)

    expect(mockFn).toHaveBeenCalledWith('arg1', 'arg2', 123)
  })

  it('should maintain correct this context', () => {
    const obj = {
      value: 'test',
      method: function (arg) {
        return this.value + arg
      }
    }

    const mockFn = vi.fn(obj.method)
    obj.debouncedMethod = debounce(mockFn, 100)

    obj.debouncedMethod(' suffix')
    vi.advanceTimersByTime(100)

    expect(mockFn).toHaveBeenCalled()
  })
})

describe('throttle', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should execute function immediately on first call', () => {
    const mockFn = vi.fn()
    const throttledFn = throttle(mockFn, 100)

    throttledFn()
    expect(mockFn).toHaveBeenCalledTimes(1)
  })

  it('should throttle subsequent calls within delay period', () => {
    const mockFn = vi.fn()
    const throttledFn = throttle(mockFn, 100)

    throttledFn()
    throttledFn()
    throttledFn()

    expect(mockFn).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(100)
    expect(mockFn).toHaveBeenCalledTimes(2)
  })

  it('should allow execution after delay period', () => {
    const mockFn = vi.fn()
    const throttledFn = throttle(mockFn, 100)

    throttledFn()
    expect(mockFn).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(150)
    throttledFn()
    expect(mockFn).toHaveBeenCalledTimes(2)
  })

  it('should pass arguments correctly', () => {
    const mockFn = vi.fn()
    const throttledFn = throttle(mockFn, 100)

    throttledFn('arg1', 'arg2')
    expect(mockFn).toHaveBeenCalledWith('arg1', 'arg2')

    throttledFn('arg3', 'arg4')
    vi.advanceTimersByTime(100)
    expect(mockFn).toHaveBeenLastCalledWith('arg3', 'arg4')
  })
})
