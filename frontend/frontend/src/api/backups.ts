import api from './client'

export const backupsApi = {
  list: () => api.get('/backups'),
  create: () => api.post('/backups'),
  download: (id: number) => api.get(`/backups/${id}/download`, { responseType: 'blob' }),
  restore: (id: number) => api.post(`/backups/${id}/restore`),
  delete: (id: number) => api.delete(`/backups/${id}`),
  upload: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/backups/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
  },
}
