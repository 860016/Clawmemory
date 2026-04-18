<template>
  <el-container style="height: 100vh">
<<<<<<< HEAD
    <el-aside width="220px" class="app-sidebar">
      <div class="sidebar-logo">
        <div class="logo-icon">
          <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" width="28" height="28">
            <rect width="32" height="32" rx="8" fill="url(#logo-grad)"/>
            <path d="M8 22V10l8 6-8 6z" fill="white"/>
            <path d="M16 22V10l8 6-8 6z" fill="white" opacity="0.6"/>
            <defs><linearGradient id="logo-grad" x1="0" y1="0" x2="32" y2="32"><stop stop-color="#4f6ef7"/><stop offset="1" stop-color="#7b93fa"/></linearGradient></defs>
=======
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <div class="logo-icon-wrap">
          <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="#00d4aa" stroke-width="2">
            <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z"/>
            <path d="M12 6v6l4 2"/>
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
          </svg>
        </div>
        <span class="logo-text">ClawMemory</span>
      </div>
<<<<<<< HEAD

      <el-menu
        :default-active="currentRoute"
        router
        class="sidebar-menu"
        background-color="transparent"
        text-color="var(--cm-sidebar-text)"
        active-text-color="var(--cm-sidebar-active)"
      >
        <el-menu-item index="/">
          <el-icon><HomeFilled /></el-icon>
          <span>主页</span>
        </el-menu-item>
        <el-menu-item index="/chat">
          <el-icon><ChatDotRound /></el-icon>
          <span>聊天</span>
        </el-menu-item>
        <el-menu-item index="/memories">
          <el-icon><Collection /></el-icon>
          <span>记忆</span>
        </el-menu-item>
        <el-menu-item index="/knowledge">
          <el-icon><Connection /></el-icon>
          <span>知识库</span>
        </el-menu-item>
        <el-menu-item index="/agents">
          <el-icon><Monitor /></el-icon>
          <span>Agent</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/admin">
          <el-icon><User /></el-icon>
          <span>管理</span>
        </el-menu-item>
      </el-menu>

      <div class="sidebar-footer">
        <div class="user-info">
          <el-icon><UserFilled /></el-icon>
          <span class="user-name">{{ auth.username }}</span>
          <el-tag size="small" :type="isAdmin ? 'warning' : 'info'" effect="dark" round>{{ auth.role }}</el-tag>
        </div>
        <el-button text class="logout-btn" @click="handleLogout">
          <el-icon><SwitchButton /></el-icon> 退出
        </el-button>
      </div>
    </el-aside>

    <el-main style="padding: 0; background: var(--cm-bg);">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
=======
      <el-menu :default-active="currentRoute" router class="sidebar-menu">
        <el-menu-item index="/">
          <el-icon><HomeFilled /></el-icon>
          <span>{{ $t('nav.dashboard') }}</span>
        </el-menu-item>
        <el-menu-item index="/memories">
          <el-icon><Collection /></el-icon>
          <span>{{ $t('nav.memories') }}</span>
        </el-menu-item>
        <el-menu-item index="/knowledge">
          <el-icon><Connection /></el-icon>
          <span>{{ $t('nav.knowledge') }}</span>
        </el-menu-item>
        <el-menu-item index="/wiki">
          <el-icon><Document /></el-icon>
          <span>{{ $t('nav.wiki') }}</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>{{ $t('nav.settings') }}</span>
        </el-menu-item>
      </el-menu>
      <div class="sidebar-footer">
        <div class="tier-badge" :class="tierClass">{{ tierLabel }}</div>
        <el-button text @click="handleLogout" class="logout-btn">
          <el-icon><SwitchButton /></el-icon> {{ $t('common.logout') }}
        </el-button>
      </div>
    </el-aside>
    <el-main class="main-content">
      <router-view />
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
<<<<<<< HEAD
import { useAuthStore } from '../stores/auth'
import {
  HomeFilled, ChatDotRound, Collection, Connection, Monitor,
  Setting, User, SwitchButton, UserFilled
} from '@element-plus/icons-vue'
=======
import { useI18n } from 'vue-i18n'
import { HomeFilled, Collection, Connection, Setting, SwitchButton, Document } from '@element-plus/icons-vue'
import axios from '../api/client'
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const currentRoute = computed(() => route.path)

const tier = ref('oss')
const tierLabel = computed(() => tier.value === 'oss' ? t('tier.free') : t('tier.pro'))
const tierClass = computed(() => tier.value === 'oss' ? 'tier-free' : 'tier-pro')

onMounted(async () => {
  try {
    const { data } = await axios.get('/api/v1/license/info')
    tier.value = data.tier || 'oss'
  } catch {
    tier.value = 'oss'
  }
})

function handleLogout() {
  localStorage.removeItem('token')
  router.push('/login')
}
</script>

<style scoped>
<<<<<<< HEAD
.app-sidebar {
  background: var(--cm-sidebar-bg);
  display: flex;
  flex-direction: column;
  border-right: none;
  overflow: hidden;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 18px 16px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}
.logo-icon { flex-shrink: 0; }
.logo-text {
  font-size: 17px;
  font-weight: 700;
  color: #fff;
  letter-spacing: -0.3px;
}

.sidebar-menu {
  flex: 1;
  border-right: none !important;
  padding: 8px;
}
.sidebar-menu .el-menu-item {
  border-radius: 8px;
  margin-bottom: 2px;
  height: 44px;
  line-height: 44px;
  transition: all 0.2s;
}
.sidebar-menu .el-menu-item:hover {
  background: rgba(255,255,255,0.06) !important;
}
.sidebar-menu .el-menu-item.is-active {
  background: var(--cm-primary) !important;
  color: #fff !important;
}

.sidebar-footer {
  padding: 12px 14px;
  border-top: 1px solid rgba(255,255,255,0.06);
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  color: var(--cm-sidebar-text);
  font-size: 13px;
}
.user-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.logout-btn {
  width: 100%;
  color: var(--cm-sidebar-text) !important;
  justify-content: flex-start;
}
.logout-btn:hover {
  color: #f56c6c !important;
=======
.sidebar {
  background: #0d1117;
  border-right: 1px solid rgba(0, 212, 170, 0.1);
  display: flex;
  flex-direction: column;
}

.logo {
  padding: 20px 16px;
  display: flex;
  align-items: center;
  gap: 10px;
  border-bottom: 1px solid rgba(0, 212, 170, 0.1);
}

.logo-icon-wrap {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 212, 170, 0.1);
  border-radius: 10px;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #e6edf3;
  letter-spacing: 0.5px;
}

.sidebar-menu {
  border-right: none;
  background: transparent;
  --el-menu-item-height: 46px;
}

.sidebar-menu .el-menu-item {
  color: #7d8590;
  margin: 3px 8px;
  border-radius: 8px;
  transition: all 0.2s;
}

.sidebar-menu .el-menu-item:hover {
  background: rgba(0, 212, 170, 0.08);
  color: #e6edf3;
}

.sidebar-menu .el-menu-item.is-active {
  background: rgba(0, 212, 170, 0.15);
  color: #00d4aa;
}

.sidebar-footer {
  margin-top: auto;
  padding: 16px;
  border-top: 1px solid rgba(0, 212, 170, 0.1);
  text-align: center;
}

.tier-badge {
  display: inline-block;
  padding: 4px 14px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
}

.tier-free {
  background: rgba(125, 133, 144, 0.15);
  color: #7d8590;
}

.tier-pro {
  background: rgba(0, 212, 170, 0.15);
  color: #00d4aa;
}

.logout-btn {
  color: #7d8590;
  width: 100%;
}

.logout-btn:hover {
  color: #f85149;
}

.main-content {
  padding: 0;
  background: #0d1117;
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
}
</style>
