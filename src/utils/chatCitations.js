// Splits assistant chat text on [[player:UID]] placeholder tokens (ticket 05/06's
// citation format) into renderable text/player segments. A token only becomes a
// 'player' part when its uid is present in referencedPlayers — an unresolvable token
// (a model slip) falls back to its literal text rather than a broken chip. Whether the
// resulting uid is actually openable (i.e. still present in the loaded dataset) is a
// separate, later check the caller makes against its own player list.
export function parseChatCitations(text, referencedPlayers = []) {
  const byUid = new Map((referencedPlayers || []).map((p) => [String(p.uid), p]))
  const parts = []
  const pattern = /\[\[player:([\w-]+)\]\]/g
  const source = text || ''
  let lastIndex = 0
  let match = pattern.exec(source)

  while (match) {
    if (match.index > lastIndex) {
      parts.push({ type: 'text', content: source.slice(lastIndex, match.index) })
    }
    const ref = byUid.get(match[1])
    if (ref) {
      parts.push({ type: 'player', uid: ref.uid, name: ref.name })
    } else {
      parts.push({ type: 'text', content: match[0] })
    }
    lastIndex = pattern.lastIndex
    match = pattern.exec(source)
  }

  if (lastIndex < source.length) {
    parts.push({ type: 'text', content: source.slice(lastIndex) })
  }
  return parts
}
