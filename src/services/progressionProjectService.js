const endpoint = '/api/progression/projects'
export default {
  async list() {
    const r = await fetch(endpoint, { headers: { Accept: 'application/json' } })
    if (!r.ok) throw new Error('Unable to load progression projects')
    return r.json()
  },
  async save(document) {
    const r = await fetch(endpoint, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify(document),
    })
    if (!r.ok) throw new Error('Unable to save progression projects')
    return r.json()
  },
}
