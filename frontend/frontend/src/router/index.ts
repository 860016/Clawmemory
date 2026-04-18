import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/setup', name: 'setup', component: () => import('../views/SetupView.vue') },
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
    {
      path: '/', component: () => import('../views/MainLayout.vue'),
      children: [
        { path: '', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
        { path: 'chat', name: 'chat', component: () => import('../views/ChatView.vue') },
        { path: 'memories', name: 'memories', component: () => import('../views/MemoriesView.vue') },
        { path: 'knowledge', name: 'knowledge', component: () => import('../views/KnowledgeView.vue') },
        { path: 'agents', name: 'agents', component: () => import('../views/AgentsView.vue') },
        { path: 'settings', name: 'settings', component: () => import('../views/SettingsView.vue') },
        { path: 'admin', name: 'admin', component: () => import('../views/AdminView.vue'), meta: { requiresAdmin: true } },
      ],
    },
  ],
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.name !== 'login' && to.name !== 'setup' && !token) {
    next({ name: 'login' })
  } else {
    next()
  }
})

export default router
