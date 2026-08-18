import DOMPurify from 'dompurify'
import { marked } from 'marked'

marked.setOptions({ breaks: true, gfm: true })

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

// Renders assistant chat text as sanitized markdown HTML, with [[player:UID]] tokens
// (ticket 05/06's citation format) swapped for clickable spans first — so their name
// text is escaped and protected before marked ever sees the source, and so unresolvable
// tokens (a model slip) fall back to their literal text rather than a broken chip. The
// player span carries a data-player-uid attribute rather than a live click handler,
// since v-html content can't host Vue component bindings — the caller attaches one
// delegated click listener and reads that attribute instead. isUidOpenable (typically a
// lookup against the currently loaded player list) decides the unresolved styling —
// distinct from citation resolution above, since a uid can be a real past citation yet
// no longer openable (e.g. dataset changed).
export function renderChatMessageHtml(text, referencedPlayers = [], isUidOpenable = () => true) {
  const byUid = new Map((referencedPlayers || []).map((p) => [String(p.uid), p]))
  const withChips = (text || '').replace(/\[\[player:([\w-]+)\]\]/g, (match, uid) => {
    const ref = byUid.get(uid)
    if (!ref) return escapeHtml(match)
    const cls = isUidOpenable(ref.uid)
      ? 'chat-player-link'
      : 'chat-player-link chat-player-link--unresolved'
    return `<span class="${cls}" data-player-uid="${escapeHtml(String(ref.uid))}">${escapeHtml(ref.name)}</span>`
  })
  const html = marked.parse(withChips)
  return DOMPurify.sanitize(html, { ADD_ATTR: ['data-player-uid'] })
}

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
