<template>
  <div class="settings-view" style="padding: 16px; max-width: 700px">
    <el-tabs v-model="activeTab">
      <!-- 个人资料 -->
      <el-tab-pane label="个人资料" name="profile">
        <el-form :model="profileForm" label-width="100px" style="margin-top: 16px">
          <el-form-item label="用户名"><el-input :model-value="auth.username" disabled /></el-form-item>
          <el-form-item label="角色"><el-input :model-value="auth.role" disabled /></el-form-item>
          <el-form-item label="邮箱"><el-input v-model="profileForm.email" placeholder="输入邮箱" /></el-form-item>
          <el-form-item label="显示名称"><el-input v-model="profileForm.display_name" placeholder="输入显示名称" /></el-form-item>
          <el-form-item><el-button type="primary" @click="handleUpdateProfile">保存</el-button></el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 修改密码 -->
      <el-tab-pane label="修改密码" name="password">
        <el-form :model="passwordForm" label-width="100px" style="margin-top: 16px">
          <el-form-item label="当前密码">
            <el-input v-model="passwordForm.old_password" type="password" show-password />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="passwordForm.new_password" type="password" show-password />
          </el-form-item>
          <el-form-item label="确认密码">
            <el-input v-model="passwordForm.confirm_password" type="password" show-password />
          </el-form-item>
          <el-form-item><el-button type="warning" @click="handleChangePassword">修改密码</el-button></el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 授权管理 -->
      <el-tab-pane label="授权管理" name="license">
        <el-descriptions :column="1" border style="margin-top: 16px">
          <el-descriptions-item label="授权状态">
            <el-tag :type="licenseInfo?.active ? 'success' : 'info'">
              {{ licenseInfo?.active ? '已激活' : '未激活' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="授权类型">
            <el-tag>{{ licenseInfo?.tier || 'OSS (开源版)' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="到期时间">{{ licenseInfo?.expires_at || '永久' }}</el-descriptions-item>
          <el-descriptions-item label="最大设备数">{{ licenseInfo?.max_devices ?? 1 }}</el-descriptions-item>
        </el-descriptions>

        <el-divider />

        <div v-if="!licenseInfo?.active || licenseInfo?.tier === 'OSS'" style="margin-top: 16px">
          <h4>升级到 Pro 版</h4>
          <p style="color: #666; margin: 8px 0">解锁知识图谱、自动备份、记忆衰减等高级功能</p>
          <el-form :inline="true">
            <el-form-item>
              <el-input v-model="licenseKey" placeholder="输入授权码" style="width: 300px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleActivateLicense">激活</el-button>
            </el-form-item>
          </el-form>
        </div>

        <div v-else style="margin-top: 16px">
          <el-button type="danger" plain @click="handleDeactivateLicense">停用授权</el-button>
        </div>

        <el-divider />

        <h4>版本功能对比</h4>
        <el-table :data="featureMatrix" stripe style="width: 100%; margin-top: 12px" size="small">
          <el-table-column prop="feature" label="功能" width="180" />
          <el-table-column label="OSS">
            <template #default="{ row }"><el-tag :type="row.oss ? 'success' : 'info'" size="small">{{ row.oss ? '✓' : '✗' }}</el-tag></template>
          </el-table-column>
          <el-table-column label="Pro">
            <template #default="{ row }"><el-tag :type="row.pro ? 'success' : 'info'" size="small">{{ row.pro ? '✓' : '✗' }}</el-tag></template>
          </el-table-column>
          <el-table-column label="Enterprise">
            <template #default="{ row }"><el-tag :type="row.ent ? 'success' : 'info'" size="small">{{ row.ent ? '✓' : '✗' }}</el-tag></template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { adminApi } from '../api/admin'
import api from '../api/client'
import { ElMessage, ElMessageBox } from 'element-plus'

const auth = useAuthStore()
const activeTab = ref('profile')
const licenseKey = ref('')
const licenseInfo = ref<any>(null)

const profileForm = reactive({ email: '', display_name: '' })
const passwordForm = reactive({ old_password: '', new_password: '', confirm_password: '' })

const featureMatrix = [
  { feature: '记忆 CRUD + FTS5', oss: true, pro: true, ent: true },
  { feature: '语义搜索', oss: true, pro: true, ent: true },
  { feature: '知识图谱可视化', oss: false, pro: true, ent: true },
  { feature: '自动备份', oss: false, pro: true, ent: true },
  { feature: '记忆衰减', oss: false, pro: true, ent: true },
  { feature: 'Token 智能路由', oss: false, pro: true, ent: true },
  { feature: '矛盾治理', oss: false, pro: true, ent: true },
  { feature: 'API 访问', oss: false, pro: false, ent: true },
  { feature: 'SSO', oss: false, pro: false, ent: true },
  { feature: '审计日志', oss: false, pro: false, ent: true },
  { feature: '记忆时间旅行', oss: false, pro: false, ent: true },
  { feature: '离线模式', oss: false, pro: false, ent: true },
]

onMounted(async () => {
  await auth.fetchMe()
  profileForm.email = ''
  profileForm.display_name = ''
  try {
    const resp = await adminApi.getLicenseInfo()
    licenseInfo.value = resp.data
  } catch { licenseInfo.value = null }
})

async function handleUpdateProfile() {
  try {
    await auth.updateMe(profileForm)
    ElMessage.success('保存成功')
  } catch { /* handled by interceptor */ }
}

async function handleChangePassword() {
  if (!passwordForm.old_password || !passwordForm.new_password) {
    return ElMessage.warning('请填写当前密码和新密码')
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    return ElMessage.warning('两次密码输入不一致')
  }
  if (passwordForm.new_password.length < 6) {
    return ElMessage.warning('密码长度至少6位')
  }
  try {
    await api.put('/auth/change-password', {
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    ElMessage.success('密码修改成功，请重新登录')
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    auth.logout()
    window.location.href = '/login'
  } catch { /* handled by interceptor */ }
}

async function handleActivateLicense() {
  if (!licenseKey.value) return ElMessage.warning('请输入授权码')
  try {
    const resp = await adminApi.activateLicense(licenseKey.value)
    licenseInfo.value = resp.data
    ElMessage.success('激活成功')
    licenseKey.value = ''
  } catch { /* handled by interceptor */ }
}

async function handleDeactivateLicense() {
  await ElMessageBox.confirm('停用授权将失去高级功能，确定继续？', '警告', { type: 'warning' })
  try {
    await adminApi.deactivateLicense()
    licenseInfo.value = { active: false, tier: 'OSS' }
    ElMessage.success('已停用')
  } catch { /* handled by interceptor */ }
}
</script>
