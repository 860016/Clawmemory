import api from './go-client'

export const wikiApi = {
  list: (params?: { category?: string; page?: number; size?: number }) =>
    api.get('/wiki', { params }),
  create: (data: any) => api.post('/wiki', data),
  get: (id: number) => api.get(`/wiki/${id}`),
  update: (id: number, data: any) => api.put(`/wiki/${id}`, data),
  delete: (id: number) => api.delete(`/wiki/${id}`),
}
