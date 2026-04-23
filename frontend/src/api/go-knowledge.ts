import api from './go-client'

export const knowledgeApi = {
  // 实体
  listEntities: (params?: { type?: string; page?: number; size?: number }) =>
    api.get('/knowledge/entities', { params }),
  createEntity: (data: any) => api.post('/knowledge/entities', data),
  updateEntity: (id: number, data: any) => api.put(`/knowledge/entities/${id}`, data),
  deleteEntity: (id: number) => api.delete(`/knowledge/entities/${id}`),

  // 关系
  listRelations: () => api.get('/knowledge/relations'),
  createRelation: (data: any) => api.post('/knowledge/relations', data),
  deleteRelation: (id: number) => api.delete(`/knowledge/relations/${id}`),

  // 图谱
  getGraph: () => api.get('/knowledge/graph'),

  // 统计
  getStats: () => api.get('/stats'),

  // AI 提取（Pro 功能）
  aiExtract: (_data: any) => Promise.reject(new Error('Pro feature not available')),

  // 消歧
  disambiguate: (_name: string) => Promise.resolve({ data: [] }),
}
