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
        { path: 'skills', name: 'skills', component: () => import('../views/SkillsView.vue') },
        { path: 'wiki', name: 'wiki', component: () => import('../views/WikiView.vue') },
        { path: 'pro', name: 'pro', component: () => import('../views/ProView.vue') },
        { path: 'settings', name: 'settings', component: () => import('../views/SettingsView.vue') },
      ],
    },
  ],
})

router.beforeEach(async (to, _from, next) => {
  // Always allow login page
  if (to.name === 'login') {
    next()
    return
  }

  const token = localStorage.getItem('token')
  if (token) {
    next()
    return
  }

  // No token — always require login
  next({ name: 'login' })
})

export default router
