<template>
  <div class="login-container">
    <div class="glow-orb orb-1"></div>
    <div class="glow-orb orb-2"></div>
    <div class="login-card">
      <div class="login-header">
        <div class="logo-icon-wrap">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="#10B981" stroke-width="2">
            <path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z"/>
            <path d="M12 6v6l4 2"/>
          </svg>
        </div>
        <h1>ClawMemory</h1>
        <p class="subtitle">{{ $t('common.appDesc') }}</p>
      </div>

      <!-- Password already set - show login form -->
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
        <button class="forgot-link" @click="handleForgotPassword">{{ $t('login.forgotPassword') }}</button>
      </div>

      <!-- No password set - require setting one -->
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
      </div>

      <!-- Reset password success message -->
      <div v-if="resetMessage" class="reset-message">
        <el-icon color="#10B981"><SuccessFilled /></el-icon>
        <span>{{ resetMessage }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { SuccessFilled } from '@element-plus/icons-vue'
import axios from '../api/client'
import { authApi } from '../api/auth'

const { t } = useI18n()
const router = useRouter()
const password = ref('')
const loading = ref(false)
const passwordSet = ref(true)
const resetMessage = ref('')

onMounted(async () => {
  try {
    const { data } = await axios.get('/auth/init-status')
    passwordSet.value = data.password_set
  } catch {
    passwordSet.value = false
  }
})

async function handleLogin() {
  if (!password.value) return
  loading.value = true
  try {
    const { data } = await axios.post('/auth/login', { password: password.value })
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
    const { data } = await axios.post('/auth/set-password', { password: password.value })
    localStorage.setItem('token', data.access_token)
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally {
    loading.value = false
  }
}

async function handleForgotPassword() {
  loading.value = true
  try {
    await authApi.resetPassword()
    resetMessage.value = t('login.resetSuccess')
    passwordSet.value = false
    password.value = ''
  } catch {
    ElMessage.error(t('login.resetFailed'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--cm-bg);
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
  background: #10B981;
  top: -100px;
  left: -100px;
  animation: float1 8s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: #06b6d4;
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
  background: rgba(var(--cm-primary-rgb), 0.03);
  backdrop-filter: blur(24px);
  border: 1px solid rgba(16, 185, 129, 0.15);
  border-radius: 20px;
  padding: 48px 40px;
  width: 400px;
  max-width: 90vw;
  position: relative;
  z-index: 1;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

[data-theme="light"] .login-card {
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(16, 185, 129, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
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
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 16px;
}

.login-header h1 {
  color: var(--cm-text);
  font-size: 26px;
  font-weight: 700;
  margin: 0;
  letter-spacing: 1px;
}

.subtitle {
  color: var(--cm-text-muted);
  font-size: 14px;
  margin-top: 8px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.hint {
  color: var(--cm-text-muted);
  font-size: 13px;
  text-align: center;
  margin: 0;
}

.login-btn {
  width: 100%;
  height: 44px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
}

.forgot-link {
  background: none;
  border: none;
  color: var(--cm-primary);
  cursor: pointer;
  font-size: 13px;
  text-align: center;
  padding: 4px;
  transition: opacity 0.2s;
}
.forgot-link:hover {
  opacity: 0.8;
}

.reset-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  margin-top: 8px;
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 10px;
  color: var(--cm-text-secondary);
  font-size: 13px;
}

@media (max-width: 480px) {
  .login-card {
    padding: 32px 24px;
    border-radius: 16px;
  }
}
</style>
