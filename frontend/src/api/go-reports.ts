import api from './go-client'

export const reportApi = {
  list: (params?: { page?: number; size?: number }) =>
    api.get('/reports', { params }),
  create: (data: any) => api.post('/reports', data),
  getByDate: (date: string) => api.get(`/reports/${date}`),
  generate: (date: string) => api.post('/reports/generate', { date }),
}
