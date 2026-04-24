import axios from './client'

export const proApi = {
  getLicenseInfo: () => axios.get('/license/info'),
  activate: (licenseKey: string) =>
    axios.post('/license/activate', { license_key: licenseKey }),
  deactivate: () => axios.post('/license/deactivate'),

  getDecayStats: () => axios.get('/pro/decay/stats'),
  applyDecay: () => axios.post('/pro/decay/apply'),
  reinforceMemory: (memoryId: number) => axios.post(`/pro/reinforce/${memoryId}`),
  getPruneSuggestions: () => axios.get('/pro/prune-suggest'),

  scanConflicts: () => axios.get('/pro/conflicts/scan'),
  resolveConflict: (conflictIndex: number, strategy?: string) =>
    axios.post(`/pro/conflicts/resolve/${conflictIndex}`, {
      strategy: strategy || 'merge',
    }),

  routeModel: (text: string, contextLength = 0) =>
    axios.post('/pro/token/route', null, {
      params: { message: text, context_length: contextLength },
    }),
  getTokenStats: () => axios.get('/pro/token/stats'),

  aiExtract: (memoryIds?: number[]) =>
    axios.post('/pro/ai/extract', { memory_ids: memoryIds }),

  autoGraph: (overwrite = false) =>
    axios.post('/pro/auto-graph', { overwrite }),

  getBackupSchedule: () => axios.get('/pro/backup/schedule'),
  setBackupSchedule: (schedule: { enabled: boolean; interval_hours: number }) =>
    axios.post('/pro/backup/schedule', schedule),

  compressPreview: (level: 'light' | 'medium' | 'deep' = 'light') =>
    axios.post('/pro/compress/preview', { level }),
  compressApply: (level: 'light' | 'medium' | 'deep' = 'light', options?: Record<string, any>) =>
    axios.post('/pro/compress/apply', { level, options: options || {} }),
  getCompressConfig: () => axios.get('/pro/compress/config'),
  setCompressConfig: (config: Record<string, any>) =>
    axios.put('/pro/compress/config', config),

  getEvolutionInsights: () => axios.get('/pro/evolution/insights'),
  discoverRelations: () => axios.post('/pro/evolution/discover'),
  inferChains: () => axios.post('/pro/evolution/infer'),
  getImportanceAdjustments: () => axios.post('/pro/evolution/importance'),
  prefetchMemories: (context: string) =>
    axios.post('/pro/evolution/prefetch', { context }),
}
