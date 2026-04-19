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
          <span class="setting-desc">{{ appVersion }}</span>
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
const appVersion = ref('2.1.0')

const featureLabels: Record<string, string> = {
  ai_extract: t('settings.featAiExtract'),
  auto_graph: t('settings.featAutoGraph'),
  unlimited_graph: t('settings.featUnlimitedGraph'),
  auto_decay: t('settings.featAutoDecay'),
  decay_report: t('settings.featDecayReport'),
  prune_suggest: t('settings.featPruneSuggest'),
  reinforce: t('settings.featReinforce'),
  conflict_scan: t('settings.featConflictScan'),
  conflict_merge: t('settings.featConflictMerge'),
  smart_router: t('settings.featSmartRouter'),
  token_stats: t('settings.featTokenStats'),
  wiki: t('settings.featWiki'),
  auto_backup: t('settings.featAutoBackup'),
  // Enterprise only
  api_access: t('settings.featApiAccess'),
  sso: t('settings.featSso'),
  audit_log: t('settings.featAuditLog'),
  time_travel: t('settings.featTimeTravel'),
  offline_mode: t('settings.featOfflineMode'),
}

onMounted(async () => {
  await Promise.all([loadLicense(), loadInitStatus(), loadInstallStatus()])
})

function changeLocale(locale: 'zh' | 'en') {
  setLocale(locale)
}

async function loadLicense() {
  try { const { data } = await axios.get('/license/info'); license.value = data } catch {}
}

async function loadInitStatus() {
  try { const { data } = await axios.get('/auth/init-status'); passwordSet.value = data.password_set } catch {}
}

async function loadInstallStatus() {
  try { const { data } = await axios.get('/install-status'); coreEngine.value = data.checks?.security_engine || 'python'; if (data.version) appVersion.value = data.version } catch {}
}

async function activateLicense() {
  if (!licenseKey.value) return
  activating.value = true
  try {
    const { data } = await axios.post('/license/activate', { license_key: licenseKey.value })
    if (data.valid || data.active) { ElMessage.success(t('settings.activated')); await loadLicense() }
    else ElMessage.error(data.message || t('common.failed'))
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { activating.value = false }
}

async function deactivateLicense() {
  try {
    await ElMessageBox.confirm(t('settings.cancelConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.post('/license/deactivate')
    ElMessage.success(t('settings.canceled')); await loadLicense()
  } catch {}
}

async function setPassword() {
  if (newPassword.value.length < 4) { ElMessage.warning(t('settings.passwordMinLen')); return }
  settingPassword.value = true
  try {
    await axios.post('/auth/set-password', { password: newPassword.value })
    ElMessage.success(t('settings.passwordSet')); showPasswordDialog.value = false; newPassword.value = ''; passwordSet.value = true
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { settingPassword.value = false }
}

async function createBackup() {
  backingUp.value = true
  try { await axios.post('/backups', { notes: '手动备份' }); ElMessage.success(t('settings.backupCreated')) }
  catch { ElMessage.error(t('settings.backupFailed')) }
  finally { backingUp.value = false }
}

async function uploadBackup(file: File) {
  const formData = new FormData(); formData.append('file', file)
  try { await axios.post('/backups/upload', formData); ElMessage.success(t('settings.backupRestored')) }
  catch { ElMessage.error(t('settings.restoreFailed')) }
  return false
}
</script>

<style scoped>
.settings-page { padding: 28px; max-width: 1000px; margin: 0 auto; }
.page-header { margin-bottom: 24px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.settings-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(440px, 1fr)); gap: 16px; }
.settings-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; }
.card-title { font-size: 16px; font-weight: 600; color: var(--cm-text); margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid var(--cm-border); }
.license-status .status-row { display: flex; justify-content: space-between; align-items: flex-start; padding: 8px 0; }
.status-label { color: var(--cm-text-muted); font-size: 13px; }
.status-value { font-size: 13px; color: var(--cm-text); }
.status-value.pro { color: #10B981; font-weight: 600; }
.feature-tags { display: flex; flex-wrap: wrap; gap: 4px; max-width: 280px; justify-content: flex-end; }
.ftag { padding: 1px 8px; background: rgba(16,185,129,0.12); color: #10B981; border-radius: 4px; font-size: 11px; }
.license-free { text-align: center; }
.free-badge { font-size: 18px; font-weight: 600; color: var(--cm-text-muted); margin-bottom: 8px; }
.free-desc { color: var(--cm-text-muted); font-size: 13px; margin-bottom: 16px; }
.pricing { display: flex; gap: 12px; margin-bottom: 16px; }
.price-card { flex: 1; background: var(--cm-bg); border: 1px solid var(--cm-border); border-radius: 10px; padding: 16px; position: relative; }
.price-card.featured { border-color: rgba(16,185,129,0.4); background: rgba(16,185,129,0.04); }
.price-badge { position: absolute; top: -8px; right: 12px; background: #10B981; color: var(--cm-bg); font-size: 10px; padding: 2px 8px; border-radius: 8px; font-weight: 600; }
.price-name { font-size: 14px; font-weight: 600; color: var(--cm-text); margin-bottom: 4px; }
.price-amount { font-size: 24px; font-weight: 700; color: #10B981; }
.price-features { list-style: none; padding: 0; margin: 8px 0 0; font-size: 12px; color: var(--cm-text-muted); line-height: 1.8; text-align: left; }
.activate-section { display: flex; gap: 8px; justify-content: center; }
.license-input { width: 260px; }
.setting-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 0; border-bottom: 1px solid var(--cm-border); font-size: 14px; color: var(--cm-text); }
.setting-desc { color: var(--cm-text-muted); font-size: 13px; }
@media (max-width: 768px) { .settings-grid { grid-template-columns: 1fr; } .pricing { flex-direction: column; } }
</style>
