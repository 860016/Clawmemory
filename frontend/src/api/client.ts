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
      // 跳转登录页（避免在登录页自身重复跳转）
      if (window.location.hash !== '#/login' && !window.location.pathname.endsWith('/login')) {
        window.location.href = '/#/login'
      }
    } else if (status === 403) {
      // 403 由具体页面处理（如免费版限制）
    } else if (status === 405) {
      ElMessage.error('请求方法不允许')
    } else {
      const msg = error.response?.data?.detail || error.response?.data?.message || 'Request failed'
      if (typeof msg === 'string' && !msg.includes('rate limit')) {
        ElMessage.error(msg)
      }
    }
    return Promise.reject(error)
  }
)

export default api
