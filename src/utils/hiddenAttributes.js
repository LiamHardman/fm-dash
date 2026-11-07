// Hidden and personality attributes carried by FM27 save imports (player.hiddenAttributes,
// keyed by FM's names). Display and filtering only: no rating, Overall, MBR or CA path
// reads them.

export const HIDDEN_ATTRIBUTE_GROUPS = [
  {
    name: 'Hidden',
    attributes: [
      'Consistency',
      'Important Matches',
      'Versatility',
      'Injury Proneness',
      'Dirtiness',
    ],
  },
  {
    name: 'Personality',
    attributes: [
      'Adaptability',
      'Ambition',
      'Controversy',
      'Loyalty',
      'Pressure',
      'Professionalism',
      'Sportsmanship',
      'Temperament',
    ],
  },
]
export const HIDDEN_ATTRIBUTES = HIDDEN_ATTRIBUTE_GROUPS.flatMap((group) => group.attributes)

// Lower is better for these: their colour scale is inverted and their filter is a maximum.
export const LOWER_IS_BETTER = new Set(['Injury Proneness', 'Dirtiness', 'Controversy'])

/** Some player in the Dataset carries hidden attributes (only save imports do). */
export function hasHiddenAttributes(players) {
  return (
    Array.isArray(players) &&
    players.some(
      (player) => player?.hiddenAttributes && Object.keys(player.hiddenAttributes).length > 0
    )
  )
}

export function hiddenAttributeValue(player, name) {
  return player?.hiddenAttributes?.[name]
}

/** The 1-20 value to feed the shared attribute colour scale (inverted where lower is better). */
export function hiddenAttributeColorValue(name, value) {
  if (value === undefined || value === null) return value
  return LOWER_IS_BETTER.has(name) ? 21 - value : value
}

const keyOf = (name) => name.replace(/\s+/g, '')

/** The currentFilters key for an attribute: minHidden<Key>, or maxHidden<Key> when lower is better. */
export function hiddenFilterKey(name) {
  return `${LOWER_IS_BETTER.has(name) ? 'max' : 'min'}Hidden${keyOf(name)}`
}

/** Filter defaults (all off): minimums 0, maximums 20. */
export function defaultHiddenFilters() {
  return Object.fromEntries(
    HIDDEN_ATTRIBUTES.map((name) => [hiddenFilterKey(name), LOWER_IS_BETTER.has(name) ? 20 : 0])
  )
}

/** The active hidden filters as [name, kind, bound] entries. */
export function activeHiddenFilters(filters) {
  const active = []
  for (const name of HIDDEN_ATTRIBUTES) {
    const bound = filters?.[hiddenFilterKey(name)]
    if (bound === undefined || bound === null) continue
    if (LOWER_IS_BETTER.has(name) ? bound < 20 : bound > 0) {
      active.push([name, LOWER_IS_BETTER.has(name) ? 'max' : 'min', bound])
    }
  }
  return active
}

/** A player missing a value fails any active hidden filter on it. */
export function passesHiddenFilters(player, active) {
  for (const [name, kind, bound] of active) {
    const value = hiddenAttributeValue(player, name)
    if (value === undefined || value === null) return false
    if (kind === 'min' ? value < bound : value > bound) return false
  }
  return true
}
