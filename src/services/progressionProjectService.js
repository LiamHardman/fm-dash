import { createProfileDocumentClient } from './profileDocumentClient'

export default createProfileDocumentClient(
  '/api/progression/projects',
  'fm_dash_progression_projects_v1',
  () => ({ version: 1, projects: [] })
)
