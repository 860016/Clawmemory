import api from './go-client'

export const memoryApi = {
  list: (params?: { layer?: string; page?: number; size?: number; status?: string; memory_type?: string; source_agent?: string; visibility?: string }) =>
    api.get('/memories', { params }),
  create: (data: any) => api.post('/memories', data),
  get: (id: number) => api.get(`/memories/${id}`),
  update: (id: number, data: any) => api.put(`/memories/${id}`, data),
  delete: (id: number) => api.delete(`/memories/${id}`),
  restore: (id: number) => api.post(`/memories/${id}/restore`),
  search: (q: string, mode?: string, limit?: number) =>
    api.get('/memories/search', { params: { q, mode: mode || 'keyword', limit } }),

  smartLoad: (params: { q?: string; token_budget?: number; load_level?: string }) =>
    api.get('/memories/smart-load', { params }),
  reinforce: (id: number) => api.post(`/memories/${id}/reinforce`),
  generateSummaries: () => api.post('/memories/generate-summaries'),
  verify: (id: number) => api.post(`/memories/${id}/verify`),
  extract: (content: string) => api.post('/memories/extract', { content }),
  extractAndSave: (content: string, autoSave: boolean = false) =>
    api.post('/memories/extract-and-save', { content, auto_save: autoSave }),
  scanSecrets: (content: string) => api.post('/memories/scan-secrets', { content }),
  batchValidate: () => api.post('/memories/validate'),

  listTemplates: () => api.get('/memories/templates'),
  createTemplate: (data: any) => api.post('/memories/templates', data),
  deleteTemplate: (name: string) => api.delete(`/memories/templates/${name}`),
  applyTemplate: (name: string, values: Record<string, string>) =>
    api.post(`/memories/templates/${name}/apply`, { values }),

  getDecaySettings: () => api.get('/memories/decay/settings'),
  getDecayStats: () => api.get('/memories/decay/stats'),
  applyDecay: () => api.post('/memories/decay/apply'),
  updateDecaySettings: (data: { enabled: boolean }) => api.put('/memories/decay/settings', data, { _silent: true } as any),
  getHealth: () => api.get('/memories/health'),
  assessQuality: () => api.get('/memories/quality'),
  autoFix: (issueTypes?: string[]) => api.post('/memories/auto-fix', { issue_types: issueTypes }),
  scanDedup: () => api.get('/memories/dedup/scan'),
  emptyTrash: () => api.delete('/memories/trash'),
  listTrash: (limit = 100) => api.get('/memories/trash', { params: { limit } }),

  getGovernanceStatus: () => api.get('/memories/governance/status'),
  runGovernance: () => api.post('/memories/governance/run'),
  updateGovernanceConfig: (config: any) => api.put('/memories/governance/config', config, { _silent: true } as any),

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
