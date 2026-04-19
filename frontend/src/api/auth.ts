import axios from './client'

export const authApi = {
  getInitStatus: () => axios.get('/auth/init-status'),
  login: (data: { password: string }) => axios.post('/auth/login', data),
  setPassword: (data: { password: string }) => axios.post('/auth/set-password', data),
  getMe: () => axios.get('/auth/me'),
  resetPassword: () => axios.post('/auth/reset-password'),
  resetPasswordWithToken: (token: string) => axios.post('/auth/reset-password', { token }),
}
