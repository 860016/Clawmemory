import api from './go-client'

export const settingsApi = {
  get: () => api.get('/settings'),
  update: (data: Record<string, any>) => api.put('/settings', data),

  getApiKeys: () => api.get('/api-keys'),
  createApiKey: (data: { name: string; expires_in_days?: number }) => api.post('/api-keys', data),
  deleteApiKey: (id: number) => api.delete(`/api-keys/${id}`),

  getInvitations: () => api.get('/invitations'),
  createInvitation: (data: { max_uses?: number; expires_in_hours?: number }) => api.post('/invitations', data),
  deleteInvitation: (id: number) => api.delete(`/invitations/${id}`),

  listUsers: () => api.get('/users'),
  resetUserPassword: (data: { user_id?: number; username?: string; new_password: string }) => api.post('/users/reset-password', data),

  getInstallStatus: () => api.get('/install-status'),
  checkUpdate: () => api.get('/check-update'),

  exportData: (password: string) => api.post('/data/export', { password }),
  importData: (data: any) => api.post('/data/import', data),
}
