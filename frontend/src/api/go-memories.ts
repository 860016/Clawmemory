import api from './go-client'

export const memoryApi = {
  list: (params?: { layer?: string; page?: number; size?: number; status?: string; memory_type?: string; source_agent?: string; visibility?: string }) =>
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

  smartLoad: (params: { q?: string; token_budget?: number; load_level?: string }) =>
    api.get('/memories/smart-load', { params }),
  reinforce: (id: number) => api.post(`/memories/${id}/reinforce`),
  generateSummaries: () => api.post('/memories/generate-summaries'),
  verify: (id: number) => api.post(`/memories/${id}/verify`),
  extract: (content: string) => api.post('/memories/extract', { content }),
  extractAndSave: (content: string, autoSave: boolean = false) =>
    api.post('/memories/extract-and-save', { content, auto_save: autoSave }),
  scanSecrets: (content: string) => api.post('/memories/scan-secrets', { content }),

  scanAgentMemories: () => api.get('/agent-memories/scan'),
  scanAgent: (agentName: string) => api.get(`/agent-memories/scan/${agentName}`),
  importAgentMemories: (data: any) => api.post('/agent-memories/import', data),
  getAgentSyncStatus: () => api.get('/agent-sync/status'),
  forceAgentSync: () => api.post('/agent-sync/force'),
  toggleAgentSync: (enabled: boolean) => api.post('/agent-sync/toggle', { enabled }),
  getAgentsMD: () => api.get('/agent/agents-md'),
}

export const sessionMemoryApi = {
  list: (params?: { session_id?: string; status?: string }) =>
    api.get('/session-memories', { params }),
  create: (data: any) => api.post('/session-memories', data),
  get: (id: number) => api.get(`/session-memories/${id}`),
  update: (id: number, data: any) => api.put(`/session-memories/${id}`, data),
  delete: (id: number) => api.delete(`/session-memories/${id}`),
}
