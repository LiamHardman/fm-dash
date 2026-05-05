import { computed, ref, watch } from 'vue'

export function useTopPlayers(filteredPlayers, statsConfig, getNumericValue) {
  const topPlayersByStat = ref({})

  const uniqueStats = computed(() => {
    // statsConfig may contain duplicate keys across categories; de-duplicate by key
    const entries = [...statsConfig.value]
    const byKey = new Map(entries.map((s) => [s.key, s]))
    return Array.from(byKey.values())
  })

  const readStatValue = (player, key) => {
    const raw = player.performanceStatsNumeric?.[key] ?? player.attributes?.[key]
    return getNumericValue(raw)
  }

  const calculateTopPerformers = () => {
    const playersToProcess = filteredPlayers.value
    const results = {}

    for (const stat of uniqueStats.value) {
      const playersWithStat = playersToProcess.filter((player) => {
        const value = readStatValue(player, stat.key)
        return value !== null && value !== undefined && !Number.isNaN(value)
      })

      const top = playersWithStat
        .map((player) => ({ player, value: readStatValue(player, stat.key) }))
        .sort((a, b) => b.value - a.value)
        .slice(0, 10)
        .map((e) => e.player)

      results[stat.key] = top
    }

    topPlayersByStat.value = results
  }

  watch(filteredPlayers, () => calculateTopPerformers(), { deep: true, immediate: true })

  return { topPlayersByStat, calculateTopPerformers }
}
