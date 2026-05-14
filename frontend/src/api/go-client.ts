import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  if (config.url && config.url.includes('/agent-memories/scan')) {
    config.timeout = 120000
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.code === 'ECONNABORTED' || error.code === 'ERR_CANCELED' || error.message?.includes('timeout')) {
      if (!error.config?._silent) {
        ElMessage.error('请求超时，请稍后重试')
      }
      return Promise.reject(error)
    }

    const status = error.response?.status
    if (status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('cm_username')
      if (!window.location.pathname.endsWith('/login') && !window.location.pathname.endsWith('/register')) {
        const current = window.location.pathname + window.location.search
        window.location.href = '/login?redirect=' + encodeURIComponent(current)
      }
    } else if (status === 403) {
      if (!error.config?._silent) {
        ElMessage.warning('权限不足，无法执行此操作')
      }
    } else if (status === 429) {
      ElMessage.warning('请求过于频繁，请稍后重试')
    } else {
      let msg = error.response?.data?.error || error.response?.data?.detail || error.response?.data?.message || 'Request failed'
      if (typeof msg === 'string') {
        if (msg.includes('Pro license required') || msg.includes('Pro feature not authorized')) {
          msg = '此功能需要 Pro 授权'
        }
        if (msg.includes('missing token') || msg.includes('invalid token')) {
          msg = '登录已过期，请重新登录'
        }
        if (!error.config?._silent) {
          ElMessage.error(msg)
        }
      }
    }
    return Promise.reject(error)
  }
)

export default api
