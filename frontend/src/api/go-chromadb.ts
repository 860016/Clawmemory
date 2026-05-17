import api from './go-client'

export const chromadbApi = {
  getStatus: () => api.get('/chromadb/status'),
  sync: () => api.post('/chromadb/sync'),
}
