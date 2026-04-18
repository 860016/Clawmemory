import api from './client'

export const nodesApi = {
  list: () => api.get('/nodes'),
  register: (data: { name: string; agent_id?: number; metadata?: Record<string, any> }) => api.post('/nodes', data),
  get: (id: number) => api.get(`/nodes/${id}`),
  updateStatus: (id: number, data: { status: string; metadata?: Record<string, any> }) => api.put(`/nodes/${id}/status`, data),
  delete: (id: number) => api.delete(`/nodes/${id}`),
}
