import api from './go-client'

export const skillsApi = {
  recordAction(data: {
    session_id?: string
    agent_name?: string
    platform?: string
    action_type: string
    action_name: string
    parameters?: string
    result?: string
    duration?: number
  }) {
    return api.post('/skills/actions', data)
  },

  recordActionBatch(data: {
    session_id?: string
    agent_name?: string
    platform?: string
    actions: Array<{
      action_type: string
      action_name: string
      parameters?: Record<string, unknown>
      result?: string
      duration?: number
    }>
  }) {
    return api.post('/skills/actions/batch', data)
  },

  detectPatterns() {
    return api.get('/skills/detect')
  },

  createSkill(data: { use_ai?: boolean; patterns?: Array<{ pattern: string; category?: string }> }) {
    return api.post('/skills/create', data, { timeout: 120000 })
  },

  listSkills(status?: string) {
    const params = status ? { params: { status } } : {}
    return api.get('/skills/list', params)
  },

  matchSkill(query: string) {
    return api.get('/skills/match', { params: { q: query } })
  },

  patchSkill(id: number, data: { field: string; old_value: string; new_value: string }) {
    return api.patch(`/skills/${id}`, data)
  },

  improveSkill(id: number) {
    return api.post(`/skills/${id}/improve`, {}, { timeout: 120000 })
  },

  recordSkillUsage(id: number, success: boolean) {
    return api.post(`/skills/${id}/usage`, { success })
  },

  getSuggestions() {
    return api.get('/skills/suggestions')
  },

  generateSuggestions() {
    return api.post('/skills/suggestions/generate')
  },

  dismissSuggestion(id: number) {
    return api.post(`/skills/suggestions/${id}/dismiss`)
  },
}
