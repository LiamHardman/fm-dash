export const referenceKey = (ref) => `${ref.datasetId}:${ref.playerUid}`

export function evaluateScenario(scenario, targets, squad, resolvePlayer) {
  const targetByRef = new Map(targets.map((item) => [referenceKey(item.playerRef), item]))
  let fees = 0
  let wages = 0
  let missingTargets = 0
  let unresolvedPlayers = 0
  const coverage = new Map()
  const adjust = (player, delta) => {
    if (!player) {
      unresolvedPlayers++
      return
    }
    const positions = player.parsedPositions?.length
      ? player.parsedPositions
      : [player.position || 'Unknown']
    for (const position of new Set(positions))
      coverage.set(position, (coverage.get(position) || 0) + delta)
  }
  for (const player of squad) adjust(player, 1)
  for (const ref of scenario.signings) {
    const item = targetByRef.get(referenceKey(ref))
    if (!item) {
      missingTargets++
      continue
    }
    fees += Number(item.targetFee || 0)
    wages += Number(item.targetWage || 0)
    adjust(resolvePlayer(ref), 1)
  }
  for (const sale of scenario.sales) {
    fees -= Number(sale.fee || 0)
    wages -= Number(sale.wage || 0)
    const player = squad.find((p) => Number(p.uid ?? p.UID) === sale.playerRef.playerUid)
    adjust(player, -1)
  }
  return {
    fees,
    wages,
    feeRemaining: scenario.feeBudget - fees,
    wageRemaining: scenario.wageBudget - wages,
    missingTargets,
    unresolvedPlayers,
    coverage: [...coverage].map(([position, count]) => ({ position, count })),
  }
}
