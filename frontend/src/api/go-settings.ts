import api from './go-client'

export const settingsApi = {
  get: () => api.get('/settings'),
  update: (data: Record<string, any>) => api.put('/settings', data, { _silent: true } as any),

  getApiKeys: () => api.get('/api-keys'),
  createApiKey: (data: { name: string; expires_in_days?: number }) => api.post('/api-keys', data, { _silent: true } as any),
  deleteApiKey: (id: number) => api.delete(`/api-keys/${id}`, { _silent: true } as any),

  getInvitations: () => api.get('/invitations'),
  createInvitation: (data: { max_uses?: number; expires_in_hours?: number }) => api.post('/invitations', data, { _silent: true } as any),
  deleteInvitation: (id: number) => api.delete(`/invitations/${id}`, { _silent: true } as any),

  listUsers: () => api.get('/users'),
  resetUserPassword: (data: { user_id?: number; username?: string; new_password: string }) => api.post('/users/reset-password', data, { _silent: true } as any),

  getInstallStatus: () => api.get('/install-status'),

  exportData: (password: string) => api.post('/data/export', { password }, { _silent: true } as any),
  importData: (data: any) => api.post('/data/import', data, { _silent: true } as any),
}
