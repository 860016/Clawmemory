<template>
  <div class="setup-container">
    <el-card class="setup-card">
      <template #header><h2>OpenClaw 初始化</h2></template>
      <el-form :model="form" @submit.prevent="handleInit" label-width="100px">
        <el-form-item label="管理员账号"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <el-form-item label="邮箱(可选)"><el-input v-model="form.email" /></el-form-item>
        <el-form-item><el-button type="primary" native-type="submit" :loading="loading">初始化系统</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = ref({ username: '', password: '', email: '' })

async function handleInit() {
  loading.value = true
  try {
    await auth.initSystem(form.value)
    ElMessage.success('初始化成功')
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || '初始化失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.setup-container { display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #f5f7fa; }
.setup-card { width: 420px; }
</style>
