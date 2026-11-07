// Personality and Media Handling labels derived from a player's hidden personality bytes.
//
// FM stores no label in the save; the game derives both from the bytes. The thresholds and
// rule order are facts published by FM Moleman 27 / FM Starship (see
// .scratch/fm27-save-extras/research/personality-mapping.md and ./THIRD_PARTY.md). This is
// an independent implementation from those facts; no FM Starship code was copied.
// Verified on an FM27 save: Personality 14/14 against in-game labels, Media Handling 143/143
// against an in-game CSV export.

/**
 * @typedef {object} PersonalityInputs
 * @property {{ambition:number, controversy:number, loyalty:number, pressure:number,
 *   professionalism:number, sportsmanship:number, temperament:number}} personality 1-20
 * @property {number} determination 1-20
 * @property {number} leadership 1-20
 * @property {number} age at the game date
 * @property {boolean} male false for players at a women's club; FM's negative labels are men-only
 */

const LEADER_AGE = 23

/**
 * FM27's Personality label, first matching rule wins.
 * @param {PersonalityInputs} inputs
 */
export function derivePersonality({ personality, determination: d, leadership, age, male }) {
  const {
    ambition: am,
    loyalty: l,
    pressure: pre,
    professionalism: pr,
    sportsmanship: s,
    temperament: t,
  } = personality
  const adult = age >= LEADER_AGE

  if (s >= 15 && l >= 15 && pr >= 15 && d >= 14 && am >= 12 && t >= 15 && pre > 13) {
    return 'Model Citizen'
  }
  if (d >= 14 && am >= 14 && pr >= 14 && t < 10) return 'Perfectionist'
  if (male && pr <= 10 && t <= 4) return 'Temperamental'
  if (adult && leadership >= 19) {
    // Neither leader variant below occurs on the test save, so both are unconfirmed in-game.
    // "Natural Leader" is the FM27 string per the source; older guides call it Charismatic.
    if (d === 20 && leadership === 20) return 'Born Leader'
    if (s >= 18 && t >= 18) return 'Natural Leader'
    return 'Leader'
  }
  if (pr >= 18 && t >= 10) return pr === 20 && adult ? 'Model Professional' : 'Professional'
  if (d >= 18) return am >= 12 ? 'Driven' : 'Determined'
  if (male && pr < 5 && d < 10) return pr === 1 ? 'Slack' : 'Casual'
  if (male && l < 7 && am >= 16) return l >= 4 ? 'Fickle' : 'Mercenary'
  if (am >= 16 && l <= 9) return am === 20 ? 'Very Ambitious' : 'Ambitious'
  if (male && am < 6 && l >= 11) return 'Unambitious'
  if (d <= 9 && s >= 18) return s === 20 ? 'Honest' : 'Sporting'
  if (male && s < 5 && d >= 11) return s === 1 ? 'Unsporting' : 'Realist'
  if (male && d < 6 && am < 10) return d === 1 ? 'Easily Discouraged' : 'Low Determination'
  // Devoted needs the player to be at a favourite club, which the inputs don't carry, so
  // it is never emitted; those players fall through to the rules below.
  if (am <= 7 && l >= 18) return l === 20 ? 'Very Loyal' : 'Loyal'
  if (d >= 15 && pre >= 17) return pre === 20 ? 'Iron Willed' : 'Resilient'
  if (male && d < 10 && pre < 4) return pre === 1 ? 'Spineless' : 'Low Self Belief'
  if (t >= 10 && pre >= 15) {
    if (s >= 15) return 'Light-Hearted'
    return pr >= 11 ? 'Spirited' : 'Jovial'
  }
  if (d >= 15) return pr >= 15 ? 'Resolute' : 'Fairly Determined'
  if (pr > 14) return 'Fairly Professional'
  if (am > 14) return 'Fairly Ambitious'
  if (l >= 15) return 'Fairly Loyal'
  return s >= 15 ? 'Fairly Sporting' : 'Balanced'
}

// Media Handling labels in the order FM joins them.
const MEDIA_ORDER = [
  'Outspoken',
  'Evasive',
  'Volatile',
  'Unflappable',
  'Media-friendly',
  'Level-headed',
  'Reserved',
  'Short-tempered',
  'Confrontational',
]

/**
 * FM27's Media Handling label: up to one label from each of three slots, "Level-headed"
 * when none fires, joined with ", " in FM's order.
 * @param {PersonalityInputs} inputs (determination, leadership and age are not used)
 */
export function deriveMediaHandling({ personality, male }) {
  const {
    controversy: c,
    loyalty: l,
    pressure: pre,
    professionalism: pr,
    sportsmanship: s,
    temperament: t,
  } = personality
  const labels = new Set()

  // Temperament slot.
  if (t < 3) labels.add('Short-tempered')
  else if (t < 7) labels.add('Volatile')

  // Attitude slot.
  if (pr >= 15 && t <= 15 && pre >= 15 && (s > 12 || c < 15)) labels.add('Evasive')
  else if (c >= 15) labels.add('Outspoken')
  else if (l < 11 || (s < 12 && pr < 13)) labels.add('Media-friendly')

  // Composure slot.
  const fiery = labels.has('Short-tempered') || labels.has('Volatile')
  if (s <= 7 && t <= 7) labels.add('Confrontational')
  else if (t >= 15 && pre > 14) labels.add('Unflappable')
  else if (c <= 5 && pr >= 15 && !fiery) labels.add('Reserved')

  if (labels.size === 0) labels.add('Level-headed')

  // Women's players never get the combative labels; FM shows them as Outspoken instead.
  if (!male && ['Short-tempered', 'Volatile', 'Confrontational'].some((x) => labels.has(x))) {
    labels.delete('Short-tempered')
    labels.delete('Volatile')
    labels.delete('Confrontational')
    labels.add('Outspoken')
  }
  return MEDIA_ORDER.filter((label) => labels.has(label)).join(', ')
}

/**
 * Both labels for a decoded save player, or blanks when it has no personality bytes.
 * @returns {[string, string]} [Personality, Media Handling]
 */
export function personalityLabels(player) {
  if (!player.personality || player.age === undefined) return ['', '']
  const inputs = {
    personality: player.personality,
    determination: player.attributes.determination,
    leadership: player.attributes.leadership,
    age: player.age,
    male: player.gender !== 1, // 1 = the club's teams are women's teams
  }
  return [derivePersonality(inputs), deriveMediaHandling(inputs)]
}
