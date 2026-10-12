import { sortPlayerIndices } from '../utils/playerSorting.js'

let rows = []
let generation = 0
self.onmessage = ({ data }) => {
  if (data.type === 'load') {
    rows = data.rows
    generation = data.generation
    return
  }
  try {
    if (data.generation !== generation) throw new Error('Dataset changed')
    const indices = Uint32Array.from(sortPlayerIndices(rows, data.spec))
    self.postMessage({ id: data.id, indices }, [indices.buffer])
  } catch (error) {
    self.postMessage({ id: data.id, error: error.message })
  }
}
