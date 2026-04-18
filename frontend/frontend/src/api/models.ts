import api from './client'

export const modelsApi = {
  list: () => api.get('/models'),
  create: (data: { name: string; provider: string; model_id: string; api_key?: string; base_url?: string; config?: Record<string, any> }) => api.post('/models', data),
  get: (id: number) => api.get(`/models/${id}`),
  update: (id: number, data: { name?: string; provider?: string; model_id?: string; api_key?: string; base_url?: string; config?: Record<string, any> }) => api.put(`/models/${id}`, data),
  delete: (id: number) => api.delete(`/models/${id}`),
}
