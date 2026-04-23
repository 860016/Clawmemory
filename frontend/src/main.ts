import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './styles/global.css'
import './styles/design-system.css'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { useThemeStore } from './stores/theme'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(i18n)

// Initialize theme store
useThemeStore()

app.mount('#app')
