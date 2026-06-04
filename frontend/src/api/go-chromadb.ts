import api from './go-client'

export const chromadbApi = {
  getStatus: () => api.get('/chromadb/status'),
  install: () => api.post('/chromadb/install', {}, { timeout: 120000 } as any),
  sync: () => api.post('/chromadb/sync'),
}
