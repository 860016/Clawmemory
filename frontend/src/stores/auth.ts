import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from '../api/go-client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const username = ref<string>(localStorage.getItem('cm_username') || '')
  const role = ref('user')

  const isLoggedIn = computed(() => !!token.value)

  async function login(user: string, password: string) {
    const { data } = await axios.post('/auth/login', { username: user || 'admin', password })
    token.value = data.access_token
    username.value = user || 'admin'
    role.value = data.role || 'user'
    localStorage.setItem('token', data.access_token)
    localStorage.setItem('cm_username', username.value)
  }

  async function register(user: string, password: string, invitationCode?: string) {
    const payload: any = { username: user, password }
    if (invitationCode) payload.invitation_code = invitationCode
    await axios.post('/auth/register-with-invitation', payload)
    await login(user, password)
  }

  async function setPassword(password: string) {
    const { data } = await axios.post('/auth/set-password', { password })
    token.value = data.access_token
    localStorage.setItem('token', data.access_token)
  }

  async function fetchMe() {
    try {
      const { data } = await axios.get('/auth/me')
      username.value = data.username || data.name || ''
      role.value = data.role || 'user'
      localStorage.setItem('cm_username', username.value)
    } catch {}
  }

  function logout() {
    token.value = null
    username.value = ''
    role.value = 'user'
    localStorage.removeItem('token')
    localStorage.removeItem('cm_username')
  }

  return {
    token, username, role, isLoggedIn,
    login, register, setPassword, fetchMe, logout,
  }
})
