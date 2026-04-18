import api from './client'

export const memoryApi = {
  list: (params?: { layer?: string; page?: number; size?: number }) => api.get('/memories', { params }),
  create: (data: any) => api.post('/memories', data),
  get: (id: number) => api.get(`/memories/${id}`),
  update: (id: number, data: any) => api.put(`/memories/${id}`, data),
  delete: (id: number) => api.delete(`/memories/${id}`),
  searchKeyword: (q: string) => api.get('/memories/search/keyword', { params: { q } }),
  searchSemantic: (q: string) => api.get('/memories/search/semantic', { params: { q } }),
}
