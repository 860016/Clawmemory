<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <div class="logo-icon-wrap">
          <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="#00d4aa" stroke-width="2">
            <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z"/>
            <path d="M12 6v6l4 2"/>
          </svg>
        </div>
        <span class="logo-text">ClawMemory</span>
      </div>
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
        <el-menu-item index="/pro">
          <el-icon><Promotion /></el-icon>
          <span>{{ $t('nav.pro') }}</span>
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
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { HomeFilled, Collection, Connection, Setting, SwitchButton, Document, Promotion } from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const currentRoute = computed(() => route.path)

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

function handleLogout() {
  localStorage.removeItem('token')
  router.push('/login')
}
</script>

<style scoped>
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
}
</style>
