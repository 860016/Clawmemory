import api from './client'

export const agentsApi = {
  list: () => api.get('/agents'),
  get: (id: number) => api.get(`/agents/${id}`),
  create: (data: { name: string; description?: string; model_id?: string }) => api.post('/agents', data),
  update: (id: number, data: { name?: string; description?: string; model_id?: string }) => api.put(`/agents/${id}`, data),
  delete: (id: number) => api.delete(`/agents/${id}`),
}
