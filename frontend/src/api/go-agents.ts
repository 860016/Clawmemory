import api from './go-client'

export const agentApi = {
  getConnected: () => api.get('/agents/connected'),
  scanMemories: () => api.get('/agent-memories/scan'),
  scanAgentMemories: (agentName: string) => api.get(`/agent-memories/scan/${agentName}`),
  importMemories: (data: any) => api.post('/agent-memories/import', data, { _silent: true } as any),
  getSyncStatus: () => api.get('/agent-sync/status'),
  forceSync: () => api.post('/agent-sync/force', {}, { _silent: true } as any),
  toggleSync: (enabled: boolean) => api.post('/agent-sync/toggle', { enabled }, { _silent: true } as any),
  getAgentsMD: () => api.get('/agent/agents-md'),
}
