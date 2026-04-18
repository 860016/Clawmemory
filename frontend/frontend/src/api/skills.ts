import api from './client'

export const skillsApi = {
  list: () => api.get('/skills'),
  create: (data: { name: string; description?: string; config?: Record<string, any> }) => api.post('/skills', data),
  get: (id: number) => api.get(`/skills/${id}`),
  update: (id: number, data: { name?: string; description?: string; config?: Record<string, any> }) => api.put(`/skills/${id}`, data),
  delete: (id: number) => api.delete(`/skills/${id}`),
}
