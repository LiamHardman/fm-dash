export function mergeProfileBackup(current, incoming, replaceConflicts = false) {
  const changes = []
  const result = { version: 1 }
  for (const [section, key] of [
    ['shortlists', 'lists'],
    ['savedSearches', 'searches'],
    ['progressionProjects', 'projects'],
  ]) {
    const merged = new Map(current[section][key].map((item) => [item.id, item]))
    for (const item of incoming[section][key]) {
      const conflict = merged.has(item.id)
      changes.push({
        section,
        id: item.id,
        name: item.name,
        action: conflict ? (replaceConflicts ? 'Replace existing' : 'Keep existing') : 'Add',
      })
      if (!conflict || replaceConflicts) merged.set(item.id, item)
    }
    result[section] = { version: 1, [key]: [...merged.values()] }
  }
  return { document: result, changes }
}
