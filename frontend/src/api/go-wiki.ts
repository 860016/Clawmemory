import api from './go-client'
import type { WikiCreateParams, WikiUpdateParams } from './types'

export const wikiApi = {
  list: (params?: { category?: string; status?: string; page?: number; size?: number }) =>
    api.get('/wiki', { params }),
  create: (data: WikiCreateParams) => api.post('/wiki', data),
  get: (id: number) => api.get(`/wiki/${id}`),
  update: (id: number, data: WikiUpdateParams) => api.put(`/wiki/${id}`, data),
  delete: (id: number) => api.delete(`/wiki/${id}`),
  tree: () => api.get('/wiki/tree'),
  search: (q: string, limit?: number) =>
    api.get('/wiki/search', { params: { q, limit } }),
  categories: () => api.get('/wiki/categories'),
  stats: () => api.get('/wiki/stats'),
  config: () => api.get('/wiki/config'),
  markComplete: (id: number) => api.post(`/wiki/${id}/mark-complete`),
  markInProgress: (id: number) => api.post(`/wiki/${id}/mark-in-progress`),
  aiExtract: (data: Record<string, unknown>) => api.post('/wiki/ai/extract', data),
  refine: (id: number, data?: Record<string, unknown>) => api.post(`/wiki/${id}/refine`, data),
}
