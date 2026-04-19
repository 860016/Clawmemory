import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDark = ref(localStorage.getItem('clawmemory-theme') !== 'light')

  function toggle() {
    isDark.value = !isDark.value
    localStorage.setItem('clawmemory-theme', isDark.value ? 'dark' : 'light')
    applyTheme()
  }

  function applyTheme() {
    document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : 'light')
  }

  // Apply on init
  applyTheme()

  return { isDark, toggle, applyTheme }
})
