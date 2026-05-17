import api from './go-client'

export const statsApi = {
  getOverview: () => api.get('/stats'),
  getUsage: (days?: number) => api.get('/stats/usage', { params: { days } }),
}
