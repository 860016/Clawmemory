import axios from './go-client'

export const authApi = {
  getInitStatus: () => axios.get('/auth/init-status'),
  login: (data: { username?: string; password: string }) =>
    axios.post('/auth/login', { username: data.username || 'admin', password: data.password }),
  setPassword: (data: { username?: string; password: string }) =>
    axios.post('/auth/set-password', { username: data.username, password: data.password }),
  getMe: () => axios.get('/auth/me'),
  resetPassword: (data: { old_password: string; new_password: string }) =>
    axios.post('/auth/reset-password', data),
  changePassword: (data: { old_password: string; new_password: string }) =>
    axios.post('/auth/change-password', data),
  forgotPassword: (data: { email?: string; username?: string; new_password?: string; confirm?: boolean }) =>
    axios.post('/auth/forgot-password', data),
  registerWithInvitation: (data: { username: string; password: string; invitation_code?: string }) =>
    axios.post('/auth/register-with-invitation', data),
}
