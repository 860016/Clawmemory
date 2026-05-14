import api from './go-client'

export const agentApi = {
  getConnected: () => api.get('/agents/connected'),
  scanMemories: () => api.get('/agent-memories/scan'),
  scanAgentMemories: (agentName: string) => api.get(`/agent-memories/scan/${agentName}`),
  importMemories: (data: any) => api.post('/agent-memories/import', data),
  getSyncStatus: () => api.get('/agent-sync/status'),
  forceSync: () => api.post('/agent-sync/force'),
  toggleSync: (enabled: boolean) => api.post('/agent-sync/toggle', { enabled }),
  getAgentsMD: () => api.get('/agent/agents-md'),

  scanOpenClaw: () => api.get('/agent-memories/scan'),
  scanOpenClawAgent: (agentName: string) => api.get(`/agent-memories/scan/${agentName}`),
  importOpenClaw: (data: any) => api.post('/agent-memories/import', data),
}
