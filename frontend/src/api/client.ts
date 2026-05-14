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
    if (!error.response) {
      if (error.code === 'ECONNABORTED' || error.message?.includes('timeout')) {
        ElMessage.error('请求超时，请检查网络后重试')
      } else {
        ElMessage.error('网络连接异常，请检查网络设置')
      }
      return Promise.reject(error)
    }

    const status = error.response.status
    if (status === 401) {
      localStorage.removeItem('token')
      if (!window.location.pathname.endsWith('/login')) {
        window.location.href = '/login'
      }
    } else if (status === 403) {
      const msg = error.response?.data?.error || ''
      if (msg.includes('permission')) {
        ElMessage.error('权限不足：' + msg)
      }
    } else if (status === 429) {
      ElMessage.warning('操作过于频繁，请稍后再试')
    } else if (status === 405) {
      ElMessage.error('请求方法不允许')
    } else if (status >= 500) {
      ElMessage.error('服务器内部错误，请稍后重试')
    } else {
      let msg = error.response?.data?.error || error.response?.data?.detail || error.response?.data?.message || '请求失败，请稍后重试'
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
