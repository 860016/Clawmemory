import api from './client'

export const adminApi = {
  // Users
  listUsers: () => api.get('/users'),
  toggleUserActive: (id: number, is_active: boolean) => api.put(`/users/${id}/active`, { is_active }),

  // License
  getLicenseInfo: () => api.get('/license/info'),
  activateLicense: (key: string) => api.post('/license/activate', { key }),
  deactivateLicense: () => api.post('/license/deactivate'),
}
