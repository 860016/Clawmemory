import api from './go-client'

export const backupApi = {
  list: () => api.get('/backups').then(res => res.data.backups || []),
  create: () => api.post('/backups'),
  download: (_filename: string) => Promise.reject(new Error('Download not implemented')),
  restore: (_filename: string) => Promise.reject(new Error('Restore not implemented')),
  delete: (_filename: string) => Promise.reject(new Error('Delete not implemented')),
}
