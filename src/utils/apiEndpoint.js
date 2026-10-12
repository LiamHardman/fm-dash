export function apiUrl(path) {
  const endpoint =
    globalThis.window?.APP_CONFIG?.API_ENDPOINT ?? import.meta.env.VITE_API_ENDPOINT ?? ''
  return `${endpoint.replace(/\/$/, '')}${path}`
}
