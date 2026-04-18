<template>
<<<<<<< HEAD
  <div class="login-page">
    <div class="login-bg"></div>
    <el-card class="login-card">
      <div class="login-header">
        <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" width="36" height="36">
          <rect width="32" height="32" rx="8" fill="url(#login-grad)"/>
          <path d="M8 22V10l8 6-8 6z" fill="white"/>
          <path d="M16 22V10l8 6-8 6z" fill="white" opacity="0.6"/>
          <defs><linearGradient id="login-grad" x1="0" y1="0" x2="32" y2="32"><stop stop-color="#4f6ef7"/><stop offset="1" stop-color="#7b93fa"/></linearGradient></defs>
        </svg>
        <h2>ClawMemory</h2>
        <p>OpenClaw 智能记忆管理系统</p>
      </div>
      <el-form :model="form" @submit.prevent="handleLogin" label-position="top">
        <el-form-item label="账号">
          <el-input v-model="form.username" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password prefix-icon="Lock" size="large" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" size="large" style="width: 100%; border-radius: 10px">
          登 录
        </el-button>
      </el-form>
    </el-card>
=======
  <div class="login-container">
    <div class="glow-orb orb-1"></div>
    <div class="glow-orb orb-2"></div>
    <div class="login-card">
      <div class="login-header">
        <div class="logo-icon-wrap">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="#00d4aa" stroke-width="2">
            <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z"/>
            <path d="M12 6v6l4 2"/>
          </svg>
        </div>
        <h1>ClawMemory</h1>
        <p class="subtitle">{{ $t('common.appDesc') }}</p>
      </div>

      <div v-if="passwordSet" class="login-form">
        <el-input
          v-model="password"
          type="password"
          :placeholder="$t('login.passwordPlaceholder')"
          @keyup.enter="handleLogin"
          size="large"
          show-password
        />
        <el-button type="primary" @click="handleLogin" :loading="loading" size="large" class="login-btn">
          {{ $t('login.login') }}
        </el-button>
      </div>

      <div v-else class="login-form">
        <p class="hint">{{ $t('login.noPassword') }}</p>
        <el-input
          v-model="password"
          type="password"
          :placeholder="$t('login.setPassword')"
          @keyup.enter="handleSetPassword"
          size="large"
          show-password
        />
        <el-button type="primary" @click="handleSetPassword" :loading="loading" size="large" class="login-btn">
          {{ $t('login.setPasswordAndEnter') }}
        </el-button>
        <el-button text @click="enterWithoutPassword" size="large" class="skip-btn">
          {{ $t('login.skipPassword') }}
        </el-button>
      </div>
    </div>
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import axios from '../api/client'

const { t } = useI18n()
const router = useRouter()
const password = ref('')
const loading = ref(false)
const passwordSet = ref(true)

onMounted(async () => {
  try {
    const { data } = await axios.get('/api/v1/auth/init-status')
    passwordSet.value = data.password_set
  } catch {
    passwordSet.value = false
  }
})

async function handleLogin() {
  if (!password.value) return
  loading.value = true
  try {
    const { data } = await axios.post('/api/v1/auth/login', { password: password.value })
    localStorage.setItem('token', data.access_token)
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('login.wrongPassword'))
  } finally {
    loading.value = false
  }
}

async function handleSetPassword() {
  if (password.value.length < 4) {
    ElMessage.warning(t('login.passwordTooShort'))
    return
  }
  loading.value = true
  try {
    const { data } = await axios.post('/api/v1/auth/set-password', { password: password.value })
    localStorage.setItem('token', data.access_token)
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally {
    loading.value = false
  }
}

function enterWithoutPassword() {
  localStorage.setItem('token', 'no-password')
  router.push('/')
}
</script>

<style scoped>
<<<<<<< HEAD
.login-page {
  display: flex; justify-content: center; align-items: center; min-height: 100vh;
  background: #f0f2f5; position: relative; overflow: hidden;
}
.login-bg {
  position: absolute; top: -50%; left: -50%; width: 200%; height: 200%;
  background: radial-gradient(ellipse at 30% 50%, rgba(79,110,247,0.08) 0%, transparent 60%),
              radial-gradient(ellipse at 70% 50%, rgba(123,147,250,0.06) 0%, transparent 60%);
}
.login-card {
  width: 400px; border-radius: 16px !important; position: relative; z-index: 1;
  box-shadow: 0 8px 32px rgba(0,0,0,0.08) !important;
}
.login-header {
  text-align: center; margin-bottom: 24px;
}
.login-header h2 {
  margin: 12px 0 4px; font-size: 22px; color: #303133; font-weight: 700;
}
.login-header p {
  color: #909399; font-size: 13px; margin: 0;
=======
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0d1117;
  position: relative;
  overflow: hidden;
}

.glow-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.15;
}

.orb-1 {
  width: 500px;
  height: 500px;
  background: #00d4aa;
  top: -100px;
  left: -100px;
  animation: float1 8s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: #00bcd4;
  bottom: -100px;
  right: -100px;
  animation: float2 10s ease-in-out infinite;
}

@keyframes float1 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(60px, 40px); }
}

@keyframes float2 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-40px, -60px); }
}

.login-card {
  background: rgba(22, 27, 34, 0.8);
  backdrop-filter: blur(24px);
  border: 1px solid rgba(0, 212, 170, 0.15);
  border-radius: 20px;
  padding: 48px 40px;
  width: 400px;
  max-width: 90vw;
  position: relative;
  z-index: 1;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-icon-wrap {
  width: 64px;
  height: 64px;
  margin: 0 auto 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 212, 170, 0.1);
  border: 1px solid rgba(0, 212, 170, 0.2);
  border-radius: 16px;
}

.login-header h1 {
  color: #e6edf3;
  font-size: 26px;
  font-weight: 700;
  margin: 0;
  letter-spacing: 1px;
}

.subtitle {
  color: #7d8590;
  font-size: 14px;
  margin-top: 8px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.hint {
  color: #7d8590;
  font-size: 13px;
  text-align: center;
}

.login-btn {
  width: 100%;
  height: 44px;
  border-radius: 10px;
  font-size: 15px;
}

.skip-btn {
  color: #7d8590;
  width: 100%;
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
}
</style>
