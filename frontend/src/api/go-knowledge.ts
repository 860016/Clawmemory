import api from './go-client'
import type { EntityCreateParams, EntityUpdateParams, RelationCreateParams } from './types'

export const knowledgeApi = {
  listEntities: (params?: { type?: string; page?: number; size?: number }) =>
    api.get('/knowledge/entities', { params }),
  createEntity: (data: EntityCreateParams) => api.post('/knowledge/entities', data),
  updateEntity: (id: number, data: EntityUpdateParams) => api.put(`/knowledge/entities/${id}`, data),
  deleteEntity: (id: number) => api.delete(`/knowledge/entities/${id}`),

  listRelations: () => api.get('/knowledge/relations'),
  createRelation: (data: RelationCreateParams) => api.post('/knowledge/relations', data),
  deleteRelation: (id: number) => api.delete(`/knowledge/relations/${id}`),

  getGraph: () => api.get('/knowledge/graph'),
}
