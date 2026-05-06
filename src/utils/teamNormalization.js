const TEAM_PREFIXES = [
  'fc',
  'cf',
  'ac',
  'sc',
  'as',
  'ca',
  'cs',
  'rc',
  'rs',
  'cd',
  'ud',
  'rcd',
  'rsd',
  'rfc',
  'afc',
  'cfc',
  'sfc',
  'club',
  'athletic',
  'atletico',
  'athletico',
  'deportivo',
  'sporting',
  'association',
  'society',
  'union',
]

const TEAM_SUFFIXES = [
  'fc',
  'cf',
  'ac',
  'sc',
  'as',
  'ca',
  'cs',
  'rc',
  'rs',
  'cd',
  'ud',
  'rcd',
  'rsd',
  'rfc',
  'afc',
  'cfc',
  'sfc',
  'club',
  'rovers',
  'wanderers',
  'albion',
  'villa',
  'county',
  'athletic',
  'atletico',
  'athletico',
  'deportivo',
  'sporting',
  'association',
  'society',
  'union',
]

const ABBREVIATIONS = {
  psg: 'paris saint germain',
  'paris saint germain': 'paris saint germain',
  'man utd': 'manchester united',
  'manchester utd': 'manchester united',
  'man united': 'manchester united',
  'man city': 'manchester city',
  tottenham: 'tottenham hotspur',
  spurs: 'tottenham hotspur',
  brighton: 'brighton hove albion',
  'west ham': 'west ham united',
  newcastle: 'newcastle united',
  'sheffield utd': 'sheffield united',
  'sheffield wed': 'sheffield wednesday',
  'nottm forest': 'nottingham forest',
  'notts forest': 'nottingham forest',
  wolves: 'wolverhampton wanderers',
  qpr: 'queens park rangers',
  barca: 'barcelona',
}

const STOP_WORDS = new Set(['football', 'soccer', 'sports', 'sport'])

function stripEdgeWords(words, edgeWords, fromStart) {
  const stripped = [...words]
  while (stripped.length > 1) {
    const edgeWord = fromStart ? stripped[0] : stripped[stripped.length - 1]
    if (!edgeWords.has(edgeWord)) break
    if (fromStart) {
      stripped.shift()
    } else {
      stripped.pop()
    }
  }
  return stripped
}

/**
 * Normalize a team name for logo lookup and fuzzy matching.
 * @param {string} name - Team name to normalize.
 * @returns {string} Normalized team name.
 */
export function normalizeTeamName(name) {
  if (!name) return ''

  const normalized = String(name)
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .trim()
    .replace(/&/g, ' and ')
    .replace(/\bf\s*\.\s*c\s*\.?/g, 'fc')
    .replace(/\bc\s*\.\s*f\s*\.?/g, 'cf')
    .replace(/\ba\s*\.\s*c\s*\.?/g, 'ac')
    .replace(/\bs\s*\.\s*c\s*\.?/g, 'sc')
    .replace(/[^\w\s]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  if (!normalized) return ''

  const expanded = ABBREVIATIONS[normalized] || normalized
  const prefixWords = new Set(TEAM_PREFIXES)
  const suffixWords = new Set(TEAM_SUFFIXES)
  let words = expanded.split(' ').filter((word) => word && !STOP_WORDS.has(word))

  words = stripEdgeWords(words, prefixWords, true)
  words = stripEdgeWords(words, suffixWords, false)

  return words.join(' ')
}
