export function useThresholds(allPlayersData, filters, getNumericValue, getPlayerDivision) {
  const calculateThresholds = () => {
    if (!allPlayersData.value.length) return { minutes: 0, overall: 0 }

    const allAvailableDivisions = [
      ...new Set(allPlayersData.value.map((p) => getPlayerDivision(p)).filter(Boolean)),
    ]

    const filteredByDivisionAndPosition = allPlayersData.value.filter((player) => {
      const division = getPlayerDivision(player)
      const divisionMatch =
        filters.selectedDivisions.value.length === 0 ||
        filters.selectedDivisions.value.length === allAvailableDivisions.length ||
        filters.selectedDivisions.value.includes(division)

      const positionMatch =
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

      return divisionMatch && positionMatch
    })

    const sortedOverall = [...filteredByDivisionAndPosition]
      .map((p) => p.Overall || 0)
      .sort((a, b) => b - a)

    const overallTargetIndex = Math.min(249, sortedOverall.length - 1)
    const overallThreshold = sortedOverall[overallTargetIndex] || 0

    const playersAfterOverallFilter = filteredByDivisionAndPosition.filter(
      (player) => (player.overall || 0) >= overallThreshold
    )

    const sortedMinutes = [...playersAfterOverallFilter]
      .map((p) => getNumericValue(p.attributes?.Mins) || 0)
      .sort((a, b) => b - a)

    const minutesTargetIndex = Math.min(99, sortedMinutes.length - 1)
    const rawMinutesThreshold = sortedMinutes[minutesTargetIndex] || 0
    const minutesThreshold = Math.round(rawMinutesThreshold / 50) * 50

    return { minutes: minutesThreshold, overall: overallThreshold }
  }

  return { calculateThresholds }
}
