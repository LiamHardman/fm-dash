import { computed, onScopeDispose, ref, shallowRef, toValue, watch } from 'vue'
import { comparePlayers, projectSortRows, sortPlayerIndices } from '../utils/playerSorting'
import { useOptimizedSorting } from './useVirtualScrolling'

export function usePlayerSorting(
  players,
  specification,
  limit = 1000,
  workerFactory = () =>
    new Worker(new URL('../workers/playerSortWorker.js', import.meta.url), { type: 'module' })
) {
  const sortedPlayers = shallowRef([])
  const isAsyncSorting = ref(false)
  const totalSortedCount = computed(() => toValue(players).length)
  const isSliced = computed(() => totalSortedCount.value > limit)
  const { sortLargeArray } = useOptimizedSorting()
  let worker,
    loadedSource,
    loadGeneration = 0,
    request = 0,
    latest = 0,
    pending,
    running = false,
    disposed = false
  const jobs = new Map()

  function stopWorker() {
    worker?.terminate()
    worker = null
    loadedSource = null
    for (const job of jobs.values()) {
      clearTimeout(job.timer)
      job.reject(new Error('Worker stopped'))
    }
    jobs.clear()
  }
  function workerSort(source, spec) {
    if (!worker) {
      worker = workerFactory()
      worker.onmessage = ({ data }) => {
        const job = jobs.get(data.id)
        if (!job) return
        jobs.delete(data.id)
        clearTimeout(job.timer)
        if (data.error) job.reject(new Error(data.error))
        else job.resolve(data.indices)
      }
      worker.onerror = stopWorker
    }
    if (loadedSource !== source) {
      loadedSource = source
      loadGeneration++
      worker.postMessage({
        type: 'load',
        rows: projectSortRows(source),
        generation: loadGeneration,
      })
    }
    const id = ++request
    return new Promise((resolve, reject) => {
      const timer = setTimeout(stopWorker, 15000)
      jobs.set(id, {
        resolve: (indices) => resolve(Array.from(indices, (index) => source[index])),
        reject,
        timer,
      })
      try {
        worker.postMessage({ type: 'sort', generation: loadGeneration, id, spec })
      } catch (error) {
        clearTimeout(timer)
        jobs.delete(id)
        reject(error)
        stopWorker()
      }
    })
  }
  async function drain() {
    running = true
    while (pending && !disposed) {
      const job = pending
      pending = null
      let result
      try {
        if (job.source.length < 2000)
          result = sortPlayerIndices(job.source, job.spec).map((index) => job.source[index])
        else result = await workerSort(job.source, job.spec)
      } catch {
        if (disposed) break
        stopWorker()
        result = await sortLargeArray([...job.source], job.spec.field, job.spec.direction, (a, b) =>
          comparePlayers(a, b, job.spec)
        )
      }
      if (!disposed && job.id === latest) sortedPlayers.value = result.slice(0, limit)
    }
    running = false
    if (!disposed) isAsyncSorting.value = false
  }
  watch(
    [() => toValue(players), () => toValue(specification)],
    ([source, spec], previous) => {
      const id = ++latest
      if (source !== previous?.[0]) sortedPlayers.value = source.slice(0, limit)
      pending = { source, spec: { ...spec }, id }
      isAsyncSorting.value = true
      if (!running) void drain()
    },
    { immediate: true }
  )
  onScopeDispose(() => {
    disposed = true
    pending = null
    latest++
    stopWorker()
  })
  return { sortedPlayers, isAsyncSorting, totalSortedCount, isSliced }
}
