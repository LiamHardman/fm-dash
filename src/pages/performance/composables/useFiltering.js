import { computed } from 'vue'

export function useFiltering(allPlayersData, filters) {
  const filteredPlayers = computed(() => {
    const allAvailableDivisions = [
      ...new Set(allPlayersData.value.map((p) => filters.getPlayerDivision(p)).filter(Boolean)),
    ]

    return allPlayersData.value.filter((player) => {
      const minutesPlayed = filters.getNumericValue(player.numericAttributes?.Mins) || 0
      const overall = player.overall || 0
      const division = filters.getPlayerDivision(player)

      const matchesMinutes = minutesPlayed >= filters.selectedMinutes.value
      const matchesOverall = overall >= filters.selectedOverall.value
      const matchesDivision =
        filters.selectedDivisions.value.length === 0 ||
        filters.selectedDivisions.value.length === allAvailableDivisions.length ||
        filters.selectedDivisions.value.includes(division)

      const matchesPosition =
        filters.selectedPositions.value.length === 0 ||
        filters.selectedPositions.value.some((selectedPos) => {
          const playerPositions = player.short_positions || player.shortPositions || []
          if (selectedPos === 'Goalkeeper') return playerPositions.includes('GK')
          if (selectedPos === 'Defender')
            return playerPositions.some((pos) => ['DC', 'DR', 'DL', 'WBR', 'WBL'].includes(pos))
          if (selectedPos === 'Midfielder')
            return playerPositions.some((pos) =>
              ['DM', 'MC', 'MR', 'ML', 'AMR', 'AMC', 'AML'].includes(pos)
            )
          if (selectedPos === 'Forward') return playerPositions.includes('ST')

          if (
            [
              'GK',
              'DC',
              'DR',
              'DL',
              'WBR',
              'WBL',
              'DM',
              'MC',
              'MR',
              'ML',
              'AMC',
              'AMR',
              'AML',
              'ST',
            ].includes(selectedPos)
          ) {
            return playerPositions.includes(selectedPos)
          }

          return false
        })

      return matchesMinutes && matchesOverall && matchesDivision && matchesPosition
    })
  })

  return { filteredPlayers }
}
