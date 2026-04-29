import api from './go-client'

export const backupApi = {
  list: () => api.get('/backups').then(res => res.data.backups || []),
  create: () => api.post('/backups'),
  download: (filename: string) => api.get(`/backups/${filename}`, { responseType: 'blob' }),
  restore: (filename: string) => api.post(`/backups/${filename}/restore`),
  delete: (filename: string) => api.delete(`/backups/${filename}`),
}
