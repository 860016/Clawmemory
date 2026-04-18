<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>{{ $t('settings.title') }}</h1>
    </div>

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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from '../api/client'
import { setLocale, getLocale } from '../i18n'

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
</style>
