import api from './go-client'

export const reasoningApi = {
  getConfig() {
    return api.get('/reasoning/config')
  },

  updateConfig(data: Record<string, any>) {
    return api.put('/reasoning/config', data)
  },

  testConnection() {
    return api.post('/reasoning/test', {}, { timeout: 30000 })
  },

  execute(query: string, depth?: number, level?: string) {
    return api.post('/reasoning/execute', {
      query,
      depth: depth || 1,
      level: level || 'medium',
    }, { timeout: 120000 })
  },
}
