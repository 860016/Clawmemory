import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((cb) => cb(token))
  refreshSubscribers = []
}

function addRefreshSubscriber(cb: (token: string) => void) {
  refreshSubscribers.push(cb)
}

async function tryRefreshToken(): Promise<string | null> {
  const refreshToken = localStorage.getItem('refresh_token')
  if (!refreshToken) return null

  try {
    const { data } = await axios.post('/api/v1/auth/refresh', {
      refresh_token: refreshToken,
    })
    localStorage.setItem('token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    return data.access_token
  } catch {
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('cm_username')
    return null
  }
}

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
  async (error) => {
    const originalRequest = error.config

    if (error.code === 'ECONNABORTED' || error.code === 'ERR_CANCELED' || error.message?.includes('timeout')) {
      if (!originalRequest?._silent) {
        ElMessage.error('请求超时，请稍后重试')
      }
      return Promise.reject(error)
    }

    const status = error.response?.status
    if (status === 401) {
      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken && !originalRequest._retry) {
        originalRequest._retry = true

        if (!isRefreshing) {
          isRefreshing = true
          const newToken = await tryRefreshToken()
          isRefreshing = false

          if (newToken) {
            onTokenRefreshed(newToken)
            originalRequest.headers.Authorization = `Bearer ${newToken}`
            return api(originalRequest)
          }
        } else {
          return new Promise((resolve) => {
            addRefreshSubscriber((token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`
              resolve(api(originalRequest))
            })
          })
        }
      }

      localStorage.removeItem('token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('cm_username')
      if (!window.location.pathname.endsWith('/login') && !window.location.pathname.endsWith('/register')) {
        const current = window.location.pathname + window.location.search
        window.location.href = '/login?redirect=' + encodeURIComponent(current)
      }
    } else if (status === 403) {
      if (!originalRequest?._silent) {
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
        if (!originalRequest?._silent) {
          ElMessage.error(msg)
        }
      }
    }
    return Promise.reject(error)
  }
)

export default api
