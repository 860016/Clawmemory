import api from './go-client'

export interface ShareRuleForm {
  from_agent?: string
  to_agent?: string
  layer?: string
  target_visibility?: string
  enabled?: boolean
  name?: string
  source_agent?: string
  target_agent?: string
  min_importance?: number
  auto_approve?: boolean
}

export const sharingApi = {
  shareMemory: (memoryId: number, toAgent: string, shareType?: string) =>
    api.post('/shares', { memory_id: memoryId, to_agent: toAgent, share_type: shareType || 'manual' }),
  getPendingShares: () => api.get('/shares/pending'),
  getOutboundShares: () => api.get('/shares/outbound'),
  approveShare: (id: number) => api.post(`/shares/${id}/approve`),
  rejectShare: (id: number) => api.post(`/shares/${id}/reject`),
  revokeShare: (id: number) => api.post(`/shares/${id}/revoke`),
  getAgentMemories: (agent: string) => api.get(`/shares/agent/${agent}`),

  listRules: () => api.get('/share-rules'),
  createRule: (data: ShareRuleForm) => api.post('/share-rules', data),
  updateRule: (id: number, data: Partial<ShareRuleForm>) => api.put(`/share-rules/${id}`, data),
  deleteRule: (id: number) => api.delete(`/share-rules/${id}`),
}

export const riskSwitchApi = {
  getSwitches: () => api.get('/risk-switches'),
  setSwitches: (switches: Record<string, boolean>) => api.put('/risk-switches', { switches }, { _silent: true } as any),
}

export const writebackApi = {
  getTargets: () => api.get('/writeback/targets'),
  preview: (agentName: string, projectPath?: string) =>
    api.post('/writeback/preview', { agent_name: agentName, project_path: projectPath }),
  execute: (agentName: string, projectPath?: string) =>
    api.post('/writeback/execute', { agent_name: agentName, project_path: projectPath }),
}
