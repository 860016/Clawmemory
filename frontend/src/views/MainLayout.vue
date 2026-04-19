<template>
  <div class="layout" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <!-- Top Navigation Bar -->
    <header class="topbar">
      <div class="topbar-left">
        <button class="hamburger-btn" @click="sidebarCollapsed = !sidebarCollapsed">
          <el-icon :size="20"><Fold v-if="!sidebarCollapsed" /><Expand v-else /></el-icon>
        </button>
        <div class="logo" @click="$router.push('/')">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="var(--cm-primary)" stroke-width="2">
            <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z"/>
            <path d="M12 6v6l4 2"/>
          </svg>
          <span class="logo-text">ClawMemory</span>
        </div>
      </div>

      <nav class="topbar-nav">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: isNavActive(item.path) }"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span class="nav-emoji">{{ item.emoji }}</span>
          <span class="nav-label">{{ $t(item.label) }}</span>
        </router-link>
      </nav>

      <div class="topbar-right">
        <button class="theme-toggle" @click="themeStore.toggle()" :title="themeStore.isDark ? 'Switch to light mode' : 'Switch to dark mode'">
          <span class="theme-icon">{{ themeStore.isDark ? '☀️' : '🌙' }}</span>
        </button>
        <div class="tier-badge" :class="tierClass">{{ tierLabel }}</div>
        <el-dropdown trigger="click" @command="handleUserCommand">
          <button class="user-btn">
            <el-icon :size="18"><User /></el-icon>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="settings">{{ $t('nav.settings') }}</el-dropdown-item>
              <el-dropdown-item command="logout" divided>{{ $t('common.logout') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <div class="layout-body">
      <!-- Left Sidebar -->
      <aside class="sidebar" v-if="currentSubNav.length">
        <div class="sidebar-inner">
          <div class="sidebar-section" v-for="group in currentSubNav" :key="group.label">
            <div class="sidebar-group-title" v-if="group.label">{{ $t(group.label) }}</div>
            <router-link
              v-for="item in group.items"
              :key="item.path"
              :to="item.path"
              class="sidebar-item"
              :class="{ active: $route.path === item.path || (item.path !== '/' && $route.path.startsWith(item.path)) }"
            >
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ $t(item.label) }}</span>
            </router-link>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '../stores/theme'
import {
  HomeFilled, Collection, Connection, Setting, Document, Promotion,
  Fold, Expand, User, DataAnalysis, Upload, Grid, Share,
  TrendCharts, Warning, Cpu, MagicStick, Connection as ConnectionIcon,
  FolderOpened, Lock, Coin, Monitor
} from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const themeStore = useThemeStore()
const sidebarCollapsed = ref(false)

const tier = ref('oss')
const tierLabel = computed(() => tier.value === 'oss' ? t('tier.free') : t('tier.pro'))
const tierClass = computed(() => tier.value === 'oss' ? 'tier-free' : 'tier-pro')

onMounted(async () => {
  try {
    const { data } = await axios.get('/license/info')
    tier.value = data.tier || 'oss'
  } catch {
    tier.value = 'oss'
  }
})

const navItems = [
  { path: '/', label: 'nav.dashboard', icon: HomeFilled, emoji: '📊' },
  { path: '/memories', label: 'nav.memories', icon: Collection, emoji: '🧠' },
  { path: '/knowledge', label: 'nav.knowledge', icon: Connection, emoji: '🕸️' },
  { path: '/skills', label: 'nav.skills', icon: MagicStick, emoji: '✨' },
  { path: '/wiki', label: 'nav.wiki', icon: Document, emoji: '📖' },
  { path: '/pro', label: 'nav.pro', icon: Promotion, emoji: '🚀' },
  { path: '/settings', label: 'nav.settings', icon: Setting, emoji: '⚙️' },
]

function isNavActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

// Sub-navigation per main route
const subNavMap: Record<string, Array<{ label?: string; items: Array<{ path: string; label: string; icon?: any }> }>> = {
  '/': [
    { items: [
      { path: '/', label: 'nav.overview', icon: HomeFilled },
      { path: '/?tab=stats', label: 'nav.stats', icon: DataAnalysis },
    ]}
  ],
  '/memories': [
    { items: [
      { path: '/memories', label: 'memories.all', icon: Collection },
      { path: '/memories?import=openclaw', label: 'memories.importOpenClaw', icon: Upload },
      { path: '/memories?tab=skills', label: 'nav.skills', icon: MagicStick },
    ]}
  ],
  '/knowledge': [
    { items: [
      { path: '/knowledge?tab=entities', label: 'knowledge.entities', icon: Grid },
      { path: '/knowledge?tab=relations', label: 'knowledge.relations', icon: Share },
      { path: '/knowledge?tab=graph', label: 'knowledge.graphView', icon: Connection },
    ]}
  ],
  '/skills': [
    { items: [
      { path: '/skills', label: 'skills.scanSkills', icon: MagicStick },
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
      { path: '/pro?section=graph', label: 'pro.autoGraph', icon: ConnectionIcon },
      { path: '/pro?section=backup', label: 'pro.autoBackup', icon: Coin },
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
.layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--cm-bg);
  color: var(--cm-text);
}

/* ===== Top Bar ===== */
.topbar {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  background: var(--cm-bg-secondary);
  border-bottom: 1px solid var(--cm-border);
  position: sticky;
  top: 0;
  z-index: 100;
  gap: 16px;
  backdrop-filter: blur(12px);
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.hamburger-btn {
  display: none;
  background: none;
  border: none;
  color: var(--cm-text-muted);
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: all 0.2s;
}
.hamburger-btn:hover {
  background: rgba(var(--cm-primary-rgb), 0.1);
  color: var(--cm-primary);
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.2s;
}
.logo:hover {
  background: rgba(var(--cm-primary-rgb), 0.06);
}

.logo-text {
  font-size: 17px;
  font-weight: 700;
  color: var(--cm-text);
  letter-spacing: 0.3px;
}

.topbar-nav {
  display: flex;
  align-items: center;
  gap: 2px;
  flex: 1;
  justify-content: center;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--cm-text-muted);
  text-decoration: none;
  transition: all 0.2s ease;
  white-space: nowrap;
  position: relative;
}
.nav-item:hover {
  color: var(--cm-text);
  background: rgba(var(--cm-primary-rgb), 0.06);
}
.nav-item.active {
  color: var(--cm-primary);
  background: rgba(var(--cm-primary-rgb), 0.1);
  font-weight: 600;
}
.nav-item.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 2px;
  background: var(--cm-primary);
  border-radius: 1px;
}
.nav-emoji { font-size: 14px; line-height: 1; }
.nav-label { font-size: 13px; }

.topbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.theme-toggle {
  background: none;
  border: 1px solid var(--cm-border);
  cursor: pointer;
  padding: 5px 8px;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
}
.theme-toggle:hover {
  border-color: var(--cm-primary);
  background: rgba(var(--cm-primary-rgb), 0.06);
}
.theme-icon {
  font-size: 16px;
  line-height: 1;
}

.tier-badge {
  padding: 3px 10px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
}
.tier-free {
  background: rgba(var(--cm-primary-rgb), 0.08);
  color: var(--cm-text-muted);
}
.tier-pro {
  background: rgba(var(--cm-primary-rgb), 0.15);
  color: var(--cm-primary);
}

.user-btn {
  background: none;
  border: 1px solid var(--cm-border);
  cursor: pointer;
  padding: 5px 8px;
  border-radius: 8px;
  color: var(--cm-text-muted);
  transition: all 0.2s;
  display: flex;
  align-items: center;
}
.user-btn:hover {
  border-color: var(--cm-primary);
  color: var(--cm-primary);
  background: rgba(var(--cm-primary-rgb), 0.06);
}

/* ===== Layout Body ===== */
.layout-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ===== Sidebar ===== */
.sidebar {
  width: 200px;
  flex-shrink: 0;
  background: var(--cm-bg-secondary);
  border-right: 1px solid var(--cm-border);
  overflow-y: auto;
  transition: width 0.3s ease;
}

.sidebar-inner {
  padding: 12px 8px;
}

.sidebar-section {
  margin-bottom: 8px;
}

.sidebar-group-title {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--cm-text-muted);
  padding: 8px 12px 4px;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 13px;
  color: var(--cm-text-muted);
  text-decoration: none;
  transition: all 0.2s ease;
}
.sidebar-item:hover {
  color: var(--cm-text);
  background: rgba(var(--cm-primary-rgb), 0.06);
}
.sidebar-item.active {
  color: var(--cm-primary);
  background: rgba(var(--cm-primary-rgb), 0.1);
  font-weight: 600;
}

/* ===== Main Content ===== */
.main-content {
  flex: 1;
  overflow-y: auto;
  background: var(--cm-bg);
}

/* ===== Mobile Responsive ===== */
@media (max-width: 768px) {
  .hamburger-btn {
    display: flex;
  }

  .topbar-nav {
    display: none;
  }

  .sidebar {
    position: fixed;
    left: 0;
    top: 56px;
    bottom: 0;
    z-index: 90;
    width: 220px;
    transform: translateX(-100%);
    transition: transform 0.3s ease;
    box-shadow: var(--cm-shadow-lg);
  }

  .layout:not(.sidebar-collapsed) .sidebar {
    transform: translateX(0);
  }

  /* 小屏幕显示图标+emoji导航 */
  .topbar-nav {
    display: flex;
  }
  .nav-label {
    display: none;
  }
  .topbar-nav .nav-item {
    padding: 8px;
  }
  .nav-item.active::after { display: none; }
}

@media (max-width: 1024px) {
  .nav-label {
    display: none;
  }
  .nav-item {
    padding: 8px 10px;
  }
}
</style>
