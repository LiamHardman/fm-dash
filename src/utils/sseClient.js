// Consumes a fetch Response's body as an SSE stream, dispatching each event to the
// matching handler by event name (e.g. { status, done, error }). Shared by chatStore.js,
// WhoToSignDialog.vue, and ScoutReportTab.vue — all three backend routes emit the same
// event shape. Who to Sign and Scout Report were converted from a single blocking
// fetch+json() call to this streaming shape to fix a Safari-specific bug: WebKit's
// default ~60s idle-fetch timeout was tripping on calls that legitimately run up to
// their 120s server-side budget without sending any bytes in the interim (see
// .scratch/llm-refinements/issues/05-safari-compatibility-investigation.md).
export async function consumeSSEStream(response, handlers) {
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })

    let boundary = buffer.indexOf('\n\n')
    while (boundary !== -1) {
      dispatchSSEEvent(buffer.slice(0, boundary), handlers)
      buffer = buffer.slice(boundary + 2)
      boundary = buffer.indexOf('\n\n')
    }
  }
}

function dispatchSSEEvent(rawEvent, handlers) {
  let eventType = 'message'
  let dataLine = ''
  for (const line of rawEvent.split('\n')) {
    if (line.startsWith('event:')) eventType = line.slice(6).trim()
    else if (line.startsWith('data:')) dataLine += line.slice(5).trim()
  }
  if (!dataLine) return
  let payload
  try {
    payload = JSON.parse(dataLine)
  } catch (_e) {
    return
  }
  handlers[eventType]?.(payload)
}
