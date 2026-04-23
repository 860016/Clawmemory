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

  // OpenClaw memory import (Go 版暂不支持)
  scanOpenClaw: () => Promise.resolve({ data: [] }),
  scanOpenClawAgent: (_agentName: string) => Promise.resolve({ data: [] }),
  importOpenClaw: (_data: any) => Promise.resolve({ data: { imported: 0 } }),
}
