import api from './client'

export const authApi = {
  getInitStatus: () => api.get('/auth/init-status'),
  init: (data: { username: string; password: string; email?: string }) => api.post('/auth/init', data),
  login: (data: { username: string; password: string }) => api.post('/auth/login', data),
  getMe: () => api.get('/auth/me'),
  updateMe: (data: { display_name?: string; email?: string }) => api.put('/auth/me', data),
  changePassword: (data: { old_password: string; new_password: string }) => api.put('/auth/change-password', data),
}
