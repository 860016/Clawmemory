import axios from './client'

export default {
  getDecayStats() {
    return axios.get('/pro/decay/stats')
  },
  applyDecay() {
    return axios.post('/pro/decay/apply')
  },
  reinforceMemory(memoryId: number) {
    return axios.post(`/pro/reinforce/${memoryId}`)
  },
  getPruneSuggestions() {
    return axios.get('/pro/prune-suggest')
  },

  scanConflicts() {
    return axios.get('/pro/conflicts/scan')
  },
  resolveConflict(conflictIndex: number, strategy?: string) {
    return axios.post(`/pro/conflicts/resolve/${conflictIndex}`, {
      strategy: strategy || 'merge',
    })
  },

  routeToken(message: string, contextLength = 0) {
    return axios.post('/pro/token/route', null, {
      params: { message, context_length: contextLength },
    })
  },
  getTokenStats() {
    return axios.get('/pro/token/stats')
  },

  aiExtract(memoryIds?: number[]) {
    return axios.post('/pro/ai/extract', { memory_ids: memoryIds })
  },

  autoGraph(overwrite = false) {
    return axios.post('/pro/auto-graph', { overwrite })
  },

  getBackupSchedule() {
    return axios.get('/pro/backup/schedule')
  },
  setBackupSchedule(schedule: { enabled: boolean; interval_hours: number }) {
    return axios.post('/pro/backup/schedule', schedule)
  },

  compressPreview(level: 'light' | 'medium' | 'deep' = 'light') {
    return axios.post('/pro/compress/preview', { level })
  },
  compressApply(level: 'light' | 'medium' | 'deep' = 'light', options?: Record<string, any>) {
    return axios.post('/pro/compress/apply', { level, options: options || {} })
  },
  getCompressConfig() {
    return axios.get('/pro/compress/config')
  },
  setCompressConfig(config: Record<string, any>) {
    return axios.put('/pro/compress/config', config)
  },

  getEvolutionInsights() {
    return axios.get('/pro/evolution/insights')
  },
  discoverRelations() {
    return axios.post('/pro/evolution/discover')
  },
  inferChains() {
    return axios.post('/pro/evolution/infer')
  },
  getImportanceAdjustments() {
    return axios.post('/pro/evolution/importance')
  },
  prefetchMemories(context: string) {
    return axios.post('/pro/evolution/prefetch', { context })
  },
}
