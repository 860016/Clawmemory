<template>
<<<<<<< HEAD
  <div class="settings-view">
    <div class="page-header">
      <h2>设置</h2>
      <p class="page-desc">个人资料、密码和授权管理</p>
    </div>
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
=======
  <div class="settings-page">
    <div class="page-header">
      <h1>{{ $t('settings.title') }}</h1>
    </div>
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)

    <div class="settings-grid">
      <!-- 授权管理 -->
      <div class="settings-card">
        <div class="card-title">◆ {{ $t('settings.license') }}</div>
        <div class="license-status" v-if="license.active">
          <div class="status-row">
            <span class="status-label">{{ $t('settings.version') }}</span>
            <span class="status-value pro">Pro {{ license.type === 'pro_lifetime' ? $t('settings.proLifetime') : $t('settings.proAnnual') }}</span>
          </div>
          <div class="status-row" v-if="license.expires_at">
            <span class="status-label">{{ $t('settings.expiresAt') }}</span>
            <span class="status-value">{{ license.expires_at }}</span>
          </div>
          <div class="status-row" v-if="license.device_slot">
            <span class="status-label">{{ $t('settings.deviceSlot') }}</span>
            <span class="status-value">{{ license.device_slot }}</span>
          </div>
          <div class="status-row">
            <span class="status-label">{{ $t('settings.features') }}</span>
            <div class="feature-tags">
              <span class="ftag" v-for="f in license.features" :key="f">{{ featureLabels[f] || f }}</span>
            </div>
          </div>
          <el-button type="danger" text @click="deactivateLicense" style="margin-top: 12px">{{ $t('settings.cancelLicense') }}</el-button>
        </div>
        <div v-else class="license-free">
          <div class="free-badge">{{ $t('settings.freeBadge') }}</div>
          <p class="free-desc">{{ $t('settings.freeDesc') }}</p>
          <div class="pricing">
            <div class="price-card">
              <div class="price-name">Pro {{ $t('settings.proAnnual') }}</div>
              <div class="price-amount">{{ $t('settings.proAnnualPrice') }}</div>
              <ul class="price-features">
                <li v-for="(f, i) in $tm('settings.proFeatures')" :key="i">{{ f }}</li>
              </ul>
            </div>
            <div class="price-card featured">
              <div class="price-badge">{{ $t('settings.recommended') }}</div>
              <div class="price-name">Pro {{ $t('settings.proLifetime') }}</div>
              <div class="price-amount">{{ $t('settings.proLifetimePrice') }}</div>
              <ul class="price-features">
                <li v-for="(f, i) in $tm('settings.lifetimeExtra')" :key="i">{{ f }}</li>
              </ul>
            </div>
          </div>
          <div class="activate-section">
            <el-input v-model="licenseKey" :placeholder="$t('settings.licensePlaceholder')" class="license-input" />
            <el-button type="primary" @click="activateLicense" :loading="activating">{{ $t('settings.activateLicense') }}</el-button>
          </div>
        </div>
      </div>

      <!-- 语言设置 -->
      <div class="settings-card">
        <div class="card-title">◇ {{ $t('settings.language') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.language') }}</span>
          <el-select v-model="currentLocale" @change="changeLocale" style="width: 140px">
            <el-option label="中文" value="zh" />
            <el-option label="English" value="en" />
          </el-select>
        </div>
      </div>

<<<<<<< HEAD
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

      <!-- OpenClaw 配置 -->
      <el-tab-pane label="OpenClaw 配置" name="openclaw">
        <div v-if="openclawLoading" style="margin-top: 16px; text-align: center">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
          <p style="color: #909399">加载配置中...</p>
        </div>
        <div v-else-if="!openclawConfig" style="margin-top: 16px">
          <el-alert title="未检测到 OpenClaw 配置" type="warning" show-icon :closable="false">
            <template #default>
              <p>请确认本机已安装 OpenClaw，且存在 <code>~/.openclaw/</code> 目录</p>
            </template>
          </el-alert>
        </div>
        <template v-else>
          <el-descriptions title="全局配置" :column="2" border style="margin-top: 16px">
            <el-descriptions-item label="默认模型">{{ openclawConfig.config?.default_model || '-' }}</el-descriptions-item>
            <el-descriptions-item label="主力模型">{{ openclawConfig.config?.model_routing?.primary || '-' }}</el-descriptions-item>
            <el-descriptions-item label="备用模型">{{ openclawConfig.config?.model_routing?.fallback || '-' }}</el-descriptions-item>
            <el-descriptions-item label="日志级别">{{ openclawConfig.config?.logging?.level || '-' }}</el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">模型列表</el-divider>
          <el-table :data="openclawModels" stripe size="small" style="width: 100%">
            <el-table-column prop="name" label="名称" width="180" />
            <el-table-column prop="provider" label="提供商" width="120" />
            <el-table-column prop="model" label="模型 ID" show-overflow-tooltip />
            <el-table-column prop="base_url" label="端点" show-overflow-tooltip />
          </el-table>

          <el-divider content-position="left">Agent 列表</el-divider>
          <el-table :data="openclawAgents" stripe size="small" style="width: 100%">
            <el-table-column prop="id" label="ID" width="120" />
            <el-table-column prop="name" label="名称" width="150" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
            <el-table-column prop="model" label="模型" width="180" />
          </el-table>

          <el-divider content-position="left">技能/插件</el-divider>
          <el-table :data="openclawSkills" stripe size="small" style="width: 100%">
            <el-table-column prop="id" label="ID" width="120" />
            <el-table-column prop="name" label="名称" width="180" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
          </el-table>
          <el-empty v-if="openclawSkills.length === 0" description="暂无技能配置" :image-size="40" />
        </template>
        <div style="margin-top: 16px">
          <el-button @click="loadOpenClawConfig">刷新配置</el-button>
        </div>
      </el-tab-pane>
    </el-tabs>
=======
      <!-- 安全设置 -->
      <div class="settings-card">
        <div class="card-title">◇ {{ $t('settings.security') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.password') }}</span>
          <el-button text type="primary" @click="showPasswordDialog = true">
            {{ passwordSet ? $t('settings.changePassword') : $t('settings.setPassword') }}
          </el-button>
        </div>
      </div>

      <!-- 数据管理 -->
      <div class="settings-card">
        <div class="card-title">◇ {{ $t('settings.data') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.dbLocation') }}</span>
          <span class="setting-desc">data/clawmemory.db</span>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.createBackup') }}</span>
          <el-button text type="primary" @click="createBackup" :loading="backingUp">{{ $t('settings.backupNow') }}</el-button>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.importBackup') }}</span>
          <el-upload :show-file-list="false" :before-upload="uploadBackup" accept=".zip" action="" :auto-upload="false">
            <el-button text type="primary">{{ $t('settings.chooseFile') }}</el-button>
          </el-upload>
        </div>
      </div>

      <!-- 系统信息 -->
      <div class="settings-card">
        <div class="card-title">◇ {{ $t('settings.system') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.version') }}</span>
          <span class="setting-desc">2.0.0</span>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.coreEngine') }}</span>
          <span class="setting-desc">{{ coreEngine }}</span>
        </div>
      </div>
    </div>

    <el-dialog v-model="showPasswordDialog" :title="$t('settings.setPassword')" width="400px">
      <el-form label-position="top">
        <el-form-item :label="$t('settings.newPassword')">
          <el-input v-model="newPassword" type="password" show-password :placeholder="$t('settings.passwordMinLen')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="setPassword" :loading="settingPassword">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
<<<<<<< HEAD
import { Loading } from '@element-plus/icons-vue'
=======
import axios from '../api/client'
import { setLocale, getLocale } from '../i18n'
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)

const { t } = useI18n()
const license = ref<any>({ active: false, tier: 'oss', features: [] })
const licenseKey = ref('')
const activating = ref(false)
const backingUp = ref(false)
const passwordSet = ref(false)
const showPasswordDialog = ref(false)
const newPassword = ref('')
const settingPassword = ref(false)
const coreEngine = ref('python')
const currentLocale = ref(getLocale())

<<<<<<< HEAD
// OpenClaw config
const openclawLoading = ref(false)
const openclawConfig = ref<any>(null)
const openclawModels = ref<any[]>([])
const openclawAgents = ref<any[]>([])
const openclawSkills = ref<any[]>([])

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
    localStorage.setItem('licenseInfo', JSON.stringify(resp.data))
  } catch { licenseInfo.value = null; localStorage.removeItem('licenseInfo') }
  await loadOpenClawConfig()
})

async function loadOpenClawConfig() {
  openclawLoading.value = true
  try {
    // Load settings
    const settingsResp = await api.get('/chat/gateway-settings')
    if (settingsResp.data?.found) {
      openclawConfig.value = settingsResp.data
    }

    // Load models, agents, skills in parallel
    const [modelsResp, agentsResp, skillsResp] = await Promise.allSettled([
      api.get('/models'),
      api.get('/agents'),
      api.get('/skills'),
    ])

    if (modelsResp.status === 'fulfilled') {
      const d = modelsResp.value.data
      openclawModels.value = Array.isArray(d) ? d : (d?.items ?? [])
    }
    if (agentsResp.status === 'fulfilled') {
      const d = agentsResp.value.data
      openclawAgents.value = Array.isArray(d) ? d : (d?.items ?? [])
    }
    if (skillsResp.status === 'fulfilled') {
      const d = skillsResp.value.data
      openclawSkills.value = Array.isArray(d) ? d : (d?.items ?? [])
    }
  } catch {
    openclawConfig.value = null
  } finally {
    openclawLoading.value = false
  }
}

async function handleUpdateProfile() {
  try {
    await auth.updateMe(profileForm)
    ElMessage.success('保存成功')
  } catch { /* handled by interceptor */ }
=======
const featureLabels: Record<string, string> = {
  ai_extract: 'AI提取',
  auto_graph: '自动图谱',
  unlimited_graph: '无限图谱',
  auto_decay: '自动衰减',
  decay_report: '衰减报告',
  prune_suggest: '修剪建议',
  reinforce: '访问强化',
  conflict_scan: '矛盾扫描',
  conflict_merge: 'AI合并',
  smart_router: '智能路由',
  token_stats: 'Token统计',
}

onMounted(async () => {
  await Promise.all([loadLicense(), loadInitStatus(), loadInstallStatus()])
})

function changeLocale(locale: string) {
  setLocale(locale)
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
}

async function loadLicense() {
  try { const { data } = await axios.get('/api/v1/license/info'); license.value = data } catch {}
}

async function loadInitStatus() {
  try { const { data } = await axios.get('/api/v1/auth/init-status'); passwordSet.value = data.password_set } catch {}
}

async function loadInstallStatus() {
  try { const { data } = await axios.get('/api/v1/install-status'); coreEngine.value = data.checks?.security_engine || 'python' } catch {}
}

async function activateLicense() {
  if (!licenseKey.value) return
  activating.value = true
  try {
    const { data } = await axios.post('/api/v1/license/activate', { license_key: licenseKey.value })
    if (data.valid || data.active) { ElMessage.success(t('settings.activated')); await loadLicense() }
    else ElMessage.error(data.message || t('common.failed'))
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { activating.value = false }
}

async function deactivateLicense() {
  try {
    await ElMessageBox.confirm(t('settings.cancelConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.post('/api/v1/license/deactivate')
    ElMessage.success(t('settings.canceled')); await loadLicense()
  } catch {}
}

async function setPassword() {
  if (newPassword.value.length < 4) { ElMessage.warning(t('settings.passwordMinLen')); return }
  settingPassword.value = true
  try {
    await axios.post('/api/v1/auth/set-password', { password: newPassword.value })
    ElMessage.success(t('settings.passwordSet')); showPasswordDialog.value = false; newPassword.value = ''; passwordSet.value = true
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { settingPassword.value = false }
}

async function createBackup() {
  backingUp.value = true
  try { await axios.post('/api/v1/backups', { notes: '手动备份' }); ElMessage.success(t('settings.backupCreated')) }
  catch { ElMessage.error(t('settings.backupFailed')) }
  finally { backingUp.value = false }
}

async function uploadBackup(file: File) {
  const formData = new FormData(); formData.append('file', file)
  try { await axios.post('/api/v1/backups/upload', formData); ElMessage.success(t('settings.backupRestored')) }
  catch { ElMessage.error(t('settings.restoreFailed')) }
  return false
}
</script>

<style scoped>
<<<<<<< HEAD
.page-header {
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.page-desc {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.settings-view :deep(.el-descriptions) {
  border-radius: 8px;
  overflow: hidden;
}
.settings-view :deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}
=======
.settings-page { padding: 28px; max-width: 1000px; margin: 0 auto; }
.page-header { margin-bottom: 24px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: #e6edf3; margin: 0; }
.settings-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(440px, 1fr)); gap: 16px; }
.settings-card { background: #161b22; border: 1px solid #21262d; border-radius: 12px; padding: 20px; }
.card-title { font-size: 16px; font-weight: 600; color: #e6edf3; margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid #21262d; }
.license-status .status-row { display: flex; justify-content: space-between; align-items: flex-start; padding: 8px 0; }
.status-label { color: #7d8590; font-size: 13px; }
.status-value { font-size: 13px; color: #e6edf3; }
.status-value.pro { color: #00d4aa; font-weight: 600; }
.feature-tags { display: flex; flex-wrap: wrap; gap: 4px; max-width: 280px; justify-content: flex-end; }
.ftag { padding: 1px 8px; background: rgba(0,212,170,0.12); color: #00d4aa; border-radius: 4px; font-size: 11px; }
.license-free { text-align: center; }
.free-badge { font-size: 18px; font-weight: 600; color: #7d8590; margin-bottom: 8px; }
.free-desc { color: #7d8590; font-size: 13px; margin-bottom: 16px; }
.pricing { display: flex; gap: 12px; margin-bottom: 16px; }
.price-card { flex: 1; background: #0d1117; border: 1px solid #21262d; border-radius: 10px; padding: 16px; position: relative; }
.price-card.featured { border-color: rgba(0,212,170,0.4); background: rgba(0,212,170,0.04); }
.price-badge { position: absolute; top: -8px; right: 12px; background: #00d4aa; color: #0d1117; font-size: 10px; padding: 2px 8px; border-radius: 8px; font-weight: 600; }
.price-name { font-size: 14px; font-weight: 600; color: #e6edf3; margin-bottom: 4px; }
.price-amount { font-size: 24px; font-weight: 700; color: #00d4aa; }
.price-features { list-style: none; padding: 0; margin: 8px 0 0; font-size: 12px; color: #7d8590; line-height: 1.8; text-align: left; }
.activate-section { display: flex; gap: 8px; justify-content: center; }
.license-input { width: 260px; }
.setting-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 0; border-bottom: 1px solid #21262d; font-size: 14px; color: #e6edf3; }
.setting-desc { color: #7d8590; font-size: 13px; }
@media (max-width: 768px) { .settings-grid { grid-template-columns: 1fr; } .pricing { flex-direction: column; } }
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
</style>
