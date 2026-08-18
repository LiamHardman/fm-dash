// Squad best-XI/depth-chart computation, extracted verbatim from TeamViewPage.vue's
// setup() (map decision, see .scratch/llm-refinements/issues/
// 03-chat-tactics-formation-display.md) so ChatWidget.vue's tactic-fit pitch display
// and TeamViewPage.vue's own Best XI/depth chart call the exact same function and can
// never disagree about who best fits a given formation. Every scoring/matching rule
// below is unchanged from the original inline implementation -- only closed-over
// component refs (teamPlayers.value) became explicit parameters.

export const fmSlotRoleMatcher = {
  GK: ['Goalkeeper'],
  'D (R)': ['Defender (Right)', 'Right Back'],
  'D (L)': ['Defender (Left)', 'Left Back'],
  'D (C)': ['Defender (Centre)', 'Centre Back'],
  'WB (R)': ['Wing-Back (Right)', 'Right Wing-Back'],
  'WB (L)': ['Wing-Back (Left)', 'Left Wing-Back'],
  'DM (C)': ['Defensive Midfielder (Centre)', 'Centre Defensive Midfielder'],
  'M (R)': ['Midfielder (Right)', 'Right Midfielder'],
  'M (L)': ['Midfielder (Left)', 'Left Midfielder'],
  'M (C)': ['Midfielder (Centre)', 'Centre Midfielder'],
  'AM (R)': ['Attacking Midfielder (Right)', 'Right Attacking Midfielder', 'Winger (Right)'],
  'AM (L)': ['Attacking Midfielder (Left)', 'Left Attacking Midfielder', 'Winger (Left)'],
  'AM (C)': ['Attacking Midfielder (Centre)', 'Centre Attacking Midfielder'],
  'ST (C)': ['Striker (Centre)', 'Striker'],
}

const fmMatcherToRoleKeyPrefix = {
  GOALKEEPER: 'GK',
  SWEEPER: 'DC',
  'DEFENDER (RIGHT)': 'DR',
  'RIGHT BACK': 'DR',
  'DEFENDER (LEFT)': 'DL',
  'LEFT BACK': 'DL',
  'DEFENDER (CENTRE)': 'DC',
  'CENTRE BACK': 'DC',
  'WING-BACK (RIGHT)': 'WBR',
  'RIGHT WING-BACK': 'WBR',
  'WING-BACK (LEFT)': 'WBL',
  'LEFT WING-BACK': 'WBL',
  'DEFENSIVE MIDFIELDER (CENTRE)': 'DM',
  'CENTRE DEFENSIVE MIDFIELDER': 'DM',
  'MIDFIELDER (RIGHT)': 'MR',
  'RIGHT MIDFIELDER': 'MR',
  'MIDFIELDER (LEFT)': 'ML',
  'LEFT MIDFIELDER': 'ML',
  'MIDFIELDER (CENTRE)': 'MC',
  'CENTRE MIDFIELDER': 'MC',
  'ATTACKING MIDFIELDER (RIGHT)': 'AMR',
  'RIGHT ATTACKING MIDFIELDER': 'AMR',
  'WINGER (RIGHT)': 'AMR',
  'ATTACKING MIDFIELDER (LEFT)': 'AML',
  'LEFT ATTACKING MIDFIELDER': 'AML',
  'WINGER (LEFT)': 'AML',
  'ATTACKING MIDFIELDER (CENTRE)': 'AMC',
  'CENTRE ATTACKING MIDFIELDER': 'AMC',
  'STRIKER (CENTRE)': 'ST',
  STRIKER: 'ST',
}

// For handling combined positions like D/WB(R)
// The first position is the PREFERRED position, others are fallbacks
export const positionSideMap = {
  'D (R)': ['DR'],
  'D (L)': ['DL'],
  'D (C)': ['DC'],
  'WB (R)': ['WBR'],
  'WB (L)': ['WBL'],
  'DM (C)': ['DM'],
  'M (R)': ['MR'],
  'M (L)': ['ML'],
  'M (C)': ['MC'],
  'AM (R)': ['AMR'],
  'AM (L)': ['AML'],
  'AM (C)': ['AMC'],
  'ST (C)': ['ST'],
  GK: ['GK'],
}

export const fallbackPositionMap = {
  'D (R)': ['DR', 'WBR', 'MR'],
  'D (L)': ['DL', 'WBL', 'ML'],
  'D (C)': ['DC', 'DM'],
  'WB (R)': ['WBR', 'DR', 'MR'],
  'WB (L)': ['WBL', 'DL', 'ML'],
  'DM (C)': ['DM', 'DC', 'MC'],
  'M (R)': ['MR', 'WBR', 'AMR'],
  'M (L)': ['ML', 'WBL', 'AML'],
  'M (C)': ['MC', 'DM'],
  'AM (R)': ['AMR', 'MR'],
  'AM (L)': ['AML', 'ML'],
  'AM (C)': ['AMC', 'MC'],
  'ST (C)': ['ST', 'AMC'],
  GK: ['GK'],
}

export const MIN_SUITABILITY_THRESHOLD = 10

export function getPlayerShortPositions(player) {
  if (Array.isArray(player?.shortPositions)) return player.shortPositions
  if (Array.isArray(player?.short_positions)) return player.short_positions
  return []
}

export function getSlotExactPositions(slotRole) {
  return positionSideMap[slotRole.toUpperCase()] || []
}

export function getSlotFallbackPositions(slotRole) {
  const exactPositions = getSlotExactPositions(slotRole)
  const fallbackPositions = fallbackPositionMap[slotRole.toUpperCase()] || []
  return fallbackPositions.filter((position) => !exactPositions.includes(position))
}

function getExactCandidateCountForSlot(slot, players) {
  const slotPositions = getSlotExactPositions(slot.role)
  return players.filter((player) =>
    getPlayerShortPositions(player).some((position) => slotPositions.includes(position))
  ).length
}

// Fills scarce natural positions first so flexible players don't block
// wing-backs/full-backs that fewer players can actually cover.
export function getSlotSelectionOrder(formationSlots, players) {
  return [...formationSlots].sort((a, b) => {
    const exactCandidateDifference =
      getExactCandidateCountForSlot(a, players) - getExactCandidateCountForSlot(b, players)
    if (exactCandidateDifference !== 0) return exactCandidateDifference
    return formationSlots.indexOf(a) - formationSlots.indexOf(b)
  })
}

export function getPlayerOverallForRole(player, slotFormationRole) {
  if (!player || !slotFormationRole) return 0

  let bestScoreForRole = 0

  if (!player.roleSpecificOveralls) {
    return player.Overall || 0
  }

  const hasRoleOveralls = Array.isArray(player.roleSpecificOveralls)
    ? player.roleSpecificOveralls.length > 0
    : Object.keys(player.roleSpecificOveralls).length > 0

  if (!hasRoleOveralls) {
    return player.Overall || 0
  }

  const upperSlotRoleOriginal = slotFormationRole.toUpperCase()
  const requiredPositions = positionSideMap[upperSlotRoleOriginal] || []

  if (player.shortPositions && player.shortPositions.length > 0) {
    const exactPositionMatches = player.shortPositions.filter((pos) =>
      requiredPositions.includes(pos)
    )

    if (exactPositionMatches.length > 0) {
      if (Array.isArray(player.roleSpecificOveralls)) {
        for (const rso of player.roleSpecificOveralls) {
          const rsoBasePosition = rso.roleName.split(' - ')[0].trim()
          if (exactPositionMatches.includes(rsoBasePosition)) {
            bestScoreForRole = Math.max(bestScoreForRole, rso.score)
          }
        }
      } else {
        for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
          const rsoBasePosition = roleName.split(' - ')[0].trim()
          if (exactPositionMatches.includes(rsoBasePosition)) {
            bestScoreForRole = Math.max(bestScoreForRole, score)
          }
        }
      }

      if (bestScoreForRole === 0) {
        bestScoreForRole = Math.max(MIN_SUITABILITY_THRESHOLD, player.Overall || 0)
      }
    }
  }

  if (bestScoreForRole > 0) {
    return bestScoreForRole
  }

  const fallbackPositions = fallbackPositionMap[upperSlotRoleOriginal] || []

  if (player.shortPositions && player.shortPositions.length > 0) {
    const fallbackMatches = player.shortPositions.filter((pos) => fallbackPositions.includes(pos))

    if (fallbackMatches.length > 0) {
      if (Array.isArray(player.roleSpecificOveralls)) {
        for (const rso of player.roleSpecificOveralls) {
          const rsoBasePosition = rso.roleName.split(' - ')[0].trim()
          if (fallbackMatches.includes(rsoBasePosition)) {
            bestScoreForRole = Math.max(bestScoreForRole, rso.score)
          }
        }
      } else {
        for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
          const rsoBasePosition = roleName.split(' - ')[0].trim()
          if (fallbackMatches.includes(rsoBasePosition)) {
            bestScoreForRole = Math.max(bestScoreForRole, score)
          }
        }
      }

      if (bestScoreForRole === 0) {
        bestScoreForRole = Math.max(MIN_SUITABILITY_THRESHOLD - 10, (player.Overall || 0) - 5)
      }
    }
  }

  if (bestScoreForRole === 0) {
    const upperSlotRole = slotFormationRole.toUpperCase()
    const fmPositionMatchers = fmSlotRoleMatcher[upperSlotRole] || [upperSlotRole]

    const targetRoleKeyPrefixes = fmPositionMatchers
      .map((matcher) => fmMatcherToRoleKeyPrefix[matcher.toUpperCase()])
      .filter((prefix) => !!prefix)
      .reduce((acc, val) => {
        if (!acc.includes(val)) acc.push(val)
        return acc
      }, [])

    if (Array.isArray(player.roleSpecificOveralls)) {
      for (const rso of player.roleSpecificOveralls) {
        const rsoBasePosition = rso.roleName.split(' - ')[0].trim()
        if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
          bestScoreForRole = Math.max(bestScoreForRole, rso.score)
        }
      }
    } else if (player.roleSpecificOveralls) {
      for (const [roleName, score] of Object.entries(player.roleSpecificOveralls)) {
        const rsoBasePosition = roleName.split(' - ')[0].trim()
        if (targetRoleKeyPrefixes.includes(rsoBasePosition)) {
          bestScoreForRole = Math.max(bestScoreForRole, score)
        }
      }
    }

    // Final fallback to player's general Overall rating if still nothing found
    if (bestScoreForRole === 0) {
      bestScoreForRole = Math.max(0, (player.Overall || 0) - 10)
    }
  }

  return bestScoreForRole
}

// computeSquadComposition mirrors TeamViewPage.vue's former calculateBestTeamAndDepth
// exactly (minus caching/UI-message state, which stay page-local): same scoring, same
// exact-match/fallback two-pass fill per depth level, same 3-deep depth chart, same
// final orphan-slot fallback pass. Returns { squadComposition, bestTeamAverageOverall }.
export function computeSquadComposition(players, formationLayout) {
  const formationSlots = formationLayout.flatMap((row) => row.positions)
  const tempSquadComposition = {}
  for (const slot of formationSlots) {
    tempSquadComposition[slot.id] = []
  }

  const playerPositionMap = new Map()
  for (const player of players) {
    playerPositionMap.set(player.name, [...getPlayerShortPositions(player)])
  }

  const allPotentialPlayerAssignments = []
  for (const slot of formationSlots) {
    for (const player of players) {
      const overallInRole = getPlayerOverallForRole(player, slot.role)

      if (overallInRole >= MIN_SUITABILITY_THRESHOLD) {
        const slotPositions = getSlotExactPositions(slot.role)
        const fallbackPositions = getSlotFallbackPositions(slot.role)
        const playerPositions = playerPositionMap.get(player.name) || []

        const isExactMatch = playerPositions.some((pos) => slotPositions.includes(pos))
        const isFallbackMatch = playerPositions.some((pos) => fallbackPositions.includes(pos))
        const canPlayInPosition = isExactMatch || isFallbackMatch

        if (canPlayInPosition) {
          const assignment = {
            player,
            slotId: slot.id,
            slotRole: slot.role,
            overallInRole,
            sortScore: overallInRole,
            exactMatch: isExactMatch,
          }
          assignment.sortScore += isExactMatch ? 10000 : -5000
          allPotentialPlayerAssignments.push(assignment)
        }
      }
    }
  }

  allPotentialPlayerAssignments.sort((a, b) => b.sortScore - a.sortScore)

  const assignedPlayersToSlots = new Set()
  const slotsByNaturalScarcity = getSlotSelectionOrder(formationSlots, players)

  const isAlreadyStarterElsewhere = (playerName) => {
    for (const sId in tempSquadComposition) {
      if (
        tempSquadComposition[sId].length > 0 &&
        tempSquadComposition[sId][0].player.name === playerName
      ) {
        return true
      }
    }
    return false
  }

  for (let depthIndex = 0; depthIndex < 3; depthIndex++) {
    // First pass: fill positions with exact matches
    for (const slot of slotsByNaturalScarcity) {
      if (tempSquadComposition[slot.id].length === depthIndex) {
        for (const assignment of allPotentialPlayerAssignments) {
          if (
            assignment.slotId === slot.id &&
            assignment.exactMatch &&
            !assignedPlayersToSlots.has(assignment.player.name)
          ) {
            if (depthIndex === 0 || !isAlreadyStarterElsewhere(assignment.player.name)) {
              tempSquadComposition[slot.id].push({
                player: assignment.player,
                overallInRole: assignment.overallInRole,
                exactMatch: assignment.exactMatch,
              })
              assignedPlayersToSlots.add(assignment.player.name)
              break
            }
          }
        }
      }
    }

    // Second pass: fill remaining positions with fallback matches
    for (const slot of slotsByNaturalScarcity) {
      if (tempSquadComposition[slot.id].length === depthIndex) {
        for (const assignment of allPotentialPlayerAssignments) {
          if (
            assignment.slotId === slot.id &&
            !assignment.exactMatch &&
            !assignedPlayersToSlots.has(assignment.player.name)
          ) {
            if (depthIndex === 0 || !isAlreadyStarterElsewhere(assignment.player.name)) {
              tempSquadComposition[slot.id].push({
                player: assignment.player,
                overallInRole: assignment.overallInRole,
                exactMatch: assignment.exactMatch,
              })
              assignedPlayersToSlots.add(assignment.player.name)
              break
            }
          }
        }
      }
    }
  }

  // Ensure each slot is sorted by overallInRole descending (exact matches first)
  for (const slotId in tempSquadComposition) {
    tempSquadComposition[slotId].sort((a, b) => {
      if (a.exactMatch !== b.exactMatch) return a.exactMatch ? -1 : 1
      return b.overallInRole - a.overallInRole
    })
  }

  // Any slot still empty after all 3 depths: find any player who can play a
  // fallback position there, as a last resort.
  for (const slot of formationSlots) {
    if (tempSquadComposition[slot.id].length === 0) {
      const fallbackPositions = fallbackPositionMap[slot.role.toUpperCase()] || []
      const fallbackAssignments = []

      for (const player of players) {
        if (!assignedPlayersToSlots.has(player.name)) {
          const playerPositions = getPlayerShortPositions(player)
          const canPlayFallback = playerPositions.some((pos) => fallbackPositions.includes(pos))

          if (canPlayFallback) {
            const overallInRole = getPlayerOverallForRole(player, slot.role)
            if (overallInRole >= MIN_SUITABILITY_THRESHOLD - 10) {
              fallbackAssignments.push({ player, overallInRole, exactMatch: false })
            }
          }
        }
      }

      fallbackAssignments.sort((a, b) => b.overallInRole - a.overallInRole)

      if (fallbackAssignments.length > 0) {
        const bestFallback = fallbackAssignments[0]
        tempSquadComposition[slot.id].push(bestFallback)
        assignedPlayersToSlots.add(bestFallback.player.name)
      }
    }
  }

  let sumOfStartersOverall = 0
  let startersCount = 0
  for (const slotPlayers of Object.values(tempSquadComposition)) {
    if (slotPlayers && slotPlayers.length > 0) {
      sumOfStartersOverall += slotPlayers[0].overallInRole
      startersCount++
    }
  }
  const bestTeamAverageOverall =
    startersCount > 0 ? Math.round(sumOfStartersOverall / startersCount) : 0

  return { squadComposition: tempSquadComposition, bestTeamAverageOverall }
}
