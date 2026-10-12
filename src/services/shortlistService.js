import { createProfileDocumentClient } from './profileDocumentClient'

export default createProfileDocumentClient('/api/shortlists', 'fm_dash_shortlists_v1', () => ({
  version: 1,
  lists: [
    {
      id: 'default',
      name: 'My shortlist',
      currency: 'USD',
      wagePeriod: 'week',
      items: [],
      scenarios: [],
    },
  ],
}))
