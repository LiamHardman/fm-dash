import { createProfileDocumentClient } from './profileDocumentClient'

export default createProfileDocumentClient(
  '/api/saved-searches',
  'fm_dash_saved_searches_v1',
  () => ({ version: 1, searches: [] })
)
