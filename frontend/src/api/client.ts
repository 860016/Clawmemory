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
      // 403 由具体页面处理（如免费版限制）
    } else if (status === 405) {
      ElMessage.error('请求方法不允许')
    } else {
      let msg = error.response?.data?.detail || error.response?.data?.message || 'Request failed'
      if (typeof msg === 'string') {
        if (msg.includes('non-JSON response') || msg.includes('<html') || msg.includes('Pro server')) {
          msg = 'Pro 云服务暂不可用，已自动切换为本地模式'
        }
        if (msg.includes('Pro license required') || msg.includes('Pro feature not authorized')) {
          msg = '此功能需要 Pro 授权，当前使用本地基础模式'
        }
        if (msg.includes('Pro server unreachable')) {
          msg = 'Pro 云服务暂不可用，已自动切换为本地模式'
        }
        if (!msg.includes('rate limit')) {
          ElMessage.error(msg)
        }
      }
    }
    return Promise.reject(error)
  }
)

export default api
