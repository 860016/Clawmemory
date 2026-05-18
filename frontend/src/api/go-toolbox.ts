import axios from './go-client'

export const toolboxApi = {
  getDecayStats: () => axios.get('/toolbox/decay/stats'),
  applyDecay: () => axios.post('/toolbox/decay/apply'),
  reinforceMemory: (memoryId: number) => axios.post(`/toolbox/reinforce/${memoryId}`),
  getPruneSuggestions: () => axios.get('/toolbox/prune-suggest'),

  scanConflicts: () => axios.get('/toolbox/conflicts/scan'),
  resolveConflict: (conflictIndex: number, strategy?: string) =>
    axios.post(`/toolbox/conflicts/resolve/${conflictIndex}`, {
      strategy: strategy || 'merge',
    }),

  routeModel: (text: string, contextLength = 0) =>
    axios.get('/toolbox/token/route', {
      params: { message: text, context_length: contextLength },
    }),
  getTokenStats: () => axios.get('/toolbox/token/stats'),

  aiExtract: (memoryIds?: number[]) =>
    axios.post('/toolbox/ai/extract', { memory_ids: memoryIds }),

  autoGraph: (overwrite = false) =>
    axios.post('/toolbox/auto-graph', { overwrite }),

  getBackupSchedule: () => axios.get('/toolbox/backup/schedule'),
  setBackupSchedule: (schedule: { enabled: boolean; interval_hours: number }) =>
    axios.post('/toolbox/backup/schedule', schedule),

  compressPreview: (level: 'light' | 'medium' | 'deep' = 'light') =>
    axios.post('/toolbox/compress/preview', { level }),
  compressApply: (level: 'light' | 'medium' | 'deep' = 'light', options?: Record<string, any>) =>
    axios.post('/toolbox/compress/apply', { level, options: options || {} }),
  getCompressConfig: () => axios.get('/toolbox/compress/config'),
  setCompressConfig: (config: Record<string, any>) =>
    axios.put('/toolbox/compress/config', config),

  getEvolutionInsights: () => axios.get('/toolbox/evolution/insights'),
  discoverRelations: () => axios.post('/toolbox/evolution/discover'),
  inferChains: () => axios.post('/toolbox/evolution/infer'),
  getImportanceAdjustments: () => axios.post('/toolbox/evolution/importance'),
  prefetchMemories: (context: string) =>
    axios.post('/toolbox/evolution/prefetch', { context }),
}
