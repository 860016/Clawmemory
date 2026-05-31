import axios from './go-client'

export const toolboxApi = {
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

  compressPreview: (level: 'light' | 'medium' | 'deep' = 'light') =>
    axios.post('/toolbox/compress/preview', { level }),
  compressApply: (level: 'light' | 'medium' | 'deep' = 'light', options?: Record<string, any>) =>
    axios.post('/toolbox/compress/apply', { level, options: options || {} }),
  getCompressConfig: () => axios.get('/toolbox/compress/config'),
  setCompressConfig: (config: Record<string, any>) =>
    axios.put('/toolbox/compress/config', config),

  getEvolutionInsights: () => axios.get('/memories/evolution/insights'),
  discoverRelations: () => axios.post('/memories/evolution/run', { action: 'discover' }),
  inferChains: () => axios.post('/memories/evolution/run', { action: 'infer' }),
  getImportanceAdjustments: () => axios.post('/memories/evolution/run', { action: 'importance' }),
}
