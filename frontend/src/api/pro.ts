import axios from './client'

export default {
  // Memory Decay
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

  // Conflict Resolution
  scanConflicts() {
    return axios.get('/pro/conflicts/scan')
  },
  resolveConflict(conflictIndex: number, strategy?: string) {
    return axios.post(`/pro/conflicts/resolve/${conflictIndex}`, null, {
      params: strategy ? { strategy } : {},
    })
  },

  // Token Router
  routeToken(message: string, contextLength = 0) {
    return axios.post('/pro/token/route', null, {
      params: { message, context_length: contextLength },
    })
  },
  getTokenStats() {
    return axios.get('/pro/token/stats')
  },

  // AI Extract
  aiExtract(memoryIds?: number[]) {
    return axios.post('/pro/ai/extract', { memory_ids: memoryIds })
  },

  // Auto Graph
  autoGraph(overwrite = false) {
    return axios.post('/pro/auto-graph', { overwrite })
  },

  // Auto Backup
  getBackupSchedule() {
    return axios.get('/pro/backup/schedule')
  },
  setBackupSchedule(schedule: { enabled: boolean; interval_hours: number }) {
    return axios.post('/pro/backup/schedule', schedule)
  },
}
