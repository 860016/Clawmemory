import axios from './go-client'

export const authApi = {
  getInitStatus: () => Promise.resolve({ data: { initialized: true } }),
  login: (data: { username?: string; password: string }) =>
    axios.post('/auth/login', { username: data.username || 'admin', password: data.password }),
  setPassword: (data: { password: string }) =>
    axios.post('/auth/register', { username: 'admin', password: data.password }),
  getMe: () => Promise.resolve({ data: { id: 1, username: 'admin' } }),
  resetPassword: (data: { old_password: string; new_password: string }) =>
    axios.post('/auth/reset-password', data),
  resetPasswordWithToken: (_token: string) =>
    Promise.resolve({ data: { message: '请使用旧密码重置' } }),
}
