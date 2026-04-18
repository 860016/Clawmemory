import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from '../api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const role = ref('admin')

  const isLoggedIn = computed(() => !!token.value)

  async function login(password: string) {
    const { data } = await axios.post('/auth/login', { password })
    token.value = data.access_token
    localStorage.setItem('token', data.access_token)
  }

  async function setPassword(password: string) {
    const { data } = await axios.post('/auth/set-password', { password })
    token.value = data.access_token
    localStorage.setItem('token', data.access_token)
  }

  function logout() {
    token.value = null
    localStorage.removeItem('token')
  }

  return {
    token, role, isLoggedIn,
    login, setPassword, logout,
  }
})
