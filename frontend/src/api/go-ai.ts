import api from './go-client'

export const aiApi = {
  getConfig() {
    return api.get('/ai/config')
  },

  updateConfig(data: Record<string, any>) {
    return api.put('/ai/config', data, { _silent: true } as any)
  },

  testConnection() {
    return api.post('/ai/test')
  },

  getUsage() {
    return api.get('/ai/usage')
  },

  getProviders() {
    return api.get('/ai/providers')
  },

  extract() {
    return api.post('/ai/extract', {}, { timeout: 120000 })
  },

  conflictScan() {
    return api.post('/ai/conflict-scan', {}, { timeout: 120000 })
  },

  decayEvaluate() {
    return api.post('/ai/decay-evaluate', {}, { timeout: 120000 })
  },

  generateDailyReport(date?: string) {
    const params = date ? { params: { date } } : {}
    return api.get('/ai/daily-report', { ...params, timeout: 120000 } as any)
  },

  generateWiki(topic: string) {
    return api.post('/ai/wiki-generate', { topic }, { timeout: 120000 })
  },

  compress(memoryIds: number[]) {
    return api.post('/ai/compress', { memory_ids: memoryIds }, { timeout: 120000 })
  },

  discoverRelations() {
    return api.post('/ai/discover-relations', {}, { timeout: 120000 })
  },

  discoverProjects() {
    return api.post('/ai/discover-projects', {}, { timeout: 120000 })
  },

  smartRoute(text: string) {
    return api.post('/ai/smart-route', { text }, { timeout: 60000 })
  },

  nudgeReflect() {
    return api.post('/ai/nudge-reflect', {}, { timeout: 120000 })
  },

  selfRefine(pressureLevel: 'low' | 'medium' | 'high' = 'medium') {
    return api.post('/ai/self-refine', { pressure_level: pressureLevel }, { timeout: 120000 })
  },

  buildUserProfile() {
    return api.post('/ai/user-profile', {}, { timeout: 120000 })
  },

  extractFacts(messages: Array<{ role: string; content: string }>) {
    return api.post('/ai/extract-facts', { messages }, { timeout: 120000 })
  },

  consolidate(newFacts: Array<{ key: string; value: string; layer?: string; importance?: number }>) {
    return api.post('/ai/consolidate', { facts: newFacts }, { timeout: 120000 })
  },

  processConversation(messages: Array<{ role: string; content: string }>) {
    return api.post('/ai/process-conversation', { messages }, { timeout: 120000 })
  },

  assembleContext(query: string, tokenBudget?: number) {
    return api.post('/ai/context-assemble', { query, token_budget: tokenBudget || 4000 }, { timeout: 120000 })
  },
}
