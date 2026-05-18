import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
    { path: '/register', name: 'register', component: () => import('../views/LoginView.vue') },
    { path: '/reset-password', name: 'reset-password', component: () => import('../views/LoginView.vue') },
    {
      path: '/', component: () => import('../views/MainLayout.vue'),
      children: [
        { path: '', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
        { path: 'memories', name: 'memories', component: () => import('../views/MemoriesView.vue') },
        { path: 'session-memories', name: 'session-memories', component: () => import('../views/SessionMemoryView.vue') },
        { path: 'sharing', name: 'sharing', component: () => import('../views/SharingView.vue') },
        { path: 'knowledge', name: 'knowledge', component: () => import('../views/KnowledgeViewV2.vue') },
        { path: 'skills', name: 'skills', component: () => import('../views/SkillsView.vue') },
        { path: 'wiki', name: 'wiki', component: () => import('../views/WikiView.vue') },
        { path: 'projects', name: 'projects', component: () => import('../views/ProjectView.vue') },
        { path: 'reports', name: 'reports', component: () => import('../views/DailyReportViewV2.vue') },
        { path: 'docs', name: 'docs', component: () => import('../views/UserGuideView.vue') },
        { path: 'advanced', name: 'advanced', component: () => import('../views/ProView.vue') },
        { path: 'settings', name: 'settings', component: () => import('../views/SettingsView.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'not-found', redirect: '/' },
  ],
})

router.beforeEach(async (to, _from, next) => {
  // Always allow login and reset-password pages
  if (to.name === 'login' || to.name === 'reset-password' || to.name === 'register') {
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
