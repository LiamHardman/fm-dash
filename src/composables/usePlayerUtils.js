export function usePlayerUtils() {
  // GK stat mapping for FIFA-style display
  const gkStatMapping = {
    PAC: 'DIV', // Diving -> Pace
    SHO: 'HAN', // Handling -> Shooting
    PAS: 'KIC', // Kicking -> Passing
    DRI: 'REF', // Reflexes -> Dribbling
    DEF: 'SPD', // Speed -> Defending
    PHY: 'POS' // Positioning -> Physical
  }

  // Position group mappings
  const positionGroups = {
    GK: ['GK'],
    DEF: ['CB', 'LB', 'RB', 'LWB', 'RWB', 'SW'],
    MID: ['CM', 'CDM', 'CAM', 'LM', 'RM', 'DM', 'AM'],
    ATT: ['ST', 'CF', 'LW', 'RW', 'IF', 'TQ', 'F9']
  }

  // Get player value with GK mapping applied
  const getPlayerValue = (player, fieldKey, columnName = null, isGoalkeeperView = false) => {
    if (!isGoalkeeperView && player.position && player.position.includes('GK')) {
      const mappedStat = gkStatMapping[columnName || fieldKey]
      if (mappedStat && player[mappedStat] !== undefined) {
        return player[mappedStat]
      }
    }

    return player[fieldKey]
  }

  // Check if player is a goalkeeper
  const isGoalkeeper = player => {
    return (
      player.position && (player.position.includes('GK') || player.position.includes('Goalkeeper'))
    )
  }

  // Get player's position group
  const getPositionGroup = position => {
    if (!position) return 'UNKNOWN'

    for (const [group, positions] of Object.entries(positionGroups)) {
      if (positions.some(pos => position.includes(pos))) {
        return group
      }
    }
    return 'UNKNOWN'
  }

  // Get position index for sorting
  const getPositionIndex = position => {
    const positionOrder = [
      'GK',
      'CB',
      'LB',
      'RB',
      'LWB',
      'RWB',
      'SW',
      'CDM',
      'CM',
      'CAM',
      'LM',
      'RM',
      'DM',
      'AM',
      'LW',
      'RW',
      'ST',
      'CF',
      'IF',
      'TQ',
      'F9'
    ]
    const index = positionOrder.indexOf(position)
    return index === -1 ? 999 : index
  }

  // Format player name with proper capitalization
  const formatPlayerName = name => {
    if (!name) return ''
    return name
      .split(' ')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ')
  }

  // Get player age category
  const getAgeCategory = age => {
    if (age < 18) return 'Youth'
    if (age < 23) return 'Young'
    if (age < 30) return 'Prime'
    if (age < 35) return 'Experienced'
    return 'Veteran'
  }

  // Calculate player potential rating
  const calculatePotentialRating = (current, potential, age) => {
    if (!potential || !current || !age) return current || 0

    const ageFactor = Math.max(0, (35 - age) / 17) // Peak at 18, decline after 35
    return Math.min(potential, current + (potential - current) * ageFactor)
  }

  return {
    gkStatMapping,
    positionGroups,
    getPlayerValue,
    isGoalkeeper,
    getPositionGroup,
    getPositionIndex,
    formatPlayerName,
    getAgeCategory,
    calculatePotentialRating
  }
}
