<template>
  <div class="sharing-view">
    <div class="page-header">
      <h2>{{ $t('sharing.title') }}</h2>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane :label="$t('sharing.pendingShares')" name="pending">
        <div v-if="pendingLoading" style="text-align: center; padding: 40px">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
        </div>
        <div v-else-if="pendingShares.length === 0" class="empty-state">
          <el-empty :description="$t('sharing.noPendingShares')" />
        </div>
        <div v-else class="share-list">
          <div v-for="share in pendingShares" :key="share.id" class="share-card">
            <div class="share-header">
              <span class="share-from">{{ $t('sharing.fromAgent') }}: {{ share.from_agent }}</span>
              <span class="share-arrow">→</span>
              <span class="share-to">{{ $t('sharing.toAgent') }}: {{ share.to_agent }}</span>
              <el-tag size="small" type="warning">{{ $t('sharing.shareType' + capitalize(share.share_type)) }}</el-tag>
            </div>
            <div class="share-preview">
              <div class="memory-key">{{ share.memory?.key }}</div>
              <div class="memory-value">{{ truncateValue(share.memory?.value) }}</div>
            </div>
            <div class="share-actions">
              <el-button size="small" type="success" @click="approveShare(share.id)">{{ $t('sharing.approve') }}</el-button>
              <el-button size="small" type="danger" @click="rejectShare(share.id)">{{ $t('sharing.reject') }}</el-button>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('sharing.outboundShares')" name="outbound">
        <div v-if="outboundLoading" style="text-align: center; padding: 40px">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
        </div>
        <div v-else-if="outboundShares.length === 0" class="empty-state">
          <el-empty :description="$t('sharing.noOutboundShares')" />
        </div>
        <div v-else class="share-list">
          <div v-for="share in outboundShares" :key="share.id" class="share-card">
            <div class="share-header">
              <span class="share-from">{{ $t('sharing.fromAgent') }}: {{ share.from_agent }}</span>
              <span class="share-arrow">→</span>
              <span class="share-to">{{ $t('sharing.toAgent') }}: {{ share.to_agent }}</span>
              <el-tag size="small" :type="share.status === 'approved' ? 'success' : share.status === 'rejected' ? 'danger' : 'warning'">
                {{ shareStatusLabel(share.status) }}
              </el-tag>
            </div>
            <div class="share-preview">
              <div class="memory-key">{{ share.memory?.key }}</div>
              <div class="memory-value">{{ truncateValue(share.memory?.value) }}</div>
            </div>
            <div class="share-actions" v-if="share.status === 'approved'">
              <el-button size="small" type="warning" @click="revokeShare(share.id)">{{ $t('sharing.revoke') }}</el-button>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('sharing.shareRules')" name="rules">
        <div class="rules-header">
          <el-button type="primary" size="small" @click="showRuleDialog = true">{{ $t('sharing.createRule') }}</el-button>
        </div>
        <div v-if="rulesLoading" style="text-align: center; padding: 40px">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
        </div>
        <div v-else-if="shareRules.length === 0" class="empty-state">
          <el-empty description="暂无自动共享规则" />
        </div>
        <div v-else class="rule-list">
          <div v-for="rule in shareRules" :key="rule.id" class="rule-card">
            <div class="rule-info">
              <span>{{ rule.from_agent }} → {{ rule.to_agent }}</span>
              <el-tag size="small" v-if="rule.layer">{{ rule.layer }}</el-tag>
              <el-tag size="small" :type="rule.enabled ? 'success' : 'info'">
                {{ rule.enabled ? $t('sharing.ruleEnabled') : $t('sharing.ruleDisabled') }}
              </el-tag>
            </div>
            <div class="rule-actions">
              <el-button size="small" @click="editRule(rule)">{{ $t('sharing.editRule') }}</el-button>
              <el-button size="small" type="danger" @click="deleteRule(rule.id)">{{ $t('sharing.deleteRule') }}</el-button>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showRuleDialog" :title="editingRule ? $t('sharing.editRule') : $t('sharing.createRule')" width="480px">
      <el-form label-width="100px">
        <el-form-item :label="$t('sharing.ruleFrom')">
          <el-input v-model="ruleForm.from_agent" />
        </el-form-item>
        <el-form-item :label="$t('sharing.ruleTo')">
          <el-input v-model="ruleForm.to_agent" />
        </el-form-item>
        <el-form-item :label="$t('sharing.ruleLayer')">
          <el-select v-model="ruleForm.layer" clearable style="width: 100%">
            <el-option label="Knowledge" value="knowledge" />
            <el-option label="Preference" value="preference" />
            <el-option label="Short Term" value="short_term" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('sharing.ruleVisibility')">
          <el-select v-model="ruleForm.target_visibility" style="width: 100%">
            <el-option :label="$t('memories.visibilityShared')" value="shared" />
            <el-option :label="$t('memories.visibilityPublic')" value="public" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('sharing.ruleEnabled')">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRuleDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveRule" :loading="ruleSaving">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { sharingApi } from '../api/go-sharing'

const { t } = useI18n()

const activeTab = ref('pending')
const pendingShares = ref<any[]>([])
const outboundShares = ref<any[]>([])
const shareRules = ref<any[]>([])
const pendingLoading = ref(false)
const outboundLoading = ref(false)
const rulesLoading = ref(false)
const showRuleDialog = ref(false)
const ruleSaving = ref(false)
const editingRule = ref<any>(null)

const ruleForm = ref({
  from_agent: '',
  to_agent: '',
  layer: '',
  target_visibility: 'shared',
  enabled: true,
})

function capitalize(s: string) {
  return s ? s.charAt(0).toUpperCase() + s.slice(1) : ''
}

function truncateValue(v?: string) {
  if (!v) return ''
  return v.length > 120 ? v.substring(0, 120) + '...' : v
}

function shareStatusLabel(status: string) {
  const map: Record<string, string> = {
    approved: t('sharing.approved'),
    rejected: t('sharing.rejected'),
    revoked: t('sharing.revoked'),
    pending: t('sharing.pendingShares'),
  }
  return map[status] || status
}

async function loadPendingShares() {
  pendingLoading.value = true
  try {
    const { data } = await sharingApi.getPendingShares()
    pendingShares.value = data.shares || data || []
  } catch {
    pendingShares.value = []
  } finally {
    pendingLoading.value = false
  }
}

async function loadOutboundShares() {
  outboundLoading.value = true
  try {
    const { data } = await sharingApi.getOutboundShares()
    outboundShares.value = data.shares || data || []
  } catch {
    outboundShares.value = []
  } finally {
    outboundLoading.value = false
  }
}

async function loadShareRules() {
  rulesLoading.value = true
  try {
    const { data } = await sharingApi.listRules()
    shareRules.value = data.rules || data || []
  } catch {
    shareRules.value = []
  } finally {
    rulesLoading.value = false
  }
}

async function approveShare(id: number) {
  try {
    await sharingApi.approveShare(id)
    ElMessage.success(t('sharing.approved'))
    await loadPendingShares()
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

async function rejectShare(id: number) {
  try {
    await sharingApi.rejectShare(id)
    ElMessage.success(t('sharing.rejected'))
    await loadPendingShares()
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

async function revokeShare(id: number) {
  try {
    await sharingApi.revokeShare(id)
    ElMessage.success(t('sharing.revoked'))
    await loadOutboundShares()
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

function editRule(rule: any) {
  editingRule.value = rule
  ruleForm.value = {
    from_agent: rule.from_agent || '',
    to_agent: rule.to_agent || '',
    layer: rule.layer || '',
    target_visibility: rule.target_visibility || 'shared',
    enabled: rule.enabled !== false,
  }
  showRuleDialog.value = true
}

async function saveRule() {
  ruleSaving.value = true
  try {
    if (editingRule.value) {
      await sharingApi.updateRule(editingRule.value.id, ruleForm.value)
    } else {
      await sharingApi.createRule(ruleForm.value)
    }
    ElMessage.success(t('common.success'))
    showRuleDialog.value = false
    editingRule.value = null
    await loadShareRules()
  } catch {
    ElMessage.error(t('common.failed'))
  } finally {
    ruleSaving.value = false
  }
}

async function deleteRule(id: number) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), { type: 'warning' })
  } catch { return }
  try {
    await sharingApi.deleteRule(id)
    ElMessage.success(t('common.success'))
    await loadShareRules()
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

onMounted(() => {
  loadPendingShares()
  loadOutboundShares()
  loadShareRules()
})
</script>

<style scoped>
.sharing-view { padding: 20px; max-width: 900px; margin: 0 auto; }
.page-header { margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 20px; }
.share-list { display: flex; flex-direction: column; gap: 12px; }
.share-card { padding: 16px; background: var(--cm-bg-secondary); border-radius: 8px; border: 1px solid var(--cm-border); }
.share-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; font-size: 13px; }
.share-arrow { color: var(--cm-text-muted); }
.share-preview { padding: 8px; background: var(--cm-bg); border-radius: 4px; margin-bottom: 8px; }
.memory-key { font-weight: 600; font-size: 13px; margin-bottom: 4px; }
.memory-value { font-size: 12px; color: var(--cm-text-muted); word-break: break-all; }
.share-actions { display: flex; gap: 8px; }
.empty-state { padding: 40px 0; }
.rules-header { margin-bottom: 16px; display: flex; justify-content: flex-end; }
.rule-list { display: flex; flex-direction: column; gap: 8px; }
.rule-card { display: flex; justify-content: space-between; align-items: center; padding: 12px; background: var(--cm-bg-secondary); border-radius: 8px; border: 1px solid var(--cm-border); }
.rule-info { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.rule-actions { display: flex; gap: 8px; }
</style>
