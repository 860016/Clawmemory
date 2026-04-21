<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>⚙️ {{ $t('settings.title') }}</h1>
    </div>

    <div class="settings-grid">
      <!-- 授权管理 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'license' }" id="settings-license">
        <div class="card-title">🔑 {{ $t('settings.license') }}</div>
        <div class="license-status" v-if="license.active">
          <div class="status-row">
            <span class="status-label">{{ $t('settings.version') }}</span>
            <span class="status-value pro">{{ license.type === 'enterprise' ? 'Enterprise' : 'Pro' }}</span>
          </div>
          <div class="status-row" v-if="license.license_key">
            <span class="status-label">{{ $t('settings.licenseKey') || 'License Key' }}</span>
            <span class="status-value">{{ license.license_key }}</span>
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
          <el-button type="danger" plain size="small" @click="deactivateLicense" style="margin-top: 12px">{{ $t('settings.cancelLicense') }}</el-button>
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
          <div v-if="proInstallStatus" class="pro-install-status">
            <div class="status-text">{{ proInstallStatus }}</div>
            <el-progress v-if="proInstalling" :percentage="proInstallProgress" :stroke-width="6" />
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
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'security' }" id="settings-security">
        <div class="card-title">◇ {{ $t('settings.security') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.password') }}</span>
          <el-button size="small" @click="showPasswordDialog = true">
            {{ passwordSet ? $t('settings.changePassword') : $t('settings.setPassword') }}
          </el-button>
        </div>
      </div>

      <!-- 数据管理 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'data' }" id="settings-data">
        <div class="card-title">💾 {{ $t('settings.data') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.dbLocation') }}</span>
          <span class="setting-desc">data/clawmemory.db</span>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.createBackup') }}</span>
          <el-button size="small" type="primary" @click="createBackup" :loading="backingUp">{{ $t('settings.backupNow') }}</el-button>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.importBackup') }}</span>
          <el-upload :show-file-list="false" :before-upload="uploadBackup" accept=".zip" action="" :auto-upload="false">
            <el-button size="small" type="warning">{{ $t('settings.chooseFile') }}</el-button>
          </el-upload>
        </div>
        <!-- 备份列表 -->
        <div v-if="backups.length" class="backup-list">
          <div class="backup-item" v-for="b in backups" :key="b.id">
            <div class="backup-info">
              <span class="backup-name">{{ b.filename || $t('settings.backup') + ' #' + b.id }}</span>
              <span class="backup-meta">{{ formatBackupTime(b.created_at) }} · {{ formatSize(b.file_size) }}</span>
            </div>
            <div class="backup-actions">
              <el-button text size="small" @click="downloadBackup(b.id)">⬇</el-button>
              <el-button text size="small" @click="restoreBackup(b.id)">{{ $t('settings.restore') }}</el-button>
              <el-button text size="small" type="danger" @click="deleteBackup(b.id)">✕</el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 记忆衰减设置 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'decay' }" id="settings-decay">
        <div class="card-title">🧠 {{ $t('settings.memoryDecay') || '记忆衰减' }}</div>
        <div class="decay-info" v-if="decayInfo">
          <div class="decay-stage-info">
            <div class="stage-item">
              <span class="stage-label">15天未访问</span>
              <span class="stage-desc">轻度衰减 10%</span>
            </div>
            <div class="stage-item">
              <span class="stage-label">30天未访问</span>
              <span class="stage-desc">中度衰减 30%，标记为不重要</span>
            </div>
            <div class="stage-item">
              <span class="stage-label">60天未访问</span>
              <span class="stage-desc">进入回收站</span>
            </div>
            <div class="stage-item">
              <span class="stage-label">回收站保留</span>
              <span class="stage-desc">30天后自动清空</span>
            </div>
          </div>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.autoDecay') || '自动衰减' }}</span>
          <el-switch v-model="decayEnabled" @change="updateDecaySettings" :loading="decayLoading" />
        </div>
        <div class="decay-stats" v-if="decayStats">
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.totalMemories') || '总记忆' }}</span>
            <span class="stats-value">{{ decayStats.total }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.activeMemories') || '正常记忆' }}</span>
            <span class="stats-value">{{ decayStats.active }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.archivedMemories') || '不重要记忆' }}</span>
            <span class="stats-value warning">{{ decayStats.archived }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.trashedMemories') || '回收站' }}</span>
            <span class="stats-value danger">{{ decayStats.trashed }}</span>
          </div>
        </div>
        <div class="decay-actions" v-if="decayStats && decayStats.trashed > 0">
          <el-button size="small" type="warning" @click="viewTrash">{{ $t('settings.viewTrash') || '查看回收站' }}</el-button>
          <el-button size="small" type="danger" @click="emptyTrash">{{ $t('settings.emptyTrash') || '清空回收站' }}</el-button>
        </div>
      </div>

      <!-- 系统信息 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'system' }" id="settings-system">
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

    <el-dialog v-model="showPasswordDialog" :title="passwordSet ? $t('settings.changePassword') : $t('settings.setPassword')" width="400px">
      <el-form label-position="top">
        <el-form-item v-if="passwordSet" :label="$t('settings.oldPassword')">
          <el-input v-model="oldPassword" type="password" show-password :placeholder="$t('settings.oldPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('settings.newPassword')">
          <el-input v-model="newPassword" type="password" show-password :placeholder="$t('settings.passwordMinLen')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSetPassword" :loading="settingPassword">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from '../api/client'
import { setLocale, getLocale } from '../i18n'

const { t } = useI18n()
const route = useRoute()
const activeSection = ref((route.query.section as string) || '')
const license = ref<any>({ active: false, tier: 'oss', features: [] })
const licenseKey = ref('')
const activating = ref(false)
const proInstalling = ref(false)
const proInstallProgress = ref(0)
const proInstallStatus = ref('')
const backingUp = ref(false)
const passwordSet = ref(false)
const showPasswordDialog = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const settingPassword = ref(false)
const coreEngine = ref('python')
const currentLocale = ref(getLocale())
const appVersion = ref('2.8.2')
const backups = ref<any[]>([])
const decayEnabled = ref(false)
const decayLoading = ref(false)
const decayStats = ref<any>(null)
const decayInfo = ref<any>(null)

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
  await Promise.all([loadLicense(), loadInitStatus(), loadInstallStatus(), loadBackups(), loadDecaySettings(), loadDecayStats()])
  if (activeSection.value) {
    nextTick(() => scrollToSection(activeSection.value))
  }
})

watch(() => route.query.section, (section) => {
  if (section && typeof section === 'string') {
    activeSection.value = section
    nextTick(() => scrollToSection(section))
  }
})

function scrollToSection(section: string) {
  const el = document.getElementById(`settings-${section}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('section-highlight')
    setTimeout(() => el.classList.remove('section-highlight'), 2000)
  }
}

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

async function loadBackups() {
  try { const { data } = await axios.get('/backups'); backups.value = data || [] } catch { backups.value = [] }
}

function formatBackupTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatSize(bytes: number) {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

async function downloadBackup(id: number) {
  try {
    const { data } = await axios.get(`/backups/${id}/download`, { responseType: 'blob' })
    const url = URL.createObjectURL(data)
    const a = document.createElement('a')
    a.href = url; a.download = `backup_${id}.zip`; a.click()
    URL.revokeObjectURL(url)
  } catch { ElMessage.error(t('common.failed')) }
}

async function restoreBackup(id: number) {
  try {
    await ElMessageBox.confirm(t('settings.restoreConfirm'), t('common.confirm'), { type: 'warning' })
    backingUp.value = true
    await axios.post(`/backups/${id}/restore`)
    ElMessage.success(t('settings.backupRestored'))
    await loadBackups()
  } catch {} finally { backingUp.value = false }
}

async function deleteBackup(id: number) {
  try {
    await ElMessageBox.confirm(t('settings.deleteBackupConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.delete(`/backups/${id}`)
    ElMessage.success(t('common.success'))
    await loadBackups()
  } catch {}
}

async function activateLicense() {
  if (!licenseKey.value) return
  activating.value = true
  try {
    const { data } = await axios.post('/license/activate', { license_key: licenseKey.value })
    if (data.valid || data.active) {
      ElMessage.success(t('settings.activated'))
      licenseKey.value = ''
      await loadLicense()
      
      // 检查是否有 Pro 下载地址
      if (data.pro_download_url) {
        await installProModule(data.pro_download_url, data.pro_fallback_urls || [])
      } else {
        // 授权服务器未配置下载地址，提示用户
        ElMessage.info('授权已激活，Pro 模块暂未配置下载地址，请联系管理员')
      }
    } else {
      ElMessage.error(data.message || t('common.failed'))
    }
  } catch (e: any) {
    const detail = e.response?.data?.detail
    if (typeof detail === 'string') {
      ElMessage.error(detail)
    } else {
      ElMessage.error(t('common.failed'))
    }
  } finally { activating.value = false }
}

async function installProModule(url: string, fallbackUrls: string[]) {
  proInstalling.value = true
  proInstallProgress.value = 0
  proInstallStatus.value = '正在下载 Pro 模块...'
  
  try {
    const fallbackParam = fallbackUrls.length > 0 ? fallbackUrls.join(',') : ''
    const { data } = await axios.post('/license/pro/install', null, {
      params: { url, fallback_urls: fallbackParam },
      timeout: 120000,
      onDownloadProgress: (e) => {
        if (e.total) {
          proInstallProgress.value = Math.round((e.loaded / e.total) * 100)
        }
      }
    })
    
    if (data.success) {
      ElMessage.success('Pro 模块安装成功')
      proInstallStatus.value = '安装完成'
      // 刷新授权状态以显示 Pro 功能
      await loadLicense()
    } else {
      ElMessage.error(data.message || 'Pro 模块安装失败')
      proInstallStatus.value = '安装失败'
    }
  } catch (e: any) {
    const detail = e.response?.data?.detail || e.message
    ElMessage.error(`Pro 模块安装失败: ${detail}`)
    proInstallStatus.value = '安装失败'
  } finally {
    proInstalling.value = false
    setTimeout(() => { proInstallStatus.value = '' }, 3000)
  }
}

async function deactivateLicense() {
  try {
    await ElMessageBox.confirm(t('settings.cancelConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.post('/license/deactivate')
    ElMessage.success(t('settings.canceled')); await loadLicense()
  } catch {}
}

async function handleSetPassword() {
  if (newPassword.value.length < 4) { ElMessage.warning(t('settings.passwordMinLen')); return }
  settingPassword.value = true
  try {
    if (passwordSet.value) {
      // 修改密码 — 验证旧密码
      await axios.post('/auth/change-password', { old_password: oldPassword.value, new_password: newPassword.value })
    } else {
      // 首次设置密码
      await axios.post('/auth/set-password', { password: newPassword.value })
    }
    ElMessage.success(t('settings.passwordSet')); showPasswordDialog.value = false; oldPassword.value = ''; newPassword.value = ''; passwordSet.value = true
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { settingPassword.value = false }
}

async function createBackup() {
  backingUp.value = true
  try { await axios.post('/backups', { notes: '手动备份' }); ElMessage.success(t('settings.backupCreated')); await loadBackups() }
  catch { ElMessage.error(t('settings.backupFailed')) }
  finally { backingUp.value = false }
}

async function uploadBackup(file: File) {
  const formData = new FormData(); formData.append('file', file)
  backingUp.value = true
  try {
    const { data } = await axios.post('/backups/upload', formData)
    if (data && data.id) {
      await axios.post(`/backups/${data.id}/restore`)
    }
    ElMessage.success(t('settings.backupRestored'))
    await loadLicense()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('settings.restoreFailed'))
  } finally {
    backingUp.value = false
  }
  return false
}

async function loadDecaySettings() {
  try {
    const { data } = await axios.get('/memories/decay/settings')
    decayEnabled.value = data.enabled
    decayInfo.value = data
  } catch {}
}

async function loadDecayStats() {
  try {
    const { data } = await axios.get('/memories/decay/stats')
    decayStats.value = data.stats
  } catch {}
}

async function updateDecaySettings() {
  decayLoading.value = true
  try {
    await axios.post('/memories/decay/settings', null, { params: { enabled: decayEnabled.value } })
    ElMessage.success(decayEnabled.value ? '自动衰减已开启' : '自动衰减已关闭')
  } catch {
    ElMessage.error(t('common.failed'))
  } finally {
    decayLoading.value = false
  }
}

async function viewTrash() {
  window.location.href = '/memories?status=trashed'
}

async function emptyTrash() {
  try {
    await ElMessageBox.confirm('确定要清空回收站吗？此操作不可恢复。', '确认', { type: 'warning' })
    await axios.delete('/memories/trash')
    ElMessage.success('回收站已清空')
    await loadDecayStats()
  } catch {}
}
</script>

<style scoped>
.settings-page { padding: 28px; max-width: 1000px; margin: 0 auto; }
.page-header { margin-bottom: 24px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.settings-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(440px, 1fr)); gap: 16px; }
.settings-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; transition: border-color 0.3s, box-shadow 0.3s; }
.settings-card.section-highlight { border-color: #10B981; box-shadow: 0 0 0 2px rgba(16,185,129,0.2); }
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
.pro-install-status { margin-top: 12px; padding: 12px; background: rgba(16,185,129,0.05); border-radius: 8px; }
.status-text { font-size: 13px; color: var(--cm-text); margin-bottom: 8px; }
.setting-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 0; border-bottom: 1px solid var(--cm-border); font-size: 14px; color: var(--cm-text); }
.setting-desc { color: var(--cm-text-muted); font-size: 13px; }
.backup-list { margin-top: 12px; border-top: 1px solid var(--cm-border); padding-top: 8px; max-height: 200px; overflow-y: auto; }
.backup-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid var(--cm-border); }
.backup-item:last-child { border-bottom: none; }
.backup-info { display: flex; flex-direction: column; gap: 2px; }
.backup-name { font-size: 13px; color: var(--cm-text); font-weight: 500; }
.backup-meta { font-size: 11px; color: var(--cm-text-muted); }
.backup-actions { display: flex; gap: 4px; }
@media (max-width: 768px) { .settings-grid { grid-template-columns: 1fr; } .pricing { flex-direction: column; } }
.decay-info { margin-bottom: 16px; }
.decay-stage-info { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
.stage-item { display: flex; justify-content: space-between; padding: 8px 12px; background: var(--cm-bg); border-radius: 8px; font-size: 12px; }
.stage-label { color: var(--cm-text-muted); }
.stage-desc { color: var(--cm-text); font-weight: 500; }
.decay-stats { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--cm-border); }
.stats-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 13px; }
.stats-label { color: var(--cm-text-muted); }
.stats-value { color: var(--cm-text); font-weight: 500; }
.stats-value.warning { color: #ffc107; }
.stats-value.danger { color: #e91e63; }
.decay-actions { margin-top: 12px; display: flex; gap: 8px; }
</style>
