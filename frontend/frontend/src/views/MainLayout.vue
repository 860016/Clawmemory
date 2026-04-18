<template>
  <el-container style="height: 100vh">
    <el-aside width="200px" style="border-right: 1px solid #e6e6e6">
      <div style="padding: 16px; font-size: 18px; font-weight: bold; text-align: center">OpenClaw</div>
      <el-menu :default-active="currentRoute" router>
        <el-menu-item index="/"><el-icon><HomeFilled /></el-icon>主页</el-menu-item>
        <el-menu-item index="/chat"><el-icon><ChatDotRound /></el-icon>聊天</el-menu-item>
        <el-menu-item index="/memories"><el-icon><Collection /></el-icon>记忆</el-menu-item>
        <el-menu-item index="/knowledge"><el-icon><Connection /></el-icon>知识库</el-menu-item>
        <el-menu-item index="/agents"><el-icon><Monitor /></el-icon>Agent</el-menu-item>
        <el-menu-item index="/settings"><el-icon><Setting /></el-icon>设置</el-menu-item>
        <el-menu-item v-if="isAdmin" index="/admin"><el-icon><User /></el-icon>管理</el-menu-item>
      </el-menu>
      <div style="position: absolute; bottom: 20px; padding: 0 16px; width: 200px; box-sizing: border-box">
        <el-button type="danger" plain @click="handleLogout" style="width: 100%">退出登录</el-button>
      </div>
    </el-aside>
    <el-main style="padding: 0"><router-view /></el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { HomeFilled, ChatDotRound, Collection, Connection, Monitor, Setting, User } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const currentRoute = computed(() => route.path)
const isAdmin = computed(() => auth.role === 'admin')

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>
