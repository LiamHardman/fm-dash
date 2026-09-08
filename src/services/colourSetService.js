const API_ENDPOINT =
  typeof window !== 'undefined' && window.APP_CONFIG?.API_ENDPOINT !== undefined
    ? window.APP_CONFIG.API_ENDPOINT
    : import.meta.env.VITE_API_ENDPOINT || ''

async function request(scale, document) {
  const response = await fetch(`${API_ENDPOINT}/api/colour-sets/${scale}`, {
    ...(document
      ? {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(document),
        }
      : {}),
  })
  if (!response.ok)
    throw new Error(`Could not ${document ? 'save' : 'load'} colour sets (${response.status}).`)
  return response.json()
}

export default {
  load: (scale) => request(scale),
  save: (scale, document) => request(scale, document),
}
