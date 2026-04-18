import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token && token !== 'no-password') {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      // 不自动跳转，让页面自行处理
    } else if (error.response?.status !== 403) {
      // 403 由具体页面处理（如免费版限制）
      // 其他错误统一提示
      const msg = error.response?.data?.detail || error.response?.data?.message || 'Request failed'
      if (typeof msg === 'string' && !msg.includes('rate limit')) {
        ElMessage.error(msg)
      }
    }
    return Promise.reject(error)
  }
)

export default api
