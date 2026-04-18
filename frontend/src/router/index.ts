import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
    {
      path: '/', component: () => import('../views/MainLayout.vue'),
      children: [
        { path: '', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
        { path: 'memories', name: 'memories', component: () => import('../views/MemoriesView.vue') },
        { path: 'knowledge', name: 'knowledge', component: () => import('../views/KnowledgeView.vue') },
        { path: 'wiki', name: 'wiki', component: () => import('../views/WikiView.vue') },
        { path: 'settings', name: 'settings', component: () => import('../views/SettingsView.vue') },
      ],
    },
  ],
})

// 缓存是否需要密码
let passwordRequired: boolean | null = null

router.beforeEach(async (to, _from, next) => {
  // 登录页始终放行
  if (to.name === 'login') {
    next()
    return
  }

  const token = localStorage.getItem('token')
  if (token) {
    next()
    return
  }

  // 没有 token，检查是否需要密码
  if (passwordRequired === null) {
    try {
      const res = await fetch('/api/v1/auth/init-status')
      const data = await res.json()
      passwordRequired = data.password_set
    } catch {
      passwordRequired = false
    }
  }

  if (!passwordRequired) {
    // 不需要密码，直接放行
    next()
  } else {
    // 需要密码，跳转登录
    next({ name: 'login' })
  }
})

export default router
