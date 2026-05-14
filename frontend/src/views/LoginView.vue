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

      <!-- Login form -->
      <div v-if="loginMode === 'login'" class="login-form">
        <el-input
          v-model="username"
          :placeholder="$t('login.usernamePlaceholder')"
          @keyup.enter="focusPassword"
          size="large"
        />
        <el-input
          ref="passwordInput"
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
        <div class="login-links">
          <button class="link-btn" @click="showForgotDialog = true">{{ $t('login.forgotPassword') }}</button>
          <button class="link-btn" @click="loginMode = 'register'">{{ $t('login.goRegister') }}</button>
        </div>
      </div>

      <!-- Register form -->
      <div v-else-if="loginMode === 'register'" class="login-form">
        <el-input
          v-model="regUsername"
          :placeholder="$t('login.usernamePlaceholder')"
          size="large"
        />
        <el-input
          v-model="regPassword"
          type="password"
          :placeholder="$t('login.regPasswordPlaceholder')"
          size="large"
          show-password
        />
        <el-input
          v-model="regConfirmPassword"
          type="password"
          :placeholder="$t('login.confirmPasswordPlaceholder')"
          @keyup.enter="handleRegister"
          size="large"
          show-password
        />
        <el-input
          v-model="invitationCode"
          :placeholder="$t('login.invitationCodePlaceholder')"
          size="large"
        />
        <el-button type="primary" @click="handleRegister" :loading="loading" size="large" class="login-btn">
          {{ $t('login.register') }}
        </el-button>
        <div class="login-links">
          <button class="link-btn" @click="loginMode = 'login'">{{ $t('login.backToLogin') }}</button>
        </div>
      </div>

      <!-- First-time setup -->
      <div v-else class="login-form">
        <p class="hint">{{ $t('login.noPassword') }}</p>
        <el-input
          v-model="setupUsername"
          :placeholder="$t('login.setupUsernamePlaceholder')"
          size="large"
        />
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

      <div v-if="resetMessage" class="reset-message">
        <el-icon color="#10B981"><SuccessFilled /></el-icon>
        <span>{{ resetMessage }}</span>
      </div>
    </div>

    <!-- Forgot Password Dialog -->
    <el-dialog v-model="showForgotDialog" :title="$t('login.forgotPassword')" width="440px" :close-on-click-modal="false">
      <div v-if="!resetMessage">
        <el-alert type="info" :closable="false" style="margin-bottom: 16px">
          <template #title>{{ $t('login.forgotStep2Title') }}</template>
          <p style="margin: 8px 0 0; font-size: 13px; color: var(--cm-text-muted)">{{ $t('login.forgotStep2Desc') }}</p>
        </el-alert>
        <div class="reset-token-section">
          <el-input v-model="forgotUsername" :placeholder="$t('login.usernamePlaceholder')" size="large" style="margin-bottom: 12px" />
          <el-input v-model="newPassword" type="password" show-password :placeholder="$t('login.newPasswordPlaceholder')" size="large" />
          <el-button type="primary" @click="handleResetPassword" :loading="loading" size="large" style="margin-top: 12px; width: 100%">
            {{ $t('login.resetPassword') }}
          </el-button>
        </div>
        <div class="cli-hint">
          <p>{{ $t('login.cliResetHint') }}</p>
          <div class="cli-commands">
            <div class="cli-platform">Windows:</div>
            <code>clawmemory.exe --reset-password NEW_PASSWORD</code>
            <div class="cli-platform">Linux/macOS:</div>
            <code>./clawmemory --reset-password NEW_PASSWORD</code>
          </div>
        </div>
      </div>
      <div v-else class="reset-success-box">
        <el-icon color="#10B981" :size="40"><SuccessFilled /></el-icon>
        <p style="margin: 12px 0 0; color: var(--cm-text); font-weight: 600">{{ resetMessage }}</p>
      </div>
      <template #footer>
        <el-button @click="showForgotDialog = false; resetMessage = ''; newPassword = ''">{{ $t('common.cancel') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { SuccessFilled } from '@element-plus/icons-vue'
import axios from '../api/go-client'
import { translateError } from '../i18n'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const loginMode = ref<'login' | 'register' | 'setup'>('login')
const username = ref('')
const password = ref('')
const setupUsername = ref('')
const loading = ref(false)
const passwordSet = ref(true)
const resetMessage = ref('')
const showForgotDialog = ref(false)
const newPassword = ref('')
const forgotUsername = ref('')

const regUsername = ref('')
const regPassword = ref('')
const regConfirmPassword = ref('')
const invitationCode = ref('')

const passwordInput = ref<any>(null)

function focusPassword() {
  passwordInput.value?.focus()
}

onMounted(async () => {
  if (route.query.mode === 'register') {
    loginMode.value = 'register'
    if (route.query.code) {
      invitationCode.value = route.query.code as string
    }
  }

  try {
    const { data } = await axios.get('/auth/init-status')
    passwordSet.value = data.password_set
    if (!data.password_set) {
      loginMode.value = 'setup'
    }
  } catch (e) {
    console.error('Failed to check init status:', e)
    passwordSet.value = false
    loginMode.value = 'setup'
  }
})

async function handleLogin() {
  if (!password.value) return
  loading.value = true
  try {
    const { data } = await axios.post('/auth/login', {
      username: username.value || 'admin',
      password: password.value,
    })
    localStorage.setItem('token', data.access_token)
    router.push('/')
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error || e.response?.data?.detail, t('login.wrongPassword')))
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  if (!regUsername.value || !regPassword.value) {
    ElMessage.warning(t('login.fillRequired'))
    return
  }
  if (regPassword.value.length < 6) {
    ElMessage.warning(t('login.passwordMinLen6'))
    return
  }
  if (regPassword.value !== regConfirmPassword.value) {
    ElMessage.warning(t('login.passwordMismatch'))
    return
  }
  loading.value = true
  try {
    const payload: any = {
      username: regUsername.value,
      password: regPassword.value,
    }
    if (invitationCode.value) {
      payload.invitation_code = invitationCode.value
    }
    await axios.post('/auth/register-with-invitation', payload)
    ElMessage.success(t('login.registerSuccess'))
    username.value = regUsername.value
    password.value = regPassword.value
    loginMode.value = 'login'
    await handleLogin()
  } catch (e: any) {
    const errMsg = e.response?.data?.error || ''
    if (errMsg.includes('invitation') || errMsg.includes('invite')) {
      ElMessage.error(t('login.invalidInvitationCode'))
    } else if (errMsg.includes('already exists') || errMsg.includes('conflict')) {
      ElMessage.error(t('login.usernameExists'))
    } else {
      ElMessage.error(translateError(errMsg, t('common.failed')))
    }
  } finally {
    loading.value = false
  }
}

async function handleSetPassword() {
  if (setupUsername.value && setupUsername.value.length < 2) {
    ElMessage.warning(t('login.usernameTooShort'))
    return
  }
  if (password.value.length < 6) {
    ElMessage.warning(t('login.passwordTooShort'))
    return
  }
  loading.value = true
  try {
    const payload: any = { password: password.value }
    if (setupUsername.value) {
      payload.username = setupUsername.value
    }
    const { data } = await axios.post('/auth/set-password', payload)
    if (data.access_token) {
      localStorage.setItem('token', data.access_token)
      passwordSet.value = true
      router.push('/')
    } else {
      ElMessage.error(t('common.failed'))
    }
  } catch (e: any) {
    const detail = e.response?.data?.error || e.response?.data?.detail || ''
    if (detail === 'password already set') {
      try {
        const loginPayload: any = { password: password.value }
        if (setupUsername.value) {
          loginPayload.username = setupUsername.value
        }
        const { data: loginData } = await axios.post('/auth/login', loginPayload)
        localStorage.setItem('token', loginData.access_token)
        passwordSet.value = true
        router.push('/')
        return
      } catch (loginErr: any) {
        ElMessage.error(translateError(loginErr.response?.data?.error || loginErr.response?.data?.detail, t('login.wrongPassword')))
      }
    } else {
      ElMessage.error(translateError(detail, t('common.failed')))
    }
  } finally {
    loading.value = false
  }
}

async function handleResetPassword() {
  if (!newPassword.value || newPassword.value.length < 6) {
    ElMessage.warning(t('login.passwordTooShort'))
    return
  }
  loading.value = true
  try {
    await axios.post('/auth/forgot-password', {
      username: forgotUsername.value || 'admin',
      new_password: newPassword.value,
      confirm: true,
    })
    resetMessage.value = t('login.resetSuccess')
    username.value = forgotUsername.value
    password.value = newPassword.value
  } catch (e: any) {
    const hint = e.response?.data?.hint
    const error = translateError(e.response?.data?.error, t('login.resetFailed'))
    ElMessage.error(hint ? `${error} (${hint})` : error)
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
  width: 420px;
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
  gap: 14px;
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

.login-links {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.link-btn {
  background: none;
  border: none;
  color: var(--cm-primary);
  cursor: pointer;
  font-size: 13px;
  padding: 4px;
  transition: opacity 0.2s;
}
.link-btn:hover {
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

.reset-token-section {
  margin-top: 8px;
}

.cli-hint {
  margin-top: 16px;
  padding: 12px;
  background: var(--cm-bg);
  border: 1px solid var(--cm-border);
  border-radius: 8px;
  font-size: 12px;
}
.cli-hint p {
  color: var(--cm-text-muted);
  margin: 0 0 8px;
}
.cli-commands {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.cli-platform {
  color: var(--cm-text-muted);
  font-size: 11px;
  margin-top: 6px;
}
.cli-platform:first-child {
  margin-top: 0;
}
.cli-hint code {
  display: block;
  padding: 6px 10px;
  background: rgba(var(--cm-primary-rgb), 0.06);
  border-radius: 4px;
  font-family: monospace;
  font-size: 11px;
  color: var(--cm-text);
  user-select: all;
  word-break: break-all;
}

.reset-success-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0;
}

@media (max-width: 768px) {
  .login-card {
    padding: 36px 28px;
    border-radius: 16px;
  }
  .login-header h1 {
    font-size: 24px;
  }
}

@media (max-width: 480px) {
  .login-card {
    padding: 28px 20px;
    border-radius: 14px;
  }
  .login-header h1 {
    font-size: 20px;
  }
}
</style>
