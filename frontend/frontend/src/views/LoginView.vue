<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header><h2>OpenClaw 登录</h2></template>
      <el-form :model="form" @submit.prevent="handleLogin" label-width="80px">
        <el-form-item label="账号"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <el-form-item><el-button type="primary" native-type="submit" :loading="loading">登录</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = ref({ username: '', password: '' })

onMounted(async () => {
  try {
    const init = await auth.checkInit()
    if (!init) router.push('/setup')
  } catch { /* ignore */ }
})

async function handleLogin() {
  loading.value = true
  try {
    await auth.login(form.value)
    router.push('/')
  } catch { /* Error handled by interceptor */ } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container { display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #f5f7fa; }
.login-card { width: 380px; }
</style>
