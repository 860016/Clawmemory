import api from './go-client'
import type {
  MemoryCreateParams,
  MemoryUpdateParams,
  MemorySearchParams,
  MemoryListParams,
  SmartLoadParams,
  DecaySettings,
  SearchResult,
} from './types'

export const memoryApi = {
  list: (params?: MemoryListParams) =>
    api.get('/memories', { params }),
  create: (data: MemoryCreateParams) => api.post('/memories', data),
  get: (id: number) => api.get(`/memories/${id}`),
  update: (id: number, data: MemoryUpdateParams) => api.put(`/memories/${id}`, data),
  delete: (id: number) => api.delete(`/memories/${id}`),
  restore: (id: number) => api.post(`/memories/${id}/restore`),
  search: (params: MemorySearchParams) =>
    api.get('/memories/search', { params: { q: params.q, mode: params.mode || 'keyword', limit: params.limit } }),

  smartLoad: (params: SmartLoadParams) =>
    api.get('/memories/smart-load', { params }),
  reinforce: (id: number) => api.post(`/memories/${id}/reinforce`),
  generateSummaries: () => api.post('/memories/generate-summaries'),
  verify: (id: number) => api.post(`/memories/${id}/verify`),
  extract: (content: string) => api.post('/memories/extract', { content }),
  extractAndSave: (content: string, autoSave: boolean = false) =>
    api.post('/memories/extract-and-save', { content, auto_save: autoSave }),
  scanSecrets: (content: string) => api.post('/memories/scan-secrets', { content }),
  batchValidate: () => api.post('/memories/validate'),

  getDecaySettings: () => api.get('/memories/decay/settings'),
  getDecayStats: () => api.get('/memories/decay/stats'),
  applyDecay: () => api.post('/memories/decay/apply'),
  updateDecaySettings: (data: DecaySettings) => api.put('/memories/decay/settings', data, { _silent: true } as any),
  getHealth: () => api.get('/memories/health'),
  assessQuality: () => api.get('/memories/quality'),
  autoFix: (issueTypes?: string[]) => api.post('/memories/auto-fix', { issue_types: issueTypes }),
  scanDedup: () => api.get('/memories/dedup/scan'),
  mergeDedup: (sourceId: number, targetId: number) => api.post('/memories/dedup/merge', { source_id: sourceId, target_id: targetId }),
  emptyTrash: () => api.delete('/memories/trash'),
  listTrash: (limit = 100) => api.get('/memories/trash', { params: { limit } }),

  getGovernanceStatus: () => api.get('/memories/governance/status'),
  runGovernance: () => api.post('/memories/governance/run'),
  updateGovernanceConfig: (config: Record<string, unknown>) => api.put('/memories/governance/config', config, { _silent: true } as any),

  getEvolutionInsights: () => api.get('/memories/evolution/insights'),
  runEvolution: (action: string, context?: string) => api.post('/memories/evolution/run', { action, context }),
  getGraphReasoning: () => api.get('/memories/evolution/graph-reasoning'),
  getCentrality: () => api.get('/memories/evolution/centrality'),
  getCommunities: () => api.get('/memories/evolution/communities'),
  communitiesToWiki: () => api.post('/memories/evolution/communities-to-wiki'),

  scanAgentMemories: () => api.get('/agent-memories/scan'),
  scanAgent: (agentName: string) => api.get(`/agent-memories/scan/${agentName}`),
  importAgentMemories: (data: Record<string, unknown>) => api.post('/agent-memories/import', data),
  getAgentSyncStatus: () => api.get('/agent-sync/status'),
  forceAgentSync: () => api.post('/agent-sync/force'),
  toggleAgentSync: (enabled: boolean) => api.post('/agent-sync/toggle', { enabled }),
  getAgentsMD: () => api.get('/agent/agents-md'),
  getConnectedAgents: () => api.get('/agents/connected'),
}

export const sessionMemoryApi = {
  list: (params?: { session_id?: string; status?: string }) =>
    api.get('/session-memories', { params }),
  create: (data: Record<string, unknown>) => api.post('/session-memories', data),
  get: (id: number) => api.get(`/session-memories/${id}`),
  update: (id: number, data: Record<string, unknown>) => api.put(`/session-memories/${id}`, data),
  delete: (id: number) => api.delete(`/session-memories/${id}`),
}
