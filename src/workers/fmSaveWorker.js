// Parses an FM27 .fm save off the main thread and turns it into an FM-style CSV that
// /api/upload ingests. The save itself never leaves the browser.
//
// in:  { file: File }
// out: { type: 'progress', stage, current?, total? }
//      { type: 'done', csv: Blob (gzipped CSV), summary }
//      { type: 'error', message, kind? }

import { blobSource } from '../utils/fmSave/container.js'
import { exportSave } from '../utils/fmSave/exportSave.js'

self.onmessage = async ({ data }) => {
  try {
    const { csv, summary } = await exportSave(blobSource(data.file), (progress) =>
      self.postMessage({ type: 'progress', ...progress })
    )
    // A whole-database export (~25 MB) exceeds the default upload limit; gzip brings it
    // to roughly a third, and /api/upload inflates `.csv.gz` uploads.
    const gzipped = await new Response(
      new Blob([csv]).stream().pipeThrough(new CompressionStream('gzip'))
    ).blob()
    self.postMessage({ type: 'done', csv: gzipped, summary })
  } catch (error) {
    self.postMessage({ type: 'error', message: error?.message ?? String(error), kind: error?.kind })
  }
}
