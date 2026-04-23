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
          <el-icon v-if="tier === 'pro'"><Trophy /></el-icon>
          <span>{{ tierLabel }}</span>
        </div>
        
        <!-- User Menu -->
        <el-dropdown trigger="click" @command="handleUserCommand">
          <button class="user-avatar">
            <span class="avatar-text">U</span>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="settings">
                <el-icon><Setting /></el-icon>
                {{ $t('nav.settings') }}
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
            <div class="sidebar-group-title" v-if="group.label">{{ $t(group.label) }}</div>
            <router-link
              v-for="item in group.items"
              :key="item.path"
              :to="item.path"
              class="sidebar-item-v2"
              :class="{ active: isSubNavActive(item) }"
            >
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ $t(item.label) }}</span>
            </router-link>
          </div>
        </div>
        
        <!-- Sidebar Footer -->
        <div class="sidebar-footer">
          <div class="storage-info">
            <div class="storage-bar">
              <div class="storage-fill" style="width: 45%"></div>
            </div>
            <span class="storage-text">45% {{ $t('common.storageUsed') }}</span>
          </div>
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
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '../stores/theme'
import {
  HomeFilled, Collection, Connection, Setting, Document, Promotion,
  Menu, User, Search, Sunny, Moon, Trophy, SwitchButton,
  DataAnalysis, MagicStick, Upload, Grid, Share,
  TrendCharts, Warning, Cpu, FolderOpened, Lock, Coin, Monitor,
  DocumentChecked, Star, Lightbulb, Timer, Compass
} from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const themeStore = useThemeStore()
const sidebarCollapsed = ref(false)
const isMobile = ref(window.innerWidth <= 768)
const showSearch = ref(false)
const searchQuery = ref('')
const searchInput = ref<HTMLInputElement>()

const tier = ref('oss')
const tierLabel = computed(() => tier.value === 'oss' ? t('tier.free') : t('tier.pro'))
const tierClass = computed(() => tier.value === 'oss' ? 'tier-free' : 'tier-pro')

// Handle window resize
const handleResize = () => {
  isMobile.value = window.innerWidth <= 768
  if (!isMobile.value) {
    sidebarCollapsed.value = false
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  handleResize()
  try {
    axios.get('/license/info').then(({ data }) => {
      tier.value = data.tier || 'oss'
    }).catch(() => {
      tier.value = 'oss'
    })
  } catch {
    tier.value = 'oss'
  }
  
  // Keyboard shortcut for search
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
  { path: '/wiki', label: 'nav.wiki', icon: Document },
  { path: '/pro', label: 'nav.pro', icon: Promotion },
]

function isNavActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

function isSubNavActive(item: { path: string }) {
  if (item.path.includes('?')) {
    return route.fullPath === item.path
  }
  return route.path === item.path && !Object.keys(route.query).length
}

// Sub-navigation
const subNavMap: Record<string, Array<{ label?: string; items: Array<{ path: string; label: string; icon?: any }> }>> = {
  '/': [
    { items: [
      { path: '/', label: 'nav.overview', icon: HomeFilled },
      { path: '/reports', label: 'nav.reports', icon: DataAnalysis },
      { path: '/guide', label: 'nav.guide', icon: Document },
    ]}
  ],
  '/memories': [
    { items: [
      { path: '/memories', label: 'memories.all', icon: Collection },
      { path: '/memories?import=openclaw', label: 'memories.importOpenClaw', icon: Upload },
    ]}
  ],
  '/knowledge': [
    { items: [
      { path: '/knowledge', label: 'knowledge.entities', icon: Grid },
      { path: '/knowledge?tab=relations', label: 'knowledge.relations', icon: Share },
      { path: '/knowledge?tab=graph', label: 'knowledge.graphView', icon: Connection },
    ]}
  ],
  '/wiki': [
    { items: [
      { path: '/wiki', label: 'wiki.allPages', icon: Document },
      { path: '/wiki?tab=categories', label: 'wiki.categories', icon: FolderOpened },
    ]}
  ],
  '/pro': [
    { items: [
      { path: '/pro?section=decay', label: 'pro.decay', icon: TrendCharts },
      { path: '/pro?section=conflicts', label: 'pro.conflicts', icon: Warning },
      { path: '/pro?section=router', label: 'pro.tokenRouter', icon: Cpu },
      { path: '/pro?section=extract', label: 'pro.aiExtract', icon: MagicStick },
      { path: '/pro?section=graph', label: 'pro.autoGraph', icon: Connection },
    ]}
  ],
  '/settings': [
    { items: [
      { path: '/settings?section=license', label: 'settings.license', icon: Promotion },
      { path: '/settings?section=security', label: 'settings.security', icon: Lock },
      { path: '/settings?section=data', label: 'settings.data', icon: Coin },
      { path: '/settings?section=system', label: 'settings.system', icon: Monitor },
    ]}
  ],
}

const currentSubNav = computed(() => {
  const path = '/' + (route.path.split('/')[1] || '')
  return subNavMap[path] || []
})

const currentSubNavItems = computed(() => {
  const items: Array<{ path: string; label: string; icon?: any }> = []
  currentSubNav.value.forEach(group => {
    items.push(...group.items)
  })
  return items
})

// Search results (mock)
const searchResults = computed(() => [
  {
    category: 'memories',
    label: t('nav.memories'),
    items: [
      { id: 1, title: 'Go 后端迁移方案', description: '技术笔记', icon: Collection },
      { id: 2, title: 'UI 设计规范', description: '设计文档', icon: Collection },
    ]
  },
  {
    category: 'knowledge',
    label: t('nav.knowledge'),
    items: [
      { id: 3, title: 'ClawMemory 项目', description: '知识实体', icon: Connection },
    ]
  }
])

function navigateTo(item: any) {
  showSearch.value = false
  searchQuery.value = ''
  // router.push(item.path)
}

function handleUserCommand(command: string) {
  if (command === 'logout') {
    localStorage.removeItem('token')
    router.push('/login')
  } else if (command === 'settings') {
    router.push('/settings')
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

.tier-badge-v2.tier-free {
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-tertiary);
}

.tier-badge-v2.tier-pro {
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
  width: 240px;
  flex-shrink: 0;
  background: var(--cm-bg-primary);
  border-right: 1px solid var(--cm-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width var(--cm-transition-normal);
}

.sidebar-v2.collapsed {
  width: 64px;
}

.sidebar-inner {
  flex: 1;
  overflow-y: auto;
  padding: var(--cm-space-3);
}

.sidebar-section {
  margin-bottom: var(--cm-space-4);
}

.sidebar-group-title {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--cm-text-tertiary);
  padding: var(--cm-space-2) var(--cm-space-3);
  margin-bottom: var(--cm-space-1);
}

.sidebar-item-v2 {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  padding: var(--cm-space-2) var(--cm-space-3);
  border-radius: var(--cm-radius-md);
  font-size: 14px;
  color: var(--cm-text-secondary);
  text-decoration: none;
  transition: all var(--cm-transition-fast);
  margin-bottom: 2px;
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
  padding: var(--cm-space-4);
  border-top: 1px solid var(--cm-border);
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
