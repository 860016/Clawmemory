import api from './client'

export const knowledgeApi = {
  listEntities: (params?: { page?: number; size?: number }) => api.get('/knowledge/entities', { params }),
  createEntity: (data: { name: string; entity_type: string; properties?: Record<string, any>; description?: string }) => api.post('/knowledge/entities', data),
  getEntity: (id: number) => api.get(`/knowledge/entities/${id}`),
  updateEntity: (id: number, data: { name?: string; entity_type?: string; properties?: Record<string, any>; description?: string }) => api.put(`/knowledge/entities/${id}`, data),
  deleteEntity: (id: number) => api.delete(`/knowledge/entities/${id}`),

  listRelations: (params?: { page?: number; size?: number }) => api.get('/knowledge/relations', { params }),
  createRelation: (data: { source_id: number; target_id: number; relation_type: string; properties?: Record<string, any>; description?: string }) => api.post('/knowledge/relations', data),
  deleteRelation: (id: number) => api.delete(`/knowledge/relations/${id}`),

  getGraphData: (params?: { depth?: number }) => api.get('/knowledge/graph', { params }),
}
