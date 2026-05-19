<template>
  <div class="layout-v2" :class="{ 'sidebar-collapsed': sidebarCollapsed, 'mobile': isMobile }">
    <!-- Top Navigation Bar -->
    <header class="topbar-v2">
      <div class="topbar-left">
        <button class="menu-toggle" @click="sidebarCollapsed = !sidebarCollapsed">
          <el-icon :size="20"><Menu /></el-icon>
        </button>
        
        <div class="logo-v2" @click="$router.push('/')">
          <div class="logo-icon">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="white" stroke-width="2.5">
              <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z"/>
              <path d="M12 6v6l4 2"/>
            </svg>
          </div>
          <span class="logo-text">ClawMemory</span>
        </div>
      </div>

      <!-- Desktop Top Nav -->
      <nav class="topbar-nav-v2">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item-v2"
          :class="{ active: isNavActive(item.path) }"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ $t(item.label) }}</span>
        </router-link>
      </nav>

      <div class="topbar-right">
        <!-- Search Trigger -->
        <button class="icon-btn" @click="showSearch = true">
          <el-icon><Search /></el-icon>
        </button>
        
        <!-- Theme Toggle -->
        <button class="icon-btn theme-btn" @click="themeStore.toggle()">
          <el-icon v-if="themeStore.isDark"><Sunny /></el-icon>
          <el-icon v-else><Moon /></el-icon>
        </button>
        
        <!-- Tier Badge -->
        <div class="tier-badge-v2" :class="tierClass">
          <el-icon v-if="tier === 'advanced'"><Trophy /></el-icon>
          <span>{{ tierLabel }}</span>
        </div>
        
        <!-- User Menu -->
        <el-dropdown trigger="click" @command="handleUserCommand">
          <button class="user-avatar">
            <span class="avatar-text">{{ userInitial }}</span>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item disabled>
                <el-icon><User /></el-icon>
                {{ authStore.username || 'admin' }}
              </el-dropdown-item>
              <el-dropdown-item command="logout" divided>
                <el-icon><SwitchButton /></el-icon>
                {{ $t('common.logout') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <div class="layout-body-v2">
      <!-- Left Sidebar -->
      <aside class="sidebar-v2" v-if="currentSubNav.length && !isMobile" :class="{ 'collapsed': sidebarCollapsed }">
        <div class="sidebar-inner">
          <div class="sidebar-section" v-for="group in currentSubNav" :key="group.label">
            <div class="sidebar-group-title" v-if="group.label && !sidebarCollapsed">{{ $t(group.label) }}</div>
            <router-link
              v-for="item in group.items.filter((i: any) => !i.adminOnly || authStore.isFounder)"
              :key="item.path"
              :to="item.path"
              class="sidebar-item-v2"
              :class="{ active: isSubNavActive(item) }"
              :title="sidebarCollapsed ? $t(item.label) : ''"
            >
              <el-icon v-if="item.icon" :size="sidebarCollapsed ? 18 : 16"><component :is="item.icon" /></el-icon>
              <span v-if="!sidebarCollapsed">{{ $t(item.label) }}</span>
            </router-link>
          </div>
        </div>
        
        <div class="sidebar-footer" v-if="!sidebarCollapsed">
          <div class="storage-info">
            <div class="storage-bar">
              <div class="storage-fill" :style="{ width: storagePercent + '%' }"></div>
            </div>
            <span class="storage-text">{{ storagePercent }}% {{ $t('common.storageUsed') }}</span>
          </div>
        </div>

        <div class="sidebar-collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed">
          <el-icon :size="14">
            <DArrowLeft v-if="!sidebarCollapsed" />
            <DArrowRight v-else />
          </el-icon>
        </div>
      </aside>

      <!-- Mobile Sub Nav -->
      <div class="mobile-sub-nav-v2" v-if="currentSubNav.length && isMobile">
        <div class="mobile-sub-nav-inner">
          <router-link
            v-for="item in currentSubNavItems"
            :key="item.path"
            :to="item.path"
            class="mobile-sub-item-v2"
            :class="{ active: isSubNavActive(item) }"
          >
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <span>{{ $t(item.label) }}</span>
          </router-link>
        </div>
      </div>

      <!-- Main Content -->
      <main class="main-content-v2">
        <router-view />
      </main>
    </div>

    <!-- Mobile Bottom Tab Bar -->
    <nav class="mobile-tab-bar-v2" v-if="isMobile">
      <router-link
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="mobile-tab-item-v2"
        :class="{ active: isNavActive(item.path) }"
      >
        <div class="tab-icon">
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
        <span class="tab-label">{{ $t(item.label) }}</span>
      </router-link>
    </nav>

    <!-- Global Search Modal -->
    <el-dialog
      v-model="showSearch"
      width="600px"
      :show-close="false"
      class="search-modal"
      :fullscreen="isMobile"
    >
      <div class="search-container">
        <div class="search-input-wrapper">
          <el-icon class="search-icon"><Search /></el-icon>
          <input 
            v-model="searchQuery"
            :placeholder="$t('common.searchPlaceholder')"
            class="global-search-input"
            ref="searchInput"
          />
          <kbd class="shortcut">ESC</kbd>
        </div>
        <div class="search-results" v-if="searchQuery">
          <div v-if="searchLoading" class="search-loading">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>{{ $t('common.searching') }}</span>
          </div>
          <div v-else-if="searchResults.length === 0" class="search-empty">
            {{ $t('common.noSearchResults') }}
          </div>
          <template v-else>
            <div class="result-category" v-for="cat in searchResults" :key="cat.category">
              <div class="category-label">{{ cat.label }}</div>
              <div 
                v-for="item in cat.items" 
                :key="item.id"
                class="result-item"
                @click="navigateTo(item)"
              >
                <el-icon><component :is="item.icon" /></el-icon>
                <div class="result-info">
                  <div class="result-title">{{ item.title }}</div>
                  <div class="result-desc">{{ item.description }}</div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '../stores/theme'
import { useAuthStore } from '../stores/auth'
import {
  HomeFilled, Collection, Connection, Setting, Document, Promotion,
  Menu, User, Search, Sunny, Moon, Trophy, SwitchButton,
  DataAnalysis, MagicStick, Upload, Grid, Share,
  TrendCharts, Warning, Cpu, FolderOpened, Lock, Coin, Monitor,
  DocumentChecked, Star, Timer, Compass, CircleCheck, SuccessFilled, Memo,
  DArrowLeft, DArrowRight, Loading
} from '@element-plus/icons-vue'
import { statsApi } from '../api/go-stats'
import { memoryApi } from '../api/go-memories'
import { wikiApi } from '../api/go-wiki'
import projectApi from '../api/project'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const themeStore = useThemeStore()
const authStore = useAuthStore()
const userInitial = computed(() => (authStore.username || 'U').charAt(0).toUpperCase())
const sidebarCollapsed = ref(false)
const isMobile = ref(window.innerWidth <= 768)
const showSearch = ref(false)
const searchQuery = ref('')
const searchInput = ref<HTMLInputElement>()

const tier = ref('advanced')
const tierLabel = computed(() => t('tier.advanced'))
const tierClass = computed(() => 'tier-advanced')
const storagePercent = ref(0)

const handleResize = () => {
  isMobile.value = window.innerWidth <= 768
  if (!isMobile.value) {
    sidebarCollapsed.value = false
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  handleResize()
  statsApi.getOverview().then(({ data }) => {
    const total = data.memoryCount || 0
    const maxMemories = data.maxMemories || 50000
    storagePercent.value = Math.min(Math.round((total / maxMemories) * 100), 100)
  }).catch(() => {
    storagePercent.value = 0
  })
  
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('keydown', handleKeydown)
})

function handleKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    showSearch.value = true
    nextTick(() => searchInput.value?.focus())
  }
  if (e.key === 'Escape') {
    showSearch.value = false
  }
}

const navItems = [
  { path: '/', label: 'nav.dashboard', icon: HomeFilled },
  { path: '/memories', label: 'nav.memories', icon: Collection },
  { path: '/knowledge', label: 'nav.knowledge', icon: Connection },
  { path: '/projects', label: 'nav.projects', icon: Document },
  { path: '/advanced', label: 'nav.advanced', icon: Promotion },
  { path: '/settings', label: 'nav.settings', icon: Setting },
]

function isNavActive(path: string) {
  if (path === '/') return route.path === '/' || route.path === '/reports' || route.path === '/docs'
  return route.path.startsWith(path)
}

function isSubNavActive(item: { path: string }) {
  if (item.path.includes('?')) {
    return route.fullPath === item.path
  }
  return route.path === item.path && !Object.keys(route.query).length
}

const subNavMap: Record<string, Array<{ label?: string; items: Array<{ path: string; label: string; icon?: any; adminOnly?: boolean }> }>> = {
  '/': [
    { items: [
      { path: '/', label: 'nav.overview', icon: HomeFilled },
      { path: '/?tab=stats', label: 'nav.stats', icon: TrendCharts },
      { path: '/?tab=skills', label: 'nav.skills', icon: Cpu },
      { path: '/reports', label: 'nav.reports', icon: DataAnalysis },
      { path: '/docs', label: 'nav.guide', icon: DocumentChecked },
    ]}
  ],
  '/memories': [
    { items: [
      { path: '/memories', label: 'memories.all', icon: Collection },
      { path: '/memories?import=agent', label: 'memories.importAgentMemories', icon: Upload },
      { path: '/sharing', label: 'sharing.title', icon: Share },
      { path: '/session-memories', label: 'sessionMemory.title', icon: Memo },
    ]}
  ],
  '/knowledge': [],
  '/projects': [
    { items: [
      { path: '/projects', label: 'project.allProjects', icon: Document },
      { path: '/projects?status=active', label: 'project.active', icon: CircleCheck },
      { path: '/projects?status=completed', label: 'project.completed', icon: SuccessFilled },
    ]}
  ],
  '/advanced': [
    { items: [
      { path: '/advanced?section=decay', label: 'advanced.decay', icon: TrendCharts },
      { path: '/advanced?section=conflicts', label: 'advanced.conflicts', icon: Warning },
      { path: '/advanced?section=router', label: 'advanced.smartRouter', icon: Cpu },
      { path: '/advanced?section=tokenStats', label: 'advanced.tokenStats', icon: Coin },
      { path: '/advanced?section=extract', label: 'advanced.aiExtract', icon: MagicStick },
      { path: '/advanced?section=graph', label: 'advanced.autoGraph', icon: Connection },
    ]}
  ],
  '/settings': [
    { items: [
      { path: '/settings?section=ai', label: 'settings.aiConfig', icon: Cpu },
      { path: '/settings?section=security', label: 'settings.security', icon: Lock },
      { path: '/settings?section=risk-switches', label: 'settings.riskControl', icon: Warning },
      { path: '/settings?section=openclaw', label: 'settings.clientConnection', icon: Connection },
      { path: '/settings?section=data', label: 'settings.dataManagement', icon: Coin },
      { path: '/settings?section=decay', label: 'settings.memoryDecay', icon: Timer },
      { path: '/settings?section=system', label: 'settings.systemInfo', icon: Monitor },
      { path: '/settings?section=invitations', label: 'settings.invitationManage', icon: Share, adminOnly: true },
      { path: '/settings?section=users', label: 'settings.userManage', icon: User, adminOnly: true },
    ]}
  ],
}

const currentSubNav = computed(() => {
  const path = '/' + (route.path.split('/')[1] || '')
  if (path === '/reports' || path === '/docs') {
    return subNavMap['/'] || []
  }
  return subNavMap[path] || []
})

const currentSubNavItems = computed(() => {
  const items: Array<{ path: string; label: string; icon?: any; adminOnly?: boolean }> = []
  currentSubNav.value.forEach(group => {
    items.push(...group.items)
  })
  return items.filter(item => !item.adminOnly || authStore.isFounder)
})

const searchLoading = ref(false)
const searchResults = ref<Array<{ category: string; label: string; items: Array<{ id: number; title: string; description: string; icon: any; path: string }> }>>([])

let searchTimer: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, (q) => {
  if (searchTimer) clearTimeout(searchTimer)
  if (!q.trim()) {
    searchResults.value = []
    return
  }
  searchTimer = setTimeout(() => performSearch(q.trim()), 300)
})

async function performSearch(q: string) {
  searchLoading.value = true
  try {
    const [memRes, wikiRes, projRes] = await Promise.allSettled([
      memoryApi.searchKeyword(q, 5),
      wikiApi.search(q, 5),
      projectApi.search(q, 5),
    ])
    const results: typeof searchResults.value = []
    if (memRes.status === 'fulfilled' && memRes.value.data?.items?.length) {
      results.push({
        category: 'memories',
        label: t('nav.memories'),
        items: memRes.value.data.items.map((m: any) => ({
          id: m.id, title: m.key, description: (m.value || '').substring(0, 80), icon: Collection, path: '/memories'
        }))
      })
    }
    if (wikiRes.status === 'fulfilled' && wikiRes.value.data?.items?.length) {
      results.push({
        category: 'wiki',
        label: t('nav.knowledge'),
        items: wikiRes.value.data.items.map((w: any) => ({
          id: w.id, title: w.title, description: (w.content || '').substring(0, 80), icon: Connection, path: '/knowledge'
        }))
      })
    }
    if (projRes.status === 'fulfilled' && projRes.value.data?.items?.length) {
      results.push({
        category: 'projects',
        label: t('nav.advancedjects'),
        items: projRes.value.data.items.map((p: any) => ({
          id: p.id, title: p.name, description: (p.description || '').substring(0, 80), icon: Document, path: '/projects'
        }))
      })
    }
    searchResults.value = results
  } catch {
    searchResults.value = []
  } finally {
    searchLoading.value = false
  }
}

function navigateTo(item: any) {
  showSearch.value = false
  searchQuery.value = ''
  if (item.path) router.push(item.path)
}

function handleUserCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.layout-v2 {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--cm-bg-secondary);
  color: var(--cm-text-primary);
}

/* ===== Top Bar ===== */
.topbar-v2 {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 var(--cm-space-5);
  background: var(--cm-bg-primary);
  border-bottom: 1px solid var(--cm-border);
  position: sticky;
  top: 0;
  z-index: var(--cm-z-sticky);
  gap: var(--cm-space-4);
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  flex-shrink: 0;
}

.menu-toggle {
  width: 36px;
  height: 36px;
  border: none;
  background: var(--cm-bg-secondary);
  color: var(--cm-text-secondary);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--cm-transition-fast);
}

.menu-toggle:hover {
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-primary);
}

.logo-v2 {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  cursor: pointer;
  padding: var(--cm-space-1) var(--cm-space-2);
  border-radius: var(--cm-radius-md);
  transition: background var(--cm-transition-fast);
}

.logo-v2:hover {
  background: var(--cm-bg-secondary);
}

.logo-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--cm-radius-md);
  background: var(--cm-primary-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: var(--cm-text-primary);
  letter-spacing: -0.3px;
}

/* ===== Navigation ===== */
.topbar-nav-v2 {
  display: flex;
  align-items: center;
  gap: 2px;
  flex: 1;
  justify-content: center;
}

.nav-item-v2 {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
  padding: var(--cm-space-2) var(--cm-space-4);
  border-radius: var(--cm-radius-md);
  font-size: 14px;
  font-weight: 500;
  color: var(--cm-text-secondary);
  text-decoration: none;
  transition: all var(--cm-transition-fast);
  white-space: nowrap;
  position: relative;
}

.nav-item-v2:hover {
  color: var(--cm-text-primary);
  background: var(--cm-bg-secondary);
}

.nav-item-v2.active {
  color: var(--cm-primary);
  background: rgba(99, 102, 241, 0.08);
  font-weight: 600;
}

.nav-item-v2.active::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 20px;
  height: 3px;
  background: var(--cm-primary-gradient);
  border-radius: 2px;
}

/* ===== Top Bar Right ===== */
.topbar-right {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
  flex-shrink: 0;
}

.icon-btn {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  color: var(--cm-text-secondary);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--cm-transition-fast);
}

.icon-btn:hover {
  background: var(--cm-bg-secondary);
  color: var(--cm-text-primary);
}

.theme-btn:hover {
  color: var(--cm-primary);
}

.tier-badge-v2 {
  display: flex;
  align-items: center;
  gap: var(--cm-space-1);
  padding: var(--cm-space-1) var(--cm-space-3);
  border-radius: var(--cm-radius-full);
  font-size: 12px;
  font-weight: 600;
}

.tier-badge-v2.tier-advanced {
  background: var(--cm-primary-gradient);
  color: white;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 2px solid var(--cm-border);
  background: var(--cm-bg-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--cm-transition-fast);
}

.user-avatar:hover {
  border-color: var(--cm-primary);
}

.avatar-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text-secondary);
}

/* ===== Layout Body ===== */
.layout-body-v2 {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ===== Sidebar ===== */
.sidebar-v2 {
  width: 160px;
  flex-shrink: 0;
  background: var(--cm-bg-primary);
  border-right: 1px solid var(--cm-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width var(--cm-transition-normal);
  position: relative;
}

.sidebar-v2.collapsed {
  width: 48px;
}

.sidebar-inner {
  flex: 1;
  overflow-y: auto;
  padding: var(--cm-space-2);
}

.sidebar-section {
  margin-bottom: var(--cm-space-3);
}

.sidebar-group-title {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--cm-text-tertiary);
  padding: var(--cm-space-1) var(--cm-space-2);
  margin-bottom: var(--cm-space-1);
}

.sidebar-item-v2 {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
  padding: 6px 8px;
  border-radius: var(--cm-radius-md);
  font-size: 13px;
  color: var(--cm-text-secondary);
  text-decoration: none;
  transition: all var(--cm-transition-fast);
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
}

.sidebar-v2.collapsed .sidebar-item-v2 {
  justify-content: center;
  padding: 8px;
}

.sidebar-item-v2:hover {
  color: var(--cm-text-primary);
  background: var(--cm-bg-secondary);
}

.sidebar-item-v2.active {
  color: var(--cm-primary);
  background: rgba(99, 102, 241, 0.08);
  font-weight: 600;
}

.sidebar-footer {
  padding: var(--cm-space-3);
  border-top: 1px solid var(--cm-border);
}

.sidebar-collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-top: 1px solid var(--cm-border);
  cursor: pointer;
  color: var(--cm-text-tertiary);
  transition: all var(--cm-transition-fast);
}

.sidebar-collapse-btn:hover {
  color: var(--cm-text-primary);
  background: var(--cm-bg-secondary);
}

.storage-info {
  display: flex;
  flex-direction: column;
  gap: var(--cm-space-2);
}

.storage-bar {
  height: 4px;
  background: var(--cm-bg-tertiary);
  border-radius: var(--cm-radius-full);
  overflow: hidden;
}

.storage-fill {
  height: 100%;
  background: var(--cm-primary-gradient);
  border-radius: var(--cm-radius-full);
  transition: width 0.3s ease;
}

.storage-text {
  font-size: 12px;
  color: var(--cm-text-tertiary);
}

/* ===== Main Content ===== */
.main-content-v2 {
  flex: 1;
  overflow-y: auto;
  background: var(--cm-bg-secondary);
}

/* ===== Mobile Sub Nav ===== */
.mobile-sub-nav-v2 {
  display: none;
  background: var(--cm-bg-primary);
  border-bottom: 1px solid var(--cm-border);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

.mobile-sub-nav-v2::-webkit-scrollbar {
  display: none;
}

.mobile-sub-nav-inner {
  display: flex;
  gap: var(--cm-space-2);
  padding: var(--cm-space-3) var(--cm-space-4);
  min-width: max-content;
}

.mobile-sub-item-v2 {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
  padding: var(--cm-space-2) var(--cm-space-3);
  border-radius: var(--cm-radius-full);
  font-size: 13px;
  color: var(--cm-text-secondary);
  text-decoration: none;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  transition: all var(--cm-transition-fast);
  white-space: nowrap;
}

.mobile-sub-item-v2:hover {
  border-color: var(--cm-primary-light);
}

.mobile-sub-item-v2.active {
  color: var(--cm-primary);
  background: rgba(99, 102, 241, 0.08);
  border-color: var(--cm-primary);
  font-weight: 600;
}

/* ===== Mobile Bottom Tab Bar ===== */
.mobile-tab-bar-v2 {
  display: none;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: var(--cm-bg-primary);
  border-top: 1px solid var(--cm-border);
  z-index: var(--cm-z-sticky);
  justify-content: space-around;
  align-items: center;
  padding: 0 var(--cm-space-2);
  padding-bottom: env(safe-area-inset-bottom, 0);
}

.mobile-tab-item-v2 {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: var(--cm-space-1) var(--cm-space-2);
  border-radius: var(--cm-radius-md);
  color: var(--cm-text-tertiary);
  text-decoration: none;
  font-size: 11px;
  transition: all var(--cm-transition-fast);
  flex: 1;
  max-width: 80px;
}

.mobile-tab-item-v2:hover {
  color: var(--cm-text-secondary);
}

.mobile-tab-item-v2.active {
  color: var(--cm-primary);
}

.tab-icon {
  font-size: 22px;
  line-height: 1;
}

.tab-label {
  font-size: 10px;
  line-height: 1.2;
  font-weight: 500;
}

/* ===== Search Modal ===== */
.search-modal :deep(.el-dialog__header) {
  display: none;
}

.search-modal :deep(.el-dialog__body) {
  padding: 0;
}

.search-container {
  padding: var(--cm-space-4);
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: var(--cm-space-4);
  color: var(--cm-text-tertiary);
  font-size: 18px;
}

.global-search-input {
  width: 100%;
  padding: var(--cm-space-3) var(--cm-space-4) var(--cm-space-3) 44px;
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-lg);
  background: var(--cm-bg-secondary);
  color: var(--cm-text-primary);
  font-size: 16px;
  outline: none;
  transition: all var(--cm-transition-fast);
}

.global-search-input:focus {
  border-color: var(--cm-primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.search-results {
  margin-top: var(--cm-space-4);
  max-height: 400px;
  overflow-y: auto;
}

.result-category {
  margin-bottom: var(--cm-space-4);
}

.category-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--cm-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: var(--cm-space-2) var(--cm-space-3);
}

.result-item {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  padding: var(--cm-space-3);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  transition: background var(--cm-transition-fast);
}

.result-item:hover {
  background: var(--cm-bg-secondary);
}

.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  color: var(--cm-text-tertiary);
  font-size: 14px;
}

.search-empty {
  text-align: center;
  padding: 24px;
  color: var(--cm-text-tertiary);
  font-size: 14px;
}

.result-info {
  flex: 1;
}

.result-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--cm-text-primary);
}

.result-desc {
  font-size: 12px;
  color: var(--cm-text-tertiary);
}

/* ===== Responsive ===== */
@media (max-width: 1024px) {
  .nav-item-v2 span {
    display: none;
  }
  
  .nav-item-v2 {
    padding: var(--cm-space-2);
  }
  
  .logo-text {
    display: none;
  }
  
  .sidebar-v2.collapsed {
    width: 0;
    overflow: hidden;
  }
}

@media (max-width: 768px) {
  .layout-v2 {
    padding-bottom: 64px;
  }

  .topbar-v2 {
    height: 56px;
    padding: 0 var(--cm-space-3);
  }

  .topbar-nav-v2 {
    display: none;
  }

  .sidebar-v2 {
    display: none;
  }

  .mobile-sub-nav-v2 {
    display: block;
  }

  .mobile-tab-bar-v2 {
    display: flex;
  }

  .main-content-v2 {
    padding: 0;
  }

  .tier-badge-v2 span {
    display: none;
  }
  
  .tier-badge-v2 {
    padding: var(--cm-space-1);
  }
}

@media (max-width: 480px) {
  .layout-v2 {
    padding-bottom: 56px;
  }

  .topbar-v2 {
    height: 52px;
  }

  .mobile-tab-bar-v2 {
    height: 56px;
  }
  
  .tab-label {
    font-size: 9px;
  }
}
</style>
