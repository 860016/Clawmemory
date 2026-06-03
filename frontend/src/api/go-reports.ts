import api from './go-client'
import type { DailyReport } from './types'

export const reportApi = {
  list: (params?: { page?: number; size?: number }) =>
    api.get('/reports', { params }),
  create: (data: Omit<DailyReport, 'id' | 'user_id' | 'created_at'>) => api.post('/reports', data),
  getByDate: (date: string) => api.get(`/reports/${date}`),
  generate: (date: string) => api.post('/reports/generate', { date }),
}
