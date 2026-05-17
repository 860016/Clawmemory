import api from './go-client'

export const openClawSkillsApi = {
  scan: () => api.get('/openclaw-skills/scan'),
  getDetail: (params: Record<string, any>) => api.get('/openclaw-skills/detail', { params }),
}
