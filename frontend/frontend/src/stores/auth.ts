import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const username = ref('')
  const role = ref('')
  const user = ref<{ id: number; username: string; role: string; is_active: boolean } | null>(null)
  const initialized = ref(false)

  const isLoggedIn = computed(() => !!token.value)

  async function checkInit() {
    const resp = await authApi.getInitStatus()
    initialized.value = resp.data.initialized
    return initialized.value
  }

  async function initSystem(data: { username: string; password: string; email?: string }) {
    const resp = await authApi.init(data)
    token.value = resp.data.access_token
    role.value = resp.data.role
    username.value = resp.data.username
    localStorage.setItem('token', token.value!)
  }

  async function login(uname: string, pwd: string) {
    const resp = await authApi.login({ username: uname, password: pwd })
    token.value = resp.data.access_token
    role.value = resp.data.role
    username.value = resp.data.username
    localStorage.setItem('token', token.value!)
  }

  async function fetchMe() {
    const resp = await authApi.getMe()
    username.value = resp.data.username
    role.value = resp.data.role
    user.value = resp.data
  }

  async function checkAuth() {
    if (!token.value) return
    try {
      await fetchMe()
    } catch {
      token.value = null
      user.value = null
      localStorage.removeItem('token')
    }
  }

  function logout() {
    token.value = null
    username.value = ''
    role.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    try {
      await authApi.changePassword({ old_password: oldPassword, new_password: newPassword })
      return true
    } catch {
      return false
    }
  }

  return {
    token, username, role, user, initialized, isLoggedIn,
    checkInit, initSystem, login, fetchMe, checkAuth, logout, changePassword,
  }
})
