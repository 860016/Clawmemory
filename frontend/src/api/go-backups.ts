import api from './go-client'

export const backupApi = {
  list: () => api.get('/backups').then(res => res.data.backups || []),
  create: () => api.post('/backups'),
  download: (filename: string) => api.get(`/backups/${filename}/download`, { responseType: 'blob' }),
  restore: (_filename: string) => Promise.reject(new Error('Restore not implemented in Go version')),
  delete: (_filename: string) => Promise.reject(new Error('Delete not implemented in Go version')),
}
