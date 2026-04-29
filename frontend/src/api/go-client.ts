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
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    if (status === 401) {
      localStorage.removeItem('token')
      if (!window.location.pathname.endsWith('/login')) {
        window.location.href = '/login'
      }
    } else if (status === 403) {
      // 403 由具体页面处理
    } else {
      let msg = error.response?.data?.error || error.response?.data?.detail || error.response?.data?.message || 'Request failed'
      if (typeof msg === 'string') {
        if (msg.includes('Pro license required') || msg.includes('Pro feature not authorized')) {
          msg = '此功能需要 Pro 授权'
        }
        if (msg.includes('missing token') || msg.includes('invalid token')) {
          msg = '登录已过期，请重新登录'
        }
        if (!msg.includes('rate limit') && !error.config?._silent) {
          ElMessage.error(msg)
        }
      }
    }
    return Promise.reject(error)
  }
)

export default api
