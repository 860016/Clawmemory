import api from './go-client'

export const memoryApi = {
  list: (params?: { layer?: string; page?: number; size?: number; status?: string }) =>
    api.get('/memories', { params }),
  create: (data: any) => api.post('/memories', data),
  get: (id: number) => api.get(`/memories/${id}`),
  update: (id: number, data: any) => api.put(`/memories/${id}`, data),
  delete: (id: number) => api.delete(`/memories/${id}`),
  restore: (id: number) => api.post(`/memories/${id}/restore`),
  searchKeyword: (q: string, limit?: number) =>
    api.get('/memories/search/keyword', { params: { q, limit } }),
  searchSemantic: (q: string, limit?: number) =>
    api.get('/memories/search/semantic', { params: { q, limit } }),

  smartLoad: (params: { q?: string; token_budget?: number; load_level?: string }) =>
    api.get('/memories/smart-load', { params }),
  reinforce: (id: number) => api.post(`/memories/${id}/reinforce`),
  generateSummaries: () => api.post('/memories/generate-summaries'),

  scanOpenClaw: () => api.get('/openclaw-memories/scan'),
  scanOpenClawAgent: (agentName: string) => api.get(`/openclaw-memories/scan/${agentName}`),
  importOpenClaw: (data: any) => api.post('/openclaw-memories/import', data),
}
